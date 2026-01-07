package prompt

import (
	"embed"
)

// Embed all skills files into the binary
//
//go:embed skills/*.md
var embeddedSkills embed.FS

// GetEmbeddedSkills returns the embedded skills filesystem
func GetEmbeddedSkills() embed.FS {
	return embeddedSkills
}
