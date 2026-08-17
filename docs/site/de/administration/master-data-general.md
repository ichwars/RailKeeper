---
title: Allgemeine Stammdaten
description: Fahrzeugklassifizierungen, Hersteller, CV8-Kennungen und Funktionssymbole verwalten.
audience: admin
status: stable
reviewedVersion: 0.1.19.1
lastReviewed: 2026-08-16
---

# Allgemeine Stammdaten

Unter **Einstellungen > Daten > Allgemein** werden die gemeinsamen Auswahlen für Fahrzeuge,
Zubehör, Decoderdaten und Messen verwaltet. RailKeeper v0.1.19.1 stellt acht Datentypen bereit.

## Datentypen

| Datentyp | Verwendung | Spezifische Felder |
| --- | --- | --- |
| Hersteller | Herstellerwahl für Fahrzeuge und Artikel, Zuordnung von Suchquellen | Name, Nenngrößen, Website, Suchdomains, Aliasse, Quell-URL |
| Fahrzeugkategorien | Grobe Fahrzeugklassifizierung und Wahl des Inventarnummernschemas | Name, Quell-URL |
| Fahrzeuggattungen | Genauere Fahrzeugart oder Gattung | Name, Quell-URL |
| Epochen | Epochenwahl für Fahrzeuge und Messeinträge | Name, Quell-URL |
| Spurweiten | Spurweitenwahl für Fahrzeuge und Artikel | Name, Quell-URL |
| Bahngesellschaften | Betreiberwahl für Fahrzeuge und Messeinträge | Name, Quell-URL |
| CV8-Hersteller | Decoderhersteller-Kennung aus CV8 | Name, Dezimal, Binär, Hexadezimal, Land |
| Funktionssymbole | Bilder und Beschreibungen für Fahrzeugfunktionen | Name, Beschreibung, Bild |

Jeder Eintrag besitzt außerdem Aktivstatus, Sortierreihenfolge, Herkunft, unveränderlichen Schlüssel
und Zeitstempel im Speicher. Das Formular zeigt die zum gewählten Typ passenden Felder. Der
Schlüssel eines neuen gewöhnlichen Eintrags wird aus der Bezeichnung erzeugt. Eine spätere
Umbenennung ändert diesen Schlüssel nicht.

## Einträge suchen, sortieren und prüfen

Nach Auswahl eines Datentyps durchsucht das lokale Suchfeld den angezeigten Namen, Schlüssel,
Quellwerte und die typabhängigen Hersteller-, CV8- oder Symbolangaben. Fahrzeug- oder
Artikeldatensätze werden dabei nicht durchsucht.

Die Tabelle ist zunächst nach Name sortiert. Mit einer sortierbaren Spalte wird der Sortierschlüssel
gewechselt, erneutes Auswählen kehrt die Reihenfolge um. Aktive und inaktive Einträge, ihre Herkunft
**Mitgeliefert** oder **Eigen** und nur die erlaubten Lebenszyklusaktionen bleiben sichtbar.

Nach einer parallelen Änderung durch einen anderen Administrator die Aktualisierungsaktion
verwenden. Beim Wechsel des Datentyps werden aktueller Entwurf, Suche und Sortierung zurückgesetzt.

## Hersteller verwalten

Der Herstellereintrag steuert sowohl Auswahlnamen als auch Hinweise für die Artikelsuche:

- **Nenngrößen** ordnet die vom Hersteller angebotenen Maßstäbe zu.
- **Website** enthält die maßgebliche Herstellerseite.
- **Suchdomains** bevorzugt passende Quellen in der Artikelsuche. Ist die Liste leer und die
  Website gültig, leitet das Formular deren Domain automatisch ab.
- **Aliasse** enthalten alternative Schreibweisen für die Zuordnung.
- **Quell-URL** dokumentiert die Herkunft des Stammdateneintrags selbst.

Mehrere Nenngrößen, Domains oder Aliasse wie im jeweiligen Feld angegeben trennen. In die
bevorzugten Domains gehören nur Hersteller- oder Katalogquellen. Händler-, Marktplatz-,
Weiterleitungs- und SEO-Domains verschlechtern die Suchqualität. Eine Warnung in der Tabelle
kennzeichnet Herstellereinträge, deren Website noch geprüft werden muss.

Das Ändern einer Herstellerbezeichnung schreibt den bereits in Fahrzeugen, Zubehörartikeln oder
Messeinträgen gespeicherten Freitext nicht um. Bei einer Korrektur eines verwendeten Namens müssen
diese Datensätze geprüft werden.

## CV8-Hersteller verwalten

CV8 kennzeichnet den Decoderhersteller mit einem Dezimalwert von 0 bis 255. Aus der Dezimalzahl
leitet RailKeeper die binäre und hexadezimale Darstellung ab. Beide bleiben sichtbar und werden vor
dem Speichern normalisiert. Der optionale Länderwert wird in Großbuchstaben umgewandelt und ist auf
acht Zeichen begrenzt.

Neue Einträge erhalten einen unveränderlichen Schlüssel der Form `cv8-NNN`, beispielsweise
`cv8-151`. Im angezeigten Namen wird eine redundante führende Dezimalzahl entfernt. Oberhalb der
Liste verlinkt RailKeeper den offiziellen NMRA-Anhang mit Herstellerkennungen. Diese Quelle vor dem
Anlegen oder Korrigieren verwenden.

## Funktionssymbole verwalten

Ein Symbol besteht aus Name, optionaler Beschreibung und optionalem Bild. Der Upload akzeptiert
SVG, PNG, JPEG oder WebP und weist Dateien über 1 MiB ab. Das Bild liegt in den Metadaten des
Eintrags und ist deshalb im Stammdatenexport und in der Anwendungssicherung enthalten.

Das Entfernen der Vorschau löscht das Bild erst beim Speichern des Eintrags. Bestehende
Fahrzeugfunktionen behalten den Symbolschlüssel. Eine Deaktivierung verhindert neue reguläre
Auswahlen, vorhandene Verweise bleiben lesbar. Der Messearbeitsbereich kann aktive Symbole abrufen,
ohne allgemeinen Zugriff auf die Einstellungen zu erhalten.

## Eintrag anlegen, bearbeiten und außer Betrieb nehmen

1. Benötigten Datentyp wählen und nach einem vorhandenen gleichwertigen Eintrag suchen.
2. Name sowie relevante Quell- oder Spezialwerte eingeben.
3. **Hinzufügen** wählen. RailKeeper erzeugt einen lokalen Eintrag mit Herkunft **Eigen**.
4. Zum Korrigieren die Bearbeitungsaktion wählen, Formular ändern und speichern.
5. Zum Außerbetriebnehmen **Deaktivieren** wählen und bestätigen. Gespeicherte Verwendungen bleiben
   unverändert.
6. Den Eintrag reaktivieren, wenn er wieder für neue Auswahlen angeboten werden soll.
7. Endgültiges Löschen nur verwenden, wenn RailKeeper es für einen ungenutzten eigenen Eintrag
   anbietet.

Mitgelieferte Einträge lassen sich bearbeiten und deaktivieren, aber nicht löschen. Ein verwendeter
eigener Eintrag kann ebenfalls nicht gelöscht werden. Wird ein Schreibvorgang für Viewer oder
Planner abgewiesen, ist ein Admin- oder Editor-Konto erforderlich.

## Verwandte Seiten

- [Stammdaten-Administration](./master-data)
- [Artikelstammdaten und Lagerorte](./master-data-articles)
- [Inventarnummernschemata](./master-data-inventory-numbers)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.19.1** und wurde zuletzt am 16.08.2026 geprüft.
