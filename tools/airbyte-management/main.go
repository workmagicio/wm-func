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

type StatusRequest struct {
	ConnectionId string `json:"connectionId"`
	Status       string `json:"status"`
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

func UpdateConnectionStatus(connectionId string, status string) {
	url := "http://internal-airbytes.workmagic.io/api/v1/web_backend/connections/update"

	data := StatusRequest{
		ConnectionId: connectionId,
		Status:       status,
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
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Origin", "http://internal-airbytes.workmagic.io")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Referer", "http://internal-airbytes.workmagic.io/workspaces/1bcb346c-c895-480a-9bcf-b77ecd57e4ff/connections?search=gmv")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-airbyte-analytic-source", "webapp")
	req.Header.Set("x-api-key", "C7986gw4eZp8sr34QuBd")

	// 设置 Cookie
	req.Header.Set("Cookie", "intercom-id-oe2kqapl=a4e37469-24f4-4f73-9e18-dc179a8a60f0; intercom-device-id-oe2kqapl=a5b4fc31-1c2b-48de-ae9c-7413b2e3bfbf; wm_client_id=wm.6ejfa4madub.1719805732; _ga=GA1.1.wm.6ejfa4madub.1719805732; _ga_HDBMVFQGBH=GS1.1.1720529831.2.0.1720529831.0.0.0; hubspotutk=ce588d9f8ef9f115719855aaa17d1644; _gcl_au=1.1.1966978318.1747636076.1120573105.1748275619.1748275621; __hstc=266511057.ce588d9f8ef9f115719855aaa17d1644.1747636080850.1748351580323.1748936291268.13; _ga_QXWRYC1ZJY=GS2.1.s1750404697$o1005$g0$t1750404697$j60$l0$h0; ajs_user_id=1bcb346c-c895-480a-9bcf-b77ecd57e4ff; ajs_anonymous_id=1281cbb5-d836-4234-a5cc-24d1f005c760")

	// 使用 HTTP 客户端发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 处理响应
	log.Printf("Response Status: %s", resp.Status)
}
