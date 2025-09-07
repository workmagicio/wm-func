package controller

import (
	"time"
	"wm-func/tools/alter-data-v2/backend/cac"
)

type TenantData struct {
	TenantId      int64              `json:"tenant_id"`
	Last30DayDiff int64              `json:"last_30_day_diff"`
	RegisterTime  string             `json:"register_time"`
	DateSequence  []cac.DateSequence `json:"date_sequence"`
	Tags          []string           `json:"tags"`
}

type AllTenantData struct {
	NewTenants       []TenantData `json:"new_tenants"`
	OldTenants       []TenantData `json:"old_tenants"`
	DataLastLoadTime time.Time    `json:"data_last_load_time"`
	DataType         string       `json:"data_type"`      // 数据类型：dual_source 或 wm_only
	LastDataDate     string       `json:"last_data_date"` // 最后有数据的日期（仅WM-only类型使用）
}
