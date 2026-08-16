---
title: Einträge und Drucken
description: Messeloks erfassen, Adresskonflikte lösen und die Betriebsliste drucken.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Einträge und Drucken

Ein Messelisteneintrag beschreibt eine Lokomotive so, wie sie während der ausgewählten
Veranstaltung benötigt wird. Der Eintrag verbindet Besitzer und Modellidentität mit Betriebstagen,
digitalen oder analogen Steuerungsdaten, einem Bild, Funktionstasten und Notizen.

## Eintragsrechte und Speicherung

Messe und Admin können Einträge anlegen und bearbeiten, solange die ausgewählte Liste offen ist.
Nur Admin sieht am Eintrag die Aktion **Löschen**. Eine gesperrte Liste weist alle drei Änderungen
zurück, auch das Löschen durch Admin.

Der Eintragsdialog ist ein gemeinsamer Entwurf mit den drei Reitern **Allgemein**,
**Bilder upload** und **Funktionstasten**. Änderungen in jedem Reiter bleiben Formularzustand im
Browser, bis **Speichern** erfolgreich war. Das Speichern übermittelt den vollständigen Eintrag,
schließt den Dialog und lädt die Einträge der ausgewählten Liste neu. **Abbrechen** oder die
Schließen-Schaltfläche verwirft die aktuellen Dialogänderungen ohne eigene Warnung.

Gespeicherte Einträge, eingebettete Bilddaten und Funktionseinstellungen sind lokale
RailKeeper-Anwendungsdaten und gehören zum App-Backup. Ungespeicherte Formularänderungen und
Druckoptionen gehören nicht dazu.

## Eintrag anlegen oder bearbeiten

Wähle eine offene Liste und verwende das Bedienelement zum Anlegen im Eintragsbereich.
**Eintrag erfassen** öffnet sich im Reiter **Allgemein**. Fülle die beiden Pflichtfelder aus:

| Feld | Regel |
| --- | --- |
| Besitzer | Pflichtfeld. Der Server entfernt führende und nachgestellte Leerzeichen. |
| Lok Bezeichnung | Pflichtfeld. Der Server entfernt führende und nachgestellte Leerzeichen. |

Mit **Bearbeiten** in einer vorhandenen Zeile öffnet sich **Eintrag bearbeiten** mit den aktuellen
Werten. v0.1.17.6 bietet in diesem Dialog keine Fahrzeugauswahl und kein sichtbares Feld für eine
Fahrzeugverknüpfung. Der Eintrag bleibt ein Messedatensatz, auch wenn seine Beschreibung einem
Fahrzeug im allgemeinen Bestand ähnelt.

Wähle **Speichern** erst, nachdem du alle drei Reiter geprüft hast. Während des Speicherns oder bei
einem erkannten DCC- oder SX-Adresskonflikt ist die Schaltfläche deaktiviert. Wurde die Liste nach
dem Öffnen des Dialogs gesperrt, weist der Server den Vorgang ebenfalls zurück.

## Allgemeine und Steuerungsdaten ausfüllen

Der Reiter **Allgemein** enthält drei Bereiche.

### Basisdaten

| Feld | Eingabe und Bedeutung |
| --- | --- |
| Besitzer | Erforderlicher Freitext für die verantwortliche Person oder den Verein. |
| Lok Bezeichnung | Erforderlicher Freitext als hauptsächliche Modellidentität. |
| Hersteller | Auswahl aus aktiven Hersteller-Stammdaten, wenn das Konto sie lesen darf. |
| Baureihe | Optionaler Freitext. |
| Gattung | Auswahl aus aktiven Fahrzeug-Gattungen, wenn verfügbar. |
| Epoche | Auswahl aus aktiven Epochen, wenn verfügbar. |
| Bahnverwaltung | Auswahl aus aktiven Bahnverwaltungen, wenn verfügbar. |

Admin kann diese allgemeinen Stammdatenauswahlen laden. Ein Messekonto mit zusätzlicher Rolle
Viewer, Editor oder Planner kann sie ebenfalls laden. Reines Messe erhält in diesen vier
Auswahllisten nur **Keine Auswahl**. Die Auswahllisten nehmen keinen beliebigen Freitext an. Kann
ein gespeicherter Wert nicht geladen werden, zeigt das Bedienelement möglicherweise
**Keine Auswahl**, obwohl der Eintragsentwurf seinen bisherigen Wert noch enthält. Wähle nicht
gezielt **Keine Auswahl** und speichere, wenn diese Klassifikation erhalten bleiben muss. Prüfe sie
zuerst mit Admin oder einem Messekonto mit bestandsberechtigter Zusatzrolle.

### Steuerung

| Feld | Eingabe und Bedeutung |
| --- | --- |
| Decoder-Typ | Optionaler Freitext. |
| Adapter / Schnittstelle | Keine Auswahl, NEM 651, NEM 652, PluX16, PluX22, MTC21, Next18, 8-polig oder 21-polig. |
| Adresse DCC | Optionaler Freitext, der mit anderen geladenen Einträgen dieser Liste verglichen wird. |
| Adresse SX | Optionaler Freitext, der getrennt mit anderen geladenen Einträgen verglichen wird. |
| Analog | Schalter für die Verfügbarkeit des Analogbetriebs. |

Tabelle und Report verbinden DCC und SX unter **Adresse**, zeigen Analog als Ja oder Nein und führen
Decoder-Typ und Schnittstelle getrennt auf. Leere Werte erscheinen als Strich.

**Baureihe** wird mit dem Eintrag gespeichert, erscheint in v0.1.17.6 jedoch weder in Tabelle und
Listenansicht noch im gedruckten Report. Öffne **Eintrag bearbeiten**, um den Wert zu prüfen.

### Nr. / Beschriftung / Merkmale

Das Textfeld **Notizen** speichert freie Betriebsinformationen. Der Server entfernt beim Speichern
an allen Textfeldern führende und nachgestellte Leerzeichen, erhält aber den Inhalt dazwischen.

## Messetage auswählen

Die Tagesauswahl bietet **Alle Tage** und **Tag 1** bis **Tag 4**.

- **Alle Tage** ersetzt eine vorhandene Einzelauswahl.
- Die Auswahl eines nummerierten Tages bei aktivem **Alle Tage** begrenzt den Eintrag auf diesen
  Tag.
- Weitere nummerierte Tage ergeben eine Mehrfachauswahl.
- Kein nummerierter Tag oder alle vier nummerierten Tage werden beim Speichern zu **Alle Tage**.
- Gespeicherte Einzeltage werden in die Reihenfolge Tag 1, Tag 2, Tag 3, Tag 4 gebracht.

Der ausgewählte Umfang erscheint unter dem Besitzer in Tabellen und Reports. Der Druck filtert
nicht nach einem Tag. Er enthält alle Reporteinträge und zeigt deren jeweiligen Tagesumfang.

## DCC- und SX-Adresskonflikte lösen

RailKeeper vergleicht jede nicht leere DCC- und SX-Angabe mit dem entsprechenden Wert der anderen
aktuell geladenen Einträge der ausgewählten Liste. Führende und nachgestellte Leerzeichen sowie
Groß- und Kleinschreibung werden beim Vergleich ignoriert. Der gerade bearbeitete Eintrag bleibt
außen vor.

Ein Konflikt erzeugt eine Feldwarnung wie **DCC-Adresse bereits bei BR 218 vergeben.** oder
**SX-Adresse bereits bei BR 218 vergeben.** Die Hauptaktion **Speichern** ist deaktiviert; der
Speicherpfad meldet **Diese Adresse ist in der ausgewählten Liste bereits vergeben.** Ändere die
Adresse oder korrigiere den anderen Eintrag.

Dies ist eine Browserprüfung über die bereits geladenen Einträge. Der stabile Server erzwingt keine
eindeutigen Adressen. Arbeiten mehrere Personen gleichzeitig, lade die Liste unmittelbar vor einer
neuen Adressvergabe neu und prüfe anschließend die gedruckte Liste auf Konflikte.

## Bild hinzufügen oder entfernen

Der Reiter **Bilder upload** speichert eine Bildquelle pro Eintrag.

| Aktion | Stabiles Verhalten |
| --- | --- |
| Bildlink eingeben | Der Browser zeigt die angegebene Quelle als Vorschau und speichert sie später mit dem Eintrag. |
| **Bild hochladen** | Akzeptiert PNG, JPEG oder WebP und liest die Datei als eingebettete Daten in den Entwurf. |
| Quelle ersetzen | Ein anderer Link oder eine andere Datei ersetzt die Quelle im Entwurf. |
| **Bild entfernen** | Leert die Quelle im Entwurf. Dauerhaft wird dies erst nach erfolgreichem Speichern des Eintrags. |

RailKeeper lädt oder validiert einen externen Link nicht vor dem Speichern. Der Browser ruft die
externe Ressource ab, wenn er Vorschau, Tabelle, Detailansicht oder Report darstellt. Prüfe Quelle,
Verfügbarkeit und Datenschutz vor der Verwendung. Ein defekter Link liefert keine nutzbare
Vorschau. Korrigiere oder entferne ihn und speichere erneut.

Kann der Browser eine ausgewählte lokale Bilddatei nicht lesen, wird die Quelle im Entwurf nicht
ersetzt. Wähle eine lesbare unterstützte Datei und prüfe die Vorschau vor **Speichern**.

## Funktionstasten F0 bis F31 konfigurieren

Der Reiter **Funktionstasten** enthält F0 bis F31. Jede Zeile kann Folgendes aufnehmen:

- einen optionalen Funktionsnamen;
- ein optionales Funktionssymbol;
- einen Typ: `standard`, `licht`, `sound`, `kupplung`, `rauch` oder `sonderfunktion`.

Bei einem neuen Eintrag beginnt F0 mit **Fahrlicht**, Typ `licht` und Lichtsymbol. Alle anderen
Tasten beginnen unkonfiguriert mit Typ `standard`. Eine Zeile gilt als belegt, sobald sie einen
Namen, ein Symbol oder einen vom Standard abweichenden Typ besitzt. Die Zusammenfassung zählt
belegte Zeilen sowie die Typen `sound` und `licht`.

Die Auswahl eines Symbols übernimmt dessen Bezeichnung als Funktionsnamen, wenn die Auswahl eine
Bezeichnung liefert. Prüfe den Namen nach der Symbolauswahl. Beim **Speichern** legt RailKeeper nur
belegte Funktionen als strukturierte Daten ab. Tabelle, Detailansicht und Report zeigen nur diese
Funktionen.

v0.1.17.6 kann außerdem frühere Klartextwerte lesen, die durch Kommas, Semikolons oder Zeilen
getrennt sind und jeweils mit einer Taste wie `F1 Sound` beginnen. Das Öffnen und Speichern eines
solchen Eintrags schreibt die aktuell konfigurierte strukturierte Darstellung.

## Einträge lesen, sortieren und löschen

Die Eintragstabelle beginnt mit **Besitzer** aufsteigend und bietet diese sortierbaren
Überschriften:

- **Besitzer**;
- **Lok Bezeichnung**;
- **Steuerung**, sortiert nach der DCC-Adresse;
- **Funktionstasten**, sortiert nach der gespeicherten Funktionsdarstellung.

Die Bildspalte ist nicht sortierbar. Die Lokzelle verbindet Lok Bezeichnung, Bahnverwaltung,
Hersteller, Gattung, Epoche und Notizen, soweit Werte vorhanden sind. Die Steuerungszelle zeigt
Adresse, Analogzustand, Decoder und Schnittstelle.

Admin kann bei einem Eintrag einer offenen Liste **Löschen** wählen und bestätigen:

> Eintrag "BR 218" wirklich löschen?

Das Löschen ist sofort und dauerhaft. Danach lädt RailKeeper Einträge und Anzahl neu. Ein
allgemeiner Fahrzeugdatensatz wird nicht gelöscht. Messe erhält das Löschbedienelement nicht. Eine
gesperrte Liste weist die Serveranfrage auch für Admin zurück.

## Report ansehen und drucken

**Ansehen** öffnet eine schreibgeschützte Tabelle für eine Liste. Sie zeigt Bild, Besitzer und Tage,
Lokdaten, Steuerungsdaten und belegte Funktionstasten. **Drucken** in diesem Dialog führt weiter zu
**Report drucken**.

Alternativ stehen **Drucken** in einer Listenzeile und **Liste drucken** über der ausgewählten
Eintragstabelle zur Verfügung. Gehört die Aktion zu einer anderen Liste, lädt RailKeeper zuerst
deren Einträge. Die Aktion der ausgewählten Liste verwendet die aktuell angezeigten Einträge.

Lasse in **Report drucken** die Option **Mit Bildern drucken** aktiviert oder deaktiviere sie, um
die Bildspalte wegzulassen. Der erzeugte Report verwendet A4 im Querformat und enthält:

- RailKeeper-Zeichen, Listenbezeichnung, Datum und Eintragsanzahl;
- optional das Bild;
- Besitzer und Tagesumfang;
- Lokidentität sowie vorhandene Angaben zu Hersteller, Gattung und Epoche;
- Adressen, Analogzustand, Decoder und Schnittstelle;
- belegte Funktionstasten und Symbole;
- Notizen.

Mit **Drucken** öffnet sich der Druckdialog des Browsers. Papierauswahl, Ränder, Skalierung,
Druckerverfügbarkeit, Abbruch und endgültige Ausgabe bleiben Entscheidungen von Browser und
Betriebssystem. Eine leere Liste erzeugt eine Reportzeile mit **Keine Einträge.**

## Eintrags- und Druckfehler beheben

| Situation | Stabiles Ergebnis und Wiederherstellung |
| --- | --- |
| Besitzer oder Lok Bezeichnung fehlt | Browser oder Server weist Speichern ab. Fülle beide Pflichtfelder aus. |
| Doppelte DCC- oder SX-Adresse | Speichern bleibt deaktiviert. Lade neu, finde den anderen Eintrag und korrigiere eine Adresse. |
| Liste wurde gesperrt | Speichern oder Löschen wird abgewiesen. Lade den Status; Admin muss vor der Korrektur entsperren. |
| Stammdatenauswahlen zeigen nur **Keine Auswahl** | Das Konto darf allgemeine Bestandsstammdaten nicht lesen. Lösche gespeicherte Auswahlwerte nicht. |
| Bild hat keine Vorschau | Prüfe den Link oder wähle vor dem Speichern eine lesbare PNG-, JPEG- oder WebP-Datei. |
| Speichern gelingt, aber der Refresh scheitert | Öffne den Arbeitsbereich neu und prüfe den Eintrag vor dem Wiederholen. |
| Detail oder Report öffnet nicht | Lade Liste und Einträge neu und wiederhole danach die Leseaktion. |
| Browser-Druckdialog wird ohne Ausgabe geschlossen | Öffne **Report drucken** erneut; RailKeeper-Daten wurden nicht geändert. |

## Verwandte Seiten

- [Messearbeitsbereich](./)
- [Listen und Sperren](./lists-and-locking)
- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)

## Dokumentierte RailKeeper-Version

Diese Seite beschreibt RailKeeper v0.1.17.6. Der Entwicklungsstand auf `main` kann abweichen und
gehört nicht zu diesem Benutzerablauf.
