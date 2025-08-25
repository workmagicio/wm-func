package main

import (
	"fmt"
	"testing"
	"wm-func/wm_account"
)

func TestRunning(t *testing.T) {
	accounts := wm_account.GetAccountsWithPlatform(Platform)
	for _, account := range accounts {
		res, err := RefreshToken(account)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}

	fmt.Println(accounts)
}
