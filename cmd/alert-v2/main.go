package main

import (
	"fmt"
	"strings"
	"wm-func/common/db/platform_db"
	"wm-func/common/model"
)

func main() {
	//tenantIds := model.GetAllTenantsId()
	//db := platform_db.GetDB()
	//for tableName, fileds := range airbyte_raw_tables {
	//	dao.GetAirbyteRawDataWithTablName(tableName, strings.Join(fileds, ","), db)
	//	time.Sleep(time.Second * 3)
	//}
	check2()
}

func check2() {
	db := platform_db.GetDB()
	tenantSlices := tenantStrSlice(model.GetAllTenantsIdWithPlatform())
	fmt.Println(tenantSlices)
	for name, sql := range over_view_sqls {
		check(name, sql, tenantSlices)
		fmt.Println(name, sql)
	}
	fmt.Println(db)
}

func check(name, sql string, slice [][]string) {
	db := platform_db.GetDB()
	for _, tenants := range slice {
		execSql := strings.ReplaceAll(sql, "{{tenant_ids}}", strings.Join(tenants, ","))
		fmt.Println(execSql)

		res := query(execSql, db)
		f
	}

}

func tenantStrSlice(ids []int64) [][]string {
	res := [][]string{}
	tmp := []string{}
	for i, id := range ids {
		tmp = append(tmp, fmt.Sprintf("%d", id))

		if len(tmp) == 15 || i == len(ids)-1 {
			res = append(res, tmp)
			tmp = []string{}
		}
	}
	return res
}
