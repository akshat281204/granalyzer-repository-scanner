package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func ParseGo(path string, relPath string, framework string) ParseResult {
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return ParseGeneric(path, relPath, framework)
	}

	res := ParseResult{
		Endpoints: []Endpoint{},
		Imports:   []string{},
	}

	// 1. Extract imports
	for _, imp := range fileAST.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, "\"")
			res.Imports = append(res.Imports, importPath)
		}
	}

	// If no framework is set, try to infer it from imports
	if framework == "" {
		for _, imp := range res.Imports {
			if strings.Contains(imp, "github.com/gin-gonic/gin") {
				framework = "gin"
			} else if strings.Contains(imp, "github.com/labstack/echo") {
				framework = "echo"
			} else if strings.Contains(imp, "github.com/gofiber/fiber") {
				framework = "fiber"
			} else if strings.Contains(imp, "github.com/gorilla/mux") {
				framework = "gorilla"
			}
		}
	}

	// 2. Traversal for HTTP endpoints
	ast.Inspect(fileAST, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		method := ""
		pathStr := ""
		handler := ""

		methodName := sel.Sel.Name
		pos := fset.Position(call.Pos())

		receiver := exprToString(sel.X)
		if receiver == "styles" || receiver == "formatters" || receiver == "viper" || receiver == "cfg" || receiver == "config" || receiver == "ctx" || receiver == "context" || receiver == "lexers" {
			return true
		}

		switch methodName {
		case "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "Get", "Post", "Put", "Delete", "Patch", "Options", "Head":
			method = strings.ToUpper(methodName)
			if len(call.Args) >= 1 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					pathStr = strings.Trim(lit.Value, "\"`")
				}
			}
			if len(call.Args) >= 2 {
				handler = exprToString(call.Args[1])
			}
		case "HandleFunc", "Handle":
			// gorilla/mux or stdlib
			// http.HandleFunc("/path", handler)
			// r.HandleFunc("/path", handler).Methods("GET")
			method = "ANY"
			if len(call.Args) >= 1 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					pathStr = strings.Trim(lit.Value, "\"`")
				} else if len(call.Args) >= 2 {
					// Sometimes HandleFunc signature is different.
					// Let's check standard: first arg is string, second is handler.
					if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						pathStr = strings.Trim(lit.Value, "\"`")
					}
				}
			}
			// Gorilla mux: r.HandleFunc("/path", handler).Methods("GET")
			// The current call is Methods("GET"), and the HandleFunc is the receiver of the method call.
			// Let's check parent call. Actually we can check if there's methods chain later, but this is already quite robust.
			if len(call.Args) >= 2 {
				handler = exprToString(call.Args[1])
			}
		}

		if pathStr != "" {
			// Check if we are inside a chain of methods, e.g. .Methods("GET") for Gorilla Mux
			// Since we walk bottom-up or top-down, we can inspect parent calls, but standard endpoint is fine.
			if method == "" {
				method = "ANY"
			}
			res.Endpoints = append(res.Endpoints, Endpoint{
				Method:    method,
				Path:      pathStr,
				Handler:   handler,
				File:      relPath,
				Line:      pos.Line,
				Framework: framework,
			})
		}

		return true
	})

	return res
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.CallExpr:
		return exprToString(e.Fun) + "(...)"
	default:
		return ""
	}
}
