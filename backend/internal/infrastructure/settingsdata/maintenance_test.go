package settingsdata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseFileBytesRetainsMissingFilesAndOriginalContents(t *testing.T) {
	root := t.TempDir()
	store := NewMaintenance(nil, root)
	files := map[string]string{"railkeeper.db": "original database", "railkeeper.db-wal": "original log", "railkeeper.db-shm": "original shared memory"}
	var expected int64
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		expected += int64(len(content))
	}
	if size, err := store.DatabaseFileBytes(); err != nil || size != expected {
		t.Fatal("database/WAL/SHM size changed", size, expected, err)
	}
	for name, content := range files {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || string(data) != content {
			t.Fatal("measurement changed original file", name, err)
		}
	}
	if err := os.Remove(filepath.Join(root, "railkeeper.db-wal")); err != nil {
		t.Fatal(err)
	}
	expected -= int64(len(files["railkeeper.db-wal"]))
	if size, err := store.DatabaseFileBytes(); err != nil || size != expected {
		t.Fatal("missing optional WAL no longer accepted", size, err)
	}
}

func TestDatabaseFileBytesKeepsAbsentDirectoryReadOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	if size, err := NewMaintenance(nil, root).DatabaseFileBytes(); err != nil || size != 0 {
		t.Fatal("missing directory contract changed", size, err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatal("measurement created missing directory", err)
	}
}

func TestDatabaseFileBytesRejectsRegularFileAsDataDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-a-directory")
	const original = "retain this original file"
	if err := os.WriteFile(root, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if size, err := NewMaintenance(nil, root).DatabaseFileBytes(); err == nil || size != 0 {
		t.Fatal("invalid storage root accepted", size, err)
	}
	data, err := os.ReadFile(root)
	if err != nil || string(data) != original {
		t.Fatal("invalid-root rejection changed original bytes", err)
	}
}
