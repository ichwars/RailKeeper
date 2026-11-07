package buildinfo

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestBuildIdentityDoesNotMistakeVersionOrUnknownSourceForRelease(t *testing.T) {
	revision := strings.Repeat("a", 40)
	for _, test := range []struct {
		commit, kind, modified, expected string
	}{
		{revision, "release", "false", "release"},
		{revision, "development", "false", "development"},
		{revision, "release", "true", "modified"},
		{revision, "release", "", "unknown"},
		{"", "release", "false", "unknown"},
		{revision, "invalid", "false", "unknown"},
		{"untrusted/value", "release", "false", "unknown"},
	} {
		info := resolve(test.commit, test.kind, test.modified, nil, "go-test", "linux/test")
		if info.Status != test.expected || info.GoVersion != "go-test" || info.Platform != "linux/test" {
			t.Fatalf("incorrect build identity: %+v", info)
		}
		if validCommit(test.commit) && info.ShortCommit != test.commit[:12] {
			t.Fatal("short identity does not match the full source identity")
		}
		if !validCommit(test.commit) && (info.Commit != "" || info.ShortCommit != "") {
			t.Fatal("unverified source text leaked into build identity")
		}
	}
}

func TestEmbeddedVCSAndExplicitBuilderIdentity(t *testing.T) {
	old, current := strings.Repeat("a", 40), strings.Repeat("b", 64)
	settings := []debug.BuildSetting{{Key: "vcs.revision", Value: old}, {Key: "vcs.modified", Value: "true"}}
	if info := resolve("", "development", "", settings, "go-test", "test"); info.Commit != old || info.Status != "modified" {
		t.Fatalf("embedded source metadata lost: %+v", info)
	}
	if info := resolve(current, "release", "false", settings, "go-test", "test"); info.Commit != current || info.Status != "release" {
		t.Fatalf("explicit builder metadata was replaced: %+v", info)
	}
	for _, invalid := range []string{strings.Repeat("A", 40), strings.Repeat("z", 40), strings.Repeat("a", 39)} {
		if validCommit(invalid) {
			t.Fatal("invalid commit accepted")
		}
	}
}

func TestCurrentBuildAlwaysIdentifiesCompilerAndPlatform(t *testing.T) {
	info := Current()
	if info.GoVersion == "" || !strings.Contains(info.Platform, "/") || info.Status == "" {
		t.Fatal("build diagnostic metadata missing")
	}
}
