package platforms

import (
	"fmt"
	"sync"
)

var (
	platformRegistry = make(map[string]Platform)
	registryMutex    sync.RWMutex
)

// RegisterPlatform 注册平台实现
func RegisterPlatform(platform Platform) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	platformRegistry[platform.GetName()] = platform
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
