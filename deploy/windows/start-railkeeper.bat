@echo off
setlocal
cd /d "%~dp0"

echo RailKeeper Windows Standalone wird gestartet.
echo Dieses Fenster waehrend der Nutzung geoeffnet lassen.
echo.

RailKeeper.exe --standalone

echo.
echo RailKeeper wurde beendet.
pause
