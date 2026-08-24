---
title: Listen und Sperren
description: Messelisten anlegen, sperren, entsperren und sicher löschen.
audience: user
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-16
---

# Listen und Sperren

Eine Messeliste bündelt die Lokomotiveinträge für eine benannte Messe oder einen Fahrtag. Admin
bereitet Listen vor und steuert sie; Messe konzentriert sich auf die Betriebseinträge einer offenen
Liste.

## Rollen für Betrieb und Verwaltung

| Aktion | Messe | Admin |
| --- | --- | --- |
| Listen auflisten, auswählen, ansehen und drucken | Ja | Ja |
| Liste anlegen oder bearbeiten | Nein | Ja |
| Liste sperren oder entsperren | Nein | Ja |
| Liste löschen | Nein | Ja |
| Einträge einer offenen Liste pflegen | Ja | Ja |
| Eintrag einer offenen Liste löschen | Nein | Ja |

Admin erhält direkten Zugriff auf den Messearbeitsbereich und benötigt keine zusätzliche Rolle
Messe. Editor, Viewer und Planner öffnen den Arbeitsbereich allein nicht und erteilen keine
Listenverwaltung.

## Liste anlegen

Wähle **Neue Liste** und fülle den Dialog **Messeliste anlegen** aus:

| Feld | Regel |
| --- | --- |
| Bezeichnung | Pflichtfeld. RailKeeper entfernt beim Speichern führende und nachgestellte Leerzeichen. |
| Datum | Pflichtfeld. Ein neuer Dialog beginnt mit dem vom Browser erzeugten UTC-Kalendertag. Prüfe ihn nahe der lokalen Mitternacht. |

Wähle **Speichern**. Eine erfolgreiche Anfrage legt eine offene Liste an, wählt sie aus, schließt
den Dialog und lädt die Listentabelle neu. Das Anlegen ist ein sofortiger Server-Schreibvorgang;
bis zum Erfolg sind die Dialogwerte nur lokaler Formularzustand.

Ist eines der Pflichtfelder leer, verhindert der Browser die normale Übermittlung. Der Server weist
auch Werte zurück, die nach dem Entfernen äußerer Leerzeichen leer sind. In v0.1.20.2 prüft der
Server beim Datum, dass der Wert nicht leer ist; das Datumsfeld liefert den normalen Kalenderwert.

## Listen auswählen, sortieren und prüfen

Wähle eine Zeile, um sie zur aktiven Liste zu machen. Die Eintragstabelle lädt anschließend die
Einträge dieser Liste. Die ausgewählte Zeile bleibt hervorgehoben.

Die Listentabelle beginnt mit **Datum** absteigend. Diese Überschriften sortieren im Browser:

- **Bezeichnung**;
- **Datum**;
- **Einträge**;
- **Status**.

Die Auswahl einer anderen Überschrift beginnt aufsteigend. Die aktive aufsteigende Sortierung wird
beim nächsten Klick absteigend und beim darauffolgenden wieder aufsteigend.

**Ansehen** lädt die aktuellen Einträge und öffnet eine schreibgeschützte Tabelle. Die Aktion steht
Messe und Admin für offene und gesperrte Listen zur Verfügung. **Drucken** verwendet denselben
Leseweg und öffnet die unter [Einträge und Drucken](./entries-and-printing) beschriebenen
Reportoptionen.

## Bezeichnung oder Datum bearbeiten

Admin wählt **Bearbeiten** in der gewünschten Zeile. **Messeliste bearbeiten** öffnet sich mit der
gespeicherten Bezeichnung und dem Datum. Ändere eines oder beide Felder und wähle **Speichern**.

Das Bearbeiten einer Liste ändert weder ihre Einträge noch ihren Sperrzustand. Es ist bei offenen
und gesperrten Listen erlaubt. Nach einem erfolgreichen Schreibvorgang bleibt die gespeicherte
Liste ausgewählt und RailKeeper lädt alle Listen neu. Schlägt das Speichern fehl, bleibt der Dialog
offen und die Meldung oberhalb des Arbeitsbereichs zeigt den Fehler. Korrigiere die Werte oder lade
neu, bevor du es erneut versuchst.

## Liste sperren oder entsperren

Wähle **Sperren**, sobald keine Einträge mehr geändert werden sollen. RailKeeper schreibt den neuen
Zustand sofort und ersetzt **offen** in der Tabelle durch **gesperrt**. Es erscheint kein
Bestätigungsdialog.

Die Sperre hat diese genaue Grenze:

- Auswahl, Lesen, **Ansehen** und **Drucken** funktionieren weiter;
- Bezeichnung und Datum können weiterhin von Admin bearbeitet werden;
- die Liste kann weiterhin von Admin gelöscht werden;
- Anlegen, Bearbeiten und Löschen von Einträgen werden abgewiesen;
- deaktivierte Bedienelemente und Servervalidierung erzwingen gemeinsam die Eintragssperre.

Wähle **Entsperren**, um die Eintragspflege wieder zuzulassen. Auch das Entsperren erfolgt sofort
und ohne Bestätigung. Schlägt eine Sperranfrage fehl oder ist der angezeigte Zustand unklar, lade
den Arbeitsbereich neu und lies die Spalte **Status**, bevor du einen Eintrag änderst.

## Liste löschen

Admin wählt **Löschen** und bestätigt:

> Messeliste "Dortmund 2026" wirklich löschen?

Das Löschen ist dauerhaft. Mit der Liste werden über die Datenbankbeziehung auch alle zugehörigen
Messelisteneinträge entfernt. Allgemeine Fahrzeugdatensätze werden nicht gelöscht. Exportiere vor
dem Löschen einer gefüllten Liste ein aktuelles validiertes App-Backup, falls eine Wiederherstellung
nötig werden könnte.

Nach erfolgreichem Löschen entfernt RailKeeper die Auswahl, wenn sie auf diese Liste zeigte, und
lädt die verbleibenden Listen neu. Brich die Bestätigung ab, um die Liste unverändert zu lassen.

## Listenfehler beheben

| Situation | Stabiles Ergebnis und Wiederherstellung |
| --- | --- |
| Leere Bezeichnung oder leeres Datum | Speichern wird abgewiesen. Fülle beide Pflichtfelder aus. |
| Liste existiert nicht mehr | Der Server meldet nicht gefunden. Lade den Arbeitsbereich neu und wähle eine vorhandene Liste. |
| Fehlende Admin-Berechtigung | Anlegen, Bearbeiten, Sperren, Entsperren und Löschen sind verboten. Verwende ein berechtigtes Adminkonto. |
| Speichern schlägt fehl | Der Dialog bleibt offen und eine Meldung erscheint. Prüfe Werte und Speicherzustand vor dem Wiederholen. |
| Sperren oder Entsperren wirkt unverändert | Lade vor der Eintragspflege neu; der Serverzustand ist maßgeblich. |
| Löschergebnis ist unklar | Lade neu und prüfe, ob die Liste noch vorhanden ist, bevor du erneut **Löschen** wählst. |
| Eintragsvorgang meldet eine gesperrte Liste | Lade den Status neu. Nur Admin kann die Liste gezielt entsperren. |

Erfolgreiches Anlegen, Bearbeiten, Sperren, Entsperren und Löschen wird im RailKeeper-Audit-Log
erfasst. Die Bedienung des Audit-Logs gehört zur Administration und wird auf dieser Benutzerseite
nicht erklärt.

## Verwandte Seiten

- [Messearbeitsbereich](./)
- [Einträge und Drucken](./entries-and-printing)
- [Übersicht des Benutzerhandbuchs](/de/guide/)

## Dokumentierte RailKeeper-Version

Diese Seite beschreibt RailKeeper v0.1.20.2. Der Entwicklungsstand auf `main` kann abweichen und
gehört nicht zu diesem Benutzerablauf.
