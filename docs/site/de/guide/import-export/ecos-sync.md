---
title: ECoS-Loks und gespeicherte CVs übernehmen
description: ECoS-Loks und gespeicherte CVs übernehmen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-29
---

# ECoS-Loks und gespeicherte CVs übernehmen

**Ziel:** Lokdaten aus der ECoS mit deinem Bestand vergleichen und gespeicherte CV-Werte übernehmen. Der Ablauf benötigt Admin-Rechte und eine [eingerichtete ECoS-Verbindung](../../administration/digital-centers).


<GuideFigure name="central-workspace" alt="Digitalzentralen ist ein eigener Arbeitsbereich; die Demo hat keine Zentrale verbunden." caption="Digitalzentralen ist ein eigener Arbeitsbereich; die Demo hat keine Zentrale verbunden. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## 1. Lokliste lesen und die richtige Lok wählen

1. Öffne **Digitalzentralen** in der Seitenleiste.
2. Prüfe die aktive ECoS und wähle **Daten lesen**.
3. Wähle die gewünschte Lok in der Arbeitsliste.
4. Vergleiche Name, Adresse, Protokoll und ECoS-Objekt-ID mit dem tatsächlichen Fahrzeug.

Eine gespeicherte externe Zuordnung ist ein wichtiger Anhaltspunkt. Adresse und Name allein beweisen keine Identität. Prüfe vorgeschlagene RailKeeper-Treffer vor jeder Übernahme. Die Vergleichsarbeitsliste liegt auf dem Server und kann später erneut geöffnet werden.

## 2. Gespeicherte Lokdaten gemeinsam prüfen

**Daten lesen** sammelt im selben Abgleich Lokdaten, gespeicherte Funktionstasten und Symbole
sowie die vollständigen CV-Exporte einschließlich Indexseiten. Die lokal eingerichtete
Sicherungsauswertung ergänzt und prüft diese Daten intern. Du musst keine Datei herunterladen
oder zwischen verschiedenen Importen wählen. Es wird kein Decoder neu ausgelesen und keine
Lok auf dem Programmiergleis benötigt.

Öffne die gewünschte Lok in der Arbeitsliste. Die Anzeige nennt die gespeicherten Werte,
ihre Herkunft und mögliche Widersprüche. Fehlende Angaben bleiben unbekannt. Profilvorgaben
und gespeicherte Werte sind keine aktuelle Messung am Decoder. Wenn eine interne Datenquelle
nicht verfügbar ist, bleibt diese Einschränkung sichtbar.

## 3. Übernahme prüfen

Bei widersprüchlichen Quellen werden die abweichenden Werte getrennt angezeigt und nicht
automatisch übernommen. Unbestätigte Speicherbytes gelten nicht als bekannte CV-Werte.
Für bereits zugeordnete Fahrzeuge ergänzt **CV-Werte ergänzen** nur fehlende vollständige
Registeradressen aus dem gelesenen Stand. Bestehende Werte und manuelle Angaben bleiben erhalten.

Kontrolliere unter **Gelieferte Werte und Herkunft ansehen** die Anzahl, Werte und Spalte **CV31 / CV32**. Beispielsweise ist CV257 auf Seite 16/3 ein anderer Eintrag als CV257 auf Seite 16/8. Vorhandene Werte derselben vollständigen Adresse bleiben mit ihrer Herkunft erhalten, auch wenn sie unter einem anderen Profil gespeichert sind. Ein erneuter Import erzeugt keine Duplikate. ECoS bleibt als Herkunft erkennbar.

Die Bibliothek berücksichtigt auch das gespeicherte ECoS-Decoderprofil. Passt dessen Name eindeutig
zu einem aktiven Modellprofil, können dessen dokumentierte CV- und Funktionsregeln verwendet werden,
auch wenn CV8 fehlt. Die Anzeige nennt das ECoS-Profil ausdrücklich als Quelle. Widersprüche zum
manuellen Decoder-Typ oder zu vorhandenen Hersteller- und Modell-CVs sperren diesen Abgleich.
Generische, unbekannte oder mehrdeutige Profile erhalten keine geratenen Modellregeln.

Manuelle Decoder-Typen, CV-Werte und Funktionszuordnungen bleiben erhalten. Fehlende CV7/CV8,
Firmware und Indexseiten werden nicht erfunden. Funktionsnamen und Symbole ergänzt du ausdrücklich
im Entwurf. Das Profil ist gespeicherter ECoS-Kontext und keine neue Messung am Decoder.
Nach dem Löschen der Vergleichsarbeitsliste oder einer JSON-Wiederherstellung ohne diese Daten liest
du die ECoS erneut ein. [CVs verstehen](../vehicles/decoder-cv) und [Decoderbibliothek](../vehicles/decoder-library).

## 4. Ein Fahrzeug anlegen oder zuordnen

Die bereits gelesenen CVs und Funktionstasten werden im selben Entwurf verwendet. Bekannte
Symbole und belegte Bedeutungen werden mit der Decoderbibliothek abgeglichen. Unbekannte
Codes und widersprüchliche Angaben bleiben sichtbar. Du prüfst die normalen Fahrzeug- und
Funktionsfelder und speicherst das Fahrzeug; ein weiterer Importdialog ist dafür nicht nötig.

Wähle die passende Übernahmeaktion der Arbeitsliste. Bei einem neuen Fahrzeug ergänzt du die Pflichtfelder, die die Zentrale nicht kennt, etwa Hersteller, Spurweite und Gattung. Prüfe die Zuordnung zu einem bestehenden Fahrzeug besonders sorgfältig. Speichere erst danach.

Die Fahrzeuganlage mit externer Zuordnung ist ein mehrstufiger Ablauf. Falls eine spätere Aktion scheitert, prüfe den bereits gespeicherten Fahrzeugstand und den Auftrag, bevor du neu anlegst. So vermeidest du Duplikate.

## Funktionstasten im Fahrzeugentwurf

Bei einer neuen Lok enthält der normale Fahrzeugentwurf bereits die gemeinsam gelesenen
Funktionstasten. Prüfe Namen, Symbole und Schaltart in den normalen Funktionsfeldern.
Unbekannte Codes bleiben mit Herkunft erkennbar. Quellen- und Bibliothekswidersprüche werden
nicht automatisch aufgelöst. Eigene vorhandene Zuordnungen bleiben erhalten.

<details>
<summary>Bestehende Fahrzeuge gezielt nachbearbeiten</summary>

1. Speichere die Zuordnung zwischen ECoS-Lok und Fahrzeug unter **Digitalzentralen**.
2. Öffne **Fahrzeug bearbeiten > Steuerung > Funktionstasten aus ECoS übernehmen**.
   Das Lesen steht Admins zur Verfügung.
3. Wähle **Aus ECoS lesen**. Alle Tasten F0–F31 werden geprüft; ausgeblendete und nicht
   verfügbare Tasten erzeugen keine neuen Einträge. Es wird keine Funktion geschaltet.
4. Prüfe ECoS, bisherige RailKeeper-Werte und die **Decoderbibliothek** nebeneinander.
   Die Bibliothek nutzt das gewählte CV-Profil, passende gespeicherte ECoS-Decoderprofile,
   belegte Decoderidentität und vorhandene Indexseiten.
   Ein fehlender Index wird nicht ergänzt oder geraten. Unbekannte Symbolcodes bleiben erkennbar.
5. Markiere die gewünschten Zeilen einzeln oder nutze **Freie Angaben auswählen**.
   Bei unterschiedlichen Symbolen oder Schaltarten aus ECoS und Decoderdaten bleibt die Zeile
   gesperrt. Prüfe diese Angaben anschließend einzeln im Funktionseditor.
6. Wähle **Auswahl in Entwurf übernehmen** und kontrolliere die Funktionseinträge.
   Manuelle Namen, Symbole, gespeicherte Schaltarten und eigene Notizen bleiben erhalten.
   Ein Herkunftsnachweis nennt ECoS, Taste, Symbolcode und Lesezeit; verwendete CV-Nachweise bleiben sichtbar.
7. Speichere das Fahrzeug gemeinsam oder verwirf den Entwurf mit **Abbrechen**.

</details>

Modellregeln gibt es auch für **LokPilot 5 nano DCC**, **LokPilot micro V4.0 / DCC**
und **LokPilot Standard**. Sie verwenden ihre jeweils belegten CV- und Mapping-Tabellen.
Fehlende Register werden nicht als Null gewertet; bei unvollständigen Mapping-Zeilen bleibt
die Funktionsbedeutung offen. Beim Standard reicht die geprüfte ältere Mapping-Tabelle bis F18.
Die ECoS-Tasten F19–F31 können weiterhin mit den verfügbaren Angaben übernommen werden.

Die Symbolzuordnung nutzt dokumentierte ECoS-Protokollcodes, etwa 10 für Rangiergang und
11 für ABV. Unbekannte Codes bleiben sichtbar und können manuell zugeordnet werden.
Das Bibliotheksupdate erhält eigene Symboleinstellungen und bestehende Fahrzeugzuordnungen.

<GuideFigure name="ecos-functions" alt="Funktionsvorschau mit geschützten manuellen Angaben und einem Widerspruch zur Decoderbibliothek." caption="Echte RailKeeper-Komponenten mit synthetischen Beispieldaten und simulierter ECoS-Antwort. Die Auswahl ergänzt freie Angaben erst im Entwurf; sie schaltet keine Funktion an der Zentrale." />

## Was kann zur ECoS zurückgeschrieben werden?

Lesen und Übernehmen in RailKeeper verändert keine Decoder-CVs. Der separate Schreibablauf zeigt Änderungen an Lokdaten vorab und verlangt eine Bestätigung. Bildübertragung und Bildzuordnung besitzen eine [eigene Vorschau](../vehicles/locomotive-images). Die CV-Übernahme programmiert keinen Decoder und ersetzt kein LokProgrammer-Projekt.

## Wenn etwas fehlt

| Beobachtung | Nächster Schritt |
| --- | --- |
| Weboberfläche erreichbar, Protokoll nicht | TCP-Port 15471 vom RailKeeper-Server prüfen; die Weboberfläche nutzt einen anderen Port |
| Weniger CVs als im LokProgrammer-Export | Die ECoS liefert nur ihren gespeicherten Cache; bei Bedarf den vollständigen TXT-Export im Fahrzeug importieren |
| Ein Wert hat keine Erklärung | Modell, Firmware und Indexkontext prüfen; fehlende Daten bleiben offen |
| Falscher Fahrzeugtreffer | Übernahme unterbrechen und die Zuordnung korrigieren |
| Abruf unvollständig | Fortschritt und Fehler prüfen, dann kontrolliert fortsetzen |

Der [CS3-Leseablauf](cs3-import) übernimmt andere Daten. Z21 und Intellibox 3 bieten noch keinen gleichwertigen Lok-/CV-Import.
