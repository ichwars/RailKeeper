---
title: Zubehör und Wartung in der Werkstatt üben
description: Zubehör und Wartung in der Werkstatt üben.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Zubehör und Wartung in der Werkstatt üben

**Ziel:** Verfolgen, wie Reservierung, Einbau und Ausbau den Bestand verändern, und anschließend
eine Fahrzeugwartung abschließen. Verwende **Admin** oder **Editor** und ausschließlich Übungsdaten.

## 1. Vier Stück als Ausgangsbestand

Lege einen Mengenartikel **Demo-Beleuchtungsplatine** mit Einheit Stück an. Erstelle unter den
Stammdaten einen aktiven Lagerort **Demo-Schublade**. Buche dort vier Stück als Anfangsbestand.
Ein vorhandenes Testfahrzeug dient als Zuordnungsziel.
[Artikel anlegen](article-records) · [Mengenbestand buchen](stock-purchases-documents).

Die Bildfolge zeigt echte lokale Demo-Buchungen. Die Zwischenstände wurden über die lokalen
Anwendungs-APIs hergestellt und danach in der tatsächlichen Oberfläche geprüft; sie enthält keine
Buchung an einer Digitalzentrale.

## 2. Ein Stück reservieren

Öffne den Artikel zum Bearbeiten und **Verwendung & Historie**. Wähle Fahrzeug, Lagerort und
Menge **1**. Optional beschreibe Platzierung und Anschluss. Wähle **Reservierung anlegen** und bestätige.

<GuideFigure name="stock-reserved" alt="Aktive Reservierung eines Stücks für das Demo-Fahrzeug." caption="Reservierung vorhanden: Vier Stück bleiben im Lager, eines ist reserviert und drei sind verfügbar." />

Eine Reservierung hält Material für ein Fahrzeug frei. Sie dokumentiert noch keinen tatsächlichen Einbau.

## 3. Die Reservierung einbauen

Wähle die aktive Reservierung im Einbauformular. Prüfe das übernommene Ziel und die Quelle.
Wähle den Zustand, dann **Einbau erfassen**, und bestätige.

<GuideFigure name="stock-installed" alt="Aktiver Einbau des zuvor reservierten Zubehörstücks." caption="Die Reservierung ist erfüllt. Drei Stück liegen noch im Lager und eines ist eingebaut." />

## 4. Wieder ins Lager ausbauen

Wähle am aktiven Einbau **Ausbauen**, als Ziel **Eingelagert** und den aktiven Lagerort.
Prüfe und bestätige den Vorgang.

<GuideFigure name="stock-returned" alt="Verwendungshistorie nach dem Ausbau zurück in den Lagerbestand." caption="Der Verlauf bleibt erhalten. Alle vier Stück liegen wieder im Lager und sind verfügbar." />

| Stand | Eingelagert | Reserviert | Eingebaut | Verfügbar |
| --- | --- | --- | --- | --- |
| Anfang | 4 | 0 | 0 | 4 |
| Reserviert | 4 | 1 | 0 | 3 |
| Eingebaut | 3 | 0 | 1 | 3 |
| Zurück im Lager | 4 | 0 | 0 | 4 |

Diese Tabelle gilt für den Mengenartikel und die gewählte Disposition **Eingelagert**.
Einzelstücke und Ausbau nach Defekt, Wartung oder Außer Betrieb haben eigene Lebenszyklen.
[Details und Fehlerbehandlung](allocations-history).

## 5. Fahrzeugwartung abschließen

Öffne das Testfahrzeug unter **Bearbeiten > Wartung**. Lege beispielsweise eine Reinigung mit
Status **geplant** und einem vergangenen Fälligkeitsdatum an. Speichere den Wartungseintrag.

<GuideFigure name="maintenance-open" alt="Offene, bereits fällige Reinigung des Demo-Fahrzeugs." caption="Die Wartung ist gespeichert und noch nicht erledigt. Andere vorhandene Wartungen können die Gesamtsummen zusätzlich beeinflussen." />

Führe die Arbeit aus und wähle an diesem Eintrag **Erledigt**.

<GuideFigure name="maintenance-done" alt="Wartung nach dem tatsächlichen Abschluss als erledigt." caption="Der Eintrag besitzt den Status erledigt und einen Abschlusszeitpunkt. Die Aktion speichert direkt." />

Öffne das Fahrzeug erneut und kontrolliere Status und Datum. Wartungsaktionen speichern unabhängig
von der allgemeinen Fahrzeug-Schaltfläche und können das Fahrzeug neu laden. Speichere oder verwirf
andere Entwürfe vorher bewusst. [Wartungsreferenz](../vehicles/maintenance).
