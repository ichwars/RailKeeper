---
title: Ausstellungen und Fahrtage organisieren
description: Ausstellungen und Fahrtage organisieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Ausstellungen und Fahrtage organisieren

Öffne **Ausstellung** und wähle links eine Veranstaltung. Oben stehen Termin und Status;
rechts verwaltest du Teilnehmer und Fahrzeuge. [Die bebilderte Übung](practice) führt durch einen vollständigen Ablauf.

<GuideFigure name="exhibition-open" alt="Aktueller Veranstaltungsarbeitsbereich mit Kennzahlen, Tagesfiltern und Fahrzeugen." caption="Die tatsächliche Oberfläche mit künstlichen Daten. Ein vorhandener Prüffall bleibt sichtbar und wird nicht als betriebsbereit ausgegeben." />

## Zugriffsrechte und Trennung

| Konto | Aktueller Umfang |
| --- | --- |
| Admin | Arbeitsbereich, Veranstaltungsplanung, Status und Löschen |
| Messe | Bestehende Veranstaltungen nutzen, Einträge während offener Planung pflegen, prüfen und drucken |
| Editor zusammen mit Messe | Zusätzlich Veranstaltungen und Status verwalten |
| Viewer zusammen mit Messe | Messe-Abläufe wie oben; Viewer allein öffnet diesen Arbeitsbereich nicht |

**Editor** allein hat in der allgemeinen Navigation keinen Ausstellungszugang. Ergänze **Messe**,
wenn dieselbe Person auch diese Aufgabe ausführen soll. Löschen ist Admin vorbehalten.
Betriebseinträge ändern keine Fahrzeug-Stammdaten. Der Fahrzeugschalter **Ausstellung** legt keinen Eintrag an.

## Listen und Einträge verstehen

Eine Veranstaltung hat Bezeichnung, Beginn, Ende, Ort und organisatorische Angaben.
Ihre Einträge beschreiben Besitzer, Lok, Zentrale, Adresse, Fahrtage und Verfügbarkeit.
Ein Bestandsfahrzeug kann optional zugeordnet werden; Gastfahrzeuge werden manuell erfasst.

Die Kennzahlen zeigen Einträge, Besitzer, Konflikte und bereite Fahrzeuge.
**Bereitschaft** trennt Adressprüfung, Funktionstasten und Bilder. Bereit bedeutet nicht,
dass eine Lok technisch gefahren oder von RailKeeper gesteuert wurde.

## Dem Messeablauf folgen

1. **Neue Veranstaltung** anlegen und Termin prüfen.
2. Unter **Teilnehmer und Fahrzeuge** Einträge hinzufügen.
3. **Alle Tage** oder einen einzelnen Fahrtag wählen; mit **Alle**, **Bereit**, **Prüfen** und der Suche eingrenzen.
4. **Konflikte prüfen** und fehlende Angaben oder Adressüberschneidungen bearbeiten.
5. Liste sperren, Bericht drucken und später den Veranstaltungsstatus weiterführen.

[Veranstaltungen und Sperren](lists-and-locking) · [Einträge und Bericht](entries-and-printing).

## Mit einer offenen oder gesperrten Liste arbeiten

| Status | Bedeutung für den Ablauf |
| --- | --- |
| Entwurf | Planung zunächst öffnen, bevor der weitere Betrieb vorbereitet wird |
| Offen | Einträge pflegen und prüfen |
| Gesperrt | Planung festgehalten, Einträge nicht bearbeitbar; Drucken bleibt möglich |
| Läuft | Veranstaltung gestartet, Einträge bleiben geschützt |
| Abgeschlossen | Betrieb beendet; Planung kann wieder geöffnet werden |
| Archiviert | Standardmäßig ausgeblendet; über **Alle Veranstaltungen anzeigen** oder das Aktionsmenü auffindbar |

Die Oberfläche erlaubt Eintragsänderungen nur in **Entwurf** oder **Offen** und ohne Sperre.
**Weitere Aktionen** bietet die zum Status passenden Übergänge. Eine Sperre schützt keine Hardware.

## Lade-, Leer- und Fehlerzustände lesen

Warte, bis die gewählte Veranstaltung geladen ist. Bei einer leeren Liste lege den ersten Eintrag an.
Bei einer veralteten Revision lade den Stand neu und vergleiche vor dem Speichern.
Bei **Angaben fehlen** öffne den Eintrag und die Konfliktanzeige; bei **Nicht verfügbar** kläre die Teilnahme.
Ein fehlendes Bild oder fehlende Funktionstasten können weiterhin **Prüfen** ergeben.

## Mit einem gezielten Ablauf fortfahren

- [Bebilderter Fahrtag](practice)
- [Planung, Status, Konfliktsperre](lists-and-locking)
- [Einträge, Bilder, Funktionen und Drucken](entries-and-printing)

## Verwandte Seiten

[Fahrzeugbestand](../vehicles/) · [Konten und Rollen](../../administration/users-sessions-security).
