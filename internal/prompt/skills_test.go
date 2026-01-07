package prompt

import (
	"strings"
	"testing"
)

func TestSkillsLoader(t *testing.T) {
	loader := NewSkillsLoader()

	// Test that embedded skills are available
	available := loader.ListAvailable()
	t.Logf("Available skills: %v", available)

	if len(available) < 10 {
		t.Errorf("Expected at least 10 embedded skills, got %d", len(available))
	}

	// Test Go skill loading
	goSkill := loader.LoadForLanguage("go")
	if goSkill == "" {
		t.Error("Go skill should be embedded and loadable")
	} else {
		t.Logf("Go skill loaded: %d bytes", len(goSkill))
		if !strings.Contains(goSkill, "Go") {
			t.Error("Go skill should contain 'Go'")
		}
	}

	// Test Elixir skill loading
	elixirSkill := loader.LoadForLanguage("elixir")
	if elixirSkill == "" {
		t.Error("Elixir skill should be embedded and loadable")
	} else {
		t.Logf("Elixir skill loaded: %d bytes", len(elixirSkill))
	}

	// Test LoadByName
	tddSkill := loader.LoadByName("tdd-bug-fix")
	if tddSkill == "" {
		t.Error("TDD Bug Fix skill should be loadable by name")
	}

	// Test GetSkillCategories
	categories := loader.GetSkillCategories()
	if len(categories["language"]) == 0 {
		t.Error("Should have language skills")
	}
	if len(categories["product"]) == 0 {
		t.Error("Should have product skills")
	}
}

func TestProductSkillsForTask(t *testing.T) {
	loader := NewSkillsLoader()

	// Test that frontend task triggers design-system skill
	skills := loader.LoadProductSkillsForTask("add a login form with button")
	if len(skills) == 0 {
		t.Error("Task with 'form' and 'button' should load design-system skill")
	}

	// Test analytics task
	analyticsSkills := loader.LoadProductSkillsForTask("add tracking and analytics")
	if len(analyticsSkills) == 0 {
		t.Error("Task with 'tracking' should load product-metrics skill")
	}

	// Test deployment task
	deploySkills := loader.LoadProductSkillsForTask("deploy with health checks")
	if len(deploySkills) == 0 {
		t.Error("Task with 'deploy' and 'health' should load production-ready skill")
	}

	// Test empty task
	emptySkills := loader.LoadProductSkillsForTask("")
	if len(emptySkills) != 0 {
		t.Error("Empty task should not load any skills")
	}
}
