package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	RabbitMQURL     = "amqp://supermart:supermart123@localhost:5672/"
	ExchangeName    = "supermart.events"
	QueueName       = "inventory_stock_reservation_queue"
	LegacyQueueName = "order.created"
	RoutingKey      = "order.event.placed"
	BindingKey      = "order.event.*"
	DLXExchangeName = "supermart.dlx"
	DLQQueueName    = "inventory_stock_reservation_queue.dlq"
	DLQRoutingKey   = "order.event.dead"
)

type RabbitMQ struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func ConnectRabbitMQ() (*RabbitMQ, error) {
	conn, err := amqp.Dial(RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("gagal membuka channel RabbitMQ: %w", err)
	}

	cleanup := func(err error) (*RabbitMQ, error) {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	// Exchange utama: durable topic exchange sesuai Laporan Tugas 3 Bagian II
	if err := ch.ExchangeDeclare(
		ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		// Jika exchange sudah pernah dideklarasikan sebagai direct pada instance berjalan
		_ = ch.Close()
		newCh, chErr := conn.Channel()
		if chErr != nil {
			return cleanup(fmt.Errorf("gagal membuka ulang channel: %w", chErr))
		}
		ch = newCh
		if errDirect := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); errDirect != nil {
			return cleanup(fmt.Errorf("gagal membuat exchange utama: %w", errDirect))
		}
	}

	// Dead-Letter Exchange: durable direct exchange
	if err := ch.ExchangeDeclare(
		DLXExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal membuat DLX: %w", err))
	}

	// Dead-Letter Queue: durable
	if _, err := ch.QueueDeclare(
		DLQQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal membuat DLQ: %w", err))
	}

	// Hubungkan DLQ dengan DLX
	if err := ch.QueueBind(
		DLQQueueName,
		DLQRoutingKey,
		DLXExchangeName,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal menghubungkan DLQ: %w", err))
	}

	// Queue utama menggunakan DLX untuk pesan yang ditolak (poison pill)
	args := amqp.Table{
		"x-dead-letter-exchange":    DLXExchangeName,
		"x-dead-letter-routing-key": DLQRoutingKey,
	}

	if _, err := ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		args,
	); err != nil {
		return cleanup(fmt.Errorf("gagal membuat queue utama: %w", err))
	}

	// Hubungkan queue utama ke exchange utama menggunakan pattern wildcard topic (order.event.*)
	_ = ch.QueueBind(QueueName, BindingKey, ExchangeName, false, nil)
	_ = ch.QueueBind(QueueName, RoutingKey, ExchangeName, false, nil)
	_ = ch.QueueBind(QueueName, "order.created", ExchangeName, false, nil)

	// Pastikan juga legacy queue jika ada tools pengujian yang mengarah ke order.created
	if _, err := ch.QueueDeclare(LegacyQueueName, true, false, false, false, args); err == nil {
		_ = ch.QueueBind(LegacyQueueName, BindingKey, ExchangeName, false, nil)
		_ = ch.QueueBind(LegacyQueueName, "order.created", ExchangeName, false, nil)
		_ = ch.QueueBind(LegacyQueueName, RoutingKey, ExchangeName, false, nil)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}, nil
}

func (r *RabbitMQ) Publish(message []byte) error {
	if r == nil || r.Channel == nil {
		return fmt.Errorf("channel RabbitMQ belum tersedia")
	}

	return r.Channel.Publish(
		ExchangeName,
		RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         message,
		},
	)
}

func (r *RabbitMQ) Close() {
	if r == nil {
		return
	}

	if r.Channel != nil {
		_ = r.Channel.Close()
	}

	if r.Connection != nil {
		_ = r.Connection.Close()
	}
}
