package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm/clause"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"
)

var platform = "tiktokMarketing"

func main() {
	accounts := wm_account.GetAccountsWithPlatform(platform)

	tenantIds := map[int64]bool{
		150092: true,
		150012: true,
		150096: true,
	}

	for _, account := range accounts {
		if tenantIds[account.TenantId] {
			continue
		}

		var campaignInfo []Data
		ids := getCampaignIdByTenantId(account.TenantId)
		for _, id := range ids {
			fmt.Println(id, campaignInfo)
			campaignInfo = append(campaignInfo, *RequestById(account.AccountId, id, account.AccessToken))
		}

		var insertData []AribyteDate
		for _, data := range campaignInfo {
			b, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}

			insertData = append(insertData, AribyteDate{
				TenantId:            account.TenantId,
				AirbyteRawId:        fmt.Sprintf("%s|%s", account.AccountId, data.CampaignId),
				AirbyteData:         b,
				AirbyteExtractedAt:  time.Now().Format("2006-01-02 15:04:05"),
				AirbyteLoadedAt:     time.Now().Format("2006-01-02 15:04:05"),
				AirbyteMeta:         `{}`,
				AirbyteGenerationId: 0,
				ItemType:            "-",
			})

		}
		db := airbyte_db.GetDB()
		if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(insertData, 1).Error; err != nil {
			panic(err)
		}

		fmt.Println(ids)
	}
}
