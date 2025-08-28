package main

import (
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

func main() {
	accounts := wm_account.GetFairingAccounts()

	FAccounts := []FAccount{}

	lock := lock2.NewMySQLLocker()
	for _, account := range accounts {
		FAccounts = append(FAccounts, FAccount{
			Account: account,
			Locker:  lock,
		})
	}

	for _, account := range FAccounts {
		//if account.TenantId != 150102 {
		//	continue
		//}

		run(account)

	}
}

func run(account FAccount) {
	RequestResponse(account)
}
