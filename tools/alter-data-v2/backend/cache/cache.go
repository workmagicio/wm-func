package cache

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"wm-func/tools/alter-data-v2/backend/redis"
)

// CacheManager manages the persistent cache for remove data
type CacheManager struct {
	mu        sync.RWMutex
	keyPrefix string
}

var (
	instance *CacheManager
	once     sync.Once
)

// GetCacheManager returns the singleton cache manager
func GetCacheManager() *CacheManager {
	once.Do(func() {
		instance = &CacheManager{
			keyPrefix: "cache:remove_data:",
		}
		// 初始化Redis连接
		redis.GetClient()
	})
	return instance
}

// getCacheKey generates a cache key for tenant and platform
func (c *CacheManager) getCacheKey(tenantId int64, platform string) string {
	return fmt.Sprintf("%s%d_%s", c.keyPrefix, tenantId, platform)
}

// GetRemoveData retrieves remove data for a tenant and platform
func (c *CacheManager) GetRemoveData(tenantId int64, platform string) map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	redisKey := c.getCacheKey(tenantId, platform)
	client := redis.GetClient()
	ctx := redis.GetContext()

	data, err := client.HGetAll(ctx, redisKey).Result()
	if err != nil {
		fmt.Printf("Error getting remove data from Redis: %v\n", err)
		return make(map[string]int64)
	}

	result := make(map[string]int64)
	for date, valueStr := range data {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			result[date] = value
		}
	}
	return result
}

// SetRemoveData stores remove data for a tenant and platform
func (c *CacheManager) SetRemoveData(tenantId int64, platform string, removeDataMap map[string]int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	redisKey := c.getCacheKey(tenantId, platform)
	client := redis.GetClient()
	ctx := redis.GetContext()

	// 准备Redis Hash数据
	hashData := make(map[string]interface{})
	for date, removeData := range removeDataMap {
		hashData[date] = strconv.FormatInt(removeData, 10)
	}

	// 添加元数据
	hashData["_tenant_id"] = strconv.FormatInt(tenantId, 10)
	hashData["_platform"] = platform
	hashData["_last_updated"] = time.Now().Format(time.RFC3339)

	// 保存到Redis
	if err := client.HSet(ctx, redisKey, hashData).Err(); err != nil {
		return fmt.Errorf("save to redis failed: %w", err)
	}

	return nil
}

// HasRemoveData checks if remove data exists for a tenant and platform
func (c *CacheManager) HasRemoveData(tenantId int64, platform string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	redisKey := c.getCacheKey(tenantId, platform)
	client := redis.GetClient()
	ctx := redis.GetContext()

	exists, err := client.Exists(ctx, redisKey).Result()
	if err != nil {
		fmt.Printf("Error checking remove data existence: %v\n", err)
		return false
	}
	return exists > 0
}

// GetCacheInfo returns cache information for debugging
func (c *CacheManager) GetCacheInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client := redis.GetClient()
	ctx := redis.GetContext()

	info := make(map[string]interface{})
	info["cache_type"] = "redis"
	info["key_prefix"] = c.keyPrefix

	// 统计缓存键数量
	pattern := c.keyPrefix + "*"
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		info["error"] = fmt.Sprintf("Error getting cache keys: %v", err)
		info["cache_count"] = 0
		info["cached_tenants"] = []string{}
	} else {
		info["cache_count"] = len(keys)
		info["cached_tenants"] = keys
	}

	return info
}
