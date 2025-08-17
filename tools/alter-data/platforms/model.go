package platforms

import (
	"wm-func/common/db/platform_db"
	"wm-func/tools/alter-data/query_sql"
)

type AlterData struct {
	TenantId int64  `gorm:"column:tenant_id"`
	RawDate  string `gorm:"column:raw_date"`
	ApiSpend int64  `gorm:"column:api_spend"`
	AdSpend  int64  `gorm:"column:ad_spend"`
}

func GetGoogleData() []AlterData {
	exec_sql := query_sql.Query_google_api_with_overview
	db := platform_db.GetDB()
	res := []AlterData{}

	if err := db.Raw(exec_sql).Limit(-1).Scan(&res).Error; err != nil {
		panic(err)
	}
	return format(res)
}

func format(data []AlterData) []AlterData {
	for i := range data {
		data[i].RawDate = data[i].RawDate[:10]
	}
	return data
}
