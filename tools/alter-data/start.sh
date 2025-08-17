#!/bin/bash

# 数据监控看板启动脚本
# 确保在正确目录下运行并启动服务
# 停止数据监控看板服务

echo "🛑 停止数据监控看板服务..."

# 查找运行在8090端口的进程
PID=$(lsof -ti:8090 2>/dev/null | head -n1 || true)

if [ -z "$PID" ]; then
    echo "ℹ️  没有发现运行在端口8090的进程"
    exit 0
fi

echo "🔍 发现进程: $PID"
echo "⏳ 正在停止..."

# 优雅停止
kill -TERM "$PID" 2>/dev/null || true
sleep 3

# 检查是否还在运行
if kill -0 "$PID" 2>/dev/null; then
    echo "⚠️  优雅停止失败，强制停止..."
    kill -9 "$PID" 2>/dev/null || true
    sleep 1
fi

# 再次检查
if lsof -ti:8090 >/dev/null 2>&1; then
    echo "❌ 停止失败，端口8090仍被占用"
    echo "请手动检查: lsof -ti:8090"
    exit 1
else
    echo "✅ 服务已成功停止"
fi

set -e  # 遇到错误立即退出

echo "🚀 启动数据监控看板服务..."

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "📁 工作目录: $SCRIPT_DIR"

# 切换到正确的工作目录
cd "$SCRIPT_DIR"

# 检查必要文件是否存在
if [ ! -f "main.go" ]; then
    echo "❌ 错误: 未找到 main.go 文件"
    exit 1
fi

if [ ! -d "static" ]; then
    echo "❌ 错误: 未找到 static 目录"
    exit 1
fi

if [ ! -f "static/index.html" ]; then
    echo "❌ 错误: 未找到 static/index.html 文件"
    exit 1
fi

echo "✅ 文件检查完成"

# 停止可能运行的旧进程
echo "🔄 检查并停止旧进程..."
OLD_PID=$(lsof -ti:8090 2>/dev/null | grep -v grep | head -n1 || true)
if [ ! -z "$OLD_PID" ]; then
    echo "🛑 发现运行在端口8090的进程: $OLD_PID，正在停止..."
    kill -9 $OLD_PID 2>/dev/null || true
    sleep 2
fi

# 编译Go程序
echo "🔨 编译程序..."
if ! go build -o alter-data-server main.go; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译完成"

# 启动服务
echo "🚀 启动服务..."
echo "📍 当前工作目录: $(pwd)"
echo "📂 静态文件目录: $(pwd)/static"
echo ""

# 启动服务并显示日志
./alter-data-server

# 如果上面的命令因为某种原因退出了
echo ""
echo "⚠️  服务已退出"
