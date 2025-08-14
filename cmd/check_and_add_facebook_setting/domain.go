package main

import (
	"gorm.io/gorm/clause"
	"log"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"
)

type Ad struct {
	TenantId int64  `json:"tenant_id"`
	AdId     string `json:"ad_id"`
}
type AdSet struct {
	TenantId int64  `json:"tenant_id"`
	AdSetId  string `json:"ad_set_id"`
}
type Campaign struct {
	TenantId   int64  `json:"tenant_id"`
	CampaignId string `json:"campaign_id"`
}

func GetAllAds() []Ad {
	db := airbyte_db.GetDB()
	var ads []Ad
	if err := db.Exec(query_loss_ad_data).Scan(&ads).Error; err != nil {
		panic(err)
	}
	return ads
}

func GetAllAdSet() []AdSet {
	db := airbyte_db.GetDB()
	var ads []AdSet
	if err := db.Exec(query_loss_adset_data).Scan(&ads).Error; err != nil {
		panic(err)
	}
	return ads
}

func GetAllCampaign() []Campaign {
	db := airbyte_db.GetDB()
	var ads []Campaign
	if err := db.Exec(query_loss_campaign).Scan(&ads).Error; err != nil {
		panic(err)
	}
	return ads
}

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

func (s *AirbyteData) TableName() string {
	//return "airbyte_destination_v2.raw_facebook_marketing_campaigns"
	return "airbyte_destination_v2.raw_facebook_marketing_ad_sets"
}

func saveFirstOrderCustomers(account wm_account.Account, customers []AirbyteData) error {
	if len(customers) == 0 {
		return nil
	}

	db := airbyte_db.GetDB()
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(customers, 1).Error; err != nil {
		return err
	}
	log.Printf("[%s] successfully inserted %d", account.GetTraceId(), len(customers))
	log.Println(customers[0].AirbyteRawId)

	return nil
}
