# RailKeeper Modulbibliothek, Modulformen und normierte Verbindungsseiten

Datum: 2026-09-06
Status: Fachlich und visuell freigegeben, schriftliche Spezifikation zur Prüfung

## Ziel

Der Anlagenbereich folgt der freigegebenen Variante 2: Die Arbeitsfläche steht im Mittelpunkt,
Werkzeuge und Eigenschaften bleiben kompakt am Rand. Das Anlegen und Bearbeiten eines Moduls wird
von einem einfachen Breite-mal-Tiefe-Formular zu einem präzisen Formeditor ausgebaut.

Module werden aus der Bindung an eine einzelne Anlage gelöst und in einer projektweiten Bibliothek
geführt. Eine versionierte Modulvorlage beschreibt Konstruktion, Geometrie, Anschlüsse und
Gleisplanung. Davon getrennte Modulexemplare repräsentieren konkrete physische Anlagenteile mit
eindeutiger Kennung, Eigentümer und Zustand. Baugleiche Exemplare können dieselbe Vorlage nutzen.

Ein Modul darf eine freie Kontur und frei wählbare Maße haben. Modulnormen gelten nicht pauschal
für das ganze Modul, sondern ausschließlich für ausdrücklich markierte Verbindungsseiten. Ein
Modul kann keine, eine oder mehrere Verbindungsseiten besitzen. Jede Verbindungsseite kann ein
anderes, versioniertes Normprofil verwenden.

Der Leitgedanke lautet:

> Freie Form, normierte Köpfe.

Die normale Anlagenplanung darf mit Vorlagen und noch nicht gebauten Modulen arbeiten. Die
Ausstellungsplanung verwendet konkrete Modulexemplare und wird fachlich im bestehenden Bereich
**Ausstellung** geführt. Ihr interaktiver, mehrstufiger Aufbau ist in der begleitenden
[Spezifikation zur Ausstellungsanlagenplanung](./2026-09-06-exhibition-layout-planning-design.md)
beschrieben.

## Verbindliche visuelle Referenzen

Die spätere Umsetzung folgt den beiden freigegebenen Bildern exakt. Sie sind keine lose
Stilinspiration, sondern die verbindliche Desktopreferenz für Aufbau, Hierarchie, Flächen,
Werkzeugpositionen, Inspektor, Typografie, Abstände, Farben und Interaktionsdichte.

### Anlagenübersicht, Variante 2

`C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-ffa8ae91-5053-4c73-b629-83a701194a95.png`

SHA-256:
`3CAB446FD1FB6D3DC24391E7FF6C050CCEFF841EC5F429CF410BBF943522ABA5`

### Moduleditor, freie Form und normierte Köpfe

`C:\Users\droth\.codex\generated_images\01a07612-659f-7632-8706-cb2b937377fa\exec-be12f866-ee67-43f9-820c-f7c5131e29ec.png`

SHA-256:
`C387A13776ED3C839BEACAEED0D1B7A186C2749D56CF9BD29D5926E6EA8F1DA2`

Funktional notwendige Zustände, Barrierefreiheit und responsive Umbrüche ergänzen diese Referenzen,
ändern aber nicht eigenmächtig ihre visuelle Richtung. Eine unvermeidbare sichtbare Abweichung wird
vor ihrer Umsetzung mit dem Benutzer abgestimmt.

Vor visuellen Implementierungsarbeiten werden Existenz und SHA-256-Prüfsumme beider Dateien
kontrolliert. Fehlt eine Datei oder stimmt die Prüfsumme nicht, wird die Arbeit angehalten und das
Original erneut angefordert. Eine Rekonstruktion aus Erinnerung ersetzt keine Referenzdatei.

## Festgelegte Produktentscheidungen

- Die visuelle Basis ist Variante 2 mit großer Zeichenfläche, schmaler Werkzeugleiste und
  kontextbezogenem Inspektor.
- Modulvorlagen und Modulexemplare gehören zu einer projektweiten Bibliothek, nicht zu einer
  einzelnen Anlage.
- Eine Vorlage enthält die versionierte technische Definition. Ein Exemplar enthält Identität,
  Eigentümer, Zustand und die verwendete Vorlagenrevision.
- Ein physisches Exemplar darf in beliebig vielen gespeicherten Aufbauten vorkommen, aber nur einmal
  innerhalb eines gleichzeitig verwendeten Aufbaus.
- Eine Modulvorlage darf in einem normalen Anlagenentwurf mehrfach vorkommen. Sobald konkrete
  Exemplare zugeordnet werden, muss jede physische Verwendung eindeutig sein.
- Die normale Anlagenplanung und die Ausstellungsanlagenplanung bleiben getrennte Arbeitsabläufe
  mit gemeinsamem Modul-, Geometrie-, Norm- und Verbindungsmodell.
- Die Modulgrundfläche wird als geschlossene Kontur in lokalen Millimeterkoordinaten gespeichert.
- Gerade und kreisbogenförmige Kontursegmente sind zulässig. Frei geformte Kurven und ein
  allgemeiner CAD-Constraint-Solver gehören nicht zum ersten Ausbau.
- Jede Konturkante besitzt eine stabile Identität. Normzuordnungen dürfen nie von einem flüchtigen
  Array-Index abhängen.
- Nur eine gerade Konturkante kann als Verbindungsseite markiert werden. Soll nur ein Teil einer
  längeren Kante verbunden werden, teilt der Benutzer die Kante vorher.
- Ein Modul besitzt null bis beliebig viele Verbindungsseiten.
- Eine Norm wird pro Verbindungsseite und pro Normbereich zugeordnet. Mechanik,
  Gleisübergang und Elektrik/Digital bleiben getrennte, kombinierbare Bereiche.
- Unterschiedliche Verbindungsseiten desselben Moduls dürfen unterschiedliche Normen und Versionen
  verwenden.
- Die übrige Kontur bleibt vollständig frei. Eine Normzuordnung formt oder korrigiert das Modul
  niemals stillschweigend um.
- Normprüfungen sind transparent und nicht destruktiv. Abweichungen dürfen gespeichert werden und
  erscheinen als Abweichung oder Warnung. Ungültige Grundgeometrie kann nicht gespeichert werden.
- Ein positives Prüfergebnis bedeutet nur, dass die in RailKeeper erfassten, planungsrelevanten
  Werte zum gewählten Profil passen. Es ist keine Fertigungsabnahme oder Zertifizierung.
- Bestehende Module, die nur Breite, Tiefe und Konturpunkte enthalten, bleiben gültig und erhalten
  keine automatische Normzuordnung.
- Modulplatzierungen verwenden `X`, `Y`, eine absolute beziehungsweise ebenenbezogene `Z`-Höhe und
  eine Drehung um die Hochachse. Komplette Module erhalten keine Neigung oder Querneigung.
- Höhenverläufe innerhalb eines Moduls bleiben relative Eigenschaften seiner Gleisplanung.
- Die Umsetzung erfolgt vollständig inline und ohne Subagenten.
- Alle sichtbaren deutschen und englischen Texte durchlaufen vor Abschluss eine gesonderte
  Rechtschreib- und Konsistenzprüfung.

## Ausgangslage

Der bestehende Modulbereich bietet eine Tabelle und ein kompaktes Formular für Name, Typ,
Eigentümer, Breite und Tiefe. Daneben bestehen bereits wichtige technische Grundlagen:

- Konturpunkte können über eine eigene API gespeichert und im digitalen Zwilling verschoben
  werden.
- Modulports für Gleis, Strom, Digital, Rückmeldung, Zubehör und Sonstiges besitzen Position,
  Richtung und einen optionalen Schnittstellenschlüssel.
- Modulversionen schützen Schreibvorgänge vor konkurrierenden Änderungen.
- Konfigurationen und der digitale Zwilling verwenden Modulkonturen und Ports bereits für
  Darstellung und Einrasten.
- Jedes heutige `LayoutUnit` gehört fest zu genau einer Anlage. Dadurch kann dasselbe physische
  Modul nicht ohne Kopie in mehreren Anlagen oder Ausstellungsaufbauten verwendet werden.
- Platzierungen in einer Anlagenkonfiguration enthalten derzeit nur `X`, `Y` und Drehung. Die
  Gleisplanung kennt zwar relative Gleishöhen, die Modulplatzierung selbst besitzt aber keine
  Ebene oder `Z`-Koordinate.

Diese Funktionen werden im neuen Editor zusammengeführt. Es entsteht kein paralleles zweites
Kontur- oder Portmodell. Der heutige Rechteck-Fallback bleibt für alte Datensätze erhalten, bis eine
Kontur im neuen Editor gespeichert wurde.

## Betrachtete Ansätze

### 1. Stabile Konturkanten mit Normprofilen pro Verbindungsseite

Dies ist der ausgewählte Ansatz. Die freie Grundform und die genormte Schnittstelle bleiben sauber
getrennt. Mehrere Köpfe, gemischte Normen, Übergangsmodule und Sonderformen lassen sich ohne
Sonderfälle abbilden.

### 2. Eine Normvorlage für das vollständige Modul

Eine vollständige Vorlage wäre bei Standardrechtecken bequem, würde aber Breite, Tiefe,
Zwischenkontur und Anzahl der Köpfe unnötig koppeln. Übergangs-, Eck-, Abzweig- und
Mehrkopfmodule würden schnell eigene Vorlagentypen benötigen. Dieser Ansatz wird verworfen.

### 3. Freie Kontur mit unstrukturierten Normnotizen

Freitext wäre sehr flexibel, könnte aber weder die Kompatibilität zweier Köpfe prüfen noch Ports,
Gleisachsen oder Maße zuverlässig an einer Seite verankern. Notizen bleiben ergänzend möglich,
ersetzen aber keine strukturierte Normzuordnung.

### Projektweite Verwendung von Modulen

#### 1. Modulvorlage und physisches Exemplar trennen

Dies ist der ausgewählte Ansatz. Die Vorlage bündelt wiederverwendbare Geometrie, Normen, Ports und
Gleisplanrevisionen. Jedes physische Exemplar erhält eine eigene Identität, einen Eigentümer, einen
Zustand und eine festgelegte Vorlagenrevision. Normale Anlagenentwürfe dürfen Vorlagen verwenden;
Ausstellungsaufbauten verwenden ausschließlich Exemplare.

#### 2. Bestehende Anlagenmodule nur projektweit verschieben

Dieser Ansatz benötigt weniger neue Entitäten, vermischt aber wiederverwendbare Konstruktion und
physisches Eigentum. Baugleiche Module müssten vollständig dupliziert werden, und Änderungen wären
nur schwer reproduzierbar.

#### 3. Module für jede Anlage kopieren

Kopien halten die heutige Anlagenbindung aufrecht, führen aber bei Geometrie, Normen,
Eigentümerdaten und Gleisplanung zu auseinanderlaufenden Daten. Wechselnde Ausstellungsaufbauten
wären unnötig aufwendig.

## Bedienkonzept

### Anlagen-Workspace

Der bestehende Anlagenkopf mit Anlage, Spurweite, Status und Versionsinformation bleibt kompakt
oberhalb der Arbeitsfläche. Die vorhandenen Arbeitsbereiche Übersicht, Gleisplanung, Module,
Aufbauten, Technik, Wartung und Dokumente bleiben erreichbar. Der aktive Bereich erhält mehr
vertikalen Platz und vermeidet verschachtelte Kartenflächen.

Der projektweite Modulbestand ist zusätzlich als eigenständige Bibliothek erreichbar. Der
kontextuelle Bereich **Module** einer Anlage verwendet dieselbe Bibliothek, filtert zunächst aber
auf die dort eingesetzten oder geplanten Vorlagen und Exemplare. Ein Wechsel auf **Alle Module**
entfernt diesen Filter, ohne Datensätze zu kopieren.

Im Bereich **Module** besteht die Ansicht aus:

1. einer kompakten Modulauswahl mit Suche, Status und Prüfhinweisen;
2. einer großen, zoombaren Zeichenfläche;
3. einer schmalen Werkzeugleiste links;
4. einem kontextbezogenen Inspektor rechts;
5. einer ruhigen Statuszeile für Koordinaten, Raster, Zoom und Prüfstatus.

Auf schmalen Ansichten wird die Modulauswahl zum Auswahlfeld. Werkzeugleiste und primäre Aktion
bleiben erreichbar, der Inspektor öffnet als seitliches Panel oder Bottom Sheet. Die Zeichenfläche
erzeugt kein unkontrolliertes horizontales Scrollen der Seite.

### Modulvorlage und Modulexemplar

Der Form- und Kopfeditor bearbeitet ausschließlich eine Modulvorlage beziehungsweise eine neue
Vorlagenrevision. Ein physisches Exemplar zeigt die technische Vorlage schreibgeschützt und ergänzt
Kennung, Eigentümer, Zustand, Standort, Verfügbarkeit und Notizen. Die Aktion **Weiteres Exemplar**
legt ein baugleiches Exemplar derselben Revision an, ohne Geometrie oder Gleisplan zu duplizieren.

Eine Vorlage besitzt einen veränderbaren Entwurf und unveränderliche freigegebene Revisionen. Nur
eine freigegebene Revision kann einem bestätigten physischen Exemplar oder einem gesperrten Aufbau
zugrunde liegen. Dadurch erzeugt nicht jede einzelne Zeichenbewegung eine dauerhafte Revision.

Änderungen an einer bereits verwendeten Vorlage erzeugen eine neue Revision. Exemplare wechseln
nicht automatisch. Der Benutzer erhält eine Vergleichsansicht und kann jedes Exemplar bewusst auf
die neue Revision umstellen. Gesperrte oder veröffentlichte Aufbauten behalten ihre bisherige
Revision.

### Modul anlegen

**Neues Modul** öffnet den Editor mit einer kleinen Startauswahl, nicht mit einem langen Wizard:

- Rechteck;
- Trapez;
- L-Form;
- Kreissegment;
- freie Kontur.

Die Vorlage erzeugt nur eine editierbare Ausgangskontur. Name, Typ, Eigentümer und optionale Notiz
stehen im Abschnitt **Grunddaten**. Der Benutzer kann sofort zeichnen und die Werte anschließend
präzisieren.

### Form und Maße

Die Werkzeugleiste enthält Auswahl, Punkt hinzufügen, Kante teilen, Gerade, Kreisbogen, Maß und
Verbindungsseite markieren. Der Inspektor zeigt abhängig von der Auswahl:

- für einen Punkt: X- und Y-Koordinate;
- für eine Gerade: Länge und Winkel sowie die Wahl des festen Endpunkts;
- für einen Kreisbogen: Radius, Richtung und Bogenwinkel;
- ohne Auswahl: Begrenzungsbreite, Begrenzungstiefe, Umfang und Fläche.

Alle gespeicherten Geometriedaten verwenden Millimeter. Die Oberfläche formatiert Dezimalwerte
sprachabhängig, überträgt sie aber numerisch. Raster, Fangabstand, Hilfslinien, Spiegeln und Drehen
sind Bedienhilfen. Sie verändern keine Geometrie ohne eine sichtbare Benutzeraktion.

Die erste Ausbaustufe ist bewusst kein vollwertiges CAD-System. Sie unterstützt exakte Punkte,
Längen, Winkel und Kreisbögen, aber keine allgemeinen Gleichungsbedingungen, Bézierkurven oder
automatische Zwangsbedingungsauflösung.

### Verbindungsseiten markieren

Der Benutzer aktiviert **Verbindungsseite** und wählt eine gerade Konturkante. Die Kante erhält eine
eindeutige Konturmarkierung sowie eine textliche Kennzeichnung, zum Beispiel **Kopf A**. Farbe ist
nur eine zusätzliche Hilfe.

Der Inspektor zeigt für die gewählte Seite:

- Name und optionale Beschreibung;
- gemessene Seitenlänge und Orientierung;
- mechanisches Kopfprofil;
- Profil für den Gleisübergang;
- Profil für Elektrik oder Digitaltechnik;
- zugehörige, an der Kante verankerte Ports;
- Quelle, Version und Prüfstatus jedes Profils.

Eine Verbindungsseite kann ohne Normprofil angelegt werden. Sie erscheint dann als **frei
definiert**. Die Auswahl eines Profils zeigt zuerst die Sollwerte und die daraus entstehenden
Abweichungen. Die Aktion **Normmaße übernehmen** erhält eine Vorschau und eine ausdrückliche
Bestätigung, bevor sie etwa die Kantenlänge oder Portpositionen anpasst.

Beim Löschen oder Zusammenführen einer markierten Kante fordert RailKeeper eine Entscheidung an,
weil sonst die Verbindungsseite verloren ginge. Beim Teilen einer bereits markierten Kante muss der
Benutzer festlegen, welcher neue Abschnitt die Zuordnung übernimmt. Normen und Ports werden nie
stillschweigend verworfen.

### Kopfansicht und Normprüfung

Die Draufsicht bleibt die primäre Zeichenfläche. Für eine markierte Verbindungsseite kann eine
kompakte **Kopfansicht** geöffnet werden. Sie zeigt das gewählte Profil in seinem lokalen
Koordinatensystem:

- horizontale Position entlang des Kopfes;
- Höhe relativ zur im Profil definierten Bezugsebene;
- Gleisachsen und Schienenoberkante, soweit das Profil sie definiert;
- Bohrungen und Befestigungspunkte als Referenz;
- elektrische oder digitale Schnittstellen als beschriftete Referenzpunkte.

RailKeeper bildet damit die planungsrelevante Schnittstelle ab, nicht die komplette 3D-Konstruktion
des Modulkastens. Die Kopfansicht dient als prüfbare Referenz und kann später für Maßblätter oder
Berichte verwendet werden.

## Geometriemodell

### Kontur

Eine Modulkontur ist ein geschlossener, nicht selbstschneidender Ring aus mindestens drei
Segmenten. Jedes Segment besitzt eine stabile `edgeId`, eine Reihenfolge und eine Geometrieart:

- `line` für eine Gerade;
- `arc` für einen Kreisbogen mit eindeutigem Radius und Drehsinn.

Die Endposition eines Segments entspricht der Startposition des folgenden Segments. Der letzte
Abschnitt schließt auf den ersten. Die API liefert die stabile Identität mit aus und übernimmt sie
bei Änderungen. Neue oder geteilte Segmente erhalten neue IDs. Eine reine Arrayposition darf
nicht als dauerhafte Referenz verwendet werden.

Mindestens folgende Regeln werden serverseitig geprüft:

- mindestens drei Segmente und eine geschlossene Kontur;
- ausschließlich endliche, begrenzte Millimeterwerte;
- keine Segmente mit praktisch null Länge;
- keine Selbstüberschneidung;
- keine widersprüchliche Kreisbogen-Geometrie;
- kein Flächeninhalt von praktisch null;
- eindeutige Segment-IDs und konsistente Reihenfolge.

Breite und Tiefe werden künftig als berechnete Begrenzungsmaße der Kontur zurückgegeben. Die
bestehenden Felder `widthMm` und `heightMm` bleiben aus Kompatibilitätsgründen erhalten, werden nach
einer Geometrieänderung aber aus der Kontur abgeleitet. Sie sind nicht länger eine zweite,
unabhängige Quelle der Wahrheit.

### Verbindungsseite

Eine Verbindungsseite referenziert genau eine stabile, gerade `edgeId`. Ihr lokales
Koordinatensystem wird aus der Kante abgeleitet:

- Ursprung in der Kantenmitte;
- positive Querachse entlang der Kante;
- positive Normalenrichtung nach außen;
- Höhenachse senkrecht zur Draufsicht in der Kopfansicht.

Damit bleiben Normmaße und verankerte Ports stabil, wenn das komplette Modul verschoben oder
gedreht wird. Wird die Kante geometrisch geändert, berechnet RailKeeper die Prüfwerte neu.

Eine Verbindungsseite enthält mindestens:

- stabile ID und Modul-ID;
- referenzierte Kanten-ID;
- Name, Beschreibung und Sortierung;
- optionale Profilrevision je Normbereich;
- berechneten Prüfstatus;
- Erstellungs- und Änderungszeit.

Der Prüfstatus wird aus der aktuellen Geometrie, den Ports und den unveränderlichen
Profilrevisionen berechnet. Er wird nicht als frei editierbare Behauptung gespeichert.

### Ports

Das vorhandene Portmodell bleibt erhalten. Ein Port kann zusätzlich an eine Verbindungsseite
gebunden werden. Für gebundene Ports werden Position und Richtung relativ zur Kante gespeichert,
zum Beispiel als Abstand von der Kantenmitte und als Richtung zur Außennormale. Die absolute
X/Y-Position wird für bestehende Verbraucher weiterhin zurückgegeben und aus der Bindung
abgeleitet.

Gleis-, Strom- und Digitalports können aus einem Normprofil vorgeschlagen werden. Das Annehmen
eines Vorschlags ist eine ausdrückliche Aktion. Zubehör- und sonstige Ports dürfen weiterhin frei
im Modul liegen. Dadurch bleiben alte Daten und nicht genormte interne Schnittstellen nutzbar.

## Normmodell

### Grundstruktur

Normdaten werden als versionierte Referenzprofile geführt. Eine Profilrevision ist nach ihrer
Verwendung unveränderlich. Eine Änderung erzeugt eine neue Revision, damit bestehende Module nicht
unbemerkt ihre Bedeutung ändern.

Ein Profil enthält:

- Herausgeber oder Organisation;
- Normsystem und Profilname;
- veröffentlichte Version beziehungsweise Ausgabestand;
- Normbereich `mechanical`, `track`, `electrical` oder `digital`;
- passende Spurweiten und optionale Kompatibilitätsgruppe;
- strukturierte Sollwerte mit Einheiten und Toleranzen;
- lesbare Beschreibung und Prüfhilfen;
- Quellentitel, Quell-URL und Verifikationsstand;
- Status `draft`, `verified` oder `deprecated`.

Mechanik, Gleisübergang und elektrische Schnittstelle werden getrennt modelliert, weil sie in der
Praxis nicht zwingend gemeinsam versioniert oder kombiniert werden. Pro Verbindungsseite ist
höchstens eine aktive Profilrevision je Normbereich zulässig. Die Kombination wird zusätzlich auf
Spurweite und ausgewiesene Kompatibilität geprüft.

Die strukturierten Regeln verwenden versionierte, serverseitig typisierte Schemas. Ein beliebiges,
ungeprüftes JSON-Regelwerk ist keine öffentliche Erweiterungsschnittstelle.

### Quellen und Starterprofile

Die Normbibliothek zeigt immer Herkunft und Version. Als fachliche Grundlage dienen offizielle
Quellen. [NEM 900](https://morop.org/downloads/nem/de/nem900_d.pdf) beschreibt das Grundprinzip
standardisierter Verbindungsflächen für Modulanlagen. Konkrete Systeme wie die
[FREMO-TT-Modulnorm](https://www.fremo-net.eu/en/modulsysteme/baugroesse-tt/tt-modulnorm) führen
mechanische und elektrische Festlegungen sowie konkrete Kopfprofile.

RailKeeper liefert nur Profile mit, deren Zahlenwerte gegen eine benannte offizielle Ausgabe
geprüft wurden und deren Nutzung im Projekt geklärt ist. Ein Profil ohne abgeschlossene Prüfung
wird als Entwurf gekennzeichnet und nie als verbindliche Normvorgabe dargestellt. Es findet kein
automatischer Abruf fremder Normseiten während der Nutzung statt.

Zusätzlich können Planer und Administratoren eigene Profile anlegen oder ein bestehendes Profil in
eine benutzerdefinierte Familie kopieren. Mitgelieferte Profile sind schreibgeschützt.
Benutzerdefinierte Änderungen erzeugen ebenfalls neue Revisionen.

### Prüfregeln

Die erste Stufe prüft nur Tatsachen, die RailKeeper verlässlich kennt:

- Kantenlänge gegen Sollmaß und Toleranz;
- Spurweite und zulässige Profilkombination;
- Anzahl, Abstand und Richtung der gebundenen Gleisports;
- geforderte elektrische oder digitale Ports und Schnittstellenschlüssel;
- paarweise Kompatibilität zweier im Aufbau verbundener Seiten;
- Vollständigkeit von Quelle und Versionsbezug.

Bohrbilder, Höhen und weitere Kopfdetails werden in der Kopfansicht als versionierte Referenzwerte
gezeigt. Sie werden erst dann automatisch als Istmaß geprüft, wenn RailKeeper dafür später echte,
vom Benutzer erfasste Istwerte besitzt. Dies verhindert eine scheinpräzise Konformitätsaussage.

Prüfergebnisse besitzen Schweregrad, Code, betroffenen Normbereich, Sollwert, Istwert und eine
konkrete Behebungshilfe. Die Oberfläche verwendet Symbol, Text und Farbe. Mögliche Zustände sind:

- **Passend**: alle erfassten, planungsrelevanten Regeln erfüllt;
- **Abweichung**: mindestens ein bekannter Istwert liegt außerhalb der Toleranz;
- **Unvollständig**: für eine Regel fehlt ein erfassbarer Wert oder eine Zuordnung;
- **Nicht geprüft**: kein Profil oder nur ein Profilentwurf zugeordnet;
- **Veraltet**: die zugeordnete Revision ist als überholt markiert, bleibt aber reproduzierbar.

Normabweichungen verhindern das Speichern nicht. Sie verhindern lediglich eine irreführende
Darstellung als passend und bleiben in Liste, Editor und Aufbau sichtbar.

## Datenhaltung und Migration

Die genaue Tabellenteilung wird im Implementierungsplan festgelegt. Das persistierte Modell muss
mindestens folgende Identitäten getrennt abbilden:

- projektweite Modulvorlagen und unveränderliche Vorlagenrevisionen;
- physische Modulexemplare mit eindeutiger Kennung und referenzierter Vorlagenrevision;
- stabile Kontursegmente einer Vorlagenrevision;
- Verbindungsseiten mit Referenz auf ein gerades Kontursegment;
- Normfamilien und unveränderliche Profilrevisionen;
- Zuordnungen zwischen Verbindungsseite und Profilrevision je Normbereich;
- optionale Bindung vorhandener Modulports an eine Verbindungsseite;
- Verwendungen einer Vorlage oder eines Exemplars in normalen Anlagenentwürfen;
- Verwendungen physischer Exemplare in Ausstellungsanlagenplänen.

Die Migration ist vorwärtsgerichtet und deterministisch:

1. Jedes bestehende `LayoutUnit` wird zu einer projektweiten Modulvorlage mit erster Revision.
2. Ein zugehöriges, zunächst als **Bestand prüfen** markiertes Exemplar bewahrt Eigentümer und
   bisherige Identität, ohne ungeprüft eine Ausstellungsverfügbarkeit zu behaupten.
3. Bestehende Anlagenkonfigurationen referenzieren anschließend die migrierte Vorlagenrevision.
   Position, Reihenfolge und Gleisplanrevision bleiben unverändert. Das zu prüfende Exemplar wird
   dadurch nicht ungeprüft als tatsächlich verwendetes physisches Modul behauptet.
4. Bestehende Konturpunkte werden in gerade Segmente mit stabilen IDs überführt.
5. Module ohne gespeicherte Kontur erhalten ihren bisherigen Breite-mal-Tiefe-Rechteck-Fallback.
6. Es werden keine Verbindungsseiten und keine Normen geraten oder automatisch zugeordnet.
7. Bestehende Ports behalten absolute Position und Richtung und bleiben zunächst ungebunden.
8. Breite und Tiefe bleiben im API-Modell erhalten und werden bei neuen Geometrien abgeleitet.

Die bisherige Anlagenzuordnung wird als Verwendung in der ursprünglichen Anlage erhalten, nicht als
Eigentumsgrenze der neuen Bibliothek. Erst nach einer ausdrücklichen Bestandsprüfung kann ein
migriertes Exemplar in der Ausstellungsplanung als verfügbar angeboten werden.

Die Vorlagenrevision schützt die gesamte technische Definition als Aggregat. Kontur,
Verbindungsseiten, Profilzuordnungen und gebundene Ports werden in einer Transaktion mit
`expectedVersion` geschrieben. Dadurch kann kein Kopf auf eine bereits geänderte oder gelöschte
Kante zeigen. Das Exemplar besitzt eine eigene Version für Eigentümer-, Zustands- und
Verfügbarkeitsdaten.

## API-Vertrag

Frontend, Backend und `openapi/railkeeper.yaml` werden gemeinsam erweitert. Der neue Editor erhält
einen aggregierten Lese- und Schreibvertrag für Modulgeometrie. Er umfasst:

- Kontursegmente mit stabilen IDs und Geometrieart;
- Verbindungsseiten;
- Profilzuordnungen;
- gebundene Ports;
- berechnete Maße und Prüfergebnisse;
- aktuelle Vorlagenrevision und Aggregatversion.

Zusätzliche Endpunkte liefern die filterbare Normbibliothek, einzelne Profilrevisionen sowie das
Anlegen und Revisionieren eigener Profile.

Die Modulbibliothek erhält getrennte, paginierte Verträge für Vorlagen, Vorlagenrevisionen und
physische Exemplare. Suche und Filter decken mindestens Kennung, Name, Eigentümer, Spurweite,
Normprofil, Zustand, Verfügbarkeit und Archivstatus ab. Anlagen- und Ausstellungsendpunkte
referenzieren diese projektweiten IDs, statt vollständige Moduldaten zu kopieren.

Der bestehende Punktkontur-Endpunkt bleibt zunächst kompatibel. Sobald eine Kontur referenzierte
Verbindungsseiten besitzt, darf ein alter, identitätsloser Komplettaustausch keine Kanten löschen.
Der Server antwortet in diesem Fall mit `409 Conflict` und verweist auf den aggregierten
Geometrievertrag. Eine bestehende Zuordnung darf nie still durch einen Legacy-Aufruf verschwinden.

Fehlerantworten unterscheiden mindestens ungültige Geometrie, verwaiste Kantenreferenz,
Profilinkompatibilität, Versionskonflikt und fehlende Berechtigung. Schreibende Aufrufe bleiben
CSRF-geschützt.

## Aufbau, digitaler Zwilling und Gleisplanung

Eine Verbindungsseite wird im Aufbau als starre Andockfläche verwendet. Beim Einrasten richtet
RailKeeper die gewählten Seiten gegeneinander aus und dreht ihre Außennormalen entgegengesetzt.
Gebundene Gleisports dienen anschließend zur Feinprüfung der Gleisachsen.

Normale Anlagenentwürfe dürfen eine Vorlagenrevision direkt platzieren, auch wenn noch kein
physisches Exemplar existiert. Sobald ein konkretes Exemplar zugeordnet wird, bleibt diese
Zuordnung eindeutig sichtbar. Ausstellungsanlagenpläne akzeptieren ausschließlich bestätigte
physische Exemplare.

Eine Platzierung enthält `X`, `Y`, `Z` und die Drehung um die Hochachse. Module bleiben als Ganzes
horizontal. Relative Gleishöhen, Steigungen und Rampen werden weiterhin innerhalb der
Gleisplanrevision des Moduls geführt. Die Verbindungskontrolle addiert Modulhöhe und relative
Port- beziehungsweise Gleishöhe, bevor sie zwei Köpfe vergleicht.

Vor dem Einrasten zeigt die Oberfläche:

- geometrische Passung der Kopfbreite;
- mechanische Profilkompatibilität;
- Spurweite und Gleisübergänge;
- elektrische beziehungsweise digitale Schnittstellen;
- konkrete Abweichungen und unvollständige Angaben.

Der Benutzer kann inkompatible Seiten für einen Entwurf trotzdem verbinden, muss die Abweichung
aber ausdrücklich bestätigen. In Prüfung und veröffentlichten Planständen bleibt sie sichtbar.
Ungebundene Altports funktionieren weiterhin mit dem bestehenden Port-Fangverhalten.

Der digitale Zwilling zeichnet Verbindungsseiten mit Name und Status. In normaler Übersicht bleibt
die Darstellung ruhig. Details erscheinen bei Auswahl oder aktivierter Prüfansicht.

Die interaktive Platzierung, Ebenendarstellung und Skalierung großer Ausstellungsaufbauten gehören
zur begleitenden Ausstellungsplanungsspezifikation. Beide Oberflächen verwenden dieselbe
Transformations- und Verbindungsauswertung.

## Berechtigungen und Audit

Die bestehenden Anlagenrechte bleiben maßgeblich:

- Admin und Planner dürfen Modulgeometrie, Verbindungsseiten, Zuordnungen und eigene Normprofile
  anlegen oder ändern.
- Admin und Planner dürfen Vorlagenrevisionen sowie physische Modulexemplare anlegen und pflegen.
- Admin, Planner, Editor und Viewer dürfen Module, Profile, Quellen und Prüfungen lesen.
- Editor und Viewer bleiben im Anlagenbereich schreibgeschützt.
- Messe erhält keinen allgemeinen Zugriff auf den Anlageneditor.
- Mitgelieferte Normprofile können nicht überschrieben werden. Nur Admin darf sie systemweit
  archivieren oder ersetzen.

Relevante Änderungen werden mit Akteur, Ziel-ID und Zeit protokolliert. Dazu gehören mindestens
Geometrieänderungen, Zuordnung oder Entfernung einer Normrevision, Erstellung einer eigenen
Profilrevision, Wechsel eines Exemplars auf eine andere Vorlagenrevision und das ausdrückliche
Bestätigen einer inkompatiblen Verbindung.

## Konflikte, Fehler und Wiederherstellung

Der Editor hält einen lokalen Entwurf und zeigt ungespeicherte Änderungen sichtbar an. Vor dem
Speichern läuft dieselbe grundlegende Geometrieprüfung wie auf dem Server. Der Server bleibt die
entscheidende Instanz.

Bei `409 Conflict` werden lokale Änderungen nicht verworfen. Die Oberfläche bietet:

- Serverstand laden;
- lokalen Entwurf als Datei oder kopierbaren JSON-Datensatz sichern;
- Unterschiede bei Kontur, Köpfen und Profilrevisionen ansehen;
- nach bewusster Übernahme erneut speichern.

Netzwerk- oder Validierungsfehler lassen den Entwurf auf der Seite bestehen. Ein Wechsel des Moduls
mit ungespeicherten Änderungen benötigt eine Bestätigung.

Eine neue Vorlagenrevision verändert weder physische Exemplare noch bestehende Aufbauten
automatisch. Die Vergleichsansicht benennt geänderte Konturen, Köpfe, Ports, Normrevisionen und
Gleisplanbezüge. Eine Umstellung ist ein eigener, protokollierter Vorgang.

## Backup und Restore

Backups umfassen:

- projektweite Modulvorlagen und sämtliche referenzierten Vorlagenrevisionen;
- physische Modulexemplare mit Kennung, Eigentümer, Zustand und Revisionsbindung;
- stabile Modulgeometrien;
- Verbindungsseiten und Portbindungen;
- die exakt referenzierten Profilrevisionen;
- alle benutzerdefinierten Normfamilien und Revisionen;
- Quellen- und Versionsinformationen.

Mitgelieferte Profile besitzen stabile Schlüssel und Revisionen. Ein Backup führt trotzdem die
referenzierten Revisionsdaten mit, damit eine Prüfung reproduzierbar bleibt. Beim Restore werden
identische mitgelieferte Revisionen über ihren stabilen Schlüssel zusammengeführt. Ein abweichender
Inhalt unter demselben Schlüssel ist ein Validierungsfehler und wird vor dem destruktiven Import
angezeigt.

Alte Backups ohne stabile Kanten oder Normdaten bleiben importierbar. Ihre Punktkonturen werden wie
bei der Datenbankmigration in gerade Segmente überführt. Bestehende Anlagenmodule werden als
**Bestand prüfen** migriert. Es werden weder Normen noch Ausstellungsverfügbarkeiten erfunden.

## Barrierefreiheit und Darstellungsqualität

- Zeichenwerkzeuge sind per Tastatur erreichbar und besitzen sichtbare Fokuszustände.
- Punkt- und Kantenwerte lassen sich vollständig über Formularfelder bearbeiten, ohne Drag-and-drop
  zu benötigen.
- Auswahl, Normstatus und Fehler werden nie nur durch Farbe vermittelt.
- Die Trefferflächen der Zeichenwerkzeuge entsprechen auf mobilen Ansichten mindestens 44 Pixeln.
- Der Editor respektiert `prefers-reduced-motion`; Animationen erklären nur räumliche Übergänge.
- Lange deutsche Profilnamen, große Maßzahlen und mehrere Prüfhinweise dürfen den Inspektor nicht
  überlaufen lassen.
- Lichtes und dunkles Thema verwenden getrennt geprüfte Flächen-, Raster-, Auswahl-, Warn- und
  Fokuskontraste.
- Sichtbare deutsche und englische Texte stammen aus der gemeinsamen i18n-Struktur. Beide
  Sprachfassungen werden auf Rechtschreibung, Terminologie, Platzhalter und abgeschnittene Texte
  geprüft.
- Desktop-Screenshots werden bei gleicher Fenstergröße direkt gegen die freigegebenen Referenzen
  verglichen. Abweichungen bei Hierarchie, Anordnung, Dichte und visueller Gewichtung gelten als
  Fehler, sofern sie nicht durch eine zuvor abgestimmte responsive oder barrierefreie Anpassung
  entstehen.

## Umsetzung in prüfbaren Stufen

Die Umsetzung wird wegen des Umfangs in vier prüfbare Stufen zerlegt. Die
Ausstellungsanlagenplanung folgt anschließend ihrer eigenen Spezifikation und ihren eigenen Plänen.

### Stufe 1: Projektweite Modulidentität

- Modulvorlagen, unveränderliche Vorlagenrevisionen und physische Exemplare;
- Migration bestehender `LayoutUnit`-Daten ohne Verlust ihrer Anlagenverwendung;
- paginierte Bibliotheksverträge, Rollen, Audit und OpenAPI;
- Backup-/Restore-Kompatibilität;
- Bestandsprüfung migrierter Exemplare.

### Stufe 2: Geometrie, Köpfe und Normverträge

- stabile Kontursegmente und Migration;
- Verbindungsseiten und zusammensetzbare Profilrevisionen;
- aggregierter Geometrievertrag und gebundene Ports;
- serverseitige Geometrie- und Normprüfung.

### Stufe 3: Modul- und Kopfeditor

- Variante-2-Arbeitsfläche im Modulbereich;
- Vorlagen, freie Kontur, Geraden, Kreisbögen und exakte Maße;
- Verbindungsseiten markieren und Profile zuordnen;
- Kopfansicht, Quellenanzeige und verständliche Prüfergebnisse;
- Konflikt-, Leer-, Lade- und Fehlerzustände.

### Stufe 4: Normale Anlagenplanung und digitaler Zwilling

- Einrasten über Verbindungsseiten;
- Platzierung von Vorlagen oder bewusst zugeordneten Exemplaren mit `X`, `Y`, `Z` und Drehung;
- Kompatibilitätsvorschau und bestätigte Abweichungen;
- Darstellung in Konfigurationen und digitalem Zwilling;
- Übernahme der ruhigen, flächenorientierten Variante-2-Struktur in die übrigen Anlagenbereiche.

Jede Stufe wird separat gebaut, getestet und visuell geprüft. Die nächste Stufe beginnt erst, wenn
die vorherige Daten- und API-Grundlage stabil ist.

Nach Freigabe dieser Gesamtspezifikation erhält jede Stufe einen eigenen Implementierungsplan. Der
erste Plan behandelt ausschließlich Stufe 1, damit Bibliotheksmodell, Migration und Verträge vor
Geometrieeditor und Ausstellungsplanung verbindlich geprüft werden können. Alle Arbeiten erfolgen
inline ohne Subagenten.

## Verifikation

Automatisierte Backendtests decken mindestens ab:

- Migration jedes bestehenden Anlagenmoduls in Vorlage, erste Revision und zu prüfendes Exemplar;
- Erhalt bestehender Anlagenkonfigurationen und Planrevisionen nach der Identitätsmigration;
- projektweite Wiederverwendung einer Vorlage und eindeutige physische Exemplare;
- bewussten Revisionswechsel eines Exemplars ohne Änderung gesperrter Aufbauten;
- Migration alter Punktkonturen zu stabilen Geradensegmenten;
- Rechteck-Fallback für Module ohne Kontur;
- Selbstüberschneidung, Nullkante, ungültigen Kreisbogen und offene Kontur;
- stabile Kantenreferenzen bei Punktverschiebung und Drehung;
- explizites Verhalten beim Teilen, Löschen oder Zusammenführen markierter Kanten;
- null, eine und mehrere Verbindungsseiten pro Modul;
- mehrere Normbereiche und unterschiedliche Profile an einem Modul;
- unveränderliche Profilrevisionen und Quellmetadaten;
- Grenzwerte der Toleranzprüfung;
- Paarprüfung kompatibler und inkompatibler Seiten;
- atomare Versionskonflikte, Rollen, CSRF und Audit;
- Export, Restore alter Backups und Wiederherstellung referenzierter Revisionen.

Frontendtests decken mindestens ab:

- projektweite Suche und Filterung von Vorlagen und Exemplaren;
- Anlegen mehrerer physischer Exemplare aus derselben Vorlage;
- Bestandsprüfung und bewussten Wechsel einer Vorlagenrevision;
- Anlegen aus jeder Startform und als freie Kontur;
- exakte Bearbeitung per Tastatur und Formular;
- Kante teilen und als Verbindungsseite markieren;
- Profile pro Normbereich auswählen, wechseln und entfernen;
- Vorschau vor **Normmaße übernehmen**;
- sichtbare Abweichung ohne Speicherverlust;
- Konfliktbehandlung mit erhaltenem lokalen Entwurf;
- schreibgeschützte Darstellung für Editor und Viewer;
- leere, ladende und fehlerhafte Normbibliothek.

Normale Projektprüfungen:

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd run build
```

Die visuelle Prüfung umfasst mindestens Desktop und Mobilansicht, helles und dunkles Thema,
Tastaturreihenfolge, lange deutsche Texte, große und kleine freie Konturen, ein Modul ohne Kopf,
ein Modul mit mehreren unterschiedlichen Köpfen sowie eine inkompatible Verbindung. Die
Desktopansicht wird direkt mit beiden freigegebenen Referenzbildern verglichen. Deutsche und
englische Oberflächentexte werden abschließend separat auf Rechtschreibung und Terminologie geprüft.

## Abnahmekriterien

- Ein Benutzer kann ein Modul mit beliebiger, gültiger Kontur aus Geraden und Kreisbögen anlegen.
- Modulvorlagen und physische Modulexemplare sind projektweit verfügbar und nicht Eigentum einer
  einzelnen Anlage.
- Baugleiche Exemplare teilen dieselbe technische Vorlagenrevision, behalten aber eigene Kennung,
  Eigentümer-, Zustands- und Verfügbarkeitsdaten.
- Normale Anlagenentwürfe können Vorlagen verwenden, bevor ein physisches Exemplar existiert.
- Eine neue Vorlagenrevision verändert bestehende Exemplare und Aufbauten nicht stillschweigend.
- Maße lassen sich grafisch und numerisch in Millimetern bearbeiten.
- Breite und Tiefe sind berechnete Begrenzungsmaße und können nicht von der Kontur abweichen.
- Jede gerade Konturkante kann unabhängig als Verbindungsseite markiert werden.
- Ein Modul unterstützt null bis beliebig viele, auch unterschiedlich genormte
  Verbindungsseiten.
- Mechanik, Gleisübergang und Elektrik/Digital lassen sich pro Seite getrennt versioniert zuordnen.
- Normquelle, Ausgabestand, Sollwert, Istwert und Abweichung sind nachvollziehbar.
- Keine Normzuordnung verändert Geometrie oder Ports ohne Vorschau und Bestätigung.
- Kantenänderungen lassen keine verwaisten oder still verlorenen Normzuordnungen zurück.
- Alte Module und Backups bleiben ohne erfundene Normdaten nutzbar.
- Aufbau und digitaler Zwilling können kompatible Köpfe erkennen und Abweichungen sichtbar halten.
- Die Oberfläche entspricht der dichten, ruhigen Variante 2 und bleibt per Tastatur sowie auf
  mobilen Ansichten bedienbar.
- Die Desktopumsetzung entspricht den freigegebenen Bildern ohne nicht abgestimmte sichtbare
  Abweichungen.
- Deutsche und englische UI-Texte enthalten keine bekannten Rechtschreib- oder
  Terminologiefehler.

## Nicht-Ziele

- Vollständiges 3D-CAD oder eine Fertigungszertifizierung.
- Allgemeine Bézierkurven, parametrische Baugruppen oder ein universeller Constraint-Solver.
- Automatisches Herunterladen oder Interpretieren externer Normdokumente.
- Stillschweigende Aktualisierung bestehender Module auf eine neue Normrevision.
- Neigung oder Querneigung vollständiger Module in einem Aufbau.
- Cloudbibliothek oder öffentlicher Austausch von Modulprofilen.
- Vermischung der normalen Anlagenplanung mit dem veranstaltungsgebundenen Ausstellungsworkflow.
- Ersatz der bestehenden Gleisplanung oder Materialplanung.
- Automatische elektrische Schreibzugriffe auf Digitalzentralen.
