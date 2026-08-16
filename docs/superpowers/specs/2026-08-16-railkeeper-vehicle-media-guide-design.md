# Zweisprachiges Benutzerhandbuch für Fahrzeugbilder und Beilagen

Datum: 16. August 2026

Status: fachlich freigegeben

Geltungsbereich: Benutzerhandbuch für RailKeeper v0.1.17.6

## Ziel

RailKeeper erhält ein vollständiges, aufgabenorientiertes Kapitel für Fahrzeugbilder und allgemeine
Beilagen. Benutzer sollen Bilder und Dateien sicher hochladen, ordnen, beschreiben, ansehen,
herunterladen und entfernen können, ohne Implementierungswissen oder andere Projektdokumente zu
benötigen.

Die englische und deutsche Fassung sind fachlich gleichwertig. Sichtbare Bezeichnungen folgen der
jeweiligen stabilen Oberfläche. Gespeicherte Kategorien und andere sprachunabhängig verwendete
Werte werden so dargestellt, wie RailKeeper sie tatsächlich speichert.

## Pfade und Navigation

Das Kapitel wird als zusammengehöriges Sprachpaar veröffentlicht:

```text
/guide/vehicles/media
/de/guide/vehicles/media
```

Die bestehenden Coverage-Pfade bleiben deshalb unverändert:

```text
guide/vehicles/media.md
de/guide/vehicles/media.md
```

Beide Seiten werden direkt nach dem Fahrzeugbestand in der jeweiligen Seitenleiste und auf der
Einstiegsseite des Benutzerhandbuchs verlinkt. Das Kapitel zum Fahrzeugbestand erhält einen
passenden Querverweis. Links auf noch nicht veröffentlichte Fachkapitel werden nicht als
funktionierende Ziele ausgegeben. Solange diese Seiten fehlen, beschreibt ein kurzer Hinweis nur
die fachliche Abgrenzung.

## Zielgruppe und Rollen

Admin, Editor, Viewer und Planner können vorhandene Bilder und Beilagen im Fahrzeugdatensatz
einsehen. Hochladen, Metadaten ändern und löschen erfordert serverseitig Admin oder Editor. Ein
sichtbares Bedienelement ersetzt die Rollenprüfung nicht. Ein reines Messe-Konto hat keinen
allgemeinen Zugriff auf den Fahrzeugbestand.

Bilder und Beilagen benötigen eine bereits gespeicherte Fahrzeug-ID. Beim Anlegen eines Fahrzeugs
erklärt der Uploads-Tab deshalb, dass lokale Dateien erst nach dem ersten Speichern hinzugefügt
werden können.

## Kapitelstruktur

Das Kapitel folgt dem tatsächlichen Arbeitsablauf im Tab **Uploads**:

1. Zweck, Rollen und Voraussetzungen
2. Einstieg über Fahrzeugbestand, Schnellmenü oder Bearbeitungsdialog
3. lokale Bilder hochladen und prüfen
4. Bildmetadaten, Reihenfolge und Hauptbild verwalten
5. Bilder mit vorhandenen Wartungen verknüpfen
6. allgemeine Beilagen auswählen oder per Drag-and-drop hochladen
7. Kategorien, Bemerkungen und automatische Zuordnung verstehen
8. Beilagen anzeigen, öffnen und herunterladen
9. Bilder und Beilagen sicher entfernen
10. Dateiformate, Größenlimits und Sicherheitsgrenzen
11. Teilfehler, leere Zustände und Fehlerbehebung
12. fachliche Grenzen, verwandte Funktionen und dokumentierte Version

## Bilder

### Upload

Der Benutzer öffnet ein gespeichertes Fahrzeug im Bearbeitungsmodus und wechselt zu **Uploads**.
**Bild hochladen** akzeptiert eine oder mehrere lokale Dateien. Unterstützt werden:

- JPG und JPEG
- PNG
- WebP

Das Standardlimit beträgt 10 MB pro Bild. Der Betreiber kann den serverseitigen Grenzwert ändern.
Die Oberfläche prüft die Dateiendung, der Server prüft zusätzlich Größe und erkannten MIME-Typ.

Uploads werden nacheinander ausgeführt. Ist noch kein Bild vorhanden, wird das erste erfolgreich
gespeicherte Bild automatisch Hauptbild. Nach Abschluss lädt RailKeeper den vollständigen
Fahrzeugdatensatz neu.

### Anzeige und Metadaten

Das Kapitel erklärt jede Bildaktion:

- Originalgröße in der Bildvorschau öffnen
- Bildbeschreibung bearbeiten
- vorhandene Quelle öffnen, wenn ein externer Import eine Quelladresse gespeichert hat
- Bild nach oben oder unten verschieben
- Bild als Hauptbild markieren
- Bild mit einer vorhandenen Wartung verknüpfen oder die Verknüpfung entfernen
- Bild entfernen

Hauptbild, Alternativbilder, Sortierreihenfolge, Beschreibung und Wartungsverknüpfung sind
Bildmetadaten. Änderungen an diesen Daten werden erst mit **Änderungen speichern** im
Fahrzeugdialog persistiert. Das Hochladen und das Entfernen eines bereits gespeicherten Bildes
wirken dagegen unmittelbar über eigene Serveranfragen.

Das Hauptbild wird im Fahrzeugbestand und in anderen kompakten Ansichten bevorzugt verwendet. Wird
das Hauptbild gelöscht, übernimmt das nächste verbleibende Bild nach Sortierreihenfolge diese Rolle.

### Wartungsgrenze

Das Medienkapitel beschreibt ausschließlich die Bildverknüpfung. Die Wartung selbst bleibt ein
eigenes Fachkapitel. Ein mit einer Wartung verknüpftes Bild kann nicht sofort gelöscht werden. Der
Benutzer muss zuerst **Keine Wartung** auswählen, den Fahrzeugdatensatz speichern und danach das
Bild entfernen.

## Allgemeine Beilagen

### Upload

Beilagen können über **Beilage hochladen** ausgewählt oder in die Drag-and-drop-Zone gelegt werden.
Mehrfachauswahl ist möglich. Unterstützt werden:

- PDF
- TXT
- CSV
- JSON
- XML
- ZIP
- JPG und JPEG
- PNG
- WebP

Das Standardlimit beträgt 25 MB pro Datei. Der Betreiber kann den serverseitigen Grenzwert und die
erlaubten Beilagenendungen einschränken. Der Server lehnt leere, ausführbare, nicht erlaubte,
inhaltlich blockierte oder zu große Dateien ab.

Eine vor dem Upload ausgewählte Kategorie und Bemerkung gilt für alle Dateien dieses Uploadvorgangs.
Ohne explizite Kategorie bestimmt RailKeeper sie anhand des Dateinamens und der Endung.

### Kategorien

Die stabile Auswahlliste enthält gespeicherte deutsche Werte:

- `Anleitung`
- `Rechnung`
- `Decoder-Datei`
- `Dokumentation`
- `Ersatzteilliste`
- `Zertifikat`
- `Sonstiges`

Die automatische Zuordnung folgt dieser Priorität:

1. `rechnung` oder `invoice` im Dateinamen ergibt `Rechnung`.
2. `decoder` im Namen sowie JSON oder XML ergibt `Decoder-Datei`.
3. `ersatzteil` im Namen ergibt `Ersatzteilliste`.
4. `zertifikat` oder `certificate` ergibt `Zertifikat`.
5. `anleitung`, `manual` oder `bedienung` ergibt `Anleitung`.
6. andere PDF-Dateien ergeben `Dokumentation`.
7. alle übrigen erlaubten Dateien ergeben `Sonstiges`.

Die Reihenfolge ist relevant. Eine frühere Regel gewinnt.

### Gespeicherte Beilagen

Die Liste zeigt Originalname, Kategorie, erkannten Dateityp und Dateigröße. Kategorie und Bemerkung
können für jede Beilage separat geändert und über deren Speichern-Aktion persistiert werden.

Das Kapitel unterscheidet drei Aktionen:

- **Anzeigen** öffnet die integrierte Vorschau.
- **Datei öffnen** öffnet eine geeignete Inline-Darstellung in einem neuen Browser-Tab.
- **Datei herunterladen** lädt die gespeicherte Originaldatei herunter.

PDFs, JPG, JPEG, PNG, WebP, TXT, CSV, JSON und XML besitzen eine Vorschau. ZIP-Dateien können
heruntergeladen werden, bieten aber keine integrierte Vorschau. Die Oberfläche zeigt in diesem Fall
den stabilen Leerzustand für nicht unterstützte Vorschautypen.

Die Aktion **Ersatzteile extrahieren** und der Bereich **Gefundene Dokumente** werden als
Übergänge genannt, aber nicht fachlich erklärt. Sie gehören zur späteren Dokument- und
Ersatzteilsuche.

## Löschen und Auswirkungen

Ein nicht verknüpftes Bild wird ohne zusätzlichen Bestätigungsdialog entfernt. Ein verknüpftes Bild
ist geschützt, bis seine Wartungsverknüpfung entfernt und gespeichert wurde. Wird das Hauptbild
gelöscht, bestimmt RailKeeper automatisch ein verbleibendes Hauptbild.

Das Löschen einer Beilage öffnet einen Bestätigungsdialog mit ihrem Originalnamen. Nach Bestätigung
entfernt RailKeeper den Datenbankeintrag und löscht die gespeicherten Dateidaten, wenn keine andere
Referenz mehr besteht. Ein fehlgeschlagener Löschvorgang wird nicht als Erfolg dargestellt.

Vor umfangreichen Aufräumarbeiten empfiehlt das Kapitel einen aktuellen RailKeeper-Export. Die
Anwendungssicherung umfasst Fahrzeug- und Uploaddaten. Das Kapitel verspricht keine außerhalb von
RailKeeper angelegten Dateikopien oder Browserdownloads zu entfernen.

## Fehler- und Teilzustände

Die englische und deutsche Seite erklären mindestens:

| Situation | Dokumentierte Reaktion |
| --- | --- |
| Fahrzeug noch nicht gespeichert | Erst Grunddatensatz anlegen, dann Uploads verwenden. |
| Keine Bilder oder Beilagen | Stabilen Leerzustand benennen und passende Uploadaktion zeigen. |
| Format nicht erlaubt | Unterstützte Formate nennen und keine Datei als gespeichert behandeln. |
| Datei leer oder zu groß | Servergrenze beachten, Datei korrigieren oder verkleinern. |
| Mehrfachupload scheitert teilweise | Frühere erfolgreiche Dateien bleiben gespeichert, die fehlerhafte und spätere Dateien prüfen und gezielt erneut hochladen. |
| Bildmetadaten nicht gespeichert | Fahrzeugdialog geöffnet lassen, Fehlermeldung beachten und erneut speichern. |
| Wartungsverknüpftes Bild kann nicht gelöscht werden | Verknüpfung entfernen, speichern und Löschung wiederholen. |
| Vorschau nicht verfügbar | Datei herunterladen oder in einem geeigneten lokalen Programm öffnen. |
| Download oder Öffnen schlägt fehl | Sitzung und Verbindung prüfen, Datensatz neu laden und erneut versuchen. |
| Beilagenmetadaten oder Löschen schlägt fehl | Fehler beachten, Datensatz neu laden und nicht von einer erfolgreichen Änderung ausgehen. |

Bei Mehrfachuploads arbeitet RailKeeper sequenziell. Schlägt eine spätere Datei fehl, bleiben zuvor
erfolgreich gespeicherte Dateien erhalten und nachfolgende Dateien werden in diesem Durchlauf nicht
mehr verarbeitet. Dieser Teilzustand wird ausdrücklich erklärt, damit keine unbemerkten Lücken
entstehen.

## Sicherheits- und Speichergrenzen

Das Kapitel beschreibt die benutzerrelevanten Grenzen, ohne interne Speicherpfade offenzulegen:

- Dateien werden lokal im privaten RailKeeper-Datenbestand gespeichert.
- Das Datenverzeichnis wird nicht als öffentliche statische Website bereitgestellt.
- Der Server bereinigt Dateinamen und prüft Endung, MIME-Typ, Inhalt und Größe.
- Ausführbare und skriptartige Dateien bleiben blockiert.
- Lese- und Schreibzugriffe verwenden die bestehende Sitzung und serverseitige Rollenprüfung.
- Betreibergrenzen können strenger sein als die in der Oberfläche genannten Standardwerte.
- Uploads gehören zum Sicherungs- und Wiederherstellungsumfang der Anwendungsdaten.

Konkrete Umgebungsvariablen und betriebliche Konfiguration gehören in das spätere
Administrationskapitel. Das Benutzerkapitel nennt nur Standardwerte und weist auf mögliche
Betreiberabweichungen hin.

## Fachliche Abgrenzung und Coverage

Das Kapitel deckt lokale Fahrzeugbilder, allgemeine Beilagen, Bildpflege, Bildvorschau und deren
öffentliche Fahrzeug-API-Abläufe ab. Der Coverage-Eintrag `vehicle-media` wechselt erst nach
vorhandenem englisch-deutschem Seitenpaar auf `documented`.

Die Übersetzungsschlüssel unter `vehicles.uploads.webDocuments`, `vehicles.uploads.webDocument` und
`vehicles.uploads.extractSpareParts` gehören fachlich zur Web-Dokument- und Ersatzteilsuche. Der
Coverage-Vertrag erhält deshalb spezifischere Präfixe für `vehicle-search-spares`, während der
allgemeinere Präfix `vehicles.uploads` weiterhin `vehicle-media` zugeordnet wird.

Dasselbe gilt für die API-Endpunkte zum Import einer Bild- oder Beilagen-URL. Die spezifischen
Präfixe `/api/v1/vehicles/{id}/images/import-url` und
`/api/v1/vehicles/{id}/attachments/import-url` werden `vehicle-search-spares` zugeordnet. Lokaler
Upload, Metadaten, Download und Löschen bleiben bei `vehicle-media`. Die jeweils längste
Präfixzuordnung bleibt damit eindeutig und verhindert eine falsche Vollständigkeitsbehauptung.

Nicht Bestandteil dieses Kapitels sind:

- Durchführung und Historie von Wartungen
- CV-Dateien, CV-Werte und Decoderprofile
- Suche, Auswahl und Import externer Webdokumente
- Artikeldatensuche und Import externer Bilder
- Extraktion und Pflege von Ersatzteilen
- betriebliche Konfiguration der Uploadgrenzen

## Prüfung und Abnahme

Vor Veröffentlichung werden folgende Nachweise erbracht:

1. Quellenabgleich aller Aussagen gegen den Git-Tag `v0.1.17.6`.
2. Semantisch gleichwertige englische und deutsche Seite mit passenden Metadaten.
3. Navigation, Einstiegsseiten, Querverweise und Coverage-Vertrag aktualisiert.
4. Erwarteter Coverage-Fehler nach Statuswechsel vor dem Anlegen der Seiten.
5. Vollständiger Lauf von `cd docs; npm.cmd run check`.
6. Prüfung auf unfertige Marker, fehlerhafte Links, unerwünschte Buildausgaben und Diff-Fehler.
7. Unabhängiger Read-only-Review mit besonderem Fokus auf Dateitypen, Persistenzzeitpunkt,
   Teiluploads, Rollen, Wartungsverknüpfung und Löschfolgen.
8. Pull Request erst nach erfolgreicher lokaler Prüfung.
9. Merge erst nach erfolgreichen GitHub-Prüfungen für CI, Trivy und CodeQL.

## Abnahmekriterien

Das Kapitel ist fertig, wenn ein neuer Benutzer ohne weitere Quelle beantworten kann:

- wann Bilder und Beilagen hinzugefügt werden können,
- welche Rollen lesen und welche Rollen schreiben dürfen,
- welche Formate und Standardlimits gelten,
- welche Bildänderungen sofort und welche erst nach Speichern wirken,
- wie Hauptbild, Reihenfolge, Beschreibung und Wartungsverknüpfung funktionieren,
- wie Kategorien automatisch oder manuell gesetzt werden,
- welche Dateitypen Vorschau, Öffnen und Download unterstützen,
- wie Bilder und Beilagen sicher entfernt werden,
- wie ein teilweise fehlgeschlagener Mehrfachupload erkannt und vervollständigt wird,
- welche Funktionen bewusst in andere Fachkapitel gehören.
