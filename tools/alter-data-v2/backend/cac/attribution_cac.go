package cac

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wm-func/common/config"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bmodel"
	"wm-func/tools/alter-data-v2/backend/tags"
)

// AttributionDateSequence 归因日期数据结构
type AttributionDateSequence struct {
	Date             string           `json:"date"`
	PlatformData     map[string]int64 `json:"platform_data"`     // 各平台归因数据
	TotalAttribution int64            `json:"total_attribution"` // 当日汇总
	IsRecentZero     bool             `json:"is_recent_zero"`    // 是否为最近3天的零值
}

// PlatformTotal 平台汇总数据
type PlatformTotal struct {
	Platform         string  `json:"platform"`
	TotalAttribution int64   `json:"total_attribution"`
	DailyAverage     float64 `json:"daily_average"`
}

// AttributionTenantData 归因租户数据
type AttributionTenantData struct {
	TenantId            int64                     `json:"tenant_id"`
	DateSequence        []AttributionDateSequence `json:"date_sequence"`
	PlatformTotals      []PlatformTotal           `json:"platform_totals"`
	TotalAttributionAvg float64                   `json:"total_attribution_avg"` // 总归因平均值
	Tags                []string                  `json:"tags"`
	RecentZeroDays      int                       `json:"recent_zero_days"` // 最近3天为0的数量
	HasRecentZeros      bool                      `json:"has_recent_zeros"` // 是否有最近的零值
	CustomerType        string                    `json:"customer_type"`    // 客户类型：new（新客户）或 old（老客户）
	RegisterTime        string                    `json:"register_time"`    // 注册时间
}

// AttributionData 解析归因JSON数据的结构
type AttributionData struct {
	Orders int64 `json:"orders"`
	Sales  int64 `json:"sales"`
}

// GenerateAttributionDateSequence 生成归因数据序列
func GenerateAttributionDateSequence() []AttributionDateSequence {
	now := time.Now()
	start := now.Add(config.DateDay * -90)
	var res []AttributionDateSequence

	for start.Before(now) {
		res = append(res, AttributionDateSequence{
			Date:             start.Format("2006-01-02"),
			PlatformData:     make(map[string]int64),
			TotalAttribution: 0,
			IsRecentZero:     false,
		})
		start = start.Add(config.DateDay)
	}

	return res
}

// calculateAttributionRecentZeroDays 计算最近3天零值数据
func calculateAttributionRecentZeroDays(dateSequences []AttributionDateSequence) (int, bool) {
	checkDays := 3 // 检查最近3天
	zeroDays := 0

	// 从最后3天开始统计零值数据
	for i := len(dateSequences) - checkDays; i < len(dateSequences); i++ {
		if i < 0 {
			continue
		}

		if dateSequences[i].TotalAttribution == 0 {
			dateSequences[i].IsRecentZero = true
			zeroDays++
		}
	}

	// 有1天或以上零值就标记为有最近零值
	hasRecentZeros := zeroDays >= 1
	return zeroDays, hasRecentZeros
}

// parseAttributionData 解析归因数据JSON
func parseAttributionData(dataStr string) (int64, error) {
	var attrData AttributionData
	if err := json.Unmarshal([]byte(dataStr), &attrData); err != nil {
		return 0, err
	}
	return attrData.Orders, nil
}

// calculatePlatformTotals 计算平台汇总数据
func calculatePlatformTotals(dateSequences []AttributionDateSequence) []PlatformTotal {
	platformTotals := make(map[string]int64)
	platformDays := make(map[string]int)

	// 统计各平台的总归因数和有效天数
	for _, seq := range dateSequences {
		for platform, value := range seq.PlatformData {
			platformTotals[platform] += value
			if value > 0 {
				platformDays[platform]++
			}
		}
	}

	var result []PlatformTotal
	for platform, total := range platformTotals {
		dailyAverage := float64(0)
		if platformDays[platform] > 0 {
			dailyAverage = float64(total) / float64(platformDays[platform])
		}

		result = append(result, PlatformTotal{
			Platform:         platform,
			TotalAttribution: total,
			DailyAverage:     dailyAverage,
		})
	}

	return result
}

// calculateTotalAttributionAverage 计算总归因平均值
func calculateTotalAttributionAverage(dateSequences []AttributionDateSequence) float64 {
	var totalSum int64
	var count int

	// 计算最近30天的平均值
	thirtyDaysAgo := time.Now().Add(config.DateDay * -30)

	for _, seq := range dateSequences {
		seqDate, err := time.Parse("2006-01-02", seq.Date)
		if err != nil {
			continue
		}

		// 只计算最近30天的数据
		if seqDate.After(thirtyDaysAgo) && seq.TotalAttribution > 0 {
			totalSum += seq.TotalAttribution
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return float64(totalSum) / float64(count)
}

// generateAttributionTags 生成归因相关标签
func generateAttributionTags(tenantId int64, platform string, hasRecentZeros bool) []string {
	var systemTags []string

	// 检查最近零值
	if hasRecentZeros {
		systemTags = append(systemTags, "err_归因缺失")
	}

	// 获取用户自定义tags
	userTags := tags.GetAllTags(tenantId, platform)
	validUserTags := filterEmptyTags(userTags)

	var errorTags []string
	var normalTags []string

	// 1. 添加系统检测的异常tags
	errorTags = append(errorTags, systemTags...)

	// 2. 分类用户tags
	for _, tag := range validUserTags {
		if strings.HasPrefix(tag, "err_") {
			errorTags = append(errorTags, tag)
		} else {
			normalTags = append(normalTags, tag)
		}
	}

	// 3. 异常tags在前，正常tags在后
	result := append(errorTags, normalTags...)
	return result
}

// ProcessAttributionData 处理归因数据并检测异常
func ProcessAttributionData(tenantId int64, needRefresh bool) AttributionTenantData {
	// 1. 获取租户信息（用于客户类型判断）
	allTenants := bmodel.GetAllTenant()
	var currentTenant *bmodel.AllTenant
	for _, tenant := range allTenants {
		if tenant.TenantId == tenantId {
			currentTenant = &tenant
			break
		}
	}

	// 2. 获取归因原始数据
	attributions := bdao.GetAttributionDataByTenantId(needRefresh, tenantId)

	// 3. 生成日期序列
	dateSequence := GenerateAttributionDateSequence()

	// 3. 填充归因数据
	attributionMap := make(map[string]map[string]int64) // date -> platform -> value
	for _, attr := range attributions {
		if attributionMap[attr.RawDate] == nil {
			attributionMap[attr.RawDate] = make(map[string]int64)
		}

		// 直接使用归因数据
		attributionMap[attr.RawDate][attr.AdsPlatform] = attr.Data
	}

	// 4. 合并数据到日期序列
	for i, seq := range dateSequence {
		if platformData, exists := attributionMap[seq.Date]; exists {
			dateSequence[i].PlatformData = platformData
			// 计算总归因
			total := int64(0)
			for _, value := range platformData {
				total += value
			}
			dateSequence[i].TotalAttribution = total
		}
	}

	// 5. 检测最近3天零值
	recentZeroDays, hasRecentZeros := calculateAttributionRecentZeroDays(dateSequence)

	// 6. 计算平台汇总
	platformTotals := calculatePlatformTotals(dateSequence)

	// 7. 计算总归因平均值
	totalAttributionAvg := calculateTotalAttributionAverage(dateSequence)

	// 8. 生成标签 (使用默认平台)
	attributionTags := generateAttributionTags(tenantId, "attribution", hasRecentZeros)

	// 9. 判断客户类型和设置注册时间
	customerType := "unknown"
	registerTime := ""
	if currentTenant != nil {
		registerTime = currentTenant.RegisterTime.Format("2006-01-02 15:04:05")
		// 判断是否为新客户（注册时间在30天内）
		last30Day := time.Now().Add(config.DateDay * -30)
		if currentTenant.RegisterTime.After(last30Day) {
			customerType = "new"
		} else {
			customerType = "old"
		}
	}

	return AttributionTenantData{
		TenantId:            tenantId,
		DateSequence:        dateSequence,
		PlatformTotals:      platformTotals,
		TotalAttributionAvg: totalAttributionAvg,
		Tags:                attributionTags,
		RecentZeroDays:      recentZeroDays,
		HasRecentZeros:      hasRecentZeros,
		CustomerType:        customerType,
		RegisterTime:        registerTime,
	}
}

// GetAttributionDataWithTenantId 获取特定租户的归因分析数据
func GetAttributionDataWithTenantId(tenantId int64, needRefresh bool) AttributionTenantData {
	return ProcessAttributionData(tenantId, needRefresh)
}

// GetAllAttributionData 获取所有租户的归因分析数据
func GetAllAttributionData(needRefresh bool) []AttributionTenantData {
	fmt.Printf("🔍 [GetAllAttributionData] 开始获取所有租户信息\n")
	tenantStartTime := time.Now()

	// 1. 获取所有租户
	allTenants := bdao.GetAllTenant()

	tenantDuration := time.Since(tenantStartTime)
	fmt.Printf("📊 [GetAllAttributionData] 获取租户信息完成 - 租户数量: %d, 耗时: %v\n",
		len(allTenants), tenantDuration)

	fmt.Printf("🔍 [GetAllAttributionData] 开始获取所有归因数据\n")
	attrStartTime := time.Now()

	// 2. 获取所有归因数据
	allAttributions := bdao.GetAttributionData(needRefresh)

	attrDuration := time.Since(attrStartTime)
	fmt.Printf("📊 [GetAllAttributionData] 获取归因数据完成 - 归因记录数: %d, 耗时: %v\n",
		len(allAttributions), attrDuration)

	fmt.Printf("🔍 [GetAllAttributionData] 开始按租户分组归因数据\n")
	groupStartTime := time.Now()

	// 3. 按租户分组归因数据
	tenantAttributionMap := make(map[int64][]bmodel.Attribution)
	for _, attr := range allAttributions {
		tenantAttributionMap[attr.TenantId] = append(tenantAttributionMap[attr.TenantId], attr)
	}

	groupDuration := time.Since(groupStartTime)
	fmt.Printf("📊 [GetAllAttributionData] 分组完成 - 有归因数据的租户: %d, 耗时: %v\n",
		len(tenantAttributionMap), groupDuration)

	var result []AttributionTenantData

	fmt.Printf("🔍 [GetAllAttributionData] 开始处理每个租户的数据\n")
	processStartTime := time.Now()

	// 4. 为每个有归因数据的租户生成分析数据
	processedCount := 0
	for _, tenant := range allTenants {
		// 只处理有归因数据的租户
		if attributions, exists := tenantAttributionMap[tenant.TenantId]; exists && len(attributions) > 0 {
			tenantData := processAttributionDataForTenant(tenant.TenantId, attributions)
			result = append(result, tenantData)
			processedCount++

			// 每处理100个租户打印一次进度
			if processedCount%100 == 0 {
				fmt.Printf("📊 [GetAllAttributionData] 已处理 %d 个租户\n", processedCount)
			}
		}
	}

	processDuration := time.Since(processStartTime)
	fmt.Printf("📊 [GetAllAttributionData] 数据处理完成 - 处理租户数: %d, 耗时: %v\n",
		processedCount, processDuration)

	return result
}

// processAttributionDataForTenant 为特定租户处理归因数据
func processAttributionDataForTenant(tenantId int64, attributions []bmodel.Attribution) AttributionTenantData {
	// 生成日期序列
	dateSequence := GenerateAttributionDateSequence()

	// 填充归因数据
	attributionMap := make(map[string]map[string]int64) // date -> platform -> value
	for _, attr := range attributions {
		if attributionMap[attr.RawDate] == nil {
			attributionMap[attr.RawDate] = make(map[string]int64)
		}

		// 直接使用归因数据
		attributionMap[attr.RawDate][attr.AdsPlatform] = attr.Data
	}

	// 合并数据到日期序列
	for i, seq := range dateSequence {
		if platformData, exists := attributionMap[seq.Date]; exists {
			dateSequence[i].PlatformData = platformData
			// 计算总归因
			total := int64(0)
			for _, value := range platformData {
				total += value
			}
			dateSequence[i].TotalAttribution = total
		}
	}

	// 检测最近3天零值
	recentZeroDays, hasRecentZeros := calculateAttributionRecentZeroDays(dateSequence)

	// 计算平台汇总
	platformTotals := calculatePlatformTotals(dateSequence)

	// 计算总归因平均值
	totalAttributionAvg := calculateTotalAttributionAverage(dateSequence)

	// 生成标签 (使用默认平台)
	attributionTags := generateAttributionTags(tenantId, "attribution", hasRecentZeros)

	// 获取租户信息以判断客户类型
	allTenants := bmodel.GetAllTenant()
	var currentTenant *bmodel.AllTenant
	for _, tenant := range allTenants {
		if tenant.TenantId == tenantId {
			currentTenant = &tenant
			break
		}
	}

	// 判断客户类型和设置注册时间
	customerType := "unknown"
	registerTime := ""
	if currentTenant != nil {
		registerTime = currentTenant.RegisterTime.Format("2006-01-02 15:04:05")
		// 判断是否为新客户（注册时间在30天内）
		last30Day := time.Now().Add(config.DateDay * -30)
		if currentTenant.RegisterTime.After(last30Day) {
			customerType = "new"
		} else {
			customerType = "old"
		}
	}

	return AttributionTenantData{
		TenantId:            tenantId,
		DateSequence:        dateSequence,
		PlatformTotals:      platformTotals,
		TotalAttributionAvg: totalAttributionAvg,
		Tags:                attributionTags,
		RecentZeroDays:      recentZeroDays,
		HasRecentZeros:      hasRecentZeros,
		CustomerType:        customerType,
		RegisterTime:        registerTime,
	}
}

// GetAttributionDataGroupedByCustomerType 获取按新老客户分组的归因数据
func GetAttributionDataGroupedByCustomerType(needRefresh bool) ([]AttributionTenantData, []AttributionTenantData) {
	fmt.Printf("🔍 [GetAttributionDataGroupedByCustomerType] 开始获取所有归因数据\n")
	allDataStartTime := time.Now()

	// 1. 获取所有归因数据
	allData := GetAllAttributionData(needRefresh)

	allDataDuration := time.Since(allDataStartTime)
	fmt.Printf("📊 [GetAttributionDataGroupedByCustomerType] 获取所有归因数据完成 - 数据量: %d, 耗时: %v\n",
		len(allData), allDataDuration)

	// 2. 按客户类型分组
	fmt.Printf("🔍 [GetAttributionDataGroupedByCustomerType] 开始按客户类型分组\n")
	groupStartTime := time.Now()

	var newCustomers []AttributionTenantData
	var oldCustomers []AttributionTenantData

	for _, tenantData := range allData {
		if tenantData.CustomerType == "new" {
			newCustomers = append(newCustomers, tenantData)
		} else if tenantData.CustomerType == "old" {
			oldCustomers = append(oldCustomers, tenantData)
		}
		// 忽略 unknown 类型的租户
	}

	groupDuration := time.Since(groupStartTime)
	fmt.Printf("📊 [GetAttributionDataGroupedByCustomerType] 分组完成 - 新客户: %d, 老客户: %d, 分组耗时: %v\n",
		len(newCustomers), len(oldCustomers), groupDuration)

	return newCustomers, oldCustomers
}
