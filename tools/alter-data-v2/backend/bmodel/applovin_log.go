package bmodel

var query_applovin_log = `
select
	tenant_id,
	cast(date(src_event_time) as varchar) as raw_date,
	count(1) as data
from
	platform_offline.dwd_attr_3p_ref_order_join_source_v20250721jz
where src_source = 'Applovin'
  and date(src_event_time) > utc_date() - interval 90 day
group by 1, 2
`
