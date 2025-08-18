package config

// QueryConfig SQL查询配置
type QueryConfig struct {
	Key         string `json:"key"`         // 查询键
	Name        string `json:"name"`        // 查询名称
	Description string `json:"description"` // 查询描述
	SQL         string `json:"sql"`         // SQL语句
}

// queryConfigs SQL查询配置映射
var queryConfigs = map[string]QueryConfig{
	"google_ads_query": {
		Key:         "google_ads_query",
		Name:        "Google Ads数据查询",
		Description: "查询Google广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'googleAds'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Google'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"meta_ads_query": {
		Key:         "meta_ads_query",
		Name:        "Meta Ads数据查询",
		Description: "查询Meta广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'facebookMarketing'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Meta'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"applovin_ads_query": {
		Key:         "applovin_ads_query",
		Name:        "AppLovin数据查询",
		Description: "查询AppLovin广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'applovin'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Applovin'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"tiktok_ads_query": {
		Key:         "tiktok_ads_query",
		Name:        "TikTok Ads数据查询",
		Description: "查询TikTok广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM in('tiktokMarketing', 'tiktokMarketing_gmv_max')
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'TikTok'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"tenant_cross_platform_query": {
		Key:         "tenant_cross_platform_query",
		Name:        "租户跨平台数据查询",
		Description: "查询指定租户在所有广告平台的数据对比",
		SQL: `
with
    google_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'googleAds'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    google_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Google'
          and TENANT_ID = ?
        group by 1, 2
    ),
    google_merge as (
        select
            google_api.TENANT_ID,
            'Google' as platform,
            google_api.RAW_DATE,
            google_api.spend as api_spend,
            coalesce(google_ads.ad_spend, 0) as ad_spend
        from
            google_api
            left join google_ads on google_api.TENANT_ID = google_ads.TENANT_ID 
            and google_api.RAW_DATE = google_ads.EVENT_DATE
    ),
    
    meta_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'facebookMarketing'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    meta_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Meta'
          and TENANT_ID = ?
        group by 1, 2
    ),
    meta_merge as (
        select
            meta_api.TENANT_ID,
            'Meta' as platform,
            meta_api.RAW_DATE,
            meta_api.spend as api_spend,
            coalesce(meta_ads.ad_spend, 0) as ad_spend
        from
            meta_api
            left join meta_ads on meta_api.TENANT_ID = meta_ads.TENANT_ID 
            and meta_api.RAW_DATE = meta_ads.EVENT_DATE
    ),
    
    applovin_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'applovin'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    applovin_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Applovin'
          and TENANT_ID = ?
        group by 1, 2
    ),
    applovin_merge as (
        select
            applovin_api.TENANT_ID,
            'AppLovin' as platform,
            applovin_api.RAW_DATE,
            applovin_api.spend as api_spend,
            coalesce(applovin_ads.ad_spend, 0) as ad_spend
        from
            applovin_api
            left join applovin_ads on applovin_api.TENANT_ID = applovin_ads.TENANT_ID 
            and applovin_api.RAW_DATE = applovin_ads.EVENT_DATE
    ),
    
    tiktok_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM in('tiktokMarketing', 'tiktokMarketing_gmv_max')
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    tiktok_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'TikTok'
          and TENANT_ID = ?
        group by 1, 2
    ),
    tiktok_merge as (
        select
            tiktok_api.TENANT_ID,
            'TikTok' as platform,
            tiktok_api.RAW_DATE,
            tiktok_api.spend as api_spend,
            coalesce(tiktok_ads.ad_spend, 0) as ad_spend
        from
            tiktok_api
            left join tiktok_ads on tiktok_api.TENANT_ID = tiktok_ads.TENANT_ID 
            and tiktok_api.RAW_DATE = tiktok_ads.EVENT_DATE
    ),
    
    shopify_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(orders), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'shopify'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    shopify_ads as (
        select
            TENANT_ID,
            event_date,
            sum(total_orders) as ad_spend
        from platform_offline.ads_view_overview_sales_and_profit_latest
        where event_date >= utc_date() - interval 90 day
          and sales_platform = 'shopify'
          and TENANT_ID = ?
        group by 1, 2
    ),
    shopify_merge as (
        select
            shopify_api.TENANT_ID,
            'Shopify' as platform,
            shopify_api.RAW_DATE,
            shopify_api.spend as api_spend,
            coalesce(shopify_ads.ad_spend, 0) as ad_spend
        from
            shopify_api
            left join shopify_ads on shopify_api.TENANT_ID = shopify_ads.TENANT_ID 
            and shopify_api.RAW_DATE = shopify_ads.EVENT_DATE
    ),
    
    snapchat_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'snapchatMarketing'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    snapchat_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Snapchat'
          and TENANT_ID = ?
        group by 1, 2
    ),
    snapchat_merge as (
        select
            snapchat_api.TENANT_ID,
            'Snapchat' as platform,
            snapchat_api.RAW_DATE,
            snapchat_api.spend as api_spend,
            coalesce(snapchat_ads.ad_spend, 0) as ad_spend
        from
            snapchat_api
            left join snapchat_ads on snapchat_api.TENANT_ID = snapchat_ads.TENANT_ID 
            and snapchat_api.RAW_DATE = snapchat_ads.EVENT_DATE
    ),
    
    tiktokshop_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(ORDERS), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'tiktokShopPartner'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    tiktokshop_ads as (
        select
            tenant_id,
            event_date,
            sum(total_orders) as ad_spend
        from platform_offline.ads_view_overview_sales_and_profit_latest
        where event_date >= utc_date() - interval 90 day
          and sales_platform = 'tiktok'
          and TENANT_ID = ?
        group by 1, 2
    ),
    tiktokshop_merge as (
        select
            tiktokshop_api.TENANT_ID,
            'TikTok Shop' as platform,
            tiktokshop_api.RAW_DATE,
            tiktokshop_api.spend as api_spend,
            coalesce(tiktokshop_ads.ad_spend, 0) as ad_spend
        from
            tiktokshop_api
            left join tiktokshop_ads on tiktokshop_api.TENANT_ID = tiktokshop_ads.TENANT_ID 
            and tiktokshop_api.RAW_DATE = tiktokshop_ads.EVENT_DATE
    ),
    
    pinterest_api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'pinterest'
          and RAW_DATE > utc_date() - interval 90 day
          and TENANT_ID = ?
        group by 1, 2
    ),
    pinterest_ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Pinterest'
          and TENANT_ID = ?
        group by 1, 2
    ),
    pinterest_merge as (
        select
            pinterest_api.TENANT_ID,
            'Pinterest' as platform,
            pinterest_api.RAW_DATE,
            pinterest_api.spend as api_spend,
            coalesce(pinterest_ads.ad_spend, 0) as ad_spend
        from
            pinterest_api
            left join pinterest_ads on pinterest_api.TENANT_ID = pinterest_ads.TENANT_ID 
            and pinterest_api.RAW_DATE = pinterest_ads.EVENT_DATE
    ),
    
    all_platforms as (
        select * from google_merge
        UNION ALL
        select * from meta_merge  
        UNION ALL
        select * from applovin_merge
        UNION ALL
        select * from tiktok_merge
        UNION ALL
        select * from shopify_merge
        UNION ALL
        select * from snapchat_merge
        UNION ALL
        select * from tiktokshop_merge
        UNION ALL
        select * from pinterest_merge
    ),
    
    result as (
        select
            all_platforms.*
        from
            all_platforms
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on all_platforms.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    platform,
    raw_date,
    api_spend,
    ad_spend
from
    result
order by platform, raw_date`,
	},

	"tenants_list_query": {
		Key:         "tenants_list_query",
		Name:        "租户列表查询",
		Description: "获取系统中所有可用租户的列表",
		SQL: `
select distinct
    TENANT_ID as tenant_id,
    concat('Tenant ', TENANT_ID) as tenant_name
from platform_offline.integration_api_data_view
where RAW_DATE > utc_date() - interval 30 day
  and TENANT_ID in (
    select tenant_id 
    from platform_offline.dwd_view_analytics_non_testing_tenants
  )
order by TENANT_ID
limit 1000`,
	},

	"recent_registered_tenants_query": {
		Key:         "recent_registered_tenants_query",
		Name:        "最近注册租户查询",
		Description: "获取最近15天注册的租户列表",
		SQL: `
select
    tenant_id,
    concat('Tenant ', tenant_id) as tenant_name,
    register_time
from
    platform_offline.dwd_view_analytics_non_testing_tenants
where register_time > utc_date() - interval 15 day
order by register_time desc
limit 50`,
	},

	"shopify_query": {
		Key:         "shopify_query",
		Name:        "Shopify订单数据查询",
		Description: "查询Shopify电商平台的API订单数据和概览订单数据对比",
		SQL: `
with
    api as (
        select
            tenant_id,
            RAW_DATE,
            round(sum(orders), 0) as api_data
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'shopify'
          and RAW_DATE > utc_date() - interval 70 day
        group by 1, 2
    ),
    ads as (
        select
            tenant_id,
            event_date,
            sum(total_orders) as overview_data
        from platform_offline.ads_view_overview_sales_and_profit_latest
        where ((event_date >= utc_date() - interval 70 day) and
               sales_platform = 'shopify')
        group by tenant_id, event_date
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.api_data as api_spend,
            coalesce(ads.overview_data, 0) as ad_spend
        from
            api
            left join ads
                      on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
                 on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"snapchat_ads_query": {
		Key:         "snapchat_ads_query",
		Name:        "Snapchat Ads数据查询",
		Description: "查询Snapchat广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'snapchatMarketing'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Snapchat'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"tiktokshop_query": {
		Key:         "tiktokshop_query",
		Name:        "TikTok Shop订单数据查询",
		Description: "查询TikTok Shop电商平台的API订单数据和概览订单数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(ORDERS), 0) as api_data
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'tiktokShopPartner'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            tenant_id,
            event_date,
            sum(total_orders) as overview_data
        from platform_offline.ads_view_overview_sales_and_profit_latest
        where event_date >= utc_date() - interval 90 day
          and sales_platform = 'tiktok'
        group by tenant_id, event_date
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.api_data as api_spend,
            coalesce(ads.overview_data, 0) as ad_spend
        from
            api
            left join ads
                      on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
                 on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},

	"pinterest_ads_query": {
		Key:         "pinterest_ads_query",
		Name:        "Pinterest Ads数据查询",
		Description: "查询Pinterest广告平台的API数据和广告数据对比",
		SQL: `
with
    api as (
        select
            TENANT_ID,
            RAW_DATE,
            round(sum(AD_SPEND), 0) as spend
        from
            platform_offline.integration_api_data_view
        where RAW_PLATFORM = 'pinterest'
          and RAW_DATE > utc_date() - interval 90 day
        group by 1, 2
    ),
    ads as (
        select
            TENANT_ID,
            event_date,
            round(sum(ad_spend), 0) as ad_spend
        from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
        where event_date > utc_date() - interval 90 day
          and json_overlaps(attr_model_array, json_array(0, 3))
          and attr_enhanced in (1, 4)
          and ADS_PLATFORM = 'Pinterest'
        group by 1, 2
    ),
    merge as (
        select
            api.TENANT_ID,
            api.RAW_DATE,
            api.spend as api_spend,
            coalesce(ads.ad_spend, 0) as ad_spend
        from
            api
            left join ads on api.TENANT_ID = ads.TENANT_ID and api.RAW_DATE = ads.EVENT_DATE
    ),
    result as (
        select
            merge.*
        from
            merge
            join platform_offline.dwd_view_analytics_non_testing_tenants as b
            on merge.TENANT_ID = b.tenant_id
    )
select
    tenant_id,
    raw_date,
    api_spend,
    ad_spend
from
    result`,
	},
}

// GetQueryConfig 根据键获取查询配置
func GetQueryConfig(key string) (QueryConfig, bool) {
	config, exists := queryConfigs[key]
	return config, exists
}

// GetQuerySQL 根据键获取SQL语句
func GetQuerySQL(key string) (string, bool) {
	config, exists := queryConfigs[key]
	if !exists {
		return "", false
	}
	return config.SQL, true
}

// GetAllQueryConfigs 获取所有查询配置
func GetAllQueryConfigs() map[string]QueryConfig {
	return queryConfigs
}
