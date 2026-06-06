package tui

import (
	"fmt"
	"strings"

	"granalyzer/internal/parser"

	"github.com/charmbracelet/lipgloss"
)

type EndpointsViewModel struct {
	Endpoints         []parser.Endpoint
	FilteredEndpoints []parser.Endpoint
	Cursor            int
	Scroll            int
	Filter            string
	Height            int
	Width             int
	Focus             bool
}

func NewEndpointsViewModel() EndpointsViewModel {
	return EndpointsViewModel{
		Endpoints:         []parser.Endpoint{},
		FilteredEndpoints: []parser.Endpoint{},
	}
}

func (m *EndpointsViewModel) SetEndpoints(eps []parser.Endpoint) {
	m.Endpoints = eps
	m.ApplyFilter()
}

func (m *EndpointsViewModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

func (m *EndpointsViewModel) SetFilter(q string) {
	m.Filter = q
	m.ApplyFilter()
}

func (m *EndpointsViewModel) ApplyFilter() {
	if m.Filter == "" {
		m.FilteredEndpoints = m.Endpoints
	} else {
		q := strings.ToLower(m.Filter)
		var filtered []parser.Endpoint
		for _, ep := range m.Endpoints {
			if strings.Contains(strings.ToLower(ep.Path), q) ||
				strings.Contains(strings.ToLower(ep.Method), q) ||
				strings.Contains(strings.ToLower(ep.File), q) ||
				strings.Contains(strings.ToLower(ep.Framework), q) {
				filtered = append(filtered, ep)
			}
		}
		m.FilteredEndpoints = filtered
	}
	m.Cursor = 0
	m.Scroll = 0
}

func (m *EndpointsViewModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
	m.adjustScroll()
}

func (m *EndpointsViewModel) MoveDown() {
	if m.Cursor < len(m.FilteredEndpoints)-1 {
		m.Cursor++
	}
	m.adjustScroll()
}

func (m *EndpointsViewModel) adjustScroll() {
	visibleHeight := m.Height - 3
	if visibleHeight <= 0 {
		return
	}

	if m.Cursor < m.Scroll {
		m.Scroll = m.Cursor
	} else if m.Cursor >= m.Scroll+visibleHeight {
		m.Scroll = m.Cursor - visibleHeight + 1
	}
}

func (m EndpointsViewModel) View() string {
	if len(m.FilteredEndpoints) == 0 {
		msg := "No endpoints detected."
		if m.Filter != "" {
			msg = fmt.Sprintf("No matches for '%s'", m.Filter)
		}
		return RenderWithTitle(
			BorderStyle,
			" Endpoints ",
			lipgloss.NewStyle().
				Height(m.Height).
				Width(m.Width).
				Align(lipgloss.Center, lipgloss.Center).
				Foreground(ColorMuted).
				Render(msg),
			m.Width,
			m.Height + 2,
		)
	}

	var lines []string

	header := fmt.Sprintf("  %-8s  %-45s  %-30s  %-10s",
		HeaderStyle.Render("Method"),
		HeaderStyle.Render("Path"),
		HeaderStyle.Render("Location"),
		HeaderStyle.Render("Framework"))
	lines = append(lines, header)
	lines = append(lines, RepeatDivider(m.Width))

	visibleHeight := m.Height - 2
	if visibleHeight <= 0 {
		visibleHeight = 1
	}

	start := m.Scroll
	end := start + visibleHeight
	if end > len(m.FilteredEndpoints) {
		end = len(m.FilteredEndpoints)
	}

	for i := start; i < end; i++ {
		ep := m.FilteredEndpoints[i]

		methodStr := ep.Method
		var methodStyle lipgloss.Style
		switch ep.Method {
		case "GET":
			methodStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
		case "POST":
			methodStyle = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
		case "PUT", "PATCH":
			methodStyle = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)
		case "DELETE":
			methodStyle = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
		default:
			methodStyle = lipgloss.NewStyle().Foreground(ColorMuted).Bold(true)
		}
		styledMethod := methodStyle.Render(fmt.Sprintf("%-6s", methodStr))

		pathStr := ep.Path
		if len(pathStr) > 43 {
			pathStr = pathStr[:40] + "..."
		}

		locationStr := fmt.Sprintf("%s:%d", ep.File, ep.Line)
		if len(locationStr) > 28 {
			locationStr = "..." + locationStr[len(locationStr)-25:]
		}

		fwStr := ep.Framework
		if fwStr == "" {
			fwStr = "unknown"
		}

		row := fmt.Sprintf("  %-6s    %-43s  %-28s  %-10s",
			styledMethod,
			pathStr,
			locationStr,
			fwStr)

		if i == m.Cursor {
			if m.Focus {
				row = AccentTextStyle.Render("> " + row[2:])
			} else {
				row = lipgloss.NewStyle().Foreground(ColorText).Background(ColorBorder).Render("> " + row[2:])
			}
		} else {
			row = "  " + row[2:]
		}

		lines = append(lines, row)
	}

	for len(lines) < m.Height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	style := BorderStyle
	if m.Focus {
		style = ActiveBorderStyle
	}

	return RenderWithTitle(style, " Endpoints ", content, m.Width, m.Height + 2)
}
