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
	files, err := filepath.Glob(filepath.Join("..", "..", "..", "frontend", "src", "shared", "api*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, ".test.ts") {
			continue
		}
		data, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, operation := range frontendAPIOperations(string(data)) {
			if operation.Path != "" && !openAPIOperationExists(operations, operation) {
				t.Fatalf("%s uses undocumented operation %s %s", filepath.Base(file), operation.Method, operation.Path)
			}
		}
	}
}

func TestOpenAPIDocumentsLayoutAndAccessorySchemas(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	want := []string{
		"Layout", "LayoutInput", "UpdateLayoutInput", "LayoutUnit", "LayoutUnitInput",
		"UpdateLayoutUnitInput", "LayoutConfiguration", "LayoutConfigurationInput",
		"UpdateLayoutConfigurationInput", "PlanVariant", "PlanVariantInput", "PlanRevision",
		"PlanRevisionInput", "PlanRevisionTransitionInput", "AccessoryProduct", "AccessoryProductInput",
		"StorageLocation", "StorageLocationInput", "AccessoryStockSummary", "AccessoryStockAdjustmentInput",
		"AccessoryAsset", "AccessoryAssetInput", "AccessoryAllocationTarget", "AccessoryReservation",
		"AccessoryReservationInput", "AccessoryInstallation", "AccessoryInstallationInput",
		"AccessoryInstallationRemovalInput", "AccessoryInstallationConditionInput", "AccessoryAllocationSummary",
		"AccessoryArticleListItem", "AccessoryOverviewMetrics", "AccessoryArticleFilterOptions",
		"AccessoryArticleListResult", "AccessoryAttributeValue", "AccessoryTextAttribute",
		"AccessoryNumberAttribute", "AccessoryBooleanAttribute", "AccessoryDateAttribute",
		"AccessorySingleSelectAttribute", "AccessoryMultiSelectAttribute", "AccessoryDuplicateCheckInput",
		"AccessoryDuplicateCandidate", "AccessoryDuplicateCheckResult", "AccessoryStockMovement",
		"AccessoryStockTransferInput", "AccessoryPurchase", "AccessoryPurchaseInput",
		"AccessoryIndividualizationInput", "AccessoryDocument", "AccessoryDocumentUpdateInput",
		"AccessoryUsageEvent", "AccessoryUsageHistory",
	}
	for _, schema := range want {
		if !strings.Contains(contract, "    "+schema+":\n") {
			t.Errorf("OpenAPI contract is missing schema %s", schema)
		}
	}
}

func TestOpenAPIDocumentsCompleteAccessoryArticleHTTPContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	fragments := []string{
		"name: articleType\n", "name: gauge\n", "name: status\n", "style: form\n", "explode: true\n",
		"enum: [available, reserved, installed, maintenance_due, defective, archived]",
		"enum: [article, type, gauge, stock, storage, updatedAt]", "enum: [asc, desc]",
		"multipart/form-data:", "format: binary", "application/octet-stream:",
		"AccessoryArticleListResult", "AccessoryDuplicateCheckResult", "AccessoryStockSummary",
		"AccessoryStockMovement", "AccessoryPurchase", "AccessoryDocument", "AccessoryUsageHistory",
		"discriminator:", "propertyName: kind", `"400":`, `"403":`, `"404":`, `"409":`, `"413":`, `"415":`,
	}
	for _, fragment := range fragments {
		if !strings.Contains(contract, fragment) {
			t.Errorf("OpenAPI article contract is missing %q", fragment)
		}
	}
}

func TestOpenAPIArticleSchemasMatchRuntimeSemantics(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)

	usageEvent := openAPIIndentedBlock(t, contract, "AccessoryUsageEvent", 4)
	if !strings.Contains(usageEvent,
		"enum: [reservation, installation, condition_changed, removal]") {
		t.Errorf("usage event enum does not match runtime: %s", usageEvent)
	}

	asset := openAPIIndentedBlock(t, contract, "AccessoryAsset", 4)
	if !strings.Contains(asset, "purchaseId:") {
		t.Errorf("AccessoryAsset schema is missing purchaseId: %s", asset)
	}

	document := openAPIIndentedBlock(t, contract, "AccessoryDocument", 4)
	if strings.Contains(document, "fileBlobId:") {
		t.Errorf("AccessoryDocument exposes internal fileBlobId: %s", document)
	}

	download := openAPIIndentedBlock(
		t, contract, "/accessory-products/{id}/documents/{documentID}/download", 2,
	)
	successStart := strings.Index(download, `        "200":`)
	successEnd := strings.Index(download, `        "403":`)
	if successStart < 0 || successEnd <= successStart {
		t.Fatalf("document download success response is malformed: %s", download)
	}
	downloadSuccess := download[successStart:successEnd]
	for _, contentType := range []string{"application/pdf:", "application/json:", "text/plain:", "image/png:"} {
		if !strings.Contains(downloadSuccess, contentType) {
			t.Errorf("document download contract is missing %s", contentType)
		}
	}
}

func openAPIIndentedBlock(t *testing.T, contract, heading string, indent int) string {
	t.Helper()
	lines := strings.Split(contract, "\n")
	prefix := strings.Repeat(" ", indent) + heading + ":"
	start := -1
	for index, line := range lines {
		if line == prefix {
			start = index
			break
		}
	}
	if start < 0 {
		t.Fatalf("OpenAPI block %s is missing", heading)
	}
	end := len(lines)
	for index := start + 1; index < len(lines); index++ {
		line := lines[index]
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineIndent := len(line) - len(strings.TrimLeft(line, " "))
		if lineIndent <= indent {
			end = index
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

type frontendAPIOperation struct {
	Method string
	Path   string
}

func TestFrontendAPIOperationsIncludeHTTPMethods(t *testing.T) {
	source := `
const api = {
  current: () => request<UserSession>("/auth/session"),
  update: () => request<UserSession>("/auth/password", { method: "PUT" }),
  dynamic: (id: string) => request<void>(` + "`/sessions/${encodeURIComponent(id)}/revoke`" + `, {
    method: "PUT"
  }),
  helper: () => request<void>("/layouts", json("POST", input))
};
`

	operations := frontendAPIOperations(source)

	expected := []frontendAPIOperation{
		{Method: "GET", Path: "/auth/session"},
		{Method: "PUT", Path: "/auth/password"},
		{Method: "PUT", Path: "/sessions/{}/revoke"},
		{Method: "POST", Path: "/layouts"},
	}
	if len(operations) != len(expected) {
		t.Fatalf("expected %#v, got %#v", expected, operations)
	}
	for index, operation := range expected {
		if operations[index] != operation {
			t.Fatalf("expected operation %d to be %#v, got %#v", index, operation, operations[index])
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

func frontendAPIOperations(source string) []frontendAPIOperation {
	seen := map[string]bool{}
	operations := []frontendAPIOperation{}
	for _, operation := range extractRequestOperations(source) {
		path := operation.Path
		normalized := normalizeFrontendPath(path)
		if normalized == "" {
			continue
		}
		operation.Path = normalized
		operation.Method = normalizeHTTPMethod(operation.Method)
		key := operation.Method + " " + operation.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		operations = append(operations, operation)
	}

	apiV1Literal := regexp.MustCompile("`/api/v1[^`]+`|\"/api/v1[^\"]+\"")
	for _, match := range apiV1Literal.FindAllString(source, -1) {
		normalized := normalizeFrontendPath(strings.Trim(match, "`\""))
		if normalized == "" {
			continue
		}
		operation := frontendAPIOperation{Method: "GET", Path: normalized}
		key := operation.Method + " " + operation.Path
		if seen[key] {
			continue
		}
		seen[key] = true
		operations = append(operations, operation)
	}
	return operations
}

func extractRequestOperations(source string) []frontendAPIOperation {
	methodPattern := regexp.MustCompile(`method:\s*["'](GET|POST|PUT|DELETE|PATCH)["']`)
	jsonHelperPattern := regexp.MustCompile(`json\(\s*["'](POST|PUT|PATCH)["']`)
	paths := []frontendAPIOperation{}
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
			method := "GET"
			if callEnd := findCallEnd(source, index+argumentStart); callEnd > next {
				call := source[next:callEnd]
				if match := methodPattern.FindStringSubmatch(call); len(match) == 2 {
					method = match[1]
				} else if match := jsonHelperPattern.FindStringSubmatch(call); len(match) == 2 {
					method = match[1]
				}
			}
			paths = append(paths, frontendAPIOperation{Method: method, Path: path})
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

func findCallEnd(source string, openParen int) int {
	depth := 0
	quote := byte(0)
	for cursor := openParen; cursor < len(source); cursor++ {
		current := source[cursor]
		if quote != 0 {
			if current == '\\' {
				cursor++
				continue
			}
			if current == quote {
				quote = 0
			}
			continue
		}
		if current == '"' || current == '\'' || current == '`' {
			quote = current
			continue
		}
		switch current {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return cursor
			}
		}
	}
	return len(source)
}

func normalizeHTTPMethod(method string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		return "GET"
	}
	return method
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

func openAPIOperationExists(operations map[string]map[string]bool, operation frontendAPIOperation) bool {
	for contractPath := range operations {
		if pathShapeMatches(contractPath, operation.Path) {
			return operations[contractPath][operation.Method]
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
		if contractDynamic && frontendDynamic {
			continue
		}
		if contractDynamic || frontendDynamic {
			return false
		}
		if contractParts[index] != frontendParts[index] {
			return false
		}
	}
	return true
}
