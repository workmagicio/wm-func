package platforms

import "wm-func/tools/alter-data/models"

// Platform 平台接口定义
type Platform interface {
	// 获取平台基本信息
	GetName() string
	GetDisplayName() string

	// 获取数据（返回原始数据格式）
	GetTenantData(tenantID int64, days int) ([]models.AlterData, error)
	GetAllTenantsData(days int) (map[int64][]models.AlterData, error)
}
