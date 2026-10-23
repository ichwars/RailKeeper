---
title: F0, F1 und F3 gemeinsam übernehmen
description: F0, F1 und F3 gemeinsam übernehmen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# F0, F1 und F3 gemeinsam übernehmen

**Ziel:** Funktionen aus einer CV-Datei ergänzen, gemeinsam speichern und eigene Angaben schützen.
Verwende ein gespeichertes Testfahrzeug ohne Funktionen und ohne CVs, mit **Admin** oder **Editor**.
Die Übung schreibt ausschließlich in RailKeeper.

<GuideDownload file="functions-cv-demo.txt">Künstliche CV-Datei für F0, F1 und F3 herunterladen</GuideDownload>

Die Datei enthält **144 Werte**, ein ESU-Profil mit Firmware **5.14.183** sowie die Indexseiten
`1/1`, `16/3` und `16/8`. Die Werte sind keine Werkseinstellungen oder Programmierempfehlungen.
Die Schaltart-/Invertierungsregel wurde im Projekt anhand getrennter F0-, F1- und F3-Vergleichsexporte
mit LokProgrammer **5.2.18** abgesichert. Das ist kein allgemeiner Nachweis für andere Decoder.

## 1. CV-Datei importieren

1. Öffne **Bearbeiten > CV** und importiere die TXT unverändert.
2. Prüfe **144 neue Werte** und `LokPilot 5 micro DCC` in der Vorschau.
3. Übernimm die Werte. Kontrolliere das Profil und die Herkunft wie in der
   [Decodererkennung](decoder-identification).
4. Öffne **Steuerung**. Lasse **Nur belegte** eingeschaltet.

<GuideFigure name="functions-recognized" alt="Steuerung mit drei erkannten, noch nicht gespeicherten Funktionstasten." caption="Die echte Demo erkennt F0, F1 und F3 aus dem Mapping. Die Funktionseinträge sind vor der gemeinsamen Übernahme noch nicht gespeichert." />

## 2. Fehlende Angaben gemeinsam ergänzen

Wähle **Fehlende Angaben aus CVs ergänzen**. Prüfe die drei Zeilen:

| Taste | Erwartete Bedeutung | Schaltart | Invertiert |
| --- | --- | --- | --- |
| F0 | Licht vorn/hinten, mit belegbarer Bezeichnung und Lichtsymbol | Dauer | Nein |
| F1 | AUX1; der tatsächliche Verbraucher muss von dir benannt werden | Dauer | Nein |
| F3 | Rangiergang, mit passendem Symbol | Dauer | Nein |

AUX1 verrät nicht, ob an der Lok beispielsweise eine Innenbeleuchtung angeschlossen ist.
Ergänze bei Bedarf einen eigenen Namen. Eine dezente Markierung mit Text zeigt fehlende Symbole;
ein Klick öffnet direkt die Symbolauswahl.

**Dauer (Vorgabe)** bedeutet bei einer neuen Zeile nur den Startwert. Ein übernommener Wert besitzt
seinen Herkunftsnachweis. Richtungsabhängigkeit ist eine eigenständige Angabe und bleibt unbekannt,
wenn sie für den Funktionseintrag nicht gesichert übernommen werden kann.

## 3. Herkunft prüfen und speichern

Öffne bei F1 **Details > Herkunft und CV-Werte**.

<GuideFigure name="functions-provenance" alt="F1-Details mit übernommenem RailComPlus-Nachweis und aktuellen CV-Werten." caption="Der Entwurf nennt Quelle, Bytepaar, Originalzeilen und Exportdatum. Gespeicherter Nachweis und aktuelle CVs sind getrennt lesbar." />

Erwarte `CV31=1 / CV32=1`, `CV259=32`, `CV260=2`, Originalzeilen und Exportdatum aus der Datei.
F0 nutzt CV257/258, F3 CV263/264. Die Indexseite gehört immer dazu.

Wähle **Änderungen speichern**, öffne das Fahrzeug erneut und prüfe F0, F1 und F3.
Erwarte weiterhin **Dauer**, **Invertiert: Nein** und den gespeicherten Herkunftsnachweis.
Der technische Mappingbereich bleibt am Seitenende eingeklappt.

## 4. Schutz eigener Angaben prüfen

1. Öffne erneut F1 und seine **Details**.
2. Wähle manuell **Moment** und **Invertiert: Ja**.
3. Prüfe die gemeinsame Ergänzungsaktion: In dieser Übung gibt es keine weiteren Lücken;
   die Aktion ist deaktiviert. Deine beiden Angaben bleiben im Entwurf erhalten.

<GuideFigure name="functions-manual-protected" alt="Manuell geänderte F1 mit Moment und Invertiert Ja; die Ergänzungsaktion ist gesperrt." caption="Eigene Entscheidungen werden nicht aus den CVs zurückgesetzt. Dies ist ein ungespeicherter Testentwurf." />

4. Wähle **Abbrechen**. Öffne das Fahrzeug erneut.
5. F1 muss wieder **Dauer** und **Invertiert: Nein** zeigen; der ursprüngliche Nachweis bleibt erhalten.

Für einen dauerhaften eigenen Wert würdest du stattdessen **Änderungen speichern** wählen.
Gespeicherte Schaltarten sind geschützt; **Unbekannt** bei Invertierung kann ergänzt werden.
Sind an anderen Tasten noch Angaben offen, darf die gemeinsame Aktion dafür aktiv bleiben.

## Funktionen als Datei austauschen

<GuideDownload file="functions-manual-demo.json">Manuelle Funktionsbeispiele als JSON herunterladen</GuideDownload>

Optional auf einem **zweiten Testfahrzeug**: Importiere diese JSON über **Steuerung > Import**.
Sie setzt F0/F1/F3 absichtlich auf **Moment** und **Invertiert: Ja** und trägt eine eigene Übungsnotiz.
Prüfe den Entwurf und speichere gemeinsam oder verwirf ihn mit **Abbrechen**.
Dies ist eine Funktionsdatei, keine CV-Liste und keine Sicherung. Sie enthält keine vorgetäuschten CV-Nachweise.

[Formate und Grenzen](decoder-cv-reference#funktionen-importieren-und-exportieren) ·
[CV-Import üben](cv-import-example)
