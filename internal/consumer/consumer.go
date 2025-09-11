package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"yourmodule/internal/cache"
	"yourmodule/internal/db"
	"yourmodule/internal/models"
)

type Consumer struct {
	reader *kafka.Reader
	store  *db.Store
	cache  *cache.Cache
}

func New(brokerAddr, topic, groupID string, store *db.Store, c *cache.Cache) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: r, store: store, cache: c}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				log.Println("consumer context canceled, exiting")
				return
			}
			log.Printf("fetch message error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var ord models.Order
		if err := json.Unmarshal(m.Value, &ord); err != nil {
			log.Printf("invalid message JSON: %v; raw: %s", err, string(m.Value))
			if err2 := c.reader.CommitMessages(ctx, m); err2 != nil {
				log.Printf("failed commit invalid message: %v", err2)
			}
			continue
		}

		if err := c.store.SaveOrder(ctx, ord, m.Value); err != nil {
			log.Printf("db save error: %v", err)
			if err2 := c.reader.CommitMessages(ctx, m); err2 != nil {
				log.Printf("failed commit after db error: %v", err2)
			}
			continue
		}

		c.cache.Set(ord.OrderUID, ord)

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("failed commit message: %v", err)
		}
	}
}
