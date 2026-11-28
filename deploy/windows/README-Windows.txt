RailKeeper Windows Standalone (ohne Installation)
=================================================

RailKeeper selbst wird nicht installiert. Das Programmfenster verwendet
die Microsoft Edge WebView2 Evergreen Runtime, die auf vielen Windows-PCs
bereits vorhanden ist. Die Runtime wird nicht mit dem ZIP ausgeliefert.

RailKeeper itself does not require installation. The desktop window uses
the Microsoft Edge WebView2 Evergreen Runtime, which is already present
on many Windows PCs. The ZIP does not include the runtime.

Lizenz und Quellcode / License and source code
--------------------------------------------

LICENSE.md enthaelt die AGPL-3.0-Lizenz von RailKeeper. In
THIRD_PARTY_NOTICES.md und THIRD_PARTY_LICENSES.txt stehen die Hinweise
und Lizenztexte zu Drittanbietern. Die Schriftlizenz liegt unter
web/fonts/LICENSE-Selawik.txt. Der Quellcode steht beim zugehoerigen
Versionstag unter https://github.com/ichwars/RailKeeper/releases bereit.

LICENSE.md contains RailKeeper's AGPL-3.0 license. See THIRD_PARTY_NOTICES.md
and THIRD_PARTY_LICENSES.txt for third-party notices and license texts, and
web/fonts/LICENSE-Selawik.txt for the font license. Source code is available
under the corresponding version tag at https://github.com/ichwars/RailKeeper/releases.

Empfohlene Nutzung
------------------

Entpacken Sie das ZIP-Archiv vollstaendig auf Ihren Rechner, zum Beispiel:

C:\Users\<IhrName>\Documents\RailKeeper

Oeffnen Sie anschliessend "Start RailKeeper.exe" im Explorer.
RailKeeper oeffnet ein eigenes Fenster. Die Konsole startet minimiert.
Schliessen Sie das RailKeeper-Fenster ueber X oder Alt+F4. RailKeeper
schliesst die Konsole automatisch, sobald das Programm beendet ist.

Fehlt WebView2, installieren Sie die Evergreen Runtime von Microsoft:
https://developer.microsoft.com/microsoft-edge/webview2/#download-section
Alternativ starten Sie start-railkeeper.bat im Terminal. Dann oeffnet sich
RailKeeper im Standardbrowser. Strg+C im Terminal beendet diese Instanz.

Open "Start RailKeeper.exe" in File Explorer. RailKeeper opens its own
window and starts its console minimized. Close the RailKeeper window
with X or Alt+F4. The console closes automatically when RailKeeper exits.

If WebView2 is missing, install the Evergreen Runtime from Microsoft:
https://developer.microsoft.com/microsoft-edge/webview2/#download-section
Alternatively, run start-railkeeper.bat in a terminal. RailKeeper opens
in your default browser. Press Ctrl+C in that terminal to exit.

Sicherer Datenordner
--------------------

Ihre Daten liegen standardmaessig nicht im austauschbaren Programmordner,
sondern hier:

%LOCALAPPDATA%\RailKeeper\data

Das Profil fuer das Programmfenster liegt separat unter
%LOCALAPPDATA%\RailKeeper\WebView2. Es gehoert nicht zum Daten-Backup.
The desktop window profile is stored separately under
%LOCALAPPDATA%\RailKeeper\WebView2. It is not part of the data backup.

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

- Pruefen Sie den aktiven Datenordner unter Einstellungen > Allgemein.
- Erstellen Sie vor Updates ein RailKeeper-Backup und nach Moeglichkeit eine
  zusaetzliche Kopie von %LOCALAPPDATA%\RailKeeper\data.
- Wenn Port 8080 belegt ist, waehlt RailKeeper automatisch einen der naechsten
  freien lokalen Ports. Die Konsole zeigt die verwendete Adresse an.
- Windows Standalone bindet standardmaessig nur lokal an 127.0.0.1.

Update
------

1. Unter Einstellungen > Allgemein > Updates nach einer neuen Version suchen.
2. Die Schaltflaeche "Version X herunterladen" laedt nur das passende ZIP
   von GitHub herunter. RailKeeper installiert, entpackt oder ersetzt nichts.
3. Ein RailKeeper-Backup erstellen und RailKeeper beenden.
4. Das heruntergeladene ZIP in einen neuen Programmordner entpacken.
5. "Start RailKeeper.exe" im neuen Ordner oeffnen.
6. Unter Einstellungen den angezeigten Datenordner und den Bestand pruefen.

Kopieren Sie keine Datenbank in den neuen Programmordner. RailKeeper verwendet
den persistenten Datenordner automatisch weiter. Fehlt eine vertrauenswuerdige,
passende ZIP-Datei, verwenden Sie stattdessen die verlinkte GitHub-Release-Seite.
