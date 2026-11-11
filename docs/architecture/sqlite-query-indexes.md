# SQLite-Indizes für RailKeeper

Die Migrationen 0082 bis 0085 ergänzen 29 Indizes und ersetzen fünf bisherige
Indizes. Primärschlüssel und Eindeutigkeitsregeln bleiben in den bestehenden
Definitionen. Die Migrationen enthalten ausschließlich Index-DDL.

## Fahrzeuge, CVs und Ausstellungen

| Index | Abfrage und Wirkung |
| --- | --- |
| `idx_vehicles_inventory_nocase` | Bestandsnummern im Transfer und Fahrzeugzuordnung beim Ausstellungsimport. Der Plan wechselt vom linearen Indexdurchlauf zur direkten Suche mit `COLLATE NOCASE`. |
| `idx_vehicle_sets_inventory_nocase` | Kollisionsprüfung für Bestandsnummern von Sets beim Transfer. Direkte Suche mit `COLLATE NOCASE`. |
| `idx_vehicles_manufacturer_nocase` | Verwendungsprüfung beim Verwalten und Löschen von Hersteller-Stammdaten. Indexsuche nach dem Herstellernamen. |
| `idx_vehicles_exhibition_active` | Rücksetzen abgelaufener Ausstellungskennzeichen in `ResetLegacyMatches`. Der äußere UPDATE-Durchlauf beschränkt sich auf Fahrzeuge mit `exhibition=1`. |
| `idx_vehicle_cv_values_order` | CV-Liste pro Fahrzeug. Liefert die Reihenfolge nach Profil, CV31, CV32 und CV-Nummer ohne Teilsortierung. Für die bisherige Spaltenfolge Fahrzeug, Profil, CV-Nummer steht der CV-Identitätsindex zur Verfügung. |
| `idx_vehicle_cv_values_source` | `CVFileSelect` sammelt die IDs der CV-Werte pro Quelldatei für die Dateirevision. Die Unterabfrage sucht die Datei-ID direkt und erhält die CV-IDs sortiert aus dem Index. |
| `idx_vehicle_cv_history_vehicle_order` | `CVHistories` liest die Änderungshistorie pro Fahrzeug. Sucht nach Fahrzeug und liefert die jüngsten Änderungen zuerst. Der Index pro CV-Wert dient weiterhin dem Einzelverlauf. |
| `idx_vehicle_images_primary_order` | Bildliste und Bildrevision eines Fahrzeugs. Liefert Hauptbild, Sortierposition und Erstellungszeit in der geforderten Reihenfolge. Die automatische Hauptbild-Beförderung verwendet die eigene Reihenfolge nach Sortierposition und Erstellungszeit. |
| `idx_exhibition_entries_list_order` | Ausstellungsdetail in `Entries`. Sucht nach Liste und liefert Position, Fahrzeugname und Besitzer mit der passenden NOCASE-Sortierung. |

## Zubehör und Datenübertragung

| Index | Abfrage und Wirkung |
| --- | --- |
| `idx_accessory_reservations_active_stock` | Verfügbarer Bestand bei Reservierung, Korrektur, Umbuchung und Import. Sucht aktive Reservierungen nach Produkt, Lagerort und fehlender Einzelstück-ID. Die Menge steht im Index. |
| `idx_accessory_installations_active_product` | `GetAllocationSummary` summiert eingebaute Mengen je Produkt. Der Teilindex berücksichtigt aktuelle Einbauten und unterstützt die Unterteilung nach Einzelstück-ID. |
| `idx_accessory_products_normalized_article` | `FindDuplicateCandidates` sucht nach `lower(trim(manufacturer))` und `lower(trim(article_number))`. Der Ausdrucksindex liefert die Kandidaten in der gewünschten Namensreihenfolge. |
| `idx_accessory_products_inventory_nocase` | Kollisionsprüfung für Zubehör-Bestandsnummern beim Transfer. Direkte Suche mit `COLLATE NOCASE`. |
| `idx_accessory_attributes_key_product` | Verwendungsprüfung und Einzelprüfung beim Ändern eines Zubehör-Freitext- oder Auswahlfelds. Der Join kann bei dem gesuchten Attributschlüssel beginnen und die Produkt-ID direkt aus dem Index lesen. |
| `idx_data_transfer_jobs_created_order` | Transfer-Verlauf in `ListJobs`. Liefert die gewünschte absteigende Reihenfolge nach Zeit und ID auch bei gleichen Zeitstempeln. |
| `idx_data_transfer_jobs_profile_order` | Transfer-Verlauf mit `profile_id=?`. Sucht nur Aufträge des gewählten Profils und liefert sie in Zeitreihenfolge. |
| `idx_data_transfer_issues_job_order` | Meldungen in Transfer-Vorschau und Bestätigung. Sucht nach Job und liefert sie nach Zeit und ID sortiert. |

## Medienreferenzen

Die sechs Indizes helfen derselben Löschprüfung in `DeleteIfUnreferenced`.
SQLite prüft jeden Referenztyp per Indexsuche. Für Bilder und Sets verknüpft
SQLite die beiden Suchen mit `MULTI-INDEX OR`.

| Index | Geprüfte Referenz |
| --- | --- |
| `idx_vehicle_images_blob` | Originaldatei eines Fahrzeugbilds. |
| `idx_vehicle_images_thumbnail_blob` | Vorschaudatei eines Fahrzeugbilds. |
| `idx_vehicle_attachments_blob` | Datei eines Fahrzeuganhangs. |
| `idx_vehicle_cv_files_blob` | Datei einer CV-Quelle. |
| `idx_vehicle_sets_image_blob` | Original des eigenen Setbilds. Der Teilindex enthält belegte Referenzen. |
| `idx_vehicle_sets_thumbnail_blob` | Vorschau des eigenen Setbilds. Der Teilindex enthält belegte Referenzen. |

## Authentifizierung und Digitalzentralen

| Index | Abfrage und Wirkung |
| --- | --- |
| `idx_audit_logs_created_order` | `ListAuditLog` liest die neuesten Ereignisse mit `LIMIT`. Der Index ermöglicht den Einstieg bei den jüngsten Einträgen ohne Sortierung des gesamten Logs. |
| `idx_rate_limit_attempts_scope_expiry` | Bereinigung in `ratelimit.Store.Allow`. Sucht abgelaufene Versuche direkt über Scope und Zeitgrenze. Die Zählung pro Schlüssel hat ihren eigenen Index. |
| `idx_login_origins_expiry` | `saveOrigin` löscht abgelaufene Login-Ursprünge mit `expires_at<=?` per Bereichssuche. |
| `idx_password_reset_pending_user` | Anforderung und Bestätigung eines Passwort-Resets markieren offene Anfragen des Nutzers. Der Teilindex enthält Anfragen mit `handled_at IS NULL`. |
| `idx_digital_center_sessions_ready_station` | Neuester erfolgreicher Lesestand, gespeichertes Decoderprofil und ECoS-Backup-Vergleich. Sucht lesebereite Sitzungen nach Provider, Host und Port. Der Lesestand kann rückwärts nach Abschlusszeit und Row-ID gelesen werden. |
| `idx_digital_center_work_items_session_order` | Arbeitsliste einer Digitalzentrale in `ListWorkItems`. Sucht nach Sitzung und liefert Name, Objekt-ID und Datensatz-ID in der geforderten Reihenfolge. |

## Prüfung und Grenzen

Der Quellscan erfasste 632 SQL-Literale und Fragmente in Go einschließlich
Lesezugriffen, schreibenden Filtern und dynamisch aufgebauten Abfragen. Die
Indexstruktur wurde nach Anwendung der Migrationen aus `sqlite_schema` und den
Index-PRAGMAs gelesen. Die ausgewählten Abfragepläne wurden vor und nach der
Migration mit synthetischen Daten und dem RailKeeper-Treiber (SQLite 3.54.1)
verglichen. Das Upgrade von Schema 0081 mit bestehenden Daten, erneutes Starten
und Datenbankintegrität deckt `TestQueryIndexMigrationsPreserveDataAndImprovePlans`
ab.

Der synthetische Datensatz enthielt 6.000 Fahrzeuge, 96.000 CV-Werte,
144.000 Historieneinträge und 80.000 Audit-Ereignisse. Die zusätzlichen Indizes
benötigen Speicherplatz und müssen bei betroffenen Schreiboperationen gepflegt werden.

Vergleich der SQLite-VM-Schritte auf demselben Datensatz mit SQLite 3.46.1.
Die VM-Schritte geben Ausführungsaufwand an. Antwortzeiten hängen zusätzlich von
Datengröße, Cache und Speicher ab.

| Abfrage | Vorher | Nachher |
| --- | ---: | ---: |
| CV-Datei mit referenzierenden CV-Werten | 480.148 | 182 |
| CV-Historie eines Fahrzeugs | 432.625 | 375 |
| Neueste 40 Audit-Ereignisse | 1.129.389 | 1.092 |
| Zuletzt gelesene Sitzung einer Digitalzentrale | 3.510 | 108 |

Weitere Felder mit wenigen Ausprägungen wurden nach ihrer Selektivität und dem
tatsächlichen Zugriffspfad beurteilt. Primär- und Fremdschlüsselpfade mit
passendem führendem Indexpräfix benötigen keinen weiteren Index. Textsuche mit
führendem `%`, JSON-Array-Suche, berechnete Sortierung nach Beständen und
vollständige Backup-Exporte erfordern andere Ansätze. Der Sortierindex auf
`vehicles(updated_at DESC, inventory_number)` sparte VM-Schritte, erhöhte aber
beim vollständigen Durchlauf und Auslesen von 6.000 Fahrzeugen die mediane Zeit
von 113 auf 124 ms (zehn Wiederholungen mit SQLite 3.54.1). Er wurde verworfen.
Der Ausdrucksindex auf
ECoS-Objekt und `julianday(updated_at)` wurde verworfen. Die geprüften
Sitzungsabfragen verwendeten den Stationsindex zusammen mit dem vorhandenen
eindeutigen Schlüssel aus Sitzung und Objekt-ID.

Bei einem passenden WHERE- oder ORDER-BY-Ausdruck berücksichtigt SQLite
Kollation, Spaltenreihenfolge und Prädikat von Teilindizes.
[SQLite Query Planner](https://sqlite.org/queryplanner.html),
[Partial Indexes](https://sqlite.org/partialindex.html),
[Indexes on Expressions](https://sqlite.org/expridx.html)
