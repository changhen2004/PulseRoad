package feedback

import "context"

const QueueFeedbackCreated = "feedback.created"

type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, payload any) error
}

type RabbitMQPublisher struct {
	messages MessagePublisher
}

func NewRabbitMQPublisher(messages MessagePublisher) *RabbitMQPublisher {
	return &RabbitMQPublisher{messages: messages}
}

func (p *RabbitMQPublisher) PublishFeedbackCreated(ctx context.Context, event FeedbackCreatedEvent) error {
	return p.messages.PublishJSON(ctx, QueueFeedbackCreated, event)
}
