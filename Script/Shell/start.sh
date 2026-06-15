#!/bin/bash
echo "启动数据微服务"
cd ../../Server/DBService && ./DBService &
sleep 1

echo "启动登录微服务"
cd ../../Server/LoginService && ./LoginService &
sleep 1

echo "启动游戏微服务"
cd ../../Server/GameService && ./GameService &
sleep 1

echo "启动网关服务"
cd ../../Server/GatewayService && ./GatewayService &
echo "所有服务启动完成"