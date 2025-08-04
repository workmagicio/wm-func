package dao

import (
	"fmt"
	"strings"
	"wm-func/common/db/platform_db"
)

type AirbyteRawData struct {
	TenantId  int64  `gorm:"column:tenant_id"`
	RawDate   string `gorm:"column:raw_date"`
	DateCount int64  `gorm:"column:date_count"`
	Avg       int64
}

func GetAirbyteRawDataWithTablName(tableName string) []AirbyteRawData {
	data := query(tableName)
	data = cacAvg(data)
	return nil
}

func cacAvg(data []AirbyteRawData) []AirbyteRawData {
	var tenantAirbyteData = map[int64][]AirbyteRawData{}

	for i, raw := range data {
		tenantAirbyteData[raw.TenantId] = append(tenantAirbyteData[raw.TenantId], data[i])
	}
	fmt.Println(tenantAirbyteData)
	return nil
}

func query(tableName string) []AirbyteRawData {
	result := make([]AirbyteRawData, 0)
	sql := strings.ReplaceAll(query_airbyte_raw_table, "{{tableName}}", tableName)
	db := platform_db.GetDB()
	if err := db.Raw(sql).Scan(&result).Limit(-1).Error; err != nil {
		panic(err)
	}
	return result
}
