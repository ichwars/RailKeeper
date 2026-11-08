---
title: Decoderexemplare und Einbauten
description: Einzelne Decoderexemplare, tatsächliche Einbauzeiträume und Zuordnungen von CV-Gruppen erfassen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-08
---

# Decoderexemplare und Einbauten

Öffne bei einem gespeicherten Fahrzeug **CV → Decoder und CV-Gruppen**.
Editor und Admin können Änderungen bestätigen, Viewer kann die Angaben lesen.
Die Einträge beschreiben tatsächliche Exemplare und Einbauten im Bestand.

<GuideFigure name="decoder-installations" alt="Einzelne Decodereinbauten und CV-Gruppen eines synthetischen Fahrzeugs." caption="Der synthetische Bestand zeigt getrennte Exemplare, erfasste Ein- und Ausbauten sowie ausdrücklich unbekannte Datumsangaben." />

## Exemplar und Einbau erfassen

1. Wähle **Decoderexemplar erfassen** und vergib eine Bezeichnung, an der du dieses
   einzelne Exemplar erkennst. Hersteller, Modell, Seriennummer und Notizen sind
   optional. Bestätige **Exemplar jetzt speichern**.
2. Öffne **Alle erfassten Decoderexemplare**, suche das Exemplar und wähle
   **Einbau erfassen**. Ein Exemplar mit aktivem Einbau steht für keinen weiteren
   gleichzeitigen Einbau zur Verfügung.
3. Trage das bekannte Einbaudatum als **JJJJ-MM-TT** ein oder lasse es bei
   unbekanntem Datum leer. Bestätige **Einbau jetzt speichern** für dieses Fahrzeug.

Gleiche Modelle, Bezeichnungen oder CV-Profile belegen kein gemeinsames Exemplar.
Ein Fahrzeug darf mehrere eingebaute Decoder haben. Die Exemplarkennung
unterscheidet Einträge auch bei gleicher Beschreibung.

**Exemplar bearbeiten** berichtigt die Beschreibung dieses Exemplars über seine
Einbauten hinweg. Bereits gespeicherte CV-Teilstände behalten ihre damalige Beschreibung.

## Ausbau und berichtigte Daten

Wähle am Einbau **Ausbau erfassen**, trage ein bekanntes Ausbaudatum ein oder lasse
es unbekannt und bestätige. Eine zusätzliche Notiz ergänzt die bisherigen
Einbaunotizen. CV-Gruppen, Quellen und gespeicherte Teilstände bleiben erhalten.

**Datumsangaben berichtigen** ändert bekannte Daten. Ein irrtümlich eingetragenes
Datum lässt sich ausdrücklich zu unbekannt leeren. Der Zustand eingebaut,
ausgebaut oder ungeklärt bleibt erhalten. Ungültige Daten und ein Ausbau vor dem
Einbau lassen sich nicht bestätigen. Die Erfassungszeit ersetzt kein unbekanntes
tatsächliches Ein- oder Ausbaudatum.

## CV-Gruppen und gespeicherte Teilstände

Vorhandene CV-Profile bleiben CV-Gruppen mit ihren genauen Bezeichnungen,
Werten und Historien. **CV-Gruppe anlegen** ergänzt eine leere Gruppe, ohne
vorhandene Daten umzubenennen oder zusammenzuführen und ohne automatische Hardwarezuordnung.

Wähle bei einer Gruppe **Zuordnung prüfen**, dann einen aktiven Einbau dieses
Fahrzeugs oder ausdrücklich keine Zuordnung und bestätige. Vor der Änderung
erfasst RailKeeper den gespeicherten Teilbestand mit seiner bisherigen
Hardwarebeschreibung. Die CV-Werte selbst bleiben in der CV-Gruppe erhalten.

**Teilstand erfassen** hält bewusst die derzeit gespeicherten CVs, echten
Historien, Quellen und die aktuelle Hardwarebeschreibung fest. Eine Notiz macht
den Stand wiedererkennbar. Offene CV-Entwürfe gehören nicht dazu. Das Ergebnis
steht im [CV-Archiv](cv-archive). Seine Erfassungszeit belegt keine vollständige Auslesung.

## Ziel für neue CV-Eingaben bestätigen

<GuideFigure name="cv-working-groups" alt="Bestätigte CV-Zielgruppe mit der aktuellen Decoderzuordnung." caption="Die Voreinstellung gilt für neue Eingaben. Bereits vorbereitete Änderungen behalten ihr eigenes Ziel." />

Im Arbeitsbestand wählst du eine **Voreinstellung für neue CV-Eingaben und
Dateiübernahmen**. Prüfe die angezeigte Gruppe und Decoderzuordnung, dann wähle
**CV-Zielgruppe jetzt bestätigen**. Die Auswahl allein ändert das Ziel noch nicht.
**Profil der jeweiligen Eingabe verwenden** erhält die getrennten Profile der Eingaben.
Signifikante Leerzeichen in alten Profilbezeichnungen erscheinen in Anführungszeichen.

Vorgemerkte Änderungen, bereits geöffnete CV-Zeilen und vorbereitete Dateiübernahmen
behalten ihre ursprüngliche Zielgruppe bei einem späteren Wechsel dieser Voreinstellung.
Die Importvorschau nennt abweichende Quellenprofile zusätzlich zum bestätigten Ziel.
Die Ablage einer Originaldatei und die Übernahme ihrer CV-Werte bleiben getrennte Bestätigungen.

Ändert sich die Zuordnung eines beteiligten Decoders, bleibt die Änderung erhalten
und wird gesperrt. Lade die CV-Gruppen neu und bestätige gegebenenfalls die
Voreinstellung erneut. Vergleiche anschließend die betroffene CV-Änderung oder den
Import erneut. Prüfe die aktuelle Zuordnung und bestätige erst dann die Übernahme.
Bei Dateien bietet **Dateiziel erneut vergleichen** denselben Ablauf. Dateien und
Beschreibung bleiben erhalten; der Vergleich startet keinen Upload.

## Bestätigung und Wiederaufnahme

Jedes Formular beschreibt die Wirkung seiner eigenen Bestätigung. Ein Wechsel
zwischen Fahrzeugreitern erhält das offene Decoderformular und andere Entwürfe.
Abbrechen am Fahrzeug verwirft unbestätigte Formulare. Bereits bestätigte
Decoderaktionen bleiben gespeichert.

Hat ein anderer Schreiber den Stand geändert oder fehlt eine eindeutige
Speicherantwort, bleibt deine Eingabe erhalten und die Bestätigung ist gesperrt.
Wähle **Aktuellen Stand laden und prüfen**, vergleiche die angezeigten Einträge
und bestätige bei Bedarf erneut. Ein bereits angelegtes Exemplar, ein Einbau,
eine Gruppe oder ein Teilstand wird an seiner ursprünglichen Kennung erkannt
und nicht erneut angelegt.

Der Vergleich übernimmt gespeicherte Berichtigungen in Angaben, die du nicht
geändert hast. Wurde dieselbe Angabe beidseitig geändert, wähle dort ausdrücklich
**Gespeicherten Wert übernehmen** oder **Meine Eingabe behalten**. Solange ein
Angabenkonflikt offen ist, bleibt Speichern gesperrt.

Eine [vollständige kompatible Sicherung](../../administration/backup-restore) erhält
Einbauten, Zuordnungen und Archiv gemeinsam. Teilstand-JSON dient als lesbarer
Datenexport. Aufbewahrte Originaldateien haben gesonderte Downloadlinks.
