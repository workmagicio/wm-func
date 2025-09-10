package main

import (
	"fmt"
	"time"
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

const (
	SubTypeRequest  = "request"
	SubTypeResponse = "response"
)

type FAccount struct {
	wm_account.Account
	lock2.Locker
}

// GetSimpleTraceId 获取简化的跟踪ID (只包含TenantId，不包含AccountId)
func (ka FAccount) GetSimpleTraceId() string {
	return fmt.Sprintf("%d", ka.TenantId)
}

// GetTraceIdWithSubType 获取包含子类型的跟踪ID
func (ka FAccount) GetTraceIdWithSubType(subType string) string {
	return fmt.Sprintf("%d-%s", ka.TenantId, subType)
}

type FairingQuestionResponse struct {
	Data []FairingQuestion `json:"data"`
}

type FairingQuestion struct {
	Id                 int64     `json:"id"`
	Type               string    `json:"type"`
	Prompt             string    `json:"prompt"`
	InsertedAt         time.Time `json:"inserted_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	RandomizeResponses bool      `json:"randomize_responses"`
	OtherPlaceholder   string    `json:"other_placeholder"`
	AllowOther         bool      `json:"allow_other"`
	PublishedAt        time.Time `json:"published_at"`
	CustomerType       string    `json:"customer_type"`
	FrequencyType      string    `json:"frequency_type"`
	SubmitText         string    `json:"submit_text"`
}

type FairingResponsesResponse struct {
	Data []FairingUserResponse `json:"data"`
	Prev *string               `json:"prev"`
	Next *string               `json:"next"`
}

type FairingUserResponse struct {
	AvailableResponses          []string  `json:"available_responses"`
	ClarificationQuestion       bool      `json:"clarification_question"`
	CouponAmount                string    `json:"coupon_amount"`
	CouponCode                  string    `json:"coupon_code"`
	CouponType                  string    `json:"coupon_type"`
	CustomerId                  string    `json:"customer_id"`
	CustomerOrderCount          int64     `json:"customer_order_count"`
	Email                       string    `json:"email"`
	Id                          string    `json:"id"`
	InsertedAt                  time.Time `json:"inserted_at"`
	LandingPagePath             string    `json:"landing_page_path"`
	OrderCurrencyCode           string    `json:"order_currency_code"`
	OrderId                     string    `json:"order_id"`
	OrderNumber                 string    `json:"order_number"`
	OrderPlatform               string    `json:"order_platform"`
	OrderSource                 string    `json:"order_source"`
	OrderTotal                  string    `json:"order_total"`
	OrderTotalUsd               string    `json:"order_total_usd"`
	Other                       bool      `json:"other"`
	OtherResponse               string    `json:"other_response"`
	Question                    string    `json:"question"`
	QuestionId                  int64     `json:"question_id"`
	QuestionType                string    `json:"question_type"`
	ReferringQuestion           string    `json:"referring_question"`
	ReferringQuestionId         int64     `json:"referring_question_id"`
	ReferringQuestionResponse   string    `json:"referring_question_response"`
	ReferringQuestionResponseId int64     `json:"referring_question_response_id"`
	ReferringSite               string    `json:"referring_site"`
	Response                    string    `json:"response"`
	ResponseId                  int64     `json:"response_id"`
	ResponsePosition            int64     `json:"response_position"`
	ResponseProvidedAt          time.Time `json:"response_provided_at"`
	SubmitDelta                 int64     `json:"submit_delta"`
	UpdatedAt                   time.Time `json:"updated_at"`
	UtmCampaign                 string    `json:"utm_campaign"`
	UtmContent                  string    `json:"utm_content"`
	UtmMedium                   string    `json:"utm_medium"`
	UtmSource                   string    `json:"utm_source"`
	UtmTerm                     string    `json:"utm_term"`
}
