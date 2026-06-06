package cmd

import (
	"fmt"
	"os"
	"strings"

	"granalyzer/internal/scanner"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var (
	ignoreFlag  string
	depthFlag   int
	noColorFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "granalyzer",
	Short: "A language/framework-agnostic repository scanner and analyzer",
	Long: `granalyzer is a static analysis CLI & TUI tool that walks your repository,
counts files, directories, lines, detects framework endpoints, and maps
internal dependencies without sending any data to external services.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColorFlag {
			lipgloss.SetColorProfile(termenv.Ascii)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// By default, launch the scan (TUI) command
		scanCmd.Run(cmd, args)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ignoreFlag, "ignore", "", "Comma-separated glob patterns to ignore")
	rootCmd.PersistentFlags().IntVar(&depthFlag, "depth", 0, "Max directory depth to scan (default: unlimited)")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable color output")
}

// Helper to construct WalkOptions from flags
func getWalkOptions() scanner.WalkOptions {
	var patterns []string
	if ignoreFlag != "" {
		for _, pat := range strings.Split(ignoreFlag, ",") {
			trimmed := strings.TrimSpace(pat)
			if trimmed != "" {
				patterns = append(patterns, trimmed)
			}
		}
	}
	return scanner.WalkOptions{
		IgnorePatterns: patterns,
		MaxDepth:       depthFlag,
	}
}

// Helper to determine the repository path from args
func getRepoPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func printErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
