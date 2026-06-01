package flagflow

import "context"

const QueueFlagChanged = "flag.changed"

type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, payload any) error
}

type RabbitMQPublisher struct {
	messages MessagePublisher
}

func NewRabbitMQPublisher(messages MessagePublisher) *RabbitMQPublisher {
	return &RabbitMQPublisher{messages: messages}
}

func (p *RabbitMQPublisher) PublishFlagChanged(ctx context.Context, event FlagChangedEvent) error {
	return p.messages.PublishJSON(ctx, QueueFlagChanged, event)
}
