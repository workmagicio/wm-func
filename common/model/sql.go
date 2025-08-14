package model

var query_all_tenants_id = `
select
    tenant_id
from
    platform_offline.dwd_view_analytics_non_testing_tenants
group by 1
`

var query_all_tenants_id_with_platform = `
select
    tenant_id, platform
from
    platform_offline.account_connection_unnest_account_level_with_no_testing
group by 1, 2
`
