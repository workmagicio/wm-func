package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 字段映射服务
type FieldMappingService struct {
	geminiClient *GeminiClient
	fileReader   *FileReader
}

// 创建字段映射服务
func NewFieldMappingService() *FieldMappingService {
	return &FieldMappingService{
		geminiClient: NewGeminiClient(),
		fileReader:   NewFileReader(),
	}
}

// 执行字段映射分析
func (fms *FieldMappingService) Analyze(filename string) (*FieldMappingResponse, error) {
	// 1. 读取文件数据
	data, err := fms.fileReader.ReadData(filename, MAX_ROWS)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 2. 创建请求
	request := FieldMappingRequest{
		Filename: filename,
		Data:     data,
	}

	// 3. 分析字段映射
	responseMap, err := fms.geminiClient.AnalyzeFieldMappingAsMap(request)
	if err != nil {
		return nil, fmt.Errorf("分析字段映射失败: %v", err)
	}

	// 4. 验证必填字段
	if err := fms.validateRequiredFieldsFromMap(responseMap); err != nil {
		return nil, fmt.Errorf("必填字段验证失败: %v", err)
	}

	// 5. 转换为结构体返回
	response := fms.mapToResponse(responseMap)
	return response, nil
}

// validateRequiredFieldsFromMap 从map验证必填字段
func (fms *FieldMappingService) validateRequiredFieldsFromMap(responseMap map[string]interface{}) error {
	var missingFields []string

	// 遍历必填字段
	for fieldName := range requered_fields_map {
		value, exists := responseMap[fieldName]
		if !exists {
			missingFields = append(missingFields, fieldName)
			continue
		}

		// 检查值是否为空
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			missingFields = append(missingFields, fieldName)
		}
	}

	// 如果有缺失的必填字段，返回错误
	if len(missingFields) > 0 {
		return fmt.Errorf("以下必填字段缺失或为空: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// GetRequiredFieldsStatusFromMap 从map获取必填字段状态
func (fms *FieldMappingService) GetRequiredFieldsStatusFromMap(responseMap map[string]interface{}) map[string]string {
	status := make(map[string]string)

	for fieldName := range requered_fields_map {
		value, exists := responseMap[fieldName]
		if !exists {
			status[fieldName] = "❌ 缺失"
			continue
		}

		if str, ok := value.(string); ok {
			if strings.TrimSpace(str) == "" {
				status[fieldName] = "❌ 缺失"
			} else {
				status[fieldName] = fmt.Sprintf("✅ %s", str)
			}
		} else {
			status[fieldName] = fmt.Sprintf("✅ %v", value)
		}
	}

	return status
}

// mapToResponse 将map转换为结构体
func (fms *FieldMappingService) mapToResponse(responseMap map[string]interface{}) *FieldMappingResponse {
	response := &FieldMappingResponse{}

	// 辅助函数：安全获取字符串值
	getString := func(key string) string {
		if value, exists := responseMap[key]; exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
		return ""
	}

	// 辅助函数：安全获取整数值
	getInt := func(key string) int {
		if value, exists := responseMap[key]; exists {
			if num, ok := value.(float64); ok {
				return int(num)
			}
			if num, ok := value.(int); ok {
				return num
			}
		}
		return 0
	}

	// 填充结构体
	response.DateType = getString("date_type")
	response.DateCode = getString("date_code")
	response.GeoType = getString("geo_type")
	response.GeoCode = getString("geo_code")
	response.GeoName = getString("geo_name")
	response.SalesPlatform = getString("sales_platform")
	response.SalesPlatformType = getString("sales_platform_type")
	response.CountryCode = getString("country_code")
	response.Orders = getString("orders")
	response.Sales = getString("sales")
	response.Profit = getString("profit")
	response.NewCustomerOrders = getString("new_customer_orders")
	response.NewCustomerSales = getString("new_customer_sales")
	response.DataStartRow = getInt("data_start_row")
	response.HeaderRow = getInt("header_row")

	return response
}

// AnalyzeAsMap 执行字段映射分析并返回map（用于调试）
func (fms *FieldMappingService) AnalyzeAsMap(filename string) (map[string]interface{}, error) {
	// 1. 读取文件数据
	fmt.Printf("📖 Reading first %d rows for AI analysis...\n", MAX_ROWS)
	data, err := fms.fileReader.ReadData(filename, MAX_ROWS)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	fmt.Printf("📊 Successfully read %d rows, %d columns\n", len(data), len(data[0]))

	// 2. 创建请求
	fmt.Println("🔄 Preparing AI analysis request...")
	request := FieldMappingRequest{
		Filename: filename,
		Data:     data,
	}

	// 3. 分析字段映射
	fmt.Println("🚀 Sending request to AI service...")
	responseMap, err := fms.geminiClient.AnalyzeFieldMappingAsMap(request)
	if err != nil {
		return nil, fmt.Errorf("分析字段映射失败: %v", err)
	}
	fmt.Println("🎯 AI analysis response received!")

	// 4. 调整AI推断的数据起始行
	if dataStartRow, exists := responseMap["data_start_row"]; exists {
		var currentRow int
		switch v := dataStartRow.(type) {
		case int:
			currentRow = v
		case float64:
			currentRow = int(v)
		case string:
			if num, err := strconv.Atoi(v); err == nil {
				currentRow = num
			}
		}

		// AI通常将表头识别为数据起始行，需要+1或+2来跳过表头和示例行
		// 根据数据特征智能调整
		if currentRow == 1 && len(data) > 2 {
			// 检查第2行是否看起来像示例行（包含非数字的描述性内容）
			if len(data) > 1 && fms.isExampleRow(data[1]) {
				// 如果第2行是示例行，数据从第3行开始
				responseMap["data_start_row"] = 3
				fmt.Println("🔧 Adjusted data_start_row from 1 to 3 (skipping header and example row)")
			} else {
				// 如果第2行就是数据，数据从第2行开始
				responseMap["data_start_row"] = 2
				fmt.Println("🔧 Adjusted data_start_row from 1 to 2 (skipping header row)")
			}
		}
	}

	// 5. 自动应用AI建议的清理操作
	fms.autoApplyDataCleaningOperations(responseMap)

	return responseMap, nil
}

// isExampleRow 判断一行是否为示例行（包含描述性内容而非真实数据）
func (fms *FieldMappingService) isExampleRow(row []string) bool {
	if len(row) == 0 {
		return false
	}

	// 检查是否包含典型的示例行特征
	for _, cell := range row {
		cellLower := strings.ToLower(strings.TrimSpace(cell))

		// 如果包含这些词汇，很可能是示例行
		if strings.Contains(cellLower, "day") ||
			strings.Contains(cellLower, "postal code") ||
			strings.Contains(cellLower, "unit sales") ||
			strings.Contains(cellLower, "sales $") ||
			strings.Contains(cellLower, "retail price") ||
			strings.Contains(cellLower, "partner") {
			return true
		}
	}

	return false
}

// autoApplyDataCleaningOperations 根据AI分析结果自动应用数据清理操作
func (fms *FieldMappingService) autoApplyDataCleaningOperations(responseMap map[string]interface{}) {
	fmt.Println("🤖 Analyzing data cleaning requirements...")

	// 检查AI建议的清理操作
	weekFormatIssues := getBoolValue(responseMap, "week_format_issues")
	monthFormatIssues := getBoolValue(responseMap, "month_format_issues")
	currencySymbols := getBoolValue(responseMap, "currency_symbols")
	numberFormatting := getBoolValue(responseMap, "number_formatting")

	appliedOperations := []string{}

	// 自动应用日期格式转换 - 总是为date_code字段应用日期格式转换以确保显示完整日期
	if fms.autoApplyDateFormatting() {
		appliedOperations = append(appliedOperations, "Date format conversion (Always applied for Date Code)")
	}

	// 自动应用周格式转换
	if weekFormatIssues {
		if fms.autoApplyWeekFormatConversion() {
			appliedOperations = append(appliedOperations, "Week format conversion (YYYYWW → YYYY-MM-DD)")
		}
	}

	// 自动应用月份格式转换
	if monthFormatIssues {
		if fms.autoApplyMonthFormatConversion() {
			appliedOperations = append(appliedOperations, "Month format conversion (YYYYMM → YYYY-MM-DD)")
		}
	}

	// 自动应用货币符号清理
	if currencySymbols {
		if fms.autoApplyCurrencyCleaning() {
			appliedOperations = append(appliedOperations, "Currency symbol removal")
		}
	}

	// 自动应用数字格式清理
	if numberFormatting {
		if fms.autoApplyNumberFormatting() {
			appliedOperations = append(appliedOperations, "Number formatting cleanup")
		}
	}

	// 显示应用的操作
	if len(appliedOperations) > 0 {
		fmt.Printf("✅ Auto-applied data cleaning operations:\n")
		for _, operation := range appliedOperations {
			fmt.Printf("   🔧 %s\n", operation)
		}
		fmt.Println("💡 You can review and modify these in the Tools configuration (T)")
	} else {
		fmt.Println("ℹ️ No automatic data cleaning operations needed")
	}
}

// autoApplyDateFormatting 自动应用日期格式转换
func (fms *FieldMappingService) autoApplyDateFormatting() bool {
	// 为date_code字段添加日期转换操作
	if fieldProcessors["date_code"] == nil {
		fieldProcessors["date_code"] = &FieldProcessor{
			FieldName:  "date_code",
			Operations: []ProcessingOperation{},
		}
	}

	// 检查是否已经有日期转换操作
	for _, op := range fieldProcessors["date_code"].Operations {
		if op.Name == "convert_date_format" {
			return false // 已经存在，不重复添加
		}
	}

	// 添加日期格式转换操作
	dateOp := ProcessingOperation{
		Type:        "predefined",
		Name:        "convert_date_format",
		Pattern:     "",
		Replacement: "",
		Description: "Convert various date formats to YYYY-MM-DD (Auto-applied)",
	}

	fieldProcessors["date_code"].Operations = append(fieldProcessors["date_code"].Operations, dateOp)
	return true
}

// autoApplyCurrencyCleaning 自动应用货币符号清理
func (fms *FieldMappingService) autoApplyCurrencyCleaning() bool {
	applied := false

	// 为sales和profit字段添加货币清理操作
	for _, fieldName := range []string{"sales", "profit"} {
		if fieldProcessors[fieldName] == nil {
			fieldProcessors[fieldName] = &FieldProcessor{
				FieldName:  fieldName,
				Operations: []ProcessingOperation{},
			}
		}

		// 检查是否已经有货币符号清理操作
		hasOp := false
		for _, op := range fieldProcessors[fieldName].Operations {
			if op.Name == "remove_currency_symbols" {
				hasOp = true
				break
			}
		}

		if !hasOp {
			// 添加货币符号清理操作
			currencyOp := ProcessingOperation{
				Type:        "predefined",
				Name:        "remove_currency_symbols",
				Pattern:     `[$¥€£₹₽₩¢]`,
				Replacement: "",
				Description: "Remove currency symbols (Auto-applied)",
			}

			fieldProcessors[fieldName].Operations = append(fieldProcessors[fieldName].Operations, currencyOp)
			applied = true
		}
	}

	return applied
}

// autoApplyWeekFormatConversion 自动应用周格式转换
func (fms *FieldMappingService) autoApplyWeekFormatConversion() bool {
	applied := false

	// 为date_code字段添加周格式转换
	if fieldProcessors["date_code"] == nil {
		fieldProcessors["date_code"] = &FieldProcessor{
			FieldName:  "date_code",
			Operations: []ProcessingOperation{},
		}
	}

	processor := fieldProcessors["date_code"]

	// 移除可能存在的月格式转换（避免冲突）
	fms.removeOperationByName(processor, "convert_month_format")

	// 检查是否已经有周格式转换操作
	hasWeekConversion := false
	for _, op := range processor.Operations {
		if op.Name == "convert_week_format" {
			hasWeekConversion = true
			break
		}
	}

	if !hasWeekConversion {
		operation := ProcessingOperation{
			Type:        "predefined",
			Name:        "convert_week_format",
			Pattern:     `^\d{6}$`,
			Replacement: "",
			Description: "Convert YYYYWW format to week start date (Auto-applied)",
		}
		processor.Operations = append(processor.Operations, operation)
		applied = true
	}

	return applied
}

// autoApplyMonthFormatConversion 自动应用月份格式转换
func (fms *FieldMappingService) autoApplyMonthFormatConversion() bool {
	applied := false

	// 为date_code字段添加月份格式转换
	if fieldProcessors["date_code"] == nil {
		fieldProcessors["date_code"] = &FieldProcessor{
			FieldName:  "date_code",
			Operations: []ProcessingOperation{},
		}
	}

	processor := fieldProcessors["date_code"]

	// 移除可能存在的周格式转换（避免冲突）
	fms.removeOperationByName(processor, "convert_week_format")

	// 检查是否已经有月份格式转换操作
	hasMonthConversion := false
	for _, op := range processor.Operations {
		if op.Name == "convert_month_format" {
			hasMonthConversion = true
			break
		}
	}

	if !hasMonthConversion {
		operation := ProcessingOperation{
			Type:        "predefined",
			Name:        "convert_month_format",
			Pattern:     `^\d{6}$`,
			Replacement: "",
			Description: "Convert YYYYMM format to month start date (Auto-applied)",
		}
		processor.Operations = append(processor.Operations, operation)
		applied = true
	}

	return applied
}

// autoApplyNumberFormatting 自动应用数字格式清理
func (fms *FieldMappingService) autoApplyNumberFormatting() bool {
	applied := false

	// 为数字字段添加格式清理操作
	for _, fieldName := range []string{"sales", "profit", "orders"} {
		if fieldProcessors[fieldName] == nil {
			fieldProcessors[fieldName] = &FieldProcessor{
				FieldName:  fieldName,
				Operations: []ProcessingOperation{},
			}
		}

		// 检查是否已经有逗号清理操作
		hasOp := false
		for _, op := range fieldProcessors[fieldName].Operations {
			if op.Name == "remove_commas" {
				hasOp = true
				break
			}
		}

		if !hasOp {
			// 添加逗号清理操作
			commaOp := ProcessingOperation{
				Type:        "predefined",
				Name:        "remove_commas",
				Pattern:     `,`,
				Replacement: "",
				Description: "Remove thousand separators (Auto-applied)",
			}

			fieldProcessors[fieldName].Operations = append(fieldProcessors[fieldName].Operations, commaOp)
			applied = true
		}
	}

	return applied
}

// removeOperationByName 从处理器中移除指定名称的操作
func (fms *FieldMappingService) removeOperationByName(processor *FieldProcessor, operationName string) {
	var newOperations []ProcessingOperation
	for _, op := range processor.Operations {
		if op.Name != operationName {
			newOperations = append(newOperations, op)
		}
	}
	processor.Operations = newOperations
}

// getBoolValue 从map中获取布尔值
func getBoolValue(responseMap map[string]interface{}, key string) bool {
	if value, exists := responseMap[key]; exists {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			return strings.ToLower(v) == "true"
		}
	}
	return false
}

// PrintResults 简洁列表格式打印分析结果
func (fms *FieldMappingService) PrintResults(responseMap map[string]interface{}) {
	// 定义所有字段的顺序
	allFields := []string{
		"date_type", "date_code", "geo_type", "geo_code", "geo_name",
		"sales_platform", "sales_platform_type", "country_code",
		"orders", "sales", "profit", "new_customer_orders", "new_customer_sales",
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("               字段映射分析结果")
	fmt.Println(strings.Repeat("=", 60))

	// 逐个显示字段映射
	for _, field := range allFields {
		value := fms.getFieldValue(responseMap, field)
		isRequired := requered_fields_map[field]

		// 格式化显示值
		displayValue := value
		if value == "" {
			if isRequired {
				displayValue = " ❌ "
			} else {
				displayValue = ""
			}
		}

		// 显示字段映射
		if displayValue != "" {
			fmt.Printf("%-25s: %s\n", field, displayValue)
		}
	}

	// 打印数据位置信息
	dataStartRow := fms.getFieldValue(responseMap, "data_start_row")

	if dataStartRow != "" {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("📊 数据位置信息: 数据起始行=%s\n", dataStartRow)
	}

	// 显示必填字段验证结果
	missingRequired := []string{}
	foundRequired := []string{}

	for _, field := range allFields {
		if requered_fields_map[field] {
			value := fms.getFieldValue(responseMap, field)
			if value == "" {
				missingRequired = append(missingRequired, field)
			} else {
				foundRequired = append(foundRequired, field)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
	if len(missingRequired) > 0 {
		fmt.Printf("❌ 缺失的必填字段: %s\n", strings.Join(missingRequired, ", "))
	}
	if len(foundRequired) > 0 {
		fmt.Printf("✅ 已找到的必填字段: %s\n", strings.Join(foundRequired, ", "))
	}

	if len(missingRequired) == 0 {
		fmt.Println("🎉 所有必填字段都已找到！")
	}
	fmt.Println()
}

// getFieldValue 安全获取字段值
func (fms *FieldMappingService) getFieldValue(responseMap map[string]interface{}, field string) string {
	if value, exists := responseMap[field]; exists {
		if str, ok := value.(string); ok {
			return str
		}
		if num, ok := value.(float64); ok {
			return fmt.Sprintf("%.0f", num)
		}
		if num, ok := value.(int); ok {
			return fmt.Sprintf("%d", num)
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// getFieldDisplayName 获取字段的友好显示名称
func (fms *FieldMappingService) getFieldDisplayName(field string) string {
	displayNames := map[string]string{
		"date_type":           "日期类型",
		"date_code":           "日期代码",
		"geo_type":            "地理类型",
		"geo_code":            "地理代码",
		"geo_name":            "地理名称",
		"sales_platform":      "销售平台",
		"sales_platform_type": "平台类型",
		"country_code":        "国家代码",
		"orders":              "订单数",
		"sales":               "销售额",
		"profit":              "利润",
		"new_customer_orders": "新客户订单数",
		"new_customer_sales":  "新客户销售额",
		"data_start_row":      "数据起始行",
		"header_row":          "表头行",
	}

	if displayName, exists := displayNames[field]; exists {
		return displayName
	}
	return field
}

// getFieldNote 获取字段的备注信息
func (fms *FieldMappingService) getFieldNote(field, value string) string {
	if value == "" {
		return ""
	}

	notes := map[string]string{
		"date_type":           " (应为 DAILY/WEEKLY)",
		"geo_type":            " (应为 DMA/ZIP/STATE)",
		"sales_platform_type": " (应为 PRIMARY/SECONDARY)",
	}

	if note, exists := notes[field]; exists {
		return note
	}
	return ""
}
