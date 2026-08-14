# Tillig-TT-Gleisplaner, Paket C

**Datum:** 2026-08-09

**Status:** Aus der fachlich abgestimmten Stage-3-Spezifikation abgeleitet und lokal freigegeben

**Abnahmestand 2026-08-09:** Paket C ist lokal umgesetzt. Die vollständige Go-Suite, 64
Frontend-Testdateien mit 342 Tests und der Produktionsbuild sind grün. Die Browserabnahme gegen den
neu gebauten lokalen Server bestätigt Vorschau, eigenen Reservierungsdialog, Leerzustand, Anmeldung
und eine leere Browserkonsole.

## Ziel

Paket C schließt Stage 3 mit einer nachvollziehbaren Änderungsvorschau und einer ausdrücklich
bestätigten, planobjektbezogenen Materialreservierung ab. Veröffentlichung, Reservierung und Einbau
bleiben getrennte Aktionen. Weder Vorschau noch Veröffentlichung verändern Bestand.

## C1: Revisionsvergleich

Jedes Gleisobjekt erhält eine revisionsübergreifende `lineage_id`. Beim Erstellen ist sie identisch
mit der Objekt-ID, beim Ableiten eines Entwurfs wird sie aus der Basisrevision übernommen. Dadurch
kann RailKeeper hinzugefügte, entfernte und geänderte Objekte deterministisch unterscheiden, obwohl
jede Revision eigene unveränderliche Objekt-IDs besitzt.

Die Vorschau vergleicht den Entwurf mit seiner `base_revision_id`. Ohne Basis gelten alle Objekte als
hinzugefügt. Sie liefert:

- hinzugefügte, entfernte und geometrisch oder in der Pose geänderte Gleisobjekte,
- Bedarfssalden je konkreter Geometrieversion,
- neu hinzugekommene und behobene Prüfhinweise, normalisiert über Lineage und Anschlusskennung,
- Aufbaukonfigurationen, die noch die Basisrevision referenzieren.

Die Vorschau ist vollständig abgeleitet und wird nicht persistiert.

## C2: Planobjektbezogene Reservierung

Eine neue Verknüpfungstabelle verbindet vorhandene Zubehörreservierungen mit genau einem
Gleisobjekt. Historische Links bleiben erhalten. Ein partieller eindeutiger Index erlaubt je
Gleisobjekt höchstens eine aktive Reservierung. Stornierung oder Erfüllung deaktiviert den Link in
derselben Transaktion und ermöglicht anschließend eine neue Reservierung.

Der bestätigte Batch-Befehl enthält für jedes Gleisobjekt Produkt, Lagerort, optionale Einzelobjekt-ID
und erwartete Objektversion. Der Server prüft in einer Transaktion:

- Revision ist Entwurf oder in Prüfung,
- Objekt gehört zur Revision und ist unverändert,
- Hersteller und Artikelnummer stimmen mit der Geometrie überein,
- Lagerort und Bestand beziehungsweise Einzelobjekt sind verfügbar,
- es existiert keine aktive planobjektbezogene Reservierung.

Bei einem Fehler entsteht keine Teilreservierung. Planner und Admin dürfen reservieren, Viewer und
Editor nur Vorschau und Status lesen, Messe bleibt ausgeschlossen. Die Oberfläche zeigt vor dem
Schreiben eine RailKeeper-eigene Bestätigung mit Objektanzahl, Produkt, Lagerort und resultierendem
freien Bestand.

## Daten- und Backup-Kompatibilität

Migration 0046 ergänzt die Lineage und die Reservierungslinks. Bestehende Gleisobjekte erhalten ihre
eigene ID als Lineage. Das Backupformat steigt auf Version 6. Backups der Versionen 1 bis 5 bleiben
importierbar; fehlende Lineage wird beim Restore beziehungsweise durch den Datenbankstandard aus der
Objekt-ID abgeleitet.

## Abgrenzung

Paket C installiert kein Material und steuert keine Digitalzentrale. Flexgleis, Höhenprofile,
Modulports und weitere Gleisgeometrien bleiben Stage 4. Eine automatische Reservierung ohne
Bestätigungsdialog ist ausdrücklich ausgeschlossen.
