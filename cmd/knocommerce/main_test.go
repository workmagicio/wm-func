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

		state, err := GetState(account, "question")
		if err != nil {
			panic(err)
		}

		SaveState(account, state)
	}

	fmt.Println(accounts)
}
