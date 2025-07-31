select event_date,
       event_hour,
       sum(ad_spend)    as `ad_spend`,
       sum(impressions) as `impressions`,
       sum(clicks)      as `clicks`,
       sum(ads_orders)  as `ads_orders`,
       sum(ads_sales)   as `ads_sales`
from ads_view_analytics_ads_account_level_metrics_hourly_v20250219
where (event_date = '2025-07-24' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0;

select event_date,
       event_hour,
       sum(page_views)                    as `page_views`,
       sum(product_collection_page_views) as `product_collection_page_views`,
       sum(product_added_to_carts)        as `product_added_to_carts`,
       sum(checkout_starts)               as `checkout_starts`,
       sum(checkout_completes)            as `checkout_completes`
from ads_view_event_metrics_hourly_v20250723
where (event_date = '2025-07-24' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0;

select event_date,
       event_hour,
       sum(orders)                        as `total_orders`,
       (sum(orders) - sum(refund_orders)) as `net_orders`,
       sum(sales)                         as `total_sales`,
       sum(sales) / sum(orders)           as `aov`
from ads_view_store_sales_hourly_v20250514v2
where (event_date = '2025-07-24' and sales_platform = 'shopify' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0;

select event_date,
       event_hour,
       sum(orders)   as `orders`,
       sum(sales)    as `total_sales`,
       sum(orders)   as `total_orders`,
       sum(ad_spend) as `ad_spend`
from ads_view_store_sales_hourly_v20250514v2
where (event_date = '2025-07-24' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0;

select event_date,
       event_hour,
       sum(ad_spend)   as `ad_spend`,
       sum(ads_orders) as `ads_orders`,
       sum(ads_sales)  as `ads_sales`
from ads_view_analytics_ads_account_level_metrics_hourly_v20250219
where (event_date = '2025-07-24' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0;

select event_date,
       event_hour,
       sum(orders)                  as `total_orders`,
       sum(sales)                   as `total_sales`,
       sum(sales) / sum(orders)     as `aov`,
       sum(subscription_orders)     as `subscription_orders`,
       sum(non_subscription_orders) as `non_subscription_orders`
from ads_view_store_sales_hourly_v20250514v2
where (event_date = '2025-07-24' and sales_platform = 'shopify' and tenant_id in (134301))
group by event_date, event_hour
having true limit 50000
offset 0