package startup

import (
	"errors"
	"strings"
	"testing"
)

func TestResolveRuntimeConfigSelectsSafeDataPaths(t *testing.T) {
	tests := []struct {
		name           string
		goos           string
		args           []string
		env            map[string]string
		bundled        bool
		wantStandalone bool
		wantMode       StorageMode
		wantData       string
		wantLegacy     string
		wantMigrations string
		wantSeeds      string
		wantStatic     string
		wantAddr       string
		wantOpenFolder bool
	}{
		{
			name:           "new Windows standalone",
			goos:           "windows",
			args:           []string{"--standalone"},
			env:            map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`},
			wantStandalone: true,
			wantMode:       StorageModeWindowsStandalone,
			wantData:       `C:\Users\Ada\AppData\Local\RailKeeper\data`,
			wantLegacy:     `C:\RailKeeper\data`,
			wantMigrations: `C:\RailKeeper\migrations`,
			wantSeeds:      `C:\RailKeeper\seeds`,
			wantStatic:     `C:\RailKeeper\web`,
			wantAddr:       "127.0.0.1:8080",
			wantOpenFolder: true,
		},
		{
			name:           "legacy portable flag keeps the safe default",
			goos:           "windows",
			args:           []string{"--portable"},
			env:            map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`},
			wantStandalone: true,
			wantMode:       StorageModeWindowsStandalone,
			wantData:       `C:\Users\Ada\AppData\Local\RailKeeper\data`,
			wantLegacy:     `C:\RailKeeper\data`,
			wantMigrations: `C:\RailKeeper\migrations`,
			wantSeeds:      `C:\RailKeeper\seeds`,
			wantStatic:     `C:\RailKeeper\web`,
			wantAddr:       "127.0.0.1:8080",
			wantOpenFolder: true,
		},
		{
			name:           "packaged layout is detected without a flag",
			goos:           "windows",
			env:            map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`},
			bundled:        true,
			wantStandalone: true,
			wantMode:       StorageModeWindowsStandalone,
			wantData:       `C:\Users\Ada\AppData\Local\RailKeeper\data`,
			wantLegacy:     `C:\RailKeeper\data`,
			wantMigrations: `C:\RailKeeper\migrations`,
			wantSeeds:      `C:\RailKeeper\seeds`,
			wantStatic:     `C:\RailKeeper\web`,
			wantAddr:       "127.0.0.1:8080",
			wantOpenFolder: true,
		},
		{
			name: "explicit data directory wins",
			goos: "windows",
			args: []string{"--standalone"},
			env: map[string]string{
				"LOCALAPPDATA":        `C:\Users\Ada\AppData\Local`,
				"RAILKEEPER_DATA_DIR": `D:\RailKeeperUSB\data`,
			},
			wantStandalone: true,
			wantMode:       StorageModeConfigured,
			wantData:       `D:\RailKeeperUSB\data`,
			wantLegacy:     "",
			wantMigrations: `C:\RailKeeper\migrations`,
			wantSeeds:      `C:\RailKeeper\seeds`,
			wantStatic:     `C:\RailKeeper\web`,
			wantAddr:       "127.0.0.1:8080",
			wantOpenFolder: false,
		},
		{
			name:           "server defaults remain inside the working directory",
			goos:           "linux",
			env:            map[string]string{},
			wantStandalone: false,
			wantMode:       StorageModeServer,
			wantData:       `/srv/railkeeper/data`,
			wantLegacy:     "",
			wantMigrations: `/srv/railkeeper/migrations`,
			wantSeeds:      `/srv/railkeeper/seeds`,
			wantStatic:     `/srv/frontend/dist`,
			wantAddr:       ":8080",
			wantOpenFolder: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			windows := tt.goos == "windows"
			config, err := ResolveRuntimeConfig(RuntimeInputs{
				GOOS:          tt.goos,
				Args:          tt.args,
				ExecutableDir: `C:\RailKeeper`,
				WorkingDir:    `/srv/railkeeper/backend`,
				LookupEnv:     mapLookup(tt.env),
				PathExists: func(path string) bool {
					if !tt.bundled {
						return false
					}
					return path == `C:\RailKeeper\web\index.html` ||
						path == `C:\RailKeeper\migrations` || path == `C:\RailKeeper\seeds`
				},
				JoinPath: pathJoin(windows),
				AbsPath:  absolutePath(windows),
			})
			if err != nil {
				t.Fatalf("ResolveRuntimeConfig() error = %v", err)
			}
			if config.Standalone != tt.wantStandalone || config.StorageMode != tt.wantMode {
				t.Fatalf("runtime mode = standalone:%v storage:%q, want standalone:%v storage:%q",
					config.Standalone, config.StorageMode, tt.wantStandalone, tt.wantMode)
			}
			if config.DataDir != tt.wantData || config.LegacyDataDir != tt.wantLegacy {
				t.Fatalf("data paths = %q, %q, want %q, %q",
					config.DataDir, config.LegacyDataDir, tt.wantData, tt.wantLegacy)
			}
			if config.MigrationsDir != tt.wantMigrations || config.SeedsDir != tt.wantSeeds ||
				config.StaticDir != tt.wantStatic {
				t.Fatalf("runtime paths = migrations:%q seeds:%q static:%q",
					config.MigrationsDir, config.SeedsDir, config.StaticDir)
			}
			if config.AddrDefault != tt.wantAddr ||
				config.OpenDataFolderSupported != tt.wantOpenFolder {
				t.Fatalf("runtime capabilities = addr:%q openFolder:%v",
					config.AddrDefault, config.OpenDataFolderSupported)
			}
		})
	}
}

func TestResolveRuntimeConfigRefusesMissingLocalAppData(t *testing.T) {
	_, err := ResolveRuntimeConfig(RuntimeInputs{
		GOOS:          "windows",
		Args:          []string{"--standalone"},
		ExecutableDir: `C:\RailKeeper`,
		WorkingDir:    `C:\RailKeeper`,
		LookupEnv:     mapLookup(map[string]string{}),
		PathExists:    func(string) bool { return false },
		JoinPath:      pathJoin(true),
		AbsPath:       absolutePath(true),
	})
	if err == nil || !strings.Contains(err.Error(), "LOCALAPPDATA") {
		t.Fatalf("expected LOCALAPPDATA error, got %v", err)
	}
}

func TestResolveRuntimeConfigUsesConfiguredDataWithoutLocalAppData(t *testing.T) {
	config, err := ResolveRuntimeConfig(RuntimeInputs{
		GOOS:          "windows",
		Args:          []string{"--standalone"},
		ExecutableDir: `C:\RailKeeper`,
		WorkingDir:    `C:\RailKeeper`,
		LookupEnv: mapLookup(map[string]string{
			"RAILKEEPER_DATA_DIR": `D:\RailKeeperUSB\data`,
		}),
		PathExists: func(string) bool { return false },
		JoinPath:   pathJoin(true),
		AbsPath:    absolutePath(true),
	})
	if err != nil {
		t.Fatalf("ResolveRuntimeConfig() error = %v", err)
	}
	if config.StorageMode != StorageModeConfigured || config.DataDir != `D:\RailKeeperUSB\data` {
		t.Fatalf("configured data path not preserved: %#v", config)
	}
}

func mapLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func pathJoin(windows bool) func(...string) string {
	separator := "/"
	if windows {
		separator = `\`
	}
	return func(parts ...string) string {
		cleaned := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.Trim(part, `/\`)
			if part != "" {
				cleaned = append(cleaned, part)
			}
		}
		if len(cleaned) == 0 {
			return ""
		}
		prefix := ""
		if strings.HasPrefix(parts[0], "/") {
			prefix = "/"
		}
		return prefix + strings.Join(cleaned, separator)
	}
}

func absolutePath(windows bool) func(string) (string, error) {
	return func(path string) (string, error) {
		if strings.TrimSpace(path) == "" {
			return "", errors.New("empty path")
		}
		if windows {
			if len(path) >= 3 && path[1] == ':' {
				return path, nil
			}
			return `C:\RailKeeper\` + strings.Trim(path, `/\`), nil
		}
		if strings.HasPrefix(path, "/") {
			return path, nil
		}
		if path == "../../frontend/dist" {
			return "/srv/frontend/dist", nil
		}
		return "/srv/railkeeper/" + strings.TrimPrefix(path, "./"), nil
	}
}
