package platforms

import (
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/models"
	"wm-func/tools/alter-data/query_sql"
)

// MetaPlatform Meta Ads 平台实现
type MetaPlatform struct{}

// GetName 返回平台标识
func (m *MetaPlatform) GetName() string {
	return "meta"
}

// GetDisplayName 返回平台显示名称
func (m *MetaPlatform) GetDisplayName() string {
	return "Meta Ads"
}

// GetTenantData 获取指定租户数据
func (m *MetaPlatform) GetTenantData(tenantID int64, days int) ([]models.AlterData, error) {
	exec_sql := query_sql.Query_meta_api_with_overview
	db := platform_db.GetDB()
	res := []models.AlterData{}

	if err := db.Raw(exec_sql+" AND TENANT_ID = ?", tenantID).Limit(-1).Scan(&res).Error; err != nil {
		return nil, err
	}
	return formatAlterData(res), nil
}

// GetAllTenantsData 获取所有租户数据
func (m *MetaPlatform) GetAllTenantsData(days int) (map[int64][]models.AlterData, error) {
	exec_sql := query_sql.Query_meta_api_with_overview
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
