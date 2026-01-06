package main

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed skills/*.md
var embeddedSkills embed.FS

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage coding skills/guidelines",
	Long:  `List, install, and view language-specific coding skills that guide the AI agent.`,
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and installed skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".gptcode", "skills")

		fmt.Println("📚 Available Skills")
		fmt.Println()

		// List embedded skills
		embedded := listEmbeddedSkills()
		if len(embedded) > 0 {
			fmt.Println("Built-in:")
			for _, name := range embedded {
				installed := isSkillInstalled(skillsDir, name)
				status := "  "
				if installed {
					status = "✓ "
				}
				fmt.Printf("  %s%s\n", status, name)
			}
		}

		// List user-installed skills
		userSkills := listUserSkills(skillsDir)
		if len(userSkills) > 0 {
			fmt.Println()
			fmt.Println("User-installed:")
			for _, name := range userSkills {
				if !containsString(embedded, name) {
					fmt.Printf("  ✓ %s\n", name)
				}
			}
		}

		fmt.Println()
		fmt.Println("Use 'gptcode skills install <name>' to install a skill")
		fmt.Println("Use 'gptcode skills show <name>' to view a skill")
		return nil
	},
}

var skillsInstallCmd = &cobra.Command{
	Use:   "install <name|url>",
	Short: "Install a skill from built-in library or URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".gptcode", "skills")

		// Ensure skills directory exists
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory: %w", err)
		}

		var content []byte
		var err error
		var skillName string

		if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
			// Install from URL
			content, skillName, err = fetchSkillFromURL(name)
			if err != nil {
				return fmt.Errorf("failed to fetch skill: %w", err)
			}
		} else {
			// Install from embedded
			content, err = getEmbeddedSkill(name)
			if err != nil {
				return fmt.Errorf("skill '%s' not found in built-in library", name)
			}
			skillName = name
		}

		// Write skill file
		destPath := filepath.Join(skillsDir, skillName+".md")
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write skill: %w", err)
		}

		fmt.Printf("✅ Installed skill '%s' to %s\n", skillName, destPath)
		return nil
	},
}

var skillsInstallAllCmd = &cobra.Command{
	Use:   "install-all",
	Short: "Install all built-in skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".gptcode", "skills")

		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory: %w", err)
		}

		embedded := listEmbeddedSkills()
		for _, name := range embedded {
			content, err := getEmbeddedSkill(name)
			if err != nil {
				fmt.Printf("⚠️  Skipping %s: %v\n", name, err)
				continue
			}

			destPath := filepath.Join(skillsDir, name+".md")
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				fmt.Printf("⚠️  Failed to install %s: %v\n", name, err)
				continue
			}

			fmt.Printf("✅ Installed %s\n", name)
		}

		fmt.Printf("\nInstalled %d skills to %s\n", len(embedded), skillsDir)
		return nil
	},
}

var skillsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Display the contents of a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".gptcode", "skills")

		// Try installed first
		installedPath := filepath.Join(skillsDir, name+".md")
		if content, err := os.ReadFile(installedPath); err == nil {
			fmt.Println(string(content))
			return nil
		}

		// Try embedded
		if content, err := getEmbeddedSkill(name); err == nil {
			fmt.Println("(Built-in, not installed)")
			fmt.Println()
			fmt.Println(string(content))
			return nil
		}

		return fmt.Errorf("skill '%s' not found", name)
	},
}

var skillsRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"uninstall", "rm"},
	Short:   "Remove an installed skill",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		home, _ := os.UserHomeDir()
		skillsDir := filepath.Join(home, ".gptcode", "skills")

		skillPath := filepath.Join(skillsDir, name+".md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			return fmt.Errorf("skill '%s' is not installed", name)
		}

		if err := os.Remove(skillPath); err != nil {
			return fmt.Errorf("failed to remove skill: %w", err)
		}

		fmt.Printf("✅ Removed skill '%s'\n", name)
		return nil
	},
}

// Helper functions

func listEmbeddedSkills() []string {
	entries, err := embeddedSkills.ReadDir("skills")
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			if name != "README" {
				names = append(names, name)
			}
		}
	}
	return names
}

func getEmbeddedSkill(name string) ([]byte, error) {
	// Try exact name
	content, err := embeddedSkills.ReadFile("skills/" + name + ".md")
	if err == nil {
		return content, nil
	}

	// Try with language suffix patterns
	patterns := []string{
		name + "-patterns.md",
		name + "-idioms.md",
	}
	for _, pattern := range patterns {
		content, err = embeddedSkills.ReadFile("skills/" + pattern)
		if err == nil {
			return content, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func isSkillInstalled(skillsDir, name string) bool {
	patterns := []string{
		filepath.Join(skillsDir, name+".md"),
		filepath.Join(skillsDir, name+"-patterns.md"),
		filepath.Join(skillsDir, name+"-idioms.md"),
	}
	for _, path := range patterns {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func listUserSkills(skillsDir string) []string {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			names = append(names, name)
		}
	}
	return names
}

func fetchSkillFromURL(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract name from URL
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".md")

	return content, name, nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsInstallCmd)
	skillsCmd.AddCommand(skillsInstallAllCmd)
	skillsCmd.AddCommand(skillsShowCmd)
	skillsCmd.AddCommand(skillsRemoveCmd)
}
