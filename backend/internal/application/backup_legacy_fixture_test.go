package application_test

import "railkeeper/backend/internal/application"

// Historical fixtures must contain only data their original format could export.
func setLegacyBackupVersion(doc *application.BackupDocument, version int) {
	doc.Version = version
	if version < 25 {
		for _, table := range []string{"catalog_articles", "catalog_article_revisions", "vehicle_article_links", "vehicle_set_article_links"} {
			delete(doc.Tables, table)
		}
		return
	}
	if version == 25 {
		for _, table := range []string{"vehicle_article_links", "vehicle_set_article_links"} {
			for _, row := range doc.Tables[table] {
				delete(row, "source_authority")
			}
		}
	}
}
