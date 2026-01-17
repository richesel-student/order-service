package models

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type Delivery struct {
	Name    string `json:"name" validate:"required,min=1,max=128"`
	Phone   string `json:"phone" validate:"required,min=7,max=32"`
	Zip     string `json:"zip" validate:"required,min=3,max=16"`
	City    string `json:"city" validate:"required,min=1,max=64"`
	Address string `json:"address" validate:"required,min=1,max=256"`
	Region  string `json:"region" validate:"required,min=1,max=64"`
	Email   string `json:"email" validate:"required,email,max=128"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required,min=1,max=128"`
	RequestID    string `json:"request_id" validate:"omitempty,max=128"`
	Currency     string `json:"currency" validate:"required,oneof=USD"`
	Provider     string `json:"provider" validate:"required,min=1,max=64"`
	Amount       int    `json:"amount" validate:"gt=0"`
	PaymentDT    int64  `json:"payment_dt" validate:"gt=0"`
	Bank         string `json:"bank" validate:"required,min=1,max=64"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"gt=0"`
	TrackNumber string `json:"track_number" validate:"required,min=1,max=64"`
	Price       int    `json:"price" validate:"gte=0"`
	RID         string `json:"rid" validate:"required,min=1,max=128"`
	Name        string `json:"name" validate:"required,min=1,max=128"`
	Sale        int    `json:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size" validate:"required,min=1,max=16"`
	TotalPrice  int    `json:"total_price" validate:"gte=0"`
	NmID        int    `json:"nm_id" validate:"gt=0"`
	Brand       string `json:"brand" validate:"required,min=1,max=64"`
	Status      int    `json:"status" validate:"gte=0"`
}

type Order struct {
	OrderUID          string   `json:"order_uid" validate:"required,min=1,max=64"`
	TrackNumber       string   `json:"track_number" validate:"required,min=1,max=64"`
	Entry             string   `json:"entry" validate:"required,min=1,max=32"`
	Delivery          Delivery `json:"delivery" validate:"required"`
	Payment           Payment  `json:"payment" validate:"required"`
	Items             []Item   `json:"items" validate:"required,min=1,dive"`
	Locale            string   `json:"locale" validate:"required,oneof=ru en"`
	InternalSignature string   `json:"internal_signature" validate:"omitempty,max=256"`
	CustomerID        string   `json:"customer_id" validate:"required,min=1,max=64"`
	DeliveryService   string   `json:"delivery_service" validate:"required,min=1,max=64"`
	ShardKey          string   `json:"shardkey" validate:"required,min=1,max=32"`
	SmID              int      `json:"sm_id" validate:"gte=0"`
	DateCreated       string   `json:"date_created" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	OofShard          string   `json:"oof_shard" validate:"required,min=1,max=32"`
}

func (o *Order) Validate() error {
	return validate.Struct(o)
}
