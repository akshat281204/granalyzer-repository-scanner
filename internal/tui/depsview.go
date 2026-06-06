package tui

import (
	"strings"

	"granalyzer/internal/deps"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type DepsViewModel struct {
	Viewport viewport.Model
	Height   int
	Width    int
}

func NewDepsViewModel() DepsViewModel {
	vp := viewport.New(0, 0)
	return DepsViewModel{
		Viewport: vp,
	}
}

func (m *DepsViewModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Viewport.Width = w
	m.Viewport.Height = h
}

func (m *DepsViewModel) SetGraph(graph *deps.DepGraph) {
	var builder strings.Builder

	var roots []string
	for path, node := range graph.Nodes {
		if node.IsRoot {
			roots = append(roots, path)
		}
	}

	if len(roots) == 0 && len(graph.Nodes) > 0 {
		for path := range graph.Nodes {
			roots = append(roots, path)
			break
		}
	}

	builder.WriteString(AccentTextStyle.Render(" INTERNAL DEPENDENCY GRAPH ") + "\n")
	builder.WriteString(RepeatDivider(m.Width) + "\n\n")

	if len(graph.Nodes) == 0 {
		builder.WriteString("  No dependencies detected.\n")
		m.Viewport.SetContent(builder.String())
		return
	}

	nodeCount := 0
	maxNodes := 200
	truncated := false

	var renderNode func(root string, prefix string, isLast bool, visited map[string]bool) string
	renderNode = func(root string, prefix string, isLast bool, visited map[string]bool) string {
		if nodeCount >= maxNodes {
			truncated = true
			return ""
		}
		nodeCount++

		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		nodeName := root
		if visited[root] {
			return prefix + connector + nodeName + " (cycle)\n"
		}

		visited[root] = true
		defer func() { visited[root] = false }()

		line := prefix + connector + nodeName + "\n"
		children := graph.ChildrenOf(root)
		for i, child := range children {
			childLine := renderNode(child, childPrefix, i == len(children)-1, visited)
			if childLine != "" {
				line += childLine
			}
		}
		return line
	}

	visited := make(map[string]bool)
	for i, root := range roots {
		if nodeCount >= maxNodes {
			truncated = true
			break
		}
		builder.WriteString(renderNode(root, "", i == len(roots)-1, visited))
	}

	if truncated {
		builder.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorWarn).Bold(true).Render(" [Dependency graph truncated to 200 nodes]"))
	}

	m.Viewport.SetContent(builder.String())
	m.Viewport.GotoTop()
}

func (m DepsViewModel) View() string {
	return RenderWithTitle(BorderStyle, " Dependency Graph ", m.Viewport.View(), m.Width, m.Height + 2)
}
