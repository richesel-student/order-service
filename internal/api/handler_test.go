package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yourmodule/internal/models"
)

/************* FAKE CACHE *************/

type fakeCache struct {
	data map[string]interface{}
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: map[string]interface{}{}}
}

func (f *fakeCache) Get(key string) (interface{}, bool) {
	v, ok := f.data[key]
	return v, ok
}

func (f *fakeCache) Set(key string, val interface{}, _ time.Duration) {
	f.data[key] = val
}

/************* FAKE STORE *************/

type fakeStore struct{}

func (f *fakeStore) GetOrder(ctx context.Context, id string) (models.Order, []byte, error) {
	if id == "123" {
		return models.Order{
			OrderUID:    "123",
			TrackNumber: "WB123",
		}, nil, nil
	}
	return models.Order{}, nil, errors.New("not found")
}

/************* TESTS *************/

func TestGetOrder_FromCache(t *testing.T) {
	cache := newFakeCache()
	cache.Set("123", models.Order{OrderUID: "123"}, time.Minute)

	server := NewServer(&fakeStore{}, cache)

	req := httptest.NewRequest(http.MethodGet, "/order/123", nil)
	w := httptest.NewRecorder()

	server.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetOrder_FromDB(t *testing.T) {
	server := NewServer(&fakeStore{}, newFakeCache())

	req := httptest.NewRequest(http.MethodGet, "/order/123", nil)
	w := httptest.NewRecorder()

	server.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var ord models.Order
	if err := json.NewDecoder(w.Body).Decode(&ord); err != nil {
		t.Fatal(err)
	}

	if ord.OrderUID != "123" {
		t.Fatalf("unexpected OrderUID: %s", ord.OrderUID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	server := NewServer(&fakeStore{}, newFakeCache())

	req := httptest.NewRequest(http.MethodGet, "/order/999", nil)
	w := httptest.NewRecorder()

	server.Routes().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
