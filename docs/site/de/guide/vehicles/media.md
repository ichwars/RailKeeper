---
title: Fahrzeugbilder und Beilagen
description: Fahrzeugbilder und Beilagen hochladen, ordnen, anzeigen, herunterladen und sicher entfernen.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Fahrzeugbilder und Beilagen

Der Tab **Uploads** verwaltet lokale Fahrzeugbilder und allgemeine Beilagen am Fahrzeugdatensatz.
Admin, Editor, Viewer und Planner können gespeicherte Medien ansehen. Nur Admin und Editor dürfen
hochladen, Metadaten ändern oder löschen. Der Server prüft diese Regel auch dann, wenn ein
Bedienelement sichtbar ist.

## Tab Uploads öffnen

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann **Uploads**. Das Fahrzeug
muss vorher gespeichert sein: Lokale Bilder und Beilagen benötigen seine gespeicherte Fahrzeug-ID.
Speichere zuerst den Grunddatensatz und fahre danach im Bearbeitungsdialog fort oder öffne ihn neu.
Der Tab kennzeichnet leere Bild- und Beilagenbereiche und bietet die passende Uploadaktion an.

## Lokale Bilder hochladen

Wähle mit **Bild hochladen** eine oder mehrere lokale Dateien. RailKeeper akzeptiert JPG/JPEG, PNG
und WebP. Das Standardlimit des Servers beträgt 10 MB pro Bild, ein Betreiber kann eine strengere
Grenze konfigurieren. Der Browser prüft die gewählte Endung, der Server zusätzlich Dateigröße und
erkannten MIME-Typ.

RailKeeper lädt ausgewählte Bilder nacheinander hoch. Wenn anfangs noch kein Bild vorhanden ist,
markiert jede Anfrage eines Mehrfachuploads ihr Bild als **Hauptbild**. Daher ist die zuletzt
erfolgreich hochgeladene Datei das Hauptbild. Frühere erfolgreiche Dateien bleiben gespeichert,
wenn eine spätere Datei fehlschlägt; weitere spätere Dateien werden in diesem Durchlauf nicht mehr
versucht. Lade oder öffne das Fahrzeug vor einem erneuten Versuch neu. Vergleiche die aktualisierte
Bildliste und das Hauptbild und lade nur weiterhin fehlende Dateien hoch. Der Bildupload wirkt
unmittelbar und wartet nicht auf **Änderungen speichern**.

## Bildmetadaten ordnen

Jedes Bild lässt sich in Originalgröße öffnen, mit einer **Bildbeschreibung** versehen, nach oben
oder unten verschieben oder als **Hauptbild** markieren. Das Hauptbild wird im Fahrzeugbestand und
in kompakten Ansichten bevorzugt verwendet. Du kannst ein Bild auch mit einer vorhandenen Wartung
verknüpfen.

Bildbeschreibung, Reihenfolge, Hauptbild-Auswahl und Wartungsverknüpfung sind Bildmetadaten. Sie
ändern sich zunächst nur im Bearbeitungsdialog und werden erst mit **Änderungen speichern** am
Fahrzeug dauerhaft gespeichert. Schlägt das Speichern fehl, lasse den Dialog geöffnet, lies die
Fehlermeldung und versuche es nach Prüfung von Sitzung und Verbindung erneut.

## Bild mit einer Wartung verknüpfen

Wähle einen vorhandenen Wartungseintrag, wenn das Bild diesen Eintrag dokumentiert. Dieses Kapitel
erklärt nur die Bildverknüpfung, nicht das Bearbeiten der Wartung selbst. Ein mit Wartung
verknüpftes Bild lässt sich nicht sofort entfernen. Wähle zuerst **Keine Wartung**, klicke
**Änderungen speichern** und entferne das Bild danach.

## Allgemeine Beilagen hochladen

Wähle mit **Beilage hochladen** eine oder mehrere Dateien oder lege sie nach dem Speichern des
Fahrzeugs per Drag-and-drop bei **Dateien hier ablegen** ab. Die vor dem Upload gewählte Kategorie
und Bemerkung gelten für jede Datei dieses Uploads. Mit **Kategorie automatisch** bestimmt
RailKeeper für jede Datei eine Kategorie.

Wie Bilder werden Beilagen nacheinander hochgeladen. Eine frühere erfolgreiche Datei bleibt
gespeichert, wenn eine spätere Datei fehlschlägt, und RailKeeper versucht weitere Dateien in diesem
Durchlauf nicht. Lade oder öffne das Fahrzeug neu, vergleiche die Liste gespeicherter Beilagen und
lade nur weiterhin fehlende Dateien hoch. Gespeicherte Beilagen zeigen Originalname, Kategorie,
MIME-Typ und Größe.

## Formate, Grenzen und Kategorien für Beilagen

Standardmäßig akzeptiert RailKeeper PDF, TXT, CSV, JSON, XML, ZIP, JPG/JPEG, PNG und WebP als
Beilagen. Das Standardlimit beträgt 25 MB pro Datei. Ein Betreiber kann eine strengere
Größenbegrenzung konfigurieren und einschränken, welche dieser Dateiendungen der Server akzeptiert.
Leere, ausführbare, nicht erlaubte, inhaltlich blockierte und zu große Dateien werden abgelehnt.
Betrachte eine ausgewählte Datei erst dann als gespeichert, wenn RailKeeper den Upload als
erfolgreich meldet.

Die gespeicherten Kategoriewerte sind `Anleitung`, `Rechnung`, `Decoder-Datei`,
`Dokumentation`, `Ersatzteilliste`, `Zertifikat` und `Sonstiges`.

Die automatische Kategorisierung nutzt diese Priorität. Die erste zutreffende Regel gewinnt:

1. `rechnung` oder `invoice` im Dateinamen: `Rechnung`.
2. `decoder` im Namen oder JSON beziehungsweise XML: `Decoder-Datei`.
3. `ersatzteil` im Namen: `Ersatzteilliste`.
4. `zertifikat` oder `certificate` im Namen: `Zertifikat`.
5. `anleitung`, `manual` oder `bedienung` im Namen: `Anleitung`.
6. Andere PDF-Datei: `Dokumentation`.
7. Jede andere erlaubte Datei: `Sonstiges`.

Kategorie und Bemerkung lassen sich für jede gespeicherte Beilage einzeln ändern. Diese Änderungen
werden nur über die Speichern-Aktion der jeweiligen Zeile gespeichert, nicht über
**Änderungen speichern** am Fahrzeug. Beim Speichern einer Beilagenzeile wird das Fahrzeug neu
geladen, wodurch andere ungespeicherte Medienänderungen verworfen werden. Speichere zuerst
ausstehende Bildmetadaten. Bearbeite und speichere danach jeweils eine Beilagenzeile, bevor du die
nächste änderst.

## Beilagen anzeigen, öffnen und herunterladen

**Anzeigen** öffnet die integrierte Vorschau von RailKeeper. Unterstützt werden PDF und Bilder
sowie TXT, CSV, JSON und XML. ZIP-Dateien haben keine Inline-Vorschau. Lade sie herunter und öffne
sie mit einem geeigneten lokalen Programm. **Datei öffnen** öffnet eine geeignete
Inline-Darstellung in einem neuen Browser-Tab. **Datei herunterladen** lädt die gespeicherte
Originaldatei herunter. Dies sind getrennte Aktionen: Du kannst in RailKeeper prüfen, im Browser
öffnen oder eine lokale Kopie behalten.

## Bilder und Beilagen löschen

Das Entfernen eines gespeicherten Bildes wirkt sofort und benötigt keine zusätzliche Bestätigung.
War es das Hauptbild, bestimmt RailKeeper das nächste Bild nach Sortierreihenfolge zum Hauptbild.
Ein verknüpftes Bild muss zuerst wie oben beschrieben getrennt und gespeichert werden. Beim Löschen
einer Beilage erscheint **Beilage löschen?**. Bestätige erst nach Prüfung des Originaldateinamens.
Schlägt die Löschung fehl, lade den Datensatz neu und gehe nicht davon aus, dass Datei oder
Metadaten entfernt wurden.

## Rollen-, Speicher- und Sicherungsgrenzen

Medien sind lokale und private RailKeeper-Daten, keine öffentlichen Website-Inhalte.
Anwendungssicherungen enthalten gespeicherte Fahrzeugmedien. RailKeeper akzeptiert in v0.1.19.2
jedoch nur Sicherungsdateien bis 250 MiB für Prüfung und Wiederherstellung. Exportiere vor
umfangreichen Aufräumarbeiten eine Sicherung und bitte einen Admin, diese Datei im
Wiederherstellungsbereich auszuwählen. Verlasse dich erst darauf, wenn RailKeeper sie als kompatibel
akzeptiert. Diese Seite verspricht nichts über Kopien, die du zuvor heruntergeladen oder außerhalb
von RailKeeper angelegt hast.

Admin, Editor, Viewer und Planner können vorhandene Medien ansehen. Serverseitige Schreibzugriffe
erfordern Admin oder Editor. **Quelle öffnen** ist der Übergang für importierte Bildquellen. Externe
Artikelbilder, **Gefundene Dokumente**, Web-Dokument-Import und **Ersatzteile extrahieren** erklärt
[Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares). CV-Dateien und
das Bearbeiten von Wartungen bleiben Fachgrenzen, die auf dieser Seite nicht erklärt werden.

## Leere, teilweise und fehlerhafte Zustände

| Situation | Vorgehen |
| --- | --- |
| Fahrzeug ist nicht gespeichert | Speichere den Grunddatensatz und verwende dann **Uploads**. |
| Keine Bilder oder Beilagen | Nutze die Uploadaktion des Leerzustands am richtigen Fahrzeug. |
| Format ist nicht erlaubt | Wähle eines der genannten Formate. Die abgelehnte Datei wird nicht gespeichert. |
| Datei ist leer oder zu groß | Korrigiere oder verkleinere die Datei und beachte die Servergrenze. |
| Mehrfachupload schlägt teilweise fehl | Frühere Dateien bleiben gespeichert. Lade das Fahrzeug neu, vergleiche die aktualisierte Liste und lade nur weiterhin fehlende Dateien hoch. |
| Bildmetadaten werden nicht gespeichert | Lasse den Fahrzeugdialog offen, lies den Fehler und nutze erneut **Änderungen speichern**. |
| Verknüpftes Bild kann nicht gelöscht werden | Wähle **Keine Wartung**, speichere das Fahrzeug und entferne danach das Bild. |
| Vorschau ist nicht verfügbar | Lade die Datei herunter oder öffne sie mit einem geeigneten lokalen Programm. |
| Öffnen oder Herunterladen schlägt fehl | Prüfe Sitzung und Verbindung, lade den Datensatz neu und versuche es erneut. |
| Beilagenmetadaten oder Löschen schlägt fehl | Lies den Fehler, lade den Datensatz neu und gehe nicht von einer erfolgreichen Änderung aus. |

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Ersteinrichtung und Anmeldung](/de/guide/getting-started/)
- [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
- [Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.19.2** und wurde zuletzt am 16.08.2026 geprüft.
