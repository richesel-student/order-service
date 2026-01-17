package models

import (
	"testing"
	"time"
)

func TestOrderValidation_OK(t *testing.T) {
	ord := Order{
		OrderUID:    "123",
		TrackNumber: "WB123",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test User",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "City",
			Address: "Street 1",
			Region:  "Region",
			Email:   "test@example.com",
		},
		Payment: Payment{
			Transaction:  "tx123",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       100,
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      1,
				TrackNumber: "WB123",
				RID:         "rid123",
				Name:        "Product",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        123,
				Brand:       "Brand",
				Status:      200,
			},
		},
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "cust123",
		DeliveryService:   "meest",
		ShardKey:          "1",
		SmID:              1,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
	}

	err := validate.Struct(ord)
	if err != nil {
		t.Fatalf("expected valid order, got error: %v", err)
	}
}

func TestOrderValidation_InvalidCurrency(t *testing.T) {
	ord := Order{
		OrderUID:    "123",
		TrackNumber: "WB123",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test User",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "City",
			Address: "Street 1",
			Region:  "Region",
			Email:   "test@example.com",
		},
		Payment: Payment{
			Transaction:  "tx123",
			RequestID:    "",
			Currency:     "ABC", // <- НЕВАЛИДНАЯ ВАЛЮТА
			Provider:     "wbpay",
			Amount:       100,
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      1,
				TrackNumber: "WB123",
				RID:         "rid123",
				Name:        "Product",
				Sale:        0,
				Size:        "M",
				TotalPrice:  100,
				NmID:        123,
				Brand:       "Brand",
				Status:      200,
			},
		},
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "cust123",
		DeliveryService:   "meest",
		ShardKey:          "1",
		SmID:              1,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
	}

	err := validate.Struct(ord)
	if err == nil {
		t.Fatal("expected validation error due to invalid currency, got nil")
	}
}

func TestOrderValidation_MissingField(t *testing.T) {
	ord := Order{
		OrderUID: "123",
		// TrackNumber отсутствует -> validation должен упасть
		Payment: Payment{
			Currency:  "USD",
			Amount:    100,
			PaymentDT: time.Now().Unix(),
		},
		Items: []Item{
			{
				ChrtID:     1,
				Name:       "Product",
				TotalPrice: 100,
				NmID:       123,
			},
		},
	}

	err := validate.Struct(ord)
	if err == nil {
		t.Fatal("expected validation error due to missing required fields, got nil")
	}
}
