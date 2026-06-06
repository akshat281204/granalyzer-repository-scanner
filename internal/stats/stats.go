package stats

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"granalyzer/internal/scanner"
)

type LangStats struct {
	Files      int
	Lines      int
	Percentage float64
}

type RepoStats struct {
	TotalFiles       int
	TotalLines       int
	TotalSize        int64
	ByLanguage       map[string]LangStats
	LargestFiles     []scanner.FileInfo // top 10
	DeepestPath      string
	AverageFileSize  int64
	AverageFileLines int
	ScannedAt        time.Time
	ScanDuration     time.Duration
}

func Analyze(result *scanner.ScanResult) RepoStats {
	stats := RepoStats{
		TotalFiles:   result.Total,
		ByLanguage:   make(map[string]LangStats),
		ScannedAt:    time.Now(),
		ScanDuration: result.Duration,
	}

	if len(result.Files) == 0 {
		return stats
	}

	var totalLines int
	var totalSize int64
	maxDepth := -1
	deepestFile := ""

	// Temporary map to calculate file count and lines per language
	langFiles := make(map[string]int)
	langLines := make(map[string]int)

	for _, file := range result.Files {
		totalLines += file.Lines
		totalSize += file.Size

		langFiles[file.Language]++
		langLines[file.Language] += file.Lines

		// Calculate depth
		depth := strings.Count(file.RelPath, string(filepath.Separator))
		if depth > maxDepth {
			maxDepth = depth
			deepestFile = file.RelPath
		}
	}

	stats.TotalLines = totalLines
	stats.TotalSize = totalSize
	stats.DeepestPath = deepestFile
	stats.AverageFileSize = totalSize / int64(len(result.Files))
	stats.AverageFileLines = totalLines / len(result.Files)

	// Calculate language percentages
	for lang, filesCount := range langFiles {
		linesCount := langLines[lang]
		pct := (float64(filesCount) / float64(result.Total)) * 100.0
		stats.ByLanguage[lang] = LangStats{
			Files:      filesCount,
			Lines:      linesCount,
			Percentage: pct,
		}
	}

	// Sort files by size to get top 10 largest
	sortedFiles := make([]scanner.FileInfo, len(result.Files))
	copy(sortedFiles, result.Files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Size > sortedFiles[j].Size
	})

	limit := 10
	if len(sortedFiles) < limit {
		limit = len(sortedFiles)
	}
	stats.LargestFiles = sortedFiles[:limit]

	return stats
}
