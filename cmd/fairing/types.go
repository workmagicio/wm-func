package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Fairing Questions API 响应结构（根据真实文档）
type FairingQuestionsResponse struct {
	Questions []FairingQuestion `json:"questions"`
	Status    string            `json:"status,omitempty"`
	Message   string            `json:"message,omitempty"`
}

// Fairing Responses API 响应结构（分页格式）
type FairingResponsesResponse struct {
	Data []FairingUserResponse `json:"data"`
	Prev *string               `json:"prev"`
	Next *string               `json:"next"`
}

// 通用响应结构体
type FairingResponse struct {
	Questions []FairingQuestion     `json:"questions,omitempty"`
	Responses []FairingUserResponse `json:"responses,omitempty"`
	Status    string                `json:"status,omitempty"`
	Message   string                `json:"message,omitempty"`
}

// Fairing Question 结构（根据API文档）
//type FairingQuestion struct {
//	Id                 int64                     `json:"id"`
//	AllowOther         bool                    `json:"allow_other"`
//	CustomerType       string                  `json:"customer_type"`
//	FrequencyType      string                  `json:"frequency_type"`
//	InsertedAt         string                  `json:"inserted_at"`
//	MaxResponses       int64                     `json:"max_responses"`
//	OtherPlaceholder   string                  `json:"other_placeholder"`
//	Prompt             string                  `json:"prompt"`
//	PublishedAt        *string                 `json:"published_at"`
//	RandomizeResponses bool                    `json:"randomize_responses"`
//	Responses          []FairingResponseOption `json:"responses"`
//	SubmitText         string                  `json:"submit_text"`
//	Type               string                  `json:"type"`
//	UpdatedAt          string                  `json:"updated_at"`
//}

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

// Fairing Response Option 结构
type FairingResponseOption struct {
	Id                    int64                         `json:"id"`
	Value                 string                        `json:"value"`
	ClarificationQuestion *FairingClarificationQuestion `json:"clarification_question"`
}

// Fairing Clarification Question 结构
type FairingClarificationQuestion struct {
	Id                 int64                   `json:"id"`
	AllowOther         bool                    `json:"allow_other"`
	MaxResponses       int64                   `json:"max_responses"`
	OtherPlaceholder   *string                 `json:"other_placeholder"`
	Prompt             string                  `json:"prompt"`
	RandomizeResponses bool                    `json:"randomize_responses"`
	Responses          []FairingResponseOption `json:"responses"`
	SubmitText         string                  `json:"submit_text"`
	Type               string                  `json:"type"`
}

// Fairing User Response 结构（根据真实API文档）
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

// Airbyte 数据库存储结构体（匹配数据库表结构）
type FairingData struct {
	TenantId            int64  `gorm:"column:wm_tenant_id"`
	AirbyteRawId        string `gorm:"column:_airbyte_raw_id"`
	AirbyteData         []byte `gorm:"column:_airbyte_data"`
	AirbyteExtractedAt  string `gorm:"column:_airbyte_extracted_at"`
	AirbyteLoadedAt     string `gorm:"column:_airbyte_loaded_at"`
	AirbyteMeta         string `gorm:"column:_airbyte_meta"`
	AirbyteGenerationId int64  `gorm:"column:_airbyte_generation_id"`
	ItemType            string `gorm:"-"` // 不映射到数据库，用于确定表名
}

// 根据数据类型返回对应的表名
func (f *FairingData) TableName() string {
	switch f.ItemType {
	case "question":
		return "airbyte_destination_v2.raw_fairing_questions"
	case "response":
		return "airbyte_destination_v2.raw_fairing_responses"
	default:
		return "airbyte_destination_v2.raw_fairing_questions" // 默认为questions表
	}
}

// 通用文本清理函数
func cleanTextFields(text string) string {
	if text == "" {
		return text
	}

	// 移除或替换可能导致问题的特殊字符
	// 移除换行符和制表符，替换为空格
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// 移除多余的空格
	text = strings.TrimSpace(text)
	// 将多个连续空格替换为单个空格
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text
}

// 转换函数 - Question
func (q FairingQuestion) TransformToFairingData(tenantId int64) FairingData {
	// 需要处理 prompt 中的特殊字符，防止数据库插入失败

	// 创建问题副本并清理 prompt 字段
	cleanedQuestion := q
	if cleanedQuestion.Prompt != "" {
		cleanedQuestion.Prompt = cleanTextFields(cleanedQuestion.Prompt)
	}

	// 将清理后的问题数据序列化为JSON
	jsonData, _ := json.Marshal(cleanedQuestion)

	// 生成唯一的Raw ID
	rawId := fmt.Sprintf("%d", q.Id)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return FairingData{
		TenantId:            tenantId,
		AirbyteRawId:        rawId,
		AirbyteData:         jsonData,
		AirbyteExtractedAt:  now,
		AirbyteLoadedAt:     now,
		AirbyteMeta:         `{}`,
		AirbyteGenerationId: 0,
		ItemType:            "question", // 设置类型用于确定表名
	}
}

// 转换函数 - Response
func (r FairingUserResponse) TransformToFairingData(tenantId int64) FairingData {
	// 创建响应副本并清理文本字段中的特殊字符
	cleanedResponse := r

	// 清理 Question 字段
	if cleanedResponse.Question != "" {
		cleanedResponse.Question = cleanTextFields(cleanedResponse.Question)
	}

	// 清理 Response 字段
	if cleanedResponse.Response != "" {
		cleanedResponse.Response = cleanTextFields(cleanedResponse.Response)
	}

	// 清理 OtherResponse 字段
	if cleanedResponse.OtherResponse != "" {
		cleanedResponse.OtherResponse = cleanTextFields(cleanedResponse.OtherResponse)
	}

	// 清理 ReferringQuestion 字段
	if cleanedResponse.ReferringQuestion != "" {
		cleanedResponse.ReferringQuestion = cleanTextFields(cleanedResponse.ReferringQuestion)
	}

	// 清理 ReferringQuestionResponse 字段
	if cleanedResponse.ReferringQuestionResponse != "" {
		cleanedResponse.ReferringQuestionResponse = cleanTextFields(cleanedResponse.ReferringQuestionResponse)
	}

	// 将清理后的响应数据序列化为JSON
	jsonData, _ := json.Marshal(cleanedResponse)

	// 使用响应的ID作为Raw ID
	rawId := r.Id

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return FairingData{
		TenantId:            tenantId,
		AirbyteRawId:        rawId,
		AirbyteData:         jsonData,
		AirbyteExtractedAt:  now,
		AirbyteLoadedAt:     now,
		AirbyteMeta:         `{}`,
		AirbyteGenerationId: 0,
		ItemType:            "response", // 设置类型用于确定表名
	}
}

// 同步状态结构体（支持增量同步）
type SyncState struct {
	Status       string     `json:"status"`
	Message      string     `json:"message"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RecordCount  int64      `json:"record_count"`   // 记录数量用于变化检测
	LastSyncTime *time.Time `json:"last_sync_time"` // 用于增量同步的时间戳
}

// Fairing 专属的同步状态结构体（简化版，不存储详细的slice信息）
type FairingSyncState struct {
	Status       string     `json:"status"`
	Message      string     `json:"message"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RecordCount  int64      `json:"record_count"`   // 总记录数量
	LastSyncTime *time.Time `json:"last_sync_time"` // 最后成功同步的时间点

	// Stream Slice 进度追踪（轻量级）
	SyncStartDate    *time.Time `json:"sync_start_date"`    // 本轮同步的起始日期
	SyncEndDate      *time.Time `json:"sync_end_date"`      // 本轮同步的结束日期
	CurrentSliceDate *time.Time `json:"current_slice_date"` // 当前处理到的日期
	CompletedSlices  int        `json:"completed_slices"`   // 已完成片段数
	TotalSlices      int        `json:"total_slices"`       // 总片段数

	// 同步配置
	InitialDays    int  `json:"initial_days"`     // 首次同步拉取天数（默认365天）
	IsInitialSync  bool `json:"is_initial_sync"`  // 是否为首次同步
	RecentSyncDays int  `json:"recent_sync_days"` // 近期数据同步天数（默认7天）
	SliceDays      int  `json:"slice_days"`       // 每个slice的天数（默认1天）

	// 兼容字段（保持向后兼容）
	FirstSyncDate     *time.Time `json:"first_sync_date,omitempty"`     // 首次同步的起始日期
	CurrentSyncDate   *time.Time `json:"current_sync_date,omitempty"`   // 当前正在同步的日期
	LastCompletedDate *time.Time `json:"last_completed_date,omitempty"` // 最后完成同步的日期
	CompletedDays     int        `json:"completed_days,omitempty"`      // 已完成同步的天数
	TotalDaysToSync   int        `json:"total_days_to_sync,omitempty"`  // 总共需要同步的天数
	CurrentBatchCount int64      `json:"current_batch_count,omitempty"` // 当前批次的记录数
}

// 创建默认的 Fairing 同步状态
func NewFairingSyncState() FairingSyncState {
	return FairingSyncState{
		Status:          STATUS_SUCCESS,
		Message:         "初始状态",
		UpdatedAt:       time.Now().UTC().Add(-2 * time.Hour), // 设为2小时前，确保可以被执行
		RecordCount:     0,
		InitialDays:     365, // 默认首次同步一年数据
		IsInitialSync:   true,
		RecentSyncDays:  7, // 默认近期数据为7天
		SliceDays:       1, // 默认每个slice为1天
		CompletedSlices: 0,
		TotalSlices:     0,
	}
}

func (s FairingSyncState) GetSliceDays() int {
	if s.SliceDays < 1 {
		s.SliceDays = 1
	}
	return s.SliceDays
}

// 初始化同步范围和片段数量
func (fs *FairingSyncState) InitializeSyncRange() error {
	now := time.Now().UTC()

	if fs.IsInitialSync {
		// 首次同步：从 InitialDays 天前到现在
		startDate := now.AddDate(0, 0, -fs.InitialDays)
		fs.SyncStartDate = &startDate
		fs.SyncEndDate = &now
		fs.CurrentSliceDate = &startDate
	} else {
		// 增量同步：从上次完成日期开始
		var startDate time.Time
		if fs.LastCompletedDate != nil {
			startDate = *fs.LastCompletedDate
		} else if fs.LastSyncTime != nil {
			startDate = *fs.LastSyncTime
		} else {
			// 没有历史记录，按近期数据处理
			startDate = now.AddDate(0, 0, -fs.RecentSyncDays)
		}
		fs.SyncStartDate = &startDate
		fs.SyncEndDate = &now
		fs.CurrentSliceDate = &startDate
	}

	// 计算总片段数
	if fs.SyncStartDate != nil && fs.SyncEndDate != nil {
		days := int(fs.SyncEndDate.Sub(*fs.SyncStartDate).Hours() / 24)
		if days < 0 {
			days = 0
		}
		sliceDays := fs.GetSliceDays()
		fs.TotalSlices = (days + sliceDays - 1) / sliceDays // 向上取整
	}

	return nil
}

// 获取当前需要处理的时间范围
func (fs *FairingSyncState) GetCurrentSliceRange() (since, until *time.Time) {
	if fs.CurrentSliceDate == nil || fs.SyncEndDate == nil {
		return nil, nil
	}

	// 检查是否已经完成
	if !fs.CurrentSliceDate.Before(*fs.SyncEndDate) {
		return nil, nil
	}

	since = fs.CurrentSliceDate
	sliceEndDate := fs.CurrentSliceDate.AddDate(0, 0, fs.GetSliceDays())
	if sliceEndDate.After(*fs.SyncEndDate) {
		sliceEndDate = *fs.SyncEndDate
	}
	until = &sliceEndDate

	return since, until
}

// 标记当前片段完成，移到下一个片段
func (fs *FairingSyncState) CompleteCurrentSlice(recordCount int64) {
	if fs.CurrentSliceDate != nil {
		// 移动到下一个slice
		nextDate := fs.CurrentSliceDate.AddDate(0, 0, fs.GetSliceDays())
		fs.CurrentSliceDate = &nextDate

		// 更新统计
		fs.CompletedSlices++
		fs.RecordCount += recordCount
		fs.LastSyncTime = &nextDate

		// 检查是否完成所有slice
		if fs.SyncEndDate != nil && !fs.CurrentSliceDate.Before(*fs.SyncEndDate) {
			// 所有slice完成
			if fs.IsInitialSync {
				fs.IsInitialSync = false
			}
			fs.LastCompletedDate = fs.SyncEndDate
		}
	}
}

// 判断是否所有片段都已完成
func (fs *FairingSyncState) IsAllSlicesCompleted() bool {
	return fs.CurrentSliceDate != nil && fs.SyncEndDate != nil &&
		!fs.CurrentSliceDate.Before(*fs.SyncEndDate)
}

// 获取同步进度百分比
func (fs *FairingSyncState) GetProgress() float64 {
	if fs.TotalSlices == 0 {
		return 0
	}
	return float64(fs.CompletedSlices) / float64(fs.TotalSlices) * 100
}

// 兼容原有接口
func (fs *FairingSyncState) CalculateSyncRange() (since, until *time.Time) {
	return fs.GetCurrentSliceRange()
}

// 判断是否需要继续同步（兼容原有接口）
func (fs *FairingSyncState) ShouldContinueSync() bool {
	return !fs.IsAllSlicesCompleted()
}

// 更新同步进度（兼容原有接口）
func (fs *FairingSyncState) UpdateProgress(recordCount int64, syncDate time.Time) {
	fs.CompleteCurrentSlice(recordCount)
	fs.UpdatedAt = time.Now().UTC()
}
