RailKeeper Windows Standalone (ohne Installation)
=================================================

Dieses Paket kann ohne Installation gestartet werden. Es werden keine
zusaetzlichen Programme, Datenbanken oder Laufzeitumgebungen benoetigt.

Empfohlene Nutzung
------------------

Entpacken Sie das ZIP-Archiv vollstaendig auf Ihren Rechner, zum Beispiel:

C:\Users\<IhrName>\Documents\RailKeeper

Starten Sie anschliessend:

start-railkeeper.bat

RailKeeper oeffnet danach automatisch den Browser. Falls der Browser nicht
automatisch startet, oeffnen Sie die im Fenster angezeigte lokale Adresse.

Sicherer Datenordner
--------------------

Ihre Daten liegen standardmaessig nicht im austauschbaren Programmordner,
sondern hier:

%LOCALAPPDATA%\RailKeeper\data

Das ZIP enthaelt absichtlich keinen Datenordner. Ein neuer Programmordner
verwendet deshalb beim naechsten Start wieder dieselben vorhandenen Daten.

Beim ersten Start nach einer aelteren Standalone- oder Portable-Version wird
ein vorhandener Ordner "data" neben RailKeeper.exe in den sicheren Ordner
uebernommen. Der bisherige Ordner bleibt dabei unveraendert erhalten. Wenn an
beiden Stellen unterschiedliche Datenbanken vorhanden sind, stoppt RailKeeper
und zeigt eine Sicherheitsseite. Keine Datenbank wird automatisch ausgewaehlt,
ueberschrieben oder zusammengefuehrt.

Expliziter Datenordner
----------------------

Fortgeschrittene Benutzer koennen RAILKEEPER_DATA_DIR vor dem Start auf einen
anderen absoluten Ordner setzen. Dieser Modus deaktiviert die automatische
Uebernahme. Liegt der Ordner innerhalb des Programmverzeichnisses oder auf
einem USB-Stick, gehen Daten beim Loeschen, Ueberschreiben oder Abziehen dieses
Speichers verloren. Regelmaessige externe Backups bleiben erforderlich.

Wichtig
-------

- Lassen Sie das RailKeeper-Fenster waehrend der Nutzung geoeffnet.
- Pruefen Sie den aktiven Datenordner unter Einstellungen > Allgemein.
- Erstellen Sie vor Updates ein RailKeeper-Backup und nach Moeglichkeit eine
  zusaetzliche Kopie von %LOCALAPPDATA%\RailKeeper\data.
- Wenn Port 8080 belegt ist, waehlt RailKeeper automatisch einen der naechsten
  freien lokalen Ports und zeigt die Adresse im Fenster an.
- Windows Standalone bindet standardmaessig nur lokal an 127.0.0.1.

Update
------

1. Unter Einstellungen > Allgemein > Updates nach einer neuen Version suchen.
2. Die Schaltflaeche "Version X herunterladen" laedt nur das passende ZIP
   von GitHub herunter. RailKeeper installiert, entpackt oder ersetzt nichts.
3. Ein RailKeeper-Backup erstellen und RailKeeper beenden.
4. Das heruntergeladene ZIP in einen neuen Programmordner entpacken.
5. start-railkeeper.bat im neuen Ordner starten.
6. Unter Einstellungen den angezeigten Datenordner und den Bestand pruefen.

Kopieren Sie keine Datenbank in den neuen Programmordner. RailKeeper verwendet
den persistenten Datenordner automatisch weiter. Fehlt eine vertrauenswuerdige,
passende ZIP-Datei, verwenden Sie stattdessen die verlinkte GitHub-Release-Seite.
