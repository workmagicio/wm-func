package alter_common

import (
	"strings"
	"wm-func/common/db/platform_db"
)

type Data struct {
	TenantId int64  `gorm:"tenant_id"`
	RawDate  string `gorm:"raw_date"`
	Data     int64  `gorm:"data"`
}

type Tenant struct {
	TenantId     int64  `gorm:"tenant_id"`
	RegisterTime string `gorm:"register_time"`
}

func GetData(execSql string) map[int64]map[string]Data {
	db := platform_db.GetDB()
	var res []Data

	if err := db.Raw(execSql).Limit(-1).Scan(&res).Error; err != nil {
		panic(err)
	}

	var result = map[int64]map[string]Data{}

	for i := 0; i < len(res); i++ {
		if _, ok := result[res[i].TenantId]; !ok {
			result[res[i].TenantId] = map[string]Data{}
		}

		result[res[i].TenantId][res[i].RawDate] = res[i]
	}

	return result
}

func GetAllTenantWithPlatform(platform string) []int64 {
	exec := strings.ReplaceAll(query_platform_with_platform, "{{platform}}", platform)
	if platform == "shopify" {
		exec = query_platform_with_shopify
	}

	var res = []Tenant{}

	db := platform_db.GetDB()
	if err := db.Raw(exec).Scan(&res).Error; err != nil {
		panic(err)
	}

	var result = []int64{}
	for i := 0; i < len(res); i++ {
		result = append(result, res[i].TenantId)
	}

	return result
}

func GetLast15DayRegisterTenant() map[int64]string {
	exec := query_new_register_tenant
	var res = []Tenant{}
	db := platform_db.GetDB()
	if err := db.Raw(exec).Scan(&res).Error; err != nil {
		panic(err)
	}
	var result = map[int64]string{}
	for i := 0; i < len(res); i++ {
		result[res[i].TenantId] = res[i].RegisterTime
	}

	return result
}
