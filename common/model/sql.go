package model

var query_all_tenants_id = `
select
    tenant_id
from
    platform_offline.dwd_view_analytics_non_testing_tenants
group by 1
`
