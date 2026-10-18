# RailKeeper Ausstellungsanlagenplanung

Datum: 2026-09-06
Status: Fachlich und visuell freigegeben, schriftliche Spezifikation zur Prüfung

## Ziel

Die Ausstellungsanlagenplanung wird als eigener Arbeitsablauf in den bestehenden Bereich
**Ausstellung** integriert. Sie kombiniert fertig gebaute, physische Modulexemplare aus der
projektweiten Modulbibliothek zu konkreten veranstaltungsgebundenen Aufbauten.

Die normale Anlagenplanung bleibt der Ort für Neuentwurf, Modulvorlagen, Geometrie, Gleisplanung
und technische Ausarbeitung. Beide Arbeitsabläufe verwenden denselben Modul-, Geometrie-, Norm-,
Verbindungs- und Transformationskern, führen aber keine doppelten Moduldaten.

Eine Veranstaltung kann keine, eine oder mehrere voneinander unabhängige Ausstellungsanlagen
enthalten. Jede Ausstellungsanlage kann mehrere alternative Aufbauvarianten und mehrere
Höhenebenen besitzen.

## Abhängige Spezifikation

Diese Funktion baut auf der [projektweiten Modulbibliothek und den normierten
Verbindungsseiten](./2026-09-06-layout-modules-connection-sides-design.md) auf. Die
Ausstellungsplanung beginnt erst, wenn Modulvorlagen, Vorlagenrevisionen, physische Exemplare und
stabile Verbindungsseiten aus deren ersten beiden Implementierungsstufen verfügbar sind.

## Verbindliche visuelle Referenzen

Die Umsetzung übernimmt die freigegebene Variante 2 ohne eigenständige visuelle Neuinterpretation.
Die Bilder sind verbindlich für Seitenaufbau, Hierarchie, dunkle Arbeitsfläche, Werkzeugleiste,
Inspektor, Statusdarstellung, Typografie, Abstände, Farben und Informationsdichte.

### Anlagenübersicht und große Planfläche

`C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-ffa8ae91-5053-4c73-b629-83a701194a95.png`

SHA-256:
`3CAB446FD1FB6D3DC24391E7FF6C050CCEFF841EC5F429CF410BBF943522ABA5`

### Werkzeug-, Geometrie- und Inspektorsprache

`C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-be12f866-ee67-43f9-820c-f7c5131e29ec.png`

SHA-256:
`C387A13776ED3C839BEACAEED0D1B7A186C2749D56CF9BD29D5926E6EA8F1DA2`

Der zusätzliche Modulpool, die Ebenenfilter und die Ausstellungsmetadaten werden in diese festgelegte
Komposition eingefügt. Sie begründen kein abweichendes Karten-, Farb- oder Navigationssystem.
Sichtbare Abweichungen werden vor ihrer Umsetzung abgestimmt. Responsive und barrierefreie
Anpassungen bewahren Hierarchie und visuelle Gewichtung der Referenzen.

Vor visuellen Implementierungsarbeiten werden Existenz und SHA-256-Prüfsumme beider Dateien
kontrolliert. Fehlt eine Datei oder stimmt die Prüfsumme nicht, wird die Arbeit angehalten und das
Original erneut angefordert. Eine Rekonstruktion aus Erinnerung ersetzt keine Referenzdatei.

## Festgelegte Produktentscheidungen

- Ausstellungsanlagenplanung und normale Anlagenplanung sind getrennte Arbeitsabläufe.
- Die Ausstellungsanlagenplanung liegt im bestehenden Bereich **Ausstellung**.
- Jede Veranstaltung kann beliebig viele getrennte Ausstellungsanlagenpläne besitzen.
- Ein Ausstellungsanlagenplan gehört genau zu einer Veranstaltung.
- Ein Plan besitzt eine Spurweite, einen Maßstab, einen optionalen Hallenbereich, benannte Ebenen
  und eine oder mehrere Aufbauvarianten.
- Ausstellungsaufbauten verwenden ausschließlich konkrete physische Modulexemplare.
- Modulgeometrie, Normen und Gleispläne sind im Ausstellungseditor schreibgeschützt. Änderungen
  erfolgen in der projektweiten Modulbibliothek als neue Revision.
- Ein physisches Exemplar darf in einem gleichzeitig aktiven Aufbau nur einmal vorkommen.
- Module können direkt auf der Planfläche verschoben, gedreht, gruppiert und über kompatible
  Verbindungsseiten eingerastet werden.
- Eine Platzierung verwendet `X`, `Y`, eine Ebene mit `Z`-Höhe und eine Drehung um die Hochachse.
- Ganze Module können weder geneigt noch quergeneigt werden.
- Steigungen, Rampen und relative Gleishöhen bleiben Bestandteil der jeweiligen Modulgleisplanung.
- Das überprüfbare Startziel umfasst 1.000 Module und 100.000 Gleis- beziehungsweise
  Technikobjekte in einem Aufbau. Es ist keine fest eingebaute Obergrenze.
- Die Umsetzung erfolgt vollständig inline und ohne Subagenten.
- Deutsche und englische Oberflächentexte werden gesondert auf Rechtschreibung, Terminologie und
  abgeschnittene Inhalte geprüft.

## Ausgangslage

Der bestehende Ausstellungsbereich verwaltet Veranstaltungen mit Zeitraum, Ort, Status, Sperrung,
Fahrzeuglisten, Bereitschaftsprüfung und begründbaren Konflikten. Er enthält noch keinen räumlichen
Anlagenplan.

Der bestehende Anlagenbereich kennt Konfigurationen aus Modulen mit `X`, `Y` und Drehung. Module
gehören derzeit fest zu einer Anlage. Das Einrasten verwendet freie Ports. Die Gleisplanung kennt
relative Start- und Endhöhen einzelner Gleisobjekte, die Modulplatzierung selbst aber keine Ebene
oder `Z`-Koordinate.

Die neue Funktion erweitert diese Grundlagen. Sie ersetzt weder Fahrzeuglisten noch normale
Anlagenentwürfe.

## Betrachtete Ansätze

### 1. Getrennte Arbeitsbereiche mit gemeinsamem Planungskern

Dies ist der ausgewählte Ansatz. Die normale Anlagenplanung bleibt entwurfsorientiert. Die
Ausstellungsplanung erhält eine auf Veranstaltung, physische Verfügbarkeit und Aufbauprüfung
zugeschnittene Oberfläche. Beide nutzen dieselben technischen Moduldefinitionen und dieselbe
Geometrieauswertung.

### 2. Eine gemeinsame Anlagenansicht mit Modusumschaltung

Ein Modusschalter würde weniger Navigation erzeugen, aber Entwurfswerkzeuge, physische Exemplare,
Veranstaltungsstatus, Sperrung und Logistik in derselben Oberfläche mischen. Rollen und
Validierungsregeln wären schwerer nachvollziehbar.

### 3. Vollständig getrennte Datenmodelle

Zwei unabhängige Systeme hätten klare Zuständigkeiten, würden aber Konturen, Normen, Ports,
Gleispläne und Kompatibilitätslogik duplizieren. Korrekturen könnten zwischen normaler und
Ausstellungsplanung auseinanderlaufen.

## Fachliches Modell

### Veranstaltung und Ausstellungsanlagenplan

Eine bestehende Veranstaltung erhält den Arbeitsbereich **Anlagenpläne**. Sie kann null bis
beliebig viele Pläne enthalten, beispielsweise für unterschiedliche Spurweiten, Hallenbereiche oder
voneinander unabhängige Aufbauten.

Ein Ausstellungsanlagenplan enthält mindestens:

- stabile ID und Veranstaltungs-ID;
- Name und optionale Beschreibung;
- Spurweite und Maßstab;
- optionalen Hallenbereich oder eine freie Flächenkontur;
- Status und Revision;
- Aufbauvarianten;
- Erstellungs- und Änderungszeit.

Veranstaltungszeitraum, Ort und übergeordneter Lebenszyklus werden nicht kopiert. Sie stammen aus
der zugehörigen Veranstaltung.

### Aufbauvariante und Revision

Eine Aufbauvariante beschreibt eine alternative räumliche Kombination innerhalb eines
Ausstellungsanlagenplans. Beispiele sind ein L-Aufbau, eine U-Form oder eine verkürzte Variante für
eine kleinere Fläche.

Jede Variante besitzt einen lokalen Entwurf und unveränderliche Revisionen. Für jeden
Veranstaltungstag kann pro Ausstellungsanlagenplan höchstens eine Revision genau einer Variante
aktiv sein. Rein alternative Varianten reservieren ihre Module nicht gleichzeitig. Beim Sperren
der Veranstaltung werden die aktiven Revisionen fixiert.

Eine Revision hält mindestens fest:

- verwendete Modulexemplare und deren Vorlagenrevisionen;
- Ebenen;
- Platzierungen;
- ausdrückliche Verbindungen zwischen Modulköpfen;
- Hallenbereich und Kamera-unabhängige Weltgrenzen;
- bestätigte Ausnahmen und Prüfstand;
- Ersteller, Zeitpunkt und Basisrevision.

### Ebene und Höhe

Eine Variante besitzt benannte Ebenen. Eine Ebene enthält Name, Basishöhe `baseZMm`, Sortierung und
optionale Darstellungsmetadaten. Eine Modulplatzierung verweist auf eine Ebene und besitzt einen
zusätzlichen Höhenversatz `zOffsetMm`, normalerweise `0`.

Die absolute Modulhöhe lautet:

`absoluteZMm = level.baseZMm + placement.zOffsetMm`

Module bleiben horizontal. Die tatsächliche Höhe eines Gleisanschlusses ergibt sich aus absoluter
Modulhöhe und der relativen Gleis- oder Porthöhe der fixierten Modulrevision. Dadurch können
mehrstöckige Anlagen, Rampenmodule und Höhenübergänge geprüft werden, ohne vollständige
3D-Transformationen einzuführen.

### Modulplatzierung

Eine Platzierung enthält:

- physisches Modulexemplar;
- fixierte Modulvorlagen- und Gleisplanrevision;
- Ebene und Höhenversatz;
- `positionXMm`, `positionYMm` und `rotationDegrees`;
- Sortierung und optionalen Sperrstatus;
- lokale Prüfhinweise.

Ein physisches Exemplar kann in beliebig vielen historischen oder alternativen Varianten
vorkommen. Innerhalb derselben Variante darf es nur einmal platziert werden. Über alle am selben
Tag gleichzeitig aktiven Ausstellungspläne hinweg darf es ebenfalls nur einmal vorkommen. Eine
begründete organisatorische Ausnahme ist möglich, eine doppelte Platzierung innerhalb derselben
Variante jedoch nicht.

### Verbindung

Eine eingerastete Verbindung wird ausdrücklich gespeichert. Sie referenziert zwei Platzierungen
und je eine stabile Verbindungsseite. Eine Verbindungsseite darf innerhalb derselben Variante nur
Teil einer Verbindung sein.

Die Verbindungsauswertung prüft:

- mechanische Kopfprofile und Toleranzen;
- Spurweite und Gleisachsen;
- elektrische beziehungsweise digitale Schnittstellen;
- absolute Gleis- und Porthöhen;
- Ausrichtung und Abstand;
- veraltete oder fehlende Modulrevisionen.

Räumliche Nähe allein erzeugt keine dauerhafte Verbindung. Das Einrasten zeigt eine Vorschau, und
erst das Loslassen beziehungsweise die Tastaturbestätigung erzeugt die Verbindung.

## Bedienkonzept

### Einstieg im Bereich Ausstellung

Die bestehende Veranstaltungsansicht erhält den Reiter **Anlagenpläne** neben Übersicht,
Fahrzeugen und Konflikten. Die Liste zeigt pro Plan Name, Spurweite, aktive Variante, Modulanzahl,
benötigte Fläche, Ebenen, Prüfstatus und letzten Änderungsstand.

**Neuer Anlagenplan** fragt nur Name, Spurweite, Maßstab und optionalen Hallenbereich ab. Danach
öffnet sich die große Arbeitsfläche mit der Variante **Standard** und der Ebene **Ebene 0** bei
`0 mm`. Ein bestehender normaler Anlagenentwurf oder ein früherer Ausstellungsplan kann als
Ausgangspunkt übernommen werden. Dabei werden Platzierungen kopiert, Moduldefinitionen aber
weiterhin referenziert.

Eine Veranstaltung kann mehrere Pläne getrennt prüfen und freigeben. Die Veranstaltung selbst kann
erst gesperrt werden, wenn alle aktiven Anlagenplanrevisionen einen zulässigen Prüfstand besitzen
oder ihre Ausnahmen begründet wurden. Die Veranstaltungssperre fixiert anschließend alle aktiven
Revisionen gemeinsam.

### Interaktive Arbeitsfläche

Die Arbeitsfläche übernimmt die freigegebene Variante 2:

- kompakter Veranstaltungs- und Plankopf oben;
- große zoombare Planfläche in der Mitte;
- schmale transparente Werkzeugleiste links;
- kontextbezogener Inspektor rechts;
- Hinweise, Material und Änderungen in einem unteren Bereich;
- ruhige Statuszeile für Koordinaten, Ebene, Zoom und Speicherzustand.

Der projektweite Modulpool öffnet aus der linken Werkzeugleiste als schmales Panel. Er bietet Suche
und Filter nach Kennung, Name, Eigentümer, Spurweite, Normprofil, Zustand und Verfügbarkeit. Bereits
in derselben Variante verwendete Exemplare sind gekennzeichnet und können nicht erneut gezogen
werden.

Ein Modul wird aus dem Pool auf die Fläche gezogen oder per Tastatur platziert. Auswahl und
Inspektor bieten:

- freie Verschiebung;
- Drehgriff und numerischen Winkel;
- numerische Eingabe von `X`, `Y`, Ebene und Höhenversatz;
- Sperren gegen versehentliches Bewegen;
- Öffnen der schreibgeschützten Moduldaten;
- Auswahl der verwendeten Gleisplanrevision;
- Lösen oder Prüfen gespeicherter Verbindungen.

### Einrasten und Verschieben

Während des Verschiebens hebt RailKeeper passende Verbindungsseiten hervor. Eine transparente
Vorschau zeigt die daraus berechnete Position und Drehung. Die Vorschau benennt Norm-, Gleisachsen-,
Elektrik- und Höhenstatus vor dem Loslassen.

Das Einrasten berechnet `X`, `Y`, `Z` und Drehung so, dass die beiden Köpfe gegeneinander liegen.
Es neigt kein Modul. Stimmen absolute Anschluss- beziehungsweise Gleishöhen nicht überein, bleibt
die Abweichung sichtbar und erfordert eine bewusste Bestätigung.

Beim Ziehen eines bereits verbundenen Moduls fragt die erste Bewegung, ob RailKeeper:

- nur dieses Modul löst und verschiebt;
- die verbundene Auswahl als Gruppe verschiebt;
- den Vorgang abbricht.

Die Auswahl wird für den laufenden Ziehvorgang beibehalten. Mehrfachauswahl, Rechteckauswahl,
Gruppenverschiebung, Ausrichten, Sperren sowie Rückgängig/Wiederholen gehören zum Grundumfang.
Rückgängig/Wiederholen arbeitet nur auf dem lokalen Entwurf und erzeugt keine Folge von
Serverrevisionen.

### Ebenenansicht

Die Ebenensteuerung zeigt alle benannten Ebenen mit Höhe, Modulanzahl und Prüfstatus. Die aktive
Ebene wird vollständig dargestellt. Andere Ebenen können:

- ausgeblendet;
- als ruhige, nicht auswählbare Kontur hinterlegt;
- gemeinsam sichtbar geschaltet werden.

Ein Seiten- oder Höhenprofil zeigt die Ebenen, absolute Modulhöhen und vertikalen Übergänge als
zweidimensionale technische Ansicht. Eine frei drehbare 3D-Ansicht ist nicht vorgesehen.

### Hallenfläche und Navigation

Der optionale Hallenbereich wird als feste, bemaßte Kontur hinter dem Aufbau dargestellt. Module
außerhalb des Bereichs oder mit überlappender Sicherheitszone erhalten einen Hinweis. Mehrere
Ausstellungsanlagen derselben Veranstaltung können getrennten Hallenbereichen zugeordnet werden.

Navigation umfasst Mausrad- und Tastaturzoom, Verschieben der Kamera, **Alles anzeigen**,
**Auswahl anzeigen**, gespeicherte Ansichten und Sprung zu Modul, Verbindung, Ebene oder Hinweis.
Weltkoordinaten werden in Millimetern gespeichert. Der Renderer verwendet kamera-relative
Koordinaten, damit auch sehr große Ausdehnungen stabil bleiben.

## Datenfluss

1. Die Veranstaltungsansicht lädt nur Anlagenplan-Zusammenfassungen und Prüfzahlen.
2. Das Öffnen einer Variante lädt Ebenen, vereinfachte Modulkonturen, Platzierungen, Verbindungen
   und Status in einem kompakten Szenenvertrag.
3. Gleise und Technikdetails werden abhängig von sichtbarem Ausschnitt und Zoomstufe nachgeladen.
4. Ziehen, Drehen, Gruppieren und Lösen ändern zunächst einen lokalen Entwurf mit Befehlsprotokoll.
5. Der Client berechnet schnelle Fangvorschläge. Der Server validiert beim Speichern Geometrie,
   Identitäten, Verfügbarkeit, Höhen, Normen und aktuelle Revisionen verbindlich.
6. Der gesamte Änderungssatz wird mit `expectedRevision` in einer Transaktion gespeichert.
7. Veröffentlichung oder Sperrung erzeugt eine unveränderliche Revision mit fixierten Modul- und
   Gleisplanrevisionen.

Der Server überträgt keine 100.000 Detailobjekte für eine weit herausgezoomte Übersicht. Pool,
Szenenzusammenfassung und Detailabfragen sind getrennte Verträge.

## Skalierungsarchitektur

Das Startziel lautet: Ein Aufbau mit 1.000 Modulplatzierungen und insgesamt 100.000 Gleis- und
Technikobjekten bleibt interaktiv navigierbar. Dieses Ziel wird mit einem reproduzierbaren großen
Testdatensatz geprüft und nicht als fachliches Mengenlimit validiert.

Die technische Grundlage umfasst:

- paginierte und serverseitig gefilterte Modulsuche;
- indizierte Begrenzungsrahmen für Modulplatzierungen und sichtbare Detailobjekte;
- Ausschnittsabfragen mit Ebene und Detailstufe;
- vereinfachte Kontur- und Statusdarstellung bei großer Entfernung;
- Verbindungsseiten und Modulnamen bei mittlerer Entfernung;
- Gleise und Technik nur für sichtbare Module bei naher Entfernung;
- Canvas-basierte Massendarstellung mit separater zugänglicher DOM-Bedienebene;
- räumlichen Index für Auswahl, Fangziele und Kollisionsprüfung;
- kamera-relative Renderkoordinaten und unveränderte Millimeter-Weltkoordinaten im Datenmodell;
- Abbruch oder Verwerfung überholter Detailanfragen beim schnellen Zoomen und Verschieben.

Die erste Implementierung verwendet keine vorsorgliche Cloud- oder verteilte Architektur. SQLite
bleibt die lokale Quelle der Wahrheit.

## Prüfungen und Konflikte

Ein Entwurf darf Warnungen und bewusst bestätigte Abweichungen enthalten. Folgende Sachverhalte
werden mindestens erkannt:

- dasselbe physische Exemplar mehrfach in einer Variante;
- dasselbe Exemplar in gleichzeitig aktiven Plänen oder zeitlich überlappenden Ausstellungen;
- nicht bestätigter Bestand, fehlende Verfügbarkeit oder fehlende Eigentümerfreigabe;
- Spurweiten-, Kopfprofil-, Gleisachsen-, Elektrik- oder Höhenabweichung;
- offene Verbindungsseite;
- mehrfach belegte Verbindungsseite;
- Modulüberlappung oder Überschreitung des Hallenbereichs;
- fehlende, veraltete oder archivierte Modul- beziehungsweise Gleisplanrevision;
- ungespeicherte Änderungen, Versionskonflikte und gesperrte Veranstaltung.

Eine doppelte Platzierung desselben Exemplars innerhalb einer Variante und strukturell ungültige
Verbindungsreferenzen blockieren das Speichern. Fachliche Abweichungen dürfen im Entwurf bestehen.
Vor Aktivierung, Veröffentlichung oder Sperrung müssen blockierende Konflikte behoben sein.
Erlaubte Ausnahmen verwenden den bestehenden Ausstellungsablauf und benötigen eine Begründung.

Bei `409 Conflict` bleibt der lokale Entwurf erhalten. Die Oberfläche zeigt Änderungen an Ebenen,
Platzierungen, Verbindungen und verwendeten Revisionen und bietet Serverstand laden, Entwurf
sichern oder bewusst erneut anwenden.

## Rollen und Lebenszyklus

- Admin und Planner dürfen Ausstellungsanlagenpläne, Varianten, Ebenen und Platzierungen anlegen
  oder ändern.
- Editor und Viewer dürfen veröffentlichte beziehungsweise freigegebene Pläne lesen.
- Messe darf die für die laufende Veranstaltung notwendigen Pläne und Modulinformationen lesen,
  erhält aber keinen allgemeinen Zugriff auf die Modulbibliothek.
- Messe darf keine Modulgeometrie oder Ausstellungsplatzierung verändern.
- Mit Sperrung der Veranstaltung werden aktive Planrevisionen schreibgeschützt.
- Änderungen nach einer Sperrung verwenden den bestehenden kontrollierten Entsperr- und
  Begründungsablauf.

Auditereignisse umfassen mindestens Planerstellung, Modulplatzierung, Verbindung, Trennung,
Gruppenbewegung, Revisionswechsel, Ausnahme, Aktivierung, Veröffentlichung und Sperrung.

## API und Datenhaltung

Frontend, Backend und `openapi/railkeeper.yaml` werden gemeinsam erweitert. Das persistierte Modell
trennt mindestens:

- Ausstellungsanlagenpläne;
- Aufbauvarianten und unveränderliche Revisionen;
- Ebenen;
- Modulplatzierungen;
- Verbindungen zwischen konkreten Verbindungsseiten;
- Hallenbereiche;
- Tagesaktivierung und begründete Ausnahmen.

Die Szenen-API liefert nur die für eine Detailstufe benötigten Daten. Schreibende Aufrufe verwenden
CSRF-Schutz, Rollenprüfung und `expectedRevision`. Fremdschlüssel und serverseitige Prüfung
verhindern verwaiste Platzierungen, Verbindungen und Modulrevisionen.

Bestehende Ausstellungen erhalten durch die Migration keine leeren oder geratenen Anlagenpläne.
Bestehende Anlagenkonfigurationen bleiben normale Anlagenentwürfe. Eine ausdrückliche Aktion kann
einen geeigneten Anlagenentwurf als neue Ausstellungsvariante übernehmen und verlangt danach die
Zuordnung konkreter Modulexemplare.

## Backup und Restore

Backups umfassen alle Ausstellungsanlagenpläne, Varianten, Revisionen, Ebenen, Platzierungen,
Verbindungen, Hallenbereiche, Ausnahmen und referenzierten Modul- beziehungsweise
Gleisplanrevisionen.

Restore validiert vor dem destruktiven Import:

- vorhandene Modul- und Vorlagenidentitäten;
- unveränderte Inhalte stabiler Revisionen;
- Veranstaltung und Tagesbezüge;
- eindeutige Exemplare innerhalb einer Variante;
- geschlossene Verbindungsreferenzen;
- unterstützte Szenen- und Ebenenschemaversionen.

Alte Backups ohne Ausstellungsanlagenpläne bleiben kompatibel. Es werden keine Pläne, Ebenen oder
physischen Verfügbarkeiten erfunden.

## Barrierefreiheit und Textqualität

- Jede Drag-and-drop-Aktion besitzt einen gleichwertigen Tastatur- und Formulareingabeweg.
- Modul, Verbindung und Ebene sind über Suche und Sprungbefehle erreichbar.
- Fokus bleibt bei Auswahl, Dialogen, Ebenenwechsel und Fehlerbehebung nachvollziehbar.
- Auswahl, Kompatibilität, Ebene und Konflikt werden nie nur durch Farbe dargestellt.
- Mobile Ansichten verwenden mindestens 44 Pixel große primäre Trefferflächen.
- `prefers-reduced-motion` reduziert räumliche Übergänge auf unmittelbare Zustandswechsel.
- Lange deutsche Modul-, Eigentümer- und Hallennamen dürfen Inspektor und Listen nicht überlaufen.
- Alle sichtbaren Texte stammen aus der gemeinsamen deutschen und englischen i18n-Struktur.
- Beide Sprachen werden auf Rechtschreibung, einheitliche Fachbegriffe, Platzhalter und
  abgeschnittene Inhalte geprüft.

## Umsetzung in prüfbaren Stufen

Die Arbeiten erfolgen vollständig inline und ohne Subagenten. Jede Stufe erhält nach Freigabe einen
eigenen Implementierungsplan.

### Voraussetzung: Projektweite Modulbibliothek

Vorlagen, Revisionen, physische Exemplare und stabile Verbindungsseiten aus Stufe 1 und 2 der
Modulbibliothek müssen verfügbar und migriert sein.

### Stufe 1: Ausstellungsplan-Domäne

- Pläne, Varianten, Revisionen, Ebenen, Platzierungen und Verbindungen;
- Rollen, Audit, Lebenszyklus und Konfliktintegration;
- OpenAPI sowie Backup-/Restore-Vertrag;
- Übernahme eines normalen Anlagenentwurfs als expliziter Vorgang.

### Stufe 2: Interaktiver Aufbaueditor

- Arbeitsfläche nach der freigegebenen Variante 2;
- Modulpool, Ziehen, Drehen, numerische Position und Sperren;
- Einrastvorschau, ausdrückliche Verbindungen und Gruppenbewegung;
- lokaler Entwurf, Rückgängig/Wiederholen und Revisionskonflikte.

### Stufe 3: Ebenen und Ausstellungsprüfung

- benannte Ebenen und absolute Höhenberechnung;
- Höhenprofil und modulübergreifende Anschlussprüfung;
- Hallenbereiche, Tagesaktivierung, Verfügbarkeit und begründete Ausnahmen;
- Veröffentlichung und Sperrung mit fixierten Revisionen.

### Stufe 4: Große Aufbauten

- Ausschnitts- und Detailstufen-API;
- Canvas-Massendarstellung und räumliche Indizierung;
- großer reproduzierbarer Testdatensatz;
- Profiling und visuelle Prüfung bei 1.000 Modulen und 100.000 Detailobjekten.

## Verifikation

Automatisierte Backendtests decken mindestens ab:

- null, einen und mehrere Anlagenpläne pro Veranstaltung;
- mehrere Varianten und genau eine aktive Revision je Plan und Tag;
- Ebenenbasis, Höhenversatz und absolute Gleisanschlusshöhe;
- keine Neigung oder Querneigung vollständiger Module;
- eindeutige physische Exemplare innerhalb einer Variante;
- Überschneidungen zwischen gleichzeitig aktiven Plänen und Veranstaltungen;
- mechanische, Gleis-, Elektrik- und Höhenkompatibilität;
- ausdrückliche Verbindungen und eindeutige Kopfbelegung;
- atomaren Variantenspeicherstand und `409 Conflict`;
- Veranstaltungsstatus, Sperrung, Rollen, CSRF und Audit;
- Backup, Restore und alte Backups ohne Anlagenpläne;
- Ausschnitts- und Detailstufenabfragen mit großem Testdatensatz.

Frontendtests decken mindestens ab:

- Pläne und Varianten anlegen, auswählen, aktivieren und lesen;
- Module per Zeiger, Tastatur und numerischer Eingabe platzieren;
- freie Bewegung, Drehung, Ebenenwechsel und Höhenversatz;
- Einrastvorschau für passende, abweichende und unvollständige Köpfe;
- verbundenes Einzelmodul lösen oder verbundene Gruppe bewegen;
- Mehrfachauswahl, Sperren und Rückgängig/Wiederholen;
- aktive, ausgeblendete und transparent hinterlegte Ebenen;
- lokale Änderungen nach Netzwerk- und Versionsfehlern;
- gesperrte und rollenabhängig schreibgeschützte Ansichten;
- Detailsprünge, Suche und große Szenen mit wechselnder Zoomstufe.

Normale Projektprüfungen:

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd run build
```

Die visuelle Prüfung vergleicht die Desktopansicht bei gleicher Fenstergröße direkt mit den beiden
freigegebenen Referenzbildern. Zusätzlich werden Mobilansicht, helles und dunkles Thema,
Tastaturreihenfolge, lange deutsche Texte, leere und fehlerhafte Zustände, mehrere Ebenen und ein
großer Aufbau geprüft. Deutsche und englische Texte erhalten eine abschließende
Rechtschreibprüfung.

## Abnahmekriterien

- Der Bereich **Ausstellung** kann pro Veranstaltung mehrere unabhängige Anlagenpläne verwalten.
- Ausstellungsplanung und normale Anlagenplanung bleiben in Navigation, Rollen und Zweck getrennt.
- Beide verwenden dieselbe projektweite Modulbibliothek und dieselben technischen Revisionen.
- Ein Ausstellungsplan platziert ausschließlich konkrete physische Modulexemplare.
- Module lassen sich interaktiv verschieben, drehen, numerisch positionieren und gruppieren.
- Kompatible Verbindungsseiten rasten mit sichtbarer Vorschau ein.
- Gespeicherte Verbindungen bleiben eindeutig und prüfbar.
- Mehrere benannte Ebenen und absolute `Z`-Höhen funktionieren ohne Modulneigung.
- Relative Gleishöhen eines Moduls werden korrekt in die Anschlussprüfung einbezogen.
- Ein physisches Exemplar kann in einem gleichzeitig aktiven Aufbau nicht doppelt verwendet werden.
- Hallenfläche, Verfügbarkeit, Normen, Spurweite, Gleisachsen, Elektrik und Höhen werden geprüft.
- Gesperrte Veranstaltungen fixieren ihre aktiven Anlagenplanrevisionen.
- 1.000 Module und 100.000 Detailobjekte sind ohne festes Mengenlimit interaktiv navigierbar.
- Die Desktopoberfläche entspricht den freigegebenen Bildern ohne nicht abgestimmte sichtbare
  Abweichungen.
- Deutsche und englische UI-Texte enthalten keine bekannten Rechtschreib- oder
  Terminologiefehler.

## Nicht-Ziele

- Modulgeometrie oder Gleispläne im Ausstellungseditor verändern.
- Vollständiges 3D-CAD oder eine frei drehbare 3D-Ansicht.
- Neigung oder Querneigung kompletter Module.
- Automatische Erzeugung physischer Exemplare aus einer bloßen Planung.
- Stillschweigende Aktualisierung gesperrter Pläne auf neue Modulrevisionen.
- Cloudbasierte Zusammenarbeit oder öffentlicher Austausch von Aufbauten.
- Ersatz der bestehenden Fahrzeug- und Funktionsplanung einer Ausstellung.
- Harte fachliche Begrenzung auf 1.000 Module oder 100.000 Detailobjekte.
