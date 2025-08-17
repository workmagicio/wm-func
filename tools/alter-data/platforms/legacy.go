package platforms

import (
	"wm-func/tools/alter-data/models"
)

// GetGoogleData 兼容原有函数，保持向后兼容
func GetGoogleData() []models.AlterData {
	google := &GooglePlatform{}
	data, err := google.GetAllTenantsData(90)
	if err != nil {
		panic(err)
	}

	var result []models.AlterData
	for _, tenantData := range data {
		result = append(result, tenantData...)
	}
	return result
}
