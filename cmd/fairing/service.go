package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"
	"wm-func/common/http_request"
	"wm-func/common/state"
	"wm-func/wm_account"
)

// 调用Fairing Questions API（全量获取）
func callFairingQuestionsAPI(account wm_account.Account) ([]FairingQuestion, error) {
	traceId := getTraceIdWithSubType(account, "question")

	headers := map[string]string{
		"Authorization": fmt.Sprintf("%s", account.SecretToken),
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	url := "https://app.fairing.co/api/questions"

	log.Printf("[%s] 调用Fairing Questions API: %s", traceId, url)

	response, err := http_request.Get(url, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("请求 Questions API 失败: %w", err)
	}

	var questions FairingQuestionResponse
	if err := json.Unmarshal(response, &questions); err != nil {
		return nil, fmt.Errorf("解析questions响应失败: %w", err)
	}

	log.Printf("[%s] Questions API响应成功，获取数量: %d", traceId, len(questions.Data))
	return questions.Data, nil
}

// 调用Fairing Responses API（增量分页获取）
func callFairingResponsesAPI(account wm_account.Account, since *time.Time, after string, limit int) (*FairingResponsesResponse, error) {
	traceId := getTraceIdWithSubType(account, "response")

	headers := map[string]string{
		"Authorization": fmt.Sprintf("%s", account.SecretToken),
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	// 构建查询参数
	params := make(map[string]string)
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if after != "" {
		params["after"] = after
	}
	if since != nil {
		params["since"] = since.Format(time.RFC3339)
	}

	apiURL := "https://app.fairing.co/api/responses"

	log.Printf("[%s] 调用Fairing Responses API: %s", traceId, apiURL)
	if since != nil {
		log.Printf("[%s] 增量同步起始时间: %s", traceId, since.Format(time.RFC3339))
	}

	response, err := http_request.Get(apiURL, headers, params)
	if err != nil {
		return nil, fmt.Errorf("请求 Responses API 失败: %w", err)
	}

	var responsesResp FairingResponsesResponse
	if err := json.Unmarshal(response, &responsesResp); err != nil {
		return nil, fmt.Errorf("解析responses响应失败: %w", err)
	}

	log.Printf("[%s] Responses API响应成功，本页数量: %d", traceId, len(responsesResp.Data))
	return &responsesResp, nil
}

// 调用Fairing Responses API（支持时间范围）
func callFairingResponsesAPIWithRange(account wm_account.Account, since, until *time.Time, after string, limit int) (*FairingResponsesResponse, error) {
	traceId := getTraceIdWithSubType(account, "response")

	headers := map[string]string{
		"Authorization": fmt.Sprintf("%s", account.SecretToken),
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	// 构建查询参数
	params := make(map[string]string)
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if after != "" {
		params["after"] = after
	}
	if since != nil {
		params["since"] = since.Format(time.RFC3339)
	}
	if until != nil {
		params["until"] = until.Format(time.RFC3339)
	}

	apiURL := "https://app.fairing.co/api/responses"

	log.Printf("[%s] 调用Fairing Responses API: %s", traceId, apiURL)
	if since != nil && until != nil {
		log.Printf("[%s] 时间范围同步: %s 到 %s",
			traceId, since.Format("2006-01-02"), until.Format("2006-01-02"))
	} else if since != nil {
		log.Printf("[%s] 增量同步起始时间: %s", traceId, since.Format(time.RFC3339))
	}

	response, err := http_request.Get(apiURL, headers, params)
	if err != nil {
		return nil, fmt.Errorf("请求 Responses API 失败: %w", err)
	}

	var responsesResp FairingResponsesResponse
	if err := json.Unmarshal(response, &responsesResp); err != nil {
		return nil, fmt.Errorf("解析responses响应失败: %w", err)
	}

	log.Printf("[%s] Responses API响应成功，本页数量: %d", traceId, len(responsesResp.Data))
	return &responsesResp, nil
}

// 处理Questions数据（全量）
func processQuestionsData(account wm_account.Account, questions []FairingQuestion) ([]FairingData, error) {
	var data []FairingData

	for _, question := range questions {
		data = append(data, question.TransformToFairingData(account.TenantId))
	}

	return data, nil
}

// 处理Responses数据（增量）
func processResponsesData(account wm_account.Account, responses []FairingUserResponse) ([]FairingData, error) {
	var data []FairingData

	for _, response := range responses {
		data = append(data, response.TransformToFairingData(account.TenantId))
	}

	return data, nil
}

// 更新同步状态
func updateSyncState(account wm_account.Account, syncState SyncState, subType string) error {
	traceId := getTraceIdWithSubType(account, subType)
	stateData, err := json.Marshal(syncState)
	if err != nil {
		return fmt.Errorf("[%s] 序列化同步状态失败: %w", traceId, err)
	}

	state.SaveSyncInfo(account.TenantId, account.AccountId, Platform, subType, stateData)
	log.Printf("[%s] 同步状态已更新, 状态: %s", traceId, syncState.Status)

	return nil
}

// 同步Questions数据（全量）
func syncQuestionsData(account wm_account.Account, syncState SyncState) error {
	traceId := getTraceIdWithSubType(account, "question")

	log.Printf("[%s] 开始全量同步Questions数据", traceId)

	// 1. 调用API获取全量数据
	questions, err := callFairingQuestionsAPI(account)
	if err != nil {
		return fmt.Errorf("Questions API调用失败: %w", err)
	}

	// 2. 处理数据
	fairingData, err := processQuestionsData(account, questions)
	if err != nil {
		return fmt.Errorf("Questions数据处理失败: %w", err)
	}

	// 3. 检查数据是否有变化
	currentCount := int64(len(fairingData))
	if currentCount == syncState.RecordCount {
		log.Printf("[%s] Questions数据无变化，跳过保存。记录数: %d", traceId, currentCount)
		syncState.Status = STATUS_SUCCESS
		syncState.Message = fmt.Sprintf("Questions数据无变化，记录数: %d", currentCount)
		syncState.UpdatedAt = time.Now().UTC().Truncate(time.Millisecond)
		updateSyncState(account, syncState, "question")
		return nil
	}

	// 4. 保存数据
	if len(fairingData) > 0 {
		if err = saveFairingData(account, fairingData, "question"); err != nil {
			return fmt.Errorf("Questions数据保存失败: %w", err)
		}
	}

	// 5. 更新状态
	syncState.Status = STATUS_SUCCESS
	syncState.Message = fmt.Sprintf("Questions全量同步完成，共%d条记录（上次: %d）", currentCount, syncState.RecordCount)
	syncState.RecordCount = currentCount
	syncState.UpdatedAt = time.Now().UTC().Truncate(time.Millisecond)

	if err := updateSyncState(account, syncState, "question"); err != nil {
		return fmt.Errorf("Questions状态更新失败: %w", err)
	}

	log.Printf("[%s] Questions全量同步完成，共%d条记录", traceId, len(fairingData))
	return nil
}

// 基于时间范围的Responses数据同步（新版本 - 轻量级状态）
func syncResponsesDataWithTimeRange(account wm_account.Account, fairingSyncState FairingSyncState) error {
	traceId := getTraceIdWithSubType(account, "response")
	config := getFairingConfig()

	log.Printf("[%s] 开始基于Stream Slice的Responses数据同步", traceId)

	// 1. 初始化同步范围（如果尚未初始化）
	if fairingSyncState.SyncStartDate == nil || fairingSyncState.CurrentSliceDate == nil {
		log.Printf("[%s] 初始化同步范围...", traceId)
		if err := fairingSyncState.InitializeSyncRange(); err != nil {
			return fmt.Errorf("初始化同步范围失败: %w", err)
		}
		log.Printf("[%s] 已初始化同步范围: %s 到 %s，共 %d 个时间片段",
			traceId,
			fairingSyncState.SyncStartDate.Format("2006-01-02"),
			fairingSyncState.SyncEndDate.Format("2006-01-02"),
			fairingSyncState.TotalSlices)

		// 保存更新后的状态
		if err := updateFairingState(account, fairingSyncState, "response"); err != nil {
			return fmt.Errorf("保存初始化状态失败: %w", err)
		}
	}

	if fairingSyncState.TotalSlices == 0 {
		log.Printf("[%s] 没有需要同步的时间片段", traceId)
		fairingSyncState.Status = STATUS_SUCCESS
		fairingSyncState.Message = "没有需要同步的数据"
		return nil
	}

	log.Printf("[%s] 开始处理时间片段（当前进度: %d/%d, %.1f%%）",
		traceId, fairingSyncState.CompletedSlices,
		fairingSyncState.TotalSlices, fairingSyncState.GetProgress())

	// 2. 循环处理时间片段
	processedInThisRun := 0
	maxSlicesPerRun := config.MaxSlicesPerRun

	for {
		// 获取当前需要处理的时间范围
		since, until := fairingSyncState.GetCurrentSliceRange()
		if since == nil || until == nil {
			log.Printf("[%s] 所有时间片段已完成", traceId)
			break
		}

		// 检查是否达到单次运行限制
		if processedInThisRun >= maxSlicesPerRun {
			log.Printf("[%s] 已达到单次运行限制(%d个slice)，保存进度并退出", traceId, maxSlicesPerRun)
			fairingSyncState.Status = "PARTIAL_SUCCESS"
			fairingSyncState.Message = fmt.Sprintf("已处理 %d 个slice，进度: %.1f%%",
				processedInThisRun, fairingSyncState.GetProgress())
			return nil
		}

		sliceIndex := fairingSyncState.CompletedSlices + 1
		log.Printf("[%s] 处理时间片段 %d/%d: %s 到 %s",
			traceId, sliceIndex, fairingSyncState.TotalSlices,
			since.Format("2006-01-02"),
			until.Format("2006-01-02"))

		// 3. 同步当前时间片段的数据
		sliceRecordCount, err := syncSingleTimeSliceByRange(account, since, until, config)
		if err != nil {
			log.Printf("[%s] 时间片段 %d 同步失败: %v", traceId, sliceIndex, err)
			fairingSyncState.Status = STATUS_FAILED
			fairingSyncState.Message = fmt.Sprintf("时间片段 %d 同步失败: %v", sliceIndex, err)
			return err
		}

		// 4. 标记片段完成并移动到下一个片段
		fairingSyncState.CompleteCurrentSlice(sliceRecordCount)
		processedInThisRun++

		log.Printf("[%s] 时间片段 %d 完成，记录数: %d，总进度: %.1f%% (%d/%d)",
			traceId, sliceIndex, sliceRecordCount,
			fairingSyncState.GetProgress(), fairingSyncState.CompletedSlices, fairingSyncState.TotalSlices)

		// 5. 立即保存状态，确保数据一致性（每个slice完成后立即保存）
		if err := updateFairingState(account, fairingSyncState, "response"); err != nil {
			log.Printf("[%s] 保存slice状态失败: %v", traceId, err)
			// 状态保存失败，但数据已保存，标记为部分成功
			fairingSyncState.Status = "PARTIAL_SUCCESS"
			fairingSyncState.Message = fmt.Sprintf("Slice %d 数据已保存但状态更新失败: %v", sliceIndex, err)
			return fmt.Errorf("slice状态保存失败: %w", err)
		}

		// 简短休息，避免API限流
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	// 6. 更新最终状态
	if fairingSyncState.IsAllSlicesCompleted() {
		fairingSyncState.Status = STATUS_SUCCESS
		if fairingSyncState.IsInitialSync {
			fairingSyncState.Message = fmt.Sprintf("首次同步完成，共处理%d个时间片段，%d条记录",
				fairingSyncState.TotalSlices, fairingSyncState.RecordCount)
		} else {
			fairingSyncState.Message = fmt.Sprintf("增量同步完成，共处理%d个时间片段，%d条记录",
				fairingSyncState.TotalSlices, fairingSyncState.RecordCount)
		}
		log.Printf("[%s] 🎉 所有时间片段同步完成！总记录数: %d", traceId, fairingSyncState.RecordCount)
	} else {
		fairingSyncState.Status = "PARTIAL_SUCCESS"
		fairingSyncState.Message = fmt.Sprintf("部分完成，本次处理%d个slice，总进度: %.1f%%",
			processedInThisRun, fairingSyncState.GetProgress())
	}

	// 7. 保存最终状态（最后一次确认性保存）
	if err := updateFairingState(account, fairingSyncState, "response"); err != nil {
		log.Printf("[%s] 保存最终状态失败: %v", traceId, err)
		// 不返回错误，因为每个slice都已经保存过状态了
	}

	log.Printf("[%s] Stream Slice同步完成，本次处理: %d个slice，累计记录: %d条",
		traceId, processedInThisRun, fairingSyncState.RecordCount)
	return nil
}

// 同步指定时间范围的数据（不再依赖TimeSlice结构体）
func syncSingleTimeSliceByRange(account wm_account.Account, since, until *time.Time, config FairingConfig) (int64, error) {
	traceId := getTraceIdWithSubType(account, "response")

	var allResponses []FairingUserResponse
	var after string
	totalPages := 0

	// 分页获取指定时间范围的数据
	for {
		totalPages++

		// 调用支持时间范围的API
		responsesResp, err := callFairingResponsesAPIWithRange(account, since, until, after, config.ResponsesPageSize)
		if err != nil {
			return 0, fmt.Errorf("第%d页API调用失败: %w", totalPages, err)
		}

		// 收集数据
		allResponses = append(allResponses, responsesResp.Data...)

		// 检查是否有下一页
		if responsesResp.Next == nil || *responsesResp.Next == "" {
			break
		}

		// 解析next URL获取after参数
		nextURL, err := url.Parse(*responsesResp.Next)
		if err != nil {
			log.Printf("[%s] 解析next URL失败: %v", traceId, err)
			break
		}
		after = nextURL.Query().Get("after")

		// 分页保护
		if totalPages > config.MaxPages {
			log.Printf("[%s] 达到最大页数限制(%d)，停止分页", traceId, config.MaxPages)
			break
		}

		// 速率限制
		time.Sleep(time.Duration(1000/config.RateLimit) * time.Millisecond)
	}

	// 处理和保存数据
	recordCount := int64(len(allResponses))
	if recordCount > 0 {
		// 转换数据
		fairingData, err := processResponsesData(account, allResponses)
		if err != nil {
			return 0, fmt.Errorf("数据处理失败: %w", err)
		}

		// 保存数据到数据库
		// 注意：只有数据成功保存后，外层循环才会标记slice完成并更新状态
		// 这确保了数据库数据和同步状态的一致性
		if err = saveFairingData(account, fairingData, "response"); err != nil {
			return 0, fmt.Errorf("数据保存失败: %w", err)
		}
	}

	return recordCount, nil
}

// 同步Responses数据（增量分页）
func syncResponsesData(account wm_account.Account, syncState SyncState) error {
	traceId := getTraceIdWithSubType(account, "response")
	config := getFairingConfig()

	log.Printf("[%s] 开始增量同步Responses数据", traceId)

	var allResponses []FairingUserResponse
	var after string
	totalPages := 0

	// 确定增量同步的起始时间
	var since *time.Time
	if syncState.LastSyncTime != nil {
		since = syncState.LastSyncTime
		log.Printf("[%s] 增量同步起始时间: %s", traceId, since.Format(time.RFC3339))
	} else {
		log.Printf("[%s] 首次同步，获取所有数据", traceId)
	}

	// 分页获取数据
	for {
		totalPages++

		// 调用API
		responsesResp, err := callFairingResponsesAPI(account, since, after, config.ResponsesPageSize)
		if err != nil {
			return fmt.Errorf("第%d页Responses API调用失败: %w", totalPages, err)
		}

		// 收集数据
		allResponses = append(allResponses, responsesResp.Data...)

		log.Printf("[%s] 第%d页处理完成，本页%d条，累计%d条",
			traceId, totalPages, len(responsesResp.Data), len(allResponses))

		// 检查是否有下一页
		if responsesResp.Next == nil || *responsesResp.Next == "" {
			log.Printf("[%s] 已获取所有页面，共%d页", traceId, totalPages)
			break
		}

		// 解析next URL获取after参数
		nextURL, err := url.Parse(*responsesResp.Next)
		if err != nil {
			log.Printf("[%s] 解析next URL失败: %v", traceId, err)
			break
		}
		after = nextURL.Query().Get("after")

		// 分页保护
		if totalPages > config.MaxPages {
			log.Printf("[%s] 达到最大页数限制(%d)，停止分页", traceId, config.MaxPages)
			break
		}

		// 速率限制
		time.Sleep(time.Duration(1000/config.RateLimit) * time.Millisecond)
	}

	// 处理收集到的所有数据
	if len(allResponses) > 0 {
		fairingData, err := processResponsesData(account, allResponses)
		if err != nil {
			return fmt.Errorf("Responses数据处理失败: %w", err)
		}

		// 保存数据
		if err = saveFairingData(account, fairingData, "response"); err != nil {
			return fmt.Errorf("Responses数据保存失败: %w", err)
		}

		log.Printf("[%s] 成功保存%d条Responses记录", traceId, len(fairingData))
	}

	// 更新同步状态
	now := time.Now().UTC().Truncate(time.Millisecond)
	syncState.Status = STATUS_SUCCESS
	syncState.Message = fmt.Sprintf("Responses增量同步完成，共%d页，%d条新记录", totalPages, len(allResponses))
	syncState.RecordCount += int64(len(allResponses))
	syncState.LastSyncTime = &now // 更新最后同步时间
	syncState.UpdatedAt = now

	if err := updateSyncState(account, syncState, "response"); err != nil {
		return fmt.Errorf("Responses状态更新失败: %w", err)
	}

	log.Printf("[%s] Responses增量同步完成，共%d条新记录", traceId, len(allResponses))
	return nil
}

// 主同步函数 - 根据数据类型选择同步方式（支持新的状态结构）
func syncFairingDataWithFairingState(account wm_account.Account, fairingSyncState FairingSyncState, dataType string) error {
	switch dataType {
	case "question":
		// Questions 继续使用原来的全量同步逻辑（转换状态格式）
		oldState := SyncState{
			Status:       fairingSyncState.Status,
			Message:      fairingSyncState.Message,
			UpdatedAt:    fairingSyncState.UpdatedAt,
			RecordCount:  fairingSyncState.RecordCount,
			LastSyncTime: fairingSyncState.LastSyncTime,
		}
		return syncQuestionsData(account, oldState)
	case "response":
		// Responses 使用新的基于时间范围的同步逻辑
		return syncResponsesDataWithTimeRange(account, fairingSyncState)
	default:
		return fmt.Errorf("不支持的数据类型: %s", dataType)
	}
}

// 主同步函数 - 根据数据类型选择同步方式（兼容旧接口）
func syncFairingData(account wm_account.Account, syncState SyncState, dataType string) error {
	// 转换为新的状态格式
	fairingSyncState := convertOldSyncState(syncState)
	return syncFairingDataWithFairingState(account, fairingSyncState, dataType)
}
