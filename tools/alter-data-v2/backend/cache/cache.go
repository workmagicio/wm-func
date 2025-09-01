package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RemoveDataItem represents a single remove data entry
type RemoveDataItem struct {
	Date       string `json:"date"`
	RemoveData int64  `json:"remove_data"`
}

// TenantRemoveData represents all remove data for a tenant
type TenantRemoveData struct {
	TenantId    int64                     `json:"tenant_id"`
	Platform    string                    `json:"platform"`
	Data        map[string]RemoveDataItem `json:"data"` // date -> RemoveDataItem
	LastUpdated time.Time                 `json:"last_updated"`
}

// CacheManager manages the persistent cache for remove data
type CacheManager struct {
	mu       sync.RWMutex
	cacheDir string
	cache    map[string]*TenantRemoveData // key: "tenantId_platform"
}

var (
	instance *CacheManager
	once     sync.Once
)

// GetCacheManager returns the singleton cache manager
func GetCacheManager() *CacheManager {
	once.Do(func() {
		instance = &CacheManager{
			cacheDir: "cache/remove_data",
			cache:    make(map[string]*TenantRemoveData),
		}
		instance.loadFromDisk()
	})
	return instance
}

// getCacheKey generates a cache key for tenant and platform
func (c *CacheManager) getCacheKey(tenantId int64, platform string) string {
	return fmt.Sprintf("%d_%s", tenantId, platform)
}

// getCacheFilePath gets the file path for a cache key
func (c *CacheManager) getCacheFilePath(key string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("%s.json", key))
}

// ensureCacheDir creates the cache directory if it doesn't exist
func (c *CacheManager) ensureCacheDir() error {
	return os.MkdirAll(c.cacheDir, 0755)
}

// loadFromDisk loads all cache files from disk
func (c *CacheManager) loadFromDisk() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCacheDir(); err != nil {
		fmt.Printf("Error creating cache directory: %v\n", err)
		return
	}

	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		fmt.Printf("Error reading cache directory: %v\n", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			key := file.Name()[:len(file.Name())-5] // remove .json extension
			c.loadCacheFile(key)
		}
	}
}

// loadCacheFile loads a specific cache file
func (c *CacheManager) loadCacheFile(key string) {
	filePath := c.getCacheFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading cache file %s: %v\n", filePath, err)
		return
	}

	var tenantData TenantRemoveData
	if err := json.Unmarshal(data, &tenantData); err != nil {
		fmt.Printf("Error unmarshaling cache file %s: %v\n", filePath, err)
		return
	}

	c.cache[key] = &tenantData
}

// saveToDisk saves a specific tenant's data to disk
func (c *CacheManager) saveToDisk(key string, data *TenantRemoveData) error {
	if err := c.ensureCacheDir(); err != nil {
		return err
	}

	filePath := c.getCacheFilePath(key)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

// GetRemoveData retrieves remove data for a tenant and platform
func (c *CacheManager) GetRemoveData(tenantId int64, platform string) map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.getCacheKey(tenantId, platform)
	tenantData, exists := c.cache[key]
	if !exists {
		return make(map[string]int64)
	}

	result := make(map[string]int64)
	for date, item := range tenantData.Data {
		result[date] = item.RemoveData
	}
	return result
}

// SetRemoveData stores remove data for a tenant and platform
func (c *CacheManager) SetRemoveData(tenantId int64, platform string, removeDataMap map[string]int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.getCacheKey(tenantId, platform)

	// Create or update tenant data
	tenantData := &TenantRemoveData{
		TenantId:    tenantId,
		Platform:    platform,
		Data:        make(map[string]RemoveDataItem),
		LastUpdated: time.Now(),
	}

	for date, removeData := range removeDataMap {
		tenantData.Data[date] = RemoveDataItem{
			Date:       date,
			RemoveData: removeData,
		}
	}

	// Store in memory cache
	c.cache[key] = tenantData

	// Save to disk
	return c.saveToDisk(key, tenantData)
}

// HasRemoveData checks if remove data exists for a tenant and platform
func (c *CacheManager) HasRemoveData(tenantId int64, platform string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.getCacheKey(tenantId, platform)
	_, exists := c.cache[key]
	return exists
}

// GetCacheInfo returns cache information for debugging
func (c *CacheManager) GetCacheInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := make(map[string]interface{})
	info["cache_dir"] = c.cacheDir
	info["cache_count"] = len(c.cache)

	tenants := make([]string, 0, len(c.cache))
	for key := range c.cache {
		tenants = append(tenants, key)
	}
	info["cached_tenants"] = tenants

	return info
}
