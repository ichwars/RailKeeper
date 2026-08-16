---
title: Fahrzeugbestand und Grunddaten
description: Fahrzeuge suchen, filtern, anlegen, pflegen, ausgeben und sicher löschen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Fahrzeugbestand und Grunddaten

Der **Fahrzeugbestand** ist der zentrale Arbeitsbereich für Modellbahnfahrzeuge. Er verbindet
Bestandsstatus, Suche, Filter, Tabellen- und Kartenansicht, Fahrzeuggrunddaten, QR-Etiketten und
druckbare Reports. Dieses Kapitel beschreibt die stabile RailKeeper-Version v0.1.17.6.

Admin, Editor, Viewer und Planner können den Bestand einsehen. Fahrzeuge anlegen, ändern und
löschen dürfen nur Admin und Editor. In v0.1.17.6 können Schreibfunktionen für Viewer und Planner
trotzdem sichtbar sein. Der Server lehnt ihre Schreibversuche ab und RailKeeper zeigt den
Fehler an.

Medien, Wartung, Decoder- und CV-Daten, Artikeldatensuche sowie Ersatzteile haben eigene Abläufe.
Dieses Kapitel erklärt deren Einstiegspunkte, konzentriert sich aber auf den Fahrzeuggrunddatensatz.

## Bestandsstatus lesen

Öffne **Fahrzeugbestand** in der Seitenleiste. Die Seite trägt die Überschrift **Bestand**. Über der
Liste stehen fünf Statusbereiche:

| Status | Bedeutung und Aktion |
| --- | --- |
| **Gesamtbestand** | Zeigt alle geladenen Fahrzeuge sowie Anzahlen nach Kategorie und Spurweite. Ein Klick entfernt alle Bestandsfilter. |
| **Digitalisierung** | Zeigt Digitalisierungsgrad sowie digitale und analoge Fahrzeuge. Ein Klick aktiviert den Bestandsfilter **Digital**. |
| **Wartung** | Zählt unerledigte, heute fällige oder überfällige Wartungen. Die kleinere Zahl umfasst unerledigte Einträge der nächsten 14 Tage. Ein Klick aktiviert **Wartung fällig**. |
| **Nächster Termin** | Zeigt den ältesten überfälligen oder nächsten anstehenden unerledigten Eintrag innerhalb von 14 Tagen. Ein Klick öffnet die schreibgeschützte Fahrzeugansicht. |
| **Bildpflege** | Zeigt Anteil und Anzahl der Fahrzeuge mit mindestens einem Bild. Ein Klick aktiviert **Ohne Bild**. |

Das Aktualisieren-Symbol lädt die aktuelle Serversuche neu. RailKeeper aktualisiert den sichtbaren
Bestand außerdem, wenn das Browserfenster wieder den Fokus erhält, online geht oder aus einem
ausgeblendeten Tab zurückkehrt.

## Bestand durchsuchen

Gib Text in **Bestand durchsuchen** ein. Jede Änderung lädt die Fahrzeugliste neu vom Server. Die
Suche findet Teilzeichenfolgen in genau vier Feldern:

- Inventarnummer
- Hersteller
- Artikelnummer
- Bezeichnung

Beschreibungen, Bahngesellschaften, Epochen, Kategorien, Decodernummern, EANs und andere
Detailfelder durchsucht v0.1.17.6 nicht. Die nachfolgenden Filter wirken im Browser auf die vom
Server gelieferten Suchergebnisse. Suche und Filter werden daher kombiniert.

Schlägt das Laden fehl, erscheint der Fehler über der Liste. Bereits geladene Zeilen können sichtbar
bleiben. Behebe den Fehler oder aktualisiere die Liste, bevor du sie als aktuellen Stand verwendest.

## Bestand filtern

Filtergruppen sind UND-verknüpft. Ein Fahrzeug muss jede aktive Gruppe erfüllen.

| Gruppe | Stabile Auswahl in v0.1.17.6 |
| --- | --- |
| Bestandsstatus | **Alle**, **Digital**, **Analog**, **Mit Bild**, **Ohne Bild** |
| Wartung | **Alle**, **Wartung fällig**, **Ohne Wartung** |
| Stammdaten | Hersteller, Kategorie, Gattung |
| Betriebsmerkmal | **Messe tauglich** |
| Datenlücke aus der Übersicht | **Ohne Artikel-Nr.**, **Ohne EAN** oder **Digital ohne Decoder-Nr.**, wenn der Bestand aus der **Übersicht** geöffnet wurde |

Die Bezeichnung **Ohne Wartung** ist in v0.1.17.6 ungenau. Der Filter zeigt Fahrzeuge ohne fälligen
Wartungseintrag. Dazu können Fahrzeuge mit erledigten oder noch nicht fälligen Wartungen gehören.

Eine ausgewählte Kategorie beschränkt die Gattungen auf die unterschiedlichen Werte der aktuell vom
Server geladenen Suchergebnisse in dieser Kategorie und entfernt die aktuelle Gattungsauswahl.
**Filter entfernen** setzt alle Gruppen zurück und entfernt einen `gap`-Parameter der Übersicht aus
der Browseradresse. Der Ergebniszähler bezieht sich immer auf alle aktiven Filter.

Zusätzliche Filter für Bahngesellschaft, Epoche und Schnittstelle aus späteren Entwicklungsständen
gehören nicht zur stabilen v0.1.17.6.

## Ansicht wählen und sortieren

Auf dem Desktop wechselst du über die Ansichtssymbole zwischen Tabelle und Karten. RailKeeper
speichert die Auswahl im lokalen Speicher des aktuellen Browsers, nicht im Benutzerkonto. Auf
schmalen Bildschirmen zeigt die Oberfläche unabhängig davon eine kompakte mobile Liste.

Die Tabelle enthält Auswahl, Bild, Inventarnummer, Hersteller, Artikelnummer, Bezeichnung,
Spurweite, Epoche, Ausstellungsstatus und Aktionen. Sortierbar sind Inventarnummer, Hersteller,
Artikelnummer, Bezeichnung, Spurweite und Epoche.

Die Anfangssortierung ist Inventarnummer aufsteigend und berücksichtigt Zahlen, sodass `...2` vor
`...10` steht. Ein weiterer Klick auf die aktive Überschrift kehrt die Richtung um. Karten- und
Mobilansicht folgen der aktuellen Tabellensortierung, bieten aber keine eigenen Sortierschalter.

## Fahrzeuge markieren

Die Kontrollkästchen der Tabelle markieren Fahrzeuge für Reports. Das Kästchen in der Kopfzeile
markiert oder entfernt alle nach Suche und Filtern aktuell sichtbaren Fahrzeuge. Eine Markierung kann
im Speicher bleiben, wenn ein späterer Filter die Zeile ausblendet. Der Reportumfang **Markierung**
verwendet jedoch nur markierte Fahrzeuge, die weiterhin sichtbar sind.

Karten- und kompakte Mobilansicht zeigen keine Kontrollkästchen. Wechsle zur Tabelle, um eine
gezielte Reportauswahl zusammenzustellen.

## Fahrzeug öffnen und prüfen

Klicke auf die Fahrzeugbezeichnung, das Bild, **Anzeigen** oder das Augensymbol. RailKeeper lädt den
vollständigen Datensatz und öffnet die **Fahrzeugansicht**. Leere Felder und Abschnitte werden nicht
angezeigt. Die Ansicht kann enthalten:

- Produkt-, Modell-, Technik-, Eigentums- und Steuerungsdaten
- konfigurierte Funktionen
- Bilder und Anhänge
- Wartungseinträge
- bis zu zwölf angezeigte CV-Werte sowie CV-Dateien

Die Kopfaktionen öffnen **Bearbeiten**, einen einzelnen Detailreport oder einen QR-Code. Das
Schnellmenü der Zeile führt außerdem direkt zu Uploads, Wartung und Ersatzteilen. Diese Fachbereiche
werden in eigenen Kapiteln beschrieben. Auch nach **Bearbeiten** benötigt das Speichern Admin- oder
Editor-Rechte.

## Fahrzeug anlegen

Admin und Editor können einen Datensatz anlegen:

1. Klicke auf **Neues Fahrzeug**.
2. Fülle die fünf Pflichtfelder **Hersteller**, **Bezeichnung**, **Spurweite**, **Kategorie** und
   **Gattung** aus.
3. Lasse die **Inventarnummer** für eine automatische Vergabe leer oder trage eine eindeutige Nummer
   ein.
4. Ergänze optionale Modell-, Detail-, Eigentums-, Steuerungs- oder QR-Einstellungen.
5. Klicke auf **Anlegen**.

Die automatische Vergabe nutzt das aktive Inventarnummernschema der ausgewählten Kategorie. Ohne
aktives Schema schlägt das Anlegen fehl. Eine manuell eingegebene Nummer muss eindeutig sein. Der
Server entfernt beim Speichern Leerraum am Anfang und Ende von Textfeldern.

Nach dem Anlegen bleibt der Dialog im Bearbeitungsmodus geöffnet und meldet, dass die weiteren Tabs
nun bearbeitet werden können. Funktionen, Geschwindigkeitskurve, CVs, Uploads, Wartungen und
Ersatzteile benötigen die gespeicherte Fahrzeug-ID und können nicht vor den Grunddaten persistiert
werden.

Die Funktionen **Barcode** und **Artikeldaten suchen** können Felder und Bilder aus externen Quellen
vorschlagen. Vorschläge müssen geprüft werden und gehören zum gesonderten Ablauf der
Artikeldatensuche.

## Referenz der Grunddaten

### Modell und Produktidentität

| Felder | Zweck und Verhalten |
| --- | --- |
| Inventarnummer | Eindeutige lokale Identität. Beim Anlegen bedeutet leer eine automatische Vergabe, beim Ändern bleibt dadurch die bisherige Nummer erhalten. |
| Artikel-Nr., Quelle / URL | Hersteller- oder Katalogidentität und optionaler Quellenlink. Die Quelle wird normalerweise durch Suche oder Import statt über ein einfaches Formularfeld gesetzt. |
| Hersteller, Bezeichnung, Spurweite | Erforderliche Stamm- und Modelldaten. Hersteller und Spurweite verwenden konfigurierte Stammdaten. |
| Bahngesellschaft, Epoche | Optionale konfigurierte Stammdaten. |
| Kategorie, Gattung | Beide sind Pflichtfelder. Die verfügbaren Gattungen hängen von der Kategorie ab. |
| Beschreibung, Baureihe, Fahrzeug-Nr. | Optionale freie Modellbeschreibung und Vorbildidentität. |
| Höchstgeschwindigkeit, Heimat-Bw / Einsatzstelle | Optionale Vorbild- und Betriebsdaten. Die Höchstgeschwindigkeit akzeptiert ganze Werte von 1 bis 1000 km/h; die Einsatzstelle ist Freitext. |
| EAN, Produktionszeit, Listenpreis | Optionale Handelsdaten. Produktionszeit und Listenpreis werden als bereinigter Text gespeichert. Die Übersicht wertet parsebare Listenpreise aus. |
| Digital, Digital / Decoder-Nr. | Der Schalter aktiviert die primäre digitale Decodernummer. Der Ausstellungsschalter prüft diese primäre Nummer. |
| DT / Decoder, DT / Decoder-Nr., Decoder-Typ | Optionaler zweiter Decoderschalter mit Nummer und Beschreibung. |
| Messe tauglich, Ausstellung | Getrennte Betriebsmerkmale. **Messe tauglich** ist ein normales Datensatzmerkmal, **Ausstellung** wird mit der Ausstellungsliste koordiniert. |
| ABC-Bremsen | Erfasst die ABC-Bremsfähigkeit. |

### Technische Details

| Felder | Zweck und Verhalten |
| --- | --- |
| Länge (mm), Gewicht (g) | Freie numerische Texteingaben für Abmessungen. |
| Farbe, Beschriftung, Beladung, Inneneinrichtung, Achsen | Beschreibende physische und optische Details. |
| Anzahl Achsen, Anzahl Haftreifen | Zählfelder mit numerischem Tastaturhinweis. Der Server speichert bereinigten Text. |
| Radsatz, Stromabnahme, Adapter / Schnittstelle | Technische Auswahl aus den nachfolgenden stabilen Listen. |
| Kupplung (V=H), Kupplung vorne, Kupplung hinten | Sind vorne und hinten gleich, folgt der hintere Wert dem vorderen und kann nicht getrennt bearbeitet werden. |
| Antrieb, Fahrlicht, Beleuchtung, Soundgenerator, Rauchgenerator | Jeder Schalter aktiviert das zugehörige Beschreibungsfeld. Ausschalten löscht eine vorhandene Beschreibung nicht automatisch. |
| QR-Code erstellen | Aktiviert die QR-Schaltfläche im Bearbeitungsformular, wenn Inventarnummer oder Bezeichnung vorhanden ist. |

### Fahrzeug, Eigentum und Zustand

| Felder | Zweck und Verhalten |
| --- | --- |
| Erwerb, von/bei | Wie und aus welcher Quelle das Fahrzeug erworben wurde. |
| Preis, Datum | Kaufpreis als bereinigter Text und Kaufdatum als Datumsfeld. |
| Standort, Details | Allgemeiner Lagerort sowie genaueres Regal, Fach oder Position. |
| Zustand, Details | Standardisierter Zustand und freie Ergänzung. |
| Verpackung | Zustand oder Art der Verpackung. |
| Zusatzinformationen | Längere Hinweise, die keinem anderen Grunddatenfeld zugeordnet sind. |

## Stabile Auswahllisten

Die folgenden Werte sind in v0.1.17.6 statisch. Sie bleiben in beiden Oberflächensprachen auf
Deutsch gespeichert.

| Feld | Werte |
| --- | --- |
| Radsatz | `2-Leiter DC`, `3-Leiter AC`, `NEM`, `RP25`, `Metall`, `Kunststoff` |
| Kupplung | `NEM-Schacht`, `Kurzkupplung`, `Bügelkupplung`, `Klauenkupplung`, `Schraubenkupplung` |
| Stromabnahme | `Schiene`, `Oberleitung`, `Batterie`, `Akku` |
| Adapter / Schnittstelle | `NEM 651`, `NEM 652`, `PluX16`, `PluX22`, `MTC21`, `Next18`, `8-polig`, `21-polig` |
| Erwerb | `Kauf`, `Tausch`, `Geschenk`, `Erbe`, `Leihgabe`, `Sonstiges` |
| von/bei | `Händler`, `Privat`, `Messe / Börse`, `Online`, `Auktion`, `Hersteller`, `Verein`, `Sonstiges` |
| Standort | `Auf Anlage`, `Vitrine`, `Lager`, `Werkstatt`, `Transportbox`, `Ausgeliehen`, `Sonstiges` |
| Zustand | `Neu`, `Neuwertig`, `Sehr gut`, `Gut`, `Gebraucht`, `Leichte Gebrauchsspuren`, `Reparaturbedürftig`, `Defekt` |
| Verpackung | `Originalverpackung`, `Ersatzverpackung`, `Ohne Verpackung`, `Transportbox`, `Sonstiges` |

Hersteller, Spurweite, Epoche, Bahngesellschaft, Kategorie und Gattung stammen aus den
konfigurierten Stammdaten der Instanz statt aus diesen statischen Listen.

## Fahrzeug bearbeiten

1. Öffne **Bearbeiten** über Tabelle, Karte, Mobilzeile, Schnellmenü oder Fahrzeugansicht.
2. Ändere Pflicht- oder optionale Felder. Pflichtfelder dürfen beim Speichern nicht leer sein.
3. Klicke auf **Änderungen speichern**.

Beim Wechsel der Kategorie entfernt RailKeeper eine Gattung, die der neuen Kategorie nicht
zugeordnet ist. **Kupplung (V=H)** kopiert die vordere Kupplung nach hinten. Spätere Änderungen vorn
aktualisieren dann beide Werte.

Eine geänderte Inventarnummer muss erneut eindeutig sein. RailKeeper schreibt alte und neue Nummer
in die Inventarnummernhistorie und protokolliert die Fahrzeugänderung im Audit-Log. Nach
erfolgreichem Speichern wird der Bestand neu geladen und der Dialog bleibt im Bearbeitungsmodus.
Der Server erzwingt die Admin- oder Editor-Berechtigung unabhängig von sichtbaren Schaltflächen.

## QR-Etiketten erstellen

Ein QR-Code benötigt mindestens Inventarnummer oder Bezeichnung. Seine Klartextnutzlast lautet:

```text
Inventar-Nr.: <Inventarnummer>
Bezeichnung: <Bezeichnung>
Decoder-Nr.: <Decodernummer, nur wenn vorhanden>
```

RailKeeper verwendet zuerst die primäre digitale Decodernummer, danach die DT-Decodernummer. Der
QR-Dialog kann PNG oder SVG herunterladen und ein druckbares Etikett öffnen. Schnellmenü und
Fahrzeugansicht können bei vorhandenen Identitätsdaten immer einen QR-Code erzeugen. Der Schalter
**QR-Code erstellen** steuert in v0.1.17.6 nur die QR-Schaltfläche im Detailbereich des
Bearbeitungsformulars.

Ist Inventarnummer oder Bezeichnung leer, schreibt die Nutzlast für dieses Feld `-`. Mindestens eines
der beiden Felder muss einen Wert enthalten, bevor RailKeeper den QR-Code erzeugt.

## Bestandsreports erstellen

Der Dialog startet mit **Übersichtsliste**, dem Titel `Fahrzeugsammlung`, dem Umfang **Alles** sowie
aktiviertem QR-Code und Bild. Passe diese Voreinstellungen bei Bedarf an:

1. Suche, filtere, sortiere und markiere bei Bedarf Fahrzeuge.
2. Klicke in der Bestandswerkzeugleiste auf **Report erstellen**.
3. Wähle **Übersichtsliste** oder **Detailliste** und trage einen Titel ein.
4. Wähle unter **Drucken** den Umfang **Alles** oder **Markierung**.
5. Schließe **QR-Code** und **Bild** ein oder aus.
6. Klicke auf **Report erstellen** und drucke anschließend über den Browser oder speichere als PDF.

**Alles** bedeutet alle aktuell gesuchten und gefilterten Fahrzeuge in der aktuellen Sortierung,
nicht jeden Datensatz der Datenbank. **Markierung** berücksichtigt nur markierte Fahrzeuge, die noch
sichtbar sind, und ist ohne sichtbare Markierung deaktiviert.

Die Übersichtsliste ist kompakt. Die Detailliste lädt vollständige Datensätze und kann Funktionen,
Wartung, CV-Daten, Bilder, Anhänge und externe Zuordnungen enthalten. Der Einzelreport aus
Schnellmenü oder Fahrzeugansicht ist immer ein Detailreport mit QR-Code und Bildern.

Existiert kein passendes Fahrzeug, erstellt RailKeeper keinen Report. Erlaube das Druckfenster, falls
der Browser Pop-ups blockiert.

## Grenze des Ausstellungsschalters

Der Schalter **Ausstellung** in der Tabelle kann nur aktiviert werden, wenn **Digital** eingeschaltet
und die primäre **Digital / Decoder-Nr.** gefüllt ist. Die DT-Decodernummer allein genügt nicht.

Beim Aktivieren öffnet RailKeeper eine entsperrte Ausstellungsliste und verhindert ein doppeltes
Fahrzeug für dieselbe Eigentümer-Namens-Kombination sowie eine doppelte Decodernummer in dieser
Liste. Der gesamte Ablauf benötigt Adminrechte oder eine Kombination aus Editor- und
Messeberechtigung.
Das Deaktivieren benötigt Schreibrechte und setzt das Fahrzeugmerkmal zurück, löscht aber keinen
bereits vorhandenen Ausstellungslisteneintrag.

## Fahrzeug löschen

Admin und Editor können **Löschen** wählen und das durch Inventarnummer und Bezeichnung
identifizierte Fahrzeug bestätigen. v0.1.17.6 bietet weder Rückgängig noch eine Eingabebestätigung.

Das Löschen entfernt das Fahrzeug und seine abhängigen Datenbankeinträge, darunter
Inventarnummernhistorie, Bilder, Anhangsmetadaten, Wartungen, Funktionen, CV-Daten, externe
Zuordnungen und Ersatzteile. Zuordnungen zu Anlagen oder Zubehör können das Löschen blockieren.
RailKeeper schreibt bei Erfolg einen Audit-Eintrag `VehicleDeleted`. Exportiere vor dem Löschen
wichtiger Datensätze eine aktuelle Sicherung. Die Anwendung verspricht nicht, jede referenzierte
Datei physisch zu entfernen.

## Leere, Lade- und Fehlerzustände

| Situation | Ergebnis und nächster Schritt |
| --- | --- |
| Erstes Laden | **Lade Fahrzeuge aus lokaler Datenbank...** erscheint bis zum Abschluss der Anfrage. |
| Noch kein Fahrzeug | **Noch keine Fahrzeuge vorhanden.** Admin oder Editor können ein Fahrzeug anlegen oder Import/Export verwenden. |
| Suche oder Filter ohne Treffer | **Keine Fahrzeuge für diesen Filter gefunden.** Entferne Filter oder ändere die Suche. |
| Liste oder Details schlagen fehl | RailKeeper zeigt den Serverfehler über dem Bestand. Prüfe Sitzung und Verbindung und aktualisiere danach. |
| Pflichtfeld fehlt | Der Bereich Modell öffnet sich und nennt Hersteller, Bezeichnung, Spurweite, Kategorie oder Gattung als fehlend. |
| Automatische Nummer schlägt fehl | Bitte einen Administrator, ein aktives Inventarnummernschema für die Kategorie einzurichten. |
| Manuelle Nummer kollidiert | Wähle eine andere eindeutige Inventarnummer. |
| Speichern oder Löschen verboten | Melde dich als Admin oder Editor an. Sichtbare Bedienelemente ersetzen die serverseitige Rollenprüfung nicht. |
| Speichern schlägt fehl | Der Fehler erscheint über dem Bestand und der Editor bleibt geöffnet. Betrachte die Änderung als ungespeichert, prüfe Sitzung und Verbindung und versuche es erneut. |
| Löschen schlägt fehl | Der Fehler erscheint über dem Bestand und die Bestätigung bleibt geöffnet. Betrachte das Fahrzeug als nicht gelöscht, behebe die gemeldete Ursache und versuche es erneut. |
| Reporterzeugung schlägt fehl | Der Fehler erscheint über dem Bestand und der Reportdialog bleibt geöffnet. Prüfe Verbindung und Pop-up-Behandlung und erstelle ihn erneut. |
| Report ohne Zeilen | Passe Suche oder Filter an oder markiere mindestens eine sichtbare Zeile. |
| Löschen blockiert | Entferne oder ändere die Anlagen- oder Zubehörzuordnung, die das Fahrzeug noch referenziert, und versuche es erneut. |

## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Ersteinrichtung und Anmeldung](/de/guide/getting-started/)
- [Übersicht, Kennzahlen und Datenqualität](/de/guide/overview/)
- [Übersicht der Administration](/de/administration/)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.17.6** und wurde zuletzt am
16.08.2026 geprüft.
