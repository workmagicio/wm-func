package bmodel

import (
	"fmt"
	"log"
)

var query_api_data = `
select
    tenant_id,
    cast(raw_date as varchar) as raw_date,
    cast(sum(if(RAW_PLATFORM in ('shopify'), ORDERS, AD_SPEND) ) as bigint) as data
from
    platform_offline.integration_api_data_view
where RAW_PLATFORM = 'shopify'
  and RAW_DATE > utc_date() - interval 365 day
group by 1, 2`

var query_shopify_data = `
select
    tenant_id,
    cast(date(order_create_time) as varchar) as raw_date,
    count(order_id) as data
from
    platform_offline.dwd_view_portrait_order_dim_shopify_latest
where order_create_time > utc_date() - interval 365 day
group by 1, 2`

// Shopify 数据比对查询，比较 API 数据源和原始订单数据
var query_shopify_comparison = `
SELECT 
    tenant_id,
    cast(date(order_create_time) as varchar) as raw_date,
    count(order_id) as data
FROM platform_offline.dwd_view_portrait_order_dim_shopify_latest
WHERE order_create_time > utc_date() - interval 365 day
GROUP BY 1, 2
ORDER BY tenant_id, raw_date`

// 获取 API 数据源的 Shopify 数据
func GetShopifyApiData() []WmData {
	log.Printf("🔍 [GetShopifyApiData] 开始查询 API 数据源")
	result := QueryWmData(query_api_data)
	log.Printf("📊 [GetShopifyApiData] 查询完成，获取到 %d 条记录", len(result))
	return result
}

// 获取原始数据源的 Shopify 订单数据
func GetShopifyOrderData() []WmData {
	log.Printf("🔍 [GetShopifyOrderData] 开始查询订单数据源")
	result := QueryWmData(query_shopify_data)
	log.Printf("📊 [GetShopifyOrderData] 查询完成，获取到 %d 条记录", len(result))
	return result
}

// 获取 Shopify 比对数据（使用原始订单数据作为基准）
func GetShopifyComparisonData() []WmData {
	log.Printf("🔍 [GetShopifyComparisonData] 开始查询比对数据")
	fmt.Printf("📝 [GetShopifyComparisonData] 执行SQL:\n%s\n", query_shopify_comparison)
	result := QueryWmData(query_shopify_comparison)
	log.Printf("📊 [GetShopifyComparisonData] 查询完成，获取到 %d 条记录", len(result))
	if len(result) > 0 {
		log.Printf("🎯 [GetShopifyComparisonData] 示例数据: TenantId=%d, Date=%s, Data=%d",
			result[0].TenantId, result[0].RawDate, result[0].Data)
	}
	return result
}
