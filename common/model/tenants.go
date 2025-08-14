package model

import (
	"wm-func/common/db/platform_db"
)

func GetAllTenantsId() []int64 {
	result := make([]int64, 0)
	db := platform_db.GetDB()
	if err := db.Raw(query_all_tenants_id).Scan(&result).Limit(-1).Error; err != nil {
		panic(err)
	}
	return result
}

type Tenants struct {
	TenantId int64  `gorm:"column:tenant_id"`
	Platform string `gorm:"column:platform"`
}

func GetAllTenantsIdWithPlatform() []Tenants {
	result := make([]Tenants, 0)
	db := platform_db.GetDB()
	if err := db.Raw(query_all_tenants_id_with_platform).Scan(&result).Limit(-1).Error; err != nil {
		panic(err)
	}
	return result
}
