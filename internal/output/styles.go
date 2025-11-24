package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
	"os"
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginTop(1).
			MarginBottom(1)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func Separator() string {
	width := GetTerminalWidth()
	return SeparatorStyle.Render(strings.Repeat("━", width))
}

func Header(text string) string {
	return HeaderStyle.Render(text)
}

func Success(text string) string {
	return SuccessStyle.Render(fmt.Sprintf("✓ %s", text))
}

func CodeBlockBox(title, code string) string {
	styledCode := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")).
		Render(code)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("11"))

	content := fmt.Sprintf("%s\n\n%s", titleStyle.Render(title), styledCode)
	return BoxStyle.Render(content)
}
