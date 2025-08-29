package cac

import (
	"sort"
	"time"
	"wm-func/common/config"
	"wm-func/tools/alter-data-v2/backend/bdao"
	"wm-func/tools/alter-data-v2/backend/bmodel"
)

type Cac struct {
}

type TenantDateSequence struct {
	TenantId      int64
	Last30DayDiff int64
	DateSequence  []DateSequence
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

func GetAlterDataWithPlatform(platform string, needRefresh bool) ([]TenantDateSequence, []TenantDateSequence) {
	b1 := bdao.GetApiDataByPlatform(needRefresh, platform)
	b2 := bdao.GetOverviewDataByPlatform(needRefresh, platform)

	var apiDataMap = map[int64]map[string]bmodel.ApiData{}
	for _, v := range b1 {
		if apiDataMap[v.TenantId] == nil {
			apiDataMap[v.TenantId] = make(map[string]bmodel.ApiData)
		}
		apiDataMap[v.TenantId][v.RawDate] = v
	}

	var overviewDataMap = map[int64]map[string]bmodel.OverViewData{}
	for _, v := range b2 {
		if overviewDataMap[v.TenantId] == nil {
			overviewDataMap[v.TenantId] = make(map[string]bmodel.OverViewData)
		}
		overviewDataMap[v.TenantId][v.EventDate] = v
	}

	// 按租户分组：30天为界限
	last30Day := time.Now().Add(config.DateDay * -30)
	var allTenant = bmodel.GetAllTenant()
	var newTenants []TenantDateSequence
	var oldTenants []TenantDateSequence

	tenantPlatformMap := bmodel.GetTenantPlatformMap()

	for _, tenant := range allTenant {

		if !tenantPlatformMap[tenant.TenantId][platform] {
			continue
		}

		tmp := GenerateDateSequence()
		for i, v := range tmp {
			if overviewData, exists := overviewDataMap[tenant.TenantId][v.Date]; exists {
				tmp[i].Data = overviewData.Value
			}
			if apiData, exists := apiDataMap[tenant.TenantId][v.Date]; exists {
				tmp[i].ApiData = apiData.AdSpend
			}
		}

		// 计算最近30天的diff
		last30DayDiff := calculateLast30DayDiff(tmp)

		tenantData := TenantDateSequence{
			TenantId:      tenant.TenantId,
			Last30DayDiff: last30DayDiff,
			DateSequence:  tmp,
		}

		// 分组：新租户 vs 老租户
		if tenant.RegisterTime.After(last30Day) {
			newTenants = append(newTenants, tenantData)
		} else {
			oldTenants = append(oldTenants, tenantData)
		}
	}

	// oldTenants 按照diff差值逆序排序
	sort.Slice(oldTenants, func(i, j int) bool {
		return oldTenants[i].Last30DayDiff > oldTenants[j].Last30DayDiff
	})

	return newTenants, oldTenants
}

func calculateLast30DayDiff(dateSequences []DateSequence) int64 {
	now := time.Now()
	last30Day := now.Add(config.DateDay * -30)

	var totalDiff int64 = 0
	for _, seq := range dateSequences {
		seqDate, err := time.Parse("2006-01-02", seq.Date)
		if err != nil {
			continue
		}
		if seqDate.After(last30Day) {
			diff := seq.Data - seq.ApiData // 以ApiData为基准
			totalDiff += diff
		}
	}
	return totalDiff
}
