package models

import "time"

// AlterData 原始数据模型 (数据库查询结果)
type AlterData struct {
	TenantId int64  `gorm:"column:tenant_id"`
	Platform string `gorm:"column:platform"` // 新增平台字段，用于跨平台查询
	RawDate  string `gorm:"column:raw_date"`
	ApiSpend int64  `gorm:"column:api_spend"`
	AdSpend  int64  `gorm:"column:ad_spend"`
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	Name        string `json:"name"`         // 平台标识: google, meta, tiktok
	DisplayName string `json:"display_name"` // 显示名称: Google Ads, Meta Ads
}

// TenantData 租户数据 (API响应格式)
type TenantData struct {
	TenantID   int64    `json:"tenant_id"`
	TenantName string   `json:"tenant_name"`
	Platform   string   `json:"platform"`
	DateRange  []string `json:"date_range"`
	APISpend   []int64  `json:"api_spend"`
	AdSpend    []int64  `json:"ad_spend"`
	Difference []int64  `json:"difference"`
}

// APIResponse 通用API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PlatformResponse 平台列表响应
type PlatformResponse struct {
	Success bool           `json:"success"`
	Data    []PlatformInfo `json:"data"`
	Message string         `json:"message"`
}

// DashboardResponse 仪表板数据响应
type DashboardResponse struct {
	Success   bool         `json:"success"`
	Platform  string       `json:"platform"`
	Data      []TenantData `json:"data"`
	Message   string       `json:"message"`
	CacheInfo *CacheInfo   `json:"cache_info,omitempty"` // 缓存信息
}

// CacheInfo 缓存信息
type CacheInfo struct {
	Platform  string    `json:"platform"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsExpired bool      `json:"is_expired"`
	DataCount int       `json:"data_count"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalItems   int           `json:"total_items"`
	ExpiredItems int           `json:"expired_items"`
	ValidItems   int           `json:"valid_items"`
	CacheDir     string        `json:"cache_dir"`
	TTL          time.Duration `json:"ttl"`
}

// TenantInfo 租户信息
type TenantInfo struct {
	TenantID   int64  `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	Success bool         `json:"success"`
	Data    []TenantInfo `json:"data"`
	Message string       `json:"message"`
}

// CrossPlatformData 跨平台数据 (按平台组织的租户数据)
type CrossPlatformData struct {
	TenantID     int64                   `json:"tenant_id"`
	TenantName   string                  `json:"tenant_name"`
	PlatformData map[string][]TenantData `json:"platform_data"` // key: platform name, value: tenant data for that platform
}

// TenantCrossPlatformResponse 租户跨平台数据响应
type TenantCrossPlatformResponse struct {
	Success    bool              `json:"success"`
	TenantID   int64             `json:"tenant_id"`
	TenantName string            `json:"tenant_name"`
	Data       CrossPlatformData `json:"data"`
	Message    string            `json:"message"`
	CacheInfo  *CacheInfo        `json:"cache_info,omitempty"`
}
