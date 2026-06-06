package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	ColorBg      = lipgloss.AdaptiveColor{Light: "#f5f5f7", Dark: "#0f0f11"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#d2d2d7", Dark: "#2a2a2f"}
	ColorAccent  = lipgloss.AdaptiveColor{Light: "#5856d6", Dark: "#7c6af7"} // Premium purple
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#86868b", Dark: "#55555e"}
	ColorText    = lipgloss.AdaptiveColor{Light: "#1d1d1f", Dark: "#e4e4e9"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#34c759", Dark: "#5faf5f"}
	ColorWarn    = lipgloss.AdaptiveColor{Light: "#ff9500", Dark: "#d4ac3a"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#ff3b30", Dark: "#d45f5f"}

	// Text Styles
	TextStyle       = lipgloss.NewStyle().Foreground(ColorText)
	MutedTextStyle  = lipgloss.NewStyle().Foreground(ColorMuted)
	AccentTextStyle = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	TitleStyle      = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Padding(0, 1)

	// Borders & Panels
	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	ActiveBorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent)

	// Tabs
	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorMuted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorAccent).
			Bold(true)

	// Custom table/list headers
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Underline(true)
)

func RenderWithTitle(style lipgloss.Style, title string, content string, width, height int) string {
	rendered := style.Width(width).Height(height).Render(content)
	lines := strings.Split(rendered, "\n")
	if len(lines) > 0 {
		lines[0] = insertTitleInTopBorder(lines[0], title)
	}
	return strings.Join(lines, "\n")
}

func insertTitleInTopBorder(topLine string, title string) string {
	if title == "" {
		return topLine
	}

	type part struct {
		isANSI bool
		text   string
	}
	var parts []part

	inANSI := false
	var current strings.Builder
	for i := 0; i < len(topLine); i++ {
		b := topLine[i]
		if b == '\x1b' {
			if current.Len() > 0 {
				parts = append(parts, part{isANSI: inANSI, text: current.String()})
				current.Reset()
			}
			inANSI = true
			current.WriteByte(b)
		} else if inANSI && b == 'm' {
			current.WriteByte(b)
			parts = append(parts, part{isANSI: true, text: current.String()})
			current.Reset()
			inANSI = false
		} else {
			current.WriteByte(b)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, part{isANSI: inANSI, text: current.String()})
	}

	type visualRune struct {
		r         rune
		partIndex int
		byteIndex int
	}
	var visualRunes []visualRune
	for pIdx, p := range parts {
		if p.isANSI {
			continue
		}
		runes := []rune(p.text)
		byteOffset := 0
		for _, r := range runes {
			visualRunes = append(visualRunes, visualRune{
				r:         r,
				partIndex: pIdx,
				byteIndex: byteOffset,
			})
			byteOffset += len(string(r))
		}
	}

	titleRunes := []rune(title)
	if len(visualRunes) < len(titleRunes)+4 {
		return topLine
	}

	partRunes := make([][]rune, len(parts))
	for i, p := range parts {
		partRunes[i] = []rune(p.text)
	}

	for i, tr := range titleRunes {
		vRune := visualRunes[2+i]
		rIdx := 0
		bOffset := 0
		for bOffset < vRune.byteIndex {
			bOffset += len(string(partRunes[vRune.partIndex][rIdx]))
			rIdx++
		}
		partRunes[vRune.partIndex][rIdx] = tr
	}

	var result strings.Builder
	for _, pr := range partRunes {
		result.WriteString(string(pr))
	}
	return result.String()
}

func RepeatDivider(width int) string {
	w := width - 4
	if w < 0 {
		w = 40
	}
	return strings.Repeat("─", w)
}


