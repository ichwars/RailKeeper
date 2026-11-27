@echo off
setlocal
pushd "%~dp0" || exit /b 1
".\RailKeeper.exe" --standalone
set "RailKeeperExitCode=%errorlevel%"
popd
exit /b %RailKeeperExitCode%
