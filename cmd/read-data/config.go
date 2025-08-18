package main

// 配置常量
const (
	MODEL_NAME   = "gemini-2.5-pro-preview-05-06"
	MAX_ROWS     = 20 // 传给AI分析的数据行数
	PREVIEW_ROWS = 50 // 用于预览的数据行数
)

var requered_fields_map = map[string]bool{
	"date_type":           true,
	"date_code":           true,
	"geo_type":            true,
	"geo_code":            true,
	"sales_platform":      true,
	"sales_platform_type": true,
	"country_code":        true,
	"order_id":            false, // 可选字段，仅在需要精确订单统计时配置
	"orders":              true,
	"sales":               true,
}
