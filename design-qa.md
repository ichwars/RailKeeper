# Design QA: Artikelübersicht im Bestandslayout

## Referenz und Prüfumgebung

- Referenz: `C:\Users\droth\AppData\Local\Temp\codex-clipboard-c38f7d9a-7452-4a08-904f-2a36101b01ff.png`
- Implementierung: `http://127.0.0.1:18083/accessories`
- Desktopvergleich: 1287 × 559 CSS-Pixel, Pixeldichte 1
- Mobileprüfung: 390 × 844 CSS-Pixel
- Zustände: Deutsch, Dark und Light Theme, ein vorhandener Tillig-Artikel
- Vergleichsbild: `.superpowers/sdd/article-overview-layout/comparison-reference-left-implementation-right.png`

## Vergleich

### Gesamthierarchie

Die Artikelübersicht folgt der Fahrzeugbestandsseite: kompakter Seitentitel mit primärer Aktion,
Bestandskennzahlen, transparenter Arbeitsbereich, Suche und Filter sowie die Datentabelle. Die
vorherige Eyebrow-Zeile und der umschließende Artikelkasten sind entfernt.

### Kennzahlen

Die vier vorhandenen Kennzahlen verwenden dieselben Kartenprimitiven, Höhen, Abstände, Icons und
Typografiestufen wie der Fahrzeugbestand. Die gebundene Menge spannt zwei Rasterspalten und erhält
damit denselben visuellen Rhythmus wie die breite Termin-Karte der Referenz. Es wurden keine
Kennzahlen ergänzt oder inhaltlich verändert.

### Werkzeugbereich

Der Listentitel steht links, die Suche rechts. Die bestehenden Filter folgen darunter in einer
kompakten Zeile; die Trefferzahl bildet den Abschluss. Auf 390 Pixeln stapeln sich die vorhandenen
Filter mit 38 Pixel Höhe, ohne ungewollte vertikale Flex-Ausdehnung.

### Tabelle

Die Tabelle verwendet Fläche, Kopfzeile, Zeilenrhythmus und Aktionsausrichtung des
Fahrzeugbestands. Spalten, Inhalte, Sortierung und Aktionen bleiben unverändert. Mobil bleibt die
760 Pixel breite Tabelle innerhalb ihres 352 Pixel breiten horizontal scrollbaren Containers.

### Themes und Responsive

Dark und Light Theme sind kontrastreich und verwenden ausschließlich bestehende Tokens. Bei
390 Pixeln beträgt die Dokumentbreite exakt 390 Pixel; es entsteht kein Dokumentüberlauf.

## Findings

- P0: keine
- P1: keine
- P2: keine
- P3: Die unveränderte Kennzahl `0 frei · 0 Lagerorte` bricht in der schmalen Desktopkarte um.
  Das ist bei der freigegebenen Inhaltsgrenze akzeptiert und bleibt auch mobil lesbar.

## Änderungshistorie während QA

1. Die gebundene Kennzahl erhielt die breite Bestandskartenvariante, damit Wert und Rhythmus der
   Referenz entsprechen.
2. Die schmalen Filter erhielten im mobilen Spaltenlayout `flex: 0 0 auto`, nachdem die visuelle
   Prüfung eine ungewollte Höhe von 160 Pixeln je Auswahlfeld zeigte.

## Ergebnis

passed

---

# Nachtrag: Bestandsformular Kompaktraster

## Referenz und Prüfumgebung

- Problemreferenz:
  `C:\Users\droth\AppData\Local\Temp\codex-clipboard-94dbed73-3e14-488c-bb24-8cf53ae27812.png`
- Gewählte Richtung: Variante B des visuellen Vergleichs, Korrektur schmal und Umbuchung breit.
- Implementierung: `http://127.0.0.1:18083/accessories`, Artikel Tillig 83101.
- Desktopaufnahme Dark:
  `C:\Users\droth\AppData\Local\Temp\railkeeper-stock-compact-grid.png`
- Desktopaufnahme Light:
  `C:\Users\droth\AppData\Local\Temp\railkeeper-stock-compact-light.png`
- Mobile Aufnahme 390 × 844:
  `C:\Users\droth\AppData\Local\Temp\railkeeper-stock-compact-mobile-bottom.png`
- Visuelle Entwurfsquelle Variante B:
  `.superpowers/brainstorm/visual-stock-balance-20260809/content/stock-layout-options-v1.html`

## Normalisierung und Zustand

- Problemreferenz: 1012 × 431 Pixel, fokussierter Ausschnitt des Bestandsblocks.
- Desktopimplementierung: 2019 × 910 Pixel bei 2243 × 1011 CSS-Pixeln und Browserdichte 0,9.
- Mobile Implementierung: 390 × 844 Pixel Bildausgabe. Die Browseremulation meldete aufgrund der
  geräteinternen Skalierung 433 × 937 CSS-Pixel; Dokument- und Clientbreite waren beide 433 Pixel.
- Zustand: angemeldeter QA-Admin, Deutsch, Tillig 83101, reiner Mengenbestand, Reiter `Bestand`.
- Themes: Dark und Light.

Die Problemreferenz ist kein Sollentwurf mit identischem Viewport, sondern der vom Nutzer markierte
Fehlerzustand. Die verbindliche Sollanordnung ist die anschließend ausgewählte Variante B. Deshalb
wird die Bestandsregion inhaltlich und proportional verglichen, nicht als Pixelkopie des fehlerhaften
Ausgangszustands.

## Vergleich

### Desktop

Die Bestandskorrektur belegt ein Drittel, die Umbuchung zwei Drittel des verfügbaren Bereichs. Von
und Nach Lagerort bilden die erste, Menge und Notizen die zweite Umbuchungszeile. Der zuvor markierte
Leerraum im Korrekturformular entfällt vollständig. Beide Aktionsschaltflächen bilden visuell eine
gemeinsame Abschlusslinie.

Die gemessenen Rasterbreiten betragen rund 347 und 695 CSS-Pixel. Das interne Umbuchungsraster
verwendet zwei Spalten mit jeweils rund 342 CSS-Pixeln.

### Mobile

In der 390 × 844 Bildausgabe stehen Bestandskorrektur und Umbuchung untereinander. Die vier
Umbuchungsfelder werden einspaltig dargestellt. Die Umbuchungsschaltfläche und der feste
Dialogfooter sind erreichbar. Dokument- und Clientbreite bleiben identisch; horizontaler
Dokumentüberlauf tritt nicht auf.

### Themes und Laufzeit

Dark und Light Theme zeigen dieselbe kompakte Hierarchie ohne Überlagerung oder abgeschnittene
Beschriftung. Die Browserkonsole enthält keine neuen Fehler oder Warnungen. Der ursprüngliche
Buchungs- und Umbuchungsablauf wurde nicht verändert.

## Vollansicht und Fokusvergleich

Die Desktopaufnahmen belegen Dialoghierarchie, feste Aktionsleiste und Einordnung zwischen
Bestandstabelle und Journal. Der fokussierte Bestandsblock ist in Referenz und Implementierung groß
genug dargestellt, um Überschriften, Feldreihen, Abstände und Schaltflächen direkt zu vergleichen;
ein zusätzlicher Ausschnitt war nicht erforderlich.

Geprüfte Interaktionen: Artikeldialog öffnen, Reiter `Bestand` aktivieren, Dark/Light umschalten,
mobile Bestandsregion scrollen und die Umbuchungsaktion erreichen. Die Browserkonsole blieb leer.

## Fidelity-Oberflächen

- Typografie: bestehende RailKeeper-Schriftfamilien, Größen, Gewichte und Zeilenhöhen unverändert.
- Abstand und Raster: der markierte Leerraum ist entfernt; äußeres 1:2-Raster und internes
  2×2-Raster entsprechen Variante B.
- Farben und Tokens: ausschließlich bestehende Panel-, Linien-, Text- und Akzenttokens in Dark und
  Light.
- Bild- und Assetqualität: im geänderten Formularbereich sind keine Rasterbilder oder neuen Assets
  vorhanden; das bestehende Umbuchungssymbol bleibt unverändert.
- Inhalt: Feldnamen, Pflichtmarkierungen und Aktionsbeschriftungen bleiben unverändert.

## Vergleichshistorie

1. Ausgangszustand, P2: Das gleich breite, auf die Umbuchungshöhe gestreckte Korrekturformular
   erzeugte große Leerflächen und einen unruhigen vertikalen Rhythmus.
2. Korrektur: Variante B wurde ausgewählt und als gewichtetes 1:2-Raster mit internem
   2×2-Umbuchungsraster umgesetzt.
3. Nachvergleich: Dark, Light und Mobile zeigen keine verbleibenden P0-, P1- oder P2-Abweichungen.

## Findings

- P0: keine
- P1: keine
- P2: keine
- P3: keine

## Ergebnis

passed

final result: passed

---

# Nachtrag: Artikeldialog Formularraster

## Referenz und Prüfumgebung

- Referenzen: vier vom Nutzer bereitgestellte Screenshots zu Bestand, Reservierung, Einbau und
  Boolean-Fachangaben.
- Implementierung: `http://127.0.0.1:18083/accessories`, Artikel Tillig 83101.
- Desktopprüfung: Dark Theme bei 1280 × 720 Bildausgabe sowie Light Theme in der normalen
  Browseransicht.
- Mobileprüfung: 390 × 844 Bildausgabe, einspaltige Allokationsformulare und kein horizontaler
  Dokumentüberlauf.
- Vergleichsbilder:
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-stock-cdp.png`
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-reservation-grid.png`
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-installation-grid.png`
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-subject-checkbox-alignment.png`
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-subject-light.png`
  - `C:\Users\droth\AppData\Local\Temp\railkeeper-subject-mobile-390x844.png`

## Vergleich

### Bestand

Bestandskorrektur und Umbuchung stehen als gleich breite Spalten nebeneinander. Die Schaltflächen
enden auf derselben Linie. Pflichtsterne bleiben Bestandteil des Feldlabels.

### Reservierung und Einbau

Beide Erfassungsformulare verwenden auf Desktop ein gleichmäßiges Zweispaltenraster. Fachlich breite
Zeilen wie Reservierung und Notizen belegen beide Spalten. Unterhalb von 920 Pixeln wechselt das
Raster auf eine Spalte.

### Boolean-Fachangaben

Die Kontrollzeilen für `Bettung` und `Digitaltauglich` besitzen Eingabefeldhöhe. Im Browservergleich
endet die Kontrollzeile von `Bettung` exakt auf derselben Unterkante wie das benachbarte Eingabefeld
`Profilhöhe`.

### Themes und Responsive

Dark und Light Theme verwenden weiterhin die bestehenden Tokens. Auf der mobilen Prüfansicht bleiben
Footer und Formularfelder erreichbar; ein Dokumentüberlauf tritt nicht auf.

## Findings

- P0: keine
- P1: keine
- P2: keine
- P3: keine

## Änderungshistorie während QA

1. Die erste Rasterfassung streckte das Eingabefeld `Mengenänderung` vertikal. Das Bestandsformular
   wurde deshalb als Flexspalte ausgeführt, während die beiden Formulare als gleich hohe
   Rasterspalten bestehen bleiben.
2. Der Theme-Wechsel wurde nach geschlossenem Dialog erneut geprüft. Dark und Light Theme zeigen
   keine Überlagerung oder abgeschnittene Beschriftung.

## Ergebnis

passed

---

# Design QA: Digitalzentralen

## Ergebnis

`passed`

Die Seite bildet das freigegebene Zielbild als funktionalen Inbetriebnahme-Workflow ab. Es gibt
keine offenen P0-, P1- oder P2-Befunde.

## Referenz und Prüfzustand

- Referenzbild: `C:\Users\droth\.codex\generated_images\01a003f1-6222-7100-a0b8-76b31be7fde7\exec-8fdd5aaf-7157-4c1b-9dcf-6999cad9d173.png`
- Implementierungsbild: `C:\Users\droth\Documents\GitHub\RailKeeper\design-qa-digital-implementation.png`
- Gesamtvergleich: `C:\Users\droth\Documents\GitHub\RailKeeper\design-qa-digital-comparison.png`
- Detailvergleich: `C:\Users\droth\Documents\GitHub\RailKeeper\design-qa-digital-focused-comparison.png`
- Mobile-Prüfbild: `C:\Users\droth\Documents\GitHub\RailKeeper\design-qa-digital-mobile.png`
- Referenz: 1680 x 945 Pixel
- Implementierung: 1680 x 944 Pixel nach Dichte-Normalisierung einer 1512 x 850 Aufnahme
- CSS-Viewport Desktop: 1680 x 944
- CSS-Viewport Mobil: 711 x 1000
- Zustand: ESU ECoS konfiguriert, noch nicht getestet, nicht aktiviert

## Geprüfte Oberflächen

- Typografie und Hierarchie entsprechen der bestehenden RailKeeper-Oberfläche und dem Zielbild.
- Abstände, Kartenraster und Workflow-Höhe wurden auf vollständige Sichtbarkeit im Desktop-Viewport
  abgestimmt.
- Farben verwenden bestehende Design-Tokens und den vorhandenen grünen Akzent.
- Alle vier Adapter sind auswählbar. Die Zustände „konfiguriert“, „getestet“ und „aktiv“ sind
  voneinander getrennt und werden nicht nur über Farbe vermittelt.
- Die Diagnoseansicht enthält Verbindungstest, letzte Diagnose, ECoS-Meldungen und technische
  Details.
- Konfiguration, Verbindungstest, Aktivierung, Deaktivierung, Live-Monitor und Entfernen der
  Verbindung sind bedienbar und an die vorhandenen APIs angebunden.
- Tabs, Schaltflächen und Statusangaben besitzen passende semantische Rollen und deaktivierte
  Zustände.
- Bei 711 Pixel CSS-Breite entsteht kein horizontaler Überlauf. Adapter, Workflow und Seitenkarten
  wechseln in ein einspaltiges Layout.

## Browserprüfung

- Konfigurationsdialog geöffnet, Host und Port geändert und gespeichert.
- Die Tabs „Letzte Diagnose“ und „ECoS-Meldungen“ geöffnet.
- Statusaktualisierung und Detailbereich bedient.
- Aktivierungsabhängigkeit, erfolgreicher Test und Entfernen der Verbindung zusätzlich durch
  Komponententests geprüft.
- Browserkonsole: keine Fehler oder Warnungen.

## Vergleichsverlauf

1. Erste Fassung: Der Adapterwähler lief über die gesamte Breite, die rechte Sicherheits- und
   Aktionsspalte begann zu tief. Außerdem drängten ein zusätzlicher Statushinweis und zu großzügige
   Innenabstände die Fußzeile aus dem sichtbaren Bereich. Behoben durch getrennte linke und rechte
   Spalten, stille initiale Statusabfrage und kompaktere Abstände.
2. Zweite Fassung: Bei 1680 x 944 CSS-Pixeln blieb ein geringer vertikaler Überlauf. Behoben durch
   spezifische Panel-Innenabstände und kompaktere Abstände im Digitalzentralen-Tab. Die gesamte
   Bedienoberfläche ist jetzt im Zielviewport sichtbar.

## Bewusste Abweichungen

- Die reale RailKeeper-App behält ihre einklappbare globale Seitenleiste. Das Zielbild zeigt nur den
  Seiteninhalt.
- Die im generierten Referenzbild angedeuteten Herstellerlogos wurden nicht als neue Markenassets
  übernommen. Die Implementierung verwendet die bestehende Icon-Sprache der App.
- Das Zielbild bezeichnet die ECoS-Karte trotz gesperrtem Aktivierungsschritt als „Aktiv“. Die
  Implementierung zeigt in diesem Zustand korrekt „Konfiguriert“.
- Zeitstempel, Benutzername und Zentralenversion stammen aus dem tatsächlichen Zustand statt aus
  den statischen Beispieldaten des Zielbilds.

## Niedrige Restabweichungen

- P3: Durch die globale Seitenleiste ist die rechte Spalte etwas schmaler als im freigestellten
  Zielbild. Inhalt, Reihenfolge und Bedienbarkeit bleiben vollständig erhalten.
- P3: Adapter werden durch einheitliche Linienicons statt durch Herstellerlogos repräsentiert.
