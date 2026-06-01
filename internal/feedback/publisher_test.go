package feedback

import (
	"context"
	"testing"
)

type fakeMessagePublisher struct {
	topic   string
	payload any
}

func (p *fakeMessagePublisher) PublishJSON(_ context.Context, topic string, payload any) error {
	p.topic = topic
	p.payload = payload
	return nil
}

func TestRabbitMQPublisherPublishesFeedbackCreatedQueue(t *testing.T) {
	messagePublisher := &fakeMessagePublisher{}
	publisher := NewRabbitMQPublisher(messagePublisher)
	event := FeedbackCreatedEvent{
		FeedbackID: 1,
		ProductID:  10,
		TeamID:     20,
		Title:      "Missing export",
		Status:     StatusOpen,
		CreatedBy:  7,
	}

	if err := publisher.PublishFeedbackCreated(context.Background(), event); err != nil {
		t.Fatalf("publish feedback created: %v", err)
	}

	if messagePublisher.topic != QueueFeedbackCreated {
		t.Fatalf("expected topic %q, got %q", QueueFeedbackCreated, messagePublisher.topic)
	}
	if messagePublisher.payload != event {
		t.Fatalf("unexpected payload: %#v", messagePublisher.payload)
	}
}
