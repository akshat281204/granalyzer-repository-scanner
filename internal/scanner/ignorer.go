package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	"__pycache__":  true,
	".venv":        true,
	"target":       true,
	"bin":          true,
	"obj":          true,
	".idea":        true,
	".vscode":      true,
}

type Ignorer struct {
	patterns []string
	root     string
}

func NewIgnorer(root string) *Ignorer {
	ig := &Ignorer{
		patterns: []string{},
		root:     root,
	}
	ig.loadIgnoreFile(filepath.Join(root, ".gitignore"))
	ig.loadIgnoreFile(filepath.Join(root, ".granalyzerignore"))
	return ig
}

func (ig *Ignorer) loadIgnoreFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ig.patterns = append(ig.patterns, line)
	}
}

func (ig *Ignorer) ShouldIgnore(path string, isDir bool) bool {
	rel, err := filepath.Rel(ig.root, path)
	if err != nil {
		return false
	}

	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if defaultIgnoreDirs[part] {
			return true
		}
	}

	for _, pattern := range ig.patterns {
		pat := pattern
		if strings.HasPrefix(pat, "!") {
			continue
		}

		dirOnly := false
		if strings.HasSuffix(pat, "/") {
			dirOnly = true
			pat = strings.TrimSuffix(pat, "/")
		}

		if dirOnly && !isDir {
			continue
		}

		if matchPattern(rel, pat, parts) {
			return true
		}
	}

	return false
}

func matchPattern(rel string, pattern string, parts []string) bool {
	if strings.Contains(pattern, "/") {
		pat := strings.TrimPrefix(pattern, "/")
		match, err := filepath.Match(pat, rel)
		if err == nil && match {
			return true
		}
		if strings.HasSuffix(pat, "**") {
			prefix := strings.TrimSuffix(pat, "**")
			if strings.HasPrefix(rel, prefix) {
				return true
			}
		}
	} else {
		for _, part := range parts {
			match, err := filepath.Match(pattern, part)
			if err == nil && match {
				return true
			}
		}
	}
	return false
}
