package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wm-func/common/state"
	"wm-func/wm_account"
)

const (
	STATUS_SUCCESS = "SUCCESS"
	STATUS_FAILED  = "FAILED"
)

// SyncState 同步状态结构体
type SyncState struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
	// 可以根据需要添加其他字段
}

// getState 获取账户的同步状态
func getState(account wm_account.Account) (SyncState, error) {
	syncInfo := state.GetSyncInfo(account.TenantId, account.AccountId, Platform, SubType)

	var syncState SyncState
	if syncInfo == nil {
		log.Printf("[%s] 首次同步，没有历史状态", account.GetTraceId())
		return syncState, nil
	}

	if err := json.Unmarshal(syncInfo, &syncState); err != nil {
		log.Printf("[%s] 解析同步状态失败: %v", account.GetTraceId(), err)
		return syncState, fmt.Errorf("解析同步状态失败: %w", err)
	}

	log.Printf("[%s] 获取同步状态成功", account.GetTraceId())
	return syncState, nil
}

// updateSyncState 更新同步状态
func updateSyncState(account wm_account.Account, syncState SyncState) error {
	syncState.UpdatedAt = time.Now().UTC()

	data, err := json.Marshal(syncState)
	if err != nil {
		return fmt.Errorf("序列化同步状态失败: %w", err)
	}

	state.SaveSyncInfo(account.TenantId, account.AccountId, Platform, SubType, data)
	log.Printf("[%s] 更新同步状态成功", account.GetTraceId())
	return nil
}

// syncAdMetrics 同步广告数据（占位函数，你可以在这里添加自己的逻辑）
func syncAdMetrics(account wm_account.Account, syncState SyncState) error {
	log.Printf("[%s] 开始同步广告数据", account.GetTraceId())

	// TODO: 在这里添加你的 Meta 广告数据同步逻辑
	// 例如：
	// 1. 调用 Meta Marketing API
	// 2. 获取广告账户数据
	// 3. 处理数据并存储

	// 模拟同步过程
	time.Sleep(1 * time.Second)

	// 更新同步状态为成功
	syncState.Status = STATUS_SUCCESS
	syncState.Message = "同步成功"

	log.Printf("[%s] 广告数据同步完成", account.GetTraceId())
	return nil
}

/*
step1
创建异步任务
curl --location --globoff --request POST 'https://graph.facebook.com/v23.0/act_{account_id}/insights?fields=date_start%2Cspend&time_range={%27since%27%3A%272025-07-01%27%2C%27until%27%3A%272025-07-26%27}&time_increment=1&access_token={access_token}'
返回结果

{
    "report_run_id": "1496466178359684"
}

step2
查询任务状态
curl --location 'https://graph.facebook.com/v23.0/{report_run_id}?access_token={access_token}'
{
"id": "1496466178359684",
"account_id": "202470630091557",
"time_ref": 1754035923,
"time_completed": 1754035925,
"async_status": "Job Completed",
"async_percent_completion": 100,
"date_start": "2025-07-01",
"date_stop": "2025-07-26"
}

step3
获取结果
https://graph.facebook.com/v19.0/{{meta_report_run_id}}/insights?access_token={{meta_access_token}}
{
    "data": [
		内容在：ads_insights.json 中
    ],
    "paging": {
        "cursors": {
            "before": "MAZDZD",
            "after": "MjQZD"
        },
        "next": "https://graph.facebook.com/v21.0/1496466178359684/insights?access_token={xxx}&limit=25&after=MjQZD" -- 本连接中自带accesstoken，直接用这个链接查就行了
    }
}
*/
