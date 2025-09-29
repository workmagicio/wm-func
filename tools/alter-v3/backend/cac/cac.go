package cac

import (
	"wm-func/tools/alter-data-v2/backend/tags"
	"wm-func/tools/alter-v3/backend/data"
)

func Ca(platform string, tenantData []*data.Analytics) {
	for _, t := range tenantData {
		last7dLossDataCheck(platform, t)
		last30dDataDiff(t)
		attachDefaultTags(t)
	}
}

func last7dLossDataCheck(platform string, analytics *data.Analytics) {

	zeroData := last30DayZeroData(analytics)
	hasErr := false
	lastSyncDate := analytics.Data[0].Date
	isContinuous := true
	for i := 0; i < 7; i++ {

		if i == 0 {
			if analytics.HaveApiData && analytics.Data[i].WMData*2 >= analytics.Data[i].APiData {
				continue
			}
		}

		// 如果小于平均值的 10% 这视为缺数
		if analytics.Data[i].WMData < zeroData {
			if platform == "amazonVendorPartner" && i <= 2 {
				continue
			}

			hasErr = true
		} else {
			isContinuous = false
		}

		if isContinuous {
			lastSyncDate = analytics.Data[i].Date
		}
	}

	if hasErr {
		analytics.Tags = append(analytics.Tags, "last_7d_no_data")
	}

	analytics.LastSyncDate = lastSyncDate

}

func last30dDataDiff(analytics *data.Analytics) {
	zeroData := last30DayZeroData(analytics)

	hasErr := false
	for i := 7; i < 30; i++ {
		if analytics.HaveApiData && analytics.Data[i].WMData == analytics.Data[i].APiData {
			continue
		}

		if analytics.Data[i].WMData < zeroData {
			hasErr = true
		}
	}

	if hasErr {
		analytics.Tags = append(analytics.Tags, "last_30d_no_data")
	}
}

func last30DayZeroData(analytics *data.Analytics) int64 {
	var total int64

	for i := 0; i < 30; i++ {
		if analytics.HaveApiData {
			total = analytics.Data[i].APiData
		} else {
			total = analytics.Data[i].WMData
		}
	}

	return total / 30 * 5 / 100

}

func attachDefaultTags(analytics *data.Analytics) {
	defaultTags := tags.GetDefaultTags()
	if _, ok := defaultTags[analytics.TenantId]; ok {
		analytics.Tags = append(analytics.Tags, defaultTags[analytics.TenantId])
	}
}
