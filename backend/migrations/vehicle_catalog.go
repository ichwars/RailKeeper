package migrations

import _ "embed"

// LegacyVehicleCatalogSQL is also used inside restore's transaction for older backups.
//
//go:embed 0074_vehicle_catalog_legacy.sql
var LegacyVehicleCatalogSQL string
