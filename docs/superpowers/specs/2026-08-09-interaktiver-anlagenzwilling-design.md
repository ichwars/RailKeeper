# Interaktiver Anlagenzwilling, Etappe 2

**GitHub-Issue:** #35

**Status:** Lokaler Arbeitsentwurf auf Basis der freigegebenen Stage-1-Architektur

## Ziel

Etappe 2 macht Anlagen und Module erstmals als interaktive, maßbezogene Arbeitsfläche sichtbar.
Anlageneinheiten erhalten Konturen und frei platzierbare technische Positionen. Eine gewählte
Aufbaukonfiguration setzt die Einheiten zu einer Gesamtansicht zusammen. Statuslayer, Filter,
Hoverinformationen und ein Inspector verbinden Planung, Reservierung, Einbau, Wartung und Defekte,
ohne Steuerbefehle an die Modellbahn zu senden.

Die Umsetzung bleibt lokal, selbst gehostet und SQLite-basiert. Sie erweitert die bestehende
Stage-1-Struktur, ersetzt weder Planrevisionen noch Zubehörbestand und greift der geometrischen
Gleisplanung aus Etappe 3 nicht vor.

## Abgrenzung

Etappe 2 enthält:

- Konturen von Anlageneinheiten im lokalen Millimeter-Koordinatensystem,
- technische Positionen mit Bezeichnung, Typ, Koordinate und optionaler Produktreferenz,
- eine transformierte Gesamtansicht der gewählten Aufbaukonfiguration,
- live abgeleitete Planungs-, Material- und Betriebszustände,
- Filter, Legende, Hoverinformationen und Inspector,
- einen ausdrücklich aktivierten Bearbeitungsmodus,
- eine touch-taugliche reine Anzeige ohne unbeabsichtigte Änderungen,
- Einbau- und Zustandsverlauf im Inspector.

Nicht enthalten sind:

- Gleisgeometrien, Anschlussfang und Fahrwege,
- Tillig-Kataloggeometrien,
- Flexgleise, Höhenprofile, Steigungen oder Durchfahrtshöhen,
- automatische Moduloptimierung,
- digitale Steuerbefehle,
- Cloud-Synchronisation oder öffentliche Freigabe.

## Geprüfte Ansätze

### Frontend-only aus vorhandenen Daten

Einheiten könnten bereits als Rechtecke aus Breite und Höhe dargestellt werden. Technische Marker
wären jedoch nicht dauerhaft speicherbar und könnten nicht zuverlässig mit Reservierungen oder
Einbauten verbunden werden. Dieser Ansatz erfüllt Issue #35 nicht.

### Großes JSON-Dokument pro Planrevision

Ein freies JSON-Dokument wäre schnell erweiterbar. Live-Zustände, referenzielle Integrität,
zielgerichtete Abfragen, Backup-Prüfung und die spätere Migration zum Gleisplaner würden jedoch
unnötig schwach. Dieser Ansatz wird nicht verwendet.

### Normalisierte technische Positionen je Anlageneinheit

Technische Positionen werden als eigene Fachobjekte an Anlageneinheiten gespeichert. Konturpunkte
liegen geordnet und maßbezogen vor. Aufbaukonfigurationen transformieren beide in die Gesamtansicht.
Reservierungen und Einbauten werden über Referenztabellen verknüpft, sodass historische Zuordnungen
erhalten bleiben und Betriebszustände live abgeleitet werden. Dieser Ansatz wird umgesetzt.

## Fachmodell

### Kontur

Jede Anlageneinheit besitzt genau eine äußere Kontur aus mindestens drei geordneten Punkten. Die
Punkte verwenden Millimeter im lokalen Koordinatensystem der Einheit. Solange keine eigene Kontur
hinterlegt ist, erzeugt die Anwendung aus `width_mm` und `height_mm` ein Rechteck.

Ausschnitte, mehrere Teilflächen und Kurven sind nicht Bestandteil dieser Etappe. Sie folgen mit der
erweiterten Geometrie. Eine benutzerdefinierte Kontur verändert nicht automatisch Breite und Höhe.

### Technische Position

Eine technische Position beschreibt einen dokumentierten Ort auf einer Anlageneinheit:

- stabile ID,
- Anlageneinheit,
- Bezeichnung,
- Typ,
- X- und Y-Koordinate in Millimetern,
- Drehung in Grad,
- optionale Zubehörproduktreferenz,
- optionale Beschreibung,
- Version für optimistische Nebenläufigkeitskontrolle,
- Archivstatus und Zeitstempel.

Typen bleiben bewusst auf dokumentierte Anlagenobjekte begrenzt: Weiche, Signal, Rückmelder,
Decoder, Beleuchtung, Stromversorgung, Sensor und Sonstiges. Sie beschreiben keine ausführbare
Steuerfunktion.

Eine Position kann mehrere Zubehörzuordnungen bündeln, beispielsweise Gleis, Antrieb und Decoder.
Eine Reservierung oder Installation gehört höchstens zu einer technischen Position. Die vorhandene
Zuordnung zu Anlage oder Anlageneinheit bleibt maßgeblich und wird serverseitig auf Konsistenz geprüft.

### Zustandsachsen

Die API liefert getrennte Zustände:

- `planned`: Position ohne aktive Reservierung oder Installation,
- `reserved`: mindestens eine aktive Reservierung,
- `installed`: mindestens eine aktive Installation,
- `maintenance_due`: mindestens eine aktive Installation mit fälliger Wartung,
- `defective`: mindestens eine aktive defekte Installation.

`maintenance_due` und `defective` ergänzen den Materialzustand und ersetzen ihn nicht. Eine Position
kann deshalb gleichzeitig eingebaut und defekt sein. Die Oberfläche zeigt neben Farben immer Symbol
und Text.

## Speicherung und Migration

Die nächste Migration ergänzt:

- `layout_unit_outline_points`,
- `layout_technical_positions`,
- `accessory_reservation_positions`,
- `accessory_installation_positions`.

Konturpunkte verwenden `(layout_unit_id, point_index)` als Schlüssel. Technische Positionen besitzen
eine eigene Version. Referenztabellen halten die historische Verbindung zu Reservierungen und
Installationen. Fremdschlüssel bleiben restriktiv, damit referenzierte Positionen nicht versehentlich
gelöscht werden. Fachlich wird archiviert statt gelöscht.

Das Anwendungsbackup erhält Version 4. Ältere Backups bleiben importierbar und erzeugen leere
Stage-2-Tabellen. Version-4-Backups verlangen die vollständigen neuen Tabellen. Benutzer, Rollen,
Sitzungen, Rate-Limits, Auditdaten und Passwortdaten bleiben ausgeschlossen.

## API und Berechtigungen

### Endpunkte

- `GET /layouts/{id}/twin?configurationId=...` liefert die transformierte Leseansicht.
- `GET /layout-units/{id}/technical-positions` listet Positionen im lokalen Koordinatensystem.
- `POST /layout-units/{id}/technical-positions` legt eine Position an.
- `PUT /layout-technical-positions/{id}` aktualisiert eine Position mit `expectedVersion`.
- `PUT /layout-units/{id}/outline` ersetzt die Kontur atomar mit `expectedVersion` der Einheit.
- Reservierungs- und Installationsinputs erhalten optional `technicalPositionId`.

Die Twin-Leseansicht enthält Konfiguration, transformierte Einheiten, Konturen, Positionen,
Zustandsachsen, verknüpfte Produkt- und Bestandsdaten sowie kompakte Verlaufssummen. Detailverläufe
werden erst beim Öffnen des Inspectors geladen, damit die Gesamtansicht kompakt bleibt.

### Rollen

- Viewer, Editor, Planner und Admin dürfen den Anlagenzwilling lesen.
- Planner und Admin dürfen Konturen und technische Positionen verwalten.
- Editor und Admin dürfen weiterhin Einbauten und Betriebszustände pflegen.
- Planner darf Reservierungen anlegen, aber keine Installation bestätigen.
- Messe erhält keinen allgemeinen Zugriff auf den Anlagenzwilling.

Alle Schreibzugriffe bleiben CSRF-geschützt, serverseitig autorisiert und auditiert.

## Frontend

### Übersicht

Das Register `Übersicht` erhält oberhalb des Profils die interaktive Arbeitsfläche. Bei mehreren
Aufbaukonfigurationen steht eine kompakte Auswahl zur Verfügung. Ohne Konfiguration zeigt RailKeeper
eine einzelne auswählbare Anlageneinheit in ihrem lokalen Koordinatensystem.

Die Arbeitsfläche bietet:

- Zoom auf Inhalt und Zurücksetzen,
- Statusfilter und Legende,
- eindeutige Symbole pro Positionstyp,
- Hoverinformation auf Desktop,
- Klick oder Tippen zum Öffnen des Inspectors,
- einen klaren Leerzustand ohne Einheiten oder Positionen.

Die Darstellung verwendet SVG. Millimeterkoordinaten bleiben im ViewBox-System erhalten. Die
Aufbaukonfiguration liefert Translation und Rotation der Einheiten. HTML-Overlays werden nicht für
die Geometrie verwendet.

### Inspector

Der Inspector erscheint als app-eigene Seitenfläche auf Desktop und als app-eigener Dialog auf
schmalen Ansichten. Er zeigt:

- Bezeichnung, Typ, Einheit und Koordinate,
- Material- und Betriebsstatus,
- Produkt, Inventarnummer beziehungsweise Menge,
- technische Angaben wie Adresse, Ausgang, Anschluss und Verdrahtung,
- aktive Reservierungen und Einbauten,
- Einbau- und Zustandsverlauf,
- rollenabhängige Aktionen.

### Bearbeitungsmodus

Die reine Anzeige ist der Standard. Nur Planner und Admin sehen `Bearbeiten`. Erst nach Aktivierung
werden Positionen ziehbar und Konturpunkte bearbeitbar. Änderungen werden mit kurzer Verzögerung
gespeichert. Während einer laufenden Speicherung bleibt der betroffene Marker gesperrt. Konflikte
überschreiben nichts still, sondern bieten Serverstand laden und lokalen Entwurf verwerfen.

Touch-Gesten in der Anzeige bewegen oder vergrößern nur die Ansicht. Sie verändern keine Fachobjekte.

## Datenfluss

1. Die Übersicht lädt Anlagen, Einheiten und Aufbaukonfigurationen wie bisher.
2. Für die gewählte Konfiguration lädt sie die Twin-Leseansicht.
3. Das Backend transformiert Kontur und Positionen aus dem lokalen Einheitensystem in die
   Konfigurationsansicht.
4. Das Backend verbindet aktive Reservierungen und Installationen und leitet beide Zustandsachsen ab.
5. Auswahl und Filter bleiben rein lokal im Frontend.
6. Der Inspector lädt bei Bedarf Detailverläufe.
7. Änderungen im Bearbeitungsmodus verwenden versionsgeprüfte Einzeloperationen.

## Fehler- und Leerzustände

- Fehlt eine Aufbaukonfiguration, kann eine einzelne Einheit angezeigt werden.
- Einheiten ohne Maße oder Kontur erscheinen in einer separaten unvollständigen Liste.
- Positionen außerhalb der Kontur werden gespeichert, aber klar als Warnung markiert.
- Fehlende Produktreferenzen verhindern die Anzeige nicht.
- Versionskonflikte behalten den lokalen Entwurf bis zur bewussten Entscheidung.
- Nicht auflösbare Konfigurationsreferenzen erscheinen als Warnung und werden nicht still verworfen.
- API-, Offline- und Ladefehler bleiben im betroffenen Panel sichtbar und blockieren nicht die übrige
  Anlagenarbeitsmappe.

## Arbeitspakete

### Paket A: Fachmodell und vertikale Positionserfassung

Migration, Repository, Anwendung, API, OpenAPI, typisierter Client und eine kompakte Positionsliste
im Anlagenregister `Technik`. Konturen verwenden zunächst das Rechteck aus Einheitsmaßen. Dieses
Paket liefert bereits Anlegen, Bearbeiten, Archivieren und Produktverknüpfung. Backup-Version 4
nimmt die neuen Tabellen sofort vollständig auf, damit kein Zwischenstand Stage-2-Daten auslässt.

### Paket B: Interaktive Twin-Übersicht

Transformierte SVG-Gesamtansicht, Konfigurationsauswahl, Statuslayer, Filter, Hover und Inspector.
Reservierungen und Einbauten werden mit Positionen verknüpft und live dargestellt.

### Paket C: Expliziter Bearbeitungsmodus und Konturen

Ziehen technischer Positionen, benutzerdefinierte Konturpunkte, verzögertes Speichern,
Konfliktbehandlung sowie Desktop- und Touch-Interaktion.

### Paket D: Verlauf, Wartung und Abnahme

Einbau- und Zustandsverlauf im Inspector, Rollenabnahme, vollständiger Backup-/Restore-Roundtrip,
Migrationstests, Deutsch/Englisch, Hell/Dunkel, Desktop/Mobil und lange deutsche Texte.

## Tests und Abnahme

Erforderlich sind:

- Domain- und Repositorytests für Koordinaten, Konturen, Versionen und Referenzen,
- Migrationstests einschließlich Fremdschlüsseln und älteren Datenständen,
- API-Tests für Rollenmatrix, CSRF, Validierung und Konflikte,
- OpenAPI-Vertragsprüfung und Clienttests,
- Backup-/Restore-Roundtrip für Version 4 und Kompatibilität mit Versionen 1 bis 3,
- Frontendtests für Filter, Inspector, Bearbeitungsmodus, Fokus und Rollen,
- Transformationstests für Translation und Rotation,
- visuelle Prüfung in Hell/Dunkel sowie Desktop und schmaler Ansicht,
- vollständige Go-Tests, Frontendtests, Build und `git diff --check`.

Issue #35 gilt als lokal umgesetzt, wenn Konturen und technische Positionen gespeichert, in einer
Aufbaukonfiguration korrekt transformiert, mit Live-Zuständen inspiziert und nur im ausdrücklichen
Bearbeitungsmodus verändert werden können.
