package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"granalyzer/internal/scanner"
	"granalyzer/internal/stats"

	"github.com/spf13/cobra"
)

var statsJsonFlag bool

var statsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Prints repository stats and exits",
	Run: func(cmd *cobra.Command, args []string) {
		path := getRepoPath(args)
		opts := getWalkOptions()

		result, err := scanner.Walk(path, opts)
		if err != nil {
			printErrorAndExit(err)
		}

		repoStats := stats.Analyze(result)

		if statsJsonFlag {
			data, err := json.MarshalIndent(repoStats, "", "  ")
			if err != nil {
				printErrorAndExit(err)
			}
			fmt.Println(string(data))
			return
		}

		// Text/CLI Output
		fmt.Printf("granalyzer Repository Stats\n")
		fmt.Printf("===========================\n")
		fmt.Printf("Scan Path:        %s\n", path)
		fmt.Printf("Total Files:      %d\n", repoStats.TotalFiles)
		fmt.Printf("Total Code Lines: %d\n", repoStats.TotalLines)
		fmt.Printf("Total Size:       %s\n", formatSize(repoStats.TotalSize))
		fmt.Printf("Deepest Path:     %s\n", repoStats.DeepestPath)
		fmt.Printf("Avg File Size:    %s\n", formatSize(repoStats.AverageFileSize))
		fmt.Printf("Avg File Lines:   %d\n\n", repoStats.AverageFileLines)

		fmt.Printf("Language Breakdown:\n")
		fmt.Printf("-------------------\n")
		fmt.Printf("%-20s %10s %10s %10s\n", "Language", "Files", "Lines", "Percentage")
		fmt.Printf("%-20s %10s %10s %10s\n", "--------", "-----", "-----", "----------")

		type langItem struct {
			name  string
			stats stats.LangStats
		}
		var list []langItem
		for name, ls := range repoStats.ByLanguage {
			list = append(list, langItem{name: name, stats: ls})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].stats.Files > list[j].stats.Files
		})

		for _, item := range list {
			fmt.Printf("%-20s %10d %10d %9.1f%%\n",
				item.name,
				item.stats.Files,
				item.stats.Lines,
				item.stats.Percentage)
		}
		fmt.Println()

		fmt.Printf("Top 5 Largest Files:\n")
		fmt.Printf("--------------------\n")
		limit := 5
		if len(repoStats.LargestFiles) < limit {
			limit = len(repoStats.LargestFiles)
		}
		for i := 0; i < limit; i++ {
			f := repoStats.LargestFiles[i]
			fmt.Printf("%d. %-50s %10d lines (%s)\n",
				i+1,
				f.RelPath,
				f.Lines,
				formatSize(f.Size))
		}
	},
}

func init() {
	statsCmd.Flags().BoolVar(&statsJsonFlag, "json", false, "Output stats as JSON")
	rootCmd.AddCommand(statsCmd)
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
