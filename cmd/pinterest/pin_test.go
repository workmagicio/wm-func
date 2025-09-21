package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/pinterest/a.json")
	if err != nil {
		panic(err)
	}
	res := map[string][]ReportData{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		panic(err)
	}
}

func TestNullTimeFields(t *testing.T) {
	// 测试 AdGroup 中 StartTime 和 EndTime 为 0 时的处理
	adGroupJSON := `{
		"id": "test123",
		"ad_account_id": "549756365874",
		"campaign_id": "626754086067",
		"name": "Test AdGroup",
		"status": "ACTIVE",
		"start_time": 0,
		"end_time": 0,
		"created_time": 1734395678,
		"updated_time": 1734395678
	}`

	var adGroup AdGroup
	err := json.Unmarshal([]byte(adGroupJSON), &adGroup)
	if err != nil {
		t.Fatalf("解析AdGroup JSON失败: %v", err)
	}

	// 验证 StartTime 和 EndTime 为 0 时是否正确处理
	t.Logf("StartTime: %v", adGroup.StartTime)
	t.Logf("EndTime: %v", adGroup.EndTime)

	// 当值为0时，指针应该指向0值
	if adGroup.StartTime != nil && *adGroup.StartTime == 0 {
		t.Log("StartTime 为 0，将在数据库中显示为 0")
	}

	if adGroup.EndTime != nil && *adGroup.EndTime == 0 {
		t.Log("EndTime 为 0，将在数据库中显示为 0")
	}

	// 测试 null 值的情况
	adGroupNullJSON := `{
		"id": "test456",
		"ad_account_id": "549756365874",
		"campaign_id": "626754086067",
		"name": "Test AdGroup Null",
		"status": "ACTIVE",
		"start_time": null,
		"end_time": null,
		"created_time": 1734395678,
		"updated_time": 1734395678
	}`

	var adGroupNull AdGroup
	err = json.Unmarshal([]byte(adGroupNullJSON), &adGroupNull)
	if err != nil {
		t.Fatalf("解析AdGroup Null JSON失败: %v", err)
	}

	// 验证 null 值是否正确处理
	if adGroupNull.StartTime == nil {
		t.Log("StartTime 为 null，将在数据库中显示为 NULL")
	}

	if adGroupNull.EndTime == nil {
		t.Log("EndTime 为 null，将在数据库中显示为 NULL")
	}

	// 测试序列化回 JSON
	serialized, err := json.Marshal(adGroupNull)
	if err != nil {
		t.Fatalf("序列化AdGroup失败: %v", err)
	}

	t.Logf("序列化后的JSON: %s", string(serialized))
}

func TestTransformAdGroupWithZeroTime(t *testing.T) {
	// 创建一个 StartTime 和 EndTime 为 0 的 AdGroup
	startTime := int64(0)
	endTime := int64(0)

	adGroup := AdGroup{
		Id:          "test123",
		AdAccountId: "549756365874",
		CampaignId:  "626754086067",
		Name:        "Test AdGroup",
		Status:      "ACTIVE",
		StartTime:   &startTime,
		EndTime:     &endTime,
		CreatedTime: 1734395678,
		UpdatedTime: 1734395678,
	}

	// 转换为 Airbyte 格式
	airbyteData := TransformAdGroupToAirbyte(adGroup, 150219)

	// 解析 JSON 数据来验证
	var parsedAdGroup AdGroup
	err := json.Unmarshal(airbyteData.AirbyteData, &parsedAdGroup)
	if err != nil {
		t.Fatalf("解析转换后的JSON失败: %v", err)
	}

	// 验证 StartTime 和 EndTime 是否被转换为 null
	if parsedAdGroup.StartTime != nil {
		t.Errorf("期望 StartTime 为 null，但得到: %v", *parsedAdGroup.StartTime)
	} else {
		t.Log("✓ StartTime 成功转换为 null")
	}

	if parsedAdGroup.EndTime != nil {
		t.Errorf("期望 EndTime 为 null，但得到: %v", *parsedAdGroup.EndTime)
	} else {
		t.Log("✓ EndTime 成功转换为 null")
	}

	t.Logf("转换后的JSON: %s", string(airbyteData.AirbyteData))
}

func TestAdStructure(t *testing.T) {
	// 测试 Ad 结构体的 JSON 解析
	adJSON := `{
		"id": "123456789",
		"ad_account_id": "549756365874",
		"ad_group_id": "987654321",
		"campaign_id": "626754086067",
		"pin_id": "pin123",
		"name": "Test Ad",
		"status": "ACTIVE",
		"type": "regular",
		"creative_type": "REGULAR",
		"destination_url": "https://example.com",
		"created_time": 1734395678,
		"updated_time": 1734395678
	}`

	var ad Ad
	err := json.Unmarshal([]byte(adJSON), &ad)
	if err != nil {
		t.Fatalf("解析Ad JSON失败: %v", err)
	}

	// 验证关键字段
	if ad.Id != "123456789" {
		t.Errorf("期望 ID 为 '123456789'，但得到: %s", ad.Id)
	}

	if ad.AdAccountId != "549756365874" {
		t.Errorf("期望 AdAccountId 为 '549756365874'，但得到: %s", ad.AdAccountId)
	}

	if ad.CampaignId != "626754086067" {
		t.Errorf("期望 CampaignId 为 '626754086067'，但得到: %s", ad.CampaignId)
	}

	t.Logf("✓ Ad结构体解析成功: ID=%s, Name=%s, Status=%s", ad.Id, ad.Name, ad.Status)

	// 测试转换为 Airbyte 格式
	airbyteData := TransformAdToAirbyte(ad, 150219)

	// 验证转换结果
	if airbyteData.AirbyteRawId != ad.Id {
		t.Errorf("期望 AirbyteRawId 为 '%s'，但得到: %s", ad.Id, airbyteData.AirbyteRawId)
	}

	t.Logf("✓ Ad转换为Airbyte格式成功")
}
