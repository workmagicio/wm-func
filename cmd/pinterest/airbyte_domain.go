package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
	"wm-func/common/model"
)

type OdsCampaignCache struct {
	model.AirbyteRawData
}

func (f OdsCampaignCache) TableName() string {
	return "airbyte_destination_v2.raw_pinterest_campaigns"
}

type RawPinterestAdGroups struct {
	model.AirbyteRawData
}

func (f RawPinterestAdGroups) TableName() string {
	return "airbyte_destination_v2.raw_pinterest_ad_groups"
}

type RawPinterestAds struct {
	model.AirbyteRawData
}

func (f RawPinterestAds) TableName() string {
	return "airbyte_destination_v2.raw_pinterest_ads"
}

type RawPinterestAdAnalytics struct {
	model.AirbyteRawData
}

func (f RawPinterestAdAnalytics) TableName() string {
	return "airbyte_destination_v2.raw_pinterest_ad_analytics"
}

type RawPinterestAdGroupAnalytics struct {
	model.AirbyteRawData
}

func (f RawPinterestAdGroupAnalytics) TableName() string {
	return "airbyte_destination_v2.raw_pinterest_ad_group_analytics"
}

func Save(data []OdsCampaignCache) error {
	return model.SaveAirbyteData(data)
}

func SaveAdGroups(data []RawPinterestAdGroups) error {
	return model.SaveAirbyteData(data)
}

func SaveAds(data []RawPinterestAds) error {
	return model.SaveAirbyteData(data)
}

func SaveAdAnalytics(data []RawPinterestAdAnalytics) error {
	return model.SaveAirbyteData(data)
}

func SaveAdGroupAnalytics(data []RawPinterestAdGroupAnalytics) error {
	return model.SaveAirbyteData(data)
}

// 生成唯一的 Raw ID
func generateCampaignRawId(campaign Campaign) string {
	// 使用 Campaign ID 和更新时间生成唯一标识
	return fmt.Sprintf("%s", campaign.Id)
}

// 将 Campaign 转换为 OdsCampaignCache
func TransformCampaignToAirbyte(campaign Campaign, tenantId int64) OdsCampaignCache {
	// 处理时间字段：将 0 值转换为 null
	if campaign.StartTime != nil && *campaign.StartTime == 0 {
		campaign.StartTime = nil
	}

	// 将 Campaign 数据序列化为 JSON（不转义HTML字符）
	jsonData, _ := marshalJSONWithoutHTMLEscape(campaign)

	// 生成唯一的 Raw ID
	rawId := generateCampaignRawId(campaign)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return OdsCampaignCache{
		AirbyteRawData: model.AirbyteRawData{
			TenantId:            tenantId,
			AirbyteRawId:        rawId,
			AirbyteData:         jsonData,
			AirbyteExtractedAt:  now,
			AirbyteLoadedAt:     now,
			AirbyteMeta:         `{"changes":[]}`,
			AirbyteGenerationId: time.Now().Unix(),
		},
	}
}

// 批量转换并保存 Campaign 数据到 Airbyte
func SaveCampaignsToAirbyte(campaigns []Campaign, tenantId int64) error {
	// 以下代码暂时被注释，等开发完成后启用
	if len(campaigns) == 0 {
		return nil
	}

	// 转换为 Airbyte 格式
	airbyteData := make([]OdsCampaignCache, len(campaigns))
	for i, campaign := range campaigns {
		airbyteData[i] = TransformCampaignToAirbyte(campaign, tenantId)
	}

	// 保存到数据库
	return Save(airbyteData)
}

// 生成唯一的 AdGroup Raw ID
func generateAdGroupRawId(adGroup AdGroup) string {
	// 使用 AdGroup ID 生成唯一标识
	return fmt.Sprintf("%s", adGroup.Id)
}

// convertZeroToNull 将 0 值转换为 null 指针
func convertZeroToNull(value int64) *int64 {
	if value == 0 {
		return nil
	}
	return &value
}

// marshalJSONWithoutHTMLEscape 序列化为JSON但不转义HTML字符
func marshalJSONWithoutHTMLEscape(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 不转义HTML字符，保持 & < > 等字符原样
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	// 移除末尾的换行符
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

// 将 AdGroup 转换为 RawPinterestAdGroups
func TransformAdGroupToAirbyte(adGroup AdGroup, tenantId int64) RawPinterestAdGroups {
	// 处理时间字段：将 0 值转换为 null
	if adGroup.StartTime != nil && *adGroup.StartTime == 0 {
		adGroup.StartTime = nil
	}
	if adGroup.EndTime != nil && *adGroup.EndTime == 0 {
		adGroup.EndTime = nil
	}

	// 将 AdGroup 数据序列化为 JSON（不转义HTML字符）
	jsonData, _ := marshalJSONWithoutHTMLEscape(adGroup)

	// 生成唯一的 Raw ID
	rawId := generateAdGroupRawId(adGroup)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return RawPinterestAdGroups{
		AirbyteRawData: model.AirbyteRawData{
			TenantId:            tenantId,
			AirbyteRawId:        rawId,
			AirbyteData:         jsonData,
			AirbyteExtractedAt:  now,
			AirbyteLoadedAt:     now,
			AirbyteMeta:         `{"changes":[]}`,
			AirbyteGenerationId: time.Now().Unix(),
		},
	}
}

// 批量转换并保存 AdGroup 数据到 Airbyte
func SaveAdGroupsToAirbyte(adGroups []AdGroup, tenantId int64) error {
	if len(adGroups) == 0 {
		return nil
	}

	// 转换为 Airbyte 格式
	airbyteData := make([]RawPinterestAdGroups, len(adGroups))
	for i, adGroup := range adGroups {
		airbyteData[i] = TransformAdGroupToAirbyte(adGroup, tenantId)
	}

	// 保存到数据库
	return SaveAdGroups(airbyteData)
}

// 生成唯一的 Ad Raw ID
func generateAdRawId(ad Ad) string {
	// 使用 Ad ID 生成唯一标识
	return fmt.Sprintf("%s", ad.Id)
}

// 将 Ad 转换为 RawPinterestAds
func TransformAdToAirbyte(ad Ad, tenantId int64) RawPinterestAds {
	// 将 Ad 数据序列化为 JSON（不转义HTML字符）
	jsonData, _ := marshalJSONWithoutHTMLEscape(ad)

	// 生成唯一的 Raw ID
	rawId := generateAdRawId(ad)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return RawPinterestAds{
		AirbyteRawData: model.AirbyteRawData{
			TenantId:            tenantId,
			AirbyteRawId:        rawId,
			AirbyteData:         jsonData,
			AirbyteExtractedAt:  now,
			AirbyteLoadedAt:     now,
			AirbyteMeta:         `{"changes":[]}`,
			AirbyteGenerationId: time.Now().Unix(),
		},
	}
}

// 批量转换并保存 Ad 数据到 Airbyte
func SaveAdsToAirbyte(ads []Ad, tenantId int64) error {
	if len(ads) == 0 {
		return nil
	}

	// 转换为 Airbyte 格式
	airbyteData := make([]RawPinterestAds, len(ads))
	for i, ad := range ads {
		airbyteData[i] = TransformAdToAirbyte(ad, tenantId)
	}

	// 保存到数据库
	return SaveAds(airbyteData)
}

// 生成唯一的 Ad Analytics Raw ID
func generateAdAnalyticsRawId(adMetrics AdMetrics) string {
	// 使用 DATE|0|CAMPAIGN_ID|AD_GROUP_ID|AD_ID 格式生成唯一标识

	return fmt.Sprintf("%s|0|%d|%d|%s", adMetrics.DATE, adMetrics.CAMPAIGNID, adMetrics.ADGROUPID, adMetrics.ADID)
}

// 将 AdMetrics 转换为 RawPinterestAdAnalytics
func TransformAdAnalyticsToAirbyte(adMetrics AdMetrics, tenantId int64) RawPinterestAdAnalytics {
	// 设置 tenant_id
	adMetrics.TenantId = tenantId

	// 将 AdMetrics 数据序列化为 JSON（不转义HTML字符）
	jsonData, _ := marshalJSONWithoutHTMLEscape(adMetrics)

	// 生成唯一的 Raw ID
	rawId := generateAdAnalyticsRawId(adMetrics)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return RawPinterestAdAnalytics{
		AirbyteRawData: model.AirbyteRawData{
			TenantId:            tenantId,
			AirbyteRawId:        rawId,
			AirbyteData:         jsonData,
			AirbyteExtractedAt:  now,
			AirbyteLoadedAt:     now,
			AirbyteMeta:         `{"changes":[]}`,
			AirbyteGenerationId: time.Now().Unix(),
		},
	}
}

// 批量转换并保存 Ad Analytics 数据到 Airbyte
func SaveAdAnalyticsToAirbyte(adMetrics []AdMetrics, tenantId int64) error {
	if len(adMetrics) == 0 {
		return nil
	}

	// 转换为 Airbyte 格式
	airbyteData := make([]RawPinterestAdAnalytics, len(adMetrics))
	for i, metrics := range adMetrics {
		airbyteData[i] = TransformAdAnalyticsToAirbyte(metrics, tenantId)
	}

	// 保存到数据库
	return SaveAdAnalytics(airbyteData)
}

// 生成唯一的 AdGroup Analytics Raw ID
func generateAdGroupAnalyticsRawId(adGroupMetrics AdGroupMetrics) string {
	// 使用 DATE|0|CAMPAIGN_ID|AD_GROUP_ID|AD_ID 格式生成唯一标识
	return fmt.Sprintf("%s|0|%d|%s", adGroupMetrics.DATE, adGroupMetrics.CAMPAIGNID, adGroupMetrics.ADGROUPID)
}

// 将 AdGroupMetrics 转换为 RawPinterestAdGroupAnalytics
func TransformAdGroupAnalyticsToAirbyte(adGroupMetrics AdGroupMetrics, tenantId int64) RawPinterestAdGroupAnalytics {
	// 设置 tenant_id
	adGroupMetrics.TenantId = tenantId

	// 将 AdGroupMetrics 数据序列化为 JSON（不转义HTML字符）
	jsonData, _ := marshalJSONWithoutHTMLEscape(adGroupMetrics)

	// 生成唯一的 Raw ID
	rawId := generateAdGroupAnalyticsRawId(adGroupMetrics)

	// 当前时间
	now := time.Now().Format("2006-01-02 15:04:05")

	return RawPinterestAdGroupAnalytics{
		AirbyteRawData: model.AirbyteRawData{
			TenantId:            tenantId,
			AirbyteRawId:        rawId,
			AirbyteData:         jsonData,
			AirbyteExtractedAt:  now,
			AirbyteLoadedAt:     now,
			AirbyteMeta:         `{"changes":[]}`,
			AirbyteGenerationId: time.Now().Unix(),
		},
	}
}

// 批量转换并保存 AdGroup Analytics 数据到 Airbyte
func SaveAdGroupAnalyticsToAirbyte(adGroupMetrics []AdGroupMetrics, tenantId int64) error {
	if len(adGroupMetrics) == 0 {
		return nil
	}

	// 转换为 Airbyte 格式
	airbyteData := make([]RawPinterestAdGroupAnalytics, len(adGroupMetrics))
	for i, metrics := range adGroupMetrics {
		airbyteData[i] = TransformAdGroupAnalyticsToAirbyte(metrics, tenantId)
	}

	// 保存到数据库
	return SaveAdGroupAnalytics(airbyteData)
}
