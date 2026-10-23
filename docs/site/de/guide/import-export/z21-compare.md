---
title: Bekannte Adressen mit Z21 vergleichen
description: Bekannte Adressen mit Z21 vergleichen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Bekannte Adressen mit Z21 vergleichen

**Ziel:** Bereits bekannte RailKeeper-Adressen lesend mit einer Z21 beziehungsweise z21 abgleichen.
Du brauchst **Admin**, vorhandene digitale Fahrzeuge mit Decoderadresse und eine erreichbare Zentrale.

<GuideFigure name="central-workspace" alt="Arbeitsbereich Digitalzentralen in der getrennten Demo ohne verbundene Zentrale." caption="Der echte Einstieg in den Arbeitsbereich. Dieses Bild zeigt keinen realen Z21-Abruf; die Hardwareabnahme ist offen." />

## Verbindung und Ausgangsbestand prüfen

1. Öffne **Einstellungen > Digitalzentralen**, wähle Z21 und prüfe Host sowie UDP-Port, normalerweise **21105**.
2. Speichere, teste die Verbindung und lies bei Bedarf die Diagnose.
3. Prüfe im Fahrzeugbestand die gespeicherten Decoderadressen. Der Leseablauf erfindet keine unbekannten Adressen.
4. Aktiviere den passenden Anschluss und öffne **Digitalzentralen**.

## Daten lesen und die begrenzte Aussage verstehen

Wähle **Daten lesen**. RailKeeper fragt die bereits bekannten, eindeutigen Adressen ab und übernimmt
nur die bestätigte Adresse in die Vergleichsarbeitsliste. Namens- und Protokollangaben fehlen in dieser
Antwort und bleiben als unvollständig gekennzeichnet. Geschwindigkeit, Fahrtrichtung und aktive
Funktionstastenzustände werden nicht in den Bestand übernommen.

Dies ist keine Abfrage der vollständigen Zentralen-Lokdatenbank. Die Antwort enthält weder
Loknamen noch gespeicherte CV-Werte oder Lokbilder. Der Ablauf unterstützt höchstens **64 verschiedene
bekannte Adressen**; bei mehr Adressen wird abgewiesen statt unbemerkt nur ein Teil abgefragt.

## Ergebnis prüfen

Vergleiche Quelladresse, vorhandenes Fahrzeug und Konfliktmeldung. Bei mehreren passenden Fahrzeugen
kläre die Identität selbst. Fehlende Namen oder Protokolle dürfen nicht als Löschvorschlag für deine
vorhandenen Fahrzeugangaben verstanden werden. Prüfe, welche Übernahmeaktion die aktuelle
Arbeitsliste tatsächlich anbietet.

| Beobachtung | Nächster Schritt |
| --- | --- |
| Kein sinnvoller Ausgangsbestand | Decoderadressen an den vorgesehenen RailKeeper-Fahrzeugen prüfen |
| Zeitüberschreitung | Erreichbarkeit vom Server und UDP-Port prüfen |
| Name oder Protokoll unbekannt | Grenze dieser Antwort; keinen Wert raten |
| Keine CVs | CV-Datei beim Fahrzeug importieren; dieser Leseweg liefert keine CV-Liste |
| Intellibox 3 ausgewählt | Dort sind nur Verbindungstest und Diagnose freigeschaltet |

Die Leseabläufe sind protokollbasiert automatisiert geprüft. Reale Modell-/Firmwareprüfungen stehen
noch aus. [Kompatibilität und Nachweise](../../reference/command-station-compatibility).
