package main

import "wm-func/wm_account"

func getTraceIdWithSubType(account wm_account.Account, subType string) string {
	return account.GetTraceId() + "-" + subType
}
