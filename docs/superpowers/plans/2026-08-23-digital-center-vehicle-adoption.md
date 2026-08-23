# Digital Center Vehicle Adoption Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an unassigned ECoS locomotive create a prefilled RailKeeper vehicle or be explicitly assigned to an existing vehicle without writing to the command station.

**Architecture:** Reuse the existing ECoS vehicle-draft pipeline for full vehicle creation and keep existing-vehicle assignment in a focused Digital Centers controller and dialog. Protect the authoritative `(provider, external_id)` mapping in the Go service, then trigger a safe ECoS read after a successful mapping so persisted comparison rows are rebuilt from current RailKeeper data.

**Tech Stack:** Go, SQLite, React, TypeScript, Vite, Vitest, Testing Library, OpenAPI YAML

## Global Constraints

- Maintain German and English i18n for every new user-facing string.
- Never write locomotive data to ECoS from this workflow and never request a device write grant.
- Reuse `ECoSVehicleDraftPayload`, `useVehicleECoSDraftController`, and `/vehicles/{id}/external-mappings`.
- Never fabricate CV values, function values, manufacturer, gauge, or vehicle class data.
- Preserve the current Admin-only Digital Centers authorization boundary; do not widen route access.
- Do not silently move an existing `(provider, external_id)` mapping to another vehicle.
- Preserve the existing Import/Export ECoS return path.
- Keep dialogs keyboard accessible with focus containment, Escape handling, and semantic labels.
- Preserve all existing uncommitted ECoS stabilization changes in this worktree.

---

## File map

- `backend/internal/application/vehicle_types.go`: mapping ownership conflict error.
- `backend/internal/application/vehicle_external_mappings.go`: idempotent external mapping ownership.
- `backend/internal/application/vehicle_external_mappings_test.go`: mapping regression tests.
- `backend/internal/api/vehicle_handlers.go`: HTTP 409 mapping.
- `backend/internal/api/openapi_contract_test.go`, `openapi/railkeeper.yaml`: conflict contract.
- `frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.ts`: pure draft, mapping, and ranking helpers.
- `frontend/src/features/digitalCenters/useDigitalCenterVehicleAdoption.ts`: vehicle loading and assignment mutation.
- `frontend/src/features/digitalCenters/VehicleAssignmentDialog.tsx`: searchable vehicle picker.
- `frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx`: adoption entry actions.
- `frontend/src/features/digitalCenters/DigitalCentersView.tsx`: navigation, assignment, and safe refresh wiring.
- `frontend/src/features/digitalCenters/digitalCenterModel.ts`: assignment dialog state.
- `frontend/src/features/vehicles/vehicleViewModel.ts`: Digital Centers return metadata.
- `frontend/src/features/vehicles/useVehicleECoSDraftController.ts`: origin-aware return navigation.
- `frontend/src/features/vehicles/vehicleMutationCommands.ts`: mapping save, return, and localized conflicts.
- `frontend/src/shared/i18n/de.ts`, `frontend/src/shared/i18n/en.ts`: bilingual copy.
- `frontend/src/styles/digital-centers.css`: dense, responsive assignment UI.

### Task 1: Protect external mapping ownership

**Files:**
- Modify: `backend/internal/application/vehicle_types.go`
- Modify: `backend/internal/application/vehicle_external_mappings.go`
- Create: `backend/internal/application/vehicle_external_mappings_test.go`
- Modify: `backend/internal/api/vehicle_handlers.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `VehicleService.UpsertExternalMapping(context.Context, string, VehicleExternalMapInput, string)`.
- Produces: `ErrVehicleExternalMappingConflict`; problem code `external_mapping_conflict` with HTTP 409.

- [ ] **Step 1: Write the failing ownership test**

Create `vehicle_external_mappings_test.go` in package `application_test`:

```go
func TestVehicleExternalMappingKeepsAuthoritativeOwner(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	create := func(name string) *application.Vehicle {
		vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
			Manufacturer: "Piko", Name: name, Gauge: "H0",
			Category: "Lokomotive", Gattung: "Diesellok",
		}, "admin-1")
		if err != nil { t.Fatal(err) }
		return vehicle
	}
	first, second := create("First"), create("Second")
	input := application.VehicleExternalMapInput{
		Provider: "ecos", ExternalID: "77", ExternalName: "BR 106",
		ExternalAddress: "3", ExternalProtocol: "DCC", SyncStatus: "linked",
	}
	created, err := service.UpsertExternalMapping(t.Context(), first.ID, input, "admin-1")
	if err != nil || created.VehicleID != first.ID { t.Fatalf("created=%#v err=%v", created, err) }
	repeated, err := service.UpsertExternalMapping(t.Context(), first.ID, input, "admin-1")
	if err != nil || repeated.VehicleID != first.ID { t.Fatalf("repeated=%#v err=%v", repeated, err) }
	if _, err := service.UpsertExternalMapping(t.Context(), second.ID, input, "admin-1");
		!errors.Is(err, application.ErrVehicleExternalMappingConflict) {
		t.Fatalf("conflict error=%v", err)
	}
}
```

- [ ] **Step 2: Run the focused test and verify RED**

```powershell
cd backend
go test ./internal/application -run TestVehicleExternalMappingKeepsAuthoritativeOwner -count=1
```

Expected: compilation fails because `ErrVehicleExternalMappingConflict` is undefined.

- [ ] **Step 3: Implement the ownership guard**

Add to `vehicle_types.go`:

```go
ErrVehicleExternalMappingConflict = errors.New("vehicle external mapping conflict")
```

Before the mapping upsert, inspect the current owner:

```go
existing, err := s.getVehicleExternalMapping(ctx, input.Provider, input.ExternalID)
switch {
case err == nil && existing.VehicleID != vehicleID:
	return nil, ErrVehicleExternalMappingConflict
case err == nil:
	// Continue so metadata and last-seen time can be refreshed idempotently.
case !errors.Is(err, sql.ErrNoRows):
	return nil, fmt.Errorf("check vehicle external mapping owner: %w", err)
}
```

Import `database/sql` and `errors`. Keep `%w` around the `QueryRowContext` error in
`getVehicleExternalMapping` so `sql.ErrNoRows` remains detectable.

- [ ] **Step 4: Add HTTP and OpenAPI conflict behavior**

Extend `upsertVehicleExternalMapping`:

```go
case errors.Is(err, application.ErrVehicleExternalMappingConflict):
	respondProblem(w, http.StatusConflict, "external_mapping_conflict",
		"This digital-center locomotive is already assigned to another vehicle.")
```

Document a `409` response using the shared `Problem` schema. Add this contract test:

```go
func TestOpenAPIDocumentsExternalMappingConflict(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil { t.Fatal(err) }
	block := openAPIIndentedBlock(t, string(data), "/vehicles/{id}/external-mappings", 2)
	if !strings.Contains(block, `"409":`) {
		t.Fatalf("external mapping conflict response is missing: %s", block)
	}
}
```

- [ ] **Step 5: Format, verify GREEN, and commit**

```powershell
gofmt -w backend/internal/application/vehicle_types.go backend/internal/application/vehicle_external_mappings.go backend/internal/application/vehicle_external_mappings_test.go backend/internal/api/vehicle_handlers.go backend/internal/api/openapi_contract_test.go
cd backend
go test ./internal/application ./internal/api -run "ExternalMapping|OpenAPI" -count=1
cd ..
git add backend/internal/application/vehicle_types.go backend/internal/application/vehicle_external_mappings.go backend/internal/application/vehicle_external_mappings_test.go backend/internal/api/vehicle_handlers.go backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml
git commit -m "fix: protect digital center vehicle mappings"
```

Expected: tests PASS; the commit contains only the listed backend and contract files.

### Task 2: Build and return from a Digital Centers vehicle draft

**Files:**
- Create: `frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.ts`
- Create: `frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.test.ts`
- Modify: `frontend/src/features/vehicles/vehicleViewModel.ts`
- Modify: `frontend/src/features/vehicles/useVehicleECoSDraftController.ts`
- Modify: `frontend/src/features/vehicles/vehicleMutationCommands.ts`
- Modify: `frontend/src/features/vehicles/vehicleMutationCommands.test.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: `DigitalCenterWorkItem`, `emptyVehicle`, and `ecosVehicleDraftStorageKey`.
- Produces: `buildDigitalCenterVehicleDraft(item): ECoSVehicleDraftPayload`, `openDigitalCenterVehicleDraft(item): void`, and `returnToDigitalCenters` metadata.

- [ ] **Step 1: Write failing draft tests**

Create a test using the existing `DigitalCenterWorkItem` fixture shape:

```ts
it("builds a bounded ECoS create draft", () => {
  const draft = buildDigitalCenterVehicleDraft({
    ...workItem, sessionId: "session-1", centerObjectId: "77", vehicleId: "",
    name: "BR 106", decoderAddress: 3, protocol: "DCC",
    center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" }
  });
  expect(draft).toMatchObject({
    source: "ecos", mode: "create",
    vehicle: { name: "BR 106", category: "Lokomotive", digital: true, digitalDecoderNumber: "3" },
    externalMapping: { provider: "ecos", externalId: "77", externalAddress: "3",
      externalProtocol: "DCC", syncStatus: "linked" },
    cvValues: [], functionValues: [],
    returnToDigitalCenters: { sessionId: "session-1", objectId: "77" }
  });
  expect(draft.vehicle.manufacturer).toBe("");
  expect(draft.vehicle.gauge).toBe("");
  expect(draft.vehicle.gattung).toBe("");
});
```

Add a second test asserting an empty or non-numeric object ID throws before session storage changes.

- [ ] **Step 2: Run the draft test and verify RED**

```powershell
cd frontend
npm.cmd test -- digitalCenterVehicleAdoption.test.ts --run
```

Expected: FAIL because the adoption module does not exist.

- [ ] **Step 3: Implement the bounded draft and navigation**

Add to `ECoSVehicleDraftPayload`:

```ts
returnToDigitalCenters?: { sessionId: string; objectId: string };
```

Implement:

```ts
export function digitalCenterExternalMapping(item: DigitalCenterWorkItem): VehicleExternalMappingInput {
  const externalId = item.centerObjectId.trim();
  if (!/^\d+$/.test(externalId)) throw new Error("invalid digital-center object id");
  const address = item.center.decoderAddress ?? item.decoderAddress;
  return { provider: "ecos", externalId, externalName: item.center.name?.trim() || item.name.trim(),
    externalAddress: address > 0 ? String(address) : "",
    externalProtocol: item.center.protocol ?? item.protocol, syncStatus: "linked" };
}

export function buildDigitalCenterVehicleDraft(item: DigitalCenterWorkItem): ECoSVehicleDraftPayload {
  const mapping = digitalCenterExternalMapping(item);
  const name = mapping.externalName || `ECoS ${mapping.externalId}`;
  return {
    source: "ecos", mode: "create",
    sourceSummary: { objectId: Number(mapping.externalId), name, address: mapping.externalAddress,
      protocol: mapping.externalProtocol, profile: "" },
    vehicle: { ...emptyVehicle, name, category: "Lokomotive", digital: true,
      digitalDecoderNumber: mapping.externalAddress },
    importedKeys: ["name", "category", "digital", "digitalDecoderNumber"],
    externalMapping: mapping, cvValues: [], functionValues: [],
    unclearFields: ["manufacturer", "gauge", "gattung"],
    returnToDigitalCenters: { sessionId: item.sessionId, objectId: mapping.externalId }
  };
}

export function openDigitalCenterVehicleDraft(item: DigitalCenterWorkItem) {
  const draft = buildDigitalCenterVehicleDraft(item);
  sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(draft));
  history.pushState(null, "", "/vehicles?source=ecos");
  dispatchEvent(new PopStateEvent("popstate"));
}
```

- [ ] **Step 4: Write failing save-and-return tests**

In `vehicleMutationCommands.test.ts`, use a draft with `returnToDigitalCenters` and assert:

```ts
expect(api.upsertVehicleExternalMapping).toHaveBeenCalledWith("created", draft.externalMapping);
expect(options.ecos.markSaved).toHaveBeenCalledWith(draft, "created");
expect(options.editor.close).toHaveBeenCalledOnce();
expect(options.ecos.returnToSession).toHaveBeenCalledWith(draft);
```

Reject the mapping with `new ApiError("conflict", "external_mapping_conflict", 409)` and expect
`vehicles.ecosDraft.mappingConflict` to be sent to `onMessage`.

- [ ] **Step 5: Implement origin-aware return and localized conflicts**

Change the command contract to:

```ts
returnToSession: (draft: ECoSVehicleDraftPayload) => void;
```

Treat either origin as a return flow:

```ts
const savedDraft = ecos.draft;
const returnsToSource = Boolean(savedDraft.returnToEcos || savedDraft.returnToDigitalCenters);
ecos.clear();
if (returnsToSource) {
  reloadVehicles();
  editor.close();
  ecos.returnToSession(savedDraft);
  return;
}
```

Navigate according to the saved draft:

```ts
const returnToImportSession = (savedDraft: ECoSVehicleDraftPayload) => {
  const destination = savedDraft.returnToDigitalCenters
    ? `/digital-centers?sessionId=${encodeURIComponent(savedDraft.returnToDigitalCenters.sessionId)}` +
      `&objectId=${encodeURIComponent(savedDraft.returnToDigitalCenters.objectId)}`
    : "/import-export?source=ecos";
  history.pushState(null, "", destination);
  dispatchEvent(new PopStateEvent("popstate"));
};
```

Translate `external_mapping_conflict` to German and English key `vehicles.ecosDraft.mappingConflict`.

- [ ] **Step 6: Verify GREEN and commit the creation path**

```powershell
cd frontend
npm.cmd test -- digitalCenterVehicleAdoption.test.ts vehicleMutationCommands.test.ts --run
cd ..
git add frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.ts frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.test.ts frontend/src/features/vehicles/vehicleViewModel.ts frontend/src/features/vehicles/useVehicleECoSDraftController.ts frontend/src/features/vehicles/vehicleMutationCommands.ts frontend/src/features/vehicles/vehicleMutationCommands.test.ts frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: prepare ECoS locomotives as vehicles"
```

Expected: focused tests PASS.

### Task 3: Assign an existing vehicle explicitly

**Files:**
- Modify: `frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.ts`
- Modify: `frontend/src/features/digitalCenters/digitalCenterVehicleAdoption.test.ts`
- Create: `frontend/src/features/digitalCenters/useDigitalCenterVehicleAdoption.ts`
- Create: `frontend/src/features/digitalCenters/useDigitalCenterVehicleAdoption.test.tsx`
- Create: `frontend/src/features/digitalCenters/VehicleAssignmentDialog.tsx`
- Create: `frontend/src/features/digitalCenters/VehicleAssignmentDialog.test.tsx`
- Modify: `frontend/src/features/digitalCenters/digitalCenterModel.ts`
- Modify: `frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx`
- Modify: `frontend/src/features/digitalCenters/DigitalCentersView.tsx`
- Modify: `frontend/src/features/digitalCenters/DigitalCentersView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/digital-centers.css`

**Interfaces:**
- Consumes: `api.vehicles("")`, `api.upsertVehicleExternalMapping`, and `digitalCenterExternalMapping`.
- Produces: `rankDigitalCenterVehicleCandidates`; a focused adoption controller; assignment dialog state `{ kind: "assignment"; itemId: string }`.

- [ ] **Step 1: Write failing ranking tests**

```ts
expect(rankDigitalCenterVehicleCandidates(item, [ordinary, nameMatch, addressMatch], "").map(v => v.id))
  .toEqual([addressMatch.id, nameMatch.id, ordinary.id]);
expect(rankDigitalCenterVehicleCandidates(item, [ordinary, nameMatch], "br 106").map(v => v.id))
  .toEqual([nameMatch.id]);
```

Also assert the function returns a new array and never selects a vehicle.

- [ ] **Step 2: Run ranking tests and verify RED**

```powershell
cd frontend
npm.cmd test -- digitalCenterVehicleAdoption.test.ts --run
```

Expected: FAIL because `rankDigitalCenterVehicleCandidates` is undefined.

- [ ] **Step 3: Implement deterministic ranking**

Normalize text by lowercasing and removing non-alphanumeric characters. Filter the query against
inventory number, name, manufacturer, article number, vehicle number, and decoder number. Sort by:

```ts
const score = (vehicle: Vehicle) => {
  if (vehicle.externalMappings?.some(mapping =>
    mapping.provider === "ecos" && mapping.externalId === item.centerObjectId)) return 0;
  if (item.decoderAddress > 0 && vehicle.digitalDecoderNumber === String(item.decoderAddress)) return 1;
  const sourceName = comparable(item.center.name || item.name);
  const vehicleName = comparable(vehicle.name);
  const vehicleNumber = comparable(vehicle.vehicleNumber);
  if (sourceName && (vehicleName === sourceName || vehicleNumber === sourceName ||
    sourceName.includes(vehicleName) || sourceName.includes(vehicleNumber))) return 2;
  return 3;
};
```

Within equal scores, sort by inventory number then name.

- [ ] **Step 4: Write failing controller tests**

```ts
await act(async () => result.current.commands.load(item));
expect(api.vehicles).toHaveBeenCalledWith("");
expect(result.current.state.vehicles).toEqual(vehicles);

await act(async () => result.current.commands.assign(item, "vehicle-2"));
expect(api.upsertVehicleExternalMapping).toHaveBeenCalledWith(
  "vehicle-2", digitalCenterExternalMapping(item)
);
expect(onAssigned).toHaveBeenCalledOnce();
```

For `external_mapping_conflict`, assert localized error text, preserved selection, and no
`onAssigned` call.

- [ ] **Step 5: Implement the focused adoption controller**

Expose exactly:

```ts
type DigitalCenterVehicleAdoptionController = {
  state: { vehicles: Vehicle[]; selectedVehicleId: string; loading: boolean; saving: boolean; error: string };
  setters: { setSelectedVehicleId(value: string): void };
  commands: {
    load(item: DigitalCenterWorkItem): Promise<void>;
    assign(item: DigitalCenterWorkItem, vehicleId: string): Promise<void>;
    reset(): void;
  };
};
```

The controller performs no navigation and no ECoS write. Call `onAssigned` only after the mapping
request succeeds.

- [ ] **Step 6: Write failing assignment-dialog tests**

Verify search filtering, visible textual match reasons, no initial selection, disabled confirmation,
one confirmed `onAssign` call, Escape close, focus placement, long labels, and `role="alert"` errors.

- [ ] **Step 7: Implement the assignment dialog**

Use `useModalDialogLayer`, a search input, and radio selection. Each candidate renders:

```tsx
<strong>{vehicle.inventoryNumber} · {vehicle.name}</strong>
<small>{vehicle.manufacturer} · {vehicle.articleNumber || "–"}</small>
<span>{t("digitalCenters.assignment.address", {
  value: vehicle.digitalDecoderNumber || "–"
})}</span>
```

Use textual match badges, not color alone. Disable assignment while saving or without selection.

- [ ] **Step 8: Write failing view integration tests**

```ts
expect(screen.getByRole("button", { name: "Neues Fahrzeug anlegen" })).toBeEnabled();
expect(screen.getByRole("button", { name: "Bestehendem Fahrzeug zuordnen" })).toBeEnabled();
```

Click create and assert the session-storage draft and `/vehicles?source=ecos` navigation. Click
assignment and assert the picker opens. Verify adoption actions are unavailable without Admin role.

- [ ] **Step 9: Wire the actions and safe refresh**

Extend the dialog union:

```ts
export type DigitalCenterWorkspaceDialog =
  | { kind: "comparison"; itemId: string }
  | { kind: "write-preview"; itemId: string }
  | { kind: "assignment"; itemId: string };
```

Pass `canAdopt={roles.includes("Admin")}`, `onCreateVehicle`, and `onAssignVehicle` into the
comparison dialog. Show adoption actions only for unassigned items with a numeric object ID.

In `DigitalCentersView`, use the adoption controller. A successful existing-vehicle assignment
closes the picker and calls `workspace.readData()`. On return from creation, consume `sessionId` and
`objectId` query parameters once, wait for `actions.canRead`, call `workspace.readData()`, and remove
the parameters with `history.replaceState` before awaiting the read so rerenders cannot repeat it.

- [ ] **Step 10: Add bilingual copy and focused styling**

Add German and English keys under `digitalCenters.assignment` for create, assign, search, no results,
match reasons, address, confirm, cancel, success, loading error, save error, and mapping conflict.
Replace the old unassigned dead-end text with guidance to create or assign a vehicle.

Add CSS for a bounded scroll area, compact candidate rows, visible keyboard focus, textual match
reasons, and a single-column narrow layout. Use existing tokens only.

- [ ] **Step 11: Verify GREEN and commit the assignment path**

```powershell
cd frontend
npm.cmd test -- digitalCenterVehicleAdoption.test.ts useDigitalCenterVehicleAdoption.test.tsx VehicleAssignmentDialog.test.tsx DigitalCentersView.test.tsx vehicleMutationCommands.test.ts --run
cd ..
git add frontend/src/features/digitalCenters frontend/src/features/vehicles/vehicleViewModel.ts frontend/src/features/vehicles/useVehicleECoSDraftController.ts frontend/src/features/vehicles/vehicleMutationCommands.ts frontend/src/features/vehicles/vehicleMutationCommands.test.ts frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/digital-centers.css
git commit -m "feat: assign ECoS locomotives to vehicles"
```

Expected: focused tests PASS.

### Task 4: Full verification and local hardware smoke test

**Files:**
- Verify only; if a check fails, modify only the smallest responsible file and rerun its focused test first.

**Interfaces:**
- Consumes: all prior task outputs.
- Produces: verified branch ready for user review.

- [ ] **Step 1: Run formatting and diff hygiene**

```powershell
gofmt -w backend/internal/application/vehicle_types.go backend/internal/application/vehicle_external_mappings.go backend/internal/application/vehicle_external_mappings_test.go backend/internal/api/vehicle_handlers.go backend/internal/api/openapi_contract_test.go
git diff --check
```

Expected: `git diff --check` has no output.

- [ ] **Step 2: Run all backend tests**

```powershell
cd backend
go test ./...
```

Expected: every Go package passes.

- [ ] **Step 3: Run all frontend tests and build**

```powershell
cd frontend
npm.cmd test -- --run
npm.cmd run build
```

Expected: all Vitest tests pass; Vite finishes without TypeScript errors.

- [ ] **Step 4: Rebuild and restart the local worktree server**

Build `frontend/dist`, then restart the existing process on `127.0.0.1:8081` with this worktree's
backend, migrations, seeds, static directory, and the current local data directory. Keep the process
hidden and do not change the user's data location.

- [ ] **Step 5: Smoke-test creation with real ECoS data**

Read ECoS data, open an unassigned locomotive, choose **Neues Fahrzeug anlegen**, verify the full
editor prefill and required fields, then cancel. Confirm no vehicle or mapping was created. Do not
save a real inventory vehicle without explicit user confirmation.

- [ ] **Step 6: Smoke-test assignment without mutating ECoS**

Use a disposable local vehicle or obtain explicit confirmation before assigning a real vehicle.
Verify search, textual match reasons, explicit selection, safe refresh, resulting mapping, retained
live telemetry, and absence of a device write preview or grant.

- [ ] **Step 7: Audit the final branch**

```powershell
git status --short
git log --oneline -5
git diff origin/main...HEAD --stat
```

Expected: only intended source, tests, OpenAPI, and approved design/plan files are present. No
`frontend/dist`, `data`, `.cache`, credentials, or backups are tracked.
