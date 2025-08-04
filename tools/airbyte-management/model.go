package main

import (
	"fmt"
	"wm-func/common/db/airbyte_db"
	"wm-func/common/db/platform_db"
)

type Connection struct {
	ConnectionId string
	Name         string
	Prefix       string
	Status       string
}

func GetAllConnectionWithPrefix(prefix string) []Connection {
	sql := fmt.Sprintf(query_all_connection_with_prefix, prefix)

	var result []Connection
	client := platform_db.GetDB()
	if err := client.Raw(sql).Limit(-1).Scan(&result).Error; err != nil {
		panic(err)
	}

	return result
}

func GetAllConnections() []Connection {
	sql := query_all_connection

	var result []Connection
	//client := platform_db.GetDB()
	client := airbyte_db.GetDB()
	if err := client.Raw(sql).Limit(-1).Scan(&result).Error; err != nil {
		panic(err)
	}

	return result
}
