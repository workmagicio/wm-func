package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"wm-func/common/apollo"
)

// Gemini API客户端
type GeminiClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// 创建Gemini客户端
func NewGeminiClient() *GeminiClient {
	// 使用安全的配置读取方法
	conf, err := apollo.GetInstance().GetLLMConfigSafe()
	if err != nil {
		// 降级处理：使用改进的Apollo客户端
		fmt.Printf("警告: 使用原Apollo客户端读取配置失败: %v\n", err)
		fmt.Println("尝试使用改进的Apollo客户端...")

		improvedConf, improvedErr := apollo.GetImprovedInstance().GetLLMConfigSafe()
		if improvedErr != nil {
			panic(fmt.Sprintf("Apollo配置读取完全失败: %v", improvedErr))
		}
		conf = improvedConf
	}

	fmt.Printf("Apollo Init Success: BaseUrl=%s, Key=***...\n", conf.BaseUrl)

	// 配置HTTP客户端，使用环境变量中的代理设置
	client := &http.Client{
		Timeout: 300 * time.Second, // 5分钟总超时
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // 从环境变量读取代理设置
		},
	}

	return &GeminiClient{
		baseURL: conf.BaseUrl,
		apiKey:  conf.Key,
		client:  client,
	}
}

// 分析字段映射
func (gc *GeminiClient) AnalyzeFieldMapping(req FieldMappingRequest) (*FieldMappingResponse, error) {
	// 1. 构建prompt
	prompt := gc.buildPrompt(req.Data)

	// 2. 创建API请求
	apiReq := gc.createAPIRequest(prompt)

	// 3. 调用API
	responseText, err := gc.callAPI(apiReq)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %v", err)
	}

	// 4. 解析响应
	var response FieldMappingResponse
	if err := json.Unmarshal([]byte(responseText), &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &response, nil
}

// AnalyzeFieldMappingAsMap 分析字段映射并返回map
func (gc *GeminiClient) AnalyzeFieldMappingAsMap(req FieldMappingRequest) (map[string]interface{}, error) {
	// 1. 构建prompt
	fmt.Println("🔧 Building AI analysis prompt...")
	prompt := gc.buildPrompt(req.Data)
	fmt.Printf("📝 Prompt built successfully (%d characters)\n", len(prompt))

	// 2. 创建API请求
	fmt.Println("📋 Creating API request structure...")
	apiReq := gc.createAPIRequest(prompt)

	// 3. 调用API
	fmt.Println("🌐 Calling Gemini AI API (this may take a few seconds)...")
	responseText, err := gc.callAPI(apiReq)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %v", err)
	}
	fmt.Printf("📥 API response received (%d characters)\n", len(responseText))

	// 4. 解析响应为map
	fmt.Println("🔍 Parsing AI response...")
	var responseMap map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &responseMap); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	fmt.Printf("✨ Successfully parsed %d field mappings\n", len(responseMap))

	return responseMap, nil
}

// 构建分析prompt
func (gc *GeminiClient) buildPrompt(data [][]string) string {
	var builder strings.Builder

	builder.WriteString(`你是一名数据分析师，需要分析Excel数据并找到字段对应的列标题。

需要找到以下字段对应的Excel列标题：

【必须映射到Excel列的字段】：
1. date_code: 日期代码列名 (🔴 重要：必须是Excel中的实际列名，绝对不能推断固定值或常量！)
2. geo_code: 地理标识符代码列名 (如邮编、州代码等 必须是Excel列名)
3. geo_name: 地理名称列名 (如城市名、州名等，必须是Excel列名)
4. sales_platform: 销售平台名称列名 (如店铺名称、平台名称等)
5. sales: 销售额列名
6. profit: 利润列名
7. orders: 订单数列名
8. new_customer_orders: 新客户订单数列名
9. new_customer_sales: 新客户销售额列名

【可以推断固定值的字段】：
10. date_type: 时间聚合级别 (DAILY/WEEKLY，可以根据数据推断)
11. geo_type: 地理粒度 (DMA/ZIP/STATE，可以根据数据推断)
12. sales_platform_type: 销售平台类型 (PRIMARY/SECONDARY，除了shopify都是SECONDARY,可以推断)
13. country_code: 国家代码 (可以推断，如US/CA等)

【系统配置字段】：
14. data_start_row: 数据开始的行号下标(需要你推断，不包含表头)
15. header_row: 表头所在的行号下标(需要你推断，表头包含列名)

特别说明：
- 请仔细分析数据，识别哪一行是表头(包含列名)，哪一行开始是真实数据
- header_row: 包含列标题的行号(从1开始计数)
- data_start_row: 实际数据开始的行号(从1开始计数，通常是header_row + 1)
- 对于【必须映射到Excel列的字段】：
  * 只返回Excel中的实际列名，不要返回推断的固定值
  * 这些字段代表具体的数据值，必须从Excel列中读取
  * 🔴 特别注意date_code：必须返回包含日期数据的Excel列标题，如"Calendar Walmart Week"、"Date"、"Week"等实际列名
  * 例如：geo_name应该映射到包含城市名称的列，而不是推断为"城市"这样的固定值
  
- 对于【可以推断固定值的字段】：
  * date_type: 通过分析日期数据的格式和间隔来判断是DAILY还是WEEKLY
  * geo_type: 通过分析地理代码的格式来判断是DMA、ZIP还是STATE  
  * sales_platform_type: 根据业务逻辑推断PRIMARY或SECONDARY
  * country_code: 根据数据特征推断国家代码
  * 这些字段代表数据的类型或分类，可以根据数据特征推断
  
- 如果是通过推断得出的值，请返回格式：推断值(inferred)，例如：DAILY(inferred)
- 如果无法找到对应的Excel列或无法推断，请返回空字符串

数据清理分析：
请分析数据中是否需要以下清理操作，并在响应中包含建议：
- date_format_issues: 日期格式是否需要转换为YYYY-MM-DD格式
- week_format_issues: 是否包含YYYYWW格式（如202502表示2025年第2周）需要转换为周开始日期
- month_format_issues: 是否包含YYYYMM格式（如202501表示2025年1月）需要转换为月开始日期
- currency_symbols: 销售额/利润字段是否包含货币符号需要清理
- number_formatting: 数字字段是否包含千位分隔符等需要清理
- suggested_operations: 建议自动应用的清理操作列表

注意：6位数字格式智能区分规则：
- YYYYWW周格式特征：
  * 后两位数字范围通常在01-53之间（一年最多53周）
  * 如果数据包含大于12的数字（如202513-202553），几乎确定是周格式
  * 周格式通常在业务报告中用于周度分析
- YYYYMM月格式特征：
  * 后两位数字范围严格在01-12之间（12个月）
  * 如果所有数据的后两位都≤12，且呈现月度规律，很可能是月格式
  * 月格式通常在财务报告中用于月度汇总
- 智能判断策略：
  * 如果发现后两位有>12的数字，直接判断为week_format_issues=true
  * 如果所有后两位都≤12，分析数据分布模式：
    - 如果数据呈现1-12的连续模式或月度间隔，判断为month_format_issues=true
    - 如果数据分布不规律或无法确定，两个都设为false让用户选择
  * 只有在95%确信时才设置为true，否则设为false

Excel数据样本：
`)

	for i, row := range data {
		builder.WriteString(fmt.Sprintf("第%d行: %s\n", i+1, strings.Join(row, " | ")))
	}

	builder.WriteString(`
请返回JSON格式结果。如果某个字段不存在，返回空字符串。
只有99%确信时才返回字段名，否则返回空字符串。`)

	return builder.String()
}

// 创建API请求
func (gc *GeminiClient) createAPIRequest(prompt string) *GeminiAPIRequest {
	return &GeminiAPIRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   createResponseSchema(),
		},
	}
}

// 调用API
func (gc *GeminiClient) callAPI(request *GeminiAPIRequest) (string, error) {
	// 构建URL
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
		gc.baseURL, MODEL_NAME, gc.apiKey)

	// 序列化请求
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	startTime := time.Now()
	resp, err := gc.client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ Request failed after %v\n", duration)
		safeErrorMsg := gc.sanitizeError(err.Error())
		return "", fmt.Errorf("发送HTTP请求失败: %v", safeErrorMsg)
	}

	fmt.Printf("✅ Got response after %v, status: %d\n", duration, resp.StatusCode)
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析API响应
	var apiResponse GeminiAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", fmt.Errorf("解析API响应失败: %v", err)
	}

	if len(apiResponse.Candidates) == 0 || len(apiResponse.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("API响应中没有内容")
	}

	return apiResponse.Candidates[0].Content.Parts[0].Text, nil
}

// 清理错误信息中的敏感信息
func (gc *GeminiClient) sanitizeError(errorMsg string) string {
	// 替换API key为***
	if gc.apiKey != "" {
		errorMsg = strings.ReplaceAll(errorMsg, gc.apiKey, "***")
	}

	// 替换可能包含key的URL参数
	re := regexp.MustCompile(`key=[^&\s]+`)
	errorMsg = re.ReplaceAllString(errorMsg, "key=***")

	return errorMsg
}
