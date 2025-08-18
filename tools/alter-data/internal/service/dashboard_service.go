package service

import (
	"fmt"
	"log"
	"strings"
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
	// 创建缓存管理器，缓存永不过期（只在手动刷新时更新）
	// TTL参数保留但不再使用，实际缓存时间设为100年
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
		ExpiresAt: updateTime.Add(100 * 365 * 24 * time.Hour), // 100年后过期，实际上永不过期
		IsExpired: isExpired,                                  // 现在总是false（除非缓存不存在）
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

// GetTenantList 获取租户列表（带长期缓存，1天更新一次）
func (s *DashboardService) GetTenantList() ([]models.TenantInfo, error) {
	return s.GetTenantListWithRefresh(false)
}

// GetTenantListWithRefresh 获取租户列表，支持强制刷新
func (s *DashboardService) GetTenantListWithRefresh(forceRefresh bool) ([]models.TenantInfo, error) {
	cacheKey := "tenant_list"

	// 检查缓存（1天有效期）
	if s.cacheManager != nil && !forceRefresh {
		if cachedItem, exists := s.cacheManager.Get(cacheKey); exists {
			// 检查是否在1天内
			if time.Since(cachedItem.UpdatedAt) < 24*time.Hour {
				log.Printf("Using cached tenant list (updated at %v)", cachedItem.UpdatedAt)
				// 将缓存的TenantData转换为TenantInfo
				if tenantInfoList := s.convertTenantDataToTenantInfo(cachedItem.Data); len(tenantInfoList) > 0 {
					return tenantInfoList, nil
				}
			}
		}
	}

	log.Printf("Fetching fresh tenant list (force_refresh=%v)", forceRefresh)

	sql, exists := config.GetQuerySQL("tenants_list_query")
	if !exists {
		return nil, fmt.Errorf("tenant list query not found")
	}

	processor := data.NewDataProcessor()
	tenantList, err := processor.ExecuteTenantListQuery(sql)
	if err != nil {
		return nil, err
	}

	// 缓存租户列表（转换为TenantData格式以便使用现有缓存系统）
	if s.cacheManager != nil && len(tenantList) > 0 {
		tenantDataList := s.convertTenantInfoToTenantData(tenantList)
		if err := s.cacheManager.Set(cacheKey, tenantDataList); err != nil {
			log.Printf("Failed to cache tenant list: %v", err)
		} else {
			log.Printf("Tenant list cached (%d tenants)", len(tenantList))
		}
	}

	return tenantList, nil
}

// GetRecentRegisteredTenants 获取最近注册的租户列表
func (s *DashboardService) GetRecentRegisteredTenants() ([]models.TenantInfo, error) {
	return s.GetRecentRegisteredTenantsWithRefresh(false)
}

// GetRecentRegisteredTenantsWithRefresh 获取最近注册的租户列表，支持强制刷新
func (s *DashboardService) GetRecentRegisteredTenantsWithRefresh(forceRefresh bool) ([]models.TenantInfo, error) {
	cacheKey := "recent_registered_tenants"

	// 检查缓存（1小时有效期，比普通租户列表更频繁更新）
	if s.cacheManager != nil && !forceRefresh {
		if cachedItem, exists := s.cacheManager.Get(cacheKey); exists {
			// 检查是否在1小时内
			if time.Since(cachedItem.UpdatedAt) < 1*time.Hour {
				log.Printf("Using cached recent tenants (updated at %v)", cachedItem.UpdatedAt)
				// 将缓存的TenantData转换为TenantInfo
				if tenantInfoList := s.convertTenantDataToTenantInfo(cachedItem.Data); len(tenantInfoList) > 0 {
					return tenantInfoList, nil
				}
			}
		}
	}

	log.Printf("Fetching fresh recent registered tenants (force_refresh=%v)", forceRefresh)

	sql, exists := config.GetQuerySQL("recent_registered_tenants_query")
	if !exists {
		return nil, fmt.Errorf("recent registered tenants query not found")
	}

	processor := data.NewDataProcessor()
	recentTenants, err := processor.ExecuteRecentTenantsQuery(sql)
	if err != nil {
		return nil, err
	}

	// 缓存最近注册租户列表（转换为TenantData格式以便使用现有缓存系统）
	if s.cacheManager != nil && len(recentTenants) > 0 {
		tenantDataList := s.convertTenantInfoToTenantData(recentTenants)
		if err := s.cacheManager.Set(cacheKey, tenantDataList); err != nil {
			log.Printf("Failed to cache recent tenants: %v", err)
		} else {
			log.Printf("Recent tenants cached (%d tenants)", len(recentTenants))
		}
	}

	return recentTenants, nil
}

// convertTenantDataToTenantInfo 将TenantData转换为TenantInfo
func (s *DashboardService) convertTenantDataToTenantInfo(cachedData []models.TenantData) []models.TenantInfo {
	var result []models.TenantInfo

	// 使用map去重
	tenantMap := make(map[int64]string)
	for _, data := range cachedData {
		tenantMap[data.TenantID] = data.TenantName
	}

	for tenantID, tenantName := range tenantMap {
		result = append(result, models.TenantInfo{
			TenantID:   tenantID,
			TenantName: tenantName,
		})
	}

	return result
}

// convertTenantInfoToTenantData 将TenantInfo转换为TenantData以便缓存
func (s *DashboardService) convertTenantInfoToTenantData(tenantList []models.TenantInfo) []models.TenantData {
	var result []models.TenantData

	for _, tenant := range tenantList {
		result = append(result, models.TenantData{
			TenantID:   tenant.TenantID,
			TenantName: tenant.TenantName,
			Platform:   "tenant_list", // 标识这是租户列表数据
		})
	}

	return result
}

// GetTenantCrossPlatformData 获取指定租户的跨平台数据
func (s *DashboardService) GetTenantCrossPlatformData(tenantID int64) (models.CrossPlatformData, error) {
	return s.GetTenantCrossPlatformDataWithRefresh(tenantID, false)
}

// GetTenantCrossPlatformDataWithRefresh 获取指定租户的跨平台数据，支持强制刷新
func (s *DashboardService) GetTenantCrossPlatformDataWithRefresh(tenantID int64, forceRefresh bool) (models.CrossPlatformData, error) {
	cacheKey := fmt.Sprintf("tenant_%d", tenantID)

	// 如果有缓存管理器且不强制刷新，先尝试从缓存获取
	if s.cacheManager != nil && !forceRefresh {
		if cachedItem, exists := s.cacheManager.Get(cacheKey); exists {
			log.Printf("Using cached data for tenant %d (updated at %v)", tenantID, cachedItem.UpdatedAt)
			// 需要转换缓存的数据格式
			if crossPlatformData, ok := s.convertTenantDataToCrossPlatform(tenantID, cachedItem.Data); ok {
				return crossPlatformData, nil
			}
		}
	}

	log.Printf("Fetching fresh cross-platform data for tenant %d (force_refresh=%v)", tenantID, forceRefresh)

	// 访问记录已移到API层，这里不再记录

	sql, exists := config.GetQuerySQL("tenant_cross_platform_query")
	if !exists {
		return models.CrossPlatformData{}, fmt.Errorf("tenant cross-platform query not found")
	}

	// 跨平台查询需要16个参数，每个平台的API和Ads查询各需要一个tenantID参数
	// Google(2) + Meta(2) + AppLovin(2) + TikTok(2) + Shopify(2) + Snapchat(2) + TikTok Shop(2) + Pinterest(2) = 16个参数
	processor := data.NewDataProcessor()
	rawData, err := processor.ExecuteQueryWithParams(sql, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID, tenantID)
	if err != nil {
		return models.CrossPlatformData{}, fmt.Errorf("failed to fetch cross-platform data for tenant %d: %w", tenantID, err)
	}

	// 按平台分组数据
	platformGroups := processor.GroupByPlatform(rawData)

	// 转换为CrossPlatformData格式
	result := models.CrossPlatformData{
		TenantID:     tenantID,
		TenantName:   fmt.Sprintf("Tenant %d", tenantID),
		PlatformData: make(map[string][]models.TenantData),
	}

	// 为每个平台转换数据格式
	transformer := data.NewDataTransformer()
	for platform, platformData := range platformGroups {
		// 将[]AlterData转换为map[int64][]AlterData格式，以便使用现有的转换器
		groupedByTenant := make(map[int64][]models.AlterData)
		groupedByTenant[tenantID] = platformData

		// 使用现有的转换器
		tenantDataList := transformer.TransformToTenantDataList(strings.ToLower(platform), groupedByTenant)
		result.PlatformData[platform] = tenantDataList
	}

	// 如果有缓存管理器，保存到缓存
	if s.cacheManager != nil {
		// 将CrossPlatformData转换为[]TenantData格式以便缓存
		if flatData := s.convertCrossPlatformToTenantData(result); len(flatData) > 0 {
			if err := s.cacheManager.Set(cacheKey, flatData); err != nil {
				log.Printf("Failed to cache data for tenant %d: %v", tenantID, err)
			} else {
				log.Printf("Data cached for tenant %d", tenantID)
			}
		}
	}

	return result, nil
}

// convertTenantDataToCrossPlatform 将缓存的TenantData转换为CrossPlatformData
func (s *DashboardService) convertTenantDataToCrossPlatform(tenantID int64, cachedData []models.TenantData) (models.CrossPlatformData, bool) {
	if len(cachedData) == 0 {
		return models.CrossPlatformData{}, false
	}

	result := models.CrossPlatformData{
		TenantID:     tenantID,
		TenantName:   fmt.Sprintf("Tenant %d", tenantID),
		PlatformData: make(map[string][]models.TenantData),
	}

	// 按平台分组
	for _, tenantData := range cachedData {
		platform := tenantData.Platform
		result.PlatformData[platform] = append(result.PlatformData[platform], tenantData)
	}

	return result, true
}

// convertCrossPlatformToTenantData 将CrossPlatformData转换为[]TenantData以便缓存
func (s *DashboardService) convertCrossPlatformToTenantData(crossPlatformData models.CrossPlatformData) []models.TenantData {
	var result []models.TenantData

	for _, platformData := range crossPlatformData.PlatformData {
		result = append(result, platformData...)
	}

	return result
}

// GetTenantCacheInfo 获取指定租户的缓存信息
func (s *DashboardService) GetTenantCacheInfo(tenantID int64) *models.CacheInfo {
	if s.cacheManager == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("tenant_%d", tenantID)
	updateTime := s.cacheManager.GetLastUpdateTime(cacheKey)
	if updateTime == nil {
		return nil
	}

	isExpired := s.cacheManager.IsExpired(cacheKey)

	return &models.CacheInfo{
		Platform:  fmt.Sprintf("tenant_%d", tenantID),
		UpdatedAt: *updateTime,
		ExpiresAt: updateTime.Add(30 * time.Minute),
		IsExpired: isExpired,
	}
}

// RefreshTenantCache 刷新指定租户的缓存
func (s *DashboardService) RefreshTenantCache(tenantID int64) error {
	if s.cacheManager == nil {
		return fmt.Errorf("cache manager not available")
	}

	// 强制刷新数据
	_, err := s.GetTenantCrossPlatformDataWithRefresh(tenantID, true)
	return err
}

// GetFrequentTenants 获取经常访问的租户列表
func (s *DashboardService) GetFrequentTenants() ([]models.TenantAccessRecord, error) {
	if s.cacheManager == nil {
		return []models.TenantAccessRecord{}, fmt.Errorf("cache manager not available")
	}

	records := s.cacheManager.GetFrequentTenants()
	return records, nil
}

// RecordTenantAccess 记录租户访问（每次API调用都记录）
func (s *DashboardService) RecordTenantAccess(tenantID int64) {
	if s.cacheManager == nil {
		return
	}

	// 获取租户名称
	tenantName := fmt.Sprintf("Tenant %d", tenantID)
	if tenantList, err := s.GetTenantList(); err == nil {
		for _, tenant := range tenantList {
			if tenant.TenantID == tenantID {
				tenantName = tenant.TenantName
				break
			}
		}
	}

	s.cacheManager.RecordTenantAccess(tenantID, tenantName)
}
