package bmodel

import (
	"fmt"
	"log"
	"wm-func/common/db/platform_db"
)

var query_knocommerce = `
select
  tenant_id,
  cast(date(response_provided_at) as varchar) as event_date,
  count(order_id) as value
  from platform_offline.dwd_view_post_survey_response_v20250910jz
where platform = 'knocommerce'
and response_provided_at > utc_timestamp() - interval 90 day
group by 1, 2;
`

var query_knocommerce_with_tenant_id = `
select
  tenant_id,
  cast(date(response_provided_at) as varchar) as event_date,
  count(order_id) as value
  from platform_offline.dwd_view_post_survey_response_v20250910jz
where platform = 'knocommerce'
  and tenant_id = %d
and response_provided_at > utc_timestamp() - interval 90 day
group by 1, 2;
`

// GetKnocommerceOverviewData 获取knocommerce的WM数据
func GetKnocommerceOverviewData() []OverViewData {
	db := platform_db.GetDB()
	var res = []OverViewData{}
	if err := db.Raw(query_knocommerce).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}
	return res
}

// GetKnocommerceOverviewDataWithTenantId 获取特定租户的knocommerce WM数据
func GetKnocommerceOverviewDataWithTenantId(tenantId int64) []OverViewData {
	db := platform_db.GetDB()
	var res = []OverViewData{}
	exec := fmt.Sprintf(query_knocommerce_with_tenant_id, tenantId)
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}
	return res
}
