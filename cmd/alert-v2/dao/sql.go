package dao

var query_airbyte_raw_table = `
select
    wm_tenant_id as tenant_id,
	date_format(_airbyte_loaded_at, '%Y-%m-%d') as raw_date,
	count(1) as date_count
from
    airbyte_destination_v2.{{tableName}}
group by 1, 2
`
