#!/bin/bash

SERVICE=$1

if [ -z "$SERVICE" ]; then
  echo "❌ 请指定要运行的服务，例如: ./run.sh order"
  exit 1
fi

SERVICE_DIR="../internal/$SERVICE"

if [ ! -d "$SERVICE_DIR" ]; then
  echo "❌ 服务目录 $SERVICE_DIR 不存在"
  exit 1
fi

cd "$SERVICE_DIR"
echo "🚀 启动服务: $SERVICE"
air .