package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"yourmodule/internal/models"
)

/************* FAKE STORE / MOCK *************/

type fakeStore struct {
	data            map[string]models.Order
	BadMessageRaw   []byte
	BadMessageErr   string
	BadMessageSaved bool
}

func newFakeStoreWithSpy() *fakeStore {
	return &fakeStore{
		data: make(map[string]models.Order),
	}
}

func (f *fakeStore) SaveBadMessage(ctx context.Context, raw []byte, errText string) error {
	f.BadMessageRaw = raw
	f.BadMessageErr = errText
	f.BadMessageSaved = true
	return nil
}

func (f *fakeStore) GetOrder(ctx context.Context, orderUID string) (models.Order, []byte, error) {
	o, ok := f.data[orderUID]
	if !ok {
		return models.Order{}, nil, errors.New("not found")
	}
	raw, _ := json.Marshal(o)
	return o, raw, nil
}

func newFakeStore() *fakeStore {
	return &fakeStore{data: map[string]models.Order{}}
}

func (f *fakeStore) SaveOrder(ctx context.Context, ord models.Order, raw []byte) error {
	if f.data == nil {
		return errors.New("data not initialized")
	}
	f.data[ord.OrderUID] = ord
	return nil
}

func (f *fakeStore) LoadAllOrders(ctx context.Context, limit int) (map[string]models.Order, error) {
	out := make(map[string]models.Order)
	for k, v := range f.data {
		out[k] = v
	}
	return out, nil
}

/************* TESTS *************/

func TestStore_SaveGetOrder(t *testing.T) {
	store := newFakeStore()
	ctx := context.Background()

	order := models.Order{OrderUID: "123"}
	raw, _ := json.Marshal(order)
	if err := store.SaveOrder(ctx, order, raw); err != nil {
		t.Fatalf("SaveOrder failed: %v", err)
	}

	got, rawGot, err := store.GetOrder(ctx, "123")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if got.OrderUID != "123" {
		t.Fatalf("expected OrderUID 123, got %s", got.OrderUID)
	}

	if len(rawGot) == 0 {
		t.Fatalf("expected raw JSON, got empty")
	}
}

func TestStore_GetOrder_NotFound(t *testing.T) {
	store := newFakeStore()
	ctx := context.Background()

	_, _, err := store.GetOrder(ctx, "999")
	if err == nil {
		t.Fatalf("expected error for missing order")
	}
}

func TestStore_LoadAllOrders(t *testing.T) {
	store := newFakeStore()
	ctx := context.Background()

	orders := []models.Order{
		{OrderUID: "1"},
		{OrderUID: "2"},
	}

	for _, o := range orders {
		raw, _ := json.Marshal(o)
		store.SaveOrder(ctx, o, raw)
	}

	all, err := store.LoadAllOrders(ctx, 0)
	if err != nil {
		t.Fatalf("LoadAllOrders failed: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(all))
	}
}

func TestStore_SaveBadMessage(t *testing.T) {
	store := newFakeStore()
	ctx := context.Background()

	err := store.SaveBadMessage(ctx, []byte("bad"), "error")
	if err != nil {
		t.Fatalf("SaveBadMessage failed: %v", err)
	}
}
func TestStore_SaveBadMessageCalled(t *testing.T) {
	store := newFakeStoreWithSpy()
	ctx := context.Background()

	raw := []byte(`{"order_uid":""}`) // невалидный JSON
	err := store.SaveBadMessage(ctx, raw, "validation failed")
	if err != nil {
		t.Fatalf("SaveBadMessage failed: %v", err)
	}

	if !store.BadMessageSaved {
		t.Fatalf("expected bad message to be saved")
	}

	if string(store.BadMessageRaw) != string(raw) {
		t.Fatalf("raw message mismatch")
	}

	if store.BadMessageErr != "validation failed" {
		t.Fatalf("error text mismatch")
	}
}
