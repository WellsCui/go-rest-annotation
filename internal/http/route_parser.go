package http

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type RouteMetadata struct {
	Operation     *RestOperation
	HandlerMethod string
	HandlerType   string
	PackagePath   string
}

// ParseRoutes parses a Go source file and extracts route metadata from @RestOperation annotations.
func ParseRoutes(filePath string) ([]*RouteMetadata, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	var routes []*RouteMetadata
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Doc == nil {
			continue
		}
		for _, comment := range fn.Doc.List {
			if !strings.Contains(comment.Text, "@RestOperation") {
				continue
			}
			op, err := ParseRestOperation(comment.Text)
			if err != nil {
				return nil, fmt.Errorf("failed to parse annotation on %s: %w", fn.Name.Name, err)
			}
			metadata := &RouteMetadata{
				Operation:     op,
				HandlerMethod: fn.Name.Name,
				PackagePath:   file.Name.Name,
			}
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				metadata.HandlerType = extractReceiverType(fn.Recv.List[0])
			}
			routes = append(routes, metadata)
		}
	}
	return routes, nil
}

func extractReceiverType(field *ast.Field) string {
	switch t := field.Type.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}
