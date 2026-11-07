package transfer

import (
	"strings"
	"unicode"
)

// DataTransferProfileNameKey returns the stable Unicode case-fold key used by
// persistence to enforce active profile name uniqueness.
func DataTransferProfileNameKey(name string) string {
	var key strings.Builder
	for _, current := range strings.TrimSpace(name) {
		canonical := current
		for folded := unicode.SimpleFold(current); folded != current; folded = unicode.SimpleFold(folded) {
			if folded < canonical {
				canonical = folded
			}
		}
		key.WriteRune(canonical)
	}
	return key.String()
}
