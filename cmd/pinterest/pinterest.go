package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wm-func/wm_account"
)

type Pinterest struct {
	Account          wm_account.Account
	IdForCampaigns   []string
	IdForAdGroups    []string  // 存储已拉取的AdGroup IDs
	IdForAds         []string  // 存储已拉取的Ad IDs
	TokenExpiresAt   time.Time // AccessToken 过期时间
	RefreshExpiresAt time.Time // RefreshToken 过期时间
}

type IdForCampaign struct {
	SPENDINDOLLAR float64 `json:"SPEND_IN_DOLLAR"`
	CAMPAIGNID    int64   `json:"CAMPAIGN_ID"`
	DATE          string  `json:"DATE"`
}

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	ResponseType          string `json:"response_type"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	Scope                 string `json:"scope"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt int64  `json:"refresh_token_expires_at"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
}

// ReportRequest 创建报告的请求结构
type ReportRequest struct {
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Granularity  string   `json:"granularity"`
	Columns      []string `json:"columns"`
	Level        string   `json:"level"`
	ReportFormat string   `json:"report_format"`
}

// ReportResponse 创建报告的响应结构
type ReportResponse struct {
	Token        string `json:"token"`
	Message      string `json:"message"`
	ReportStatus string `json:"report_status"`
	URL          string `json:"url,omitempty"`
	Size         int64  `json:"size,omitempty"`
}

// ReportData 报告数据结构
type ReportData struct {
	SpendInDollar float64 `json:"SPEND_IN_DOLLAR"`
	CampaignID    int64   `json:"CAMPAIGN_ID"`
	Date          string  `json:"DATE"`
}

// NewPinterest 创建新的 Pinterest 实例
func NewPinterest(account wm_account.Account) *Pinterest {
	return &Pinterest{
		Account:          account,
		TokenExpiresAt:   time.Time{}, // 初始为零值，表示需要检查
		RefreshExpiresAt: time.Time{}, // 初始为零值
	}
}

// IsAccessTokenExpired 检查 AccessToken 是否过期
func (p *Pinterest) IsAccessTokenExpired() bool {
	if p.TokenExpiresAt.IsZero() {
		return true // 如果没有过期时间记录，认为已过期
	}
	// 提前5分钟判断为过期，避免临界时间问题
	return time.Now().Add(5 * time.Minute).After(p.TokenExpiresAt)
}

// IsRefreshTokenExpired 检查 RefreshToken 是否过期
func (p *Pinterest) IsRefreshTokenExpired() bool {
	if p.RefreshExpiresAt.IsZero() {
		return false // RefreshToken 通常有很长的有效期，如果没有记录就假设未过期
	}
	return time.Now().After(p.RefreshExpiresAt)
}

// GetAccessToken 获取或刷新 AccessToken
func (p *Pinterest) GetAccessToken() error {
	// 检查当前 AccessToken 是否仍然有效
	if !p.IsAccessTokenExpired() {
		return nil // Token 未过期，无需刷新
	}

	// 检查 RefreshToken 是否过期
	if p.IsRefreshTokenExpired() {
		return fmt.Errorf("RefreshToken 已过期，需要重新授权")
	}

	// 刷新 AccessToken
	return p.refreshAccessToken()
}

// refreshAccessToken 刷新 AccessToken 的内部实现
func (p *Pinterest) refreshAccessToken() error {
	const apiURL = "https://api.pinterest.com/v5/oauth/token"
	const clientCredentials = "MTQ5NDYwOTpiZjk2M2I1MGE3NjRlZjUwNzhiMDY4N2ZlOWQxNWU5MDFkNzc1OWNl"

	// 准备表单数据
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {p.Account.RefreshToken},
		"refresh_on":    {"true"},
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": "Basic " + clientCredentials,
		"Content-Type":  "application/x-www-form-urlencoded",
		"User-Agent":    "pinterest-api-client/1.0",
		"Cookie":        "_ir=0",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second, // 设置超时时间
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 响应错误，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 更新账户信息和过期时间
	p.updateTokenInfo(tokenResp)

	return nil
}

// updateTokenInfo 更新 token 信息和过期时间
func (p *Pinterest) updateTokenInfo(tokenResp TokenResponse) {
	now := time.Now()

	// 更新 AccessToken 和过期时间
	p.Account.AccessToken = tokenResp.AccessToken
	p.TokenExpiresAt = now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// 更新 RefreshToken 和过期时间
	if tokenResp.RefreshToken != "" {
		p.Account.RefreshToken = tokenResp.RefreshToken
		if tokenResp.RefreshTokenExpiresAt > 0 {
			p.RefreshExpiresAt = time.Unix(tokenResp.RefreshTokenExpiresAt, 0)
		}
	}
}

// GetValidAccessToken 获取有效的 AccessToken，如果过期会自动刷新
func (p *Pinterest) GetValidAccessToken() (string, error) {
	if err := p.GetAccessToken(); err != nil {
		return "", err
	}
	return p.Account.AccessToken, nil
}

// GetTokenInfo 获取 token 信息，用于调试和监控
func (p *Pinterest) GetTokenInfo() map[string]interface{} {
	return map[string]interface{}{
		"access_token_expired":  p.IsAccessTokenExpired(),
		"refresh_token_expired": p.IsRefreshTokenExpired(),
		"token_expires_at":      p.TokenExpiresAt,
		"refresh_expires_at":    p.RefreshExpiresAt,
		"has_access_token":      p.Account.AccessToken != "",
		"has_refresh_token":     p.Account.RefreshToken != "",
	}
}
