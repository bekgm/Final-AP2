package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type JobAcceptedEvent struct {
	JobID        string `json:"job_id"`
	FreelancerID string `json:"freelancer_id"`
	ClientID     string `json:"client_id"`
}

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	exchange string
}

func NewPublisher(url, exchange string) (*Publisher, error) {
	const (
		connectRetries  = 20
		connectInterval = 1 * time.Second
	)

	var (
		conn *amqp.Connection
		err  error
	)
	for attempt := 1; attempt <= connectRetries; attempt++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		if attempt < connectRetries {
			time.Sleep(connectInterval)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", connectRetries, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare a topic exchange so other services can subscribe to specific routing keys
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &Publisher{conn: conn, channel: ch, exchange: exchange}, nil
}

func (p *Publisher) PublishJobAccepted(jobID, freelancerID, clientID string) error {
	event := JobAcceptedEvent{
		JobID:        jobID,
		FreelancerID: freelancerID,
		ClientID:     clientID,
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		p.exchange,   // exchange
		"job.accepted", // routing key — Messaging Service subscribes to this
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() {
	if p == nil {
		return
	}
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

// NoopPublisher is a no-op implementation for testing.
type NoopPublisher struct{}

func (n *NoopPublisher) PublishJobAccepted(_, _, _ string) error { return nil }
func (n *NoopPublisher) Close()                                   {}
