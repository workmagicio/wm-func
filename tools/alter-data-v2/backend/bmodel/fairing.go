package bmodel

type WmData struct {
	TenantId int64  `json:"tenant_id"`
	RawDate  string `json:"raw_date"`
	Data     int64  `json:"data"`
}

var fairing_query = `
select
    tenant_id,
    cast(date(response_provided_at) as varchar) as raw_date,
    count(1) as data
from
    platform_offline.dwd_view_post_survey_response_latest
where response_provided_at > utc_timestamp() - interval 90 day
group by 1, 2
`
