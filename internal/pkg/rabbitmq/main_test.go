package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakePublishChannel struct {
	queueName string
	durable   bool
	key       string
	message   amqp.Publishing
}

func (c *fakePublishChannel) QueueDeclare(name string, durable bool, _ bool, _ bool, _ bool, _ amqp.Table) (amqp.Queue, error) {
	c.queueName = name
	c.durable = durable
	return amqp.Queue{Name: name}, nil
}

func (c *fakePublishChannel) PublishWithContext(_ context.Context, _ string, key string, _ bool, _ bool, message amqp.Publishing) error {
	c.key = key
	c.message = message
	return nil
}

func TestValidateURLAcceptsAMQPURL(t *testing.T) {
	if err := ValidateURL("amqp://guest:guest@127.0.0.1:5672/"); err != nil {
		t.Fatalf("expected valid rabbitmq url: %v", err)
	}
}

func TestValidateURLRejectsInvalidURL(t *testing.T) {
	if err := ValidateURL("http://127.0.0.1:5672/"); err == nil {
		t.Fatal("expected invalid rabbitmq url error")
	}
}

func TestDialWithRetryRetriesUntilDialerSucceeds(t *testing.T) {
	attempts := 0
	client, err := dialWithRetry(
		context.Background(),
		"amqp://guest:guest@127.0.0.1:5672/",
		3,
		time.Nanosecond,
		func(string) (*Client, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("connection refused")
			}
			return &Client{}, nil
		},
	)
	if err != nil {
		t.Fatalf("dial with retry: %v", err)
	}
	if client == nil || attempts != 3 {
		t.Fatalf("expected success on third attempt, client=%#v attempts=%d", client, attempts)
	}
}

func TestDialWithRetryReturnsLastErrorAfterAttempts(t *testing.T) {
	_, err := dialWithRetry(
		context.Background(),
		"amqp://guest:guest@127.0.0.1:5672/",
		2,
		time.Nanosecond,
		func(string) (*Client, error) {
			return nil, errors.New("connection refused")
		},
	)
	if err == nil {
		t.Fatal("expected retry error")
	}
}

func TestPublishJSONDeclaresDurableQueueAndPublishesPersistentMessage(t *testing.T) {
	channel := &fakePublishChannel{}
	publisher := NewJSONPublisher(channel)
	payload := map[string]any{"feedback_id": float64(1)}

	if err := publisher.PublishJSON(context.Background(), "feedback.created", payload); err != nil {
		t.Fatalf("publish json: %v", err)
	}

	if channel.queueName != "feedback.created" || !channel.durable {
		t.Fatalf("expected durable feedback.created queue, got name=%q durable=%v", channel.queueName, channel.durable)
	}
	if channel.key != "feedback.created" {
		t.Fatalf("expected routing key feedback.created, got %q", channel.key)
	}
	if channel.message.ContentType != "application/json" {
		t.Fatalf("expected json content type, got %q", channel.message.ContentType)
	}
	if channel.message.DeliveryMode != amqp.Persistent {
		t.Fatalf("expected persistent delivery mode, got %d", channel.message.DeliveryMode)
	}

	var got map[string]any
	if err := json.Unmarshal(channel.message.Body, &got); err != nil {
		t.Fatalf("message body is not json: %v", err)
	}
	if got["feedback_id"] != float64(1) {
		t.Fatalf("unexpected message body: %#v", got)
	}
}
