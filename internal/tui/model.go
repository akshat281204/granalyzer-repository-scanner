package tui

import (
	"fmt"

	"granalyzer/internal/scanner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	choices    []string
	cursor     int
	fileCursor int
	stats      scanner.Stats
	activeView string
}

func InitialModel() Model {

	stats, _ := scanner.ScanRepository(".")

	return Model{
		choices: []string{
			"Analyze Repository",
			"Dependency Graph",
			"Endpoints",
			"Exit",
		},
		stats:      stats,
		activeView: "menu",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":

			if m.activeView == "menu" {

				if m.cursor > 0 {
					m.cursor--
				}

			} else if m.activeView == "analyze" {

				if m.fileCursor > 0 {
					m.fileCursor--
				}
			}

		case "down", "j":

			if m.activeView == "menu" {

				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			} else if m.activeView == "analyze" {

				visibleNodes := []VisibleNode{}
				flattenTree(m.stats.Tree, 0, &visibleNodes)

				if m.fileCursor < len(visibleNodes)-1 {
					m.fileCursor++
				}
			}

		case "enter":
			if m.activeView == "menu" {
				selected := m.choices[m.cursor]

				switch selected {

				case "Analyze Repository":
					m.activeView = "analyze"

				case "Dependency Graph":
					m.activeView = "graph"

				case "Endpoints":
					m.activeView = "endpoints"

				case "Exit":
					return m, tea.Quit
				}
			} else if m.activeView == "analyze" {
				visibleNodes := []VisibleNode{}
				flattenTree(m.stats.Tree, 0, &visibleNodes)

				if len(visibleNodes) > 0 && m.fileCursor < len(visibleNodes) {
					selectedNode := visibleNodes[m.fileCursor].Node
					if selectedNode.IsDir {
						selectedNode.Expanded = !selectedNode.Expanded
					}
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {

	if m.activeView == "analyze" {

		leftStyle := lipgloss.NewStyle().
			Width(60).
			Border(lipgloss.RoundedBorder()).
			Padding(1)

		rightStyle := lipgloss.NewStyle().
			Width(40).
			Border(lipgloss.RoundedBorder()).
			Padding(1)

		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

		fileTree := "File Tree\n\n"

		visibleNodes := []VisibleNode{}
		flattenTree(m.stats.Tree, 0, &visibleNodes)

		for i, visible := range visibleNodes {
			node := visible.Node
			depth := visible.Depth
			prefix := ""

			for j := 0; j < depth; j++ {
				prefix += "  "
			}
			icon := "📄"
			if node.IsDir {
				if node.Expanded {
					icon = "🗂️"
				} else {
					icon = "🗂️"
				}
			}
			line := prefix + icon + " " + node.Name
			if i == m.fileCursor {
				line = selectedStyle.Render("> " + line)
			} else {
				line = "  " + line
			}

			fileTree += line + "\n"	
		}

		selectedFile := ""

		if len(visibleNodes) > 0 && m.fileCursor < len(visibleNodes) {
			selectedFile = visibleNodes[m.fileCursor].Node.Path
		}

		stats := "Repository Stats\n\n"

		stats += fmt.Sprintf("Files: %d\n", m.stats.Files)
		stats += fmt.Sprintf("Directories: %d\n\n", m.stats.Dirs)

		stats += "Languages\n\n"

		for lang, count := range m.stats.Languages {
			stats += fmt.Sprintf("%s : %d\n", lang, count)
		}

		stats += "\nSelected File\n\n"
		stats += selectedFile

		leftPanel := leftStyle.Render(fileTree)
		rightPanel := rightStyle.Render(stats)

		ui := lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPanel,
			rightPanel,
		)

		ui += "\n\nPress ESC to go back."

		return ui
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	s := titleStyle.Render("Granalyzer") + "\n\n"

	for i, choice := range m.choices {

		cursor := " "

		if m.cursor == i {
			cursor = ">"
		}

		line := fmt.Sprintf("%s %s", cursor, choice)

		if m.cursor == i {
			line = cursorStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\nPress q to quit.\n"

	return s
}

func renderTree(node *scanner.TreeNode, depth int, lines *[]string) {

	prefix := ""

	for i := 0; i < depth; i++ {
		prefix += "  "
	}

	icon := "📄"

	if node.IsDir {
		icon = "🗂️"
	}

	*lines = append(*lines, prefix+icon+" "+node.Name)

	if node.IsDir && node.Expanded {

		for _, child := range node.Children {
			renderTree(child, depth+1, lines)
		}
	}
}

type VisibleNode struct {
	Node *scanner.TreeNode
	Depth int
}

func flattenTree(node *scanner.TreeNode, depth int, nodes *[]VisibleNode) {

	*nodes = append(*nodes, VisibleNode{
		Node: node,
		Depth: depth,
	})

	if node.IsDir && node.Expanded {

		for _, child := range node.Children {
			flattenTree(child, depth + 1, nodes)
		}
	}
}