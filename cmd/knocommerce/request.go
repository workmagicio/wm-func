package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"wm-func/wm_account"
)

func RefreshToken(account wm_account.Account) (*RefreshTokenResponse, error) {
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

// GetAllKnoCommerceResponses 函数会自动处理分页，获取指定日期范围内的所有回复
func GetAllKnoCommerceResponses(accessToken, startDate, endDate string) ([]Result, error) {
	var allResults []Result
	page := 1
	// 设定一个合理的页面大小，例如 50，以减少 API 调用次数
	const pageSize = 250

	for {
		fmt.Printf("正在获取第 %d 页数据...\n", page)
		response, err := GetKnoCommerceResponses(accessToken, startDate, endDate, page, pageSize)
		if err != nil {
			// 如果在获取某一页时出错，返回已获取的数据和错误
			return allResults, fmt.Errorf("获取第 %d 页数据时出错: %w", page, err)
		}

		// 如果当前页没有结果，说明已经获取完毕
		if len(response.Results) == 0 {
			break
		}

		// 将当前页的结果追加到总结果列表中
		allResults = append(allResults, response.Results...)

		// 如果已获取的结果数量大于或等于总数，说明已经获取完毕
		if response.Total > 0 && len(allResults) >= response.Total {
			break
		}

		// 准备获取下一页
		page++
		time.Sleep(time.Second * 5)
	}

	return allResults, nil
}

// GetKnoCommerceResponses 函数用于获取调查问卷的回复列表
// 它接收分页参数、日期范围和访问令牌
func GetKnoCommerceResponses(accessToken, startDate, endDate string, page, pageSize int) (*APIResponse, error) {
	// 1. 定义 API 基础 URL
	baseURL := "https://app-api.knocommerce.com/api/rest/responses"

	// 2. 准备 URL 查询参数
	params := url.Values{}
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))
	params.Add("startDate", startDate)
	params.Add("endDate", endDate)
	params.Add("expand[]", "order")

	// 3. 构造完整的请求 URL
	fullURL := baseURL + "?" + params.Encode()

	// 4. 创建一个新的 GET 请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 responses 请求失败: %w", err)
	}

	// 5. 设置 Authorization 请求头，使用 Bearer Token
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// 6. 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送 responses 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 7. 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("responses API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 8. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 responses 响应体失败: %w", err)
	}

	// 9. 将 JSON 响应体解析到 APIResponse 结构体中
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("解析 responses JSON 失败: %w", err)
	}

	// 10. 返回解析后的数据
	return &apiResponse, nil
}

// GetKnoCommerceResponsesCount 函数用于获取指定日期范围内的回复总数
func GetKnoCommerceResponsesCount(accessToken, startDate, endDate string) (int64, error) {
	startDate = fmt.Sprintf("%sT00:00:00.000Z", startDate)
	endDate = fmt.Sprintf("%sT23:59:59.999Z", endDate)
	// 1. 定义 API 基础 URL
	baseURL := "https://app-api.knocommerce.com/api/rest/responses/count"

	// 2. 准备 URL 查询参数
	params := url.Values{}
	// API 需要 ISO 8601 格式 (YYYY-MM-DDTHH:MM:SS.sssZ)
	params.Add("createdAt[gte]", startDate)
	params.Add("createdAt[lte]", endDate)

	// 3. 构造完整的请求 URL
	fullURL := baseURL + "?" + params.Encode()

	// 4. 创建一个新的 GET 请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("创建 count 请求失败: %w", err)
	}

	// 5. 设置 Authorization 请求头
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// 6. 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("发送 count 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 7. 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("count API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 8. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取 count 响应体失败: %w", err)
	}

	// 9. 将 JSON 响应体解析到 ResponseCount 结构体中
	var responseCount ResponseCount
	if err := json.Unmarshal(body, &responseCount); err != nil {
		return 0, fmt.Errorf("解析 count JSON 失败: %w", err)
	}

	// 10. 返回计数
	return responseCount.Count, nil
}

func GetKnoCommerceQuestion(accessToken string) (*BenchmarkResponse, error) {
	// 1. 定义 API URL
	apiURL := "https://app-api.knocommerce.com/api/rest/questions/benchmarks"

	// 2. 创建一个新的 GET 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 benchmarks 请求失败: %w", err)
	}

	// 3. 设置 Authorization 请求头
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// 4. 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送 benchmarks 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 5. 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("benchmarks API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 6. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 benchmarks 响应体失败: %w", err)
	}

	// 7. 将 JSON 响应体解析到 BenchmarkResponse 结构体中
	var benchmarkResponse BenchmarkResponse
	if err := json.Unmarshal(body, &benchmarkResponse); err != nil {
		return nil, fmt.Errorf("解析 benchmarks JSON 失败: %w", err)
	}

	// 8. 返回解析后的数据
	return &benchmarkResponse, nil
}
func GetAllKnoCommerceSurveys(accessToken string) ([]Survey, error) {
	var allSurveys []Survey
	page := 1
	const pageSize = 50 // 使用一个合理的页面大小以减少API调用次数

	for {
		fmt.Printf("正在获取第 %d 页的调查问卷...\n", page)
		response, err := GetKnoCommerceSurveys(accessToken, page, pageSize)
		if err != nil {
			return allSurveys, fmt.Errorf("获取第 %d 页调查问卷时出错: %w", page, err)
		}

		// 如果当前页没有结果，说明已经获取完毕
		if len(response.Results) == 0 {
			break
		}

		// 将当前页的结果追加到总结果列表中
		allSurveys = append(allSurveys, response.Results...)

		// 如果已获取的结果数量大于或等于总数，说明已经获取完毕
		if response.Total > 0 && len(allSurveys) >= response.Total {
			break
		}

		// 准备获取下一页
		page++
		time.Sleep(time.Second * 5)
	}

	return allSurveys, nil
}

// GetKnoCommerceSurveys 函数用于获取调查问卷列表
func GetKnoCommerceSurveys(accessToken string, page, pageSize int) (*SurveysResponse, error) {
	// 1. 定义 API 基础 URL
	baseURL := "https://app-api.knocommerce.com/api/rest/surveys"

	// 2. 准备 URL 查询参数
	params := url.Values{}
	params.Add("page", fmt.Sprintf("%d", page))
	params.Add("pageSize", fmt.Sprintf("%d", pageSize))

	// 3. 构造完整的请求 URL
	fullURL := baseURL + "?" + params.Encode()

	// 4. 创建一个新的 GET 请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 surveys 请求失败: %w", err)
	}

	// 5. 设置 Authorization 请求头
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// 6. 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送 surveys 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 7. 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("surveys API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
	}

	// 8. 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 surveys 响应体失败: %w", err)
	}

	// 9. 将 JSON 响应体解析到 SurveysResponse 结构体中
	var surveysResponse SurveysResponse
	if err := json.Unmarshal(body, &surveysResponse); err != nil {
		return nil, fmt.Errorf("解析 surveys JSON 失败: %w", err)
	}

	// 10. 返回解析后的数据
	return &surveysResponse, nil
}
