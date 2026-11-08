package ecosmodel

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"railkeeper/backend/internal/decoder"
)

type CachedDecoderProfile struct {
	Data, Status, Protocol, Completed, MappedAddress, MappedProtocol string
	Address                                                          int
	Source                                                           decoder.ECoSDecoderProfileEvidence
}

// ValidatedDecoderProfile omits invalid optional evidence without changing manual decoder data.
func ValidatedDecoderProfile(observation CachedDecoderProfile, maximumProtocolBytes int) *decoder.ECoSDecoderProfileEvidence {
	data, status, protocol, completed := observation.Data, observation.Status, observation.Protocol, observation.Completed
	mappedAddress, mappedProtocol, address, source := observation.MappedAddress, observation.MappedProtocol, observation.Address, observation.Source

	// A newer missing/incomplete or conflicting snapshot invalidates older context.
	if (status != "read" && status != "incomplete") || len(data) > 8<<20 {
		return nil
	}
	if mappedAddress != "" {
		mapped, err := strconv.Atoi(mappedAddress)
		if err != nil || mapped != address {
			return nil
		}
	}
	stationProtocol, err := NormalizeProtocol(protocol, maximumProtocolBytes)
	if err != nil || stationProtocol != "DCC" {
		return nil
	}
	if mappedProtocol != "" {
		mapped, err := NormalizeProtocol(mappedProtocol, maximumProtocolBytes)
		if err != nil || mapped != stationProtocol {
			return nil
		}
	}
	var snapshot struct {
		Profile string            `json:"profile"`
		Read    *ECoSCVReadState  `json:"cvRead"`
		Sources *ECoSCVSourceInfo `json:"cvSources"`
	}
	if json.Unmarshal([]byte(data), &snapshot) != nil || !utf8.ValidString(snapshot.Profile) || len(snapshot.Profile) > 160 {
		return nil
	}
	if snapshot.Sources != nil && len(snapshot.Sources.Conflicts) != 0 {
		return nil // The latest contradictory observation must not revive older model context.
	}
	source.Name = strings.Join(strings.Fields(snapshot.Profile), " ")
	source.ReadAt = completed
	if status == "incomplete" && (snapshot.Read == nil || snapshot.Read.Source != "database") {
		return nil
	}
	if snapshot.Read != nil && snapshot.Read.Source == "database" {
		if snapshot.Read.SourceHost != source.Host || snapshot.Read.NextCV != 1025 || snapshot.Read.Unavailable != 0 {
			return nil
		}
		source.ReadAt = snapshot.Read.ReadAt
	}
	if _, err := time.Parse(time.RFC3339Nano, source.ReadAt); err != nil || source.Name == "" {
		return nil
	}
	return &source
}
