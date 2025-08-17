package platform

import "wm-func/tools/alter-data/models"

// Platform 平台接口定义
type Platform interface {
	// GetInfo 获取平台基本信息
	GetInfo() PlatformInfo

	// GetTenantData 获取指定租户数据
	GetTenantData(tenantID int64, days int) ([]models.AlterData, error)

	// GetAllTenantsData 获取所有租户数据
	GetAllTenantsData(days int) (map[int64][]models.AlterData, error)

	// IsEnabled 检查平台是否启用
	IsEnabled() bool
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	Name        string `json:"name"`         // 平台标识
	DisplayName string `json:"display_name"` // 显示名称
	QueryKey    string `json:"query_key"`    // SQL查询键
	Enabled     bool   `json:"enabled"`      // 是否启用
	Description string `json:"description"`  // 平台描述
}
