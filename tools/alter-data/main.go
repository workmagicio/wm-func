package main

import (
	"fmt"
	"wm-func/tools/alter-data/platforms"
)

func main() {
	res := platforms.GetGoogleData()
	fmt.Println(res)

}
