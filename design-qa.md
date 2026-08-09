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
