package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/common/http_request"
	"wm-func/common/state"
	"wm-func/wm_account"

	"gorm.io/gorm/clause"
)

const (
	STATUS_SUCCESS = "SUCCESS"
	STATUS_FAILED  = "FAILED"

	// Meta API 相关常量
	META_API_VERSION  = "v23.0"
	BASE_URL          = "https://graph.facebook.com"
	POLL_INTERVAL     = 5 * time.Second
	MAX_POLL_ATTEMPTS = 20
)

// SyncState 同步状态结构体
type SyncState struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
	// 可以根据需要添加其他字段
}

// MetaReportResponse 创建报告的响应
type MetaReportResponse struct {
	ReportRunId string `json:"report_run_id"`
}

// MetaReportStatus 报告状态响应
type MetaReportStatus struct {
	Id                     string `json:"id"`
	AccountId              string `json:"account_id"`
	TimeRef                int64  `json:"time_ref"`
	TimeCompleted          int64  `json:"time_completed"`
	AsyncStatus            string `json:"async_status"`
	AsyncPercentCompletion int    `json:"async_percent_completion"`
	DateStart              string `json:"date_start"`
	DateStop               string `json:"date_stop"`
}

// MetaInsightData 广告洞察数据
type MetaInsightData struct {
	AccountId       string  `json:"account_id"`
	AccountName     string  `json:"account_name"`
	AccountCurrency string  `json:"account_currency"`
	CampaignId      string  `json:"campaign_id"`
	CampaignName    string  `json:"campaign_name"`
	AdsetId         string  `json:"adset_id"`
	AdsetName       string  `json:"adset_name"`
	AdId            string  `json:"ad_id"`
	AdName          string  `json:"ad_name"`
	DateStart       string  `json:"date_start"`
	DateStop        string  `json:"date_stop"`
	Impressions     int64   `json:"impressions,string"`
	Clicks          int64   `json:"clicks,string"`
	Spend           float64 `json:"spend,string"`
	Reach           int64   `json:"reach,string"`
	Frequency       float64 `json:"frequency,string"`
	CTR             float64 `json:"ctr,string"`
	CPC             float64 `json:"cpc,string"`
	CPM             float64 `json:"cpm,string"`
	CreatedTime     string  `json:"created_time"`
	UpdatedTime     string  `json:"updated_time"`
}

// MetaInsightsResponse 洞察数据响应
type MetaInsightsResponse struct {
	Data   []MetaData `json:"data"`
	Paging struct {
		Cursors struct {
			Before string `json:"before"`
			After  string `json:"after"`
		} `json:"cursors"`
		Next string `json:"next"`
	} `json:"paging"`
}

// MetaAirbyteData airbyte 数据库存储结构体
type MetaAirbyteData struct {
	TenantId            int64  `gorm:"column:wm_tenant_id"`
	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
	AirbyteData         []byte `gorm:"column:_airbyte_data"`
	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
	AirbyteGenerationId int64  `gorm:"column:_airbyte_generation_id"`
}

// 生成 Raw ID
func generateRawId(data MetaData) string {
	// 使用账户ID、日期和创建时间生成唯一标识
	source := fmt.Sprintf("%s|%s|%s", data.DateStart, data.AccountId, data.AdId)
	return source
}

// 转换函数 - 将 MetaData 转换为 MetaAirbyteData
func (m MetaData) TransformToAirbyteData(tenantId int64) MetaAirbyteData {
	// 将 Meta 数据序列化为 JSON
	jsonData, _ := json.Marshal(m)

	// 生成唯一的 Raw ID
	rawId := generateRawId(m)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return MetaAirbyteData{
		TenantId:            tenantId,
		AirbyteRawId:        rawId,
		AirbyteData:         jsonData,
		AirbyteExtractedAt:  now,
		AirbyteLoadedAt:     now,
		AirbyteMeta:         `{}`,
		AirbyteGenerationId: 0,
	}
}

// getState 获取账户的同步状态
func getState(account wm_account.Account) (SyncState, error) {
	syncInfo := state.GetSyncInfo(account.TenantId, account.AccountId, Platform, SubType)

	var syncState SyncState
	if syncInfo == nil {

		syncState.UpdatedAt = time.Now().Add(time.Hour * 24 * -1)

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

// syncAdMetrics 同步广告数据 - 使用stream slice按天拉取
func syncAdMetrics(account wm_account.Account, syncState SyncState) error {
	log.Printf("[%s] 开始同步广告数据", account.GetTraceId())

	// 生成日期切片，按天处理
	dateSlices := generateDateSlices()
	log.Printf("[%s] 生成 %d 个日期切片，准备按天同步数据", account.GetTraceId(), len(dateSlices))

	successCount := 0
	totalSlices := len(dateSlices)

	// 按天循环处理数据
	for i, date := range dateSlices {
		log.Printf("[%s] 开始处理第 %d/%d 天的数据: %s", account.GetTraceId(), i+1, totalSlices, date)

		err := syncAdMetricsForDate(account, date)
		if err != nil {
			log.Printf("[%s] 处理日期 %s 的数据失败: %v", account.GetTraceId(), date, err)
			// 继续处理下一天的数据，不中断整个流程
			continue
		}

		successCount++
		log.Printf("[%s] 成功处理日期 %s 的数据 (%d/%d)", account.GetTraceId(), date, successCount, totalSlices)
	}

	// 更新同步状态
	if successCount == totalSlices {
		syncState.Status = STATUS_SUCCESS
		syncState.Message = fmt.Sprintf("同步成功，处理了 %d 天的数据", successCount)
	} else {
		syncState.Status = STATUS_SUCCESS // 部分成功也认为是成功状态
		syncState.Message = fmt.Sprintf("部分同步成功，成功处理 %d/%d 天的数据", successCount, totalSlices)
	}

	err := updateSyncState(account, syncState)
	if err != nil {
		log.Printf("[%s] 更新同步状态失败: %v", account.GetTraceId(), err)
	}

	log.Printf("[%s] 广告数据同步完成，成功处理 %d/%d 天的数据", account.GetTraceId(), successCount, totalSlices)
	return nil
}

// syncAdMetricsForDate 同步指定日期的广告数据
func syncAdMetricsForDate(account wm_account.Account, date string) error {
	// Step 1: 创建指定日期的异步任务
	reportRunId, err := createAdInsightsReportForDate(account, date)
	if err != nil {
		return fmt.Errorf("创建日期 %s 的广告洞察报告失败: %w", date, err)
	}

	log.Printf("[%s] 创建日期 %s 的报告任务成功，report_run_id: %s", account.GetTraceId(), date, reportRunId)

	// Step 2: 轮询任务状态直到完成
	err = waitForReportCompletion(account, reportRunId)
	if err != nil {
		return fmt.Errorf("等待日期 %s 的报告完成失败: %w", date, err)
	}

	log.Printf("[%s] 日期 %s 的报告任务已完成", account.GetTraceId(), date)

	time.Sleep(time.Second * 2) // 稍微减少等待时间
	// Step 3: 获取并存储结果
	err = fetchAndStoreInsights(account, reportRunId)
	if err != nil {
		return fmt.Errorf("获取并存储日期 %s 的洞察数据失败: %w", date, err)
	}

	return nil
}

// createAdInsightsReport 创建广告洞察报告 - 原有函数保留
func createAdInsightsReport(account wm_account.Account) (string, error) {
	// 构建请求URL
	apiUrl := fmt.Sprintf("%s/%s/act_%s/insights", BASE_URL, META_API_VERSION, account.AccountId)

	// 设置查询参数
	params := map[string]string{
		"access_token":   account.AccessToken,
		"fields":         "account_id,actions,action_values,account_name,account_currency,campaign_id,campaign_name,adset_id,adset_name,ad_id,ad_name,date_start,date_stop,impressions,clicks,spend,reach,frequency,ctr,cpc,cpm,created_time,updated_time",
		"time_increment": "1",
		"time_range":     fmt.Sprintf(`{"since":"%s","until":"%s"}`, getStartDate(), getEndDate()),
		"level":          "ad",
	}

	// 发送POST请求
	respData, err := http_request.Post(apiUrl, nil, params, nil)
	if err != nil {
		return "", fmt.Errorf("发送创建报告请求失败: %w", err)
	}

	// 解析响应
	var reportResp MetaReportResponse
	if err := json.Unmarshal(respData, &reportResp); err != nil {
		return "", fmt.Errorf("解析创建报告响应失败: %w, 响应内容: %s", err, string(respData))
	}

	if reportResp.ReportRunId == "" {
		return "", fmt.Errorf("创建报告响应中没有report_run_id, 响应内容: %s", string(respData))
	}

	return reportResp.ReportRunId, nil
}

// createAdInsightsReportForDate 为指定日期创建广告洞察报告
func createAdInsightsReportForDate(account wm_account.Account, date string) (string, error) {
	// 构建请求URL
	apiUrl := fmt.Sprintf("%s/%s/act_%s/insights", BASE_URL, META_API_VERSION, account.AccountId)

	// 设置查询参数 - 单日数据，since和until都是同一天
	params := map[string]string{
		"access_token":   account.AccessToken,
		"fields":         "account_id,actions,action_values,account_name,account_currency,campaign_id,campaign_name,adset_id,adset_name,ad_id,ad_name,date_start,date_stop,impressions,clicks,spend,reach,frequency,ctr,cpc,cpm,created_time,updated_time",
		"time_increment": "1",
		"time_range":     fmt.Sprintf(`{"since":"%s","until":"%s"}`, date, date),
		"level":          "ad",
	}

	// 发送POST请求
	respData, err := http_request.Post(apiUrl, nil, params, nil)
	if err != nil {
		return "", fmt.Errorf("发送创建报告请求失败: %w", err)
	}

	// 解析响应
	var reportResp MetaReportResponse
	if err := json.Unmarshal(respData, &reportResp); err != nil {
		return "", fmt.Errorf("解析创建报告响应失败: %w, 响应内容: %s", err, string(respData))
	}

	if reportResp.ReportRunId == "" {
		return "", fmt.Errorf("创建报告响应中没有report_run_id, 响应内容: %s", string(respData))
	}

	return reportResp.ReportRunId, nil
}

// waitForReportCompletion 等待报告完成
func waitForReportCompletion(account wm_account.Account, reportRunId string) error {
	apiUrl := fmt.Sprintf("%s/%s/%s", BASE_URL, META_API_VERSION, reportRunId)
	params := map[string]string{
		"access_token": account.AccessToken,
	}

	for attempt := 0; attempt < MAX_POLL_ATTEMPTS; attempt++ {
		respData, err := http_request.Get(apiUrl, nil, params)
		if err != nil {
			return fmt.Errorf("查询报告状态失败: %w", err)
		}

		var status MetaReportStatus
		if err := json.Unmarshal(respData, &status); err != nil {
			return fmt.Errorf("解析报告状态响应失败: %w, 响应内容: %s", err, string(respData))
		}

		log.Printf("[%s] 报告状态: %s, 完成度: %d%%", account.GetTraceId(), status.AsyncStatus, status.AsyncPercentCompletion)

		if status.AsyncStatus == "Job Completed" {
			return nil
		}

		if status.AsyncStatus == "Job Failed" {
			return fmt.Errorf("报告任务失败")
		}

		// 等待一段时间后再次查询
		time.Sleep(POLL_INTERVAL)
	}

	return fmt.Errorf("报告任务超时，已尝试 %d 次", MAX_POLL_ATTEMPTS)
}

// fetchAndStoreInsights 获取并存储洞察数据
func fetchAndStoreInsights(account wm_account.Account, reportRunId string) error {
	apiUrl := fmt.Sprintf("%s/%s/%s/insights", BASE_URL, META_API_VERSION, reportRunId)
	params := map[string]string{
		"access_token": account.AccessToken,
		"limit":        "250",
	}

	allData := make([]MetaData, 0)
	currentUrl := apiUrl

	// 处理分页获取所有数据
	for currentUrl != "" {
		respData, err := http_request.Get(currentUrl, nil, params)
		if err != nil {
			return fmt.Errorf("获取洞察数据失败: %w", err)
		}

		var insightsResp MetaInsightsResponse
		if err := json.Unmarshal(respData, &insightsResp); err != nil {
			return fmt.Errorf("解析洞察数据响应失败: %w, 响应内容: %s", err, string(respData))
		}

		// 添加当前页数据到总数据
		allData = append(allData, insightsResp.Data...)
		log.Printf("[%s] 获取到洞察数据 %d 条， 总计 %d 条", account.GetTraceId(), len(insightsResp.Data), len(allData))

		// 检查是否有下一页
		currentUrl = insightsResp.Paging.Next
		if currentUrl != "" {
			// 从next URL中提取参数，避免重复设置access_token
			params = make(map[string]string)
		}
	}

	log.Printf("[%s] 总共获取到洞察数据 %d 条", account.GetTraceId(), len(allData))

	// 存储数据到数据库
	if len(allData) > 0 {
		err := storeInsightsData(account, allData)
		if err != nil {
			return fmt.Errorf("存储洞察数据失败: %w", err)
		}
	}

	return nil
}

// storeInsightsData 存储洞察数据到数据库 (airbyte 格式)
func storeInsightsData(account wm_account.Account, data []MetaData) error {
	if len(data) == 0 {
		return nil
	}

	db := airbyte_db.GetDB()

	// 转换为 airbyte 格式
	airbyteRecords := make([]MetaAirbyteData, len(data))
	for i, item := range data {
		airbyteRecords[i] = item.TransformToAirbyteData(account.TenantId)
	}

	// 生成动态表名 (根据租户ID)
	tableName := "airbyte_destination_v2.raw_facebook_marketing_ads_insights"

	// 批量插入数据，使用 OnConflict 处理重复数据
	if err := db.Table(tableName).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "wm_tenant_id"}, {Name: "_airbyte_raw_id"}},
			UpdateAll: true,
		}).CreateInBatches(airbyteRecords, 500).Error; err != nil {
		return fmt.Errorf("批量插入洞察数据失败: %w", err)
	}

	log.Printf("[%s] 成功存储 %d 条洞察数据到表 %s", account.GetTraceId(), len(airbyteRecords), tableName)
	return nil
}

// generateDateSlices 生成日期切片（从30天前到昨天，每天一个切片）
func generateDateSlices() []string {
	dates := make([]string, 0)
	now := time.Now()

	// 从30天前开始，到昨天结束
	for i := 5; i >= 1; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)
	}

	return dates
}

// getStartDate 获取开始日期（过去30天）
func getStartDate() string {
	//return "2025-07-31"
	return time.Now().AddDate(0, 0, -30).Format("2006-01-02")
}

// getEndDate 获取结束日期（昨天）
func getEndDate() string {
	//return "2025-07-31"
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
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

type MetaData struct {
	AccountCurrency string `json:"account_currency"`
	AccountId       string `json:"account_id"`
	AdId            string `json:"ad_id"`
	AdsetId         string `json:"adset_id"`
	CampaignId      string `json:"campaign_id"`
	AccountName     string `json:"account_name"`
	ActionValues    []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"action_values"`
	Actions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"actions"`
	AttributionSetting    string `json:"attribution_setting"`
	CanvasAvgViewPercent  string `json:"canvas_avg_view_percent"`
	CanvasAvgViewTime     string `json:"canvas_avg_view_time"`
	Clicks                string `json:"clicks"`
	ConversionRateRanking string `json:"conversion_rate_ranking"`
	ConversionValues      []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"conversion_values"`
	Conversions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"conversions"`
	CostPer15SecVideoView []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_15_sec_video_view"`
	CostPer2SecContinuousVideoView []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_2_sec_continuous_video_view"`
	CostPerActionType []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_action_type"`
	CostPerConversion []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_conversion"`
	CostPerInlineLinkClick      string `json:"cost_per_inline_link_click"`
	CostPerInlinePostEngagement string `json:"cost_per_inline_post_engagement"`
	CostPerOutboundClick        []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_outbound_click"`
	CostPerThruplay []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_thruplay"`
	CostPerUniqueActionType []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_unique_action_type"`
	CostPerUniqueClick           string `json:"cost_per_unique_click"`
	CostPerUniqueInlineLinkClick string `json:"cost_per_unique_inline_link_click"`
	CostPerUniqueOutboundClick   []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_unique_outbound_click"`
	Cpc                   string `json:"cpc"`
	Cpm                   string `json:"cpm"`
	Cpp                   string `json:"cpp"`
	CreatedTime           string `json:"created_time"`
	Ctr                   string `json:"ctr"`
	DateStart             string `json:"date_start"`
	DateStop              string `json:"date_stop"`
	EngagementRateRanking string `json:"engagement_rate_ranking"`
	Frequency             string `json:"frequency"`
	FullViewImpressions   string `json:"full_view_impressions"`
	FullViewReach         string `json:"full_view_reach"`
	Impressions           string `json:"impressions"`
	InlineLinkClickCtr    string `json:"inline_link_click_ctr"`
	InlineLinkClicks      string `json:"inline_link_clicks"`
	InlinePostEngagement  string `json:"inline_post_engagement"`
	MobileAppPurchaseRoas []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"mobile_app_purchase_roas"`
	Objective        string `json:"objective"`
	OptimizationGoal string `json:"optimization_goal"`
	OutboundClicks   []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"outbound_clicks"`
	OutboundClicksCtr []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"outbound_clicks_ctr"`
	PurchaseRoas []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"purchase_roas"`
	QualityRanking string `json:"quality_ranking"`
	Reach          string `json:"reach"`
	SocialSpend    string `json:"social_spend"`
	Spend          string `json:"spend"`
	UniqueActions  []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"unique_actions"`
	UniqueClicks             string `json:"unique_clicks"`
	UniqueCtr                string `json:"unique_ctr"`
	UniqueInlineLinkClickCtr string `json:"unique_inline_link_click_ctr"`
	UniqueInlineLinkClicks   string `json:"unique_inline_link_clicks"`
	UniqueLinkClicksCtr      string `json:"unique_link_clicks_ctr"`
	UniqueOutboundClicks     []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"unique_outbound_clicks"`
	UniqueOutboundClicksCtr []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"unique_outbound_clicks_ctr"`
	UpdatedTime              string `json:"updated_time"`
	Video15SecWatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_15_sec_watched_actions"`
	Video30SecWatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_30_sec_watched_actions"`
	VideoAvgTimeWatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_avg_time_watched_actions"`
	VideoContinuous2SecWatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_continuous_2_sec_watched_actions"`
	VideoP100WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p100_watched_actions"`
	VideoP25WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p25_watched_actions"`
	VideoP50WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p50_watched_actions"`
	VideoP75WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p75_watched_actions"`
	VideoP95WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p95_watched_actions"`
	VideoPlayActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_play_actions"`
	WebsiteCtr []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"website_ctr"`
	WebsitePurchaseRoas []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"website_purchase_roas"`
}
