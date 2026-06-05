RailKeeper Portable fuer Windows
================================

Dieses Paket kann ohne Installation gestartet werden. Es werden keine
zusaetzlichen Programme, Datenbanken oder Laufzeitumgebungen benoetigt.

Empfohlene Nutzung
------------------

Entpacken Sie das ZIP-Archiv vollstaendig auf Ihren Rechner, zum Beispiel:

C:\Users\<IhrName>\Documents\RailKeeper

Starten Sie anschliessend:

start-railkeeper.bat

RailKeeper oeffnet danach automatisch den Browser. Falls der Browser nicht
automatisch startet, oeffnen Sie diese Adresse manuell:

http://127.0.0.1:8080

USB-Stick
---------

RailKeeper kann grundsaetzlich auch direkt von einem USB-Stick gestartet
werden. Fuer den taeglichen Betrieb wird dies jedoch nicht empfohlen, da
USB-Sticks je nach Qualitaet langsamer sein koennen und bei versehentlichem
Abziehen Daten beschaedigt werden koennen.

Wichtig
-------

- Lassen Sie das RailKeeper-Fenster waehrend der Nutzung geoeffnet.
- Ihre Daten werden im Ordner "data" neben RailKeeper.exe gespeichert.
- Sichern Sie den gesamten RailKeeper-Ordner regelmaessig.
- Wenn Port 8080 belegt ist, waehlt RailKeeper automatisch einen der naechsten
  freien lokalen Ports und zeigt die Adresse im Fenster an.
- RailKeeper bindet im Portable-Modus nur lokal an 127.0.0.1.

Update
------

1. RailKeeper beenden.
2. Den alten Ordner sichern.
3. Das neue ZIP in einen neuen Ordner entpacken.
4. Den Ordner "data" aus der alten Version in den neuen Ordner kopieren.
5. start-railkeeper.bat starten.
