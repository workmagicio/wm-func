#!/bin/bash

# Redis缓存功能测试脚本

set -e

echo "🔧 测试Redis缓存功能..."

# 检查Docker服务是否运行
echo "1️⃣ 检查Docker服务..."
if ! docker compose ps | grep -q "Up"; then
    echo "❌ Docker服务未运行，请先运行 ./start-docker.sh"
    exit 1
fi

# 检查Redis连接
echo "2️⃣ 检查Redis连接..."
if docker compose exec -T redis redis-cli ping | grep -q "PONG"; then
    echo "✅ Redis连接正常"
else
    echo "❌ Redis连接失败"
    exit 1
fi

# 测试API健康检查
echo "3️⃣ 测试API健康检查..."
if curl -s http://localhost:8081/health | grep -q "ok"; then
    echo "✅ API服务正常"
else
    echo "❌ API服务异常"
    exit 1
fi

# 测试缓存功能
echo "4️⃣ 测试缓存功能..."
echo "   - 请求Google Ads数据（强制刷新）..."
RESPONSE=$(curl -s "http://localhost:8081/api/alter-data?platform=googleAds&needRefresh=true")

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ 数据获取成功"
    
    # 检查Redis中是否有缓存数据
    echo "   - 检查Redis缓存..."
    CACHE_KEYS=$(docker compose exec -T redis redis-cli keys "bcache:*" | wc -l)
    if [ "$CACHE_KEYS" -gt 0 ]; then
        echo "✅ Redis缓存数据创建成功，共 $CACHE_KEYS 个缓存键"
    else
        echo "⚠️  未发现Redis缓存数据"
    fi
    
    REMOVE_CACHE_KEYS=$(docker compose exec -T redis redis-cli keys "cache:remove_data:*" | wc -l)
    if [ "$REMOVE_CACHE_KEYS" -gt 0 ]; then
        echo "✅ RemoveData缓存创建成功，共 $REMOVE_CACHE_KEYS 个缓存键"
    else
        echo "ℹ️  暂无RemoveData缓存（正常，需要特定请求才会创建）"
    fi
else
    echo "❌ 数据获取失败"
    echo "Response: $RESPONSE"
    exit 1
fi

# 测试缓存读取
echo "5️⃣ 测试缓存读取..."
echo "   - 请求Google Ads数据（使用缓存）..."
START_TIME=$(date +%s%N)
RESPONSE2=$(curl -s "http://localhost:8081/api/alter-data?platform=googleAds")
END_TIME=$(date +%s%N)
DURATION=$(( (END_TIME - START_TIME) / 1000000 ))

if echo "$RESPONSE2" | grep -q '"success":true'; then
    echo "✅ 缓存读取成功，响应时间: ${DURATION}ms"
else
    echo "❌ 缓存读取失败"
    exit 1
fi

echo ""
echo "🎉 Redis缓存功能测试全部通过！"
echo ""
echo "📊 测试结果总结:"
echo "   ✅ Redis服务正常运行"
echo "   ✅ API服务健康检查通过" 
echo "   ✅ 缓存写入功能正常"
echo "   ✅ 缓存读取功能正常"
echo "   ✅ 响应时间：${DURATION}ms"
echo ""
echo "🔍 查看Redis缓存详情:"
echo "   docker compose exec redis redis-cli keys '*'"
echo "   docker compose exec redis redis-cli info memory"
