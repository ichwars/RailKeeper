package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
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

func TestOpenAPIMasterDataActiveBatchContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	operation := openAPIIndentedBlock(t, contract, "/master-data/{type}/active", 2)
	for _, expected := range []string{
		"patch:", "MasterDataActiveBatchInput", `"200":`, `"400":`, `"404":`,
	} {
		if !strings.Contains(operation, expected) {
			t.Errorf("master-data batch operation is missing %q: %s", expected, operation)
		}
	}
	schema := openAPIIndentedBlock(t, contract, "MasterDataActiveBatchInput", 4)
	for _, expected := range []string{
		"required: [keys, active]", "minItems: 1", "maxItems: 5000", "uniqueItems: true", "enum: [false]",
	} {
		if !strings.Contains(schema, expected) {
			t.Errorf("master-data batch schema is missing %q: %s", expected, schema)
		}
	}
}

func TestOpenAPIDataTransferPreviewEnumsMatchRuntimeValues(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	actionSchema := openAPIIndentedBlock(t, contract, "TransferProposedAction", 4)
	proposalSchema := openAPIIndentedBlock(t, contract, "TransferProposedResolution", 4)
	selectionSchema := openAPIIndentedBlock(t, contract, "TransferIssueResolution", 4)

	for _, action := range []string{"create", "replace", "use_existing", "copy"} {
		payload, err := json.Marshal(application.DataTransferPreviewRecord{ProposedAction: action})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(payload), `"proposedAction":"`+action+`"`) ||
			!openAPIEnumContains(actionSchema, action) {
			t.Errorf("runtime proposed action %q is missing from OpenAPI: %s", action, actionSchema)
		}
	}
	issuePayload, err := json.Marshal(application.DataTransferIssue{ProposedResolution: "replace_or_copy"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(issuePayload), `"proposedResolution":"replace_or_copy"`) ||
		!openAPIEnumContains(proposalSchema, "replace_or_copy") {
		t.Errorf("runtime proposal replace_or_copy is missing from OpenAPI: %s", proposalSchema)
	}
	if openAPIEnumContains(selectionSchema, "replace_or_copy") {
		t.Errorf("non-selectable suggestion leaked into resolution input: %s", selectionSchema)
	}
	for _, resolution := range []string{"replace", "copy", "skip", "use_existing", "create", "link"} {
		if !openAPIEnumContains(selectionSchema, resolution) {
			t.Errorf("selectable resolution %q is missing: %s", resolution, selectionSchema)
		}
	}
}

func openAPIEnumContains(schema, value string) bool {
	marker := "enum: ["
	start := strings.Index(schema, marker)
	if start < 0 {
		return false
	}
	start += len(marker)
	end := strings.Index(schema[start:], "]")
	if end < 0 {
		return false
	}
	for _, candidate := range strings.Split(schema[start:start+end], ",") {
		if strings.TrimSpace(candidate) == value {
			return true
		}
	}
	return false
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
		"UpdateLayoutUnitInput", "LayoutUnitPort", "LayoutUnitPortInput", "UpdateLayoutUnitPortInput",
		"LayoutTechnicalPosition", "LayoutTechnicalPositionInput",
		"UpdateLayoutTechnicalPositionInput", "LayoutConfiguration", "LayoutConfigurationInput",
		"UpdateLayoutConfigurationInput", "PlanVariant", "PlanVariantInput", "PlanRevision",
		"ModulePortConnection", "ModulePortIssue", "ModulePortAnalysis", "ConfigurationUnitSnapPreviewInput",
		"ModulePortSnapResult",
		"PlanRevisionInput", "PlanRevisionTransitionInput", "TrackPoint", "TrackPort", "TrackRoute",
		"TrackGeometry", "TrackGeometryDefinition", "PlanTrackObject", "TrackPlan",
		"TrackPlanConnection", "TrackPlanIssue", "TrackBOMLine", "TrackGrade", "TrackMaterialStatus", "TrackPlanAnalysis",
		"CreatePlanTrackObjectInput", "UpdatePlanTrackObjectInput", "AccessoryProduct", "AccessoryProductInput",
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

func TestOpenAPIDocumentsTrackElevationMismatch(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	issue := openAPIIndentedBlock(t, contract, "TrackPlanIssue", 4)
	if !strings.Contains(issue,
		"enum: [open_end, incompatible_connection, overlap, broken_geometry, elevation_mismatch, grade_limit_exceeded, insufficient_clearance, flex_radius_below_limit]") {
		t.Errorf("TrackPlanIssue is missing elevation_mismatch: %s", issue)
	}
	if !strings.Contains(issue, "elevationDifferenceMm:") || !strings.Contains(issue, "minimum: 0") {
		t.Errorf("TrackPlanIssue is missing elevation difference details: %s", issue)
	}
	change := openAPIIndentedBlock(t, contract, "TrackPlanIssueChange", 4)
	if !strings.Contains(change,
		"enum: [open_end, incompatible_connection, overlap, broken_geometry, elevation_mismatch, grade_limit_exceeded, insufficient_clearance, flex_radius_below_limit]") {
		t.Errorf("TrackPlanIssueChange is missing elevation_mismatch: %s", change)
	}
}

func TestOpenAPIDocumentsFlexTrackPreview(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{"FlexTrackPath", "FlexTrackPreviewInput", "FlexTrackPreview"} {
		if !strings.Contains(contract, "    "+schema+":\n") {
			t.Errorf("OpenAPI contract is missing schema %s", schema)
		}
	}
	issue := openAPIIndentedBlock(t, contract, "TrackPlanIssue", 4)
	for _, field := range []string{"flex_radius_below_limit", "radiusMm:", "radiusLimitMm:"} {
		if !strings.Contains(issue, field) {
			t.Errorf("TrackPlanIssue is missing %s: %s", field, issue)
		}
	}
}

func TestOpenAPIDocumentsTransitionCurvePreview(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{
		"TransitionCurvePath", "TransitionCurvePreviewInput", "TransitionCurvePreview",
	} {
		if !strings.Contains(contract, "    "+schema+":\n") {
			t.Errorf("OpenAPI contract is missing schema %s", schema)
		}
	}
	for _, schema := range []string{"PlanTrackObject", "CreatePlanTrackObjectInput", "UpdatePlanTrackObjectInput"} {
		block := openAPIIndentedBlock(t, contract, schema, 4)
		if !strings.Contains(block, "transitionPath:") {
			t.Errorf("%s is missing transitionPath: %s", schema, block)
		}
	}
	if !strings.Contains(contract, "  /plan-track-objects/{id}/transition-preview:\n") {
		t.Error("OpenAPI contract is missing transition preview path")
	}
}

func TestOpenAPIDocumentsFreePlanObjects(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{
		"FreePlanObjectShape", "PlanFreeObject", "CreateFreePlanObjectInput",
		"UpdateFreePlanObjectInput", "PlanFreeObjectChange",
	} {
		if !strings.Contains(contract, "    "+schema+":\n") {
			t.Errorf("OpenAPI contract is missing schema %s", schema)
		}
	}
	for _, path := range []string{
		"  /plan-revisions/{id}/free-objects:\n", "  /plan-free-objects/{id}:\n",
	} {
		if !strings.Contains(contract, path) {
			t.Errorf("OpenAPI contract is missing path %s", strings.TrimSpace(path))
		}
	}
	plan := openAPIIndentedBlock(t, contract, "TrackPlan", 4)
	if !strings.Contains(plan, "freeObjects:") {
		t.Errorf("TrackPlan is missing freeObjects: %s", plan)
	}
	preview := openAPIIndentedBlock(t, contract, "TrackPlanChangePreview", 4)
	if !strings.Contains(preview, "freeObjectChanges:") {
		t.Errorf("TrackPlanChangePreview is missing freeObjectChanges: %s", preview)
	}
}

func TestOpenAPIDocumentsLayoutGradeLimitAndWarning(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{"Layout", "LayoutInput"} {
		block := openAPIIndentedBlock(t, contract, schema, 4)
		if !strings.Contains(block, "maxGradePercent:") || !strings.Contains(block, "maximum: 100") ||
			!strings.Contains(block, "exclusiveMinimum: 0") {
			t.Errorf("%s is missing maximum grade constraints: %s", schema, block)
		}
	}
	issue := openAPIIndentedBlock(t, contract, "TrackPlanIssue", 4)
	if !strings.Contains(issue, "grade_limit_exceeded") || !strings.Contains(issue, "gradePercent:") ||
		!strings.Contains(issue, "gradeLimitPercent:") {
		t.Errorf("TrackPlanIssue is missing grade limit details: %s", issue)
	}
	change := openAPIIndentedBlock(t, contract, "TrackPlanIssueChange", 4)
	if !strings.Contains(change, "grade_limit_exceeded") {
		t.Errorf("TrackPlanIssueChange is missing grade_limit_exceeded: %s", change)
	}
}

func TestOpenAPIDocumentsTrackClearanceLimitAndWarning(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, schema := range []string{"Layout", "LayoutInput"} {
		block := openAPIIndentedBlock(t, contract, schema, 4)
		if !strings.Contains(block, "minimumTrackClearanceMm:") ||
			!strings.Contains(block, "exclusiveMinimum: 0") {
			t.Errorf("%s is missing minimum clearance constraints: %s", schema, block)
		}
	}
	issue := openAPIIndentedBlock(t, contract, "TrackPlanIssue", 4)
	for _, token := range []string{"insufficient_clearance", "clearanceMm:", "clearanceLimitMm:",
		"intersectionXMm:", "intersectionYMm:"} {
		if !strings.Contains(issue, token) {
			t.Errorf("TrackPlanIssue is missing %s: %s", token, issue)
		}
	}
	change := openAPIIndentedBlock(t, contract, "TrackPlanIssueChange", 4)
	if !strings.Contains(change, "insufficient_clearance") {
		t.Errorf("TrackPlanIssueChange is missing insufficient_clearance: %s", change)
	}
}

func TestOpenAPIDocumentsMasterDataValidationResponses(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, operation := range []struct {
		path   string
		method string
	}{
		{path: "/master-data/{type}", method: "post"},
		{path: "/master-data/{type}/{key}", method: "put"},
		{path: "/master-data/{type}/{key}", method: "delete"},
	} {
		pathBlock := openAPIIndentedBlock(t, contract, operation.path, 2)
		methodBlock := openAPIIndentedBlock(t, pathBlock, operation.method, 4)
		if !strings.Contains(methodBlock, `        "400":`) ||
			!strings.Contains(methodBlock, `$ref: "#/components/schemas/Problem"`) {
			t.Errorf("%s %s is missing its 400 Problem response: %s", operation.method, operation.path, methodBlock)
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
		"[article, image, inventoryNumber, manufacturer, articleNumber, name,",
		"type, gauge, stock, storage, updatedAt]", "default: inventoryNumber", "enum: [asc, desc]",
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

func TestOpenAPIDocumentsVehicleSetMemberRequestAndRemainingSet(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	request := openAPIIndentedBlock(t, contract, "CreateVehicleSetRequest", 4)
	if !strings.Contains(request, `$ref: "#/components/schemas/CreateVehicleSetMemberRequest"`) {
		t.Errorf("set members do not use their dedicated request schema: %s", request)
	}
	member := openAPIIndentedBlock(t, contract, "CreateVehicleSetMemberRequest", 4)
	if strings.Contains(member, "required: [manufacturer") {
		t.Errorf("member schema incorrectly requires shared set data: %s", member)
	}
	for _, field := range []string{"inventoryNumber:", "name:", "vehicleNumber:", "images:"} {
		if !strings.Contains(member, field) {
			t.Errorf("member schema is missing %s: %s", field, member)
		}
	}
	vehicle := openAPIIndentedBlock(t, contract, "Vehicle", 4)
	setSizeStart := strings.Index(vehicle, "vehicleSetSize:")
	if setSizeStart < 0 || !strings.Contains(vehicle[setSizeStart:], "minimum: 1") {
		t.Errorf("vehicleSetSize does not allow a one-member remainder: %s", vehicle)
	}
	for _, field := range []string{"vehicleSet:", "inventoryNumber:", "memberCount:", "position:"} {
		if !strings.Contains(contract, field) {
			t.Errorf("vehicle set contract is missing %s", field)
		}
	}
	setPath := openAPIIndentedBlock(t, contract, "/vehicle-sets/{id}", 2)
	for _, operation := range []string{"get:", "patch:"} {
		if !strings.Contains(setPath, operation) {
			t.Errorf("vehicle set path is missing %s: %s", operation, setPath)
		}
	}
}

func TestOpenAPIUsesIndividualItemTerminologyForAllocations(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	publicStringPattern := regexp.MustCompile(`(?im)^\s+(?:summary|description):\s*(.+)$`)
	assetTermPattern := regexp.MustCompile(`(?i)\bassets?\b`)
	stableTechnicalIdentifiers := strings.NewReplacer(
		"AccessoryAsset", "",
		"assetId", "",
		"/assets", "",
	)
	for _, match := range publicStringPattern.FindAllStringSubmatch(contract, -1) {
		publicText := stableTechnicalIdentifiers.Replace(match[1])
		if assetTermPattern.MatchString(publicText) {
			t.Errorf("OpenAPI public text still uses asset terminology: %q", match[1])
		}
	}
	for _, summary := range []string{
		"summary: Reserve accessory stock or an individual item",
		"summary: Install accessory stock or an individual item",
	} {
		if !strings.Contains(contract, summary) {
			t.Errorf("OpenAPI is missing %q", summary)
		}
	}
}

func TestOpenAPIArticlePathInventoryMatchesRegisteredRoutes(t *testing.T) {
	operations := readOpenAPIOperations(t)
	registered := map[string]map[string]bool{}
	for _, route := range apiRouteSpecs() {
		path := strings.TrimPrefix(route.Path, "/api/v1")
		if !strings.HasPrefix(path, "/accessory-") {
			continue
		}
		if registered[path] == nil {
			registered[path] = map[string]bool{}
		}
		registered[path][route.Method] = true
	}

	for path, methods := range registered {
		for method := range methods {
			if !operations[path][method] {
				t.Errorf("registered article route is undocumented: %s %s", method, path)
			}
		}
	}
	for path, methods := range operations {
		if !strings.HasPrefix(path, "/accessory-") {
			continue
		}
		for method := range methods {
			if !registered[path][method] {
				t.Errorf("documented article route is unregistered: %s %s", method, path)
			}
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

func TestOpenAPIAccessoryImageImportContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := openAPIIndentedBlock(
		t, string(data), "/accessory-products/{id}/documents/import-url", 2,
	)
	for _, expected := range []string{
		"post:", "AccessoryDocumentImportURLInput", `"201":`, `"400":`, `"403":`, `"413":`, `"415":`, `"502":`,
	} {
		if !strings.Contains(contract, expected) {
			t.Errorf("accessory image import contract is missing %q: %s", expected, contract)
		}
	}
}

func TestOpenAPIArticleListDocumentsValidatedStringFilters(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	articleList := openAPIIndentedBlock(t, string(data), "/accessory-products", 2)
	for _, expectation := range []struct {
		name      string
		maxLength int
		scalar    bool
	}{
		{name: "query", maxLength: 200, scalar: true},
		{name: "manufacturer", maxLength: 200, scalar: true},
		{name: "locationId", maxLength: 128, scalar: true},
		{name: "gauge", maxLength: 128},
	} {
		parameter := openAPIParameterBlock(t, articleList, expectation.name)
		for _, fragment := range []string{
			"minLength: 1", fmt.Sprintf("maxLength: %d", expectation.maxLength),
			"must not contain control characters",
		} {
			if !strings.Contains(parameter, fragment) {
				t.Errorf("%s parameter is missing %q: %s", expectation.name, fragment, parameter)
			}
		}
		if expectation.scalar && !strings.Contains(parameter, "must be supplied at most once") {
			t.Errorf("%s parameter does not document scalar cardinality: %s", expectation.name, parameter)
		}
	}
}

func openAPIParameterBlock(t *testing.T, operation, name string) string {
	t.Helper()
	marker := "        - name: " + name + "\n"
	start := strings.Index(operation, marker)
	if start < 0 {
		t.Fatalf("OpenAPI parameter %s is missing", name)
	}
	remainder := operation[start+len(marker):]
	end := strings.Index(remainder, "        - name: ")
	if end < 0 {
		end = len(remainder)
	}
	return marker + remainder[:end]
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

func TestOpenAPIDocumentsExternalMappingConflict(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	block := openAPIIndentedBlock(t, string(data), "/vehicles/{id}/external-mappings", 2)
	if !strings.Contains(block, `"409":`) {
		t.Fatalf("external mapping conflict response is missing: %s", block)
	}
}
