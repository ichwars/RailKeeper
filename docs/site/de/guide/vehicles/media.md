---
title: Fahrzeugbilder und Beilagen
description: Fahrzeugbilder und Beilagen hochladen, ordnen, anzeigen, herunterladen und sicher entfernen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-06
---

# Fahrzeugbilder und Beilagen

1. Öffne ein gespeichertes Fahrzeug unter **Bearbeiten > Uploads**.
2. Lade ein Foto als **Bild** oder eine Anleitung als **Beilage** hoch.
3. Warte auf die Erfolgsmeldung und prüfe die gespeicherte Datei. Änderungen an Bildmetadaten mit dem Fahrzeug speichern. [Bild für eine Zentrale erstellen](locomotive-images).

<GuideFigure name="media" alt="Unter Uploads hinterlegst du Fahrzeugbilder und Dokumente." caption="Unter Uploads hinterlegst du Fahrzeugbilder und Dokumente. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

Der Tab **Uploads** verwaltet lokale Fahrzeugbilder und allgemeine Beilagen am Fahrzeugdatensatz.
Admin, Editor, Viewer können gespeicherte Medien ansehen. Nur Admin und Editor dürfen
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

Wähle höchstens 20 Bilder pro Vorgang. Prüfe im Bestätigungsdialog Fahrzeug, Dateien und
Wartungsverknüpfung. Erst **Ausgewählte Dateien speichern** beginnt den Upload. Abbrechen davor
speichert nichts. Die Dateien werden nacheinander gespeichert. Ist der Bildentwurf anfangs leer,
bleibt das zuletzt erfolgreich gespeicherte Bild Hauptbild.

Das Ergebnis unterscheidet **Gespeichert**, **Abgelehnt**, **Noch nicht versucht**, **Übersprungen**
und **Antwort unklar**. Nach einer Ablehnung kannst du die Datei überspringen und offene Dateien
fortsetzen. Bestätigte Dateien werden nicht erneut gesendet. Bei unklarer Antwort prüfe zuerst
die aktualisierte Dateiliste, bevor du eine fehlende Datei neu auswählst. Bei einer parallelen
Bildänderung wird die alte Bestätigung gesperrt; prüfe dann die Liste und wähle nur fehlende Bilder neu.
Gespeicherte Dateien bleiben beim Abbrechen des Fahrzeugeditors erhalten.

## Bildmetadaten ordnen

Jedes Bild lässt sich in Originalgröße öffnen, mit einer **Bildbeschreibung** versehen, nach oben
oder unten verschieben oder als **Hauptbild** markieren. Das Hauptbild wird im Fahrzeugbestand und
in kompakten Ansichten bevorzugt verwendet. Du kannst ein Bild auch mit einer vorhandenen Wartung
verknüpfen.

Bildbeschreibung, Reihenfolge, Hauptbild-Auswahl und Wartungsverknüpfung sind Bildmetadaten. Sie
ändern sich zunächst nur im Bearbeitungsdialog und werden erst mit **Änderungen speichern** am
Fahrzeug dauerhaft gespeichert. Schlägt das Speichern fehl, lasse den Dialog geöffnet, lies die
Fehlermeldung und versuche es nach Prüfung von Sitzung und Verbindung erneut.

## Lokbild für eine Digitalzentrale erstellen

Wähle an einer Bildkarte **Zentralenbild erstellen**. Der gemeinsame Dialog lädt das lokale
Original und prüft Auflösung, Ausschnitt und technische Verwendbarkeit. Bei nötiger Vergrößerung,
wenig genutzter Bildfläche oder großen transparenten Rändern erscheinen konkrete Hinweise.
Defekte, zu große oder vollständig transparente Bilder lassen sich nicht exportieren.

| Ziel | Ausgabedatei und Verwendung |
| --- | --- |
| ESU ECoS 50200 / 50210 / 50220 | Unkomprimiertes BMP, 24 Bit RGB, 190 × 40 Pixel. Im ECoS-Webinterface unter Lokbilder hochladen; 50200 ab Firmware 3.3.0. |
| Märklin CS3 / CS3 plus | PNG, Voreinstellung 384 × 132 Pixel. Im CS3-Webinterface hochladen. Die CS3 erlaubt andere Größen; unsere Voreinstellung ist keine Herstellergrenze. |
| Uhlenbrock Intellibox 3 | PNG, 340 × 72 Pixel nach den Beispielen des Herstellerpakets 1.3.03. Im Lokbilder-Editor von **IB3 Konfiguration** öffnen; das Programm wandelt ins Geräteformat um und überträgt. |
| Z21-App | PNG, Voreinstellung 960 × 400 Pixel. Auf dem Mobilgerät speichern und über die Bildauswahl der App zuordnen. Die Zentrale selbst speichert diese Bilder nicht. |

Ziehe einen Rahmen um die Lok oder öffne **Ausschnitt in Pixeln einstellen** für die Tastaturbedienung.
**Ganzes Bild** setzt den Ausschnitt zurück. **Transparente Ränder entfernen** schlägt bei geeigneten
PNG/WebP-Bildern einen engeren Ausschnitt vor. Prüfe dabei auch feine Teile wie Pantographen.
Die Proportionen bleiben erhalten. Der Hintergrund füllt transparente Bereiche und freien Rand;
ein fotografierter Hintergrund wird dadurch nicht entfernt.

Prüfe das **Ergebnis in Originalgröße**, besonders die vollständige Sichtbarkeit und Erkennbarkeit
der Lok. Große Vorschauen können innerhalb ihres Rahmens gescrollt werden. **Technisch verwendbar**
bedeutet, dass die Datei aufbereitet werden kann; Motiv, Perspektive und Schärfe werden nicht
automatisch erkannt. Die Herstellerquelle ist im Dialog verlinkt.

**Bild herunterladen** erstellt eine separate Datei mit Zielkürzel im Namen. Original, Bildmetadaten
und Fahrzeugdaten bleiben erhalten. Es erfolgt keine automatische Übertragung an die Zentrale.
Extern verlinkte Artikelbilder müssen zuerst lokal importiert oder hochgeladen werden.
Die Aufbereitung läuft im Browser, ohne Übermittlung an einen Bild- oder KI-Dienst.

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

Auch Beilagen werden vor dem Upload bestätigt, mit höchstens 20 Dateien und den festgehaltenen
Angaben. Die Ergebnisliste und Fortsetzen-Funktion verhalten sich wie bei Bildern. Bereits
bestätigte Dateien bleiben gespeichert. Dateiauswahl und erfolgreiche Bilduploads erhalten offene
Profil- und Bemerkungsfelder für Decoderdateien.

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
**Änderungen speichern** am Fahrzeug. Das anschließende Aktualisieren erhält offene Fahrzeug-,
Funktions-, Bild- und weitere Anhangsentwürfe. **Abbrechen** verwirft diese offenen Angaben und erhält
bereits gespeicherte Einzelaktionen.

## Parallele Änderungen vergleichen

Hat jemand inzwischen andere Werte gespeichert, bleibt dein Entwurf erhalten. Ein Vergleich zeigt
deine Angaben und die gespeicherten Werte. Wähle für widersprüchliche Felder ausdrücklich die
gewünschte Seite. Das gilt auch für konkurrierende Hauptbilder und Bildreihenfolgen. Unberührte
Angaben werden aktualisiert. Bildkonflikte müssen vor **Änderungen speichern** geklärt werden;
Anhangskonflikte vor der eigenen Speicheraktion des betroffenen Anhangs. Noch offene Angaben zu einem
inzwischen entfernten Medium bleiben zum Prüfen sichtbar und können ausdrücklich verworfen werden.

## Beilagen anzeigen, öffnen und herunterladen

**Anzeigen** öffnet die integrierte Vorschau von RailKeeper. Unterstützt werden PDF und Bilder
sowie TXT, CSV, JSON und XML. ZIP-Dateien haben keine Inline-Vorschau. Lade sie herunter und öffne
sie mit einem geeigneten lokalen Programm. **Datei öffnen** öffnet eine geeignete
Inline-Darstellung in einem neuen Browser-Tab. **Datei herunterladen** lädt die gespeicherte
Originaldatei herunter. Dies sind getrennte Aktionen: Du kannst in RailKeeper prüfen, im Browser
öffnen oder eine lokale Kopie behalten.

## Bilder und Beilagen löschen

Bilder und Beilagen werden erst nach **Gespeicherte Datei löschen?** entfernt. Prüfe Fahrzeug
und Dateiname. Eine inzwischen geänderte Datei erfordert nach Aktualisieren der Liste eine neue
Bestätigung. Ein verknüpftes Bild muss zuerst wie oben beschrieben getrennt und gespeichert werden.
War das gelöschte Bild Hauptbild, bestimmt RailKeeper das nächste Bild nach Sortierreihenfolge.
Die bestätigte Löschung bleibt beim Abbrechen des Fahrzeugeditors bestehen.

## Rollen-, Speicher- und Sicherungsgrenzen

Medien sind lokale und private RailKeeper-Daten, keine öffentlichen Website-Inhalte.
Anwendungssicherungen enthalten gespeicherte Fahrzeugmedien. RailKeeper akzeptiert derzeit
jedoch nur Sicherungsdateien bis 250 MiB für Prüfung und Wiederherstellung. Exportiere vor
umfangreichen Aufräumarbeiten eine Sicherung und bitte einen Admin, diese Datei im
Wiederherstellungsbereich auszuwählen. Verlasse dich erst darauf, wenn RailKeeper sie als kompatibel
akzeptiert. Diese Seite verspricht nichts über Kopien, die du zuvor heruntergeladen oder außerhalb
von RailKeeper angelegt hast.

Admin, Editor, Viewer können vorhandene Medien ansehen. Serverseitige Schreibzugriffe
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
| Mehrfachupload schlägt teilweise fehl | Prüfe die Einzelergebnisse, überspringe abgelehnte Dateien bei Bedarf und setze offene Dateien fort. Unklare Antworten zuerst in der Dateiliste prüfen. |
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
