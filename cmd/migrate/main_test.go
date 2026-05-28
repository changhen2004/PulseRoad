package main

import (
	"testing"

	"pulseroad/internal/pkg/database"
)

func TestMigrateCommandRegistersApplicationModels(t *testing.T) {
	if got := database.RegisteredModelCount(); got < 5 {
		t.Fatalf("expected migrate command to register at least 5 application models, got %d", got)
	}
}
