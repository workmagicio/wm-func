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
	TenantID     int64  `json:"tenant_id"`
	TenantName   string `json:"tenant_name"`
	RegisterTime string `json:"register_time,omitempty"` // 注册时间，用于最近注册租户
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	Success bool         `json:"success"`
	Data    []TenantInfo `json:"data"`
	Message string       `json:"message"`
}

// RecentTenantsResponse 最近注册租户响应
type RecentTenantsResponse struct {
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

// TenantAccessRecord 租户访问记录
type TenantAccessRecord struct {
	TenantID    int64     `json:"tenant_id"`
	TenantName  string    `json:"tenant_name"`
	AccessCount int       `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	FirstAccess time.Time `json:"first_access"`
}

// FrequentTenantsResponse 经常访问租户响应
type FrequentTenantsResponse struct {
	Success bool                 `json:"success"`
	Data    []TenantAccessRecord `json:"data"`
	Message string               `json:"message"`
}

// === 归因订单分析相关数据模型 ===

// AttributionOrderRawData 归因订单原始数据模型 (数据库查询结果)
type AttributionOrderRawData struct {
	TenantId    int64  `gorm:"column:tenant_id"`
	EventDate   string `gorm:"column:event_date"`
	AdsPlatform string `gorm:"column:ads_platform"`
	AttrOrders  int64  `gorm:"column:attr_orders"`
}

// AttributionOrderData 归因订单租户数据 (API响应格式)
type AttributionOrderData struct {
	TenantID     int64              `json:"tenant_id"`
	TenantName   string             `json:"tenant_name"`
	DateRange    []string           `json:"date_range"`
	PlatformData map[string][]int64 `json:"platform_data"` // platform -> orders array
	Platforms    []string           `json:"platforms"`     // 该tenant涉及的所有平台
	TotalOrders  map[string]int64   `json:"total_orders"`  // 每个平台的订单总数
	HasConcave   bool               `json:"has_concave"`   // 是否存在凹字形异常
	ConcaveCount int                `json:"concave_count"` // 凹字形异常数量
}

// AttributionOrderResponse 归因订单分析响应
type AttributionOrderResponse struct {
	Success   bool                   `json:"success"`
	Data      []AttributionOrderData `json:"data"`
	Message   string                 `json:"message"`
	CacheInfo *CacheInfo             `json:"cache_info,omitempty"`
}

// === Amazon订单分析相关数据模型 ===

// AmazonOrderRawData Amazon订单原始数据模型 (数据库查询结果)
type AmazonOrderRawData struct {
	TenantId int64  `gorm:"column:tenant_id"`
	StatDate string `gorm:"column:stat_date"`
	Orders   int64  `gorm:"column:orders"`
}

// AmazonOrderData Amazon订单租户数据 (API响应格式)
type AmazonOrderData struct {
	TenantID        int64    `json:"tenant_id"`
	TenantName      string   `json:"tenant_name"`
	DateRange       []string `json:"date_range"`
	OrdersData      []int64  `json:"orders_data"`      // 订单数量数组
	DailyAverage    float64  `json:"daily_average"`    // 90天日平均值
	WarningLevel    string   `json:"warning_level"`    // 预警级别：normal, warning, critical
	TotalOrders     int64    `json:"total_orders"`     // 总订单数
	ZeroDaysCount   int      `json:"zero_days_count"`  // 掉0天数
	ConcaveCount    int      `json:"concave_count"`    // 凹形问题数量
	HasAnomalies    bool     `json:"has_anomalies"`    // 是否存在异常
	ProcessedOrders []int64  `json:"processed_orders"` // 处理后的订单数（异常标记）
	CacheTimestamp  int64    `json:"cache_timestamp"`  // 缓存时间戳
}

// AmazonOrderResponse Amazon订单分析响应
type AmazonOrderResponse struct {
	Success   bool              `json:"success"`
	Data      []AmazonOrderData `json:"data"`
	Message   string            `json:"message"`
	CacheInfo *CacheInfo        `json:"cache_info,omitempty"`
}

// === Fairing分析相关数据模型 ===

// FairingRawData Fairing原始数据模型 (数据库查询结果)
type FairingRawData struct {
	TenantId int64  `gorm:"column:tenant_id"`
	StatDate string `gorm:"column:stat_date"`
	Cnt      int64  `gorm:"column:cnt"`
}

// FairingData Fairing租户数据 (API响应格式)
type FairingData struct {
	TenantID           int64    `json:"tenant_id"`
	TenantName         string   `json:"tenant_name"`
	DateRange          []string `json:"date_range"`
	ResponseData       []int64  `json:"response_data"`       // 响应数量数组
	DailyAverage       float64  `json:"daily_average"`       // 90天日平均值
	WarningLevel       string   `json:"warning_level"`       // 预警级别：normal, warning, critical
	TotalResponses     int64    `json:"total_responses"`     // 总响应数
	ZeroDaysCount      int      `json:"zero_days_count"`     // 掉0天数
	ConcaveCount       int      `json:"concave_count"`       // 凹形问题数量
	HasAnomalies       bool     `json:"has_anomalies"`       // 是否存在异常
	ProcessedResponses []int64  `json:"processed_responses"` // 处理后的响应数（异常标记）
	CacheTimestamp     int64    `json:"cache_timestamp"`     // 缓存时间戳
}

// FairingResponse Fairing分析响应
type FairingResponse struct {
	Success   bool          `json:"success"`
	Data      []FairingData `json:"data"`
	Message   string        `json:"message"`
	CacheInfo *CacheInfo    `json:"cache_info,omitempty"`
}
