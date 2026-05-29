/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"path/filepath"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("please provide a path to analyze")
			return
		}

		path := args[0]

		fileCount := 0
		dirCount := 0

		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				fileCount++
			} else {
				dirCount++
			}

			if err != nil {
				return err
			}

			fmt.Println(path)
			return nil
		})

		fmt.Printf("\nTotal files: %d\n", fileCount)
		fmt.Printf("Total directories: %d\n", dirCount)

		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", path, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// analyzeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// analyzeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
