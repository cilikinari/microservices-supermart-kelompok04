package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	RabbitMQURL     = "amqp://supermart:supermart123@localhost:5672/"
	ExchangeName    = "supermart.events"
	QueueName       = "order.created"
	RoutingKey      = "order.created"
	DLXExchangeName = "supermart.dlx"
	DLQQueueName    = "order.created.dlq"
	DLQRoutingKey   = "order.created.dead"
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

	// Exchange utama: durable.
	if err := ch.ExchangeDeclare(
		ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal membuat exchange utama: %w", err))
	}

	// Dead-Letter Exchange: durable.
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

	// Dead-Letter Queue: durable.
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

	// Hubungkan DLQ dengan DLX.
	if err := ch.QueueBind(
		DLQQueueName,
		DLQRoutingKey,
		DLXExchangeName,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal menghubungkan DLQ: %w", err))
	}

	// Queue utama menggunakan DLX untuk pesan yang ditolak.
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

	// Hubungkan queue utama ke exchange utama.
	if err := ch.QueueBind(
		QueueName,
		RoutingKey,
		ExchangeName,
		false,
		nil,
	); err != nil {
		return cleanup(fmt.Errorf("gagal menghubungkan queue utama: %w", err))
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
