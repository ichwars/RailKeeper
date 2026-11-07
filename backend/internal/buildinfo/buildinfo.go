package buildinfo

import (
	"encoding/hex"
	"runtime"
	"runtime/debug"
)

// Release builders set these with -X. Ordinary Go builds can use embedded VCS settings.
var (
	commit   string
	kind     = "development"
	modified string
)

type Info struct {
	Commit      string `json:"commit,omitempty"`
	ShortCommit string `json:"shortCommit,omitempty"`
	Status      string `json:"status"`
	GoVersion   string `json:"goVersion"`
	Platform    string `json:"platform"`
}

func Current() Info {
	var settings []debug.BuildSetting
	if build, ok := debug.ReadBuildInfo(); ok {
		settings = build.Settings
	}
	return resolve(commit, kind, modified, settings, runtime.Version(), runtime.GOOS+"/"+runtime.GOARCH)
}

func resolve(revision, channel, dirty string, settings []debug.BuildSetting, goVersion, platform string) Info {
	if !validCommit(revision) {
		revision = ""
		for _, setting := range settings {
			if setting.Key == "vcs.revision" && validCommit(setting.Value) {
				revision = setting.Value
			}
		}
	}
	if dirty == "" {
		for _, setting := range settings {
			if setting.Key == "vcs.modified" {
				dirty = setting.Value
			}
		}
	}
	result := Info{Commit: revision, Status: "unknown", GoVersion: goVersion, Platform: platform}
	if revision != "" {
		result.ShortCommit = revision[:12]
	}
	if dirty == "true" {
		result.Status = "modified"
	} else if revision != "" && dirty == "false" && (channel == "release" || channel == "development") {
		result.Status = channel
	}
	return result
}

func validCommit(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && hex.EncodeToString(decoded) == value
}
