package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"wm-func/common/state"
	"wm-func/wm_account"
)

// 获取账户的同步状态
func getState(account wm_account.Account, subType string) (SyncState, error) {
	traceId := getTraceIdWithSubType(account, subType)
	syncInfo := state.GetSyncInfo(account.TenantId, account.AccountId, Platform, subType)

	var syncState SyncState
	if syncInfo == nil {
		log.Printf("[%s] 首次同步，没有历史状态", traceId)
		return syncState, nil
	}

	if err := json.Unmarshal(syncInfo, &syncState); err != nil {
		log.Printf("[%s] 解析同步状态失败: %v", traceId, err)
		return syncState, fmt.Errorf("解析同步状态失败: %w", err)
	}

	// 根据数据类型显示不同的状态信息
	switch subType {
	case "question":
		log.Printf("[%s] 获取同步状态成功，上次记录数: %d", traceId, syncState.RecordCount)
	case "response":
		if syncState.LastSyncTime != nil {
			log.Printf("[%s] 获取同步状态成功，上次记录数: %d，上次同步时间: %s",
				traceId, syncState.RecordCount, syncState.LastSyncTime.Format(time.RFC3339))
		} else {
			log.Printf("[%s] 获取同步状态成功，首次增量同步，记录数: %d", traceId, syncState.RecordCount)
		}
	default:
		log.Printf("[%s] 获取同步状态成功，记录数: %d", traceId, syncState.RecordCount)
	}

	return syncState, nil
}

// 获取 Fairing 专属的同步状态
func getFairingState(account wm_account.Account, subType string) (FairingSyncState, error) {
	traceId := getTraceIdWithSubType(account, subType)
	syncInfo := state.GetSyncInfo(account.TenantId, account.AccountId, Platform, subType)

	var fairingSyncState FairingSyncState
	if syncInfo == nil {
		log.Printf("[%s] 首次同步，创建新的状态", traceId)
		fairingSyncState = NewFairingSyncState()
		return fairingSyncState, nil
	}

	// 先尝试解析为新的 FairingSyncState
	if err := json.Unmarshal(syncInfo, &fairingSyncState); err != nil {
		// 解析失败，可能是旧的 SyncState 格式，尝试兼容
		var oldSyncState SyncState
		if oldErr := json.Unmarshal(syncInfo, &oldSyncState); oldErr == nil {
			log.Printf("[%s] 检测到旧版本状态，进行兼容转换", traceId)
			fairingSyncState = convertOldSyncState(oldSyncState)
		} else {
			log.Printf("[%s] 解析同步状态失败: %v，创建新状态", traceId, err)
			fairingSyncState = NewFairingSyncState()
		}
	}

	// 根据数据类型显示不同的状态信息
	switch subType {
	case "question":
		log.Printf("[%s] 获取同步状态成功，上次记录数: %d", traceId, fairingSyncState.RecordCount)
	case "response":
		if fairingSyncState.IsInitialSync {
			log.Printf("[%s] 首次同步状态，计划同步%d天，已完成%d个slice，记录数: %d",
				traceId, fairingSyncState.InitialDays, fairingSyncState.CompletedSlices, fairingSyncState.RecordCount)
		} else if fairingSyncState.LastCompletedDate != nil {
			log.Printf("[%s] 增量同步状态，上次完成: %s，记录数: %d",
				traceId, fairingSyncState.LastCompletedDate.Format("2006-01-02"), fairingSyncState.RecordCount)
		} else {
			log.Printf("[%s] 获取同步状态成功，记录数: %d", traceId, fairingSyncState.RecordCount)
		}

		// 如果有进行中的同步任务，显示详细进度
		if fairingSyncState.CurrentSliceDate != nil && !fairingSyncState.IsAllSlicesCompleted() {
			log.Printf("[%s] 同步进度: %d/%d (%.1f%%)，当前处理到: %s",
				traceId, fairingSyncState.CompletedSlices, fairingSyncState.TotalSlices,
				fairingSyncState.GetProgress(), fairingSyncState.CurrentSliceDate.Format("2006-01-02"))
		}
	}

	return fairingSyncState, nil
}

// 兼容旧版本状态转换
func convertOldSyncState(oldState SyncState) FairingSyncState {
	newState := NewFairingSyncState()
	newState.Status = oldState.Status
	newState.Message = oldState.Message
	newState.UpdatedAt = oldState.UpdatedAt
	newState.RecordCount = oldState.RecordCount
	newState.LastSyncTime = oldState.LastSyncTime

	// 如果有历史同步时间，说明不是首次同步
	if oldState.LastSyncTime != nil {
		newState.IsInitialSync = false
		newState.LastCompletedDate = oldState.LastSyncTime
	}

	return newState
}

// 更新 Fairing 专属的同步状态
func updateFairingState(account wm_account.Account, fairingSyncState FairingSyncState, subType string) error {
	traceId := getTraceIdWithSubType(account, subType)
	stateData, err := json.Marshal(fairingSyncState)
	if err != nil {
		return fmt.Errorf("[%s] 序列化同步状态失败: %w", traceId, err)
	}

	state.SaveSyncInfo(account.TenantId, account.AccountId, Platform, subType, stateData)

	// 详细的状态日志
	if fairingSyncState.IsInitialSync && fairingSyncState.CurrentSliceDate != nil {
		log.Printf("[%s] 同步状态已更新 - 首次同步进行中，当前日期: %s，已完成%d个slice",
			traceId, fairingSyncState.CurrentSliceDate.Format("2006-01-02"), fairingSyncState.CompletedSlices)
	} else if fairingSyncState.LastCompletedDate != nil {
		log.Printf("[%s] 同步状态已更新 - 增量同步，最后完成日期: %s，状态: %s",
			traceId, fairingSyncState.LastCompletedDate.Format("2006-01-02"), fairingSyncState.Status)
	} else {
		log.Printf("[%s] 同步状态已更新，状态: %s", traceId, fairingSyncState.Status)
	}

	return nil
}

// 检查网络连接
func checkConnection() {
	instanceConfig := getInstanceConfig()
	log.Printf("[%s] 正在检查fairing服务连接...", instanceConfig.InstanceId)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 使用真实的Fairing API端点进行连接检查
	testURL := "https://app.fairing.co/api/questions"

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		log.Printf("[%s] 创建请求失败: %v", instanceConfig.InstanceId, err)
		return
	}

	// 如果有测试账户，添加认证头
	if testAccount := getTestAccount(); testAccount != nil {
		req.Header.Set("Authorization", fmt.Sprintf("%s", testAccount.SecretToken))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] fairing服务连接失败: %v", instanceConfig.InstanceId, err)
		return
	}
	defer resp.Body.Close()

	// 401表示服务可达但需要认证，这是正常的
	if resp.StatusCode == 401 {
		log.Printf("[%s] fairing服务连接正常(需要认证), 状态码: %d", instanceConfig.InstanceId, resp.StatusCode)
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("[%s] fairing服务连接正常, 状态码: %d", instanceConfig.InstanceId, resp.StatusCode)
	} else {
		log.Printf("[%s] fairing服务响应异常, 状态码: %d", instanceConfig.InstanceId, resp.StatusCode)
	}
}

// 生成实例ID（改进版）
func generateInstanceId() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	pid := os.Getpid()
	timestamp := time.Now().Unix()

	return fmt.Sprintf("fairing-%s-%d-%d", hostname, pid, timestamp)
}

// 任务统计结构
type TaskStats struct {
	TotalAccounts   int
	ProcessedTasks  int
	SuccessfulTasks int
	FailedTasks     int
	SkippedTasks    int
	StartTime       time.Time
	EndTime         time.Time
}

// 全局任务统计
var globalStats = &TaskStats{
	StartTime: time.Now(),
}

// 更新任务统计
func updateTaskStats(taskType string, success bool) {
	globalStats.ProcessedTasks++
	if success {
		globalStats.SuccessfulTasks++
	} else {
		globalStats.FailedTasks++
	}
}

// 跳过任务统计
func skipTaskStats() {
	globalStats.SkippedTasks++
}

// 打印最终统计信息
func printFinalStats() {
	globalStats.EndTime = time.Now()
	duration := globalStats.EndTime.Sub(globalStats.StartTime)

	instanceConfig := getInstanceConfig()

	log.Printf("[%s] ===== 任务执行统计 =====", instanceConfig.InstanceId)
	log.Printf("[%s] 总账户数: %d", instanceConfig.InstanceId, globalStats.TotalAccounts)
	log.Printf("[%s] 处理任务数: %d", instanceConfig.InstanceId, globalStats.ProcessedTasks)
	log.Printf("[%s] 成功任务数: %d", instanceConfig.InstanceId, globalStats.SuccessfulTasks)
	log.Printf("[%s] 失败任务数: %d", instanceConfig.InstanceId, globalStats.FailedTasks)
	log.Printf("[%s] 跳过任务数: %d", instanceConfig.InstanceId, globalStats.SkippedTasks)
	log.Printf("[%s] 执行时长: %v", instanceConfig.InstanceId, duration)
	log.Printf("[%s] ========================", instanceConfig.InstanceId)
}

// 环境变量辅助函数
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 其他通用工具函数可以在这里添加

// 获取测试账户信息的辅助函数
func getTestAccount() *wm_account.Account {
	// 这里可以根据实际情况获取测试账户
	// 暂时返回nil，让连接检查不使用认证
	return nil
}
