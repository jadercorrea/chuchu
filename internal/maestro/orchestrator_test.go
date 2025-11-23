package maestro

import "testing"

func TestParsePlan_Basic(t *testing.T) {
	m := NewMaestro(nil, ".", "")
	steps := m.parsePlan("# Title\n\n## Step A\ncontent A\n\n## Step B\ncontent B\n")
	if len(steps) != 2 {
		t.Fatalf("want 2 steps, got %d", len(steps))
	}
	if steps[0].Title != "Step A" || steps[1].Title != "Step B" {
		t.Fatalf("titles mismatch: %#v", steps)
	}
}
