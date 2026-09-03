---
title: Artikelsuche, Web-Dokumente und Ersatzteile
description: Externe Artikeldaten prüfen, Web-Dokumente importieren und Fahrzeugersatzteile pflegen.
audience: user
status: stable
reviewedVersion: 0.1.20.4
lastReviewed: 2026-08-16
---

# Artikelsuche, Web-Dokumente und Ersatzteile

RailKeeper kann externe Vorschläge verwenden, um ein Fahrzeug zu vervollständigen,
Referenzdokumente zu importieren und Ersatzteile zu pflegen. Diese Ergebnisse sind Ausgangspunkte,
keine verbindlichen Modelldaten. Öffne die Quelle und prüfe Hersteller und Artikelidentität, bevor
du etwas speicherst.

Der Ablauf unterscheidet vier Zustände:

| Zustand | Bedeutung |
| --- | --- |
| Suchergebnis | Externer Vorschlag, nicht gespeichert |
| Übernommenes Feld oder ausgewähltes Bild | Lokaler Zustand im Fahrzeugeditor, noch nicht gespeichert |
| Importiertes Dokument oder Ersatzteilaktion | Sofortiger Schreibvorgang auf dem Server |
| Gespeichertes Fahrzeug | Grunddaten und vorgemerkte Bilder wurden durch den Fahrzeugspeicherablauf geschrieben |

## Voraussetzungen, Einstellungen und Zugriffsrechte

**Artikeldaten-Websuche** ist standardmäßig aktiviert, sofern sie nicht in den Einstellungen
deaktiviert wurde. Der Aktivierungszustand und die ausgewählten Quellen sind Einstellungen des
aktuellen Browsers. Ein anderer Browser oder ein anderes Gerät kann daher andere Einstellungen
verwenden.

Die Standardquellen sind Herstellerseiten, Modellbahn-Fokus-Kataloge, Händlerseiten und die
allgemeine Websuche. Modellbau Wiki ist eine optionale fünfte Quelle. Normale Artikel- und
Dokumentsuchen benötigen die folgenden Werte. RailKeeper migriert erkannte ältere
Standardkombinationen beim Lesen der Browsereinstellung auf die aktuellen Vorgaben.

- Hersteller;
- Spurweite;
- Artikelnummer oder Bezeichnung.

Eine reine EAN-Barcodesuche ist die Ausnahme und kann ohne diese Identitätswerte ausgeführt werden.

Admin und Editor können den bearbeitbaren Fahrzeugdialog nutzen, Dateien importieren und
Ersatzteile ändern. Viewer und Planner können gespeicherte Fahrzeuge, Beilagen und Ersatzteile
ansehen. Die Artikelsuche-API erlaubt Lesezugriff ab Viewer, der schreibgeschützte Fahrzeugdialog
deaktiviert Such- und Übernahmeaktionen dennoch. Schreibzugriffe auf Bilder, Beilagen und
Ersatzteile benötigen serverseitig mindestens Editor. Messe erhält keinen allgemeinen
Fahrzeugzugriff.

## Nach Artikeldaten suchen

Öffne einen Fahrzeugeditor und nutze **Artikeldaten suchen** unter **Modell**. Die Aktion bleibt
deaktiviert, bis Hersteller, Spurweite und entweder Artikelnummer oder Bezeichnung vorhanden sind.
RailKeeper sendet diese Angaben zusammen mit weiteren nicht leeren suchbaren Fahrzeugfeldern und
den in diesem Browser ausgewählten Quellen.

Der Server bereinigt die Eingabe und kann bekannte Hersteller um Aliase und bevorzugte Domains
ergänzen. Er erstellt mehrere gezielte Suchanfragen, verwendet ein Dienstzeitlimit von zehn
Sekunden, bewertet und sortiert die Ergebnisse, entfernt doppelte URLs und liefert höchstens zehn
Ergebnisse. Suchen für Piko und Roco können außerdem direkte herstellerspezifische
Ersatzteilergebnisse enthalten.

Der Ergebnisdialog zeigt die Belege hinter den Vorschlägen:

- endgültige Suchanfrage, aktive Quellklassen, bevorzugte Domains und geplante Suchanfragen;
- Ergebnistitel, Quelle, Bewertung, Textauszug und Anzahl gefundener Felder;
- ob die Detailseite geladen wurde, fehlgeschlagen ist oder übersprungen wurde;
- **Quelle öffnen** zur unabhängigen Prüfung;
- gefundene Bilder und ihre Vorschauen;
- aktuelle und gefundene Werte, ihren Status und Konflikte.

Die stabilen Ergebnisgruppen enthalten diese Felder:

| Gruppe | Sichtbare Felder |
| --- | --- |
| Modell | Bezeichnung, Artikelnummer, Hersteller, Spurweite, EAN, Bahngesellschaft, Epoche, Baureihe, Fahrzeugnummer, Untertyp, Kategorie |
| Maße / Aufbau | Länge, Gewicht, Farbe, Beschriftung, Ladung, Inneneinrichtung, Achsen, Achsanzahl, Haftreifen |
| Technik | Schnittstelle, Stromabnehmer, Digital-/Decoderdaten, Sound, Beleuchtung und technische Beschreibungen |
| Weitere Daten | Beschreibung, weitere Informationen, Produktionszeitraum, Listenpreis und Quell-URL |

RailKeeper entfernt offensichtlich leere, werbliche, Cookie-, Navigations- und unplausible
Feldwerte vor der Anzeige. Filterung und eine hohe Bewertung erleichtern die Prüfung, beweisen aber
nicht, dass ein verbleibender Wert zum physischen Modell gehört.

## Mit Barcode oder Kamera suchen

Nutze **Barcode suchen** unter **Modell**, um ein EAN-Feld mit dem aktuellen Wert aus dem
Fahrzeugformular zu öffnen. Tippe den Wert ein, füge ihn ein, nutze einen Tastaturscanner oder
starte die Kamera.

Das Absenden eines nicht leeren Codes kopiert ihn in das lokale EAN-Feld, lässt die Artikelnummer
unverändert, schließt den Barcodedialog und startet eine reine EAN-Suche. Die EAN wird erst beim
Speichern des Fahrzeugs dauerhaft geschrieben.

Die Kamerasuche benötigt:

- einen sicheren Browserkontext wie HTTPS oder localhost;
- Kameraunterstützung und Berechtigung im Browser;
- einen lesbaren Code mit mindestens acht Ziffern, nachdem Nichtziffern entfernt wurden.

RailKeeper fordert nach Möglichkeit die rückseitige Kamera an. Eine Erkennung füllt das Feld,
sendet die Suche aber nicht automatisch ab. Schlägt Berechtigung, Browserunterstützung oder
Erkennung fehl, gib die EAN manuell ein oder nutze einen Tastaturscanner.

## Ein Suchergebnis prüfen und übernehmen

Auswahlen gehören jeweils zu einem Ergebnis. Ein gefundenes Feld ist anfangs nur ausgewählt, wenn
das aktuelle Fahrzeugfeld leer ist. Gleiche Werte und Konflikte werden nicht automatisch
ausgewählt. Bilder sind nie vorausgewählt. Ein Bild mit fehlgeschlagener Vorschau verschwindet und
wird aus der Auswahl entfernt.

Öffne die Quelle und vergleiche die Artikelidentität, bevor du etwas auswählst. Prüfe Konflikte und
Plausibilität über alle Feldgruppen. **Ausgewählte Felder übernehmen** überträgt anschließend nur
markierte, gültige Felder dieses Ergebnisses. Wahrheitswerte werden in die stabile
RailKeeper-Darstellung umgewandelt. Ausgewählte Bilder werden als vorgemerkte externe Bilder
geführt. Ist
noch kein Bild vorgemerkt, wird das erste ausgewählte Bild zum Kandidaten für das Hauptbild.

Die Übernahme schließt den Ergebnisdialog, schreibt aber nichts auf den Server. Prüfe das gesamte
Fahrzeugformular und die vorgemerkten Bilder und nutze danach **Erstellen** oder **Änderungen
speichern**. Schließen oder Neuladen vor diesem Speichern verwirft übernommene Felder, EAN und
vorgemerkte Bilder.

Beim Speichern schreibt RailKeeper zuerst das Grundfahrzeug und lädt danach ausgewählte externe
Bilder nacheinander herunter. Schlägt ein späteres Bild fehl, bleiben Fahrzeug und frühere Bilder
gespeichert. Die normale abschließende Aktualisierung erfolgt dann nicht. Lade das Fahrzeug neu,
vergleiche die gespeicherten Bilder und wiederhole nur fehlende Importe. Sortierung, Metadaten,
Vorschauen, Wartungsverknüpfungen und Löschen nach dem Import erklärt
[Fahrzeugbilder und Beilagen](/de/guide/vehicles/media).

## Web-Dokumente finden und importieren

Öffne am gespeicherten Fahrzeug **Uploads** und nutze **Gefundene Dokumente**. Die Suche verwendet
die gespeicherte Fahrzeugidentität und die normale Voraussetzung aus Hersteller, Spurweite und
Artikelnummer oder Bezeichnung. Ergebnisse werden anhand URL oder Titel zusammengeführt und zeigen
Titel, Art und Quelle.

RailKeeper erkennt ein bereits importiertes Dokument, wenn seine URL in einer
Beilagenbeschreibung vorkommt. Eine solche Zeile kann nicht erneut ausgewählt werden und bietet
lokale Aktionen zum Herunterladen und Öffnen. Importiere ein einzelnes verbleibendes Ergebnis oder
wähle alle noch nicht gespeicherten Ergebnisse aus.

Ein Import lädt den externen Inhalt herunter und speichert ihn als echte lokale Fahrzeugbeilage.
Die Beschreibung enthält Quelle und Quell-URL. RailKeeper leitet eine dieser Kategorien ab:

- `Ersatzteilliste` bei Ersatzteilsignalen;
- `Anleitung` bei Handbuch- oder Anleitungssignalen;
- `Dokumentation` in allen anderen Fällen.

Der Server akzeptiert nur öffentliche HTTP(S)-URLs, prüft Weiterleitungsziele erneut und blockiert
private oder interne Ziele. Er erzwingt außerdem die konfigurierte Größen- und Typbegrenzung für
Beilagen, lehnt leere Dateien ab und verwendet für externe Anfragen ein Zeitlimit von zehn Sekunden.

Der Import mehrerer ausgewählter Dokumente sendet die Anfragen nacheinander. Ein späterer Fehler
lässt frühere Dokumente gespeichert und verhindert die normale abschließende Aktualisierung und
das Zurücksetzen der Auswahl. Lade das Fahrzeug neu, vergleiche die Beilagenliste und wiederhole
nur fehlende URLs.

## Ersatzteile aus einer Beilage extrahieren

Jede gespeicherte Beilage bietet **Ersatzteile extrahieren**. Die Aktion speichert zuerst die
aktuelle Beschreibung, Kategorie und Wartungsverknüpfung dieser ausgewählten Beilage. Danach wird
nur diese Beilage analysiert und die akzeptierten Vorschläge werden erstellt. Erst nach Erfolg der
gesamten Folge lädt RailKeeper das Fahrzeug neu und wechselt zu **Ersatzteile**.

Die Extraktion benötigt eine gespeicherte Fahrzeug-Artikelnummer. Der Server liest höchstens
12 MiB aus der Beilage und liefert höchstens 80 eindeutige Vorschläge. Wahrscheinliche
Ersatzteillisten, Anleitungen, Serviceblätter, HTML und unterstützte Textinhalte können analysiert
werden. PDF-Text wird zuerst direkt extrahiert. Gescannte PDFs können eine optionale, vom Betreiber
bereitgestellte OCR verwenden. Ohne funktionierende OCR kann ein Scan berechtigterweise keine
Vorschläge liefern.

Die stabile Aktion besitzt weder Bestätigung noch Zeilenvorschau. Sie bereinigt Beschreibungen,
verwirft leere und dokumenttitelähnliche Zeilen, entfernt bereits am Fahrzeug vorhandene Duplikate
und erstellt alle verbleibenden Kandidaten nacheinander. Eine Quell-URL zur internen
RailKeeper-Beilagenadresse wird nicht als externer Ersatzteillink gespeichert.

Beilagenmetadaten und früher erstellte Ersatzteile bleiben gespeichert, wenn eine spätere Anfrage
fehlschlägt. Das normale Neuladen und der Tabwechsel bleiben dann aus. Lade das Fahrzeug manuell
neu und vergleiche Beilage und Ersatzteile vor einem erneuten Versuch.

## Ersatzteile manuell pflegen

Der Tab **Ersatzteile** kann beschrieben werden, sobald das Fahrzeug existiert. Der Editor enthält
Artikelnummer, Beschreibung, Preis als freien Text und externen Link. Mindestens Artikelnummer,
Beschreibung oder Link ist erforderlich. Ein Preis allein ist ungültig.

Bei neuen und geänderten Werten werden führende und nachgestellte Leerzeichen entfernt. Anfangs ist
die Tabelle aufsteigend nach Artikelnummer sortiert. Wähle eine Spaltenüberschrift, um
Artikelnummer, Beschreibung, Preis oder Link auf- oder absteigend zu sortieren.

Das Erstellen eines Ersatzteils mit vorhandener Identität aktualisiert die passende Zeile, statt
eine weitere anzulegen. Die Identität verwendet zuerst die normalisierte Artikelnummer, sonst die
normalisierte URL, sonst die Beschreibung. Die Artikelnormalisierung ignoriert ein `ET`-Präfix,
Satzzeichen, Leerzeichen und Groß-/Kleinschreibung. Beim Erstellen bleiben vorhandene nicht leere
Werte erhalten. Nur noch leere Felder werden aus der neuen Eingabe ergänzt. Das Bearbeiten einer
ausgewählten Zeile schreibt genau ihre vier gesendeten Felder.

Speichern, Übernehmen und Löschen wirken sofort und laden das vollständige Fahrzeug nach Erfolg
neu. Löschen hat keine weitere Bestätigung. Da das vollständige Neuladen andere ungespeicherte
Änderungen im Fahrzeugeditor ersetzt, speichere oder verwirf sie vorher bewusst.

## Ein gespeichertes Ersatzteil nachschlagen

Ein gespeichertes Ersatzteil mit Artikelnummer bietet **Dieses Ersatzteil suchen**. Die Suche
verwendet Fahrzeughersteller, Ersatzteil-Artikelnummer, Hersteller- und Katalogquellen und bei
Bedarf einen Piko- oder Roco-Modus. Die stabile Oberfläche zeigt höchstens fünf Kandidaten und
bevorzugt nacheinander Ergebnisse mit Preis, Verfügbarkeit und Link.

Ein Ergebnis kann Preis, Verfügbarkeit, Quelle und externen Link zeigen. **Preis und Link
übernehmen** aktualisiert das gespeicherte Ersatzteil sofort. Artikelnummer und Beschreibung bleiben
unverändert. Die Verfügbarkeit wird angezeigt, aber nicht gespeichert. Fehlt dem ausgewählten
Kandidaten ein Preis, kann der stabile Controller Preis und URL eines anderen bepreisten
Kandidaten derselben Ergebnisliste übernehmen.

Beim Öffnen des Tabs prüft RailKeeper verlinkte Ersatzteile einmal für den aktuellen
Fahrzeugzustand. Piko und Roco verwenden eine gemeinsame Herstellerübersichtsabfrage. Bei anderen
Herstellern werden nur die ersten vier extern verlinkten Ersatzteile geprüft, die auch eine
Artikelnummer besitzen. Der Indikator ordnet erkennbaren Text als verfügbar, begrenzt, nicht
verfügbar oder unbekannt ein. Er bleibt ein aktueller externer Vorschlag, kein lokaler Bestand und
keine Lieferzusage.

## Eine Piko- oder Roco-Herstellerübersicht importieren

**Verfügbare Ersatzteile suchen** ist nur aktiv, wenn alle Bedingungen erfüllt sind:

- das Fahrzeug ist gespeichert;
- sein Herstellername enthält Piko oder Roco;
- die Fahrzeug-Artikelnummer ist nicht leer;
- der Benutzer hat als Admin oder Editor Zugriff auf den bearbeitbaren Dialog.

Die Schaltfläche kann auch bei deaktivierter Artikelsuche gewählt werden. RailKeeper bricht dann
vor der externen Anfrage ab, zeigt den Hinweis auf die Einstellungen und speichert nichts.

RailKeeper durchsucht die Herstellerübersicht und erstellt einen zurückhaltenden Importplan. Die
Zuordnung bevorzugt normalisierte Ersatzteil-Artikelnummer, dann Beschreibung, dann URL. Wiederholte
Vorschläge werden zusammengeführt. Vorhandene Treffer werden zuerst aktualisiert, fehlende
Ersatzteile danach erstellt.

Vorhandene Artikelnummer und Beschreibung bleiben erhalten. Vorhandener Preis und externe URL
haben ebenfalls Vorrang. Nur fehlender Preis oder fehlende URL werden ergänzt. Aktualisierungen und
Erstellungen laufen nacheinander ohne gemeinsame Transaktion oder Bestätigungsvorschau. Ein später
Fehler lässt frühere Zeilen gespeichert und verhindert die normale abschließende Aktualisierung.
Lade neu und vergleiche vor einem erneuten Versuch. **Keine fehlenden Ersatzteile in der
Herstellerübersicht gefunden** kann bedeuten, dass alle Herstellervorschläge bereits gespeicherten
Zeilen entsprechen, nicht dass der Hersteller keine Ersatzteile führt.

## Daten bei mehrstufigen Aktionen schützen

Zeitpunkt von Speicherung und Aktualisierung unterscheidet sich je Aktion:

| Aktion | Speichert Daten | Lädt das vollständige Fahrzeug neu |
| --- | --- | --- |
| Artikel- oder Barcodesuche ausführen | Nein | Nein |
| Ergebnisfelder oder Bilder übernehmen | Nein, nur lokaler Editor | Nein |
| Fahrzeug nach Ergebnisübernahme speichern | Grundfahrzeug, dann externe Bilder nacheinander | Nur nach Gesamterfolg |
| Gefundene Dokumente suchen | Nein | Nein |
| Ein Web-Dokument importieren | Sofort | Nach Erfolg |
| Ausgewählte Web-Dokumente importieren | Nacheinander | Nur nach Gesamterfolg |
| Ersatzteile extrahieren | Beilagenmetadaten, dann Ersatzteile nacheinander | Nur nach Gesamterfolg |
| Ein Ersatzteil hinzufügen, aktualisieren, aus Suche übernehmen oder löschen | Sofort | Nach Erfolg |
| Preis oder Verfügbarkeit prüfen | Nein | Nein |
| Piko-/Roco-Übersicht importieren | Aktualisierungen, dann Erstellungen nacheinander | Nur nach Gesamterfolg |

Keine später fehlgeschlagene Anfrage macht eine frühere erfolgreiche Anfrage desselben
nacheinander ausgeführten Ablaufs rückgängig. Lade nach einem teilweisen Fehler neu, vergleiche die
gespeicherten Daten und wiederhole nur fehlende Arbeiten.

Suchantworten, Ergebnisauswahlen und ungespeicherte Formularänderungen sind flüchtiger
Browserzustand und gehören nicht zu einer Sicherung. Gespeicherte, aus der Artikelsuche abgeleitete
Fahrzeugfelder, importierte Bilder, Beilagenmetadaten und -dateien sowie Fahrzeugersatzteile gehören
zu den lokalen RailKeeper-Anwendungsdaten und sind in der Anwendungssicherung enthalten.

Erstelle und validiere vor einem großen Dokumentstapel, einer Extraktion, einem Herstellerimport
oder einer Aufräumaktion eine Anwendungssicherung. Externe Quell-URLs und Ersatzteillinks können
außerhalb RailKeepers dennoch verschwinden und werden durch eine lokale Sicherung nicht dauerhaft.

## Leere, teilweise und fehlerhafte Zustände beheben

| Situation | Vorgehen |
| --- | --- |
| Artikelsuche ist deaktiviert | Aktiviere sie für diesen Browser in den Einstellungen oder arbeite ohne externe Vorschläge weiter. |
| Normale Suche ist deaktiviert | Trage Hersteller, Spurweite und Artikelnummer oder Bezeichnung ein. |
| Kamera startet nicht | Nutze HTTPS oder localhost, prüfe Berechtigung und Unterstützung oder gib den Code manuell ein. |
| Kein Artikelergebnis erscheint | Präzisiere Identität und ausgewählte Quellen, ohne die Quellenprüfung abzuschwächen. |
| Detailseite ist fehlgeschlagen oder wurde übersprungen | Öffne die Quelle und nutze nur überprüfbare Felder. |
| Ein gefundenes Bild verschwindet | Seine Vorschau ist fehlgeschlagen. Wähle ein anderes Quellbild. |
| Übernommene Werte verschwinden | Sie waren lokaler Entwurfszustand. Übernimm sie erneut und speichere das Fahrzeug. |
| Speichern externer Bilder schlägt teilweise fehl | Lade neu und vergleiche gespeicherte Bilder, bevor du nur fehlende wiederholst. |
| Keine Web-Dokumente erscheinen | Prüfe gespeicherte Fahrzeugidentität und Quellen. Nicht jede Quelle stellt Dokumente bereit. |
| Dokumentimport schlägt fehl | Prüfe öffentliche URL, Weiterleitungen, Typ, Größe und externe Verfügbarkeit. |
| Dokumentstapel schlägt teilweise fehl | Frühere Dokumente bleiben gespeichert. Lade neu und wiederhole nur fehlende URLs. |
| Extraktion findet nichts | Prüfe Fahrzeug-Artikelnummer, Inhalt, Textextraktion und optionale OCR. |
| Extraktion schlägt teilweise fehl | Metadaten und frühere Ersatzteile können gespeichert sein. Lade neu und vergleiche. |
| Ersatzteilformular wird abgelehnt | Trage Artikelnummer, Beschreibung oder Link ein. Ein Preis allein reicht nicht. |
| Einzelsuche ist deaktiviert | Speichere eine Ersatzteil-Artikelnummer und aktiviere die Artikelsuche. |
| Verfügbarkeit ist unbekannt | Behandle sie als ungeprüften externen Status und öffne die Quelle. |
| Herstellerübersicht ist deaktiviert | Sie benötigt Piko oder Roco und die gespeicherte Fahrzeug-Artikelnummer. |
| Herstellerimport schlägt teilweise fehl | Lade neu, vergleiche und wiederhole erst nach Prüfung der gespeicherten Ergebnisse. |
| Ersatzteil wurde ohne Nachfrage gelöscht | Löschen wirkt sofort. Wiederherstellung benötigt eine geeignete Sicherung. |

Einige stabile Backend- und Frontendzweige zeigen auch im englischen Sprachmodus deutsche Fehler.
Behandle dies als Einschränkung der stabilen Oberfläche und leite aus den Speicherregeln oben ab,
was bereits geschrieben worden sein kann.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Zubehörübersicht](/de/guide/accessories/) für Zubehörartikel und -bestand, getrennt von
  fahrzeugspezifischen Ersatzteildatensätzen
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
- [Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.20.3** und wurde zuletzt am
16.08.2026 geprüft.
