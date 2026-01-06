package prompt

import (
	"os"
	"path/filepath"
	"strings"
)

// SkillsLoader loads language-specific skills from the skills directory
type SkillsLoader struct {
	skillsDir string
}

// NewSkillsLoader creates a loader that looks for skills in the CLI skills directory
func NewSkillsLoader() *SkillsLoader {
	// Skills are in the CLI installation directory or embedded
	// First try relative to executable, then fallback locations
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	
	possiblePaths := []string{
		filepath.Join(execDir, "skills"),
		filepath.Join(execDir, "..", "skills"),
		"skills",
		"./skills",
	}
	
	// Also check home directory for user-defined skills
	home, _ := os.UserHomeDir()
	if home != "" {
		possiblePaths = append(possiblePaths, filepath.Join(home, ".gptcode", "skills"))
	}
	
	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return &SkillsLoader{skillsDir: path}
		}
	}
	
	return &SkillsLoader{skillsDir: ""}
}

// LoadForLanguage returns the skill content for a given language
// Returns empty string if no skill exists for the language
func (sl *SkillsLoader) LoadForLanguage(lang string) string {
	if sl.skillsDir == "" {
		return ""
	}
	
	langLower := strings.ToLower(lang)
	
	// Map language names to skill files
	skillFiles := map[string]string{
		"go":         "go.md",
		"golang":     "go.md",
		"elixir":     "elixir.md",
		"ex":         "elixir.md",
		"ruby":       "ruby.md",
		"rb":         "ruby.md",
		"rails":      "rails.md",
		"python":     "python.md",     // Future
		"py":         "python.md",
		"typescript": "typescript.md", // Future
		"ts":         "typescript.md",
		"javascript": "javascript.md", // Future
		"js":         "javascript.md",
	}
	
	fileName, ok := skillFiles[langLower]
	if !ok {
		return ""
	}
	
	skillPath := filepath.Join(sl.skillsDir, fileName)
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return ""
	}
	
	return string(content)
}

// LoadByName returns a specific skill by name (e.g., "tdd-bug-fix", "code-review")
func (sl *SkillsLoader) LoadByName(name string) string {
	if sl.skillsDir == "" {
		return ""
	}
	
	// Try with and without .md extension
	paths := []string{
		filepath.Join(sl.skillsDir, name),
		filepath.Join(sl.skillsDir, name+".md"),
	}
	
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}
	
	return ""
}

// ListAvailable returns the names of all available skills
func (sl *SkillsLoader) ListAvailable() []string {
	if sl.skillsDir == "" {
		return nil
	}
	
	entries, err := os.ReadDir(sl.skillsDir)
	if err != nil {
		return nil
	}
	
	var skills []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension for display
			name := strings.TrimSuffix(entry.Name(), ".md")
			skills = append(skills, name)
		}
	}
	
	return skills
}
