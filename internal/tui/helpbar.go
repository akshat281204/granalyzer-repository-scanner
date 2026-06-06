package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HelpBinding struct {
	Key  string
	Desc string
}

func RenderHelpBar(width int, bindings []HelpBinding) string {
	var rendered []string
	for _, b := range bindings {
		k := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render(b.Key)
		d := lipgloss.NewStyle().Foreground(ColorMuted).Render(b.Desc)
		rendered = append(rendered, k+" "+d)
	}

	content := strings.Join(rendered, "  •  ")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1).
		Width(width - 2).
		Render(content)
}
