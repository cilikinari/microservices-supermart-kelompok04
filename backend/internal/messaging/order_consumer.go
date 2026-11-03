package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type OrderEventHandler func(ctx context.Context, payload []byte) error

func (r *RabbitMQ) ConsumeOrderEvents(handler OrderEventHandler, identityAuditDB *sqlx.DB) error {
	// Batasi pesan unacknowledged sesuai QoS (prefetch=10) sesuai Laporan Tugas 3 Bagian III.2
	if err := r.Channel.Qos(10, 0, false); err != nil {
		return fmt.Errorf("gagal mengatur QoS: %w", err)
	}

	messages, err := r.Channel.Consume(
		QueueName,
		"inventory-order-consumer",
		false, // Manual ACK
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("gagal memulai consumer: %w", err)
	}

	go func() {
		for message := range messages {
			var event OrderCreatedEvent

			// Validasi 1: Format JSON (jika corrupt, NACK tanpa requeue -> otomatis dialihkan ke DLQ)
			if err := json.Unmarshal(message.Body, &event); err != nil {
				log.Printf("[DLQ Alert] JSON tidak valid, mengalihkan ke DLQ: %v", err)
				_ = message.Nack(false, false)
				continue
			}

			// Validasi 2: Kontrak Skema Event (Poison Pill Mitigation)
			if event.EventID == "" ||
				event.Data.ReferenceOrderID == "" ||
				event.Data.ProductID == "" ||
				event.Data.ReservedQuantity <= 0 {
				log.Printf("[DLQ Alert] Payload OrderCreated cacat kontrak, mengalihkan ke DLQ: %s", event.EventID)
				_ = message.Nack(false, false)
				continue
			}

			// Jalankan handler bisnis Inventory Consumer (Idempotensi & Reservasi Stok)
			if handler != nil {
				ctx := context.Background()
				if err := handler(ctx, message.Body); err != nil {
					log.Printf("Gagal memproses event di Inventory: %v", err)
					// Jika gagal karena stok tidak mencukupi atau payload tidak valid, kirim ke DLQ
					_ = message.Nack(false, false)
					continue
				}
			}

			// Audit backward compatibility jika tabel uji coba identity masih aktif
			if identityAuditDB != nil {
				_, _ = identityAuditDB.Exec(`
					INSERT IGNORE INTO rabbitmq_processed_events
						(event_id, event_type, order_id, product_id, quantity)
					VALUES (?, ?, ?, ?, ?)
				`,
					event.EventID,
					event.EventType,
					event.Data.ReferenceOrderID,
					event.Data.ProductID,
					event.Data.ReservedQuantity,
				)
			}

			// ACK hanya setelah transaksi database inventaris berhasil
			if err := message.Ack(false); err != nil {
				log.Printf("Transaksi sukses tapi ACK RabbitMQ gagal: %v", err)
				continue
			}

			log.Printf(
				"Event tersimpan dan di-ACK: event=%s order=%s product=%s quantity=%d",
				event.EventID,
				event.Data.ReferenceOrderID,
				event.Data.ProductID,
				event.Data.ReservedQuantity,
			)
		}
	}()

	return nil
}

func (r *RabbitMQ) ConsumeOrderCreated(db *sqlx.DB) error {
	return r.ConsumeOrderEvents(nil, db)
}
