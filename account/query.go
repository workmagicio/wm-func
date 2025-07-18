package account

var query_shopify_accounts = `
select
    a.tenant_id,
    shop_domain,
    access_token
from
    platform.shopify_connection as a
join platform_offline.dwd_view_analytics_non_testing_tenants as b
on a.tenant_id = b.tenant_id
where a.tenant_id > 0
and access_token != ''
and connected > 0
`
