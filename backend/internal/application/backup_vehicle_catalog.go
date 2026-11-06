package application

import (
	"fmt"
	"strconv"
	"time"
)

var vehicleCatalogBackupTables = []string{
	"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links",
}

func isVehicleCatalogBackupTable(table string) bool {
	for _, known := range vehicleCatalogBackupTables {
		if table == known {
			return true
		}
	}
	return false
}

func validateBackupVehicleCatalog(doc *BackupDocument) []string {
	var problems []string
	for _, table := range vehicleCatalogBackupTables {
		if doc.Version < 25 && len(doc.Tables[table]) > 0 {
			return []string{"Artikeldaten benötigen Backupformat 25 oder neuer. / Catalog data requires backup format 25 or newer."}
		}
	}
	identities := func(table, column string) map[string]bool {
		ids := map[string]bool{}
		for index, row := range doc.Tables[table] {
			id, valid := backupNonEmptyString(row[column])
			if !valid || ids[id] {
				problems = append(problems, fmt.Sprintf("%s Zeile %d: fehlende oder doppelte Identität. / Missing or duplicate identity.", table, index+1))
			}
			ids[id] = true
		}
		return ids
	}
	articles := identities("catalog_articles", "id")
	revisions := identities("catalog_article_revisions", "id")
	vehicles, sets := map[string]bool{}, map[string]bool{}
	for _, row := range doc.Tables["vehicles"] {
		id, _ := row["id"].(string)
		vehicles[id] = true
	}
	for _, row := range doc.Tables["vehicle_sets"] {
		id, _ := row["id"].(string)
		sets[id] = true
	}
	timestamp := func(row map[string]any, key, label string) {
		value, ok := row[key].(string)
		if _, err := time.Parse(time.RFC3339Nano, value); !ok || err != nil {
			problems = append(problems, label+": ungültiger Zeitpunkt. / Invalid timestamp.")
		}
	}
	sources := map[string]bool{}
	for _, row := range doc.Tables["catalog_articles"] {
		origin, _ := row["origin"].(string)
		source, validSource := backupNonEmptyString(row["source_entity_id"])
		key := origin + ":" + source
		if (origin != "manual" && origin != "legacy_vehicle" && origin != "legacy_set") ||
			(origin == "manual" && row["source_entity_id"] != nil) ||
			(origin != "manual" && (!validSource || sources[key])) {
			problems = append(problems, "Katalogartikel: ungültige oder doppelte Herkunft. / Invalid or duplicate catalog origin.")
		}
		if origin != "manual" {
			sources[key] = true
		}
		timestamp(row, "created_at", "Katalogartikel")
	}
	numbers := map[string]bool{}
	for _, row := range doc.Tables["catalog_article_revisions"] {
		articleID, _ := row["article_id"].(string)
		number, err := strconv.ParseInt(fmt.Sprint(row["revision_number"]), 10, 32)
		key := fmt.Sprintf("%s:%d", articleID, number)
		if !articles[articleID] || err != nil || number < 1 || numbers[key] {
			problems = append(problems, "Artikelrevision: ungültige Artikelbeziehung oder Revisionsnummer. / Invalid article relationship or revision number.")
		}
		numbers[key] = true
		fields, ok := row["fields_json"].(string)
		if _, err := decodeVehicleCatalogFields(fields, true); !ok || err != nil {
			problems = append(problems, "Artikelrevision: unvollständige oder unbekannte Katalogfelder. / Incomplete or unknown catalog fields.")
		}
		if _, ok := row["note"].(string); !ok {
			problems = append(problems, "Artikelrevision: Revisionsnotiz fehlt. / Revision note is missing.")
		}
		timestamp(row, "created_at", "Artikelrevision")
	}
	for _, relation := range []struct {
		table, targetColumn string
		targets             map[string]bool
	}{
		{"vehicle_article_links", "vehicle_id", vehicles}, {"vehicle_set_article_links", "vehicle_set_id", sets},
	} {
		identities(relation.table, relation.targetColumn)
		for _, row := range doc.Tables[relation.table] {
			target, _ := row[relation.targetColumn].(string)
			revision, _ := row["revision_id"].(string)
			authority, _ := row["authority"].(string)
			if !relation.targets[target] || !revisions[revision] || (authority != "legacy" && authority != "catalog") {
				problems = append(problems, relation.table+": ungültiges Ziel, Revision oder maßgebliche Darstellung. / Invalid target, revision or authority.")
			}
			overrides, ok := row["overrides_json"].(string)
			if _, err := decodeVehicleCatalogFields(overrides, false); !ok || err != nil {
				problems = append(problems, relation.table+": ungültige individuelle Katalogabweichungen. / Invalid individual catalog overrides.")
			}
			timestamp(row, "updated_at", relation.table)
		}
	}
	return problems
}
