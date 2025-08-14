package account

var query_need_sync_account = `with tenant as (select a.tenant_id,
                       cast(json_extract(a.accounts, '$.0.id') as varchar )      as account_id,
                       cast(json_extract(tokens, '$.accessToken') as varchar)  as access_token,
                       cast(json_extract(tokens, '$.refreshToken') as varchar) as refresh_token
                from platform.account_connection as a
                         join platform_offline.dwd_view_analytics_key_tenant_connections as b
                              on a.tenant_id = b.tenant_id and a.platform = b.raw_platform and b.connected = 1
                where platform = '%s'),
     res as (select tenant.*, ii.sync_info
             from tenant
                      left join platform_offline.thirds_integration_sync_increment_info ii
                                on tenant.tenant_id = ii.tenant_id and ii.raw_platform = '%s'
                                    and tenant.account_id = ii.account_id and
                                   date(cast(json_extract(sync_info, '$.report_date') as varchar)) = date(now()))

select tenant_id, account_id, refresh_token, access_token
from res
where sync_info is null
`
