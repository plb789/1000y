@echo off
cd /d %~dp0
set BIN_DIR=%~dp0bin
set CONFIG_DIR=%~dp0Config

echo Compiling services...
cd DBService && go build -o ../bin/DBService.exe . && cd ..
cd LoginService && go build -o ../bin/LoginService.exe . && cd ..
cd GameService && go build -o ../bin/GameService.exe . && cd ..
cd GatewayService && go build -o ../bin/GatewayService.exe . && cd ..

echo Starting DBService...
start cmd /k "cd /d %BIN_DIR% && set CONFIG_PATH=%CONFIG_DIR% && DBService.exe"
timeout /t 1 >nul

echo Starting LoginService...
start cmd /k "cd /d %BIN_DIR% && set CONFIG_PATH=%CONFIG_DIR% && LoginService.exe"
timeout /t 1 >nul

echo Starting GameService...
start cmd /k "cd /d %BIN_DIR% && set CONFIG_PATH=%CONFIG_DIR% && GameService.exe"
timeout /t 1 >nul

echo Starting GatewayService...
start cmd /k "cd /d %BIN_DIR% && set CONFIG_PATH=%CONFIG_DIR% && GatewayService.exe"

echo All services started!