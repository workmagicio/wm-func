package main

import (
	"gorm.io/gorm/clause"
	"log"
	"wm-func/common/db/airbyte_db"
)

type AirbyteData struct {
	TenantId            int64  `gorm:"column:wm_tenant_id"`
	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
	AirbyteData         []byte `gorm:"column:_airbyte_data"`
	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
	AirbyteGenerationId int64  `gorm:"column:_airbyte_generation_id"`
	ItemType            string `gorm:"-"` // 不映射到数据库，用于确定表名
}

func GetTableNameWithType(subType string) string {
	if subType == SubTypeRequest {
		return "airbyte_destination_v2.raw_fairing_questions"
	}
	return "airbyte_destination_v2.raw_fairing_responses"
}

func SaveToAirbyte(account FAccount, data []AirbyteData, subType string) error {
	if len(data) == 0 {
		return nil
	}

	traceId := account.GetTraceIdWithSubType(subType)

	db := airbyte_db.GetDB()
	table := GetTableNameWithType(subType)
	if err := db.Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "wm_tenant_id"}, {Name: "_airbyte_raw_id"}},
			UpdateAll: true,
		}).CreateInBatches(data, 500).Error; err != nil {
		return err
	}
	log.Printf("[%s] successfully inserted %d knocommerce %s records", traceId, len(data), subType)
	return nil
}
