package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// GetAllKnoCommerceResponses 函数会自动处理分页，获取指定日期范围内的所有回复
func GetAllKnoCommerceResponses(account KAccount, token *TokenManager, startDate, endDate string) ([]Result, error) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE)
	var allResults []Result
	page := 1
	// 设定一个合理的页面大小，例如 50，以减少 API 调用次数
	const pageSize = 250
	var totalCount int = 0

	log.Printf("[%s] 开始分页获取回复数据，日期范围: %s 至 %s", traceId, startDate, endDate)

	for {
		response, err := GetKnoCommerceResponses(token, startDate, endDate, page, pageSize)
		if err != nil {
			// 如果在获取某一页时出错，返回已获取的数据和错误
			return allResults, fmt.Errorf("获取第 %d 页数据时出错: %w", page, err)
		}

		// 第一页时记录总数
		if page == 1 {
			totalCount = response.Total
			var estimatedPages int
			if totalCount > 0 && pageSize > 0 {
				estimatedPages = (totalCount + pageSize - 1) / pageSize
			} else {
				estimatedPages = 1
			}
			log.Printf("[%s] API返回总计数量: %d，预计页数: %d",
				traceId, totalCount, estimatedPages)
		}

		// 如果当前页没有结果，说明已经获取完毕
		if len(response.Results) == 0 {
			log.Printf("[%s] 第 %d 页无数据，获取完毕", traceId, page)
			break
		}

		// 将当前页的结果追加到总结果列表中
		allResults = append(allResults, response.Results...)

		var progressText string
		if totalCount > 0 {
			progressText = fmt.Sprintf("累计获取: %d/%d 条 (%.1f%%)",
				len(allResults), totalCount, float64(len(allResults))/float64(totalCount)*100)
		} else {
			progressText = fmt.Sprintf("累计获取: %d 条", len(allResults))
		}
		log.Printf("[%s] 成功获取第 %d 页，当前页数据: %d 条，%s",
			traceId, page, len(response.Results), progressText)

		// 如果已获取的结果数量大于或等于总数，说明已经获取完毕
		if response.Total > 0 && len(allResults) >= response.Total {
			log.Printf("[%s] 已获取全部数据，总共: %d 条", traceId, len(allResults))
			break
		}

		// 准备获取下一页
		page++
		time.Sleep(time.Second * 2)
	}

	log.Printf("[%s] 分页获取完成，最终获得 %d 条回复数据", traceId, len(allResults))
	return allResults, nil
}

// GetKnoCommerceResponses 函数用于获取调查问卷的回复列表
// 它接收分页参数、日期范围和访问令牌
func GetKnoCommerceResponses(token *TokenManager, startDate, endDate string, page, pageSize int) (*APIResponse, error) {
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
	req.Header.Add("Authorization", "Bearer "+token.GetAccessToken())

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
func GetKnoCommerceResponsesCount(token *TokenManager, startDate, endDate string) (int64, error) {
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
	req.Header.Add("Authorization", "Bearer "+token.GetAccessToken())

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
func GetAllKnoCommerceSurveys(account KAccount, token *TokenManager) ([]Survey, error) {
	traceId := account.GetTraceIdWithSubType(SUBTYPE_SURVEY)
	var allSurveys []Survey
	page := 1
	const pageSize = 50 // 使用一个合理的页面大小以减少API调用次数
	var totalCount int = 0

	log.Printf("[%s] 开始分页获取调查问卷数据", traceId)

	for {
		response, err := GetKnoCommerceSurveys(token.GetAccessToken(), page, pageSize)
		if err != nil {
			return allSurveys, fmt.Errorf("获取第 %d 页调查问卷时出错: %w", page, err)
		}

		// 第一页时记录总数
		if page == 1 {
			totalCount = response.Total
			var estimatedPages int
			if totalCount > 0 && pageSize > 0 {
				estimatedPages = (totalCount + pageSize - 1) / pageSize
			} else {
				estimatedPages = 1
			}
			log.Printf("[%s] API返回调查问卷总数: %d，每页大小: %d，预计页数: %d",
				traceId, totalCount, pageSize, estimatedPages)
		}

		// 如果当前页没有结果，说明已经获取完毕
		if len(response.Results) == 0 {
			log.Printf("[%s] 第 %d 页无数据，获取完毕", traceId, page)
			break
		}

		// 将当前页的结果追加到总结果列表中
		allSurveys = append(allSurveys, response.Results...)

		var progressText string
		if totalCount > 0 {
			progressText = fmt.Sprintf("累计获取: %d/%d 条 (%.1f%%)",
				len(allSurveys), totalCount, float64(len(allSurveys))/float64(totalCount)*100)
		} else {
			progressText = fmt.Sprintf("累计获取: %d 条", len(allSurveys))
		}
		log.Printf("[%s] 成功获取第 %d 页，当前页数据: %d 条，%s",
			traceId, page, len(response.Results), progressText)

		// 如果已获取的结果数量大于或等于总数，说明已经获取完毕
		if response.Total > 0 && len(allSurveys) >= response.Total {
			log.Printf("[%s] 已获取全部调查问卷数据，总共: %d 条", traceId, len(allSurveys))
			break
		}

		// 准备获取下一页
		page++
		time.Sleep(time.Second * 2)
	}

	log.Printf("[%s] 分页获取完成，最终获得 %d 条调查问卷数据", traceId, len(allSurveys))
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
