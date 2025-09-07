package bcache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"wm-func/tools/alter-data-v2/backend/redis"

	redisv8 "github.com/go-redis/redis/v8"
)

// Cache 缓存结构
type Cache struct {
	CreateTime time.Time   `json:"create_time"`
	Data       interface{} `json:"data"`
}

// CacheManager 缓存管理器
type CacheManager struct {
	keyPrefix string
	mutex     sync.RWMutex
}

var (
	defaultManager *CacheManager
	once           sync.Once
)

// GetManager 获取缓存管理器单例
func GetManager() *CacheManager {
	once.Do(func() {
		defaultManager = &CacheManager{
			keyPrefix: "bcache:",
		}
		// 初始化Redis连接
		redis.GetClient()
	})
	return defaultManager
}

// SaveCache 保存缓存
func SaveCache(key string, value interface{}) error {
	return GetManager().Save(key, value)
}

// LoadCache 加载缓存
func LoadCache(key string) (*Cache, error) {
	return GetManager().Load(key)
}

// LoadTyped 类型安全的加载方法
func LoadTyped[T any](key string) (T, error) {
	var zero T

	cache, err := GetManager().Load(key)
	if err != nil {
		return zero, err
	}

	// 尝试类型断言
	if data, ok := cache.Data.(T); ok {
		return data, nil
	}

	// 如果直接断言失败，尝试通过JSON转换
	jsonData, err := json.Marshal(cache.Data)
	if err != nil {
		return zero, fmt.Errorf("marshal cache data failed: %w", err)
	}

	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return zero, fmt.Errorf("unmarshal to target type failed: %w", err)
	}

	return result, nil
}

// Save 保存缓存
func (cm *CacheManager) Save(key string, value interface{}) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cache := Cache{
		CreateTime: time.Now(),
		Data:       value,
	}

	data, err := json.Marshal(&cache)
	if err != nil {
		return fmt.Errorf("marshal cache failed: %w", err)
	}

	redisKey := cm.getRedisKey(key)
	client := redis.GetClient()
	ctx := redis.GetContext()

	if err := client.Set(ctx, redisKey, string(data), 0).Err(); err != nil {
		return fmt.Errorf("save to redis failed: %w", err)
	}

	return nil
}

// Load 加载缓存
func (cm *CacheManager) Load(key string) (*Cache, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	redisKey := cm.getRedisKey(key)
	client := redis.GetClient()
	ctx := redis.GetContext()

	data, err := client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redisv8.Nil {
			return nil, fmt.Errorf("cache not found: %s", key)
		}
		return nil, fmt.Errorf("read from redis failed: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, fmt.Errorf("unmarshal cache failed: %w", err)
	}

	return &cache, nil
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	redisKey := cm.getRedisKey(key)
	client := redis.GetClient()
	ctx := redis.GetContext()

	if err := client.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("delete from redis failed: %w", err)
	}
	return nil
}

// getRedisKey 获取Redis键
func (cm *CacheManager) getRedisKey(key string) string {
	return cm.keyPrefix + key
}
