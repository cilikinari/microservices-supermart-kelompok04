package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderPlacedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		OrderID     string `json:"reference_order_id"`
		ProductID   string `json:"product_id"`
		WarehouseID string `json:"warehouse_id"`
		Quantity    int    `json:"reserved_quantity"`
	} `json:"data"`
}

type InventoryConsumer struct {
	db *sqlx.DB
}

func NewInventoryConsumer(db *sqlx.DB) *InventoryConsumer {
	return &InventoryConsumer{db: db}
}

func (c *InventoryConsumer) HandleOrderPlaced(ctx context.Context, payload []byte) error {
	var event OrderPlacedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	if event.EventID == "" {
		return errors.New("event_id is required")
	}
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT IGNORE INTO inventory_processed_events
			(event_id, event_type, consumer_name, aggregate_id)
		VALUES (?, ?, ?, ?)
	`, event.EventID, event.EventType, "inventory-order-placed", event.Data.OrderID)
	if err != nil {
		return err
	}
	claimed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if claimed == 0 {
		return nil
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE inventory_stocks
		SET quantity_reserved = quantity_reserved + ?, updated_at = NOW()
		WHERE product_id = ? AND warehouse_id = ?
		  AND (quantity_on_hand - quantity_reserved) >= ?
	`, event.Data.Quantity, event.Data.ProductID, event.Data.WarehouseID,
		event.Data.Quantity)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("insufficient stock")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_reservations
			(id, product_id, reference_order_id, reserved_quantity, status)
		VALUES (?, ?, ?, ?, 'HOLD')
	`, uuid.NewString(), event.Data.ProductID, event.Data.OrderID,
		event.Data.Quantity)
	if err != nil {
		return err
	}
	return tx.Commit()
}
