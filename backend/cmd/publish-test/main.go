package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nusantara-supermart/backend/internal/messaging"
)

func main() {
	mode := flag.String("mode", "batch", "Mode pengujian: batch (50 events), duplicate (uji idempotensi), poison (uji DLQ)")
	count := flag.Int("count", 50, "Jumlah pesan untuk mode batch")
	customEventID := flag.String("event-id", "", "Event ID kustom (opsional, default: otomatis UUID)")
	flag.Parse()

	rabbit, err := messaging.ConnectRabbitMQ()
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ: %v", err)
	}
	defer rabbit.Close()

	switch *mode {
	case "duplicate":
		fmt.Println("=== PENGUJIAN IDEMPOTENSI (DUPLICATE MESSAGE) ===")
		eventID := *customEventID
		if eventID == "" {
			eventID = uuid.NewString()
		}
		orderID := "TEST-DUP-ORDER-001"
		event := messaging.OrderCreatedEvent{
			EventID:   eventID,
			EventType: "OrderCreated",
			Timestamp: time.Now().UTC(),
			Version:   "1.0",
			Producer:  "order-service",
			Data: messaging.OrderData{
				ReferenceOrderID: orderID,
				ProductID:        "TEST-PRODUCT-001",
				WarehouseID:      "TEST-WAREHOUSE-001",
				ReservedQuantity: 2,
			},
		}
		body, _ := json.Marshal(event)

		fmt.Printf("1. Mengirim event PERTAMA (EventID: %s)...\n", eventID)
		if err := rabbit.Publish(body); err != nil {
			log.Fatalf("Gagal kirim event 1: %v", err)
		}
		fmt.Println("   -> Event 1 terkirim. Menunggu 1 detik...")
		time.Sleep(1 * time.Second)

		fmt.Printf("2. Mengirim event KEDUA yang SAMA PERSIS (EventID: %s)...\n", eventID)
		if err := rabbit.Publish(body); err != nil {
			log.Fatalf("Gagal kirim event 2: %v", err)
		}
		fmt.Println("   -> Event 2 terkirim.")
		fmt.Println("Hasil yang diharapkan: Consumer mencatat '[Idempotency] ... Melewati mutasi stok ganda' dan stok hanya dipotong 1 kali.")

	case "poison":
		fmt.Println("=== PENGUJIAN POISON PILL & DEAD-LETTER QUEUE (DLQ) ===")
		// 1. Pesan JSON rusak
		corruptJSON := []byte(`{"event_id": "corrupt-123", "event_type": "OrderCreated", "data": { broken json ...`)
		fmt.Println("1. Mengirim pesan JSON corrupt...")
		if err := rabbit.Publish(corruptJSON); err != nil {
			log.Fatalf("Gagal kirim JSON corrupt: %v", err)
		}

		// 2. Pesan valid JSON tapi melanggar kontrak skema (misal reserved_quantity = 0)
		invalidContract := messaging.OrderCreatedEvent{
			EventID:   uuid.NewString(),
			EventType: "OrderCreated",
			Data: messaging.OrderData{
				ReferenceOrderID: "TEST-BAD-001",
				ProductID:        "", // ProductID kosong (cacat kontrak)
				ReservedQuantity: 0,  // Kuantitas tidak valid
			},
		}
		badBody, _ := json.Marshal(invalidContract)
		fmt.Println("2. Mengirim pesan cacat kontrak skema...")
		if err := rabbit.Publish(badBody); err != nil {
			log.Fatalf("Gagal kirim payload cacat: %v", err)
		}
		fmt.Println("Hasil yang diharapkan: Consumer menolak pesan via NACK (requeue=false) dan pesan otomatis dialihkan ke inventory_stock_reservation_queue.dlq.")

	default: // batch
		fmt.Printf("=== PENGUJIAN BATCH EVENT (%d PESAN) ===\n", *count)
		for i := 1; i <= *count; i++ {
			orderID := fmt.Sprintf("TEST-ORDER-%03d", i)
			if err := rabbit.PublishOrderCreated(
				orderID,
				"TEST-PRODUCT-001",
				"TEST-WAREHOUSE-001",
				1,
			); err != nil {
				log.Fatalf("Gagal publish pesan %d: %v", i, err)
			}
		}
		fmt.Printf("Berhasil mem-publish %d event OrderCreated ke RabbitMQ.\n", *count)
	}
}