package maestro

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckpoint_SaveAndRestore(t *testing.T) {
	dir := t.TempDir()
	cs := NewCheckpointSystem(dir)

	file := filepath.Join(dir, "a.txt")
	os.WriteFile(file, []byte("v1"), 0644)

	ckpt, err := cs.Save(0, []string{file})
	if err != nil {
		t.Fatalf("save error: %v", err)
	}

	os.WriteFile(file, []byte("v2"), 0644)

	if err := cs.Restore(ckpt.ID); err != nil {
		t.Fatalf("restore error: %v", err)
	}

	b, _ := os.ReadFile(file)
	if string(b) != "v1" {
		t.Fatalf("expected v1, got %s", string(b))
	}
}
