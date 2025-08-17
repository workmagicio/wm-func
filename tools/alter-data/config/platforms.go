package config

import (
	"fmt"
	"wm-func/tools/alter-data/models"
)

// GetAvailablePlatforms 获取所有可用平台列表（写死配置）
func GetAvailablePlatforms() []models.PlatformInfo {
	platforms := []models.PlatformInfo{
		{Name: "google", DisplayName: "Google Ads"},
		{Name: "meta", DisplayName: "Meta Ads"},
		{Name: "applovin", DisplayName: "AppLovin"},
		{Name: "tiktok", DisplayName: "TikTok Ads"},
		{Name: "pinterest", DisplayName: "Pinterest Ads"},
	}
	fmt.Printf("🔍 DEBUG: GetAvailablePlatforms returning %d platforms\n", len(platforms))
	for i, p := range platforms {
		fmt.Printf("  %d: %s -> %s\n", i+1, p.Name, p.DisplayName)
	}
	return platforms
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
