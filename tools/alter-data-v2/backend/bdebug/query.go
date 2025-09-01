package bdebug

import (
	"fmt"
	"log"
	"strings"
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data-v2/backend"
)

/*+resource_group=job_production*/
var query_data = `
with
  tiktok_gmv_max_metrics as (
    select
      wm_tenant_id                                                                            as tenant_id,
      cast(json_extract(_airbyte_data, '$.campaign_id') as varchar)                           as campaign_id,
      cast(json_extract(_airbyte_data, '$.advertiser_id') as varchar)                         as advertiser_id,
      date(cast(json_extract(_airbyte_data, '$.stat_time_day') as varchar))                           as stat_date,
      cast(json_extract(_airbyte_data, '$.metrics.cost') as double)                           as ad_spend,
      cast(json_extract(_airbyte_data, '$.metrics.gross_revenue') as double)                  as gross_revenue,
      cast(json_extract(_airbyte_data, '$.metrics.orders') as bigint)                         as orders
    from airbyte_destination_v2.raw_tiktok_marketing_gmv_max_metrics
  ),
  applovin_metrics as (
    select
      wm_tenant_id as tenant_id,
      date(cast(json_extract (_airbyte_data, '$.day') as varchar)) as stat_date,
      cast(json_extract (_airbyte_data, '$.account_id') as varchar) as account_id,
      cast(json_extract (_airbyte_data, '$.campaign_id_external') as varchar) as campaign_id,
      '' as adset_id,
      cast(json_extract (_airbyte_data, '$.ad_id') as varchar) as ad_id,
      cast(json_extract (_airbyte_data, '$.cost') as double) as ad_spend,
      cast(json_extract (_airbyte_data, '$.impressions') as bigint) as impressions,
      cast(json_extract (_airbyte_data, '$.clicks') as bigint) as clicks,
      cast(json_extract (_airbyte_data, '$.chka_7d') as bigint) as ads_orders,
      cast(json_extract (_airbyte_data, '$.chka_usd_7d') as double) as ads_sales
    from
      airbyte_destination_v2.raw_applovin_ads_ads_metrics as api
  ),
  applovin_metrics_v2 as (
    select
      wm_tenant_id as tenant_id,
      date(cast(json_extract (_airbyte_data, '$.day') as varchar)) as stat_date,
      cast(json_extract (_airbyte_data, '$.account_id') as varchar) as account_id,
      cast(json_extract (_airbyte_data, '$.campaign_id') as varchar) as campaign_id,
      '' as adset_id,
      cast(json_extract (_airbyte_data, '$.creative_set_id') as varchar) as ad_id,
      cast(json_extract (_airbyte_data, '$.cost') as double) as ad_spend,
      cast(json_extract (_airbyte_data, '$.impressions') as bigint) as impressions,
      cast(json_extract (_airbyte_data, '$.clicks') as bigint) as clicks,
      cast(json_extract (_airbyte_data, '$.chka_7d') as bigint) as ads_orders,
      cast(json_extract (_airbyte_data, '$.chka_usd_7d') as double) as ads_sales
    from
      airbyte_destination_v2.raw_applovin_ads_v2_ads_metrics as api
  ),
  applovin_metrics_merge as (
    select
      *
    from
      applovin_metrics
    where stat_date < '2025-06-20'
    union all
    select
      *
    from
      applovin_metrics_v2
    where stat_date >= '2025-06-20'
  ),
  snapchat_metrics as (
    select
      cast(daily.wm_tenant_id as bigint) as tenant_id,
      json_unquote (daily._airbyte_data ->> '$.id') as ad_id,
      cast(STR_TO_DATE (substring(json_unquote (daily._airbyte_data ->> '$.start_time'), 1, 19), '%Y-%m-%dT%H:%i:%s') as date) as stat_date,
      cast(json_unquote (daily._airbyte_data ->> '$.spend') / 1000000 as double) as ad_spend,
      cast(json_unquote (daily._airbyte_data ->> '$.impressions') as bigint) as impressions,
      cast(json_unquote (daily._airbyte_data ->> '$.swipes') as bigint) as clicks,
      cast(json_unquote (daily._airbyte_data ->> '$.video_views') as bigint) as video_views,
      cast(json_unquote (daily._airbyte_data ->> '$.quartile_1') as bigint) as video_views_p25,
      cast(json_unquote (daily._airbyte_data ->> '$.quartile_2') as bigint) as video_views_p50,
      cast(json_unquote (daily._airbyte_data ->> '$.quartile_3') as bigint) as video_views_p75,
      cast(json_unquote (daily._airbyte_data ->> '$.view_completion') as bigint) as video_views_p100,
      CAST(json_unquote(daily._airbyte_data ->> '$.conversion_purchases') as bigint) AS ads_orders,
#       coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_web_swipe_up') as bigint), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_app_swipe_up') as bigint), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_offline_swipe_up') as bigint), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_web_view') as bigint), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_app_view') as bigint), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_offline_view') as bigint), 0) as ads_orders,
      CAST(json_unquote(daily._airbyte_data ->> '$.conversion_purchases_value') / 1000000 as double) AS ads_sales,
#       coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_web_swipe_up') / 1000000 as double), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_app_swipe_up') / 1000000 as double), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_offline_swipe_up') / 1000000 as double), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_web_view') / 1000000 as double), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_app_view') / 1000000 as double), 0) + coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_offline_view') / 1000000 as double), 0) as ads_sales,
      (
        case JSON_UNQUOTE (adset_settings ->> '$.optimization_goal')
          when 'PIXEL_PAGE_VIEW' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_page_views')
          when 'PIXEL_ADD_TO_CART' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_add_cart')
          when 'PIXEL_PURCHASE' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_purchases')
          end
        ) as conversion,
      (
        case JSON_UNQUOTE (adset_settings ->> '$.optimization_goal')
          when 'PIXEL_PAGE_VIEW' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_page_views_value')
          when 'PIXEL_ADD_TO_CART' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_add_cart_value')
          when 'PIXEL_PURCHASE' then JSON_UNQUOTE (daily._airbyte_data ->> '$.conversion_purchases_value')
          end
        ) as conversion_value,
      JSON_OBJECT (
        'orders_details',
        JSON_OBJECT (
          'website',
          JSON_OBJECT (
            'cta',
            coalesce(cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_web_swipe_up') as bigint), 0),
            'vta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_web_view') as bigint)
          ),
          'app',
          JSON_OBJECT ('cta', cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_app_swipe_up') as bigint), 'vta', cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_app_view') as bigint)),
          'offline',
          JSON_OBJECT (
            'cta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_offline_swipe_up') as bigint),
            'vta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_offline_view') as bigint)
          )
        ),
        'sales_details',
        JSON_OBJECT (
          'website',
          JSON_OBJECT (
            'cta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_web_swipe_up') / 1000000 as double),
            'vta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_web_view') / 1000000 as double)
          ),
          'app',
          JSON_OBJECT (
            'cta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_app_swipe_up') / 1000000 as double),
            'vta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_app_view') / 1000000 as double)
          ),
          'offline',
          JSON_OBJECT (
            'cta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_offline_swipe_up') / 1000000 as double),
            'vta',
            cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_offline_view') / 1000000 as double)
          )
        ),
        'vta_orders',
        (cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_web_view') as bigint) + cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_app_view') as bigint)) + cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_offline_view') as bigint),
        'vta_sales',
        (
          cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_web_view') / 1000000 as double) + cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_app_view') / 1000000 as double) + cast(json_unquote (daily._airbyte_data ->> '$.conversion_purchases_value_offline_view') / 1000000 as double)
          )
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_snapchat_marketing_ads_stats_daily as daily
        join platform_offline.dwd_view_analytics_ads_all_level_settings_v20241225 ad on daily.wm_tenant_id = ad.tenant_id
        and json_extract (daily._airbyte_data, '$.id') = ad.ad_id
        and ad.type = 'ad'
        and ad.ads_platform = 'Snapchat'
  ),
  pinterest_ad_metrics as (
    select
      cast(r.wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (r._airbyte_data ->> '$.DATE') as date) as stat_date,
      cast(json_unquote (r._airbyte_data ->> '$.AD_ID') as bigint) as ad_id,
      cast(json_unquote (r._airbyte_data ->> '$.SPEND_IN_DOLLAR') as double) as ad_spend,
      cast(json_unquote (r._airbyte_data ->> '$.IMPRESSION_1') as bigint) as impressions,
      cast(json_unquote (r._airbyte_data ->> '$.CLICKTHROUGH_1') as bigint) as clicks,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CHECKOUT') as bigint) as ads_orders,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CHECKOUT_VALUE_IN_MICRO_DOLLAR') / 1000000 as bigint) as ads_sales,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_3SEC_VIEWS') as bigint) as video_views,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P25_COMBINED') as bigint) as video_views_p25,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P50_COMBINED') as bigint) as video_views_p50,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P75_COMBINED') as bigint) as video_views_p75,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P100_COMPLETE') as bigint) as video_views_p100,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CONVERSIONS') as bigint) as conversion,
      0 as conversion_value,
      JSON_OBJECT (
        'orders_details',
        JSON_OBJECT (
          'vta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT') as bigint), 0),
          'cta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CLICK_CHECKOUT') as bigint), 0),
          'engagement',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_ENGAGEMENT_CHECKOUT') as bigint), 0)
        ),
        'sales_details',
        JSON_OBJECT (
          'vta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0),
          'cta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0),
          'engagement',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_ENGAGEMENT_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0)
        ),
        'vta_orders',
        coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT') as bigint), 0),
        'vta_sales',
        coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0)
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_pinterest_ad_analytics r
  ),
  pinterest_adset_metrics as (
    select
      cast(r.wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (r._airbyte_data ->> '$.DATE') as date) as stat_date,
      cast(json_unquote (r._airbyte_data ->> '$.CAMPAIGN_ID') as bigint) as campaign_id,
      campaign_type as objective_type,
      cast(json_unquote (r._airbyte_data ->> '$.AD_GROUP_ID') as bigint) as adset_id,
      cast(json_unquote (r._airbyte_data ->> '$.SPEND_IN_DOLLAR') as double) as ad_spend,
      cast(json_unquote (r._airbyte_data ->> '$.IMPRESSION_1') as bigint) as impressions,
      cast(json_unquote (r._airbyte_data ->> '$.CLICKTHROUGH_1') as bigint) as clicks,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CHECKOUT') as bigint) as ads_orders,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CHECKOUT_VALUE_IN_MICRO_DOLLAR') / 1000000 as bigint) as ads_sales,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_3SEC_VIEWS') as bigint) as video_views,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P25_COMBINED') as bigint) as video_views_p25,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P50_COMBINED') as bigint) as video_views_p50,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P75_COMBINED') as bigint) as video_views_p75,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIDEO_P100_COMPLETE') as bigint) as video_views_p100,
      cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CONVERSIONS') as bigint) as conversion,
      0 as conversion_value,
      JSON_OBJECT (
        'orders_details',
        JSON_OBJECT (
          'vta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT') as bigint), 0),
          'cta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CLICK_CHECKOUT') as bigint), 0),
          'engagement',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_ENGAGEMENT_CHECKOUT') as bigint), 0)
        ),
        'sales_details',
        JSON_OBJECT (
          'vta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0),
          'cta',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0),
          'engagement',
          coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_ENGAGEMENT_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0)
        ),
        'vta_orders',
        coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT') as bigint), 0),
        'vta_sales',
        coalesce(cast(json_unquote (r._airbyte_data ->> '$.TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR') as double) / 1000000, 0)
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_pinterest_ad_group_analytics r
        join platform_offline.dwd_view_analytics_ads_all_level_settings_v20241225 campaign on r.wm_tenant_id = campaign.tenant_id
        and json_extract (r._airbyte_data, '$.CAMPAIGN_ID') = campaign_id
        and campaign.type = 'campaign'
        and campaign.ads_platform = 'Pinterest'
  ),
  tt_metrics as (
    select
      cast(r.wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (r._airbyte_data ->> '$.stat_time_day') as date) as stat_date,
      cast(json_unquote (r._airbyte_data ->> '$.ad_id') as bigint) as ad_id,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.spend') as double) as ad_spend,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.impressions') as bigint) as impressions,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.clicks') as bigint) as clicks,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') as bigint) as ads_orders,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') * json_unquote (r._airbyte_data ->> '$.metrics.value_per_complete_payment') as double) as ads_sales,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.video_watched_2s') as bigint) as video_views,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.video_views_p25') as bigint) as video_views_p25,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.video_views_p50') as bigint) as video_views_p50,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.video_views_p75') as bigint) as video_views_p75,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.video_views_p100') as bigint) as video_views_p100,
      cast(json_extract(_airbyte_data, '$.metrics.onsite_shopping') as double ) onsite_shopping,
      cast(json_extract(_airbyte_data, '$.metrics.total_onsite_shopping_value') as double ) as total_onsite_shopping_value,
      cast(json_unquote (r._airbyte_data ->> '$.metrics.conversion') as bigint) as conversion,
      0 as conversion_value,
      JSON_OBJECT (
        'cta_conversion',
        cast(json_unquote (r._airbyte_data ->> '$.metrics.cta_conversion') as bigint),
        'vta_conversion',
        cast(json_unquote (r._airbyte_data ->> '$.metrics.vta_conversion') as bigint),
        'orders_details',
        JSON_OBJECT (
          'website',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Website" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') as bigint)
              else 0
              end,
            'vta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Website" then cast(json_unquote (r._airbyte_data ->> '$.metrics.vta_conversion') as bigint)
              else 0
              end
          ),
          'other',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Other" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') as bigint)
              else 0
              end,
            'vta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Other" then cast(json_unquote (r._airbyte_data ->> '$.metrics.vta_conversion') as bigint)
              else 0
              end
          ),
          'app',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "App" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') as bigint)
              else 0
              end,
            'vta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "App" then cast(json_unquote (r._airbyte_data ->> '$.metrics.vta_conversion') as bigint)
              else 0
              end
          ),
          'cta',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.complete_payment') as bigint), 0) - coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint), 0) - coalesce(cast(json_extract (r._airbyte_data, '$.metrics.evta_payments_completed') as bigint), 0),
          'vta',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint), 0),
          'engagement',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.evta_payments_completed') as bigint), 0)
        ),
        'sales_details',
        JSON_OBJECT (
          'website',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Website" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') * json_unquote (r._airbyte_data ->> '$.metrics.value_per_complete_payment') as double)
              else 0
              end
          ),
          'other',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "Other" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') * json_unquote (r._airbyte_data ->> '$.metrics.value_per_complete_payment') as double)
              else 0
              end
          ),
          'app',
          JSON_OBJECT (
            'cta',
            case
              when json_unquote (r._airbyte_data ->> '$.metrics.promotion_type') = "App" then cast(json_unquote (r._airbyte_data ->> '$.metrics.complete_payment') * json_unquote (r._airbyte_data ->> '$.metrics.value_per_complete_payment') as double)
              else 0
              end
          ),
          'cta',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.complete_payment') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.value_per_complete_payment') as double), 0) - coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.cost_per_vta_payments_completed') as double), 0) - coalesce(cast(json_extract (r._airbyte_data, '$.metrics.evta_payments_completed') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.cost_per_evta_payments_completed') as double), 0),
          'vta',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.cost_per_vta_payments_completed') as double), 0),
          'engagement',
          coalesce(cast(json_extract (r._airbyte_data, '$.metrics.evta_payments_completed') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.cost_per_evta_payments_completed') as double), 0)
        ),
        'vta_orders',
        coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint), 0),
        'vta_sales',
        coalesce(cast(json_extract (r._airbyte_data, '$.metrics.vta_complete_payment') as bigint) * cast(json_extract (r._airbyte_data, '$.metrics.cost_per_vta_payments_completed') as double), 0)
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_tiktok_marketing_ads_reports_daily r
  ),
  meta_data as (
    select * from airbyte_destination_v2.raw_facebook_marketing_ads_insights
    where json_extract (_airbyte_data, '$.date_start') >= date_format(utc_date() - interval 65 day, '%Y-%m-%d')
  ),
  meta_metrics as (
    select
      cast(a.wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (_airbyte_data ->> '$."date_start"') as date) as stat_date,
      cast(json_unquote (_airbyte_data ->> '$."ad_id"') as bigint) as ad_id,
      cast(json_unquote (_airbyte_data ->> '$."adset_id"') as bigint) as adset_id,
      cast(json_unquote (_airbyte_data ->> '$."campaign_id"') as bigint) as campaign_id,
      cast(replace(json_unquote (_airbyte_data ->> '$."account_id"'), 'act_', '') as bigint) as account_id,
      cast(json_unquote (_airbyte_data ->> '$."spend"') as double) as ad_spend,
      cast(json_unquote (_airbyte_data ->> '$."impressions"') as bigint) as impressions,
      cast(json_unquote (_airbyte_data ->> '$."clicks"') as bigint) as clicks,
      actions_onsite_conversion_purchase + action_app_custom_fb_mobile_purchase + actions_offsite_fb_pixel_purchase as ads_orders,
      actions_values_onsite_conversion_purchase + action_value_fb_mobile_purchase + action_value_offsite_fb_pixel_purchase as ads_sales,
      video_views,
      (
        case optimization_goal -- 后续开发，先这样写, 根据 optimization_goal 来判断 不同的 conversion 怎么计算
          when 'LANDING_PAGE_VIEWS' then actions_landing_page_view
          else actions_onsite_conversion_purchase + actions_offsite_fb_pixel_purchase
          end
        ) as conversion,
      (
        case optimization_goal
          when 'LANDING_PAGE_VIEWS' then action_value_landing_page_view
          else actions_values_onsite_conversion_purchase + action_value_offsite_fb_pixel_purchase
          end
        ) as conversion_value,
      video_views_p25,
      video_views_p50,
      video_views_p75,
      video_views_p100,
      JSON_OBJECT (
        'unique_clicks',
        cast(json_unquote (_airbyte_data ->> '$."unique_clicks"') as bigint),
        'unique_inline_link_clicks',
        cast(json_unquote (_airbyte_data ->> '$."unique_inline_link_clicks"') as bigint),
        'actions_cta_landing_page_view',
        actions_cta_landing_page_view,
        'actions_cta_view_content',
        actions_cta_view_content,
        'actions_onsite_web_purchase',
        actions_onsite_web_purchase,
        'actions_cta_onsite_web_purchase',
        actions_cta_onsite_web_purchase,
        'actions_vta_onsite_web_purchase',
        actions_vta_onsite_web_purchase,
        'actions_onsite_web_app_purchase',
        actions_onsite_web_app_purchase,
        'actions_cta_onsite_web_app_purchase',
        actions_cta_onsite_web_app_purchase,
        'actions_vta_onsite_web_app_purchase',
        actions_vta_onsite_web_app_purchase,
        'actions_onsite_conversion_purchase',
        actions_onsite_conversion_purchase,
        'actions_cta_onsite_conversion_purchase',
        actions_cta_onsite_conversion_purchase,
        'actions_vta_onsite_conversion_purchase',
        actions_vta_onsite_conversion_purchase,
        'actions_offsite_fb_pixel_purchase',
        actions_offsite_fb_pixel_purchase,
        'actions_cta_offsite_fb_pixel_purchase',
        actions_cta_offsite_fb_pixel_purchase,
        'actions_vta_offsite_fb_pixel_purchase',
        actions_vta_offsite_fb_pixel_purchase,
        'actions_values_onsite_web_purchase',
        actions_values_onsite_web_purchase,
        'actions_values_onsite_conversion_purchase',
        actions_values_onsite_conversion_purchase,
        'cta_omni_purchase',
        cta_omni_purchase,
        'vta_omni_purchase',
        vta_omni_purchase,
        'vta_1d_view',
        coalesce(vta_1d_view_purchase, vta_1d_actions_offsite_fb_pixel_purchase, vta_1d_view_omni_purchase),
        'vta_orders',
        action_app_custom_fb_mobile_purchase_1d_view + actions_onsite_conversion_purchase_1d_view + actions_offsite_fb_pixel_purchase_1d_view + action_app_custom_fb_mobile_purchase_1d_ev + actions_onsite_conversion_purchase_1d_ev + actions_offsite_fb_pixel_purchase_1d_ev,
        'vta_1d_view_value',
        actions_values_onsite_conversion_purchase_1d_view + action_value_offsite_fb_pixel_purchase_1d_view,
        'vta_sales',
        actions_values_onsite_conversion_purchase_1d_view + action_value_fb_mobile_purchase_1d_view + action_value_offsite_fb_pixel_purchase_1d_view + actions_values_onsite_conversion_purchase_1d_ev + action_value_fb_mobile_purchase_1d_ev + action_value_offsite_fb_pixel_purchase_1d_ev,
        'unique_omni_purchase',
        unique_omni_purchase,
        'unique_cta_omni_purchase',
        unique_cta_omni_purchase,
        'orders_details',
        JSON_OBJECT (
          'onsite',
          JSON_OBJECT (
            '1d_click',
            coalesce(actions_onsite_conversion_purchase_1d_click, 0),
            '7d_click',
            coalesce(actions_onsite_conversion_purchase_7d_click, 0),
            '28d_click',
            coalesce(actions_onsite_conversion_purchase_28d_click, 0),
            '1d_view',
            coalesce(actions_onsite_conversion_purchase_1d_view, 0),
            '1d_ev',
            coalesce(actions_onsite_conversion_purchase_1d_ev, 0)
          ),
          'app',
          JSON_OBJECT (
            '1d_click',
            coalesce(action_app_custom_fb_mobile_purchase_1d_click, 0),
            '7d_click',
            coalesce(action_app_custom_fb_mobile_purchase_7d_click, 0),
            '28d_click',
            coalesce(action_app_custom_fb_mobile_purchase_28d_click, 0),
            '1d_view',
            coalesce(action_app_custom_fb_mobile_purchase_1d_view, 0),
            '1d_ev',
            coalesce(action_app_custom_fb_mobile_purchase_1d_ev, 0)
          ),
          'website',
          JSON_OBJECT (
            '1d_click',
            coalesce(actions_offsite_fb_pixel_purchase_1d_click, 0),
            '7d_click',
            coalesce(actions_offsite_fb_pixel_purchase_7d_click, 0),
            '28d_click',
            coalesce(actions_offsite_fb_pixel_purchase_28d_click, 0),
            '1d_view',
            coalesce(actions_offsite_fb_pixel_purchase_1d_view, 0),
            '1d_ev',
            coalesce(actions_offsite_fb_pixel_purchase_1d_ev, 0)
          )
        ),
        'sales_details',
        JSON_OBJECT (
          'onsite',
          JSON_OBJECT (
            '1d_click',
            coalesce(actions_values_onsite_conversion_purchase_1d_click, 0),
            '7d_click',
            coalesce(actions_values_onsite_conversion_purchase_7d_click, 0),
            '28d_click',
            coalesce(actions_values_onsite_conversion_purchase_28d_click, 0),
            '1d_view',
            coalesce(actions_values_onsite_conversion_purchase_1d_view, 0),
            '1d_ev',
            coalesce(actions_values_onsite_conversion_purchase_1d_ev, 0)
          ),
          'app',
          JSON_OBJECT (
            '1d_click',
            coalesce(action_value_fb_mobile_purchase_1d_click, 0),
            '7d_click',
            coalesce(action_value_fb_mobile_purchase_7d_click, 0),
            '28d_click',
            coalesce(action_value_fb_mobile_purchase_28d_click, 0),
            '1d_view',
            coalesce(action_value_fb_mobile_purchase_1d_view, 0),
            '1d_ev',
            coalesce(action_value_fb_mobile_purchase_1d_ev, 0)
          ),
          'website',
          JSON_OBJECT (
            '1d_click',
            coalesce(action_value_offsite_fb_pixel_purchase_1d_click, 0),
            '7d_click',
            coalesce(action_value_offsite_fb_pixel_purchase_7d_click, 0),
            '28d_click',
            coalesce(action_value_offsite_fb_pixel_purchase_28d_click, 0),
            '1d_view',
            coalesce(action_value_offsite_fb_pixel_purchase_1d_view, 0),
            '1d_ev',
            coalesce(action_value_offsite_fb_pixel_purchase_1d_ev, 0)
          )
        )
      ) as extra_metrics
    from
      meta_data a
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          optimization_goal,
          sum(
            cast(
              case concat('omni_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as omni_conversion,
          sum(
            cast(
              case concat('app_custom_event.fb_mobile_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as fb_mobile_conversion,
          sum(
            cast(
              case concat('offsite_conversion.fb_pixel_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as fb_pixel_conversion,
          sum(
            cast(
              case concat('onsite_app_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as onsite_app_conversion,
          sum(
            cast(
              case concat('onsite_web_app_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as onsite_web_app_conversion,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as ads_orders,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as cta_omni_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.28d_view')
                else 0
                end as bigint
            )
          ) as vta_omni_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as bigint
            )
          ) as vta_1d_view_omni_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as bigint
            )
          ) as vta_1d_view_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'video_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as video_views,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as double
            )
          ) as action_app_custom_fb_mobile_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as actions_onsite_web_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_onsite_web_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_purchase' then json_unquote (action ->> '$.28d_view')
                else 0
                end as bigint
            )
          ) as actions_vta_onsite_web_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_app_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as actions_onsite_web_app_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_app_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_onsite_web_app_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_app_purchase' then json_unquote (action ->> '$.28d_view')
                else 0
                end as bigint
            )
          ) as actions_vta_onsite_web_app_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as bigint
            )
          ) as actions_onsite_conversion_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_onsite_conversion_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.28d_view')
                else 0
                end as bigint
            )
          ) as actions_vta_onsite_conversion_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as bigint
            )
          ) as actions_offsite_fb_pixel_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as bigint
            )
          ) as vta_1d_actions_offsite_fb_pixel_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_offsite_fb_pixel_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.28d_view')
                else 0
                end as bigint
            )
          ) as actions_vta_offsite_fb_pixel_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'landing_page_view' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_landing_page_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'landing_page_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as actions_landing_page_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'view_content' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as actions_cta_view_content
        from
          (
            select
              ins._airbyte_raw_id,
              ins.wm_tenant_id,
              json_unquote (adset_settings ->> '$.optimization_goal') optimization_goal,
              lower(json_unquote (coalesce(adset_settings ->> '$.promoted_object.custom_event_type', adset_settings ->> '$.promoted_object.omnichannel_object.pixel[0].custom_event_type'))) as conversion_event,
              cast(json_extract (ins._airbyte_data, '$.actions') as ARRAY < JSON >) as actions_array
            from
              meta_data ins
                left join platform_offline.dwd_view_analytics_ads_all_level_settings_v20241225 adset on ins.wm_tenant_id = adset.tenant_id
                and json_extract (ins._airbyte_data, '$.adset_id') = adset.adset_id
                and adset.type = 'adset'
                and adset.raw_platform = 'facebookMarketing'
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) b on a._airbyte_raw_id = b._airbyte_raw_id
        and a.wm_tenant_id = b.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case concat('omni_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as omni_conversion_value,
          sum(
            cast(
              case concat('app_custom_event.fb_mobile_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as fb_mobile_conversion_value,
          sum(
            cast(
              case concat('offsite_conversion.fb_pixel_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as fb_pixel_conversion_value,
          sum(
            cast(
              case concat('onsite_app_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as onsite_app_conversion_value,
          sum(
            cast(
              case concat('onsite_web_app_', conversion_event)
                when json_unquote (action ->> '$.action_type') then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as onsite_web_app_conversion_value,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as ads_sales,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'app_custom_event.fb_mobile_purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as double
            )
          ) as action_value_fb_mobile_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_web_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as actions_values_onsite_web_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'onsite_conversion.purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as double
            )
          ) as actions_values_onsite_conversion_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_click')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase_1d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.7d_click')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase_7d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase_28d_click,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_ev')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase_1d_ev,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'offsite_conversion.fb_pixel_purchase' then json_unquote (action ->> '$.1d_view')
                else 0
                end as double
            )
          ) as action_value_offsite_fb_pixel_purchase_1d_view,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'landing_page_view' then json_unquote (action ->> '$.value')
                else 0
                end as double
            )
          ) as action_value_landing_page_view
        from
          (
            select
              ins._airbyte_raw_id,
              ins.wm_tenant_id,
              lower(json_unquote (coalesce(adset_settings ->> '$.promoted_object.custom_event_type', adset_settings ->> '$.promoted_object.omnichannel_object.pixel[0].custom_event_type'))) as conversion_event,
              cast(json_extract (ins._airbyte_data, '$.action_values') as ARRAY < JSON >) as actions_array
            from
              meta_data ins
                left join platform_offline.dwd_view_analytics_ads_all_level_settings_v20241225 adset on ins.wm_tenant_id = adset.tenant_id
                and json_extract (ins._airbyte_data, '$.adset_id') = adset.adset_id
                and adset.type = 'adset'
                and adset.raw_platform = 'facebookMarketing'
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) c on a._airbyte_raw_id = c._airbyte_raw_id
        and a.wm_tenant_id = c.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as unique_omni_purchase,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'omni_purchase' then json_unquote (action ->> '$.28d_click')
                else 0
                end as bigint
            )
          ) as unique_cta_omni_purchase
        from
          (
            select
              _airbyte_raw_id,
              wm_tenant_id,
              cast(json_extract (_airbyte_data, '$.unique_actions') as ARRAY < JSON >) as actions_array
            from
              meta_data
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) c1 on a._airbyte_raw_id = c1._airbyte_raw_id
        and a.wm_tenant_id = c1.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'video_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as video_views_p25
        from
          (
            select
              _airbyte_raw_id,
              wm_tenant_id,
              cast(json_extract (_airbyte_data, '$.video_p25_watched_actions') as ARRAY < JSON >) as actions_array
            from
              meta_data
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) d on a._airbyte_raw_id = d._airbyte_raw_id
        and a.wm_tenant_id = d.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'video_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as video_views_p50
        from
          (
            select
              _airbyte_raw_id,
              wm_tenant_id,
              cast(json_extract (_airbyte_data, '$.video_p50_watched_actions') as ARRAY < JSON >) as actions_array
            from
              meta_data
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) e on a._airbyte_raw_id = e._airbyte_raw_id
        and a.wm_tenant_id = e.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'video_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as video_views_p75
        from
          (
            select
              _airbyte_raw_id,
              wm_tenant_id,
              cast(json_extract (_airbyte_data, '$.video_p75_watched_actions') as ARRAY < JSON >) as actions_array
            from
              meta_data
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) f on a._airbyte_raw_id = f._airbyte_raw_id
        and a.wm_tenant_id = f.wm_tenant_id
        left join (
        select
          _airbyte_raw_id,
          wm_tenant_id,
          sum(
            cast(
              case
                when json_unquote (action ->> '$.action_type') = 'video_view' then json_unquote (action ->> '$.value')
                else 0
                end as bigint
            )
          ) as video_views_p100
        from
          (
            select
              _airbyte_raw_id,
              wm_tenant_id,
              cast(json_extract (_airbyte_data, '$.video_p100_watched_actions') as ARRAY < JSON >) as actions_array
            from
              meta_data
          ) as ads
            cross join UNNEST (ads.actions_array) as t (action)
        group by
          1,
          2
      ) g on a._airbyte_raw_id = g._airbyte_raw_id
        and a.wm_tenant_id = g.wm_tenant_id
  ),
  google_ad_category_conversion as (
    select
      wm_tenant_id tenant_id,
      json_extract (_airbyte_data, '$."ad_group_ad.ad.id"') ad_id,
      json_extract (_airbyte_data, '$."ad_group.id"') adset_id,
      json_extract (_airbyte_data, '$."segments.date"') stat_date,
      json_extract (_airbyte_data, '$."segments.conversion_action_category"') category,
      sum(cast(json_unquote (_airbyte_data ->> '$."metrics.conversions"') as double)) conversions,
      sum(cast(json_unquote (_airbyte_data ->> '$."metrics.conversions_value"') as double)) conversions_value
    from
      airbyte_destination_v2.raw_google_ads_ad_conversions
    group by
      1,
      2,
      3,
      4,
      5
  ),
  google_campaign_category_conversion as (
    select
      wm_tenant_id tenant_id,
      json_extract (_airbyte_data, '$."campaign.id"') campaign_id,
      json_extract (_airbyte_data, '$."segments.date"') stat_date,
      json_extract (_airbyte_data, '$."segments.conversion_action_category"') category,
      sum(cast(json_unquote (_airbyte_data ->> '$."metrics.conversions"') as double)) conversions,
      sum(cast(json_unquote (_airbyte_data ->> '$."metrics.conversions_value"') as double)) conversions_value
    from
      airbyte_destination_v2.raw_google_ads_campaign_conversions
    group by
      1,
      2,
      3,
      4
  ),
  google_ad_metrics as (
    select
      cast(wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (_airbyte_data ->> '$."segments.date"') as date) as stat_date,
      cast(json_extract (_airbyte_data, '$."ad_group_ad.ad.id"') as varchar) as ad_id,
      cast(json_unquote (_airbyte_data ->> '$."ad_group.id"') as bigint) as adset_id,
      cast(json_unquote (_airbyte_data ->> '$."campaign.id"') as bigint) as campaign_id,
      cast(json_unquote (_airbyte_data ->> '$."customer.id"') as bigint) as account_id,
      json_unquote (_airbyte_data ->> '$."campaign.advertising_channel_type"') as advertising_channel_type,
      cast(json_unquote (_airbyte_data ->> '$."metrics.cost_micros"') / 1000000 as double) as ad_spend,
      cast(json_unquote (_airbyte_data ->> '$."metrics.impressions"') as bigint) as impressions,
      cast(json_unquote (_airbyte_data ->> '$."metrics.video_views"') as bigint) as video_views,
      cast(json_unquote (_airbyte_data ->> '$."metrics.clicks"') as bigint) as clicks,
      coalesce(cc.conversions, cast(json_unquote (_airbyte_data ->> '$."metrics.orders"') as double)) ads_orders,
      cast(json_unquote (_airbyte_data ->> '$."metrics.conversions"') as double) as conversion,
      cast(json_unquote (_airbyte_data ->> '$."metrics.conversions_value"') as double) as conversion_value,
      coalesce(cc.conversions_value, cast(json_unquote (_airbyte_data ->> '$."metrics.average_order_value_micros"') * json_unquote (_airbyte_data ->> '$."metrics.orders"') / 1000000 as double)) ads_sales,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p25_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p25,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p50_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p50,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p75_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p75,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p100_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p100,
      JSON_OBJECT (
        'orders_details',
        JSON_OBJECT (
          'cta',
          coalesce(cc.conversions, cast(json_unquote (_airbyte_data ->> '$."metrics.orders"') as double), 0),
          'vta',
          coalesce(cast(json_unquote (_airbyte_data ->> '$."metrics.view_through_conversions"') as double), 0)
        ),
        'sales_details',
        JSON_OBJECT (
          'cta',
          coalesce(cc.conversions_value, cast(json_unquote (_airbyte_data ->> '$."metrics.average_order_value_micros"') * json_unquote (_airbyte_data ->> '$."metrics.orders"') / 1000000 as double), 0)
        )
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_google_ads_ad_metrics m
        left join google_ad_category_conversion cc on m.wm_tenant_id = cc.tenant_id
        and json_unquote (m._airbyte_data ->> '$."ad_group_ad.ad.id"') = cc.ad_id
        and json_unquote (m._airbyte_data ->> '$."ad_group.id"') = cc.adset_id
        and json_unquote (m._airbyte_data ->> '$."segments.date"') = cc.stat_date
        and cc.category = 'PURCHASE'
  ),
  google_campaign_metrics as (
    select
      cast(wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (_airbyte_data ->> '$."segments.date"') as date) as stat_date,
      cast(json_unquote (_airbyte_data ->> '$."campaign.id"') as bigint) as campaign_id,
      cast(json_unquote (_airbyte_data ->> '$."customer.id"') as bigint) as account_id,
      cast(json_unquote (_airbyte_data ->> '$."ad_group.id"') as bigint) as adset_id,
      json_unquote (_airbyte_data ->> '$."campaign.advertising_channel_type"') as advertising_channel_type,
      cast(json_unquote (_airbyte_data ->> '$."metrics.cost_micros"') / 1000000 as double) as ad_spend,
      cast(json_unquote (_airbyte_data ->> '$."metrics.impressions"') as bigint) as impressions,
      cast(json_unquote (_airbyte_data ->> '$."metrics.video_views"') as bigint) as video_views,
      cast(json_unquote (_airbyte_data ->> '$."metrics.clicks"') as bigint) as clicks,
      coalesce(cc.conversions, cast(json_unquote (_airbyte_data ->> '$."metrics.orders"') as double)) ads_orders,
      cast(json_unquote (_airbyte_data ->> '$."metrics.conversions"') as double) as conversion,
      cast(json_unquote (_airbyte_data ->> '$."metrics.conversions_value"') as double) as conversion_value,
      coalesce(cc.conversions_value, cast(json_unquote (_airbyte_data ->> '$."metrics.average_order_value_micros"') * json_unquote (_airbyte_data ->> '$."metrics.orders"') / 1000000 as double)) ads_sales,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p25_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p25,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p50_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p50,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p75_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p75,
      cast(round(json_unquote (_airbyte_data ->> '$."metrics.video_quartile_p100_rate"') * json_unquote (_airbyte_data ->> '$."metrics.video_views"')) as bigint) as video_views_p100,
      JSON_OBJECT (
        'orders_details',
        JSON_OBJECT (
          'cta',
          coalesce(cc.conversions, cast(json_unquote (_airbyte_data ->> '$."metrics.orders"') as double), 0),
          'vta',
          coalesce(cast(json_unquote (_airbyte_data ->> '$."metrics.view_through_conversions"') as double), 0)
        ),
        'sales_details',
        JSON_OBJECT (
          'cta',
          coalesce(cc.conversions_value, cast(json_unquote (_airbyte_data ->> '$."metrics.average_order_value_micros"') * json_unquote (_airbyte_data ->> '$."metrics.orders"') / 1000000 as double), 0)
        )
      ) as extra_metrics
    from
      airbyte_destination_v2.raw_google_ads_campaign_metrics m
        left join google_campaign_category_conversion cc on m.wm_tenant_id = cc.tenant_id
        and json_extract (m._airbyte_data, '$."campaign.id"') = cc.campaign_id
        and json_extract (m._airbyte_data, '$."segments.date"') = cc.stat_date
        and cc.category = 'PURCHASE'
  ),
  bing_ads as (
    select
      cast(wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (_airbyte_data ->> '$."TimePeriod"') as date) as stat_date,
      cast(json_unquote (_airbyte_data ->> '$."AdId"') as bigint) as ad_id,
      cast(json_unquote (_airbyte_data ->> '$."AdGroupId"') as bigint) as adset_id,
      cast(json_unquote (_airbyte_data ->> '$."CampaignId"') as bigint) as campaign_id,
      cast(json_unquote (_airbyte_data ->> '$."AccountId"') as bigint) as account_id,
      sum(cast(json_unquote (_airbyte_data ->> '$."Spend"') as double)) as ad_spend,
      sum(cast(json_unquote (_airbyte_data ->> '$."Impressions"') as bigint)) as impressions,
      sum(cast(json_unquote (_airbyte_data ->> '$."Clicks"') as bigint)) as clicks,
      sum(cast(json_unquote (_airbyte_data ->> '$."ConversionsQualified"') as double)) as ads_orders,
      sum(cast(json_unquote (_airbyte_data ->> '$."Revenue"') as double)) as ads_sales,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllConversionsQualified"') as double)) as conversion,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllRevenue"') as double)) as conversion_value,
      sum(cast(json_unquote (_airbyte_data ->> '$."VideoViews"') as double)) as video_views,
      sum(cast(json_unquote (_airbyte_data ->> '$."VideoViewsAt25Percent"') as double)) as video_views_p25,
      sum(cast(json_unquote (_airbyte_data ->> '$."VideoViewsAt50Percent"') as double)) as video_views_p50,
      sum(cast(json_unquote (_airbyte_data ->> '$."VideoViewsAt75Percent"') as double)) as video_views_p75,
      sum(cast(json_unquote (_airbyte_data ->> '$."CompletedVideoViews"') as double)) as video_views_p100,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllConversionsQualified"') as double)) as cta,
      sum(cast(json_unquote (_airbyte_data ->> '$."ViewThroughConversionsQualified"') as double)) as vta,
      sum(cast(json_unquote (_airbyte_data ->> '$."Revenue"') as double)) as revenue,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllRevenue"') as double)) as all_revenue
    from
      airbyte_destination_v2.raw_bing_ads_ad_performance_report_daily
    group by
      1,
      2,
      3,
      4,
      5,
      6
  ),
  bing_ads_metrics as (
    select
      tenant_id,
      stat_date,
      ad_id,
      adset_id,
      campaign_id,
      account_id,
      ad_spend,
      impressions,
      clicks,
      ads_orders,
      ads_sales,
      conversion,
      conversion_value,
      video_views,
      video_views_p25,
      video_views_p50,
      video_views_p75,
      video_views_p100,
      cta,
      vta,
      revenue,
      all_revenue,
      JSON_OBJECT ('orders_details', JSON_OBJECT ('cta', ads_orders, 'vta', vta), 'sales_details', JSON_OBJECT ('cta', revenue), 'vta_orders', vta, 'vta_sales', all_revenue) as extra_metrics
    from
      bing_ads
  ),
  bing_campaign as (
    select
      cast(wm_tenant_id as bigint) as tenant_id,
      cast(json_unquote (_airbyte_data ->> '$."TimePeriod"') as date) as stat_date,
      cast(json_unquote (_airbyte_data ->> '$."CampaignId"') as bigint) as campaign_id,
      cast(json_unquote (_airbyte_data ->> '$."AccountId"') as bigint) as account_id,
      json_unquote (_airbyte_data ->> '$."CampaignType"') as campaign_type,
      sum(cast(json_unquote (_airbyte_data ->> '$."Spend"') as double)) as ad_spend,
      sum(cast(json_unquote (_airbyte_data ->> '$."Impressions"') as bigint)) as impressions,
      sum(cast(json_unquote (_airbyte_data ->> '$."Clicks"') as bigint)) as clicks,
      sum(cast(json_unquote (_airbyte_data ->> '$."ConversionsQualified"') as double)) as ads_orders,
      sum(cast(json_unquote (_airbyte_data ->> '$."Revenue"') as double)) as ads_sales,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllConversionsQualified"') as double)) as conversion,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllRevenue"') as double)) as conversion_value,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllConversionsQualified"') as double)) as cta,
      sum(cast(json_unquote (_airbyte_data ->> '$."ViewThroughConversionsQualified"') as double)) as vta,
      sum(cast(json_unquote (_airbyte_data ->> '$."Revenue"') as double)) as revenue,
      sum(cast(json_unquote (_airbyte_data ->> '$."AllRevenue"') as double)) as all_revenue
    from
      airbyte_destination_v2.raw_bing_ads_campaign_performance_report_daily
    group by
      1,
      2,
      3,
      4,
      5
  ),
  bing_campaign_metrics as (
    select
      tenant_id,
      stat_date,
      campaign_id,
      account_id,
      campaign_type,
      ad_spend,
      impressions,
      clicks,
      ads_orders,
      ads_sales,
      conversion,
      conversion_value,
      0 as video_views,
      0 as video_views_p25,
      0 as video_views_p50,
      0 as video_views_p75,
      0 as video_views_p100,
      JSON_OBJECT ('orders_details', JSON_OBJECT ('cta', ads_orders, 'vta', vta), 'sales_details', JSON_OBJECT ('cta', revenue), 'vta_orders', vta, 'vta_sales', all_revenue) as extra_metrics
    from
      bing_campaign
  ),
  all_ad_level_metrics as (
    -- snapchat
    select
      'Snapchat' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      setting.adset_id,
      setting.campaign_id,
      setting.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      null as ads_conversions,
      null as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      snapchat_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.ad_id = setting.ad_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Snapchat'
    union all

    -- tiktok gmv max
    select
      'TikTok' as ads_platform,
      m.tenant_id,
      m.stat_date,
      cast(null as varchar) as ad_id,
      cast(null as varchar) as adset_id,
      m.campaign_id,
      setting.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      0 as impressions,
      0 as clicks,
      0 as ads_orders,
      0 as ads_sales,
      m.orders as ads_secondary_orders,
      m.gross_revenue as ads_secondary_sales,
      m.orders as ads_conversions,
      m.gross_revenue as ads_conversions_value,
      0 as video_views,
      0 as video_views_p25,
      0 as video_views_p50,
      0 as video_views_p75,
      0 as video_views_p100,
      setting.exchange_rate,
      cast(null as json) as extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      cast(null as varchar) as adset_name,
      cast(null as varchar) as ad_name,
      setting.campaign_status,
      cast(null as varchar) as adset_status,
      cast(null as varchar) as ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      tiktok_gmv_max_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk as setting on m.campaign_id = setting.campaign_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'campaign'
        and setting.ads_platform = 'TikTok'

    -- tiktok
    union all
    select
      'TikTok' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      setting.adset_id,
      setting.campaign_id,
      setting.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders as ads_orders,
      m.ads_sales as ads_sales,
      m.onsite_shopping as ads_secondary_orders,
      m.total_onsite_shopping_value as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      m.total_onsite_shopping_value as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      tt_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.ad_id = setting.ad_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'TikTok'
    -- WHERE
    --   extra_metrics ->> '$.cta_conversion' > 0
    union all
    -- meta ads
    select
      'Meta' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      m.adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      coalesce(m.conversion_value, 0) as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      meta_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.ad_id = setting.ad_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Meta'
    union all
    -- google ads - non-pmax
    select
      'Google' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      m.adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      JSON_OBJECT () as creative_base_settings,
      JSON_OBJECT () as adset_base_settings,
      setting.campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      coalesce(m.conversion_value, 0) as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      google_ad_metrics m
        left join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on setting.tenant_id = m.tenant_id
        and m.ad_id = setting.ad_id
        and m.adset_id = setting.adset_id
        and setting.campaign_id = m.campaign_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Google'
    union all
    -- google ads - pmax
    select
      'Google' as ads_platform,
      m.tenant_id,
      m.stat_date,
      '' as ad_id,
      '' as adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      JSON_OBJECT () as creative_base_settings,
      JSON_OBJECT () as adset_base_settings,
      setting.campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      coalesce(m.conversion_value, 0) as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      google_campaign_metrics m
        left join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on setting.tenant_id = m.tenant_id
        and setting.campaign_id = m.campaign_id
        and setting.ads_join_level = 'campaign'
        and setting.ads_platform = 'Google'
    where
      m.advertising_channel_type = 'PERFORMANCE_MAX'
    union all
    select
      'Pinterest' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      setting.adset_id,
      setting.campaign_id,
      setting.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      null as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      pinterest_ad_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.ad_id = setting.ad_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Pinterest'
    union all
    select
      'Pinterest' as ads_platform,
      m.tenant_id,
      m.stat_date,
      null as ad_id,
      setting.adset_id,
      setting.campaign_id,
      setting.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      null as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      pinterest_adset_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.adset_id = setting.adset_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'adset'
        and setting.ads_platform = 'Pinterest'
    where
      m.objective_type = 'CATALOG_SALES'
    union all
    select
      'Microsoft' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      m.adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      coalesce(m.conversion_value, 0) as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      bing_ads_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.ad_id = setting.ad_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Microsoft'
    union all
    select
      'Microsoft' as ads_platform,
      m.tenant_id,
      m.stat_date,
      '' as ad_id,
      '' as adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      setting.creative_base_settings,
      JSON_OBJECT () adset_base_settings,
      JSON_OBJECT () campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      coalesce(m.conversion, 0) as ads_conversions,
      coalesce(m.conversion_value, 0) as ads_conversions_value,
      m.video_views,
      m.video_views_p25,
      m.video_views_p50,
      m.video_views_p75,
      m.video_views_p100,
      setting.exchange_rate,
      m.extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      bing_campaign_metrics m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on m.campaign_id = setting.campaign_id
        and m.tenant_id = setting.tenant_id
        and setting.ads_join_level = 'campaign'
        and setting.ads_platform = 'Microsoft'
    where
      m.campaign_type = 'Performance max'
    union all
    select
      'Applovin' as ads_platform,
      m.tenant_id,
      m.stat_date,
      m.ad_id,
      m.adset_id,
      m.campaign_id,
      m.account_id,
      setting.account_base_settings,
      JSON_OBJECT () as creative_base_settings,
      JSON_OBJECT () as adset_base_settings,
      setting.campaign_base_settings,
      campaign_settings_snapshot,
      adset_settings_snapshot,
      ad_settings_snapshot,
      m.ad_spend,
      m.impressions,
      m.clicks,
      m.ads_orders,
      m.ads_sales,
      cast(0 as double) as ads_secondary_orders,
      cast(0 as double) as ads_secondary_sales,
      m.ads_orders as ads_conversions,
      m.ads_sales as ads_conversions_value,
      0,
      0,
      0,
      0,
      0,
      setting.exchange_rate,
      JSON_OBJECT ('orders_details', JSON_OBJECT ('cta', m.ads_orders, 'vta', 0), 'sales_details', JSON_OBJECT ('cta', m.ads_sales, 'vta', 0), 'vta_orders', 0, 'vta_sales', m.ads_sales) as extra_metrics,
      setting.campaign_type,
      setting.account_name,
      setting.campaign_name,
      setting.adset_name,
      setting.ad_name,
      setting.campaign_status,
      setting.adset_status,
      setting.ad_status,
      setting.campaign_status_with_finish,
      setting.campaign_start_time,
      setting.campaign_end_time,
      setting.ads_join_level,
      setting.ads_setting_key,
      setting.tactic_name
    from
      applovin_metrics_merge m
        join platform_offline.dwd_view_analytics_ads_ad_level_settings_v20250725xk setting on setting.tenant_id = m.tenant_id
        and m.ad_id = setting.ad_id
        and setting.campaign_id = m.campaign_id
        and setting.ads_join_level = 'ad'
        and setting.ads_platform = 'Applovin'
  )
select
  a.tenant_id,
  cast(stat_date as varchar) as stat_date,
  cast(sum(ad_spend) * exchange_rate as bigint) as spend
from
  all_ad_level_metrics as a
join platform_offline.ods_campaign_action_cache as b
on a.tenant_id = b.tenant_id and a.campaign_id = b.campaign_id and a.adset_id = b.adset_id and a.ad_id = b.ad_id
where stat_date >= utc_date() - interval 30 day
  and a.tenant_id = {{tenant_id}}
  and ads_platform = '{{ads_platform}}'
  and ad_spend > 0
  and account_base_settings is null
group by 1, 2
`

type LossData struct {
	TenantId int64  `json:"tenant_id"`
	StatDate string `json:"stat_date"`
	Spend    int64  `json:"spend"`
}

func GetDataWithPlatform(tenantId int64, platform string) []LossData {
	platform = backend.PlatformMap[platform]
	db := platform_db.GetDB()
	var res = []LossData{}

	exec := strings.ReplaceAll(query_data, "{{tenant_id}}", fmt.Sprintf("%d", tenantId))
	exec = strings.ReplaceAll(exec, "{{ads_platform}}", platform)

	if err := db.Raw(exec).Limit(-1).Scan(&res).Error; err != nil {
		log.Println(err)
		panic(err)
	}

	//for index := range res {
	//	if res[index].Spend == 0 {
	//		res[index].Spend = -1
	//	}
	//}

	return res
}
