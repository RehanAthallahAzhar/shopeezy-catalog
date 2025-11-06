package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/models"
	"github.com/streadway/amqp"
)

const (
	OrderExchange = "product_events"
)

type RabbitMQPublisher struct {
	channel *amqp.Channel
}

// publisher constructor
func NewRabbitMQPublisher(ch *amqp.Channel) (*RabbitMQPublisher, error) {
	// Pastikan "Papan Pengumuman" (Exchange) ada.
	// Jika belum ada, perintah ini akan membuatnya.
	err := ch.ExchangeDeclare(
		OrderExchange, // name
		"fanout",      // type: fanout akan mengirim pesan ke semua queue yang terikat
		true,          // durable: exchange akan tetap ada jika RabbitMQ restart
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mendeklarasikan exchange: %w", err)
	}

	return &RabbitMQPublisher{channel: ch}, nil
}

// mengubah event menjadi JSON dan mengirimkannya ke exchange.
func (p *RabbitMQPublisher) PublishOrderCreated(ctx context.Context, event models.OrderCreatedEvent) error {
	// Ubah struct Go menjadi JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publikasikan pesan
	err = p.channel.Publish(
		OrderExchange, // exchange: Kirim ke "papan pengumuman" kita
		"",            // routing key: tidak perlu untuk fanout
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("OrderCreated event was successfully published for Order ID: %s", event.OrderID)
	return nil
}
