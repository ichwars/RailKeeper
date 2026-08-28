---
title: CS3-Lokdaten read-only lesen
description: Märklin-CS3-Lokomotiven sicher lesen und im Digitalzentralen-Arbeitsbereich vergleichen.
audience: user
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-28
---

# CS3-Lokdaten read-only lesen

RailKeeper liest die Lokliste einer aktiven Märklin CS3 oder CS3 Plus per HTTP in den bestehenden
Arbeitsbereich **Digitalzentralen**. Der Ablauf erstellt eine dauerhafte Vergleichsvorschau. Er
schreibt nichts zur CS3 und steuert keine Lokomotive.

Alle CS3-Routen sind ausschließlich für Admin verfügbar. Host und Port stammen aus der
Serverkonfiguration unter **Einstellungen > Digitalzentralen**. Werte aus dem Leseaufruf können den
gespeicherten Zielhost nicht überschreiben.

## Unterstützte CS3-API-Generationen

| CS3-Firmwaregeneration | Read-only Endpunkt | Verhalten |
| --- | --- | --- |
| ab 2.6 | `/app/api/locos` | Wird immer zuerst geprüft und verwendet. |
| vor 2.6 | `/app/api/loks` | Wird nur verwendet, wenn der aktuelle Endpunkt eindeutig mit HTTP 404 antwortet. |

Andere Fehler lösen keinen stillen Fallback aus. Eine Anmeldung, Weiterleitung, HTML-Seite oder
unbekannte Antwort gilt nicht als erfolgreiche CS3-Verbindung.

Märklin beschreibt CS3 und CS3 Plus als lokale Zentralen mit Lokdatenbank, veröffentlicht jedoch
keinen stabilen Vertrag für diese Web-App-Endpunkte. Die Endpunktformen und Feldnamen wurden aus
öffentlich dokumentierten TrainControl-Kompatibilitätsdaten abgeleitet. Die RailKeeper-Fixtures
`cs3_loks_pre_2_6_anonymized.json` und `cs3_locos_2_6_anonymized.json` bilden diese Antwortformen
anonymisiert nach. Sie enthalten keine privaten Anlagen- oder Lokdaten. Für diese Fixtures liegt
keine direkte Hardware-Verifikation durch das RailKeeper-Projekt vor.

## Gelesene und bewusst ausgelassene Daten

RailKeeper übernimmt pro Lok nur:

- `uid` als stabile externe CS3-ID;
- `name` oder ersatzweise den dekodierten `internname`;
- `address` als Decoderadresse;
- `dectyp` als normalisiertes Protokoll, zum Beispiel MFX, Motorola oder DCC.

RailKeeper ignoriert Geschwindigkeit, Richtung, aktive Funktionen, Icons, CVs und Anlagenobjekte.
Es startet kein Live-Monitoring und sendet keine Steuer- oder Schreibbefehle. Eine Decoderadresse
ist nur ein Vergleichsmerkmal. Namen allein erzeugen niemals einen automatischen Treffer.

## Sicher lesen und vergleichen

1. Unter **Einstellungen > Digitalzentralen** die CS3 mit Host und HTTP-Port konfigurieren.
2. **Verbindung testen** ausführen. Erfolg setzt eine kompatible, gültige JSON-Lokliste voraus.
3. Unter **Letzte Diagnose** optional **Diagnosedaten lesen** wählen. RailKeeper zeigt API-Pfad,
   Generation, HTTP-Status, Content-Type, Anzahl und den Hinweis `readOnly`.
4. Den Adapter aktivieren und den Arbeitsbereich **Digitalzentralen** öffnen.
5. **Daten lesen** wählen. RailKeeper legt eine neue Read-Session und Vergleichsarbeitsliste an.
6. Neue, fehlende, abweichende und mehrdeutige Einträge prüfen. Mehrere mögliche Adresstreffer
   bleiben als sichtbarer Konflikt bestehen.

Der Abruf ist auf HTTP GET, 8 MiB Antwortgröße und 5.000 Lokomotiven begrenzt. Weiterleitungen
werden abgelehnt. Nur JSON-Content-Types und vollständig validierte UIDs, Namen, Adressen und
Protokolle gelangen in die Vorschau.

## Fehlerbehebung

| Diagnose | Prüfen |
| --- | --- |
| Netzwerk- oder Timeoutfehler | CS3-IP, Port, lokales Netz und Erreichbarkeit vom RailKeeper-Server prüfen. |
| Authentifizierungsfehler | Zugriffsschutz der CS3-Webanwendung prüfen. RailKeeper umgeht keine Anmeldung. |
| Weiterleitung abgelehnt | Den direkten lokalen CS3-Host konfigurieren. RailKeeper folgt keiner Weiterleitung. |
| Kein JSON oder HTML erhalten | Prüfen, ob Host und Port die CS3-API statt einer Login- oder Proxyseite erreichen. |
| Keine unterstützte Loklisten-API | Firmware und Web-App-Verfügbarkeit prüfen. Beide bekannten Endpunkte lieferten 404. |
| Ungültige Lokdaten | UID, Name, Decoderadresse oder Protokoll einer Lok liegt außerhalb der sicheren Grenzen. Die gesamte Antwort wird verworfen. |
| Lok erscheint als Konflikt | Mehrere RailKeeper-Fahrzeuge passen zu Adresse und Protokoll. Zuordnung manuell prüfen. |

## Dokumentierte RailKeeper-Version

Dieses Kapitel dokumentiert RailKeeper **v0.1.20.2** und wurde zuletzt am 28.08.2026 geprüft.
