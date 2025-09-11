package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	broker := envOr("KAFKA_BROKERS", "kafka:9092")
	topic := envOr("KAFKA_TOPIC", "orders")

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	order := map[string]any{
		"order_uid":    "b563feb7b2b84b6test",
		"track_number": "WBILMTESTTRACK",
		"entry":        "WBIL",
		"delivery": map[string]any{
			"name": "Test Testov", "phone": "+9720000000", "zip": "2639809",
			"city": "Kiryat Mozkin", "address": "Ploshad Mira 15",
			"region": "Kraiot", "email": "test@gmail.com",
		},
		"payment": map[string]any{
			"transaction": "b563feb7b2b84b6test", "request_id": "", "currency": "USD",
			"provider": "wbpay", "amount": 1817, "payment_dt": 1637907727, "bank": "alpha",
			"delivery_cost": 1500, "goods_total": 317, "custom_fee": 0,
		},
		"items": []map[string]any{
			{"chrt_id": 9934930, "track_number": "WBILMTESTTRACK", "price": 453,
				"rid": "ab4219087a764ae0btest", "name": "Mascaras", "sale": 30,
				"size": "0", "total_price": 317, "nm_id": 2389212, "brand": "Vivienne Sabo", "status": 202},
		},
		"locale": "en", "internal_signature": "", "customer_id": "test",
		"delivery_service": "meest", "shardkey": "9", "sm_id": 99,
		"date_created": "2021-11-26T06:22:19Z", "oof_shard": "1",
	}

	data, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	log.Printf("producer -> %s topic=%s", broker, topic)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order["order_uid"].(string)),
		Value: data,
	}); err != nil {
		log.Fatalf("failed write: %v", err)
	}
	log.Println("sent:", order["order_uid"])
}
