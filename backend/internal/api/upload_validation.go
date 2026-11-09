package api

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func cleanOriginalFileName(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimSpace(path.Base(value))
	if value == "" || value == "." || value == "/" {
		return "beilage"
	}
	return value
}

func safeAttachmentFileName(value string) string {
	value = cleanOriginalFileName(value)
	if value == "" {
		return "beilage"
	}
	value = safeFileNamePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-")
	if value == "" {
		return "beilage"
	}
	return value
}

func confinedDataPath(dataDir, relativePath string) (string, error) {
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(base, relativePath))
	if err != nil {
		return "", err
	}
	if target != base && !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		return "", errors.New("path escapes data directory")
	}
	return target, nil
}

func isBlockedAttachmentName(value string) bool {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".dll", ".ps1", ".vbs", ".js", ".jar", ".sh":
		return true
	default:
		return false
	}
}

func isBlockedAttachmentMime(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(value, "x-msdownload") ||
		strings.Contains(value, "x-dosexec") ||
		strings.Contains(value, "x-sh") ||
		strings.Contains(value, "javascript") ||
		strings.Contains(value, "ecmascript") ||
		strings.Contains(value, "x-msdos-program")
}

func isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, allowedAttachmentExtensions)
}

func (a *App) isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, a.allowedAttachmentExtensions)
}

func isAllowedAttachmentUploadWithExtensions(filename, mimeType string, extensions map[string]struct{}) bool {
	if isBlockedAttachmentName(filename) || isBlockedAttachmentMime(mimeType) {
		return false
	}
	extension := strings.ToLower(filepath.Ext(filename))
	if _, ok := extensions[extension]; !ok {
		return false
	}
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	switch extension {
	case ".pdf":
		return mimeType == "application/pdf"
	case ".jpg", ".jpeg":
		return mimeType == "image/jpeg"
	case ".png":
		return mimeType == "image/png"
	case ".webp":
		return mimeType == "image/webp"
	case ".zip":
		return mimeType == "application/zip" || mimeType == "application/x-zip-compressed" || mimeType == "application/octet-stream"
	case ".txt", ".csv", ".json", ".xml":
		return strings.HasPrefix(mimeType, "text/") ||
			mimeType == "application/json" ||
			mimeType == "application/xml"
	default:
		return false
	}
}

func isAllowedImageMime(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}
