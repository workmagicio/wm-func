package main

import (
	"encoding/json"
	"fmt"
	"time"
	"wm-func/common/db/platform_db"
)

type ReportDate struct {
	Date string `gorm:"column:last_sync_date"`
}

func GetSyncInfo(tenantId int64, accountId string, platform string) string {
	sql := `select
    cast(json_extract(sync_info, '$.report_date') as varchar) as last_sync_date
from platform_offline.thirds_integration_sync_increment_info
where tenant_id = %d
and account_id = '%s'
and raw_platform = '%s'
limit 1`
	sql = fmt.Sprintf(sql, tenantId, accountId, platform)

	res := []ReportDate{}
	client := platform_db.GetDB()
	if err := client.Raw(sql).Scan(&res).Error; err != nil {
		panic(err)
	}
	if len(res) == 0 {
		return ""
	}
	return res[0].Date
}

type SyncInfo struct {
	ReportDate string `json:"report_date"`
	Success    bool   `json:"success"`
	Error      string `json:"error"`
}

type Records struct {
	TenantId    int64  `gorm:"primaryKey;column:tenant_id"`
	AccountId   string `gorm:"primaryKey;column:account_id"`
	RawPlatform string `gorm:"primaryKey;column:raw_platform"`
	SubType     string `gorm:"primaryKey;column:sub_type"`
	SyncInfo    []byte `gorm:"column:sync_info"`
	CreateTime  string `gorm:"column:create_time"`
}

func (r Records) TableName() string {
	return "thirds_integration_sync_increment_info"
}

func SaveSyncInfoWithTime(tenantId int64, accountId string, platform string, point time.Time) {
	info := SyncInfo{ReportDate: point.Format("2006-01-02"), Success: true}

	b, _ := json.Marshal(info)
	record := Records{
		TenantId:    tenantId,
		AccountId:   accountId,
		RawPlatform: platform,
		SubType:     "-",
		SyncInfo:    b,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
	}
	Save(record)
}

func Save(records Records) {
	conn := platform_db.GetDB()
	if err := conn.Save(records).Error; err != nil {
		panic(err)
	}
}
