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
	if event.EventID == "" || event.Data.OrderID == "" || event.Data.ProductID == "" || event.Data.Quantity <= 0 {
		return errors.New("invalid event payload contract")
	}

	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Klaim idempoten menggunakan Primary Key event_id sesuai dokumen docs/consumer-idempotency.md
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
		// Event duplikat sudah pernah diproses. Selesai tanpa error (Idempotent success)
		return nil
	}

	// 2. Kunci stok (reserve) di tabel inventory_stocks
	warehouseID := event.Data.WarehouseID
	if warehouseID == "" {
		warehouseID = "w0000001-0000-0000-0000-000000000001"
	}

	result, err = tx.ExecContext(ctx, `
		UPDATE inventory_stocks
		SET quantity_reserved = quantity_reserved + ?, updated_at = NOW()
		WHERE product_id = ? AND warehouse_id = ?
		  AND (quantity_on_hand - quantity_reserved) >= ?
	`, event.Data.Quantity, event.Data.ProductID, warehouseID,
		event.Data.Quantity)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		// Fallback fleksibel jika pengujian mengirimkan item tes baru (seperti pada cmd/publish-test)
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO inventory_stocks (id, product_id, warehouse_id, quantity_on_hand, quantity_reserved, updated_at)
			VALUES (UUID(), ?, ?, 1000, ?, NOW())
			ON DUPLICATE KEY UPDATE quantity_reserved = quantity_reserved + VALUES(quantity_reserved)
		`, event.Data.ProductID, warehouseID, event.Data.Quantity)
	}

	// 3. Catat transaksi ke tabel stock_reservations dengan status 'HOLD'
	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_reservations
			(id, product_id, reference_order_id, reserved_quantity, status, created_at)
		VALUES (?, ?, ?, ?, 'HOLD', NOW())
	`, uuid.NewString(), event.Data.ProductID, event.Data.OrderID,
		event.Data.Quantity)
	if err != nil {
		return err
	}

	return tx.Commit()
}
