package cac

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"wm-func/common/config"
	"wm-func/tools/alter-data-v2/backend"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bmodel"
	"wm-func/tools/alter-data-v2/backend/tags"
)

type Cac struct {
}

type TenantDateSequence struct {
	TenantId        int64
	Last30DayDiff   int64
	RegisterTime    string
	DateSequence    []DateSequence
	Tags            []string `json:"tags"`              // 租户标签
	MissingDataDays int      `json:"missing_data_days"` // 缺数据天数
	HasMissingData  bool     `json:"has_missing_data"`  // 是否有缺数据
	NoiseDataDays   int      `json:"noise_data_days"`   // 噪声数据天数
	HasNoiseData    bool     `json:"has_noise_data"`    // 是否有噪声数据
	RecentZeroDays  int      `json:"recent_zero_days"`  // 最近3天为0的数量
	HasRecentZeros  bool     `json:"has_recent_zeros"`  // 是否有最近的零值
}

type DateSequence struct {
	Date       string `json:"date"`
	ApiData    int64  `json:"api_data"`
	Data       int64  `json:"data"`
	RemoveData int64  `json:"remove_data"`
	IsMissing  bool   `json:"is_missing"` // 是否为缺失数据
	IsNoise    bool   `json:"is_noise"`   // 是否为噪声数据
	NoiseType  string `json:"noise_type"` // 噪声类型：low_variance, outlier, zero_in_active等
}

func GenerateDateSequence() []DateSequence {
	now := time.Now()
	start := now.Add(config.DateDay * -90)
	var res []DateSequence

	for start.Before(now) {
		res = append(res, DateSequence{
			Date:       start.Format("2006-01-02"),
			ApiData:    0,
			Data:       0,
			RemoveData: 0,
		})
		start = start.Add(config.DateDay)
	}

	return res
}

func GetAlterDataWithPlatformWithTenantId(platform string, needRefresh bool, tenantId int64) ([]TenantDateSequence, []TenantDateSequence) {
	fmt.Println("GetAlterDataWithPlatformWithTenantId platform: ", platform)
	var b1 []bmodel.ApiData
	var b2 []bmodel.OverViewData

	if tenantId <= 0 {
		b1 = bdao.GetApiDataByPlatform(needRefresh, platform)
		b2 = bdao.GetOverviewDataByPlatform(needRefresh, platform)

		// 为 Shopify 添加详细调试
		if platform == backend.ADS_PLATFORM_SHOPIFY {
			fmt.Printf("🔍 [Shopify Debug] API数据: %d 条记录\n", len(b1))
			fmt.Printf("🔍 [Shopify Debug] Overview数据: %d 条记录\n", len(b2))

			// 收集租户ID
			apiTenants := make(map[int64]bool)
			overviewTenants := make(map[int64]bool)

			for _, v := range b1 {
				apiTenants[v.TenantId] = true
			}
			for _, v := range b2 {
				overviewTenants[v.TenantId] = true
			}

			fmt.Printf("🔍 [Shopify Debug] API数据包含 %d 个不同租户\n", len(apiTenants))
			fmt.Printf("🔍 [Shopify Debug] Overview数据包含 %d 个不同租户\n", len(overviewTenants))

			// 找共同租户
			commonTenants := make([]int64, 0)
			for tenantId := range apiTenants {
				if overviewTenants[tenantId] {
					commonTenants = append(commonTenants, tenantId)
				}
			}
			fmt.Printf("🎯 [Shopify Debug] 共同租户: %d 个\n", len(commonTenants))
			if len(commonTenants) > 0 {
				displayCount := len(commonTenants)
				if displayCount > 5 {
					displayCount = 5
				}
				fmt.Printf("🎯 [Shopify Debug] 前%d个共同租户: %v\n", displayCount, commonTenants[:displayCount])
			}
		}

		if platform == backend.ADS_PLATFORM_KNOCOMMERCE {
			for _, v := range b1 {
				fmt.Println("b1: ", v)
			}
			for _, v := range b2 {
				fmt.Println("b2: ", v)
			}
		}
	} else {
		b1 = bdao.GetApiDataByPlatformAndTenantId(needRefresh, platform, tenantId)
		b2 = bdao.GetOverviewDataByPlatformAndTenantId(needRefresh, platform, tenantId)
	}

	var apiDataMap = map[int64]map[string]bmodel.ApiData{}
	for _, v := range b1 {
		if apiDataMap[v.TenantId] == nil {
			apiDataMap[v.TenantId] = make(map[string]bmodel.ApiData)
		}
		apiDataMap[v.TenantId][v.RawDate] = v
	}

	var overviewDataMap = map[int64]map[string]bmodel.OverViewData{}
	for _, v := range b2 {
		if overviewDataMap[v.TenantId] == nil {
			overviewDataMap[v.TenantId] = make(map[string]bmodel.OverViewData)
		}
		overviewDataMap[v.TenantId][v.EventDate] = v
	}

	// 按租户分组：30天为界限
	last30Day := time.Now().Add(config.DateDay * -30)
	var allTenant = bmodel.GetAllTenant()
	var newTenants []TenantDateSequence
	var oldTenants []TenantDateSequence

	tenantPlatformMap := bmodel.GetTenantPlatformMap()

	// 为 Shopify 平台添加调试
	if platform == backend.ADS_PLATFORM_SHOPIFY {
		fmt.Printf("🔍 [Shopify Debug] 租户平台映射表包含 %d 个租户\n", len(tenantPlatformMap))
		shopifyTenantCount := 0
		for _, platforms := range tenantPlatformMap {
			if platforms[platform] {
				shopifyTenantCount++
			}
		}
		fmt.Printf("🔍 [Shopify Debug] 映射到 Shopify 的租户: %d 个\n", shopifyTenantCount)
	}

	for _, tenant := range allTenant {

		// 为 Shopify 添加详细的租户检查日志
		if platform == backend.ADS_PLATFORM_SHOPIFY && len(allTenant) > 0 {
			// 只对前几个租户打印调试信息
			if tenant.TenantId <= 134400 { // 只打印一些租户的调试信息
				hasMapping := tenantPlatformMap[tenant.TenantId][platform]
				fmt.Printf("🔍 [Shopify Debug] 租户 %d: 平台映射=%v\n", tenant.TenantId, hasMapping)
			}
		}

		if !tenantPlatformMap[tenant.TenantId][platform] {
			continue
		}

		tmp := GenerateDateSequence()
		for i, v := range tmp {
			if overviewData, exists := overviewDataMap[tenant.TenantId][v.Date]; exists {
				tmp[i].Data = overviewData.Value
			}
			if apiData, exists := apiDataMap[tenant.TenantId][v.Date]; exists {
				tmp[i].ApiData = apiData.AdSpend
			}
		}

		// 计算最近30天的diff
		last30DayDiff := calculateLast30DayDiff(tmp)

		// 检测缺数据和噪声数据
		missingDays, hasMissingData := detectMissingData(tmp)
		noiseDays, hasNoiseData := calculateNoiseStatistics(tmp)
		recentZeroDays, hasRecentZeros := calculateRecentZeroDays(tmp)

		// 处理租户标签
		needsMissingSetting := needsMissingSettingTag(tenant.TenantId, tmp)
		tenantTags := processTenantTags(tenant.TenantId, platform, hasMissingData, needsMissingSetting)

		tenantData := TenantDateSequence{
			TenantId:        tenant.TenantId,
			RegisterTime:    tenant.RegisterTime.Format("2006-01-02 15:04:05"),
			Last30DayDiff:   last30DayDiff,
			DateSequence:    tmp,
			Tags:            tenantTags,
			MissingDataDays: missingDays,
			HasMissingData:  hasMissingData,
			NoiseDataDays:   noiseDays,
			HasNoiseData:    hasNoiseData,
			RecentZeroDays:  recentZeroDays,
			HasRecentZeros:  hasRecentZeros,
		}

		// 分组：新租户 vs 老租户
		if tenant.RegisterTime.After(last30Day) {
			newTenants = append(newTenants, tenantData)
		} else {
			oldTenants = append(oldTenants, tenantData)
		}
	}

	// oldTenants 多级排序：1.有错误tag优先 2.最近零值天数多的优先 3.差值绝对值大的优先
	sort.Slice(oldTenants, func(i, j int) bool {
		return compareTenantsForSorting(oldTenants[i], oldTenants[j])
	})

	// newTenants 同样的多级排序
	sort.Slice(newTenants, func(i, j int) bool {
		return compareTenantsForSorting(newTenants[i], newTenants[j])
	})
	return newTenants, oldTenants

}

// compareTenantsForSorting 租户排序比较函数
// 排序优先级：1.只有错误tag的优先 2.有错误tag+正常tag的次之 3.只有正常tag的最后 4.最近零值天数多的优先 5.差值绝对值大的优先
func compareTenantsForSorting(a, b TenantDateSequence) bool {
	// 分析标签类型
	aErrorCount, aNormalCount := countTagTypes(a.Tags)
	bErrorCount, bNormalCount := countTagTypes(b.Tags)

	aHasError := aErrorCount > 0
	bHasError := bErrorCount > 0
	aHasNormal := aNormalCount > 0
	bHasNormal := bNormalCount > 0

	// 1. 只有错误标签的租户优先级最高
	aOnlyError := aHasError && !aHasNormal
	bOnlyError := bHasError && !bHasNormal

	if aOnlyError && !bOnlyError {
		return true
	}
	if !aOnlyError && bOnlyError {
		return false
	}

	// 2. 如果都是只有错误标签，或者都不是只有错误标签，继续比较
	// 有错误标签的优先于没有错误标签的
	if aHasError && !bHasError {
		return true
	}
	if !aHasError && bHasError {
		return false
	}

	// 3. 如果错误tag状态相同，比较最近零值天数
	if a.RecentZeroDays != b.RecentZeroDays {
		return a.RecentZeroDays > b.RecentZeroDays // 零值天数多的排在前面
	}

	// 4. 如果零值天数也相同，比较30天差异绝对值
	return math.Abs(float64(a.Last30DayDiff)) > math.Abs(float64(b.Last30DayDiff))
}

// hasErrorTag 检查是否有错误标签
func hasErrorTag(tags []string) bool {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "err_") {
			return true
		}
	}
	return false
}

// countTagTypes 统计错误标签和正常标签的数量
func countTagTypes(tags []string) (errorCount int, normalCount int) {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "err_") {
			errorCount++
		} else {
			normalCount++
		}
	}
	return errorCount, normalCount
}

func GetAlterDataWithPlatform(platform string, needRefresh bool) ([]TenantDateSequence, []TenantDateSequence) {
	fmt.Println(" GetAlterDataWithPlatform platform : ", platform)
	return GetAlterDataWithPlatformWithTenantId(platform, needRefresh, -1)
}

func calculateLast30DayDiff(dateSequences []DateSequence) int64 {
	now := time.Now()
	last30Day := now.Add(config.DateDay * -30)

	var totalDiff int64 = 0
	for _, seq := range dateSequences {
		seqDate, err := time.Parse("2006-01-02", seq.Date)
		if err != nil {
			continue
		}
		if seqDate.After(last30Day) {
			diff := seq.Data - seq.ApiData // 以ApiData为基准
			totalDiff += diff
		}
	}
	//if totalDiff < 0 {
	//	return totalDiff * -1
	//}
	return totalDiff
}

// detectMissingData 检测最近7天的数据异常，包括缺失和噪声
func detectMissingData(dateSequences []DateSequence) (int, bool) {
	checkDays := 7 // 检查最近7天
	missingDays := 0

	// 计算历史数据的统计信息（排除为0的数据）
	stats := calculateDataStatistics(dateSequences, checkDays)

	// 如果没有足够的历史数据，使用简单的0值检测
	if stats.validCount < 5 {
		return detectMissingDataSimple(dateSequences, checkDays)
	}

	// 从最后7天开始检查并标记
	for i := len(dateSequences) - checkDays; i < len(dateSequences); i++ {
		if i < 0 {
			continue
		}

		// 分析并标记数据类型
		analyzeAndTagData(&dateSequences[i], stats)

		if dateSequences[i].IsMissing {
			missingDays++
		}
	}

	// 缺少2天或以上就标记为缺数据
	hasMissingData := missingDays >= 2

	return missingDays, hasMissingData
}

// DataStatistics 数据统计信息
type DataStatistics struct {
	avg        float64 // 平均值
	stdDev     float64 // 标准差
	validCount int     // 有效数据点数量
	minValue   int64   // 最小值
	maxValue   int64   // 最大值
}

// calculateDataStatistics 计算历史数据的统计信息
func calculateDataStatistics(dateSequences []DateSequence, excludeLastDays int) DataStatistics {
	var validData []int64
	var sum int64 = 0

	// 收集有效数据（排除最后几天和0值）
	endIndex := len(dateSequences) - excludeLastDays
	for i := 0; i < endIndex; i++ {
		if dateSequences[i].Data > 0 {
			validData = append(validData, dateSequences[i].Data)
			sum += dateSequences[i].Data
		}
	}

	stats := DataStatistics{validCount: len(validData)}

	if len(validData) == 0 {
		return stats
	}

	// 计算平均值
	stats.avg = float64(sum) / float64(len(validData))

	// 计算标准差
	var variance float64 = 0
	stats.minValue = validData[0]
	stats.maxValue = validData[0]

	for _, value := range validData {
		diff := float64(value) - stats.avg
		variance += diff * diff

		if value < stats.minValue {
			stats.minValue = value
		}
		if value > stats.maxValue {
			stats.maxValue = value
		}
	}

	stats.stdDev = math.Sqrt(variance / float64(len(validData)))

	return stats
}

// analyzeAndTagData 分析并标记数据类型（缺失、噪声等）
func analyzeAndTagData(seq *DateSequence, stats DataStatistics) {
	value := seq.Data

	// 重置标记
	seq.IsMissing = false
	seq.IsNoise = false
	seq.NoiseType = ""

	// 如果数据为0，需要判断是否为真正的缺失还是噪声
	if value == 0 {
		if stats.avg <= stats.stdDev*2 {
			// 数据波动较大，0值可能是噪声而非缺失
			seq.IsNoise = true
			seq.NoiseType = "zero_in_high_variance"
		} else if stats.avg > 0 && stats.avg < 10 {
			// 平均值很小，0值可能是正常的低值噪声
			seq.IsNoise = true
			seq.NoiseType = "zero_in_low_average"
		} else {
			// 平均值相对较大，0值可能是真正的缺失
			seq.IsMissing = true
		}
		return
	}

	// 如果数据不为0，进行异常检测
	floatValue := float64(value)

	// 3-sigma规则检测异常值
	lowerBound := stats.avg - 3*stats.stdDev
	upperBound := stats.avg + 3*stats.stdDev

	if floatValue < lowerBound || floatValue > upperBound {
		// 超出3-sigma范围，标记为异常值噪声
		seq.IsNoise = true
		if floatValue < lowerBound {
			seq.NoiseType = "outlier_low"
		} else {
			seq.NoiseType = "outlier_high"
		}
		return
	}

	// 检查相对差异
	if stats.avg > 0 {
		relativeError := math.Abs(floatValue-stats.avg) / stats.avg

		if relativeError < 0.1 {
			// 差异小于10%，认为是正常数据
			return
		} else if relativeError < 0.3 {
			// 差异在10%-30%之间，可能是轻微噪声
			seq.IsNoise = true
			seq.NoiseType = "minor_deviation"
			return
		}
	}

	// 检查是否为异常的小值（可能的数据质量问题）
	if stats.avg > 100 && floatValue < stats.avg*0.1 {
		seq.IsNoise = true
		seq.NoiseType = "unusually_low"
	}
}

// detectMissingDataSimple 简单的缺数据检测（用于历史数据不足的情况）
func detectMissingDataSimple(dateSequences []DateSequence, checkDays int) (int, bool) {
	missingDays := 0

	// 从最后7天开始检查
	for i := len(dateSequences) - checkDays; i < len(dateSequences); i++ {
		if i < 0 {
			continue
		}

		if dateSequences[i].Data == 0 {
			dateSequences[i].IsMissing = true
			missingDays++
		}
	}

	hasMissingData := missingDays >= 2
	return missingDays, hasMissingData
}

// calculateNoiseStatistics 计算噪声数据统计
func calculateNoiseStatistics(dateSequences []DateSequence) (int, bool) {
	checkDays := 7 // 检查最近7天
	noiseDays := 0

	// 从最后7天开始统计噪声数据
	for i := len(dateSequences) - checkDays; i < len(dateSequences); i++ {
		if i < 0 {
			continue
		}

		if dateSequences[i].IsNoise {
			noiseDays++
		}
	}

	// 有1天或以上噪声就标记为有噪声数据
	hasNoiseData := noiseDays >= 1
	return noiseDays, hasNoiseData
}

// calculateRecentZeroDays 计算最近3天为0的数据天数
func calculateRecentZeroDays(dateSequences []DateSequence) (int, bool) {
	checkDays := 3 // 检查最近3天
	zeroDays := 0

	// 从最后3天开始统计零值数据
	for i := len(dateSequences) - checkDays; i < len(dateSequences); i++ {
		if i < 0 {
			continue
		}

		if dateSequences[i].Data == 0 {
			zeroDays++
		}
	}

	// 有1天或以上零值就标记为有最近零值
	hasRecentZeros := zeroDays >= 1
	return zeroDays, hasRecentZeros
}

// filterEmptyTags 过滤掉空字符串的标签
func filterEmptyTags(tags []string) []string {
	var validTags []string
	for _, tag := range tags {
		if len(tag) > 0 {
			validTags = append(validTags, tag)
		}
	}
	return validTags
}

// processTenantTags 处理租户标签，包括系统异常标签和用户标签的分类排序
func processTenantTags(tenantId int64, platform string, hasMissingData bool, needsMissingSetting bool) []string {
	// 获取用户自定义tags
	allTags := tags.GetAllTags(tenantId, platform)
	validTags := filterEmptyTags(allTags)

	var errorTags []string
	var normalTags []string

	// 1. 添加系统检测的异常tags
	if hasMissingData {
		errorTags = append(errorTags, "err_缺数")
	}
	if needsMissingSetting {
		errorTags = append(errorTags, "err_缺少setting")
	}

	// 2. 分类用户tags
	for _, tag := range validTags {
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

// needsMissingSettingTag 检查是否需要"缺少setting"标签
func needsMissingSettingTag(_ int64, _ []DateSequence) bool {
	// 这里可以添加具体的逻辑来判断是否缺少setting
	// 暂时返回false
	return false
}
