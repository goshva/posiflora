package model
import "time"

type OrderPayload struct {
	Number       string  `json:"number"`
	Total        float64 `json:"total"`
	CustomerName string  `json:"customer_name"`
}

type Order struct {
	ID           int64     `json:"id"`
	ShopID       int64     `json:"shopId"`
	Number       string    `json:"number"`
	Total        float64   `json:"total"`
	CustomerName string    `json:"customerName"`
	CreatedAt    time.Time `json:"createdAt"` // <- должно быть time.Time
}
