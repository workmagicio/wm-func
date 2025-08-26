package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func RefreshToken(account KAccount) (*RefreshTokenResponse, error) {
	traceId := account.GetTraceId()
	log.Printf("[%s] 开始RefreshToken，请求授权token", traceId)

	// 1. 定义 API 的基础 URL
	apiURL := "https://app-api.knocommerce.com/api/oauth2/token"

	// 2. 准备 URL 查询参数
	params := url.Values{}
	params.Add("grant_type", "client_credentials")
	params.Add("scope", "attribution responses surveys")

	// 3. 构造完整的请求 URL
	fullURL := apiURL + "?" + params.Encode()
	log.Printf("[%s] 构建请求URL: %s", traceId, fullURL)

	// 4. 创建一个新的 POST 请求
	// 第三个参数是请求体，这里我们没有请求体，所以是 nil
	req, err := http.NewRequest("POST", fullURL, nil)
	if err != nil {
		log.Printf("[%s] 创建HTTP请求失败: %v", traceId, err)
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 5. 对 AccessKeyId 和 SecretToken 进行 Base64 编码以用于 Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(account.AccessKeyId + ":" + account.SecretToken))

	// 6. 设置 Authorization 请求头
	req.Header.Add("Authorization", "Basic "+auth)
	log.Printf("[%s] 设置Authorization请求头完成", traceId)

	// 7. 发送 HTTP 请求
	log.Printf("[%s] 开始发送HTTP请求到Knocommerce API", traceId)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] 发送HTTP请求失败: %v", traceId, err)
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	// 确保在函数结束时关闭响应体
	defer resp.Body.Close()

	log.Printf("[%s] 收到HTTP响应，状态码: %d", traceId, resp.StatusCode)

	// 8. 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		// 如果状态码不是 200 OK，读取响应体并返回错误
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[%s] API返回错误状态码: %d, 响应内容: %s", traceId, resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 9. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] 读取响应体失败: %v", traceId, err)
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	log.Printf("[%s] 开始解析JSON响应", traceId)

	// 10. 将 JSON 格式的响应体解析到 RefreshToken 结构体中
	var token RefreshTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		log.Printf("[%s] JSON解析失败: %v", traceId, err)
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	log.Printf("[%s] RefreshToken成功，获得access_token", traceId)
	// 11. 返回解析后的 token
	return &token, nil
}

// TokenManager 管理token的生命周期，支持过期自动续期
type TokenManager struct {
	mutex      sync.RWMutex          // 读写锁，保证并发安全
	account    KAccount              // 账户信息
	token      *RefreshTokenResponse // 当前token
	obtainedAt time.Time             // token获取时间
}

// NewTokenManager 创建新的token管理器
func NewTokenManager(account KAccount) *TokenManager {
	return &TokenManager{
		account: account,
	}
}

// GetValidToken 获取有效的token，如果过期则自动刷新
func (tm *TokenManager) GetValidToken() (*RefreshTokenResponse, error) {
	traceId := tm.account.GetTraceId()

	// 先用读锁检查token是否有效
	tm.mutex.RLock()
	if tm.token != nil && tm.isTokenValid() {
		token := tm.token
		tm.mutex.RUnlock()
		//log.Printf("[%s] 使用缓存的有效token", traceId)
		return token, nil
	}
	tm.mutex.RUnlock()

	// 需要刷新token，使用写锁
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 双重检查，防止并发情况下重复刷新
	if tm.token != nil && tm.isTokenValid() {
		log.Printf("[%s] 其他协程已刷新token，使用最新token", traceId)
		return tm.token, nil
	}

	log.Printf("[%s] token已过期或不存在，开始刷新token", traceId)

	// 刷新token
	newToken, err := RefreshToken(tm.account)
	if err != nil {
		return nil, fmt.Errorf("刷新token失败: %w", err)
	}

	// 更新token和获取时间
	tm.token = newToken
	tm.obtainedAt = time.Now()

	log.Printf("[%s] token刷新成功，有效期: %d秒", traceId, newToken.ExpiresIn)
	return newToken, nil
}

// isTokenValid 检查token是否仍然有效（内部方法，调用时需要持有锁）
func (tm *TokenManager) isTokenValid() bool {
	if tm.token == nil {
		return false
	}

	// 提前30秒刷新token，避免在使用过程中过期
	expirationTime := tm.obtainedAt.Add(time.Duration(tm.token.ExpiresIn-30) * time.Second)
	return time.Now().Before(expirationTime)
}

// GetAccessToken 便捷方法，直接获取access_token字符串
func (tm *TokenManager) GetAccessToken() string {
	token, err := tm.GetValidToken()
	if err != nil {
		panic(err)
	}
	return token.AccessToken
}

// 使用示例：
/*
// 1. 创建token管理器（一般在应用启动时创建，作为全局实例）
tokenManager := NewTokenManager(account)

// 2. 在需要使用token的地方，直接调用GetValidToken或GetAccessToken
// 方法会自动处理token过期和刷新逻辑

// 方式1：获取完整的token信息
token, err := tokenManager.GetValidToken()
if err != nil {
    log.Printf("获取token失败: %v", err)
    return
}
fmt.Printf("Access Token: %s, 过期时间: %d秒", token.AccessToken, token.ExpiresIn)

// 方式2：直接获取access_token字符串（推荐用于API调用）
accessToken, err := tokenManager.GetAccessToken()
if err != nil {
    log.Printf("获取access token失败: %v", err)
    return
}

// 在HTTP请求中使用
req.Header.Add("Authorization", "Bearer " + accessToken)
*/
