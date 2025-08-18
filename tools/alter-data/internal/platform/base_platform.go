package platform

import (
	"fmt"
	"wm-func/tools/alter-data/internal/config"
	"wm-func/tools/alter-data/internal/data"
	"wm-func/tools/alter-data/models"
)

// BasePlatform 基础平台实现
type BasePlatform struct {
	config    config.PlatformConfig
	processor *data.DataProcessor
}

// NewBasePlatform 创建基础平台实例
func NewBasePlatform(platformConfig config.PlatformConfig) *BasePlatform {
	return &BasePlatform{
		config:    platformConfig,
		processor: data.NewDataProcessor(),
	}
}

// GetInfo 获取平台基本信息
func (p *BasePlatform) GetInfo() PlatformInfo {
	return PlatformInfo{
		Name:        p.config.Name,
		DisplayName: p.config.DisplayName,
		QueryKey:    p.config.QueryKey,
		Enabled:     p.config.Enabled,
		Description: p.config.Description,
	}
}

// GetTenantData 获取指定租户数据
func (p *BasePlatform) GetTenantData(tenantID int64, days int) ([]models.AlterData, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("platform %s is not enabled", p.config.Name)
	}

	sql, exists := config.GetQuerySQL(p.config.QueryKey)
	if !exists {
		return nil, fmt.Errorf("query configuration not found for platform %s", p.config.Name)
	}

	return p.processor.ExecuteQueryForTenant(sql, tenantID)
}

// GetAllTenantsData 获取所有租户数据
func (p *BasePlatform) GetAllTenantsData(days int) (map[int64][]models.AlterData, error) {
	if !p.config.Enabled {
		return nil, fmt.Errorf("platform %s is not enabled", p.config.Name)
	}

	sql, exists := config.GetQuerySQL(p.config.QueryKey)
	if !exists {
		return nil, fmt.Errorf("query configuration not found for platform %s", p.config.Name)
	}

	rawData, err := p.processor.ExecuteQuery(sql)
	if err != nil {
		return nil, err
	}

	return p.processor.GroupByTenant(rawData), nil
}

// IsEnabled 检查平台是否启用
func (p *BasePlatform) IsEnabled() bool {
	return p.config.Enabled
}
