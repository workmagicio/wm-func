package main

import (
	"log"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"

	"gorm.io/gorm/clause"
)

const (
	STATUS_SUCCESS = "SUCCESS"
	STATUS_RUNNING = "RUNNING"
	STATUS_FAILED  = "FAILED"
)

// 保存fairing数据到数据库
func saveFairingData(account wm_account.Account, data []FairingData, subType string) error {
	if len(data) == 0 {
		return nil
	}

	traceId := getTraceIdWithSubType(account, subType)
	db := airbyte_db.GetDB()

	// 根据数据类型分组保存到不同的表
	questionsData := make([]FairingData, 0)
	responsesData := make([]FairingData, 0)

	for _, item := range data {
		switch item.ItemType {
		case "question":
			questionsData = append(questionsData, item)
		case "response":
			responsesData = append(responsesData, item)
		}
	}

	// 保存 questions 数据
	if len(questionsData) > 0 {
		if err := db.Table("airbyte_destination_v2.raw_fairing_questions").
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "wm_tenant_id"}, {Name: "_airbyte_raw_id"}},
				UpdateAll: true,
			}).CreateInBatches(questionsData, 500).Error; err != nil {
			return err
		}
		log.Printf("[%s] successfully inserted %d fairing question records", traceId, len(questionsData))
	}

	// 保存 responses 数据
	if len(responsesData) > 0 {
		if err := db.Table("airbyte_destination_v2.raw_fairing_responses").
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "wm_tenant_id"}, {Name: "_airbyte_raw_id"}},
				UpdateAll: true,
			}).CreateInBatches(responsesData, 500).Error; err != nil {
			return err
		}
		log.Printf("[%s] successfully inserted %d fairing response records", traceId, len(responsesData))
	}

	log.Printf("[%s] successfully inserted %d fairing records total", traceId, len(data))
	return nil
}

// create table airbyte_destination_v2.raw_fairing_questions
// (
//     wm_tenant_id           bigint                    null,
//     _airbyte_raw_id        varchar(256) charset utf8 null,
//     _airbyte_data          json charset utf8 null,
//     _airbyte_extracted_at  timestamp(6)              null,
//     _airbyte_loaded_at     timestamp(6)              null,
//     _airbyte_meta          json charset utf8 null,
//     _airbyte_generation_id bigint                    null,
//     primary key (wm_tenant_id, _airbyte_raw_id)
// )
//     engine = InnoDB
//     collate = utf8_bin;

// create table airbyte_destination_v2.raw_fairing_responses
// (
//     wm_tenant_id           bigint                    null,
//     _airbyte_raw_id        varchar(256) charset utf8 null,
//     _airbyte_data          json charset utf8 null,
//     _airbyte_extracted_at  timestamp(6)              null,
//     _airbyte_loaded_at     timestamp(6)              null,
//     _airbyte_meta          json charset utf8 null,
//     _airbyte_generation_id bigint                    null,
//     primary key (wm_tenant_id, _airbyte_raw_id)
// )
//     engine = InnoDB
//     collate = utf8_bin;

// type PinAdAnalyticsAirbyteData struct {
// 	TenantId            int64  `gorm:"column:wm_tenant_id"`
// 	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
// 	AirbyteData         string `gorm:"column:_airbyte_data"`
// 	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
// 	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
// 	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
// 	AirbyteGenerationId int    `gorm:"column:_airbyte_generation_id"`
// }

// func (p PinAdAnalyticsAirbyteData) TableName() string {
// 	return "raw_pinterest_ad_analytics"
// }

// dbData = append(dbData, model.PinAdAnalyticsAirbyteData{
// 	TenantId:            r.tenantId,
// 	AirbyteRawId:        analytics[j].GetKey(),
// 	AirbyteData:         string(b),
// 	AirbyteExtractedAt:  time.Now().Format("2006-01-02 15:04:05"),
// 	AirbyteLoadedAt:     time.Now().Format("2006-01-02 15:04:05"),
// 	AirbyteMeta:         `{}`,
// 	AirbyteGenerationId: 0,
// })

// dbClient := airbyte_db.GetDB()
// 	if err := dbClient.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(dbData, 500).Error; err != nil {
// 		panic(err)
// 	}
