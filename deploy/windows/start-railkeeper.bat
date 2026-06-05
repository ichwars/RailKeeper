@echo off
setlocal
cd /d "%~dp0"

echo RailKeeper Portable wird gestartet.
echo Dieses Fenster waehrend der Nutzung geoeffnet lassen.
echo.

RailKeeper.exe --portable

echo.
echo RailKeeper wurde beendet.
pause
