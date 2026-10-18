// Package migrations exposes the idempotent legacy data conversion used during backup restore.
package migrations

import _ "embed"

// LegacyLayoutRetirementSQL preserves original allocation records and returns retired stock.
//
//go:embed 0069_retire_layout_planning.sql
var LegacyLayoutRetirementSQL string
