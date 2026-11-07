package migrations

import (
	_ "embed"
	"errors"
	"strings"
)

// LegacyVehicleCatalogSQL is also used inside restore's transaction for older backups.
//
//go:embed 0074_vehicle_catalog_legacy.sql
var LegacyVehicleCatalogSQL string

//go:embed 0075_vehicle_catalog_authority.sql
var VehicleCatalogAuthoritySQL string

// VehicleCatalogProjectionSQL excludes the one-time column/data migration.
// Restore recreates these exact trigger definitions inside its data transaction.
func VehicleCatalogProjectionSQL() (string, error) {
	_, definitions, ok := strings.Cut(VehicleCatalogAuthoritySQL, "-- BEGIN catalog projection triggers\n")
	if !ok {
		return "", errors.New("catalog projection definitions are absent from migration")
	}
	return definitions, nil
}
