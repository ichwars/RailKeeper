---
title: Zubehörübersicht
description: Zubehörartikel finden, filtern, prüfen und sicher verwalten.
audience: user
status: stable
reviewedVersion: 0.1.20.2
lastReviewed: 2026-08-16
---

# Zubehörübersicht

**Zubehör** ist RailKeepers Katalog- und Bestandsbereich für Gleise, Signale, Decoder, elektrische
Bauteile, Landschaftsbau, Beleuchtung und weitere Modellbahnartikel. Er verbindet Produktidentität,
gelagerte und gebundene Mengen, Bilder, Lagerorte und Aktionen im Artikellebenszyklus. Dieses
Kapitel beschreibt den stabilen RailKeeper-Stand v0.1.20.2.

Ein Artikel ist der gemeinsame Produktdatensatz. Sein Bestand kann aus austauschbaren Mengen,
einzeln geführten Stücken oder beidem bestehen. Reservierungen und Einbauten binden einen Teil
davon an ein Fahrzeug oder ein anderes konfiguriertes Ziel. Die Übersicht fasst diese Zustände
zusammen, verändert den Bestand aber nicht selbst.

## Zugriffsrechte und Arbeitsbereich

Admin, Editor, Viewer und Planner können **Zubehör** öffnen und Artikeldaten einsehen. Ihre
Schreibrechte unterscheiden sich:

| Rolle | Stabiler Zugriff in der Übersicht |
| --- | --- |
| Viewer | Artikel und ihre gespeicherten Ressourcen ansehen. |
| Planner | Artikel ansehen und später Reservierungen anlegen oder stornieren. Die allgemeine Artikel- und Bestandsbearbeitung bleibt schreibgeschützt. |
| Editor | Artikel anlegen und bearbeiten, archivieren oder wiederherstellen sowie Bestand und Lebenszyklus verwalten. |
| Admin | Besitzt die Editor-Rechte und kann zusätzlich einen vollständig unbenutzten Artikel endgültig löschen. |
| Messe | Erhält allein durch die Rolle Messe keinen Zugriff auf Zubehör. |

Die Seite zeigt Viewern und Plannern einen ausdrücklichen Schreibschutz-Hinweis. Die serverseitigen
Rollenprüfungen sind maßgeblich. Ein sichtbarer Wert oder eine Leseaktion verleiht kein
Schreibrecht.

Kennzahlen und Zeilen stammen aus der lokalen RailKeeper-Datenbank. Jede Änderung an Suche,
Filterung oder Sortierung startet eine neue Serveranfrage. Der Ergebniszähler beschreibt deshalb
die aktuelle gesuchte und gefilterte Liste, während die vier Kennzahlen immer alle nicht
archivierten Artikel zusammenfassen.

## Kennzahlen lesen

Oberhalb der Artikelliste stehen vier Kennzahlenkarten:

| Kennzahl | Stabile Bedeutung und Aktion |
| --- | --- |
| **Artikelbestand** | Anzahl nicht archivierter Artikeldatensätze und unterschiedlicher Artikelarten. Die Auswahl entfernt alle aktiven Filter. |
| **Verfügbar** | Gesamte freie Menge und Anzahl unterschiedlicher Lagerorte mit Bezug zu aktiven Artikeln. Die Auswahl zeigt Artikel mit mehr als null verfügbaren Einheiten. |
| **Gebunden** | Summe aktiver Reservierungen und aktiver Einbauten. Die Auswahl zeigt Artikel mit reservierter oder eingebauter Menge. |
| **Pflegehinweise** | Gesamtzahl unvollständiger Angaben in aktiven Artikeln. Die Karte dient nur der Information und besitzt keine Filteraktion. |

**Pflegehinweise** zählt fehlende Felder, nicht betroffene Artikel. Ein Artikel kann mehrere
Hinweise beitragen. Der stabile Stand prüft fehlenden Hersteller, fehlende Artikelnummer,
Artikelart und Bestandseinheit. Bei Gleis-, Signal-, Decoder-, Elektrik-/Steuerungs-,
Gebäude-/Ausstattungs- und Beleuchtungsartikeln erzeugt auch eine fehlende Spurweite einen Hinweis.
Für Landschaftsverbrauchsmaterial und **Sonstiges** verlangt diese Kennzahl keine Spurweite.

Die Kennzahlen ignorieren die aktuelle Suche und Filterauswahl. Die Größe des sichtbaren
Ergebnisses steht im Ergebniszähler darunter.

## Artikel suchen und filtern

Gib Text unter **Artikel suchen** ein. Der Server sucht ohne Beachtung der Groß-/Kleinschreibung
nach Teilzeichenfolgen in genau fünf Feldern:

- Inventarnummer
- Hersteller
- Artikelnummer
- Name
- EAN

Beschreibung, Schlagwörter, alternative Nummern, Fachangaben, Lagerhinweise, Kaufdaten und
Verwendungshistorie gehören im stabilen Stand nicht zur Freitextsuche.

Zusätzlich stehen diese Filter bereit:

| Filter | Auswahl und Verhalten |
| --- | --- |
| Artikelart | Arten, die in aktiven Artikeln vorkommen. |
| Hersteller | Hersteller, die in aktiven Artikeln vorkommen. |
| Spurweite | Spurweiten, die in aktiven Artikeln vorkommen. |
| Status | Verfügbar, Reserviert, Eingebaut, Wartung fällig, Defekt oder Archiviert. |
| Lagerort | Aktive Lagerorte. Ein Artikel passt, wenn Mengenbestand oder ein Einzelstück diesem Lagerort zugeordnet ist. |

Unterschiedliche Filtergruppen werden mit UND verknüpft. Eine Zeile muss den Suchtext und jede
ausgewählte Gruppe erfüllen. **Gebunden** aus der Kennzahlenkarte ist der einzige kombinierte
Status: Er entspricht **Reserviert** ODER **Eingebaut**. Die normale Liste schließt archivierte
Artikel aus. **Archiviert** ersetzt diesen Standard und zeigt archivierte Zeilen, die auch zu den
übrigen Filtern passen.

Die Statusfilter bedeuten:

| Status | Enthaltene Artikel |
| --- | --- |
| Verfügbar | Nach aktiven Reservierungen sind mehr als null Einheiten frei. |
| Reserviert | Mindestens eine Einheit gehört zu einer aktiven Reservierung. |
| Eingebaut | Mindestens eine Einheit gehört zu einem aktiven Einbau. |
| Wartung fällig | Ein Einzelstück oder aktiver Einbau hat den Zustand **Wartung fällig**. |
| Defekt | Ein Einzelstück oder aktiver Einbau hat den Zustand **Defekt**. |
| Archiviert | Der Artikeldatensatz ist archiviert. |

**Filter zurücksetzen** entfernt Suchtext und alle fünf Filtergruppen. Desktopansicht, sichtbare
Spalten und Sortierung bleiben unverändert.

## Spalten, Ansicht und Sortierung wählen

Auf breiten Bildschirmen kannst du zwischen Tabellen- und Kachelansicht wechseln. RailKeeper
speichert diese Wahl im lokalen Speicher des aktuellen Browsers. Sie gehört nicht zum Benutzerkonto
und folgt dir nicht in einen anderen Browser.

Bis zu einer Breite von 900 Pixeln zeigt RailKeeper immer die kompakte Liste. Die gespeicherte
Tabellen-/Kachelwahl gilt wieder, sobald das Fenster breiter wird.

Die Tabelle bietet neun Datenspalten:

| Spalte | Angezeigter Wert |
| --- | --- |
| Bild | Primäres Zubehörbild oder allgemeines Bildsymbol. |
| Inventarnummer | Lokale Artikelidentität in RailKeeper. |
| Hersteller | Produkthersteller. |
| Artikelnummer | Hersteller- oder Katalognummer. |
| Name | Artikelname und Einstieg in die Leseansicht. |
| Art / Unterart | Konfigurierte Artikelklassifikation. |
| Spur | Alle dem Artikel zugeordneten Spurweiten. |
| Bestand | Gesamtbestand sowie freie, reservierte und eingebaute Mengen. |
| Lagerung | Erster Lagerort und Anzahl weiterer Lagerorte. |

Jede sichtbare Datenspalte ist sortierbar. Die Anfangsreihenfolge ist Inventarnummer aufsteigend.
Wähle die aktive Überschrift erneut, um die Richtung umzukehren. Eine andere Überschrift beginnt
aufsteigend. Bei sonst gleichen Werten verwendet RailKeeper die interne ID als stabilen
Tie-Breaker.

Die Spaltenauswahl ist nur in der Tabellenansicht verfügbar. Auch sie wird im aktuellen Browser
gespeichert. Anfangs sind alle neun Spalten sichtbar. Inventarnummer und Name können nicht
gleichzeitig ausgeblendet werden, damit mindestens eine Artikelidentität sichtbar bleibt.

Die Kachelansicht betont Bild, Identität, Art, Spurweite, Lagerung und Bestand. Die kompakte Liste
zeigt eine kleinere Identitäts- und Bestandszusammenfassung. Beide folgen der aktuellen
Serversortierung, bieten aber keine eigenen Sortieraktionen.

## Artikel auswählen, öffnen und verwalten

Kontrollkästchen in der Tabelle wählen eine Zeile oder alle aktuell sichtbaren Zeilen aus. Eine
Auswahl verschwindet, wenn ein neues Such- oder Filterergebnis den Artikel nicht mehr enthält. Der
stabile Stand v0.1.20.2 bietet keine Sammelaktion für diese Auswahl. Sie ist nur visueller Zustand
und keine vorgemerkte Archivierung, Auswertung oder Bestandsbuchung.

Öffne über Artikelname, Bild oder **Artikel ansehen** den schreibgeschützten Artikeldialog. Abhängig
von Rolle und Artikelzustand bieten die Zeilen außerdem:

| Aktion | Rolle und Speicherung |
| --- | --- |
| Artikel ansehen | Jede Rolle mit Zubehör-Lesezugriff. Kein Schreibvorgang. |
| Artikel bearbeiten | Admin oder Editor. Das Formular bleibt Entwurf, bis es gespeichert wird. |
| Artikel archivieren | Admin oder Editor. Schreibt sofort ohne Nachfrage und lädt die Übersicht neu. |
| Artikel wiederherstellen | Admin oder Editor. Schreibt sofort ohne Nachfrage und lädt die Übersicht neu. |
| Artikel löschen | Nur Admin. Öffnet eine Bestätigung für das endgültige Löschen. |

Endgültiges Löschen gelingt nur bei einem vollständig unbenutzten Artikel. Bestand ungleich null,
Einzelstücke, Bestandsbewegungen, Käufe, Reservierungen, Einbauten oder eine technische
Layoutposition mit Artikelbezug blockieren es. Gespeicherte Artikeldokumente blockieren einen
ansonsten unbenutzten Artikel nicht. RailKeeper entfernt beim Löschen ihre Metadaten und versucht
anschließend, auch die gespeicherten Dateien zu entfernen.

Löschen kann nicht rückgängig gemacht werden. Erstelle und validiere vor dem Entfernen eines
wichtigen Datensatzes eine aktuelle Anwendungssicherung. Die vollständigen Lösch- und
Wiederherstellungsregeln gehören zum Artikellebenszyklus.

## Lade-, Leer- und Fehlerzustände beheben

| Situation | Verhalten und nächster Schritt |
| --- | --- |
| Erstes Laden | **Artikel werden geladen ...** erscheint bis zum Abschluss der ersten Anfrage. |
| Kein Artikel vorhanden | **Noch keine Artikel vorhanden.** Admin oder Editor kann den ersten Artikel anlegen. |
| Filter ohne Treffer | **Keine Artikel entsprechen den aktiven Filtern.** Suche oder Filter anpassen oder zurücksetzen. |
| Laden schlägt vor einem Ergebnis fehl | Nur der Fehler erscheint oberhalb des leeren Listenbereichs. Sitzung und Verbindung prüfen, danach durch Ändern oder Zurücksetzen eines Filters erneut anfragen. |
| Laden schlägt nach einer Anzeige fehl | Der Fehler erscheint, während frühere Zeilen sichtbar bleiben können. Nicht als aktuell behandeln, sondern die Anfrage vor einer Aktion wiederholen. |
| Stammdatenbezeichnungen laden nicht | Zeilen können auf stabile übersetzte Art-/Unterartbezeichnungen zurückfallen. Einen Artikel erst bearbeiten, wenn die erforderlichen Editor-Ressourcen geladen sind. |
| Archivieren oder Wiederherstellen scheitert | Der Fehler erscheint oberhalb der Liste. Den Artikel bis zu einem erfolgreichen Neuladen als unverändert behandeln. |
| Löschen ist blockiert | Der Datensatz bleibt bestehen. Vor dem nächsten Versuch alle gemeldeten Bestands- oder Verwendungsbezüge auflösen. |
| Schreiben ist verboten | Mit der erforderlichen Editor- oder Admin-Berechtigung anmelden. Lesezugriff ersetzt keine serverseitige Schreibberechtigung. |

## Mit einem Artikelablauf fortfahren

Der ausführliche Zubehörbereich trennt drei Arbeitsabläufe:

- Artikeldatensatz und Fachangaben anlegen und pflegen;
- Bestand, Käufe, Einzelstücke, Bilder und Dokumente verwalten;
- Bestand reservieren oder einbauen und die Verwendungshistorie lesen.

Diese Abläufe teilen sich einen Artikeldialog, besitzen aber unterschiedliche Speicher- und
Berechtigungsregeln. Speichere oder verwirf ungespeicherte Artikelfelder bewusst, bevor du eine
sofortige Bestands- oder Lebenszyklusaktion startest.

## Verwandte Seiten

- [Überblick zum Benutzerhandbuch](/de/guide/)
- [Artikelstammdaten und Fachangaben](./article-records)
- [Bestand, Käufe und Dokumente](./stock-purchases-documents)
- [Reservierungen, Einbauten und Verwendung](./allocations-history)
- [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierter RailKeeper-Stand

Diese Seite dokumentiert den stabilen RailKeeper-Stand **v0.1.20.2** und wurde zuletzt am
2026-08-16 geprüft.
