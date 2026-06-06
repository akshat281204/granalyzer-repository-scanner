package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"granalyzer/internal/deps"
	"granalyzer/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	depsJsonFlag bool
	depsDotFlag  bool
)

var depsCmd = &cobra.Command{
	Use:   "deps [path]",
	Short: "Prints the dependency graph and exits",
	Run: func(cmd *cobra.Command, args []string) {
		path := getRepoPath(args)
		opts := getWalkOptions()

		result, err := scanner.Walk(path, opts)
		if err != nil {
			printErrorAndExit(err)
		}

		graph := deps.Resolve(result)

		if depsJsonFlag {
			data, err := json.MarshalIndent(graph, "", "  ")
			if err != nil {
				printErrorAndExit(err)
			}
			fmt.Println(string(data))
			return
		}

		if depsDotFlag {
			fmt.Print(generateDOT(graph))
			return
		}

		// Text/ASCII Output (Tree style)
		// Find roots
		var roots []string
		for path, node := range graph.Nodes {
			if node.IsRoot {
				roots = append(roots, path)
			}
		}

		// Fallback if cyclic
		if len(roots) == 0 && len(graph.Nodes) > 0 {
			for path := range graph.Nodes {
				roots = append(roots, path)
				break
			}
		}

		fmt.Printf("Dependency Tree:\n")
		fmt.Printf("================\n")
		for _, root := range roots {
			fmt.Print(graph.RenderTree(root))
		}
	},
}

func generateDOT(graph *deps.DepGraph) string {
	var builder strings.Builder
	builder.WriteString("digraph G {\n")
	builder.WriteString("  node [shape=box, style=filled, color=lightgrey];\n")
	for _, edge := range graph.Edges {
		builder.WriteString(fmt.Sprintf("  %q -> %q;\n", edge.From, edge.To))
	}
	builder.WriteString("}\n")
	return builder.String()
}

func init() {
	depsCmd.Flags().BoolVar(&depsJsonFlag, "json", false, "Output dependency graph as JSON")
	depsCmd.Flags().BoolVar(&depsDotFlag, "dot", false, "Output dependency graph in Graphviz DOT format")
	rootCmd.AddCommand(depsCmd)
}
