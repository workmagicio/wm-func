select event_date,
       event_hour,
       sum(ad_spend)    as `ad_spend`,
       sum(impressions) as `impressions`,
       sum(clicks)      as `clicks`,
       sum(ads_orders)  as `ads_orders`,
       sum(ads_sales)   as `ads_sales`
from ads_view_analytics_ads_account_level_metrics_hourly_latest
where (event_date = '2025-07-24' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0