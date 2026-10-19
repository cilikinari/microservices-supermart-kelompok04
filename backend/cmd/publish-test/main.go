package main

import (
	"fmt"
	"log"

	"github.com/nusantara-supermart/backend/internal/messaging"
)

func main() {
	rabbit, err := messaging.ConnectRabbitMQ()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	for i := 1; i <= 50; i++ {
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

	fmt.Println("Berhasil publish 50 event OrderCreated")
}