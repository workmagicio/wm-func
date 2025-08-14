package wm_account

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
-- and a.tenant_id in (150083, 133222, 150156, 150158, 150078, 131532, 150077, 134510, 150161, 150178)
`
var query_account_with_platform = `
select
    a.tenant_id,
    a.account_id,
    a.platform,
    a.access_token,
    a.refresh_token,
	cast(json_extract(a.tokens, '$.secretToken') as varchar) as secret_token,
	coalesce(a.cipher, json_extract(account_model, '$.cipher')) as cipher
    from
             platform_offline.
account_connection_unnest_account_level as a
where platform = '%s'
-- and (access_token is not null or refresh_token is not null)
`

var query_account_with_platform_not_null = `
select
    a.tenant_id,
    a.account_id,
    a.platform,
    a.access_token,
    a.refresh_token,
	cast(json_extract(a.tokens, '$.secretToken') as varchar) as secret_token
    from
             platform_offline.
account_connection_unnest_account_level as a
where platform = '%s'
-- and (access_token is not null or refresh_token is not null)
`
