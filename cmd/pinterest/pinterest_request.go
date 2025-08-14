package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"wm-func/common/http_request"
)

const (
	// Pinterest API 基础配置
	PinterestAPIBase = "https://api.pinterest.com/v5"
	DefaultTimeout   = 30 * time.Second
	MaxRetries       = 3
	RetryDelay       = 5 * time.Second

	// 分页和批处理配置
	DefaultPageSize      = 100 // API每页默认大小
	CampaignBatchSize    = 50  // Campaign ID批处理大小
	AdGroupSaveBatchSize = 500 // AdGroup保存批处理大小
	MaxPagesLimit        = 100 // 最大页数限制，防止无限循环
)

// CreateReport 创建Pinterest广告报告
// 返回报告token，用于后续查询报告状态
func (p *Pinterest) CreateReport() (*ReportResponse, error) {
	traceId := p.getTraceId()

	// 固定时间范围：最近180天到今天（UTC时间）
	now := time.Now().UTC()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -180).Format("2006-01-02")

	log.Printf("[%s] 开始创建Pinterest报告，时间范围: %s 到 %s", traceId, startDate, endDate)

	// 获取有效的访问令牌
	token, err := p.GetValidAccessToken()
	if err != nil {
		log.Printf("[%s] 获取访问令牌失败: %v", traceId, err)
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建报告请求
	reportReq := ReportRequest{
		StartDate:    startDate,
		EndDate:      endDate,
		Granularity:  "TOTAL",
		Columns:      []string{"SPEND_IN_DOLLAR"},
		Level:        "CAMPAIGN",
		ReportFormat: "JSON",
	}

	// 序列化请求数据
	reqData, err := json.Marshal(reportReq)
	if err != nil {
		log.Printf("[%s] 序列化报告请求失败: %v", traceId, err)
		return nil, fmt.Errorf("序列化报告请求失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/reports", PinterestAPIBase, p.Account.AccountId)
	log.Printf("[%s] 调用创建报告API: %s", traceId, apiURL)

	// 发送HTTP请求
	respData, err := p.makeHTTPRequestWithRetry("POST", apiURL, headers, nil, reqData)
	if err != nil {
		log.Printf("[%s] 创建报告请求失败: %v", traceId, err)
		return nil, fmt.Errorf("创建报告请求失败: %w", err)
	}

	// 解析响应
	var reportResp ReportResponse
	if err := json.Unmarshal(respData, &reportResp); err != nil {
		log.Printf("[%s] 解析创建报告响应失败: %v", traceId, err)
		return nil, fmt.Errorf("解析创建报告响应失败: %w", err)
	}

	log.Printf("[%s] 报告创建成功，token: %s, 状态: %s", traceId, reportResp.Token, reportResp.ReportStatus)
	return &reportResp, nil
}

// QueryReport 查询报告状态和获取下载链接
// 轮询检查报告是否完成，返回最终的报告响应
func (p *Pinterest) QueryReport(token string) (*ReportResponse, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始查询报告状态，token: %s", traceId, token)

	// 获取有效的访问令牌
	accessToken, err := p.GetValidAccessToken()
	if err != nil {
		log.Printf("[%s] 获取访问令牌失败: %v", traceId, err)
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建查询参数
	params := map[string]string{
		"token": token,
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/reports", PinterestAPIBase, p.Account.AccountId)

	// 轮询查询报告状态
	maxAttempts := 60 // 最多查询60次，每次间隔10秒，总共10分钟
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("[%s] 第%d次查询报告状态", traceId, attempt)

		// 发送查询请求
		respData, err := p.makeHTTPRequestWithRetry("GET", apiURL, headers, params, nil)
		if err != nil {
			log.Printf("[%s] 查询报告状态失败: %v", traceId, err)
			return nil, fmt.Errorf("查询报告状态失败: %w", err)
		}

		// 解析响应
		var reportResp ReportResponse
		if err := json.Unmarshal(respData, &reportResp); err != nil {
			log.Printf("[%s] 解析查询报告响应失败: %v", traceId, err)
			return nil, fmt.Errorf("解析查询报告响应失败: %w", err)
		}

		log.Printf("[%s] 报告状态: %s", traceId, reportResp.ReportStatus)

		// 检查报告状态
		switch reportResp.ReportStatus {
		case "FINISHED":
			log.Printf("[%s] 报告生成完成，下载URL: %s, 大小: %d bytes",
				traceId, reportResp.URL, reportResp.Size)
			return &reportResp, nil
		case "FAILED":
			log.Printf("[%s] 报告生成失败: %s", traceId, reportResp.Message)
			return nil, fmt.Errorf("报告生成失败: %s", reportResp.Message)
		case "IN_PROGRESS", "PENDING":
			// 报告仍在处理中，等待后继续查询
			log.Printf("[%s] 报告仍在处理中，等待10秒后重试", traceId)
			time.Sleep(10 * time.Second)
			continue
		default:
			log.Printf("[%s] 未知报告状态: %s", traceId, reportResp.ReportStatus)
			time.Sleep(10 * time.Second)
			continue
		}
	}

	// 超时处理
	log.Printf("[%s] 查询报告状态超时，已尝试%d次", traceId, maxAttempts)
	return nil, fmt.Errorf("查询报告状态超时，报告可能仍在处理中")
}

// DownloadReport 下载并处理报告数据
// 从URL下载报告文件，解析数据并返回结构化数据
func (p *Pinterest) DownloadReport(reportURL string) (map[string][]ReportData, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始下载报告数据，URL: %s", traceId, reportURL)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	// 发送下载请求
	resp, err := client.Get(reportURL)
	if err != nil {
		log.Printf("[%s] 下载报告失败: %v", traceId, err)
		return nil, fmt.Errorf("下载报告失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Printf("[%s] 下载报告响应错误，状态码: %d", traceId, resp.StatusCode)
		return nil, fmt.Errorf("下载报告响应错误，状态码: %d", resp.StatusCode)
	}
	log.Printf("[%s] 开始解析报告数据", traceId)

	res := map[string][]ReportData{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		panic(err)
	}

	log.Printf("[%s] 报告数据解析成功，记录数: %d", traceId, len(res))
	return res, nil
}

// makeHTTPRequestWithRetry 带重试机制的HTTP请求
func (p *Pinterest) makeHTTPRequestWithRetry(method, url string, headers, params map[string]string, body []byte) ([]byte, error) {
	traceId := p.getTraceId()

	var lastErr error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		log.Printf("[%s] 发送HTTP请求，第%d次尝试: %s %s", traceId, attempt, method, url)

		var respData []byte
		var err error

		if method == "GET" {
			respData, err = http_request.Get(url, headers, params)
		} else if method == "POST" {
			respData, err = http_request.Post(url, headers, params, body)
		} else {
			return nil, fmt.Errorf("不支持的HTTP方法: %s", method)
		}

		if err == nil {
			log.Printf("[%s] HTTP请求成功", traceId)
			return respData, nil
		}

		lastErr = err
		log.Printf("[%s] HTTP请求失败，第%d次尝试: %v", traceId, attempt, err)

		// 如果不是最后一次尝试，等待后重试
		if attempt < MaxRetries {
			log.Printf("[%s] 等待%v后重试", traceId, RetryDelay)
			time.Sleep(RetryDelay)
		}
	}

	return nil, fmt.Errorf("HTTP请求重试%d次后仍然失败: %w", MaxRetries, lastErr)
}

// getTraceId 获取跟踪ID
func (p *Pinterest) getTraceId() string {
	return fmt.Sprintf("%d-%s", p.Account.TenantId, p.Account.AccountId)
}

// ProcessReportData 完整的报告处理流程
// 创建报告 -> 查询状态 -> 下载数据
func (p *Pinterest) ProcessReportData() ([]string, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始完整的报告处理流程", traceId)

	// 1. 创建报告
	createResp, err := p.CreateReport()
	if err != nil {
		return nil, fmt.Errorf("创建报告失败: %w", err)
	}

	// 2. 查询报告状态直到完成
	queryResp, err := p.QueryReport(createResp.Token)
	if err != nil {
		return nil, fmt.Errorf("查询报告状态失败: %w", err)
	}

	// 3. 下载并解析报告数据
	reportData, err := p.DownloadReport(queryResp.URL)
	if err != nil {
		return nil, fmt.Errorf("下载报告数据失败: %w", err)
	}

	log.Printf("[%s] 报告处理流程完成，获取到%d条记录", traceId, len(reportData))

	var res = []string{}
	for key := range reportData {
		res = append(res, key)
	}

	return res, nil
}

// ListCampaigns 获取Pinterest广告活动列表
// ids: 可选的campaign ID列表，如果为空则获取所有campaigns
func (p *Pinterest) ListCampaigns(ids []string) (*CampaignListResponse, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始获取Campaign列表，指定IDs: %v", traceId, ids)

	// 获取有效的访问令牌
	token, err := p.GetValidAccessToken()
	if err != nil {
		log.Printf("[%s] 获取访问令牌失败: %v", traceId, err)
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建查询参数
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", DefaultPageSize),
	}

	// 如果指定了campaign IDs，添加到查询参数中
	if len(ids) > 0 {
		// 将ID数组转换为逗号分隔的字符串
		campaignIds := make([]string, len(ids))
		copy(campaignIds, ids)
		params["campaign_ids"] = strings.Join(campaignIds, ",")
		log.Printf("[%s] 查询指定的Campaign IDs: %s", traceId, params["campaign_ids"])
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/campaigns", PinterestAPIBase, p.Account.AccountId)
	log.Printf("[%s] 调用获取Campaign列表API: %s", traceId, apiURL)

	// 发送HTTP请求
	respData, err := p.makeHTTPRequestWithRetry("GET", apiURL, headers, params, nil)
	if err != nil {
		log.Printf("[%s] 获取Campaign列表请求失败: %v", traceId, err)
		return nil, fmt.Errorf("获取Campaign列表请求失败: %w", err)
	}

	// 解析响应
	var campaignResp CampaignListResponse
	if err := json.Unmarshal(respData, &campaignResp); err != nil {
		log.Printf("[%s] 解析Campaign列表响应失败: %v", traceId, err)
		return nil, fmt.Errorf("解析Campaign列表响应失败: %w", err)
	}

	log.Printf("[%s] 成功获取Campaign列表，共%d个campaigns", traceId, len(campaignResp.Items))

	return &campaignResp, nil
}

// ListAdGroups 获取Pinterest广告组列表，支持bookmark分页
// campaignIds: 可选的campaign ID列表，如果为空则获取所有ad groups
// bookmark: 分页游标，首次请求传空字符串
func (p *Pinterest) ListAdGroups(campaignIds []string, bookmark string) (*AdGroupListResponse, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始获取AdGroup列表 Bookmark: %s", traceId, bookmark)

	// 获取有效的访问令牌
	token, err := p.GetValidAccessToken()
	if err != nil {
		log.Printf("[%s] 获取访问令牌失败: %v", traceId, err)
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建查询参数
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", DefaultPageSize),
	}

	// 如果指定了campaign IDs，添加到查询参数中
	if len(campaignIds) > 0 {
		params["campaign_ids"] = strings.Join(campaignIds, ",")
		log.Printf("[%s] 查询指定的Campaign IDs: %s", traceId, params["campaign_ids"])
	}

	// 如果有bookmark，添加到查询参数中
	if bookmark != "" {
		params["bookmark"] = bookmark
		log.Printf("[%s] 使用bookmark进行分页: %s", traceId, bookmark)
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/ad_groups", PinterestAPIBase, p.Account.AccountId)
	log.Printf("[%s] 调用获取AdGroup列表API: %s", traceId, apiURL)

	// 发送HTTP请求
	respData, err := p.makeHTTPRequestWithRetry("GET", apiURL, headers, params, nil)
	if err != nil {
		log.Printf("[%s] 获取AdGroup列表请求失败: %v", traceId, err)
		return nil, fmt.Errorf("获取AdGroup列表请求失败: %w", err)
	}

	// 解析响应
	var adGroupResp AdGroupListResponse
	if err := json.Unmarshal(respData, &adGroupResp); err != nil {
		log.Printf("[%s] 解析AdGroup列表响应失败: %v", traceId, err)
		return nil, fmt.Errorf("解析AdGroup列表响应失败: %w", err)
	}

	log.Printf("[%s] 成功获取AdGroup列表，共%d个ad groups，下一页bookmark: %s",
		traceId, len(adGroupResp.Items), adGroupResp.Bookmark)

	return &adGroupResp, nil
}

// PullAllAdGroupsAndSave 批量获取所有AdGroup数据并保存到数据库
// 将campaign IDs分批处理，每批最多50个，避免URL过长和API限制
func (p *Pinterest) PullAllAdGroupsAndSave() error {
	traceId := p.getTraceId()
	campaignIds := p.IdForCampaigns

	if len(campaignIds) == 0 {
		log.Printf("[%s] 没有Campaign IDs需要处理", traceId)
		return nil
	}

	log.Printf("[%s] 开始批量获取AdGroup数据，共%d个Campaign IDs", traceId, len(campaignIds))

	const batchSize = CampaignBatchSize
	var totalAdGroups int64
	var errors []error

	// 分批处理campaign IDs
	for i := 0; i < len(campaignIds); i += batchSize {
		end := i + batchSize
		if end > len(campaignIds) {
			end = len(campaignIds)
		}

		batch := campaignIds[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(campaignIds) + batchSize - 1) / batchSize

		log.Printf("[%s] 处理第%d/%d批，包含%d个Campaign IDs", traceId, batchNum, totalBatches, len(batch))

		// 处理当前批次
		if err := p.ListAllAdGroups(batch); err != nil {
			log.Printf("[%s] 第%d批处理失败: %v", traceId, batchNum, err)
			errors = append(errors, fmt.Errorf("批次%d处理失败: %w", batchNum, err))
			continue // 继续处理下一批，不中断整个流程
		}

		log.Printf("[%s] 第%d批处理完成", traceId, batchNum)
	}

	// 汇总结果
	if len(errors) > 0 {
		log.Printf("[%s] 批量处理完成，但有%d个批次失败", traceId, len(errors))
		return fmt.Errorf("批量处理部分失败，%d个错误: %v", len(errors), errors)
	}

	log.Printf("[%s] 批量获取AdGroup数据完成，总计获取%d个AdGroup", traceId, totalAdGroups)
	return nil
}

// ListAllAdGroups 获取指定Campaign的所有AdGroup数据并保存到数据库
// 支持bookmark分页，自动处理所有分页数据
func (p *Pinterest) ListAllAdGroups(campaignIds []string) error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始获取AdGroup数据，Campaign数量: %d", traceId, len(campaignIds))

	var allAdGroups []AdGroup
	var adGroupIds = p.IdForAdGroups
	bookmark := ""
	pageCount := 0
	const maxPages = MaxPagesLimit // 防止无限循环的安全限制

	// 分页获取所有数据
	for pageCount < maxPages {
		pageCount++
		log.Printf("[%s] 获取第%d页AdGroup数据", traceId, pageCount)

		// 获取当前页数据，带重试机制
		resp, err := p.listAdGroupsWithRetry(campaignIds, bookmark, 3)
		if err != nil {
			log.Printf("[%s] 获取第%d页AdGroup数据失败: %v", traceId, pageCount, err)
			return fmt.Errorf("获取第%d页AdGroup数据失败: %w", pageCount, err)
		}

		// 添加到总结果中
		if len(resp.Items) > 0 {
			allAdGroups = append(allAdGroups, resp.Items...)
			// 收集AdGroup IDs
			for _, adGroup := range resp.Items {
				adGroupIds = append(adGroupIds, adGroup.Id)
			}
			log.Printf("[%s] 第%d页获取到%d个AdGroup，累计%d个",
				traceId, pageCount, len(resp.Items), len(allAdGroups))
		} else {
			log.Printf("[%s] 第%d页没有获取到AdGroup数据", traceId, pageCount)
		}

		// 检查是否还有下一页
		if resp.Bookmark == "" {
			log.Printf("[%s] 没有更多页面，数据获取完成", traceId)
			break
		}

		// 更新bookmark用于下一页
		bookmark = resp.Bookmark
	}

	// 检查是否达到最大页数限制
	if pageCount >= maxPages {
		log.Printf("[%s] 警告：达到最大页数限制(%d)，可能还有未获取的数据", traceId, maxPages)
	}

	// 存储AdGroup IDs到结构体中
	p.IdForAdGroups = adGroupIds
	log.Printf("[%s] 共获取到%d个AdGroup，%d个AdGroup IDs已存储", traceId, len(allAdGroups), len(adGroupIds))

	// 分批保存到数据库，避免单次插入数据过多
	if len(allAdGroups) > 0 {
		if err := p.saveAdGroupsInBatches(allAdGroups, traceId); err != nil {
			return fmt.Errorf("保存AdGroup数据到数据库失败: %w", err)
		}
		log.Printf("[%s] 成功保存%d个AdGroup到数据库", traceId, len(allAdGroups))
	} else {
		log.Printf("[%s] 没有AdGroup数据需要保存", traceId)
	}

	return nil
}

// listAdGroupsWithRetry 带重试机制的AdGroup获取方法
func (p *Pinterest) listAdGroupsWithRetry(campaignIds []string, bookmark string, maxRetries int) (*AdGroupListResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := p.ListAdGroups(campaignIds, bookmark)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second // 递增等待时间
			log.Printf("AdGroup获取失败，%d秒后重试 (第%d/%d次): %v", waitTime/time.Second, attempt, maxRetries, err)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("重试%d次后仍然失败: %w", maxRetries, lastErr)
}

// saveAdGroupsInBatches 分批保存AdGroup数据，避免单次插入过多数据
func (p *Pinterest) saveAdGroupsInBatches(allAdGroups []AdGroup, traceId string) error {
	const batchSize = AdGroupSaveBatchSize // 每批保存的AdGroup数量

	for i := 0; i < len(allAdGroups); i += batchSize {
		end := i + batchSize
		if end > len(allAdGroups) {
			end = len(allAdGroups)
		}

		batch := allAdGroups[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(allAdGroups) + batchSize - 1) / batchSize

		log.Printf("[%s] 保存第%d/%d批AdGroup数据，包含%d条记录", traceId, batchNum, totalBatches, len(batch))

		if err := SaveAdGroupsToAirbyte(batch, p.Account.TenantId); err != nil {
			return fmt.Errorf("保存第%d批数据失败: %w", batchNum, err)
		}
	}

	return nil
}

// CampaignListResponse Campaign列表API响应结构
type CampaignListResponse struct {
	Items    []Campaign `json:"items"`
	Bookmark string     `json:"bookmark,omitempty"`
}

// Campaign Pinterest广告活动结构
type Campaign struct {
	Id                           string `json:"id"`
	AdAccountId                  string `json:"ad_account_id"`
	Name                         string `json:"name"`
	Status                       string `json:"status"`
	ObjectiveType                string `json:"objective_type"`
	LifetimeSpendCap             int64  `json:"lifetime_spend_cap"`
	DailySpendCap                int64  `json:"daily_spend_cap"`
	OrderLineId                  string `json:"order_line_id"` // 可能为null
	TrackingUrls                 string `json:"tracking_urls"` // 可能为null
	CreatedTime                  int64  `json:"created_time"`
	UpdatedTime                  int64  `json:"updated_time"`
	Type                         string `json:"type"`
	IsFlexibleDailyBudgets       bool   `json:"is_flexible_daily_budgets"`
	SummaryStatus                string `json:"summary_status"`
	IsCampaignBudgetOptimization bool   `json:"is_campaign_budget_optimization"`
	StartTime                    *int64 `json:"start_time"`
	EndTime                      string `json:"end_time"` // 可能为null
	IsPerformancePlus            bool   `json:"is_performance_plus"`
	IsAutomatedCampaign          bool   `json:"is_automated_campaign"`
	BidOptions                   string `json:"bid_options"` // 可能为null
}

// AdGroupListResponse AdGroup列表API响应结构
type AdGroupListResponse struct {
	Items    []AdGroup `json:"items"`
	Bookmark string    `json:"bookmark,omitempty"`
}

// AdGroup Pinterest广告组结构
type AdGroup struct {
	Name                     string `json:"name"`
	Status                   string `json:"status"`
	BudgetInMicroCurrency    int64  `json:"budget_in_micro_currency"`
	BidInMicroCurrency       int64  `json:"bid_in_micro_currency"`
	OptimizationGoalMetadata struct {
		ConversionTagV3GoalMetadata struct {
			AttributionWindows struct {
				ClickWindowDays      int64 `json:"click_window_days"`
				EngagementWindowDays int64 `json:"engagement_window_days"`
				ViewWindowDays       int64 `json:"view_window_days"`
			} `json:"attribution_windows"`
			ConversionEvent             string `json:"conversion_event"`
			ConversionTagId             string `json:"conversion_tag_id"`
			CpaGoalValueInMicroCurrency string `json:"cpa_goal_value_in_micro_currency"`
			IsRoasOptimized             bool   `json:"is_roas_optimized"`
			LearningModeType            string `json:"learning_mode_type"`
			ReportingEvent              string `json:"reporting_event"`
		} `json:"conversion_tag_v3_goal_metadata"`
		FrequencyGoalMetadata struct {
			Frequency int64  `json:"frequency"`
			Timerange string `json:"timerange"`
		} `json:"frequency_goal_metadata"`
		ScrollupGoalMetadata struct {
			ScrollupGoalValueInMicroCurrency string `json:"scrollup_goal_value_in_micro_currency"`
		} `json:"scrollup_goal_metadata"`
	} `json:"optimization_goal_metadata"`
	BudgetType    string `json:"budget_type"`
	StartTime     *int64 `json:"start_time"`
	EndTime       *int64 `json:"end_time"`
	TargetingSpec struct {
		AGEBUCKET           []string `json:"AGE_BUCKET"`
		APPTYPE             []string `json:"APPTYPE"`
		AUDIENCEEXCLUDE     []string `json:"AUDIENCE_EXCLUDE"`
		AUDIENCEINCLUDE     []string `json:"AUDIENCE_INCLUDE"`
		GENDER              []string `json:"GENDER"`
		GEO                 []string `json:"GEO"`
		INTEREST            []string `json:"INTEREST"`
		LOCALE              []string `json:"LOCALE"`
		LOCATION            []string `json:"LOCATION"`
		SHOPPINGRETARGETING []struct {
			LookbackWindow  int64   `json:"lookback_window"`
			ExclusionWindow int64   `json:"exclusion_window"`
			TagTypes        []int64 `json:"tag_types"`
		} `json:"SHOPPING_RETARGETING"`
		TARGETINGSTRATEGY []interface{} `json:"TARGETING_STRATEGY"`
	} `json:"targeting_spec"`
	LifetimeFrequencyCap int64 `json:"lifetime_frequency_cap"`
	TrackingUrls         struct {
		Impression           []string `json:"impression"`
		Click                []string `json:"click"`
		Engagement           []string `json:"engagement"`
		BuyableButton        []string `json:"buyable_button"`
		AudienceVerification []string `json:"audience_verification"`
	} `json:"tracking_urls"`
	AutoTargetingEnabled       bool     `json:"auto_targeting_enabled"`
	PlacementGroup             string   `json:"placement_group"`
	PacingDeliveryType         string   `json:"pacing_delivery_type"`
	CampaignId                 string   `json:"campaign_id"`
	BillableEvent              string   `json:"billable_event"`
	BidStrategyType            string   `json:"bid_strategy_type"`
	TargetingTemplateIds       []string `json:"targeting_template_ids"`
	IsCreativeOptimization     bool     `json:"is_creative_optimization"`
	PromotionId                string   `json:"promotion_id"`
	PromotionApplicationLevel  string   `json:"promotion_application_level"`
	Id                         string   `json:"id"`
	AdAccountId                string   `json:"ad_account_id"`
	CreatedTime                int64    `json:"created_time"`
	UpdatedTime                int64    `json:"updated_time"`
	Type                       string   `json:"type"`
	ConversionLearningModeType string   `json:"conversion_learning_mode_type"`
	SummaryStatus              string   `json:"summary_status"`
	FeedProfileId              string   `json:"feed_profile_id"`
	DcaAssets                  string   `json:"dca_assets"`
	BidMultiplier              int64    `json:"bid_multiplier"`
}

// AdListResponse Ad列表API响应结构
type AdListResponse struct {
	Items    []Ad   `json:"items"`
	Bookmark string `json:"bookmark,omitempty"`
}

// Ad Pinterest广告结构
type Ad struct {
	Id                       string   `json:"id"`
	AdAccountId              string   `json:"ad_account_id"`
	AdGroupId                string   `json:"ad_group_id"`
	CampaignId               string   `json:"campaign_id"`
	PinId                    string   `json:"pin_id"`
	Name                     string   `json:"name"`
	Status                   string   `json:"status"`
	Type                     string   `json:"type"`
	CreativeType             string   `json:"creative_type"`
	DestinationUrl           string   `json:"destination_url"`
	AndroidDeepLink          string   `json:"android_deep_link"`
	IosDeepLink              string   `json:"ios_deep_link"`
	CarouselAndroidDeepLinks []string `json:"carousel_android_deep_links"`
	CarouselDestinationUrls  []string `json:"carousel_destination_urls"`
	CarouselIosDeepLinks     []string `json:"carousel_ios_deep_links"`
	ClickTrackingUrl         string   `json:"click_tracking_url"`
	ViewTrackingUrl          string   `json:"view_tracking_url"`
	TrackingUrls             struct {
		Impression           []string `json:"impression"`
		Click                []string `json:"click"`
		Engagement           []string `json:"engagement"`
		BuyableButton        []string `json:"buyable_button"`
		AudienceVerification []string `json:"audience_verification"`
	} `json:"tracking_urls"`
	IsPinDeleted                          bool   `json:"is_pin_deleted"`
	IsRemovable                           bool   `json:"is_removable"`
	LeadFormId                            string `json:"lead_form_id"`
	GridClickType                         string `json:"grid_click_type"`
	CustomizableCtaType                   string `json:"customizable_cta_type"`
	CollectionItemsDestinationUrlTemplate string `json:"collection_items_destination_url_template"`
	QuizPinData                           struct {
		Questions []struct {
			QuestionId   int    `json:"question_id"`
			QuestionText string `json:"question_text"`
			Options      []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"questions"`
		Results []struct {
			OrganicPinId    string `json:"organicPinId"`
			AndroidDeepLink string `json:"android_deep_link"`
			IOSDeepLink     string `json:"iOS_deep_link"`
			DestinationUrl  string `json:"destination_url"`
			ResultId        int    `json:"result_id"`
		} `json:"results"`
	} `json:"quiz_pin_data"`
	RejectedReasons []string `json:"rejected_reasons"`
	RejectionLabels []string `json:"rejection_labels"`
	ReviewStatus    string   `json:"review_status"`
	SummaryStatus   string   `json:"summary_status"`
	CreatedTime     int64    `json:"created_time"`
	UpdatedTime     int64    `json:"updated_time"`
}

// ListAds 获取Pinterest广告列表，支持bookmark分页
// campaignIds: 可选的campaign ID列表，如果为空则获取所有ads
// bookmark: 分页游标，首次请求传空字符串
func (p *Pinterest) ListAds(campaignIds []string, bookmark string) (*AdListResponse, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始获取Ad列表，Campaign IDs: %v, Bookmark: %s", traceId, campaignIds, bookmark)

	// 获取有效的访问令牌
	token, err := p.GetValidAccessToken()
	if err != nil {
		log.Printf("[%s] 获取访问令牌失败: %v", traceId, err)
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建查询参数
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", DefaultPageSize),
	}

	// 如果指定了campaign IDs，添加到查询参数中
	if len(campaignIds) > 0 {
		params["campaign_ids"] = strings.Join(campaignIds, ",")
		log.Printf("[%s] 查询指定的Campaign IDs: %s", traceId, params["campaign_ids"])
	}

	// 如果有bookmark，添加到查询参数中
	if bookmark != "" {
		params["bookmark"] = bookmark
		log.Printf("[%s] 使用bookmark进行分页: %s", traceId, bookmark)
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/ads", PinterestAPIBase, p.Account.AccountId)
	log.Printf("[%s] 调用获取Ad列表API: %s", traceId, apiURL)

	// 发送HTTP请求
	respData, err := p.makeHTTPRequestWithRetry("GET", apiURL, headers, params, nil)
	if err != nil {
		log.Printf("[%s] 获取Ad列表请求失败: %v", traceId, err)
		return nil, fmt.Errorf("获取Ad列表请求失败: %w", err)
	}

	// 解析响应
	var adResp AdListResponse
	if err := json.Unmarshal(respData, &adResp); err != nil {
		log.Printf("[%s] 解析Ad列表响应失败: %v", traceId, err)
		return nil, fmt.Errorf("解析Ad列表响应失败: %w", err)
	}

	log.Printf("[%s] 成功获取Ad列表，共%d个ads，下一页bookmark: %s",
		traceId, len(adResp.Items), adResp.Bookmark)

	return &adResp, nil
}

// ListAllAds 获取指定Campaign的所有Ad数据并保存到数据库
// 支持bookmark分页，自动处理所有分页数据
func (p *Pinterest) ListAllAds(campaignIds []string) error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始获取Ad数据，Campaign数量: %d", traceId, len(campaignIds))

	var allAds []Ad
	var adIds = p.IdForAds
	bookmark := ""
	pageCount := 0
	const maxPages = MaxPagesLimit // 防止无限循环的安全限制

	// 分页获取所有数据
	for pageCount < maxPages {
		pageCount++
		log.Printf("[%s] 获取第%d页Ad数据", traceId, pageCount)

		// 获取当前页数据，带重试机制
		resp, err := p.listAdsWithRetry(campaignIds, bookmark, 3)
		if err != nil {
			log.Printf("[%s] 获取第%d页Ad数据失败: %v", traceId, pageCount, err)
			return fmt.Errorf("获取第%d页Ad数据失败: %w", pageCount, err)
		}

		// 添加到总结果中，同时收集IDs
		if len(resp.Items) > 0 {
			allAds = append(allAds, resp.Items...)
			// 收集Ad IDs
			for _, ad := range resp.Items {
				adIds = append(adIds, ad.Id)
			}
			log.Printf("[%s] 第%d页获取到%d个Ad，累计%d个",
				traceId, pageCount, len(resp.Items), len(allAds))
		} else {
			log.Printf("[%s] 第%d页没有获取到Ad数据", traceId, pageCount)
		}

		// 检查是否还有下一页
		if resp.Bookmark == "" {
			log.Printf("[%s] 没有更多页面，数据获取完成", traceId)
			break
		}

		// 更新bookmark用于下一页
		bookmark = resp.Bookmark
	}

	// 检查是否达到最大页数限制
	if pageCount >= maxPages {
		log.Printf("[%s] 警告：达到最大页数限制(%d)，可能还有未获取的数据", traceId, maxPages)
	}

	// 存储Ad IDs到结构体中
	p.IdForAds = adIds
	log.Printf("[%s] 共获取到%d个Ad，%d个Ad IDs已存储", traceId, len(allAds), len(adIds))

	// 分批保存到数据库，避免单次插入数据过多
	if len(allAds) > 0 {
		if err := p.saveAdsInBatches(allAds, traceId); err != nil {
			return fmt.Errorf("保存Ad数据到数据库失败: %w", err)
		}
		log.Printf("[%s] 成功保存%d个Ad到数据库", traceId, len(allAds))
	} else {
		log.Printf("[%s] 没有Ad数据需要保存", traceId)
	}

	return nil
}

// listAdsWithRetry 带重试机制的Ad获取方法
func (p *Pinterest) listAdsWithRetry(campaignIds []string, bookmark string, maxRetries int) (*AdListResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := p.ListAds(campaignIds, bookmark)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * 2 * time.Second // 递增等待时间
			log.Printf("Ad获取失败，%d秒后重试 (第%d/%d次): %v", waitTime/time.Second, attempt, maxRetries, err)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("重试%d次后仍然失败: %w", maxRetries, lastErr)
}

// saveAdsInBatches 分批保存Ad数据，避免单次插入过多数据
func (p *Pinterest) saveAdsInBatches(allAds []Ad, traceId string) error {
	const batchSize = AdGroupSaveBatchSize // 复用AdGroup的批处理大小

	for i := 0; i < len(allAds); i += batchSize {
		end := i + batchSize
		if end > len(allAds) {
			end = len(allAds)
		}

		batch := allAds[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(allAds) + batchSize - 1) / batchSize

		log.Printf("[%s] 保存第%d/%d批Ad数据，包含%d条记录", traceId, batchNum, totalBatches, len(batch))

		if err := SaveAdsToAirbyte(batch, p.Account.TenantId); err != nil {
			return fmt.Errorf("保存第%d批数据失败: %w", batchNum, err)
		}
	}

	return nil
}

// PullAllAdsAndSave 批量获取所有Ad数据并保存到数据库
// 将campaign IDs分批处理，每批最多50个，避免URL过长和API限制
func (p *Pinterest) PullAllAdsAndSave() error {
	traceId := p.getTraceId()
	campaignIds := p.IdForCampaigns

	if len(campaignIds) == 0 {
		log.Printf("[%s] 没有Campaign IDs需要处理", traceId)
		return nil
	}

	log.Printf("[%s] 开始批量获取Ad数据，共%d个Campaign IDs", traceId, len(campaignIds))

	const batchSize = CampaignBatchSize
	var errors []error

	// 分批处理campaign IDs
	for i := 0; i < len(campaignIds); i += batchSize {
		end := i + batchSize
		if end > len(campaignIds) {
			end = len(campaignIds)
		}

		batch := campaignIds[i:end]
		batchNum := (i / batchSize) + 1
		totalBatches := (len(campaignIds) + batchSize - 1) / batchSize

		log.Printf("[%s] 处理第%d/%d批，包含%d个Campaign IDs", traceId, batchNum, totalBatches, len(batch))

		// 处理当前批次
		if err := p.ListAllAds(batch); err != nil {
			log.Printf("[%s] 第%d批处理失败: %v", traceId, batchNum, err)
			errors = append(errors, fmt.Errorf("批次%d处理失败: %w", batchNum, err))
			continue // 继续处理下一批，不中断整个流程
		}

		log.Printf("[%s] 第%d批处理完成", traceId, batchNum)
	}

	// 汇总结果
	if len(errors) > 0 {
		log.Printf("[%s] 批量处理完成，但有%d个批次失败", traceId, len(errors))
		return fmt.Errorf("批量处理部分失败，%d个错误: %v", len(errors), errors)
	}

	log.Printf("[%s] 批量获取Ad数据完成", traceId)
	return nil
}
