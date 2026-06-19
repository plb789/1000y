@echo off
echo =========================================
echo   Server Build Script
echo =========================================
echo.

set SRV_DIR=%~dp0
set BIN_DIR=%SRV_DIR%bin

if not exist "%BIN_DIR%" (
    mkdir "%BIN_DIR%"
    echo Created bin directory: %BIN_DIR%
)

cd /d "%SRV_DIR%"

echo [1/5] Building DBService...
go build -o "%BIN_DIR%\DBService.exe" ./DBService
if errorlevel 1 (
    echo Error: DBService build failed!
    pause
    exit /b 1
)
echo Success: DBService.exe

echo [2/5] Building LoginService...
go build -o "%BIN_DIR%\LoginService.exe" ./LoginService
if errorlevel 1 (
    echo Error: LoginService build failed!
    pause
    exit /b 1
)
echo Success: LoginService.exe

echo [3/5] Building GameService...
go build -o "%BIN_DIR%\GameService.exe" ./GameService
if errorlevel 1 (
    echo Error: GameService build failed!
    pause
    exit /b 1
)
echo Success: GameService.exe

echo [4/5] Building GatewayService...
go build -o "%BIN_DIR%\GatewayService.exe" ./GatewayService
if errorlevel 1 (
    echo Error: GatewayService build failed!
    pause
    exit /b 1
)
echo Success: GatewayService.exe

echo [5/5] Building RegistryService...
go build -o "%BIN_DIR%\RegistryService.exe" ./RegistryService
if errorlevel 1 (
    echo Error: RegistryService build failed!
    pause
    exit /b 1
)
echo Success: RegistryService.exe

echo.
echo =========================================
echo   Build completed! Output: %BIN_DIR%
echo =========================================
echo.
echo Generated files:
dir /b "%BIN_DIR%\*.exe"
echo.
pause
