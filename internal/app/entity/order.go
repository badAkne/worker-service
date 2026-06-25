package entity

const (
	EventOrderHeaderTypeKey = "type"

	EventOrderHeaderCreated            = "order.created"
	EventOrderHeaderDeliveryCalculated = "order.delivery.calculated"
)

type EventOrderCreated struct {
	OrderID     string           `json:"order_id"`
	UserID      string           `json:"user_id"`
	TotalAmount float64          `json:"total_amount"`
	Currency    string           `json:"currency"`
	Items       []EventOrderItem `json:"items"`
	CreatedAt   string           `json:"created_at"`
}

type EventOrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type EventOrderDeliveryCalculated struct {
	OrderID       string  `json:"order_id"`
	DeliveryPrice float64 `json:"delivery_price"`
	Currency      string  `json:"currency"`
	CalculatedAt  string  `json:"calculated_at"`
}
