@echo off
chcp 65001 >nul
echo ========================================
echo   千年江湖 - 客户端发布编译脚本
echo ========================================
echo.

REM 设置变量
set FRONTEND_DIR=%~dp0Frontend\Electron
set BUILD_DIR=%~dp0Frontend\Build

REM 进入Electron目录
cd /d "%FRONTEND_DIR%"

echo [1/2] 检查依赖...
if not exist "node_modules" (
    echo 正在安装依赖...
    call npm.cmd install
    if errorlevel 1 (
        echo [错误] npm install 失败!
        pause
        exit /b 1
    )
) else (
    echo 依赖已存在，跳过安装
)

echo.
echo [2/2] 编译发布版...
echo 正在打包 Electron 应用...

REM 使用electron-packager打包
call npm.cmd run package
if errorlevel 1 (
    echo [错误] 打包失败!
    pause
    exit /b 1
)

echo.
echo ========================================
echo   编译完成！输出目录: %BUILD_DIR%
echo ========================================
echo.

REM 显示生成的文件
if exist "%BUILD_DIR%" (
    echo 生成的文件:
    dir /b "%BUILD_DIR%"
) else (
    echo [警告] 未找到输出目录
)

echo.
pause