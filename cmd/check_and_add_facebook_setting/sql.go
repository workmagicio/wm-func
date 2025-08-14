package main

var query_loss_ad_data = `with
    meta_data as (
        select a.* from airbyte_destination_v2.raw_facebook_marketing_ads_insights as a
                 join platform_offline.dwd_view_analytics_non_testing_tenants  as b
                   on a.wm_tenant_id = b.tenant_id
        where json_extract (_airbyte_data, '$.date_start') >= date_format(utc_date() - interval 65 day, '%Y-%m-%d')
    ),
    meta_metrics as (select cast(a.wm_tenant_id as bigint)                                                          as tenant_id,
                            cast(json_unquote(_airbyte_data ->> '$."date_start"') as date)                        as stat_date,
                            cast(json_unquote(_airbyte_data ->> '$."ad_id"') as varchar)                           as ad_id,
                            cast(json_unquote(_airbyte_data ->> '$."adset_id"') as varchar)                        as adset_id,
                            cast(json_unquote(_airbyte_data ->> '$."campaign_id"') as varchar)                     as campaign_id,
                            cast(replace(json_unquote(_airbyte_data ->> '$."account_id"'), 'act_',
                                         '') as varchar)                                                             as account_id,
                            cast(json_unquote(_airbyte_data ->> '$."spend"') as double)                           as ad_spend
                     from meta_data as a)
,ad_level as (
    select
        tenant_id,
        ad_id
    from meta_metrics
    where true
    and ad_spend > 0
    group by 1, 2
)
,merge as (
    select
        a.tenant_id,
        a.ad_id,
        b.ad_id as ods_ad_id
    from ad_level as a
    left join platform_offline.ods_campaign_action_cache as b
    on a.tenant_id = b.tenant_id and a.ad_id = b.ad_id
    and b.platform = 'facebookMarketing' and b.type = 'ad'
)
select
    * from merge
where ods_ad_id is null
limit 50
`

var query_loss_adset_data = `
with
    meta_data as (
        select a.* from airbyte_destination_v2.raw_facebook_marketing_ads_insights as a
                 join platform_offline.dwd_view_analytics_non_testing_tenants  as b
                   on a.wm_tenant_id = b.tenant_id
        where json_extract (_airbyte_data, '$.date_start') >= date_format(utc_date() - interval 65 day, '%Y-%m-%d')
    ),
    meta_metrics as (select cast(a.wm_tenant_id as bigint)                                                          as tenant_id,
                            cast(json_unquote(_airbyte_data ->> '$."date_start"') as date)                        as stat_date,
                            cast(json_unquote(_airbyte_data ->> '$."ad_id"') as varchar)                           as ad_id,
                            cast(json_unquote(_airbyte_data ->> '$."adset_id"') as varchar)                        as adset_id,
                            cast(json_unquote(_airbyte_data ->> '$."campaign_id"') as varchar)                     as campaign_id,
                            cast(replace(json_unquote(_airbyte_data ->> '$."account_id"'), 'act_',
                                         '') as varchar)                                                             as account_id,
                            cast(json_unquote(_airbyte_data ->> '$."spend"') as double)                           as ad_spend
                     from meta_data as a)
,ad_level as (
    select
        tenant_id,
        adset_id
    from meta_metrics
    where true
    and ad_spend > 0
    group by 1, 2
)
,merge as (
    select
        a.tenant_id,
        a.adset_id,
        b.ad_id as ods_ad_id
    from ad_level as a
    left join platform_offline.ods_campaign_action_cache as b
    on a.tenant_id = b.tenant_id and a.adset_id = b.adset_id
    and b.platform = 'facebookMarketing' and b.type = 'adset'
)
select
    * from merge
where ods_ad_id is null
`

var query_loss_campaign = `
with
    meta_data as (
        select a.* from airbyte_destination_v2.raw_facebook_marketing_ads_insights as a
                 join platform_offline.dwd_view_analytics_non_testing_tenants  as b
                   on a.wm_tenant_id = b.tenant_id
        where json_extract (_airbyte_data, '$.date_start') >= date_format(utc_date() - interval 65 day, '%Y-%m-%d')
    ),
    meta_metrics as (select cast(a.wm_tenant_id as bigint)                                                          as tenant_id,
                            cast(json_unquote(airbyte_data ->> '$."date_start"') as date)                        as stat_date,
                            cast(json_unquote(_airbyte_data ->> '$."ad_id"') as varchar)                           as ad_id,
                            cast(json_unquote(_airbyte_data ->> '$."adset_id"') as varchar)                        as adset_id,
                            cast(json_unquote(_airbyte_data ->> '$."campaign_id"') as varchar)                     as campaign_id,
                            cast(replace(json_unquote(_airbyte_data ->> '$."account_id"'), 'act_',
                                         '') as varchar)                                                             as account_id,
                            cast(json_unquote(_airbyte_data ->> '$."spend"') as double)                           as ad_spend
                     from meta_data as a)
,ad_level as (
    select
        tenant_id,
        campaign_id
    from meta_metrics
    where true
    and ad_spend > 0
    group by 1, 2
)
,merge as (
    select
        a.tenant_id,
        a.campaign_id,
        b.ad_id as ods_ad_id
    from ad_level as a
    left join platform_offline.ods_campaign_action_cache as b
    on a.tenant_id = b.tenant_id and a.campaign_id = b.campaign_id
    and b.platform = 'facebookMarketing' and b.type = 'campaign'
)
select
    * from merge
where ods_ad_id is null
`
