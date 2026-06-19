@echo off
cd /d %~dp0
set BIN_DIR=%~dp0bin

echo Starting DBService...
start cmd /k "cd /d %BIN_DIR% && DBService.exe"
timeout /t 1 >nul

echo Starting LoginService...
start cmd /k "cd /d %BIN_DIR% && LoginService.exe"
timeout /t 1 >nul

echo Starting RegistryService...
start cmd /k "cd /d %BIN_DIR% && RegistryService.exe"
timeout /t 1 >nul

echo Starting GameService...
start cmd /k "cd /d %BIN_DIR% && GameService.exe"
timeout /t 1 >nul

echo Starting GatewayService...
start cmd /k "cd /d %BIN_DIR% && GatewayService.exe"

echo All services started!
pause