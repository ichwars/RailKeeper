package api

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestStaticAssetCompressionNegotiation(t *testing.T) {
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.Mkdir(assets, 0o700); err != nil {
		t.Fatal(err)
	}
	content := []byte(strings.Repeat("console.log('railkeeper');\n", 64))
	var compressed bytes.Buffer
	encoder := gzip.NewWriter(&compressed)
	if _, err := encoder.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{"main-hash.js": content, "main-hash.js.gz": compressed.Bytes(), "old-hash.js": content} {
		if err := os.WriteFile(filepath.Join(assets, name), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	handler := staticHandler(root)
	request := func(method, name, accept string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/assets/"+name, nil)
		if accept != "" {
			r.Header.Set("Accept-Encoding", accept)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != http.StatusOK || w.Header().Get("Vary") != "Accept-Encoding" ||
			w.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
			t.Fatalf("%s %s: HTTP %d, headers %v", method, name, w.Code, w.Header())
		}
		return w
	}

	for _, accept := range []string{"gzip", "br, GZip;q=0.5", "*;q=0.4"} {
		w := request(http.MethodGet, "main-hash.js", accept)
		if w.Header().Get("Content-Encoding") != "gzip" || !strings.Contains(w.Header().Get("Content-Type"), "javascript") ||
			w.Header().Get("Content-Length") != strconv.Itoa(compressed.Len()) {
			t.Fatalf("%q: headers %v", accept, w.Header())
		}
		decoder, err := gzip.NewReader(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := io.ReadAll(decoder)
		_ = decoder.Close()
		if err != nil || !bytes.Equal(decoded, content) {
			t.Fatalf("%q: decoded bytes differ, error %v", accept, err)
		}
	}

	for _, accept := range []string{"", "br", "*;q=1, gzip;q=0", "gzip;q=invalid"} {
		w := request(http.MethodGet, "main-hash.js", accept)
		if w.Header().Get("Content-Encoding") != "" || !bytes.Equal(w.Body.Bytes(), content) {
			t.Fatalf("%q: expected uncompressed content", accept)
		}
	}
	old := request(http.MethodGet, "old-hash.js", "gzip")
	if old.Header().Get("Content-Encoding") != "" || !bytes.Equal(old.Body.Bytes(), content) {
		t.Fatal("builds without a compressed file must still be served")
	}
	head := request(http.MethodHead, "main-hash.js", "gzip")
	if head.Header().Get("Content-Encoding") != "gzip" || head.Body.Len() != 0 {
		t.Fatal("HEAD must describe the negotiated representation without a body")
	}
}

func TestStaticAssetsStayInsideWebRoot(t *testing.T) {
	root := t.TempDir()
	assets := filepath.Join(root, "assets")
	if err := os.Mkdir(assets, 0o700); err != nil {
		t.Fatal(err)
	}
	const index = "safe index"
	const script = "safe script"
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte(index), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "main.js"), []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	private := filepath.Join(t.TempDir(), "private")
	if err := os.WriteFile(private, []byte("private file"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"main.js.gz", "outside.js"} {
		if err := os.Symlink(private, filepath.Join(assets, name)); err != nil {
			if runtime.GOOS == "windows" {
				t.Skipf("symlink creation is unavailable: %v", err)
			}
			t.Fatal(err)
		}
	}

	handler := staticHandler(root)
	for _, test := range []struct{ path, body string }{
		{"/assets/main.js", script},
		{"/assets/main.js.gz", index},
		{"/assets/outside.js", index},
	} {
		t.Run(test.path, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, test.path, nil)
			r.Header.Set("Accept-Encoding", "gzip")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != http.StatusOK || w.Header().Get("Content-Encoding") != "" || w.Body.String() != test.body {
				t.Fatalf("unexpected response: status %d, encoding %q, body %q", w.Code,
					w.Header().Get("Content-Encoding"), w.Body.String())
			}
		})
	}
}
