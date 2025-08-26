package main

import (
	"log"
	"time"
	"wm-func/common/db/airbyte_db"

	"gorm.io/gorm/clause"
)

var dateFormatDate = "2006-01-02"

func RequestResponseCount(account KAccount, token *TokenManager) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE_COUNT)
	log.Printf("[%s] 开始RequestResponseCount，获取回复统计数据", traceId)

	st, err := GetState(account, SUBTYPE_RESPONSE_COUNT)
	if err != nil {
		log.Printf("[%s] GetState失败: %v", traceId, err)
		panic(err)
	}

	if time.Now().Before(st.NextRunningTime) {
		log.Printf("[%s] 还未到执行时间，跳过本次运行", traceId)
		return
	}

	insertData := []Count{}
	lastSyncTime := time.Now().UTC()
	slice := GetStreamSlice(st.LastSync, time.Now().UTC(), dateFormatDate, SUBTYPE_RESPONSE_COUNT)

	log.Printf("[%s] 需要处理的时间分片数量: %d", traceId, len(slice))

	for i, v := range slice {
		log.Printf("[%s] 正在处理第%d个分片，时间范围: %s - %s", traceId, i+1, v.Start, v.End)

		// 获取开始日期的count
		var count int64
		count, err = GetKnoCommerceResponsesCount(token, v.Start, v.Start)
		if err != nil {
			log.Printf("[%s] GetKnoCommerceResponsesCount失败(开始日期%s): %v", traceId, v.Start, err)
			panic(err)
		}
		log.Printf("[%s] 获取到%s的回复数量: %d", traceId, v.Start, count)

		insertData = append(insertData, Count{
			Count:    count,
			StatDate: v.Start,
		})
		time.Sleep(time.Second * 2)

		// 获取结束日期的count
		count, err = GetKnoCommerceResponsesCount(token, v.End, v.End)
		if err != nil {
			log.Printf("[%s] GetKnoCommerceResponsesCount失败(结束日期%s): %v", traceId, v.End, err)
			panic(err)
		}
		log.Printf("[%s] 获取到%s的回复数量: %d", traceId, v.End, count)

		insertData = append(insertData, Count{
			Count:    count,
			StatDate: v.End,
		})
		time.Sleep(time.Second * 2)
	}

	log.Printf("[%s] 总共收集到统计数据条数: %d", traceId, len(insertData))
	log.Printf("[%s] 开始保存统计数据到Airbyte", traceId)

	var airbyteData []AirbyteData
	for _, v := range insertData {
		airbyteData = append(airbyteData, *TransToAirbyte(account, v))
	}
	SaveAirbyteData(account, airbyteData, SUBTYPE_RESPONSE_COUNT)

	log.Printf("[%s] 更新同步状态", traceId)

	// 判断lastSyncTime是否在最近7天内，如果是才减去30天
	sevenDaysAgo := time.Now().UTC().Add(Day * -7)
	if lastSyncTime.After(sevenDaysAgo) {
		st.LastSync = lastSyncTime.Add(Day * -30)
		log.Printf("[%s] lastSyncTime在最近7天内，减去30天后设置为: %v", traceId, st.LastSync)
	} else {
		st.LastSync = lastSyncTime
		log.Printf("[%s] lastSyncTime不在最近7天内，直接设置为: %v", traceId, st.LastSync)
	}

	SaveState(account, st)
	log.Printf("[%s] RequestResponseCount完成", traceId)
}

func RequestResponse(account KAccount, token *TokenManager) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE)
	log.Printf("[%s] 开始RequestResponse，获取回复数据", traceId)

	st, err := GetState(account, SUBTYPE_RESPONSE)
	if err != nil {
		log.Printf("[%s] GetState失败: %v", traceId, err)
		panic(err)
	}

	if time.Now().Before(st.NextRunningTime) {
		log.Printf("[%s] 还未到执行时间，跳过本次运行", traceId)
		return
	}

	insertData := []Result{}
	lastSyncTime := time.Now().UTC()
	slice := GetStreamSlice(st.LastSync, time.Now().UTC(), dateFormatDate, SUBTYPE_RESPONSE)

	log.Printf("[%s] 需要处理的时间分片数量: %d", traceId, len(slice))

	for i, v := range slice {
		log.Printf("[%s] 正在处理第%d个分片，时间范围: %s - %s", traceId, i+1, v.Start, v.End)

		var tmp []Result
		tmp, err = GetAllKnoCommerceResponses(account, token, v.Start, v.End)
		if err != nil {
			log.Printf("[%s] GetAllKnoCommerceResponses失败: %v", traceId, err)
			panic(err)
		}

		log.Printf("[%s] 第%d个分片获取到回复数量: %d", traceId, i+1, len(tmp))

		if len(tmp) > 0 {
			lastSyncTime = tmp[len(tmp)-1].CreatedAt
			insertData = append(insertData, tmp...)
		}

		if len(insertData) >= 500 || i == len(slice)-1 {
			if len(insertData) > 0 {
				log.Printf("[%s] 开始保存批次数据到Airbyte，数据量: %d", traceId, len(insertData))
				airbyteData := []AirbyteData{}
				for _, d := range insertData {
					airbyteData = append(airbyteData, *TransToAirbyte(account, d))
				}
				SaveAirbyteData(account, airbyteData, SUBTYPE_RESPONSE)
				insertData = []Result{}

				log.Printf("[%s] 更新同步状态", traceId)

				// 判断lastSyncTime是否在最近7天内，如果是才减去30天
				sevenDaysAgo := time.Now().UTC().Add(Day * -7)
				if lastSyncTime.After(sevenDaysAgo) {
					st.LastSync = lastSyncTime.Add(Day * -30)
					log.Printf("[%s] lastSyncTime在最近7天内，减去30天后设置为: %v", traceId, st.LastSync)
				} else {
					st.LastSync = lastSyncTime
					log.Printf("[%s] lastSyncTime不在最近7天内，直接设置为: %v", traceId, st.LastSync)
				}

				SaveState(account, st)
			}
		}
	}

	log.Printf("[%s] RequestResponse完成", traceId)
}

func RequestQuestion(account KAccount, token *TokenManager) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_QUESTION)
	log.Printf("[%s] 开始RequestQuestion，获取问题基准数据", traceId)

	res, err := GetKnoCommerceQuestion(token.GetAccessToken())
	if err != nil {
		log.Printf("[%s] GetKnoCommerceQuestion失败: %v", traceId, err)
		return
	}

	log.Printf("[%s] 成功获取到问题数据，问题数量: %d", traceId, len(res.Data.Questions))

	var data []AirbyteData
	for _, d := range res.Data.Questions {
		data = append(data, *TransToAirbyte(account, d))
	}

	log.Printf("[%s] 开始保存问题数据到Airbyte", traceId)
	SaveAirbyteData(account, data, SUBTYPE_QUESTION)
	log.Printf("[%s] RequestQuestion完成", traceId)
}

func RequestSurvey(account KAccount, token *TokenManager) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_SURVEY)
	log.Printf("[%s] 开始RequestSurvey，获取调查问卷数据", traceId)

	res, err := GetAllKnoCommerceSurveys(account, token)
	if err != nil {
		log.Printf("[%s] GetAllKnoCommerceSurveys失败: %v", traceId, err)
		return
	}

	log.Printf("[%s] 成功获取到调查问卷数据，问卷数量: %d", traceId, len(res))

	var data []AirbyteData
	for _, d := range res {
		data = append(data, *TransToAirbyte(account, d))
	}

	log.Printf("[%s] 开始保存调查问卷数据到Airbyte", traceId)
	SaveAirbyteData(account, data, SUBTYPE_SURVEY)
	log.Printf("[%s] RequestSurvey完成", traceId)
}

func SaveAirbyteData(account KAccount, data []AirbyteData, subType string) error {
	if len(data) == 0 {
		return nil
	}

	traceId := account.GetTraceIdWithSubType(subType)

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
