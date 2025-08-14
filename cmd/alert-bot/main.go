package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"wm-func/cmd/alert-bot/exec_sql"

	"google.golang.org/genai"
)

func main() {
	Run2()
	return
	sqls := GetAllDataQualityChecks()

	for _, sql := range sqls {
		tenants := GetCheckTenantIds()
		b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/alert-bot/promot.txt")
		if err != nil {
			panic(err)
		}

		if strings.Contains(sql.DataSQL, "{{tenant_ids}}") {
			sql.DataSQL = strings.ReplaceAll(sql.DataSQL, "{{tenant_ids}}", strings.Join(tenants, ","))
			sql.CheckSQL = strings.ReplaceAll(sql.CheckSQL, "{{tenant_ids}}", strings.Join(tenants, ","))
		}

		prompt := string(b)
		prompt = strings.ReplaceAll(prompt, "{{MAIN_DATA_JSON}}", exec_sql.Exec(sql.DataSQL))
		prompt = strings.ReplaceAll(prompt, "{{CHECK_DATA_JSON}}", exec_sql.Exec(sql.CheckSQL))
		prompt = strings.ReplaceAll(prompt, "{{CUSTOMER_RULE}}", sql.AIPrompt)
		generateWithText(prompt)
	}

}

func Run2() {
	prompt := "你是一名数据清洗工程师，我现在有一些excel交给你，你需要从我提供的数据中找到我需要找到下面几个字段" +
		"1. date_type： The time aggregation level of the data record. Priority: High. Usage: Used to determine grouping level. 你需要返回给我 DAILY 或者 WEEKLY" +
		"2. date_code: type, use the start date of the week (e.g., Monday). The system will evenly distribute the aggregated weekly data across each day of that week. Priority: High. Usage: Used to filter or join time-based metrics. pattern:\n\\d\\d\\d\\d-\\d\\d-\\d\\d" +
		"3. geo_type: The geographic granularity of the data. Priority: High. Usage: Determines spatial segmentation in lift test. 你需要返回给我 DMA 或者 ZIP 或者 STATE" +
		"4. geo_code： The geographic identifier code, matched to the geo_type." +
		"5: sales_platform： The name of the platform where the sale occurred (e.g., Shopify, Amazon). If not platform-specific, use \"Online Store\" for PRIMARY sales_platform_type or \"Retail\" for SECONDARY. Priority: High. Usage: Segmentation of sales data by platform." +
		"6: sales_platform_type：Specifies whether the sales occurred on a direct-to-consumer (DTC) or non-DTC platform. Use \"PRIMARY\" for DTC channels where marketing directly drives purchases. Use \"SECONDARY\" for indirect channels, such as marketplaces, where marketing may influence sales through halo effects. Priority: High. Usage: Controls the attribution logic used in lift tests—PRIMARY enables direct incremental lift analysis, while SECONDARY supports halo effect evaluation. 你可以返回给我 PRIMARY 或者 SECONDARY" +
		"7. country_code： The country code where sales occurred in ISO 3166-1 alpha-2 format. Priority: High. Usage: Used to identify the country where the lift test should be conducted." +
		"8. orders: The number of orders placed. Priority: High. Usage: Used for calculating incremental orders in lift test, and predict orders in MMM. type:\nnumber" +
		"9. sales: The total value of sales (in your store currency). Priority: High. Usage: Used for calculating incremental sales in lift test, and predict sales in MMM. type:\nnumber" +
		"10. profit: The profit earned after subtracting costs (in your store currency). Priority: Medium. Usage: Profit prediction in MMM." +
		"11. new_customer_orders: The number of orders placed by new customers. Priority: Medium. Usage: Used for calculating incremental new customer orders in lift test, and predict new customer orders in MMM.\n\n" +
		"12. new_customer_sales: The total value of sales placed by new customers (in your store currency). Priority: Medium. Usage: Used for calculating incremental new customer sales in lift test, and predict new customer sales in MMM.\n\n" +
		"" +
		"当你有99%的概率确认时才返回，否则返回一个空，我还需要你找到数据的起始行是多少"

	// 使用 ReadFileData 函数，然后取前10条数据拼到 prompt 上
	// 可以只读excel 的前10条记录，避免读整个文件

	// 这里需要指定要读取的文件路径，可以从命令行参数或配置中获取
	filename := "/Users/xukai/Downloads/cs-data/TGT Sales Data - Zip Code.xlsx" // 请替换为实际文件路径
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	// 读取文件数据，只读取前11行（表头 + 10条数据）
	maxRows := 11 // 表头 + 10条数据
	data, err := ReadFileDataWithLimit(filename, maxRows)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	// 将读取的数据格式化为字符串拼接到prompt
	var dataStr strings.Builder
	dataStr.WriteString("\n\n以下是Excel文件的前10条数据样本：\n")
	for i := 0; i < len(data); i++ {
		if i == 0 {
			dataStr.WriteString("表头: ")
		} else {
			dataStr.WriteString(fmt.Sprintf("第%d行: ", i))
		}
		dataStr.WriteString(strings.Join(data[i], " | "))
		dataStr.WriteString("\n")
	}

	// 将数据拼接到原有prompt
	prompt += dataStr.String()

	generateWithText(prompt)
}

// generateWithText shows how to generate text using a text prompt.
func generateWithText(prompt string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      "AIzaSyBIWCe93sSdEFu8w_X6MDgLYNQz6W8fdms",
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		panic(err)
		//return fmt.Errorf("failed to create genai client: %w", err)
	}

	resp, err := client.Models.GenerateContent(ctx,
		"gemini-2.5-pro",
		genai.Text(prompt),
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		panic(err)
	}

	respText := resp.Text()

	fmt.Println(respText)

	return nil
}
