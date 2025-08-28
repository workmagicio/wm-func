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

// FairingService Fairing分析服务
type FairingService struct {
	cache map[string]*models.FairingData
	mu    sync.RWMutex
}

// NewFairingService 创建Fairing分析服务实例
func NewFairingService() *FairingService {
	return &FairingService{
		cache: make(map[string]*models.FairingData),
	}
}

// GetAllTenantsFairing 获取所有租户的Fairing数据
func (s *FairingService) GetAllTenantsFairing(days int, forceRefresh bool) ([]models.FairingData, error) {
	// 固定使用90天进行分析
	analysisDays := 90
	cacheKey := fmt.Sprintf("fairing_all_%d_days", analysisDays)

	// 尝试从缓存读取
	if !forceRefresh {
		if _, found := s.getFromCache(cacheKey); found {
			log.Printf("Fairing数据缓存命中: %s", cacheKey)
			return s.convertCacheToSlice(), nil
		}
	}

	log.Printf("从数据库查询Fairing数据: %d 天", analysisDays)

	// 从数据库查询原始数据
	rawData, err := s.queryFairingRawData(analysisDays)
	if err != nil {
		return nil, fmt.Errorf("查询Fairing原始数据失败: %v", err)
	}

	// 数据转换和处理
	tenantData := s.processFairingData(rawData)

	// 按异常优先级排序
	s.sortTenantsByAnomalies(tenantData)

	// 永久缓存结果
	s.setPermanentCache(cacheKey, tenantData)
	log.Printf("Fairing数据已永久缓存: %s, 租户数量: %d", cacheKey, len(tenantData))

	return tenantData, nil
}

// GetTenantFairing 获取指定租户的Fairing数据
func (s *FairingService) GetTenantFairing(tenantID int64, forceRefresh bool) (*models.FairingData, error) {
	cacheKey := fmt.Sprintf("fairing_tenant_%d", tenantID)

	// 尝试从缓存读取
	if !forceRefresh {
		if cachedData, found := s.getFromCache(cacheKey); found {
			log.Printf("租户Fairing数据缓存命中: %s", cacheKey)
			return cachedData, nil
		}
	}

	log.Printf("从数据库查询租户Fairing数据: tenant_id=%d", tenantID)

	// 固定90天分析
	analysisDays := 90
	rawData, err := s.queryTenantFairingRawData(tenantID, analysisDays)
	if err != nil {
		return nil, fmt.Errorf("查询租户Fairing原始数据失败: %v", err)
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("未找到租户 %d 的Fairing数据", tenantID)
	}

	// 数据处理
	tenantDataMap := s.processFairingData(rawData)
	if len(tenantDataMap) == 0 {
		return nil, fmt.Errorf("租户 %d 的Fairing数据处理失败", tenantID)
	}

	// 获取第一个（也是唯一一个）租户的数据
	for _, data := range tenantDataMap {
		// 永久缓存结果
		s.setPermanentCache(cacheKey, []models.FairingData{data})
		log.Printf("租户Fairing数据已永久缓存: %s", cacheKey)
		return &data, nil
	}

	return nil, fmt.Errorf("租户 %d 的Fairing数据为空", tenantID)
}

// queryFairingRawData 查询Fairing原始数据
func (s *FairingService) queryFairingRawData(days int) ([]models.FairingRawData, error) {
	db := platform_db.GetDB()

	sql := `
		select
		    tenant_id,
		    date(response_provided_at) as stat_date,
		    count(1) as cnt
		from
		    platform_offline.dwd_view_post_survey_response_latest
		where response_provided_at > utc_timestamp() - interval ? day
		group by 1, 2
		order by tenant_id, stat_date
	`

	var rawData []models.FairingRawData
	err := db.Raw(sql, days).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行Fairing查询失败: %v", err)
	}

	log.Printf("查询到Fairing原始数据: %d 条", len(rawData))
	return rawData, nil
}

// queryTenantFairingRawData 查询指定租户的Fairing原始数据
func (s *FairingService) queryTenantFairingRawData(tenantID int64, days int) ([]models.FairingRawData, error) {
	db := platform_db.GetDB()

	sql := `
		select
		    tenant_id,
		    date(response_provided_at) as stat_date,
		    count(1) as cnt
		from
		    platform_offline.dwd_view_post_survey_response_latest
		where response_provided_at > utc_timestamp() - interval ? day
		and tenant_id = ?
		group by 1, 2
		order by stat_date
	`

	var rawData []models.FairingRawData
	err := db.Raw(sql, days, tenantID).Scan(&rawData).Error
	if err != nil {
		return nil, fmt.Errorf("执行租户Fairing查询失败: %v", err)
	}

	log.Printf("查询到租户 %d Fairing原始数据: %d 条", tenantID, len(rawData))
	return rawData, nil
}

// processFairingData 处理Fairing原始数据
func (s *FairingService) processFairingData(rawData []models.FairingRawData) []models.FairingData {
	// 按租户分组
	tenantMap := make(map[int64][]models.FairingRawData)
	for _, data := range rawData {
		tenantMap[data.TenantId] = append(tenantMap[data.TenantId], data)
	}

	var result []models.FairingData
	for tenantID, tenantRawData := range tenantMap {
		tenantData := s.processTenantData(tenantID, tenantRawData)
		if tenantData != nil {
			result = append(result, *tenantData)
		}
	}

	return result
}

// processTenantData 处理单个租户的数据
func (s *FairingService) processTenantData(tenantID int64, rawData []models.FairingRawData) *models.FairingData {
	if len(rawData) == 0 {
		return nil
	}

	// 构建日期映射
	dateSet := make(map[string]bool)
	dataMap := make(map[string]int64)

	for _, data := range rawData {
		dateSet[data.StatDate] = true
		dataMap[data.StatDate] = data.Cnt
	}

	// 转换为有序数组
	var dates []string
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// 构建响应数量数据序列
	responseData := make([]int64, len(dates))
	var totalResponses int64
	for i, date := range dates {
		if cnt, exists := dataMap[date]; exists {
			responseData[i] = cnt
			totalResponses += cnt
		} else {
			responseData[i] = 0
		}
	}

	// 计算90天平均值
	dailyAverage := float64(totalResponses) / float64(len(dates))

	// 检测异常和预警
	processedResponses, zeroDays, concaveCount := s.detectAnomalies(responseData, dailyAverage)
	warningLevel := s.calculateWarningLevel(zeroDays, concaveCount)
	hasAnomalies := zeroDays > 0 || concaveCount > 0

	return &models.FairingData{
		TenantID:           tenantID,
		TenantName:         fmt.Sprintf("Tenant %d", tenantID),
		DateRange:          dates,
		ResponseData:       responseData,
		ProcessedResponses: processedResponses,
		DailyAverage:       dailyAverage,
		WarningLevel:       warningLevel,
		TotalResponses:     totalResponses,
		ZeroDaysCount:      zeroDays,
		ConcaveCount:       concaveCount,
		HasAnomalies:       hasAnomalies,
		CacheTimestamp:     time.Now().Unix(),
	}
}

// detectAnomalies 检测异常：掉0和凹形问题
func (s *FairingService) detectAnomalies(responseData []int64, avgResponses float64) ([]int64, int, int) {
	processedResponses := make([]int64, len(responseData))
	copy(processedResponses, responseData)

	zeroDays := 0
	concaveCount := 0

	// 检测凹形问题（优先级高）
	concaveRanges := s.detectConcavePattern(responseData, avgResponses)

	// 标记凹形区域
	for _, concaveRange := range concaveRanges {
		for i := concaveRange.start; i <= concaveRange.end; i++ {
			if responseData[i] == 0 {
				processedResponses[i] = -200 // 凹形中的0用-200标记
				concaveCount++
			}
		}
	}

	// 检测普通掉0
	for i := 0; i < len(responseData); i++ {
		if responseData[i] == 0 && processedResponses[i] == 0 { // 未被凹形标记的0
			// 检查是否为掉0：前面至少连续2天有数据，然后突然变成0
			if i >= 2 && responseData[i-1] > 0 && responseData[i-2] > 0 {
				processedResponses[i] = -100 // 普通掉0用-100标记
			}
			zeroDays++
		}
	}

	return processedResponses, zeroDays, concaveCount
}

// FairingConcaveRange Fairing凹形范围
type FairingConcaveRange struct {
	start int // 开始位置
	end   int // 结束位置
}

// detectConcavePattern 检测凹形模式：有数据 -> 连续低值或0 -> 又有数据
func (s *FairingService) detectConcavePattern(responseData []int64, avgResponses float64) []FairingConcaveRange {
	var concaveRanges []FairingConcaveRange
	threshold := avgResponses * 0.3 // 低于平均值30%认为是异常低值

	for i := 0; i < len(responseData); i++ {
		if float64(responseData[i]) < threshold {
			// 找到低值的开始，检查前面是否有正常数据
			if i >= 2 && float64(responseData[i-1]) >= threshold && float64(responseData[i-2]) >= threshold {
				// 寻找连续低值的结束位置
				lowStart := i
				lowEnd := i

				// 向后查找连续的低值
				for j := i + 1; j < len(responseData) && float64(responseData[j]) < threshold; j++ {
					lowEnd = j
				}

				// 检查低值结束后是否有数据恢复（至少连续2天）
				if lowEnd+2 < len(responseData) {
					if float64(responseData[lowEnd+1]) >= threshold && float64(responseData[lowEnd+2]) >= threshold {
						// 确认是凹形：正常数据 -> 连续低值 -> 正常数据
						concaveRanges = append(concaveRanges, FairingConcaveRange{
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
func (s *FairingService) calculateWarningLevel(zeroDays, concaveCount int) string {
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
func (s *FairingService) sortTenantsByAnomalies(tenants []models.FairingData) {
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

		// 最后按总响应量降序排序
		return tenants[i].TotalResponses > tenants[j].TotalResponses
	})

	// 记录排序结果
	anomalyCount := 0
	for _, tenant := range tenants {
		if tenant.HasAnomalies {
			anomalyCount++
		}
	}
	log.Printf("Fairing租户排序完成: 异常租户 %d 个已优先排序", anomalyCount)
}

// getFromCache 从永久缓存获取数据
func (s *FairingService) getFromCache(key string) (*models.FairingData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.cache[key]
	return value, found
}

// setPermanentCache 设置永久缓存数据
func (s *FairingService) setPermanentCache(key string, data []models.FairingData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 为所有租户设置缓存
	for _, tenantData := range data {
		tenantKey := fmt.Sprintf("fairing_tenant_%d", tenantData.TenantID)
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
func (s *FairingService) convertCacheToSlice() []models.FairingData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.FairingData
	tenantMap := make(map[int64]bool)

	for key, data := range s.cache {
		// 只处理租户级别的缓存
		if len(key) > 15 && key[:15] == "fairing_tenant_" {
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
func (s *FairingService) GetCacheInfo() *models.CacheInfo {
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
		Platform:  "fairing",
		UpdatedAt: earliestTime,
		ExpiresAt: now.Add(100 * 365 * 24 * time.Hour), // 100年后过期，实际上永不过期
		IsExpired: false,
		DataCount: len(s.cache),
	}
}

// ClearCache 清除缓存
func (s *FairingService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*models.FairingData)
	log.Printf("Fairing分析缓存已清除")
}
