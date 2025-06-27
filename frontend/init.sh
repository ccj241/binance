#!/bin/bash
# 设置淘宝镜像
yarn config set registry https://registry.npmmirror.com

if [ ! -d "binance-frontend/src" ]; then
  echo "Running create-vue..."
  CI=true yarn create vue || {
    echo "create-vue failed, exiting"
    exit 1
  }
  echo "Changing to binance-frontend directory..."
  cd binance-frontend || {
    echo "Failed to change to binance-frontend directory, exiting"
    exit 1
  }
  echo "Installing dependencies with Yarn..."
  yarn add axios vue-router@4 vite || {
    echo "yarn add failed, exiting"
    exit 1
  }
fi
# 启动Vite开发服务器
echo "Starting Vite..."
cd /app/binance-frontend
yarn dev --host 0.0.0.0 --port 8080  # 容器内部使用 8080