package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"yourmodule/internal/models"

	"github.com/segmentio/kafka-go"
)

// Message обёртка Kafka-сообщения
type Message struct {
	Value     []byte
	Partition int
	Offset    int64
}

// OrderStore интерфейс для работы с БД
type OrderStore interface {
	SaveOrder(ctx context.Context, ord models.Order, raw []byte) error
	SaveBadMessage(ctx context.Context, raw []byte, errText string) error
}

// Cache интерфейс для кэша
type Cache interface {
	Set(key string, value any, ttl time.Duration)
	Get(key string) (any, bool)
}

// Reader интерфейс для kafka.Reader
type Reader interface {
	FetchMessage(ctx context.Context) (Message, error)
	CommitMessages(ctx context.Context, msgs ...Message) error
	Close() error
}

// Consumer основной потребитель сообщений
type Consumer struct {
	reader Reader
	store  OrderStore
	cache  Cache
}

// New создаёт нового Consumer
func New(reader Reader, store OrderStore, cache Cache) *Consumer {
	return &Consumer{reader: reader, store: store, cache: cache}
}

// Run запускает цикл обработки сообщений
func (c *Consumer) Run(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Println("consumer context canceled")
				return
			}
			log.Printf("fetch message error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var ord models.Order
		if err := json.Unmarshal(m.Value, &ord); err != nil {
			log.Printf("invalid JSON: %v", err)
			_ = c.store.SaveBadMessage(ctx, m.Value, "json_unmarshal: "+err.Error())
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := ord.Validate(); err != nil {

			log.Printf(
				"invalid order uid=%s err=%v",
				ord.OrderUID,
				err,
			)
			_ = c.store.SaveBadMessage(ctx, m.Value, "validation: "+err.Error())
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := c.store.SaveOrder(ctx, ord, m.Value); err != nil {
			log.Printf("DB save error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		c.cache.Set(ord.OrderUID, ord, time.Minute)
		_ = c.reader.CommitMessages(ctx, m)
	}
}

// Close закрывает reader
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// KafkaReaderWrapper оборачивает kafka.Reader под интерфейс Consumer
type KafkaReaderWrapper struct {
	R *kafka.Reader
}

// FetchMessage получает сообщение из Kafka
func (k *KafkaReaderWrapper) FetchMessage(ctx context.Context) (Message, error) {
	m, err := k.R.FetchMessage(ctx)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Value:     m.Value,
		Partition: m.Partition,
		Offset:    m.Offset,
	}, nil

	// return Message{Value: m.Value}, nil

}

// CommitMessages подтверждает обработанные сообщения
func (k *KafkaReaderWrapper) CommitMessages(ctx context.Context, msgs ...Message) error {
	km := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		km[i] = kafka.Message{
			Value:     m.Value,
			Partition: m.Partition,
			Offset:    m.Offset,
		}

		// km[i] = kafka.Message{Value: m.Value}
	}
	return k.R.CommitMessages(ctx, km...)
}

// Close закрывает reader
func (k *KafkaReaderWrapper) Close() error {
	return k.R.Close()
}
