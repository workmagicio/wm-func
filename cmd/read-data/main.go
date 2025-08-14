package main

import (
	"fmt"
	"os"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./read-data <excel_file_path>")
		fmt.Println("Example: ./read-data /path/to/your/file.xlsx")
		fmt.Println("         ./read-data /path/to/your/file.csv")
		os.Exit(1)
	}

	// 从命令行参数获取文件路径
	filename := os.Args[1]

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("❌ File not found: %s\n", filename)
		os.Exit(1)
	}

	// 创建服务
	service := NewFieldMappingService()

	// 执行分析
	fmt.Printf("Analyzing file: %s\n", filename)

	// 一次性读取预览数据（50行）
	fileReader := NewFileReader()
	fullData, err := fileReader.ReadData(filename, PREVIEW_ROWS)
	if err != nil {
		fmt.Printf("Failed to read file data: %v\n", err)
		return
	}

	if len(fullData) == 0 {
		fmt.Println("No data found in file")
		return
	}

	// 获取原始map数据用于显示
	fmt.Println("🤖 Starting AI analysis...")
	fmt.Println("📊 Analyzing Excel structure and data patterns...")
	responseMap, err := service.AnalyzeAsMap(filename)
	if err != nil {
		fmt.Printf("❌ AI analysis failed: %v\n", err)
		return
	}
	fmt.Println("✅ AI analysis completed successfully!")

	// 创建交互式字段映射管理器并运行
	interactiveMapper := NewInteractiveFieldMapper(service)
	interactiveMapper.Run(responseMap, filename, fullData)
}

// 全局函数，用于兼容现有代码

// 预览数据映射效果
func previewDataMapping(responseMap map[string]interface{}, data [][]string) {
	dataProcessor := NewDataProcessingManager()
	dataProcessor.PreviewDataMapping(responseMap, data)
}

// 下载完整CSV文件
func downloadCSV(responseMap map[string]interface{}, filename string) {
	dataProcessor := NewDataProcessingManager()
	dataProcessor.DownloadCSV(responseMap, filename)
}

// 配置字段处理工具
func configureFieldTools() {
	fieldToolsManager := NewFieldToolsManager()
	fieldToolsManager.Configure()
}

// 配置日期格式检测
func configureDateFormatDetection(responseMap map[string]interface{}, fullData [][]string) {
	dateFormatManager := NewDateFormatManager()
	dateFormatManager.Configure(responseMap, fullData)
}
