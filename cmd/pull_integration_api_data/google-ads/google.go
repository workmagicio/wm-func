package google_ads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wm-func/common/db/platform_db"
)

const (
	GoogleAds      = "googleAds"
	DeveloperToken = "xhBTEWbeFONUH5kvHWcl2A"
	ID             = "TODO"
	KEY            = "TODO"
	DateFormat     = "2006-01-02"
)

// Connection 连接配置
type Connection struct {
	TenantID int64     `json:"tenant_id"`
	Accounts []Account `json:"accounts"`
	Tokens   Tokens    `json:"tokens"`
}

// Account 账户信息
type Account struct {
	ID     string `json:"id"`
	Cipher string `json:"cipher"`
}

// Tokens 认证令牌
type Tokens struct {
	RefreshToken string `json:"refreshToken"`
}

// GoogleAdsResponse Google Ads API响应
type GoogleAdsResponse struct {
	Results []GoogleAdsResult `json:"results"`
}

// GoogleAdsResult Google Ads 结果
type GoogleAdsResult struct {
	Segments struct {
		Date                     string `json:"date"`
		ConversionActionCategory string `json:"conversionActionCategory,omitempty"`
	} `json:"segments"`
	Customer struct {
		ID string `json:"id"`
	} `json:"customer"`
	Metrics struct {
		CostMicros              string  `json:"costMicros"`
		Orders                  float64 `json:"orders"`
		AverageOrderValueMicros string  `json:"averageOrderValueMicros"`
		Conversions             float64 `json:"conversions,omitempty"`
		ConversionsValue        float64 `json:"conversionsValue,omitempty"`
	} `json:"metrics"`
}

// IntegrationAPIData 集成API数据
type IntegrationAPIData struct {
	TenantID  int64       `json:"tenant_id"`
	AccountID string      `json:"account_id"`
	Date      string      `json:"date"`
	Data      interface{} `json:"data"`
	Platform  string      `json:"platform"`
}

// TokenResponse OAuth2令牌响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GoogleAdsPullData 拉取Google Ads数据
func GoogleAdsPullData(connections []Connection, startTime, endTime time.Time, env string, testTenantIDs []int64) error {
	for _, conn := range connections {
		// 测试环境过滤
		if env == "TEST" && !contains(testTenantIDs, conn.TenantID) {
			continue
		}

		var result []IntegrationAPIData
		for _, account := range conn.Accounts {
			// 一次性查询整个时间范围的数据
			startFormat := startTime.Format(DateFormat)
			endFormat := endTime.Format(DateFormat)

			log.Printf("%d %s 开始查询时间范围: %s 到 %s", conn.TenantID, GoogleAds, startFormat, endFormat)

			data, err := getGoogleAccountInsightsRange(account.ID, conn.Tokens.RefreshToken, startFormat, endFormat, conn.TenantID, account.Cipher)
			if err != nil {
				log.Printf("获取Google Ads数据失败: %v", err)
				return err
			}

			if data != nil && len(data.Results) > 0 {
				for _, dd := range data.Results {
					// 使用API返回的实际日期
					dateStr := dd.Segments.Date
					if dateStr == "" {
						dateStr = startFormat // 如果没有日期，使用开始日期作为默认值
					}

					result = append(result, IntegrationAPIData{
						TenantID:  conn.TenantID,
						AccountID: account.ID,
						Date:      dateStr,
						Data:      dd,
						Platform:  GoogleAds,
					})
				}
				log.Printf("%d %s 获取报告成功，记录数: %d", conn.TenantID, GoogleAds, len(data.Results))
			} else {
				log.Printf("%d %s 时间范围内无数据", conn.TenantID, GoogleAds)
			}
		}

		if err := insert(result); err != nil {
			log.Printf("插入数据失败: %v", err)
			return err
		}
		log.Printf("%d %s 插入成功，记录数(%d)......", conn.TenantID, GoogleAds, len(result))
	}
	return nil
}

// getGoogleAccountInsightsRange 获取Google账户指定时间范围的洞察数据
func getGoogleAccountInsightsRange(accountID, refreshToken, startDate, endDate string, _ int64, mccID string) (*GoogleAdsResponse, error) {
	accessToken, err := refreshTokenFunc(refreshToken)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://googleads.googleapis.com/v19/customers/%s/googleAds:search", accountID)

	headers := map[string]string{
		"Authorization":   "Bearer " + accessToken,
		"Content-Type":    "application/json",
		"developer-token": DeveloperToken,
	}
	if mccID != "" {
		headers["login-customer-id"] = mccID
	}

	// 基础查询 - 查询指定时间范围内的数据
	baseQuery := fmt.Sprintf("SELECT segments.date, customer.id, metrics.cost_micros, metrics.orders, metrics.average_order_value_micros FROM customer WHERE segments.date >= '%s' AND segments.date <= '%s'", startDate, endDate)
	baseResp, err := makeGoogleAdsRequest(url, headers, baseQuery)
	if err != nil {
		return nil, err
	}

	// 转换查询 - 查询指定时间范围内的转换数据
	conversionQuery := fmt.Sprintf("SELECT segments.date, customer.id, segments.conversion_action_category, metrics.conversions, metrics.conversions_value FROM customer WHERE segments.date >= '%s' AND segments.date <= '%s' AND segments.conversion_action_category = 'PURCHASE'", startDate, endDate)
	conversionResp, err := makeGoogleAdsRequest(url, headers, conversionQuery)
	if err != nil {
		return nil, err
	}

	// 合并数据 - 按日期匹配合并转换数据
	if len(baseResp.Results) > 0 && len(conversionResp.Results) > 0 {
		// 创建转换数据的日期映射
		conversionMap := make(map[string]*GoogleAdsResult)
		for i := range conversionResp.Results {
			conversionMap[conversionResp.Results[i].Segments.Date] = &conversionResp.Results[i]
		}

		// 将转换数据合并到基础数据中
		for i := range baseResp.Results {
			if convResult, exists := conversionMap[baseResp.Results[i].Segments.Date]; exists {
				if convResult.Metrics.Conversions > 0 {
					baseResp.Results[i].Metrics.Conversions = convResult.Metrics.Conversions
					baseResp.Results[i].Metrics.ConversionsValue = convResult.Metrics.ConversionsValue
				}
			}
		}
	}

	return baseResp, nil
}

// makeGoogleAdsRequest 发送Google Ads API请求
func makeGoogleAdsRequest(url string, headers map[string]string, query string) (*GoogleAdsResponse, error) {
	body := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google-request-error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result GoogleAdsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// refreshTokenFunc 刷新访问令牌
func refreshTokenFunc(refreshToken string) (string, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", ID)
	data.Set("client_secret", KEY)
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token refresh failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

// insert 插入数据到数据库
func insert(data []IntegrationAPIData) error {
	if len(data) == 0 {
		return nil
	}

	db := platform_db.GetDB()

	// 生成主键ID列表用于删除
	var pkIDs []string
	for _, da := range data {
		pkID := fmt.Sprintf("%s|%s|%s", da.Platform, da.AccountID, strings.ReplaceAll(da.Date, "-", ""))
		pkIDs = append(pkIDs, "'"+pkID+"'")
	}

	// 删除现有数据
	deleteSql := fmt.Sprintf("DELETE FROM platform_offline.integration_api_data WHERE _pk_id IN (%s)", strings.Join(pkIDs, ","))
	if err := db.Exec(deleteSql).Error; err != nil {
		log.Printf("删除数据失败: %v", err)
		return err
	}

	// 构建插入语句
	var values []string
	for _, da := range data {
		pkID := fmt.Sprintf("%s|%s|%s", da.Platform, da.AccountID, strings.ReplaceAll(da.Date, "-", ""))
		dataJSON, err := json.Marshal(da.Data)
		if err != nil {
			log.Printf("序列化数据失败: %v", err)
			return err
		}
		value := fmt.Sprintf("('%s', %d, '%s', '%s', '%s', '%s')",
			pkID, da.TenantID, da.AccountID, da.Date, da.Platform, string(dataJSON))
		values = append(values, value)
	}

	insertSQL := fmt.Sprintf("INSERT INTO platform_offline.integration_api_data(_pk_id, tenant_id, account_id, raw_date, raw_platform, raw_data) VALUES %s",
		strings.Join(values, ","))

	if err := db.Exec(insertSQL).Error; err != nil {
		log.Printf("插入SQL: %s", insertSQL)
		log.Printf("删除SQL: %s", deleteSql)
		log.Printf("插入数据失败: %v", err)
		return err
	}

	return nil
}

// contains 检查切片是否包含元素
func contains(slice []int64, item int64) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
