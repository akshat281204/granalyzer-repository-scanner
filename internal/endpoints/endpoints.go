package endpoints

import (
	"sort"

	"granalyzer/internal/parser"
	"granalyzer/internal/scanner"
)

func Extract(result *scanner.ScanResult) []parser.Endpoint {
	var list []parser.Endpoint

	for _, file := range result.Files {
		res := parser.ParseFile(file.Path, file.RelPath, file.Language, file.Framework)
		list = append(list, res.Endpoints...)
	}

	// Deduplicate exact duplicates if any
	seen := make(map[string]bool)
	var deduped []parser.Endpoint

	for _, ep := range list {
		// Key on method, path, file, line
		key := ep.Method + ":" + ep.Path + ":" + ep.File
		if !seen[key] {
			seen[key] = true
			deduped = append(deduped, ep)
		}
	}

	// Sort by Path, then Method, then File/Line
	sort.Slice(deduped, func(i, j int) bool {
		if deduped[i].Path != deduped[j].Path {
			return deduped[i].Path < deduped[j].Path
		}
		if deduped[i].Method != deduped[j].Method {
			return deduped[i].Method < deduped[j].Method
		}
		return deduped[i].File < deduped[j].File
	})

	return deduped
}
