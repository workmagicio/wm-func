#!/bin/bash

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
