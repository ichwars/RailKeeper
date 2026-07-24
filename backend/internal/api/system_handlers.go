package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type storageUsageResponse struct {
	TotalBytes int64                  `json:"totalBytes"`
	Categories []storageUsageCategory `json:"categories"`
	UpdatedAt  string                 `json:"updatedAt"`
}

type storageUsageCategory struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Bytes int64  `json:"bytes"`
	Files int    `json:"files"`
}

type systemPrintersResponse struct {
	Status         string          `json:"status"`
	Message        string          `json:"message"`
	DefaultPrinter string          `json:"defaultPrinter,omitempty"`
	Printers       []systemPrinter `json:"printers"`
}

type systemPrinter struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

func (a *App) systemStorage(w http.ResponseWriter, r *http.Request) {
	categories := map[string]*storageUsageCategory{
		"database":     {Key: "database", Label: "Datenbank"},
		"images":       {Key: "images", Label: "Bilder"},
		"thumbnails":   {Key: "thumbnails", Label: "Vorschaubilder"},
		"attachments":  {Key: "attachments", Label: "Beilagen"},
		"decoderFiles": {Key: "decoderFiles", Label: "Decoder-Dateien"},
		"other":        {Key: "other", Label: "Sonstiges"},
	}

	var total int64
	err := filepath.WalkDir(a.dataDir, func(filePath string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(a.dataDir, filePath)
		if err != nil {
			relativePath = filePath
		}
		key := storageCategoryKey(relativePath)
		category := categories[key]
		category.Bytes += info.Size()
		category.Files += 1
		total += info.Size()
		return nil
	})
	if err != nil {
		a.logger.Error("storage usage scan failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "storage_usage_failed", "Speichernutzung konnte nicht gelesen werden.")
		return
	}

	order := []string{"database", "images", "thumbnails", "attachments", "decoderFiles", "other"}
	result := make([]storageUsageCategory, 0, len(order))
	for _, key := range order {
		category := categories[key]
		if category.Files == 0 && key != "database" {
			continue
		}
		result = append(result, *category)
	}

	respondJSON(w, http.StatusOK, storageUsageResponse{
		TotalBytes: total,
		Categories: result,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *App) optimizeSystemStorage(w http.ResponseWriter, r *http.Request) {
	if a.databaseMaintenance == nil {
		respondProblem(w, http.StatusInternalServerError, "storage_optimize_unavailable", "Datenbankoptimierung ist nicht konfiguriert.")
		return
	}
	result, err := a.databaseMaintenance.Optimize(r.Context())
	if err != nil {
		a.logger.Error("database optimize failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "storage_optimize_failed", "Datenbank konnte nicht optimiert werden.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) systemPrinters(w http.ResponseWriter, r *http.Request) {
	response := discoverSystemPrinters()
	respondJSON(w, http.StatusOK, response)
}

func discoverSystemPrinters() systemPrintersResponse {
	if response := printersFromEnv(); len(response.Printers) > 0 {
		return response
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		return printersFromLPStat()
	case "windows":
		return printersFromPowerShell()
	default:
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Druckerabfrage ist auf dieser Plattform nicht verfügbar. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
}

func printersFromEnv() systemPrintersResponse {
	configured := strings.TrimSpace(os.Getenv("RAILKEEPER_PRINTERS"))
	if configured == "" {
		return systemPrintersResponse{}
	}
	defaultPrinter := strings.TrimSpace(os.Getenv("RAILKEEPER_DEFAULT_PRINTER"))
	printers := []systemPrinter{}
	seen := map[string]struct{}{}
	for _, part := range strings.Split(configured, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		id := printerID(name)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		printers = append(printers, systemPrinter{
			ID:        id,
			Name:      name,
			IsDefault: defaultPrinter != "" && strings.EqualFold(defaultPrinter, name),
		})
	}
	if defaultPrinter == "" && len(printers) > 0 {
		printers[0].IsDefault = true
		defaultPrinter = printers[0].Name
	}
	return systemPrintersResponse{
		Status:         "configured",
		Message:        "Druckerliste wurde aus der RailKeeper-Konfiguration gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func printersFromLPStat() systemPrintersResponse {
	namesOut, err := exec.CommandContext(context.Background(), "lpstat", "-e").Output()
	if err != nil {
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Keine Systemdrucker im Container oder auf dem Host ermittelbar. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
	defaultPrinter := ""
	if defaultOut, err := exec.CommandContext(context.Background(), "lpstat", "-d").Output(); err == nil {
		defaultPrinter = parseLPStatDefault(string(defaultOut))
	}
	printers := printersFromNames(strings.Fields(string(namesOut)), defaultPrinter)
	return systemPrintersResponse{
		Status:         "available",
		Message:        "Systemdrucker wurden über CUPS gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func printersFromPowerShell() systemPrintersResponse {
	script := `Get-CimInstance Win32_Printer | Select-Object Name,Default | ConvertTo-Json -Compress`
	output, err := exec.CommandContext(context.Background(), "powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return systemPrintersResponse{
			Status:   "unavailable",
			Message:  "Windows-Drucker konnten nicht gelesen werden. Der Browser-Systemdialog bleibt aktiv.",
			Printers: []systemPrinter{},
		}
	}
	type windowsPrinter struct {
		Name    string `json:"Name"`
		Default bool   `json:"Default"`
	}
	var many []windowsPrinter
	if err := json.Unmarshal(output, &many); err != nil {
		var one windowsPrinter
		if err := json.Unmarshal(output, &one); err != nil {
			return systemPrintersResponse{
				Status:   "unavailable",
				Message:  "Windows-Druckerantwort konnte nicht ausgewertet werden.",
				Printers: []systemPrinter{},
			}
		}
		many = []windowsPrinter{one}
	}
	printers := []systemPrinter{}
	defaultPrinter := ""
	for _, printer := range many {
		name := strings.TrimSpace(printer.Name)
		if name == "" {
			continue
		}
		if printer.Default {
			defaultPrinter = name
		}
		printers = append(printers, systemPrinter{
			ID:        printerID(name),
			Name:      name,
			IsDefault: printer.Default,
		})
	}
	return systemPrintersResponse{
		Status:         "available",
		Message:        "Windows-Systemdrucker wurden gelesen.",
		DefaultPrinter: defaultPrinter,
		Printers:       printers,
	}
}

func parseLPStatDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if index := strings.LastIndex(value, ":"); index >= 0 {
		return strings.TrimSpace(value[index+1:])
	}
	return ""
}

func printersFromNames(names []string, defaultPrinter string) []systemPrinter {
	printers := []systemPrinter{}
	seen := map[string]struct{}{}
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		id := printerID(name)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		printers = append(printers, systemPrinter{
			ID:        id,
			Name:      name,
			IsDefault: defaultPrinter != "" && strings.EqualFold(defaultPrinter, name),
		})
	}
	return printers
}

func printerID(name string) string {
	id := strings.ToLower(strings.TrimSpace(name))
	id = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	if id == "" {
		return "printer"
	}
	return id
}

func storageCategoryKey(relativePath string) string {
	clean := filepath.ToSlash(relativePath)
	name := path.Base(clean)
	if name == "railkeeper.db" || name == "railkeeper.db-wal" || name == "railkeeper.db-shm" {
		return "database"
	}
	parts := strings.Split(clean, "/")
	if pathContains(parts, "thumbs") {
		return "thumbnails"
	}
	if pathContains(parts, "images") {
		return "images"
	}
	if pathContains(parts, "cv") {
		return "decoderFiles"
	}
	if len(parts) > 0 && parts[0] == "uploads" {
		return "attachments"
	}
	return "other"
}

func pathContains(parts []string, needle string) bool {
	for _, part := range parts {
		if part == needle {
			return true
		}
	}
	return false
}
