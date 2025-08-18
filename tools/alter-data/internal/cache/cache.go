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
	cacheDir      string
	cacheTTL      time.Duration
	mu            sync.RWMutex
	memoryCache   map[string]*CacheItem
	accessRecords map[int64]*models.TenantAccessRecord // 租户访问记录
}

// NewCacheManager 创建缓存管理器实例
func NewCacheManager(cacheDir string, ttl time.Duration) (*CacheManager, error) {
	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	cm := &CacheManager{
		cacheDir:      cacheDir,
		cacheTTL:      ttl,
		memoryCache:   make(map[string]*CacheItem),
		accessRecords: make(map[int64]*models.TenantAccessRecord),
	}

	// 启动时加载所有缓存文件到内存
	if err := cm.loadCacheFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %v", err)
	}

	// 加载访问记录
	if err := cm.loadAccessRecordsFromDisk(); err != nil {
		// 访问记录加载失败不影响缓存管理器初始化，只记录错误
		fmt.Printf("Warning: failed to load access records: %v\n", err)
	}

	// 启动定时保存访问记录的协程（每3分钟保存一次）
	go cm.startPeriodicSave()

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
		ExpiresAt: now.Add(100 * 365 * 24 * time.Hour), // 100年后过期，实际上永不过期
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

// IsExpired 检查缓存是否过期 (现在永不过期，只在手动刷新时更新)
func (cm *CacheManager) IsExpired(platform string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, exists := cm.memoryCache[platform]
	if !exists {
		return true // 如果不存在，视为过期
	}

	return false // 缓存永不过期，只能手动刷新
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
			IsExpired: false, // 缓存永不过期
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

	loadedCount := 0
	updatedCount := 0

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

			// 加载所有缓存，并更新为永不过期
			// 如果是旧的过期缓存，更新过期时间为100年后
			now := time.Now()
			if now.After(item.ExpiresAt) {
				// 更新过期时间为永不过期
				item.ExpiresAt = now.Add(100 * 365 * 24 * time.Hour)
				// 保存更新后的缓存到磁盘
				if err := cm.saveToDisk(item.Platform, &item); err != nil {
					// 保存失败只记录日志，不影响加载
					fmt.Printf("Warning: failed to update cache file %s: %v\n", file.Name(), err)
				} else {
					updatedCount++
				}
			}
			cm.memoryCache[item.Platform] = &item
			loadedCount++
		}
	}

	// 输出加载统计信息
	if loadedCount > 0 {
		fmt.Printf("Cache startup: loaded %d platforms, updated %d expired caches to never-expire\n", loadedCount, updatedCount)
	} else {
		fmt.Printf("Cache startup: no cache files found, starting with empty cache\n")
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
	expiredItems := 0 // 现在没有过期项，因为缓存永不过期

	return models.CacheStats{
		TotalItems:   totalItems,
		ExpiredItems: expiredItems,
		ValidItems:   totalItems, // 所有项都有效
		CacheDir:     cm.cacheDir,
		TTL:          cm.cacheTTL, // 保留原TTL信息（虽然不再使用）
	}
}

// RecordTenantAccess 记录租户访问
func (cm *CacheManager) RecordTenantAccess(tenantID int64, tenantName string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()

	if record, exists := cm.accessRecords[tenantID]; exists {
		// 更新现有记录
		record.AccessCount++
		record.LastAccess = now
		record.TenantName = tenantName // 更新租户名称（可能会变）
	} else {
		// 创建新记录
		cm.accessRecords[tenantID] = &models.TenantAccessRecord{
			TenantID:    tenantID,
			TenantName:  tenantName,
			AccessCount: 1,
			LastAccess:  now,
			FirstAccess: now,
		}
	}

	// 不再每次访问都保存，改为定时保存（每3分钟一次）
}

// GetFrequentTenants 获取经常访问的租户（按访问次数排序，取前20个）
func (cm *CacheManager) GetFrequentTenants() []models.TenantAccessRecord {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 收集所有访问记录
	var records []models.TenantAccessRecord
	for _, record := range cm.accessRecords {
		// 只返回最近30天内有访问的租户
		if time.Since(record.LastAccess) <= 30*24*time.Hour {
			records = append(records, *record)
		}
	}

	// 按访问次数排序（降序）
	for i := 0; i < len(records)-1; i++ {
		for j := 0; j < len(records)-1-i; j++ {
			if records[j].AccessCount < records[j+1].AccessCount {
				records[j], records[j+1] = records[j+1], records[j]
			}
		}
	}

	// 返回前20个
	if len(records) > 20 {
		records = records[:20]
	}

	return records
}

// getAccessRecordsFilePath 获取访问记录文件路径
func (cm *CacheManager) getAccessRecordsFilePath() string {
	return filepath.Join(cm.cacheDir, "access_records.json")
}

// saveAccessRecordsToDisk 保存访问记录到磁盘
func (cm *CacheManager) saveAccessRecordsToDisk() error {
	filePath := cm.getAccessRecordsFilePath()

	data, err := json.MarshalIndent(cm.accessRecords, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal access records: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write access records file: %v", err)
	}

	return nil
}

// startPeriodicSave 启动定时保存访问记录的协程（每3分钟保存一次）
func (cm *CacheManager) startPeriodicSave() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 定时保存访问记录
			if err := cm.saveAccessRecordsToDisk(); err != nil {
				// 保存失败只记录日志，不影响程序运行
				fmt.Printf("Warning: failed to save access records: %v\n", err)
			} else {
				fmt.Printf("Access records saved successfully at %v\n", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}
}

// loadAccessRecordsFromDisk 从磁盘加载访问记录
func (cm *CacheManager) loadAccessRecordsFromDisk() error {
	filePath := cm.getAccessRecordsFilePath()

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，初始化为空记录
		cm.accessRecords = make(map[int64]*models.TenantAccessRecord)
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read access records file: %v", err)
	}

	var records map[int64]*models.TenantAccessRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to unmarshal access records: %v", err)
	}

	cm.accessRecords = records
	return nil
}
