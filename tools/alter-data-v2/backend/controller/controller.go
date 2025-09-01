package controller

import (
	"fmt"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bdebug"
	"wm-func/tools/alter-data-v2/backend/cac"
	"wm-func/tools/alter-data-v2/backend/cache"
	"wm-func/tools/alter-data-v2/backend/tags"
)

func GetAlterDataWithPlatformWithTenantId(needRefresh bool, platform string, tenantId int64) AllTenantData {
	var res = AllTenantData{}

	var newTenants, oldTenants = []cac.TenantDateSequence{}, []cac.TenantDateSequence{}
	if tenantId < 0 {
		newTenants, oldTenants = cac.GetAlterDataWithPlatform(platform, needRefresh)
	} else {
		newTenants, oldTenants = cac.GetAlterDataWithPlatformWithTenantId(platform, needRefresh, tenantId)
	}

	defaultTags := tags.GetDefaultTags()

	// 获取数据最后加载时间
	res.DataLastLoadTime = bdao.GetDataLastLoadTime(platform)

	for _, tenant := range newTenants {
		res.NewTenants = append(res.NewTenants, TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
	}

	for _, tenant := range oldTenants {
		res.OldTenants = append(res.OldTenants, TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
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
	var res = AllTenantData{}

	newTenants, oldTenants := cac.GetAlterDataWithPlatform(platform, needRefresh)
	defaultTags := tags.GetDefaultTags()

	// 获取数据最后加载时间
	res.DataLastLoadTime = bdao.GetDataLastLoadTime(platform)

	for _, tenant := range newTenants {
		res.NewTenants = append(res.NewTenants, TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
	}

	for _, tenant := range oldTenants {
		res.OldTenants = append(res.OldTenants, TenantData{
			TenantId:      tenant.TenantId,
			RegisterTime:  tenant.RegisterTime,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
	}

	// 读取缓存：为所有租户加载 RemoveData
	cacheManager := cache.GetCacheManager()

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

	return res

}
