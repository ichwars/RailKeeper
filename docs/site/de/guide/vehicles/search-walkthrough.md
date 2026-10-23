---
title: Suchvorschläge und Ersatzteile sicher übernehmen
description: Suchvorschläge und Ersatzteile sicher übernehmen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Suchvorschläge und Ersatzteile sicher übernehmen

**Ziel:** Gefundene Angaben bewusst auswählen und eine lokale Ersatzteilliste ausprobieren.
Zum Übernehmen brauchst du **Admin** oder **Editor** und ein gespeichertes Testfahrzeug.
Die folgenden Suchtreffer sind im Browser simuliert; die TXT-Extraktion nutzt den echten lokalen Import.

## 1. Suche mit passenden Ausgangsangaben öffnen

Speichere zunächst andere Fahrzeugänderungen. Für die Artikelsuche brauchst du geeignete Angaben,
beispielsweise Hersteller, Spurweite und Artikelnummer. Prüfe, dass die Artikeldaten-Websuche unter
**Einstellungen > Allgemein** aktiviert ist, und wähle die gewünschten Quellklassen.

Öffne das Fahrzeug und wähle **Artikeldaten suchen**. Bei echten Treffern prüfe die Originalquelle,
exakte Artikelnummer, Spurweite und Variante. Rangfolge und Vertrauensangabe ersetzen diese Prüfung nicht.

<GuideFigure name="article-result-review" alt="Suchdialog mit vorhandenem Namen und abweichendem synthetischem Vorschlag." caption="Echte Suchoberfläche, aber kontrolliert simulierte Treffer von example.invalid. Das Bild belegt Feldvergleich und Auswahl, keine Qualität eines Suchanbieters." />

Ein abweichender bestehender Name ist ein Konflikt und zunächst nicht ausgewählt.
Wähle nur Felder, die du ersetzen oder ergänzen willst. **Ausgewählte Felder übernehmen** ändert den
Fahrzeugentwurf. Erst **Änderungen speichern** sichert diese Fahrzeugangaben.
Mit **Abbrechen** kannst du den ungespeicherten Entwurf verwerfen.

## 2. Barcode als Suchhilfe verwenden

Wähle **Barcode suchen**. Du kannst den Code eintippen, mit einem Tastaturscanner einlesen oder
mit einer Scanner-App übertragen. Die Kamera benötigt Browserunterstützung und eine sichere
Umgebung, normalerweise HTTPS oder localhost, sowie deine Freigabe.

<GuideFigure name="article-barcode" alt="Barcodedialog mit einer künstlichen EAN-Eingabe vor dem Suchen." caption="Die manuelle Eingabe ist in der Demo geprüft. Eine Kameraaufnahme oder physische Scannerabnahme wird nicht behauptet." />

Prüfe den Code vor dem Suchen. Er dient als **EAN**; die Artikelnummer bleibt getrennt.
Vergleiche auch bei einem EAN-Treffer die genaue Ausführung des Modells.

## 3. Keine Treffer oder ungeeignete Quellen

<GuideFigure name="article-no-results" alt="Suchdialog ohne Treffer für die kontrollierte Beispielantwort." caption="Simulierter Leerzustand in der echten Oberfläche. Ein leerer Trefferbestand beweist nicht, dass das Produkt nicht existiert." />

Kontrolliere Schreibweise, Artikelnummer und Quellenauswahl. Entferne unpassende Zusatzbegriffe.
Bei einem Fehler notiere die konkrete Meldung. Überschreibe keine verlässlichen Bestandsangaben
mit einem nur ähnlich klingenden Produkt.

## 4. Ein Web-Dokument dauerhaft aufbewahren

Nutze in den Uploads die **Web-Dokumente** des Fahrzeugs. Prüfe Trefferquelle, Dokumenttitel,
Artikelbezug und Kategorie, bevor du importierst. Ein gespeicherter Link allein ist noch keine lokal
verfügbare Beilage. Nach einem Import öffne die gespeicherte Datei und kontrolliere den Inhalt.
Die [Detailanleitung](search-and-spares#web-dokumente-finden-und-importieren) beschreibt Auswahl,
Import, Fehler und den getrennten Speicherzeitpunkt. Diese Bildfolge enthält keinen echten Web-Download.

## 5. Ersatzteile lokal ausprobieren

<GuideDownload file="spare-parts-demo.txt">Künstliche Ersatzteilliste zum lokalen Test herunterladen</GuideDownload>

1. Lege ein Testfahrzeug mit Artikelnummer **99001** an. Dies ist eine Übungsnummer, keine Produktempfehlung.
2. Öffne **Uploads** und lade die TXT als Beilage hoch. Wähle **Ersatzteilliste** als Kategorie.
3. Prüfe Dateiname und Inhalt. Die Liste gehört ausdrücklich zum Modellartikel 99001.

<GuideFigure name="spare-document" alt="TXT-Ersatzteilliste als gespeicherte Beilage am Übungsfahrzeug." caption="Tatsächlich hochgeladene, künstliche TXT. Dafür sind Internet, Herstellerdownload und OCR nicht erforderlich." />

4. Speichere oder verwirf vorher andere Fahrzeugentwürfe.
5. Wähle an dieser Datei **Ersatzteile extrahieren**.

::: warning Diese Aktion speichert direkt
Die Aktion speichert Beilagenangaben, legt erkannte neue Ersatzteile an und lädt das Fahrzeug neu.
Sie wartet nicht auf **Änderungen speichern**. Bei einem Teilfehler prüfe die bereits gespeicherten
Teile vor einer Wiederholung.
:::

<GuideFigure name="spare-extraction-result" alt="Zwei tatsächlich aus der TXT erzeugte Ersatzteile im Fahrzeugdialog." caption="Die echte lokale Extraktion hat 99001-01 und 99001-02 angelegt. Es wurden keine Herstellerdaten oder Preise erfunden." />

**Erwartung:** `99001-01` für die Kupplung und `99001-02` für den Radsatz, ohne Preis.
Kontrolliere und korrigiere die Bezeichnungen bei Bedarf. Die Übung sagt nichts über die
OCR-Qualität gescannter PDFs aus. Ohne passende Artikelnummer oder bei nicht lesbarem Text kann
kein verwendbarer Vorschlag entstehen. Verwende dann die manuelle Ersatzteilanlage.

[Ausführliche Suche, Dokumente und Ersatzteile](search-and-spares).
