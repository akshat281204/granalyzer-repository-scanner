package tui

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type PreviewModel struct {
	Viewport viewport.Model
	FilePath string
	Lines    []string
	Height   int
	Width    int
	Focus    bool
}

func NewPreviewModel() PreviewModel {
	vp := viewport.New(0, 0)
	return PreviewModel{
		Viewport: vp,
	}
}

func (m *PreviewModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Viewport.Width = w
	m.Viewport.Height = h
}

func (m *PreviewModel) LoadFile(path string) {
	if path == m.FilePath {
		return
	}
	m.FilePath = path
	m.Lines = []string{}

	file, err := os.Open(path)
	if err != nil {
		m.Viewport.SetContent(fmt.Sprintf("Error opening file: %v", err))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	var rawContent strings.Builder

	for scanner.Scan() && lineCount < 500 {
		line := scanner.Text()
		rawContent.WriteString(line)
		rawContent.WriteString("\n")
		lineCount++
	}

	truncated := false
	if scanner.Scan() {
		truncated = true
	}

	highlighted := highlight(rawContent.String(), filepath.Base(path))
	highlightedLines := strings.Split(highlighted, "\n")

	var finalContent strings.Builder
	numStyle := lipgloss.NewStyle().Foreground(ColorMuted)

	for i, line := range highlightedLines {
		if i >= lineCount {
			break
		}
		lineNum := numStyle.Render(fmt.Sprintf("%4d │ ", i+1))
		finalContent.WriteString(lineNum + line + "\n")
	}

	if truncated {
		finalContent.WriteString("\n" + lipgloss.NewStyle().Foreground(ColorWarn).Bold(true).Render(" [File preview truncated to first 500 lines]"))
	}

	m.Viewport.SetContent(finalContent.String())
	m.Viewport.GotoTop()
}

func (m *PreviewModel) Clear() {
	m.FilePath = ""
	m.Lines = []string{}
	m.Viewport.SetContent("")
}

func (m PreviewModel) View() string {
	style := BorderStyle
	if m.Focus {
		style = ActiveBorderStyle
	}

	title := " File Preview "
	if m.FilePath != "" {
		title = fmt.Sprintf(" Preview: %s ", filepath.Base(m.FilePath))
	}

	return RenderWithTitle(style, title, m.Viewport.View(), m.Width, m.Height + 2)
}

func highlight(content, filename string) string {
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula") // clean premium style
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	var buf bytes.Buffer
	iter, err := lexer.Tokenise(nil, content)
	if err == nil {
		formatter.Format(&buf, style, iter)
		return buf.String()
	}
	return content
}
