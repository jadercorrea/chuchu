package catalog

import _ "embed"

// defaultModelsJSON contains a curated list of models that ship with the CLI.
// This is used as a fallback when the user hasn't run 'gptcode model update --all'.
// The file is automatically updated by CI on a weekly basis.
//
//go:embed default_models.json
var defaultModelsJSON []byte

// GetDefaultModels returns the embedded default models catalog.
// This is useful for first-time users or when the user catalog is corrupted.
func GetDefaultModels() []byte {
	return defaultModelsJSON
}
