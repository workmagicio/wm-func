package main

import (
	"fmt"
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

type KAccount struct {
	wm_account.Account
	lock2.Locker
}

// GetSimpleTraceId 获取简化的跟踪ID (只包含TenantId，不包含AccountId)
func (ka KAccount) GetSimpleTraceId() string {
	return fmt.Sprintf("%d", ka.TenantId)
}

// GetTraceIdWithSubType 获取包含子类型的跟踪ID
func (ka KAccount) GetTraceIdWithSubType(subType string) string {
	return fmt.Sprintf("%d-%s", ka.TenantId, subType)
}
