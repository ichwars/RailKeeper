package startup

import (
	"errors"
	"fmt"
	"strings"
)

type StorageMode string

const (
	StorageModeWindowsStandalone StorageMode = "windows_standalone"
	StorageModeConfigured        StorageMode = "configured"
	StorageModeServer            StorageMode = "server"
)

type RuntimeInputs struct {
	GOOS          string
	Args          []string
	ExecutableDir string
	WorkingDir    string
	LookupEnv     func(string) (string, bool)
	PathExists    func(string) bool
	JoinPath      func(...string) string
	AbsPath       func(string) (string, error)
}

type RuntimeConfig struct {
	Standalone              bool
	StorageMode             StorageMode
	DataDir                 string
	LegacyDataDir           string
	MigrationsDir           string
	SeedsDir                string
	StaticDir               string
	AddrDefault             string
	OpenDataFolderSupported bool
}

func ResolveRuntimeConfig(inputs RuntimeInputs) (RuntimeConfig, error) {
	if inputs.LookupEnv == nil || inputs.PathExists == nil || inputs.JoinPath == nil || inputs.AbsPath == nil {
		return RuntimeConfig{}, errors.New("runtime path dependencies are required")
	}

	standalone := standaloneRequested(inputs)
	config := RuntimeConfig{
		Standalone:    standalone,
		StorageMode:   StorageModeServer,
		AddrDefault:   ":8080",
		MigrationsDir: "./migrations",
		SeedsDir:      "./seeds",
		StaticDir:     "../../frontend/dist",
	}
	dataDir := "./data"
	if standalone {
		config.AddrDefault = "127.0.0.1:8080"
		config.MigrationsDir = inputs.JoinPath(inputs.ExecutableDir, "migrations")
		config.SeedsDir = inputs.JoinPath(inputs.ExecutableDir, "seeds")
		config.StaticDir = inputs.JoinPath(inputs.ExecutableDir, "web")
	}

	if value := configuredValue(inputs.LookupEnv, "RAILKEEPER_MIGRATIONS_DIR"); value != "" {
		config.MigrationsDir = value
	}
	if value := configuredValue(inputs.LookupEnv, "RAILKEEPER_SEEDS_DIR"); value != "" {
		config.SeedsDir = value
	}
	if value := configuredValue(inputs.LookupEnv, "RAILKEEPER_STATIC_DIR"); value != "" {
		config.StaticDir = value
	}

	if configuredData := configuredValue(inputs.LookupEnv, "RAILKEEPER_DATA_DIR"); configuredData != "" {
		dataDir = configuredData
		config.StorageMode = StorageModeConfigured
	} else if standalone && inputs.GOOS == "windows" {
		localAppData := configuredValue(inputs.LookupEnv, "LOCALAPPDATA")
		if localAppData == "" {
			return RuntimeConfig{}, errors.New("LOCALAPPDATA is required for Windows Standalone")
		}
		dataDir = inputs.JoinPath(localAppData, "RailKeeper", "data")
		config.LegacyDataDir = inputs.JoinPath(inputs.ExecutableDir, "data")
		config.StorageMode = StorageModeWindowsStandalone
		config.OpenDataFolderSupported = true
	}

	var err error
	config.DataDir, err = inputs.AbsPath(dataDir)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("resolve data directory: %w", err)
	}
	config.MigrationsDir, err = inputs.AbsPath(config.MigrationsDir)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("resolve migrations directory: %w", err)
	}
	config.SeedsDir, err = inputs.AbsPath(config.SeedsDir)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("resolve seeds directory: %w", err)
	}
	config.StaticDir, err = inputs.AbsPath(config.StaticDir)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("resolve static directory: %w", err)
	}
	if config.LegacyDataDir != "" {
		config.LegacyDataDir, err = inputs.AbsPath(config.LegacyDataDir)
		if err != nil {
			return RuntimeConfig{}, fmt.Errorf("resolve legacy data directory: %w", err)
		}
	}

	return config, nil
}

func standaloneRequested(inputs RuntimeInputs) bool {
	if strings.EqualFold(configuredValue(inputs.LookupEnv, "RAILKEEPER_PORTABLE"), "true") {
		return true
	}
	for _, arg := range inputs.Args {
		if arg == "--standalone" || arg == "--portable" {
			return true
		}
	}
	return inputs.PathExists(inputs.JoinPath(inputs.ExecutableDir, "web", "index.html")) &&
		inputs.PathExists(inputs.JoinPath(inputs.ExecutableDir, "migrations")) &&
		inputs.PathExists(inputs.JoinPath(inputs.ExecutableDir, "seeds"))
}

func configuredValue(lookup func(string) (string, bool), key string) string {
	value, ok := lookup(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}
