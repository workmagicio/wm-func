package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Request struct {
	ConnectionId string       `json:"connectionId"`
	ScheduleType string       `json:"scheduleType"`
	ScheduleData ScheduleData `json:"scheduleData"`
}

type RequestName struct {
	ConnectionId string `json:"connectionId"`
	Name         string `json:"name"`
}

type Cron struct {
	CronTimeZone   string `json:"cronTimeZone"`
	CronExpression string `json:"cronExpression"`
}
type ScheduleData struct {
	Cron Cron `json:"cron"`
}

func UpdateSetting(connectionId string, cronExpression string) {
	url := "http://internal-airbytes.workmagic.io/api/v1/web_backend/connections/update"

	data := Request{
		ConnectionId: connectionId,
		ScheduleType: "cron",
		ScheduleData: ScheduleData{
			Cron: Cron{
				CronTimeZone:   "UTC",
				CronExpression: cronExpression,
				//CronExpression: "0 1 * * * ?",
			},
		},
	}
	// 将数据编码为 JSON 格式
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	// 创建新的 HTTP 请求
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Basic d29ya21hZ2ljOjh1YW5xbk1SUkhMM1BZVkZWbWJY")
	req.Header.Set("Content-Type", "application/json")

	// 使用 HTTP 客户端发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 处理响应
	// 你可以在这里根据需要读取响应内容
	log.Printf("Response Status: %s", resp.Status)
}
