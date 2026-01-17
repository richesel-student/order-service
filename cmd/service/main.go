package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yourmodule/internal/api"
	"yourmodule/internal/cache"
	"yourmodule/internal/consumer"
	"yourmodule/internal/db"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Конфиг
	pgDSN := envOr("PG_DSN", "postgres://"+os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@"+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+"/"+os.Getenv("DB_NAME")+"?sslmode=disable")
	kafkaBroker := envOr("KAFKA_BROKERS", "kafka:9092")
	kafkaTopic := envOr("KAFKA_TOPIC", "orders")
	kafkaGroup := envOr("KAFKA_GROUP", "order-service-group")
	httpAddr := envOr("HTTP_PORT", ":8082")

	log.Printf("Config: PG_DSN=%s", pgDSN)
	log.Printf("Config: Kafka=%s Topic=%s Group=%s", kafkaBroker, kafkaTopic, kafkaGroup)
	log.Printf("Config: HTTP=%s", httpAddr)

	// Подключение к БД
	var store *db.Store
	var err error
	for i := 0; i < 20; i++ {
		store, err = db.NewStore(ctx, pgDSN)
		if err == nil {
			break
		}
		log.Printf("DB connect failed (try %d/20): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer store.Close()

	// Кэш
	cacheTTL := 5 * time.Minute
	cleanupInterval := 1 * time.Minute
	c := cache.New(cacheTTL, cleanupInterval)

	// Прогрев кэша
	go warmCache(ctx, store, c, cacheTTL)

	// Kafka consumer через обёртку kafkaReaderWrapper
	go runConsumerWithRetry(ctx, func() *consumer.Consumer {
		reader := &consumer.KafkaReaderWrapper{
			R: kafka.NewReader(kafka.ReaderConfig{
				Brokers: []string{kafkaBroker},
				Topic:   kafkaTopic,
				GroupID: kafkaGroup,
			}),
		}
		return consumer.New(reader, store, c)
	})

	// HTTP сервер
	srv := api.NewServer(store, c)
	httpSrv := &http.Server{
		Addr:         httpAddr,
		Handler:      srv.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("HTTP listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	cancel()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = httpSrv.Shutdown(ctxShutdown)

	log.Println("done")
}

// envOr возвращает переменную окружения или дефолт
func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// runConsumerWithRetry запускает consumer с бэкоффом при падении
func runConsumerWithRetry(ctx context.Context, create func() *consumer.Consumer) {
	backoff := 500 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		cons := create()
		log.Println("starting kafka consumer")
		cons.Run(ctx)
		_ = cons.Close()

		log.Println("consumer stopped, retrying...")
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}
}

// warmCache прогревает кэш
func warmCache(ctx context.Context, store *db.Store, c *cache.Cache, cacheTTL time.Duration) {
	ctxWarm, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Println("starting cache warmup...")

	backoff := 300 * time.Millisecond
	attempt := 1

	for {
		loaded, err := store.LoadAllOrders(ctxWarm, 1000)
		if err == nil {
			for k, v := range loaded {
				c.Set(k, v, cacheTTL)
			}
			log.Printf("cache warmed with %d orders", len(loaded))
			return
		}

		if ctxWarm.Err() != nil {
			log.Printf("warn: cache warmup stopped: %v", ctxWarm.Err())
			return
		}

		log.Printf("warn: warmup failed (attempt %d): %v", attempt, err)
		attempt++
		time.Sleep(backoff)
		if backoff < 3*time.Second {
			backoff *= 2
		}
	}
}
