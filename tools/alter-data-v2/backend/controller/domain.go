package controller

import "wm-func/tools/alter-data-v2/backend/cac"

type TenantData struct {
	TenantId      int64              `json:"tenant_id"`
	Last30DayDiff int64              `json:"last_30_day_diff"`
	DateSequence  []cac.DateSequence `json:"date_sequence"`
	Tags          []string           `json:"tags"`
}

type AllTenantData struct {
	NewTenants []TenantData `json:"new_tenants"`
	OldTenants []TenantData `json:"old_tenants"`
}
