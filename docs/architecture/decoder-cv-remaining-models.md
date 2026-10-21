# Herstellerzuordnung der 26 zuletzt offenen Decodermodelle

Stand: 21.10.2026. Alle 26 Modelle sind einzeln geprüft. Für elf Modelle sind erstmals
**begrenzte gemeinsame Firmware-CVs** belegt: fünf ESU- und sechs PIKO-Modelle.
Die übrigen 15 Modelle bleiben ohne ausreichend belegte CV-Zuordnung. Das Prüfergebnis
wird für alle 26 Modelle festgehalten, auch wenn eine Suche keine passende Originalunterlage ergab.
Eine abgeschlossene Prüfung bedeutet keine vollständige Dokumentation jedes Modells.

Die Ergänzung umfasst **33 gemeinsame Referenzzeilen** (ESU 5, PIKO 28). Sie liefert weder
vollständige Varianten-/Soundprojekt-Tabellen noch neue automatische Auswertungsregeln.
Die Tabellen sind in der Bibliothek ausdrücklich teilweise erfasst. Bestehende Rohwerte,
Herkunft, manuelle Beschreibungen und Funktionseinstellungen bleiben unverändert.

## ESU: gemeinsame Firmwarehinweise, genaue Variantentabellen fehlen

Die [offizielle LokProgrammer-Firmwarehistorie](https://www.esu.eu/download/software/lokprogrammer/)
beschreibt die CVs ausdrücklich für LokPilot 5 beziehungsweise die LokPilot-5-Familie.
Die Hersteller-Produktseiten ordnen alle fünf Modelle dieser Familie zu. Übernommen werden
nur diese fünf eindeutigen gemeinsamen Angaben, jeweils als Handbuch-/Firmwarehinweis:

| CV | Bedeutung | Dokumentierter Familien-Firmwarestand |
| --- | --- | --- |
| 11 | Zeit bis zum Analogwechsel ohne Datenempfang | ab 5.14.183, LokProgrammer 5.2.18 |
| 114 | Einblendzeit der Lichtausgänge | ab 5.4.139, LokProgrammer 5.1.4 |
| 115 | Ausblendzeit der Lichtausgänge | ab 5.4.139, LokProgrammer 5.1.4 |
| 131 | Helligkeit bei aktivem Dimmer | ab 5.4.139, LokProgrammer 5.1.4 |
| 199 | Lampenzahl der Class-Light-Logik | ab 5.8.156, LokProgrammer 5.2.4 |

Diese Versionsgrenzen bestätigen keine tatsächlich geladene Fahrzeugfirmware und keine
historische Lieferbarkeit jeder Variante. Ausgangsbelegungen, Wertebereiche, Werkseinstellungen,
Sound-CVs und die vollständige Tabelle der älteren Sammelanleitung werden nicht übertragen.
Die Hersteller-Downloadliste nennt die speziellen Varianten nicht als eigene LokPilot-Anleitungen;
bei Händlerlinks mit variantenspezifischem Titel führt der Download weiterhin zur Sammelanleitung.

| Modell / Artikel | Ergebnis | Weiter offen |
| --- | --- | --- |
| [LokPilot 5 micro DCC Direct](https://www.esu.eu/produkte/lokpilot/lokpilot-5-micro-dcc-direct/) · 59721 | Fünf gemeinsame Firmware-CVs mit obigen Grenzen. | Platinen-/LED-Zuordnung und vollständige Direct-Tabelle. |
| [LokPilot 5 micro DCC Direct Atlas Legacy](https://www.esu.eu/produkte/lokpilot/lokpilot-5-micro-dcc-direct-atlas-legacy/) · 59751 | Fünf gemeinsame Firmware-CVs; eigene Produktfassung nachgewiesen. | Vollständige Atlas-Legacy-Belegung; keine Übertragung aus der Soundversion. |
| [LokPilot 5 micro DCC Direct Kato USA Widebody](https://www.esu.eu/produkte/lokpilot/lokpilot-5-micro-dcc-direct-kato-usa-widebody/) · 59941 | Fünf gemeinsame Firmware-CVs. | Genaue Widebody-Hardware-/Ausgangstabelle. |
| [LokPilot 5 micro DCC Kato USA](https://www.esu.eu/produkte/lokpilot/lokpilot-5-micro-dcc-kato-usa/) · 59741 | Fünf gemeinsame Firmware-CVs. | Vollständige Kato-USA-Belegung; keine Gleichsetzung nur über den Anschluss. |
| [LokPilot 5 Nano DCC](https://www.esu.eu/produkte/lokpilot/lokpilot-5-nano-dcc/) · 59925 | Fünf gemeinsame Firmware-CVs; Nano-5-Familienbezug nachgewiesen. | Genaue E24-/Nano-5-Tabelle. Nano Standard/Fx Nano 536xx ist eine andere Generation. |

## PIKO: gemeinsame XP-CVs, keine übernommenen Projektwerte

Die [offizielle XP-Firmwarehistorie](https://www.piko-shop.de/de/seite/decoder-firmwarestand-122.html)
und die [H0-Hersteller-FAQ](https://www.piko-shop.de/de/seite/fragen-zu-h0-57.html) belegen
28 gemeinsame Referenzzeilen für 37 CV-Adressen. Die Firmwarehistorie reicht von 1.1.0 bis
1.21.0 (29.07.2026); jede Zeile nennt den einschlägigen Stand. Der XP-Bezug stammt aus der
[Hersteller-Familienseite](https://www.piko-shop.de/de/seite/piko-smartdecoder-xp-5.1-110.html),
den konkreten Produktseiten und für 56579 zusätzlich dem
[Original-Ersatzteilblatt 51775](https://www.piko-shop.de/is.php?id=37608), Seite 2.

Erfasst sind CV3/4, 7, 11/12, 15–18, 22, 29/30, 47–49, 53, 62, 96, 115, 121/122,
147–149, 152/153, 196/197, 240–245, 248, 250 und 254. Zusammengefasste Bereiche erhalten
keine erfundenen Einzelbyte-Bedeutungen. Ausgelesene Werte werden weder ergänzt noch verändert.

- **CV48:** Bei Wert 0 nennt Firmware 1.16 eine ABC-Schwelle von 1,0 V, Firmware 1.18 dagegen
  0,5 V. Daher keine universelle Umrechnung und keine automatische Bremszustandsableitung.
- **CV30:** Die FAQ nennt Soundübertragungsfehler ab 1.19, die Firmwarehistorie führt Bit4 erst
  mit 1.20 ein. Beide Quellen bleiben sichtbar; der Widerspruch wird nicht in eine Bitregel aufgelöst.
- **CV15/16:** Die Entsperränderung ist ab 1.21 dokumentiert. Sie wird beschrieben, nicht ausgeführt.
- **CV254:** Der Vorrang der Sanftanfahrrampe gegenüber Anfahr-/Bremsverzögerung ist ab 1.20 belegt.
- Keine projektabhängigen Soundslot-, Funktionstasten- oder Ausgangsbelegungen, Werkswerte oder
  XP-G-spezifischen Poti-/Pufferspeicher-CVs. Offene Bereiche wie CV191ff/224ff bleiben unbestimmt.

| Modell / Artikel | Ergebnis | Weiter offen |
| --- | --- | --- |
| [Rh 2800 PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-s-rh-2800-plux22-inkl.-lautsprecher-35119.html) · 56560 | 28 gemeinsame XP-Firmwarezeilen; explizite Nennung auf der XP-5.1-Familienseite. | Projektbelegung; älteres Rh-2800-Projekt 56360 ist kein Ersatz. |
| [Rh 1110.5 PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-s-rh-1110.5-plux22-inkl.-lautsprecher-35138.html) · 56579 | 28 gemeinsame XP-Firmwarezeilen; Original-Teileblatt bestätigt XP 5.1 Sound. | Projektbelegung; Blank-Decoder 56505 und BR491 36520 belegen keine Projektwerte. |
| [CC 65000 PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-xp-sound-cc-65000-plux22-inkl.-lautsprecher-49608.html) · 56658 | 28 gemeinsame XP-Firmwarezeilen, an die dokumentierten Stände gebunden. | Eigene Projekt-/Auslieferungsfassung; aktuell noch angekündigt. |
| [MZ PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-xp-sound-mz-plux22-inkl.-lautsprecher-49609.html) · 56659 | 28 gemeinsame XP-Firmwarezeilen, an die dokumentierten Stände gebunden. | Eigene Projekt-/Auslieferungsfassung; aktuell noch angekündigt. |
| [Tr12 PKP PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-xp-sound-tr12-pkp-plux22-inkl.-lautsprecher-49615.html) · 56663 | 28 gemeinsame XP-Firmwarezeilen, an die dokumentierten Stände gebunden. | Eigene Projekt-/Auslieferungsfassung; aktuell noch angekündigt. |
| [BR 110 CSD PluX22](https://www.piko-shop.de/de/artikel/piko-smartdecoder-xp-sound-br-110-csd-plux22-inkl.-lautsprecher-51829.html) · 56666 | 28 gemeinsame XP-Firmwarezeilen, an die dokumentierten Stände gebunden. | Eigene Projekt-/Auslieferungsfassung; aktuell noch angekündigt. |

## Lenz: zehn genaue Modell-/Generationszuordnungen weiter ungeklärt

Die [Original-Identifikationstabelle](https://www.lenz-elektronik.de/media/e7/58/f9/1733842032/Decoderidenttabelle_01.pdf?ts=1733842032)
belegt Modellnamen, Artikel und Hardwaremerkmale. Das allein belegt keine CV-Tabelle.
Das [Original-Sammelhandbuch, 5. Auflage 01/2002](https://www.mjhobby.se/webarkiv/Lenz_Digital_Plus/dekoderinfo_02-02.pdf)
unterscheidet auf Seite 4 frühe Registergenerationen von späteren CV-Decodern. Seiten 14–16
verweisen bei optionalen und herstellerspezifischen CVs auf das jeweilige Modellhandbuch.
Daher wird seine gemeinsame Tabelle nicht pauschal auf ungeklärte Altmodelle übertragen.

| Modell / Artikel | Geprüfter Nachweis und verbleibende Grenze |
| --- | --- |
| LE075 · 10075 | Identtabelle S.2 und Sammelhandbuch S.4/7 belegen frühe Bauform und Generationsunterschiede. Keine genaue CV-Tabelle; Registerpositionen sind keine geratenen CV-Nummern. |
| LE1000 · 10100 | Identtabelle S.1 belegt LE1000. Die [Herstellerübersicht 24.11.2003](https://mefm.de/decoder/declist.pdf) nennt LE1000A; keine ausdrückliche Gleichsetzung der Fassungen. Kein Übernehmen von LE1000A-/Bachmann-Tabellen. |
| LE103X · 10115 | Identtabelle S.3 unterscheidet LE103X (40 mm) von LE103XF/104XF (38 mm) trotz gemeinsamer Artikelnummer. Keine passende X-Unterlage belegt. |
| LE103XFJST · 10117 | Identtabelle S.3 belegt JST-Modell. Die auffindbare spanische LE102XF-JST-Übersetzung belegt weder Hersteller-Autorisierung noch Gleichheit mit LE103XFJST. Keine Übertragung allein über Artikel/Stecker. |
| LE110 v1 · 10120 | [Originalanleitung 10/1997](https://www.mjhobby.se/webarkiv/Lenz_Digital_Plus/le110.pdf), PDF-S.1: LE110, Art.10120, 9-poliger Decoderanschluss und 26,5×17 mm; im Text Firmware 4.1. Identtabelle: v1 mit Litzen und 25×16 mm. Die Firmwarezahl allein ist kein Hardware-Gegenbeweis; der konkrete Hardwarewiderspruch bleibt aber ungelöst. |
| LE200 v1 · 10200 | Öffentliches Foto des [originalen Beiblatts 05/94](https://www.stummiforum.de/t231163f5-Lenz-LE-Manual-CV-Belegung.html), erster Beitrag: vier Grundeinstellungen, keine CV-Nummern. Keine sichere Zuordnung zu v1; LE230-Antworttabelle ist kein LE200-Beleg. |
| LE200 v2 · 10200 | Dasselbe Original-Beiblatt identifiziert LE200 ohne v1/v2-Unterscheidung. Keine sichere CV-Tabelle für v2; gleicher Artikel begründet keine Belegungsgleichheit. |
| LE911 · 10911 | Identtabelle S.4 belegt die Bauform. Keine genaue CV-/Protokollanleitung gefunden; weder Händlerbehauptungen noch LE900-/LE930-Unterlagen ersetzen den Nachweis. |
| LF100 · 10104 | Identtabelle S.5 unterscheidet LF100 und LF100XF; Herstellerübersicht 2003 bezeichnet XF als Nachfolger. Keine Übertragung von LF100XF/LF101XF derselben Artikelnummer. |
| SILVERmini+ V2 · 10310-03/10311-03 | [Exakte Hersteller-Produktseite](https://www.lenz-elektronik.de/en/Digital-Plus/Lokdecoder/Locomotive-decoder-in-detail/NEW-SILVERmini-V2/) beschreibt neue Funktionen ohne CV-Nummern. Aktuelle Downloadliste und CV-Editor-Datenpaket liefern keinen ausdrücklichen mini-V2-Nachweis; Plus-, PluX12- und andere V2-Modelle sind kein Ersatz. |

## Kühn, Uhlenbrock und Roco: fünf weitere genaue Tabellen fehlen

| Modell / Artikel | Geprüfter Nachweis und verbleibende Grenze |
| --- | --- |
| Kühn T62 · 82510/82520/82560/82570 | [Originalkatalog 2021](https://kuehn-digital.de/files/Katalog_2021_web.pdf), S.90–91, belegt T62 als T125-Überarbeitung mit zusätzlichen Ausgängen und RailCom. Keine Aussage identischer CVs; kein passendes Originalhandbuch auffindbar. T125/T65/N45 und Anschlussdaten aus Händlerseiten werden nicht übernommen. |
| Uhlenbrock IntelliSound 8 33250 · 33250/33254 | [Originalkatalog](https://www.uhlenbrock.de/de_DE/I14EE3E1-003.apd/Decoder%2BSound%2B2026.pdf) belegt e24-Modell. [Hersteller-Startseite](https://www.uhlenbrock.de/de_DE/index.htm) kennzeichnet e24 als in Vorbereitung; kein Eintrag der Artikel in der [Anleitungsübersicht](https://www.uhlenbrock.de/de_DE/service/download/handbook/de/INDEX.HTM). Keine Übertragung aus IntelliSound 6. |
| Uhlenbrock IntelliSound 8 34720 · 34720/34724 | Katalog belegt NEM652; Hersteller-Startseite nennt 34720 als in Vorbereitung. Keine passende Tabellenfassung in der öffentlichen Anleitungsübersicht. |
| Uhlenbrock IntelliSound 8 34760 · 34760/34764 | Katalog belegt PluX22; Hersteller-Startseite nennt 34760 als in Vorbereitung. Keine passende Tabellenfassung in der öffentlichen Anleitungsübersicht. |
| Roco 10734 | [Herstellerübersicht](https://www.roco.cc/rde/produkte.html?cat=183%2C194%2C252) belegt den Artikel. Keine passende 10734-Anleitung gefunden; direkter 10734-PDF-Pfad ohne Handbuch. Original 10735 benennt nur 10735; Steckerähnlichkeit oder vermuteter Kühn-OEM-Typ ist kein Gleichheitsnachweis. |

## Speicherung und Qualitätssicherung

Die sieben vorhandenen Referenzgruppen bleiben unter ihren stabilen IDs erhalten.
Die bisherigen ESU-/PIKO-Lückengruppen heißen sichtbar nun Firmwarehinweise und tragen
`partial`. Der historische interne Schlüssel `cv-gap-*` ist keine Aussage über den aktuellen Status.
Alle 26 Modellzuordnungen bleiben erhalten; keine Artikel werden umgehängt oder aus der
Statistik entfernt. Unveränderte gebündelte Vorgänger werden über exakte kanonische Prüfsummen
aktualisiert, eigene Metadaten und lokale Einstellungen bleiben geschützt.

Web-Firmwarehinweise erhalten Quelle, Abschnitt, Abrufdatum und Seite 0, keinen erfundenen
PDF-Hash. Die heruntergeladenen Lenz-PDFs behalten ihre tatsächlichen Prüfsummen, werden
aber wegen ungeklärter Modellanwendbarkeit nicht als CV-Beleg ausgegeben. Es entsteht kein
Schemawechsel, keine neue Firmwareerkennung und kein Geräte-Schreibzugriff.
