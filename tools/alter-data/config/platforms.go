package config

import "wm-func/tools/alter-data/models"

// GetAvailablePlatforms 获取所有可用平台列表（写死配置）
func GetAvailablePlatforms() []models.PlatformInfo {
	return []models.PlatformInfo{
		{Name: "google", DisplayName: "Google Ads"},
		{Name: "meta", DisplayName: "Meta Ads"},
		{Name: "tiktok", DisplayName: "TikTok Ads"},
		{Name: "pinterest", DisplayName: "Pinterest Ads"},
	}
}

// GetPlatformByName 根据名称获取平台信息
func GetPlatformByName(name string) (models.PlatformInfo, bool) {
	platforms := GetAvailablePlatforms()
	for _, platform := range platforms {
		if platform.Name == name {
			return platform, true
		}
	}
	return models.PlatformInfo{}, false
}

// IsPlatformSupported 检查平台是否支持
func IsPlatformSupported(name string) bool {
	_, exists := GetPlatformByName(name)
	return exists
}
