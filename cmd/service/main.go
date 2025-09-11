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
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgDSN := envOr("PG_DSN", "postgres://"+os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@"+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+"/"+os.Getenv("DB_NAME")+"?sslmode=disable")
	kafkaBroker := envOr("KAFKA_BROKERS", "kafka:9092")
	kafkaTopic := envOr("KAFKA_TOPIC", "orders")
	kafkaGroup := envOr("KAFKA_GROUP", "order-service-group")
	httpAddr := envOr("HTTP_PORT", ":8082")

	log.Printf("Config: PG_DSN=%s", pgDSN)
	log.Printf("Config: Kafka=%s Topic=%s Group=%s", kafkaBroker, kafkaTopic, kafkaGroup)
	log.Printf("Config: HTTP=%s", httpAddr)

	var store *db.Store
	var err error
	for i := 1; i <= 20; i++ {
		store, err = db.NewStore(ctx, pgDSN)
		if err == nil {
			break
		}
		log.Printf("db connect failed (try %d/20): %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer store.Close()

	c := cache.New()
	loaded, err := store.LoadAllOrders(ctx, 1000)
	if err != nil {
		log.Printf("warn: failed load orders for cache: %v", err)
	} else {
		for k, v := range loaded {
			c.Set(k, v)
		}
		log.Printf("cache warmed with %d orders", len(loaded))
	}

	cons := consumer.New(kafkaBroker, kafkaTopic, kafkaGroup, store, c)
	go cons.Run(ctx)
	defer cons.Close()

	srv := api.NewServer(store, c)
	httpSrv := &http.Server{
		Addr:         httpAddr,
		Handler:      srv.Routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("http listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	cancel()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = httpSrv.Shutdown(ctxShutdown)
	log.Println("done.")
}

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
