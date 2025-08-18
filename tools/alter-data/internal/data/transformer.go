package data

import (
	"sort"
	"strconv"
	"wm-func/tools/alter-data/models"
)

// DataTransformer 数据转换器
type DataTransformer struct{}

// NewDataTransformer 创建数据转换器实例
func NewDataTransformer() *DataTransformer {
	return &DataTransformer{}
}

// TransformToTenantDataList 将原始数据转换为租户数据列表
func (t *DataTransformer) TransformToTenantDataList(platformName string, rawDataMap map[int64][]models.AlterData) []models.TenantData {
	var result []models.TenantData

	for tenantID, rawDataList := range rawDataMap {
		if len(rawDataList) == 0 {
			continue
		}

		tenantData := t.TransformSingleTenantData(platformName, tenantID, rawDataList)
		result = append(result, tenantData)
	}

	// 按差异值排序
	t.sortByDifference(result)

	return result
}

// TransformSingleTenantData 转换单个租户的数据
func (t *DataTransformer) TransformSingleTenantData(platformName string, tenantID int64, rawData []models.AlterData) models.TenantData {
	tenantData := models.TenantData{
		TenantID:   tenantID,
		TenantName: t.generateTenantName(tenantID),
		Platform:   platformName,
		DateRange:  make([]string, 0, len(rawData)),
		APISpend:   make([]int64, 0, len(rawData)),
		AdSpend:    make([]int64, 0, len(rawData)),
		Difference: make([]int64, 0, len(rawData)),
	}

	// 转换数据
	for _, data := range rawData {
		tenantData.DateRange = append(tenantData.DateRange, data.RawDate)
		tenantData.APISpend = append(tenantData.APISpend, data.ApiSpend)
		tenantData.AdSpend = append(tenantData.AdSpend, data.AdSpend)
		tenantData.Difference = append(tenantData.Difference, data.ApiSpend-data.AdSpend)
	}

	return tenantData
}

// generateTenantName 生成租户名称（可以后续从数据库获取真实名称）
func (t *DataTransformer) generateTenantName(tenantID int64) string {
	return "Tenant " + strconv.FormatInt(tenantID, 10)
}

// sortByDifference 按差异值排序租户数据，差异最多的排在前面
func (t *DataTransformer) sortByDifference(tenantDataList []models.TenantData) {
	sort.Slice(tenantDataList, func(i, j int) bool {
		// 计算每个租户的总差异值
		totalDiffI := t.calculateTotalDifference(tenantDataList[i])
		totalDiffJ := t.calculateTotalDifference(tenantDataList[j])

		// 按总差异值降序排序（差异最多的在前面）
		return totalDiffI > totalDiffJ
	})
}

// calculateTotalDifference 计算租户的总差异值
func (t *DataTransformer) calculateTotalDifference(tenantData models.TenantData) int64 {
	var totalDiff int64 = 0

	// 计算所有差异值的绝对值之和
	for _, diff := range tenantData.Difference {
		if diff < 0 {
			totalDiff += -diff // 取绝对值
		} else {
			totalDiff += diff
		}
	}

	return totalDiff
}
