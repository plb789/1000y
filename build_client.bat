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

echo [1/3] 同步最新源码到Electron目录...
copy /Y "%~dp0Frontend\Game.js" "%FRONTEND_DIR%\Game.js" >nul
copy /Y "%~dp0Frontend\index.html" "%FRONTEND_DIR%\index.html" >nul
echo 源码同步完成

echo.
echo [2/3] 检查依赖...
if not exist "node_modules" (
    echo 正在安装依赖...
    call npm.cmd install
    if errorlevel 1 (
        echo [错误] npm install 失败!
        pause
        exit /b 1
    )
) else (
    echo 依赖已存在，检查是否需要更新terser...
    call npm.cmd install terser --save-dev
)

echo.
echo [3/3] 编译发布版...
echo 正在打包 Electron 应用...

REM 使用淘宝镜像下载 Electron（解决网络问题，避免禁用SSL验证）
set ELECTRON_MIRROR=https://npmmirror.com/mirrors/electron/

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