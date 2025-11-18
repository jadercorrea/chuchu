package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetCatalogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".chuchu", "models.json")
}

func Load() (*OutputJSON, error) {
	path := GetCatalogPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog: %w", err)
	}

	var catalog OutputJSON
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	return &catalog, nil
}

func GetModelsForBackend(backend string) ([]ModelOutput, error) {
	catalog, err := Load()
	if err != nil {
		return nil, err
	}

	var models []ModelOutput
	switch strings.ToLower(backend) {
	case "groq":
		models = catalog.Groq.Models
	case "openrouter":
		models = catalog.OpenRouter.Models
	case "ollama":
		models = catalog.Ollama.Models
	case "openai":
		models = catalog.OpenAI.Models
	case "deepseek":
		models = catalog.DeepSeek.Models
	default:
		return nil, fmt.Errorf("unknown backend: %s", backend)
	}

	return models, nil
}

func SearchModels(backend string, query string) ([]ModelOutput, error) {
	models, err := GetModelsForBackend(backend)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return models, nil
	}

	query = strings.ToLower(query)
	filtered := []ModelOutput{}

	for _, model := range models {
		if matchesQuery(model, query) {
			filtered = append(filtered, model)
		}
	}

	return filtered, nil
}

func matchesQuery(model ModelOutput, query string) bool {
	if strings.Contains(strings.ToLower(model.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(model.ID), query) {
		return true
	}
	for _, tag := range model.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func FilterByTag(models []ModelOutput, tag string) []ModelOutput {
	filtered := []ModelOutput{}
	tag = strings.ToLower(tag)

	for _, model := range models {
		for _, t := range model.Tags {
			if strings.ToLower(t) == tag {
				filtered = append(filtered, model)
				break
			}
		}
	}

	return filtered
}

func GetRecommendedForAgent(backend string, agent string) ([]ModelOutput, error) {
	models, err := GetModelsForBackend(backend)
	if err != nil {
		return nil, err
	}

	filtered := []ModelOutput{}
	for _, model := range models {
		for _, rec := range model.RecommendedFor {
			if rec == agent {
				filtered = append(filtered, model)
				break
			}
		}
	}

	return filtered, nil
}
