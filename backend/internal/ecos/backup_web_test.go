package ecos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBackupWebUsesOnlyBoundedSaveGET(t *testing.T) {
	oversized := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.RequestURI() != "/cgi/save?lang=en" || r.Header.Get("Cache-Control") != "no-cache" {
			t.Errorf("unexpected save request: %s %s", r.Method, r.URL.RequestURI())
		}
		if oversized {
			_, _ = w.Write([]byte(strings.Repeat("x", EncryptedBackupLimit+1)))
			return
		}
		_, _ = w.Write([]byte("synthetic encrypted bytes"))
	}))
	defer server.Close()
	reader := &CVDatabaseReader{web: &ImageWeb{Host: "192.168.2.10", base: server.URL, client: server.Client()}}
	data, err := reader.ReadBackup(context.Background())
	if err != nil || string(data) != "synthetic encrypted bytes" || reader.Address() != "192.168.2.10" {
		t.Fatalf("backup changed its pinned address or bytes: %v", err)
	}
	oversized = true
	if _, err := reader.ReadBackup(context.Background()); err == nil {
		t.Fatal("accepted oversized stored backup")
	}
}
