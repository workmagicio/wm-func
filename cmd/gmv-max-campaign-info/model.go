package main

import (
	"fmt"
	"wm-func/common/db/airbyte_db"
)

func getCampaignIdByTenantId(tenantId int64) []string {
	db := airbyte_db.GetDB()
	res := []string{}

	if err := db.Raw(fmt.Sprintf(queryCampaign, tenantId)).Scan(&res).Error; err != nil {
		panic(err)
	}

	return res
}

var queryCampaign = `
select
    cast(json_extract(_airbyte_data, '$.campaign_id') as varchar) as campaign_id
from
    airbyte_destination_v2.raw_tiktok_marketing_gmv_max_campaigns
where wm_tenant_id = %d`

type Campaign struct {
	TenantId int64  `json:"wm_tenant_id"`
	Campaign string `json:"campaign"`
}

type AribyteDate struct {
	TenantId            int64  `gorm:"column:wm_tenant_id"`
	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
	AirbyteData         []byte `gorm:"column:_airbyte_data"`
	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
	AirbyteGenerationId int64  `gorm:"column:_airbyte_generation_id"`
	ItemType            string `gorm:"-"` // 不映射到数据库，用于确定表名
}

func (AribyteDate) TableName() string {
	return "airbyte_destination_v2.raw_tiktok_marketing_gmv_max_campaign_info"
}
