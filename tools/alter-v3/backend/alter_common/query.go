package alter_common

var query_platform_with_platform = `
select
    a.tenant_id as tenant_id
from
    platform.account_connection as a
join platform_offline.dwd_view_analytics_non_testing_tenants as b
on a.tenant_id = b.tenant_id
and platform = '{{platform}}'
# and status = 'connected'
and accounts not like '%ACCESS_REMOVE%'
group by 1
`

var query_platform_with_shopify = `
select
    a.tenant_id as tenant_id
from
    platform.shopify_connection as a
        join platform_offline.dwd_view_analytics_non_testing_tenants as b
             on a.tenant_id = b.tenant_id
where a.tenant_id > 0
and installed = 1
and access_token != ''
and shop_id > 0
group by 1;
`

var query_new_register_tenant = `
select
    tenant_id,
    cast(date(register_time) as varchar) as register_time
from
    platform_offline.dwd_view_analytics_non_testing_tenants
where register_time > utc_date() - interval 15 day
`
