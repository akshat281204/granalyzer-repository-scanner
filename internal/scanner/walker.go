package scanner

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileInfo struct {
	Path      string // Absolute path
	RelPath   string // Relative path from root
	Language  string
	Framework string
	Size      int64
	Lines     int
}

type TreeNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*TreeNode
	Expanded bool
}

type ScanResult struct {
	Root     string
	Files    []FileInfo
	Total    int
	Dirs     int
	ByLang   map[string]int // language -> file count
	Tree     *TreeNode
	Duration time.Duration
}

type WalkOptions struct {
	IgnorePatterns []string
	MaxDepth       int
}

func Walk(root string, opts WalkOptions) (*ScanResult, error) {
	startTime := time.Now()
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}

	ignorer := NewIgnorer(absRoot)
	// Add extra CLI ignore patterns
	if len(opts.IgnorePatterns) > 0 {
		ignorer.patterns = append(ignorer.patterns, opts.IgnorePatterns...)
	}

	result := &ScanResult{
		Root:   absRoot,
		ByLang: make(map[string]int),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 32) // limit concurrency to 32 goroutines

	// We'll collect all file paths first, then process them concurrently, OR walk and process concurrently.
	// Walking is fast, let's spawn goroutines during walking for text file analysis (line counting).
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // ignore individual path errors
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}

		if rel == "." {
			return nil
		}

		// Check depth
		if opts.MaxDepth > 0 {
			depth := strings.Count(rel, string(filepath.Separator)) + 1
			if depth > opts.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		isDir := d.IsDir()

		// Check if it should be ignored
		if ignorer.ShouldIgnore(path, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}

		if isDir {
			mu.Lock()
			result.Dirs++
			mu.Unlock()
			return nil
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(filePath string, relPath string) {
			defer wg.Done()
			defer func() { <-sem }()

			info, err := d.Info()
			if err != nil {
				return
			}

			lang, framework := ClassifyFile(filePath)
			lines := 0
			if isTextFile(filePath, lang) {
				lines = countFileLines(filePath)
			}

			fileInf := FileInfo{
				Path:      filePath,
				RelPath:   relPath,
				Language:  lang,
				Framework: framework,
				Size:      info.Size(),
				Lines:     lines,
			}

			mu.Lock()
			result.Files = append(result.Files, fileInf)
			result.ByLang[lang]++
			mu.Unlock()
		}(path, rel)

		return nil
	})

	wg.Wait()

	// Build the tree from the collected files
	result.Tree = &TreeNode{
		Name:     filepath.Base(absRoot),
		Path:     absRoot,
		IsDir:    true,
		Expanded: true,
	}

	for _, file := range result.Files {
		insertPath(result.Tree, file.Path, false)
	}

	result.Total = len(result.Files)
	result.Duration = time.Since(startTime)
	return result, err
}

func isTextFile(path string, lang string) bool {
	if lang == "Other" {
		// check extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".txt", ".conf", ".cfg", ".ini", ".log":
			return true
		default:
			return false
		}
	}
	return true
}

func countFileLines(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	// Fast line count using buffer
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := io.ReadFull(file, buf)
		count += bytes.Count(buf[:c], lineSep)

		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// check if last line does not end with newline
				if c > 0 && buf[c-1] != '\n' {
					count++
				}
				break
			}
			return 0
		}
	}
	return count
}

func insertPath(root *TreeNode, fullPath string, isDir bool) {
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
