package dao

import (
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

func GetOverviewDataByPlatform(isNeedRefresh bool, platform string) []bmodel.OverViewData {
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
