package api

import (
	"os"
	"strings"
	"testing"
)

func TestRouterCoreStaysFocused(t *testing.T) {
	data, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	source := string(data)
	if lines := strings.Count(source, "\n") + 1; lines > 400 {
		t.Fatalf("router.go has %d lines, want at most 400", lines)
	}
	for _, handler := range []string{
		"func (a *App) versionInfo",
		"func (a *App) login",
		"func (a *App) listVehicles",
		"func (a *App) exportBackup",
	} {
		if strings.Contains(source, handler) {
			t.Errorf("router.go still contains handler %q", handler)
		}
	}
}
