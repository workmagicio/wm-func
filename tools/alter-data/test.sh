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
# 测试租户功能
echo ""
echo "🔍 测试租户视图 API..."
echo "📋 租户列表:"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/tenants)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 租户列表API访问成功 (HTTP $RESPONSE)"
    TENANT_COUNT=$(curl -s http://localhost:8090/api/tenants | python3 -c "import json,sys; data=json.load(sys.stdin); print(len(data['data']) if data['success'] else 0)" 2>/dev/null || echo "0")
    echo "📊 发现 $TENANT_COUNT 个租户"
else
    echo "❌ 租户列表API访问失败 (HTTP $RESPONSE)"
fi

echo "👤 租户跨平台数据:"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8090/api/tenant/150075)
if [ "$RESPONSE" = "200" ]; then
    echo "✅ 租户跨平台数据API访问成功 (HTTP $RESPONSE)"
    curl -s http://localhost:8090/api/tenant/150075 | python3 -c "
import json,sys
try:
    data = json.load(sys.stdin)
    if data['success']:
        platform_count = len(data['data']['platform_data'])
        platforms = list(data['data']['platform_data'].keys())
        print(f'📊 租户 {data[\"tenant_id\"]} 包含 {platform_count} 个平台: {platforms}')
        cache_status = '已过期' if data.get('cache_info', {}).get('is_expired', True) else '有效'
        print(f'🗂️  缓存状态: {cache_status}')
    else:
        print('❌ 数据获取失败')
except:
    print('❌ 解析失败')
    " 2>/dev/null || echo "❌ 数据解析失败"
else
    echo "❌ 租户跨平台数据API访问失败 (HTTP $RESPONSE)"
fi

echo ""
echo "🎯 功能特性:"
echo "📊 平台视图:"
echo "   - ✅ 数据缓存（30分钟TTL）"
echo "   - ✅ 最后更新时间显示"
echo "   - ✅ 刷新按钮功能"
echo "   - ✅ 缓存状态标识（最新/缓存/已过期）"
echo "   - ✅ 键盘快捷键 Ctrl+R 强制刷新"
echo "👤 租户视图 (NEW!):"
echo "   - ✅ 租户列表选择器"
echo "   - ✅ 跨平台数据对比展示" 
echo "   - ✅ 独立的缓存管理"
echo "   - ✅ 视图模式一键切换"
echo "   - ✅ 键盘快捷键 Ctrl+T 切换视图"
echo "🔍 租户输入预测 (NEW!):"
echo "   - ✅ 智能输入框替代下拉选择器"
echo "   - ✅ 实时搜索和自动完成"
echo "   - ✅ 支持键盘导航 (↑↓ 方向键)"
echo "   - ✅ 支持任意租户ID输入（包括不存在的）"
echo "   - ✅ 租户列表长期缓存（1天更新）"
echo "⚡ SQL优化:"
echo "   - ✅ 单次查询获取租户所有平台数据"
echo "   - ✅ 避免多次数据库调用"
echo "   - ✅ 支持参数化查询防止SQL注入"
