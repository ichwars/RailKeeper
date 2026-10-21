package application

import (
	"cmp"
	"slices"
	"strings"
)

// Sort only the decoded entries. Sorting large metadata blobs inside SQLite can
// spill to temporary files, which are unavailable in the read-only container.
func sortMasterDataEntries(entries []MasterDataEntry) {
	slices.SortStableFunc(entries, func(a, b MasterDataEntry) int {
		if a.Active != b.Active {
			if a.Active {
				return -1
			}
			return 1
		}
		if order := cmp.Compare(a.SortOrder, b.SortOrder); order != 0 {
			return order
		}
		return strings.Compare(a.Label, b.Label)
	})
}
