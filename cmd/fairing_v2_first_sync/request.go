package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"wm-func/common/http_request"
)

func getTraceIdWithSubType(account FAccount, subType string) string {
	return account.GetTraceId() + "-" + subType
}

func callFairingQuestionsAPI(account FAccount) ([]FairingQuestion, error) {
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

func callFairingResponsesAPI(account FAccount, since string, until string, limit int) (*FairingResponsesResponse, error) {
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
	params["since"] = since
	//params["until"] = until

	apiURL := "https://app.fairing.co/api/responses"

	log.Printf("[%s] 调用Fairing Responses API: %s", traceId, apiURL)

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

func callFairingResponsesAPINext(account FAccount, url string) (*FairingResponsesResponse, error) {
	traceId := getTraceIdWithSubType(account, "response")

	headers := map[string]string{
		"Authorization": fmt.Sprintf("%s", account.SecretToken),
		"Content-Type":  "application/json",
		"Accept":        "application/json",
	}

	apiURL := url

	log.Printf("[%s] 调用Fairing Responses API: %s", traceId, apiURL)

	response, err := http_request.Get(apiURL, headers, map[string]string{})
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
