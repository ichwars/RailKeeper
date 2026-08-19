---
title: Persönliche Einstellungen
description: Sprache, Startseite, Datum, Zeit, Druckausgabe und Seitenleisten-Reihenfolge konfigurieren.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Persönliche Einstellungen

Die persönlichen Vorgaben befinden sich unter **Einstellungen > Allgemein**. Änderungen wirken
sofort; eine getrennte Speichern-Schaltfläche gibt es nicht.

## Sprache

**Deutsch** oder **English** wählen. RailKeeper ändert sofort die Oberflächentexte und die vom
Browser verwendete Dokumentensprache.

Die Sprachwahl wird in RailKeeper v0.1.19.2 nur im aktuellen Browser gespeichert. Sie folgt dem
Benutzerprofil nicht in einen anderen Browser oder auf ein anderes Gerät. Ein Browser ohne
gespeicherte Auswahl startet auf Deutsch.

## Standardansicht

Die Standardansicht wird verwendet, wenn RailKeeper über seine Stammadresse ohne genaueren Pfad
startet. Zur Auswahl stehen Übersicht, Fahrzeugbestand, Zubehör, Anlage, Ausstellung,
Import/Export und Einstellungen.

Dabei gelten folgende Grenzen:

- Ein direkter Link wie `/vehicles` oder `/settings` hat Vorrang vor der Standardansicht.
- Direkt nach der Anmeldung leitet RailKeeper das Konto zunächst in seinen ersten zulässigen
  Arbeitsbereich. Die Standardansicht umgeht diese rollenabhängige Entscheidung nicht.
- Darf die aktuelle Rolle ein gespeichertes Ziel nicht öffnen, verwendet RailKeeper die erste
  zulässige Ansicht.
- **Anlage** ist in v0.1.19.2 auswählbar, bleibt aber ein unveröffentlichter Entwicklungsbereich und
  wird in der normalen Seitenleiste nicht angezeigt. Für den stabilen Benutzerablauf eine andere
  Startseite wählen.

## Datums- und Zeitformat

Beim Datum stehen **Systemstandard**, ein deutsches Tag-Monat-Jahr-Format und das ISO-Format
Jahr-Monat-Tag zur Auswahl. Für die Zeit gibt es **Systemstandard**, 24 Stunden und 12 Stunden.

RailKeeper speichert und synchronisiert beide Vorgaben. In v0.1.19.2 sind sie noch nicht mit allen
Datums- und Zeitausgaben der Anwendung verbunden. Viele Ansichten formatieren Werte weiterhin nach
der gewählten Oberflächensprache oder mit einem eigenen festen Formatierer. Die Auswahl ist daher
eine vorbereitete Vorgabe und noch keine verlässliche globale Überschreibung.

## Standarddrucker

Die Auswahl enthält:

- **Systemdialog / Standarddrucker**;
- einen benannten Drucker, wenn ein Admin konfigurierte oder vom Betriebssystem gemeldete Drucker
  lesen kann;
- **Jedes Mal fragen**;
- **Als PDF speichern**.

Kann RailKeeper bei aktiver Auswahl **Systemdialog / Standarddrucker** einen Systemstandard
ermitteln, speichert es den gefundenen Druckernamen als Vorgabe. Die Druckererkennung selbst ist
nur für Admins zugänglich. Andere Rollen können die allgemeine Druckvorgabe weiterhin behalten und
ändern.

RailKeeper v0.1.19.2 speichert diesen Wert, leitet aber noch nicht jeden Druckvorgang automatisch an
das gewählte Ziel. Fahrzeugbestandsdruck, Messeberichte und die Druckausgabe unter Import/Export
öffnen weiterhin den Browser-Druckdialog. Browser und Betriebssystem bestimmen dort den
tatsächlichen Drucker oder das Ziel **Als PDF speichern**.

## Reihenfolge und Sichtbarkeit der Seitenleiste

Die **Seitenleisten-Reihenfolge** steuert die Hauptnavigation des aktuellen Benutzers:

1. Mit den Pfeilen einen Eintrag nach oben oder unten verschieben.
2. Mit dem Augen-Symbol einen Eintrag aus- oder einblenden.
3. **Zurücksetzen** wählen, um Standardreihenfolge und alle Einträge wiederherzustellen.

**Einstellungen** kann nicht ausgeblendet werden, damit die Konfigurationsseite erreichbar bleibt.
Das Ausblenden entfernt nur den Link aus der Seitenleiste. Mit der erforderlichen Rolle lässt sich
die zugrunde liegende Seite weiterhin direkt öffnen. Für die aktuelle Rolle unzulässige Einträge
bleiben unabhängig von ihrer gespeicherten Position herausgefiltert.

Die Liste kann **Anlage** enthalten, obwohl RailKeeper v0.1.19.2 diesen Arbeitsbereich nicht in der
normalen Seitenleiste veröffentlicht. Verschieben oder Einblenden macht ihn nicht zu einer
veröffentlichten Benutzerfunktion.

Reihenfolge und ausgeblendete Einträge werden für jeden Benutzernamen getrennt gespeichert und über
dessen Profil synchronisiert. Der kleine Pfeil am unteren Ende der Seitenleiste klappt sie nur im
aktuellen Browser ein oder aus; dieser Zustand gehört nicht zu den Profileinstellungen.

## Fehlerbehebung

| Symptom | Prüfen |
| --- | --- |
| Eine Vorgabe änderte sich lokal, aber nicht in einem anderen Browser | Einstellungen mit demselben Benutzer neu laden. Der Hintergrundzugriff auf das Profil kann gescheitert sein, obwohl der lokale Wert aktiv blieb. |
| Die Oberflächensprache unterscheidet sich auf einem anderen Gerät | Dort erneut einstellen. Die Sprache ist in v0.1.19.2 browserlokal. |
| Die gewählte Startseite öffnet sich nach der Anmeldung nicht | Rollenabhängige Anmeldung und direkte URLs haben Vorrang. Zum Testen der Vorgabe die Stammadresse öffnen. |
| Eine ausgeblendete Seite lässt sich weiterhin öffnen | Ausblenden ändert nur die Navigation, nicht Berechtigung oder Routing. |
| Es erscheinen keine benannten Drucker | Die Systemdrucker-Erkennung erfordert Admin und hängt vom Server-Betriebssystem oder der Druckerkonfiguration ab. |
| Gewählter Drucker oder Datumsformat zeigen keine Wirkung | Die Vorgaben werden gespeichert, aber in v0.1.19.2 noch nicht von jedem Arbeitsablauf global verwendet. |

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.19.2** und wurde zuletzt am 16.08.2026 geprüft.
