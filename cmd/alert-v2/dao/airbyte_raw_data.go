package dao

import (
	"fmt"
	"gorm.io/gorm"
	"sort"
	"strings"
	"time"
)

type AirbyteRawData struct {
	TenantId  int64  `gorm:"column:tenant_id"`
	RawDate   string `gorm:"column:raw_date"` // 2025-07-15
	DateCount int64  `gorm:"column:date_count"`
	Avg       int64
}

func GetAirbyteRawDataWithTablName(tableName string, fields string, db *gorm.DB) []AirbyteRawData {
	data := query(tableName, fields, db)
	data = checkDateContinuity(tableName, data)
	return nil
}

func checkDateContinuity(tableName string, data []AirbyteRawData) []AirbyteRawData {
	// 根据表名设置起始日期
	var startDate time.Time
	if tableName == "raw_tiktok_marketing_gmv_max_metrics" {
		// 从8月1号开始
		startDate = time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	}
	// 从90天前开始
	if startDate.Before(time.Now().UTC().Add(time.Hour * 24 * 90 * -1)) {
		startDate = time.Now().AddDate(0, 0, -90)
	}

	var tenantAirbyteData = map[int64][]AirbyteRawData{}

	// 过滤数据，只保留起始日期之后的数据
	for i, raw := range data {
		rawDate, err := time.Parse("2006-01-02", raw.RawDate)
		if err != nil {
			fmt.Printf("解析日期失败: %s, %v\n", raw.RawDate, err)
			continue
		}

		if rawDate.After(startDate) || rawDate.Equal(startDate) {
			tenantAirbyteData[raw.TenantId] = append(tenantAirbyteData[raw.TenantId], data[i])
		}
	}
	// 将数组中的内容排序，按照raw_date排序
	for _, raw := range tenantAirbyteData {
		sort.Slice(raw, func(i, j int) bool {
			return raw[i].RawDate < raw[j].RawDate
		})
	}

	// 判断是否存在不连续的日期
	for tenantId, rawList := range tenantAirbyteData {
		if len(rawList) <= 1 {
			continue
		}

		// 当表名为 raw_tiktok_marketing_gmv_max_metrics 时，排除特定租户
		if tableName == "raw_tiktok_marketing_gmv_max_metrics" {
			excludeTenants := []int64{150101, 150015}
			skip := false
			for _, excludeId := range excludeTenants {
				if tenantId == excludeId {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		for i := 1; i < len(rawList); i++ {
			prevDate, err := time.Parse("2006-01-02", rawList[i-1].RawDate)
			if err != nil {
				fmt.Printf("解析日期失败: %s, %v\n", rawList[i-1].RawDate, err)
				continue
			}

			currentDate, err := time.Parse("2006-01-02", rawList[i].RawDate)
			if err != nil {
				fmt.Printf("解析日期失败: %s, %v\n", rawList[i].RawDate, err)
				continue
			}

			// 计算日期差异
			diff := currentDate.Sub(prevDate).Hours() / 24
			if diff > 1 {
				fmt.Printf("tableName: %s 租户 %d 存在不连续日期: %s -> %s (间隔 %.0f 天)\n", tableName,
					tenantId, rawList[i-1].RawDate, rawList[i].RawDate, diff-1)
			}
		}
	}

	return nil
}

func query(tableName string, fields string, db *gorm.DB) []AirbyteRawData {
	result := make([]AirbyteRawData, 0)
	sql := strings.ReplaceAll(query_airbyte_raw_table, "{{tableName}}", tableName)
	sql = strings.ReplaceAll(sql, "{{fields}}", fields)

	fmt.Println(sql)
	if err := db.Raw(sql).Scan(&result).Limit(-1).Error; err != nil {
		panic(err)
	}
	return result
}
