---
title: Decoder, Funktionen und CV-Daten
description: Decoder, Funktionen und CV-Daten.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Decoder, Funktionen und CV-Daten

Wähle den Ablauf, den du gerade brauchst. Die praktischen Anleitungen verwenden künstliche
Übungsdaten und zeigen das erwartete Ergebnis. Keine dieser Übungen programmiert einen Decoder.

| Aufgabe | Anleitung |
| --- | --- |
| CV-Liste importieren und Indexseiten prüfen | [CV-Import ausprobieren](cv-import-example) |
| Modell, Firmware und Erkennung prüfen | [Den richtigen Decoder erkennen](decoder-identification) |
| F0/F1/F3 gemeinsam ergänzen und manuelle Angaben schützen | [Funktionstasten übernehmen](functions-walkthrough) |
| Ein Modell oder eine Quelle nachschlagen | [Decoderbibliothek](decoder-library) |
| Hersteller- und Firmwaregrenzen verstehen | [Modellhinweise](decoder-model-notes) |
| Formate, Rohwerte und Sonderfälle nachlesen | [Technische Referenz](decoder-cv-reference) |

<GuideFigure name="cv-analysis" alt="CV-Werte verständlich erklärt mit Quelle und Ergebnis." caption="Die echte Demo zeigt CV-Ergebnisse; technische Einzelheiten öffnen sich über die CV-Quelle." />

## Der gemeinsame Arbeitsablauf

1. **CV**: Importiere die Werteliste und prüfe Profil, Firmware und Indexseiten.
2. **Modell > Digital**: Vergleiche erkannte Werte mit den gespeicherten Fahrzeugangaben.
3. **Steuerung**: Ergänze fehlende Angaben gemeinsam; wähle fehlende Symbole und eigene Namen.
4. **Änderungen speichern**: Sichere den Fahrzeug- und Funktionsentwurf gemeinsam.
5. Öffne das Fahrzeug erneut und prüfe Ergebnis und Herkunft.

CV-Importe und Dateiaktionen speichern eigene Daten und können das Fahrzeug neu laden.
Speichere oder verwirf andere Entwürfe vorher bewusst. **Abbrechen** setzt einen bereits
abgeschlossenen CV-Import nicht zurück.

## Wenn eine Angabe unbekannt bleibt

Die Bibliothek erklärt vorhandene Werte nur innerhalb belegter Modell-, Firmware- und Indexgrenzen.
CV8 allein nennt den Hersteller. Eine fehlende Indexseite wird nicht rekonstruiert.
Ein Herstellerhinweis bedeutet noch keine automatisch berechnete Fahrzeugfunktion.

## Bisherige Detailverweise

Vorhandene Links bleiben erreichbar. Wähle unten die entsprechende Stelle der Referenz.

<span id="decoderbibliothek"></span>

- [Decoderbibliothek](decoder-cv-reference#decoderbibliothek)

<span id="voraussetzungen-und-zugriffsrechte"></span>

- [Voraussetzungen und Zugriffsrechte](decoder-cv-reference#voraussetzungen-und-zugriffsrechte)

<span id="digitalfunktionen-f0-f31-zuordnen"></span>

- [Digitalfunktionen F0-F31 zuordnen](decoder-cv-reference#digitalfunktionen-f0-f31-zuordnen)

<span id="funktionen-importieren-und-exportieren"></span>

- [Funktionen importieren und exportieren](decoder-cv-reference#funktionen-importieren-und-exportieren)

<span id="fahrkurve-lesen"></span>

- [Fahrkurve lesen](decoder-cv-reference#fahrkurve-lesen)

<span id="cv-werte-manuell-verwalten"></span>

- [CV-Werte manuell verwalten](decoder-cv-reference#cv-werte-manuell-verwalten)

<span id="railcomplus-anmeldename-und-tastensymbole"></span>

- [RailComPlus-Anmeldename und Tastensymbole](decoder-cv-reference#railcomplus-anmeldename-und-tastensymbole)

<span id="dateien-fur-einen-sauberen-import"></span>

- [Dateien für einen sauberen Import](decoder-cv-reference#dateien-fur-einen-sauberen-import)

<span id="cv-werte-importieren-und-exportieren"></span>

- [CV-Werte importieren und exportieren](decoder-cv-reference#cv-werte-importieren-und-exportieren)

<span id="decoder-dateien-prufen-ubernehmen-und-speichern"></span>

- [Decoder-Dateien prüfen, übernehmen und speichern](decoder-cv-reference#decoder-dateien-prufen-ubernehmen-und-speichern)

<span id="ecos-vorschau-als-eingangspfad-verwenden"></span>

- [ECoS-Vorschau als Eingangspfad verwenden](decoder-cv-reference#ecos-vorschau-als-eingangspfad-verwenden)

<span id="daten-bei-schreibvorgangen-schutzen"></span>

- [Daten bei Schreibvorgängen schützen](decoder-cv-reference#daten-bei-schreibvorgangen-schutzen)

<span id="fehler-bei-decoderdaten-beheben"></span>

- [Fehler bei Decoderdaten beheben](decoder-cv-reference#fehler-bei-decoderdaten-beheben)

<span id="verwandte-seiten"></span>

- [Verwandte Seiten](decoder-cv-reference#verwandte-seiten)

<span id="digitaldaten-im-fahrzeugdialog"></span>

- [Digitaldaten im Fahrzeugdialog](decoder-model-notes#digitaldaten-im-fahrzeugdialog)

<span id="kuhn-n45-geprufte-auswertung"></span>

- [Kühn N45: geprüfte Auswertung](decoder-model-notes#kuhn-n45-geprufte-auswertung)

<span id="uhlenbrock-intellidrive-76420"></span>

- [Uhlenbrock IntelliDrive 76420](decoder-model-notes#uhlenbrock-intellidrive-76420)

<span id="roco-10882-anleitung-2012-zimo"></span>

- [Roco 10882 (Anleitung 2012, ZIMO)](decoder-model-notes#roco-10882-anleitung-2012-zimo)

<span id="piko-56500-plux22"></span>

- [PIKO 56500 PluX22](decoder-model-notes#piko-56500-plux22)

<span id="d-h-dh05c-und-dh10c"></span>

- [D&H DH05C und DH10C](decoder-model-notes#d-h-dh05c-und-dh10c)

<span id="d-h-dh16a"></span>

- [D&H DH16A](decoder-model-notes#d-h-dh16a)

<span id="d-h-dh12a-alterer-plux12-decoder"></span>

- [D&H DH12A (älterer PluX12-Decoder)](decoder-model-notes#d-h-dh12a-alterer-plux12-decoder)

<span id="d-h-dh14b-mtc14"></span>

- [D&H DH14B (mTc14)](decoder-model-notes#d-h-dh14b-mtc14)

<span id="d-h-dh21a-generationen-und-anschlussvarianten"></span>

- [D&H DH21A (Generationen und Anschlussvarianten)](decoder-model-notes#d-h-dh21a-generationen-und-anschlussvarianten)

<span id="weitere-d-h-modelle-und-herstellerhinweise"></span>

- [Weitere D&H-Modelle und Herstellerhinweise](decoder-model-notes#weitere-d-h-modelle-und-herstellerhinweise)
