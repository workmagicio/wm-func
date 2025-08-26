package main

func getTraceIdWithSubType(account KAccount, subType string) string {
	return account.GetTraceId() + "-" + subType
}
