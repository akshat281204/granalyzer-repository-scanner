package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var pyImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*import\s+([a-zA-Z0-9_\.,\s]+)`),
	regexp.MustCompile(`^\s*from\s+([a-zA-Z0-9_\.]+)\s+import`),
}

var pyRoutePatterns = []struct {
	regex     *regexp.Regexp
	framework string
	methodIdx int
	pathIdx   int
}{
	// @app.route('/path', methods=['GET']) -> Flask
	{regexp.MustCompile(`@(?:app|bp|blueprint)\.route\(\s*['"]([^'"]+)['"](?:\s*,\s*methods\s*=\s*\[\s*['"]([a-zA-Z]+)['"]\s*\])?`), "flask", 2, 1},
	// @app.get('/path') or @router.post('/path') -> FastAPI / Flask
	{regexp.MustCompile(`@(?:app|router|bp|blueprint)\.(get|post|put|delete|patch)\(\s*['"]([^'"]+)['"]`), "fastapi", 1, 2},
	// path('endpoint/', view) -> Django
	{regexp.MustCompile(`path\(\s*['"]([^'"]+)['"]\s*,\s*([a-zA-Z0-9_\.]+)`), "django", -1, 1},
}

func ParsePython(path string, relPath string, framework string) ParseResult {
	file, err := os.Open(path)
	if err != nil {
		return ParseResult{}
	}
	defer file.Close()

	res := ParseResult{
		Endpoints: []Endpoint{},
		Imports:   []string{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 1. Match imports
		for _, pat := range pyImportPatterns {
			matches := pat.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				// Handle multiple imports on a single line, e.g. "import os, sys"
				parts := strings.Split(matches[1], ",")
				for _, p := range parts {
					res.Imports = append(res.Imports, strings.TrimSpace(p))
				}
			}
		}

		// Infer framework if not set
		if framework == "" {
			if strings.Contains(trimmed, "fastapi") {
				framework = "fastapi"
			} else if strings.Contains(trimmed, "flask") {
				framework = "flask"
			} else if strings.Contains(trimmed, "django") {
				framework = "django"
			}
		}

		// 2. Match routes
		for _, pattern := range pyRoutePatterns {
			matches := pattern.regex.FindStringSubmatch(trimmed)
			if len(matches) > 0 {
				pathVal := ""
				methodVal := "ANY"

				// Extract path
				if pattern.pathIdx < len(matches) {
					pathVal = matches[pattern.pathIdx]
				}

				// Extract method if any
				if pattern.methodIdx != -1 && pattern.methodIdx < len(matches) && matches[pattern.methodIdx] != "" {
					methodVal = strings.ToUpper(matches[pattern.methodIdx])
				}

				if pathVal != "" {
					fw := framework
					if fw == "" {
						fw = pattern.framework
					}
					res.Endpoints = append(res.Endpoints, Endpoint{
						Method:    methodVal,
						Path:      pathVal,
						File:      relPath,
						Line:      lineNum,
						Framework: fw,
					})
				}
			}
		}
	}

	return res
}
