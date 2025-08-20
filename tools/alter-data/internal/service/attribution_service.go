package service

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/models"
)

// AttributionOrderService 归因订单分析服务
type AttributionOrderService struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

// NewAttributionOrderService 创建归因订单分析服务实例
func NewAttributionOrderService() *AttributionOrderService {
	return &AttributionOrderService{
		cache: make(map[string]interface{}),
	}
}

// GetAllTenantsAttributionOrders 获取所有租户的归因订单数据
func (s *AttributionOrderService) GetAllTenantsAttributionOrders(days int, forceRefresh bool) ([]models.AttributionOrderData, error) {
	cacheKey := fmt.Sprintf("all_tenants_%d_days", days)

	// 尝试从缓存读取
	if !forceRefresh {
		if cachedData, found := s.getFromCache(cacheKey); found {
			if data, ok := cachedData.([]models.AttributionOrderData); ok {
				log.Printf("归因订单数据缓存命中: %s", cacheKey)
				return data, nil
			}
		}
	}

	log.Printf("从数据库查询归因订单数据: %d 天", days)

	// 从数据库查询原始数据
	rawData, err := s.queryAttributionOrderRawData(days)
	if err != nil {
		return nil, fmt.Errorf("查询归因订单原始数据失败: %v", err)
	}

	// 数据转换和处理
	tenantData := s.processAttributionOrderData(rawData)

	// 按总订单量降序排序（优先显示订单量大的租户）
	s.sortTenantsByTotalOrders(tenantData)

	// 缓存结果
	s.setToCache(cacheKey, tenantData)
	log.Printf("归因订单数据已缓存: %s, 租户数量: %d", cacheKey, len(tenantData))

	return tenantData, nil
}

// GetTenantAttributionOrders 获取指定租户的归因订单数据
func (s *AttributionOrderService) GetTenantAttributionOrders(tenantID int64, days int, forceRefresh bool) (*models.AttributionOrderData, error) {
	cacheKey := fmt.Sprintf("tenant_%d_%d_days", tenantID, days)

	// 尝试从缓存读取
	if !forceRefresh {
		if cachedData, found := s.getFromCache(cacheKey); found {
			if data, ok := cachedData.(*models.AttributionOrderData); ok {
				log.Printf("租户归因订单数据缓存命中: %s", cacheKey)
				return data, nil
			}
		}
	}

	log.Printf("从数据库查询租户归因订单数据: tenant_id=%d, days=%d", tenantID, days)

	// 从数据库查询原始数据
	rawData, err := s.queryTenantAttributionOrderRawData(tenantID, days)
	if err != nil {
		return nil, fmt.Errorf("查询租户归因订单原始数据失败: %v", err)
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("未找到租户 %d 的归因订单数据", tenantID)
	}

	// 数据转换和处理
	tenantDataMap := s.processAttributionOrderData(rawData)
	if len(tenantDataMap) == 0 {
		return nil, fmt.Errorf("租户 %d 的归因订单数据处理失败", tenantID)
	}

	// 获取第一个（也是唯一一个）租户的数据
	for _, data := range tenantDataMap {
		// 缓存结果
		s.setToCache(cacheKey, &data)
		log.Printf("租户归因订单数据已缓存: %s", cacheKey)
		return &data, nil
	}

	return nil, fmt.Errorf("租户 %d 的归因订单数据为空", tenantID)
}

// queryAttributionOrderRawData 查询归因订单原始数据
func (s *AttributionOrderService) queryAttributionOrderRawData(days int) ([]models.AttributionOrderRawData, error) {
	db := platform_db.GetDB()

	sql := `
		SELECT tenant_id,
		       event_date,
		       ads_platform,
		       sum(attr_orders + extra_orders) as attr_orders
		FROM platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
		WHERE (event_date >= utc_date() - interval ? day)
		  AND json_overlaps(attr_model_array, json_array(0, 3))
		  AND attr_enhanced in (1, 4)
		  AND ads_platform != 'Unmatched'
		GROUP BY tenant_id, event_date, ads_platform
		ORDER BY tenant_id, event_date, ads_platform
	`

	var rawData []models.AttributionOrderRawData
	err := db.Raw(sql, days).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行归因订单查询失败: %v", err)
	}

	log.Printf("查询到归因订单原始数据: %d 条", len(rawData))
	return rawData, nil
}

// queryTenantAttributionOrderRawData 查询指定租户的归因订单原始数据
func (s *AttributionOrderService) queryTenantAttributionOrderRawData(tenantID int64, days int) ([]models.AttributionOrderRawData, error) {
	db := platform_db.GetDB()

	sql := `
		SELECT tenant_id,
		       event_date,
		       ads_platform,
		       sum(attr_orders + extra_orders) as attr_orders
		FROM platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
		WHERE (event_date >= utc_date() - interval ? day)
		  AND json_overlaps(attr_model_array, json_array(0, 3))
		  AND attr_enhanced in (1, 4)
		  AND ads_platform != 'Unmatched'
		  AND tenant_id = ?
		GROUP BY tenant_id, event_date, ads_platform
		ORDER BY event_date, ads_platform
	`

	var rawData []models.AttributionOrderRawData
	err := db.Raw(sql, days, tenantID).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行租户归因订单查询失败: %v", err)
	}

	log.Printf("查询到租户 %d 归因订单原始数据: %d 条", tenantID, len(rawData))
	return rawData, nil
}

// processAttributionOrderData 处理归因订单原始数据
func (s *AttributionOrderService) processAttributionOrderData(rawData []models.AttributionOrderRawData) []models.AttributionOrderData {
	// 按租户分组
	tenantMap := make(map[int64][]models.AttributionOrderRawData)
	for _, data := range rawData {
		tenantMap[data.TenantId] = append(tenantMap[data.TenantId], data)
	}

	var result []models.AttributionOrderData
	for tenantID, tenantRawData := range tenantMap {
		tenantData := s.processTenantData(tenantID, tenantRawData)
		if tenantData != nil {
			result = append(result, *tenantData)
		}
	}

	return result
}

// processTenantData 处理单个租户的数据
func (s *AttributionOrderService) processTenantData(tenantID int64, rawData []models.AttributionOrderRawData) *models.AttributionOrderData {
	if len(rawData) == 0 {
		return nil
	}

	// 构建日期和平台的映射
	dateSet := make(map[string]bool)
	platformSet := make(map[string]bool)
	dataMap := make(map[string]map[string]int64) // date -> platform -> orders

	for _, data := range rawData {
		dateSet[data.EventDate] = true
		platformSet[data.AdsPlatform] = true

		if dataMap[data.EventDate] == nil {
			dataMap[data.EventDate] = make(map[string]int64)
		}
		dataMap[data.EventDate][data.AdsPlatform] = data.AttrOrders
	}

	// 转换为有序数组
	var dates []string
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var platforms []string
	for platform := range platformSet {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)

	// 构建平台数据和总订单数
	platformData := make(map[string][]int64)
	totalOrders := make(map[string]int64)

	// 第一步：计算所有平台的原始总订单数，用于阈值判断
	rawTotalOrders := make(map[string]int64)
	var tenantTotalOrders int64 = 0

	for _, platform := range platforms {
		var platformTotal int64
		for _, date := range dates {
			if dataMap[date] != nil && dataMap[date][platform] > 0 {
				platformTotal += dataMap[date][platform]
			}
		}
		rawTotalOrders[platform] = platformTotal
		tenantTotalOrders += platformTotal
	}

	// 第二步：处理每个平台的数据
	for _, platform := range platforms {
		var orders []int64
		var total int64

		// 先构建原始数据序列
		rawOrders := make([]int64, len(dates))
		for i, date := range dates {
			if dataMap[date] != nil && dataMap[date][platform] > 0 {
				rawOrders[i] = dataMap[date][platform]
			} else {
				rawOrders[i] = 0
			}
		}

		// 判断是否需要进行异常检测
		platformTotal := rawTotalOrders[platform]
		avgOrders := float64(platformTotal) / float64(len(dates))
		shouldDetectAnomalies := true

		if platformTotal < tenantTotalOrders/100 || avgOrders < 50 {
			shouldDetectAnomalies = false
			log.Printf("平台 %s 数据量过小，跳过异常检测 (总订单: %d, 平均: %.1f, 租户总数: %d)",
				platform, platformTotal, avgOrders, tenantTotalOrders)
		}

		// 处理数据（根据是否需要异常检测）
		if shouldDetectAnomalies {
			// 智能检测"掉0"和"凹字形"异常模式
			orders = s.detectAnomalies(rawOrders)
		} else {
			// 直接使用原始数据，不进行异常检测
			orders = rawOrders
		}

		// 计算总订单数（排除异常标记）
		for _, orderCount := range orders {
			if orderCount > 0 {
				total += orderCount
			}
		}

		platformData[platform] = orders
		totalOrders[platform] = total
	}

	// 统计数据缺失异常
	hasConcave := false
	concaveCount := 0
	for _, orders := range platformData {
		for _, order := range orders {
			if order == -200 {
				hasConcave = true
				concaveCount++
			}
		}
	}

	return &models.AttributionOrderData{
		TenantID:     tenantID,
		TenantName:   fmt.Sprintf("Tenant %d", tenantID),
		DateRange:    dates,
		PlatformData: platformData,
		Platforms:    platforms,
		TotalOrders:  totalOrders,
		HasConcave:   hasConcave,
		ConcaveCount: concaveCount,
	}
}

// sortTenantsByTotalOrders 智能排序：数据缺失异常优先，然后按总订单量降序排序
func (s *AttributionOrderService) sortTenantsByTotalOrders(tenants []models.AttributionOrderData) {
	sort.Slice(tenants, func(i, j int) bool {
		// 第一优先级：数据缺失异常的排在前面
		if tenants[i].HasConcave != tenants[j].HasConcave {
			return tenants[i].HasConcave // true 排在前面
		}

		// 如果都有数据缺失异常，按异常数量降序排序
		if tenants[i].HasConcave && tenants[j].HasConcave {
			if tenants[i].ConcaveCount != tenants[j].ConcaveCount {
				return tenants[i].ConcaveCount > tenants[j].ConcaveCount
			}
		}

		// 第二优先级：按总订单量降序排序
		totalI := int64(0)
		for _, orders := range tenants[i].TotalOrders {
			totalI += orders
		}

		totalJ := int64(0)
		for _, orders := range tenants[j].TotalOrders {
			totalJ += orders
		}

		return totalI > totalJ
	})

	// 记录排序结果
	concaveTenantsCount := 0
	for _, tenant := range tenants {
		if tenant.HasConcave {
			concaveTenantsCount++
		}
	}
	log.Printf("租户排序完成: 数据缺失异常租户 %d 个已优先排序", concaveTenantsCount)
}

// getFromCache 从缓存获取数据
func (s *AttributionOrderService) getFromCache(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.cache[key]
	return value, found
}

// setToCache 设置缓存数据
func (s *AttributionOrderService) setToCache(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = value
}

// GetCacheInfo 获取缓存信息
func (s *AttributionOrderService) GetCacheInfo(cacheKey string) *models.CacheInfo {
	// 简化实现，返回基本缓存信息
	now := time.Now()
	return &models.CacheInfo{
		Platform:  "attribution_orders",
		UpdatedAt: now,
		ExpiresAt: now.Add(100 * 365 * 24 * time.Hour), // 100年后过期，实际上永不过期
		IsExpired: false,
		DataCount: len(s.cache),
	}
}

// detectAnomalies 智能检测异常模式：掉0和数据缺失
func (s *AttributionOrderService) detectAnomalies(rawOrders []int64) []int64 {
	result := make([]int64, len(rawOrders))
	copy(result, rawOrders)

	// 第一步：检测数据缺失模式（优先级高）
	concaveRanges := s.detectConcavePattern(rawOrders)

	// 标记数据缺失区域
	for _, concaveRange := range concaveRanges {
		for i := concaveRange.start; i <= concaveRange.end; i++ {
			if rawOrders[i] == 0 {
				result[i] = -200 // 数据缺失中的0用-200标记
			}
		}
	}

	// 第二步：检测普通掉0（不在数据缺失范围内的）
	for i := 0; i < len(rawOrders); i++ {
		if rawOrders[i] == 0 && result[i] == 0 { // 未被数据缺失标记的0
			// 检查是否为普通掉0：前面至少连续2天有数据，然后突然变成0
			if i >= 2 && rawOrders[i-1] > 0 && rawOrders[i-2] > 0 {
				result[i] = -100 // 普通掉0用-100标记
			}
		}
	}

	return result
}

// ConcaveRange 数据缺失范围
type ConcaveRange struct {
	start int // 开始位置（第一个0）
	end   int // 结束位置（最后一个0）
}

// detectConcavePattern 检测数据缺失模式：有数据 -> 连续0 -> 又有数据
func (s *AttributionOrderService) detectConcavePattern(rawOrders []int64) []ConcaveRange {
	var concaveRanges []ConcaveRange

	for i := 0; i < len(rawOrders); i++ {
		if rawOrders[i] == 0 {
			// 找到0的开始，检查前面是否有数据
			if i >= 2 && rawOrders[i-1] > 0 && rawOrders[i-2] > 0 {
				// 寻找连续0的结束位置
				zeroStart := i
				zeroEnd := i

				// 向后查找连续的0
				for j := i + 1; j < len(rawOrders) && rawOrders[j] == 0; j++ {
					zeroEnd = j
				}

				// 检查0结束后是否有数据恢复（至少连续2天）
				if zeroEnd+2 < len(rawOrders) {
					if rawOrders[zeroEnd+1] > 0 && rawOrders[zeroEnd+2] > 0 {
						// 确认是数据缺失：有数据 -> 连续0(至少1天) -> 有数据
						concaveRanges = append(concaveRanges, ConcaveRange{
							start: zeroStart,
							end:   zeroEnd,
						})
						log.Printf("检测到数据缺失异常: 位置 %d-%d", zeroStart, zeroEnd)
					}
				}

				// 跳过已处理的连续0
				i = zeroEnd
			}
		}
	}

	return concaveRanges
}

// ClearCache 清除缓存
func (s *AttributionOrderService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]interface{})
}
