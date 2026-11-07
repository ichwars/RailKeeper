package backup

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzBackupFileValidation(f *testing.F) {
	for _, path := range []string{"uploads/example.csv", "uploads/a/b.pdf", "../outside", "uploads/../outside", "/uploads/absolute", "uploads/", "uploads/a\\b", "uploads/a\x00b"} {
		f.Add(path, []byte("synthetic original"))
	}
	f.Fuzz(func(t *testing.T, path string, content []byte) {
		if len(path) > 4096 || len(content) > 65536 {
			t.Skip()
		}
		file := BackupFile{Path: path, SizeBytes: int64(len(content)), ContentBase64: base64.StdEncoding.EncodeToString(content), SHA256: fmt.Sprintf("%x", sha256.Sum256(content))}
		if ValidateBackupFiles([]BackupFile{file}) != nil {
			return
		}
		normalized := filepath.ToSlash(strings.TrimSpace(path))
		relative, err := filepath.Rel("uploads", filepath.FromSlash(normalized))
		if err != nil || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(filepath.ToSlash(relative), "../") || !strings.HasPrefix(normalized, "uploads/") || normalized != PathClean(normalized) {
			t.Fatal("accepted file escaped the upload directory or retained an ambiguous path")
		}
		// Integrity failure must remain a rejection, including a zero-length original.
		file.SHA256 = strings.Repeat("0", 64)
		if ValidateBackupFiles([]BackupFile{file}) == nil {
			t.Fatal("corrupted checksum accepted")
		}
	})
}
