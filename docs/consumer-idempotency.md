# Idempotency Consumer Inventory

## Skenario

`order-service` menerbitkan `OrderPlacedEvent`. Consumer pada `inventory-service`
mengonsumsi event tersebut untuk membuat reservasi stok. Broker boleh mengirim
event yang sama lebih dari satu kali, sehingga `event_id` harus dipakai sebagai
identitas deduplikasi, bukan `order_id` saja.

Implementasi menggunakan **Processed Events Table** di database Inventory.
Tabel dibuat oleh [init-inventory.sql](../deployments/docker/init-scripts/inventory/init-inventory.sql).

## Skema tabel

```sql
CREATE TABLE inventory_processed_events (
    event_id      VARCHAR(36) NOT NULL,
    event_type    VARCHAR(100) NOT NULL,
    consumer_name VARCHAR(100) NOT NULL,
    aggregate_id  VARCHAR(36) NOT NULL,
    processed_at  DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (event_id),
    KEY idx_inventory_processed_events_aggregate (aggregate_id, event_type),
    KEY idx_inventory_processed_events_processed_at (processed_at)
) ENGINE=InnoDB;
```

Arti kolom:

- `event_id`: id unik dari event Order Service dan primary key untuk klaim atomik.
- `event_type`: misalnya `OrderPlaced`.
- `consumer_name`: nama consumer, berguna jika beberapa consumer memiliki ledger
  pada database yang sama.
- `aggregate_id`: `reference_order_id`, untuk audit dan pencarian berdasarkan order.
- `processed_at`: waktu reservasi berhasil di-commit.

Ledger sengaja tidak memiliki status `PROCESSING`. Baris ledger dan mutasi stok
harus berada dalam transaksi yang sama. Jika reservasi gagal, transaksi di-rollback
dan klaim event ikut hilang sehingga broker boleh mengirim ulang. Jika commit
berhasil, redelivery menemukan primary key yang sama dan tidak mengubah stok lagi.

## Potongan kode Consumer Go

Contoh berikut memakai `sqlx` dan gaya handler generik (Kafka/RabbitMQ tidak
memengaruhi bagian idempotensinya). `ack` hanya dipanggil setelah fungsi
mengembalikan `nil`; error dikembalikan ke broker agar pesan di-retry.

```go
package inventoryconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type OrderPlacedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Data      struct {
		OrderID  string `json:"reference_order_id"`
		ProductID string `json:"product_id"`
		WarehouseID string `json:"warehouse_id"`
		Quantity int `json:"reserved_quantity"`
	} `json:"data"`
}

type Consumer struct {
	db *sqlx.DB
}

func (c *Consumer) HandleOrderPlaced(
	ctx context.Context,
	payload []byte,
) error {
	var event OrderPlacedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode OrderPlacedEvent: %w", err)
	}
	if event.EventID == "" || event.Data.OrderID == "" ||
		event.Data.ProductID == "" || event.Data.WarehouseID == "" ||
		event.Data.Quantity <= 0 {
		return errors.New("invalid OrderPlacedEvent")
	}

	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin inventory transaction: %w", err)
	}
	defer tx.Rollback()

	// INSERT IGNORE adalah klaim atomik berdasarkan PRIMARY KEY event_id.
	result, err := tx.ExecContext(ctx, `
		INSERT IGNORE INTO inventory_processed_events
			(event_id, event_type, consumer_name, aggregate_id)
		VALUES (?, ?, ?, ?)
	`, event.EventID, event.EventType, "inventory-order-placed", event.Data.OrderID)
	if err != nil {
		return fmt.Errorf("claim event %q: %w", event.EventID, err)
	}

	claimed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read event claim result: %w", err)
	}
	if claimed == 0 {
		// Event sudah committed sebelumnya. Ack pesan tanpa mutasi bisnis kedua.
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
		return fmt.Errorf("reserve stock for event %q: %w", event.EventID, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read stock reservation result: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("insufficient stock for order %q", event.Data.OrderID)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO stock_reservations
			(id, product_id, reference_order_id, reserved_quantity, status)
		VALUES (?, ?, ?, ?, 'HOLD')
	`, uuid.NewString(), event.Data.ProductID, event.Data.OrderID,
		event.Data.Quantity)
	if err != nil {
		return fmt.Errorf("create reservation for event %q: %w", event.EventID, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit OrderPlacedEvent %q: %w", event.EventID, err)
	}
	return nil
}
```

Urutan operasi yang wajib dipertahankan:

1. Validasi payload sebelum transaksi.
2. Klaim `event_id` dengan unique constraint.
3. Mutasi `inventory_stocks` dan insert `stock_reservations`.
4. Commit ledger dan mutasi dalam transaksi yang sama.
5. Ack hanya setelah `HandleOrderPlaced` sukses.

Jika database sudah pernah diinisialisasi, `CREATE TABLE IF NOT EXISTS` pada
script bootstrap tidak otomatis dijalankan ulang oleh Docker. Jalankan DDL
yang sama sebagai migration pada database `inventory_service_db`.
