package backup

import (
	"context"
	"errors"
	"testing"
)

type memoryStore struct {
	snapshot     Snapshot
	schemaError  error
	restoreCalls int
}

func (*memoryStore) Configured() bool { return true }
func (m *memoryStore) Schema(_ context.Context, tables []string) (map[string]map[string]bool, error) {
	result := map[string]map[string]bool{}
	for _, table := range tables {
		result[table] = map[string]bool{"id": true}
	}
	return result, m.schemaError
}
func (m *memoryStore) Snapshot(context.Context) (Snapshot, error) { return m.snapshot, nil }
func (*memoryStore) ExportFiles() ([]BackupFile, error)           { return []BackupFile{}, nil }
func (m *memoryStore) Restore(context.Context, *BackupDocument) (*BackupImportResult, error) {
	m.restoreCalls++
	return &BackupImportResult{}, nil
}

type testRules struct{ problem string }

func (testRules) Table(*BackupDocument, string, []map[string]any, []string) ([]string, []string) {
	return nil, nil
}
func (r testRules) Document(*BackupDocument) []string {
	if r.problem != "" {
		return []string{r.problem}
	}
	return nil
}

func validDocument(version int) *BackupDocument {
	doc := &BackupDocument{Format: Format, Version: version, Tables: map[string][]map[string]any{}, Profiles: []map[string]any{}, Files: []BackupFile{}}
	for _, table := range TableOrder() {
		doc.Tables[table] = []map[string]any{}
	}
	return doc
}

func TestInvalidInputNeverReachesTheRestorePort(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*BackupDocument)
	}{
		{"format", func(d *BackupDocument) { d.Format = "other" }},
		{"future version", func(d *BackupDocument) { d.Version = Version + 1 }},
		{"missing table", func(d *BackupDocument) { delete(d.Tables, "vehicles") }},
		{"corrupt file", func(d *BackupDocument) { d.Files = []BackupFile{{Path: "uploads/demo.txt", ContentBase64: "invalid"}} }},
		{"path escape", func(d *BackupDocument) { d.Files = []BackupFile{{Path: "uploads/../auth.db"}} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &memoryStore{}
			doc := validDocument(Version)
			test.edit(doc)
			if _, err := New(store, testRules{}).Import(t.Context(), doc); err == nil {
				t.Fatal("invalid input accepted")
			}
			if store.restoreCalls != 0 {
				t.Fatal("invalid input reached destructive storage")
			}
		})
	}
}

func TestSchemaAndDomainFailureDoNotStartRestore(t *testing.T) {
	for _, problem := range []struct {
		store *memoryStore
		rules testRules
	}{
		{&memoryStore{schemaError: errors.New("schema unavailable")}, testRules{}},
		{&memoryStore{}, testRules{problem: "invalid immutable decoder snapshot"}},
	} {
		if _, err := New(problem.store, problem.rules).Import(t.Context(), validDocument(Version)); err == nil {
			t.Fatal("validation failure ignored")
		}
		if problem.store.restoreCalls != 0 {
			t.Fatal("destructive restore started before complete validation")
		}
	}
}

func TestLegacyValuationAbsenceRemainsCompatible(t *testing.T) {
	doc := validDocument(27)
	delete(doc.Tables, "vehicle_valuations")
	store := &memoryStore{}
	result, err := New(store, testRules{}).Validate(t.Context(), doc)
	if err != nil || !result.Compatible || len(result.Warnings) == 0 {
		t.Fatalf("legacy backup rejected: %v %#v", err, result)
	}
	if _, err := New(store, testRules{}).Import(t.Context(), doc); err != nil || store.restoreCalls != 1 {
		t.Fatal("compatible restore did not reach its owner")
	}
}

func TestExportNeverIncludesAuthenticationFromAnAdapter(t *testing.T) {
	doc := validDocument(Version)
	doc.Tables["users"] = []map[string]any{{"id": "synthetic-auth-only"}}
	store := &memoryStore{snapshot: Snapshot{Tables: doc.Tables, Profiles: doc.Profiles}}
	exported, err := New(store, testRules{}).Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := exported.Tables["users"]; exists {
		t.Fatal("authentication escaped the backup boundary")
	}
	if len(exported.Tables) != len(TableOrder()) {
		t.Fatal("application table omitted")
	}
	delete(store.snapshot.Tables, "vehicles")
	if _, err := New(store, testRules{}).Export(t.Context()); err == nil {
		t.Fatal("missing snapshot table silently exported")
	}
}
