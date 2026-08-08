# Einstellungen: Daten-Navigation und dynamischer Seitenkopf

**Datum:** 2026-08-08

**Status:** Fachlich freigegeben, bereit für den Implementierungsplan

**Geltungsbereich:** Einstellungen, Stammdatenverwaltung und Artikelstammdaten

## Ausgangslage

Die vollständige Herstellerverwaltung wurde aus dem Einstellungsbereich „Daten“ entfernt und durch
eine generische, fachlich deutlich schwächere Herstellerliste unter „Artikelverwaltung“ ersetzt.
Die Herstellerdaten selbst blieben in `master_data_entries` erhalten. Die bisherige spezialisierte
Oberfläche für Herstellerseite, Suchdomains, Aliase, Quelle und Spurweiten ist weiterhin im Frontend
vorhanden, aber nicht erreichbar.

„Artikelverwaltung“ ist zudem kein eigenständiger Einstellungsbereich. Bestandseinheiten,
Artikelarten, Unterarten, kontrollierte Zusatzfelder und Lagerorte sind Stammdaten und gehören
inhaltlich unter „Daten“.

Der feste Seitenkopf „Einstellungen“ mit allgemeiner Beschreibung und Versionsnummer wiederholt
zusätzlich die Überschrift des jeweils gewählten Einstellungsbereichs im Inhaltsbereich.

## Zielbild

Die Einstellungen erhalten eine klare, datenorientierte Informationsarchitektur:

1. Der obere Seitenkopf zeigt Titel und Beschreibung des aktiven Hauptreiters.
2. Die Versionsnummer erscheint nicht mehr im Seitenkopf. Die vorhandene Versionsanzeige in der
   Seitenleiste beziehungsweise im Systembereich bleibt erhalten.
3. „Artikelverwaltung“ entfällt als eigener Hauptreiter.
4. „Daten“ enthält allgemeine Stammdaten und Artikelstammdaten in zwei überschaubaren Gruppen.
5. Die vollständige Herstellerdatenbank wird unter „Daten“ wieder zum einzigen fachlichen
   Pflegeort für Hersteller.
6. Es findet keine Datenmigration und keine Veränderung des Hersteller-Datenmodells statt.

## Dynamischer Seitenkopf

Der Seitenkopf wird aus dem aktiven Eintrag der Hauptnavigation abgeleitet. Er verwendet dieselbe
Typografie und dieselben Abstände wie der bisherige Kopf „Einstellungen“.

Beispiele:

| Aktiver Hauptreiter | Titel im Seitenkopf | Beschreibung |
| --- | --- | --- |
| Allgemein | Allgemein | Sprache, Startseite, Datumsformat und Druckausgabe. |
| Daten | Daten | Stammdaten für Fahrzeuge, Artikel und Anlagen zentral pflegen. |
| Digitalzentralen | Digitalzentralen | Verbindungen und sichere Lese- beziehungsweise Vorschauabläufe verwalten. |
| Import/Export | Import/Export | Daten, Sicherungen und Austauschformate verwalten. |
| Darstellung | Darstellung | Erscheinungsbild und Anzeigeoptionen festlegen. |
| Authentifizierung | Authentifizierung | Benutzer, Rollen, Sitzungen und Sicherheitsfunktionen verwalten. |

Der Inhaltsbereich wiederholt Titel und Beschreibung nicht. Abschnittsüberschriften innerhalb des
aktiven Bereichs bleiben erhalten, beispielsweise „Herstellerdatenbank“.

Die sprachabhängigen Titel und Beschreibungen liegen vollständig in den bestehenden deutschen und
englischen Sprachdateien.

## Hauptnavigation der Einstellungen

Die Hauptreiter lauten anschließend:

1. Allgemein
2. Daten
3. Digitalzentralen
4. Import/Export
5. Darstellung
6. Authentifizierung

Der bisherige Hauptreiter „Artikelverwaltung“ sowie sein direkter Query-Pfad werden aus der
sichtbaren Navigation entfernt. Bestehende Aufrufe mit `?tab=articleManagement` werden ohne
Fehler auf `?tab=data&group=article&type=stock_unit` gelenkt. Damit bleiben alte Lesezeichen und
während der Entwicklung weitergegebene Links nutzbar.

Die Datenansicht verwendet einen stabilen Query-Zustand:

- `group=general|article` bestimmt die Gruppe.
- `type` bestimmt die Datenart innerhalb der Gruppe.
- `?tab=data` ohne weitere Parameter öffnet `group=general&type=manufacturer`.
- Unbekannte oder nicht zur Gruppe gehörende Werte fallen auf den Standardreiter der Gruppe zurück.
- Beim Gruppen- oder Datenartwechsel wird die URL ersetzt, ohne einen vollständigen Seitenreload.

## Daten-Gruppen

Direkt unter der Einstellungs-Hauptnavigation liegt ein kompakter, app-eigener Gruppenumschalter:

- **Allgemeine Stammdaten**
- **Artikelstammdaten**

Darunter folgt eine einzelne Reiterzeile für die Datenarten der aktiven Gruppe. Gruppenumschalter
und Datenreiter verwenden die bestehende ruhige Unterstreichungsdarstellung, keine Boxen oder
Karten-Schaltflächen.

### Allgemeine Stammdaten

Die Gruppe enthält:

1. Hersteller
2. Kategorien
3. Gattungen
4. Epochen
5. Spurweiten
6. Bahngesellschaften
7. CV8-Hersteller
8. Symbole

„Hersteller“ ist der Standardreiter dieser Gruppe.

### Artikelstammdaten

Die Gruppe enthält:

1. Bestandseinheiten
2. Artikelarten und Unterarten
3. Kontrollierte Zusatzfelder
4. Lagerorte

„Bestandseinheiten“ ist der Standardreiter dieser Gruppe. Artikelarten und Unterarten bleiben in
einem gemeinsamen Reiter, weil Unterarten fachlich von der Artikelart abhängen.

## Herstellerdatenbank

Die spezialisierte Herstellerverwaltung aus dem bisherigen Datenbereich wird vollständig wieder
erreichbar. Sie bleibt die kanonische Herstellerpflege und umfasst mindestens:

- Herstellerbezeichnung und stabilen Schlüssel,
- zugeordnete Spurweiten,
- offizielle Herstellerseite,
- Suchdomains,
- Aliase,
- Quelle und Quellenhinweise,
- Aktivstatus,
- Suche, Sortierung, Anlegen und Bearbeiten,
- vorhandene Prüf- und Warnhinweise für unvollständige oder zu überprüfende Herstellerdaten.

Die reduzierte Herstellerliste aus `ArticleManagementSettings` wird entfernt. Es gibt keine zweite
Herstellerpflege und keinen parallelen Schreibpfad. Artikel-Dialoge und Artikelsuche greifen weiter
auf denselben Master-Data-Typ `manufacturer` zu.

## Komponenten- und Verantwortungsgrenzen

`SettingsView` ist bereits sehr groß. Die neue Informationsarchitektur darf dort keine weitere
vollständige Fachansicht ansammeln.

- `SettingsView` besitzt weiterhin den aktiven Hauptreiter und den dynamischen Seitenkopf.
- Eine fokussierte Datenansicht besitzt Gruppen- und Datenreiter sowie deren URL-/Auswahlzustand.
- Die bestehende spezialisierte Herstellerdarstellung wird als kanonische Herstellerkomponente
  erhalten beziehungsweise aus dem großen View extrahiert, nicht nachgebaut.
- Allgemeine generische Stammdaten verwenden weiterhin den vorhandenen Datenpfad.
- Bestandseinheiten, Artikelarten/Unterarten und Zusatzfelder verwenden die bestehende
  Artikelstammdatenlogik ohne eigenen Herstellerzweig.
- Lagerorte verwenden weiterhin `StorageLocationsSettings`.

Die genaue Dateiaufteilung wird im Implementierungsplan nach einem Größen- und Abhängigkeitscheck
festgelegt. Es wird keine zweite API, kein Adapter und kein zusätzliches Datenmodell eingeführt.

## Zustände und Bedienung

- Gruppen und Datenarten sind semantische Tabs mit `tablist`, `tab`, `tabpanel`, `aria-selected`
  und roving `tabIndex`.
- Pfeiltasten sowie Pos1 und Ende funktionieren innerhalb der jeweils aktiven Reiterzeile.
- Beim Gruppenwechsel wird der Standardreiter der neuen Gruppe aktiviert.
- Ein aktiver Reiter wird auf schmalen Bildschirmen automatisch sichtbar gescrollt.
- Reiterzeilen dürfen horizontal scrollen, das Dokument selbst erzeugt keinen horizontalen
  Überlauf.
- Lade-, Leer-, Fehler- und Nur-Lese-Zustände bleiben innerhalb des zugehörigen Datenpanels.
- Rollenregeln bleiben unverändert: Viewer lesen, Admin und Editor pflegen. Messe erhält keinen
  zusätzlichen Zugriff.

## Sprache

Alle sichtbaren Navigations-, Titel-, Beschreibungs-, Status- und Fehlermeldungen werden in Deutsch
und Englisch gepflegt. Technische Stammdatenschlüssel werden nicht übersetzt. Bereits bewusst
umbenannte oder benutzerdefinierte Labels bleiben unverändert.

## Kompatibilität und Datenintegrität

- Es gibt keine Datenbankmigration.
- Herstellerdatensätze werden weder kopiert noch neu erzeugt.
- API, Backup/Restore und Master-Data-Import/Export bleiben unverändert.
- Die Korrektur verändert ausschließlich Navigation, Komponentenkomposition und Präsentation.
- Der alte sichtbare Hersteller-Ersatzpfad wird entfernt, damit nur eine kanonische Pflegeoberfläche
  bestehen bleibt.

## Test- und Abnahmekriterien

Die Umsetzung ist abgeschlossen, wenn:

1. „Artikelverwaltung“ nicht mehr als Hauptreiter erscheint.
2. `?tab=articleManagement` kompatibel zur Datenansicht mit Artikelstammdaten weiterleitet.
3. jeder Hauptreiter seinen eigenen Titel und seine eigene Beschreibung im Seitenkopf zeigt.
4. die Versionsnummer nicht mehr neben dem Seitentitel steht.
5. Titel und Beschreibung nicht im Inhaltsbereich wiederholt werden.
6. „Daten“ genau die Gruppen „Allgemeine Stammdaten“ und „Artikelstammdaten“ anbietet.
7. die vollständige Herstellerverwaltung wieder erreichbar und funktional ist.
8. Herstellerseite, Suchdomains, Aliase, Quelle und Spurweiten weiterhin angezeigt und bearbeitet
   werden können.
9. die reduzierte Herstellerliste unter Artikelstammdaten nicht mehr existiert.
10. alle vier Artikelstammdatenbereiche unter „Daten“ erreichbar sind.
11. Rollen-, Lade-, Fehler- und Nur-Lese-Zustände erhalten bleiben.
12. Deutsch und Englisch vollständig sind.
13. Desktop, Dark/Light und 390-Pixel-Mobilansicht keinen Dokumentüberlauf erzeugen.
14. fokussierte Tests, vollständige Frontend-Suite und Produktionsbuild grün sind.

## Freigabe

Freigegeben wurde Variante A „Gruppenreiter“ mit der Ergänzung, dass der feste Kopf
„Einstellungen“ durch Titel und Beschreibung des aktiven Hauptreiters ersetzt wird.
