# Artikelinventarnummer, Tabellenstruktur und Stammdatenfelder

**Datum:** 2026-08-09

**Status:** Fachlich und technisch freigegeben

## Ziel

Die Artikelverwaltung wird enger an den Fahrzeugbestand angelehnt. Jeder Artikelstamm erhält eine
eigene, automatisch vergebene Inventarnummer. Die Übersicht zeigt die Artikel als vollständig
sortierbare Bestandsliste mit Auswahl und Bild. Der Anlagedialog verwendet vorhandene Stammdaten
statt frei gepflegter Parallelwerte. Die Kennzahlen behalten ihre Kartenstruktur, zeigen aber wie
im Fahrzeugbestand einen großen Summenwert und darunter dezente Teilwerte.

Artikelstamm und Einzelstück bleiben getrennte Ebenen. Die neue Artikelinventarnummer identifiziert
den Katalogartikel. Bereits vorhandene Inventarnummern einzelner physischer Einzelstücke behalten
ihre heutige Bedeutung und ihren eigenen Lebenszyklus.

## Inventarnummer des Artikelstamms

- `accessory_products` erhält das eindeutige, verpflichtende Feld `inventory_number`.
- Die bestehende Inventarnummernverwaltung erhält die Kategorie `Artikel` mit dem Standardpräfix
  `RK-ART`, Startwert `1`, sechs Stellen und aktivem Status.
- Die Nummer wird bei der Artikelanlage automatisch und in derselben SQLite-Transaktion wie der
  Artikel vergeben. Ein fehlgeschlagener Schreibvorgang verbraucht keine Nummer.
- Bestehende Artikel werden in stabiler Reihenfolge nach Erstellungszeit und ID ab
  `RK-ART-000001` nummeriert. Der Schemazähler wird anschließend auf den nächsten freien Wert
  gesetzt.
- Die Nummer ist im Artikeldialog schreibgeschützt sichtbar. Bei der Neuanlage weist das Feld auf
  die automatische Vergabe hin.
- Ein deaktiviertes oder fehlendes Schema `Artikel` verhindert die Neuanlage mit einer
  verständlichen Fehlermeldung. Es gibt keinen stillen Fallback auf freie Texteingabe oder UUID.
- Suche, Sortierung, API-Ausgaben, OpenAPI und Backups führen die Artikelinventarnummer mit.
- Neue Backups bewahren die Nummer. Ältere Backups ohne Artikelinventarnummer erhalten beim Restore
  deterministisch neue Nummern. Die Wiederherstellung älterer Backupstände bleibt möglich.

## Kennzahlen

Die vier Karten bleiben im Raster und verwenden die bestehende Bestandsdarstellung. Der primäre
Summenwert steht groß, die ergänzenden Einzelwerte stehen darunter kleiner und farblich
zurückgenommen:

- Artikelbestand: `24 Artikel`, darunter `5 Arten`.
- Verfügbar: `81 frei`, darunter `7 Lagerorte`.
- Gebunden: `20 gebunden`, darunter `6 reserviert · 14 eingebaut`.
- Pflegehinweise: `3 Hinweise`, darunter `3 unvollständig`.

Die bestehenden Filteraktionen der Karten bleiben erhalten. Aktive Zustände orientieren sich am
Fahrzeugbestand und dürfen die Wertehierarchie nicht wieder zu einer lauten, gleichgewichteten
Textzeile verdichten.

## Artikelübersicht

Die Tabelle verwendet folgende verbindliche Spaltenfolge:

1. Auswahl
2. Bild
3. Inventarnummer
4. Hersteller
5. Artikelnummer
6. Bezeichnung
7. Artikelart und Unterart
8. Spurweite
9. Bestand
10. Lagerort
11. Aktionen

Alle fachlichen Datenköpfe sind sortierbar. Auswahl und Aktionen sind ausgenommen. Die Bildspalte
sortiert nach vorhandenem bzw. fehlendem Primärbild. Die übrigen Sortierungen werden serverseitig
und deterministisch mit ID als letzter Vergleichsstufe ausgeführt.

Die Kopfcheckbox wählt alle aktuell sichtbaren Artikel aus. Jede Zeile besitzt eine eigene
Checkbox. Auswahlzustände werden auf tatsächlich geladene Artikel begrenzt und barrierefrei
beschriftet. Die Spalten Bild und Aktionen bleiben nicht Teil der Textsuche. Die Suche umfasst
mindestens Inventarnummer, Hersteller, Artikelnummer und Bezeichnung.

Die Primärbilder verwenden die vorhandene Artikel-Dokumentlogik und die kompakte
Bestandsvorschau. Fehlt ein Bild, erscheint der vorhandene ruhige Platzhalter. Desktop bleibt dicht
lesbar, mobil bleibt die Tabelle kontrolliert horizontal scrollbar.

## Stammdatengebundene Formularfelder

Der Artikeldialog lädt und verwendet die bestehenden autoritativen Datenquellen:

- Hersteller aus `manufacturer`.
- Spurweiten aus `gauge`, weiterhin als Mehrfachauswahl.
- Bestandseinheiten aus `stock_unit`.
- Artikelarten aus `article_type`.
- Unterarten aus `accessory_subtype`.
- Kontrollierte Zusatzfelder und deren Optionen aus `accessory_custom_field`.

Es werden nur aktive Werte für eine neue Auswahl angeboten. Ein am bestehenden Artikel
gespeicherter, inzwischen inaktiver Wert bleibt beim Bearbeiten sichtbar, damit keine Daten
unbemerkt verloren gehen. Bezeichnungen folgen der vorhandenen deutschen und englischen
Stammdatenlokalisierung; gespeichert werden die bereits kanonisch verwendeten Werte.

Hersteller, Spurweite und Bestandseinheit werden mit den vorhandenen app-eigenen Auswahlkomponenten
umgesetzt. Freitextfelder bleiben bestehen, wenn keine passende Datenbankquelle existiert, etwa für
Maßstab. Bereits kontrollierte Felder wie Herstellerstatus werden nicht in neue Stammdatenarten
überführt.

Wenn eine Datenquelle nicht geladen werden kann, wird nur das betroffene Feld deaktiviert und mit
einem gezielten Ladefehler versehen. Andere Reiter und Funktionen bleiben nutzbar. Ein erneuter
Ladeversuch ist möglich. Bereits geladene oder bestehende Werte werden nicht verworfen.

## Verträge und Kompatibilität

- Backendmodell, Querymodell, Frontendtypen und `openapi/railkeeper.yaml` verwenden einheitlich
  `inventoryNumber`.
- Die Artikelanlage reserviert die nächste Nummer über die bestehende Schemenverwaltung. Es
  entsteht keine zweite Nummernquelle.
- Rollen, CSRF, Archivierung, Bestandsstrategien, Reservierungen, Einbauten und
  Einzelstück-Inventarnummern bleiben unverändert.
- Alte Artikel und ältere Backups werden vorwärtskompatibel ergänzt. Historische Migrationen werden
  nicht umgeschrieben; die Änderung erfolgt in der nächsten Migration `0043`.
- Die lokale, selbst gehostete SQLite-Architektur bleibt unverändert.

## Fehler- und Sicherheitsverhalten

- Nummernvergabe und Artikelanlage sind atomar und konkurrierende Anlagen dürfen keine doppelten
  Nummern erzeugen.
- Eindeutigkeitsverletzungen werden als fachlicher Konflikt gemeldet.
- Unbekannte oder inaktive neue Stammdatenwerte werden serverseitig abgewiesen, soweit diese
  Invariante bereits für die jeweilige Stammdatenart gilt.
- Restore und Migration dürfen bestehende Artikel weder verlieren noch doppelt nummerieren.
- Die neue Tabellenauswahl führt keine destruktive Aktion aus.

## Verifikation

- Migrationstest für Schema, stabile Altbestandsnummerierung, Zählerstand und Eindeutigkeit.
- Backendtests für atomare sowie konkurrierende Vergabe, Suche, alle Sortierungen und Fehler bei
  fehlendem oder deaktiviertem Schema.
- Backup- und Restoretests für neue Nummern sowie ältere Backups ohne Nummer.
- OpenAPI-Vertragstests für Listen-, Detail-, Anlage- und Änderungsantworten.
- Frontendtests für Kennzahlenhierarchie, Spaltenreihenfolge, Sortieraktionen, Bilddarstellung,
  Einzel- und Gesamtauswahl sowie die schreibgeschützte Inventarnummer.
- Frontendtests für Hersteller-, Spurweiten- und Bestandseinheiten-Stammdaten, inaktive Altwerte,
  Ladefehler und Retry.
- Vollständiger Lauf von `go test ./... -count=1`, `npm.cmd run test:run` und
  `npm.cmd run build`.
- Browser-QA in Dark und Light Theme auf Desktop und 390 × 844 Pixeln, einschließlich horizontaler
  Tabellenführung, langer deutscher Texte und fehlender Bilder.

## Nicht Bestandteil dieser Umsetzung

- Sammelaktionen für ausgewählte Artikel fehlen weiterhin. Dieser offene Folgepunkt ist bewusst
  dokumentiert; insbesondere werden Archivieren, Wiederherstellen, Exportieren oder Löschen nicht
  als Mehrfachaktion ergänzt.
- Artikelinventarnummern ersetzen keine Einzelstück-Inventarnummern.
- Es entstehen keine neue Stammdatenart für Maßstab und keine neue UI-Bibliothek.
- Es gibt keine Kartenansicht, Anlagensteuerung oder Cloudfunktion.
