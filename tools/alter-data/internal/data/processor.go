package data

import (
	"fmt"
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/models"
)

// DataProcessor 数据处理器
type DataProcessor struct{}

// NewDataProcessor 创建数据处理器实例
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{}
}

// ExecuteQuery 执行SQL查询获取原始数据
func (p *DataProcessor) ExecuteQuery(sql string) ([]models.AlterData, error) {
	db := platform_db.GetDB()
	var result []models.AlterData

	if err := db.Raw(sql).Limit(-1).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return p.formatAlterData(result), nil
}

// ExecuteQueryForTenant 执行SQL查询获取指定租户的数据
func (p *DataProcessor) ExecuteQueryForTenant(sql string, tenantID int64) ([]models.AlterData, error) {
	db := platform_db.GetDB()
	var result []models.AlterData

	queryWithCondition := sql + " AND TENANT_ID = ?"
	if err := db.Raw(queryWithCondition, tenantID).Limit(-1).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to execute query for tenant %d: %w", tenantID, err)
	}

	return p.formatAlterData(result), nil
}

// GroupByTenant 将原始数据按租户ID分组
func (p *DataProcessor) GroupByTenant(data []models.AlterData) map[int64][]models.AlterData {
	grouped := make(map[int64][]models.AlterData)

	for _, item := range data {
		grouped[item.TenantId] = append(grouped[item.TenantId], item)
	}

	return grouped
}

// formatAlterData 格式化日期数据（只截取日期部分）
func (p *DataProcessor) formatAlterData(data []models.AlterData) []models.AlterData {
	for i := range data {
		if len(data[i].RawDate) > 10 {
			data[i].RawDate = data[i].RawDate[:10]
		}
	}
	return data
}
