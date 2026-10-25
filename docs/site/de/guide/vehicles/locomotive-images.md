---
title: Lokbilder für die Digitalzentrale
description: Lokbilder für die Digitalzentrale.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-25
---

# Lokbilder für die Digitalzentrale

**Ziel:** Aus einem Fahrzeugbild eine passende Bilddatei erstellen und sie bei Bedarf einer Lok
auf der ECoS zuordnen. Das Originalbild bleibt erhalten. Für den Direkttransfer brauchst du
**Admin**; Bilder am Fahrzeug pflegen können **Admin** und **Editor**.

Die Aufnahmen zeigen die echte RailKeeper-Oberfläche mit einer künstlich gezeichneten Demo-Lok.
Bei den ECoS-Vorschau-, Pausen- und Ergebnisbildern sind die Zentralenantworten simuliert.
Sie erklären die Bedienung und sind kein zusätzlicher Hardwaretest.

## 1. Ein geeignetes Bild auswählen

Öffne ein gespeichertes Fahrzeug unter **Bearbeiten > Uploads**. Hinterlege ein Foto als
Fahrzeugbild. Eine Seitenaufnahme mit ruhigem Hintergrund eignet sich besonders gut.
Die Lok sollte vollständig sichtbar sein und möglichst wenig freien Rand besitzen.
Eine schräge Perspektive oder ein unscharfes Foto lässt sich durch Vergrößern nicht reparieren.

Zum Ausprobieren kannst du diese künstlichen Bilddateien verwenden:

- <GuideDownload file="locomotive-demo.png">Demo-Lok mit ausreichender Auflösung</GuideDownload> (1200 × 400 Pixel).
- <GuideDownload file="locomotive-low-resolution-demo.png">Demo-Lok mit zu geringer Auflösung</GuideDownload> (40 × 14 Pixel).

Wähle am gewünschten Fahrzeugbild **Zentralenbild erstellen**.

## 2. Ziel, Ausschnitt und Hintergrund einstellen

Wähle unter **Zielgerät / Anwendung** die gewünschte Ausgabe. Bei Admins wird für gespeicherte
Fahrzeuge die konfigurierte Zentrale vorbelegt. Links siehst du den Ausschnitt, rechts das
vergrößerte **Ergebnis**. **1:1 anzeigen** zeigt die tatsächliche Ausgabegröße.
Formatdetails und Herstellerangaben lassen sich unter der Gerätewahl aufklappen.

<GuideFigure name="image-export" alt="Lokbild-Dialog mit Ausschnitt links und fertiger ECoS-Vorschau rechts." caption="Die große Demo-Datei wird im echten Bilddialog verarbeitet. Die rechte Vorschau ist vergrößert; die Ausgabedatei behält das Geräteformat. Künstliches Motiv; anklicken für die volle Größe." />

1. Prüfe den automatisch vorgeschlagenen Ausschnitt.
2. Verwende **Außenränder entfernen**, wenn gleichmäßige freie Ränder stören.
3. Ziehe bei Bedarf mit Maus oder Finger einen anderen Ausschnitt. Unter **Ausschnitt in Pixeln
   einstellen** kannst du ihn genauer festlegen. **Ganzes Bild** setzt die Auswahl zurück.
4. Prüfe, dass Puffer, Räder und Aufbauten erhalten bleiben. Nutze möglichst die verfügbare Breite.
5. **Automatisch** wählt den Hintergrund passend zum Ziel: Grau für ECoS, transparent für
   CS3/Intellibox 3 und Weiß für die Z21-App. Eine andere Farbe lässt sich weiterhin auswählen.
6. **Gleichmäßigen Fotohintergrund ersetzen** passt erkannte, mit dem Bildrand verbundene
   Hintergrundflächen an diese Farbe an. Abschalten behält den Fotohintergrund. Bei uneindeutigen
   Hintergründen wird das Foto beibehalten. Das gespeicherte Original bleibt immer erhalten.

RailKeeper skaliert innerhalb der Zielgröße unter Beibehaltung der Proportionen.
Ein hohes Motiv kann nicht gleichzeitig die volle Breite ausfüllen und vollständig sichtbar bleiben.
Verbessere dann die Aufnahme oder den Ausschnitt; ziehe die Lok nicht in die Breite.

| Ziel | Ausgabe und weiterer Weg |
| --- | --- |
| ECoS 50200 / 50210 / 50220 | 190 × 40 Pixel, BMP mit 24 Bit RGB; festes Geräteformat. Manuell importieren oder den bestätigten Direkttransfer verwenden. |
| CS3 / CS3 plus | PNG, RailKeeper-Voreinstellung 384 × 132 Pixel; im CS3-Webinterface importieren. |
| Intellibox 3 | PNG, Voreinstellung 340 × 72 Pixel; mit IB3 Konfiguration weiterverarbeiten und übertragen. |
| Z21-App | PNG, Voreinstellung 960 × 400 Pixel; auf dem Mobilgerät in der App beim Fahrzeug auswählen. |

Die PNG-Voreinstellungen sind keine allgemeingültigen Herstellergrenzen. Die Z21-Zentrale selbst
speichert diese App-Bilder nicht. Im Dialog führen **Herstellerangaben** zur jeweiligen Quelle.

## 3. Automatische Aufbereitung

RailKeeper entfernt gleichmäßige Außenränder, passt die Proportionen in das Geräteformat ein
und prüft Bildgröße, Ausschnitt und Dateiausgabe automatisch. Ein ungültiger Ausschnitt oder ein
nicht verarbeitbares Bild verhindert die Ausgabe und zeigt eine Fehlermeldung.

<GuideFigure name="image-low-resolution" alt="Automatisch aufbereitete kleine Demo-Datei im Bilddialog." caption="Auch das kleine Beispiel wird verarbeitet. Die vergrößerte Vorschau fügt der Datei keine fehlenden Bilddetails hinzu." />

Für scharfe Details möglichst das größere Original verwenden. Perspektive und unscharfe
Motivdetails werden durch Skalieren nicht repariert.

## 4. Datei herunterladen und manuell importieren

Wähle **Bild herunterladen**. Verwende für die ECoS die erzeugte BMP-Datei, nicht die originale PNG.
Öffne ihre Weboberfläche, gehe zu **Lokbilder** und lade die Datei auf einen freien Benutzerbildplatz.
Ordne das Bild danach an der ECoS der richtigen Lok zu. Ein hochgeladenes Bild ist allein noch
keine bestätigte Lokzuordnung.

Bei CS3, Intellibox 3 und Z21-App verwendest du den in der Tabelle genannten Weg.
Der folgende Direkttransfer ist ausschließlich für die ECoS vorgesehen.

## 5. Direkt an die ECoS übertragen

Die ECoS muss [eingerichtet und erreichbar](../../administration/digital-centers) sein.
Die Protokollverbindung benötigt TCP-Port **15471**, der Bildtransfer zusätzlich die
Weboberfläche auf Port **80**, jeweils vom RailKeeper-Server aus.

1. Wähle **Direkt an ECoS übertragen …**.
2. Wähle die **Ziel-Lok auf der ECoS**. Prüfe Name, Decoderadresse, Protokoll und Objekt-ID.
3. Prüfe die Bildbezeichnung und wähle **Übertragung prüfen und Vorschau laden**.
4. Vergleiche bisheriges und neues Bild sowie den angezeigten Bildplatz. Diese Vorschau lädt
   noch kein neues Bild hoch.

<GuideFigure name="image-ecos-preview" alt="ECoS-Zielauswahl mit Vorher-/Nachher-Bild und noch nicht bestätigtem Transfer." caption="Echte RailKeeper-Vorschau mit simulierten Zentralenantworten und reservierter Beispieladresse 192.0.2.10. Kein Zugriff auf eine echte ECoS." />

5. Wähle die Ziel-Lok an beiden ECoS-Fahrreglern und in anderen Steuerprogrammen ab.
6. Setze **Ziel-Lok und Bildvorschau geprüft** und wähle **Jetzt hochladen und zuordnen**.
7. Warte auf das Ergebnis. RailKeeper fordert zuerst die Loksteuerung an und prüft nach dem
   Transfer Bild und Zuordnung durch Zurücklesen.

Ist bereits ein identisches Bild vorhanden, zeigt die Vorschau dessen Bildplatz und bietet
**Vorhandenes Bild jetzt zuordnen** an. Dabei wird kein zweites Bild hochgeladen.
Seine vorhandene Bezeichnung auf der ECoS bleibt erhalten.

## 6. Bei belegter Loksteuerung fortsetzen

<GuideFigure name="image-ecos-pause" alt="Pausierter Bildtransfer mit Hinweis auf die belegte Loksteuerung und Fortsetzen-Schaltfläche." caption="Simulierter Belegungsfall im echten Dialog. Bis hier wurde weder ein neues Bild hochgeladen noch eine Zuordnung geschrieben." />

Wähle die Lok an beiden Fahrreglern und gegebenenfalls in anderen Steuerprogrammen ab.
Klicke dann bewusst auf **Fortsetzen**. RailKeeper prüft eine neue Vorschau und setzt den bereits
bestätigten Vorgang fort, wenn Ziel, Bild und Bildplatz unverändert sind. Bei Änderungen musst du
die neue Vorschau erneut bestätigen. Es gibt keine erzwungene Übernahme der Loksteuerung.

## 7. Ergebnis prüfen und bei unklarer Antwort reagieren

<GuideFigure name="image-ecos-result" alt="Erfolgsmeldung nach geprüfter Bildübertragung und Lokzuordnung." caption="Simuliertes Erfolgsergebnis im echten RailKeeper-Dialog. Bild und Zuordnung werden getrennt bestätigt; die Aufnahme belegt keinen realen Gerätetest." />

| Anzeige | Nächster Schritt |
| --- | --- |
| Bild und Zuordnung geprüft | Kontrolliere das Bild anschließend auch auf der ECoS. |
| Weboberfläche nicht erreichbar | Adresse, Port 80 und Erreichbarkeit vom RailKeeper-Server prüfen. Ein erfolgreicher Protokolltest allein genügt nicht. |
| Vorschau abgelaufen oder geändert | Neu prüfen und die aktuelle Vorschau bewusst bestätigen. |
| Bild geprüft, Zuordnung nicht bestätigt | Bildplatz und Lok auf der ECoS prüfen. Falls angeboten, nur die Zuordnung erneut prüfen; ein bereits vorhandenes Bild wiederverwenden. |
| Antwort fehlt oder Zustand unklar | Nicht ungeprüft erneut hochladen. Zuerst Bildplatz und Lokzuordnung kontrollieren. |

Ein Fehler bei einer späten Zuordnung bedeutet nicht automatisch, dass ein vorheriger Upload
rückgängig gemacht wurde. Bewahre die heruntergeladene Datei auf, bis das Ergebnis geklärt ist.
