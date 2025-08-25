package main

import (
	"log"
	"wm-func/wm_account"
)

const Platform = "knocommerce"

func main() {
	accounts := wm_account.GetAccountsWithPlatform(Platform)

	for _, account := range accounts {
		run(account)
	}
}

func run(account wm_account.Account) {
	token, err := RefreshToken(account)
	if err != nil {
		log.Println(err)
		return
	}

	//RequestQuestion(account, token.AccessToken)
	RequestSurvey(account, token.AccessToken)
	//for _, subType := range subTypeList {
	//
	//}
}
