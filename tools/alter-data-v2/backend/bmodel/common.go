package bmodel

import (
	"fmt"
	"log"
	"time"
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data-v2/backend"
)

var data_view_query = `
select
    tenant_id,
    cast(raw_date as varchar) as raw_date,
    cast(sum(if(RAW_PLATFORM in ('knocommerce'), ORDERS, AD_SPEND) ) as bigint) as ad_spend
--     cast(sum(ORDERS) as bigint) as orders
from
    platform_offline.integration_api_data_view
where RAW_PLATFORM = '%s'
and RAW_DATE > utc_date() - interval 90 day
group by 1, 2
`

type ApiData struct {
	TenantId int64  `json:"tenant_id"`
	RawDate  string `json:"raw_date"`
	AdSpend  int64  `json:"ad_spend"`
	Orders   int64  `json:"orders"`
}

func GetDataWithPlatform(platform string) []ApiData {
	db := platform_db.GetDB()
	var res = []ApiData{}
	exec := fmt.Sprintf(data_view_query, platform)
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}

var data_view_query_with_tenant_id = `
select
    tenant_id,
    cast(raw_date as varchar) as raw_date,
    cast(sum(if(RAW_PLATFORM in ('knocommerce'), ORDERS, AD_SPEND) ) as bigint) as ad_spend
    cast(sum(ORDERS) as bigint) as orders
from
    platform_offline.integration_api_data_view
where RAW_PLATFORM = '%s'
  and tenant_id = %d
and RAW_DATE > utc_date() - interval 90 day
group by 1, 2
`

func GetDataWithPlatformAndTenantId(platform string, tenantId int64) []ApiData {
	db := platform_db.GetDB()
	var res = []ApiData{}
	exec := fmt.Sprintf(data_view_query_with_tenant_id, platform, tenantId)
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}

var query_overview_data = `
select
    tenant_id,
    cast(event_date as varchar) as event_date,
    cast(sum(ad_spend) as bigint) as value
from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
where event_date > utc_date() - interval 90 day
  and json_overlaps(attr_model_array, json_array(0, 3))
  and attr_enhanced in (1, 4)
  and ADS_PLATFORM = '%s'
group by 1, 2
`

type OverViewData struct {
	TenantId  int64  `json:"tenant_id"`
	EventDate string `json:"event_date"`
	Value     int64  `json:"value"`
}

func GetOverviewDataWithPlatform(platform string) []OverViewData {
	// knocommerce平台使用特殊的查询
	if platform == backend.ADS_PLATFORM_KNOCOMMERCE {
		return GetKnocommerceOverviewData()
	}

	db := platform_db.GetDB()
	var res = []OverViewData{}
	exec := fmt.Sprintf(query_overview_data, platform)
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}

var query_overview_data_with_tenant_id = `
select
    tenant_id,
    cast(event_date as varchar) as event_date,
    cast(sum(ad_spend) as bigint) as value
from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
where event_date > utc_date() - interval 90 day
  and json_overlaps(attr_model_array, json_array(0, 3))
  and attr_enhanced in (1, 4)
  and ADS_PLATFORM = '%s'
  and tenant_id = %d
group by 1, 2
`

func GetOverviewDataWithPlatformAndTenantId(platform string, tenantId int64) []OverViewData {
	// knocommerce平台使用特殊的查询
	if platform == backend.ADS_PLATFORM_KNOCOMMERCE {
		return GetKnocommerceOverviewDataWithTenantId(tenantId)
	}

	db := platform_db.GetDB()
	var res = []OverViewData{}
	exec := fmt.Sprintf(query_overview_data_with_tenant_id, platform, tenantId)
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}

var query_all_tenant = `
select
    tenant_id,
    main_client_name,
    register_time
from
    platform_offline.dwd_view_analytics_non_testing_tenants
`

type AllTenant struct {
	TenantId       int64     `json:"tenant_id"`
	MainClientName string    `json:"name"`
	RegisterTime   time.Time `json:"register_time"`
}

func GetAllTenant() []AllTenant {
	db := platform_db.GetDB()
	var res = []AllTenant{}
	if err := db.Raw(query_all_tenant).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}

var query_tenant_platform = `
select tenant_id,
       platform
from platform_offline.account_connection_unnest_account_level_with_no_testing
group by 1, 2
`

type TenantPlatform struct {
	TenantId int64  `json:"tenant_id"`
	Platform string `json:"platform"`
}

func GetTenantPlatformMap() map[int64]map[string]bool {
	db := platform_db.GetDB()
	var res = []TenantPlatform{}
	if err := db.Raw(query_tenant_platform).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	result := make(map[int64]map[string]bool)
	for _, item := range res {
		if result[item.TenantId] == nil {
			result[item.TenantId] = make(map[string]bool)
		}
		result[item.TenantId][item.Platform] = true
	}

	return result
}

func GetSingleDataWithPlatform(platform string) []WmData {
	var exec = ""
	switch platform {
	case backend.PLATFORN_SNAPCHAT_AMAZONVENDOR:
		exec = query_amazon_vonder
	case backend.PLATFORN_SNAPCHAT_FAIRING:
		exec = fairing_query
	case backend.PLATFORN_AMAZONADS:
		exec = amazonads_query
	case "applovinLog":
		exec = query_applovin_log
	case backend.PLATFORN_SHOPIFY:
		// 对于 Shopify，返回比对数据（以订单数据为基准）
		return GetShopifyComparisonData()
	}

	return QueryWmData(exec)
}

func QueryWmData(exec string) []WmData {
	db := platform_db.GetDB()
	var res = []WmData{}
	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	return res
}
