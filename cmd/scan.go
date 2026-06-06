package cmd

import (
	"granalyzer/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Launches the full interactive TUI (default: current directory)",
	Run: func(cmd *cobra.Command, args []string) {
		path := getRepoPath(args)
		opts := getWalkOptions()

		p := tea.NewProgram(tui.InitialModel(path, opts), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			printErrorAndExit(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
