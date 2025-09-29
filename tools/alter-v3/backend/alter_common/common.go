package alter_common

import "wm-func/common/db/platform_db"

type Data struct {
	TenantId int64  `gorm:"tenant_id"`
	RawDate  string `gorm:"raw_date"`
	Data     int64  `gorm:"data"`
}

func GetData(execSql string) []Data {
	db := platform_db.GetDB()
	var res []Data

	if err := db.Raw(execSql).Limit(-1).Scan(&res).Error; err != nil {
		panic(err)
	}

	var result = map[int64]map[string]Data{}

	return res
}

func GetAllTenantWithPlatform(platform string) []int64 {
	exec := query_platform_with_platform
	if platform == "shopify" {
		exec = query_platform_with_shopify
	}

	var res []int64

	db := platform_db.GetDB()
	if err := db.Raw(exec).Scan(res).Error; err != nil {
		panic(err)
	}

	//result := map[int64]bool{}
	//for _, id := range res {
	//	result[id] = true
	//}
	return res
}
