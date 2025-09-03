package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// AdGroupAnalyticsRequest AdGroup Analytics API请求参数
type AdGroupAnalyticsRequest struct {
	StartDate            string   `json:"start_date"`
	EndDate              string   `json:"end_date"`
	AdGroupIds           []string `json:"ad_group_ids"`
	Columns              []string `json:"columns"`
	Granularity          string   `json:"granularity"`
	ClickWindowDays      int      `json:"click_window_days"`
	EngagementWindowDays int      `json:"engagement_window_days"`
	ViewWindowDays       int      `json:"view_window_days"`
	PageSize             int      `json:"page_size"`
}

// AdGroupMetrics AdGroup Analytics数据结构
type AdGroupMetrics struct {
	TenantId                                int64   `json:"tenant_id"` // 这个在转化的时候写入
	ADGROUPID                               string  `json:"AD_GROUP_ID"`
	ADGROUPENTITYSTATUS                     int64   `json:"AD_GROUP_ENTITY_STATUS"`
	CPMINMICRODOLLAR                        float64 `json:"CPM_IN_MICRO_DOLLAR"`
	CAMPAIGNENTITYSTATUS                    int64   `json:"CAMPAIGN_ENTITY_STATUS"`
	CPMINDOLLAR                             float64 `json:"CPM_IN_DOLLAR"`
	PAIDIMPRESSION                          int64   `json:"PAID_IMPRESSION"`
	VIDEOLENGTH                             int64   `json:"VIDEO_LENGTH"`
	TOTALVIDEO3SECVIEWS                     int64   `json:"TOTAL_VIDEO_3SEC_VIEWS"`
	SPENDINMICRODOLLAR                      int64   `json:"SPEND_IN_MICRO_DOLLAR"`
	ECPVINDOLLAR                            float64 `json:"ECPV_IN_DOLLAR"`
	ECPMINMICRODOLLAR                       float64 `json:"ECPM_IN_MICRO_DOLLAR"`
	TOTALIMPRESSIONFREQUENCY                float64 `json:"TOTAL_IMPRESSION_FREQUENCY"`
	TOTALIMPRESSIONUSER                     int64   `json:"TOTAL_IMPRESSION_USER"`
	IMPRESSION1                             int64   `json:"IMPRESSION_1"`
	TOTALVIDEOAVGWATCHTIMEINSECOND          float64 `json:"TOTAL_VIDEO_AVG_WATCHTIME_IN_SECOND"`
	ADVERTISERID                            int64   `json:"ADVERTISER_ID"`
	PINPROMOTIONID                          int64   `json:"PIN_PROMOTION_ID"`
	CAMPAIGNID                              int64   `json:"CAMPAIGN_ID"`
	IMPRESSION1GROSS                        int64   `json:"IMPRESSION_1_GROSS"`
	TOTALVIDEOP75COMBINED                   int64   `json:"TOTAL_VIDEO_P75_COMBINED"`
	TOTALVIDEOP95COMBINED                   int64   `json:"TOTAL_VIDEO_P95_COMBINED"`
	PINID                                   int64   `json:"PIN_ID"`
	TOTALVIDEOP25COMBINED                   int64   `json:"TOTAL_VIDEO_P25_COMBINED"`
	TOTALVIDEOP50COMBINED                   int64   `json:"TOTAL_VIDEO_P50_COMBINED"`
	ECPCVP95INDOLLAR                        float64 `json:"ECPCV_P95_IN_DOLLAR"`
	TOTALVIDEOP0COMBINED                    int64   `json:"TOTAL_VIDEO_P0_COMBINED"`
	TOTALVIDEOMRCVIEWS                      int64   `json:"TOTAL_VIDEO_MRC_VIEWS"`
	CAMPAIGNLIFETIMESPENDCAP                int64   `json:"CAMPAIGN_LIFETIME_SPEND_CAP"`
	SPENDINDOLLAR                           float64 `json:"SPEND_IN_DOLLAR"`
	CAMPAIGNDAILYSPENDCAP                   int64   `json:"CAMPAIGN_DAILY_SPEND_CAP"`
	ADID                                    string  `json:"AD_ID"`
	DATE                                    string  `json:"DATE"`
	ADACCOUNTID                             int64   `json:"AD_ACCOUNT_ID"`
	CAMPAIGNNAME                            string  `json:"CAMPAIGN_NAME"`
	TOTALCLICKADDTOCART                     int64   `json:"TOTAL_CLICK_ADD_TO_CART,omitempty"`
	TOTALWEBCLICKCHECKOUT                   int64   `json:"TOTAL_WEB_CLICK_CHECKOUT,omitempty"`
	CLICKTHROUGH1                           int64   `json:"CLICKTHROUGH_1,omitempty"`
	TOTALWEBCLICKCHECKOUTVALUEINMICRODOLLAR int64   `json:"TOTAL_WEB_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	REPIN1                                  int64   `json:"REPIN_1,omitempty"`
	CLICKTHROUGH1GROSS                      int64   `json:"CLICKTHROUGH_1_GROSS,omitempty"`
	OUTBOUNDCLICK1                          int64   `json:"OUTBOUND_CLICK_1,omitempty"`
	TOTALVIEWADDTOCART                      int64   `json:"TOTAL_VIEW_ADD_TO_CART,omitempty"`
	PAGEVISITCOSTPERACTION                  float64 `json:"PAGE_VISIT_COST_PER_ACTION,omitempty"`
	TOTALCLICKTHROUGH                       int64   `json:"TOTAL_CLICKTHROUGH,omitempty"`
	TOTALENGAGEMENT                         int64   `json:"TOTAL_ENGAGEMENT,omitempty"`
	ENGAGEMENT1                             int64   `json:"ENGAGEMENT_1,omitempty"`
	ECTR                                    float64 `json:"ECTR,omitempty"`
	ECPCINDOLLAR                            float64 `json:"ECPC_IN_DOLLAR,omitempty"`
	CTR                                     float64 `json:"CTR,omitempty"`
	ECPCINMICRODOLLAR                       float64 `json:"ECPC_IN_MICRO_DOLLAR,omitempty"`
	ECPCVINDOLLAR                           float64 `json:"ECPCV_IN_DOLLAR,omitempty"`
	CPCINMICRODOLLAR                        float64 `json:"CPC_IN_MICRO_DOLLAR,omitempty"`
	VIDEOMRCVIEWS2                          int64   `json:"VIDEO_MRC_VIEWS_2,omitempty"`
	TOTALCLICKCHECKOUT                      int64   `json:"TOTAL_CLICK_CHECKOUT,omitempty"`
	TOTALCLICKSIGNUP                        int64   `json:"TOTAL_CLICK_SIGNUP,omitempty"`
	TOTALCHECKOUTVALUEINMICRODOLLAR         int64   `json:"TOTAL_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	VIDEO3SECVIEWS2                         int64   `json:"VIDEO_3SEC_VIEWS_2,omitempty"`
	EENGAGEMENTRATE                         float64 `json:"EENGAGEMENT_RATE,omitempty"`
	ECPEINDOLLAR                            float64 `json:"ECPE_IN_DOLLAR,omitempty"`
	ENGAGEMENTRATE                          float64 `json:"ENGAGEMENT_RATE,omitempty"`
	TOTALCHECKOUT                           int64   `json:"TOTAL_CHECKOUT,omitempty"`
	TOTALPAGEVISIT                          int64   `json:"TOTAL_PAGE_VISIT,omitempty"`
	TOTALSIGNUP                             int64   `json:"TOTAL_SIGNUP,omitempty"`
	IMPRESSION2                             int64   `json:"IMPRESSION_2,omitempty"`
	REPINRATE                               float64 `json:"REPIN_RATE,omitempty"`
	TOTALWEBCHECKOUT                        int64   `json:"TOTAL_WEB_CHECKOUT,omitempty"`
	TOTALWEBCHECKOUTVALUEINMICRODOLLAR      int64   `json:"TOTAL_WEB_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	WEBCHECKOUTCOSTPERACTION                float64 `json:"WEB_CHECKOUT_COST_PER_ACTION,omitempty"`
	TOTALREPINRATE                          float64 `json:"TOTAL_REPIN_RATE,omitempty"`
	VIDEOP0COMBINED2                        int64   `json:"VIDEO_P0_COMBINED_2,omitempty"`
	WEBCHECKOUTROAS                         float64 `json:"WEB_CHECKOUT_ROAS,omitempty"`
	TOTALCONVERSIONS                        int64   `json:"TOTAL_CONVERSIONS,omitempty"`
	CHECKOUTROAS                            float64 `json:"CHECKOUT_ROAS,omitempty"`
	TOTALVIDEOP100COMPLETE                  int64   `json:"TOTAL_VIDEO_P100_COMPLETE,omitempty"`
	TOTALVIEWSIGNUP                         int64   `json:"TOTAL_VIEW_SIGNUP,omitempty"`
	TOTALCLICKCHECKOUTVALUEINMICRODOLLAR    int64   `json:"TOTAL_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	ENGAGEMENT2                             int64   `json:"ENGAGEMENT_2,omitempty"`
	CTR2                                    float64 `json:"CTR_2,omitempty"`
	CLICKTHROUGH2                           int64   `json:"CLICKTHROUGH_2,omitempty"`
	VIDEOP25COMBINED2                       int64   `json:"VIDEO_P25_COMBINED_2,omitempty"`
	VIDEOP50COMBINED2                       int64   `json:"VIDEO_P50_COMBINED_2,omitempty"`
	TOTALWEBVIEWCHECKOUT                    int64   `json:"TOTAL_WEB_VIEW_CHECKOUT,omitempty"`
	TOTALWEBVIEWCHECKOUTVALUEINMICRODOLLAR  int64   `json:"TOTAL_WEB_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	TOTALVIEWCHECKOUT                       int64   `json:"TOTAL_VIEW_CHECKOUT,omitempty"`
	TOTALVIEWCHECKOUTVALUEINMICRODOLLAR     int64   `json:"TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR,omitempty"`
	OUTBOUNDCLICK2                          int64   `json:"OUTBOUND_CLICK_2,omitempty"`
	VIDEOP75COMBINED2                       int64   `json:"VIDEO_P75_COMBINED_2,omitempty"`
	VIDEOP100COMPLETE2                      int64   `json:"VIDEO_P100_COMPLETE_2,omitempty"`
	VIDEOP95COMBINED2                       int64   `json:"VIDEO_P95_COMBINED_2,omitempty"`
}

// GetAllAdGroupIds 从结构体中直接获取已拉取的AdGroup IDs
func (p *Pinterest) GetAllAdGroupIds() ([]string, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 从结构体中获取AdGroup IDs", traceId)

	if len(p.IdForAdGroups) == 0 {
		log.Printf("[%s] 结构体中没有AdGroup IDs，可能还没有拉取AdGroup数据", traceId)
		return nil, fmt.Errorf("没有可用的AdGroup IDs，请先拉取AdGroup数据")
	}

	log.Printf("[%s] 从结构体中获取到%d个AdGroup IDs", traceId, len(p.IdForAdGroups))
	return p.IdForAdGroups, nil
}

// PullAdGroupAnalytics 拉取AdGroup Analytics数据的主函数
func (p *Pinterest) PullAdGroupAnalytics() error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始拉取AdGroup Analytics数据", traceId)

	// 1. 获取所有AdGroup IDs
	adGroupIds, err := p.GetAllAdGroupIds()
	if err != nil {
		log.Printf("[%s] 获取AdGroup IDs失败: %v", traceId, err)
		return fmt.Errorf("获取AdGroup IDs失败: %w", err)
	}

	if len(adGroupIds) == 0 {
		log.Printf("[%s] 没有AdGroup IDs需要处理", traceId)
		return nil
	}

	log.Printf("[%s] 开始处理%d个AdGroup IDs的Analytics数据", traceId, len(adGroupIds))

	// 2. 分组处理AdGroup IDs（每组4个）
	const groupSize = 10
	const maxWorkers = 6

	groups := make([][]string, 0)
	for i := 0; i < len(adGroupIds); i += groupSize {
		end := i + groupSize
		if end > len(adGroupIds) {
			end = len(adGroupIds)
		}
		groups = append(groups, adGroupIds[i:end])
	}

	log.Printf("[%s] 将%d个AdGroup IDs分成%d组，每组最多%d个", traceId, len(adGroupIds), len(groups), groupSize)

	// 3. 使用工作池并发处理
	return p.processAdGroupAnalyticsWithWorkerPool(groups, maxWorkers, traceId)
}

// processAdGroupAnalyticsWithWorkerPool 使用工作池并发处理AdGroup Analytics
func (p *Pinterest) processAdGroupAnalyticsWithWorkerPool(groups [][]string, maxWorkers int, traceId string) error {
	// 创建工作通道和结果通道
	jobChan := make(chan []string, len(groups))
	resultChan := make(chan error, len(groups))

	maxWorkers = 1
	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for adGroupIds := range jobChan {
				log.Printf("[%s] Worker %d 开始处理%d个AdGroup IDs: %v", traceId, workerID, len(adGroupIds), adGroupIds)
				err := p.fetchAndSaveAdGroupAnalytics(adGroupIds, workerID, traceId)
				resultChan <- err
				if err != nil {
					log.Printf("[%s] Worker %d 处理失败: %v", traceId, workerID, err)
				} else {
					log.Printf("[%s] Worker %d 处理完成", traceId, workerID)
				}
			}
		}(i)
	}

	// 发送任务到工作通道
	for _, group := range groups {
		jobChan <- group
	}
	close(jobChan)

	// 等待所有工作协程完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var errors []error
	for err := range resultChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		log.Printf("[%s] AdGroup Analytics处理完成，但有%d个错误", traceId, len(errors))
		return fmt.Errorf("部分AdGroup Analytics处理失败，%d个错误: %v", len(errors), errors)
	}

	log.Printf("[%s] 所有AdGroup Analytics数据处理完成", traceId)
	return nil
}

// fetchAndSaveAdGroupAnalytics 获取并保存指定AdGroup IDs的Analytics数据
func (p *Pinterest) fetchAndSaveAdGroupAnalytics(adGroupIds []string, workerID int, traceId string) error {
	log.Printf("[%s] Worker %d 开始获取AdGroup Analytics数据，AdGroup IDs: %v", traceId, workerID, adGroupIds)

	// 获取最近30天的数据
	now := time.Now().UTC()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -30).Format("2006-01-02")

	// 构建所有需要的列
	columns := []string{
		"ADVERTISER_ID", "AD_ACCOUNT_ID", "AD_GROUP_ENTITY_STATUS", "AD_GROUP_ID", "AD_ID",
		"CAMPAIGN_DAILY_SPEND_CAP", "CAMPAIGN_ENTITY_STATUS", "CAMPAIGN_ID", "CAMPAIGN_LIFETIME_SPEND_CAP",
		"CAMPAIGN_NAME", "CHECKOUT_ROAS", "CLICKTHROUGH_1", "CLICKTHROUGH_1_GROSS", "CLICKTHROUGH_2",
		"CPC_IN_MICRO_DOLLAR", "CPM_IN_DOLLAR", "CPM_IN_MICRO_DOLLAR", "CTR", "CTR_2",
		"ECPCV_IN_DOLLAR", "ECPCV_P95_IN_DOLLAR", "ECPC_IN_DOLLAR", "ECPC_IN_MICRO_DOLLAR",
		"ECPE_IN_DOLLAR", "ECPM_IN_MICRO_DOLLAR", "ECPV_IN_DOLLAR", "ECTR", "EENGAGEMENT_RATE",
		"ENGAGEMENT_1", "ENGAGEMENT_2", "ENGAGEMENT_RATE", "IDEA_PIN_PRODUCT_TAG_VISIT_1",
		"IDEA_PIN_PRODUCT_TAG_VISIT_2", "IMPRESSION_1", "IMPRESSION_1_GROSS", "IMPRESSION_2",
		"INAPP_CHECKOUT_COST_PER_ACTION", "OUTBOUND_CLICK_1", "OUTBOUND_CLICK_2",
		"PAGE_VISIT_COST_PER_ACTION", "PAGE_VISIT_ROAS", "PAID_IMPRESSION", "PIN_ID", "PIN_PROMOTION_ID",
		"REPIN_1", "REPIN_2", "REPIN_RATE", "SPEND_IN_DOLLAR", "SPEND_IN_MICRO_DOLLAR",
		"TOTAL_CHECKOUT", "TOTAL_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_CLICKTHROUGH",
		"TOTAL_CLICK_ADD_TO_CART", "TOTAL_CLICK_CHECKOUT", "TOTAL_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR",
		"TOTAL_CLICK_LEAD", "TOTAL_CLICK_SIGNUP", "TOTAL_CLICK_SIGNUP_VALUE_IN_MICRO_DOLLAR",
		"TOTAL_CONVERSIONS", "TOTAL_CUSTOM", "TOTAL_ENGAGEMENT", "TOTAL_ENGAGEMENT_CHECKOUT",
		"TOTAL_ENGAGEMENT_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_ENGAGEMENT_LEAD",
		"TOTAL_ENGAGEMENT_SIGNUP", "TOTAL_ENGAGEMENT_SIGNUP_VALUE_IN_MICRO_DOLLAR",
		"TOTAL_IDEA_PIN_PRODUCT_TAG_VISIT", "TOTAL_IMPRESSION_FREQUENCY", "TOTAL_IMPRESSION_USER",
		"TOTAL_LEAD", "TOTAL_OFFLINE_CHECKOUT", "TOTAL_PAGE_VISIT", "TOTAL_REPIN_RATE",
		"TOTAL_SIGNUP", "TOTAL_SIGNUP_VALUE_IN_MICRO_DOLLAR", "TOTAL_VIDEO_3SEC_VIEWS",
		"TOTAL_VIDEO_AVG_WATCHTIME_IN_SECOND", "TOTAL_VIDEO_MRC_VIEWS", "TOTAL_VIDEO_P0_COMBINED",
		"TOTAL_VIDEO_P100_COMPLETE", "TOTAL_VIDEO_P25_COMBINED", "TOTAL_VIDEO_P50_COMBINED",
		"TOTAL_VIDEO_P75_COMBINED", "TOTAL_VIDEO_P95_COMBINED", "TOTAL_VIEW_ADD_TO_CART",
		"TOTAL_VIEW_CHECKOUT", "TOTAL_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_VIEW_LEAD",
		"TOTAL_VIEW_SIGNUP", "TOTAL_VIEW_SIGNUP_VALUE_IN_MICRO_DOLLAR", "TOTAL_WEB_CHECKOUT",
		"TOTAL_WEB_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_WEB_CLICK_CHECKOUT",
		"TOTAL_WEB_CLICK_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_WEB_ENGAGEMENT_CHECKOUT",
		"TOTAL_WEB_ENGAGEMENT_CHECKOUT_VALUE_IN_MICRO_DOLLAR", "TOTAL_WEB_SESSIONS",
		"TOTAL_WEB_VIEW_CHECKOUT", "TOTAL_WEB_VIEW_CHECKOUT_VALUE_IN_MICRO_DOLLAR",
		"VIDEO_3SEC_VIEWS_2", "VIDEO_LENGTH", "VIDEO_MRC_VIEWS_2", "VIDEO_P0_COMBINED_2",
		"VIDEO_P100_COMPLETE_2", "VIDEO_P25_COMBINED_2", "VIDEO_P50_COMBINED_2",
		"VIDEO_P75_COMBINED_2", "VIDEO_P95_COMBINED_2", "WEB_CHECKOUT_COST_PER_ACTION",
		"WEB_CHECKOUT_ROAS", "WEB_SESSIONS_1", "WEB_SESSIONS_2",
	}

	// 获取有效的访问令牌
	token, err := p.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建请求头
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	// 构建查询参数
	params := map[string]string{
		"start_date":             startDate,
		"end_date":               endDate,
		"ad_group_ids":           strings.Join(adGroupIds, ","),
		"columns":                strings.Join(columns, ","),
		"granularity":            "DAY",
		"click_window_days":      "30",
		"engagement_window_days": "30",
		"view_window_days":       "30",
		"page_size":              "250",
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/ad_accounts/%s/ad_groups/analytics", PinterestAPIBase, p.Account.AccountId)

	log.Printf("[%s] Worker %d 调用AdGroup Analytics API: %s", traceId, workerID, apiURL)

	// 发送HTTP请求
	respData, err := p.makeHTTPRequestWithRetry("GET", apiURL, headers, params, nil)
	if err != nil {
		return fmt.Errorf("获取AdGroup Analytics数据失败: %w", err)
	}

	// 解析响应
	var analyticsResp []AdGroupMetrics
	if err := json.Unmarshal(respData, &analyticsResp); err != nil {
		return fmt.Errorf("解析AdGroup Analytics响应失败: %w", err)
	}

	log.Printf("[%s] Worker %d 获取到%d条AdGroup Analytics数据", traceId, workerID, len(analyticsResp))

	// 保存到数据库
	if len(analyticsResp) > 0 {
		if err := SaveAdGroupAnalyticsToAirbyte(analyticsResp, p.Account.TenantId); err != nil {
			return fmt.Errorf("保存AdGroup Analytics数据失败: %w", err)
		}
		log.Printf("[%s] Worker %d 成功保存%d条AdGroup Analytics数据", traceId, workerID, len(analyticsResp))
	}

	return nil
}
