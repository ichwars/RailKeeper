---
title: Fehler melden
description: Direkt aus RailKeeper ein GitHub-Issue erstellen und die enthaltenen Diagnosedaten verstehen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-02
---

# Fehler melden

<GuideFigure name="preferences" alt="Die globale Fehleraktion ist auch im Arbeitsbereich Einstellungen erreichbar." caption="Einstieg über die globale Fehleraktion. Der gezeigte Basisbildschirm enthält künstliche Demodaten; das Meldeformular ist unten beschrieben." />

Wenn der Betreiber In-App-Fehlerberichte eingerichtet hat, können alle angemeldeten
RailKeeper-Rollen die schwebende Schaltfläche **Fehler melden** öffnen. Das Formular erstellt ein
beschriftetes Issue im RailKeeper-GitHub-Repository. Eine Anmeldung bei GitHub im Browser ist dafür
nicht erforderlich.

Ist die Funktion nicht verfügbar oder ist ihre Startprüfung fehlgeschlagen, öffnet dieselbe
Schaltfläche stattdessen die normale GitHub-Issue-Seite. Formulareingaben werden niemals unbemerkt
an einen anderen Dienst umgeleitet.

## Einen hilfreichen Bericht senden

1. Die Seite öffnen, auf der das Problem aufgetreten ist, und **Fehler melden** wählen.
2. Unter **Wo tritt der Fehler auf?** den betroffenen Teil von RailKeeper wählen. Die aktuelle Seite
   ist vorausgewählt. Bei Unsicherheit **Sonstiges** wählen.
3. Das Problem mit mindestens 20 Zeichen beschreiben.
4. Optional den erwarteten Zustand und Reproduktionsschritte ergänzen.
5. Eine E-Mail-Adresse nur angeben, wenn der Maintainer Kontakt aufnehmen darf. Das Feld erscheint
   nur, wenn der Betreiber die Kontaktverschlüsselung eingerichtet hat.
6. Optional einen PNG-, JPEG- oder statischen WebP-Screenshot auswählen, einfügen oder ablegen. Das
   Bild prüfen und seiner öffentlichen Speicherung zustimmen.
7. **Welche Daten werden im Bericht gesendet?** öffnen und die aktuelle Angabe prüfen.
8. Den Bericht senden und den angezeigten Issue-Link aufbewahren.

Nach einem behebbaren Fehler bleiben die Texte im Formular. Existiert das Issue bereits, während
Screenshot oder Kommentar noch unvollständig sind, setzt **Erneut versuchen** denselben Bericht und
dasselbe GitHub-Issue fort. Es wird nicht absichtlich ein Duplikat erzeugt.

## Automatische Diagnose abschalten

**Technische Diagnosedaten automatisch beifügen** ist im Formular abschaltbar. Die Auswahl bleibt
in diesem Browser gespeichert. Ausschalten leert den lokalen Fehlerpuffer, stoppt die weitere
automatische Fehlererfassung und entfernt Browser- und Serverdiagnose aus dem Bericht. Selbst
mitgesendete Diagnosefelder verwirft der Server bei ausgeschalteter Option.

Bei eingeschalteter Option zeigt die aufklappbare Vorschau den aktuellen Diagnoseumfang.
Laufzeit- und Speicherwerte sind Momentaufnahmen; beim Absenden erfasst der Server sie frisch.
Es werden keine Daten im Hintergrund zu GitHub hochgeladen.

## Enthaltene Diagnosedaten

RailKeeper sendet einen begrenzten technischen Zustand zum Zeitpunkt des Absendens:

- RailKeeper- und Go-Laufzeitversion, Bereitstellungsart, Datenbankschema-Version, Betriebssystem
  und Architektur
- Erfassungszeit, Prozesslaufzeit, CPU-Anzahl, Goroutinen und Go-Speicherbedarf
- SQLite-Version, Datenbankgröße und Journalmodus
- Browserfamilie und Hauptversion, Viewport, Darstellungsmodus, Sprache und Zeitzone
- aktueller RailKeeper-Pfad und Funktionsbereich
- Rollennamen, nicht den Benutzernamen
- Typnamen aktivierter Digitalzentralen, zum Beispiel `ecos` oder `z21`
- höchstens 30 aktuelle bereinigte Browserfehler mit abstrakten API-Routen, Status und sicherer
  Fehlerklasse

## Ausdrücklich ausgeschlossene Daten

Die automatische Diagnose enthält nicht:

- Fahrzeug-, Zubehör-, Anlagen- oder Ausstellungsdatensätze
- Bestandszahlen oder andere Bestandsstatistiken
- Gerätezahlen, Zustände, Adressen, Hosts, Ports oder Verbindungsstatistiken verbundener Geräte
- Druckernamen, Seriennummern, Zugangsdaten, Passwörter, Tokens, Cookies, API-Schlüssel oder
  Zugangscodes
- Anfrage- oder Antwortinhalte, Webhook-URLs, lokale Pfade, Hostnamen, Benutzernamen oder
  E-Mail-Adressen

Optional selbst eingegebene Texte, E-Mail-Adresse und Screenshot gehören nicht zur automatischen
Diagnose. Freitext und Bilder deshalb vor dem Senden prüfen.

## Screenshots und Kontaktdaten

Das Screenshotfeld erscheint nur, wenn ein separates öffentliches Asset-Repository eingerichtet ist.
Screenshots sind optional und auf 5 MiB begrenzt. RailKeeper akzeptiert PNG, JPEG und statisches
WebP, prüft die Abmessungen, dekodiert das Bild und kodiert es neu. Dadurch werden eingebettete
Metadaten entfernt. Sichtbare Informationen in den Bildpixeln kann diese Bereinigung nicht
verbergen. Das bereinigte Bild wird erst nach ausdrücklicher Zustimmung in einem öffentlichen
GitHub-Repository gespeichert.

Wenn verfügbar, wird die optionale E-Mail-Adresse an den eigenen RailKeeper-Server gesendet und dort
für den vom Betreiber konfigurierten age-Empfänger verschlüsselt, bevor sie in das Issue gelangt.
GitHub erhält nur den ASCII-armierten Geheimtext, nicht die lesbare Adresse. Entfernte
Bereitstellungen sollten HTTPS verwenden, damit die Adresse auch zwischen Browser und RailKeeper
geschützt ist.

Keine Screenshots anhängen, die Passwörter, Tokens, private Dokumente, persönliche Nachrichten oder
andere nicht zur Veröffentlichung bestimmte Inhalte zeigen.

## Grenzen und Fehlerfälle

Jeder Benutzer kann höchstens fünf neue Bericht-IDs pro Stunde starten. Ein erneuter Versuch für
eine bestehende Bericht-ID verbraucht keinen weiteren Platz. GitHub-Ausfälle, widerrufene Tokens,
fehlende Labels oder ein nicht verfügbares öffentliches Asset-Repository können das Senden
verhindern. Die Texte bleiben im Formular. Bei deaktivierter In-App-Funktion bleibt die externe
GitHub-Issue-Seite als Ausweichweg erhalten.
