package tui

import (
	"fmt"
	"strings"

	"granalyzer/internal/deps"
	"granalyzer/internal/endpoints"
	"granalyzer/internal/scanner"
	"granalyzer/internal/stats"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScanCompleteMsg struct {
	Result *scanner.ScanResult
	Err    error
}

type Model struct {
	Path string
	Opts scanner.WalkOptions

	width, height int
	activeTab     int
	focusLeft     bool // True = FileTree, False = Preview

	scanning   bool
	scanResult *scanner.ScanResult
	spinner    spinner.Model
	scanError  error

	// Sub-panels (now in package tui)
	fileTree      FileTreeModel
	codePreview   PreviewModel
	statsPanel    StatsViewModel
	depsPanel     DepsViewModel
	endpointsList EndpointsViewModel

	searchMode  bool
	searchInput textinput.Model
}

func InitialModel(path string, opts scanner.WalkOptions) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorAccent)

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ColorAccent)
	ti.TextStyle = lipgloss.NewStyle().Foreground(ColorText)

	return Model{
		Path:        path,
		Opts:        opts,
		scanning:    true,
		spinner:     s,
		searchInput: ti,
		focusLeft:   true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startScanCmd(),
	)
}

func (m Model) startScanCmd() tea.Cmd {
	return func() tea.Msg {
		res, err := scanner.Walk(m.Path, m.Opts)
		return ScanCompleteMsg{Result: res, Err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

		if m.searchMode {
			switch msg.String() {
			case "enter", "esc":
				m.searchMode = false
				m.searchInput.Blur()
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)

				query := m.searchInput.Value()
				if m.activeTab == 0 {
					m.fileTree.SetFilter(query)
					m.updatePreview()
				} else if m.activeTab == 3 {
					m.endpointsList.SetFilter(query)
				}
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "1", "2", "3", "4":
			m.activeTab = int(msg.Runes[0] - '1')
			m.searchMode = false
			m.searchInput.Blur()

		case "tab":
			if m.activeTab == 0 {
				m.focusLeft = !m.focusLeft
				m.fileTree.Focus = m.focusLeft
				m.codePreview.Focus = !m.focusLeft
			}

		case "r":
			m.scanning = true
			m.scanError = nil
			return m, tea.Batch(m.spinner.Tick, m.startScanCmd())

		case "/":
			if m.activeTab == 0 || m.activeTab == 3 {
				m.searchMode = true
				m.searchInput.Focus()
				m.searchInput.SetValue("")
				if m.activeTab == 0 {
					m.fileTree.SetFilter("")
				} else if m.activeTab == 3 {
					m.endpointsList.SetFilter("")
				}
				return m, textinput.Blink
			}

		default:
			m = m.delegateKeys(msg, &cmds)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateSizes()

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case ScanCompleteMsg:
		m.scanning = false
		if msg.Err != nil {
			m.scanError = msg.Err
			return m, nil
		}
		m.scanResult = msg.Result
		m.populateViews()
		m.recalculateSizes()
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updatePreview() {
	selected := m.fileTree.GetSelectedNode()
	if selected != nil && !selected.IsDir {
		m.codePreview.LoadFile(selected.Path)
	} else {
		m.codePreview.Clear()
	}
}

func (m Model) delegateKeys(msg tea.KeyMsg, cmds *[]tea.Cmd) Model {
	switch m.activeTab {
	case 0: // Files Tab
		if m.focusLeft {
			switch msg.String() {
			case "up", "k":
				m.fileTree.MoveUp()
				m.updatePreview()
			case "down", "j":
				m.fileTree.MoveDown()
				m.updatePreview()
			case "enter":
				m.fileTree.ToggleExpand()
				m.updatePreview()
			}
		} else {
			var cmd tea.Cmd
			m.codePreview.Viewport, cmd = m.codePreview.Viewport.Update(msg)
			*cmds = append(*cmds, cmd)
		}

	case 1: // Stats Tab
		var cmd tea.Cmd
		m.statsPanel.Viewport, cmd = m.statsPanel.Viewport.Update(msg)
		*cmds = append(*cmds, cmd)

	case 2: // Deps Tab
		var cmd tea.Cmd
		m.depsPanel.Viewport, cmd = m.depsPanel.Viewport.Update(msg)
		*cmds = append(*cmds, cmd)

	case 3: // Endpoints Tab
		switch msg.String() {
		case "up", "k":
			m.endpointsList.MoveUp()
		case "down", "j":
			m.endpointsList.MoveDown()
		}
	}

	return m
}

func (m *Model) populateViews() {
	if m.scanResult == nil {
		return
	}

	m.fileTree = NewFileTreeModel(m.scanResult.Tree)
	m.fileTree.Focus = m.focusLeft

	m.codePreview = NewPreviewModel()
	m.codePreview.Focus = !m.focusLeft
	m.updatePreview()

	repoStats := stats.Analyze(m.scanResult)
	m.statsPanel = NewStatsViewModel()
	m.statsPanel.SetStats(repoStats)

	depGraph := deps.Resolve(m.scanResult)
	m.depsPanel = NewDepsViewModel()
	m.depsPanel.SetGraph(depGraph)

	eps := endpoints.Extract(m.scanResult)
	m.endpointsList = NewEndpointsViewModel()
	m.endpointsList.SetEndpoints(eps)
	m.endpointsList.Focus = true
}

func (m *Model) recalculateSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	panelHeight := m.height - 6
	if panelHeight < 1 {
		panelHeight = 1
	}

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 2
	m.fileTree.SetSize(leftWidth, panelHeight)
	m.codePreview.SetSize(rightWidth, panelHeight)

	m.statsPanel.SetSize(m.width-2, panelHeight)
	m.depsPanel.SetSize(m.width-2, panelHeight)
	m.endpointsList.SetSize(m.width-2, panelHeight)
}

func (m Model) View() string {
	if m.scanning {
		return fmt.Sprintf("\n  %s Scanning repository %s...\n", m.spinner.View(), m.Path)
	}

	if m.scanError != nil {
		return fmt.Sprintf("\n  Error scanning repository: %v\n\n  Press q to quit.\n", m.scanError)
	}

	var builder strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent).
		Background(lipgloss.Color("#22222a")).
		Padding(0, 2)
	headerText := fmt.Sprintf("granalyzer v0.1.0  |  repo: %s", m.Path)
	builder.WriteString(headerStyle.Render(headerText) + "\n\n")

	var tabs []string
	tabNames := []string{"[1] Files", "[2] Stats", "[3] Deps", "[4] Endpoints"}
	for i, name := range tabNames {
		if i == m.activeTab {
			tabs = append(tabs, ActiveTabStyle.Render(name))
		} else {
			tabs = append(tabs, TabStyle.Render(name))
		}
	}
	builder.WriteString(strings.Join(tabs, " ") + "\n")

	var panelContent string
	switch m.activeTab {
	case 0:
		panelContent = lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.fileTree.View(),
			m.codePreview.View(),
		)
	case 1:
		panelContent = m.statsPanel.View()
	case 2:
		panelContent = m.depsPanel.View()
	case 3:
		panelContent = m.endpointsList.View()
	}
	builder.WriteString(panelContent + "\n")

	if m.searchMode {
		builder.WriteString(
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorAccent).
				Padding(0, 1).
				Width(m.width - 2).
				Render(m.searchInput.View()),
		)
	} else {
		var bindings []HelpBinding
		switch m.activeTab {
		case 0:
			bindings = []HelpBinding{
				{Key: "1-4", Desc: "Switch Tabs"},
				{Key: "↑/↓", Desc: "Navigate"},
				{Key: "Enter", Desc: "Expand/Collapse"},
				{Key: "Tab", Desc: "Switch Focus"},
				{Key: "/", Desc: "Search"},
				{Key: "r", Desc: "Re-scan"},
				{Key: "q", Desc: "Quit"},
			}
		case 1:
			bindings = []HelpBinding{
				{Key: "1-4", Desc: "Switch Tabs"},
				{Key: "↑/↓", Desc: "Scroll"},
				{Key: "r", Desc: "Re-scan"},
				{Key: "q", Desc: "Quit"},
			}
		case 2:
			bindings = []HelpBinding{
				{Key: "1-4", Desc: "Switch Tabs"},
				{Key: "↑/↓", Desc: "Scroll"},
				{Key: "r", Desc: "Re-scan"},
				{Key: "q", Desc: "Quit"},
			}
		case 3:
			bindings = []HelpBinding{
				{Key: "1-4", Desc: "Switch Tabs"},
				{Key: "↑/↓", Desc: "Navigate"},
				{Key: "/", Desc: "Search"},
				{Key: "r", Desc: "Re-scan"},
				{Key: "q", Desc: "Quit"},
			}
		}
		builder.WriteString(RenderHelpBar(m.width, bindings))
	}

	return builder.String()
}
