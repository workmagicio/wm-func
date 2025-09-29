package data

import "wm-func/tools/alter-v3/backend/alter_common"

type DataSlice struct {
	Date    string
	APiData int64
	WMData  int64
}

type Analytics struct {
	TenantId int64
	Data     []DataSlice
}

func NewAnalytics(apiData, wmData []alter_common.Data, slices []string) Analytics {

}
