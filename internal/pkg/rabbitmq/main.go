package rabbitmq

import (
	"fmt"
	"net/url"
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
