package main

const (
	// 平台相关常量
	Platform = "shopify"
	SubType  = "first_order_customers"

	// API 相关常量
	ShopifyAPIVersion = "2023-10"
	MaxPageSize       = 250

	// 任务调度相关常量
	SyncInterval = 60 * 10 // 秒
	MaxWorkers   = 5

	// 日志相关常量
	LogPrefix = "[shopify-customer-first-order]"
)

// ShopifyConfig Shopify API配置
type ShopifyConfig struct {
	APIVersion string
	PageSize   int
}

// getShopifyConfig 获取Shopify配置
func getShopifyConfig() ShopifyConfig {
	return ShopifyConfig{
		APIVersion: ShopifyAPIVersion,
		PageSize:   MaxPageSize,
	}
}

// buildShopifyURL 构建Shopify GraphQL API URL
func buildShopifyURL(shopDomain string) string {
	return "https://" + shopDomain + "/admin/api/" + ShopifyAPIVersion + "/graphql.json"
}
