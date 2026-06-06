package parser

func ParseGeneric(path string, relPath string, framework string) ParseResult {
	// Fallback returns empty arrays for endpoints and imports
	return ParseResult{
		Endpoints: []Endpoint{},
		Imports:   []string{},
	}
}
