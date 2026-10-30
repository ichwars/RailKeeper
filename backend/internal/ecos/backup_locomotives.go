package ecos

import (
	"bytes"
	"encoding/binary"
	"unicode"
	"unicode/utf8"
)

const (
	BackupStoredLocomotivesAvailable   = "available"
	BackupStoredLocomotivesUnavailable = "unavailable"
	backupStoredRecordLimit            = 5000
)

// BackupStoredLocomotive holds v151 station metadata, not decoder measurements.
// Only F0-F23 have stored mode slots; bits outside that range in MomentaryMask
// must not be interpreted as a known latching mode. Icon codes are not decoded.
type BackupStoredLocomotive struct {
	ObjectID      int
	Name          string
	Protocol      string
	FunctionMask  uint32
	MomentaryMask uint32
	// The legacy byte area has no verified validity/freshness information. These
	// nonzero candidates must never be merged into BackupLocomotive.Values.
	UnverifiedCVs []DatabaseCV
}

// decodeBackupStoredLocomotives supports only the fully bounded v151 layout.
// Unsupported/malformed metadata is optional: callers retain the text CV data
// and report metadata as unavailable rather than scanning for plausible records.
func decodeBackupStoredLocomotives(data []byte) ([]BackupStoredLocomotive, bool) {
	if len(data) < 52 || len(data) > backupPlaintextLimit ||
		binary.LittleEndian.Uint32(data[4:8]) != 151 {
		return nil, false
	}
	countField := binary.LittleEndian.Uint32(data[48:52])
	if countField > backupStoredRecordLimit {
		return nil, false
	}
	count := int(countField)
	position := 52
	seen := map[uint32]bool{}
	locomotives := make([]BackupStoredLocomotive, 0, count)
	for range count {
		if len(data)-position < 4 {
			return nil, false
		}
		rawType := binary.LittleEndian.Uint32(data[position : position+4])
		typ := rawType & 0x7fffffff
		position += 4
		body := position
		size, nameOffset := 0, 0
		switch typ {
		case 5: // Observed DCC station objects, including DCC28.
			size, nameOffset = 948, 28
		case 11: // MULTI consists, validated but not locomotive imports.
			if len(data)-body < 40 {
				return nil, false
			}
			members := binary.LittleEndian.Uint32(data[body+36 : body+40])
			if members > backupStoredRecordLimit {
				return nil, false
			}
			size, nameOffset = 40+4*int(members), 8
		default:
			return nil, false
		}
		if rawType&0x80000000 != 0 {
			size += 8 // Optional persistent UID; no record checksum is implied.
		}
		if len(data)-body < size {
			return nil, false
		}
		objectID := binary.LittleEndian.Uint32(data[body : body+4])
		if objectID < 1 || objectID > 2_000_000_000 || seen[objectID] {
			return nil, false
		}
		name := bytes.SplitN(data[body+nameOffset:body+nameOffset+20], []byte{0}, 2)[0]
		if len(name) == 0 || !utf8.Valid(name) {
			return nil, false
		}
		for _, character := range string(name) {
			if unicode.IsControl(character) {
				return nil, false
			}
		}
		if typ == 5 {
			locomotives = append(locomotives, backupStoredLocomotive(data[body:body+948], int(objectID), string(name)))
		}
		seen[objectID] = true
		position = body + size
	}
	if position != len(data) {
		return nil, false
	}
	return locomotives, true
}

func backupStoredLocomotive(body []byte, objectID int, name string) BackupStoredLocomotive {
	locomotive := BackupStoredLocomotive{ObjectID: objectID, Name: name, Protocol: "DCC",
		FunctionMask: binary.LittleEndian.Uint32(body[52:56]), UnverifiedCVs: []DatabaseCV{}}
	for function := range 24 {
		flags := binary.LittleEndian.Uint32(body[60+function*12 : 64+function*12])
		if flags&0x8000 != 0 {
			locomotive.MomentaryMask |= 1 << function
		}
	}
	for number := 1; number <= 255; number++ {
		value := body[0x168+number]
		if value != 0 {
			locomotive.UnverifiedCVs = append(locomotive.UnverifiedCVs, DatabaseCV{Number: number, Value: int(value)})
		}
	}
	return locomotive
}
