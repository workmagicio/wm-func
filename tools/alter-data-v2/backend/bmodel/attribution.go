package bmodel

import (
	"log"
	"time"
	"wm-func/common/db/platform_db"
)

var attr_query = `
select tenant_id,
       cast(event_date as varchar)     as raw_date,
       ads_platform,
       sum(attr_orders + extra_orders) as data
from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
where (event_date >= utc_date() - interval 60 day)
  and json_overlaps(attr_model_array, json_array(0, 3))
  and attr_enhanced in (1, 4)
group by 1, 2, 3
`

type Attribution struct {
	TenantId    int64  `json:"tenant_id"`
	RawDate     string `json:"raw_date"`
	AdsPlatform string `json:"ads_platform"`
	Data        int64  `json:"data"`
}

func GetAttrData() []Attribution {
	db := platform_db.GetDB()
	var res = []Attribution{}

	log.Printf("🔍 [GetAttrData] 开始执行归因数据查询")
	startTime := time.Now()

	if err := db.Raw(attr_query).Limit(-1).Scan(&res).Error; err != nil {
		log.Printf("❌ [GetAttrData] 归因数据查询失败: %v", err)
		panic(err)
	}

	duration := time.Since(startTime)
	log.Printf("✅ [GetAttrData] 归因数据查询成功，获取到 %d 条记录，耗时: %v", len(res), duration)
	return res
}
