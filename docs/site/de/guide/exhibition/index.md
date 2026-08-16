---
title: Messearbeitsbereich
description: Messelisten sicher vorbereiten, pflegen, prüfen und drucken.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Messearbeitsbereich

Der Arbeitsbereich **Messeliste** bündelt die für eine Messe oder einen Fahrtag benötigten Angaben
in einer Betriebsansicht. Jede Liste besitzt ein Datum, den Zustand offen oder gesperrt und eigene
Lokomotiveinträge. Das Betriebspersonal kann eine offene Liste pflegen, ohne den allgemeinen
Fahrzeugbestand zu öffnen.

Dieses Kapitel dokumentiert RailKeeper v0.1.17.6. Benutzer- und Stammdatenverwaltung,
Backup-Bedienung sowie der Anlagen-Arbeitsbereich gehören nicht zu diesem Kapitel.

## Zugriffsrechte und Trennung

RailKeeper stellt den Arbeitsbereich zwei Rollen mit unterschiedlichen Aufgaben bereit:

| Rolle | Stabiler Zugriff |
| --- | --- |
| Messe | Öffnet den getrennten Messearbeitsbereich. Die Rolle kann Listen lesen und drucken sowie Einträge pflegen, solange die Liste offen ist. |
| Admin | Öffnet den Arbeitsbereich und kann zusätzlich Listen anlegen, bearbeiten, sperren, entsperren und löschen. Admin kann auch Einträge einer offenen Liste löschen. |

Editor, Viewer und Planner öffnen den Messearbeitsbereich allein nicht. Wird eine dieser Rollen mit
Messe kombiniert, kann das Eintragsformular zusätzlich allgemeine Stammdatenauswahlen des
Fahrzeugbestands laden. Ein reines Messekonto bleibt von diesen Auswahlen getrennt, kann jedoch die
im Dialog vorhandenen Messefelder verwenden. Funktionssymbole stehen Messe weiterhin zur
Verfügung.

Der Server prüft dieselbe Trennung. Admin wird für jede geschützte Messeanfrage akzeptiert, Messe
für das Lesen von Listen sowie das Anlegen und Bearbeiten von Einträgen. Ein ausgeblendetes oder
deaktiviertes Bedienelement ist daher nicht der einzige Schutz gegen unerlaubte Schreibvorgänge.

## Listen und Einträge verstehen

Die Seite enthält zwei zusammengehörige Tabellen:

| Bereich | Aufgabe |
| --- | --- |
| Listen | Zeigt **Bezeichnung**, **Datum**, Anzahl der **Einträge**, **Status** und verfügbare Aktionen. |
| Einträge | Zeigt Bild, Besitzer und Betriebstage, Lokdaten, Steuerungsdaten, belegte Funktionstasten und Aktionen der ausgewählten Liste. |

Die Auswahl einer Listenzeile lädt deren Einträge in die benachbarte Tabelle. Nach der ersten
Listenanfrage wählt RailKeeper die erste gelieferte Liste aus, falls noch keine Auswahl besteht.
Die stabile Speicherreihenfolge beginnt mit dem neuesten Datum und danach der Bezeichnung; auch der
Browser startet mit **Datum** absteigend.

Wähle eine sortierbare Listenüberschrift, um nach Bezeichnung, Datum, Eintragsanzahl oder Status zu
sortieren. Ein erneuter Klick auf die aktive Überschrift kehrt die Reihenfolge um. Die
Eintragssortierung beginnt mit **Besitzer** aufsteigend. Die genauen Spalten und Regeln beschreibt
[Einträge und Drucken](./entries-and-printing).

## Dem Messeablauf folgen

Diese Reihenfolge hält den Betrieb nachvollziehbar:

1. Ein Admin legt eine Liste mit Bezeichnung und Datum an.
2. Messe oder Admin wählt die offene Liste aus.
3. Das Betriebspersonal ergänzt Einträge und pflegt Besitzer, Lok-, Betriebs-, Bild- und
   Funktionsdaten.
4. Adresskonflikte werden vor dem Speichern eines Eintrags gelöst.
5. Mit **Ansehen** oder **Liste drucken** werden die aktuellen Einträge geprüft.
6. Ein Admin sperrt die Liste, sobald keine Einträge mehr geändert werden sollen.
7. Die gesperrte Liste bleibt les- und druckbar. Für eine gezielte Korrektur kann ein Admin sie
   wieder entsperren.

Listen- und Eintragsaktionen sind getrennte Server-Schreibvorgänge. Das Schließen eines Dialogs
ohne Absenden speichert dessen Formular nicht. Nach dem Anlegen, Bearbeiten oder Löschen einer
Liste lädt RailKeeper die Listentabelle neu; Sperren und Entsperren aktualisieren die betroffene
Zeile direkt. Nach einem erfolgreichen Eintragsvorgang lädt es die Einträge der ausgewählten Liste
neu und aktualisiert die angezeigte Anzahl.

## Mit einer offenen oder gesperrten Liste arbeiten

Die Spalte **Status** zeigt **offen** oder **gesperrt**.

| Vorgang | Offene Liste | Gesperrte Liste |
| --- | --- | --- |
| Auswählen und lesen | Ja | Ja |
| Vollständige Liste ansehen | Ja | Ja |
| Drucken | Ja | Ja |
| Eintrag anlegen oder bearbeiten | Ja | Nein |
| Eintrag als Admin löschen | Ja | Nein |
| Listenfelder als Admin bearbeiten | Ja | Ja |
| Als Admin sperren oder entsperren | Sperren | Entsperren |
| Liste als Admin löschen | Ja | Ja |

Die Bedienelemente zum Anlegen und Bearbeiten von Einträgen sind bei einer gesperrten Liste
deaktiviert. Der Server weist Eintragsänderungen ebenfalls zurück, wenn die Liste erst nach dem
Laden der Seite gesperrt wurde. Entsperre die Liste vor einer Korrektur; ein deaktiviertes
Bedienelement bedeutet nicht, dass Daten fehlen.

Jede Liste bietet **Ansehen** und **Drucken**. Admin sieht zusätzlich **Bearbeiten**, **Sperren**
oder **Entsperren** sowie **Löschen**. Der Eintragsbereich bietet außerdem **Liste drucken** und bei
einer offenen Auswahl das Bedienelement zum Anlegen eines Eintrags.

## Lade-, Leer- und Fehlerzustände lesen

| Zustand | Bedeutung und nächster Schritt |
| --- | --- |
| **Wird geladen...** | Die Listenanfrage läuft noch. Warte, bevor du eine Liste auswählst. |
| **Noch keine Messeliste angelegt.** | Es ist keine Liste vorhanden. Ein Admin muss die erste Liste anlegen. |
| **Bitte eine Liste auswählen.** / **Keine Liste ausgewählt.** | Wähle eine Zeile der Listentabelle, bevor du Einträge bearbeitest. |
| **Noch keine Einträge in dieser Liste.** | Die ausgewählte Liste ist leer. Lege einen Eintrag an, solange sie offen ist. |
| Meldung oberhalb der Tabellen | Eine Listen-, Eintrags-, Symbol- oder Stammdatenanfrage ist fehlgeschlagen. Lies die Meldung, bevor du fortfährst. |

Einige Folgeanfragen entfernen zuvor angezeigte Daten nicht, bevor sie einen Fehler melden. Schlägt
das Neuladen einer Liste oder ihrer Einträge fehl, lade den Arbeitsbereich neu. Prüfe anschließend
Auswahl, Status, Eintragsanzahl und gespeicherte Einträge, bevor du einen Schreibvorgang wiederholst.
So wird ein möglicherweise bereits erfolgreicher Vorgang nach einem fehlgeschlagenen Refresh nicht
versehentlich doppelt ausgeführt.

## Mit einem gezielten Ablauf fortfahren

- [Listen und Sperren](./lists-and-locking) erklärt Vorbereitung, Sperrwirkung und Löschen durch
  Admin.
- [Einträge und Drucken](./entries-and-printing) erklärt alle Eintragsfelder, Adresskonflikte,
  Bilder, Funktionstasten, Ansicht und Reportdruck.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Listen und Sperren](./lists-and-locking)
- [Einträge und Drucken](./entries-and-printing)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)

## Dokumentierte RailKeeper-Version

Diese Seite beschreibt RailKeeper v0.1.17.6. Der Entwicklungsstand auf `main` kann abweichen und
gehört nicht zu diesem Benutzerablauf.
