package api

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsRegisteredAPIRoutes(t *testing.T) {
	operations := readOpenAPIOperations(t)
	routes := map[string]map[string]bool{}

	for _, route := range apiRouteSpecs() {
		if route.Path == "/health" {
			continue
		}
		path := strings.TrimPrefix(route.Path, "/api/v1")
		if routes[path] == nil {
			routes[path] = map[string]bool{}
		}
		routes[path][route.Method] = true

		methods := operations[path]
		if !methods[route.Method] {
			t.Fatalf("OpenAPI contract is missing %s %s", route.Method, path)
		}
	}

	for path, methods := range operations {
		for method := range methods {
			if !routes[path][method] {
				t.Fatalf("OpenAPI contract documents unregistered route %s %s", method, path)
			}
		}
	}
}

func TestFrontendAPIAdapterUsesDocumentedRoutes(t *testing.T) {
	operations := readOpenAPIOperations(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "frontend", "src", "shared", "api.ts"))
	if err != nil {
		t.Fatal(err)
	}

	for _, rawPath := range frontendAPIPaths(string(data)) {
		path := strings.TrimPrefix(rawPath, "/api/v1")
		path = strings.Split(path, "?")[0]
		if path == "" {
			continue
		}
		if !openAPIPathExists(operations, path) {
			t.Fatalf("frontend API adapter uses undocumented path %s", rawPath)
		}
	}
}

func readOpenAPIOperations(t *testing.T) map[string]map[string]bool {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	operations := map[string]map[string]bool{}
	currentPath := ""
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "  /") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			currentPath = strings.TrimSuffix(strings.TrimSpace(line), ":")
			if operations[currentPath] == nil {
				operations[currentPath] = map[string]bool{}
			}
			continue
		}
		if currentPath == "" || !strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "      ") {
			continue
		}
		method := strings.TrimSuffix(strings.TrimSpace(line), ":")
		switch method {
		case "get", "post", "put", "delete", "patch":
			operations[currentPath][strings.ToUpper(method)] = true
		}
	}
	return operations
}

func frontendAPIPaths(source string) []string {
	seen := map[string]bool{}
	paths := []string{}
	for _, path := range extractRequestPaths(source) {
		normalized := normalizeFrontendPath(path)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		paths = append(paths, normalized)
	}

	apiV1Literal := regexp.MustCompile("`/api/v1[^`]+`|\"/api/v1[^\"]+\"")
	for _, match := range apiV1Literal.FindAllString(source, -1) {
		normalized := normalizeFrontendPath(strings.Trim(match, "`\""))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		paths = append(paths, normalized)
	}
	return paths
}

func extractRequestPaths(source string) []string {
	paths := []string{}
	searchFrom := 0
	for {
		index := strings.Index(source[searchFrom:], "request<")
		if index < 0 {
			break
		}
		index += searchFrom
		argumentStart := strings.Index(source[index:], "(")
		if argumentStart < 0 {
			break
		}
		cursor := index + argumentStart + 1
		for cursor < len(source) && (source[cursor] == ' ' || source[cursor] == '\n' || source[cursor] == '\r' || source[cursor] == '\t') {
			cursor++
		}
		if cursor >= len(source) {
			break
		}
		quote := source[cursor]
		if quote != '"' && quote != '`' {
			searchFrom = cursor + 1
			continue
		}
		path, next := readQuotedPath(source, cursor, quote)
		if strings.HasPrefix(path, "/") {
			paths = append(paths, path)
		}
		searchFrom = next
	}
	return paths
}

func readQuotedPath(source string, start int, quote byte) (string, int) {
	var builder strings.Builder
	for cursor := start + 1; cursor < len(source); cursor++ {
		if source[cursor] == quote {
			return builder.String(), cursor + 1
		}
		if source[cursor] == '\\' && cursor+1 < len(source) {
			cursor++
		}
		builder.WriteByte(source[cursor])
	}
	return builder.String(), len(source)
}

func normalizeFrontendPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "/api/v1")
	path = stripTemplateExpressions(path)
	path = strings.Split(path, "?")[0]
	path = strings.TrimRight(path, "/")
	if path == "" {
		return ""
	}
	return path
}

func stripTemplateExpressions(path string) string {
	var builder strings.Builder
	for cursor := 0; cursor < len(path); cursor++ {
		if cursor+1 < len(path) && path[cursor] == '$' && path[cursor+1] == '{' {
			if strings.HasSuffix(builder.String(), "/") {
				builder.WriteString("{}")
			}
			cursor += 2
			depth := 1
			for cursor < len(path) && depth > 0 {
				if path[cursor] == '{' {
					depth++
				}
				if path[cursor] == '}' {
					depth--
				}
				cursor++
			}
			cursor--
			continue
		}
		builder.WriteByte(path[cursor])
	}
	return builder.String()
}

func openAPIPathExists(operations map[string]map[string]bool, frontendPath string) bool {
	for contractPath := range operations {
		if pathShapeMatches(contractPath, frontendPath) {
			return true
		}
	}
	return false
}

func pathShapeMatches(contractPath, frontendPath string) bool {
	contractParts := strings.Split(strings.Trim(contractPath, "/"), "/")
	frontendParts := strings.Split(strings.Trim(frontendPath, "/"), "/")
	if len(contractParts) != len(frontendParts) {
		return false
	}
	for index := range contractParts {
		contractDynamic := strings.HasPrefix(contractParts[index], "{") && strings.HasSuffix(contractParts[index], "}")
		frontendDynamic := frontendParts[index] == "{}"
		if contractDynamic || frontendDynamic {
			continue
		}
		if contractParts[index] != frontendParts[index] {
			return false
		}
	}
	return true
}
