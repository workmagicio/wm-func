package services

import (
	"fmt"
	"sort"
	"strconv"
	"wm-func/tools/alter-data/config"
	"wm-func/tools/alter-data/models"
	"wm-func/tools/alter-data/platforms"
)

// DashboardService 仪表板业务逻辑服务
type DashboardService struct{}

// NewDashboardService 创建仪表板服务实例
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetAvailablePlatforms 获取所有可用平台
func (s *DashboardService) GetAvailablePlatforms() []models.PlatformInfo {
	return config.GetAvailablePlatforms()
}

// GetPlatformData 获取指定平台的所有租户数据
func (s *DashboardService) GetPlatformData(platformName string) ([]models.TenantData, error) {
	// 检查平台是否存在于配置中
	if !config.IsPlatformSupported(platformName) {
		return nil, fmt.Errorf("platform %s is not supported", platformName)
	}

	// 获取平台实现（目前只有google实现了）
	platform, err := platforms.GetPlatform(platformName)
	if err != nil {
		return nil, fmt.Errorf("platform %s is not implemented yet", platformName)
	}

	// 获取原始数据
	rawDataMap, err := platform.GetAllTenantsData(90)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from platform %s: %w", platformName, err)
	}

	// 转换为前端格式
	tenantDataList := s.convertToTenantData(platformName, rawDataMap)

	// 按差异值排序，差异最多的排在前面
	s.sortByDifference(tenantDataList)

	return tenantDataList, nil
}

// GetTenantData 获取指定租户的平台数据
func (s *DashboardService) GetTenantData(platformName string, tenantID int64) (models.TenantData, error) {
	// 检查平台是否存在
	if !config.IsPlatformSupported(platformName) {
		return models.TenantData{}, fmt.Errorf("platform %s is not supported", platformName)
	}

	// 获取平台实现
	platform, err := platforms.GetPlatform(platformName)
	if err != nil {
		return models.TenantData{}, fmt.Errorf("platform %s is not implemented yet", platformName)
	}

	// 获取原始数据
	rawData, err := platform.GetTenantData(tenantID, 90)
	if err != nil {
		return models.TenantData{}, fmt.Errorf("failed to fetch data for tenant %d from platform %s: %w", tenantID, platformName, err)
	}

	// 转换为前端格式
	return s.convertSingleTenantData(platformName, tenantID, rawData), nil
}

// convertToTenantData 将原始数据转换为前端需要的格式
func (s *DashboardService) convertToTenantData(platformName string, rawDataMap map[int64][]models.AlterData) []models.TenantData {
	var result []models.TenantData

	for tenantID, rawDataList := range rawDataMap {
		if len(rawDataList) == 0 {
			continue
		}

		tenantData := s.convertSingleTenantData(platformName, tenantID, rawDataList)
		result = append(result, tenantData)
	}

	return result
}

// convertSingleTenantData 转换单个租户的数据
func (s *DashboardService) convertSingleTenantData(platformName string, tenantID int64, rawData []models.AlterData) models.TenantData {
	tenantData := models.TenantData{
		TenantID:   tenantID,
		TenantName: s.generateTenantName(tenantID),
		Platform:   platformName,
		DateRange:  make([]string, 0, len(rawData)),
		APISpend:   make([]int64, 0, len(rawData)),
		AdSpend:    make([]int64, 0, len(rawData)),
		Difference: make([]int64, 0, len(rawData)),
	}

	// 转换数据
	for _, data := range rawData {
		tenantData.DateRange = append(tenantData.DateRange, data.RawDate)
		tenantData.APISpend = append(tenantData.APISpend, data.ApiSpend)
		tenantData.AdSpend = append(tenantData.AdSpend, data.AdSpend)
		tenantData.Difference = append(tenantData.Difference, data.ApiSpend-data.AdSpend)
	}

	return tenantData
}

// generateTenantName 生成租户名称（可以后续从数据库获取真实名称）
func (s *DashboardService) generateTenantName(tenantID int64) string {
	return "Tenant " + strconv.FormatInt(tenantID, 10)
}

// sortByDifference 按差异值排序租户数据，差异最多的排在前面
func (s *DashboardService) sortByDifference(tenantDataList []models.TenantData) {
	sort.Slice(tenantDataList, func(i, j int) bool {
		// 计算每个租户的总差异值
		totalDiffI := s.calculateTotalDifference(tenantDataList[i])
		totalDiffJ := s.calculateTotalDifference(tenantDataList[j])

		// 按总差异值降序排序（差异最多的在前面）
		return totalDiffI > totalDiffJ
	})
}

// calculateTotalDifference 计算租户的总差异值
func (s *DashboardService) calculateTotalDifference(tenantData models.TenantData) int64 {
	var totalDiff int64 = 0

	// 计算所有差异值的绝对值之和
	for _, diff := range tenantData.Difference {
		if diff < 0 {
			totalDiff += -diff // 取绝对值
		} else {
			totalDiff += diff
		}
	}

	return totalDiff
}
