package bdao

import (
	"fmt"
	"time"
	"wm-func/tools/alter-data-v2/backend"
	"wm-func/tools/alter-data-v2/backend/bcache"
	"wm-func/tools/alter-data-v2/backend/bmodel"
)

func GetAllTenant() []bmodel.AllTenant {
	fmt.Printf("🔍 [GetAllTenant] 开始获取租户数据\n")
	startTime := time.Now()

	data := bmodel.GetAllTenant()

	duration := time.Since(startTime)
	fmt.Printf("📊 [GetAllTenant] 租户数据获取完成 - 租户数量: %d, 耗时: %v\n", len(data), duration)

	return data
}

func GetApiDataByPlatform(isNeedRefresh bool, platform string) []bmodel.ApiData {
	cacheKey := "apidata_" + platform

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.ApiData](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	}
}

func GetApiDataByPlatformAndTenantId(isNeedRefresh bool, platform string, tenantId int64) []bmodel.ApiData {
	cacheKey := fmt.Sprintf("apidata_%s_%d", platform, tenantId)

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetDataWithPlatformAndTenantId(platform, tenantId)
		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.ApiData](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetDataWithPlatformAndTenantId(platform, tenantId)
		bcache.SaveCache(cacheKey, data)
		return data
	}
}

func GetOverviewDataByPlatform(isNeedRefresh bool, platform string) []bmodel.OverViewData {
	platform = backend.PlatformMap[platform]
	cacheKey := "overview_" + platform

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetOverviewDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.OverViewData](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetOverviewDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	}
}

func GetOverviewDataByPlatformAndTenantId(isNeedRefresh bool, platform string, tenantId int64) []bmodel.OverViewData {
	platform = backend.PlatformMap[platform]
	cacheKey := fmt.Sprintf("overview_%s_%d", platform, tenantId)

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetOverviewDataWithPlatformAndTenantId(platform, tenantId)
		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.OverViewData](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetOverviewDataWithPlatformAndTenantId(platform, tenantId)
		bcache.SaveCache(cacheKey, data)
		return data
	}
}

// GetDataLastLoadTime 获取指定平台数据的最后加载时间
func GetDataLastLoadTime(platform string) time.Time {
	apiDataKey := "apidata_" + platform
	overviewKey := "overview_" + backend.PlatformMap[platform]

	var latestTime time.Time

	// 检查API数据缓存时间
	if apiCache, err := bcache.LoadCache(apiDataKey); err == nil {
		if apiCache.CreateTime.After(latestTime) {
			latestTime = apiCache.CreateTime
		}
	}

	// 检查wm_data缓存时间
	if overviewCache, err := bcache.LoadCache(overviewKey); err == nil {
		if overviewCache.CreateTime.After(latestTime) {
			latestTime = overviewCache.CreateTime
		}
	}

	// 如果没有找到缓存，返回当前时间
	if latestTime.IsZero() {
		latestTime = time.Now()
	}

	return latestTime
}

func QuerySingleData(platform string) ([]bmodel.WmData, error) {
	data := bmodel.GetSingleDataWithPlatform(platform)
	return data, nil
}

func GetWmOnlyDataByPlatform(isNeedRefresh bool, platform string) []bmodel.WmData {
	cacheKey := "wmdata_" + platform

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetSingleDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.WmData](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetSingleDataWithPlatform(platform)
		bcache.SaveCache(cacheKey, data)
		return data
	}
}

// GetAttributionData 获取归因数据（带缓存）
func GetAttributionData(isNeedRefresh bool) []bmodel.Attribution {
	cacheKey := "attribution_data"

	if isNeedRefresh {
		fmt.Printf("🔍 [GetAttributionData] 强制刷新 - 从数据库获取最新数据\n")
		dbStartTime := time.Now()

		// 强制刷新：直接从DB获取最新数据并更新缓存
		data := bmodel.GetAttrData()

		dbDuration := time.Since(dbStartTime)
		fmt.Printf("📊 [GetAttributionData] 数据库查询完成 - 记录数: %d, 耗时: %v\n", len(data), dbDuration)

		bcache.SaveCache(cacheKey, data)
		return data
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.Attribution](cacheKey); err == nil {
			fmt.Printf("✅ [GetAttributionData] 从缓存获取数据 - 记录数: %d\n", len(cachedData))
			return cachedData
		}

		fmt.Printf("🔍 [GetAttributionData] 缓存未命中 - 从数据库获取数据\n")
		dbStartTime := time.Now()

		// 缓存不存在，从DB获取并保存缓存
		data := bmodel.GetAttrData()

		dbDuration := time.Since(dbStartTime)
		fmt.Printf("📊 [GetAttributionData] 数据库查询完成 - 记录数: %d, 耗时: %v\n", len(data), dbDuration)

		bcache.SaveCache(cacheKey, data)
		return data
	}
}

// GetAttributionDataByTenantId 获取特定租户的归因数据（带缓存）
func GetAttributionDataByTenantId(isNeedRefresh bool, tenantId int64) []bmodel.Attribution {
	cacheKey := fmt.Sprintf("attribution_data_%d", tenantId)

	if isNeedRefresh {
		// 强制刷新：直接从DB获取最新数据并更新缓存
		allData := bmodel.GetAttrData()
		var tenantData []bmodel.Attribution
		for _, attr := range allData {
			if attr.TenantId == tenantId {
				tenantData = append(tenantData, attr)
			}
		}
		bcache.SaveCache(cacheKey, tenantData)
		return tenantData
	} else {
		// 优先缓存：尝试从缓存加载，失败则从DB获取并缓存
		if cachedData, err := bcache.LoadTyped[[]bmodel.Attribution](cacheKey); err == nil {
			return cachedData
		}

		// 缓存不存在，从DB获取并保存缓存
		allData := bmodel.GetAttrData()
		var tenantData []bmodel.Attribution
		for _, attr := range allData {
			if attr.TenantId == tenantId {
				tenantData = append(tenantData, attr)
			}
		}
		bcache.SaveCache(cacheKey, tenantData)
		return tenantData
	}
}
