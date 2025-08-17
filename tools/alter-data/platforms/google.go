package platforms

import (
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/models"
	"wm-func/tools/alter-data/query_sql"
)

// GooglePlatform Google Ads 平台实现
type GooglePlatform struct{}

// GetName 返回平台标识
func (g *GooglePlatform) GetName() string {
	return "google"
}

// GetDisplayName 返回平台显示名称
func (g *GooglePlatform) GetDisplayName() string {
	return "Google Ads"
}

// GetTenantData 获取指定租户数据
func (g *GooglePlatform) GetTenantData(tenantID int64, days int) ([]models.AlterData, error) {
	exec_sql := query_sql.Query_google_api_with_overview
	db := platform_db.GetDB()
	res := []models.AlterData{}

	if err := db.Raw(exec_sql+" AND TENANT_ID = ?", tenantID).Limit(-1).Scan(&res).Error; err != nil {
		return nil, err
	}
	return formatAlterData(res), nil
}

// GetAllTenantsData 获取所有租户数据
func (g *GooglePlatform) GetAllTenantsData(days int) (map[int64][]models.AlterData, error) {
	exec_sql := query_sql.Query_google_api_with_overview
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

// formatAlterData 格式化日期数据（只截取日期部分）
func formatAlterData(data []models.AlterData) []models.AlterData {
	for i := range data {
		if len(data[i].RawDate) > 10 {
			data[i].RawDate = data[i].RawDate[:10]
		}
	}
	return data
}
