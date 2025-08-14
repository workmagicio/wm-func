package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// 字段工具配置管理器
type FieldToolsManager struct {
	scanner *bufio.Scanner
}

// 创建字段工具配置管理器
func NewFieldToolsManager() *FieldToolsManager {
	return &FieldToolsManager{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// 配置字段处理工具
func (ftm *FieldToolsManager) Configure() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                        Field Processing Tools")
	fmt.Println("                   Configure data cleaning operations")
	fmt.Println(strings.Repeat("=", 80))

	for {
		// 显示支持工具的字段
		fmt.Println("\n🔧 Fields with available tools:")
		supportedFields := []string{"date_code", "geo_code", "sales", "profit", "orders"}
		for i, field := range supportedFields {
			displayName := getFieldDisplayName(field)
			fmt.Printf("  %d. %s", i+1, displayName)

			// 显示当前配置的工具
			if processor, exists := fieldProcessors[field]; exists && len(processor.Operations) > 0 {
				autoCount := 0
				for _, op := range processor.Operations {
					if strings.Contains(op.Description, "Auto-applied") {
						autoCount++
					}
				}
				if autoCount > 0 {
					fmt.Printf(" [%d tools (%d auto-applied)]", len(processor.Operations), autoCount)
				} else {
					fmt.Printf(" [%d tools configured]", len(processor.Operations))
				}
			}
			fmt.Println()
		}

		fmt.Println("\n🛠️ Options:")
		fmt.Println("  1-5 - Configure tools for field")
		fmt.Println("  0   - Back to main menu")
		fmt.Print("\n💡 Enter your choice: ")

		if !ftm.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(ftm.scanner.Text())
		if choice == "0" {
			break
		}

		// 处理字段选择
		switch choice {
		case "1":
			ftm.configureFieldProcessor("date_code")
		case "2":
			ftm.configureFieldProcessor("geo_code")
		case "3":
			ftm.configureFieldProcessor("sales")
		case "4":
			ftm.configureFieldProcessor("profit")
		case "5":
			ftm.configureFieldProcessor("orders")
		default:
			fmt.Println("❌ Invalid selection, please try again")
		}
	}
}

// 配置特定字段的处理器
func (ftm *FieldToolsManager) configureFieldProcessor(fieldName string) {
	fmt.Print("\n" + strings.Repeat("-", 60))
	fmt.Printf("\n        Configuring tools for: %s", getFieldDisplayName(fieldName))
	fmt.Print("\n" + strings.Repeat("-", 60))

	// 初始化字段处理器（如果不存在）
	if fieldProcessors[fieldName] == nil {
		fieldProcessors[fieldName] = &FieldProcessor{
			FieldName:  fieldName,
			Operations: []ProcessingOperation{},
		}
	}

	for {
		processor := fieldProcessors[fieldName]

		// 显示当前配置的操作
		fmt.Printf("\n📋 Current operations for %s:\n", getFieldDisplayName(fieldName))
		if len(processor.Operations) == 0 {
			fmt.Println("  (No operations configured)")
		} else {
			for i, op := range processor.Operations {
				autoApplied := ""
				if strings.Contains(op.Description, "Auto-applied") {
					autoApplied = " 🤖"
				}
				fmt.Printf("  %d. [%s] %s%s\n", i+1, op.Type, op.Description, autoApplied)
			}
		}

		// 显示预定义操作
		if predefined, exists := predefinedOperations[fieldName]; exists {
			fmt.Printf("\n🎯 Available predefined operations:\n")
			for i, op := range predefined {
				fmt.Printf("  P%d. %s\n", i+1, op.Description)
			}
		}

		fmt.Println("\n🛠️ Options:")
		fmt.Println("  P1-P3 - Add predefined operation")
		fmt.Println("  R     - Add custom regex operation")
		fmt.Println("  C     - Clear all operations")
		fmt.Println("  T     - Test operations with sample data")
		fmt.Println("  0     - Back to field list")
		fmt.Print("\n💡 Enter your choice: ")

		if !ftm.scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(strings.ToUpper(ftm.scanner.Text()))
		if choice == "0" {
			break
		}

		switch choice {
		case "P1", "P2", "P3":
			ftm.addPredefinedOperation(fieldName, choice)
		case "R":
			ftm.addCustomRegexOperation(fieldName)
		case "C":
			ftm.clearFieldOperations(fieldName)
		case "T":
			ftm.testFieldOperations(fieldName)
		default:
			fmt.Println("❌ Invalid selection, please try again")
		}
	}
}

// 添加预定义操作
func (ftm *FieldToolsManager) addPredefinedOperation(fieldName, choice string) {
	predefined, exists := predefinedOperations[fieldName]
	if !exists {
		fmt.Println("❌ No predefined operations available for this field")
		return
	}

	var opIndex int
	switch choice {
	case "P1":
		opIndex = 0
	case "P2":
		opIndex = 1
	case "P3":
		opIndex = 2
	default:
		fmt.Println("❌ Invalid predefined operation")
		return
	}

	if opIndex >= len(predefined) {
		fmt.Println("❌ Operation not available")
		return
	}

	operation := predefined[opIndex]
	processor := fieldProcessors[fieldName]

	// 检查是否已经添加
	for _, existing := range processor.Operations {
		if existing.Name == operation.Name {
			fmt.Printf("⚠️ Operation '%s' is already configured\n", operation.Description)
			return
		}
	}

	processor.Operations = append(processor.Operations, operation)
	fmt.Printf("✅ Added operation: %s\n", operation.Description)
}

// 添加自定义正则操作
func (ftm *FieldToolsManager) addCustomRegexOperation(fieldName string) {
	fmt.Println("\n" + strings.Repeat("-", 40))
	fmt.Println("       Custom Regex Operation")
	fmt.Println(strings.Repeat("-", 40))

	fmt.Print("📝 Enter operation description: ")
	if !ftm.scanner.Scan() {
		return
	}
	description := strings.TrimSpace(ftm.scanner.Text())
	if description == "" {
		fmt.Println("❌ Description cannot be empty")
		return
	}

	fmt.Print("🔍 Enter regex pattern (e.g., [^\\d.] to keep only numbers and dots): ")
	if !ftm.scanner.Scan() {
		return
	}
	pattern := strings.TrimSpace(ftm.scanner.Text())
	if pattern == "" {
		fmt.Println("❌ Pattern cannot be empty")
		return
	}

	// 验证正则表达式
	_, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("❌ Invalid regex pattern: %v\n", err)
		return
	}

	fmt.Print("🔄 Enter replacement text (leave empty to remove matched text): ")
	if !ftm.scanner.Scan() {
		return
	}
	replacement := ftm.scanner.Text() // 允许空字符串

	operation := ProcessingOperation{
		Type:        "custom",
		Name:        "custom_regex",
		Pattern:     pattern,
		Replacement: replacement,
		Description: description,
	}

	processor := fieldProcessors[fieldName]
	processor.Operations = append(processor.Operations, operation)
	fmt.Printf("✅ Added custom operation: %s\n", description)
}

// 清空字段操作
func (ftm *FieldToolsManager) clearFieldOperations(fieldName string) {
	processor := fieldProcessors[fieldName]
	if len(processor.Operations) == 0 {
		fmt.Println("ℹ️ No operations to clear")
		return
	}

	processor.Operations = []ProcessingOperation{}
	fmt.Printf("✅ Cleared all operations for %s\n", getFieldDisplayName(fieldName))
}

// 测试字段操作
func (ftm *FieldToolsManager) testFieldOperations(fieldName string) {
	processor := fieldProcessors[fieldName]

	if len(processor.Operations) == 0 {
		fmt.Println("❌ No operations configured to test")
		return
	}

	fmt.Printf("\n🧪 Testing operations for %s\n", getFieldDisplayName(fieldName))
	fmt.Print("📝 Enter sample data: ")
	if !ftm.scanner.Scan() {
		return
	}

	sampleData := ftm.scanner.Text()
	result := applyFieldProcessing(sampleData, fieldName)

	fmt.Printf("\n📊 Test Results:\n")
	fmt.Printf("  Input:  '%s'\n", sampleData)
	fmt.Printf("  Output: '%s'\n", result)

	if sampleData != result {
		fmt.Println("✅ Operations applied successfully")
	} else {
		fmt.Println("ℹ️ No changes made (operations may not match input)")
	}
}
