package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"granalyzer/internal/scanner"

	"github.com/charmbracelet/lipgloss"
)

type VisibleNode struct {
	Node  *scanner.TreeNode
	Depth int
}

type FileTreeModel struct {
	Root         *scanner.TreeNode
	FilteredRoot *scanner.TreeNode
	Cursor       int
	Scroll       int
	Filter       string
	Focus        bool
	Height       int
	Width        int
}

func NewFileTreeModel(root *scanner.TreeNode) FileTreeModel {
	return FileTreeModel{
		Root:         root,
		FilteredRoot: root,
		Focus:        true,
	}
}

func (m *FileTreeModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

func (m *FileTreeModel) SetHeight(h int) {
	m.Height = h
}

func (m *FileTreeModel) SetWidth(w int) {
	m.Width = w
}

func (m *FileTreeModel) SetFilter(q string) {
	m.Filter = q
	if q == "" {
		m.FilteredRoot = m.Root
	} else {
		m.FilteredRoot = filterTree(m.Root, strings.ToLower(q))
	}
	m.Cursor = 0
	m.Scroll = 0
}

func (m *FileTreeModel) GetSelectedNode() *scanner.TreeNode {
	nodes := m.GetVisibleNodes()
	if len(nodes) == 0 || m.Cursor < 0 || m.Cursor >= len(nodes) {
		return nil
	}
	return nodes[m.Cursor].Node
}

func (m *FileTreeModel) GetVisibleNodes() []VisibleNode {
	var nodes []VisibleNode
	flattenTree(m.FilteredRoot, 0, &nodes)
	return nodes
}

func (m *FileTreeModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
	m.adjustScroll()
}

func (m *FileTreeModel) MoveDown() {
	nodes := m.GetVisibleNodes()
	if m.Cursor < len(nodes)-1 {
		m.Cursor++
	}
	m.adjustScroll()
}

func (m *FileTreeModel) ToggleExpand() {
	node := m.GetSelectedNode()
	if node != nil && node.IsDir {
		node.Expanded = !node.Expanded
		m.SetFilter(m.Filter)
	}
}

func (m *FileTreeModel) adjustScroll() {
	if m.Height <= 0 {
		return
	}
	visibleHeight := m.Height

	if m.Cursor < m.Scroll {
		m.Scroll = m.Cursor
	} else if m.Cursor >= m.Scroll+visibleHeight {
		m.Scroll = m.Cursor - visibleHeight + 1
	}
}

func (m FileTreeModel) View() string {
	nodes := m.GetVisibleNodes()
	if len(nodes) == 0 {
		emptyMsg := "No files found"
		if m.Filter != "" {
			emptyMsg = fmt.Sprintf("No matches for '%s'", m.Filter)
		}
		return lipgloss.NewStyle().
			Height(m.Height).
			Width(m.Width).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(ColorMuted).
			Render(emptyMsg)
	}

	var lines []string
	start := m.Scroll
	end := start + m.Height
	if end > len(nodes) {
		end = len(nodes)
	}

	for i := start; i < end; i++ {
		visible := nodes[i]
		node := visible.Node
		depth := visible.Depth

		indent := strings.Repeat("  ", depth)
		icon := "📄"
		if node.IsDir {
			if node.Expanded {
				icon = "📂"
			} else {
				icon = "📁"
			}
		}

		name := node.Name
		if node.Path == m.Root.Path {
			name = filepath.Base(node.Path)
		}

		line := fmt.Sprintf("%s%s %s", indent, icon, name)

		if len(line) > m.Width-4 {
			line = line[:m.Width-7] + "..."
		}

		if i == m.Cursor {
			if m.Focus {
				line = AccentTextStyle.Render("> " + line)
			} else {
				line = lipgloss.NewStyle().Foreground(ColorText).Background(ColorBorder).Render("> " + line)
			}
		} else {
			line = "  " + line
		}

		lines = append(lines, line)
	}

	for len(lines) < m.Height {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	style := BorderStyle
	if m.Focus {
		style = ActiveBorderStyle
	}

	return RenderWithTitle(style, " File Tree ", content, m.Width, m.Height + 2)
}

func flattenTree(node *scanner.TreeNode, depth int, nodes *[]VisibleNode) {
	if node == nil {
		return
	}
	*nodes = append(*nodes, VisibleNode{
		Node:  node,
		Depth: depth,
	})

	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			flattenTree(child, depth+1, nodes)
		}
	}
}

func filterTree(node *scanner.TreeNode, query string) *scanner.TreeNode {
	if node == nil {
		return nil
	}

	if !node.IsDir {
		if strings.Contains(strings.ToLower(node.Name), query) {
			return &scanner.TreeNode{
				Name:  node.Name,
				Path:  node.Path,
				IsDir: false,
			}
		}
		return nil
	}

	var filteredChildren []*scanner.TreeNode
	for _, child := range node.Children {
		filteredChild := filterTree(child, query)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, filteredChild)
		}
	}

	if len(filteredChildren) > 0 || strings.Contains(strings.ToLower(node.Name), query) {
		return &scanner.TreeNode{
			Name:     node.Name,
			Path:     node.Path,
			IsDir:    true,
			Children: filteredChildren,
			Expanded: true,
		}
	}

	return nil
}
