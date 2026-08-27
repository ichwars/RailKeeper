# Import/Export: Set-Roundtrip und Prüfkorrekturen

Datum: 2026-08-27
Status: freigegeben

## Ziel

Der Import/Export-Arbeitsbereich soll konsistent bedienbar sein und Fahrzeugsets verlustfrei
zwischen RailKeeper-Installationen übertragen. Die Umsetzung behebt zugleich die in Issue #159
gemeldete Rückkehr von der Prüfung zur CSV-Zuordnung, den Versatz der Spalte `Datensatz`, die
uneinheitliche Anordnung der Profilaktionen und die fehlende direkte Bearbeitung per Zeilenklick.

Issue #140 wird um einen vollständigen Set-Roundtrip für JSON und CSV abgeschlossen. Ein Set umfasst
dabei seine Stammdaten, seine Mitglieder, die Mitgliedsreihenfolge und die optionalen
Mitgliedsbezeichnungen.

## Nicht im Umfang

- Bilder, Anhänge, Wartungen, CV-Dateien und andere Binär- oder Detaildaten bleiben außerhalb des
  Transferprofils. Dafür bleibt die Anwendungssicherung zuständig.
- Profile erhalten keinen neuen auswählbaren Bereich `Fahrzeugsets`. Sets gehören strukturell zum
  vorhandenen Bereich `Fahrzeuge`.
- Alte Transferpakete werden nicht nachträglich um Set-Informationen ergänzt.

## Bedienoberfläche

### Profilbearbeitung

Jede bearbeitbare Profilzeile öffnet den Profildialog per Mausklick, `Enter` oder Leertaste. Das
Startsymbol und das Drei-Punkte-Menü bleiben eigenständige Schaltflächen. Ihre Maus- und
Tastaturereignisse dürfen den Zeilenklick nicht zusätzlich auslösen.

Im Profildialog steht die destruktive Aktion links. `Abbrechen` und die primäre Speicheraktion stehen
rechts in derselben Zeile und dürfen auf Desktopbreiten nicht umbrechen. Auf kleinen Mobilbreiten
werden die Aktionen kontrolliert auf volle Breite gestapelt. Fokusreihenfolge, Beschriftung und
sichtbarer Tastaturfokus bleiben erhalten.

### Importprüfung

Die zweite Spalte der Prüftabelle bleibt eine reguläre Tabellenzelle. Datensatzschlüssel und
Zeilennummer liegen in einem inneren Layoutcontainer. Dadurch verwenden Kopf und Inhalt weiterhin
dasselbe Tabellenraster.

Nach einer Konfliktentscheidung bleibt die Ansicht auf `Prüfung` oder wechselt nach der letzten
Entscheidung zu `Bestätigung`. Eine Aktualisierung desselben Auftrags darf die bereits bestätigte
CSV-Zuordnung nicht zurücksetzen. Nur eine andere Quelldatei, eine geänderte Zuordnung, ein anderer
Auftrag oder ein echter Reupload-Konflikt setzt diese Bestätigung zurück.

Die Auswahltexte verwenden fachliche Aktionen statt technischer Auflösungsbegriffe:

- `Aktion wählen`
- `Bestehenden Datensatz überschreiben`
- `Als neuen Datensatz mit neuer Inventarnummer importieren`
- `Diesen Datensatz nicht importieren`
- bei passenden Hersteller-/Artikelnummern: `Vorhandenes Fahrzeug verwenden`
- bei passenden Hersteller-/Artikelnummern: `Zusätzlich als neues Fahrzeug importieren`

Die englischen Texte werden sinngleich gepflegt.

### Set-Prüfung

Die Vorschau erhält einen Abschnitt `Erkannte Fahrzeugsets`. Jede Gruppe zeigt mindestens
Set-Inventarnummer, Name, Mitgliederzahl, Status und eine gegebenenfalls erforderliche Gesamtaktion.
Die vorhandene Kennzahl `Datensätze` zählt weiterhin Fahrzeuge und wird durch Set-Metadaten nicht
künstlich erhöht.

## Transferformat

### JSON

Die Paketversion steigt von 2 auf 3. Innerhalb von `areas` wird neben `vehicles` die optionale
Sammlung `vehicleSets` ausgegeben, sobald der Bereich `Fahrzeuge` ausgewählt wurde und Sets vorhanden
sind.

Ein Set-Datensatz enthält:

- Quell-ID und Set-Inventarnummer,
- alle Felder aus `VehicleSetInput`,
- eine geordnete Mitgliederliste mit Quell-Fahrzeug-ID, Fahrzeug-Inventarnummer, Position und Label.

Mitgliedsreferenzen verwenden die Quell-ID zur eindeutigen Paketverknüpfung und die
Inventarnummer zur lesbaren Kontrolle. Zielsystem-IDs werden nie aus dem Paket übernommen.

Pakete der Versionen 1 und 2 bleiben lesbar. Da sie keine Set-Struktur besitzen, werden ihre
Fahrzeuge wie bisher einzeln importiert.

### CSV

CSV bleibt eine Datei mit einer Zeile pro Fahrzeug. Set-Mitglieder erhalten zusätzliche,
eindeutig präfixierte Spalten:

- `vehicleSetInventoryNumber`, `vehicleSetName`, `vehicleSetManufacturer`,
  `vehicleSetArticleNumber`, `vehicleSetArticleSourceUrl`, `vehicleSetGauge`, `vehicleSetEpoch`,
  `vehicleSetRailwayCompany`, `vehicleSetCategory`, `vehicleSetGattung`, `vehicleSetDescription`,
  `vehicleSetEAN`, `vehicleSetProductionPeriod`, `vehicleSetListPrice`,
  `vehicleSetAcquisitionType`, `vehicleSetAcquiredFrom`, `vehicleSetPurchasePrice`,
  `vehicleSetPurchaseDate`, `vehicleSetStorageLocation`, `vehicleSetStorageDetails`,
  `vehicleSetCondition`, `vehicleSetConditionDetails`, `vehicleSetPackaging`,
- `vehicleSetPosition`, `vehicleSetMemberCount` und `vehicleSetMemberLabel`.

Set-freie Fahrzeuge lassen diese Spalten leer. Beim Import gruppiert die Set-Inventarnummer die
Mitglieder. Die Reihenfolge der CSV-Zeilen ist dafür unerheblich. Die Spalten nehmen an der
bestehenden Alias-, Profil- und manuellen Zuordnung teil und erhalten deutsche sowie englische
Aliase.

## Validierung

Ein erkannter Set-Verbund ist nur gültig, wenn:

- die Set-Inventarnummer gesetzt ist,
- Name, Hersteller, Spurweite, Kategorie und Gattung gesetzt sind,
- mindestens zwei unterschiedliche Mitglieder vorliegen,
- Positionen positiv, eindeutig und lückenlos von 1 bis zur Mitgliederzahl laufen,
- alle Zeilen dieselbe Mitgliederzahl und dieselben Set-Stammdaten enthalten,
- jedes Mitglied genau einem Set angehört,
- jede JSON-Mitgliedsreferenz auf ein enthaltenes Fahrzeug zeigt.

Teilweise vorhandene Set-Spalten werden nicht stillschweigend ignoriert. Sie erzeugen eine
Prüfmeldung. Ein fehlerhaftes oder übersprungenes Mitglied verhindert, dass ein unvollständiges Set
geschrieben wird.

## Konflikte und Schreibregeln

Set-Konflikte werden auf Gruppenebene entschieden:

- `Bestehendes Set aktualisieren`: Set-Stammdaten werden aktualisiert. Mitglieder werden nur mit
  Fahrzeugen innerhalb dieses Sets abgeglichen. Fehlende Mitglieder werden angelegt. Nicht mehr im
  Import enthaltene bisherige Mitglieder bleiben als Fahrzeuge erhalten, werden aber aus dem Set
  gelöst.
- `Als neues Set importieren`: RailKeeper vergibt eine neue Set-Inventarnummer und bei notwendigen
  Kopien neue Fahrzeug-Inventarnummern. Es werden keine bestehenden Fahrzeuge umgehängt.
- `Dieses Set nicht importieren`: Das Set und alle zugehörigen Fahrzeugdatensätze werden
  übersprungen.

Kollidiert ein importiertes Mitglied mit einem Fahrzeug außerhalb des zu aktualisierenden Sets,
darf dieses Fahrzeug nicht automatisch ersetzt oder umgehängt werden. Die Vorschau verlangt dann
eine sichere Set-Gesamtaktion oder das Überspringen.

Fahrzeuge werden zuerst vorbereitet, die Set-Mitgliedschaften anschließend innerhalb derselben
SQLite-Transaktion geschrieben. Die Anwendung führt unmittelbar vor dem Schreiben erneut
Fingerabdruck-, Konflikt- und Vollständigkeitsprüfungen aus. Bei einer Abweichung wird die gesamte
Transaktion verworfen.

## Technische Einordnung

- `application` definiert Paketversion 3, Set-Transfermodelle, Parser, Gruppierung, Vorschau und
  Validierungsregeln.
- `infrastructure` liest Set-Stammdaten und Mitgliedschaften im konsistenten Snapshot und schreibt
  Fahrzeuge sowie Sets atomar.
- `api` behält die bestehenden Importendpunkte. Neue Vorschaufelder und Konfliktcodes werden im
  OpenAPI-Vertrag dokumentiert.
- `frontend` rendert Set-Gruppen, bewahrt den bestätigten Mappingstatus und behebt Tabellen- sowie
  Profilinteraktionen.

Die vorhandenen Profile und Bereichsauswahlen bleiben kompatibel. Ein Fahrzeugprofil nimmt Sets
automatisch mit, sofern sie im Exportbestand oder in der Importdatei vorkommen.

## Fehlerbehandlung

- Ungültige Set-Strukturen werden in der Vorschau nachvollziehbar dem Set und den betroffenen
  Mitgliedern zugeordnet.
- Unbekannte oder widersprüchliche Set-Spalten bleiben vor der Bestätigung blockierend.
- Revisionskonflikte laden die persistente Vorschau neu. Nur wenn Quelle oder Zuordnung tatsächlich
  nicht mehr dieselbe ist, ist ein erneuter Upload erforderlich.
- Serverfehler lassen Dialog und Auswahlzustand sichtbar, damit der Nutzer korrigieren oder erneut
  versuchen kann.

## Tests und Abnahme

Backendtests decken mindestens ab:

- JSON-Roundtrip eines Sets mit Set-Stammdaten, Mitgliedern, Labels und Reihenfolge,
- CSV-Roundtrip mit gruppierten Set-Spalten,
- unveränderte Einzelfahrzeugimporte,
- Kompatibilität der Paketversionen 1 und 2,
- unvollständige Sets, doppelte oder lückenhafte Positionen und widersprüchliche Stammdaten,
- Aktualisieren, Kopieren und Überspringen eines Set-Konflikts,
- Kollision eines Mitglieds außerhalb des Zielsets,
- vollständigen Transaktionsabbruch bei einem Set- oder Mitgliedsfehler.

Frontendtests decken mindestens ab:

- Profilbearbeitung per Klick, `Enter` und Leertaste,
- isolierte Start- und Menüaktionen innerhalb einer Profilzeile,
- stabile Desktop- und mobile Aktionsanordnung,
- unveränderte Tabellengeometrie der Spalte `Datensatz`,
- Verbleib auf `Prüfung` beziehungsweise Wechsel zu `Bestätigung` nach einer Auswahl,
- verständliche deutsche und englische Aktionstexte,
- Anzeige und Gesamtentscheidung erkannter Set-Gruppen.

Abschließend laufen `go test ./...`, die vollständige Frontend-Testfolge, der Frontend-Build und
die Dokumentationsprüfung. Die betroffenen Ansichten werden auf Desktop und Mobil, in hellem und
dunklem Theme, mit Tastaturfokus und langen deutschen Texten visuell geprüft.
