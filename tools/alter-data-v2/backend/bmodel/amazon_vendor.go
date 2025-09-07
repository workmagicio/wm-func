package bmodel

var query_amazon_vonder = `
select
	tenant_id,
	cast(stat_date as varchar) as raw_date,
	sum(shipped_units) as data
from
	platform_offline.amazon_vendor_zip_code_daily_report
where stat_date > utc_date() - interval 90 day
group by 1, 2
`
