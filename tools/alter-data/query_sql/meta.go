package query_sql

var Query_applovin_api_with_overview = `
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
    )
   , ads as (
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
)
   ,merge as (
    select
        api.TENANT_ID,
        api.RAW_DATE,
        api.spend as api_spend,
        coalesce(ads.ad_spend, 0) as ad_spend
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
    result
`

var Query_google_api_with_overview = `
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
    )
   , ads as (
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
)
   ,merge as (
    select
        api.TENANT_ID,
        api.RAW_DATE,
        api.spend as api_spend,
        coalesce(ads.ad_spend, 0) as ad_spend
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
    result
`
var Query_meta_api_with_overview = `
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
    )
   , ads as (
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
)
   ,merge as (
    select
        api.TENANT_ID,
        api.RAW_DATE,
        api.spend as api_spend,
        coalesce(ads.ad_spend, 0) as ad_spend
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
    result
`
