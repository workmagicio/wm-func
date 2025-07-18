package state

import (
	"fmt"
	"gorm.io/gorm"
	"time"
	"wm-func/common/db/platform_db"
)

type Records struct {
	TenantId    int64  `gorm:"primaryKey;column:tenant_id"`
	AccountId   string `gorm:"primaryKey;column:account_id"`
	RawPlatform string `gorm:"primaryKey;column:raw_platform"`
	SubType     string `gorm:"primaryKey;column:sub_type"`
	SyncInfo    []byte `gorm:"column:sync_info"`
	CreateTime  string `gorm:"column:create_time"`
}

func (r Records) TableName() string {
	return "platform_offline.thirds_integration_sync_increment_info"
}

func Save(records Records) {
	conn := platform_db.GetDB()
	if err := conn.Save(records).Error; err != nil {
		panic(err)
	}
}

func SaveSyncInfo(tenantId int64, accountId string, platform string, subtype string, info []byte) {
	record := Records{
		TenantId:    tenantId,
		AccountId:   accountId,
		RawPlatform: platform,
		SubType:     subtype,
		SyncInfo:    info,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
	}
	Save(record)
}

type ReportDate struct {
	Date string `gorm:"column:last_sync_date"`
}

// SyncInfoResult 用于接收单个 sync_info 字段的查询结果
type SyncInfoResult struct {
	SyncInfo []byte `gorm:"column:sync_info"`
}

func GetSyncInfo(tenantId int64, accountId, platform, subType string) []byte {
	sql := `select
    sync_info
from platform_offline.thirds_integration_sync_increment_info
where tenant_id = %d
and account_id = '%s'
and raw_platform = '%s'
and sub_type = '%s'
`
	sql = fmt.Sprintf(sql, tenantId, accountId, platform, subType)

	var result SyncInfoResult
	client := platform_db.GetDB()
	if err := client.Raw(sql).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		panic(err)
	}
	if len(result.SyncInfo) == 0 {
		return nil
	}
	return result.SyncInfo
}
