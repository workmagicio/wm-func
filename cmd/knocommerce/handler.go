package main

import (
	"gorm.io/gorm/clause"
	"log"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"
)

var dateFormatDate = "2006-01-02"

func RequestResponseCount(account wm_account.Account, accessToken string) {
	st, err := GetState(account, SUBTYPE_RESPONSE_COUNT)
	if err != nil {
		panic(err)
	}

	if time.Now().Before(st.NextRunningTime) {
		log.Println("")
		return
	}

	insertData := []Count{}
	lastSyncTime := time.Now().UTC()
	slice := GetStreamSlice(st.LastSync, time.Now().UTC(), dateFormatDate, SUBTYPE_RESPONSE_COUNT)
	for _, v := range slice {
		var count int64
		count, err = GetKnoCommerceResponsesCount(accessToken, v.Start, v.Start)
		if err != nil {
			panic(err)
		}
		insertData = append(insertData, Count{
			Count:    count,
			StatDate: v.Start,
		})
		time.Sleep(time.Second * 5)
		count, err = GetKnoCommerceResponsesCount(accessToken, v.End, v.End)
		if err != nil {
			panic(err)
		}
		insertData = append(insertData, Count{
			Count:    count,
			StatDate: v.End,
		})
		time.Sleep(time.Second * 5)
	}

	var airbyteData []AirbyteData
	for _, v := range insertData {
		airbyteData = append(airbyteData, *TransToAirbyte(account, v))
	}
	SaveAirbyteData(account, airbyteData, SUBTYPE_RESPONSE_COUNT)

	st.LastSync = lastSyncTime.Add(Day * -30)
	SaveState(account, st)

}

func RequestResponse(account wm_account.Account, accessToken string) {
	st, err := GetState(account, SUBTYPE_RESPONSE)
	if err != nil {
		panic(err)
	}

	if time.Now().Before(st.NextRunningTime) {
		log.Println("")
		return
	}

	insertData := []Result{}
	lastSyncTime := time.Now().UTC()
	slice := GetStreamSlice(st.LastSync, time.Now().UTC(), dateFormatDate, SUBTYPE_RESPONSE)
	for i, v := range slice {
		var tmp []Result
		tmp, err = GetAllKnoCommerceResponses(accessToken, v.Start, v.End)
		if err != nil {
			panic(err)
		}

		lastSyncTime = tmp[len(tmp)-1].CreatedAt

		insertData = append(insertData, tmp...)

		if len(insertData) == 500 || i == len(slice)-1 {
			airbyteData := []AirbyteData{}
			for _, d := range insertData {
				airbyteData = append(airbyteData, *TransToAirbyte(account, d))
			}
			SaveAirbyteData(account, airbyteData, SUBTYPE_RESPONSE)
			insertData = []Result{}
		}
	}

	st.LastSync = lastSyncTime.Add(Day * -30)
	SaveState(account, st)
}

func RequestQuestion(account wm_account.Account, accessToken string) {
	res, err := GetKnoCommerceQuestion(accessToken)
	if err != nil {
		log.Println(err)
	}

	var data []AirbyteData
	for _, d := range res.Data.Questions {
		data = append(data, *TransToAirbyte(account, d))
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
		data = append(data, *TransToAirbyte(account, d))
	}
	SaveAirbyteData(account, data, SUBTYPE_SURVEY)
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
