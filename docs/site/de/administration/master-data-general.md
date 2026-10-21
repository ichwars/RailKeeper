---
title: Allgemeine Stammdaten
description: Fahrzeugklassifizierungen, Hersteller, CV8-Kennungen und Funktionssymbole verwalten.
audience: admin
status: stable
reviewedVersion: 0.1.21
lastReviewed: 2026-10-21
---

# Allgemeine Stammdaten

Unter **Einstellungen > Daten > Allgemein** werden die gemeinsamen Auswahlen für Fahrzeuge,
Zubehör, Decoderdaten und Messen verwaltet. Die folgenden Datentypen stehen bereit.

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

## Mehrere Einträge deaktivieren

Auswahlfelder erscheinen nur bei aktiven Einträgen, die deaktiviert werden dürfen. Einzelne Zeilen
auswählen oder über das Feld im Tabellenkopf alle aktuell durch die Suche sichtbaren, geeigneten
Einträge markieren. Eine bereits gesetzte Auswahl bleibt beim Ändern des Suchbegriffs erhalten; die
Aktionsleiste zeigt deshalb immer die Gesamtzahl der im aktuellen Datentyp ausgewählten Einträge.
Beim Wechsel des Datentyps wird die Auswahl geleert.

**Ausgewählte deaktivieren** öffnet eine Bestätigung mit Datentyp und Anzahl. Erst die Bestätigung
sendet den vollständigen Satz an den Server. RailKeeper verarbeitet ihn atomar: Fehlt ein Eintrag
oder schlägt die Verarbeitung fehl, wird keiner der ausgewählten Einträge geändert und die Auswahl
bleibt für eine Korrektur erhalten. Nach erfolgreicher Deaktivierung wird sie geleert. Eine
Mehrfach-Reaktivierung ist nicht vorgesehen; inaktive Einträge einzeln reaktivieren.

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

RailKeeper liefert 94 eigenständig gezeichnete Symbole der **Werkstatt-Linie** mit. Acht davon sind
allgemeine Rückfallsymbole, 86 behalten numerische ECoS-Funktionsbeschreibungscodes ausschließlich
als Kompatibilitätsmetadaten. Die Grafiken selbst gehören zu RailKeeper und wurden nicht aus der
Digitalzentrale übernommen.

Ein Symbol besteht aus Name, optionaler Beschreibung und optionalem Bild. Administratoren können
weiterhin eigene Symbolschlüssel anlegen und SVG-, PNG-, JPEG- oder WebP-Bilder bis 1 MiB hochladen.
Das Bild liegt in den Metadaten des Eintrags und ist deshalb im Stammdatenexport und in der
Anwendungssicherung enthalten.

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
eigener Eintrag kann ebenfalls nicht gelöscht werden. Wird ein Schreibvorgang für Viewer abgewiesen, ist ein Admin- oder Editor-Konto erforderlich.

## Verwandte Seiten

- [Stammdaten-Administration](./master-data)
- [Artikelstammdaten und Lagerorte](./master-data-articles)
- [Inventarnummernschemata](./master-data-inventory-numbers)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert RailKeeper **v0.1.21** und wurde zuletzt am 20.10.2026 geprüft.

## Decoderbibliothek

Die Navigation führt über **Hersteller → Decoderfamilie → Modell**. Der recherchierte Bestand
umfasst ESU, ZIMO, Doehler & Haass, Lenz, Kühn, Uhlenbrock, Roco und PIKO. Er enthält
40 Familien, 599 Katalogeinträge und 747 belegte Artikelnummer-/Anschlusszuordnungen in den Modelldatensätzen.
Dies ist eine Auswahl aktueller und älterer Decoder, kein vollständiges Herstellerarchiv.
Der Erfassungsumfang und noch offene Sortimente stehen bei jedem Hersteller.
Katalogeinträge umfassen Hardwaretypen, Anschlussausführungen und separat bestellbare
Sounddecoder mit Fahrzeugvorbelegung. Gleiche Artikelnummern können zu mehreren Hardwareständen
gehören; sie sind kein eindeutiger Generationsnachweis. Unbelegte Anschlüsse oder Nummern bleiben offen.

| Hersteller | Einträge | Belegte Bestellzuordnungen | Noch offener Erfassungsumfang |
| --- | ---: | ---: | --- |
| ESU | 58 | 132 | V1.0, LokSound classic/V2.x, Select, L-V4.0/OEM; einzelne V2.0-Anschlüsse |
| ZIMO | 177 | 178 | Frühe MX620/MX631/MX640–647, MX69/MX690 und weitere OEM-Ausführungen |
| Doehler & Haass | 41 | 101 | Weitere OEM-/Archivvarianten; reine SelecTRIX- und SUSI-Module außerhalb der DCC-Auswertung |
| Lenz | 49 | 54 | Zusammengefasste LE-/A/E-Tabellenzeilen, weitere Revisionen und Fahrzeugplatinen |
| Kühn | 7 | 20 | F060-Artikelnummer, eindeutige Unterlagen für N020/T121/T140 |
| Uhlenbrock | 52 | 60 | Widersprüchliche 73800-Datei, weitere frühere IntelliSound-/OEM-Ausführungen |
| Roco | 30 | 30 | Frühere 107xx-, Werksdecoder- und Ersatzteilvarianten; Motorola-only 10738 außerhalb DCC |
| PIKO | 185 | 172 | Individuelle Anschlüsse bei 13 G-Soundartikeln, ältere Sound-/OEM- und Ersatzteilvarianten |

Herstellerunterlagen enthalten gelegentlich Widersprüche. Für ESU 54617/64617 wird die neuere
Variantenübersicht im LokPilot-V4.0-Handbuch (11. Auflage, S. 7) verwendet. Die ältere Katalogangabe
wird nicht als zusätzliche Anschlussvariante ausgegeben. Der Kühn F060 ist als Funktionsdecoder
mit Litzen erfasst, seine Artikelnummer bleibt unbekannt. Bei DH21A und älteren Lenz-Decodern
bestimmt die Artikelnummer allein die Hardwaregeneration nicht zuverlässig.

Die Quellen stehen jeweils beim Modell und bei der Bestellzuordnung. Der Herstellertext benennt
weitere Lücken. Auch eine große Trefferliste bedeutet keine vollständige Abdeckung aller Generationen.

Die Suche berücksichtigt Modellnamen, Artikelnummern, Anschlüsse, CV8-Kennungen und CV-Nummern.
Filter grenzen Decodertyp und Prüfumfang ein. Die DCC-Grundlagen sind separat erreichbar.

Jedes Modell bietet **Übersicht, CV-Definitionen, Erkennung und Quellen**. „CV-Regeln geprüft“ gilt
für LokPilot 5 micro DCC, DH18A, DH05C, DH10C, DH16A, DH12A, DH14B, DH21A,
DH21B, DH22A, DH22B, DH24A, DHSP10A, MN180N18, STANDARD+ V2, N45, IntelliDrive 76420,
Roco 10882 (Anleitung2012) und
PIKO 56500 PluX22. Die erfassten Tabellen umfassen zusammen mit den DCC-Grundlagen
2.091 Datensätze (2.088 funktionale Definitionen und drei ESU-Reservierungsdatensätze).
Geprüft sind die jeweils aufgeführten Register und Speicherbereiche,
keine vollständige Semantik jedes Decoderbytes oder aller Firmwarestände. Weitere Modelle
sind mit Stammdaten und Quellen erfasst; Herstellerregeln werden nicht von verwandten Modellen
übernommen. Nicht recherchierte Anschlüsse und Artikelnummern bleiben ausdrücklich offen.

Die strukturierte Datenbasis bewahrt bei geprüften Tabellenzeilen zusätzlich Quellenstelle,
Werkseinstellung, Zugriffsart, Einheit und gegebenenfalls einzelne erlaubte Werte. Variable
oder widersprüchliche Angaben bleiben offen. Werkseinstellungen ersetzen keine ausgelesenen
Fahrzeugwerte. Zusätzliche Tabellenfelder sind in den Stammdaten und Sicherungen enthalten;
die bestehende Anzeige nutzt weiterhin Bezeichnung, Erklärung und Bitangaben.

Unter **CV-Definitionen > CV-Tabellen des Herstellers** stehen zusätzlich 30.410
quellenbezogene CV-Zuordnungen für 580 der 599 Katalogmodelle. Nach CV-Nummer oder Begriff suchen;
ein CV-Bereich wird auch bei einer Nummer innerhalb dieses Bereichs gefunden. Jede Zeile nennt
Quelle, PDF-Seite und den Stand der Indexzuordnung. Die Angaben sind teilweise erfasst:
Modell-, Hardware-, Soundprojekt- und Firmwarebedingungen stehen im verlinkten Original.
19 Modelle haben keine CV-Zeilen: Für 15 fehlt der genaue Nachweis; vier sind Zusatzmodule
oder Motorola-Empfänger ohne eigene DCC-CV-Tabelle. Für die fünf älteren Lenz-Modelle
GOLD maxi, SILVER mini, SILVER direct, SILVER 21 und LF101XF sind weitere CV-Zuordnungen
aus den Originalanleitungen hinterlegt. Ausgabe und Generation beachten; Plus-/Nachfolgemodelle bleiben getrennt.
Auch LE030 sowie LE0511A/LE0511D besitzen eigene Originalnachweise mit Druckseiten und
unterschiedlichen Geschwindigkeitstabellen. Die indexierten Originaltexte sind als solche
gekennzeichnet; die alte direkte LE0511-PDF-URL war bei der Prüfung nicht verfügbar.
Für Roco/Fleischmann 685404/685504/686101/686201 ist die ausdrückliche MX622-Zuordnung
aus dem Herstellerkatalog 2022 hinterlegt. Die separate ZIMO-Anleitung belegt 40 CV-Einträge
mit Seitenangabe. Firmware und Hardwareumfang beachten; OEM-Werkswerte bleiben offen.
Für fünf spezielle ESU-LokPilot-5-Varianten und sechs PIKO-XP-Soundartikel sind zusätzlich
familienweit belegte Firmware-CVs hinterlegt. Die Zeilen nennen ihre Firmwaregrenzen;
vollständige Varianten-/Soundprojekt-Tabellen und Projekt-Werkswerte bleiben offen.
Bei PIKO-CV48 änderte sich die ABC-Schwelle mit der Firmware; für CV30 ist ein Versionswiderspruch
zwischen Herstellerquellen ausdrücklich vermerkt. Solche Angaben sind Hinweise mit Einschränkungen.
Die Referenzen erweitern die automatische Auswertung erst nach gesonderter Prüfung;
sie ändern keine Fahrzeugwerte oder manuellen Funktionen.
Lenz-CV-Editor-Vorlagen sind als solche bezeichnet. Ihre Editor-Standardwerte gelten nicht als
Decoder-Werkseinstellungen. Alle Herstellerreferenzen werden als `decoder_reference` zusammen
mit Stammdaten und Vollsicherungen gespeichert. Lokale Anpassungen bleiben beim Neustart erhalten.

Roco 10882 besitzt einen gesonderten Dokumentationsverweis, dessen geprüfte Definitionen
im Modell und in der Suche erreichbar sind. Seine Anleitung nennt als technischen
Hersteller ZIMO mit CV8=145. Daraus werden weder Regeln für andere Roco-Decoder noch eine
Erkennung anhand der Roco-Markenkennung abgeleitet.

„Bestellvarianten ansehen“ öffnet die konkrete Zuordnung **Artikelnummer → Anschluss → Quelle**.
Die beiden zusammengefassten Listen in der Übersicht sind keine paarweise Zuordnung. Bei ZIMO
ist die Bestellbezeichnung zugleich die Artikelnummer. Historische Dokumentationsstände und
angekündigte Produkte sind gekennzeichnet; das ist keine Zusicherung heutiger Lieferbarkeit.
Roco-Vertriebsartikel und ausgewählte frühere PIKO-Decoder bleiben als Markenartikel kenntlich:
Die registrierte Marken-ID bestätigt nicht die tatsächlich ausgelesene CV8. Ein OEM-Typname
gibt keine zusätzlichen CV-Regeln frei und erzeugt allein keinen falschen Herstellerkonflikt.

Quellen enthalten Handbuchstand und Geltungsbereich; der Katalog nennt sein Prüfdatum. Die Daten
wurden aus Herstellerunterlagen und dem NMRA-Herstellerverzeichnis zusammengefasst. Verlinkte
Dokumente können sich später ändern. Neue Quellen werden erst nach redaktioneller Prüfung als
CV-Regeln aufgenommen. Es erfolgt kein automatischer Abruf im Hintergrund.

Admin und Editor können die Auswertung eines geprüften Profils aktivieren oder deaktivieren.
Inaktive Definitionen bleiben in der Bibliothek lesbar. Rohwerte und Herkunft am Fahrzeug ändern
sich nicht. Ein freier Profil-Editor ist nicht enthalten. Stammdatentransfer und vollständige Backups
führen Profile als `decoder_profile` und das Modellverzeichnis als `decoder_catalog/library` mit.
Lokale Änderungen und Deaktivierungen bleiben beim erneuten Start erhalten.

CV8 allein bestätigt kein Modell. Für 39 der erfassten Modelle sind zusätzlich dokumentierte Typregeln
hinterlegt: D&H über CV261, ZIMO über CV250 und Lenz über CV7. **Erkennung** zeigt die Werte und
Grenzen des jeweiligen Modells. ZIMO bestätigt damit nur den Grundtyp, nicht die Anschlussvariante.
ESU benötigt weiterhin einen passenden Profilnamen und CV8.

Neue mitgelieferte Erklärungen aktualisieren beim Start ausschließlich bekannte, unveränderte
Vorversionen. Eigene Metadaten, Bezeichnungen, Reihenfolge und Deaktivierungen bleiben erhalten.
Details stehen unter
[Decoderbibliothek am Fahrzeug](/de/guide/vehicles/decoder-cv#decoderbibliothek).
