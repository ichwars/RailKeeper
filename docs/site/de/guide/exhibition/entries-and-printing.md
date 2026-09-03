---
title: Einträge und Drucken
description: Messeloks erfassen, Adresskonflikte lösen und die Betriebsliste drucken.
audience: user
status: stable
reviewedVersion: 0.1.20.4
lastReviewed: 2026-08-31
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
**Fahrzeugbild** und **Funktionstasten**. Änderungen in jedem Reiter bleiben Formularzustand im
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

Benutzer mit Leserecht auf den Bestand können optional ein **Bestandsfahrzeug** wählen. RailKeeper
übernimmt dessen Modell- und Decoderfelder in den Messeentwurf und speichert die
Fahrzeugverknüpfung. Veranstaltungsdaten bleiben unabhängig und ändern keine Fahrzeug-Stammdaten.
Reine Messe-Konten können **Gastfahrzeug / manuelle Eingabe** verwenden, aber nicht im allgemeinen
Bestand suchen.

Mit **Bearbeiten** in einer vorhandenen Zeile öffnet sich **Eintrag bearbeiten** mit den aktuellen
Werten.

Wähle **Speichern** erst, nachdem du alle drei Reiter geprüft hast. Während des Speicherns oder bei
dem laufenden Zugriff ist die Schaltfläche deaktiviert. Wurde die Liste nach dem Öffnen des Dialogs
gesperrt, weist der Server den Vorgang ebenfalls zurück.

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

Die Tabelle verbindet DCC und SX unter **Adresse**. Der aktuelle Druckreport zeigt beide Adressen,
Analogzustand, Decoder-Typ, Adapter und Zentrale getrennt. Leere Werte erscheinen als Strich.

**Baureihe** wird mit dem Eintrag gespeichert und im aktuellen Druckreport ausgegeben. In der
kompakten Bildschirmtabelle ist sie nicht enthalten. Öffne **Eintrag bearbeiten**, um den Wert zu ändern.

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

## Betriebskonflikte lösen

Der Server prüft alle Einträge der ausgewählten Liste und meldet fehlende Pflichtangaben, ein
mehrfach verwendetes Bestandsfahrzeug und überschneidende Digitaladressen. Ein Adresskonflikt
erfordert dieselbe Schnittstelle und Adresse an überlappenden Messetagen. Analoge oder nicht
verfügbare Einträge erzeugen keinen Adresskonflikt.

Nach dem Speichern aktualisiert RailKeeper den Arbeitsbereich und seine Konfliktanzahl. Öffne
**Konflikte prüfen**, um die betroffenen Datensätze zu sehen. Korrigiere den Eintrag, wenn möglich.
Ist die Überschneidung beabsichtigt, trage eine Begründung ein und speichere eine dokumentierte
Ausnahme. Der Konflikt bleibt als Ausnahme sichtbar. Das Sperren einer Liste mit ungelösten
Konflikten erfordert eine ausdrückliche Bestätigung und eine Begründung für jede Ausnahme.

## Bild hinzufügen oder entfernen

Der Reiter **Fahrzeugbild** speichert ein eingebettetes Bild pro Eintrag.

| Aktion | Stabiles Verhalten |
| --- | --- |
| **Bild auswählen** | Akzeptiert PNG, JPEG oder WebP bis 10 MB und liest die Datei als eingebettete Daten in den Entwurf. |
| Bild ersetzen | Eine andere unterstützte Datei ersetzt das Bild im Entwurf. |
| **Bild entfernen** | Leert die Quelle im Entwurf. Dauerhaft wird dies erst nach erfolgreichem Speichern des Eintrags. |

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

v0.1.20.3 kann außerdem frühere Klartextwerte lesen, die durch Kommas, Semikolons oder Zeilen
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

## Ausstellungsliste drucken

Die separate Druckansicht ist ab RailKeeper v0.1.20.4 verfügbar.

Wähle die Veranstaltung und anschließend **Drucken** in ihrer Übersicht. RailKeeper erzeugt ein
separates Dokument im A4-Querformat. Navigation, andere Veranstaltungen, Filter und Schaltflächen
erscheinen darin nicht. Gedruckt werden sämtliche gespeicherten Einträge der ausgewählten
Ausstellung, unabhängig von Suchtext, Tages- und Statusfiltern auf dem Bildschirm.

Der Report enthält:

- Veranstaltungsname, Zeitraum, Ort, Status, Eintragsanzahl, Beschreibung und Organisationshinweise;
- je Fahrzeug Bild, Besitzer, Fahrtage, Verfügbarkeit und Prüfstatus;
- Lokbezeichnung, Hersteller, Baureihe, Gattung, Epoche und Bahnverwaltung;
- Digitaldecoder, Decoder-Typ, Adapter, DCC- und SX-Adresse, Zentrale und Analogzustand;
- alle gespeicherten Funktionstasten mit Beschreibung, Typ und vorhandenem Symbol;
- vollständige Notizen mit Zeilenumbrüchen.

Funktionstasten stehen über die volle Listenbreite. Ältere Freitextbelegungen bleiben erhalten;
fehlende Belegungen werden nicht durch Standardfunktionen ergänzt. Leere Listen zeigen
**Keine Einträge.** Beim Wechsel der Veranstaltung ist Drucken gesperrt, bis ihre Daten geladen sind.

RailKeeper wartet auf das Laden der Druckansicht einschließlich ihrer Bilder. Danach öffnet sich
der Druckdialog des Browsers. Papierauswahl, Skalierung, Druckerverfügbarkeit, Abbruch und endgültige
Ausgabe bleiben Entscheidungen von Browser und Betriebssystem. Verwende den **Drucken**-Knopf der
Ausstellung, nicht den allgemeinen Browserbefehl zum Drucken der Webseite.

## Eintrags- und Druckfehler beheben

| Situation | Stabiles Ergebnis und Wiederherstellung |
| --- | --- |
| Besitzer oder Lok Bezeichnung fehlt | Browser oder Server weist Speichern ab. Fülle beide Pflichtfelder aus. |
| Konfliktanzahl steigt nach dem Speichern | **Konflikte prüfen** öffnen, Datensätze korrigieren oder eine begründete Ausnahme dokumentieren. |
| Liste wurde gesperrt | Speichern oder Löschen wird abgewiesen. Lade den Status; Admin muss vor der Korrektur entsperren. |
| Stammdatenauswahlen zeigen nur **Keine Auswahl** | Das Konto darf allgemeine Bestandsstammdaten nicht lesen. Lösche gespeicherte Auswahlwerte nicht. |
| Bild hat keine Vorschau | Vor dem Speichern eine lesbare PNG-, JPEG- oder WebP-Datei bis 10 MB wählen. |
| Speichern gelingt, aber der Refresh scheitert | Öffne den Arbeitsbereich neu und prüfe den Eintrag vor dem Wiederholen. |
| Detail oder Report öffnet nicht | Lade Liste und Einträge neu und wiederhole danach die Leseaktion. |
| Browser-Druckdialog wird ohne Ausgabe geschlossen | Wähle **Drucken** erneut; RailKeeper-Daten wurden nicht geändert. |

## Verwandte Seiten

- [Messearbeitsbereich](./)
- [Listen und Sperren](./lists-and-locking)
- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)

## Dokumentierte RailKeeper-Version

Diese Seite beschreibt RailKeeper v0.1.20.4.
