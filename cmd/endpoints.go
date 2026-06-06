package cmd

import (
	"encoding/json"
	"fmt"

	"granalyzer/internal/endpoints"
	"granalyzer/internal/scanner"

	"github.com/spf13/cobra"
)

var endpointsJsonFlag bool

var endpointsCmd = &cobra.Command{
	Use:   "endpoints [path]",
	Short: "Prints detected endpoints and exits",
	Run: func(cmd *cobra.Command, args []string) {
		path := getRepoPath(args)
		opts := getWalkOptions()

		result, err := scanner.Walk(path, opts)
		if err != nil {
			printErrorAndExit(err)
		}

		eps := endpoints.Extract(result)

		if endpointsJsonFlag {
			data, err := json.MarshalIndent(eps, "", "  ")
			if err != nil {
				printErrorAndExit(err)
			}
			fmt.Println(string(data))
			return
		}

		// Text/CLI Output
		fmt.Printf("Detected HTTP Endpoints (%d found):\n", len(eps))
		fmt.Printf("===================================\n")
		fmt.Printf("%-8s  %-40s  %-30s  %-10s\n", "Method", "Path", "File:Line", "Framework")
		fmt.Printf("%-8s  %-40s  %-30s  %-10s\n", "------", "----", "---------", "---------")

		for _, ep := range eps {
			location := fmt.Sprintf("%s:%d", ep.File, ep.Line)
			if len(location) > 30 {
				location = "..." + location[len(location)-27:]
			}
			fw := ep.Framework
			if fw == "" {
				fw = "unknown"
			}
			fmt.Printf("%-8s  %-40s  %-30s  %-10s\n", ep.Method, ep.Path, location, fw)
		}
	},
}

func init() {
	endpointsCmd.Flags().BoolVar(&endpointsJsonFlag, "json", false, "Output endpoints as JSON")
	rootCmd.AddCommand(endpointsCmd)
}
