# Änderungsprotokoll

[English](CHANGELOG.md)

Diese Datei dokumentiert alle wesentlichen Änderungen an RailKeeper.

## Unveröffentlicht

## [0.1.20.4] - 2026-09-03

### Geändert

- Ausstellungen erhalten eine separate Druckansicht im A4-Querformat. Sie enthält alle
  gespeicherten Einträge der ausgewählten Veranstaltung mit Bildern, Modell- und Steuerungsdaten,
  Fahrtagen, Verfügbarkeit, Status, vollständigen Funktionstasten und Notizen. Bildschirmfilter
  begrenzen den Ausdruck nicht; Navigation und Bedienelemente werden nicht mitgedruckt.
- Z21 bietet eine zeitlich und mengenmäßig begrenzte, rein lesende Vorschau bekannter
  Lokadressen. Intellibox-Transportauswahl und Z21-Diagnose grenzen verfügbare und noch geplante
  Fähigkeiten deutlicher ab. Diese Abläufe schreiben nichts auf die Zentrale.

### Behoben

- Drucken wartet auf die Bilder und bleibt beim Wechsel zu einer noch nicht geladenen
  Ausstellung gesperrt. Ältere Freitext-Funktionsbelegungen bleiben erhalten. Versteckte
  Zeilenaktionen verursachen in der Ausstellungsansicht keinen Seitenüberlauf mehr.
- ECoS weist unvollständige Kompatibilitätsantworten ab und unterscheidet sie von vollständigen
  Antworten mit noch nicht verifizierter Firmware.
- Parallele Änderungen und Autosaves im Anlagenzwilling bewahren andere ungespeicherte Entwürfe.
  Positionshistorie, Konflikt-Neuladen und verspätete Antworten nach Modulwechseln wurden korrigiert.
- Der Gleisplaner sichert Drag-Vorgänge, Spurweitenprüfung, Materialanalyse und transaktionales
  Snapping ab. Modulport-Vorschauen berücksichtigen ungespeicherte Platzierungen;
  Dimensionsänderungen dürfen vorhandene Ports nicht außerhalb der Modulgrenzen verschieben.

## [0.1.20.3] - 2026-08-28

### Hinzugefügt

- Märklin CS3 und CS3 Plus können ihre Loklisten über die bekannten read-only Endpunkte in den
  Digitalzentralen-Arbeitsbereich einlesen. Diagnose, Vergleich, Konfliktprüfung sowie das Anlegen
  und Zuordnen von RailKeeper-Fahrzeugen bleiben vom Schreiben auf die Zentrale getrennt.
- Fahrzeug-CSV-Dateien erkennen vollständige Sets, Set-Mitgliedschaften und individuelle
  Mitgliedsbezeichnungen. Die Importvorschau zeigt Setänderungen vor der Übernahme ausdrücklich an.
- Stammdaten lassen sich über eine zugängliche Mehrfachauswahl gesammelt deaktivieren.

### Geändert

- Fahrzeug- und Zubehörtabellen verwenden persistente, größenverstellbare Spalten. Mobile
  Fahrzeugansicht, Pflichtfeldkennzeichnung, kontextbezogene Ausstellungssteuerung, Typografie,
  Design-Tokens und reduzierte Bewegung wurden konsolidiert.
- Statusmeldungen der Artikelsuche unterscheiden tatsächlichen Fortschritt klarer von
  unbestimmten Verarbeitungsphasen. Temporäre Arbeitsunterlagen wurden aus `docs/` entfernt und
  die Dokumentationsregeln präzisiert.

### Behoben

- Set-Mitgliedsbezeichnungen und nicht zugeordnete Bestandsfelder bleiben beim CSV-Ersetzen
  erhalten. Set-Zeilen, Auswahl-, Daten- und Aktionsspalten bleiben auch nach Größenänderungen
  ausgerichtet.
- CS3-Fahrzeugentwürfe und externe Zuordnungen bewahren den Anbieter `cs3`, statt fälschlich als
  ECoS-Zuordnung gespeichert zu werden.

### Sicherheit

- CS3-HTTP-Ziele werden vor dem Abruf auf private LAN-Adressen begrenzt. Loopback-, Link-Local-,
  öffentliche und gemischte DNS-Ziele werden abgewiesen, geprüfte DNS-Ergebnisse für den Abruf
  fest gebunden und Weiterleitungen bleiben gesperrt.
- Das Laufzeit-Image erzwingt aktualisierte Alpine-OpenSSL-Pakete.

## [0.1.20.2] - 2026-08-24

### Geändert

- Import- und Exportprofile zeigen Richtung, Status, Bereiche und Format in einer gemeinsamen,
  einheitlich ausgerichteten Tabelle.
- Startaktionen verwenden kompakte, zugängliche Symbole. Lange CSV-Zuordnungen und Prüfergebnisse
  bleiben in den Importdialogen vollständig scrollbar und bedienbar.
- Go-, SQLite-, Frontend- und GitHub-Actions-Abhängigkeiten wurden aktualisiert.

### Behoben

- Tabellenversatz und gequetschte untere Bereiche im Import/Export-Arbeitsbereich wurden beseitigt.
- Die Importprüfung kann auch bei umfangreichen CSV-Zuordnungen bis zur Bestätigung fortgesetzt
  werden.

### Sicherheit

- Administratoren können ausschließlich abgebrochene Transferaufträge nach Bestätigung dauerhaft
  löschen. Die Einschränkung wird serverseitig durchgesetzt und der Vorgang protokolliert.

## [0.1.20.1] - 2026-08-24

### Hinzugefügt

- Fahrzeugsets können ein eigenes Hauptbild erhalten, das in Bestand, Mobilansicht und Setdialogen
  verwendet und über einen eigenen Upload-Ablauf verwaltet wird.
- Der Digitalzentralen-Arbeitsbereich kann geprüfte ECoS-Lokomotiven sicher übernehmen und bei
  ausdrücklicher Freigabe fehlende Lokomotiven kontrolliert auf dem Gerät anlegen.

### Geändert

- Import- und Exportprofile erscheinen gemeinsam mit eindeutiger Richtung, Bereichen, Format und
  Status. Fahrzeug-CSV-Dateien unterstützen alle 62 Transferfelder sowie gespeicherte,
  benutzergeprüfte Spaltenzuordnungen.
- Navigation und persönliche Arbeitsbereichssichtbarkeit sind konsistenter synchronisiert.
- Sicherungen verwenden Formatversion 20 und enthalten die Hauptbildzuordnung von Fahrzeugsets.

### Behoben

- Teilweise und ältere Fahrzeug-CSV-Dateien bewahren bei Ersetzungen alle nicht zugeordneten
  Bestandsfelder. Unvollständige Zuordnungen, unbekannte Wahrheitswerte und widersprüchliche
  Profilangaben werden vor der Übernahme abgewiesen.
- CSV-Schutzzeichen bleiben über Export und Rückimport verlustfrei. JSON-Pakete der Version 1
  werden strikt nach ihrem historischen Schema verarbeitet.
- ECoS-Lesevorgänge, Fahrzeugzuordnung, Schreibkonflikte und Rückmeldungen wurden gegen veraltete,
  unvollständige oder mehrdeutige Gerätezustände gehärtet.

### Sicherheit

- ECoS-Schreibvorgänge bleiben auf explizit freigegebene Abweichungen begrenzt und werden nach dem
  Schreiben am Gerät verifiziert.
- CSV-Formelinjektion, Upload-Reihenfolge und transaktionale Importübernahme werden ohne stille
  Datenänderungen behandelt.

## [0.1.20] - 2026-08-22

### Hinzugefügt

- Die neu gestaltete Übersicht bündelt Bestandskennzahlen, Wartung, Wertentwicklung und
  Bestandsstruktur in kompakten Karten und Diagrammen. Detaildialoge und direkte Filtersprünge
  führen von den Kennzahlen zu den betroffenen Fahrzeugen und Zubehörartikeln.
- Persistente Datentransferprofile und eine nachvollziehbare Auftragshistorie unterstützen
  versionierte Exporte, Importvorschauen, Konfliktentscheidungen und die transaktionale Übernahme
  geprüfter Fahrzeug-, Zubehör- und Messedaten.
- Der Arbeitsbereich **Digitalzentralen** bietet gespeicherte Sitzungen, Lok-Arbeitslisten,
  Soll-Ist-Vergleiche und passive Live-Telemetrie. Schreibvorgänge erfordern eine ausdrückliche,
  zeitlich begrenzte Freigabe und eine abschließende Bestätigung.
- Der Messebetrieb besitzt einen eigenen responsiven Arbeitsbereich für Listen, Einträge,
  Sperrzustände und bewusst bestätigte Konfliktausnahmen.
- Die eigenständig gezeichnete **RailKeeper Werkstatt-Linie** ersetzt alle mitgelieferten
  Funktionsgrafiken. 94 Symbole mit aktiver, inaktiver und druckoptimierter Palette decken die 86
  bekannten ECoS-Funktionsbeschreibungscodes sowie neutrale Rückfalltypen ab.

### Geändert

- Datentransfer, Digitalzentralen und Messe verwenden kompakte, zweisprachige und responsive
  Bedienoberflächen mit app-eigenen Steuerelementen und dauerhaft gespeicherten Arbeitsständen.
- Sicherungen verwenden Formatversion 19, enthalten Datentransferprofile und
  Messe-Konfliktausnahmen und bewahren eigene Symbolschlüssel. Beim Einspielen älterer Sicherungen
  bleibt die aktuell installierte mitgelieferte Symbolbibliothek geschützt.
- ECoS-Codes bleiben ausschließlich als neutrale Kompatibilitätsmetadaten erhalten. RailKeeper
  übernimmt keine Funktionsgrafiken oder grafischen Quelldateien des Herstellers.

### Sicherheit

- Rollen- und Messegrenzen werden für die neuen Arbeitsbereiche serverseitig durchgesetzt.
  Importvorschau, Pfadprüfung, atomare Exportreservierung und transaktionale Übernahme behandeln
  Transferdateien als nicht vertrauenswürdige Eingaben.
- Digitalzentralen lesen standardmäßig nur passiv. Schreibvorgänge bleiben auf geprüfte
  Abweichungen, kurzlebige Freigaben und eine explizite Benutzerbestätigung begrenzt.

## [0.1.19.2] - 2026-08-19

### Geändert

- Fahrzeugsets sind im Bestand standardmäßig eingeklappt. Unterschiedliche Kategorien und
  Gattungen der Mitglieder werden ausdrücklich als verschieden zusammengefasst. Hierarchie,
  Auswahlfelder und Setdialoge folgen dem kompakten, ausgerichteten Layout.
- Die Artikelsuche priorisiert Treffer mit den meisten gefundenen Feldern, verwendet den Score als
  stabilen Gleichstandsentscheid und zeigt auf dem Desktop dezentere Trefferaktionen.

### Behoben

- Setdaten lassen sich im vollständigen Seteditor mit denselben validierten Auswahllisten wie bei
  Einzelfahrzeugen bearbeiten und speichern.
- Setwerte dienen nur noch als Vorgaben beim Anlegen und überschreiben keine Mitgliedsdaten.
  Einzelne Fahrzeuge eines Sets können eigene Kategorie-, Gattungs-, Modell-, Technik-, Lager- und
  Zustandswerte behalten und bearbeiten, ohne das kanonische Set zu verändern.
- Setansicht und Seteditor, Mitgliedsausrichtung, Kontrollkästchen, Hierarchielinien und responsive
  Layouts werden auf Desktop und Mobilgeräten konsistent dargestellt.

## [0.1.19.1] - 2026-08-17

### Geändert

- Der dreistufige Dialog zur Fahrzeuganlage folgt jetzt dem abgestimmten kompakten Desktop- und
  Mobil-Layout. Dazu gehören eine klarere Schrittnavigation, dezente Aktionen, responsive
  Suchtreffer und eine fokussierte Prüfung übernommener Artikeldaten.
- Sets erhalten eine dauerhaft gespeicherte, automatisch vergebene `RK-SET`-Inventarnummer und
  erscheinen in Tabellen-, Karten- und Mobilansicht als übergeordnete Zeile mit sortierten Wagen.

### Behoben

- Die Setanlage bewahrt geprüfte Artikelfelder und Bilder, validiert jeden Wagen, trennt Entwürfe
  anhand des exakten Benutzernamens und beachtet die konfigurierten Quellen der Artikelsuche.
- Gruppierung, Sortierung, optionale Spalten, gefilterter Mitgliederzugriff, Duplizieren, Bearbeiten,
  Löschen, Wiederherstellung und Bestandsbewertung bleiben auch bei Teil- und Altdaten konsistent.
- Kompakte Aktionen, Vorschaubild-Platzhalter, Spaltenausrichtung, Suchfeldbreite, Hierarchielinien
  und schmale Ansichten entsprechen jetzt dem abgestimmten Bestandslayout.

## [0.1.19] - 2026-08-16

### Hinzugefügt

- Die Fahrzeuganlage führt in drei Schritten durch Typ und Grunddaten, optionale Barcode- oder
  Websuche sowie die übrigen Fahrzeugdaten.
- Zusammengehörige Fahrzeuge können als Set mit gemeinsamen Artikel-, Erwerbs-, Lager- und
  Zustandsdaten sowie eigenen Inventar- und Fahrzeugnummern angelegt werden. Sets erscheinen in
  Tabellen- und Mobilansicht als aufklappbare Gruppe.
- Das zweisprachige Handbuch dokumentiert jetzt zusätzlich Import und Export, allgemeine
  Einstellungen sowie die Verwaltung allgemeiner und artikelbezogener Stammdaten.

### Geändert

- Die Setanlage validiert alle Mitglieder vor externen Bildabrufen. Gemeinsame Setdaten bleiben bei
  der Bearbeitung einzelner Mitglieder geschützt, und Setpreise fließen genau einmal in die
  Bestandsbewertung ein.
- Der OpenAPI-Vertrag, Sicherungen und Wiederherstellungen enthalten das Setmodell und seine
  geordneten Mitglieder.

### Behoben

- Ausgewählte Artikelfotos bleiben bei der Setanlage erhalten; ECoS-Entwürfe verwenden weiterhin
  den vollständigen Speicherpfad für Zuordnung, CV-Werte und Funktionen.
- Fahrzeug-Kurzmenüs bleiben auch bei geringer Höhe vollständig erreichbar.
- Die Aktionsspalte der Stammdatenverwaltung bleibt am rechten Tabellenrand sichtbar und verursacht
  keinen horizontalen Seitenüberlauf mehr.

## [0.1.18] - 2026-08-16

### Hinzugefügt

- Die Fahrzeugübersicht bietet pro Benutzer gespeicherte Spaltenauswahl und -reihenfolge,
  zusätzliche Bestandsfilter, eine erweiterte Suche sowie kompakte, aufklappbare Mobilkarten.
- Fahrzeuge erfassen jetzt Erwerb, Kaufpreis und -datum, Lagerung, Zustand und Verpackung. Listen-
  und Kaufwerte werden für Fahrzeuge und Zubehör getrennt ausgewertet und in der Übersicht
  nachvollziehbar zusammengefasst.
- Stammdaten können sicher deaktiviert, reaktiviert und nur dann gelöscht werden, wenn keine
  fachlichen Referenzen bestehen. Herkunft und Status bleiben in Exporten und Sicherungen erhalten.
- Die Windows-Standalone-Ausgabe kann verfügbare Versionen anzeigen und das passende ZIP direkt
  herunterladen. Die Anwendung schützt bestehende Installationen durch dauerhafte Datenpfade,
  validierte SQLite-Sicherungen und automatische Sicherungskopien vor Migrationen.
- Das zweisprachige RailKeeper-Handbuch dokumentiert Einrichtung, Übersicht, Fahrzeuge, Wartung,
  Medien, Decoder/CV, Suche, Zubehör und Ausstellungsbetrieb.

### Geändert

- CSV-Import und -Export umfassen alle skalaren Fahrzeugfelder. Eindeutige deutsche und englische
  Überschriften, zusätzliche Aliase und die bestehende Änderungsvorschau ermöglichen einen
  verlustfreien RailKeeper-CSV-Rundlauf.
- Die Fahrzeugübersicht unterstützt `PluX12`; Aktionsmenüs bleiben in Tabellen- und Mobilansichten
  sichtbar und tastaturbedienbar.

### Behoben

- Kennzahlen der Übersicht bleiben auf breiten Ansichten in einer gemeinsamen Zeile.
- Fahrzeug-Kurzmenüs werden nicht mehr an Tabellen- oder Bildschirmrändern verdeckt.
- Windows-Aktualisierungen trennen Programmpaket und Benutzerdaten, statt Daten im
  Installationsordner durch entpackte Dateien zu gefährden.

### Sicherheit

- Vor Datenbankmigrationen wird eine konsistente SQLite-Sicherung erstellt; unsichere oder
  mehrdeutige Windows-Datenpfade blockieren den Start mit einer verständlichen Diagnose.
- Stammdatenimporte und -änderungen bewahren referenzierte historische Werte und erzwingen die
  serverseitigen Lebenszyklusregeln.

## [0.1.17.6] - 2026-08-15

### Geändert

- Die Konfiguration von Digitalzentralen führt jetzt in vier klaren Schritten von der Adapterwahl
  über Konfiguration und Verbindungstest bis zur bewussten Aktivierung.
- Verbindungstest, Diagnose, technische Details, ECoS-Meldungen und Sicherheitsumfang sind in
  einem gemeinsamen Arbeitsbereich zusammengeführt.
- Adapter können aktiviert, deaktiviert und vollständig entfernt werden. Der ECoS-Live-Monitor
  bleibt an einen erfolgreichen Test und eine aktive Verbindung gebunden.
- Die neue Oberfläche passt sich ohne horizontalen Überlauf an schmale Ansichten an.

## [0.1.17.5] - 2026-08-15

### Geändert

- Der Anlagen-Arbeitsbereich bleibt über den Direktzugriff verfügbar, wird aber nicht mehr in der
  Hauptnavigation angezeigt.

## [0.1.17.4] - 2026-08-14

### Behoben

- Die Spaltenüberschriften für Hersteller, Artikelnummer und Lagerung behalten ihre vorgesehenen
  Breiten und überlagern sich in der konfigurierbaren Zubehör-Tabelle nicht mehr.

## [0.1.17.3] - 2026-08-14

### Hinzugefügt

- Dateien für Repository-Zuständigkeit, Beiträge, Verhaltensregeln, Support, Projektkennzeichen und
  Drittanbieterhinweise wurden ergänzt.
- Die Zubehör-Tabelle bietet eine persistente Spaltenauswahl für Bild, Inventarnummer, Hersteller,
  Artikelnummer, Name, Art, Spur, Bestand und Lagerung. Inventarnummer und Name können nicht
  gleichzeitig ausgeblendet werden.

### Geändert

- Neue RailKeeper-Fassungen stehen unter `AGPL-3.0-only`, damit veränderte, über ein Netzwerk
  angebotene Fassungen offen bleiben und ihren Nutzern den korrespondierenden Quellcode anbieten.
  Bestehende Releases behalten ihre bisherigen Lizenzbedingungen; die AGPL erlaubt weiterhin
  kommerzielle Nutzung.
- Die Projektunterstützung verweist nur noch auf GitHub Sponsors und Ko-fi als freiwillige
  Zuwendungen ohne Vorteile, bezahlten Support, SLA oder besonderen Zugang.
- Die ECoS-Anbindung liest nur noch Lokstammdaten, CV-Werte und statische Funktionsdefinitionen.
  Der Schreibabgleich bleibt nach Prüfung und Bestätigung auf Name, Adresse und Protokoll begrenzt.
- Aktuelle Geschwindigkeit, Richtung, aktive Funktionszustände, Lokbilder, Schaltartikel,
  Fahrwege, S88, Booster und weitere ECoS-Objektmanager wurden aus Abfragen und UI-Verträgen
  entfernt.
- Archivieren, Wiederherstellen und endgültiges Löschen stehen in Tabellen-, Kachel- und
  Kompaktansicht als direkte, beschriftete Symbolaktionen zur Verfügung. Das bisherige
  Drei-Punkte-Menü entfällt.

## [0.1.17.2] - 2026-08-14

### Hinzugefügt

- Der Anlagen-Arbeitsbereich bietet jetzt einen interaktiven digitalen Zwilling, bearbeitbare
  technische Positionen, Modulports, Gleispläne, Revisionsvorschauen, Materialreservierungen und
  den Austausch von Gleisbibliotheken.
- Die Gleisplanung unterstützt geprüfte Tillig-Geometrien, Flexgleise, Übergangsbögen,
  Höhenprofile, freie Planobjekte, Abstandsprüfungen, Steigungsgrenzen und Verbindungsanalysen.
- Die Zubehörübersicht bietet eine persistente Tabellen- und Kachelansicht sowie eine kompakte
  Mobilansicht.
- Administratoren können vollständig unbenutzte Zubehörartikel nach ausdrücklicher Bestätigung
  endgültig löschen.

### Geändert

- Die Anlagen-Navigation ist wieder aktiv und öffnet den erweiterten Arbeitsbereich.
- Aktionen in Zubehörzeilen und -karten verwenden dasselbe tastaturbedienbare Menüverhalten.

### Behoben

- Hersteller importierter Gleise bleiben erhalten und Gleisbibliotheken werden stabil geladen.
- Zubehör-Aktionsmenüs bleiben an Tabellen- und Mobilkartenrändern sichtbar, lange Metadatenfelder
  in Karten bleiben klar getrennt.

### Sicherheit

- Die endgültige Zubehörlöschung wird serverseitig auf Administratoren begrenzt und abgewiesen,
  sobald Bestand, Käufe, Bewegungen, Reservierungen, Einbauten, Historie, Assets oder
  Anlagenreferenzen vorhanden sind.

## [0.1.17.1] - 2026-08-13

### Hinzugefügt

- Die Zubehör-Artikelsuche erfasst beschriftete Gleis-Fachangaben wie Gleissystem, Abmessungen,
  Richtung, Bettung, Anschlüsse und Digitaltauglichkeit als einzeln auswählbare, typisierte
  Vorschläge.
- Die Auswahl einer Spurweite trägt den konfigurierten Maßstab automatisch ein, bis das
  Maßstabsfeld manuell bearbeitet wird.
- Neue oder bisher unverschlagwortete Artikel erhalten synchronisierte Schlagwortvorschläge aus
  Bezeichnung, Hersteller, Artikelart und Unterart, bis das Schlagwortfeld manuell bearbeitet wird.

### Behoben

- Der Zubehör-Artikelsuchdialog wird jetzt über dem Editor angezeigt, statt hinter dessen
  Modalebene verborgen zu bleiben.
- Spurweiten-Mehrfachauswahlen reagieren auf eingegebene Optionsbezeichnungen und begrenzen Escape
  auf die geöffnete Auswahlliste, statt den gesamten Artikeleditor zu schließen.
- Fehlerhafte oder unpassende Gleis-Fachangaben externer Seiten lassen sich nicht auswählen oder
  importieren.

## [0.1.17] - 2026-08-13

### Hinzugefügt

- Artikeldaten- und Barcodesuche sind jetzt direkt beim Anlegen und Bearbeiten von Zubehör
  verfügbar, mit Quellenvorschau und ausdrücklicher felweiser Auswahl vor der Übernahme.
- Ausgewählte Produktbilder aus Suchtreffern lassen sich nach dem Speichern als private
  Zubehördokumente importieren.
- Die GitHub-Sponsoring-Links enthalten jetzt Buy Me a Coffee und PayPal.

### Geändert

- Der bisherige Bereich `Artikelübersicht` heißt jetzt `Zubehör`; die bestehende Route
  `/accessories` bleibt unverändert.
- Gemeinsame Artikelsuch- und Barcodedialoge verwenden in Fahrzeug- und Zubehörabläufen jetzt
  anwendungseigene Modal-, Tastatur-, Fokusfallen- und Fokuswiederherstellungslogik.
- `modernc.org/sqlite` wurde auf 1.56.0, `lucide-react` auf 1.29.0 und Vite auf 8.2.1 aktualisiert.

### Behoben

- Unbekannte oder inaktive Hersteller- und Spurweitenvorschläge lassen sich nicht mehr auswählen,
  wenn der entsprechende Stammdatenwert nicht verfügbar ist.
- Wiederholte Bildimporte sind idempotent, erhalten vorhandene Hauptbilder und liefern bereits
  gespeicherte Dokumente ohne erneuten Download zurück.
- Escape und Tab bleiben im aktiven untergeordneten Suchdialog, statt den dahinterliegenden
  Zubehördialog zu erreichen.

### Sicherheit

- Externe Zubehörbilder sind auf öffentliche HTTP(S)-Ziele beschränkt und werden vor der
  Speicherung anhand von URL, DNS, Weiterleitungen, MIME-Typ und Anhangsgröße geprüft.

## [0.1.16] - 2026-08-09

### Hinzugefügt

- Vollständiger Artikelkatalog und Artikelbestand mit generierten Inventarnummern, Mengen- oder
  Einzelverfolgung, kontrollierten Zusatzfeldern, Käufen und Bestandseinheiten.
- Lagerortbestände, transaktionale Bestandsbewegungen, Reservierungen, Einbauten, Zustandshistorie,
  Ausbauten, Dokumente und zusammengeführte Nutzungshistorie.
- Sortierbare Artikelübersicht mit Auswahlspalte an erster Stelle, Produktbildern und betrieblichen
  Bestandszusammenfassungen.
- Fundamente für Anlagen, Module, Aufbauten, Planvarianten und Planrevisionen mit Schutz vor
  Versionskonflikten.
- Backup-Format Version 3 für die erweiterten Artikeldaten bei erhaltener Importkompatibilität mit
  den Backup-Versionen 1 und 2.

### Geändert

- Artikelformulare verwenden konfigurierte Stammdaten-Dropdowns für Hersteller, Spurweiten,
  Artikelarten, Unterarten, Bestandseinheiten und kontrollierte Zusatzfelder.
- Kennzahlen der Artikelübersicht zeigen Summen deutlich und Einzelwerte zurückhaltender.
- Artikelbearbeitung, Bestandsbuchung, Reservierung, Einbau und gleisspezifische Formulare verwenden
  kompakte, ausgerichtete Mehrspaltenlayouts.
- Artikelstammdaten sind unter `Einstellungen > Daten` in getrennten Gruppen für allgemeine Daten
  und Artikeldaten zusammengeführt.
- Der Eintrag der Hauptnavigation heißt `Anlage` und bleibt sichtbar, aber vorübergehend deaktiviert,
  solange der Arbeitsbereich weiter verfeinert wird. Der Direktzugriff auf `/layouts` und die
  Anlagen-API bleiben verfügbar.

### Behoben

- Transaktionale Bestands- und Zuordnungsregeln bleiben bei Wiederholungen, Umbuchungen,
  Reservierungen, Einbauten, Ausbauten und Wiederherstellungen erhalten.
- Inaktive kontrollierte Zusatzfelder und historische Werte bleiben bei Bearbeitung und Backups
  erhalten.
- Artikeleingaben, Dokumentprüfung, Suchabfragen, ältere Aktualisierungen und die
  Fokuswiederherstellung nach Bestätigungsdialogen wurden gehärtet.
- Die Kompatibilität mit Backup-Version 1 und 2 bleibt erhalten, ohne die Vollständigkeitsprüfung von
  Version 3 abzuschwächen.

### Sicherheit

- API-Validierung und geschützte Stammdatenregeln beim Import wurden verschärft.
- Die Vorprüfung der Backup-Wiederherstellung bleibt konservativ, Authentifizierungsdaten bleiben
  aus Exporten ausgeschlossen.

[0.1.20.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.20.1...v0.1.20.2
[0.1.20.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.20...v0.1.20.1
[0.1.20]: https://github.com/ichwars/RailKeeper/compare/v0.1.19.2...v0.1.20
[0.1.19.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.19.1...v0.1.19.2
[0.1.19.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.19...v0.1.19.1
[0.1.19]: https://github.com/ichwars/RailKeeper/compare/v0.1.18...v0.1.19
[0.1.18]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.6...v0.1.18
[0.1.17.6]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.5...v0.1.17.6
[0.1.17.5]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.4...v0.1.17.5
[0.1.17.4]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.3...v0.1.17.4
[0.1.17.3]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.2...v0.1.17.3
[0.1.17.2]: https://github.com/ichwars/RailKeeper/compare/v0.1.17.1...v0.1.17.2
[0.1.17.1]: https://github.com/ichwars/RailKeeper/compare/v0.1.17...v0.1.17.1
[0.1.17]: https://github.com/ichwars/RailKeeper/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/ichwars/RailKeeper/compare/v0.1.15...v0.1.16
