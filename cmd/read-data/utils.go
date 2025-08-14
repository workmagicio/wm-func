package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 获取字段的建议值
func getFieldSuggestions(fieldName string) []string {
	suggestions := map[string][]string{
		"date_type":           {"DAILY", "WEEKLY", "MONTHLY"},
		"geo_type":            {"DMA", "ZIP", "STATE", "COUNTY"},
		"sales_platform_type": {"PRIMARY", "SECONDARY"},
		"country_code":        {"US", "CA", "UK", "AU"},
	}

	return suggestions[fieldName]
}

// 安全获取map中的值
func getMapValue(m map[string]interface{}, key string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// 获取字段的特殊说明
func getFieldSpecialNote(field string) string {
	specialNotes := map[string]string{
		"data_start_row": " [Data Configuration]",
	}

	if note, exists := specialNotes[field]; exists {
		return note
	}
	return ""
}

// 获取字段的友好显示名称
func getFieldDisplayName(field string) string {
	displayNames := map[string]string{
		"date_type":           "Date Type",
		"date_code":           "Date Code",
		"geo_type":            "Geo Type",
		"geo_code":            "Geo Code",
		"geo_name":            "Geo Name",
		"sales_platform":      "Sales Platform",
		"sales_platform_type": "Platform Type",
		"country_code":        "Country Code",
		"orders":              "Orders",
		"sales":               "Sales",
		"profit":              "Profit",
		"new_customer_orders": "New Customer Orders",
		"new_customer_sales":  "New Customer Sales",
		"data_start_row":      "Data Start Row",
	}

	if displayName, exists := displayNames[field]; exists {
		return displayName
	}
	return field
}

// 获取数据起始行
func getDataStartRow(responseMap map[string]interface{}) int {
	if value, exists := responseMap["data_start_row"]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if num, err := strconv.Atoi(v); err == nil {
				return num
			}
		}
	}
	return 2 // 默认值
}

// 查找列索引
func findColumnIndex(headers []string, columnName string) int {
	for i, header := range headers {
		if header == columnName {
			return i
		}
	}
	return -1
}

// 清理推断标记，用于预览显示
func cleanInferredValue(value string) string {
	// 去掉 (inferred) 标记
	if strings.Contains(value, "(inferred)") {
		return strings.TrimSpace(strings.Replace(value, "(inferred)", "", -1))
	}
	return value
}

// 判断是否为固定值（非Excel列映射）
func isFixedValue(value string, data [][]string, responseMap map[string]interface{}) bool {
	if len(data) == 0 {
		return true // 如果没有数据，假设是固定值
	}

	// 清理推断标记
	cleanValue := cleanInferredValue(value)

	// 获取正确的表头行
	headerRowIndex := getHeaderRowIndex(responseMap)
	if headerRowIndex >= len(data) {
		headerRowIndex = 0
	}

	headers := data[headerRowIndex]

	// 精确匹配和模糊匹配
	for _, header := range headers {
		// 精确匹配
		if header == cleanValue {
			return false // 在表头中找到，说明是列映射
		}

		// 忽略大小写和空格的匹配
		if strings.EqualFold(strings.TrimSpace(header), strings.TrimSpace(cleanValue)) {
			return false // 在表头中找到，说明是列映射
		}
	}

	return true // 不在表头中，说明是固定值
}

// 特殊处理：检查date_code字段是否被正确映射为列名
func validateDateCodeMapping(responseMap map[string]interface{}, data [][]string) bool {
	dateCodeValue := getMapValue(responseMap, "date_code")
	if dateCodeValue == "" {
		return false // 未映射
	}

	cleanValue := cleanInferredValue(dateCodeValue)

	// 如果date_code被标记为固定值，这通常是错误的
	if isFixedValue(cleanValue, data, responseMap) {
		fmt.Printf("⚠️  Warning: date_code '%s' appears to be a fixed value, but should be an Excel column name\n", cleanValue)
		return false
	}

	return true
}

// 获取表头行索引（从AI响应中）
func getHeaderRowIndex(responseMap map[string]interface{}) int {
	if headerRow, exists := responseMap["header_row"]; exists {
		switch v := headerRow.(type) {
		case int:
			return v - 1 // 转换为0-based索引
		case float64:
			return int(v) - 1 // 转换为0-based索引
		case string:
			if num, err := strconv.Atoi(v); err == nil {
				return num - 1 // 转换为0-based索引
			}
		}
	}
	return 0 // 默认使用第一行
}

// 检测是否为product level数据
func detectProductLevel(data [][]string, responseMap map[string]interface{}) *ProductLevelDetection {
	detection := &ProductLevelDetection{
		IsProductLevel:  false,
		OrderIDFields:   []string{},
		ProductIDFields: []string{},
		SKUFields:       []string{},
		ConfidenceScore: 0.0,
	}

	if len(data) < 2 {
		return detection
	}

	// 获取表头行
	headerRowIndex := getHeaderRowIndex(responseMap)
	if headerRowIndex >= len(data) {
		headerRowIndex = 0
	}
	headers := data[headerRowIndex]

	// 检测相关字段
	orderIDPatterns := []string{
		"order.?id", "order.?number", "order.?ref", "transaction.?id",
		"receipt.?id", "purchase.?id", "invoice.?id",
	}

	productIDPatterns := []string{
		"product.?id", "item.?id", "asin", "upc", "ean", "gtin",
		"product.?code", "item.?code", "catalog.?id",
	}

	skuPatterns := []string{
		"sku", "stock.?keeping.?unit", "variant.?id", "model.?number",
		"part.?number", "item.?number",
	}

	// 检查表头中的字段
	for _, header := range headers {
		headerLower := strings.ToLower(header)

		// 检查order ID字段
		for _, pattern := range orderIDPatterns {
			if matched, _ := regexp.MatchString(pattern, headerLower); matched {
				detection.OrderIDFields = append(detection.OrderIDFields, header)
				detection.ConfidenceScore += 0.3
			}
		}

		// 检查product ID字段
		for _, pattern := range productIDPatterns {
			if matched, _ := regexp.MatchString(pattern, headerLower); matched {
				detection.ProductIDFields = append(detection.ProductIDFields, header)
				detection.ConfidenceScore += 0.25
			}
		}

		// 检查SKU字段
		for _, pattern := range skuPatterns {
			if matched, _ := regexp.MatchString(pattern, headerLower); matched {
				detection.SKUFields = append(detection.SKUFields, header)
				detection.ConfidenceScore += 0.2
			}
		}
	}

	// 检查数据行中是否有重复的order ID（product level的特征）
	if len(detection.OrderIDFields) > 0 {
		orderIDIndex := findColumnIndex(headers, detection.OrderIDFields[0])
		if orderIDIndex >= 0 {
			orderIDs := make(map[string]int)
			dataStartRow := getDataStartRow(responseMap)

			for i := dataStartRow - 1; i < len(data) && i < dataStartRow+10; i++ {
				if i >= 0 && orderIDIndex < len(data[i]) {
					orderID := strings.TrimSpace(data[i][orderIDIndex])
					if orderID != "" {
						orderIDs[orderID]++
					}
				}
			}

			// 如果有重复的order ID，很可能是product level数据
			duplicateCount := 0
			for _, count := range orderIDs {
				if count > 1 {
					duplicateCount++
				}
			}

			if duplicateCount > 0 {
				detection.ConfidenceScore += 0.4
			}
		}
	}

	// 判断是否为product level数据
	detection.IsProductLevel = detection.ConfidenceScore >= 0.5 &&
		(len(detection.OrderIDFields) > 0 || len(detection.ProductIDFields) > 0 || len(detection.SKUFields) > 0)

	return detection
}

// 辅助函数
func getEnabledStatus(enabled bool) string {
	if enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 根据列索引获取字段名
func getFieldNameByColumnIndex(colIndex int, headers []string, responseMap map[string]interface{}) string {
	if colIndex >= len(headers) {
		return ""
	}

	columnName := headers[colIndex]

	// 遍历所有字段映射，找到对应的字段名
	for _, fieldName := range allMappableFields {
		if fieldName == "data_start_row" {
			continue
		}

		mappedValue := getMapValue(responseMap, fieldName)
		cleanMappedValue := cleanInferredValue(mappedValue)

		if cleanMappedValue == columnName {
			return fieldName
		}
	}

	return ""
}

// 获取字段对应的列索引
func getFieldColumnIndex(fieldName string, headers []string, responseMap map[string]interface{}) int {
	mappedValue := getMapValue(responseMap, fieldName)
	cleanMappedValue := cleanInferredValue(mappedValue)
	return findColumnIndex(headers, cleanMappedValue)
}

// 判断值是否为列映射（而不是固定值）
func isColumnMapping(value string, headers []string) bool {
	for _, header := range headers {
		if header == value {
			return true
		}
	}
	return false
}

// 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// 日期格式分析结果
type DateFormatAnalysis struct {
	TotalValues         int
	SixDigitCount       int
	MinValue            string
	MaxValue            string
	MinLastTwo          int
	MaxLastTwo          int
	LastTwoDistribution map[int]int
	SuggestedFormat     string // "WEEK", "MONTH", "UNKNOWN"
}

// 应用字段处理操作
func applyFieldProcessing(value, fieldName string) string {
	processor, exists := fieldProcessors[fieldName]
	if !exists || len(processor.Operations) == 0 {
		return value
	}

	result := value
	for _, operation := range processor.Operations {
		// 特殊处理日期转换
		if operation.Name == "convert_date_format" {
			if converted := convertDateFormat(result); converted != "" {
				result = converted
				continue
			}
		}

		// 特殊处理周格式转换
		if operation.Name == "convert_week_format" {
			if converted := convertWeekFormat(result); converted != result {
				result = converted
				continue
			}
		}

		// 特殊处理月份格式转换
		if operation.Name == "convert_month_format" {
			if converted := convertMonthFormat(result); converted != result {
				result = converted
				continue
			}
		}

		// 特殊处理geo_code州代码大写转换
		if operation.Name == "uppercase_state_code" {
			if converted := uppercaseStateCode(result); converted != result {
				result = converted
				continue
			}
		}

		// 普通正则表达式处理
		if operation.Pattern != "" {
			re, err := regexp.Compile(operation.Pattern)
			if err != nil {
				fmt.Printf("⚠️ Invalid regex in operation '%s': %v\n", operation.Description, err)
				continue
			}
			result = re.ReplaceAllString(result, operation.Replacement)
		}
	}

	return strings.TrimSpace(result)
}

// 转换YYYYWW格式为当周第一天
func convertWeekFormat(weekStr string) string {
	weekStr = strings.TrimSpace(weekStr)
	if len(weekStr) != 6 {
		return weekStr // 不是YYYYWW格式，返回原值
	}

	// 解析年份和周数
	yearStr := weekStr[:4]
	weekStr_num := weekStr[4:]

	year, err1 := strconv.Atoi(yearStr)
	week, err2 := strconv.Atoi(weekStr_num)

	if err1 != nil || err2 != nil || week < 1 || week > 53 {
		return weekStr // 解析失败，返回原值
	}

	// 计算该年第一周的开始日期
	jan1 := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// 找到第一个周一（ISO 8601标准，周一为一周开始）
	daysToMonday := (8 - int(jan1.Weekday())) % 7
	if jan1.Weekday() == time.Sunday {
		daysToMonday = 1
	} else if jan1.Weekday() == time.Monday {
		daysToMonday = 0
	}

	firstMonday := jan1.AddDate(0, 0, daysToMonday)

	// 如果1月1日在周四之后，第一周从下一个周一开始
	if jan1.Weekday() > time.Thursday {
		firstMonday = firstMonday.AddDate(0, 0, 7)
		week = week - 1 // 调整周数
	}

	// 计算目标周的开始日期
	targetDate := firstMonday.AddDate(0, 0, (week-1)*7)

	return targetDate.Format("2006-01-02")
}

// 转换YYYYMM格式为月份第一天
func convertMonthFormat(monthStr string) string {
	monthStr = strings.TrimSpace(monthStr)
	if len(monthStr) != 6 {
		return monthStr // 不是YYYYMM格式，返回原值
	}

	// 解析年份和月份
	yearStr := monthStr[:4]
	monthStr_num := monthStr[4:]

	year, err1 := strconv.Atoi(yearStr)
	month, err2 := strconv.Atoi(monthStr_num)

	if err1 != nil || err2 != nil || month < 1 || month > 12 {
		return monthStr // 解析失败，返回原值
	}

	// 创建该月第一天的日期
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	return firstDay.Format("2006-01-02")
}

// 转换各种日期格式为 YYYY-MM-DD
func convertDateFormat(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return ""
	}

	// 如果已经是标准格式，直接返回
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); matched {
		return dateStr
	}

	// 常见的日期格式模式（按优先级排序）
	dateFormats := []string{
		// 标准ISO格式
		"2006-01-02", // 2025-01-01 (目标格式)
		"2006/01/02", // 2025/01/01
		"2006.01.02", // 2025.01.01
		"20060102",   // 20250101

		// 美式格式
		"01/02/2006", // 01/01/2025 (MM/DD/YYYY)
		"1/2/2006",   // 1/1/2025 (M/D/YYYY)
		"01-02-2006", // 01-01-2025 (MM-DD-YYYY)
		"1-2-2006",   // 1-1-2025 (M-D-YYYY)

		// 欧洲格式
		"02/01/2006", // 01/01/2025 (DD/MM/YYYY)
		"2/1/2006",   // 1/1/2025 (D/M/YYYY)
		"02-01-2006", // 01-01-2025 (DD-MM-YYYY)
		"2-1-2006",   // 1-1-2025 (D-M-YYYY)

		// 带时间的格式
		"Jan 2, 2006 3:04:05 PM MST", // Jan 1, 2025 12:06:27 AM PST
		"Jan 2, 2006 15:04:05 MST",   // Jan 1, 2025 00:06:27 PST
		"Jan 2, 2006",                // Jan 1, 2025
		"January 2, 2006",            // January 1, 2025

		// 短年份格式
		"06-01-02", // 25-01-01
		"06/01/02", // 25/01/01
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// 确保返回完整的YYYY-MM-DD格式
			return t.Format("2006-01-02")
		}
	}

	// 如果都解析失败，尝试提取数字并智能组合
	return extractAndFormatDate(dateStr)
}

// 从字符串中提取数字并尝试组合成日期
func extractAndFormatDate(dateStr string) string {
	// 提取所有数字
	re := regexp.MustCompile(`\d+`)
	numbers := re.FindAllString(dateStr, -1)

	if len(numbers) < 3 {
		return dateStr // 无法解析，返回原值
	}

	// 尝试不同的组合方式
	var year, month, day int
	var err error

	// 解析数字
	nums := make([]int, len(numbers))
	for i, numStr := range numbers {
		if nums[i], err = strconv.Atoi(numStr); err != nil {
			return dateStr // 解析失败，返回原值
		}
	}

	// 智能判断年月日
	for _, num := range nums {
		if num > 31 { // 可能是年份
			if num < 100 { // 两位年份
				if num > 50 {
					year = 1900 + num
				} else {
					year = 2000 + num
				}
			} else {
				year = num
			}
			break
		}
	}

	// 如果没找到年份，使用当前年份
	if year == 0 {
		year = time.Now().Year()
	}

	// 找月份和日期
	remaining := []int{}
	for _, num := range nums {
		if num != year && num != year-1900 && num != year-2000 {
			remaining = append(remaining, num)
		}
	}

	if len(remaining) >= 2 {
		// 假设第一个是月份，第二个是日期
		month = remaining[0]
		day = remaining[1]

		// 如果月份大于12，交换月日
		if month > 12 && day <= 12 {
			month, day = day, month
		}

		// 验证月份和日期的有效性
		if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
			return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		}
	}

	return dateStr // 无法解析，返回原值
}

// 转换州代码为大写格式
func uppercaseStateCode(geoCode string) string {
	// 去除前后空格
	geoCode = strings.TrimSpace(geoCode)

	// 检查是否为2位字母的州代码
	if len(geoCode) == 2 {
		// 检查是否全为字母
		for _, char := range geoCode {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				return geoCode // 不是纯字母，返回原值
			}
		}
		return strings.ToUpper(geoCode)
	}

	return geoCode // 不是2位字符，返回原值
}
