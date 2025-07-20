package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
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

func checkApollo() {
	// 【最终诊断代码】: 在一切开始前，手动测试网络连接
	apolloMetaServer := "http://internal-apollo-meta-server-preview.workmagic.io"
	log.Printf("正在诊断网络: 尝试连接 %s", apolloMetaServer)

	// 设置一个5秒的超时
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(apolloMetaServer)
	if err != nil {
		// 如果网络不通，这里会打印出决定性的错误信息
		log.Fatalf("诊断失败: 无法连接到 Apollo Meta Server。根本错误: %v", err)
	}
	resp.Body.Close()
	log.Printf("诊断成功: 成功连接到 Apollo Meta Server, 状态码: %d", resp.StatusCode)
}
