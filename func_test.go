package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBase64(t *testing.T) {
	type TenantConfig struct {
		SfccClientID     string
		SfccClientSecret string
	}

	tenantConfig := TenantConfig{
		SfccClientID:     "eeefa2d0-44a2-451e-964a-f1dd58e90b56",
		SfccClientSecret: "Aas@X8XMJFQ2uEi",
	}

	// Python: auth_string = f"{tenant_config.sfcc_client_id}:{tenant_config.sfcc_client_secret}"
	// Go: 使用 fmt.Sprintf 格式化字符串
	authString := fmt.Sprintf("%s:%s", tenantConfig.SfccClientID, tenantConfig.SfccClientSecret)

	// Python:
	// auth_bytes = auth_string.encode('ascii')
	// auth_b64 = base64.b64encode(auth_bytes).decode('ascii')
	// Go: 使用 base64.StdEncoding.EncodeToString 一步完成
	// 它接收一个字节切片 []byte(authString)，并返回编码后的字符串
	authB64 := base64.StdEncoding.EncodeToString([]byte(authString))

	// 打印结果进行验证
	fmt.Println("原始字符串:", authString)
	fmt.Println("Base64 编码后:", authB64)
}

func TestRequestOrder(t *testing.T) {
	url := "https://store-api.workmagic.io/api/orders"
	concurrent := 100 // 并发数
	requests := 2000  // 总请求数

	fmt.Printf("开始压测订单接口: %s\n", url)
	fmt.Printf("并发数: %d, 总请求数: %d\n", concurrent, requests)

	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan OrderTestResult, requests)

	// 控制并发数的通道
	semaphore := make(chan struct{}, concurrent)

	begin := 56789
	for i := begin; i < begin+requests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			result := sendOrder(id)
			results <- result
		}(i)
	}

	wg.Wait()
	close(results)

	// 统计结果
	var success, failed int
	var totalTime time.Duration

	for result := range results {
		if result.Success {
			success++
		} else {
			failed++
			fmt.Printf("请求失败 #%d: %s\n", result.ID, result.Error)
		}
		totalTime += result.Duration
	}

	elapsed := time.Since(start)
	avgTime := totalTime / time.Duration(requests)
	qps := float64(requests) / elapsed.Seconds()

	fmt.Printf("\n=== 订单接口压测结果 ===\n")
	fmt.Printf("总耗时: %v\n", elapsed)
	fmt.Printf("成功请求: %d\n", success)
	fmt.Printf("失败请求: %d\n", failed)
	fmt.Printf("平均响应时间: %v\n", avgTime)
	fmt.Printf("QPS: %.2f\n", qps)
}

type OrderTestResult struct {
	ID       int
	Success  bool
	Duration time.Duration
	Error    string
}

func sendOrder(i int) OrderTestResult {
	start := time.Now()
	url := "https://store-api.workmagic.io/api/orders"

	newData := strings.ReplaceAll(order, "{{order_id}}", fmt.Sprintf("id_%d_%d", i, time.Now().Unix()))
	payload := bytes.NewReader([]byte(newData))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return OrderTestResult{
			ID:       i,
			Success:  false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-API-KEY", "mstak_tbizyrfjfrn5exymgavqgfimjx133622")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return OrderTestResult{
			ID:       i,
			Success:  false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return OrderTestResult{
			ID:       i,
			Success:  false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}

	// 检查响应状态
	if res.StatusCode >= 400 {
		return OrderTestResult{
			ID:       i,
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Sprintf("HTTP %d: %s", res.StatusCode, string(body)),
		}
	}

	return OrderTestResult{
		ID:       i,
		Success:  true,
		Duration: time.Since(start),
		Error:    "",
	}
}

// 压测 applovinLog 接口
func TestApplovinLogStress(t *testing.T) {
	url := "http://10.168.0.10/api/alter-data?platform=applovinLog"
	concurrent := 15  // 并发数
	requests := 20000 // 总请求数

	fmt.Printf("开始压测: %s\n", url)
	fmt.Printf("并发数: %d, 总请求数: %d\n", concurrent, requests)

	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan TestResult, requests)

	// 控制并发数的通道
	semaphore := make(chan struct{}, concurrent)

	for i := 10000; i < requests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			result := sendRequest(url, id)
			results <- result
		}(i)
	}

	wg.Wait()
	close(results)

	// 统计结果
	var success, failed int
	var totalTime time.Duration

	for result := range results {
		if result.Success {
			success++
		} else {
			failed++
			fmt.Printf("请求失败 #%d: %s\n", result.ID, result.Error)
		}
		totalTime += result.Duration
	}

	elapsed := time.Since(start)
	avgTime := totalTime / time.Duration(requests)
	qps := float64(requests) / elapsed.Seconds()

	fmt.Printf("\n=== 压测结果 ===\n")
	fmt.Printf("总耗时: %v\n", elapsed)
	fmt.Printf("成功请求: %d\n", success)
	fmt.Printf("失败请求: %d\n", failed)
	fmt.Printf("平均响应时间: %v\n", avgTime)
	fmt.Printf("QPS: %.2f\n", qps)
}

type TestResult struct {
	ID       int
	Success  bool
	Duration time.Duration
	Error    string
}

func sendRequest(url string, id int) TestResult {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		return TestResult{
			ID:       id,
			Success:  false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TestResult{
			ID:       id,
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}

	// 读取响应内容以确保完整请求
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			ID:       id,
			Success:  false,
			Duration: time.Since(start),
			Error:    err.Error(),
		}
	}

	return TestResult{
		ID:       id,
		Success:  true,
		Duration: time.Since(start),
		Error:    "",
	}
}

var order = `
{
"orders":[
    {
        "id": "{{order_id}}",
        "created_at": "2025-09-22T23:50:25.867Z",
        "updated_at": "2025-09-22T23:50:28Z",
        "browser_ip": "204.9.253.8",
        "order_number": "2025092216171311",
        "landing_site": "www.xukai.ink",
        "referring_site": "xukai@workmagic.io",
        "checkout_token": "2aaba7884ab47453c23374",
        "financial_status": "paid",
        "total_price": 214.96,
        "current_total_price": 214.96,
        "subtotal_price": 214.96,
        "current_subtotal_price": 214.96,
        "current_total_discounts": 20,
        "current_total_tax": 0,
        "discount_codes": [
            {
                "code": "WEB_ATC150D"
            }
        ],
        "shipping_address": {
            "zip": "80",
            "city": "us",
            "country": "United States",
            "country_code": "US",
            "province_code": "ac",
            "province": "good"
        },
        "line_items": [
            {
                "id": "c99d23b4a66da6",
                "title": "Pocket Bakinny ",
                "price": 59.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "09bae9b331b5",
                "title": "d Waist Straight Leg Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "2bf442dbd7a88",
                "title": "Ripped Double Button Mid Waist Skinny Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "cd2151055ed",
                "title": "Button Up Roll Hem Low Waist Straight Leg Jeans",
                "price": 56.99,
                "quantity": 1,
                "current_quantity": 1
            }
        ],
        "customer": {
            "id": "cce98fdf67c1",
            "email": "xukai@workmagic.io"
        },
        "extensions": {
            "order_app_name": "Web",
            "order_cost_cogs": 0,
            "order_cost_handling": 0,
            "order_cost_shipping": 0,
            "order_cost_transaction": 0,
            "order_refund_amount": 0,
            "order_refund_product_count": 0,
            "is_subscription_order": false
        }
    },
   {
        "id": "{{order_id}}",
        "created_at": "2025-09-22T23:50:25.867Z",
        "updated_at": "2025-09-22T23:50:28Z",
        "browser_ip": "204.9.253.8",
        "order_number": "2025092216171311",
        "landing_site": "www.xukai.ink",
        "referring_site": "xukai@workmagic.io",
        "checkout_token": "2aaba7884ab47453c23374",
        "financial_status": "paid",
        "total_price": 214.96,
        "current_total_price": 214.96,
        "subtotal_price": 214.96,
        "current_subtotal_price": 214.96,
        "current_total_discounts": 20,
        "current_total_tax": 0,
        "discount_codes": [
            {
                "code": "WEB_ATC150D"
            }
        ],
        "shipping_address": {
            "zip": "80",
            "city": "us",
            "country": "United States",
            "country_code": "US",
            "province_code": "ac",
            "province": "good"
        },
        "line_items": [
            {
                "id": "c99d23b4a66da6",
                "title": "Pocket Bakinny ",
                "price": 59.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "09bae9b331b5",
                "title": "d Waist Straight Leg Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "2bf442dbd7a88",
                "title": "Ripped Double Button Mid Waist Skinny Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "cd2151055ed",
                "title": "Button Up Roll Hem Low Waist Straight Leg Jeans",
                "price": 56.99,
                "quantity": 1,
                "current_quantity": 1
            }
        ],
        "customer": {
            "id": "cce98fdf67c1",
            "email": "xukai@workmagic.io"
        },
        "extensions": {
            "order_app_name": "Web",
            "order_cost_cogs": 0,
            "order_cost_handling": 0,
            "order_cost_shipping": 0,
            "order_cost_transaction": 0,
            "order_refund_amount": 0,
            "order_refund_product_count": 0,
            "is_subscription_order": false
        }
    },   {
        "id": "{{order_id}}",
        "created_at": "2025-09-22T23:50:25.867Z",
        "updated_at": "2025-09-22T23:50:28Z",
        "browser_ip": "204.9.253.8",
        "order_number": "2025092216171311",
        "landing_site": "www.xukai.ink",
        "referring_site": "xukai@workmagic.io",
        "checkout_token": "2aaba7884ab47453c23374",
        "financial_status": "paid",
        "total_price": 214.96,
        "current_total_price": 214.96,
        "subtotal_price": 214.96,
        "current_subtotal_price": 214.96,
        "current_total_discounts": 20,
        "current_total_tax": 0,
        "discount_codes": [
            {
                "code": "WEB_ATC150D"
            }
        ],
        "shipping_address": {
            "zip": "80",
            "city": "us",
            "country": "United States",
            "country_code": "US",
            "province_code": "ac",
            "province": "good"
        },
        "line_items": [
            {
                "id": "c99d23b4a66da6",
                "title": "Pocket Bakinny ",
                "price": 59.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "09bae9b331b5",
                "title": "d Waist Straight Leg Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "2bf442dbd7a88",
                "title": "Ripped Double Button Mid Waist Skinny Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "cd2151055ed",
                "title": "Button Up Roll Hem Low Waist Straight Leg Jeans",
                "price": 56.99,
                "quantity": 1,
                "current_quantity": 1
            }
        ],
        "customer": {
            "id": "cce98fdf67c1",
            "email": "xukai@workmagic.io"
        },
        "extensions": {
            "order_app_name": "Web",
            "order_cost_cogs": 0,
            "order_cost_handling": 0,
            "order_cost_shipping": 0,
            "order_cost_transaction": 0,
            "order_refund_amount": 0,
            "order_refund_product_count": 0,
            "is_subscription_order": false
        }
    },   {
        "id": "{{order_id}}",
        "created_at": "2025-09-22T23:50:25.867Z",
        "updated_at": "2025-09-22T23:50:28Z",
        "browser_ip": "204.9.253.8",
        "order_number": "2025092216171311",
        "landing_site": "www.xukai.ink",
        "referring_site": "xukai@workmagic.io",
        "checkout_token": "2aaba7884ab47453c23374",
        "financial_status": "paid",
        "total_price": 214.96,
        "current_total_price": 214.96,
        "subtotal_price": 214.96,
        "current_subtotal_price": 214.96,
        "current_total_discounts": 20,
        "current_total_tax": 0,
        "discount_codes": [
            {
                "code": "WEB_ATC150D"
            }
        ],
        "shipping_address": {
            "zip": "80",
            "city": "us",
            "country": "United States",
            "country_code": "US",
            "province_code": "ac",
            "province": "good"
        },
        "line_items": [
            {
                "id": "c99d23b4a66da6",
                "title": "Pocket Bakinny ",
                "price": 59.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "09bae9b331b5",
                "title": "d Waist Straight Leg Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "2bf442dbd7a88",
                "title": "Ripped Double Button Mid Waist Skinny Jeans",
                "price": 58.99,
                "quantity": 1,
                "current_quantity": 1
            },
            {
                "id": "cd2151055ed",
                "title": "Button Up Roll Hem Low Waist Straight Leg Jeans",
                "price": 56.99,
                "quantity": 1,
                "current_quantity": 1
            }
        ],
        "customer": {
            "id": "cce98fdf67c1",
            "email": "xukai@workmagic.io"
        },
        "extensions": {
            "order_app_name": "Web",
            "order_cost_cogs": 0,
            "order_cost_handling": 0,
            "order_cost_shipping": 0,
            "order_cost_transaction": 0,
            "order_refund_amount": 0,
            "order_refund_product_count": 0,
            "is_subscription_order": false
        }
    }
]
}
`
