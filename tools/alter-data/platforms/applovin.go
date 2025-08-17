package platforms

import (
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/models"
	"wm-func/tools/alter-data/query_sql"
)

// AppLovinPlatform AppLovin 平台实现
type AppLovinPlatform struct{}

// GetName 返回平台标识
func (a *AppLovinPlatform) GetName() string {
	return "applovin"
}

// GetDisplayName 返回平台显示名称
func (a *AppLovinPlatform) GetDisplayName() string {
	return "AppLovin"
}

// GetTenantData 获取指定租户数据
func (a *AppLovinPlatform) GetTenantData(tenantID int64, days int) ([]models.AlterData, error) {
	exec_sql := query_sql.Query_applovin_api_with_overview
	db := platform_db.GetDB()
	res := []models.AlterData{}

	if err := db.Raw(exec_sql+" AND TENANT_ID = ?", tenantID).Limit(-1).Scan(&res).Error; err != nil {
		return nil, err
	}
	return formatAlterData(res), nil
}

// GetAllTenantsData 获取所有租户数据
func (a *AppLovinPlatform) GetAllTenantsData(days int) (map[int64][]models.AlterData, error) {
	exec_sql := query_sql.Query_applovin_api_with_overview
	db := platform_db.GetDB()
	res := []models.AlterData{}

	if err := db.Raw(exec_sql).Limit(-1).Scan(&res).Error; err != nil {
		return nil, err
	}

	formatted := formatAlterData(res)

	// 按租户ID分组
	tenantData := make(map[int64][]models.AlterData)
	for _, data := range formatted {
		tenantData[data.TenantId] = append(tenantData[data.TenantId], data)
	}

	return tenantData, nil
}
