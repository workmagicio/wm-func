package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wm-func/tools/alter-data/models"
)

// CacheItem 缓存项结构
type CacheItem struct {
	Platform  string              `json:"platform"`
	Data      []models.TenantData `json:"data"`
	UpdatedAt time.Time           `json:"updated_at"`
	ExpiresAt time.Time           `json:"expires_at"`
}

// CacheManager 缓存管理器
type CacheManager struct {
	cacheDir    string
	cacheTTL    time.Duration
	mu          sync.RWMutex
	memoryCache map[string]*CacheItem
}

// NewCacheManager 创建缓存管理器实例
func NewCacheManager(cacheDir string, ttl time.Duration) (*CacheManager, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	cm := &CacheManager{
		cacheDir:    cacheDir,
		cacheTTL:    ttl,
		memoryCache: make(map[string]*CacheItem),
	}

	// 启动时加载所有缓存文件到内存
	if err := cm.loadCacheFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %v", err)
	}

	return cm, nil
}

// Get 获取缓存数据
func (cm *CacheManager) Get(platform string) (*CacheItem, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	item, exists := cm.memoryCache[platform]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item, true
}

// Set 设置缓存数据
func (cm *CacheManager) Set(platform string, data []models.TenantData) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	item := &CacheItem{
		Platform:  platform,
		Data:      data,
		UpdatedAt: now,
		ExpiresAt: now.Add(cm.cacheTTL),
	}

	// 更新内存缓存
	cm.memoryCache[platform] = item

	// 保存到磁盘
	return cm.saveToDisk(platform, item)
}

// GetLastUpdateTime 获取最后更新时间
func (cm *CacheManager) GetLastUpdateTime(platform string) *time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	item, exists := cm.memoryCache[platform]
	if !exists {
		return nil
	}

	return &item.UpdatedAt
}

// IsExpired 检查缓存是否过期
func (cm *CacheManager) IsExpired(platform string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	item, exists := cm.memoryCache[platform]
	if !exists {
		return true
	}

	return time.Now().After(item.ExpiresAt)
}

// Delete 删除缓存
func (cm *CacheManager) Delete(platform string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.memoryCache, platform)

	// 删除磁盘文件
	filePath := cm.getCacheFilePath(platform)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %v", err)
	}

	return nil
}

// GetAllPlatformsCacheInfo 获取所有平台的缓存信息
func (cm *CacheManager) GetAllPlatformsCacheInfo() map[string]models.CacheInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]models.CacheInfo)
	for platform, item := range cm.memoryCache {
		result[platform] = models.CacheInfo{
			Platform:  platform,
			UpdatedAt: item.UpdatedAt,
			ExpiresAt: item.ExpiresAt,
			IsExpired: time.Now().After(item.ExpiresAt),
			DataCount: len(item.Data),
		}
	}

	return result
}

// getCacheFilePath 获取缓存文件路径
func (cm *CacheManager) getCacheFilePath(platform string) string {
	filename := fmt.Sprintf("%s_cache.json", platform)
	return filepath.Join(cm.cacheDir, filename)
}

// saveToDisk 保存到磁盘
func (cm *CacheManager) saveToDisk(platform string, item *CacheItem) error {
	filePath := cm.getCacheFilePath(platform)

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}

// loadCacheFromDisk 从磁盘加载缓存
func (cm *CacheManager) loadCacheFromDisk() error {
	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(cm.cacheDir, file.Name())

			data, err := os.ReadFile(filePath)
			if err != nil {
				continue // 忽略读取失败的文件
			}

			var item CacheItem
			if err := json.Unmarshal(data, &item); err != nil {
				continue // 忽略解析失败的文件
			}

			// 只加载未过期的缓存
			if time.Now().Before(item.ExpiresAt) {
				cm.memoryCache[item.Platform] = &item
			} else {
				// 删除过期的缓存文件
				os.Remove(filePath)
			}
		}
	}

	return nil
}

// ClearExpired 清理过期缓存
func (cm *CacheManager) ClearExpired() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for platform, item := range cm.memoryCache {
		if now.After(item.ExpiresAt) {
			delete(cm.memoryCache, platform)

			// 删除磁盘文件
			filePath := cm.getCacheFilePath(platform)
			os.Remove(filePath) // 忽略错误
		}
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (cm *CacheManager) GetCacheStats() models.CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	totalItems := len(cm.memoryCache)
	expiredItems := 0
	now := time.Now()

	for _, item := range cm.memoryCache {
		if now.After(item.ExpiresAt) {
			expiredItems++
		}
	}

	return models.CacheStats{
		TotalItems:   totalItems,
		ExpiredItems: expiredItems,
		ValidItems:   totalItems - expiredItems,
		CacheDir:     cm.cacheDir,
		TTL:          cm.cacheTTL,
	}
}
