# Zubehörübersicht: Tabellen-, Kachel- und Mobile-Ansicht

## Ziel

Die Zubehörübersicht erhält auf Desktop dieselbe Wahl zwischen Tabellen- und Kachelansicht wie der
Fahrzeugbestand. Auf kleinen Bildschirmen zeigt RailKeeper unabhängig von der gespeicherten
Desktop-Auswahl automatisch eine kompakte Artikelliste.

Die Änderung betrifft ausschließlich die Darstellung bereits geladener Artikel. Filter, Sortierung,
Auswahl, Rollen, API-Verträge und Backend bleiben unverändert.

## Produktentscheidungen

- Die Tabellenansicht bleibt der Standard.
- Die Desktop-Auswahl wird lokal im Browser gespeichert.
- Die Kachelansicht verwendet die aktuelle Filterung und Sortierung ohne zusätzliche API-Anfrage.
- Unterhalb des bestehenden Mobile-Breakpoints wird immer die kompakte Mobile-Liste angezeigt.
- Ansehen, Bearbeiten, Archivieren und Wiederherstellen bleiben in allen Ansichten rollenabhängig.
- Eine geöffnete Artikelkarte oder Aktion verwendet dieselben Controller wie die Tabelle.
- Die Tabellen-Selektion bleibt erhalten, wenn zwischen Tabelle und Kacheln gewechselt wird.
  Kacheln und Mobile-Einträge erhalten keine zusätzliche Auswahl-Checkbox, solange keine
  ansichtsübergreifende Sammelaktion existiert.

## Architektur und Komponenten

### Ansichtsmodus

`AccessoriesView` besitzt den Ansichtsmodus `table | cards`. Der Initialwert wird über eine kleine
Zubehör-Hilfsfunktion aus `localStorage` gelesen. Der Schlüssel lautet
`railkeeper.accessories.view`. Unbekannte oder fehlende Werte fallen auf `table` zurück.

Ein Wechsel aktualisiert React-State und `localStorage` gemeinsam. Der Modus verändert weder
Filter noch Sortierung und löst keinen neuen Request aus.

### Toolbar

`ArticleToolbar` erhält den aktuellen Modus und einen Änderungs-Callback. Neben der Suche erscheint
ein kompakter Umschalter mit `Table2` und `Grid2X2`, entsprechend dem Fahrzeugbestand.

Die Schalter verwenden lokalisierte `aria-label`- und `title`-Texte, einen sichtbaren aktiven Zustand
und app-eigene transparente Icon-Buttons. Auf Mobilgeräten wird der Umschalter ausgeblendet, weil
dort die kompakte Liste verbindlich ist.

### Darstellungsgrenze

Eine fokussierte Zubehör-Darstellung koordiniert drei bestehende beziehungsweise neue Oberflächen:

1. `ArticleTable` bleibt Eigentümer der Desktop-Tabelle und ihrer Sortierköpfe.
2. `ArticleCardGrid` rendert das Desktop-Kachelraster.
3. `ArticleCompactList` rendert die automatische Mobile-Liste.

`AccessoriesView` wählt nur zwischen Tabelle und Kacheln für Desktop. CSS steuert, dass auf kleinen
Bildschirmen die Desktop-Darstellung ausgeblendet und die kompakte Liste eingeblendet wird. Dadurch
bleibt die Breakpoint-Logik konsistent mit dem Fahrzeugbestand und benötigt keine
`window.innerWidth`-Abfrage in React.

### Gemeinsame Aktionen

Die wiederkehrenden Aktionen werden in eine kleine `ArticleActions`-Komponente extrahiert. Sie
enthält:

- Ansehen
- Bearbeiten, nur für Admin und Editor
- Drei-Punkte-Menü mit Archivieren oder Wiederherstellen
- Fokusführung, Escape-Verhalten und Schließen bei Außenklick

Tabelle, Kacheln und Mobile-Liste verwenden dieselbe Komponente. Damit bleiben Rollen- und
Tastaturverhalten an einem kanonischen Ort und die zuletzt korrigierte Menü-Sichtbarkeit wird nicht
durch drei getrennte Implementierungen erneut gefährdet.

## Inhalte der Kacheln

Jede Kachel zeigt in ruhiger, dichter Form:

- primäres Artikelbild oder den vorhandenen Platzhalter
- Inventarnummer und Hersteller
- Artikelname
- Artikelnummer
- Art und Unterart
- Spurweite oder Spurweiten
- Gesamtbestand sowie frei, reserviert und eingebaut
- primären Lagerort und Hinweis auf weitere Lagerorte
- Artikelaktionen am unteren Rand

Der Name und das Bild öffnen die Leseansicht. Lange Werte werden gekürzt, bleiben aber über
`title` und zugängliche Beschriftungen vollständig erreichbar.

## Inhalte der Mobile-Liste

Jeder kompakte Eintrag zeigt:

- kleines Artikelbild oder Platzhalter
- Inventarnummer und Artikelname
- Hersteller und Artikelnummer
- Art oder Unterart
- Spurweite
- kompakten Bestandswert
- dieselben rollenabhängigen Artikelaktionen

Die Liste folgt dem bestehenden Fahrzeugmuster, verwendet aber Zubehörbegriffe und
Zubehördaten. Sie bleibt auch dann aktiv, wenn auf Desktop zuletzt die Tabellenansicht gewählt war.

## Datenfluss und Zustände

`useArticleOverview` bleibt alleiniger Eigentümer von Laden, Filtern, Sortieren, Archivieren und
Wiederherstellen. Alle drei Darstellungen erhalten dieselbe `overview.data.items`-Liste und dieselben
Callbacks.

Lade-, Fehler-, Leer- und Kein-Treffer-Zustände bleiben oberhalb der Darstellungswahl in
`AccessoriesView`. Es wird nie gleichzeitig ein Leerzustand und eine der drei Artikeldarstellungen
gerendert.

## Styling und Responsive-Verhalten

- Bestehende Design-Tokens und Inventar-Stile werden wiederverwendet.
- Zubehörspezifische Klassen verhindern Seiteneffekte auf Fahrzeugkarten.
- Kacheln verwenden ein responsives Desktop-Raster mit einer Mindestbreite um 260 Pixel.
- Desktop-Tabelle und Kachelraster werden am vorhandenen Mobile-Breakpoint ausgeblendet.
- Die kompakte Liste wird unterhalb dieses Breakpoints eingeblendet.
- Dark Mode, lange deutsche Texte, fehlende Bilder und geöffnete Aktionsmenüs bleiben prüfbar.
- Icon-Buttons bleiben transparent und verwenden farbige Hover- und Fokuszustände.

## Internationalisierung

Deutsch und Englisch erhalten Texte für:

- Ansicht wechseln
- Tabellenansicht
- Kachelansicht
- kompakte Artikelliste

Bereits vorhandene Artikel-, Bestands- und Aktionsbegriffe werden wiederverwendet.

## Tests und Abnahme

Automatisierte Tests decken ab:

- Standardmodus Tabelle
- Wechsel in die Kachelansicht
- Persistenz und Wiederherstellung über `localStorage`
- unveränderte Filterung und Sortierung beim Ansichtswechsel
- Kachelinhalte mit und ohne Bild
- Rollen für Ansehen, Bearbeiten und Archivieren
- gemeinsame Tastatur- und Fokuslogik des Drei-Punkte-Menüs
- CSS-Vertrag für Desktop-Darstellungen und automatische Mobile-Liste
- deutsche und englische Beschriftungen

Zusätzlich erfolgt eine Browserprüfung auf Desktop und in einem schmalen Viewport. Dabei werden
Tabelle, Kacheln, Mobile-Liste, Dark Mode, lange Texte und das geöffnete Drei-Punkte-Menü geprüft.

## Nicht-Ziele

- keine Backend- oder OpenAPI-Änderung
- keine neue Sammelaktion oder Kartenauswahl
- keine freie Wahl der Mobile-Darstellung
- keine generische gemeinsame Fahrzeug- und Zubehör-Inventarkomponente
- keine Änderung an Filtern, Sortierregeln oder Artikelberechtigungen
