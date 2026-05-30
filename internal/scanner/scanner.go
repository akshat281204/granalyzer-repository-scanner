package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type TreeNode struct {
	Name    string
	Path	string
	IsDir    bool
	Children []*TreeNode
	Expanded bool
}

type Stats struct {
	Files     int
	Dirs      int
	Languages map[string]int
	Tree      *TreeNode
}

func ScanRepository(root string) (Stats, error) {
	rootNode := &TreeNode{
		Name: 		"root",
		Path: 		root,
		IsDir:  	true,
		Expanded: 	true,
	}

	stats := Stats{
		Languages: make(map[string]int),
		Tree:      rootNode,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		if info.IsDir() {
			switch name {
			case "node_modules", ".git", "vendor":
				return filepath.SkipDir
			}
			stats.Dirs++
		} else {
			stats.Files++
			ext := filepath.Ext(name)
			if ext != "" {
				stats.Languages[ext]++
			}
		}

		InsertPath(rootNode, path, info.IsDir())
		return nil
	})

	return stats, err
}

func InsertPath(root *TreeNode, fullPath string, isDir bool) {
	rel, err := filepath.Rel(root.Path, fullPath)
	if err != nil || rel == "." {
		return
	}

	parts := strings.Split(rel, string(filepath.Separator))
	current := root
	currentPath := root.Path

	for i, part := range parts {
		currentPath = filepath.Join(currentPath, part)
		var existing *TreeNode
		for _, child := range current.Children {
			if child.Name == part {
				existing = child
				break
			}
		}

		if existing != nil {
			current = existing
		} else {
			node := &TreeNode{
				Name:     part,
				Path:     currentPath,
				IsDir:    i < len(parts)-1 || isDir,
				Expanded: false,
			}
			current.Children = append(current.Children, node)
			current = node
		}
	}
}