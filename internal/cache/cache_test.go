package cache

import (
	"testing"
	"time"
)

/************* BASIC SET / GET *************/

func TestCache_SetGet(t *testing.T) {
	c := New(time.Minute, 0)

	c.Set("key", "value", 0)

	v, ok := c.Get("key")
	if !ok {
		t.Fatalf("expected key to exist")
	}

	if v.(string) != "value" {
		t.Fatalf("unexpected value: %v", v)
	}
}

/************* TTL EXPIRATION *************/

func TestCache_Expiration(t *testing.T) {
	c := New(50*time.Millisecond, 0)

	c.Set("key", "value", 50*time.Millisecond)

	time.Sleep(70 * time.Millisecond)

	if _, ok := c.Get("key"); ok {
		t.Fatalf("expected key to be expired")
	}
}

/************* DELETE *************/

func TestCache_Delete(t *testing.T) {
	c := New(time.Minute, 0)

	c.Set("key", "value", 0)

	if err := c.Delete("key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := c.Get("key"); ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestCache_Delete_NotFound(t *testing.T) {
	c := New(time.Minute, 0)

	if err := c.Delete("missing"); err == nil {
		t.Fatalf("expected error on delete missing key")
	}
}

/************* GC *************/

func TestCache_GC_RemovesExpired(t *testing.T) {
	c := New(20*time.Millisecond, 10*time.Millisecond)
	defer c.StopGC()

	c.Set("key", "value", 20*time.Millisecond)

	time.Sleep(50 * time.Millisecond)

	if _, ok := c.Get("key"); ok {
		t.Fatalf("expected GC to remove expired key")
	}
}

func TestCache_StopGC(t *testing.T) {
	c := New(time.Millisecond, 10*time.Millisecond)

	c.StopGC() // просто проверяем, что не паникует
}
