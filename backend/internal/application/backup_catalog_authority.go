package application

import "fmt"

var vehicleCatalogColumns = map[string]string{
	"name": "name", "manufacturer": "manufacturer", "articleNumber": "article_number",
	"articleSourceUrl": "article_source_url", "gauge": "gauge", "epoch": "epoch",
	"railwayCompany": "railway_company", "category": "category", "gattung": "gattung",
	"description": "description", "ean": "ean", "productionPeriod": "production_period", "listPrice": "list_price",
}

func validateBackupCatalogAuthority(doc *BackupDocument) []string {
	var problems []string
	revisions := map[string]map[string]any{}
	for _, row := range doc.Tables["catalog_article_revisions"] {
		id, _ := row["id"].(string)
		revisions[id] = row
	}
	for _, target := range []struct{ links, entities, column string }{
		{"vehicle_article_links", "vehicles", "vehicle_id"}, {"vehicle_set_article_links", "vehicle_sets", "vehicle_set_id"},
	} {
		entities := map[string]map[string]any{}
		for _, row := range doc.Tables[target.entities] {
			id, _ := row["id"].(string)
			entities[id] = row
		}
		for index, row := range doc.Tables[target.links] {
			source, present := row["source_authority"]
			if doc.Version < 26 {
				if present {
					problems = append(problems, "Artikel-Schreibmodus benötigt Format 26. / Catalog writer mode requires format 26.")
				}
				continue
			}
			if !present || (source != nil && source != "legacy" && source != "catalog") {
				problems = append(problems, "Ursprünglicher Artikelmodus fehlt oder ist ungültig. / Missing or invalid original catalog authority.")
			}
			if row["authority"] != "catalog" {
				continue
			}
			id, _ := row[target.column].(string)
			revisionID, _ := row["revision_id"].(string)
			entity, revision := entities[id], revisions[revisionID]
			if entity == nil || revision == nil {
				continue
			}
			rawFields, _ := revision["fields_json"].(string)
			fields, err := decodeVehicleCatalogFields(rawFields, true)
			if err != nil {
				continue
			}
			rawOverrides, _ := row["overrides_json"].(string)
			overrides, err := decodeVehicleCatalogFields(rawOverrides, false)
			if err != nil {
				continue
			}
			for key, value := range overrides {
				fields[key] = value
			}
			for key, column := range vehicleCatalogColumns {
				value, valid := entity[column].(string)
				if entity[column] == nil {
					value, valid = "", true
				}
				if !valid || fields[key] != value {
					problems = append(problems, fmt.Sprintf("%s Zeile %d: Artikel und Kompatibilitätsansicht widersprechen sich. / Catalog and compatibility projection differ.", target.links, index+1))
					break
				}
			}
		}
	}
	return problems
}
