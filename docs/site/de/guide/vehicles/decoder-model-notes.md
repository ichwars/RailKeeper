---
title: Modell- und Firmwarehinweise
description: Modell- und Firmwarehinweise.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Modell- und Firmwarehinweise

<GuideFigure name="decoder-library" alt="Decoderbibliothek mit Modellwahl und Quellen." caption="Die Regeln gelten für den jeweils genannten Quellen- und Firmwareumfang." />

## Digitaldaten im Fahrzeugdialog

Unter **Modell → Digital** bleiben die manuelle Digitaladresse, DT-Adresse für Doppeltraktion,
Decodertyp und ABC-Fahrzeugangabe bearbeitbar. Daneben zeigt RailKeeper die aus dem gewählten
Decoderprofil ermittelten Werte mit CV-Quelle: aktive DCC-Adresse, Verbundadresse und Richtung,
ausschließlich freigegebene Verbundfunktionen (CV21/CV22), ABC und konstanten Bremsweg.
DT wird nicht aus CV19 übernommen. Fehlende Werte bleiben unbekannt; Profile und Indexseiten
werden für die Auswertung nicht vermischt. **Aus CVs übernehmen** kopiert bekannte Adresse,
erkannten Decodertyp und ABC-Status ins Formular; **Änderungen speichern** sichert diese Angaben.

Der Reiter **CV** bietet **Verständlich erklärt**, **Alle CV-Werte** und **Dateien & Import**.
Die Aktionen **CV-Liste importieren** und **Decoderdatei hochladen** sind getrennt. Eine originale
ESUX-Datei bleibt als Decoderprojekt hinterlegbar; vollständige Werte benötigen die CV-Liste.
Die gemeinsame Decoderübersicht zeigt Profil, aktive Adresse, Wertezahl und Indexseiten. Unter
**Verständlich erklärt** stehen Thema, Erklärung, Ergebnis und CV-Quelle nebeneinander; ein Klick
auf die Quelle öffnet rechts eine Seitenleiste für genau diesen CV. Sie zeigt Dezimal-, Hex- und
Binärwert, Indexseite, Decoderprofil und die tatsächlich gespeicherte Quelle. Nur beim Basis-CV29
werden dessen Bits erklärt. **CV-Wert bearbeiten** öffnet denselben Datensatz einschließlich
Indexseite im Editor. Beim CV-Import werden die Originalzeile (einschließlich Kopf- und Leerzeilen)
und der Exportzeitpunkt aus dem TXT-Dateikopf gespeichert und in der Seitenleiste angezeigt.
Die Uhrzeit bleibt ohne angenommene Zeitzone erhalten. JSON-Exporte und Backups enthalten diese
Angaben; Importe oder Bearbeitungen ohne Herkunftsangaben erhalten bereits gespeicherte Angaben.
Bei älteren Datensätzen ohne Quelle steht weiterhin „Nicht erfasst“; das Speicherdatum wird nicht
als Exportdatum verwendet. Backupformat 22 erhält die Herkunft, ältere Backups bleiben einlesbar.
Zusätzliche Verbund-/Decoderangaben sind eingeklappt erreichbar. Die ESU-Anzeige für DC-Analogbetrieb
berücksichtigt CV29 und CV50 gemeinsam. **Dateien & Import** erklärt die beiden Dateiarten und
zeigt den übernommenen Wertebestand sowie gespeicherte Originaldateien.
In der Rohwerttabelle sucht `CV29` genau nach CV29; auch Bedeutungen lassen sich suchen.
Profil-, Index- und Nullwertfilter grenzen die Liste ein. **CV-Werte exportieren** steht hier bereit. Erklärungen zu Basiswerten und den Bits von CV29 ergänzen die unveränderten Rohwerte.
Der TXT-Dateikopf bleibt als Quellenbeschreibung bei CV7 erhalten und liefert nach erneutem Import
die vollständige Firmwarekennung; CV7=255 wird nicht als Firmwareversion interpretiert.

Bei erkannten ESU LokPilot 5 (Herstellerkennung 151 und passendes Profil) werden einfache positive
Tastenzuordnungen aus den Mappingseiten im Reiter **Steuerung** angezeigt. Richtung, Ausgang und
CV-Quelle bleiben sichtbar. Vorschläge ergänzen fehlende Angaben im Entwurf. **Änderungen speichern** sichert die
geänderten Funktionen gemeinsam. Fehlende Felder und komplexe Bedingungen werden nicht ergänzt.
AUX3 im Ausgangsmodus 31 wird als PowerPack Control gekennzeichnet, ohne einen tatsächlichen
PowerPack-Einbau zu behaupten. Die reale Belegung anderer AUX-Ausgänge bleibt zu bestätigen.

Fachliche Grundlage: [ESU LokPilot 5, 8. Auflage](https://www.esu.eu/en/downloads/instruction-manuals/digital-decoders/?tx_esudownloads_pi1%5BdownloadItem%5D=0318b68e0cb850d9101947622d1041ea)
und [NMRA S-9.2.2](https://www.nmra.org/sites/default/files/s-9.2.2_2012_10.pdf).

Beim **ZIMO MN180N18** sind 68 Definitionsgruppen mit Fundstellen geprüft, einschließlich
Decoder-Sperre, PID-Regelung, Ausgangszuordnung, ABC-Pendeln und Bremswegregistern.
CV5=0/1 bedeutet jeweils die höchste interne Fahrstufe; CV6=1 wählt eine geglättete
Standardkennlinie. CV140 enthält eine Betriebsart (0/1/2/3/11/12/13), kein Richtungs-Bitfeld.
Die Fahrzeugwerte bleiben unverändert. Die Anleitung enthält auch Vorschauen auf SW6.00
und widersprüchliche Zeilen; daraus entstehen keine automatisch bestätigten Funktionen,
Schaltarten oder Invertierungen. Die Verbundadresse benötigt CV19 und CV20 gemeinsam. RailCom-Bit und Fahrtrichtung bleiben
getrennt; fehlt CV20, bleibt die Verbundadresse unbekannt. Freigaben F13–F27 werden aus dem
Gruppenbit ergänzt; Auto-Consist und das ungeklärte CV22-Bit 2 erscheinen getrennt davon.

Die Bremsübersicht zeigt die Betriebsart aus CV140 und richtungsgetrennte Wegparameter:
CV141 > 0 gilt für beide Richtungen; bei CV141 = 0 werden CV830/831 und CV832/833 kombiniert.
Fehlt ein Byte, bleibt diese Richtung unbekannt. Die Parameter sind keine gemessenen Meterwerte.
Die Fahrkurvenansicht erklärt die geglätteten ZIMO-Stützparameter mit unveränderten Rohwerten.
Eine 28-Punkt-Tabelle wird nur bei entsprechender Auswahl in CV29 gezeichnet. Der DC-Analogstatus
berücksichtigt CV12 und CV29 gemeinsam.

Beim **Lenz STANDARD+ V2** sind77 Definitionsgruppen mit genauen Fundstellen geprüft. Die
Auswertung trennt ABC-Bremsweg (CV51 Bits0/1) und Bremsweg bei Fahrstufe0 (Bit7). CV52 bleibt
ein Stellparameter ohne erfundene Meterangabe. DC-Anhalten benötigt CV51 Bit5 und deaktivierten
Analogbetrieb in CV29. Beide Pendelarten gleichzeitig bleiben ungeklärt.

Ausgänge A–D und ihre Tasten F0–F28 erscheinen mit Richtung und CV-Quelle. Die Belegung
von Bit4 bleibt ungeklärt; daraus wird kein Ausgang E angenommen. RailCom benötigt CV29
und freigegebene Kanäle in CV28. Motoroptionen, gespeicherte Fehlerflags, dezimale Lichteffekte
und Kupplungszeiten werden getrennt erklärt. Keine Tasten-Schaltart oder Invertierung wird
daraus übernommen. Die Fahrkurvenansicht zeigt die intern geglätteten Rohparameter oder
eine ausdrücklich gewählte CV67–94-Tabelle. Bereits gespeicherte Werte brauchen keinen Neuimport.


## Kühn N45: geprüfte Auswertung

Das N45-Profil umfasst **59 Definitionsgruppen** mit Fundstellen in der Herstelleranleitung,
einschließlich CV55 und CV60–63. Ausgewertet werden Motorfrequenz und Lastregelung,
ABC-/DC-Bremsoptionen, dezimale Dimm-/Kupplungsparameter und die Ausgangszuordnung von F0–F12.
Die Ausgangsbits verschieben sich ab F4 und F9. E/F sind Logikpegel und bei aktiviertem SUSI
gesperrt; ohne CV56 bleibt ihre Verfügbarkeit unbekannt. Schaltart, Invertierung und eingebaute
Verbraucher werden daraus nicht abgeleitet. Die Funktionsangaben am Fahrzeug bleiben erhalten.

Die Fahrkurve zeigt interne CV5/CV6-Sonderwerte als Rohparameter, ohne einen exakten Verlauf
zu erfinden. Nur die ausdrücklich ausgewählte gespeicherte Tabelle wird dargestellt. Laut Handbuch
gilt sie bei14/28 Fahrstufen, bei128 nicht; CV29 allein unterscheidet28/128 nicht.

Widersprüche der Anleitung (CV49 gegenüber CV51/52 und der CV112-Bereich gegenüber Bit5) sind
in den Erklärungen benannt. Die offizielle Weblesefassung wurde geprüft, ein PDF-Dateihash wird
wegen der nicht abrufbaren Datei nicht behauptet. Die Regeln gelten für N45/N45-P/N45-P8,
nicht pauschal für N45-18, N025, T62, T65, T125/T145 oder F060. Deren eigene Referenztabellen
bleiben getrennt. Vorhandene CVs werden ohne erneuten Import ausgewertet.

### Uhlenbrock IntelliDrive 76420

Für das genaue Modell **76420** mit CV8=85 werden die CV-Zuordnungen anhand der deutschen
Originalanleitung ausgewertet. Unter **Adresse & Betrieb** berücksichtigt DC-Analogbetrieb
CV29 Bit2 und CV51; ein gesetztes CV29 Bit3 wird nicht als RailCom ausgelegt. Die Verbundadresse
und ihre Richtungsumkehr stehen in CV19. Verbund-Funktionstasten bleiben mangels Nachweis unbekannt.

**Fahrverhalten**, **Bremsen** und **Ausgänge & Tasten** erklären die Motorregelung CV49/53–58,
die Bedingungen für Märklin-Gleichspannungsbremsen, das Bremsziel CV49/52 und die einheitlichen
Funktionsmasken CV33–42 für F0–F8. CV98 wählt zeitbegrenzte Ausgänge A1/A2, CV99 ihre Dauer
in 100-ms-Schritten. Diese Ausgangszeit belegt keine Momentfunktion oder Invertierung der Taste;
manuelle Funktionen bleiben erhalten. Der Lichtanschlusstausch CV49 Bit6 wird getrennt angezeigt.

Im Reiter **Fahrkurve** bleiben CV2/5/6 Rohparameter auf der Skala 1–63. Die gespeicherte
28-Punkt-Tabelle wird nur gezeichnet, wenn CV29 Bit4 sie auswählt. Die Geschwindigkeitskorrektur
steht in CV47/48; CV66 bezeichnet bei diesem Modell eine Motorola-Programmierseite. Fehlende
oder ungültige Messwerte werden nicht mit Werkseinstellungen ersetzt. Jede Erklärung verweist
auf die Original-CVs und die Herstellerfundstelle; Exportdatum und Originalzeile bleiben erhalten.

### Roco 10882 (Anleitung 2012, ZIMO)

Das geprüfte Profil benötigt die technische ZIMO-Kennung **CV8=145** und den passenden
Modellnamen. Die Roco-Markenkennung161 oder CV8 allein genügen nicht. Aktuelle ZIMO-MN-Regeln
werden nicht übertragen. Die Softwarebytes CV7/65 bleiben getrennte Angaben.

Unter **Bremsen** erscheinen die Freigabe aus CV29 Bit2=0 und CV124 Bit5=1 sowie die gespeicherte
Bremsstreckenart aus CV112 Bit6, mit der Bezeichnung der Roco-Anleitung. Die fehlerhafte
Bitnummer der dortigen Kurzfassung wird erklärt. ABC, konstanter Bremsweg und Verbundfunktionen
bleiben ohne passenden Nachweis unbekannt.

**Ausgänge & Tasten** berücksichtigt die wechselnden Mapping-Bitpositionen in CV33–46. Beim10882
sind nur Licht vorn/hinten und AUX1/AUX2 vorhanden; Zuordnungen zu anderen Ausgängen oder grauen
Tabellenfeldern werden ausdrücklich markiert. Dimmung CV60/114, der DELTA-Lichthinweis zu CV51 und
belegte Rangier-/Verzögerungstasten CV155/156 werden erklärt. Wert0 in CV155/156 bleibt ungeklärt.
Die Anzeige ändert keine manuelle Fahrzeugfunktion und leitet keine Schaltart oder Invertierung ab.

**Fahrkurve** zeichnet nur die in CV29 ausgewählten, gespeicherten und gültigen Punkte.
Die Quellenansicht trennt Originalzeile, Exportdatum und Rohwert von Werkseinstellung und
Herstellerfundstelle. Ein erneuter Import ist für die neue Auswertung nicht erforderlich.

### PIKO 56500 PluX22

Bei exakt zugeordnetem PIKO56500 werden das einfache Mapping (nur CV96=1), ABC,
konstanter Bremsweg und die Ausgangsparameter auf Indexseite18/0 erklärt. Der Bremsweg in
CV141 allein aktiviert keine Bremsung: CV145 wählt die Auslöser, ABC/DC benötigen zusätzlich
CV27; CV142 enthält eine separate Testtaste. Fehlende Voraussetzungen bleiben unbekannt.

Unter **Ausgänge & Tasten** stehen beide Parametersätze A/B für Licht vorn/hinten und A1–A7.
Die Auswertung trennt Effekt, PWM und automatische Abschaltung. Auto-Aus braucht Richtungs-
und Geschwindigkeitsflags. Bei Raucheffekten40/41 sind die Flags unbenutzt. Nicht vorhandene
Ausgänge werden gekennzeichnet. Erweitertes Mapping (CV96=6) erhält keine geratene einfache
Zuordnung. Verbraucher, Symbole, Schaltart und Invertierung bleiben unbekannt; manuelle
Funktionen bleiben erhalten. Der Quellenknopf öffnet den genauen gespeicherten Bankwert
mit Originalzeile und Exportdatum.

Die Herstellerunterlagen widersprechen sich bei CV101 und den Analogschwellen CV102–104.
Ohne eindeutigen Firmwarebezug bleiben die Bedeutungen, Einheiten und Werkswerte ungeklärt.
CV101 wird nicht zugeordnet. Quellen und Abweichungen sind in der Bibliothek hinterlegt.
Die Fahrkurve zeigt nur die in CV29 gewählte Kennlinie, ohne auf eine inaktive auszuweichen.

### D&H DH05C und DH10C

DH05C und DH10C besitzen eigene geprüfte Profile ab Firmware3.15. CV8=97 zusammen mit der
unindizierten CV261 erkennt DH05C bei52 und DH10C bei102. Ein Indexwert gleicher Nummer
bestätigt den Decoder nicht. Die Auswertung nutzt die D&H-Motorkennlinie CV48 statt einer
geratenen DCC-3-Punkt-Kurve und zeigt den Bremsrampenstatus aus CV154/155.

Der Typcode unterscheidet keine Hardwaregeneration. Angaben zu AUX3–6 und SUSI sind deshalb
mit den Grenzen aus der Anleitung versehen. Besonders CV137 hat hier Werkseinstellung0,
beim DH18A dagegen1. Ausgangsmasken, Abschalttimer und SUSI-Richtungsumkehr belegen weder
Schaltart noch Invertierung einer Funktionstaste. Eigene Funktionsangaben bleiben erhalten.
Adress- und Verbundregister haben einen direkten Herstellerbeleg; Rohwerte und Herkunft bleiben
unverändert. Eine unbekannte Firmware wird nicht aus beliebigen CV7-Werten geraten.

### D&H DH16A

Der DH16A hat ein eigenes geprüftes Profil mit120 CV-Definitionen ab Firmware3.15.
CV8=97 und die unindizierte CV261=160 erkennen den Modelltyp auch ohne Profilnamen im Export.
Die Hardwaregeneration bleibt dabei offen. Adresse, Verbund, ABC und Bremsrampe werden aus
den gespeicherten Werten ausgewertet; CV48 bleibt ein Kennlinienparameter ohne geratene Kurvenpunkte.

Laut Hersteller sind AUX3/AUX4 in beiden Generationen verstärkt. AUX5/AUX6 sind Logikausgänge,
deren Anschlüsse mit SUSI geteilt werden. Deshalb CV137 und die tatsächliche Verdrahtung beachten.
Die dokumentierte CV137-Werkseinstellung ist0; die DH18A-Sondervorgabe1 wird nicht übernommen.
Ausgangsbedingungen und SUSI-Richtungsumkehr belegen keine Funktionstasten-Invertierung.
Manuelle Angaben sowie Originalzeile und Exportdatum bleiben erhalten.


### D&H DH12A (älterer PluX12-Decoder)

Der abgekündigte DH12A hat ein eigenes geprüftes Profil mit120 CV-Definitionen ab Firmware3.15.
CV8=97 und die unindizierte CV261=120 erkennen den Modelltyp auch ohne Profilnamen im Export.
Eine unbekannte Firmware wird nicht geraten. Adresse, Verbund, ABC und Bremsrampe werden aus den
vorhandenen Werten ausgewertet; CV48 bleibt ein Kennlinienparameter ohne erfundene Kurvenpunkte.

AUX3–6 sind beim DH12A unverstärkte Logikausgänge. AUX3/AUX4 haben eigene Anschlüsse;
ZCLK/ZDAT dienen alternativ SUSI oder Logikausgängen. CV137 Bits0/4/5 gemeinsam und die tatsächliche
Verdrahtung beachten. Die dokumentierte Werkseinstellung0 ist kein ausgelesener Fahrzeugwert.
Ausgangstimer, Bedingungen und SUSI-Richtungsumkehr belegen weder Schaltart noch Invertierung einer
Funktionstaste. Manuelle Angaben sowie Originalzeile und Exportdatum bleiben erhalten.


### D&H DH14B (mTc14)

Der DH14B hat ein eigenes geprüftes Profil mit120 CV-Definitionen ab Firmware3.15. CV8=97 und
die unindizierte CV261=141 bestätigen das Modell. Adresse, Verbund, ABC und Bremsrampe werden aus
den gespeicherten Werten ausgewertet; CV48 bleibt ein Kennlinienparameter ohne geratene Kurvenpunkte.

AUX3/AUX4 sind unverstärkt und teilen ZCLK/ZDAT mit SUSI. Die CVs für AUX5/AUX6 beschreiben
alternative logische Belegungen derselben Pins, keine zusätzlichen Anschlüsse. CV137 Bits0/4/5
müssen zusammen betrachtet werden; Bit0 allein reicht für den SUSI-Status nicht aus. Die Bibliothek
zeigt die Tabelle der belegten Kombinationen und nennt die Hardwaregrenzen direkt bei Ausgangs-CVs.
Schaltart, Tasteninvertierung und angeschlossene Verbraucher werden daraus nicht abgeleitet.
Manuelle Angaben sowie Originalzeile und Exportdatum bleiben erhalten.


### D&H DH21A (Generationen und Anschlussvarianten)

CV8=97 und die unindizierte CV261=200 erkennen den DH21A. Das eigene Profil enthält121 belegte
Registerdefinitionen ab Firmware3.15, darunter die teilweise geklärte CV144. Die Bezeichnungen
DH21A-0/-2/-3/-4/-5 wählen dasselbe Profil; daraus wird keine Hardwaregeneration abgeleitet.

AUX3/AUX4 sind bei -4 in beiden Generationen verstärkt, bei -5 dagegen Logikausgänge. Die Angaben
zu AUX5/AUX6 in der Anleitung widersprechen sich teilweise. Diese Pegel bleiben ausdrücklich
ungeklärt. CV137 und die Ausgangs-CVs nennen die Varianten und Grenzen direkt; keine angeschlossenen
Verbraucher oder Funktionstasten-Schaltart daraus ableiten.

Bei CV144 sind nur die übereinstimmend belegten Bits1–5 erklärt; Bit0 bleibt offen. GPIO/AUX8 ist
nur für Generation2 dokumentiert. Gespeicherte GPIO-Konfiguration bedeutet keine erkannte
Bremsverdrahtung. Auch die dokumentierte AC-Analogfähigkeit sagt nichts über die aktuell verwendete
Versorgung aus. Originalzeile, Exportdatum und manuelle Angaben bleiben unverändert.


### Weitere D&H-Modelle und Herstellerhinweise

DH21B, DH22A, DH22B, DH24A und DHSP10A besitzen eigene geprüfte Profile ab Firmware3.15.
CV8=97 und die unindizierte CV261=201/202/203/204/110 erkennen das jeweilige Modell.
Anschlussvariante, Generation, Firmware und tatsächlich angeschlossene Verbraucher bleiben offen.
AUX11/12 teilen bei DH21B/DH22B/DH24A die SUSI-Pins; die allgemeine CV137-Bit5-Auswahl gilt dort
nicht. CV101 wird nur beim DH24A mit seinem dokumentierten AUX10 erklärt. Ausgangsbedingungen
und Tastenzuordnungen belegen keine Schaltart oder Invertierung einer Funktionstaste.

Für weitere eindeutig benannte oder erkannte Modelle ergänzt RailKeeper die Werteliste und
CV-Detailansicht um **Herstellerhinweise** aus den hinterlegten CV-Tabellen. Die Bezeichnung ist
eine Handbuchreferenz und wird als solche markiert. Modell-/Firmwarebedingungen und Quelle mit
Seite stehen in der Detailansicht; Dokumentangaben ab Werk bleiben von ausgelesenen Werten
getrennt. Diese Hinweise aktivieren keine automatische Auswertung und ändern keine manuellen
Angaben. Geprüfte Erklärungen und eigene Beschreibungen behalten Vorrang. Unbekannte Indexseiten,
widersprüchliche Zuordnungen und fehlende Unterlagen bleiben offen. Ein erneuter CV-Import ist
für diese Anzeige nicht erforderlich.
