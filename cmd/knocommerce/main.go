package main

import (
	"log"
	"wm-func/wm_account"
)

const Platform = "knocommerce"

func main() {

}

func init() {
	accounts := wm_account.GetAccountsWithPlatform(Platform)

	for _, account := range accounts {
		run(account)
	}
}

func run(account wm_account.Account) {
	accessToken, err := RefreshToken(account)
	if err != nil {
		log.Println(err)
		return
	}

	for _, subType := range subTypeList {

	}
}
