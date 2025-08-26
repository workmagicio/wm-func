package bcache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache 缓存结构
type Cache struct {
	CreateTime time.Time   `json:"create_time"`
	Data       interface{} `json:"data"`
}

// CacheManager 缓存管理器
type CacheManager struct {
	cacheDir string
	mutex    sync.RWMutex
}

var (
	defaultManager *CacheManager
	once           sync.Once
)

// GetManager 获取缓存管理器单例
func GetManager() *CacheManager {
	once.Do(func() {
		defaultManager = &CacheManager{
			cacheDir: "./cache",
		}
		// 确保缓存目录存在
		os.MkdirAll(defaultManager.cacheDir, 0755)
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

	data, err := json.MarshalIndent(&cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache failed: %w", err)
	}

	fileName := cm.getCacheFilePath(key)

	if err := os.WriteFile(fileName, data, 0644); err != nil {
		return fmt.Errorf("write cache file failed: %w", err)
	}

	return nil
}

// Load 加载缓存
func (cm *CacheManager) Load(key string) (*Cache, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	fileName := cm.getCacheFilePath(key)

	// 检查文件是否存在
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache not found: %s", key)
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read cache file failed: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("unmarshal cache failed: %w", err)
	}

	return &cache, nil
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	fileName := cm.getCacheFilePath(key)
	if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete cache file failed: %w", err)
	}
	return nil
}

// getCacheFilePath 获取缓存文件路径
func (cm *CacheManager) getCacheFilePath(key string) string {
	return filepath.Join(cm.cacheDir, key+".json")
}
