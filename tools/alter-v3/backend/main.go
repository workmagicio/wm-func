package main

import "wm-func/tools/alter-v3/backend/api"

func main() {
	router := api.SetupRouter()
	router.Run(":8090")

}
