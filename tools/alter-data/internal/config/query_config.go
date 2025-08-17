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
