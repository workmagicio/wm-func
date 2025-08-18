package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

// 数据处理管理器
type DataProcessingManager struct {
	scanner *bufio.Scanner
}

// 创建数据处理管理器
func NewDataProcessingManager() *DataProcessingManager {
	return &DataProcessingManager{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// 预览数据映射效果
func (dpm *DataProcessingManager) PreviewDataMapping(responseMap map[string]interface{}, data [][]string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                           Data Preview")
	fmt.Println("                   Showing mapped data with current settings")
	fmt.Println(strings.Repeat("=", 80))

	if len(data) == 0 {
		fmt.Println("❌ No data available for preview")
		return
	}

	processedData := data

	// 获取数据起始行
	dataStartRow := getDataStartRow(responseMap)
	if dataStartRow <= 0 || dataStartRow > len(processedData) {
		fmt.Printf("❌ Invalid data start row: %d (file has %d rows)\n", dataStartRow, len(processedData))
		return
	}

	// 显示映射配置
	dpm.showMappingConfiguration(responseMap)

	// 显示预览数据
	dpm.showPreviewData(processedData, responseMap, dataStartRow)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📝 Preview completed. Press Enter to continue...")
	fmt.Scanln() // 等待用户按回车
}

// 显示映射配置
func (dpm *DataProcessingManager) showMappingConfiguration(responseMap map[string]interface{}) {
	fmt.Println("\n📋 Current Field Mapping:")
	fmt.Println(strings.Repeat("-", 60))

	for _, field := range allMappableFields {
		if field == "data_start_row" {
			continue // 跳过系统配置字段
		}

		value := getMapValue(responseMap, field)
		displayName := getFieldDisplayName(field)

		if value != "" {
			fmt.Printf("  %-20s -> %s\n", displayName, cleanInferredValue(value))
		} else {
			fmt.Printf("  %-20s -> \n", displayName)
		}
	}

	dataStartRow := getDataStartRow(responseMap)
	fmt.Printf("\n🔧 Data Start Row: %d\n", dataStartRow)
}

// 显示预览数据
func (dpm *DataProcessingManager) showPreviewData(data [][]string, responseMap map[string]interface{}, dataStartRow int) {
	fmt.Println("\n📊 Data Preview (first 10 rows):")

	// 定义不同字段的列宽度
	fieldWidths := map[string]int{
		"Date Type":           8,  // WEEKLY
		"Date Code":           12, // 2025-01-06 (完整日期显示)
		"Geo Type":            8,  // ZIP/DMA/STATE
		"Geo Code":            10, // 地理代码
		"Geo Name":            10, // 地理名称
		"Sales Platform":      10, // 平台名称
		"Country Code":        8,  // US/CA
		"Orders":              8,  // 订单数
		"Sales":               10, // 销售额
		"Profit":              10, // 利润
		"New Customer Orders": 8,  // 新客户订单
		"New Customer Sales":  10, // 新客户销售额
	}

	// 显示表头
	headers := []string{}
	for _, field := range allMappableFields {
		if field == "data_start_row" {
			continue
		}
		displayName := getFieldDisplayName(field)
		headers = append(headers, displayName)
	}

	// 打印表头
	totalWidth := 0
	for i, header := range headers {
		width := fieldWidths[header]
		if width == 0 {
			width = 10 // 默认宽度
		}
		fmt.Printf("%-*s", width, truncateString(header, width-1))
		if i < len(headers)-1 {
			fmt.Print(" | ")
		}
		totalWidth += width + 3 // 3 for " | "
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", totalWidth-3))

	// 获取正确的表头行
	var headerRowIndex int
	if dataStartRow == 1 {
		// 如果数据从第1行开始，第1行既是表头又是数据起始行
		headerRowIndex = 0
	} else {
		// 表头在数据起始行的前一行
		headerRowIndex = dataStartRow - 2 // 转换为0基索引
		if headerRowIndex < 0 || headerRowIndex >= len(data) {
			// 如果表头行无效，使用第一行作为表头
			headerRowIndex = 0
		}
	}
	excelHeaders := data[headerRowIndex]

	// 显示数据行（最多10行）
	maxRows := 10
	actualDataRows := len(data) - dataStartRow + 1
	if actualDataRows > maxRows {
		actualDataRows = maxRows
	}

	for i := 0; i < actualDataRows && (dataStartRow-1+i) < len(data); i++ {
		row := data[dataStartRow-1+i]
		mappedRow := dpm.mapRowData(row, responseMap, excelHeaders) // 使用正确的表头行

		for j, value := range mappedRow {
			// 使用与表头相同的宽度
			header := headers[j]
			width := fieldWidths[header]
			if width == 0 {
				width = 10 // 默认宽度
			}
			fmt.Printf("%-*s", width, truncateString(value, width-1))
			if j < len(mappedRow)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Println()
	}

	if len(data)-dataStartRow+1 > maxRows {
		fmt.Printf("\n... and %d more rows\n", len(data)-dataStartRow+1-maxRows)
	}
}

// 根据映射配置转换行数据
func (dpm *DataProcessingManager) mapRowData(row []string, responseMap map[string]interface{}, headers []string) []string {
	result := []string{}

	for _, field := range allMappableFields {
		if field == "data_start_row" {
			continue
		}

		mappedValue := getMapValue(responseMap, field)
		var cellValue string

		if mappedValue == "" {
			cellValue = "" // 直接显示为空
		} else if mappedValue == "VIRTUAL_COUNT" {
			// 虚拟计算字段，显示为1
			cellValue = "1"
		} else {
			// 查找对应的列索引
			colIndex := findColumnIndex(headers, cleanInferredValue(mappedValue))
			if colIndex >= 0 && colIndex < len(row) {
				cellValue = row[colIndex]
				// 空值直接显示为空字符串
			} else {
				// 如果不是列映射，可能是固定值，清理推断标记
				cellValue = cleanInferredValue(mappedValue)
			}

			// 应用字段处理操作
			cellValue = applyFieldProcessing(cellValue, field)
		}

		result = append(result, cellValue)
	}

	return result
}

// 下载完整CSV文件
func (dpm *DataProcessingManager) DownloadCSV(responseMap map[string]interface{}, filename string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                           CSV Download")
	fmt.Println("                    Generating complete CSV file")
	fmt.Println(strings.Repeat("=", 80))

	// 验证必填字段
	if !dpm.validateAllRequiredFields(responseMap) {
		fmt.Println("❌ Cannot download CSV: Required fields are missing")
		fmt.Println("Please complete all required field mappings before downloading.")
		fmt.Println("\nPress Enter to continue...")
		fmt.Scanln()
		return
	}

	// 获取数据起始行，确定需要读取多少数据
	dataStartRow := getDataStartRow(responseMap)

	// 读取完整数据，需要包含表头行和所有数据行
	fmt.Printf("📖 Reading complete Excel data (starting from row %d)...\n", dataStartRow)

	// 使用ReadFileData读取所有数据（不限制行数）
	completeData, err := ReadFileData(filename)
	if err != nil {
		fmt.Printf("❌ Failed to read complete data: %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	if len(completeData) == 0 {
		fmt.Println("❌ No data found in file")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	// 验证数据起始行是否有效
	if dataStartRow < 1 || dataStartRow > len(completeData) {
		fmt.Printf("❌ Invalid data start row: %d (file has %d rows)\n", dataStartRow, len(completeData))
		fmt.Println("Please check your data start row setting.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	processedData := completeData

	// 询问用户是否需要分组操作
	needGrouping := dpm.askForGrouping(responseMap)
	var csvContent [][]string

	if needGrouping {
		fmt.Println("🔄 Processing data with advanced grouping...")
		result, groupErr := dpm.processWithSQLiteGrouping(processedData, responseMap)
		if groupErr != nil {
			fmt.Printf("❌ Advanced processing failed: %v\n", groupErr)
			fmt.Println("Falling back to normal processing...")
			csvContent = dpm.generateCSVContent(processedData, responseMap)
		} else {
			csvContent = result
		}
	} else {
		fmt.Println("🔄 Processing data with current mapping...")
		csvContent = dpm.generateCSVContent(processedData, responseMap)
	}

	// 生成文件名
	outputFilename := dpm.generateOutputFilename(filename)

	// 写入CSV文件
	fmt.Printf("💾 Writing CSV file: %s\n", outputFilename)
	err = dpm.writeCSVFile(outputFilename, csvContent)
	if err != nil {
		fmt.Printf("❌ Failed to write CSV file: %v\n", err)
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	// 显示完成信息
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("✅ CSV file generated successfully!\n")
	fmt.Printf("📁 File location: %s\n", outputFilename)
	fmt.Printf("📊 Total rows processed: %d (including header)\n", len(csvContent))
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎉 Download completed! Program will exit.")
	fmt.Println("Thank you for using the field mapping tool!")

	// 下载完成后退出程序
	os.Exit(0)
}

// 验证所有必填字段是否已完成
func (dpm *DataProcessingManager) validateAllRequiredFields(responseMap map[string]interface{}) bool {
	for fieldName, isRequired := range requered_fields_map {
		if isRequired {
			value := getMapValue(responseMap, fieldName)
			if value == "" {
				return false
			}
		}
	}
	return true
}

// 生成CSV内容
func (dpm *DataProcessingManager) generateCSVContent(data [][]string, responseMap map[string]interface{}) [][]string {
	if len(data) == 0 {
		return [][]string{}
	}

	result := [][]string{}

	// 生成CSV表头
	csvHeader := []string{}
	for _, field := range allMappableFields {
		if field == "data_start_row" {
			continue
		}
		displayName := getFieldDisplayName(field)
		csvHeader = append(csvHeader, displayName)
	}
	result = append(result, csvHeader)

	// 获取数据起始行
	dataStartRow := getDataStartRow(responseMap)
	if dataStartRow < 1 || dataStartRow > len(data) {
		return result // 只返回表头
	}

	// 表头通常在数据起始行的前一行
	var headerRowIndex int
	if dataStartRow == 1 {
		// 如果数据从第1行开始，第1行既是表头又是数据起始行
		headerRowIndex = 0
	} else {
		// 表头在数据起始行的前一行
		headerRowIndex = dataStartRow - 2 // 转换为0基索引
		if headerRowIndex < 0 || headerRowIndex >= len(data) {
			// 如果表头行无效，使用第一行作为表头
			headerRowIndex = 0
		}
	}
	headers := data[headerRowIndex]

	// 处理数据行（从数据起始行开始）
	for i := dataStartRow - 1; i < len(data); i++ {
		row := data[i]
		csvRow := dpm.mapRowToCSV(row, responseMap, headers)
		result = append(result, csvRow)
	}

	return result
}

// 将Excel行映射为CSV行
func (dpm *DataProcessingManager) mapRowToCSV(row []string, responseMap map[string]interface{}, headers []string) []string {
	result := []string{}

	for _, field := range allMappableFields {
		if field == "data_start_row" {
			continue
		}

		mappedValue := getMapValue(responseMap, field)
		var cellValue string

		if mappedValue == "" {
			cellValue = "" // 未映射的字段为空
		} else if mappedValue == "VIRTUAL_COUNT" {
			// 虚拟计算字段，设置为1（每行代表1个订单）
			cellValue = "1"
		} else {
			// 查找对应的列索引
			cleanMappedValue := cleanInferredValue(mappedValue)
			colIndex := findColumnIndex(headers, cleanMappedValue)
			if colIndex >= 0 && colIndex < len(row) {
				cellValue = row[colIndex]
			} else {
				// 如果不是列映射，可能是固定值
				cellValue = cleanMappedValue
			}

			// 应用字段处理操作
			cellValue = applyFieldProcessing(cellValue, field)
		}

		result = append(result, cellValue)
	}

	return result
}

// 生成输出文件名
func (dpm *DataProcessingManager) generateOutputFilename(inputFilename string) string {
	// 获取文件名（不含路径）
	baseName := inputFilename
	if lastSlash := strings.LastIndex(inputFilename, "/"); lastSlash >= 0 {
		baseName = inputFilename[lastSlash+1:]
	}
	if lastSlash := strings.LastIndex(baseName, "\\"); lastSlash >= 0 {
		baseName = baseName[lastSlash+1:]
	}

	// 移除扩展名
	if lastDot := strings.LastIndex(baseName, "."); lastDot >= 0 {
		baseName = baseName[:lastDot]
	}

	// 添加时间戳和CSV扩展名
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	return fmt.Sprintf("%s_mapped_%s.csv", baseName, timestamp)
}

// 写入CSV文件
func (dpm *DataProcessingManager) writeCSVFile(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// 询问用户是否需要分组操作
func (dpm *DataProcessingManager) askForGrouping(responseMap map[string]interface{}) bool {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("                    Data Grouping Option")
	fmt.Println(strings.Repeat("=", 60))

	// 检查orders字段的配置
	ordersValue := getMapValue(responseMap, "orders")
	hasVirtualOrders := ordersValue == "VIRTUAL_COUNT"

	if hasVirtualOrders {
		fmt.Println("\n🔍 Detected virtual orders calculation in your mapping.")
		fmt.Println("📊 Grouping is RECOMMENDED to properly aggregate your data.")
		fmt.Println("")
		fmt.Println("   With grouping enabled:")
		fmt.Println("   • Your product-level data will be aggregated to order-level")
		fmt.Println("   • Sales and profit will be summed by dimension groups")
		fmt.Println("   • Orders will show the actual count per group")
		fmt.Println("")
	} else {
		fmt.Println("\n📊 Do you want to group/aggregate your data?")
		fmt.Println("   This will group by dimension fields and sum numeric fields.")
		fmt.Println("   Only needed if you have duplicate dimension combinations.")
		fmt.Println("")
	}

	// 检查是否配置了order_id字段
	orderIDValue := getMapValue(responseMap, "order_id")
	hasOrderID := orderIDValue != "" && orderIDValue != "VIRTUAL_COUNT"

	fmt.Println("   Grouping configuration:")
	fmt.Println("   • Group by: date_type, date_code, geo_type, geo_code,")
	fmt.Println("             sales_platform, sales_platform_type, country_code")
	fmt.Println("   • Sum: sales, profit")
	if hasOrderID {
		fmt.Printf("   • Count: orders (unique Order IDs from '%s' column)\n", cleanInferredValue(orderIDValue))
		fmt.Println("             📊 Using COUNT(DISTINCT order_id) for accurate order counting")
	} else {
		fmt.Println("   • Count: orders (number of rows in each group)")
		fmt.Println("             📊 Using COUNT(*) - each row counts as one order")
	}
	fmt.Println("   • Keep: All other fields (first non-empty value per group)")
	fmt.Println("")

	if hasVirtualOrders {
		fmt.Print("💡 Enable grouping? (Y/n): ")
	} else {
		fmt.Print("💡 Enable grouping? (y/N): ")
	}

	if !dpm.scanner.Scan() {
		return hasVirtualOrders // 默认值基于是否有虚拟orders
	}

	choice := strings.TrimSpace(strings.ToLower(dpm.scanner.Text()))
	if choice == "" {
		return hasVirtualOrders // 空输入使用默认值
	}
	return choice == "y" || choice == "yes"
}

// 使用SQLite进行分组处理
func (dpm *DataProcessingManager) processWithSQLiteGrouping(data [][]string, responseMap map[string]interface{}) ([][]string, error) {
	// 创建SQLite服务
	sqliteService, err := NewSQLiteService()
	if err != nil {
		return nil, fmt.Errorf("创建数据处理服务失败: %v", err)
	}
	defer sqliteService.Close()

	// 导入数据进行处理
	fmt.Println("📥 Importing data for processing...")
	err = sqliteService.ImportData(data, responseMap)
	if err != nil {
		return nil, fmt.Errorf("导入数据失败: %v", err)
	}

	// 显示导入统计
	rowCount, _ := sqliteService.GetRowCount()
	fmt.Printf("📊 Processing %d rows of data\n", rowCount)

	// 定义默认的分组和聚合字段
	groupByFields := []string{"date_type", "date_code", "geo_type", "geo_code", "sales_platform", "sales_platform_type", "country_code"}
	sumFields := []string{"sales", "profit", "orders"}

	// 执行分组操作
	fmt.Println("🔍 Executing grouping operation...")
	result, err := sqliteService.ExecuteGroupQuery(groupByFields, sumFields, responseMap)
	if err != nil {
		return nil, fmt.Errorf("执行分组操作失败: %v", err)
	}

	fmt.Printf("✅ Grouping completed: %d groups created\n", len(result)-1)
	return result, nil
}
