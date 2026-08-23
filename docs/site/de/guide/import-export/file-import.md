---
title: Fahrzeugdateien importieren
description: Fahrzeugzeilen aus CSV, TSV, XML oder JSON zuordnen, prüfen und speichern.
audience: user
status: stable
reviewedVersion: 0.1.20.1
lastReviewed: 2026-08-16
---

# Fahrzeugdateien importieren

Der Dateiimport ist ein Prüfvorgang, kein direkter Datenbankimport. RailKeeper liest die gewählte
Datei, ordnet ihre Spalten zu, prüft jede Zeile gegen die bereits geladenen Fahrzeuge und lässt
einen berechtigten Benutzer auswählen, welche Zeilen angelegt oder aktualisiert werden.

Zum Speichern ist ein Editor- oder Admin-Konto erforderlich. Viewer und Planner können die
Vorschau prüfen, ihre Fahrzeug-Schreibvorgänge werden aber vom Server abgelehnt.

## Unterstützte Eingabeformate

| Endung | Stabile Auswertung |
| --- | --- |
| `.csv` | Verwendet den häufigsten Trenner in der ersten nicht leeren Zeile: Tabulator, Semikolon oder Komma. Bei Gleichstand zwischen Semikolon und Komma gewinnt das Semikolon. |
| `.tsv` | Verwendet immer einen Tabulator als Trenner. |
| `.xml` | Liest Elemente mit den Namen `vehicle`, `fahrzeug`, `locomotive` oder `lok`. Fehlen diese, verwendet RailKeeper das am häufigsten wiederholte verschachtelte Element als Datensatzform. Attribute und direkte Text-Kindelemente werden zu Spalten. |
| `.json` | Akzeptiert ein Fahrzeug-Array oder ein Objekt mit einem `vehicles`-Array. Der stabile Import übernimmt nur die unten genannte JSON-Teilmenge. |

Getrennte Werte können in doppelte Anführungszeichen eingeschlossen werden. Doppelte
Anführungszeichen innerhalb eines solchen Feldes werden korrekt aufgelöst. Führende und
nachgestellte Leerzeichen werden aus jeder Zelle entfernt, vollständig leere Zeilen werden
ignoriert.

Ungültige JSON- oder XML-Dokumente werden nicht unterstützt. XML ohne auswertbare Datensätze meldet
eine leere Tabelle. Nach der Korrektur muss die Datei erneut ausgewählt werden.

## Teilmenge beim JSON-Import

Der JSON-Export verpackt die Fahrzeugliste als `{ "format": "railkeeper-vehicles", "version": 1,
"vehicles": [...] }`. Der stabile JSON-Import prüft `format` und `version` nicht und übernimmt nicht
jede Eigenschaft. Er erzeugt seine Prüftabelle nur aus diesen 14 Feldern:

| Identität | Einordnung | Betrieb und Wert |
| --- | --- | --- |
| Inventarnummer, Hersteller, Artikel-Nr., Bezeichnung | Spurweite, Epoche, Bahngesellschaft, Kategorie, Gattung | Höchstgeschwindigkeit, Heimat-Bw / Einsatzstelle, Digital, Digital / Decoder-Nr., Listenpreis |

Verwende CSV für den vollständigen unterstützten Austausch der 62 skalaren Felder. Verwende eine
Anwendungssicherung, wenn wiederherstellbare Anwendungsdaten und Uploads benötigt werden.

## Spaltenzuordnung

Die erste Zeile liefert die Spaltennamen. RailKeeper normalisiert Groß- und Kleinschreibung,
Leerzeichen, deutsche Umlaute, übliche Satzzeichen sowie bekannte deutsche und englische Aliasse,
bevor nach einem Zielfeld gesucht wird.

Nach dem Laden einer Datei:

1. Unter **Erkannte CSV-Zuordnung** jede Quellüberschrift prüfen.
2. Jede benötigte offene Spalte einem RailKeeper-Feld zuweisen.
3. Eine nicht benötigte Quellspalte ausdrücklich auf **Spalte ignorieren** setzen.
4. Bei wiederkehrenden Dateien optional **Zuordnung im Profil speichern** wählen.
5. **Zuordnung prüfen** und danach die neu berechnete Importprüfung kontrollieren.

Ein Zielfeld kann nur einer Quellspalte gehören. Bereits verwendete Ziele stehen bei anderen
Spalten nicht erneut zur Auswahl. Die Zuordnungsprüfung liest dieselbe Datei serverseitig neu und
ersetzt die dauerhafte Vorschau, daher anschließend Konflikte und Korrekturen erneut prüfen.

## Alle 62 CSV- und Tabellenziele

Der vollständige stabile CSV-Export verwendet dieselben Ziele und kann deshalb ohne manuelle
Neuzuordnung importiert werden. Andere Dateien dürfen jede beliebige Teilmenge verwenden.

| Gruppe | Unterstützte Felder |
| --- | --- |
| Identität und Katalog | Inventarnummer; Hersteller; Artikel-Nr.; Quelle / URL; Bezeichnung; Spurweite; Epoche; Bahngesellschaft; Kategorie; Gattung; Beschreibung; Baureihe; Fahrzeug-Nr.; EAN; Produktionszeit; Listenpreis |
| Betrieb, Digital und Messe | Höchstgeschwindigkeit; Heimat-Bw / Einsatzstelle; Digital; Digital / Decoder-Nr.; Decoder-Typ; DT / Decoder; DT / Decoder-Nr.; Messe tauglich; Ausstellung; ABC-Bremsen; QR-Code erstellen |
| Erwerb, Lagerung und Zustand | Erwerbsart; Erworben von/bei; Kaufpreis; Kaufdatum; Lagerort; Lagerdetails; Zustand; Zustandsdetails; Verpackung |
| Physische und mechanische Daten | Länge (mm); Gewicht (g); Farbe; Beschriftung; Beladung; Inneneinrichtung; Achsen; Anzahl Achsen; Anzahl Haftreifen; Radsatz; Kupplung (V=H); Kupplung vorne; Kupplung hinten; Stromabnahme; Adapter / Schnittstelle |
| Ausstattung und Hinweise | Antrieb; Antrieb Beschreibung; Fahrlicht; Fahrlicht Beschreibung; Beleuchtung; Beleuchtung Beschreibung; Soundgenerator; Soundgenerator Beschreibung; Rauchgenerator; Rauchgenerator Beschreibung; Zusatzinformationen |

Bilder sowie andere verschachtelte oder dateibasierte Datensätze sind keine Ziele.

## Boolesche Werte

Bei Digital, DT / Decoder, Messe tauglich, Ausstellung, ABC-Bremsen, Kupplung (V=H), Antrieb,
Fahrlicht, Beleuchtung, Soundgenerator, Rauchgenerator und QR-Code erstellen bedeuten diese nicht
leeren Werte unabhängig von Groß- und Kleinschreibung **ja**:

`1`, `ja`, `yes`, `true`, `wahr`, `digital`, `d`, `x`, `vorhanden`

Jeder andere nicht leere Wert wird zu **nein**. Eine leere Zelle wird ignoriert. Verwende möglichst
die eindeutigen Ja-/Nein-Werte des RailKeeper-CSV-Exports, um mehrdeutige Quelldaten zu vermeiden.

## Prüfung und Vorauswahl

Eine neue Zeile benötigt Hersteller, Bezeichnung, Spurweite, Kategorie und Gattung, bevor sie in
der Vorschau ausgewählt werden kann. Beim Speichern kann der Server weitere Fahrzeugregeln prüfen.

Zwei feldbezogene Prüfungen laufen bereits im Browser:

- Die Höchstgeschwindigkeit muss eine ganze Zahl von 1 bis 1000 km/h sein.
- Heimat-Bw / Einsatzstelle darf höchstens 200 Unicode-Zeichen enthalten.

Eine ungültige Zeile wird als Fehler markiert, abgewählt und kann bis zur Korrektur des sichtbaren
Problems oder der Quellzuordnung nicht ausgewählt werden. In der Prüftabelle lassen sich nur
Inventarnummer, Hersteller, Artikel-Nr., Bezeichnung, Spurweite, Kategorie und Gattung direkt
bearbeiten. Andere Werte müssen in der Quelldatei oder über die Spaltenzuordnung korrigiert und bei
Bedarf neu geladen werden.

## Neue Fahrzeuge und erkannte Duplikate

RailKeeper vergleicht eine nicht leere Inventarnummer ohne Beachtung der Groß- und Kleinschreibung
mit der aktuellen Fahrzeugliste.

| Ergebnis | Standardverhalten |
| --- | --- |
| Kein Inventarnummer-Treffer und kein Prüfproblem | Aktion **Neu**, Zeile ausgewählt |
| Bestehende Inventarnummer | Aktion **Aktualisieren**, Warnung, Zeile nicht ausgewählt |
| Bestehende Inventarnummer plus weiteres Prüfproblem | Aktion **Aktualisieren**, Fehler, Zeile gesperrt |

Bei einem Update die Feldvorschau aufklappen und **gleich**, **ergänzt leeres Feld** sowie
**überschreibt** prüfen. Übernommen werden nur zugeordnete, nicht leere Zeichenketten, gültige
Zahlen und ausgewertete boolesche Werte. Leere Quellzellen löschen vorhandene Zeichenketten nicht.
Ein nicht leerer boolescher Wert außerhalb der Ja-Liste übernimmt ausdrücklich **nein**.

Wird eine Duplikatzeile auf **Neu** umgestellt, entsteht kein zweites Fahrzeug mit derselben
Inventarnummer. RailKeeper markiert die Zeile stattdessen als Fehler.

## Geprüfte Auswahl speichern

1. Sichtbare Grundfelder korrigieren und jeden Update-Vergleich prüfen.
2. Nur die zu schreibenden Zeilen auswählen.
3. **Auswahl speichern** wählen.
4. Warten, bis jede erfolgreiche Zeile **gespeichert** anzeigt.
5. Zeilenbezogene Fehler lesen und bereits gespeicherte Fahrzeuge vor einem erneuten Versuch
   prüfen.

Zeilen werden nacheinander verarbeitet. Jede ausgewählte Zeile ruft entweder die Fahrzeuganlage
oder die Fahrzeugaktualisierung auf. Scheitert eine Zeile, wird sie zum Fehler und RailKeeper fährt
mit späteren ausgewählten Zeilen fort. Der Vorgang ist daher nicht atomar und besitzt kein
gemeinsames Rückgängig.

Dateiimporte erzeugen keine Funktionszuordnungen, CV-Einträge, ECoS-Zuordnungen, Bilder, Beilagen,
Wartungen, Ersatzteile oder Decoder-Dateien. Ergänze diese über die jeweiligen Fahrzeugabläufe.

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.20.1** und wurde zuletzt am 16.08.2026 geprüft.
