package controller

import (
	"fmt"
	"strings"
	"time"
	"wm-func/common/config"
	"wm-func/tools/alter-data-v2/backend"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bdebug"
	"wm-func/tools/alter-data-v2/backend/bmodel"
	"wm-func/tools/alter-data-v2/backend/cac"
	"wm-func/tools/alter-data-v2/backend/cache"
	"wm-func/tools/alter-data-v2/backend/tags"
)

// applovinLog 平台需要监控的租户ID列表
var applovinLogTenantIds = []int64{150090}

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
func needsMissingSettingTag(tenant cac.TenantDateSequence) bool {
	// 检查过去30天的数据
	for _, item := range tenant.DateSequence {
		// 如果有 RemoveData，检查补齐后是否仍与 API数据 不匹配
		if item.RemoveData > 0 {
			dataWithRemove := item.Data + item.RemoveData
			apiData := item.ApiData
			// 如果补齐后仍然与API数据不匹配（允许小的误差）
			if apiData > 0 && dataWithRemove > 0 {
				diff := dataWithRemove - apiData
				if diff < 0 {
					diff = -diff
				}
				threshold := apiData / 20 // 5%的误差范围
				if threshold < 10 {
					threshold = 10
				}
				if diff > threshold {
					return true
				}
			}
		}
	}
	return false
}

// processTenantList 统一处理租户列表，添加标签和分组
func processTenantList(tenants []cac.TenantDateSequence, platform string) ([]TenantData, []TenantData) {
	tenantsWithTags := []TenantData{}
	tenantsWithoutTags := []TenantData{}

	for _, tenant := range tenants {
		// 检查是否需要"缺少setting"标签
		needsMissingSetting := needsMissingSettingTag(tenant)

		// 处理所有标签（包括系统异常标签）
		processedTags := processTenantTags(tenant.TenantId, platform, tenant.HasMissingData, needsMissingSetting)

		tenantData := TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          processedTags,
		}

		if len(processedTags) > 0 {
			tenantsWithTags = append(tenantsWithTags, tenantData)
		} else {
			tenantsWithoutTags = append(tenantsWithoutTags, tenantData)
		}
	}

	return tenantsWithTags, tenantsWithoutTags
}

func GetAlterDataWithPlatformWithTenantId(needRefresh bool, platform string, tenantId int64) AllTenantData {
	// 添加调试信息
	fmt.Printf("🎯 [GetAlterDataWithPlatformWithTenantId] 开始处理平台: %s, TenantId: %d, NeedRefresh: %v\n", platform, tenantId, needRefresh)

	// 根据平台类型选择不同的处理逻辑
	isWmOnly := backend.IsWmOnlyPlatform(platform)
	fmt.Printf("🔍 [GetAlterDataWithPlatformWithTenantId] 平台 %s 是否为 WM-Only: %v\n", platform, isWmOnly)

	if isWmOnly {
		fmt.Printf("📊 [GetAlterDataWithPlatformWithTenantId] 使用 WM-Only 处理逻辑\n")
		return getWmOnlyAlterData(needRefresh, platform, tenantId)
	} else {
		fmt.Printf("📊 [GetAlterDataWithPlatformWithTenantId] 使用双数据源处理逻辑\n")
		return getDualSourceAlterData(needRefresh, platform, tenantId)
	}
}

func getDualSourceAlterData(needRefresh bool, platform string, tenantId int64) AllTenantData {
	var res = AllTenantData{
		DataType: "dual_source",
	}

	var newTenants, oldTenants = []cac.TenantDateSequence{}, []cac.TenantDateSequence{}
	if tenantId < 0 {
		newTenants, oldTenants = cac.GetAlterDataWithPlatform(platform, needRefresh)
	} else {
		newTenants, oldTenants = cac.GetAlterDataWithPlatformWithTenantId(platform, needRefresh, tenantId)
	}

	// 获取数据最后加载时间
	res.DataLastLoadTime = bdao.GetDataLastLoadTime(platform)

	// 处理租户数据
	newTenantsWithTags, newTenantsWithoutTags := processTenantList(newTenants, platform)
	oldTenantsWithTags, oldTenantsWithoutTags := processTenantList(oldTenants, platform)

	// 合并数据：无tag的在前，有tag的在后
	res.NewTenants = append(newTenantsWithoutTags, newTenantsWithTags...)
	res.OldTenants = append(oldTenantsWithoutTags, oldTenantsWithTags...)

	// 确保返回的数组不为nil
	if res.NewTenants == nil {
		res.NewTenants = []TenantData{}
	}
	if res.OldTenants == nil {
		res.OldTenants = []TenantData{}
	}

	cacheManager := cache.GetCacheManager()

	if tenantId < 0 {
		// 读取缓存模式：为所有租户加载 RemoveData
		for i := range res.NewTenants {
			removeDataMap := cacheManager.GetRemoveData(res.NewTenants[i].TenantId, platform)
			for j := range res.NewTenants[i].DateSequence {
				if removeData, exists := removeDataMap[res.NewTenants[i].DateSequence[j].Date]; exists {
					res.NewTenants[i].DateSequence[j].RemoveData = removeData
				}
			}
		}

		for i := range res.OldTenants {
			removeDataMap := cacheManager.GetRemoveData(res.OldTenants[i].TenantId, platform)
			for j := range res.OldTenants[i].DateSequence {
				if removeData, exists := removeDataMap[res.OldTenants[i].DateSequence[j].Date]; exists {
					res.OldTenants[i].DateSequence[j].RemoveData = removeData
				}
			}
		}
	} else if tenantId > 0 {
		// 更新缓存模式：获取数据并缓存
		fmt.Printf("正在为租户 %d 获取 RemoveData...\n", tenantId)
		removeDataResult := bdebug.GetDataWithPlatform(tenantId, platform)

		// 将结果转换为缓存格式并存储
		removeDataMap := make(map[string]int64)
		for _, item := range removeDataResult {
			removeDataMap[item.StatDate] = item.Spend
		}

		// 存储到缓存
		err := cacheManager.SetRemoveData(tenantId, platform, removeDataMap)
		if err != nil {
			fmt.Printf("缓存 RemoveData 失败: %v\n", err)
		} else {
			fmt.Printf("成功缓存租户 %d 的 RemoveData，共 %d 条记录\n", tenantId, len(removeDataMap))
		}

		// 将新获取的数据合并到结果中
		for i := range res.NewTenants {
			if res.NewTenants[i].TenantId == tenantId {
				for j := range res.NewTenants[i].DateSequence {
					if removeData, exists := removeDataMap[res.NewTenants[i].DateSequence[j].Date]; exists {
						res.NewTenants[i].DateSequence[j].RemoveData = removeData
					}
				}
				break
			}
		}

		for i := range res.OldTenants {
			if res.OldTenants[i].TenantId == tenantId {
				for j := range res.OldTenants[i].DateSequence {
					if removeData, exists := removeDataMap[res.OldTenants[i].DateSequence[j].Date]; exists {
						res.OldTenants[i].DateSequence[j].RemoveData = removeData
					}
				}
				break
			}
		}
	}

	return res
}

func GetAlterDataWithPlatform(needRefresh bool, platform string) AllTenantData {
	return GetAlterDataWithPlatformWithTenantId(needRefresh, platform, -1)
}

// getWmOnlyAlterData 处理仅WM数据的平台
func getWmOnlyAlterData(needRefresh bool, platform string, tenantId int64) AllTenantData {
	fmt.Printf("🚀 [getWmOnlyAlterData] 开始处理 WM-Only 平台: %s, TenantId: %d\n", platform, tenantId)

	var res = AllTenantData{
		DataType: "wm_only",
	}

	// 只获取WM数据，不获取API数据
	var wmData []bmodel.WmData
	if tenantId < 0 {
		fmt.Printf("🔍 [getWmOnlyAlterData] 获取所有租户的WM数据\n")
		wmData = bdao.GetWmOnlyDataByPlatform(needRefresh, platform)
		fmt.Printf("📊 [getWmOnlyAlterData] 获取到 %s 平台的WM数据: %d 条记录\n", platform, len(wmData))
		if len(wmData) > 0 {
			fmt.Printf("🎯 [getWmOnlyAlterData] 示例WM数据: TenantId=%d, Date=%s, Data=%d\n",
				wmData[0].TenantId, wmData[0].RawDate, wmData[0].Data)
		}
	} else {
		// 对于特定租户，暂时获取所有数据然后过滤
		fmt.Printf("🔍 [getWmOnlyAlterData] 获取特定租户 %d 的WM数据\n", tenantId)
		allWmData := bdao.GetWmOnlyDataByPlatform(needRefresh, platform)
		fmt.Printf("📊 [getWmOnlyAlterData] 获取到 %s 平台的所有WM数据: %d 条记录\n", platform, len(allWmData))
		for _, data := range allWmData {
			if data.TenantId == tenantId {
				wmData = append(wmData, data)
			}
		}
		fmt.Printf("📊 [getWmOnlyAlterData] 过滤后租户 %d 的WM数据: %d 条记录\n", tenantId, len(wmData))
	}

	// 构建数据映射
	var wmDataMap = map[int64]map[string]bmodel.WmData{}
	for _, v := range wmData {
		if wmDataMap[v.TenantId] == nil {
			wmDataMap[v.TenantId] = make(map[string]bmodel.WmData)
		}
		wmDataMap[v.TenantId][v.RawDate] = v
	}

	var newTenants []TenantData
	var oldTenants []TenantData
	var lastDataDate string

	allTenant := bmodel.GetAllTenant()
	tenantPlatformMap := bmodel.GetTenantPlatformMap()
	last30Day := time.Now().Add(config.DateDay * -30)

	// 为 applovinLog 平台手动添加租户映射
	if platform == "applovinLog" {
		for _, tenantId := range applovinLogTenantIds {
			if tenantPlatformMap[tenantId] == nil {
				tenantPlatformMap[tenantId] = make(map[string]bool)
			}
			tenantPlatformMap[tenantId]["applovinLog"] = true
			fmt.Printf("为租户 %d 手动添加 applovinLog 平台映射\n", tenantId)
		}
	}

	// 为 shopify 平台手动添加租户映射
	if platform == "shopify" {
		fmt.Printf("🔍 [shopify] 开始处理 shopify 平台租户映射，当前WM数据包含的租户:\n")
		shopifyTenantIds := make(map[int64]bool)
		for _, data := range wmData {
			if !shopifyTenantIds[data.TenantId] {
				shopifyTenantIds[data.TenantId] = true
				fmt.Printf("   - 租户 %d\n", data.TenantId)
			}
		}

		fmt.Printf("🔧 [shopify] 为 %d 个租户添加 shopify 平台映射\n", len(shopifyTenantIds))
		for shopifyTenantId := range shopifyTenantIds {
			if tenantPlatformMap[shopifyTenantId] == nil {
				tenantPlatformMap[shopifyTenantId] = make(map[string]bool)
			}
			tenantPlatformMap[shopifyTenantId]["shopify"] = true
			fmt.Printf("   ✅ 为租户 %d 添加 shopify 平台映射\n", shopifyTenantId)
		}
	}

	for _, tenant := range allTenant {
		// 如果指定了租户ID，只处理该租户
		if tenantId > 0 && tenant.TenantId != tenantId {
			continue
		}
		//if tenant.TenantId != 150122 {
		//	continue
		//}

		if !tenantPlatformMap[tenant.TenantId][platform] {
			continue
		}

		tmp := cac.GenerateDateSequence()
		for i, v := range tmp {

			// 只填充WM数据，ApiData和RemoveData保持为0
			if dd, exists := wmDataMap[tenant.TenantId][v.Date]; exists {
				tmp[i].Data = dd.Data
				// tmp[i].ApiData = 0  // 保持默认值0
				// tmp[i].RemoveData = 0  // 保持默认值0
			}
		}

		// 计算最后有数据的日期
		tenantLastDataDate := getLastDataDate(tmp)
		if tenantLastDataDate > lastDataDate {
			lastDataDate = tenantLastDataDate
		}

		// WM-only特有的数据质量检测
		last7DaysHasData := checkLast7DaysHasData(tmp)

		// WM-only特有的标签处理
		wmOnlyTags := processWmOnlyTags(tenant.TenantId, platform, last7DaysHasData)

		tenantData := TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime.Format("2006-01-02 15:04:05"),
			Last30DayDiff: 0, // WM-only数据不计算差值
			DateSequence:  tmp,
			Tags:          wmOnlyTags,
		}

		if tenant.RegisterTime.After(last30Day) {
			newTenants = append(newTenants, tenantData)
		} else {
			oldTenants = append(oldTenants, tenantData)
		}
	}

	// 确保返回的数组不为nil
	if newTenants == nil {
		newTenants = []TenantData{}
	}
	if oldTenants == nil {
		oldTenants = []TenantData{}
	}

	res.NewTenants = newTenants
	res.OldTenants = oldTenants
	res.DataLastLoadTime = time.Now()
	res.LastDataDate = lastDataDate

	fmt.Printf("返回WM-only数据: 新租户 %d 个, 老租户 %d 个\n", len(newTenants), len(oldTenants))
	return res
}

// checkLast7DaysHasData 检查最近7天是否有数据
func checkLast7DaysHasData(sequences []cac.DateSequence) bool {
	checkDays := 7
	for i := len(sequences) - checkDays; i < len(sequences); i++ {
		if i >= 0 && sequences[i].Data > 0 {
			return true
		}
	}
	return false
}

// getLastDataDate 获取最后有数据的日期
func getLastDataDate(sequences []cac.DateSequence) string {
	for i := len(sequences) - 1; i >= 0; i-- {
		if sequences[i].Data > 0 {
			return sequences[i].Date
		}
	}
	return ""
}

// processWmOnlyTags 处理WM-only数据的标签
func processWmOnlyTags(tenantId int64, platform string, last7DaysHasData bool) []string {
	// 获取用户自定义tags
	allTags := tags.GetAllTags(tenantId, platform)
	validTags := filterEmptyTags(allTags)

	var errorTags []string
	var normalTags []string

	// WM-only特有的系统标签
	if !last7DaysHasData {
		errorTags = append(errorTags, "err_最近7天无数据")
	}

	// 分类用户tags
	for _, tag := range validTags {
		if strings.HasPrefix(tag, "err_") {
			errorTags = append(errorTags, tag)
		} else {
			normalTags = append(normalTags, tag)
		}
	}

	return append(errorTags, normalTags...)
}
