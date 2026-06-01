package main

import (
	"context"
	"testing"
)

func TestHandleFeedbackCreatedMessageAcceptsValidEvent(t *testing.T) {
	message := []byte(`{"feedback_id":1,"product_id":10,"team_id":20,"title":"Missing export","status":"open","created_by":7,"occurred_at":"2026-05-30T00:00:00Z"}`)

	if err := handleFeedbackCreatedMessage(context.Background(), message); err != nil {
		t.Fatalf("handle message: %v", err)
	}
}

func TestHandleFeedbackCreatedMessageRejectsInvalidJSON(t *testing.T) {
	if err := handleFeedbackCreatedMessage(context.Background(), []byte(`{`)); err == nil {
		t.Fatal("expected invalid json error")
	}
}

func TestHandleFlagChangedMessageAcceptsValidEvent(t *testing.T) {
	message := []byte(`{"flag_id":1,"product_id":10,"team_id":20,"key":"new_dashboard","environment":"production","enabled":true,"rollout_percentage":50,"action":"toggled","changed_by":7,"occurred_at":"2026-05-30T00:00:00Z"}`)

	if err := handleFlagChangedMessage(context.Background(), message); err != nil {
		t.Fatalf("handle flag message: %v", err)
	}
}

func TestHandleFlagChangedMessageRejectsInvalidJSON(t *testing.T) {
	if err := handleFlagChangedMessage(context.Background(), []byte(`{`)); err == nil {
		t.Fatal("expected invalid json error")
	}
}
