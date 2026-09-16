package processors

import (
	"context"
	"quocantran/gojob/internal/models"
	"strings"
	"testing"
)

func TestProcessUnsupportedType(t *testing.T) {
	ctx := context.Background()
	p := NewJobProcessor()

	err := p.Process(ctx, &models.Job{Type: "invoice"})
	if err == nil {
		t.Fatal("expected error for unsupported job type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported job type") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
