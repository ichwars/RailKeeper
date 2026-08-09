# Anlagen, Zubehör und Gleisplaner

**Datum:** 2026-08-07

**Status:** Fachlich abgestimmt

**Umsetzung:** Stufenweise, beginnend mit dem gemeinsamen Fundament

**Umsetzungsstand 2026-08-09:** Gleisplaner Paket A ist lokal umgesetzt. Es enthält die versionierte,
geprüfte Tillig-TT-Modellgleis-Bibliothek mit dem geraden Gleis G1 (83101, 166 mm),
revisionsgebundene Planobjekte, eine maßhaltige SVG-Arbeitsfläche und Backup-Format 5. Paket A ist
bewusst auf diese eine verifizierte Geometrie begrenzt. Anschluss-Snapping, Planvalidierung,
Materialbedarf und Stückliste folgen in Paket B.

## Ziel

RailKeeper erweitert die bestehende Fahrzeugverwaltung um einen eigenständigen Bereich für Zubehör,
Anlagen und Anlagenplanung. Nutzer können Zubehör lagern, reservieren, einbauen und historisch
nachverfolgen. Private Anlagen und Clubanlagen lassen sich aus wiederverwendbaren Modulen aufbauen.
Ein maßhaltiger 2D-Gleisplan verbindet später die Zeichnung mit Produktkatalog, Bestand, Einbauzustand,
Wartung und Dokumentation.

Der zentrale Produktgedanke ist ein lebender Anlagenzwilling: Ein gezeichnetes Objekt ist zugleich
Planbedarf, konkreter Artikel, möglicher Lagerbestand, Einbauort und wartbares Anlagenobjekt.

## Abgrenzung

RailKeeper plant und dokumentiert Anlagen. Es steuert keine Modellbahn und sendet keine Befehle an
Digitalzentralen, Decoder oder Rückmeldesysteme. Adressen, Ausgänge, Verkabelung und technische
Schnittstellen sind reine Dokumentationsdaten.

Die Erweiterung bleibt lokal, selbst gehostet und SQLite-basiert. Eine Clubanlage ist eine fachliche
Anlagenart innerhalb einer RailKeeper-Installation. Sie führt nicht zu Mandantenfähigkeit, Cloud-Sync
oder öffentlicher Freigabe.

Nicht Bestandteil der ersten Umsetzungsetappe sind:

- ein grafischer Gleisplaner,
- die Tillig-Gleisgeometrien,
- Flexgleisberechnung und Höhenprofile,
- Importe aus AnyRail, SCARM oder WinTrack,
- weitere Gleissysteme,
- digitale Anlagensteuerung.

Diese Fähigkeiten sind Teil der abgestimmten Zielarchitektur und werden in späteren Etappen ergänzt.

## Fachliche Struktur

### Zubehör

Zubehör ist ein eigenständiger Inventarbereich und kein Unterbereich eines Fahrzeugs. Er besteht aus:

- **Produkt:** Herstellerartikel mit Artikelnummer, Kategorie, technischen Daten und Dokumenten.
- **Mengenbestand:** Gleichartige, nicht einzeln verfolgte Artikel oder Verbrauchsmaterial.
- **Einzelobjekt:** Seriennummern-, wartungs- oder lebenslaufrelevantes Gerät, etwa Decoder,
  Weichenantrieb oder Rückmeldemodul.
- **Lagerort:** Benannter, optional hierarchischer physischer Aufbewahrungsort.
- **Reservierung:** Zuordnung eines verfügbaren Bestands zu einem geplanten Bedarf.
- **Einbau:** Historische Zuordnung von Bestand zu Fahrzeug, Anlage, Modul oder später einem konkreten
  Planobjekt.

Ein Mengenartikel kann bei Bedarf in ein Einzelobjekt überführt werden. Das ist vorgesehen, wenn ein
konkretes Exemplar beim Einbau individuell dokumentiert werden soll. Bereits individuell relevante
Geräte werden von Anfang an als Einzelobjekt geführt.

### Anlagen und Module

Eine **Anlage** beschreibt die fachliche Identität einer privaten oder gemeinschaftlich betriebenen
Modellbahnanlage. Sie enthält Stammdaten, Dokumente, Module und Aufbaukonfigurationen.

Eine **Anlageneinheit** besitzt ein eigenes lokales Koordinatensystem. Sie kann sein:

- eine kompakte Grundplatte,
- ein wiederverwendbares Modul,
- ein Segment,
- ein technisch oder räumlich abgegrenzter Anlagenbereich.

Eine kompakte Anlage verwendet dieselbe Technik wie eine Modulanlage und besteht lediglich aus einer
großen Anlageneinheit.

Eine **Aufbaukonfiguration** ordnet Anlageneinheiten für einen konkreten Aufbau an. Sie speichert
Position, Drehung, verwendete Revision und verbundene Modulübergänge. Dadurch können mehrere
Clubaufbauten aus denselben Modulen entstehen, ohne die Modulpläne zu kopieren.

Optionale Angaben wie Eigentümer oder verantwortliches Clubmitglied sind Metadaten. Sie verändern
keine Benutzer- oder Zugriffsrechte.

### Planvarianten und Revisionen

Jede Anlageneinheit kann mehrere Planvarianten besitzen. Eine Variante besteht aus fortlaufenden
Revisionen mit folgenden Zuständen:

- **Entwurf:** bearbeitbar und automatisch gespeichert,
- **zur Prüfung:** optionaler fachlicher Zwischenzustand,
- **veröffentlicht:** unveränderliche, aktive Revision,
- **archiviert:** weiterhin lesbar, aber nicht aktiv.

Eine veröffentlichte Revision wird nie direkt geändert. Eine Bearbeitung erzeugt einen neuen Entwurf.
Eine ältere veröffentlichte Revision kann wieder aktiviert werden. Aufbaukonfigurationen referenzieren
konkrete veröffentlichte Revisionen, damit ein historischer Aufbau reproduzierbar bleibt.

## Getrennte Zustandsachsen

Planungs- und Betriebszustand dürfen nicht in einem einzigen Status vermischt werden.

Der Planungs- und Materialzustand lautet:

- geplant,
- reserviert,
- eingebaut,
- ausgebaut.

Der Zustand eines eingebauten oder individuell geführten Objekts lautet:

- betriebsbereit,
- Wartung fällig,
- defekt,
- unbekannt.

Dadurch kann eine Weiche gleichzeitig eingebaut und defekt sein. Die Übersicht leitet ihre Darstellung
aus beiden Achsen ab und zeigt neben Farbe immer ein Symbol oder eine textuelle Kennzeichnung.

## Materialfluss

Die Planung erzeugt Bedarf, verändert aber nicht automatisch den physischen Bestand:

1. Ein Plan oder eine manuelle Anlagenplanung meldet einen Artikelbedarf.
2. Verfügbarer Bestand kann reserviert werden.
3. Eine Reservierung reduziert die verfügbare Menge, nicht den physischen Lagerbestand.
4. Erst eine ausdrücklich bestätigte Installation ordnet Material einem Einbauort zu.
5. Beim Ausbau wird das Material ins Lager zurückgeführt, in Wartung gegeben, als defekt markiert oder
   ausgemustert.

Für jedes Produkt werden mindestens Gesamtbestand, frei verfügbar, reserviert, eingebaut und fehlend
ausgewiesen. Fehlender Bestand verhindert die Planung oder Veröffentlichung nicht.

Ein Planobjekt kann eine technische Baugruppe beschreiben. Eine Weichenposition kann beispielsweise
Gleis, Unterflurantrieb, Decoder-Ausgang und Laterne bündeln. Adressen und Ausgangsnummern bleiben
Dokumentation und lösen keinerlei Steuerbefehl aus.

## Integration in Fahrzeuge

Ein Zubehör-Einzelobjekt kann in ein Fahrzeug oder in eine Anlage eingebaut werden. Das Gerät wird
nicht als zweiter Decoderdatensatz dupliziert.

Die vorhandenen fahrzeugbezogenen CV-Werte, Funktionen und Decoderdateien bleiben am Fahrzeug, weil
sie dessen konkrete Konfiguration dokumentieren. Das Zubehör-Einzelobjekt verwaltet dagegen
Geräteidentität, Kaufdaten, Garantie, Installationshistorie und aktuellen Ort. Eine optionale Verbindung
verknüpft beide Sichten.

## Benutzeroberfläche

Die Hauptnavigation erhält zwei eigenständige Bereiche:

- **Zubehör:** Produktkatalog, Bestand, Einzelobjekte, Lagerorte, Reservierungen und Einbauhistorie.
- **Anlagen:** private Anlagen, Clubanlagen, Module, Aufbauten und Planvarianten.

Die Anlagenarbeitsmappe verwendet die Register:

`Übersicht | Planer | Module | Aufbauten | Technik | Wartung | Dokumente`

Die Übersicht zeigt den veröffentlichten Plan der gewählten Aufbaukonfiguration. Ein Hovereffekt zeigt
kompakte Informationen. Klick oder Tippen öffnet einen Inspector mit Details und Aktionen. Änderungen
an Planobjekten erfordern einen ausdrücklich aktivierten Bearbeitungsmodus. Damit löst bloßes Erkunden
keine versehentlichen Änderungen aus.

Statuslayer, Filter und eine Legende machen Planung, Reservierung, Einbau, Wartung und Defekte sichtbar.
Die Darstellung darf sich nicht allein auf Farben stützen. Komplexe Konstruktion ist für Desktop
optimiert; Anzeige, Inspektion und einfache Statuspflege bleiben touch-tauglich.

## Maßhaltiger 2D-Gleisplaner

Der Planer arbeitet intern in Millimetern. Maßstab und Spurweite gehören zur Anlage. Jede
Anlageneinheit besitzt ein lokales Koordinatensystem, Aufbaukonfigurationen transformieren dieses in
den Gesamtaufbau.

### Gleisbibliotheken

Produktkatalog und Planungsgeometrie sind getrennt. Eine Geometriedefinition enthält:

- Länge, Radius und Winkel,
- geometrische Anschlusspositionen und Anschlussrichtungen,
- Fahrwege für Weichen und Kreuzungen,
- eine Bibliotheksversion,
- Quelle und Prüfstatus.

Nur geometrisch geprüfte Artikel sind im Planer platzierbar. Ungeprüfte Artikel dürfen bereits im
Katalog und Bestand existieren. Bestehende Pläne referenzieren eine konkrete Geometrieversion und
werden durch Bibliotheksaktualisierungen nicht rückwirkend verändert.

Tillig TT Modellgleis ist die erste integrierte Bibliothek. Ziel ist der vollständige Katalog. Die
Freigabe erfolgt in geometrisch geprüften Paketen statt als ungeprüfter Komplettimport.

### Platzierung und Prüfung

Kompatible Gleisanschlüsse rasten anhand von Position, Richtung und Typ ein. Das Raster unterstützt die
Orientierung, bestimmt aber nicht die fachliche Verbindung. Der Planer erkennt mindestens:

- offene und unverbundene Gleisenden,
- geometrische Überlappungen,
- nicht kompatible Anschlüsse,
- beschädigte oder fehlende Bibliotheksreferenzen,
- nicht auflösbare Modulübergänge.

Entwürfe dürfen Warnungen und vorübergehend ungültige Geometrien enthalten. Fehlerhafte Referenzen oder
intern ungültige Geometrie blockieren die Veröffentlichung. Offene Gleisenden, fehlendes Material und
fachliche Warnungen blockieren sie nicht.

### Flexgleis, Höhen und Konturen

Die spätere Flexgleisfunktion berücksichtigt reale Länge, minimalen Radius, Endrichtungen,
Zwischenpunkte und Übergangsbögen. Sie darf einen optimierten Verlauf vorschlagen, ersetzt aber keine
Benutzerentscheidung.

Mehrere Ebenen und optionale Höhenwerte ermöglichen die Berechnung von Steigungen, Gefällen,
Übergängen, Durchfahrtshöhen und möglichen Kollisionen. Grenzwerte werden je Anlage konfiguriert.

Neben Gleisen unterstützt der Planer Anlagen- und Modulkonturen, Ausschnitte, Gebäudegrundrisse,
Bahnsteige, Landschaftsbereiche, Signale, Zubehörmarker, Texte, Maße und technische Hinweise.

### Modulports

Ein Modulport enthält Position, Richtung und typisierte Schnittstellen. Dazu gehören Gleisübergänge und
rein dokumentierte technische Verbindungen. Kompatible Ports rasten in einer Aufbaukonfiguration ein.
Inkompatible oder unvollständige Ports erzeugen Warnungen. Eine automatische Optimierung der
Modulreihenfolge ist nicht vorgesehen.

## Änderungsvorschau und Veröffentlichung

Vor der Veröffentlichung vergleicht RailKeeper den Entwurf mit der vorherigen veröffentlichten
Revision. Die Vorschau zeigt mindestens:

- hinzugefügte, entfernte und geänderte Planobjekte,
- veränderten Materialbedarf,
- verfügbare, fehlende und neu zu reservierende Artikel,
- neue oder behobene Geometrie- und Anschlusswarnungen,
- betroffene Aufbaukonfigurationen.

Das Veröffentlichen selbst ändert keine Installationen und entnimmt kein Material. Reservieren und
Einbauen bleiben gesonderte, nachvollziehbare Aktionen.

## Berechtigungen

Alle Berechtigungen werden serverseitig geprüft. Schreibzugriffe bleiben CSRF-geschützt und werden wie
bisher auditiert.

| Rolle | Pläne ansehen | Pläne bearbeiten | Veröffentlichen | Bestand pflegen | Einbau bestätigen |
|---|---:|---:|---:|---:|---:|
| Viewer | Ja | Nein | Nein | Nein | Nein |
| Messe | Nur freigegebene Messeansichten | Nein | Nein | Nein | Nein |
| Editor | Ja | Nein | Nein | Ja | Ja |
| Planner | Ja | Ja | Ja | Nur lesen und reservieren | Nein |
| Admin | Ja | Ja | Ja | Ja | Ja |

Der Planner darf Anlagen, Module, Aufbaukonfigurationen und Entwürfe verwalten sowie Revisionen
veröffentlichen. Er darf Bestand lesen und reservieren, aber keine Bestandsmengen korrigieren oder
Installationen bestätigen. Der Editor pflegt den operativen Anlagenzustand, Wartung, Zubehörbestand und
Einbauten, verändert jedoch keine Plangeometrie. Der Admin besitzt beide Rechtebereiche.

Der Zustand „zur Prüfung“ ist optional. RailKeeper erzwingt kein Vier-Augen-Prinzip. Die bestehende
Isolation der Messe-Rolle von allgemeinen Inventarrouten bleibt erhalten.

## Speicherung, Konsistenz und Fehlerbehandlung

Die Erweiterung folgt der bestehenden modularen Monolith-Architektur:

- HTTP, Validierung und Rollenprüfungen in `backend/internal/api`,
- Anwendungsfälle und Transaktionen in `backend/internal/application`,
- fachliche Zustandsregeln in `backend/internal/domain`,
- SQLite-Persistenz und Migrationen in `backend/internal/infrastructure`,
- getrennte Frontend-Features unter `frontend/src/features`.

Zubehörverwaltung, Anlagenverwaltung und Planer bleiben eigenständige Anwendungsbereiche mit expliziten
Schnittstellen. Gemeinsame Identifikatoren verbinden sie, ohne Fahrzeug- oder Planerlogik in zentrale
Dateien zu schieben.

Schreibvorgänge verwenden Transaktionen. Änderbare Entwürfe tragen eine Revisionsnummer für
optimistische Nebenläufigkeitskontrolle. Bei einem Konflikt lädt der Client den aktuellen Stand und
fordert eine bewusste Übernahme oder Wiederholung der Änderung. Stille Überschreibungen sind verboten.

Der Planer speichert Entwurfsänderungen automatisch in der lokalen RailKeeper-Instanz. Rückgängig und
Wiederholen umfasst alle Aktionen der laufenden Editor-Sitzung. Die dauerhafte Revisionshistorie entsteht
erst durch ausdrücklich gespeicherte beziehungsweise veröffentlichte Revisionen.

Das Löschen veröffentlichter Pläne, referenzierter Module oder Anlageneinheiten mit Einbauten benötigt
eine explizite Bestätigung. Wo Historie oder Referenzen erhalten werden müssen, wird archiviert statt
physisch gelöscht.

## Backup, Export und Portabilität

Anlagen, Zubehör, Bibliotheksversionen, Planrevisionen, Dokumente und Einbauhistorien werden in
Anwendungsbackup und Restore aufgenommen. Die bestehende Regel bleibt bestehen: Benutzer, Rollen,
Sitzungen, Rate-Limits, Audit-Logs und Passwortdaten werden nicht importiert oder überschrieben.
Das Backupformat erhält eine neue Version. Restore prüft diese vor jeder Veränderung. Bereits
unterstützte ältere RailKeeper-Backups bleiben importierbar und erzeugen leere neue Fachbereiche.

Das portable Planformat ist ein versioniertes RailKeeper-JSON-Format. Spätere Exportformate sind:

- PDF und SVG für maßstäbliche Pläne,
- PNG für schnelle Weitergabe,
- CSV und XLSX für Stücklisten und Materialbedarf.

Fremdformate werden erst ergänzt, wenn das interne Modell stabil ist.

## Umsetzungsetappen

### Etappe 1: Gemeinsames Fundament

Etappe 1 ist der Umfang des anschließenden Implementierungsplans. Sie liefert:

- die neue Planner-Rolle samt serverseitiger Berechtigungen,
- Zubehörprodukte, Mengenbestand und Einzelobjekte,
- Lagerorte, manuelle Reservierungen und Installationshistorie,
- Installationsziele für Fahrzeuge, Anlagen und Anlageneinheiten,
- Anlagen, Anlageneinheiten und Aufbaukonfigurationen,
- Planvarianten und Revisionsmetadaten ohne grafische Geometrie,
- Listen-, Detail- und Bearbeitungsansichten für Zubehör und Anlagen,
- Auditierung, Backup, Restore und versionierten Datenaustausch für die neuen Daten.

Reservierungen können in Etappe 1 einer Anlage oder Anlageneinheit zugeordnet werden. Die feinere
Zuordnung zu einem Planobjekt folgt mit dem grafischen Planer. Das Datenmodell muss diese spätere
Spezialisierung ermöglichen, darf aber noch keine ungenutzte Geometrieimplementierung enthalten.

### Etappe 2: Interaktiver Anlagenzwilling

Konturen, frei platzierbare technische Positionen, Statuslayer, Hoverinformationen, Inspector,
Wartungsanzeige und Einbauhistorie machen Anlagen und Module interaktiv.

### Etappe 3: Tillig-TT-Gleisplaner

Geprüfte Geometrien für Tillig TT Modellgleis, Platzierung, Einrasten, Planprüfung, Stückliste,
planobjektbezogene Reservierung und Änderungsvorschau bilden den ersten vollständigen Planer.

### Etappe 4: Erweiterte Geometrie

Flexgleisoptimierung, Höhenprofile, Steigungen, Durchfahrtshöhen, Modulports und weitere freie
Planobjekte erweitern den Planer.

### Etappe 5: Bibliotheken und Fremdformate

Weitere Gleissysteme, dokumentierte Bibliotheksimporte und Fremdformatkonverter folgen auf Basis des
stabilen internen Formats.

## Test- und Abnahmestrategie

Jede Etappe ergänzt Tests auf der niedrigsten sinnvollen Ebene.

Für Etappe 1 sind erforderlich:

- Go-Tests für Bestands-, Reservierungs-, Einbau- und Revisionsregeln,
- API-Tests für die vollständige Rollenmatrix und CSRF-geschützte Schreibzugriffe,
- Migrationstests sowie Backup- und Restore-Roundtrips,
- Tests für Mengenbestand und Einzelobjekte einschließlich Einbau und Ausbau,
- Frontendtests für rollenabhängige Aktionen, Fehler- und Leerzustände,
- Abgleich von Backend-Routen, Frontend-API und OpenAPI-Vertrag,
- Produktion-Build des Frontends und vollständige Go-Test-Suite.

Für den Planer kommen später geometrische Golden-Tests, Eigenschaftstests für Transformationen und
Snapping, Bibliotheksfixtures, Interaktionstests sowie visuelle Prüfungen für Hell/Dunkel, Desktop,
Touch und lange deutsche Texte hinzu.

## Abnahme von Etappe 1

Etappe 1 gilt als abgeschlossen, wenn folgende Abläufe über Oberfläche und API funktionieren:

- Ein Admin legt Produkt, Lagerort, Bestand und Einzelobjekt an.
- Bestand kann für eine Anlage oder Anlageneinheit reserviert werden, ohne den Lagerbestand zu
  reduzieren.
- Ein Editor bestätigt Einbau, Ausbau und Zustandswechsel; die vollständige Historie bleibt sichtbar.
- Ein Planner legt Anlage, Module, Aufbaukonfiguration und Planrevisionen an und veröffentlicht eine
  Revision, kann aber keine Installation bestätigen.
- Ein Editor kann operative Anlagendaten pflegen, aber keine Planrevision verändern oder
  veröffentlichen.
- Viewer sehen die freigegebenen Daten nur lesend; Messe erhält keinen neuen allgemeinen
  Inventarzugriff.
- Ein Backup-Restore-Roundtrip stellt alle neuen Fachdaten und Referenzen verlustfrei wieder her.
- Der Restore eines bisher unterstützten älteren Backups bleibt möglich.
- Sicherheitsrelevante Änderungen erscheinen im Audit-Log.

## Erfolgskriterien

Die Gesamtfunktion ist erfolgreich, wenn:

- Zubehör unabhängig von Fahrzeugen vollständig verwaltet werden kann,
- ein konkretes Gerät ohne Datenkopie zwischen Lager, Fahrzeug und Anlage wechseln kann,
- kompakte und modulare Anlagen dasselbe Datenmodell verwenden,
- veröffentlichte Anlagenstände reproduzierbar und frühere Revisionen erhalten bleiben,
- Planung, Reservierung, Einbau und Betriebszustand klar getrennt sind,
- der Tillig-Plan später aus derselben Quelle Stückliste, Bestand und Anlagenstatus ableitet,
- keine Funktion digitale Steuerbefehle erzeugt,
- lokale Backups alle neuen Fachdaten sicher exportieren und wiederherstellen.
