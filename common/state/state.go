package state

import (
	"fmt"
	"time"
	"wm-func/common/db/platform_db"

	"gorm.io/gorm"
)

type Records struct {
	TenantId    int64  `gorm:"primaryKey;column:tenant_id"`
	AccountId   string `gorm:"primaryKey;column:account_id"`
	RawPlatform string `gorm:"primaryKey;column:raw_platform"`
	SubType     string `gorm:"primaryKey;column:sub_type"`
	SyncInfo    []byte `gorm:"column:sync_info"`
	CreateTime  string `gorm:"column:create_time"`
	IsRunning   int    `gorm:"column:is_running"` // 0/1  1 is running
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

func (s *SyncInfoResult) TableName() string {
	return "platform_offline.thirds_integration_sync_increment_info"
}

func GetSyncInfo(tenantId int64, accountId, platform, subType string) []byte {
	sql := `select
    sync_info
from platform_offline.thirds_integration_sync_increment_info
where tenant_id = %d
-- and account_id = '%s'
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

// TaskStatus 任务状态枚举
type TaskStatus int

const (
	TaskStatusNotFound       TaskStatus = iota // 没有找到任务
	TaskStatusAlreadyRunning                   // 已经在运行
	TaskStatusAcquired                         // 成功获取到任务
)

// TaskResult 获取任务的结果
type TaskResult struct {
	Status TaskStatus
	Record *Records
}

// GetAvailableTask 获取可执行的任务，使用乐观锁
func GetAvailableTask(tenantId int64, accountId, platform, subType string) TaskResult {
	conn := platform_db.GetDB()

	// 计算1小时前的时间
	oneHourAgo := time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05")

	// 查询符合条件的记录 (is_running = 0 或 create_time > 1小时前)
	var record Records
	err := conn.Where("tenant_id = ? AND account_id = ? AND raw_platform = ? AND sub_type = ? AND (is_running = 0 OR create_time < ?)",
		tenantId, accountId, platform, subType, oneHourAgo).First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return TaskResult{Status: TaskStatusNotFound}
		}
		panic(err)
	}

	// 如果记录的create_time大于1小时且is_running=1，先重置为0
	if record.IsRunning == 1 && record.CreateTime < oneHourAgo {
		// 使用乐观锁先更新为0
		result := conn.Model(&Records{}).
			Where("tenant_id = ? AND account_id = ? AND raw_platform = ? AND sub_type = ? AND is_running = ?",
				tenantId, accountId, platform, subType, 1, record.CreateTime).
			Updates(map[string]interface{}{
				"is_running":  0,
				"create_time": time.Now().Format("2006-01-02 15:04:05"),
			})

		if result.Error != nil {
			panic(result.Error)
		}

		if result.RowsAffected == 0 {
			// 乐观锁失败，说明其他进程已经处理了
			return TaskResult{Status: TaskStatusAlreadyRunning}
		}

		// 更新本地记录状态
		record.IsRunning = 0
		record.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	}

	// 如果当前状态不是0，说明正在运行
	if record.IsRunning != 0 {
		return TaskResult{Status: TaskStatusAlreadyRunning}
	}

	// 使用乐观锁更新状态为运行中
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	result := conn.Model(&Records{}).
		Where("tenant_id = ? AND account_id = ? AND raw_platform = ? AND sub_type = ? AND is_running = ?",
			tenantId, accountId, platform, subType, record.IsRunning).
		Updates(map[string]interface{}{
			"is_running":  1,
			"create_time": currentTime,
		})

	if result.Error != nil {
		panic(result.Error)
	}

	if result.RowsAffected == 0 {
		// 乐观锁失败，说明其他进程已经获取了这个任务
		return TaskResult{Status: TaskStatusAlreadyRunning}
	}

	// 更新本地记录
	record.IsRunning = 1
	record.CreateTime = currentTime

	return TaskResult{
		Status: TaskStatusAcquired,
		Record: &record,
	}
}

// SetRunning 设置任务为运行状态
func SetRunning(tenantId int64, accountId, platform, subType string) bool {
	conn := platform_db.GetDB()
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	result := conn.Model(&Records{}).
		Where("tenant_id = ? AND account_id = ? AND raw_platform = ? AND sub_type = ?",
			tenantId, accountId, platform, subType).
		Updates(map[string]interface{}{
			"is_running":  1,
			"create_time": currentTime,
		})

	if result.Error != nil {
		panic(result.Error)
	}

	return result.RowsAffected > 0
}

// SetStop 设置任务为停止状态
func SetStop(tenantId int64, accountId, platform, subType string) bool {
	conn := platform_db.GetDB()
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	result := conn.Model(&Records{}).
		Where("tenant_id = ? AND account_id = ? AND raw_platform = ? AND sub_type = ?",
			tenantId, accountId, platform, subType).
		Updates(map[string]interface{}{
			"is_running":  0,
			"create_time": currentTime,
		})

	if result.Error != nil {
		panic(result.Error)
	}

	return result.RowsAffected > 0
}
