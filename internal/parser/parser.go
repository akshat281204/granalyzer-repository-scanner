package parser

type Endpoint struct {
	Method    string // GET, POST, PUT, DELETE, PATCH, ANY, WS
	Path      string // e.g. /api/users/:id
	Handler   string // handler function name if detectable
	File      string // relative file path
	Line      int
	Framework string // gin, express, fastapi, spring, etc.
}

type ParseResult struct {
	Endpoints []Endpoint
	Imports   []string // Imported package paths or file paths
}

// Dispatch parses the file at the given absolute path for the given language and framework.
// It returns endpoints and imports.
func ParseFile(path string, relPath string, lang string, framework string) ParseResult {
	switch lang {
	case "Go":
		return ParseGo(path, relPath, framework)
	case "JavaScript", "TypeScript":
		return ParseJS(path, relPath, framework)
	case "Python":
		return ParsePython(path, relPath, framework)
	case "Rust":
		return ParseRust(path, relPath, framework)
	case "Java":
		return ParseJava(path, relPath, framework)
	default:
		return ParseGeneric(path, relPath, framework)
	}
}
