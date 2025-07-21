package main

import (
	"log"
	"os"
	"strconv"
	"time"
	t_pool "wm-func/common/pool"
	"wm-func/common/state"
	"wm-func/wm_account"
)

// getTraceIdWithSubType 生成包含 subtype 的 trace ID
func getTraceIdWithSubType(account wm_account.Account, subType string) string {
	return account.GetTraceId() + "-" + subType
}

func main() {
	instanceConfig := getInstanceConfig()
	log.Printf("[%s] Fairing数据同步程序启动", instanceConfig.InstanceId)

	// 检查是否是测试模式
	if os.Getenv("FAIRING_TEST_MODE") == "true" {
		log.Println("运行在测试模式...")
		runTestMode()
		return
	}

	log.Println("start run fairing data sync...")

	// 确保程序结束时打印统计信息
	defer printFinalStats()

	run()
	log.Println("end run fairing data sync...")
}

// 测试模式 - 用于验证API逻辑
func runTestMode() {
	log.Println("=== Fairing API 测试模式 ===")

	// 获取测试账户
	accounts := wm_account.GetFairingAccounts()
	if len(accounts) == 0 {
		log.Println("没有找到Fairing账户，无法进行测试")
		return
	}

	// 使用第一个账户进行测试
	testAccount := accounts[0]
	log.Printf("使用账户进行测试: %s", testAccount.GetTraceId())

	// 测试 Questions API
	log.Println("\n--- 测试 Questions API ---")
	testQuestionsAPI(testAccount)

	// 测试 Responses API
	log.Println("\n--- 测试 Responses API ---")
	testResponsesAPI(testAccount)

	log.Println("=== 测试完成 ===")
}

// 测试 Questions API
func testQuestionsAPI(account wm_account.Account) {
	traceId := getTraceIdWithSubType(account, "question")

	log.Printf("[%s] 开始测试 Questions API", traceId)

	// 调用API
	questions, err := callFairingQuestionsAPI(account)
	if err != nil {
		log.Printf("[%s] API调用失败: %v", traceId, err)
		return
	}

	// 处理数据
	fairingData, err := processQuestionsData(account, questions)
	if err != nil {
		log.Printf("[%s] 数据处理失败: %v", traceId, err)
		return
	}

	log.Printf("[%s] 成功处理 %d 条 question 数据", traceId, len(fairingData))

	// 显示数据样本
	if len(fairingData) > 0 {
		sample := fairingData[0]
		log.Printf("[%s] 数据样本:", traceId)
		log.Printf("  - TenantId: %d", sample.TenantId)
		log.Printf("  - AirbyteRawId: %s", sample.AirbyteRawId)
		log.Printf("  - ItemType: %s", sample.ItemType)
		log.Printf("  - 数据长度: %d bytes", len(sample.AirbyteData))

		// 显示原始问题数据
		if len(questions) > 0 {
			log.Printf("  - 问题ID: %d", questions[0].Id)
			log.Printf("  - 问题内容: %s", questions[0].Prompt)
			log.Printf("  - 问题类型: %s", questions[0].Type)
		}
	}

	// 如果环境变量允许，可以尝试保存数据
	if os.Getenv("FAIRING_TEST_SAVE") == "true" {
		log.Printf("[%s] 尝试保存测试数据...", traceId)
		if err := saveFairingData(account, fairingData, "question"); err != nil {
			log.Printf("[%s] 保存数据失败: %v", traceId, err)
		} else {
			log.Printf("[%s] 测试数据保存成功", traceId)
		}
	}
}

// 测试 Responses API
func testResponsesAPI(account wm_account.Account) {
	traceId := getTraceIdWithSubType(account, "response")
	config := getFairingConfig()

	log.Printf("[%s] 开始测试 Responses API", traceId)

	// 测试分页获取（第一页）
	responsesResp, err := callFairingResponsesAPI(account, nil, "", config.ResponsesPageSize)
	if err != nil {
		log.Printf("[%s] API调用失败: %v", traceId, err)
		return
	}

	log.Printf("[%s] 成功获取第一页数据，共 %d 条", traceId, len(responsesResp.Data))

	// 显示分页信息
	if responsesResp.Next != nil {
		log.Printf("[%s] 有下一页数据: %s", traceId, *responsesResp.Next)
	} else {
		log.Printf("[%s] 这是最后一页", traceId)
	}

	if responsesResp.Prev != nil {
		log.Printf("[%s] 有上一页数据: %s", traceId, *responsesResp.Prev)
	}

	// 处理数据
	if len(responsesResp.Data) > 0 {
		fairingData, err := processResponsesData(account, responsesResp.Data)
		if err != nil {
			log.Printf("[%s] 数据处理失败: %v", traceId, err)
			return
		}

		log.Printf("[%s] 成功处理 %d 条 response 数据", traceId, len(fairingData))

		// 显示数据样本
		sample := fairingData[0]
		log.Printf("[%s] 数据样本:", traceId)
		log.Printf("  - TenantId: %d", sample.TenantId)
		log.Printf("  - AirbyteRawId: %s", sample.AirbyteRawId)
		log.Printf("  - ItemType: %s", sample.ItemType)
		log.Printf("  - 数据长度: %d bytes", len(sample.AirbyteData))

		// 显示原始响应数据
		response := responsesResp.Data[0]
		log.Printf("  - 响应ID: %s", response.Id)
		log.Printf("  - 问题: %s", response.Question)
		log.Printf("  - 回答: %s", response.Response)
		log.Printf("  - 客户ID: %s", response.CustomerId)
		log.Printf("  - 订单总额: %s", response.OrderTotal)

		// 如果环境变量允许，可以尝试保存数据
		if os.Getenv("FAIRING_TEST_SAVE") == "true" {
			log.Printf("[%s] 尝试保存测试数据...", traceId)
			if err := saveFairingData(account, fairingData, "response"); err != nil {
				log.Printf("[%s] 保存数据失败: %v", traceId, err)
			} else {
				log.Printf("[%s] 测试数据保存成功", traceId)
			}
		}
	} else {
		log.Printf("[%s] 当前没有response数据", traceId)
	}
}

// 安全的字符串处理函数
func safeString(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func run() {
	accounts := wm_account.GetFairingAccounts()
	log.Printf("start total accounts: %d", len(accounts))

	// 更新统计信息
	globalStats.TotalAccounts = len(accounts)

	// 为了支持多实例并发，使用较小的worker pool
	pool := t_pool.NewWorkerPool(MaxWorkers)
	pool.Run()
	defer pool.Close()

	for _, account := range accounts {
		// 避免闭包问题，复制account变量
		acc := account
		pool.AddTask(func() {
			log.Printf("[%s] start process account", acc.GetTraceId())
			processAccount(acc)
			log.Printf("[%s] end process account", acc.GetTraceId())
		})
	}

	pool.Wait()
}

// processAccount 处理单个账户的所有数据类型
func processAccount(account wm_account.Account) {
	// 处理question和response两种数据类型
	for _, subType := range subTypes {
		processTask(account, subType)
	}
}

// processTask 处理单个任务（账户+数据类型）
func processTask(account wm_account.Account, subType string) {
	traceId := getTraceIdWithSubType(account, subType)
	log.Printf("[%s] 尝试获取任务锁", traceId)

	// 1. 尝试获取任务锁
	taskResult := state.GetAvailableTask(account.TenantId, account.AccountId, Platform, subType)

	switch taskResult.Status {
	case state.TaskStatusNotFound:
		log.Printf("[%s] 任务不存在，创建初始状态", traceId)
		// 创建初始状态并尝试获取锁
		if subType == "response" {
			// response 使用新的状态结构
			createInitialFairingState(account, subType)
		} else {
			// 其他类型使用原有逻辑
			createInitialState(account, subType)
		}
		// 重新尝试获取任务
		taskResult = state.GetAvailableTask(account.TenantId, account.AccountId, Platform, subType)
		if taskResult.Status != state.TaskStatusAcquired {
			log.Printf("[%s] 任务创建后仍无法获取，跳过", traceId)
			skipTaskStats()
			return
		}
	case state.TaskStatusAlreadyRunning:
		log.Printf("[%s] 任务正在其他实例运行，跳过", traceId)
		skipTaskStats()
		return
	case state.TaskStatusAcquired:
		log.Printf("[%s] 任务锁获取成功", traceId)
	}

	// 2. 确保在函数结束时释放锁
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[%s] 任务执行出现panic: %v，释放锁", traceId, err)
			updateTaskStats(subType, false)
		}
		state.SetStop(account.TenantId, account.AccountId, Platform, subType)
		log.Printf("[%s] 任务锁已释放", traceId)
	}()

	// 3. 执行具体的同步任务（优先使用支持时间范围的版本）
	var err error
	if subType == "response" {
		// response 使用新的时间范围同步逻辑
		err = execTaskWithTimeRange(account, subType)
	} else {
		// question 继续使用原有逻辑
		err = execTask(account, subType)
	}
	success := err == nil
	updateTaskStats(subType, success)

	if err != nil {
		log.Printf("[%s] 任务执行失败: %v", traceId, err)
		return
	}

	log.Printf("[%s] 任务执行成功", traceId)
}

// execTaskWithTimeRange 执行支持时间范围同步的任务
func execTaskWithTimeRange(account wm_account.Account, subType string) error {
	traceId := getTraceIdWithSubType(account, subType)
	log.Printf("[%s] 开始处理数据（Stream Slice模式）", traceId)

	// 1. 获取 Fairing 专属的同步状态
	fairingSyncState, err := getFairingState(account, subType)
	if err != nil {
		log.Printf("[%s] 获取同步状态失败: %v", traceId, err)
		return err
	}

	// 2. 检查是否需要跳过或继续
	shouldSkip := false
	if fairingSyncState.UpdatedAt.Add(time.Hour).After(time.Now().UTC()) &&
		fairingSyncState.Status == STATUS_SUCCESS && fairingSyncState.IsAllSlicesCompleted() {
		log.Printf("[%s] 同步时间小于1小时且所有slice已完成，跳过", traceId)
		shouldSkip = true
	}

	// 检查是否有未完成的同步任务
	if !shouldSkip && (fairingSyncState.CurrentSliceDate != nil && !fairingSyncState.IsAllSlicesCompleted()) {
		log.Printf("[%s] 检测到未完成的同步任务，继续之前的进度: %d/%d (%.1f%%)",
			traceId, fairingSyncState.CompletedSlices, fairingSyncState.TotalSlices,
			fairingSyncState.GetProgress())
		shouldSkip = false
	}

	// 检查是否需要创建新的同步任务
	if !shouldSkip && fairingSyncState.CurrentSliceDate == nil {
		// 如果没有当前同步任务且距离上次同步时间足够长，创建新的同步任务
		timeSinceLastSync := time.Hour * 24 // 默认值，如果没有历史同步记录
		if fairingSyncState.LastSyncTime != nil {
			timeSinceLastSync = time.Now().UTC().Sub(*fairingSyncState.LastSyncTime)
		}

		if timeSinceLastSync < time.Hour {
			log.Printf("[%s] 距离上次同步时间过短(%v)，跳过", traceId, timeSinceLastSync)
			shouldSkip = true
		}
	}

	if shouldSkip {
		return nil
	}

	// 3. 对于 responses，开始或继续同步任务
	if subType == "response" {
		log.Printf("[%s] 开始Stream Slice同步任务", traceId)

		// 执行同步，可能会执行多个 slice
		err = syncFairingDataWithFairingState(account, fairingSyncState, subType)
		if err != nil {
			fairingSyncState.Status = STATUS_FAILED
			fairingSyncState.Message = err.Error()
			_ = updateFairingState(account, fairingSyncState, subType)
			return err
		}

		// 检查是否还有未完成的slice，决定是否需要再次调度
		if !fairingSyncState.IsAllSlicesCompleted() {
			log.Printf("[%s] 还有未完成的slice，将在下次调度时继续", traceId)
		} else {
			log.Printf("[%s] 🎉 所有slice处理完成！", traceId)
		}
	}

	log.Printf("[%s] 完成处理数据", traceId)
	return nil
}

// execTask 执行具体的同步任务
func execTask(account wm_account.Account, subType string) error {
	traceId := getTraceIdWithSubType(account, subType)
	log.Printf("[%s] 开始处理数据", traceId)

	// 1. 获取同步状态
	syncState, err := getState(account, subType)
	if err != nil {
		log.Printf("[%s] 获取同步状态失败: %v", traceId, err)
		return err
	}

	// 检查是否需要跳过（1小时内已成功同步）
	if syncState.UpdatedAt.Add(time.Hour).After(time.Now().UTC()) &&
		syncState.Status == STATUS_SUCCESS {
		log.Printf("[%s] 同步时间小于1小时，跳过", traceId)
		return nil
	}

	// 2. 同步数据
	err = syncFairingData(account, syncState, subType)
	if err != nil {
		syncState.Status = STATUS_FAILED
		syncState.Message = err.Error()
		_ = updateSyncState(account, syncState, subType)
		return err
	}

	log.Printf("[%s] 完成处理数据", traceId)
	return nil
}

// createInitialState 创建初始同步状态
func createInitialState(account wm_account.Account, subType string) {
	var initialState SyncState

	switch subType {
	case "question":
		// Questions使用全量同步，不需要LastSyncTime
		initialState = SyncState{
			Status:      STATUS_SUCCESS,
			Message:     "初始状态",
			UpdatedAt:   time.Now().UTC().Add(-2 * time.Hour), // 设为2小时前，确保可以被执行
			RecordCount: 0,
		}
	case "response":
		// Responses使用增量同步，需要LastSyncTime
		initialState = SyncState{
			Status:       STATUS_SUCCESS,
			Message:      "初始状态",
			UpdatedAt:    time.Now().UTC().Add(-2 * time.Hour), // 设为2小时前，确保可以被执行
			RecordCount:  0,
			LastSyncTime: nil, // 首次同步时为nil，表示获取所有数据
		}
	default:
		initialState = SyncState{
			Status:      STATUS_SUCCESS,
			Message:     "初始状态",
			UpdatedAt:   time.Now().UTC().Add(-2 * time.Hour),
			RecordCount: 0,
		}
	}

	updateSyncState(account, initialState, subType)
}

// createInitialFairingState 创建初始的 Fairing 同步状态（支持时间范围同步）
func createInitialFairingState(account wm_account.Account, subType string) {
	var initialFairingState FairingSyncState

	switch subType {
	case "question":
		// Questions使用全量同步，转换为新格式但保持兼容性
		initialFairingState = NewFairingSyncState()
		initialFairingState.IsInitialSync = false // Questions不需要按时间范围同步
	case "response":
		// Responses使用新的时间范围同步逻辑
		initialFairingState = NewFairingSyncState()

		// 获取配置
		config := getFairingConfig()
		initialFairingState.SliceDays = config.SliceDays

		// 可以通过环境变量配置初始同步天数
		if initialDaysStr := os.Getenv("FAIRING_INITIAL_DAYS"); initialDaysStr != "" {
			if initialDays, err := strconv.Atoi(initialDaysStr); err == nil && initialDays > 0 {
				initialFairingState.InitialDays = initialDays
				log.Printf("[%s] 使用配置的初始同步天数: %d", getTraceIdWithSubType(account, subType), initialDays)
			}
		}

		// 可以通过环境变量配置近期同步天数
		if recentDaysStr := os.Getenv("FAIRING_RECENT_DAYS"); recentDaysStr != "" {
			if recentDays, err := strconv.Atoi(recentDaysStr); err == nil && recentDays > 0 {
				initialFairingState.RecentSyncDays = recentDays
				log.Printf("[%s] 使用配置的近期同步天数: %d", getTraceIdWithSubType(account, subType), recentDays)
			}
		}

		// 可以通过环境变量配置slice天数
		if sliceDaysStr := os.Getenv("FAIRING_SLICE_DAYS"); sliceDaysStr != "" {
			if sliceDays, err := strconv.Atoi(sliceDaysStr); err == nil && sliceDays > 0 {
				initialFairingState.SliceDays = sliceDays
				log.Printf("[%s] 使用配置的slice天数: %d", getTraceIdWithSubType(account, subType), sliceDays)
			}
		}
	default:
		initialFairingState = NewFairingSyncState()
		initialFairingState.IsInitialSync = false
	}

	updateFairingState(account, initialFairingState, subType)
}
