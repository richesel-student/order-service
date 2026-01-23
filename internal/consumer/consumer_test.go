package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	"yourmodule/internal/models"
)

// --------- FAKE CACHE ---------
type fakeCache struct {
	data map[string]interface{}
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: map[string]interface{}{}}
}

func (f *fakeCache) Set(key string, val any, _ time.Duration) {
	f.data[key] = val
}

func (f *fakeCache) Get(key string) (any, bool) {
	v, ok := f.data[key]
	return v, ok
}

// --------- FAKE STORE ---------
type fakeStore struct{}

func (f *fakeStore) SaveOrder(ctx context.Context, ord models.Order, raw []byte) error {
	return nil
}

func (f *fakeStore) SaveBadMessage(ctx context.Context, raw []byte, errText string) error {
	return nil
}

// --------- FAKE READER ---------
type fakeReader struct {
	messages []Message
	idx      int
}

func (f *fakeReader) FetchMessage(ctx context.Context) (Message, error) {
	if f.idx >= len(f.messages) {
		return Message{}, context.Canceled
	}
	m := f.messages[f.idx]
	f.idx++
	return m, nil
}

func (f *fakeReader) CommitMessages(ctx context.Context, msgs ...Message) error {
	return nil
}

func (f *fakeReader) Close() error { return nil }

// --------- TESTS ---------
func TestConsumerRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	validOrder := models.Order{
		OrderUID:    "123",
		TrackNumber: "TN12345",
		Entry:       "entry",
		Delivery: models.Delivery{
			Name:    "John Doe",
			Phone:   "1234567", // минимум 7 символов
			Zip:     "12345",
			City:    "City",
			Address: "addr",
			Region:  "reg",
			Email:   "email@example.com",
		},
		Payment: models.Payment{
			Transaction: "tr",
			Currency:    "USD",
			Provider:    "prov",
			Amount:      100,
			PaymentDT:   1234567890,
			Bank:        "bank",
		},
		Items: []models.Item{
			{
				ChrtID:      1,
				TrackNumber: "TN12345",
				Price:       100,
				RID:         "RID123",
				Name:        "item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        1,
				Brand:       "BrandX",
				Status:      0,
			},
		},
		Locale:          "en",
		CustomerID:      "cust123",
		DeliveryService: "dService",
		ShardKey:        "shard",
		DateCreated:     "2026-01-16T15:58:00Z",
		OofShard:        "oof",
	}

	validJSON, err := json.Marshal(validOrder)
	if err != nil {
		t.Fatalf("failed to marshal order: %v", err)
	}

	reader := &fakeReader{
		messages: []Message{
			{Value: validJSON},
		},
	}

	store := &fakeStore{}
	cache := newFakeCache()

	cons := New(reader, store, cache)

	go cons.Run(ctx)

	time.Sleep(100 * time.Millisecond)
	cancel()

	v, ok := cache.Get("123")
	if !ok {
		t.Fatalf("order not cached")
	}
	orderCached, ok := v.(models.Order)
	if !ok {
		t.Fatalf("cached value has wrong type")
	}
	if orderCached.OrderUID != "123" {
		t.Fatalf("cached order UID mismatch, got %s", orderCached.OrderUID)
	}
}
