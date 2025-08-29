package cac

import (
	"fmt"
	"time"
	"wm-func/common/config"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bmodel"
)

type Cac struct {
}

type TenantDateSequence struct {
	TenantId int64
	DateSequence
}

type DateSequence struct {
	Date    string `json:"date"`
	ApiData int64  `json:"api_data"`
	Data    int64  `json:"data"`
}

func GenerateDateSequence() []DateSequence {
	now := time.Now()
	start := now.Add(config.DateDay * -90)
	var res []DateSequence

	for start.Before(now) {
		res = append(res, DateSequence{
			Date:    start.Format("2006-01-02"),
			ApiData: 0,
			Data:    0,
		})
		start = start.Add(config.DateDay)
	}

	return res
}

func GetAlterDataWithPlatform(platform string, needRefresh bool) {
	b1 := bdao.GetApiDataByPlatform(needRefresh, platform)
	b2 := bdao.GetOverviewDataByPlatform(needRefresh, platform)
	fmt.Println(b1, b2)

	var apiDataMap = map[int64]map[string]bmodel.ApiData{}
	for _, v := range b1 {
		if apiDataMap[v.TenantId] == nil {
			apiDataMap[v.TenantId] = make(map[string]bmodel.ApiData)
		}
		apiDataMap[v.TenantId][v.RawDate] = v
	}

	var overviewDataMap = map[int64]map[string]bmodel.ApiData{}
	for _, v := range b1 {
		if overviewDataMap[v.TenantId] == nil {
			overviewDataMap[v.TenantId] = make(map[string]bmodel.ApiData)
		}
		overviewDataMap[v.TenantId][v.RawDate] = v
	}

	var res = map[int64][]DateSequence{}
	var allTenant = bmodel.GetAllTenant()
	for _, tenant := range allTenant {
		tmp := GenerateDateSequence()
		for i, v := range tmp {
			tmp[i].Data = overviewDataMap[tenant.TenantId][v.Date].AdSpend
			tmp[i].ApiData = overviewDataMap[tenant.TenantId][v.Date].AdSpend
		}
		res[tenant.TenantId] = tmp
	}

	fmt.Println(apiDataMap, overviewDataMap)

}
