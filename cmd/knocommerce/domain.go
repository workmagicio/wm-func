package main

import (
	"fmt"
	"time"
)

const (
	SUBTYPE_QUESTION       = "questions"
	SUBTYPE_RESPONSE       = "responses"
	SUBTYPE_RESPONSE_COUNT = "response_count"
	SUBTYPE_SURVEY         = "surveys"
)

func GetAirbyteDbNameWithSubType(subType string) string {
	return fmt.Sprintf("airbyte_destination_v2.raw_knocommerce_%s", subType)
}

var subTypeList = []string{SUBTYPE_QUESTION, SUBTYPE_RESPONSE, SUBTYPE_RESPONSE_COUNT, SUBTYPE_SURVEY}

var timeRageHourMap = map[string]int{
	SUBTYPE_RESPONSE:       1,
	SUBTYPE_SURVEY:         1,
	SUBTYPE_RESPONSE_COUNT: 12,
	SUBTYPE_QUESTION:       1,
}

var timeStepDailyMap = map[string]int{
	SUBTYPE_RESPONSE:       -1,
	SUBTYPE_SURVEY:         -1,
	SUBTYPE_RESPONSE_COUNT: 1,
}

type ResponseCount struct {
	Count int `json:"count"`
}

type RefreshTokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

// BenchmarkResponse 结构是 /questions/benchmarks 端点的顶层响应结构
type BenchmarkResponse struct {
	Data BenchmarkData `json:"data"`
}

// BenchmarkData 包含问题列表
type BenchmarkData struct {
	Questions []BenchmarkQuestion `json:"questions"`
}

// BenchmarkQuestion 代表一个基准问题
type BenchmarkQuestion struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
}

type QuestionResponse struct {
	Data []Questions `json:"data"`
}

type Questions []struct {
	Id    string `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
}

// =================================================================
// Section 2: Structures for API Responses
// =================================================================

// APIResponse 结构是 /responses 端点的顶层响应结构
type APIResponse struct {
	Total   int      `json:"total"`
	Results []Result `json:"results"`
}

// Result 结构代表返回结果数组中的单个调查回复
type Result struct {
	ID                     string      `json:"id"`
	AccountID              string      `json:"account_id"`
	CreatedAt              time.Time   `json:"created_at"`
	CompletedAt            time.Time   `json:"completed_at"`
	CustomerID             string      `json:"customer_id"`
	CustomerEmail          string      `json:"customer_email"`
	CustomerPhone          interface{} `json:"customer_phone"` // 使用 interface{} 因为它可能是 null
	CustomerShop           string      `json:"customer_shop"`
	CustomerLifetimeSpent  int         `json:"customer_lifetime_spent"`
	CustomerLifetimeOrders int         `json:"customer_lifetime_orders"`
	TimeSpent              interface{} `json:"time_spent"` // 使用 interface{} 因为它可能是 null
	SurveyID               string      `json:"survey_id"`
	Order                  Order       `json:"order"`
	Response               []Response  `json:"response"`
}

// Order 结构代表与调查回复关联的订单信息
type Order struct {
	ID          string      `json:"id"`
	OrderID     string      `json:"order_id"`
	OrderNumber string      `json:"order_number"`
	TotalPrice  float64     `json:"total_price"` // 使用 float64 以处理可能的浮点数价格
	Currency    string      `json:"currency"`
	BrowserIP   interface{} `json:"browser_ip"` // 使用 interface{} 因为它可能是 null
	UserAgent   string      `json:"user_agent"`
}

// Response 结构代表调查问卷中的具体问题和答案
type Response struct {
	Value      string `json:"value"`
	Type       string `json:"type"`
	Label      string `json:"label"`
	QuestionID string `json:"question_id"`
}
