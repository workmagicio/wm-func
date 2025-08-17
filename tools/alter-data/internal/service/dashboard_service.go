package service

import (
	"fmt"
	"log"
	"time"
	"wm-func/tools/alter-data/internal/cache"
	"wm-func/tools/alter-data/internal/config"
	"wm-func/tools/alter-data/internal/data"
	"wm-func/tools/alter-data/internal/platform"
	"wm-func/tools/alter-data/models"
)

// DashboardService 仪表板业务逻辑服务
type DashboardService struct {
	transformer  *data.DataTransformer
	cacheManager *cache.CacheManager
}

// NewDashboardService 创建仪表板服务实例
func NewDashboardService() *DashboardService {
	// 创建缓存管理器，缓存30分钟
	cacheManager, err := cache.NewCacheManager("./cache", 30*time.Minute)
	if err != nil {
		log.Printf("Failed to initialize cache manager: %v", err)
		// 如果缓存初始化失败，继续运行但不使用缓存
		cacheManager = nil
	}

	return &DashboardService{
		transformer:  data.NewDataTransformer(),
		cacheManager: cacheManager,
	}
}

// GetAvailablePlatforms 获取所有可用平台
func (s *DashboardService) GetAvailablePlatforms() []models.PlatformInfo {
	return config.GetAvailablePlatforms()
}

// GetPlatformData 获取指定平台的所有租户数据
func (s *DashboardService) GetPlatformData(platformName string) ([]models.TenantData, error) {
	return s.GetPlatformDataWithRefresh(platformName, false)
}

// GetPlatformDataWithRefresh 获取指定平台的所有租户数据，支持强制刷新
func (s *DashboardService) GetPlatformDataWithRefresh(platformName string, forceRefresh bool) ([]models.TenantData, error) {
	// 检查平台是否支持
	if !config.IsPlatformSupported(platformName) {
		return nil, fmt.Errorf("platform %s is not supported", platformName)
	}

	// 检查平台是否启用
	if !config.IsPlatformEnabled(platformName) {
		return nil, fmt.Errorf("platform %s is not enabled", platformName)
	}

	// 如果有缓存管理器且不强制刷新，先尝试从缓存获取
	if s.cacheManager != nil && !forceRefresh {
		if cachedItem, exists := s.cacheManager.Get(platformName); exists {
			log.Printf("Using cached data for platform %s (updated at %v)", platformName, cachedItem.UpdatedAt)
			return cachedItem.Data, nil
		}
	}

	log.Printf("Fetching fresh data for platform %s (force_refresh=%v)", platformName, forceRefresh)

	// 获取平台实现
	platformImpl, err := platform.GetPlatform(platformName)
	if err != nil {
		return nil, fmt.Errorf("platform %s is not implemented yet: %w", platformName, err)
	}

	// 获取原始数据
	rawDataMap, err := platformImpl.GetAllTenantsData(90)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from platform %s: %w", platformName, err)
	}

	// 转换为前端格式
	tenantDataList := s.transformer.TransformToTenantDataList(platformName, rawDataMap)

	// 如果有缓存管理器，保存到缓存
	if s.cacheManager != nil {
		if err := s.cacheManager.Set(platformName, tenantDataList); err != nil {
			log.Printf("Failed to cache data for platform %s: %v", platformName, err)
			// 缓存失败不影响主流程
		} else {
			log.Printf("Data cached for platform %s", platformName)
		}
	}

	return tenantDataList, nil
}

// GetTenantData 获取指定租户的平台数据
func (s *DashboardService) GetTenantData(platformName string, tenantID int64) (models.TenantData, error) {
	// 检查平台是否支持
	if !config.IsPlatformSupported(platformName) {
		return models.TenantData{}, fmt.Errorf("platform %s is not supported", platformName)
	}

	// 检查平台是否启用
	if !config.IsPlatformEnabled(platformName) {
		return models.TenantData{}, fmt.Errorf("platform %s is not enabled", platformName)
	}

	// 获取平台实现
	platformImpl, err := platform.GetPlatform(platformName)
	if err != nil {
		return models.TenantData{}, fmt.Errorf("platform %s is not implemented yet: %w", platformName, err)
	}

	// 获取原始数据
	rawData, err := platformImpl.GetTenantData(tenantID, 90)
	if err != nil {
		return models.TenantData{}, fmt.Errorf("failed to fetch data for tenant %d from platform %s: %w", tenantID, platformName, err)
	}

	// 转换为前端格式
	return s.transformer.TransformSingleTenantData(platformName, tenantID, rawData), nil
}

// GetCacheInfo 获取指定平台的缓存信息
func (s *DashboardService) GetCacheInfo(platformName string) *models.CacheInfo {
	if s.cacheManager == nil {
		return nil
	}

	updateTime := s.cacheManager.GetLastUpdateTime(platformName)
	if updateTime == nil {
		return nil
	}

	isExpired := s.cacheManager.IsExpired(platformName)

	return &models.CacheInfo{
		Platform:  platformName,
		UpdatedAt: *updateTime,
		ExpiresAt: updateTime.Add(30 * time.Minute), // TTL是30分钟
		IsExpired: isExpired,
	}
}

// GetAllCacheInfo 获取所有平台的缓存信息
func (s *DashboardService) GetAllCacheInfo() map[string]models.CacheInfo {
	if s.cacheManager == nil {
		return make(map[string]models.CacheInfo)
	}

	return s.cacheManager.GetAllPlatformsCacheInfo()
}

// RefreshPlatformCache 刷新指定平台的缓存
func (s *DashboardService) RefreshPlatformCache(platformName string) error {
	if s.cacheManager == nil {
		return fmt.Errorf("cache manager not available")
	}

	// 强制刷新数据
	_, err := s.GetPlatformDataWithRefresh(platformName, true)
	return err
}

// ClearPlatformCache 清空指定平台的缓存
func (s *DashboardService) ClearPlatformCache(platformName string) error {
	if s.cacheManager == nil {
		return fmt.Errorf("cache manager not available")
	}

	return s.cacheManager.Delete(platformName)
}

// GetCacheStats 获取缓存统计信息
func (s *DashboardService) GetCacheStats() models.CacheStats {
	if s.cacheManager == nil {
		return models.CacheStats{
			TotalItems:   0,
			ValidItems:   0,
			ExpiredItems: 0,
			CacheDir:     "缓存未启用",
			TTL:          0,
		}
	}

	return s.cacheManager.GetCacheStats()
}
