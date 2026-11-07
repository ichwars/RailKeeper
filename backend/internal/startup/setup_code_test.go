package startup

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSetupCodeSurvivesRestartAndIsPrivate(t *testing.T) {
	dir := t.TempDir()
	code, err := ReadOrCreateSetupCode(dir)
	if err != nil || len(code) != 43 {
		t.Fatalf("create setup code: length=%d err=%v", len(code), err)
	}
	again, err := ReadOrCreateSetupCode(dir)
	if err != nil || again != code {
		t.Fatalf("restart changed setup authorization: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, SetupCodeFile))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("code file permissions=%v", info.Mode())
	}
}

func TestSetupCodeRejectsInvalidFilesWithoutChangingTheirContent(t *testing.T) {
	for _, content := range []string{"", "invalid", strings.Repeat("a", 300)} {
		dir := t.TempDir()
		path := filepath.Join(dir, SetupCodeFile)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadOrCreateSetupCode(dir); err == nil {
			t.Fatal("invalid code file was accepted")
		}
		stored, err := os.ReadFile(path)
		if err != nil || string(stored) != content {
			t.Fatal("invalid existing file was overwritten")
		}
	}
}

func TestSetupCodeRejectsSymlinkAndPreservesItsTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs separate Windows privileges")
	}
	dir, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "preserve.txt")
	if err := os.WriteFile(target, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, SetupCodeFile)); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadOrCreateSetupCode(dir); err == nil {
		t.Fatal("symlink was accepted for setup authorization")
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "keep" {
		t.Fatal("setup modified a symlink target")
	}
}
