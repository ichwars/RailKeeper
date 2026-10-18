# Retirement of layout planning / Entfernung der Anlagenplanung

The layout workspace, module/setup editor, digital twin, track planner, geometry engines,
planning APIs and Planner role have been removed. Accessory inventory, track article metadata,
article search, vehicle reservations and vehicle installations remain supported.

Der Anlagenarbeitsbereich einschließlich Modulen, Aufbauten, digitalem Zwilling, Gleisplaner,
Geometrieberechnung, Planungs-APIs und Planner-Rolle ist entfernt. Zubehörbestand, Gleisartikel,
Artikelsuche, Fahrzeugreservierungen und Fahrzeugeinbauten bleiben erhalten.

## Existing installations / Bestehende Installationen

Migration `0069_retire_layout_planning.sql` runs transactionally after the existing migration
safety snapshot. It preserves the complete original rows of former layout reservations and
installations in `retired_allocation_records` before changing operational state:

- Active layout reservations are cancelled. Individually reserved items return to their source.
- Counted layout installations return to their original storage location, with a stock movement.
- Individual installations return to their original location. Defective/maintenance-due items
  retain their condition and use the maintenance lifecycle; other items become stored.
- Old installations are marked removed. Vehicle reservations, installations and histories are
  unchanged. Repeating the conversion does not return stock twice.

Die Migration archiviert die vollständigen Originalzuordnungen, gibt Anlagenreservierungen frei
und führt Anlageneinbauten ins Ursprungslager zurück. Zustände von Einzelstücken bleiben erhalten.
Fahrzeugeinbauten und deren Historie werden nicht verändert. Wiederholungen buchen nichts doppelt.

The old planning tables and allocation columns deliberately remain as dormant compatibility
storage. Historical migrations are not rewritten. No planning service, frontend or HTTP endpoint
can create new plans or allocations to them. Read/write APIs accept only `vehicleId`; stale
`layoutId`, `layoutUnitId` and `technicalPositionId` inputs are rejected. Archived references can
still prevent permanent article deletion, preserving their referential integrity.

Alte Planungstabellen bleiben bewusst als inaktiver Datenbestand erhalten. Ein späterer physischer
Schemaabbau wäre ein eigener Schritt mit ausdrücklicher Datenaufbewahrungsentscheidung. Die
Anwendung bietet keine Planungsfunktion und keine Bearbeitung dieser Altbestände mehr an.

Migration `0070_retire_planner_role.sql` grants existing Planner accounts Viewer membership and
removes Planner. Other memberships are preserved, with no new inventory write permission.
Unknown/retired start-page preferences fall back to Overview; old sidebar entries are normalized.

## Backups / Sicherungen

Backup version 23 includes `retired_allocation_records` and the existing legacy planning tables.
Older supported versions remain importable. Preflight explains the archival/stock conversion
when planning data is present. Restore imports the original data, backfills legacy snapshots and
runs the same idempotent allocation conversion before committing, under deferred foreign-key
checks. Invalid restores roll back their database transaction and leave uploads unchanged.

Backupversion 23 sichert auch archivierte Originalzuordnungen. Ältere Sicherungen bleiben lesbar;
die Vorschau weist auf die Übernahme hin. Restore und Datenbankupgrade verwenden dieselbe
Bestandsübernahme. Benutzer, Rollen, Sitzungen und andere lokale Authentifizierungsdaten werden
weiterhin weder exportiert noch aus Backups überschrieben. Die Rollenübernahme erfolgt allein
beim regulären Datenbankupgrade.

## Source boundaries / Quellcodegrenzen

Planning names may remain in historical migrations, the scoped legacy backup/conversion code,
compatibility tests, this document and published release history. They are not runtime planning
capabilities. The separate Tillig research and uncommitted library work were preserved outside
the checkout before removing their planner-only integration test and materials.
