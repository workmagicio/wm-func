#!/bin/bash

# 测试数据监控看板服务

echo "🧪 测试数据监控看板服务..."

# 检查服务是否在运行
if ! lsof -ti:8090 >/dev/null 2>&1; then
    echo "❌ 服务未运行在端口8090"
    echo "请先运行: ./start.sh"
    exit 1
fi

echo "✅ 服务正在运行"

# 测试主页
echo ""
echo "🔍 测试主页 (http://localhost:8090/)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 主页访问成功 (HTTP $RESPONSE)"
else
    echo "❌ 主页访问失败 (HTTP $RESPONSE)"
fi

# 测试API - 平台列表
echo ""
echo "🔍 测试API - 平台列表 (http://localhost:8090/api/platforms)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/platforms)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 平台列表API访问成功 (HTTP $RESPONSE)"
    echo "📋 平台列表内容:"
    curl -s http://localhost:8090/api/platforms | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8090/api/platforms
else
    echo "❌ 平台列表API访问失败 (HTTP $RESPONSE)"
fi

# 测试API - Google平台数据
echo ""
echo "🔍 测试API - Google平台数据 (http://localhost:8090/api/data/google)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/data/google)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ Google平台数据API访问成功 (HTTP $RESPONSE)"
else
    echo "⚠️  Google平台数据API响应 (HTTP $RESPONSE)"
    echo "📝 响应内容:"
    curl -s http://localhost:8090/api/data/google | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8090/api/data/google
fi

# 测试静态文件
echo ""
echo "🔍 测试静态文件 (http://localhost:8090/static/css/style.css)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/static/css/style.css)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 静态文件访问成功 (HTTP $RESPONSE)"
else
    echo "❌ 静态文件访问失败 (HTTP $RESPONSE)"
fi

# 测试缓存功能
echo ""
echo "🔍 测试缓存统计 (http://localhost:8090/api/cache/stats)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/cache/stats)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 缓存统计API访问成功 (HTTP $RESPONSE)"
    echo "📊 缓存统计信息:"
    curl -s http://localhost:8090/api/cache/stats | python3 -m json.tool 2>/dev/null || curl -s http://localhost:8090/api/cache/stats
else
    echo "❌ 缓存统计API访问失败 (HTTP $RESPONSE)"
fi

# 测试强制刷新
echo ""
echo "🔍 测试强制刷新 (http://localhost:8090/api/data/google?refresh=true)..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8090/api/data/google?refresh=true")
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 强制刷新API访问成功 (HTTP $RESPONSE)"
else
    echo "⚠️  强制刷新API响应 (HTTP $RESPONSE)"
fi

echo ""
echo "🌐 在浏览器中访问: http://localhost:8090/"
echo "📚 API文档:"
echo "   - GET  /api/platforms               # 获取平台列表"
echo "   - GET  /api/data/{platform}         # 获取平台数据（使用缓存）"
echo "   - GET  /api/data/{platform}?refresh=true # 强制刷新平台数据"
echo "   - GET  /api/data/{platform}/{tenant_id}  # 获取租户数据"
echo "   - POST /api/refresh/{platform}      # 刷新平台缓存"
echo "   - GET  /api/cache/stats             # 获取缓存统计"
echo ""
echo "🎯 新增功能:"
echo "   - ✅ 数据缓存（30分钟TTL）"
echo "   - ✅ 最后更新时间显示"
echo "   - ✅ 刷新按钮（页面右上角）"
echo "   - ✅ 缓存状态标识（最新/缓存/已过期）"
echo "   - ✅ 键盘快捷键 Ctrl+R 强制刷新"
