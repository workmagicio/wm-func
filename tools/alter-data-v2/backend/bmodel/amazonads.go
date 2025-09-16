package bmodel

var amazonads_query = `select
    tenant_id,
    cast(event_date as varchar) as raw_date,
    cast(sum(ad_spend) as bigint) as data
from platform_offline.dws_view_amazon_ads_country_level_metrics_latest
where EVENT_DATE > utc_date() - interval 90 day
group by 1, 2`
