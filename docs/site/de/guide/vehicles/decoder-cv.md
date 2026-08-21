---
title: Decoder, Funktionen und CV-Daten
description: Digitalfunktionen zuordnen, Fahrkurven prüfen, CV-Werte pflegen und Decoder-Dateien speichern.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Decoder, Funktionen und CV-Daten

RailKeeper speichert Funktionszuordnungen, Fahrkurvendaten, CV-Werte und Decoder-Projektdateien am
Fahrzeug. Der zusammengehörige Ablauf befindet sich in **Steuerung**, **Fahrkurve** und **CV**.

## Voraussetzungen und Zugriffsrechte

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann den benötigten Tab.
Allgemeine Felder wie **Digital**, Decodernummer, Decodertyp und ABC-Bremsung erklärt
[Fahrzeugbestand und Grunddaten](/de/guide/vehicles/).

Ein Fahrzeug muss normalerweise einmal gespeichert sein, bevor RailKeeper Funktionen, CV-Werte
oder Dateien speichern kann. Ein ungespeicherter ECoS-Entwurf kann CVs und eine abgeleitete Kurve
anzeigen. Seine normalen Schreibaktionen bleiben jedoch deaktiviert.

Admin, Editor, Viewer und Planner können gespeicherte Decoderdaten ansehen. Auf Viewer-Ebene sind
auch Funktions- und CV-Export sowie der Download von Decoder-Dateien möglich. Nur Admin und Editor
dürfen Daten speichern, importieren, übernehmen, hochladen oder löschen. Der Server erzwingt diese
Grenze.

::: warning Andere Änderungen zuerst speichern
Jede erfolgreiche Schreibaktion an Decoderdaten lädt das vollständige ausgewählte Fahrzeug neu.
Speichere oder verwirf bewusst ausstehende Grunddaten, Funktionsänderungen, Bildmetadaten und
Änderungen anderer Tabs vor einer Schreibaktion.
:::

## Digitalfunktionen F0-F31 zuordnen

Öffne **Steuerung**. **Digitalfunktionen** enthält eine Zeile für jede Taste von F0 bis F31. Die
Übersicht zählt belegte, Sound- und Lichtfunktionen. Aktiviere **Nur belegte**, um ungenutzte Zeilen
auszublenden.

Jede Zeile enthält:

- **Funktionsname**
- **Symbol**
- **Betriebsart**
- **Invertiert**
- **Notiz**
- **Speichern** und **Löschen**

Die Auswahl eines Symbols ersetzt den aktuellen Funktionsnamen durch die Symbolbezeichnung und
leitet den gespeicherten Funktionstyp ab. Wähle das Symbol vor einem eigenen Namen oder stelle den
gewünschten Namen nach einem Symbolwechsel wieder her. Der Typ ist in dieser stabilen Ansicht kein
eigenes Bedienelement.

Die mitgelieferte Auswahl verwendet die eigenständig gezeichnete RailKeeper **Werkstatt-Linie**.
ECoS-Importe ordnen numerische Funktionsbeschreibungscodes stabilen RailKeeper-Schlüsseln zu. Der
Code dient nur der Interoperabilität und bezeichnet keine Grafikquelle. Administratoren können in
den Stammdateneinstellungen eigene Symbolschlüssel und Bilder ergänzen.

| Gespeicherter Typ | Deutsche Bedeutung |
| --- | --- |
| `standard` | Standard |
| `sound` | Sound |
| `licht` | Licht |
| `kupplung` | Kupplung |
| `rauch` | Rauch |
| `sonderfunktion` | Sonderfunktion |

Die Betriebsarten werden als `dauer` und `moment` gespeichert und als **Dauer** und **Moment**
angezeigt. **Invertiert** speichert das richtungsabhängige beziehungsweise invertierte Merkmal der
Zeile.

F0 beginnt mit dem Namen `Fahrlicht`, dem Lichtsymbol und dem Typ `licht`. Andere neue Zeilen
beginnen mit `standard`. Jede neue Zeile verwendet zunächst `dauer`. Eine neue Zeile benötigt
mindestens Name, Symbol oder Notiz, bevor sie gespeichert werden kann. Der lokale F0-Standard zählt
daher bereits vor **F0 speichern** als belegt.

Der Server akzeptiert nur F0-F31, bekannte Typen und Betriebsarten, Namen bis 120 UTF-8-Bytes,
Symbolschlüssel bis 80 UTF-8-Bytes und Notizen bis 1.000 UTF-8-Bytes. Speichern oder Löschen einer Zeile
wirkt sofort, lädt das vollständige Fahrzeug neu und besitzt keine zusätzliche Löschbestätigung.

### Funktionen importieren und exportieren

**Export** lädt `<inventarnummer>-funktionen.json` herunter. Ohne Inventarnummer lautet der Name
`railkeeper-funktionen.json`. Die Datei enthält Inventarnummer, Fahrzeugname, Decodernummer und alle
belegten Zuordnungen. Die Decodernummer verwendet zuerst die primäre digitale Nummer und ersatzweise
die DT-Decodernummer. Der Export verwendet die aktuellen Zeilen einschließlich ungespeicherter
lokaler Funktionsänderungen, speichert sie aber nicht in RailKeeper.

**Import** liest die erste ausgewählte JSON-Datei. Er akzeptiert ein Array auf oberster Ebene oder
ein Array unter `functions` beziehungsweise `functionMappings`. Funktionstasten werden in
Großbuchstaben umgewandelt. Zeilen mit ungültigen Tasten, Typen oder Betriebsarten werden
übersprungen. Gültige Zeilen werden ohne Vorschau oder Bestätigung nacheinander geschrieben.
Doppelte Tasten bleiben in der Reihenfolge, daher überschreibt eine spätere Zeile die frühere.

Schlägt eine Anfrage fehl, bleiben frühere Zeilen gespeichert, spätere werden nicht versucht und
die normale Aktualisierung läuft nicht. Lade das Fahrzeug neu, vergleiche die Zuordnungen und
wiederhole nur fehlende Zeilen. Die stabilen Schaltflächen **Import** und **Export** sowie manche
Importfehler bleiben unabhängig von der Oberflächensprache englisch oder deutsch.

## Fahrkurve lesen

Öffne **Fahrkurve**. Dieser Tab ist **Nur lesen**. Er berechnet eine Geschwindigkeitskennlinie aus
gespeicherten CV-Werten oder einem ECoS-Entwurf und schreibt niemals in RailKeeper, einen Decoder
oder eine Digitalzentrale.

RailKeeper gruppiert relevante CVs nach Decoderprofil und Protokoll. Wähle ein Profil, um Folgendes
zu sehen:

- Anzahl relevanter CVs in dieser Gruppe;
- Kurvenart;
- Zustand von CV 29;
- Anzahl dargestellter Punkte;
- Vorwärts-/Rückwärtstrimmung;
- Diagramm und zugrunde liegende CV-Listen;
- fehlende CVs.

Die **3-Punkt-Kurve** verwendet CV 2 bei Fahrstufe 1, CV 6 bei Stufe 14 und CV 5 bei Stufe 28. Die
**28-Punkt-Speedtable** verwendet CV 67 bis CV 94. CV 66 liefert die Vorwärts- und CV 95 die
Rückwärtstrimmung.

Ist CV 29 bekannt, wählt Bit 4 die 28-Punkt-Tabelle oder 3-Punkt-Kurve. Enthält die gewählte Kurve
keine Punkte oder ist CV 29 unbekannt, verwendet RailKeeper die sinnvollsten verfügbaren Daten:
eine vollständige 28-Punkt-Tabelle, mindestens zwei 3-Punkt-Werte, beliebige Tabellenwerte und
danach einen beliebigen 3-Punkt-Wert. Diese Auswahl ändert nur die Anzeige.

## CV-Werte manuell verwalten

Öffne **CV**. Die Übersicht zeigt die Anzahl der **CV-Werte**, **Profile** und **Dateien**.

Das manuelle Formular enthält:

| Feld | Regel |
| --- | --- |
| CV-Nummer | Erforderliche Ganzzahl von 1 bis 1024 |
| Wert | Erforderliche Ganzzahl von 0 bis 255 |
| Kategorie | Optionale gespeicherte deutsche Kategorie, bis 80 UTF-8-Bytes |
| Protokoll | Optionales Protokoll, bis 80 UTF-8-Bytes |
| Decoderprofil | Optionaler Freitext, bis 160 UTF-8-Bytes |
| Quelldatei | Optionale Decoder-Datei dieses Fahrzeugs |
| Beschreibung | Optionaler Text, bis 1.000 UTF-8-Bytes |

Stabile Kategorien sind `Adresse`, `Fahrverhalten`, `Motor`, `Licht`, `Sound`, `Funktion`,
`Decoder` und `Sonstiges`.

Protokolle sind `Motorola 14`, `Motorola 27`, `Motorola 28`, `Motorola FX 14`, `DCC 14`,
`DCC 28`, `DCC 128`, `LGB` und `Selectrix`.

Häufige Profilvorschläge sind ESU LokPilot 5, ESU LokSound 5, Zimo MS, Zimo MX, D&H SD, D&H DH,
Märklin mLD3, Märklin mSD3 und Lenz Standard+. Bereits in CV-Werten oder Dateien verwendete Profile
erscheinen als Schnellwahl. Ein Profil ist beschreibender Freitext und keine Prüfung des
physischen Decoders.

RailKeeper identifiziert eine CV-Zeile durch CV-Nummer und normalisiertes Decoderprofil. Das
Protokoll gehört nicht zu dieser Identität. **CV hinzufügen** aktualisiert eine vorhandene passende
Zeile, anstatt ein Duplikat anzulegen. **CV speichern**, **CV bearbeiten** und **CV löschen** wirken
sofort und laden das vollständige Fahrzeug neu. Löschen besitzt keine zusätzliche Bestätigung.

Ändert eine Aktualisierung den Zahlenwert, legt RailKeeper einen Historieneintrag an. Reine
Metadatenänderungen erzeugen keinen. Die stabile Oberfläche zeigt die fünf neuesten
Historieneinträge einer Zeile.

## CV-Werte importieren und exportieren

Der CV-**Import** liest die erste ausgewählte JSON-, CSV- oder TXT-Datei.

- JSON darf ein Array auf oberster Ebene oder ein Objekt mit `cvValues` sein.
- CSV oder TXT verwendet eine Zeile je Wert in der Reihenfolge CV-Nummer, Wert, Beschreibung,
  Kategorie und Decoderprofil.
- Als Trenner sind Semikolon und Komma möglich.
- Eine mit `cv` beginnende Zeile gilt als Kopfzeile.
- Beim Textimport bleiben Protokoll und Quelldatei leer.

Die Vorschau markiert jede Zeile als **neu**, **geändert**, **gleich** oder **ungültig**. Neue und
geänderte Zeilen sind zunächst ausgewählt, gleiche und ungültige nicht. Doppelte Kombinationen aus
CV-Nummer und Profil sind nach dem ersten Vorkommen ungültig.

Nutze **Nur neue**, **Alle auswählen**, **Keine auswählen** oder einzelne Kontrollkästchen und
danach **Ausgewählte Felder übernehmen**. Ausgewählte Zeilen werden nacheinander geschrieben. Ein
später Fehler macht frühere Zeilen nicht rückgängig und verhindert die normale Aktualisierung. Lade
neu, vergleiche und wiederhole nur fehlende Zeilen.

**Export** lädt `<inventarnummer>-cv.json` oder `railkeeper-cv.json` herunter. Die Datei enthält
Fahrzeugidentität, bevorzugte Decodernummer und alle mit dem Fahrzeug gelieferten CV-Datensätze
einschließlich Metadaten und Historie. Das aktuelle ungespeicherte CV-Formular,
Funktionszuordnungen und Decoder-Dateiinhalte sind nicht enthalten.

Einige stabile Vorschau-, Status- und Validierungsmeldungen bleiben in der englischen Oberfläche
deutsch. Die sichtbaren CV-Schaltflächen **Import** und **Export** bleiben in beiden Sprachmodi
englisch.

## Decoder-Dateien prüfen, übernehmen und speichern

Trage unter **CV-Dateien** ein optionales Decoderprofil und eine Bemerkung ein und wähle
**CV-Datei hochladen**. Mehrere Dateien sind möglich.

Unterstützte Endungen:

- JSON, CSV, TXT und XML
- Z21
- ESU und ESUX
- LokProgrammer
- ZIP

Die normale Grenze beträgt 25 MiB je Datei. Ein Betreiber kann eine andere Beilagengrenze
konfigurieren. RailKeeper weist nicht unterstützte Endungen, blockierte ausführbare oder
skriptartige Inhalte, leere und zu große Dateien zurück.

Die Auswahl erzeugt zuerst eine **Upload-Vorschau**. Sie kann Größe, MIME-Typ, Vorschaubild,
Projekt, Decoder, Adresse, Typ, Hersteller, LokProgrammer-Metadaten sowie die Anzahl erkannter
CV-Werte und Funktionstasten anzeigen. Eine Vorschau speichert die Originaldatei nicht.

Jede Dateivorschau liefert höchstens 32 erkannte CV-Werte und 32 erkannte Funktionen. Eine
angezeigte Anzahl von 32 kann daher bedeuten, dass die Quelle weitere Einträge enthält. Vergleiche
die Vorschau mit der Projektdatei. Importiere übrige CVs über den direkten CV-Import und übrige
Funktionen per Funktions-JSON oder manueller Eingabe.

Die Vorschauaktionen sind unabhängig:

1. **Vorschlag übernehmen** kopiert das erste erkannte Profil und die Beschreibung in die noch
   ungespeicherten Dateifelder.
2. **CVs prüfen** übergibt erkannte Werte an die normale CV-Importvorschau. Erst das Übernehmen
   ausgewählter CV-Zeilen schreibt Daten. Ein erkanntes Profil gewinnt, andernfalls verwendet
   RailKeeper nacheinander aktuelles Dateiprofil, erkannten Decoder oder erkannten Projektnamen.
3. **Funktionen übernehmen** schreibt gültige erkannte Funktionen sofort. Sie verwenden erkannten
   Namen und Typ, ein leeres Symbol, Betriebsart `dauer`, keine Richtungsabhängigkeit und den
   Vorschaudateinamen als Notiz.
   Doppelte Tasten werden zusammengeführt, wobei die später erkannte Zuordnung gewinnt.
4. **Dateien speichern** speichert die ausgewählten Originaldateien mit aktuellem Profil und
   aktueller Bemerkung.

Das Übernehmen von CVs oder Funktionen speichert nicht die Originaldatei. Das Speichern von Dateien
übernimmt nicht automatisch erkannte CVs oder Funktionen. Erkannte ESU/LokProgrammer-Metadaten
können beim Speichern leer gelassene Felder füllen.

Schlägt eine Datei während der Vorschauerzeugung fehl, wurde keine Datei gespeichert und der Stapel
erzeugt nicht die normale Vorschau. **Funktionen übernehmen** und **Dateien speichern** senden
Anfragen nacheinander. Ein später Fehler lässt frühere Ergebnisse gespeichert und verhindert die
normale Aktualisierung. Lade neu und vergleiche vor einem erneuten Versuch.

Gespeicherte Dateien zeigen Originalname, Profil, MIME-Typ, Größe und Beschreibung. **Download**
liefert die Originaldatei. **Löschen** entfernt sie sofort ohne weitere Bestätigung und entfernt die
gespeicherten Dateidaten, wenn keine Referenz verbleibt.

Ein CV-Wert kann unter **Quelldatei** eine Decoder-Datei nennen. Bearbeite vor dem Löschen einer
Datei die CV-Zeilen, prüfe dieses Feld, wähle für jede passende Zeile **Ohne Datei** und speichere
die Änderung. Die stabile CV-Tabelle zeigt die Quellenzuordnung nicht. Das Löschen einer Datei
löscht weder ihre CV-Werte noch deren gespeicherte Quellen-ID. Ohne diese Reihenfolge können
veraltete Referenzen verbleiben.

## ECoS-Vorschau als Eingangspfad verwenden

Ein ungespeicherter ECoS-Lokomotiventwurf kann CV-Werte und Funktionszuordnungen liefern, bevor das
Fahrzeug existiert. Der Tab **CV** zeigt die ersten 18 Werte, die Anzahl weiterer Werte und die
Quelllokomotive. **Fahrkurve** kann daraus seine schreibgeschützte Anzeige ableiten. **Steuerung**
lädt die Funktionszuordnungen des Entwurfs in die bearbeitbaren Funktionszeilen.

Mit aktivem ECoS-Entwurf ist das Speichern des Grunddatensatzes ein besonderer mehrstufiger Ablauf.
RailKeeper speichert zuerst das Fahrzeug und andere ausstehende Fahrzeugdaten, danach nacheinander
die externe ECoS-Zuordnung, die CV-Werte des Entwurfs und alle derzeit belegten Funktionszeilen. Nur
bei vollständigem Erfolg markiert RailKeeper den ECoS-Import als gespeichert und lädt das Fahrzeug
neu. Ein später Fehler macht weder das Fahrzeug noch frühere ECoS-Schreibvorgänge rückgängig. Lade
neu, vergleiche gespeicherte Zuordnung, CVs und Funktionen mit dem Entwurf und wiederhole nur
fehlende Arbeiten.

Nach erfolgreichem Abschluss verwenden normale Funktions-, CV- und Dateiaktionen das gespeicherte
Fahrzeug. Dieses Kapitel erklärt keine ECoS-Verbindung, Rohprüfung, Synchronisierung,
Konfliktbehandlung oder Schreibvorgänge zur Digitalzentrale. Diese Abläufe gehören zum geplanten
Kapitel Digitalzentralen.

## Daten bei Schreibvorgängen schützen

| Aktion | Speichert Daten | Lädt das vollständige Fahrzeug neu |
| --- | --- | --- |
| Funktionsfeld ohne Speichern bearbeiten | Nein | Nein |
| Einzelne Funktion speichern oder löschen | Sofort | Nach Erfolg |
| Funktions-JSON importieren | Nacheinander | Nur nach vollständigem Erfolg |
| Funktions-JSON exportieren | Nein | Nein |
| Fahrkurve ansehen | Nein | Nein |
| Mit aktivem ECoS-Entwurf speichern | Fahrzeug, Zuordnung, CVs und Funktionen nacheinander | Nach Gesamterfolg |
| CV-Importvorschau erzeugen oder auswählen | Nein | Nein |
| Ausgewählte CV-Zeilen übernehmen | Nacheinander | Nur nach vollständigem Erfolg |
| Einzelnen CV hinzufügen, speichern oder löschen | Sofort | Nach Erfolg |
| CV-JSON exportieren | Nein | Nein |
| Decoder-Dateivorschau erzeugen | Nein | Nein |
| Metadatenvorschlag übernehmen | Nein | Nein |
| Erkannte CVs prüfen | Nein, bis Zeilen übernommen werden | Nein |
| Erkannte Funktionen übernehmen | Nacheinander | Nur nach vollständigem Erfolg |
| Decoder-Dateien speichern | Nacheinander | Nur nach vollständigem Erfolg |
| Decoder-Datei herunterladen | Nein | Nein |
| Decoder-Datei löschen | Sofort | Nach Erfolg |

Funktionen, CV-Werte, CV-Historie, Decoder-Dateimetadaten und Decoder-Dateiinhalte gehören zur
lokalen RailKeeper-Anwendungssicherung. Erstelle und validiere vor großen Importen oder
Aufräumarbeiten eine aktuelle Sicherung. Funktions- und CV-JSON-Exporte sind Austauschdateien,
keine vollständigen RailKeeper-Sicherungen.

## Fehler bei Decoderdaten beheben

| Situation | Reaktion |
| --- | --- |
| Fahrzeug ist nicht gespeichert | Speichere den Grunddatensatz vor dauerhaften Funktions-, CV- oder Dateiaktionen. |
| Keine Funktion unter **Nur belegte** | Deaktiviere den Filter, trage Name, Symbol oder Notiz ein und speichere. |
| Funktionsimport findet nichts Gültiges | Prüfe JSON-Struktur, F0-F31, gespeicherte Typnamen und Betriebsarten. |
| Fahrkurve ist leer | Erfasse oder importiere CV 2/5/6 oder CV 67-94 in einer Profil-/Protokollgruppe. |
| Kurvenauswahl wirkt falsch | Prüfe CV 29, Profil-/Protokollgruppe und fehlende CVs. |
| CV-Eingabe wird abgelehnt | Nutze eine ganzzahlige CV von 1-1024 und einen Wert von 0-255. |
| CV-Import meldet ein Duplikat | Behalte eine Kombination aus CV-Nummer und Profil in der Quelle. |
| Dateivorschau enthält keine Metadaten | Die Datei kann trotzdem gespeichert werden, es gibt aber keinen Vorschlag. |
| Datei wird abgelehnt | Prüfe Endung, Inhalt, leere Datei und die Größengrenze des Betreibers. |
| Datei ist noch CV-Quelle | Entferne **Quelldatei** an passenden CV-Zeilen vor dem Löschen. |
| Stapel schlägt teilweise fehl | Lade neu, vergleiche gespeicherte Ergebnisse und wiederhole nur fehlende Elemente. |
| Andere Änderungen verschwinden | Neuladen ersetzt Ungespeichertes. Nutze bei Bedarf eine Sicherung. |
| CV oder Datei wurde ohne Nachfrage gelöscht | Löschen wirkt sofort. Wiederherstellung benötigt eine Sicherung. |

Keine später fehlgeschlagene Anfrage macht eine frühere erfolgreiche Anfrage derselben
nacheinander ausgeführten Aktion rückgängig.

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.19.2** und wurde zuletzt am
16.08.2026 geprüft.
