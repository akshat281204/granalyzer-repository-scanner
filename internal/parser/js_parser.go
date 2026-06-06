package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var jsImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]`),
	regexp.MustCompile(`import\s+['"]([^'"]+)['"]`),
	regexp.MustCompile(`require\(\s*['"]([^'"]+)['"]\s*\)`),
}

var jsRoutePatterns = []struct {
	regex     *regexp.Regexp
	framework string
}{
	{regexp.MustCompile(`(?:app|router|server|route|r)\.(get|post|put|delete|patch|options|use)\(\s*['"]([^'"]+)['"]`), "express"},
	{regexp.MustCompile(`fastify\.(get|post|put|delete|patch)\(\s*['"]([^'"]+)['"]`), "fastify"},
	{regexp.MustCompile(`app\.(get|post|put|delete|patch)\(\s*['"]([^'"]+)['"]`), "hono"},
}

func ParseJS(path string, relPath string, framework string) ParseResult {
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

		// Clean comments for simpler matching
		cleanLine := stripComments(line)
		if cleanLine == "" {
			continue
		}

		// 1. Match imports
		for _, pat := range jsImportPatterns {
			matches := pat.FindStringSubmatch(cleanLine)
			if len(matches) > 1 {
				res.Imports = append(res.Imports, matches[1])
			}
		}

		// Infer framework if not set
		if framework == "" {
			if strings.Contains(cleanLine, "express") {
				framework = "express"
			} else if strings.Contains(cleanLine, "fastify") {
				framework = "fastify"
			} else if strings.Contains(cleanLine, "hono") {
				framework = "hono"
			}
		}

		// 2. Match routes
		for _, pattern := range jsRoutePatterns {
			matches := pattern.regex.FindStringSubmatch(cleanLine)
			if len(matches) > 2 {
				fw := framework
				if fw == "" {
					fw = pattern.framework
				}
				res.Endpoints = append(res.Endpoints, Endpoint{
					Method:    strings.ToUpper(matches[1]),
					Path:      matches[2],
					File:      relPath,
					Line:      lineNum,
					Framework: fw,
				})
			}
		}
	}

	return res
}

func stripComments(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "//") {
		return ""
	}
	// Note: Block comments across lines aren't fully handled by line-by-line,
	// but this is standard and efficient for regex scanners.
	return line
}
