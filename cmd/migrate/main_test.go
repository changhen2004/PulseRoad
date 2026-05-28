package main

import (
	"testing"

	"pulseroad/internal/pkg/database"
)

func TestMigrateCommandRegistersApplicationModels(t *testing.T) {
	if got := database.RegisteredModelCount(); got == 0 {
		t.Fatal("expected migrate command to register application models")
	}
}
