package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 日期格式配置管理器
type DateFormatManager struct {
	scanner *bufio.Scanner
}

// 创建日期格式配置管理器
func NewDateFormatManager() *DateFormatManager {
	return &DateFormatManager{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// 配置日期格式检测
func (dfm *DateFormatManager) Configure(responseMap map[string]interface{}, fullData [][]string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                    Date Format Detection Configuration")
	fmt.Println("                  Configure YYYYWW/YYYYMM format handling")
	fmt.Println(strings.Repeat("=", 80))

	// 获取date_code字段映射
	dateCodeField := getMapValue(responseMap, "date_code")
	if dateCodeField == "" {
		fmt.Println("❌ Date Code field is not mapped. Please map the date field first.")
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		return
	}

	fmt.Printf("📅 Date Code Field: %s\n", cleanInferredValue(dateCodeField))

	// 分析日期数据
	dateAnalysis := dfm.analyzeDateFormats(fullData, responseMap)

	fmt.Println("\n🔍 Date Format Analysis Results:")
	fmt.Printf("  📊 Total date values analyzed: %d\n", dateAnalysis.TotalValues)
	fmt.Printf("  📅 6-digit format values: %d\n", dateAnalysis.SixDigitCount)

	if dateAnalysis.SixDigitCount > 0 {
		fmt.Printf("  📈 Range: %s - %s\n", dateAnalysis.MinValue, dateAnalysis.MaxValue)
		fmt.Printf("  🔢 Last 2 digits range: %02d - %02d\n", dateAnalysis.MinLastTwo, dateAnalysis.MaxLastTwo)

		if dateAnalysis.MaxLastTwo > 12 {
			fmt.Printf("  ✅ Detected WEEK format (values > 12 found)\n")
		} else {
			fmt.Printf("  ❓ Ambiguous format (all values ≤ 12)\n")
			fmt.Printf("  📋 Distribution: %v\n", dateAnalysis.LastTwoDistribution)
		}
	}

	fmt.Println("\n🛠️ Current Configuration:")
	weekEnabled := dfm.isDateProcessingEnabled("convert_week_format")
	monthEnabled := dfm.isDateProcessingEnabled("convert_month_format")
	fmt.Printf("  📅 Week format conversion (YYYYWW): %s\n", getEnabledStatus(weekEnabled))
	fmt.Printf("  📆 Month format conversion (YYYYMM): %s\n", getEnabledStatus(monthEnabled))

	for {
		fmt.Println("\n⚙️ Configuration Options:")
		fmt.Println("  1 - Enable WEEK format conversion (YYYYWW → YYYY-MM-DD)")
		fmt.Println("  2 - Enable MONTH format conversion (YYYYMM → YYYY-MM-DD)")
		fmt.Println("  3 - Disable both conversions (keep original values)")
		fmt.Println("  4 - Auto-detect based on data analysis")
		fmt.Println("  5 - Show sample conversions")
		fmt.Println("  0 - Back to main menu")
		fmt.Print("\n💡 Enter your choice: ")

		if !dfm.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(dfm.scanner.Text())
		if choice == "0" {
			break
		}

		switch choice {
		case "1":
			dfm.enableWeekFormatConversion()
			fmt.Println("✅ Week format conversion enabled")
		case "2":
			dfm.enableMonthFormatConversion()
			fmt.Println("✅ Month format conversion enabled")
		case "3":
			dfm.disableDateFormatConversions()
			fmt.Println("✅ Date format conversions disabled")
		case "4":
			dfm.autoConfigureDateFormat(dateAnalysis)
		case "5":
			dfm.showSampleConversions(dateAnalysis)
		default:
			fmt.Println("❌ Invalid selection, please try again")
		}
	}
}

// 分析日期格式
func (dfm *DateFormatManager) analyzeDateFormats(data [][]string, responseMap map[string]interface{}) *DateFormatAnalysis {
	analysis := &DateFormatAnalysis{
		LastTwoDistribution: make(map[int]int),
		MinLastTwo:          99,
		MaxLastTwo:          0,
	}

	// 获取date_code字段对应的列
	dateCodeField := cleanInferredValue(getMapValue(responseMap, "date_code"))
	if dateCodeField == "" {
		return analysis
	}

	// 获取表头和数据起始行
	headerRowIndex := getHeaderRowIndex(responseMap)
	if headerRowIndex >= len(data) {
		headerRowIndex = 0
	}
	headers := data[headerRowIndex]

	dateColumnIndex := findColumnIndex(headers, dateCodeField)
	if dateColumnIndex == -1 {
		return analysis
	}

	dataStartRow := getDataStartRow(responseMap)

	// 分析日期数据
	for i := dataStartRow - 1; i < len(data) && i < dataStartRow+50; i++ { // 分析最多50行
		if i >= 0 && dateColumnIndex < len(data[i]) {
			dateValue := strings.TrimSpace(data[i][dateColumnIndex])
			if dateValue == "" {
				continue
			}

			analysis.TotalValues++

			// 检查是否为6位数字格式
			if len(dateValue) == 6 {
				if _, err := strconv.Atoi(dateValue); err == nil {
					analysis.SixDigitCount++

					if analysis.MinValue == "" || dateValue < analysis.MinValue {
						analysis.MinValue = dateValue
					}
					if analysis.MaxValue == "" || dateValue > analysis.MaxValue {
						analysis.MaxValue = dateValue
					}

					// 分析后两位数字
					if lastTwoStr := dateValue[4:]; len(lastTwoStr) == 2 {
						if lastTwo, err := strconv.Atoi(lastTwoStr); err == nil {
							analysis.LastTwoDistribution[lastTwo]++
							if lastTwo < analysis.MinLastTwo {
								analysis.MinLastTwo = lastTwo
							}
							if lastTwo > analysis.MaxLastTwo {
								analysis.MaxLastTwo = lastTwo
							}
						}
					}
				}
			}
		}
	}

	// 智能判断格式
	if analysis.SixDigitCount > 0 {
		if analysis.MaxLastTwo > 12 {
			analysis.SuggestedFormat = "WEEK"
		} else if analysis.MaxLastTwo <= 12 {
			// 检查分布模式
			if dfm.isMonthlyPattern(analysis.LastTwoDistribution) {
				analysis.SuggestedFormat = "MONTH"
			} else {
				analysis.SuggestedFormat = "UNKNOWN"
			}
		}
	}

	return analysis
}

// 判断是否为月度模式
func (dfm *DateFormatManager) isMonthlyPattern(distribution map[int]int) bool {
	// 如果包含1-12的连续数字，很可能是月份
	monthCount := 0
	for month := 1; month <= 12; month++ {
		if distribution[month] > 0 {
			monthCount++
		}
	}

	// 如果有超过6个月的数据，认为是月度模式
	return monthCount >= 6
}

// 检查日期处理是否启用
func (dfm *DateFormatManager) isDateProcessingEnabled(operationName string) bool {
	processor, exists := fieldProcessors["date_code"]
	if !exists {
		return false
	}

	for _, op := range processor.Operations {
		if op.Name == operationName {
			return true
		}
	}
	return false
}

// 启用周格式转换
func (dfm *DateFormatManager) enableWeekFormatConversion() {
	if fieldProcessors["date_code"] == nil {
		fieldProcessors["date_code"] = &FieldProcessor{
			FieldName:  "date_code",
			Operations: []ProcessingOperation{},
		}
	}

	processor := fieldProcessors["date_code"]

	// 移除月格式转换（避免冲突）
	dfm.removeOperationByName(processor, "convert_month_format")

	// 检查是否已存在周格式转换
	for _, op := range processor.Operations {
		if op.Name == "convert_week_format" {
			return // 已存在
		}
	}

	// 添加周格式转换
	operation := ProcessingOperation{
		Type:        "predefined",
		Name:        "convert_week_format",
		Pattern:     `^\d{6}$`,
		Replacement: "",
		Description: "Convert YYYYWW format to week start date (User configured)",
	}
	processor.Operations = append(processor.Operations, operation)
}

// 启用月格式转换
func (dfm *DateFormatManager) enableMonthFormatConversion() {
	if fieldProcessors["date_code"] == nil {
		fieldProcessors["date_code"] = &FieldProcessor{
			FieldName:  "date_code",
			Operations: []ProcessingOperation{},
		}
	}

	processor := fieldProcessors["date_code"]

	// 移除周格式转换（避免冲突）
	dfm.removeOperationByName(processor, "convert_week_format")

	// 检查是否已存在月格式转换
	for _, op := range processor.Operations {
		if op.Name == "convert_month_format" {
			return // 已存在
		}
	}

	// 添加月格式转换
	operation := ProcessingOperation{
		Type:        "predefined",
		Name:        "convert_month_format",
		Pattern:     `^\d{6}$`,
		Replacement: "",
		Description: "Convert YYYYMM format to month start date (User configured)",
	}
	processor.Operations = append(processor.Operations, operation)
}

// 禁用日期格式转换
func (dfm *DateFormatManager) disableDateFormatConversions() {
	processor, exists := fieldProcessors["date_code"]
	if !exists {
		return
	}

	// 移除周格式和月格式转换
	dfm.removeOperationByName(processor, "convert_week_format")
	dfm.removeOperationByName(processor, "convert_month_format")
}

// 根据分析结果自动配置
func (dfm *DateFormatManager) autoConfigureDateFormat(analysis *DateFormatAnalysis) {
	fmt.Printf("\n🤖 Auto-configuration based on data analysis...\n")

	if analysis.SixDigitCount == 0 {
		fmt.Println("ℹ️ No 6-digit date formats detected. No conversion needed.")
		dfm.disableDateFormatConversions()
		return
	}

	switch analysis.SuggestedFormat {
	case "WEEK":
		fmt.Printf("✅ Auto-detected WEEK format (values up to %02d found)\n", analysis.MaxLastTwo)
		dfm.enableWeekFormatConversion()
	case "MONTH":
		fmt.Printf("✅ Auto-detected MONTH format (monthly pattern detected)\n")
		dfm.enableMonthFormatConversion()
	default:
		fmt.Printf("⚠️ Unable to auto-detect format. Manual selection recommended.\n")
		fmt.Printf("   - Last 2 digits range: %02d - %02d\n", analysis.MinLastTwo, analysis.MaxLastTwo)
		fmt.Printf("   - Consider the business context of your data\n")
	}
}

// 显示转换示例
func (dfm *DateFormatManager) showSampleConversions(analysis *DateFormatAnalysis) {
	fmt.Println("\n📋 Sample Conversions:")

	if analysis.SixDigitCount == 0 {
		fmt.Println("ℹ️ No 6-digit formats found in your data")
		return
	}

	// 显示周格式转换示例
	fmt.Println("\n📅 Week Format (YYYYWW) Examples:")
	weekSamples := []string{"202501", "202513", "202525", "202552"}
	for _, sample := range weekSamples {
		converted := convertWeekFormat(sample)
		fmt.Printf("  %s → %s\n", sample, converted)
	}

	// 显示月格式转换示例
	fmt.Println("\n📆 Month Format (YYYYMM) Examples:")
	monthSamples := []string{"202501", "202506", "202512"}
	for _, sample := range monthSamples {
		converted := convertMonthFormat(sample)
		fmt.Printf("  %s → %s\n", sample, converted)
	}

	// 显示实际数据的转换效果
	if analysis.MinValue != "" {
		fmt.Println("\n🔍 Your Data Conversion Preview:")
		fmt.Printf("  Week format: %s → %s\n", analysis.MinValue, convertWeekFormat(analysis.MinValue))
		fmt.Printf("  Month format: %s → %s\n", analysis.MinValue, convertMonthFormat(analysis.MinValue))
		if analysis.MaxValue != analysis.MinValue {
			fmt.Printf("  Week format: %s → %s\n", analysis.MaxValue, convertWeekFormat(analysis.MaxValue))
			fmt.Printf("  Month format: %s → %s\n", analysis.MaxValue, convertMonthFormat(analysis.MaxValue))
		}
	}

	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}

// 从处理器中移除指定名称的操作
func (dfm *DateFormatManager) removeOperationByName(processor *FieldProcessor, operationName string) {
	var newOperations []ProcessingOperation
	for _, op := range processor.Operations {
		if op.Name != operationName {
			newOperations = append(newOperations, op)
		}
	}
	processor.Operations = newOperations
}
