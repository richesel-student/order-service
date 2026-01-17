package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	"yourmodule/internal/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

// envOr возвращает значение переменной окружения или дефолт
func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// envFloat возвращает float из окружения или дефолт
func envFloat(k string, d float64) float64 {
	if v := os.Getenv(k); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return d
}

// envInt возвращает int из окружения или дефолт
func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}

func main() {
	broker := envOr("KAFKA_BROKERS", "localhost:9092")
	topic := envOr("KAFKA_TOPIC", "orders")
	count := envInt("COUNT", 10)
	invalidRate := envFloat("INVALID_RATE", 0.2) // 20% сообщений будут "сломаны"

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	gofakeit.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())

	log.Printf("producer -> %s topic=%s count=%d invalid_rate=%.2f", broker, topic, count, invalidRate)

	for i := 0; i < count; i++ {
		uid := gofakeit.UUID()
		makeInvalid := rand.Float64() < invalidRate
		order := fakeOrder(uid, makeInvalid)

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("marshal error: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(uid),
			Value: data,
		})
		cancel()

		if err != nil {
			log.Printf("failed send uid=%s: %v", uid, err)
			continue
		}

		if makeInvalid {
			log.Printf("sent INVALID: %s", uid)
		} else {
			log.Printf("sent: %s", uid)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func fakeOrder(uid string, makeInvalid bool) models.Order {
	track := gofakeit.LetterN(10)
	order := models.Order{
		OrderUID:    uid,
		TrackNumber: track,
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  uid,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 5000),
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: gofakeit.Number(50, 500),
			GoodsTotal:   gofakeit.Number(50, 5000),
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      gofakeit.Number(1, 999999),
				TrackNumber: track,
				Price:       gofakeit.Number(10, 1000),
				RID:         gofakeit.UUID(),
				Name:        gofakeit.Word(),
				Sale:        gofakeit.Number(0, 50),
				Size:        "0",
				TotalPrice:  gofakeit.Number(10, 1000),
				NmID:        gofakeit.Number(1, 999999),
				Brand:       gofakeit.Company(),
				Status:      200,
			},
		},
		Locale:          "en",
		CustomerID:      gofakeit.Username(),
		DeliveryService: "meest",
		ShardKey:        "1",
		SmID:            1,
		DateCreated:     time.Now().UTC().Format(time.RFC3339),
		OofShard:        "1",
	}

	if makeInvalid {
		// ломаем данные для теста валидатора
		switch gofakeit.Number(1, 4) {
		case 1:
			order.Payment.Currency = "RUB" // неправильная валюта
		case 2:
			order.Payment.Amount = -10 // отрицательная сумма
		case 3:
			order.Items = []models.Item{} // пустые items
		default:
			order.OrderUID = "" // пустой order_uid
		}
	}

	return order
}
