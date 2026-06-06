package deps

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"granalyzer/internal/parser"
	"granalyzer/internal/scanner"
)

func Resolve(result *scanner.ScanResult) *DepGraph {
	graph := NewDepGraph()

	// 1. First read go.mod to extract Go module name
	goModuleName := readGoModuleName(result.Root)

	// Keep a map of relative file path -> FileInfo for easy lookups
	filesMap := make(map[string]scanner.FileInfo)
	// Keep a map of package/directory path -> list of Go files in that directory
	goDirMap := make(map[string][]string)

	for _, file := range result.Files {
		filesMap[file.RelPath] = file
		if file.Language == "Go" {
			dir := filepath.Dir(file.RelPath)
			goDirMap[dir] = append(goDirMap[dir], file.RelPath)
		}
	}

	// 2. Add all files as nodes in the graph
	fileImports := make(map[string][]string)
	for _, file := range result.Files {
		parseRes := parser.ParseFile(file.Path, file.RelPath, file.Language, file.Framework)
		graph.AddNode(file.RelPath, file.Language, parseRes.Imports)
		fileImports[file.RelPath] = parseRes.Imports
	}

	// 3. Resolve each import
	for relPath, imports := range fileImports {
		fileInfo := filesMap[relPath]
		dirOfFile := filepath.Dir(relPath)

		for _, imp := range imports {
			resolved := false

			switch fileInfo.Language {
			case "Go":
				// Standard format: "granalyzer/internal/scanner"
				if goModuleName != "" && strings.HasPrefix(imp, goModuleName) {
					// Extract relative directory path
					subPath := strings.TrimPrefix(imp, goModuleName)
					subPath = strings.TrimPrefix(subPath, "/")
					subPath = filepath.Clean(subPath)

					// Go files in that directory are dependencies
					if filesInDir, ok := goDirMap[subPath]; ok {
						for _, targetFile := range filesInDir {
							graph.AddEdge(relPath, targetFile)
							resolved = true
						}
					}
				}

			case "JavaScript", "TypeScript":
				// Standard formats: "./helper", "../utils", "express"
				if strings.HasPrefix(imp, ".") {
					// Clean relative target path
					targetRel := filepath.Clean(filepath.Join(dirOfFile, imp))
					// Try common extensions
					extensions := []string{"", ".ts", ".js", ".tsx", ".jsx", "/index.ts", "/index.js"}
					for _, ext := range extensions {
						checkPath := targetRel + ext
						if _, exists := filesMap[checkPath]; exists {
							graph.AddEdge(relPath, checkPath)
							resolved = true
							break
						}
					}
				}

			case "Python":
				// Standard formats: "import os", "from .module import something", "from package.subpackage import module"
				if strings.HasPrefix(imp, ".") {
					// Relative python import
					// e.g. .module or ..sub.module
					dots := 0
					for _, char := range imp {
						if char == '.' {
							dots++
						} else {
							break
						}
					}
					cleanImp := imp[dots:]
					cleanImp = strings.ReplaceAll(cleanImp, ".", string(filepath.Separator))

					// Calculate parent directory based on number of dots
					currDir := dirOfFile
					for i := 1; i < dots; i++ {
						currDir = filepath.Dir(currDir)
					}

					targetRel := filepath.Clean(filepath.Join(currDir, cleanImp))
					extensions := []string{"", ".py", "/__init__.py"}
					for _, ext := range extensions {
						checkPath := targetRel + ext
						if _, exists := filesMap[checkPath]; exists {
							graph.AddEdge(relPath, checkPath)
							resolved = true
							break
						}
					}
				} else {
					// Absolute python import, see if it matches any scanned python files
					// e.g., import mymodule -> check for mymodule.py in root or source dirs
					importPath := strings.ReplaceAll(imp, ".", string(filepath.Separator))
					for checkPath := range filesMap {
						if strings.HasSuffix(checkPath, importPath+".py") || strings.HasSuffix(checkPath, importPath+"/__init__.py") {
							graph.AddEdge(relPath, checkPath)
							resolved = true
							break
						}
					}
				}
			}

			// If it's not resolved locally and we want external grouping,
			// we can add a dependency to an external node if desired, but we only draw local detail.
			_ = resolved
		}
	}

	graph.BuildRelationships()
	return graph
}

func readGoModuleName(root string) string {
	goModPath := filepath.Join(root, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
