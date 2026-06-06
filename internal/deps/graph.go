package deps

import (
	"path/filepath"
)

type Node struct {
	Path     string
	Language string
	Imports  []string
	IsRoot   bool // has no incoming dependencies
	IsLeaf   bool // has no outgoing dependencies
}

type Edge struct {
	From string
	To   string
}

type DepGraph struct {
	Nodes map[string]*Node
	Edges []Edge
}

func NewDepGraph() *DepGraph {
	return &DepGraph{
		Nodes: make(map[string]*Node),
		Edges: []Edge{},
	}
}

func (g *DepGraph) AddNode(path string, lang string, imports []string) {
	if _, ok := g.Nodes[path]; !ok {
		g.Nodes[path] = &Node{
			Path:     path,
			Language: lang,
			Imports:  imports,
			IsRoot:   true,
			IsLeaf:   len(imports) == 0,
		}
	}
}

func (g *DepGraph) AddEdge(from, to string) {
	g.Edges = append(g.Edges, Edge{From: from, To: to})
}

// BuildRelationships computes IsRoot and IsLeaf for all nodes
func (g *DepGraph) BuildRelationships() {
	hasImporter := make(map[string]bool)
	hasImports := make(map[string]bool)

	for _, edge := range g.Edges {
		hasImporter[edge.To] = true
		hasImports[edge.From] = true
	}

	for path, node := range g.Nodes {
		node.IsRoot = !hasImporter[path]
		node.IsLeaf = !hasImports[path]
	}
}

// ChildrenOf returns the outgoing dependencies for a node
func (g *DepGraph) ChildrenOf(nodePath string) []string {
	var children []string
	seen := make(map[string]bool) // prevent duplicate child edges
	for _, edge := range g.Edges {
		if edge.From == nodePath {
			if !seen[edge.To] {
				seen[edge.To] = true
				children = append(children, edge.To)
			}
		}
	}
	return children
}

// HasCycle detects if there are dependency cycles in the graph using DFS.
func (g *DepGraph) HasCycle() bool {
	visited := make(map[string]int) // 0 = unvisited, 1 = visiting, 2 = visited

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = 1 // visiting
		for _, child := range g.ChildrenOf(node) {
			if visited[child] == 1 {
				return true // found cycle
			}
			if visited[child] == 0 {
				if dfs(child) {
					return true
				}
			}
		}
		visited[node] = 2 // visited
		return false
	}

	for path := range g.Nodes {
		if visited[path] == 0 {
			if dfs(path) {
				return true
			}
		}
	}
	return false
}

// RenderTree renders the graph starting from root as an ASCII tree
func (g *DepGraph) RenderTree(root string) string {
	return g.renderTreeHelper(root, "", true, make(map[string]bool))
}

func (g *DepGraph) renderTreeHelper(root string, prefix string, isLast bool, visited map[string]bool) string {
	if visited[root] {
		// Prevent infinite recursion in case of cycles
		connector := "└── "
		if !isLast {
			connector = "├── "
		}
		return prefix + connector + filepath.Base(root) + " (cycle)\n"
	}
	visited[root] = true
	defer func() { visited[root] = false }()

	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	line := prefix + connector + filepath.Base(root) + "\n"
	children := g.ChildrenOf(root)
	for i, child := range children {
		line += g.renderTreeHelper(child, childPrefix, i == len(children)-1, visited)
	}
	return line
}
