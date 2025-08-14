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
)
   ,data_analysis as (
  select
    a.TENANT_ID,
    RAW_DATE,
    api_spend,
    ad_spend,
    abs(api_spend - ad_spend) as spend_diff,
    case
      when ad_spend = 0 and api_spend > 0 then 100.0
      when ad_spend = 0 and api_spend = 0 then 0.0
      else round(abs(api_spend - ad_spend) * 100.0 / ad_spend, 2)
      end as diff_percentage,
    case
      when RAW_DATE <= utc_date() - interval 60 day then '60天以前'
      when RAW_DATE <= utc_date() - interval 30 day then '30天到60天'
      when RAW_DATE <= utc_date() - interval 15 day then '30天内'
      else '15天内'
      end as time_period,
    case
      when abs(api_spend - ad_spend) = 0 then '数据一致'
      when abs(api_spend - ad_spend) <= 10 then '轻微差异'
      when abs(api_spend - ad_spend) <= 100 then '中等差异'
      else '严重差异'
      end as consistency_level
  from merge as a
  join platform_offline.dwd_view_analytics_non_testing_tenants as b
  on a.TENANT_ID = b.tenant_id

  where api_spend > 0 or ad_spend > 0  -- 过滤掉两边都为0的记录
),
     result as (
     select
       TENANT_ID,
       time_period,
       case
         when consistency_level = '数据一致' then '<span style="color: #52c41a; font-weight: bold;">✅ 数据一致</span>'
         when consistency_level = '轻微差异' then '<span style="color: #faad14; font-weight: bold;">⚠️ 轻微差异</span>'
         when consistency_level = '中等差异' then '<span style="color: #fa8c16; font-weight: bold;">🔸 中等差异</span>'
         else '<span style="color: #ff4d4f; font-weight: bold;">❌ 严重差异</span>'
       end as consistency_status,
       consistency_level,
       count(*) as record_count,
       round(avg(spend_diff), 0) as avg_diff_amount,
       round(avg(diff_percentage), 2) as avg_diff_percentage,
       round(max(spend_diff), 0) as max_diff_amount,
       round(max(diff_percentage), 2) as max_diff_percentage,
       round(sum(api_spend), 0) as total_api_spend,
       round(sum(ad_spend), 0) as total_ads_spend,
       round(sum(spend_diff), 0) as total_diff_amount,
       case
         when avg(diff_percentage) > 50 then '<div style="background-color: #ffebee; padding: 5px; border-radius: 3px; color: #d32f2f; font-weight: bold;">🚨 高风险：平均差异超过50%</div>'
         when avg(diff_percentage) > 20 then '<div style="background-color: #fff3e0; padding: 5px; border-radius: 3px; color: #f57c00; font-weight: bold;">⚠️ 中风险：平均差异超过20%</div>'
         when avg(diff_percentage) > 5 then '<div style="background-color: #f3e5f5; padding: 5px; border-radius: 3px; color: #7b1fa2;">ℹ️ 低风险：平均差异超过5%</div>'
         else '<div style="background-color: #e8f5e8; padding: 5px; border-radius: 3px; color: #2e7d32;">✅ 正常范围</div>'
       end as risk_alert
     from data_analysis
     group by TENANT_ID, time_period, consistency_level
     order by
       TENANT_ID,
       field(time_period, '15天内', '30天内', '30天到60天', '60天以前'),
       field(consistency_level, '数据一致', '轻微差异', '中等差异', '严重差异')
   ),
   summary as (
     select
       '<h3 style="color: #1890ff;">📊 数据一致性监控汇总</h3>' as summary_title,
       concat(
         '<div style="background-color: #f6ffed; border-left: 4px solid #52c41a; padding: 10px; margin: 5px 0;">',
         '<strong>✅ 数据一致：</strong> ', count(case when consistency_level = '数据一致' then 1 end), ' 条记录',
         '</div>'
       ) as consistent_summary,
       concat(
         '<div style="background-color: #fffbe6; border-left: 4px solid #faad14; padding: 10px; margin: 5px 0;">',
         '<strong>⚠️ 轻微差异：</strong> ', count(case when consistency_level = '轻微差异' then 1 end), ' 条记录',
         '</div>'
       ) as minor_diff_summary,
       concat(
         '<div style="background-color: #fff2e8; border-left: 4px solid #fa8c16; padding: 10px; margin: 5px 0;">',
         '<strong>🔸 中等差异：</strong> ', count(case when consistency_level = '中等差异' then 1 end), ' 条记录',
         '</div>'
       ) as medium_diff_summary,
       concat(
         '<div style="background-color: #fff1f0; border-left: 4px solid #ff4d4f; padding: 10px; margin: 5px 0;">',
         '<strong>❌ 严重差异：</strong> ', count(case when consistency_level = '严重差异' then 1 end), ' 条记录',
         '</div>'
       ) as severe_diff_summary
     from data_analysis
   )
select 
  TENANT_ID,
  time_period,
  consistency_status,
  record_count,
  avg_diff_amount,
  avg_diff_percentage,
  max_diff_amount,
  max_diff_percentage,
  total_api_spend,
  total_ads_spend,
  total_diff_amount,
  risk_alert
from result
where consistency_level != '数据一致'
order by 
  case when consistency_level = '严重差异' then 1
       when consistency_level = '中等差异' then 2
       when consistency_level = '轻微差异' then 3
       else 4 end,
  avg_diff_percentage desc, 
  total_diff_amount desc
