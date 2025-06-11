#!/usr/bin/env bash

set -euo pipefail

echo "🔧 开始安装 protoc / gRPC / OpenAPI 相关工具..."

# 安装 protoc（Protocol Buffers 编译器）v21.12 版本，适用于大多数场景
PROTOC_VERSION=21.12
INSTALL_DIR="/usr/local"

if ! command -v protoc &> /dev/null; then
  echo "📦 安装 protoc..."
  curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
  unzip -o protoc-${PROTOC_VERSION}-linux-x86_64.zip -d protoc-tmp
  sudo cp -r protoc-tmp/bin/* ${INSTALL_DIR}/bin/
  sudo cp -r protoc-tmp/include/* ${INSTALL_DIR}/include/
  rm -rf protoc-${PROTOC_VERSION}-linux-x86_64.zip protoc-tmp
else
  echo "✅ protoc 已安装：$(protoc --version)"
fi

# 安装 Go 插件
echo "📦 安装 protoc-gen-go..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

echo "📦 安装 protoc-gen-go-grpc..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo "📦 安装 oapi-codegen..."
# go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# 添加 Go bin 到 PATH
GOBIN=$(go env GOPATH)/bin
if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
  echo "📌 将 Go bin 添加到 PATH 中（$GOBIN）"
  echo "export PATH=\"\$PATH:$GOBIN\"" >> ~/.bashrc
  export PATH="$PATH:$GOBIN"
  export PATH=$PATH:$(go env GOPATH)/bin
fi

# 检查是否安装成功
echo "🧪 检查工具版本..."
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
oapi-codegen --version

echo "🎉 所有工具安装完成！你现在可以运行 make gen 来生成代码了。"