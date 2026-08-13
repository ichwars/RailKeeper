# Zubehör-Artikelsuche

**Datum:** 2026-08-13
**Branch:** `dev/accessory-article-search`
**Basis:** `origin/main`
**Bereich:** Zubehörübersicht und Zubehör-Artikel-Dialog

## Ziel

Die bisherige Artikelübersicht heißt in der deutschen Oberfläche künftig `Zubehör`. Im Dialog zum
Anlegen und Bearbeiten eines Zubehörartikels steht im Reiter `Artikel` derselbe Suchablauf wie im
Fahrzeugbestand zur Verfügung: eine Barcode- beziehungsweise EAN-Suche und eine Artikeldatensuche
über die konfigurierten Quellen.

Die Umsetzung bleibt vollständig von den lokalen Anlage-Änderungen getrennt. Der neue Branch basiert
direkt auf `origin/main` und enthält keine Commits des lokalen Anlage-Branches.

## Benennung und Navigation

- Der deutsche Navigationseintrag `Artikelübersicht` wird zu `Zubehör`.
- Der Seitentitel und die zugehörigen Zugriffs- und Bereichsbezeichnungen verwenden ebenfalls
  `Zubehör`, sofern sie den gesamten Produktbereich benennen.
- Fachbegriffe innerhalb des Bereichs bleiben präzise: `Artikel`, `Neuer Artikel`, `Artikelart`,
  `Artikelnummer` und `Artikel bearbeiten` werden nicht pauschal umbenannt.
- Der bestehende Pfad `/accessories`, interne Feature-Namen und API-Routen bleiben stabil.
- Die englische Oberfläche verwendet `Accessories` für Navigation und Bereichstitel. Fachbegriffe
  wie `article` bleiben in Tabellen, Formularen und Aktionen erhalten.

## Bedienablauf

Der Suchblock erscheint im Zubehör-Artikel-Dialog am Anfang des Reiters `Artikel`, vor Produktbild
und Stammdatenraster. Aufbau, Dichte, Symbole und Zustände entsprechen dem vorhandenen Suchblock im
Fahrzeugdialog.

Der Block bietet zwei Aktionen:

1. `Barcode suchen` öffnet den app-eigenen Barcode-Dialog. Der Code kann manuell, mit einem
   Tastatur-Scanner oder über die Browserkamera erfasst werden. Nach Bestätigung wird er als EAN in
   den aktuellen Formularentwurf übernommen und die Artikelsuche mit diesem Wert gestartet.
2. `Artikeldaten suchen` startet eine Suche aus den bereits eingetragenen Artikelwerten. Die Aktion
   ist deaktiviert, solange keine geeigneten Suchkriterien vorhanden sind.

Die Suche ist im Ansichtsmodus und für Rollen ohne Bearbeitungsrecht deaktiviert. Laufende Suchen
deaktivieren beide Aktionen, damit keine konkurrierenden Ergebnisse entstehen.

## Gemeinsame Suchbasis

Die vorhandenen Fahrzeugkomponenten werden nicht in das Zubehörfeature kopiert. Stattdessen werden
Barcode-Dialog, Ergebnisdarstellung und allgemeine Auswahlmechanik so extrahiert, dass beide
Fachbereiche sie mit einem Adapter konfigurieren können.

Der jeweilige Adapter definiert:

- verfügbare Suchkriterien,
- sichtbare und übernehmbare Ergebnisfelder,
- Bezeichnungen und Feldgruppen,
- Lesen des aktuellen Formularwerts,
- Umwandlung eines Suchwerts in das jeweilige Formularformat,
- Import ausgewählter Bilder nach erfolgreichem Speichern.

Die bestehende Fahrzeugfunktion bleibt fachlich und visuell unverändert. Vorhandene Fahrzeugtests
sichern diese Rückwärtskompatibilität.

## Zubehör-Suchkriterien und Feldzuordnung

Die Zubehörsuche verwendet die vorhandene Route `/api/v1/article-search` und die bereits
konfigurierten Suchquellen. Als Kriterien werden übergeben:

- Hersteller,
- Artikelnummer,
- Bezeichnung,
- erste ausgewählte Spurweite,
- EAN sowie
- weitere bereits befüllte, kompatible Zubehörfelder.

Folgende Trefferfelder dürfen in den Zubehörentwurf übernommen werden:

| Suchfeld | Zubehörfeld | Verhalten |
| --- | --- | --- |
| `manufacturer` | Hersteller | Nur vorhandene Stammdatenwerte übernehmen |
| `articleNumber` | Artikelnummer | Textwert übernehmen |
| `name` | Bezeichnung | Textwert übernehmen |
| `ean` | EAN | Bereinigten Textwert übernehmen |
| `gauge` | Spurweiten | Als einzelne Auswahl ergänzen, sofern in den Stammdaten vorhanden |
| `scale` | Maßstab | Textwert übernehmen |
| `description` | Beschreibung | Ungeeignete Webseitentexte wie Cookie-Hinweise herausfiltern |
| `articleSourceUrl` | Produkt-URL | Nur sichere HTTP- oder HTTPS-Quellen anzeigen und übernehmen |

Artikelart und Unterart werden nicht automatisch aus allgemeinen Suchbegriffen abgeleitet. Diese
Werte steuern Pflichtfelder und Fachangaben und bleiben deshalb eine bewusste Auswahl des Nutzers.
Unbekannte Hersteller und Spurweiten werden nicht still in Dropdownfelder geschrieben. Sie bleiben
im Treffer sichtbar, sind aber erst nach Pflege der Stammdaten übernehmbar.

## Ergebnisprüfung und Konflikte

Die Suchergebnisse verwenden denselben app-eigenen Dialog wie im Fahrzeugbestand. Jeder Treffer
zeigt Quelle, Trefferwert, Detailstatus, verfügbare Felder, bestehende Werte, gefundene Werte und
Konfliktstatus.

- Leere Formularfelder sind zunächst zur Übernahme ausgewählt.
- Identische Werte werden als übereinstimmend gekennzeichnet.
- Abweichende bestehende Werte werden nicht automatisch ausgewählt.
- Nur ausdrücklich ausgewählte Felder werden in den Entwurf übernommen.
- Die Übernahme speichert den Artikel nicht automatisch. Der Nutzer kann das Ergebnis im Formular
  prüfen und verwendet anschließend die normale Speichern-Aktion.

Fehlerhafte oder fachlich unplausible Werte werden vor der Anzeige entfernt. Die bestehende
Bereinigung der Fahrzeugdaten wird in gemeinsame Regeln und domänenspezifische Regeln getrennt.

## Bilder

Trefferbilder werden wie beim Fahrzeugbestand angezeigt und können einzeln ausgewählt werden. Die
Auswahl bleibt zunächst Bestandteil des ungespeicherten Artikelentwurfs.

Nach erfolgreichem Anlegen oder Aktualisieren importiert das Backend die ausgewählten Bilder als
Zubehördokumente der Kategorie `image`. Das erste ausgewählte Bild wird nur dann Primärbild, wenn
noch kein Primärbild vorhanden ist. Der Import verwendet die vorhandenen Sicherheitsregeln für
externe Downloads, insbesondere URL-Prüfung, Redirect-Prüfung, Größenbegrenzung, MIME-Erkennung,
Dateitypbegrenzung und Schutz vor privaten beziehungsweise internen Zieladressen.

Schlägt ein Bildimport fehl, bleibt der gespeicherte Artikel erhalten. Der Dialog zeigt den
fehlgeschlagenen Bildimport verständlich an und ermöglicht einen erneuten Versuch, ohne die bereits
gespeicherten Stammdaten zurückzunehmen.

## Fehler- und Leerzustände

- Ist die Websuche in den Einstellungen deaktiviert, öffnet der Ergebnisdialog mit einem klaren
  Hinweis und ohne Anfrage an externe Quellen.
- Fehlen Suchkriterien, bleibt die Artikeldaten-Aktion deaktiviert und erklärt den Grund über ihren
  Hilfetext.
- Kameraunterstützung, Berechtigung und sicherer Kontext verwenden dieselben Meldungen und
  Fallbacks wie der Fahrzeugbestand.
- Keine Treffer, Netzwerkfehler und ungültige Treffer erscheinen innerhalb des app-eigenen
  Ergebnisdialogs.
- Ein geschlossener Suchdialog verwirft nur das aktuelle Suchergebnis, nicht den Artikelentwurf.

## API und Sicherheit

Die bestehende Artikelsuche benötigt keine neue Suchroute. Für den dauerhaften Bildimport wird eine
schmale, editorgeschützte Zubehörroute ergänzt und im OpenAPI-Vertrag dokumentiert. Schreibzugriffe
bleiben CSRF-geschützt und serverseitig auf Admin- und Editorrollen begrenzt.

Externe Treffer bleiben unverbindliche Vorschläge. RailKeeper übernimmt keine Daten ohne sichtbare
Nutzerauswahl. Externe URLs und Inhalte gelten weiterhin als nicht vertrauenswürdig.

## Tests und Abnahme

Die Umsetzung erfolgt testgetrieben. Abgedeckt werden mindestens:

- Navigation und Seitentitel `Zubehör` auf Deutsch sowie `Accessories` auf Englisch,
- unveränderte Route `/accessories`,
- Suchblock und beide Aktionen im Zubehör-Artikel-Dialog,
- Rollen- und Ansichtsmodus,
- Suchkriterien aus dem aktuellen Entwurf,
- Barcodeübernahme und anschließende EAN-Suche,
- Vorauswahl leerer Felder und Schutz bestehender Konfliktwerte,
- Stammdatenprüfung für Hersteller und Spurweite,
- selektive Feldübernahme ohne automatisches Speichern,
- Auswahl, sicherer Import und Primärbildregel für Trefferbilder,
- Lade-, Fehler- und Leerzustände,
- unverändertes Verhalten der bestehenden Fahrzeugsuche,
- OpenAPI-Vertrag und serverseitige Rollenprüfung.

Vor der Übergabe laufen `go test ./...`, der vollständige Frontend-Testlauf und
`npm.cmd run build`. Zusätzlich werden Zubehöranlage, Barcode-Dialog, Ergebnisübernahme und
Bildimport lokal im Browser geprüft.

## Nicht Bestandteil

- keine Änderung der lokalen Anlage-Funktionen,
- keine automatische Klassifikation von Artikelart oder Unterart,
- keine neue externe Suchquelle,
- keine Änderung der Suchanbieter-Konfiguration,
- keine automatische Speicherung nach der Trefferübernahme,
- keine öffentliche Freigabe oder Cloud-Synchronisierung.
