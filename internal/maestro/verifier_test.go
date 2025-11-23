package maestro

import (
	"context"
	"testing"
)

func TestBuildVerifier_NoProject(t *testing.T) {
	v := NewBuildVerifier(t.TempDir())
	res, err := v.Verify(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success for empty dir")
	}
}
