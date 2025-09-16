package model

import (
	"log"
	"wm-func/common/db/airbyte_db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AirbyteRawData 通用的 Airbyte 原始数据结构
// 所有 Airbyte 表都使用相同的字段结构
type AirbyteRawData struct {
	TenantId            int64  `gorm:"column:wm_tenant_id"`
	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
	AirbyteData         []byte `gorm:"column:_airbyte_data"`
	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
	AirbyteGenerationId int64  `gorm:"column:_airbyte_generation_id"`
}

// AirbyteTable 接口，所有需要保存到 Airbyte 的结构体都需要实现此接口
type AirbyteTable interface {
	TableName() string
}

// SaveAirbyteData 通用的 Airbyte 数据保存方法
// 支持任何实现了 AirbyteTable 接口的数据类型
func SaveAirbyteData[T AirbyteTable](data []T) error {
	if len(data) == 0 {
		return nil
	}

	db := airbyte_db.GetDB()
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(data, 500).Error; err != nil {
		log.Printf("保存 Airbyte 数据失败: %v", err)
		return err
	}
	return nil
}

// SaveAirbyteDataWithCustomDB 使用自定义数据库连接保存 Airbyte 数据
func SaveAirbyteDataWithCustomDB[T AirbyteTable](data []T, db *gorm.DB) error {
	if len(data) == 0 {
		return nil
	}

	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(data, 500).Error; err != nil {
		log.Printf("保存 Airbyte 数据失败: %v", err)
		return err
	}
	return nil
}

// SaveAirbyteDataWithBatchSize 使用自定义批次大小保存 Airbyte 数据
func SaveAirbyteDataWithBatchSize[T AirbyteTable](data []T, batchSize int) error {
	if len(data) == 0 {
		return nil
	}

	db := airbyte_db.GetDB()
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(data, batchSize).Error; err != nil {
		log.Printf("保存 Airbyte 数据失败: %v", err)
		return err
	}
	return nil
}
