---
title: Stammdaten-Administration
description: Stammdaten, Inventarnummern, Lagerorte und Transfer in RailKeeper sicher verwalten.
audience: admin
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-16
---

# Stammdaten-Administration

RailKeeper verwendet kontrollierte Stammdaten überall dort, wo Datensätze gemeinsame Namen,
Klassifizierungen, Symbole oder Nummerierungsregeln benötigen. Der größte Teil wird unter
**Einstellungen > Daten** verwaltet. Inventarnummernschemata stehen unter
**Einstellungen > Allgemein**, der JSON-Stammdatentransfer unter
**Einstellungen > Import/Export**.

Dieser Bereich dokumentiert RailKeeper v0.1.20.2 und trennt vier administrative Abläufe:

- [Allgemeine Stammdaten](./master-data-general) für Fahrzeuge, Hersteller, CV8-Kennungen und
  Funktionssymbole.
- [Artikelstammdaten und Lagerorte](./master-data-articles) für Zubehörkatalog und Bestandsstruktur.
- [Inventarnummernschemata](./master-data-inventory-numbers) für automatische Fahrzeug- und
  Artikelkennungen.
- [Stammdatentransfer](./master-data-transfer) für vollständigen JSON-Export und Abgleich.

## Zugriffsrechte

| Aktion | Admin | Editor | Viewer | Planner | Messe |
| --- | --- | --- | --- | --- | --- |
| Einstellungen öffnen und Stammdaten lesen | Ja | Ja | Ja | Ja | Nein |
| Einträge anlegen, bearbeiten, deaktivieren oder reaktivieren | Ja | Ja | Nein | Nein | Nein |
| Geeigneten eigenen Eintrag endgültig löschen | Ja | Ja | Nein | Nein | Nein |
| Lagerorte verwalten | Ja | Ja | Nein | Nein | Nein |
| Inventarnummernschemata verwalten | Ja | Ja | Nein | Nein | Nein |
| Stammdatendokument exportieren oder importieren | Ja | Nein | Nein | Nein | Nein |

Viewer und Planner erben den Lesezugriff. Die Artikelverwaltung zeigt ausdrücklich einen
Nur-Lese-Zustand. Einige Bedienelemente für allgemeine Stammdaten und Inventarnummern können für
diese Rollen in v0.1.20.2 trotzdem sichtbar bleiben, der Server weist aber jeden Schreibvorgang ab.
Ein sichtbares Bedienelement erteilt niemals eine Berechtigung.

Ein reines Messe-Konto kann die Einstellungen nicht öffnen. Es kann nur die aktiven
Funktionssymbole über den isolierten Messeablauf lesen.

## Herkunft und Lebenszyklus verstehen

Jeder kontrollierte Eintrag besitzt eine Herkunft:

| Herkunft | Bedeutung | Endgültiges Löschen |
| --- | --- | --- |
| **Mitgeliefert** | Wird von RailKeeper ausgeliefert und abgeglichen | Niemals erlaubt |
| **Eigen** | Lokal angelegt oder ohne passende mitgelieferte Kennung importiert | Nur ungenutzt erlaubt |

Einträge beider Herkünfte lassen sich bearbeiten, deaktivieren und reaktivieren. Eine Deaktivierung
entfernt einen Eintrag aus neuen Auswahlen. Bestehende Datensätze behalten ihren gespeicherten Wert
und zeigen ihn weiterhin als inaktiv an. So wird ein historisch noch benötigter Wert sicher außer
Betrieb genommen.

Das endgültige Löschen ist absichtlich enger gefasst. RailKeeper bietet es nur für einen eigenen
Eintrag an, dessen Datentyp eine bekannte Verwendungsprüfung besitzt und dessen Schlüssel oder
Bezeichnung nicht referenziert wird. Die Standardschlüssel der Artikelarten sind zusätzlich
geschützt. Beim Löschen eines geeigneten Eintrags werden auch seine Stammdatenbeziehungen entfernt.
Der Vorgang lässt sich nicht rückgängig machen.

## Konservativer Änderungsablauf

1. Vor umfangreichen Änderungen eine aktuelle Anwendungssicherung erstellen und prüfen.
2. Bei vielen Klassifizierungsänderungen zusätzlich das aktuelle Stammdaten-JSON exportieren.
3. Einen Schreibfehler bevorzugt korrigieren, statt einen nahezu gleichen Eintrag anzulegen.
4. Einen veralteten Wert zuerst deaktivieren und betroffene Fahrzeug-, Artikel- und Messeansichten
   prüfen.
5. Nur einen bestätigt ungenutzten eigenen Eintrag löschen.
6. Vor weiteren Massenänderungen ein betroffenes Fahrzeug oder einen Artikel testweise anlegen.

Inventarnummernschemata und Lagerorte gehören nicht zum Stammdaten-JSON. Wenn diese Einstellungen
zusammen mit dem Bestand wiederherstellbar sein müssen, ist eine Anwendungssicherung erforderlich.

## Verwandte Seiten

- [Einstellungen und Berechtigungen](/de/guide/settings/)
- [Zubehörübersicht](/de/guide/accessories/)
- [Fahrzeugbestand](/de/guide/vehicles/)
- [Installation und Administration](./)

## Dokumentierte RailKeeper-Version

Dieser Bereich dokumentiert RailKeeper **v0.1.20.2** und wurde zuletzt am 16.08.2026 geprüft.
