package main

import (
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

type KAccount struct {
	wm_account.Account
	lock2.Locker
}
