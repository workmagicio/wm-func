package bmodel

import (
	"fmt"
	"log"
	"wm-func/common/db/platform_db"
)

// Shopify API 数据源查询（从 integration_api_data_view 获取，只使用 ORDERS 字段）
var shopify_api_query = `
select
    tenant_id,
    cast(raw_date as varchar) as raw_date,
    cast(sum(ORDERS) as bigint) as ad_spend
from
    platform_offline.integration_api_data_view
where RAW_PLATFORM = 'shopify'
  and RAW_DATE > utc_date() - interval 90 day
group by 1, 2`

// Shopify 原始数据源查询（从订单表获取，用于对比）
var shopify_overview_query = `
select
    tenant_id,
    cast(date(order_create_time) as varchar) as event_date,
    count(order_id) as value
from
    platform_offline.dwd_view_portrait_order_dim_shopify_latest
where order_create_time > utc_date() - interval 90 day
group by 1, 2`

// 获取 Shopify Overview 数据
func GetShopifyOverviewData() []OverViewData {
	log.Printf("🔍 [GetShopifyOverviewData] 开始查询 Shopify Overview 数据")
	fmt.Printf("📝 [GetShopifyOverviewData] 执行查询:\n%s\n", shopify_overview_query)
	result := QueryOverviewData(shopify_overview_query)
	log.Printf("📊 [GetShopifyOverviewData] 查询完成，获取到 %d 条记录", len(result))
	if len(result) > 0 {
		log.Printf("🎯 [GetShopifyOverviewData] 示例数据: TenantId=%d, EventDate=%s, Value=%d",
			result[0].TenantId, result[0].EventDate, result[0].Value)
	}
	return result
}

// 获取指定租户的 Shopify Overview 数据
func GetShopifyOverviewDataWithTenantId(tenantId int64) []OverViewData {
	query := shopify_overview_query + fmt.Sprintf(" and tenant_id = %d", tenantId)
	log.Printf("🔍 [GetShopifyOverviewDataWithTenantId] 开始查询租户 %d 的 Overview 数据", tenantId)
	fmt.Printf("📝 [GetShopifyOverviewDataWithTenantId] 执行查询:\n%s\n", query)
	result := QueryOverviewData(query)
	log.Printf("📊 [GetShopifyOverviewDataWithTenantId] 查询完成，获取到 %d 条记录", len(result))
	return result
}

// 通用查询 Overview 数据的函数
func QueryOverviewData(query string) []OverViewData {
	db := platform_db.GetDB()
	var res = []OverViewData{}
	if err := db.Raw(query).Limit(-1).Scan(&res).Error; err != nil {
		log.Printf("❌ [QueryOverviewData] 查询失败: %v", err)
		log.Printf("📝 [QueryOverviewData] 失败的查询: %s", query)
		panic(err)
	}
	return res
}
