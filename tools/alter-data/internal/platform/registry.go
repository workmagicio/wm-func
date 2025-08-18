package platform

import (
	"fmt"
	"sync"
	"wm-func/tools/alter-data/internal/config"
)

var (
	platformRegistry = make(map[string]Platform)
	registryMutex    sync.RWMutex
)

// InitializePlatforms 初始化所有平台
func InitializePlatforms() error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	// 清空现有注册
	platformRegistry = make(map[string]Platform)

	// 获取所有平台配置并注册
	platformConfigs := config.GetAllPlatformConfigs()
	for _, platformConfig := range platformConfigs {
		platform := NewBasePlatform(platformConfig)
		platformRegistry[platformConfig.Name] = platform
	}

	return nil
}

// RegisterPlatform 注册平台实现
func RegisterPlatform(platform Platform) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	info := platform.GetInfo()
	platformRegistry[info.Name] = platform
}

// GetPlatform 获取指定平台实现
func GetPlatform(name string) (Platform, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	platform, exists := platformRegistry[name]
	if !exists {
		return nil, fmt.Errorf("platform implementation for %s not found", name)
	}
	return platform, nil
}

// IsImplemented 检查平台是否已实现
func IsImplemented(name string) bool {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	_, exists := platformRegistry[name]
	return exists
}

// GetImplementedPlatformNames 获取所有已实现的平台名称
func GetImplementedPlatformNames() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	var names []string
	for name := range platformRegistry {
		names = append(names, name)
	}
	return names
}

// GetEnabledPlatforms 获取所有启用的平台
func GetEnabledPlatforms() []Platform {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	var platforms []Platform
	for _, platform := range platformRegistry {
		if platform.IsEnabled() {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}
