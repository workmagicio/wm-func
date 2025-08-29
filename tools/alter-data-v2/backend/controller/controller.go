package controller

import (
	"fmt"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bdebug"
	"wm-func/tools/alter-data-v2/backend/cac"
	"wm-func/tools/alter-data-v2/backend/tags"
)

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

	// 获取所有 diff > 0 的客户

	var diffTenants = []int64{}
	for _, tag := range append(res.OldTenants, res.NewTenants...) {
		if tag.Last30DayDiff < -100 {
			diffTenants = append(diffTenants, tag.TenantId)
		}
	}

	for _, tenantId := range diffTenants {
		tmp := bdebug.GetDataWithPlatform(tenantId, platform)
		fmt.Println(tmp)
	}

	return res

}
