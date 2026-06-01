package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("rabbitmq url is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse rabbitmq url: %w", err)
	}

	if parsed.Scheme != "amqp" && parsed.Scheme != "amqps" {
		return fmt.Errorf("rabbitmq url must use amqp or amqps scheme")
	}
	if parsed.Host == "" {
		return fmt.Errorf("rabbitmq url host is required")
	}

	return nil
}

type publishChannel interface {
	QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
}

type JSONPublisher struct {
	ch publishChannel
}

func NewJSONPublisher(ch publishChannel) *JSONPublisher {
	return &JSONPublisher{ch: ch}
}

func (p *JSONPublisher) PublishJSON(ctx context.Context, queueName string, payload any) error {
	if _, err := p.ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if err := p.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		return fmt.Errorf("publish message to %s: %w", queueName, err)
	}
	return nil
}

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	pub  *JSONPublisher
}

func Dial(rawURL string) (*Client, error) {
	if err := ValidateURL(rawURL); err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(rawURL)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	return &Client{conn: conn, ch: ch, pub: NewJSONPublisher(ch)}, nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.ch != nil {
		if err := c.ch.Close(); err != nil {
			_ = c.conn.Close()
			return err
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) PublishJSON(ctx context.Context, queueName string, payload any) error {
	return c.pub.PublishJSON(ctx, queueName, payload)
}

func (c *Client) ConsumeJSON(ctx context.Context, queueName string, handler func(context.Context, []byte) error) error {
	if _, err := c.ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	deliveries, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", queueName, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("consumer for %s closed", queueName)
			}
			if err := handler(ctx, delivery.Body); err != nil {
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}
