package config

import "wm-func/tools/alter-data/models"

// PlatformConfig 平台配置信息
type PlatformConfig struct {
	Name        string `json:"name"`         // 平台标识
	DisplayName string `json:"display_name"` // 显示名称
	QueryKey    string `json:"query_key"`    // SQL查询键
	Enabled     bool   `json:"enabled"`      // 是否启用
	Description string `json:"description"`  // 平台描述
}

// platformConfigs 平台配置列表
var platformConfigs = []PlatformConfig{
	{
		Name:        "google",
		DisplayName: "Google Ads",
		QueryKey:    "google_ads_query",
		Enabled:     true,
		Description: "Google广告平台数据",
	},
	{
		Name:        "meta",
		DisplayName: "Meta Ads",
		QueryKey:    "meta_ads_query",
		Enabled:     true,
		Description: "Meta广告平台数据",
	},
	{
		Name:        "applovin",
		DisplayName: "AppLovin",
		QueryKey:    "applovin_ads_query",
		Enabled:     true,
		Description: "AppLovin广告平台数据",
	},
	{
		Name:        "tiktok",
		DisplayName: "TikTok Ads",
		QueryKey:    "tiktok_ads_query",
		Enabled:     false,
		Description: "TikTok广告平台数据（待实现）",
	},
	{
		Name:        "pinterest",
		DisplayName: "Pinterest Ads",
		QueryKey:    "pinterest_ads_query",
		Enabled:     false,
		Description: "Pinterest广告平台数据（待实现）",
	},
}

// GetAllPlatformConfigs 获取所有平台配置
func GetAllPlatformConfigs() []PlatformConfig {
	return platformConfigs
}

// GetEnabledPlatformConfigs 获取启用的平台配置
func GetEnabledPlatformConfigs() []PlatformConfig {
	var enabled []PlatformConfig
	for _, config := range platformConfigs {
		if config.Enabled {
			enabled = append(enabled, config)
		}
	}
	return enabled
}

// GetPlatformConfig 根据名称获取平台配置
func GetPlatformConfig(name string) (PlatformConfig, bool) {
	for _, config := range platformConfigs {
		if config.Name == name {
			return config, true
		}
	}
	return PlatformConfig{}, false
}

// IsPlatformSupported 检查平台是否支持
func IsPlatformSupported(name string) bool {
	_, exists := GetPlatformConfig(name)
	return exists
}

// IsPlatformEnabled 检查平台是否启用
func IsPlatformEnabled(name string) bool {
	config, exists := GetPlatformConfig(name)
	return exists && config.Enabled
}

// GetAvailablePlatforms 获取可用平台信息（用于API响应）
func GetAvailablePlatforms() []models.PlatformInfo {
	configs := GetEnabledPlatformConfigs()
	platforms := make([]models.PlatformInfo, 0, len(configs))

	for _, config := range configs {
		platforms = append(platforms, models.PlatformInfo{
			Name:        config.Name,
			DisplayName: config.DisplayName,
		})
	}

	return platforms
}
