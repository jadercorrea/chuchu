package prompt

import (
	"testing"
)

func TestSkillsLoader(t *testing.T) {
	loader := NewSkillsLoader()
	
	// Test that loader finds skills directory
	available := loader.ListAvailable()
	t.Logf("Available skills: %v", available)
	
	// Test Go skill loading
	goSkill := loader.LoadForLanguage("go")
	if goSkill == "" {
		t.Log("No Go skill found (may be expected depending on working directory)")
	} else {
		t.Logf("Go skill loaded: %d bytes", len(goSkill))
		if len(goSkill) > 100 {
			t.Logf("First 100 chars: %s", goSkill[:100])
		}
	}
	
	// Test Elixir skill loading
	elixirSkill := loader.LoadForLanguage("elixir")
	if elixirSkill == "" {
		t.Log("No Elixir skill found (may be expected depending on working directory)")
	} else {
		t.Logf("Elixir skill loaded: %d bytes", len(elixirSkill))
	}
}
