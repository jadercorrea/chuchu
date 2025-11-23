package maestro

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMaestro_E2E_SimplePlan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dir := t.TempDir()

	goMod := `module testproject

go 1.21
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import "fmt"

func main() {
	fmt.Println(Add(2, 3))
}

func Add(a, b int) int {
	return 0
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	testGo := `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Fatalf("expected 5, got %d", Add(2, 3))
	}
}
`
	if err := os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(testGo), 0644); err != nil {
		t.Fatal(err)
	}

	simplePlan := `# Simple Plan

## Fix Add function

The Add function currently returns 0 but should return a + b.

Implementation:
- Update the Add function to return the sum of a and b
`

	m := NewMaestro(nil, dir, "")
	steps := m.parsePlan(simplePlan)

	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}

	if steps[0].Title != "Fix Add function" {
		t.Fatalf("unexpected title: %s", steps[0].Title)
	}

	if !fileExists(filepath.Join(dir, "go.mod")) {
		t.Fatal("go.mod missing")
	}

	result, err := NewBuildVerifier(dir).Verify(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("build should succeed initially: %s", result.Output)
	}
}

func TestMaestro_ParsePlan_WithSubSteps(t *testing.T) {
	plan := `# Title

## Phase 1

Phase content

### Step 1.1
Step 1.1 content

### Step 1.2
Step 1.2 content

## Phase 2

Phase 2 content
`

	m := NewMaestro(nil, ".", "")
	steps := m.parsePlan(plan)

	if len(steps) != 3 {
		t.Fatalf("expected 3 flattened steps, got %d: %#v", len(steps), steps)
	}

	if steps[0].Title != "Phase 1 / Step 1.1" {
		t.Fatalf("unexpected title: %s", steps[0].Title)
	}

	if steps[1].Title != "Phase 1 / Step 1.2" {
		t.Fatalf("unexpected title: %s", steps[1].Title)
	}

	if steps[2].Title != "Phase 2" {
		t.Fatalf("unexpected title: %s", steps[2].Title)
	}
}
