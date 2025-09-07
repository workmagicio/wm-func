package bdao

import (
	"fmt"
	"time"
	"wm-func/tools/alter-data-v2/backend"
	"wm-func/tools/alter-data-v2/backend/bcache"
	"wm-func/tools/alter-data-v2/backend/bmodel"
)

func GetAllTenant() []bmodel.AllTenant {
	return bmodel.GetAllTenant()
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
