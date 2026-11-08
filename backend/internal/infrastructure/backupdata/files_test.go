package backupdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStageRestoreFilesPreservesNestedFilesAndLegacyChecksums(t *testing.T) {
	dataDir := t.TempDir()
	content := []byte{0, 1, '\r', '\n', 0xff}
	sum := sha256.Sum256(content)
	files := []BackupFile{
		{Path: "uploads/decoder/Original ä.txt", ContentBase64: base64.StdEncoding.EncodeToString(content),
			SHA256: strings.ToUpper(hex.EncodeToString(sum[:]))},
		{Path: "uploads/legacy.txt", ContentBase64: base64.StdEncoding.EncodeToString([]byte("legacy"))},
	}
	store := &Store{dataDir: dataDir}
	staged, err := store.stageRestoreFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(staged.root) })
	if staged.restoredFiles != 2 {
		t.Fatalf("restored files = %d, want 2", staged.restoredFiles)
	}
	for i, file := range files {
		path := filepath.Join(staged.root, filepath.FromSlash(file.Path))
		actual, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := content
		if i == 1 {
			want = []byte("legacy")
		}
		if !bytes.Equal(actual, want) {
			t.Fatalf("file %q changed bytes", file.Path)
		}
		if runtime.GOOS != "windows" {
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("file %q is not private: %v", file.Path, err)
			}
		}
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(staged.root)
		if err != nil || info.Mode().Perm() != 0o700 {
			t.Fatalf("staging directory is not private: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(dataDir, "uploads")); !os.IsNotExist(err) {
		t.Fatalf("staging activated uploads before restore: %v", err)
	}
}

func TestWriteStagedRestoreFileRejectsSymlinkEscapes(t *testing.T) {
	for _, name := range []string{"linked-dir/new/child.txt", "linked-dir/sentinel.txt", "linked-file.txt"} {
		t.Run(name, func(t *testing.T) {
			outside := t.TempDir()
			sentinel := filepath.Join(outside, "sentinel.txt")
			original := []byte("keep original outside bytes")
			if err := os.WriteFile(sentinel, original, 0o600); err != nil {
				t.Fatal(err)
			}
			rootDir := t.TempDir()
			for link, target := range map[string]string{"linked-dir": outside, "linked-file.txt": sentinel} {
				if err := os.Symlink(target, filepath.Join(rootDir, link)); err != nil {
					if runtime.GOOS == "windows" {
						t.Skipf("Windows symlink privilege unavailable: %v", err)
					}
					t.Fatal(err)
				}
			}
			root, err := os.OpenRoot(rootDir)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = root.Close() })
			file := BackupFile{Path: "uploads/" + name, ContentBase64: base64.StdEncoding.EncodeToString([]byte("overwrite"))}
			if err := writeStagedRestoreFile(root, file); err == nil {
				t.Fatal("restore followed a symlink outside its staging root")
			}
			actual, err := os.ReadFile(sentinel)
			if err != nil || !bytes.Equal(actual, original) {
				t.Fatalf("outside sentinel changed: %v", err)
			}
			entries, err := os.ReadDir(outside)
			if err != nil || len(entries) != 1 {
				t.Fatalf("restore created directories outside staging: %v", err)
			}
		})
	}
}
