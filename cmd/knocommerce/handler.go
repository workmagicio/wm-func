package main

import (
	"log"
	"wm-func/wm_account"
)

func RequestQuestion(account wm_account.Account, accessToken string) {
	res, err := GetKnoCommerceQuestion(accessToken)
	if err != nil {
		log.Println(err)
	}

}
