package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func serveCompressedAsset(w http.ResponseWriter, r *http.Request, path string) bool {
	var contentType string
	switch filepath.Ext(path) {
	case ".js":
		contentType = "text/javascript; charset=utf-8"
	case ".css":
		contentType = "text/css; charset=utf-8"
	default:
		return false
	}
	w.Header().Add("Vary", "Accept-Encoding")
	if (r.Method != http.MethodGet && r.Method != http.MethodHead) || !acceptsGzip(r.Header.Values("Accept-Encoding")) {
		return false
	}

	file, err := os.Open(path + ".gz")
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	http.ServeContent(w, r, filepath.Base(path), info.ModTime(), file)
	return true
}

func acceptsGzip(values []string) bool {
	wildcard := false
	for _, value := range values {
		for _, entry := range strings.Split(value, ",") {
			coding, parameter, hasParameter := strings.Cut(entry, ";")
			coding = strings.ToLower(strings.TrimSpace(coding))
			if coding != "gzip" && coding != "*" {
				continue
			}
			allowed := true
			if hasParameter {
				name, weight, _ := strings.Cut(strings.TrimSpace(parameter), "=")
				quality, err := strconv.ParseFloat(strings.TrimSpace(weight), 64)
				allowed = strings.EqualFold(strings.TrimSpace(name), "q") && err == nil && quality > 0 && quality <= 1
			}
			if coding == "gzip" {
				return allowed
			}
			wildcard = allowed
		}
	}
	return wildcard
}
