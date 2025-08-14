package main

var airbyte_raw_tables = map[string][]string{
	"raw_tiktok_marketing_gmv_max_metrics": {
		"date_format(cast(json_extract(_airbyte_data, '$.stat_time_day') as varchar), '%Y-%m-%d') as raw_date",
	},
	"raw_applovin_ads_v2_ads_metrics": {
		"date_format(cast(json_extract (_airbyte_data, '$.day') as varchar), '%Y-%m-%d') as raw_date",
	},
	"raw_snapchat_marketing_ads_stats_daily": {
		"date_format(STR_TO_DATE (substring(json_unquote (_airbyte_data ->> '$.start_time'), 1, 19), '%Y-%m-%dT%H:%i:%s'), '%Y-%m-%d') as raw_date",
	},

	"raw_pinterest_ad_analytics": {
		"date_format(json_unquote (_airbyte_data ->> '$.DATE'), '%Y-%m-%d') as raw_date",
	},
	"raw_pinterest_ad_group_analytics": {
		"date_format(json_unquote (_airbyte_data ->> '$.DATE'), '%Y-%m-%d') as raw_date",
	},
	"raw_tiktok_marketing_ads_reports_daily": {
		"date_format(json_unquote (_airbyte_data ->> '$.stat_time_day'), '%Y-%m-%d') as raw_date",
	},
	"raw_facebook_marketing_ads_insights": {
		"date_format(json_unquote (_airbyte_data ->> '$.stat_time_day'), '%Y-%m-%d') as raw_date",
	},
	"raw_google_ads_ad_metrics": {
		"date_format(json_unquote (`_airbyte_data` ->> '$.\"segments.date\"'), '%Y-%m-%d') as raw_date",
	},
	"raw_bing_ads_ad_performance_report_daily": {
		"date_format(json_unquote (`_airbyte_data` ->> '$.\"TimePeriod\"'), '%Y-%m-%d') as raw_date",
	},
	"raw_bing_ads_campaign_performance_report_daily": {
		"date_format(json_unquote (`_airbyte_data` ->> '$.\"TimePeriod\"'), '%Y-%m-%d') as raw_date",
	},
	"raw_google_ads_campaign_metrics": {
		"date_format(json_unquote (`_airbyte_data` ->> '$.\"segments.date\"'), '%Y-%m-%d') as raw_date",
	},
}

var over_view_sqls map[string]string = map[string]string{
	"amazon_ads_check": `
select a.tenant_id,
       date(event_date) as event_date,
       sum(ad_spend) as ad_spend
from dws_view_amazon_ads_country_level_metrics_v2024092301 as a
         join platform_offline.dwd_view_analytics_non_testing_tenants as b
              on a.tenant_id = b.tenant_id
where ad_spend > 0
  and a.tenant_id in (150082,150022,134484,150043,150141,150189,131532,150186,150133,130915,150023,150075,150151,150155,150138)
group by 1, 2
`,
}
