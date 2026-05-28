package rabbitmq

import "testing"

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
