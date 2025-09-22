package main

import (
	"gorm.io/gorm/clause"
	"log"
	"wm-func/common/db/platform_db"
)

type OrderJoinSource struct {
	TenantId      int64  `gorm:"column:tenant_id"`
	ImportingType string `gorm:"column:importing_type"`
	OrderId       string `gorm:"column:order_id"`
	SrcEntityType string `gorm:"column:src_entity_type"`
	SrcEntityId   string `gorm:"column:src_entity_id"`
	SrcEventTime  string `gorm:"column:src_event_time"`
	SrcChannel    string `gorm:"column:src_channel"`
	SrcSource     string `gorm:"column:src_source"`
	SrcAdId       string `gorm:"column:src_ad_id"`
	SrcAdsetId    string `gorm:"column:src_adset_id"`
	SrcCampaignId string `gorm:"column:src_campaign_id"`
	MetaData      string `gorm:"column:meta_data"`
}

func (o *OrderJoinSource) TableName() string {
	return "platform_offline.dwd_attr_3p_ref_order_join_source_20250926"
}

func InsertOrderJoinSource(reports []OrderJoinSource) {
	batchSize := 500
	db := platform_db.GetDB()
	log.Printf("Start inserting OrderJoinSource into DB, batch size: %d", batchSize)
	for i := 0; i < len(reports); i += batchSize {
		end := i + batchSize
		if end > len(reports) {
			end = len(reports)
		}
		batch := reports[i:end]
		if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(batch, len(batch)).Error; err != nil {
			log.Printf("failed to insert batch [%d:%d]: %v", i, end, err)
			panic(err)
		} else {
			log.Printf("successfully inserted batch [%d:%d]", i, end)
		}
	}
}
