@echo off
chcp 65001 >nul 2>&1
echo ========================================
echo   千年江湖 - 服务端健康检查工具
echo ========================================
echo.

:: 检查各服务状态
echo [1/5] 检查 DBService (端口 8083)...
curl -s http://localhost:8083/api/item/get_base -X POST -H "Content-Type: application/json" -d "{\"item_id\":1}" >nul 2>&1
if %errorlevel% equ 0 (
    echo     ✅ DBService 运行正常
) else (
    echo     ❌ DBService 未运行或无法访问！
    echo        请先启动: cd Server\DBService && go run main.go
)

echo.
echo [2/5] 检查 GameService (端口 8082)...
curl -s http://localhost:8082/api/item/base/list >nul 2>&1
if %errorlevel% equ 0 (
    echo     ✅ GameService 运行正常
) else (
    echo     ❌ GameService 未运行或无法访问！
    echo        请先启动: cd Server\GameService && go run main.go
)

echo.
echo [3/5] 检查 GatewayService (端口 8080)...
curl -s http://localhost:8080/ws >nul 2>&1
if %errorlevel% equ 0 (
    echo     ✅ GatewayService 运行正常
) else (
    echo     ❌ GatewayService 未运行或无法访问！
    echo        请先启动: cd Server\GatewayService && go run main.go
)

echo.
echo [4/5] 检查 Redis (端口 6379)...
redis-cli ping >nul 2>&1
if %errorlevel% equ 0 (
    echo     ✅ Redis 运行正常
) else (
    echo     ⚠️ Redis 未运行（可选，用于缓存和跨网关广播）
    echo        启动命令: redis-server
)

echo.
echo [5/5] 检查 RabbitMQ (端口 5672)...
netstat -an | findstr ":5672" >nul 2>&1
if %errorlevel% equ 0 (
    echo     ✅ RabbitMQ 运行正常
) else (
    echo     ⚠️ RabbitMQ 未运行（可选，用于消息队列）
    echo        启动命令: rabbitmq-server
)

echo.
echo ========================================
echo   测试物品接口连通性
echo ========================================
echo.

echo 测试获取背包数据 (RoleID=1)...
curl -s -X POST http://localhost:8083/api/item/get_bag -H "Content-Type: application/json" -d "{\"role_id\":1}" 2>nul
echo.

echo.
echo 测试获取装备数据 (RoleID=1)...
curl -s -X POST http://localhost:8083/api/item/get_equipped -H "Content-Type: application/json" -d "{\"role_id\":1}" 2>nul
echo.

echo.
echo ========================================
echo   健康检查完成
echo ========================================
echo.
echo 如果所有核心服务都显示 ✅，请刷新游戏页面测试背包功能。
echo 如果有 ❌ 服务未运行，请按顺序启动：
echo   1. Redis (可选)
echo   2. DBService
echo   3. GameService
echo   4. GatewayService
echo.
pause
