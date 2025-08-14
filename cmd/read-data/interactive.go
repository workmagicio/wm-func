package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 交互式字段映射管理器
type InteractiveFieldMapper struct {
	service *FieldMappingService
	scanner *bufio.Scanner
}

// 创建交互式字段映射管理器
func NewInteractiveFieldMapper(service *FieldMappingService) *InteractiveFieldMapper {
	return &InteractiveFieldMapper{
		service: service,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// 交互式字段映射补充功能
func (ifm *InteractiveFieldMapper) Run(responseMap map[string]interface{}, filename string, fullData [][]string) {
	// 记录AI推断的字段
	aiInferredFields := make(map[string]bool)
	for fieldName, value := range responseMap {
		if str, ok := value.(string); ok && strings.Contains(str, "(inferred)") {
			aiInferredFields[fieldName] = true
		}
	}

	// 从已读取的数据中获取表头
	if len(fullData) == 0 {
		fmt.Println("No data available for headers")
		return
	}

	// 获取AI推断的表头行
	headerRowIndex := getHeaderRowIndex(responseMap)
	if headerRowIndex < 0 || headerRowIndex >= len(fullData) {
		fmt.Printf("⚠️ Invalid header row index: %d, using first row as fallback\n", headerRowIndex+1)
		headerRowIndex = 0
	}
	excelHeaders := fullData[headerRowIndex]
	fmt.Printf("📋 Using row %d as header row\n", headerRowIndex+1)

	// 检测product level数据
	detection := detectProductLevel(fullData, responseMap)
	if detection.IsProductLevel {
		fmt.Printf("🔍 Product-level data detected (confidence: %.1f%%)\n", detection.ConfidenceScore*100)
		if len(detection.OrderIDFields) > 0 {
			fmt.Printf("   📦 Order ID fields: %s\n", strings.Join(detection.OrderIDFields, ", "))
		}
		if len(detection.ProductIDFields) > 0 {
			fmt.Printf("   🏷️  Product ID fields: %s\n", strings.Join(detection.ProductIDFields, ", "))
		}
		if len(detection.SKUFields) > 0 {
			fmt.Printf("   📋 SKU fields: %s\n", strings.Join(detection.SKUFields, ", "))
		}
		fmt.Println("   💡 Consider using Group operation to merge duplicate orders")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    Interactive Field Mapping")
	fmt.Println("            Select field to modify by entering number (1-14)")
	fmt.Println("")
	fmt.Println("⚠️  NOTE: AI-inferred values may not be accurate. Please review and verify.")
	fmt.Println(strings.Repeat("=", 80))

	for {
		// 显示当前可以修改的字段
		ifm.showAvailableFields(responseMap, fullData, aiInferredFields)

		// 显示必填字段状态摘要
		ifm.showRequiredFieldsSummary(responseMap)

		fmt.Println("\n🔍 Special Options:")
		fmt.Println("  P - Preview data with current mapping")
		fmt.Println("  D - Download complete CSV file")
		fmt.Println("  T - Configure field processing tools")
		fmt.Println("  F - Configure date format detection (YYYYWW/YYYYMM)")
		fmt.Print("\n💡 Enter field number (1-14), 'P' for preview, 'D' for download, 'T' for tools, 'F' for format, or 0 to exit: ")
		if !ifm.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(ifm.scanner.Text())
		if choice == "0" {
			fmt.Println("Exiting field mapping...")
			break
		}

		// 处理预览选项
		if strings.ToUpper(choice) == "P" {
			previewDataMapping(responseMap, fullData)
			continue
		}

		// 处理下载选项
		if strings.ToUpper(choice) == "D" {
			downloadCSV(responseMap, filename)
			continue
		}

		// 处理工具配置选项
		if strings.ToUpper(choice) == "T" {
			configureFieldTools()
			continue
		}

		// 处理日期格式配置选项
		if strings.ToUpper(choice) == "F" {
			configureDateFormatDetection(responseMap, fullData)
			continue
		}

		// 处理用户选择的字段
		fieldName := ifm.getFieldNameByChoice(choice)
		if fieldName == "" {
			fmt.Println("❌ Invalid selection, please try again")
			continue
		}

		// 显示当前字段信息
		currentValue := getMapValue(responseMap, fieldName)
		displayName := getFieldDisplayName(fieldName)
		specialNote := getFieldSpecialNote(fieldName)

		fmt.Printf("\n🔧 Modifying: %s (%s)%s\n", displayName, fieldName, specialNote)

		// 为数据配置字段添加特殊说明
		if fieldName == "data_start_row" {
			fmt.Printf("   📋 This configures which row your data starts in the Excel file\n")
		}

		if currentValue != "" {
			fmt.Printf("   Current value: ✅ %s\n", currentValue)
		} else {
			if requered_fields_map[fieldName] {
				fmt.Printf("   Current value: ❌ Missing\n")
			} else {
				fmt.Printf("   Current value: 📝 Optional\n")
			}
		}

		var newValue string
		// 对于数据起始行字段，直接设置固定值
		if fieldName == "data_start_row" {
			newValue = ifm.setFixedValue(fieldName)
		} else if fieldName == "orders" {
			// 为orders字段提供特殊选项
			mappingType := ifm.chooseMappingTypeForOrders(excelHeaders)
			if mappingType == "" {
				continue // 用户取消，继续主循环
			}

			if mappingType == "excel" {
				// 映射到Excel列
				newValue = ifm.chooseExcelColumn(excelHeaders)
			} else if mappingType == "virtual" {
				// 设置为虚拟计算
				newValue = "VIRTUAL_COUNT"
			} else {
				// 设置固定值
				newValue = ifm.setFixedValue(fieldName)
			}
		} else {
			// 让用户选择映射方式
			mappingType := ifm.chooseMappingType()
			if mappingType == "" {
				continue // 用户取消，继续主循环
			}

			if mappingType == "excel" {
				// 映射到Excel列
				newValue = ifm.chooseExcelColumn(excelHeaders)
			} else {
				// 设置固定值
				newValue = ifm.setFixedValue(fieldName)
			}
		}

		if newValue != "" {
			// 更新映射
			responseMap[fieldName] = newValue
			fmt.Printf("✅ Updated %s = %s\n", fieldName, newValue)

			// 显示更新后的结果
			fmt.Println("\nUpdated mapping results:")
			ifm.service.PrintResults(responseMap)
		}
	}

	fmt.Println("\n🎉 Field mapping completed! Final results have been saved.")
}

// 显示可修改的字段列表
func (ifm *InteractiveFieldMapper) showAvailableFields(responseMap map[string]interface{}, fullData [][]string, aiInferredFields map[string]bool) {
	fmt.Println("\n📝 Available Fields (select by number):")
	fmt.Println("    ✅ = User configured, ✨ = AI inferred - Please verify accuracy")
	fmt.Println("    [Data Configuration] = System settings, not business data fields")
	fmt.Println("    [Column] = Mapped to Excel column, [Fixed] = Fixed value, [System] = AI-inferred setting")
	fmt.Println(strings.Repeat("-", 80))

	for i, field := range allMappableFields {
		value := getMapValue(responseMap, field)
		var status string

		if value != "" {
			// 对于data_start_row字段，显示为系统配置
			if field == "data_start_row" {
				status = fmt.Sprintf("✅ %s [System]", value)
			} else if value == "VIRTUAL_COUNT" {
				status = fmt.Sprintf("✅ Virtual Count [Auto]")
			} else {
				// 判断是否为AI推断的值（检查原始标记和记录）
				isInferred := strings.Contains(value, "(inferred)") || aiInferredFields[field]
				cleanValue := cleanInferredValue(value)

				// 特殊处理必须映射到Excel列的字段，强制显示为列映射
				mustBeColumnFields := []string{"date_code", "geo_code", "geo_name", "sales_platform", "sales", "profit", "orders", "new_customer_orders", "new_customer_sales"}
				isMustBeColumn := false
				for _, mustField := range mustBeColumnFields {
					if field == mustField {
						isMustBeColumn = true
						break
					}
				}

				if isMustBeColumn {
					// 这些字段总是应该显示为列映射，不管检测结果如何
					if isInferred {
						status = fmt.Sprintf("✨ %s [Column]", cleanValue)
					} else {
						status = fmt.Sprintf("✅ %s [Column]", cleanValue)
					}
				} else if isFixedValue(cleanValue, fullData, responseMap) {
					if isInferred {
						status = fmt.Sprintf("✨ %s [Fixed]", cleanValue)
					} else {
						status = fmt.Sprintf("✅ %s [Fixed]", cleanValue)
					}
				} else {
					if isInferred {
						status = fmt.Sprintf("✨ %s [Column]", cleanValue)
					} else {
						status = fmt.Sprintf("✅ %s [Column]", cleanValue)
					}
				}
			}
		} else {
			// 根据字段类型显示不同的缺失状态
			if requered_fields_map[field] {
				status = "❌ Missing"
			} else {
				status = "📝 Optional"
			}
		}

		isRequired := ""
		if requered_fields_map[field] {
			isRequired = " [Required]"
		}

		// 获取友好的字段名称和特殊说明
		displayName := getFieldDisplayName(field)
		specialNote := getFieldSpecialNote(field)

		fmt.Printf(" %2d. %-20s (%s): %s%s%s\n", i+1, displayName, field, status, isRequired, specialNote)
	}
}

// 显示必填字段状态摘要
func (ifm *InteractiveFieldMapper) showRequiredFieldsSummary(responseMap map[string]interface{}) {
	missingRequired := []string{}
	foundRequired := []string{}

	for _, field := range allMappableFields {
		if requered_fields_map[field] {
			value := getMapValue(responseMap, field)
			if value == "" {
				missingRequired = append(missingRequired, field)
			} else {
				foundRequired = append(foundRequired, field)
			}
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	totalRequired := len(missingRequired) + len(foundRequired)
	fmt.Printf("📊 Required Fields Progress: %d/%d completed\n", len(foundRequired), totalRequired)

	if len(missingRequired) > 0 {
		fmt.Printf("⚠️  Missing: %s\n", strings.Join(missingRequired, ", "))
	}

	if len(missingRequired) == 0 {
		fmt.Println("🎉 All required fields completed!")
	}
}

// 根据用户选择获取字段名
func (ifm *InteractiveFieldMapper) getFieldNameByChoice(choice string) string {
	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(allMappableFields) {
		return ""
	}

	return allMappableFields[index-1]
}

// 让用户选择映射方式
func (ifm *InteractiveFieldMapper) chooseMappingType() string {
	fmt.Println("\nSelect mapping type:")
	fmt.Println("1. Map to Excel column")
	fmt.Println("2. Set fixed value")
	fmt.Print("Enter choice (1 or 2, 0 to cancel): ")

	if !ifm.scanner.Scan() {
		return ""
	}

	choice := strings.TrimSpace(ifm.scanner.Text())
	switch choice {
	case "1":
		return "excel"
	case "2":
		return "fixed"
	case "0":
		return ""
	default:
		fmt.Println("❌ Invalid choice")
		return ""
	}
}

// 为orders字段选择映射方式（包含虚拟列选项）
func (ifm *InteractiveFieldMapper) chooseMappingTypeForOrders(headers []string) string {
	fmt.Println("\n🔧 Configuring Orders field:")

	// 检查是否有可能的orders相关列
	var possibleOrdersColumns []string
	orderPatterns := []string{"order", "orders", "count", "quantity", "qty", "units", "件数", "订单"}

	for _, header := range headers {
		headerLower := strings.ToLower(header)
		for _, pattern := range orderPatterns {
			if strings.Contains(headerLower, pattern) {
				possibleOrdersColumns = append(possibleOrdersColumns, header)
				break
			}
		}
	}

	if len(possibleOrdersColumns) > 0 {
		fmt.Printf("💡 Found potential orders columns: %s\n", strings.Join(possibleOrdersColumns, ", "))
	}

	fmt.Println("\nSelect mapping type:")
	fmt.Println("1. Map to Excel column (recommended if you have order count data)")
	fmt.Println("2. Set fixed value (e.g., 1 if each row = 1 order)")
	fmt.Println("3. Use virtual calculation (for product-level data needing aggregation)")
	fmt.Print("Enter choice (1, 2, or 3, 0 to cancel): ")

	if !ifm.scanner.Scan() {
		return ""
	}

	choice := strings.TrimSpace(ifm.scanner.Text())
	switch choice {
	case "1":
		return "excel"
	case "2":
		return "fixed"
	case "3":
		return "virtual"
	case "0":
		return ""
	default:
		fmt.Println("❌ Invalid choice")
		return ""
	}
}

// 让用户选择Excel列
func (ifm *InteractiveFieldMapper) chooseExcelColumn(headers []string) string {
	fmt.Println("\nAvailable Excel columns:")
	for i, header := range headers {
		fmt.Printf("%2d. %s\n", i+1, header)
	}

	fmt.Print("Select column number (0 to cancel): ")
	if !ifm.scanner.Scan() {
		return ""
	}

	choice := strings.TrimSpace(ifm.scanner.Text())
	if choice == "0" {
		return ""
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(headers) {
		fmt.Println("❌ Invalid column number")
		return ""
	}

	return headers[index-1]
}

// 让用户设置固定值
func (ifm *InteractiveFieldMapper) setFixedValue(fieldName string) string {
	// 为特定字段提供建议值
	suggestions := getFieldSuggestions(fieldName)
	if len(suggestions) > 0 {
		fmt.Printf("\nSuggested values for %s:\n", fieldName)
		for i, suggestion := range suggestions {
			fmt.Printf("%2d. %s\n", i+1, suggestion)
		}
		fmt.Print("Select suggestion number or enter custom value (0 to cancel): ")
	} else {
		// 为数据起始行字段提供特殊提示
		if fieldName == "data_start_row" {
			fmt.Printf("Enter data start row number (e.g., 2, 3, 4..., 0 to cancel): ")
		} else {
			fmt.Printf("Enter fixed value for %s (0 to cancel): ", fieldName)
		}
	}

	if !ifm.scanner.Scan() {
		return ""
	}

	input := strings.TrimSpace(ifm.scanner.Text())
	if input == "0" {
		return ""
	}

	// 如果输入的是数字且有建议值，则选择建议值
	if len(suggestions) > 0 {
		if index, err := strconv.Atoi(input); err == nil && index > 0 && index <= len(suggestions) {
			return suggestions[index-1]
		}
	}

	// 对于数据起始行字段，验证输入是否为有效的正整数
	if fieldName == "data_start_row" {
		if rowNum, err := strconv.Atoi(input); err != nil || rowNum < 1 {
			fmt.Println("❌ Invalid row number. Please enter a positive integer (1, 2, 3...)")
			return ""
		}
	}

	// 否则直接返回用户输入
	return input
}
