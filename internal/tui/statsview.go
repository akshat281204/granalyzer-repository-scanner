package tui

import (
	"fmt"
	"sort"
	"strings"

	"granalyzer/internal/stats"

	"github.com/charmbracelet/bubbles/viewport"
)

type StatsViewModel struct {
	Viewport viewport.Model
	Height   int
	Width    int
}

func NewStatsViewModel() StatsViewModel {
	vp := viewport.New(0, 0)
	return StatsViewModel{
		Viewport: vp,
	}
}

func (m *StatsViewModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.Viewport.Width = w
	m.Viewport.Height = h
}

func (m *StatsViewModel) SetStats(s stats.RepoStats) {
	var builder strings.Builder

	builder.WriteString(AccentTextStyle.Render(" REPOSITORY OVERVIEW ") + "\n")
	builder.WriteString(RepeatDivider(m.Width) + "\n")

	sizeStr := formatSize(s.TotalSize)
	builder.WriteString(fmt.Sprintf("  %-22s %d\n", "Total Files:", s.TotalFiles))
	builder.WriteString(fmt.Sprintf("  %-22s %d\n", "Total Code Lines:", s.TotalLines))
	builder.WriteString(fmt.Sprintf("  %-22s %s\n", "Total Size:", sizeStr))
	builder.WriteString(fmt.Sprintf("  %-22s %s\n", "Deepest File Path:", s.DeepestPath))
	builder.WriteString(fmt.Sprintf("  %-22s %s\n", "Avg File Size:", formatSize(s.AverageFileSize)))
	builder.WriteString(fmt.Sprintf("  %-22s %d lines\n\n", "Avg Lines per File:", s.AverageFileLines))

	builder.WriteString(AccentTextStyle.Render(" LANGUAGE BREAKDOWN ") + "\n")
	builder.WriteString(RepeatDivider(m.Width) + "\n")

	builder.WriteString(fmt.Sprintf("  %-20s %10s %10s %10s\n",
		HeaderStyle.Render("Language"),
		HeaderStyle.Render("Files"),
		HeaderStyle.Render("Lines"),
		HeaderStyle.Render("Percentage")))

	type langItem struct {
		name  string
		stats stats.LangStats
	}
	var list []langItem
	for name, ls := range s.ByLanguage {
		list = append(list, langItem{name: name, stats: ls})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].stats.Files > list[j].stats.Files
	})

	for _, item := range list {
		builder.WriteString(fmt.Sprintf("  %-20s %10d %10d %9.1f%%\n",
			item.name,
			item.stats.Files,
			item.stats.Lines,
			item.stats.Percentage))
	}
	builder.WriteString("\n")

	builder.WriteString(AccentTextStyle.Render(" LARGEST FILES ") + "\n")
	builder.WriteString(RepeatDivider(m.Width) + "\n")

	for i, file := range s.LargestFiles {
		builder.WriteString(fmt.Sprintf("  %2d. %-50s %10d lines (%s)\n",
			i+1,
			truncatePath(file.RelPath, 50),
			file.Lines,
			formatSize(file.Size)))
	}

	m.Viewport.SetContent(builder.String())
	m.Viewport.GotoTop()
}

func (m StatsViewModel) View() string {
	return RenderWithTitle(BorderStyle, " Repository Stats ", m.Viewport.View(), m.Width, m.Height + 2)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
