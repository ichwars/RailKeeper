package api

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/safefetch"
)

func effectiveLimit(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}

var safeFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

var allowedAttachmentExtensions = map[string]struct{}{
	".csv":  {},
	".jpeg": {},
	".jpg":  {},
	".json": {},
	".pdf":  {},
	".png":  {},
	".txt":  {},
	".webp": {},
	".xml":  {},
	".zip":  {},
}

func effectiveAttachmentExtensions(input map[string]struct{}) map[string]struct{} {
	if len(input) == 0 {
		return allowedAttachmentExtensions
	}
	out := map[string]struct{}{}
	for extension := range input {
		extension = strings.ToLower(strings.TrimSpace(extension))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		if isBlockedAttachmentName("file" + extension) {
			continue
		}
		out[extension] = struct{}{}
	}
	if len(out) == 0 {
		return allowedAttachmentExtensions
	}
	return out
}

func remoteImageFileName(image application.VehicleImageInput, mimeType string) string {
	extension := ".jpg"
	switch mimeType {
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	}
	base := strings.TrimSpace(image.Title)
	if base == "" {
		if parsed, err := url.Parse(image.URL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" || base == "." || base == "/" {
		base = "artikelbild"
	}
	return safeAttachmentFileName(base + extension)
}

func remoteAttachmentFileName(input importVehicleAttachmentInput, rawURL, mimeType string) string {
	base := strings.TrimSpace(input.Title)
	if base == "" {
		if parsed, err := url.Parse(rawURL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	if base == "" || base == "." || base == "/" {
		base = "dokument"
	}
	extension := strings.ToLower(filepath.Ext(base))
	if extension == "" {
		extension = attachmentExtensionForMime(mimeType)
		base += extension
	}
	return safeAttachmentFileName(base)
}

func attachmentExtensionForMime(mimeType string) string {
	switch strings.ToLower(strings.Split(mimeType, ";")[0]) {
	case "application/pdf":
		return ".pdf"
	case "application/json", "text/json":
		return ".json"
	case "application/xml", "text/xml":
		return ".xml"
	case "application/zip", "application/x-zip-compressed":
		return ".zip"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "text/csv":
		return ".csv"
	default:
		return ".txt"
	}
}

func attachmentCategoryForRemoteDocument(fileName, title string) string {
	lower := strings.ToLower(fileName + " " + title)
	if strings.Contains(lower, "ersatzteil") || strings.Contains(lower, "spare") || strings.Contains(lower, "et-blatt") {
		return "Ersatzteilliste"
	}
	if strings.Contains(lower, "anleitung") || strings.Contains(lower, "manual") || strings.Contains(lower, "bedienung") {
		return "Anleitung"
	}
	return "Dokumentation"
}

func isPublicImageURL(ctx context.Context, value string) bool {
	return safefetch.IsPublicHTTPURL(ctx, value)
}

func remoteDocumentHTTPClient(ctx context.Context) *http.Client {
	return safefetch.NewHTTPClient(ctx, safefetch.Options{Timeout: 10 * time.Second, MaxRedirects: 5})
}
