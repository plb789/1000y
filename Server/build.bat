@echo off
chcp 65001 >nul
echo ========================================
echo   千年江湖 - 服务端编译脚本
echo ========================================
echo.

REM 设置变量
set SRV_DIR=%~dp0
set BIN_DIR=%SRV_DIR%bin

REM 创建bin目录
if not exist "%BIN_DIR%" (
    mkdir "%BIN_DIR%"
    echo 创建bin目录: %BIN_DIR%
)

REM 进入Server目录
cd /d "%SRV_DIR%"

echo [1/4] 编译 DBService...
cd DBService
go build -o "%BIN_DIR%\DBService.exe" .
if errorlevel 1 (
    echo [错误] DBService 编译失败!
    pause
    exit /b 1
)
echo [成功] DBService.exe
cd /d "%SRV_DIR%"

echo [2/4] 编译 LoginService...
cd LoginService
go build -o "%BIN_DIR%\LoginService.exe" .
if errorlevel 1 (
    echo [错误] LoginService 编译失败!
    pause
    exit /b 1
)
echo [成功] LoginService.exe
cd /d "%SRV_DIR%"

echo [3/4] 编译 GameService...
cd GameService
go build -o "%BIN_DIR%\GameService.exe" .
if errorlevel 1 (
    echo [错误] GameService 编译失败!
    pause
    exit /b 1
)
echo [成功] GameService.exe
cd /d "%SRV_DIR%"

echo [4/4] 编译 GatewayService...
cd GatewayService
go build -o "%BIN_DIR%\GatewayService.exe" .
if errorlevel 1 (
    echo [错误] GatewayService 编译失败!
    pause
    exit /b 1
)
echo [成功] GatewayService.exe
cd /d "%SRV_DIR%"

echo.
echo ========================================
echo   编译完成！输出目录: %BIN_DIR%
echo ========================================
echo.
echo 生成的文件:
dir /b "%BIN_DIR%\*.exe"
echo.
pause