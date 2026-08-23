# Sicheres ECoS-Zurückschreiben / Safe ECoS Write-Back Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Geprüfte RailKeeper-Lokstammdaten sicher, konfliktgeschützt und verifiziert in eine reale
ESU ECoS zurückschreiben. / Safely write reviewed RailKeeper locomotive master data back to a real
ESU ECoS with conflict protection and verification.

**Architecture:** `DigitalCenterWorkspaceService` koordiniert eine exklusive ECoS-Operation,
pausiert einen laufenden Live-Monitor, validiert Vorschau und Adressbelegung erneut, schreibt genau
einen kombinierten Befehl und verifiziert das Zielobjekt über eine schlanke Kontroll-Lesung.
`ECoSService` bleibt die Geräteprotokollgrenze; Repository, OpenAPI und React erhalten nur die für
strukturierte Ergebnisse notwendigen Erweiterungen. / `DigitalCenterWorkspaceService` coordinates
one exclusive ECoS operation, pauses a running live monitor, revalidates preview and address
ownership, sends exactly one combined command, and verifies the target through a lightweight read.
`ECoSService` remains the device protocol boundary; repository, OpenAPI, and React receive only the
structured-result extensions they require.

**Tech Stack:** Go 1.24, `net`/TCP, SQLite, OpenAPI 3.1, React, TypeScript strict mode, Vitest,
Testing Library, Vite.

## Global Constraints

- Schreibumfang ausschließlich Name, Decoder-Adresse und Protokoll. / Write scope is limited to
  name, decoder address, and protocol.
- `WriteCVs` bleibt `false`; CVs, Funktionen, Fahrzustände, Bilder und Anlagenobjekte werden weder
  gelesen noch geschrieben. / `WriteCVs` stays `false`; CVs, functions, runtime state, images, and
  layout objects are neither read nor written by this workflow.
- Host und Port stammen ausschließlich aus serverseitigen Einstellungen. / Host and port come only
  from server-side settings.
- Schreiben bleibt Admin- und CSRF-geschützt. / Writing remains Admin-only and CSRF-protected.
- Jede Mutation benötigt frische Lesesitzung, konfliktfreie Zuordnung, benutzergebundene
  Einmalfreigabe und ausdrückliche Bestätigung. / Every mutation requires a fresh read session,
  conflict-free mapping, actor-bound single-use grant, and explicit confirmation.
- Kein automatischer Retry nach einem gesendeten Schreibbefehl. / Never retry automatically after a
  write command has been sent.
- Backend, OpenAPI, strikter TypeScript-Client sowie deutsche und englische UI bleiben synchron. /
  Backend, OpenAPI, strict TypeScript client, and German/English UI stay aligned.
- Rohbefehle, Tokens, Hostnamen, IP-Adressen und lokale Loknamen gelangen nicht in Audit oder
  persistente Meldungen. / Raw commands, tokens, host names, IP addresses, and local locomotive
  names do not enter audit or persistent messages.

## File Structure

- Create `backend/internal/application/ecos_write_safety.go`: schlanke Master-Reads und
  Klassifizierung unklarer Schreibzustände. / Lightweight master reads and unknown-write
  classification.
- Create `backend/internal/application/ecos_write_safety_test.go`: TCP-Vertragstests für diese
  Geräteoperationen. / TCP contract tests for these device operations.
- Create `backend/internal/application/digital_center_ecos_operation.go`: Operationssperre,
  Live-Pause/-Fortsetzung und koordinierter Cleanup. / Operation gate, live pause/resume, and
  coordinated cleanup.
- Create `backend/internal/application/digital_center_ecos_operation_test.go`: Reihenfolge- und
  Nebenläufigkeitstests. / Ordering and concurrency tests.
- Create `backend/internal/application/digital_center_write_conflicts.go`: typisierter
  Adresskonflikt und reine Konfliktprüfung. / Typed address conflict and pure conflict detection.
- Create `backend/internal/application/digital_center_write_conflicts_test.go`: Konfliktfälle für
  Vorschau und Bestätigung. / Preview and confirmation conflict cases.
- Modify `backend/internal/application/digital_center_write.go`: bestehende Vorschau und Bestätigung
  an die neuen Bausteine anbinden, ohne weitere Protokolllogik aufzunehmen. / Connect the existing
  preview and confirmation to the new units without adding more protocol logic.
- Modify `backend/internal/application/digital_center_workspace.go`: Service-Sperre und interne
  Live-Start/-Stop-Helfer. / Service gate and internal live start/stop helpers.
- Modify `backend/internal/application/digital_center_workspace_types.go`: Meldungscodes und
  `UpdateWorkItem`-Repositoryvertrag. / Message codes and `UpdateWorkItem` repository contract.
- Modify `backend/internal/infrastructure/digital_center_workspace_repository.go`: atomare
  Aktualisierung eines Arbeitslisteneintrags. / Atomic work-item update.
- Modify `backend/internal/infrastructure/digital_center_workspace_repository_test.go`: SQLite-
  Rundlauf der Aktualisierung. / SQLite round trip for the update.
- Modify `backend/internal/api/http_helpers.go`: optionale strukturierte Problemdetails. / Optional
  structured problem details.
- Modify `backend/internal/api/digital_center_workspace_handlers.go`: neue Fehlercodes und sichere
  Details. / New error codes and safe details.
- Modify `backend/internal/api/digital_center_write_handlers_test.go`: HTTP- und Detailvertrag. /
  HTTP and details contract.
- Modify `backend/internal/api/digital_center_openapi_test.go`: neue Ergebniswerte und Felder. / New
  result values and fields.
- Modify `openapi/railkeeper.yaml`: Problem-, Vorschau- und Bestätigungsschemas. / Problem, preview,
  and confirmation schemas.
- Modify `frontend/src/features/digitalCenters/digitalCenterModel.ts`: strikte Ergebnistypen. /
  Strict result types.
- Modify `frontend/src/shared/api.ts`: typisierte Problemdetails. / Typed problem details.
- Modify `frontend/src/features/digitalCenters/useDigitalCentersWorkspace.ts`: konflikt- und
  ergebnisbewusste Zustandsaktualisierung. / Conflict- and result-aware state updates.
- Modify `frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx`: Fortschritt,
  verifizierte Werte, unklarer Zustand und Neustartwarnung. / Progress, verified values, unknown
  state, and restart warning.
- Modify `frontend/src/features/digitalCenters/DigitalCentersFlows.test.tsx` and
  `useDigitalCentersWorkspace.test.tsx`: Benutzerablauf und Hook-Rennen. / User flow and hook races.
- Modify `frontend/src/shared/i18n/de.ts`, `frontend/src/shared/i18n/en.ts`, and
  `frontend/src/shared/i18n.test.ts`: vollständige DE/EN-Texte. / Complete German/English copy.
- Create `backend/internal/application/testdata/ecos/real-device-success.txt`,
  `real-device-incomplete.txt`, `real-device-malformed.txt`, and `real-device-interrupted.txt`:
  anonymisierte Antwortblöcke der realen Abnahme. / Anonymized reply blocks from real-device
  acceptance.
- Create `backend/internal/application/ecos_real_device_fixture_test.go`: Regressionstest des
  aufgenommenen Korpus. / Regression test for the captured corpus.
- Modify `docs/site/de/guide/import-export/ecos-sync.md` and
  `docs/site/guide/import-export/ecos-sync.md`: Kompatibilitätsmatrix und Geräteabnahme. /
  Compatibility matrix and device acceptance.

---

### Task 1: Schlanke ECoS-Stammdatenlesung und unklarer Schreibzustand / Lightweight ECoS Master Reads and Unknown Write State

**Files:**
- Create: `backend/internal/application/ecos_write_safety.go`
- Create: `backend/internal/application/ecos_write_safety_test.go`
- Modify: `backend/internal/application/ecos.go:645-810`
- Test: `backend/internal/application/ecos_test.go`

**Interfaces:**
- Produces: `ListLocomotives(context.Context, ECoSConnectionInput) ([]ECoSLocomotive, error)`
- Produces: `ReadLocomotive(context.Context, ECoSConnectionInput, int) (ECoSLocomotive, error)`
- Produces: `ErrECoSWriteStateUnknown`
- Consumes: existing `eCoSLocomotiveListCommand`, `fetchLocomotiveDetails`, and
  `exchangeRequestedCommands`.

- [ ] **Step 1: Write failing master-read TCP tests / Fehlende TCP-Tests schreiben**

```go
func TestECoSListLocomotivesUsesOnlyApprovedMasterFields(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		if command != eCoSLocomotiveListCommand {
			t.Fatalf("command = %q", command)
		}
		return []string{
			"<REPLY queryObjects(10, addr, name, protocol)>",
			`1001 addr[3] name["Testlok A"] protocol[DCC]`,
			`1002 addr[4] name["Testlok B"] protocol[MM27]`,
			"<END 0 (OK)>",
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	locomotives, err := service.ListLocomotives(t.Context(), ECoSConnectionInput{Host: host, Port: port})
	if err != nil || len(locomotives) != 2 || locomotives[1].Address != 4 {
		t.Fatalf("locomotives=%#v err=%v", locomotives, err)
	}
}

func TestECoSReadLocomotiveUsesTargetedMasterGet(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch command {
		case "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case eCoSLocomotiveDetailCommand(1001):
			return []string{"<REPLY get(1001, addr, name, protocol, profile)>",
				`1001 name["Testlok A"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	locomotive, err := service.ReadLocomotive(t.Context(),
		ECoSConnectionInput{Host: host, Port: port}, 1001)
	if err != nil || locomotive.ObjectID != 1001 || locomotive.Name != "Testlok A" {
		t.Fatalf("locomotive=%#v err=%v", locomotive, err)
	}
}
```

- [ ] **Step 2: Run the focused tests and confirm RED / Fokustests ausführen und RED bestätigen**

Run: `cd backend; go test ./internal/application -run 'TestECoS(ListLocomotives|ReadLocomotive)'`

Expected: FAIL because `ListLocomotives` and `ReadLocomotive` do not exist.

- [ ] **Step 3: Implement the two bounded reads / Beide begrenzten Leseoperationen implementieren**

```go
var ErrECoSWriteStateUnknown = errors.New("ECoS write state is unknown")

func (s *ECoSService) ListLocomotives(
	ctx context.Context,
	input ECoSConnectionInput,
) ([]ECoSLocomotive, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return nil, err
	}
	lines, err := s.exchange(ctx, target.Host, target.Port, eCoSLocomotiveListCommand)
	if err != nil {
		return nil, err
	}
	return parseECoSLocomotives(lines), nil
}

func (s *ECoSService) ReadLocomotive(
	ctx context.Context,
	input ECoSConnectionInput,
	objectID int,
) (ECoSLocomotive, error) {
	target, err := normalizeECoSInput(input)
	if err != nil {
		return ECoSLocomotive{}, err
	}
	if objectID < 1 || objectID > maxDigitalCenterObjectID {
		return ECoSLocomotive{}, ErrDigitalCenterDeviceOutput
	}
	locomotive, err := s.fetchLocomotiveDetails(ctx, target.Host, target.Port, objectID)
	if err != nil {
		return ECoSLocomotive{}, err
	}
	return *locomotive, nil
}
```

Keep this code in `ecos_write_safety.go`; `ecos.go` continues to own parsing and transport helpers.

- [ ] **Step 4: Write and run the unknown-state tests / Tests für unklaren Zustand schreiben und ausführen**

```go
func TestECoSSyncMarksMissingWriteReplyAsUnknown(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{"<REPLY get(1001, addr, name, protocol, profile)>",
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return nil // command was accepted by TCP, but the reply is deliberately dropped
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	service.timeout = 100 * time.Millisecond
	_, err := service.SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New"}, Confirm: true,
	})
	if !errors.Is(err, ErrECoSWriteStateUnknown) {
		t.Fatalf("error=%v", err)
	}
}

func TestECoSSyncKeepsExplicitRejectionDefinite(t *testing.T) {
	listener := startECoSTestServer(t, func(command string) []string {
		switch {
		case command == "request(1001, view)":
			return []string{"<REPLY request(1001, view)>", "<END 0 (OK)>"}
		case command == eCoSLocomotiveDetailCommand(1001):
			return []string{"<REPLY get(1001, addr, name, protocol, profile)>",
				`1001 name["Old"] addr[3] protocol[DCC]`, "<END 0 (OK)>"}
		case strings.HasPrefix(command, "set(1001,"):
			return []string{"<REPLY " + command + ">", "<END 11 (unsupported)>"}
		case command == "release(1001, view)":
			return []string{"<REPLY release(1001, view)>", "<END 0 (OK)>"}
		default:
			t.Fatalf("command = %q", command)
			return nil
		}
	})
	defer func() { _ = listener.Close() }()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	_, err := service.SyncLocomotive(t.Context(), ECoSLocomotiveSyncInput{
		Host: host, Port: port, ObjectID: 1001,
		Desired: ECoSLocomotiveSyncDesired{Name: "New"}, Confirm: true,
	})
	if err == nil || errors.Is(err, ErrECoSWriteStateUnknown) {
		t.Fatalf("error=%v", err)
	}
}
```

Run: `cd backend; go test ./internal/application -run 'TestECoSSync(Marks|Keeps)'`

Expected before implementation: first test FAIL. After classifying send/read ambiguity with
`fmt.Errorf("%w: write reply unavailable", ErrECoSWriteStateUnknown)`: PASS.

- [ ] **Step 5: Run application tests / Anwendungstests ausführen**

Run: `cd backend; go test ./internal/application`

Expected: PASS.

- [ ] **Step 6: Commit / Committen**

```powershell
git add `
  backend/internal/application/ecos.go `
  backend/internal/application/ecos_write_safety.go `
  backend/internal/application/ecos_write_safety_test.go `
  backend/internal/application/ecos_test.go
git commit -m "feat: sichere ECoS-Stammdatenlesung ergänzen / add safe ECoS master reads"
```

### Task 2: Exklusive Operation und kontrollierte Live-Pause / Exclusive Operation and Controlled Live Pause

**Files:**
- Create: `backend/internal/application/digital_center_ecos_operation.go`
- Create: `backend/internal/application/digital_center_ecos_operation_test.go`
- Modify: `backend/internal/application/digital_center_workspace.go:15-280`
- Modify: `backend/internal/application/digital_center_compare.go:48-110`
- Modify: `backend/internal/application/ecos.go:170-450`
- Test: `backend/internal/application/ecos_live_telemetry_test.go`

**Interfaces:**
- Produces: `PauseLive(context.Context) (ECoSLiveStatus, error)` on `ECoSService`.
- Produces: `withECoSOperation(func() error) error` behavior through `operationMu`.
- Consumes: existing `StartLiveWithInterruption`, `StopLive`, and `LiveStatus`.

- [ ] **Step 1: Write failing live-pause and serialization tests / Fehlende Pause- und Sperrtests schreiben**

```go
func TestECoSPauseLiveWaitsForSessionShutdown(t *testing.T) {
	service, closed := runningLiveService(t)
	status, err := service.PauseLive(t.Context())
	if err != nil || status.State != ECoSLiveStopped {
		t.Fatalf("status=%#v err=%v", status, err)
	}
	select {
	case <-closed:
	default:
		t.Fatal("PauseLive returned before the live connection closed")
	}
}

func TestDigitalCenterECoSOperationsAreSerialized(t *testing.T) {
	service := &DigitalCenterWorkspaceService{}
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	go service.runECoSOperation(t.Context(), func() error {
		close(firstEntered)
		<-releaseFirst
		return nil
	})
	<-firstEntered
	secondEntered := make(chan struct{})
	go service.runECoSOperation(t.Context(), func() error { close(secondEntered); return nil })
	select {
	case <-secondEntered:
		t.Fatal("second operation entered before first completed")
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseFirst)
	<-secondEntered
}

func runningLiveService(t *testing.T) (*ECoSService, <-chan struct{}) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			close(closed)
			return
		}
		defer close(closed)
		defer func() { _ = conn.Close() }()
		reader := bufio.NewReader(conn)
		for {
			if _, readErr := reader.ReadString('\n'); readErr != nil {
				return
			}
		}
	}()
	host, port := splitTestAddress(t, listener.Addr().String())
	service := NewECoSService()
	if _, err := service.StartLive(t.Context(), ECoSConnectionInput{Host: host, Port: port}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { service.StopLive(); _ = listener.Close() })
	return service, closed
}
```

- [ ] **Step 2: Run the focused tests and confirm RED / Fokustests ausführen und RED bestätigen**

Run: `cd backend; go test ./internal/application -run 'Test(ECoSPauseLive|DigitalCenterECoSOperations)'`

Expected: FAIL because pause acknowledgement and the operation gate do not exist.

- [ ] **Step 3: Add live completion tracking / Abschluss des Live-Laufs nachverfolgbar machen**

Add `liveDone chan struct{}` to `ECoSService`. On each successful `StartLiveWithInterruption`, create
a new channel for that generation. The live goroutine closes only its own channel when the
connection is closed. Implement:

```go
func (s *ECoSService) PauseLive(ctx context.Context) (ECoSLiveStatus, error) {
	s.liveMu.Lock()
	done := s.liveDone
	s.stopLiveLocked()
	s.liveGeneration++
	s.liveStatus.Connected = false
	s.liveStatus.State = ECoSLiveStopped
	s.liveStatus.Diagnosis.ConnectionState = ECoSLiveStopped
	status := cloneECoSLiveStatus(s.liveStatus)
	s.liveMu.Unlock()
	if done == nil {
		return status, nil
	}
	select {
	case <-done:
		return status, nil
	case <-ctx.Done():
		return status, fmt.Errorf("pause ECoS live monitor: %w", ctx.Err())
	}
}
```

Guard channel closure with the live generation so an old goroutine cannot close a new run's
channel.

- [ ] **Step 4: Add the workspace operation gate and unlocked helpers / Arbeitsbereichssperre und interne Helfer ergänzen**

Add `operationMu sync.Mutex` to `DigitalCenterWorkspaceService` and implement:

```go
func (service *DigitalCenterWorkspaceService) runECoSOperation(
	ctx context.Context,
	operation func() error,
) error {
	if service == nil {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	service.operationMu.Lock()
	defer service.operationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	return operation()
}
```

Extract `startLiveMonitorUnlocked` and `stopLiveMonitorUnlocked`; public start/stop methods acquire
the gate exactly once. Apply the same gate to `StartReadSession`, `PreviewWrite`, and `ConfirmWrite`.
Do not hold this mutex for the lifetime of the passive monitor.

- [ ] **Step 5: Run live and workspace tests / Live- und Arbeitsbereichstests ausführen**

Run: `cd backend; go test ./internal/application -run 'Test(ECoSLive|ECoSPauseLive|DigitalCenterECoSOperations|DigitalCenterWorkspace)'`

Expected: PASS without race or timeout.

- [ ] **Step 6: Commit / Committen**

```powershell
git add `
  backend/internal/application/digital_center_ecos_operation.go `
  backend/internal/application/digital_center_ecos_operation_test.go `
  backend/internal/application/digital_center_workspace.go `
  backend/internal/application/digital_center_compare.go `
  backend/internal/application/ecos.go `
  backend/internal/application/ecos_live_telemetry_test.go
git commit -m "feat: ECoS-Operationen exklusiv koordinieren / coordinate ECoS operations exclusively"
```

### Task 3: Blockierende Decoder-Adresskonflikte / Blocking Decoder Address Conflicts

**Files:**
- Create: `backend/internal/application/digital_center_write_conflicts.go`
- Create: `backend/internal/application/digital_center_write_conflicts_test.go`
- Modify: `backend/internal/application/digital_center_write.go:113-151,354-410`
- Modify: `backend/internal/api/http_helpers.go:138-155`
- Modify: `backend/internal/api/digital_center_workspace_handlers.go:180-250`
- Modify: `backend/internal/api/digital_center_write_handlers_test.go`
- Modify: `openapi/railkeeper.yaml:36-45,11235-11290`

**Interfaces:**
- Produces: `DigitalCenterAddressConflictError` with `ObjectID`, `Name`, and `Address`.
- Produces: optional `Problem.details` object and `respondProblemDetails`.
- Consumes: Task 1 `ListLocomotives` and Task 2 operation gate.

- [ ] **Step 1: Write failing address-conflict tests / Fehlende Adresskonflikttests schreiben**

```go
func TestDigitalCenterWritePreviewBlocksAddressOwnedByAnotherObject(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	targetObjectID, err := strconv.Atoi(fixture.item.CenterObjectID)
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: targetObjectID, Name: "Target", Address: 3, Protocol: "DCC"},
		{ObjectID: 2002, Name: "Other", Address: 18, Protocol: "DCC"},
	}
	_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"address"}}, "admin-1")
	var conflict *DigitalCenterAddressConflictError
	if !errors.As(err, &conflict) || conflict.ObjectID != 2002 || conflict.Name != "Other" {
		t.Fatalf("error=%#v", err)
	}
}

func TestDigitalCenterWritePreviewAllowsNameOnlyWithExistingAddressCollision(t *testing.T) {
	fixture := newDigitalCenterWriteFixture(t)
	targetObjectID, err := strconv.Atoi(fixture.item.CenterObjectID)
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: targetObjectID, Name: "Target", Address: 3, Protocol: "DCC"},
		{ObjectID: 2002, Name: "Other", Address: 3, Protocol: "DCC"},
	}
	_, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: []string{"name"}}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run the focused tests and confirm RED / Fokustests ausführen und RED bestätigen**

Run: `cd backend; go test ./internal/application -run 'TestDigitalCenterWritePreview(Block|Allow)'`

Expected: FAIL because no address ownership check exists.

- [ ] **Step 3: Implement pure address conflict detection / Reine Adresskonfliktprüfung implementieren**

```go
type DigitalCenterAddressConflictError struct {
	ObjectID int
	Name     string
	Address  int
}

func (err *DigitalCenterAddressConflictError) Error() string {
	return "decoder address is already assigned to another ECoS object"
}

func normalizeSafeConflictName(value string) string {
	name, err := normalizeDigitalCenterName(value)
	if err != nil || name == "" {
		return "ECoS locomotive"
	}
	return name
}

func findECoSAddressConflict(
	locomotives []ECoSLocomotive,
	targetObjectID int,
	desiredAddress int,
) *DigitalCenterAddressConflictError {
	for _, locomotive := range locomotives {
		if locomotive.ObjectID != targetObjectID && locomotive.Address == desiredAddress {
			return &DigitalCenterAddressConflictError{
				ObjectID: locomotive.ObjectID,
				Name: normalizeSafeConflictName(locomotive.Name),
				Address: desiredAddress,
			}
		}
	}
	return nil
}
```

Call it only when `changes` contains `address`. Validate the locomotive count, object IDs, address
range, and safe display name before creating a grant.

Extend the existing write stub with the exact Task 1 interface:

```go
type digitalCenterWriteECoSStub struct {
	// existing fields remain
	locomotives []ECoSLocomotive
}

func (stub *digitalCenterWriteECoSStub) ListLocomotives(
	context.Context,
	ECoSConnectionInput,
) ([]ECoSLocomotive, error) {
	return append([]ECoSLocomotive(nil), stub.locomotives...), nil
}
```

- [ ] **Step 4: Add structured safe HTTP details / Strukturierte sichere HTTP-Details ergänzen**

```go
func respondProblemDetails(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	details map[string]any,
) {
	respondJSON(w, status, map[string]any{
		"error": code,
		"message": message,
		"details": details,
	})
}
```

Map `DigitalCenterAddressConflictError` to HTTP 409 code
`digital_center_address_conflict` and details `{objectId, name, decoderAddress}`. Add optional
`details: {type: object, additionalProperties: true}` to OpenAPI `Problem`.

- [ ] **Step 5: Run application, API, and OpenAPI tests / Anwendung, API und OpenAPI testen**

Run: `cd backend; go test ./internal/application ./internal/api -run 'DigitalCenter.*(Address|Write|OpenAPI)'`

Expected: PASS; the 409 body contains only the three approved detail fields.

- [ ] **Step 6: Commit / Committen**

```powershell
git add `
  backend/internal/application/digital_center_write_conflicts.go `
  backend/internal/application/digital_center_write_conflicts_test.go `
  backend/internal/application/digital_center_write.go `
  backend/internal/api/http_helpers.go `
  backend/internal/api/digital_center_workspace_handlers.go `
  backend/internal/api/digital_center_write_handlers_test.go `
  openapi/railkeeper.yaml
git commit -m "feat: ECoS-Adresskonflikte blockieren / block ECoS address conflicts"
```

### Task 4: Koordiniertes Schreiben, Verifikation und Persistenz / Coordinated Write, Verification, and Persistence

**Files:**
- Modify: `backend/internal/application/digital_center_write.go:25-245,500-616`
- Modify: `backend/internal/application/digital_center_workspace_types.go:15-126`
- Modify: `backend/internal/application/digital_center_ecos_operation.go`
- Modify: `backend/internal/application/digital_center_write_test.go`
- Modify: `backend/internal/infrastructure/digital_center_workspace_repository.go:180-410`
- Modify: `backend/internal/infrastructure/digital_center_workspace_repository_test.go`
- Modify: `backend/internal/api/digital_center_workspace_handlers.go:116-250`
- Modify: `backend/internal/api/digital_center_write_handlers_test.go`

**Interfaces:**
- Produces result enum: `verified | verification_failed | unknown | failed`.
- Produces sentinel: `ErrDigitalCenterLivePauseFailed` for a guaranteed pre-mutation abort.
- Produces: `DigitalCenterWriteLiveResult{WasRunning, Restarted bool}`.
- Produces: confirmation fields `VerifiedValues *ECoSLocomotiveSyncSnapshot`,
  `LiveMonitor DigitalCenterWriteLiveResult`, and `WorkItem *DigitalCenterWorkItem`.
- Produces repository method:
  `UpdateWorkItem(context.Context, DigitalCenterWorkItem) (DigitalCenterWorkItem, error)`.
- Produces:
  `captureAndPauseLive(context.Context, digitalCenterWriteTarget) (DigitalCenterWriteLiveResult, error)`.
- Produces:
  `resumeLiveAfterWrite(context.Context, digitalCenterWriteTarget, *DigitalCenterWriteLiveResult)`.
- Produces `applyAndVerifyDigitalCenterWrite` with inputs `context.Context`,
  `digitalCenterWriteTarget`, `[]ECoSLocomotiveSyncChange`, `DigitalCenterWriteLiveResult`, and
  actor `string`; it returns `(DigitalCenterWriteConfirmation, error)`.
- Consumes Tasks 1–3 interfaces.

- [ ] **Step 1: Write failing orchestration outcome tests / Fehlende Ablauf- und Ergebnistests schreiben**

```go
func TestDigitalCenterConfirmWritePausesWritesVerifiesAndResumes(t *testing.T) {
	fixture := newCoordinatedDigitalCenterWriteFixture(t, ECoSLiveRunning)
	preview := fixture.preview(t, []string{"name", "address"})
	result, err := fixture.confirm(t, preview)
	if err != nil {
		t.Fatal(err)
	}
	wantEvents := []string{"pause", "list", "dry-run", "write", "read-target", "resume"}
	if !reflect.DeepEqual(fixture.ecos.events, wantEvents) {
		t.Fatalf("events=%#v", fixture.ecos.events)
	}
	if result.Result != DigitalCenterWriteVerified || result.WorkItem == nil ||
		!result.LiveMonitor.WasRunning || !result.LiveMonitor.Restarted {
		t.Fatalf("result=%#v", result)
	}
}

func TestDigitalCenterConfirmWriteReturnsUnknownWithoutRetry(t *testing.T) {
	fixture := newCoordinatedDigitalCenterWriteFixture(t, ECoSLiveStopped)
	fixture.ecos.writeErr = ErrECoSWriteStateUnknown
	result, err := fixture.confirm(t, fixture.preview(t, []string{"name"}))
	if err != nil || result.Result != DigitalCenterWriteUnknown || fixture.ecos.writeCalls != 1 {
		t.Fatalf("result=%#v err=%v calls=%d", result, err, fixture.ecos.writeCalls)
	}
}

func TestDigitalCenterConfirmWriteKeepsVerifiedResultWhenResumeFails(t *testing.T) {
	fixture := newCoordinatedDigitalCenterWriteFixture(t, ECoSLiveRunning)
	fixture.ecos.resumeErr = errors.New("restart failed")
	result, err := fixture.confirm(t, fixture.preview(t, []string{"name"}))
	if err != nil || result.Result != DigitalCenterWriteVerified || result.LiveMonitor.Restarted {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestDigitalCenterConfirmWriteAbortsBeforeMutationWhenPauseFails(t *testing.T) {
	fixture := newCoordinatedDigitalCenterWriteFixture(t, ECoSLiveRunning)
	preview := fixture.preview(t, []string{"name"})
	fixture.ecos.pauseErr = errors.New("live connection did not close")
	_, err := fixture.confirm(t, preview)
	if !errors.Is(err, ErrDigitalCenterLivePauseFailed) || fixture.ecos.writeCalls != 0 {
		t.Fatalf("error=%v writes=%d", err, fixture.ecos.writeCalls)
	}
}

func TestDigitalCenterConfirmWriteRechecksAddressOwnershipAfterPreview(t *testing.T) {
	fixture := newCoordinatedDigitalCenterWriteFixture(t, ECoSLiveStopped)
	preview := fixture.preview(t, []string{"address"})
	fixture.ecos.locomotives = append(fixture.ecos.locomotives,
		ECoSLocomotive{ObjectID: 2002, Name: "Other", Address: 18, Protocol: "DCC"})
	_, err := fixture.confirm(t, preview)
	var conflict *DigitalCenterAddressConflictError
	if !errors.As(err, &conflict) || fixture.ecos.writeCalls != 0 {
		t.Fatalf("error=%v writes=%d", err, fixture.ecos.writeCalls)
	}
}

type coordinatedDigitalCenterWriteFixture struct {
	*digitalCenterWriteFixture
}

func newCoordinatedDigitalCenterWriteFixture(
	t *testing.T,
	state ECoSLiveMonitorState,
) *coordinatedDigitalCenterWriteFixture {
	t.Helper()
	fixture := newDigitalCenterWriteFixture(t)
	fixture.ecos.liveStatus = ECoSLiveStatus{
		Provider: "ecos", State: state, Connected: state == ECoSLiveRunning,
	}
	fixture.ecos.locomotives = []ECoSLocomotive{
		{ObjectID: 3, Name: "Alte Lok", Address: 3, Protocol: "DCC"},
	}
	return &coordinatedDigitalCenterWriteFixture{digitalCenterWriteFixture: fixture}
}

func (fixture *coordinatedDigitalCenterWriteFixture) preview(
	t *testing.T,
	fields []string,
) DigitalCenterWritePreview {
	t.Helper()
	preview, err := fixture.service.PreviewWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWritePreviewInput{Fields: fields}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	fixture.ecos.events = nil
	return preview
}

func (fixture *coordinatedDigitalCenterWriteFixture) confirm(
	t *testing.T,
	preview DigitalCenterWritePreview,
) (DigitalCenterWriteConfirmation, error) {
	t.Helper()
	return fixture.service.ConfirmWrite(t.Context(), fixture.session.ID, fixture.item.ID,
		DigitalCenterWriteConfirmInput{Token: preview.Token, Confirm: true, Fields: preview.Fields},
		"admin-1")
}
```

Extend `digitalCenterWriteECoSStub` with these exact fields and interfaces; its existing
`SyncLocomotive` appends `dry-run` or `write`, increments `writeCalls` only for the confirmed call,
and returns `writeErr` without a retry:

```go
events     []string
liveStatus ECoSLiveStatus
resumeErr  error
pauseErr   error
writeErr   error
writeCalls int

func (stub *digitalCenterWriteECoSStub) LiveStatus() ECoSLiveStatus {
	return stub.liveStatus
}

func (stub *digitalCenterWriteECoSStub) PauseLive(context.Context) (ECoSLiveStatus, error) {
	stub.events = append(stub.events, "pause")
	if stub.pauseErr != nil {
		return stub.liveStatus, stub.pauseErr
	}
	stub.liveStatus.State = ECoSLiveStopped
	stub.liveStatus.Connected = false
	return stub.liveStatus, nil
}

func (stub *digitalCenterWriteECoSStub) StartLiveWithInterruption(
	context.Context,
	ECoSConnectionInput,
	func(),
) (*ECoSLiveStatus, error) {
	stub.events = append(stub.events, "resume")
	if stub.resumeErr != nil {
		return nil, stub.resumeErr
	}
	stub.liveStatus.State = ECoSLiveRunning
	stub.liveStatus.Connected = true
	return &stub.liveStatus, nil
}

func (stub *digitalCenterWriteECoSStub) ReadLocomotive(
	context.Context,
	ECoSConnectionInput,
	int,
) (ECoSLocomotive, error) {
	stub.events = append(stub.events, "read-target")
	return ECoSLocomotive{ObjectID: 3, Name: stub.verificationName, Address: 18, Protocol: "DCC"}, nil
}
```

- [ ] **Step 2: Run the focused tests and confirm RED / Fokustests ausführen und RED bestätigen**

Run: `cd backend; go test ./internal/application -run 'TestDigitalCenterConfirmWrite(Pauses|Returns|Keeps|Aborts|Rechecks)'`

Expected: FAIL because coordinated pause/resume and structured outcomes are not implemented.

- [ ] **Step 3: Extend result types and safe message codes / Ergebnistypen und sichere Meldungscodes erweitern**

```go
const (
	DigitalCenterWriteVerified           DigitalCenterWriteResultStatus = "verified"
	DigitalCenterWriteVerificationFailed DigitalCenterWriteResultStatus = "verification_failed"
	DigitalCenterWriteUnknown            DigitalCenterWriteResultStatus = "unknown"
	DigitalCenterWriteFailed             DigitalCenterWriteResultStatus = "failed"

	DigitalCenterMessageWriteUnknown     DigitalCenterMessageCode = "write.unknown"
	DigitalCenterMessageLiveRestartFailed DigitalCenterMessageCode = "live.restart_failed"
)

var ErrDigitalCenterLivePauseFailed = errors.New("digital center live monitor could not be paused")

type DigitalCenterWriteLiveResult struct {
	WasRunning bool `json:"wasRunning"`
	Restarted  bool `json:"restarted"`
}
```

Add the two message codes to the infrastructure allowlist. Messages remain fixed safe text and do
not include network errors.

- [ ] **Step 4: Add atomic work-item persistence / Atomare Aktualisierung des Arbeitslisteneintrags ergänzen**

Implement `UpdateWorkItem` using one `UPDATE` for name, address, protocol, compare status, station
status, JSON payloads, conflicts, and `updated_at`, guarded by `WHERE id=? AND session_id=?`. Return
`GetWorkItem` after `requireDigitalCenterUpdate`.

```go
func verifiedDigitalCenterWorkItem(
	item DigitalCenterWorkItem,
	locomotive ECoSLocomotive,
) DigitalCenterWorkItem {
	item.Name = locomotive.Name
	item.Address = locomotive.Address
	item.Protocol = locomotive.Protocol
	item.Center = map[string]any{
		"objectId": locomotive.ObjectID,
		"name": locomotive.Name,
		"decoderAddress": locomotive.Address,
		"protocol": locomotive.Protocol,
	}
	item.CompareStatus = compareDigitalCenterPayloads(item.Center, item.RailKeeper)
	item.StationStatus = "read"
	item.Conflicts = []map[string]any{}
	return item
}

func compareDigitalCenterPayloads(center map[string]any, railKeeper map[string]any) DigitalCenterCompareStatus {
	centerName, _ := digitalCenterMapString(center, "name")
	railKeeperName, _ := digitalCenterMapString(railKeeper, "name")
	centerAddress, centerAddressOK := digitalCenterMapPositiveInt(center, "decoderAddress")
	railKeeperAddress, railKeeperAddressOK := digitalCenterMapPositiveInt(railKeeper, "decoderAddress")
	centerProtocol, _ := digitalCenterMapString(center, "protocol")
	railKeeperProtocol, _ := digitalCenterMapString(railKeeper, "protocol")
	if centerName == railKeeperName && centerAddressOK && railKeeperAddressOK &&
		centerAddress == railKeeperAddress && centerProtocol == railKeeperProtocol {
		return DigitalCompareOK
	}
	return DigitalCompareDeviation
}
```

- [ ] **Step 5: Implement the coordinated confirmation / Koordinierte Bestätigung implementieren**

Refactor `ConfirmWrite` into these explicit phases under Task 2's gate:

```go
func (service *DigitalCenterWorkspaceService) confirmWriteLocked(
	ctx context.Context,
	target digitalCenterWriteTarget,
	grant DigitalCenterWriteGrant,
	actor string,
) (result DigitalCenterWriteConfirmation, returnErr error) {
	live, err := service.captureAndPauseLive(ctx, target)
	if err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	defer func() {
		service.resumeLiveAfterWrite(context.WithoutCancel(ctx), target, &live)
		result.LiveMonitor = live
	}()

	changes, err := service.previewDigitalCenterChanges(ctx, target)
	if err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	if err := service.checkAddressConflict(ctx, target, changes); err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	if err := verifyGrantHash(target, changes, grant); err != nil {
		return DigitalCenterWriteConfirmation{}, err
	}
	result, returnErr = service.applyAndVerifyDigitalCenterWrite(ctx, target, changes, live, actor)
	return result, returnErr
}
```

`applyAndVerifyDigitalCenterWrite` sends one command. Map `ErrECoSWriteStateUnknown` and a failed
verification read to `unknown` with HTTP 200. Map an explicit ECoS rejection to `failed` with HTTP
200. Do not call the writer again. A verification mismatch returns `verification_failed` and the
actual values. A match updates external mapping and work item, then returns both.

- [ ] **Step 6: Test persistence, audit, and safe messages / Persistenz, Audit und sichere Meldungen testen**

Run:

```powershell
cd backend
go test ./internal/application -run 'TestDigitalCenterConfirmWrite'
go test ./internal/infrastructure -run 'TestDigitalCenterWorkspaceRepository.*UpdateWorkItem'
go test ./internal/api -run 'TestDigitalCenterWrite'
```

Expected: PASS. Audit details contain provider, object ID, fields, and result but no token, command,
host, IP address, or locomotive name.

- [ ] **Step 7: Commit / Committen**

```powershell
git add `
  backend/internal/application/digital_center_write.go `
  backend/internal/application/digital_center_workspace_types.go `
  backend/internal/application/digital_center_ecos_operation.go `
  backend/internal/application/digital_center_write_test.go `
  backend/internal/infrastructure/digital_center_workspace_repository.go `
  backend/internal/infrastructure/digital_center_workspace_repository_test.go `
  backend/internal/api/digital_center_workspace_handlers.go `
  backend/internal/api/digital_center_write_handlers_test.go
git commit -m "feat: ECoS-Schreiben koordinieren und verifizieren / coordinate and verify ECoS writes"
```

### Task 5: OpenAPI und bedienbarer DE/EN-Schreibdialog / OpenAPI and Usable DE/EN Write Dialog

**Files:**
- Modify: `openapi/railkeeper.yaml:36-45,985-1070,11235-11355`
- Modify: `backend/internal/api/digital_center_openapi_test.go`
- Modify: `frontend/src/features/digitalCenters/digitalCenterModel.ts:1-40,175-225`
- Modify: `frontend/src/shared/api.ts:1170-1245,1800-1825`
- Modify: `frontend/src/features/digitalCenters/useDigitalCentersWorkspace.ts:470-555`
- Modify: `frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx:1-170`
- Modify: `frontend/src/features/digitalCenters/DigitalCentersFlows.test.tsx`
- Modify: `frontend/src/features/digitalCenters/useDigitalCentersWorkspace.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts:2970-3020`
- Modify: `frontend/src/shared/i18n/en.ts:2970-3020`
- Modify: `frontend/src/shared/i18n.test.ts`
- Modify: `frontend/src/styles/digital-centers.css`

**Interfaces:**
- Consumes Task 3 `Problem.details` and Task 4 confirmation structure.
- Produces strict frontend result and detail types matching OpenAPI exactly.

- [ ] **Step 1: Write failing model and flow tests / Fehlende Modell- und Ablauftests schreiben**

```tsx
it("shows a localized blocking ECoS address conflict", async () => {
  vi.mocked(api.previewDigitalCenterWrite).mockRejectedValueOnce(new ApiError(
    "decoder address conflict",
    "digital_center_address_conflict",
    409,
    { objectId: 2002, name: "Testlok B", decoderAddress: 18 }
  ));
  const user = userEvent.setup();
  render(<DigitalCentersView roles={["Admin"]} />);
  await openComparison(user);
  await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Decoder-Adresse 18 wird bereits von Testlok B (ECoS-Objekt 2002) verwendet"
  );
});

it("shows unknown write state without retrying", async () => {
  vi.mocked(api.confirmDigitalCenterWrite).mockResolvedValueOnce(confirmationFixture({
    result: "unknown", applied: false, verified: false
  }));
  const user = userEvent.setup();
  render(<DigitalCentersView roles={["Admin"]} />);
  await confirmWrite(user);
  expect(await screen.findByText("Schreibzustand unklar")).toBeInTheDocument();
  expect(screen.getByText(/nicht automatisch erneut schreiben/i)).toBeInTheDocument();
  expect(api.confirmDigitalCenterWrite).toHaveBeenCalledTimes(1);
});

it("keeps verified success and shows a separate live restart warning", async () => {
  vi.mocked(api.confirmDigitalCenterWrite).mockResolvedValueOnce(confirmationFixture({
    result: "verified", applied: true, verified: true,
    liveMonitor: { wasRunning: true, restarted: false }
  }));
  const user = userEvent.setup();
  render(<DigitalCentersView roles={["Admin"]} />);
  await confirmWrite(user);
  expect(await screen.findByText("Schreiben verifiziert")).toBeInTheDocument();
  expect(screen.getByText("Live-Monitor konnte nicht fortgesetzt werden")).toBeInTheDocument();
});

async function confirmWrite(user: ReturnType<typeof userEvent.setup>) {
  await openComparison(user);
  await user.click(screen.getByRole("button", { name: "Schreibvorschau erstellen" }));
  await user.click(await screen.findByRole("checkbox", {
    name: "Ich bestätige, dass die angezeigten Werte in die Digitalzentrale geschrieben werden."
  }));
  await user.click(screen.getByRole("button", { name: "In die Digitalzentrale schreiben" }));
}
```

- [ ] **Step 2: Run focused frontend tests and confirm RED / Fokustests ausführen und RED bestätigen**

Run: `cd frontend; npm.cmd test -- DigitalCentersFlows.test.tsx useDigitalCentersWorkspace.test.tsx --run`

Expected: FAIL because `ApiError.details`, `unknown`, live restart result, and localized states are
missing.

- [ ] **Step 3: Align OpenAPI and strict types / OpenAPI und strikte Typen angleichen**

Define these exact additions in OpenAPI and TypeScript:

```ts
export type DigitalCenterWriteResult =
  "verified" | "verification_failed" | "unknown" | "failed";

export type DigitalCenterWriteLiveResult = {
  wasRunning: boolean;
  restarted: boolean;
};

export type DigitalCenterWriteConfirmation = {
  sessionId: string;
  itemId: string;
  provider: DigitalCenterProvider;
  objectId: string;
  direction: DigitalCenterWriteDirection;
  fields: DigitalCenterWriteField[];
  applied: boolean;
  verified: boolean;
  result: DigitalCenterWriteResult;
  message: string;
  verifiedValues?: DigitalCenterStationSnapshot;
  liveMonitor: DigitalCenterWriteLiveResult;
  workItem?: DigitalCenterWorkItem;
};
```

Extend `ApiError` with `details: Record<string, unknown> | null`; parse only JSON objects, never
arrays or primitives.

- [ ] **Step 4: Implement state-safe hook updates / Zustandssichere Hook-Aktualisierung implementieren**

On address-conflict 409, clear only preview and confirmation; keep the read session and selected
item. For stale grant/session conflicts, retain the existing full read-session invalidation. When a
confirmation contains `workItem`, replace it in `selectedItem` and the current `workItems` page.
Always update `liveStatus` through the existing polling path rather than inventing a frontend live
state.

- [ ] **Step 5: Render progress and structured outcomes / Fortschritt und strukturierte Ergebnisse darstellen**

Use the existing dialog and button hierarchy. While `loading`, render an `aria-live="polite"`
status and disable close/write actions that could create ambiguous user feedback. Use result-based
localized copy instead of displaying backend prose as the primary message. Render returned actual
values for `verification_failed`; render the manual live-start instruction only when
`wasRunning && !restarted`.

Add exact DE/EN keys:

```ts
"digitalCenters.write.inProgress": "ECoS wird pausiert, geschrieben und geprüft…",
"digitalCenters.write.unknown": "Schreibzustand unklar",
"digitalCenters.write.unknownHelp": "Nicht automatisch erneut schreiben. Daten zuerst neu lesen.",
"digitalCenters.write.addressConflict": "Decoder-Adresse {address} wird bereits von {name} (ECoS-Objekt {objectId}) verwendet.",
"digitalCenters.write.liveRestartFailed": "Live-Monitor konnte nicht fortgesetzt werden",
"digitalCenters.write.liveRestartHelp": "Schreibergebnis bleibt gültig. Live-Monitor manuell neu starten."
```

```ts
"digitalCenters.write.inProgress": "Pausing ECoS, writing, and verifying…",
"digitalCenters.write.unknown": "Write state unknown",
"digitalCenters.write.unknownHelp": "Do not retry automatically. Read the data again first.",
"digitalCenters.write.addressConflict": "Decoder address {address} is already used by {name} (ECoS object {objectId}).",
"digitalCenters.write.liveRestartFailed": "Live monitor could not be resumed",
"digitalCenters.write.liveRestartHelp": "The write result remains valid. Restart the live monitor manually."
```

- [ ] **Step 6: Run API contract, frontend tests, and build / API-Vertrag, Frontendtests und Build ausführen**

```powershell
cd backend
go test ./internal/api -run 'DigitalCenter.*OpenAPI|DigitalCenterWrite'
cd ..\frontend
npm.cmd test -- DigitalCentersFlows.test.tsx useDigitalCentersWorkspace.test.tsx i18n.test.ts --run
npm.cmd run build
```

Expected: all commands PASS; TypeScript reports no unchecked cast or `any` addition.

- [ ] **Step 7: Commit / Committen**

```powershell
git add `
  openapi/railkeeper.yaml `
  backend/internal/api/digital_center_openapi_test.go `
  frontend/src/features/digitalCenters/digitalCenterModel.ts `
  frontend/src/shared/api.ts `
  frontend/src/features/digitalCenters/useDigitalCentersWorkspace.ts `
  frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx `
  frontend/src/features/digitalCenters/DigitalCentersFlows.test.tsx `
  frontend/src/features/digitalCenters/useDigitalCentersWorkspace.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  frontend/src/shared/i18n.test.ts frontend/src/styles/digital-centers.css
git commit -m "feat: sicheren ECoS-Schreibablauf anzeigen / present safe ECoS write workflow"
```

### Task 6: Realgeräte-Fixture, Kompatibilitätsmatrix und Endabnahme / Real-Device Fixture, Compatibility Matrix, and Final Acceptance

**Files:**
- Create: `backend/internal/application/testdata/ecos/real-device-success.txt`
- Create: `backend/internal/application/testdata/ecos/real-device-incomplete.txt`
- Create: `backend/internal/application/testdata/ecos/real-device-malformed.txt`
- Create: `backend/internal/application/testdata/ecos/real-device-interrupted.txt`
- Create: `backend/internal/application/ecos_real_device_fixture_test.go`
- Modify: `docs/site/de/guide/import-export/ecos-sync.md`
- Modify: `docs/site/guide/import-export/ecos-sync.md`
- Test: all backend and frontend suites.

**Interfaces:**
- Consumes the complete Tasks 1–5 workflow.
- Produces one sanitized parser fixture and one documented hardware/firmware matrix entry.

- [ ] **Step 1: Add a failing fixture regression test / Fehlenden Fixture-Regressionstest schreiben**

```go
func TestECoSRealDeviceWriteCycleFixture(t *testing.T) {
	lines := readECoSFixture(t, "testdata/ecos/real-device-success.txt")
	blocks, err := ecospkg.ParseBlocks(lines)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) < 4 {
		t.Fatalf("blocks=%d, want list, preview, write and verification replies", len(blocks))
	}
	assertSanitizedECoSFixture(t, lines)
}

func TestECoSRealDeviceFailureFixturesRemainNonSuccessful(t *testing.T) {
	for _, name := range []string{
		"real-device-incomplete.txt", "real-device-malformed.txt", "real-device-interrupted.txt",
	} {
		t.Run(name, func(t *testing.T) {
			lines := readECoSFixture(t, "testdata/ecos/"+name)
			assertSanitizedECoSFixture(t, lines)
			blocks, err := ecospkg.ParseBlocks(lines)
			if err == nil && allECoSFixtureBlocksSuccessful(blocks) {
				t.Fatalf("fixture %s was accepted as a complete successful cycle", name)
			}
		})
	}
}

func readECoSFixture(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
}

func assertSanitizedECoSFixture(t *testing.T, lines []string) {
	t.Helper()
	joined := strings.Join(lines, "\n")
	for _, forbidden := range []string{"192.168.", "10.0.", "172.16.", "token", "password"} {
		if strings.Contains(strings.ToLower(joined), strings.ToLower(forbidden)) {
			t.Fatalf("fixture contains forbidden local value %q", forbidden)
		}
	}
}

func allECoSFixtureBlocksSuccessful(blocks []ecospkg.Block) bool {
	if len(blocks) == 0 {
		return false
	}
	for _, block := range blocks {
		if !strings.Contains(block.EndLine, "(OK)") {
			return false
		}
	}
	return true
}
```

Run: `cd backend; go test ./internal/application -run TestECoSRealDeviceWriteCycleFixture`

Expected: FAIL because the fixture does not exist.

- [ ] **Step 2: Perform read-only device identification / Geräteidentifikation ausschließlich lesend durchführen**

Use the configured server-side ECoS target. Record the connection-test hardware, application, and
protocol versions. Read the work list and select no locomotive for writing yet. Confirm that raw
output shown in development diagnostics contains no credential and is not persisted.

- [ ] **Step 3: Obtain explicit target approval and preserve originals / Ziel ausdrücklich freigeben lassen und Originalwerte sichern**

Before any device mutation, ask the user to name the disposable/test locomotive. Record its ECoS
object ID, name, address, and protocol in a temporary local note outside Git. Confirm the proposed
new name, demonstrably unused address, and protocol change one operation at a time. Do not continue
if the user has not selected the target.

- [ ] **Step 4: Execute the approved real-device cycle / Freigegebenen Realgerätezyklus ausführen**

For the selected locomotive:

1. change name only and verify;
2. change to the approved free decoder address and verify;
3. change protocol separately and verify;
4. verify automatic live pause and resume;
5. attempt an approved conflicting address and verify that no command is sent;
6. restore original protocol, address, and name through the same previews.

Stop after any `unknown` result. Read again and report the observed state; never retry automatically.

- [ ] **Step 5: Sanitize and save the fixture / Fixture anonymisieren und speichern**

Save the successful list/current read, explicit write reply, and verification read in
`real-device-success.txt`. Save one bounded incomplete, malformed, and interrupted reply in the
three corresponding files. Normalize object IDs to `1001` and `1002`; replace local locomotive
names with `Testlok A` and `Testlok B`; remove host, IP, user, inventory, and filesystem data. Keep
the actual hardware and firmware versions only in the documentation matrix, not inside raw protocol
lines.

Run: `cd backend; go test ./internal/application -run TestECoSRealDeviceWriteCycleFixture`

Expected: PASS and `assertSanitizedECoSFixture` finds no IPv4/IPv6 address, hostname, token, or local
vehicle name.

- [ ] **Step 6: Update the bilingual compatibility matrix / Zweisprachige Kompatibilitätsmatrix aktualisieren**

In both ECoS guide files, add one row containing the exact hardware model, firmware/application
version, test date `2026-08-23`, and results for read, name write, address write, protocol write,
verification, live resume, and address-conflict block. State any observed deviation in both
languages. Do not mark unexecuted checks as successful.

- [ ] **Step 7: Run full verification and browser QA / Vollständige Verifikation und Browser-QA durchführen**

```powershell
cd backend
go test ./...
cd ..\frontend
npm.cmd test -- --run --reporter=dot
npm.cmd run build
```

Then verify in the local browser at desktop and narrow widths:

- address conflict remains inside the dialog and preserves the read session;
- write progress prevents repeated submission;
- verified values update the row without a full page reload;
- unknown state gives no retry action;
- live restart warning is visually separate from verified success;
- German and English text fit without clipping;
- console contains no errors or warnings.

Expected: Go tests PASS, 134 frontend files and at least 762 tests PASS, build PASS, browser console
clean.

- [ ] **Step 8: Commit / Committen**

```powershell
git add `
  backend/internal/application/testdata/ecos/real-device-success.txt `
  backend/internal/application/testdata/ecos/real-device-incomplete.txt `
  backend/internal/application/testdata/ecos/real-device-malformed.txt `
  backend/internal/application/testdata/ecos/real-device-interrupted.txt `
  backend/internal/application/ecos_real_device_fixture_test.go `
  docs/site/de/guide/import-export/ecos-sync.md `
  docs/site/guide/import-export/ecos-sync.md
git commit -m "test: ECoS-Realgeräteabnahme dokumentieren / document ECoS real-device acceptance"
```

## Final Review / Abschlussprüfung

- [ ] Compare every implementation result with
  `docs/superpowers/specs/2026-08-23-ecos-safe-write-design.md` and Issue #132.
- [ ] Run `gofmt` on every changed Go file.
- [ ] Run `git diff --check` and confirm no generated output, `frontend/dist`, `data`, `.cache`, raw
  device logs, IP addresses, or local credentials are tracked.
- [ ] Confirm `capabilitiesForProvider("ecos").WriteLocomotives == true` and
  `.WriteCVs == false` in tests and OpenAPI/UI behavior.
- [ ] Request code review, resolve findings, then push and create one bilingual PR.
