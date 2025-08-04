package main

import "wm-func/cmd/alert-v2/dao"

func main() {
	//tenantIds := model.GetAllTenantsId()

	for _, table := range airbyte_raw_tables {
		dao.GetAirbyteRawDataWithTablName(table)
	}
}
