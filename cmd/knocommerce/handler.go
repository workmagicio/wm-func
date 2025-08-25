package main

import (
	"gorm.io/gorm/clause"
	"log"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"
)

func RequestResponse(account wm_account.Account, accessToken string) {
	st, err := GetState(account, SUBTYPE_RESPONSE)
	if err != nil {
		panic(err)
	}

	if time.Now().Before(st.NextRunningTime) {
		log.Println("")
		return
	}

}

func RequestQuestion(account wm_account.Account, accessToken string) {
	res, err := GetKnoCommerceQuestion(accessToken)
	if err != nil {
		log.Println(err)
	}

	var data []AirbyteData
	for _, d := range res.Data.Questions {
		dd, e := TransToAirbyte(account, d)
		if e != nil {
			panic(e)
		}
		data = append(data, *dd)
	}
	SaveAirbyteData(account, data, SUBTYPE_QUESTION)
}

func RequestSurvey(account wm_account.Account, accessToken string) {
	res, err := GetAllKnoCommerceSurveys(accessToken)
	if err != nil {
		log.Println(err)
	}

	var data []AirbyteData
	for _, d := range res {
		dd, e := TransToAirbyte(account, d)
		if e != nil {
			panic(e)
		}
		data = append(data, *dd)
	}
	SaveAirbyteData(account, data, SUBTYPE_QUESTION)
}

func SaveAirbyteData(account wm_account.Account, data []AirbyteData, subType string) error {
	if len(data) == 0 {
		return nil
	}

	traceId := getTraceIdWithSubType(account, subType)

	db := airbyte_db.GetDB()
	table := GetAirbyteTableNameWithSubType(subType)
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
