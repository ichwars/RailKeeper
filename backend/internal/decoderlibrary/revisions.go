package decoderlibrary

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
)

//go:embed previous-bundled-digests.json
var previousBundledJSON []byte

// PreviousBundled recognizes only exact released metadata, never a locally edited derivative.
func PreviousBundled(typeName, key, metadata string) bool {
	var known map[string][]string
	var document map[string]any
	if json.Unmarshal(previousBundledJSON, &known) != nil || json.Unmarshal([]byte(metadata), &document) != nil {
		return false
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return false
	}
	digest := sha256.Sum256(canonical)
	for _, expected := range known[typeName+":"+key] {
		if hex.EncodeToString(digest[:]) == expected {
			return true
		}
	}
	return false
}
