package output

import (
	"github.com/charmbracelet/glamour"
)

func RenderMarkdown(md string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return md, err
	}

	rendered, err := r.Render(md)
	if err != nil {
		return md, err
	}

	return rendered, nil
}
