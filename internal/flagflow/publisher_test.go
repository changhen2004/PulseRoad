package flagflow

import (
	"context"
	"testing"
)

type fakeFlagMessagePublisher struct {
	queue   string
	payload any
}

func (p *fakeFlagMessagePublisher) PublishJSON(_ context.Context, queue string, payload any) error {
	p.queue = queue
	p.payload = payload
	return nil
}

func TestRabbitMQPublisherPublishesFlagChangedQueue(t *testing.T) {
	messagePublisher := &fakeFlagMessagePublisher{}
	publisher := NewRabbitMQPublisher(messagePublisher)
	event := FlagChangedEvent{
		FlagID:            1,
		ProductID:         10,
		TeamID:            20,
		Key:               "new_dashboard",
		Environment:       "production",
		Enabled:           true,
		RolloutPercentage: 50,
		Action:            EventActionToggled,
		ChangedBy:         7,
	}

	if err := publisher.PublishFlagChanged(context.Background(), event); err != nil {
		t.Fatalf("publish flag changed: %v", err)
	}

	if messagePublisher.queue != QueueFlagChanged {
		t.Fatalf("expected queue %q, got %q", QueueFlagChanged, messagePublisher.queue)
	}
	if messagePublisher.payload != event {
		t.Fatalf("unexpected payload: %#v", messagePublisher.payload)
	}
}
