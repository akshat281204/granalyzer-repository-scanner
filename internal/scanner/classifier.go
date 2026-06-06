package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var extToLang = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".ts":    "TypeScript",
	".jsx":   "JavaScript",
	".tsx":   "TypeScript",
	".py":    "Python",
	".rs":    "Rust",
	".java":  "Java",
	".rb":    "Ruby",
	".php":   "PHP",
	".cs":    "C#",
	".cpp":   "C++",
	".c":     "C",
	".swift": "Swift",
	".kt":    "Kotlin",
	".html":  "HTML",
	".css":   "CSS",
	".scss":  "SCSS",
	".sql":   "SQL",
	".sh":    "Shell",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".toml":  "TOML",
	".md":    "Markdown",
}

func ClassifyFile(path string) (string, string) {
	ext := strings.ToLower(filepath.Ext(path))
	lang, ok := extToLang[ext]
	if !ok {
		return "Other", ""
	}

	framework := detectFramework(path, lang)
	return lang, framework
}

func detectFramework(path string, lang string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	linesRead := 0
	const maxPeekLines = 50

	for scanner.Scan() && linesRead < maxPeekLines {
		line := scanner.Text()
		linesRead++

		switch lang {
		case "Go":
			if strings.Contains(line, "github.com/gin-gonic/gin") {
				return "gin"
			}
			if strings.Contains(line, "github.com/labstack/echo") {
				return "echo"
			}
			if strings.Contains(line, "github.com/gorilla/mux") {
				return "gorilla"
			}
			if strings.Contains(line, "github.com/gofiber/fiber") {
				return "fiber"
			}
		case "JavaScript", "TypeScript":
			if strings.Contains(line, "express") {
				return "express"
			}
			if strings.Contains(line, "fastify") {
				return "fastify"
			}
			if strings.Contains(line, "hono") {
				return "hono"
			}
		case "Python":
			if strings.Contains(line, "fastapi") {
				return "fastapi"
			}
			if strings.Contains(line, "flask") {
				return "flask"
			}
			if strings.Contains(line, "django") {
				return "django"
			}
		case "Rust":
			if strings.Contains(line, "actix_web") || strings.Contains(line, "actix-web") {
				return "actix"
			}
			if strings.Contains(line, "axum") {
				return "axum"
			}
			if strings.Contains(line, "rocket") {
				return "rocket"
			}
		case "Java":
			if strings.Contains(line, "org.springframework") {
				return "spring"
			}
		}
	}

	return ""
}
