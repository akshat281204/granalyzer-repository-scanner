package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var rustImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*use\s+([a-zA-Z0-9_::\{\},\s\*]+);`),
	regexp.MustCompile(`mod\s+([a-zA-Z0-9_]+);`),
}

var rustRoutePatterns = []struct {
	regex     *regexp.Regexp
	framework string
	method    string
}{
	// #[get("/path")] -> Actix / Rocket
	{regexp.MustCompile(`#\[(get|post|put|delete|patch|options|head)\(\s*['"]([^'"]+)['"]\s*\)\]`), "actix/rocket", ""},
	// .route("/path", get(handler)) -> Axum
	{regexp.MustCompile(`\.route\(\s*['"]([^'"]+)['"]\s*,\s*(get|post|put|delete|patch|options|head)`), "axum", ""},
}

func ParseRust(path string, relPath string, framework string) ParseResult {
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
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// 1. Match imports
		for _, pat := range rustImportPatterns {
			matches := pat.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				res.Imports = append(res.Imports, matches[1])
			}
		}

		// Infer framework if not set
		if framework == "" {
			if strings.Contains(trimmed, "axum") {
				framework = "axum"
			} else if strings.Contains(trimmed, "actix_web") || strings.Contains(trimmed, "actix-web") {
				framework = "actix"
			} else if strings.Contains(trimmed, "rocket") {
				framework = "rocket"
			}
		}

		// 2. Match routes
		for _, pattern := range rustRoutePatterns {
			matches := pattern.regex.FindStringSubmatch(trimmed)
			if len(matches) > 0 {
				pathVal := ""
				methodVal := ""

				if pattern.framework == "actix/rocket" {
					methodVal = strings.ToUpper(matches[1])
					pathVal = matches[2]
				} else if pattern.framework == "axum" {
					pathVal = matches[1]
					methodVal = strings.ToUpper(matches[2])
				}

				if pathVal != "" {
					fw := framework
					if fw == "" {
						if pattern.framework == "actix/rocket" {
							fw = "actix" // default guess
						} else {
							fw = pattern.framework
						}
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
