package main

import (
	"encoding/json"
	"fmt"
	"log"
	"wm-func/common/state"
	"wm-func/wm_account"
)

// getState 获取账户的同步状态
func getState(account wm_account.ShopifyAccount) (SyncState, error) {
	syncInfo := state.GetSyncInfo(account.TenantId, account.ShopDomain, Platform, SubType)

	var syncState SyncState
	if syncInfo == nil {
		log.Printf("[%s] 首次同步，没有历史状态", account.GetTraceId())
		return syncState, nil
	}

	if err := json.Unmarshal(syncInfo, &syncState); err != nil {
		log.Printf("[%s] 解析同步状态失败: %v", account.GetTraceId(), err)
		return syncState, fmt.Errorf("解析同步状态失败: %w", err)
	}

	log.Printf("[%s] 获取同步状态成功，最后游标: %s", account.GetTraceId(), syncState.LastCursor)
	return syncState, nil
}

// 可以在此处添加其他通用的辅助函数
