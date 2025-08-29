package controller

import (
	"wm-func/tools/alter-data-v2/backend/cac"
	"wm-func/tools/alter-data-v2/backend/tags"
)

func GetAlterDataWithPlatform(needRefresh bool, platform string) AllTenantData {
	var res = AllTenantData{}

	newTenants, oldTenants := cac.GetAlterDataWithPlatform(platform, needRefresh)
	defaultTags := tags.GetDefaultTags()

	for _, tenant := range newTenants {
		res.NewTenants = append(res.NewTenants, TenantData{
			TenantId:      tenant.TenantId,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
	}

	for _, tenant := range oldTenants {
		res.OldTenants = append(res.OldTenants, TenantData{
			TenantId:      tenant.TenantId,
			Last30DayDiff: tenant.Last30DayDiff,
			DateSequence:  tenant.DateSequence,
			Tags:          []string{defaultTags[tenant.TenantId]},
		})
	}

	return res

}
