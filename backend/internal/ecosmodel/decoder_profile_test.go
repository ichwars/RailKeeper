package ecosmodel

import (
	"strings"
	"testing"

	"railkeeper/backend/internal/decoder"
)

func TestCachedDecoderProfileRequiresCompleteConsistentEvidence(t *testing.T) {
	original := CachedDecoderProfile{
		Data:   `{"profile":"  LokPilot   5 ","cvRead":{"source":"database","sourceHost":"saved.test","nextCV":1025,"unavailable":0,"readAt":"2026-10-27T07:01:00Z"}}`,
		Status: "incomplete", Protocol: "DCC128", Completed: "2026-10-27T07:00:00Z",
		MappedAddress: "3", MappedProtocol: "DCC 128", Address: 3,
		Source: decoder.ECoSDecoderProfileEvidence{ObjectID: "1002", Host: "saved.test"},
	}
	value := ValidatedDecoderProfile(original, 32)
	if value == nil || value.Name != "LokPilot 5" || value.ReadAt != "2026-10-27T07:01:00Z" || value.Host != "saved.test" || value.ObjectID != "1002" {
		t.Fatal("complete saved reading lost its exact provenance", value)
	}
	for _, mode := range []string{"missing", "address", "bad-address", "protocol", "mapped-protocol", "malformed", "oversized", "partial", "source-host", "source-time", "unavailable", "conflicting", "empty-profile"} {
		t.Run(mode, func(t *testing.T) {
			observation := original
			switch mode {
			case "missing":
				observation.Status = "missing"
			case "address":
				observation.MappedAddress = "4"
			case "bad-address":
				observation.MappedAddress = "unknown"
			case "protocol":
				observation.Protocol = "MM"
			case "mapped-protocol":
				observation.MappedProtocol = "MM"
			case "malformed":
				observation.Data = "{"
			case "oversized":
				observation.Data = strings.Repeat(" ", 8<<20+1)
			case "partial":
				observation.Data = strings.Replace(original.Data, `"nextCV":1025`, `"nextCV":1024`, 1)
			case "source-host":
				observation.Data = strings.Replace(original.Data, `"sourceHost":"saved.test"`, `"sourceHost":"foreign.test"`, 1)
			case "source-time":
				observation.Data = strings.Replace(original.Data, "2026-10-27T07:01:00Z", "unknown", 1)
			case "unavailable":
				observation.Data = strings.Replace(original.Data, `"unavailable":0`, `"unavailable":1`, 1)
			case "conflicting":
				observation.Data = `{"profile":"newer disputed model","cvSources":{"conflicts":[{"number":8}]}}`
			case "empty-profile":
				observation.Data = `{"profile":" "}`
			}
			if got := ValidatedDecoderProfile(observation, 32); got != nil {
				t.Fatal("inconsistent optional evidence was accepted", mode, got)
			}
		})
	}
	if got := ValidatedDecoderProfile(original, 32); got == nil || *got != *value {
		t.Fatal("rejected evidence changed an earlier independent observation")
	}
}
