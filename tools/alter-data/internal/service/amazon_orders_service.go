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

// AmazonOrdersService Amazon订单分析服务
type AmazonOrdersService struct {
	cache map[string]*models.AmazonOrderData
	mu    sync.RWMutex
}

// NewAmazonOrdersService 创建Amazon订单分析服务实例
func NewAmazonOrdersService() *AmazonOrdersService {
	return &AmazonOrdersService{
		cache: make(map[string]*models.AmazonOrderData),
	}
}

// GetAllTenantsAmazonOrders 获取所有租户的Amazon订单数据
func (s *AmazonOrdersService) GetAllTenantsAmazonOrders(days int, forceRefresh bool) ([]models.AmazonOrderData, error) {
	// 固定使用90天进行分析
	analysisDays := 90
	cacheKey := fmt.Sprintf("amazon_orders_all_%d_days", analysisDays)

	// 尝试从缓存读取
	if !forceRefresh {
		if _, found := s.getFromCache(cacheKey); found {
			log.Printf("Amazon订单数据缓存命中: %s", cacheKey)
			return s.convertCacheToSlice(), nil
		}
	}

	log.Printf("从数据库查询Amazon订单数据: %d 天", analysisDays)

	// 从数据库查询原始数据
	rawData, err := s.queryAmazonOrderRawData(analysisDays)
	if err != nil {
		return nil, fmt.Errorf("查询Amazon订单原始数据失败: %v", err)
	}

	// 数据转换和处理
	tenantData := s.processAmazonOrderData(rawData)

	// 按异常优先级排序
	s.sortTenantsByAnomalies(tenantData)

	// 永久缓存结果
	s.setPermanentCache(cacheKey, tenantData)
	log.Printf("Amazon订单数据已永久缓存: %s, 租户数量: %d", cacheKey, len(tenantData))

	return tenantData, nil
}

// GetTenantAmazonOrders 获取指定租户的Amazon订单数据
func (s *AmazonOrdersService) GetTenantAmazonOrders(tenantID int64, forceRefresh bool) (*models.AmazonOrderData, error) {
	cacheKey := fmt.Sprintf("amazon_orders_tenant_%d", tenantID)

	// 尝试从缓存读取
	if !forceRefresh {
		if cachedData, found := s.getFromCache(cacheKey); found {
			log.Printf("租户Amazon订单数据缓存命中: %s", cacheKey)
			return cachedData, nil
		}
	}

	log.Printf("从数据库查询租户Amazon订单数据: tenant_id=%d", tenantID)

	// 固定90天分析
	analysisDays := 90
	rawData, err := s.queryTenantAmazonOrderRawData(tenantID, analysisDays)
	if err != nil {
		return nil, fmt.Errorf("查询租户Amazon订单原始数据失败: %v", err)
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("未找到租户 %d 的Amazon订单数据", tenantID)
	}

	// 数据处理
	tenantDataMap := s.processAmazonOrderData(rawData)
	if len(tenantDataMap) == 0 {
		return nil, fmt.Errorf("租户 %d 的Amazon订单数据处理失败", tenantID)
	}

	// 获取第一个（也是唯一一个）租户的数据
	for _, data := range tenantDataMap {
		// 永久缓存结果
		s.setPermanentCache(cacheKey, []models.AmazonOrderData{data})
		log.Printf("租户Amazon订单数据已永久缓存: %s", cacheKey)
		return &data, nil
	}

	return nil, fmt.Errorf("租户 %d 的Amazon订单数据为空", tenantID)
}

// queryAmazonOrderRawData 查询Amazon订单原始数据
func (s *AmazonOrdersService) queryAmazonOrderRawData(days int) ([]models.AmazonOrderRawData, error) {
	db := platform_db.GetDB()

	sql := `
		select
		    tenant_id,
		    stat_date,
		    sum(shipped_units) as orders
		from
		    platform_offline.amazon_vendor_zip_code_daily_report
		where stat_date > utc_date() - interval ? day
		group by 1, 2
		order by tenant_id, stat_date
	`

	var rawData []models.AmazonOrderRawData
	err := db.Raw(sql, days).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行Amazon订单查询失败: %v", err)
	}

	log.Printf("查询到Amazon订单原始数据: %d 条", len(rawData))
	return rawData, nil
}

// queryTenantAmazonOrderRawData 查询指定租户的Amazon订单原始数据
func (s *AmazonOrdersService) queryTenantAmazonOrderRawData(tenantID int64, days int) ([]models.AmazonOrderRawData, error) {
	db := platform_db.GetDB()

	sql := `
		select
		    tenant_id,
		    stat_date,
		    sum(shipped_units) as orders
		from
		    platform_offline.amazon_vendor_zip_code_daily_report
		where stat_date > utc_date() - interval ? day
		and tenant_id = ?
		group by 1, 2
		order by stat_date
	`

	var rawData []models.AmazonOrderRawData
	err := db.Raw(sql, days, tenantID).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行租户Amazon订单查询失败: %v", err)
	}

	log.Printf("查询到租户 %d Amazon订单原始数据: %d 条", tenantID, len(rawData))
	return rawData, nil
}

// processAmazonOrderData 处理Amazon订单原始数据
func (s *AmazonOrdersService) processAmazonOrderData(rawData []models.AmazonOrderRawData) []models.AmazonOrderData {
	// 按租户分组
	tenantMap := make(map[int64][]models.AmazonOrderRawData)
	for _, data := range rawData {
		tenantMap[data.TenantId] = append(tenantMap[data.TenantId], data)
	}

	var result []models.AmazonOrderData
	for tenantID, tenantRawData := range tenantMap {
		tenantData := s.processTenantData(tenantID, tenantRawData)
		if tenantData != nil {
			result = append(result, *tenantData)
		}
	}

	return result
}

// processTenantData 处理单个租户的数据
func (s *AmazonOrdersService) processTenantData(tenantID int64, rawData []models.AmazonOrderRawData) *models.AmazonOrderData {
	if len(rawData) == 0 {
		return nil
	}

	// 构建日期映射
	dateSet := make(map[string]bool)
	dataMap := make(map[string]int64)

	for _, data := range rawData {
		dateSet[data.StatDate] = true
		dataMap[data.StatDate] = data.Orders
	}

	// 转换为有序数组
	var dates []string
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// 构建订单数据序列
	ordersData := make([]int64, len(dates))
	var totalOrders int64
	for i, date := range dates {
		if orders, exists := dataMap[date]; exists {
			ordersData[i] = orders
			totalOrders += orders
		} else {
			ordersData[i] = 0
		}
	}

	// 计算90天平均值
	dailyAverage := float64(totalOrders) / float64(len(dates))

	// 检测异常和预警
	processedOrders, zeroDays, concaveCount := s.detectAnomalies(ordersData, dailyAverage)
	warningLevel := s.calculateWarningLevel(zeroDays, concaveCount)
	hasAnomalies := zeroDays > 0 || concaveCount > 0

	return &models.AmazonOrderData{
		TenantID:        tenantID,
		TenantName:      fmt.Sprintf("Tenant %d", tenantID),
		DateRange:       dates,
		OrdersData:      ordersData,
		ProcessedOrders: processedOrders,
		DailyAverage:    dailyAverage,
		WarningLevel:    warningLevel,
		TotalOrders:     totalOrders,
		ZeroDaysCount:   zeroDays,
		ConcaveCount:    concaveCount,
		HasAnomalies:    hasAnomalies,
		CacheTimestamp:  time.Now().Unix(),
	}
}

// detectAnomalies 检测异常：掉0和凹形问题
func (s *AmazonOrdersService) detectAnomalies(ordersData []int64, avgOrders float64) ([]int64, int, int) {
	processedOrders := make([]int64, len(ordersData))
	copy(processedOrders, ordersData)

	zeroDays := 0
	concaveCount := 0

	// 检测凹形问题（优先级高）
	concaveRanges := s.detectConcavePattern(ordersData, avgOrders)

	// 标记凹形区域
	for _, concaveRange := range concaveRanges {
		for i := concaveRange.start; i <= concaveRange.end; i++ {
			if ordersData[i] == 0 {
				processedOrders[i] = -200 // 凹形中的0用-200标记
				concaveCount++
			}
		}
	}

	// 检测普通掉0
	for i := 0; i < len(ordersData); i++ {
		if ordersData[i] == 0 && processedOrders[i] == 0 { // 未被凹形标记的0
			// 检查是否为掉0：前面至少连续2天有数据，然后突然变成0
			if i >= 2 && ordersData[i-1] > 0 && ordersData[i-2] > 0 {
				processedOrders[i] = -100 // 普通掉0用-100标记
			}
			zeroDays++
		}
	}

	return processedOrders, zeroDays, concaveCount
}

// AmazonConcaveRange Amazon凹形范围
type AmazonConcaveRange struct {
	start int // 开始位置
	end   int // 结束位置
}

// detectConcavePattern 检测凹形模式：有数据 -> 连续低值或0 -> 又有数据
func (s *AmazonOrdersService) detectConcavePattern(ordersData []int64, avgOrders float64) []AmazonConcaveRange {
	var concaveRanges []AmazonConcaveRange
	threshold := avgOrders * 0.3 // 低于平均值30%认为是异常低值

	for i := 0; i < len(ordersData); i++ {
		if float64(ordersData[i]) < threshold {
			// 找到低值的开始，检查前面是否有正常数据
			if i >= 2 && float64(ordersData[i-1]) >= threshold && float64(ordersData[i-2]) >= threshold {
				// 寻找连续低值的结束位置
				lowStart := i
				lowEnd := i

				// 向后查找连续的低值
				for j := i + 1; j < len(ordersData) && float64(ordersData[j]) < threshold; j++ {
					lowEnd = j
				}

				// 检查低值结束后是否有数据恢复（至少连续2天）
				if lowEnd+2 < len(ordersData) {
					if float64(ordersData[lowEnd+1]) >= threshold && float64(ordersData[lowEnd+2]) >= threshold {
						// 确认是凹形：正常数据 -> 连续低值 -> 正常数据
						concaveRanges = append(concaveRanges, AmazonConcaveRange{
							start: lowStart,
							end:   lowEnd,
						})
						log.Printf("检测到凹形异常: 位置 %d-%d (阈值: %.2f)", lowStart, lowEnd, threshold)
					}
				}

				// 跳过已处理的连续低值
				i = lowEnd
			}
		}
	}

	return concaveRanges
}

// calculateWarningLevel 计算预警级别（基于掉0和凹形异常）
func (s *AmazonOrdersService) calculateWarningLevel(zeroDays, concaveCount int) string {
	// 有凹形问题直接critical
	if concaveCount > 0 {
		return "critical"
	}

	// 有掉0天数为warning
	if zeroDays > 0 {
		return "warning"
	}

	return "normal"
}

// sortTenantsByAnomalies 按异常优先级排序
func (s *AmazonOrdersService) sortTenantsByAnomalies(tenants []models.AmazonOrderData) {
	sort.Slice(tenants, func(i, j int) bool {
		// 第一优先级：凹形异常优先（critical级别）
		if tenants[i].ConcaveCount != tenants[j].ConcaveCount {
			return tenants[i].ConcaveCount > tenants[j].ConcaveCount
		}

		// 第二优先级：掉0天数多的排在前面
		if tenants[i].ZeroDaysCount != tenants[j].ZeroDaysCount {
			return tenants[i].ZeroDaysCount > tenants[j].ZeroDaysCount
		}

		// 第三优先级：有异常的排在前面
		if tenants[i].HasAnomalies != tenants[j].HasAnomalies {
			return tenants[i].HasAnomalies
		}

		// 最后按总订单量降序排序
		return tenants[i].TotalOrders > tenants[j].TotalOrders
	})

	// 记录排序结果
	anomalyCount := 0
	for _, tenant := range tenants {
		if tenant.HasAnomalies {
			anomalyCount++
		}
	}
	log.Printf("Amazon订单租户排序完成: 异常租户 %d 个已优先排序", anomalyCount)
}

// getFromCache 从永久缓存获取数据
func (s *AmazonOrdersService) getFromCache(key string) (*models.AmazonOrderData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.cache[key]
	return value, found
}

// setPermanentCache 设置永久缓存数据
func (s *AmazonOrdersService) setPermanentCache(key string, data []models.AmazonOrderData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 为所有租户设置缓存
	for _, tenantData := range data {
		tenantKey := fmt.Sprintf("amazon_orders_tenant_%d", tenantData.TenantID)
		cachedData := tenantData
		cachedData.CacheTimestamp = time.Now().Unix()
		s.cache[tenantKey] = &cachedData
	}

	// 设置全量查询的缓存标记
	if len(data) > 0 {
		s.cache[key] = &data[0] // 使用第一个数据作为标记
	}
}

// convertCacheToSlice 将缓存转换为数组
func (s *AmazonOrdersService) convertCacheToSlice() []models.AmazonOrderData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.AmazonOrderData
	tenantMap := make(map[int64]bool)

	for key, data := range s.cache {
		// 只处理租户级别的缓存
		if len(key) > 20 && key[:20] == "amazon_orders_tenant" {
			if !tenantMap[data.TenantID] {
				result = append(result, *data)
				tenantMap[data.TenantID] = true
			}
		}
	}

	// 重新排序
	s.sortTenantsByAnomalies(result)
	return result
}

// GetCacheInfo 获取缓存信息
func (s *AmazonOrdersService) GetCacheInfo() *models.CacheInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	// 找到最早的缓存时间
	var earliestTime time.Time
	for _, data := range s.cache {
		cacheTime := time.Unix(data.CacheTimestamp, 0)
		if earliestTime.IsZero() || cacheTime.Before(earliestTime) {
			earliestTime = cacheTime
		}
	}

	if earliestTime.IsZero() {
		earliestTime = now
	}

	return &models.CacheInfo{
		Platform:  "amazon_orders",
		UpdatedAt: earliestTime,
		ExpiresAt: now.Add(100 * 365 * 24 * time.Hour), // 100年后过期，实际上永不过期
		IsExpired: false,
		DataCount: len(s.cache),
	}
}

// ClearCache 清除缓存
func (s *AmazonOrdersService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*models.AmazonOrderData)
	log.Printf("Amazon订单分析缓存已清除")
}
