package output

import (
	"regexp"
	"strings"
)

type CodeBlock struct {
	Language string
	Code     string
	Index    int
}

type ParsedResponse struct {
	RenderedText string
	CodeBlocks   []CodeBlock
}

func ParseMarkdown(md string) ParsedResponse {
	codeBlockRegex := regexp.MustCompile("```(bash|sh|shell|zsh)\\n([\\s\\S]*?)\\n```")
	matches := codeBlockRegex.FindAllStringSubmatch(md, -1)

	var blocks []CodeBlock
	for i, match := range matches {
		if len(match) >= 3 {
			blocks = append(blocks, CodeBlock{
				Language: match[1],
				Code:     strings.TrimSpace(match[2]),
				Index:    i + 1,
			})
		}
	}

	return ParsedResponse{
		RenderedText: md,
		CodeBlocks:   blocks,
	}
}
