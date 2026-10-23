---
title: Technische Referenz für Decoder und CVs
description: Digitalfunktionen zuordnen, Fahrkurven prüfen, CV-Werte pflegen und Decoder-Dateien speichern.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Technische Referenz für Decoder und CVs

1. Öffne **Bearbeiten > CV > Dateien & Import**. Importiere deine exportierte CV-TXT; eine Decoderprojektdatei wird getrennt abgelegt.
2. Prüfe unter **Verständlich erklärt** Adresse, Fahrverhalten und Bremsen. **Alle CV-Werte** zeigt die vollständigen Rohwerte mit Indexseiten.
3. Öffne **Steuerung**, prüfe die angebotenen Ergänzungen und Symbole und speichere gemeinsam mit **Änderungen speichern**. Manuelle Angaben bleiben geschützt.

Du brauchst nicht jeden Rohwert zu verstehen. Eine Erklärung erscheint nur, wenn Modell, Firmware und Adressierung dazu passen. [Decoder nachschlagen](decoder-library).

[CV-Import mit zwölf Beispielwerten, Indexseiten und Herkunft ausprobieren](cv-import-example).


<GuideFigure name="cv-analysis" alt="Gespeicherte CVs werden verständlich erklärt; technische Details öffnen sich separat." caption="Gespeicherte CVs werden verständlich erklärt; technische Details öffnen sich separat. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

RailKeeper speichert Funktionszuordnungen, Fahrkurvendaten, CV-Werte und Decoder-Projektdateien am
Fahrzeug. Der zusammengehörige Ablauf befindet sich in **Steuerung**, **Fahrkurve** und **CV**.

## Decoderbibliothek

Unter **Einstellungen > Daten > Allgemein > Decoderbibliothek** führt die Navigation über Hersteller,
Familie und Modell. Erfasst sind ESU, ZIMO, Doehler & Haass, Lenz, Kühn, Uhlenbrock, Roco und PIKO.
Eigene CV-Regeln liegen unter anderem für **LokPilot 5 micro DCC**, **MN180N18**, **DH18A**, **STANDARD+ V2**,
**N45**, **76420**, **10882** und **56500** vor; deren Modell-, OEM- und Firmwaregrenzen bleiben
maßgeblich. Weitere Modelle sind mit Stammdaten und Quellen dokumentiert. Die DCC-Grundlagen stehen separat bereit.
Unter **CV-Definitionen** findest du CV-Nummern, Indexseiten, Wertebereiche und Bitbedeutungen;
die Suche akzeptiert `CV 29`, `291` oder `ABC`. **Erkennung** erklärt benötigte Angaben,
**Quellen** zeigt Handbuchstand und Geltungsbereich. Admin und Editor können Profile für die
Auswertung aktivieren oder deaktivieren; Viewer können sie lesen. Bestehende Werte ändern sich
dadurch nicht. [Aufbau der Bibliothek](/de/administration/master-data-general#decoderbibliothek).

Die Importvorschau sowie **Modell > Digital**, **Steuerung** und **CV** verwenden dieselben
Definitionen. Die Erkennung nutzt die Herstellerkennung aus der unindizierten CV8 und entweder
einen passenden Profilnamen oder eine dokumentierte Typkennung. Beispiel: CV8=97 und CV261=180
erkennen DH18A; CV8=99 und CV7=77 oder 79 erkennen STANDARD+ V2. CV7 ist bei Lenz eine Typkennung,
keine vollständige Firmwareversion. ESU benötigt weiterhin den passenden Profilnamen.

CV8=145 und CV250=127 bestätigen den ZIMO-Grundtyp MN180, aber nicht den Anschluss N18.
Modellspezifische Regeln für MN180N18 benötigen zusätzlich den eindeutigen Profilnamen. Erkannte
Modelle ohne eigene geprüfte CV-Regeln verwenden ausschließlich die DCC-Grundlagen.
Herstellerkennungen allein bestätigen kein Modell; indizierte Werte dienen nie als Typnachweis.
Fehlende oder widersprüchliche Angaben bleiben sichtbar. Bei Konflikten werden ausschließlich die
aktiven DCC-Grundlagen verwendet. Andere Protokolle erhalten keine DCC-Deutung.

Motorregelung, Lichthelligkeit und Bremsregister werden nach dem ausgewählten Modell erklärt.
Beispielsweise ist CV51 bei Lenz ein Bremsregister, bei D&H eine Anschlussvertauschung und bei ESU
ein Motorparameter. **Aus CVs übernehmen** überträgt erkannte Fahrzeugangaben erst nach deinem Klick
ins Formular; gespeichert werden sie mit **Änderungen speichern**.

Die Quellenbereiche beginnen bei ESU mit Firmware **5.3.128**, bei DH18A mit **3.15** und bei
MN180N18 mit **5.29**. Für STANDARD+ V2 ist die zugehörige Anleitung maßgeblich. Eine erkannte ältere
Firmware schaltet die herstellerspezifische Deutung aus. Fehlende Firmware bleibt als Unsicherheit
gekennzeichnet; eine neuere Version ist keine Zusicherung identischer Eigenschaften.

Beim **DH18A** sind Ausgangszuordnungen, Ausschaltmasken, Pendelbetrieb und Ausgangsbedingungen
getrennt erklärt. Abschalttimer belegen keine Momentfunktion; die SUSI-Richtungsumkehr belegt
keine Invertierung einer Funktionstaste. CV137 beeinflusst die Bedeutung der Mapping-Bits.
Die Fahrkurvenansicht zeigt den Kennlinienparameter **CV48**, statt eine unbelegte DCC-Kurve zu
zeichnen. Die Bremsrampe wird anhand **CV154/155** angezeigt; die Rohwerte bleiben vollständig
im Reiter **CV** zugänglich. Fehlende Werte bleiben unbekannt.

Die Abdeckungszahl zählt sowohl einzelne CV-Definitionen als auch dokumentierte Speicherbereiche.
Eine Zuordnung zu „Mapping-Bedingungen“ erklärt noch nicht jedes Byte vollständig. Unbekannte CVs
bleiben ohne geratene Bedeutung. Eine AUX-Zuordnung sagt nicht, welche Funktion in der Lok tatsächlich
angeschlossen ist; Vorschläge werden weiterhin ausdrücklich übernommen.

Die CV-Detailansicht nennt den Beleg der ausgewählten Erklärung mit Quellenstelle. Dokumentierte
Einheiten und **Dokumentangabe ab Werk** stehen getrennt vom ausgelesenen Wert; sie ergänzen oder
ersetzen keine Fahrzeugwerte. Beim LokPilot 5 micro DCC erscheinen CV57/58, CV141–143 und CV250
als **Reserviert laut ESU-Handbuch**. Sie erhöhen die Zahl funktional zugeordneter Werte nicht.
Ein gesetztes Mapping-Bit bedeutet eine Zuordnung zu dieser Zeile, keinen aktuellen Schaltzustand.

Beim **LokPilot 5 micro DCC** unterscheiden die Erklärungen Motor-PWM, Trimmwerte, EMK-Messung,
Bremsfunktionen, Analogparameter, die Ausgangskonfiguration auf **CV31=16/CV32=0** und die acht
Zufallsfunktionen auf **CV31=16/CV32=13**. Mapping-Bedingungen und -Aktionen erhalten eigene
Spaltenbedeutungen. Ein angezeigter CV-Bereich mit **Abstand 16** umfasst nur jede 16. Adresse,
nicht die Lücken dazwischen. Ein Ausgangsparameter kann je nach Betriebsart Helligkeit oder eine
Servozeit bedeuten; eine konfigurierte AUX-Nummer belegt keine angeschlossene Fahrzeugfunktion.

Neuere Register nennen ihren eigenen Firmwarebeginn. Bekannte ältere Firmware erhält daraus
keine Deutung: CV114/115/131 ab **5.4.139**, die dokumentierte SUSI-Option in CV122 ab **5.6.142**
und CV199 ab **5.8.156**. Grundlage sind die
[ESU-Anleitung, 8. Auflage](https://www.esu.eu/download/betriebsanleitungen/digitaldecoder/?tx_esudownloads_pi1%5BdownloadItem%5D=803bc0046a0244b6416c82c6abfd385e)
und die [LokProgrammer-Versionshinweise](https://www.esu.eu/en/downloads/software/lokprogrammer/).
CV11 beschreibt ab **5.14.183** die Wartezeit bei ausbleibendem DCC-Datenempfang; eine vollständige
Werteskala ist in den [deutschen Versionshinweisen](https://www.esu.eu/download/software/lokprogrammer/)
nicht belegt. CV245 beschreibt die Wartezeit zwischen Andrücken und Abrücken beim Kupplungswalzer,
belegt in der [ESU-Anleitung 41002, Abschnitt 7.4, Seite 4](https://esu.eu/download/betriebsanleitungen/zubehoer/?tx_esudownloads_pi1%5BdownloadItem%5D=db7553a3b89dceb341d13f510ce0bb5e).
Beide Anzeigen verwenden den Rohbytebereich und rechnen Werte nicht ohne gesicherten Beleg in Sekunden um.
Unveröffentlichte Indexseiten und widersprüchliche Tabellenadressen bleiben offen. Dazu gehören
CV272 auf Seite 16/0 sowie die ersten G-Adressen der Mapping-Seiten und einzelne Q-Adressen;
offensichtliche Druckfehler werden nicht stillschweigend als gesicherte Werte übernommen.

Die Bibliothek wird lokal gespeichert und im Stammdaten-JSON sowie in der Anwendungssicherung
mitgeführt. Sie ist von den ausgelesenen Fahrzeugwerten getrennt. CV-Wert, Decoderprofil, Indexpaar,
Originalzeile und Exportdatum bleiben unverändert. Es erfolgt kein Schreiben an eine Digitalzentrale
oder einen Decoder.

## Voraussetzungen und Zugriffsrechte

Öffne ein Fahrzeug im **Fahrzeugbestand**, wähle **Bearbeiten** und dann den benötigten Tab.
Allgemeine Felder wie **Digital**, Decodernummer, Decodertyp und ABC-Bremsung erklärt
[Fahrzeugbestand und Grunddaten](/de/guide/vehicles/).

Ein Fahrzeug muss normalerweise einmal gespeichert sein, bevor RailKeeper Funktionen, CV-Werte
oder Dateien speichern kann. Ein ungespeicherter ECoS-Entwurf kann CVs und eine abgeleitete Kurve
anzeigen. Seine normalen Schreibaktionen bleiben jedoch deaktiviert.

Admin, Editor, Viewer können gespeicherte Decoderdaten ansehen. Auf Viewer-Ebene sind
auch Funktions- und CV-Export sowie der Download von Decoder-Dateien möglich. Nur Admin und Editor
dürfen Daten speichern, importieren, übernehmen, hochladen oder löschen. Der Server erzwingt diese
Grenze.

::: warning Andere Änderungen zuerst speichern
Schreibaktionen an CVs und Decoder-Dateien laden das vollständige ausgewählte Fahrzeug neu.
Speichere oder verwirf bewusst ausstehende Grunddaten, Funktionsänderungen, Bildmetadaten und
Änderungen anderer Tabs vor einer Schreibaktion.
:::

## Digitalfunktionen F0-F31 zuordnen

<GuideFigure name="functions" alt="Steuerung: Funktionstasten gemeinsam prüfen und speichern." caption="Steuerung: Funktionstasten gemeinsam prüfen und speichern. Künstliche Demodaten. Anklicken für die volle Größe." />

Öffne **Steuerung**. **Funktionstasten** ist die gemeinsame Bearbeitungsliste. **Nur belegte** ist
standardmäßig eingeschaltet und zeigt gespeicherte, manuell angelegte und zuverlässig aus dem
CV-Mapping erkannte Tasten. Die Übersicht unterscheidet erkannte und gespeicherte Funktionen.
Schalte den Filter aus, um eine weitere Taste von F0 bis F31 hinzuzufügen. Mit dem Plus-Symbol
wird auch eine unveränderte Vorgabe ausdrücklich in den Entwurf aufgenommen.

Jede Zeile enthält **Funktionsname**, **Symbol**, **Schaltart** und **Löschen**. Fehlende Symbole sind
mit Text und einer dezenten Amber-Markierung gekennzeichnet; ein Klick öffnet die Symbolauswahl.
Ein vorhandener eigener Name bleibt beim Symbolwechsel erhalten. Nur eine leere Bezeichnung wird
mit der Symbolbezeichnung gefüllt. Der Funktionstyp wird aus dem Symbol abgeleitet.

Unter **Details** liegen **Richtungsabhängig**, **Invertiert**, **Eigene Notiz** und die aufklappbare
**Herkunft und CV-Werte**. Bereits gespeicherte technische Nachweise erscheinen dort vollständig
und bleiben bei Änderungen an der eigenen Notiz erhalten. Historische Notizen werden beim Öffnen
nicht umgeschrieben. Aktuelle CV-Werte dienen getrennt davon zum Vergleich.

**Fehlende Angaben aus CVs ergänzen** führt passende Symbole, belegbare Bezeichnungen und für das
geprüfte Profil Schaltart/Invertierung in einem Schritt im Entwurf zusammen. Bekannte Angaben und
manuell gewählte Schaltarten bleiben geschützt. AUX-Nummern erhalten keine geratenen Fahrzeugnamen.
Die **Funktionszuordnung aus CVs** steht am Ende eingeklappt als technische Nachschlageansicht.

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
angezeigt. **Richtungsabhängig** beschreibt eine Abhängigkeit von der Fahrtrichtung;
**Invertiert** ist eine davon unabhängige Angabe zur Funktionslogik. Beide starten mit **Unbekannt**.
Nur eine ausdrücklich gewählte Antwort wird als Ja oder Nein gespeichert. Richtungsabhängigkeit wird daraus nicht automatisch bestätigt. Schaltart und Invertierung
können beim ausdrücklich geprüften Profil mit Versionsnachweis ergänzt werden; siehe
[Funktionsübung](functions-walkthrough).

Die frühere Checkbox war uneindeutig beschriftet. Bestehende Werte bleiben als **Altangabe** sichtbar;
beide getrennten Felder werden zunächst unbekannt. Namen, Symbole, Betriebsart und Notizen bleiben
erhalten. Auch das Speichern einer anderen Funktionseigenschaft deutet die Altangabe nicht um.

F0 bietet beim Anlegen den Namen `Fahrlicht`, das Lichtsymbol und den Typ `licht` als Vorgabe.
Der unberührte F0-Standard allein zählt nicht als belegte Funktion. Neue Zeilen verwenden zunächst
`dauer`; **Dauer (Vorgabe)** unterscheidet diesen Startwert von einer gewählten oder übernommenen
Schaltart.

Alle Änderungen, Ergänzungen und Löschungen bleiben im Entwurf. **Änderungen speichern** sichert
Fahrzeugdaten und geänderte Funktionen gemeinsam. Eine Löschung lässt sich vorher zurücknehmen;
**Abbrechen** verwirft ungespeicherte Funktionsänderungen. Beim Anlegen aus einem ECoS-Entwurf werden Fahrzeug und übernommene Funktionen ebenfalls
gemeinsam gespeichert. Wurde eine Funktion zwischenzeitlich anderweitig geändert,
wird nichts aus diesem Speichervorgang übernommen; der Entwurf bleibt zum Vergleichen erhalten.

Der Server akzeptiert F0–F31, bekannte Typen und Betriebsarten, Namen bis 120 UTF-8-Bytes,
Symbolschlüssel bis 80 UTF-8-Bytes und Notizen einschließlich Herkunft bis 1.000 UTF-8-Bytes.

### Funktionen importieren und exportieren

**Export** lädt `<inventarnummer>-funktionen.json` herunter. Ohne Inventarnummer lautet der Name
`railkeeper-funktionen.json`. Die Datei enthält Inventarnummer, Fahrzeugname, Decodernummer und alle
belegten Zuordnungen. Die Decodernummer verwendet zuerst die primäre digitale Nummer und ersatzweise
die DT-Decodernummer. Der Export verwendet die aktuellen Zeilen einschließlich ungespeicherter
lokaler Funktionsänderungen, speichert sie aber nicht in RailKeeper.

**Import** liest die erste ausgewählte JSON-Datei als Array oder als Array unter `functions`
beziehungsweise `functionMappings` in den Bearbeitungsentwurf. Ungültige Einträge oder doppelte
Tasten blockieren die gesamte Übernahme; der bisherige Entwurf bleibt erhalten. Prüfe die Zeilen
und wähle **Änderungen speichern**. Bis dahin kann **Abbrechen** auch den Import verwerfen.

Neue Funktions-JSON-Exporte kennzeichnen die getrennten Angaben mit `flagsVersion: 2`.
`directionDependent` und `inverted` enthalten unabhängig `true`, `false` oder `null` (unbekannt).
Eine vorhandene `legacyDirectionDependent` bleibt eine uneindeutige Referenz. Ältere Dateien ohne
Versionskennzeichnung bestätigen keine der beiden Angaben und überschreiben bereits bestätigte
Werte nicht. App-Backups ab Version 24 erhalten alle drei Zustände; beim Import älterer Backups
werden frühere Werte ebenfalls als Altangabe behandelt.

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

<GuideFigure name="cv-details" alt="Die Seitenansicht zeigt Rohwert, Indexseite und gespeicherte Herkunft." caption="Die Seitenansicht zeigt Rohwert, Indexseite und gespeicherte Herkunft. Künstliche Demodaten. Anklicken für die volle Größe." />

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

RailKeeper identifiziert eine CV-Zeile durch Fahrzeug, getrimmtes Decoderprofil (einschließlich
Groß-/Kleinschreibung), CV-Nummer und das optionale Paar **CV31/CV32**. Das Protokoll gehört nicht
zur Identität. Basiswerte haben keine Indexangabe; Seite `0/0` ist davon verschieden. Für CVs
257–512 können beide Indexfelder gemeinsam Werte von 0 bis 255 enthalten. Bestehende Werte ohne
Index bleiben unverändert erhalten. Bei mehreren Decodern am selben Fahrzeug unterschiedliche
Profilbezeichnungen verwenden, auch wenn es derselbe Decodertyp ist.

**CV hinzufügen** aktualisiert einen passenden Wert. Bearbeiten und Löschen wirken sofort;
Löschen besitzt keine zusätzliche Bestätigung. Nur eine Änderung des Zahlenwerts erzeugt einen
Historieneintrag. Die Tabelle zeigt die fünf neuesten Änderungen je Wert.

Die Wertetabelle zeigt Indexseiten ausdrücklich, lässt sich nach Decoderprofil, Seite und Text
filtern und zeigt jeweils 50 Zeilen. Die Übersicht **Wichtige Decoderwerte** zeigt für ein
gewähltes Profil die aus den gespeicherten Basis-CVs abgeleitete DCC-Adresse, Herstellerkennung,
CV29 mit Bits sowie Analog- und RailCom-Freigabe. Sie meldet keine Live-Zustände und ändert weder
Fahrzeugadresse noch Funktionsbezeichnungen. Bei mehreren Profilen zuerst eines im Filter wählen.

## RailComPlus-Anmeldename und Tastensymbole

Für den geprüften **ESU LokPilot 5 micro DCC** zeigt **Modell → Digital** den RailComPlus-Anmeldenamen
aus Seite `1/0` an. Dieser Name ist im Decoder änderbar und kein Typnachweis. Der Fahrzeugname
bleibt erhalten. Die Auswertung unterstützt den geprüften ASCII-Bereich CV261–288 mit Nullabschluss;
fehlende oder doppelte Bytes und ungeprüfte Zeichenkodierungen erhalten einen Hinweis.

Unter **Steuerung** erscheinen die Symbolvorschläge aus Seite `1/1` mit Taste, Symbolcode,
Indexseite, Rohwerten und vorhandener Originalzeile bzw. Exportdatum. Bei der geprüften Liste
sind das F0 Stirnbeleuchtung, F1/F2/F5/F6 Allgemeine Funktion, F3 Rangiergang und F4 ABV.
Die 25 nicht zugeordneten Tastenbeschreibungen F7–F31 werden nicht als Funktionen angelegt.
Die Mappingauswertung daneben beschreibt die tatsächlichen Ausgangs- und Logikzuordnungen;
ein Symbol allein bestätigt keine angeschlossene Lampe oder Kupplung.

Die gemeinsame Aktion **Fehlende Angaben aus CVs ergänzen** füllt im Bearbeitungsentwurf ein
fehlendes Symbol nur bei genau einem passenden aktiven Stammdatensymbol. Eigene Namen, vorhandene
Symbole und bekannte Schaltangaben bleiben erhalten. Anschließend sichert **Änderungen speichern**
den gesamten Entwurf. Reine Lesekonten können die Details ansehen.

Die Byteaufteilung folgt einer [empirischen JMRI-Vergleichsdefinition](https://github.com/JMRI/JMRI/blob/a9e4d386488cb79c0811d479aad3d75b0b33b068/xml/decoders/esu/v4fnDescription.xsl).
Für Symbolnamen gilt die [ESU-Symbolübersicht](https://esu.eu/download/betriebsanleitungen/digitalsysteme/?tx_esudownloads_pi1%5BdownloadItem%5D=014a5a9436295b8c1130b6eafa0cb993).
Für **LokPilot 5 micro DCC mit Firmware 5.14.183** ergänzt die gemeinsame Aktion
**Fehlende Angaben aus CVs ergänzen** auch diese Angaben im Bearbeitungsentwurf. Je vier Exporte
für **F0 (Licht)**, **F1 (Physikalisch)** und **F3 (Logisch)** aus LokProgrammer 5.2.18 bestätigen Momentbetrieb
in Bit 7 und Invertierung in Bit 4 des ersten Deskriptorbytes auf Seite `1/1`.
Bei F0 nimmt CV257 die Werte 0, 128, 144 und 16 an; CV258 bleibt 3.
Bei F3 nimmt CV263 die Werte 96, 224, 240 und 112 an; CV264 bleibt 9. Kategorie und Symbol bleiben
von beiden Einstellungen unabhängig. Hinweis und neue Herkunftsnotizen nennen alle drei Vergleiche.
Bereits gespeicherte Notizen bleiben erhalten. Die anderen Tasten folgen demselben JMRI-Deskriptorformat;
sie wurden nicht einzeln mit positiven Vergleichsexporten geprüft.

Gespeicherte oder im Formular gewählte Schaltarten bleiben erhalten. Eine neue, noch nicht gewählte
Standardeinstellung darf ergänzt werden. Die Invertierung wird ausschließlich bei **Unbekannt**
gefüllt; sowohl **Ja** als auch **Nein** bleiben geschützt. Namen, Symbole und Richtungsabhängigkeit
bleiben erhalten. Die Notizen dokumentieren ausschließlich die übernommenen Felder, beide CV-Adressen
und Rohwerte, vorhandene Originalzeilen und das Exportdatum sowie die geprüfte Decoder-/Softwareversion.
Speichere anschließend mit **Änderungen speichern**. Bereits gespeicherte CVs benötigen keinen Neuimport.

Bei anderen oder unbekannten Firmwareständen werden diese Flags weiterhin nur zur Gegenprüfung
angezeigt und nicht übernommen. Unvollständige Bytepaare,
widersprüchliche bekannte Exportdaten und nicht passende Decoderprofile bleiben ohne Deutung.

Beide Importvorschauen verwenden dieselbe Auswertung unter **RailComPlus-Daten der Importdatei
prüfen**. Sie bezieht sich auf gültige Dateizeilen, unabhängig von der Auswahl zum Speichern.
Nach dem CV-Import stehen die Symbolvorschläge unter Steuerung zur Verfügung. Bereits gespeicherte
CVs werden beim Öffnen ausgewertet; ein erneuter Import ist nicht nötig.

## Dateien für einen sauberen Import

Für Lokomotiven und Wagen mit Funktionsdecoder gilt derselbe Ablauf. Die Zuordnung erfolgt immer
am ausgewählten, bereits gespeicherten Fahrzeug.

| Datei | Erforderlich | Verarbeitung und Anzeige |
| --- | --- | --- |
| Vollständige LokProgrammer-CV-Liste als TXT | Für den vollständigen ESU-CV-Import | Alle Werte mit Indexseiten in Vorschau und CV-Tabelle |
| CSV oder RailKeeper-CV-JSON | Alternative zur TXT-Liste | Strukturierte Werte, Profile, optionale Indexfelder |
| ESUX / ESU | Optional | Originalprojekt unter CV-Dateien; sichere Metadatenvorschau. Der proprietäre CV-Inhalt wird nicht dekodiert |
| Kopierter Text der Decoderinformationen | Optional | Fahrzeugbeilage für Seriennummer, Firmware und Rohkennungen; kein automatischer Identitätsimport |
| Screenshot, Foto, Anleitung | Optional | Fahrzeugbeilagen bzw. Bilder als prüfbare Nachweise; keine automatische OCR oder CV-Erkennung |
| Funktions-JSON | Optional | Namen, Symbole und Schaltarten über den getrennten Funktionsimport |

Eine ESUX-Datei ist keine Voraussetzung für den TXT-Import. Projektname und Typ wie `steam`
identifizieren die wirkliche Lok nicht zuverlässig. Eine rohe CV-Liste ersetzt auch keine bestätigte
Zuordnung von AUX-Ausgängen zu tatsächlich angeschlossenen Lampen oder Kupplungen.

## CV-Werte importieren und exportieren

**Import** liest die erste ausgewählte Datei: TXT/CSV bis 2 MiB, CV-JSON bis 32 MiB,
jeweils höchstens 10.000 Werte. Die höhere lokale JSON-Grenze berücksichtigt exportierte
Datensatzkennungen, Zeitstempel und Historien. Der Server erhält nur die importierbaren Felder;
für die atomare Importanfrage gilt weiterhin die Grenze von 8 MiB.

- LokProgrammer-TXT wird mit `Decoder:`, `CV 001 = 003` und
  `Index: 4099 (CV31=16, CV32=3)` vollständig verarbeitet. Der Decoderkopf liefert das Profil;
  eine Firmware-Endung wird davon getrennt. Fehlerhafte oder widersprüchliche Indexzeilen blockieren
  die gesamte Dateivorschau, statt eine Teilmenge zu liefern. CVs oberhalb 256 benötigen in diesem
  ESU-Format eine Indexseite. Unindizierte höhere CVs anderer Hersteller, etwa D&H CV261, werden
  über strukturierte CSV- oder JSON-Listen importiert.
- JSON darf ein Array oder ein Objekt mit `cvValues` sein. Optionale Felder heißen `indexCV31`
  und `indexCV32`. Alte Dateien ohne diese Felder bleiben importierbar.
- CSV unterstützt benannte Spalten, beispielsweise
  `cvNumber;value;decoderProfile;indexCV31;indexCV32`. Auch `cv_number`, `cv31`, `cv32`, `CV` und
  `Wert` werden erkannt. Ohne Kopfzeile gilt die Reihenfolge CV-Nummer, Wert, Beschreibung,
  Kategorie, Decoderprofil, CV31, CV32. Komma oder Semikolon sind möglich.
- Die Liste liefert ihre Decoderprofile und Herkunftsangaben. Profil und Dateizuordnung im
  Einzelwerteditor gelten nur für den dort bearbeiteten Wert und werden nicht auf Listenimporte
  übertragen. Aus JSON exportierte Datei-IDs werden nicht automatisch übertragen.

Die Vorschau markiert Zeilen als **neu**, **geändert**, **gleich** oder **ungültig**. Neue und
geänderte Zeilen sind ausgewählt. Nur echte Wiederholungen derselben vollständigen Identität gelten
als doppelt; CV257 auf Seite 16/3 und CV257 auf Seite 16/4 sind verschiedene Werte.

**Nur neue**, **Alle auswählen**, **Keine auswählen** und einzelne Kontrollkästchen steuern die
Auswahl über alle Vorschauseiten. **Ausgewählte Felder übernehmen** speichert die Auswahl samt
Historie in einer Transaktion. Ein Fehler rollt sämtliche Änderungen dieses Imports zurück.
Die Originaldatei wird dabei nicht gespeichert; sie kann zusätzlich unter CV-Dateien abgelegt werden.

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

LokProgrammer-TXT liefert bis zu 10.000 CV-Werte einschließlich Indexseiten. Für JSON/CSV/TXT
liest **CVs prüfen** die Originaldatei erneut über den vollständigen CV-Importweg. Andere
Vorschauformate sind Extraktionsvorschläge und können unvollständig sein. Funktionsvorschläge
bleiben auf 32 begrenzt. Die binäre ESUX-Datei liefert Metadaten, keinen vollständigen CV-Import.

Die Vorschauaktionen sind unabhängig:

1. **Vorschlag übernehmen** kopiert das erste erkannte Profil und die Beschreibung in die noch
   ungespeicherten Dateifelder.
2. **CVs prüfen** übergibt erkannte Werte an die normale CV-Importvorschau. Erst das Übernehmen
   ausgewählter CV-Zeilen schreibt Daten. Ein erkanntes Profil gewinnt, andernfalls verwendet
   RailKeeper nacheinander aktuelles Dateiprofil, erkannten Decoder oder erkannten Projektnamen.
3. **Funktionen übernehmen** schreibt gültige erkannte Funktionen sofort. Sie verwenden erkannten
   Namen und Typ, ein leeres Symbol, Betriebsart `dauer`, unbekannte Funktionsangaben und den
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
Fahrzeug existiert. Der Tab **CV** zeigt die vollständige Werteliste mit Seitenwahl und die Quelllokomotive. **Fahrkurve** kann daraus seine schreibgeschützte Anzeige ableiten. **Steuerung**
lädt die Funktionszuordnungen des Entwurfs in die bearbeitbaren Funktionszeilen.

Mit aktivem ECoS-Entwurf ist das Speichern des Grunddatensatzes ein besonderer mehrstufiger Ablauf.
RailKeeper speichert zuerst Fahrzeugdaten und geänderte Funktionen gemeinsam, danach nacheinander
die externe ECoS-Zuordnung und die CV-Werte des Entwurfs. Nur
bei vollständigem Erfolg markiert RailKeeper den ECoS-Import als gespeichert und lädt das Fahrzeug
neu. Ein später Fehler macht weder das Fahrzeug noch frühere ECoS-Schreibvorgänge rückgängig. Lade
neu, vergleiche gespeicherte Zuordnung, CVs und Funktionen mit dem Entwurf und wiederhole nur
fehlende Arbeiten.

Nach erfolgreichem Abschluss verwenden normale Funktions-, CV- und Dateiaktionen das gespeicherte
Fahrzeug. Dieses Kapitel erklärt keine ECoS-Verbindung, Rohprüfung, Synchronisierung,
Konfliktbehandlung oder Schreibvorgänge zur Digitalzentrale. Diese Abläufe stehen im [ECoS-Abgleich](../import-export/ecos-sync) und in der [Zentraleneinrichtung](../../administration/digital-centers).

ECoS-Werte behalten ihre eigene Herkunft. Bei eindeutigem CV8-Hersteller und passendem gespeicherten Decoder-Typ können Modellregeln greifen. Fehlende Indexseiten bleiben unbekannt; eine Bibliothekszuordnung erfindet sie nicht.

## Daten bei Schreibvorgängen schützen

| Aktion | Speichert Daten | Lädt das vollständige Fahrzeug neu |
| --- | --- | --- |
| Funktionsfeld bearbeiten, ergänzen, importieren oder löschen | Nein, zunächst Entwurf | Nein |
| Änderungen speichern | Fahrzeug und geänderte Funktionen gemeinsam | Nach Erfolg |
| Funktions-JSON importieren | Erst mit Änderungen speichern | Nein |
| Funktions-JSON exportieren | Nein | Nein |
| Fahrkurve ansehen | Nein | Nein |
| Mit aktivem ECoS-Entwurf speichern | Fahrzeug/Funktionen gemeinsam, danach Zuordnung/CVs | Nach Gesamterfolg |
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

[Modell- und Firmwarehinweise](decoder-model-notes) · [Einfache Arbeitsabläufe](decoder-cv)
