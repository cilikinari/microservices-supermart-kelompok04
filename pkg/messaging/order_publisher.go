package messaging

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OrderCreatedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Producer  string    `json:"producer"`
	Data      OrderData `json:"data"`
}

type OrderData struct {
	ReferenceOrderID string `json:"reference_order_id"`
	ProductID        string `json:"product_id"`
	WarehouseID      string `json:"warehouse_id"`
	ReservedQuantity int    `json:"reserved_quantity"`
}

func (r *RabbitMQ) PublishOrderCreated(orderID, productID, warehouseID string, quantity int) error {
	event := OrderCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "OrderCreated",
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		Producer:  "order-service",
		Data: OrderData{
			ReferenceOrderID: orderID,
			ProductID:        productID,
			WarehouseID:      warehouseID,
			ReservedQuantity: quantity,
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.Publish(body)
}
