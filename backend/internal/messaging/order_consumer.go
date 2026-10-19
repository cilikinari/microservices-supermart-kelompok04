package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

func (r *RabbitMQ) ConsumeOrderCreated(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database identity tidak tersedia")
	}

	// Batasi pesan yang belum di-ACK untuk mengendalikan penggunaan memori.
	if err := r.Channel.Qos(10, 0, false); err != nil {
		return fmt.Errorf("gagal mengatur QoS: %w", err)
	}

	messages, err := r.Channel.Consume(
		QueueName,
		"order-consumer",
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

			// Pesan dengan format JSON yang rusak masuk ke DLQ.
			if err := json.Unmarshal(message.Body, &event); err != nil {
				log.Printf("JSON tidak valid, mengirim ke DLQ: %v", err)
				if nackErr := message.Nack(false, false); nackErr != nil {
					log.Printf("Gagal NACK pesan: %v", nackErr)
				}
				continue
			}

			// Tolak payload yang tidak memenuhi kontrak event.
			if event.EventID == "" ||
				event.EventType != "OrderCreated" ||
				event.Data.ReferenceOrderID == "" ||
				event.Data.ProductID == "" ||
				event.Data.ReservedQuantity <= 0 {
				log.Printf("Payload OrderCreated tidak valid, mengirim ke DLQ")
				if nackErr := message.Nack(false, false); nackErr != nil {
					log.Printf("Gagal NACK pesan: %v", nackErr)
				}
				continue
			}

			// Simpan event dalam transaksi database.
			tx, err := db.Beginx()
			if err != nil {
				log.Printf("Gagal memulai transaksi: %v", err)
				_ = message.Nack(false, true)
				continue
			}

			// INSERT IGNORE membuat pemrosesan ulang event_id yang sama
			// tidak menghasilkan catatan duplikat.
			_, err = tx.Exec(`
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

			if err != nil {
				_ = tx.Rollback()
				log.Printf("Gagal menyimpan event ke database: %v", err)
				_ = message.Nack(false, true)
				continue
			}

			if err := tx.Commit(); err != nil {
				log.Printf("Gagal commit transaksi: %v", err)
				_ = message.Nack(false, true)
				continue
			}

			// ACK hanya setelah transaksi database berhasil di-commit.
			if err := message.Ack(false); err != nil {
				log.Printf("Database sudah commit, tetapi ACK gagal: %v", err)
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
