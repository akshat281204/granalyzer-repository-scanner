package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var javaImportPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*import\s+([a-zA-Z0-9_\.\*]+);`),
	regexp.MustCompile(`^\s*package\s+([a-zA-Z0-9_\.]+);`),
}

var javaRoutePatterns = []struct {
	regex     *regexp.Regexp
	framework string
}{
	{regexp.MustCompile(`@(Get|Post|Put|Delete|Patch)Mapping\(\s*["']([^"']+)["']`), "spring"},
	{regexp.MustCompile(`@RequestMapping\(\s*(?:value\s*=\s*)?["']([^"']+)["']`), "spring"},
}

func ParseJava(path string, relPath string, framework string) ParseResult {
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
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// 1. Match imports/package
		for _, pat := range javaImportPatterns {
			matches := pat.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				res.Imports = append(res.Imports, matches[1])
			}
		}

		// Infer framework if not set
		if framework == "" {
			if strings.Contains(trimmed, "org.springframework") {
				framework = "spring"
			}
		}

		// 2. Match routes
		for _, pattern := range javaRoutePatterns {
			matches := pattern.regex.FindStringSubmatch(trimmed)
			if len(matches) > 0 {
				methodVal := "ANY"
				pathVal := ""

				if len(matches) == 3 {
					// e.g. @GetMapping
					methodVal = strings.ToUpper(matches[1])
					pathVal = matches[2]
				} else if len(matches) == 2 {
					// e.g. @RequestMapping
					pathVal = matches[1]
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
