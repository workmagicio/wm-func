package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ReadFileData 根据文件后缀读取CSV或Excel文件
// 返回二维字符串数组，第一行为表头
func ReadFileData(filename string) ([][]string, error) {
	ext := strings.ReplaceAll(strings.ToLower(filepath.Ext(filename)), "'", "")

	switch ext {
	case ".csv":
		return readCSV(filename)
	case ".xlsx", ".xls":
		return readExcel(filename)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// ReadFileDataWithLimit 根据文件后缀读取CSV或Excel文件的前N行数据
// maxRows: 最大读取行数（包含表头）
// 返回二维字符串数组，第一行为表头
func ReadFileDataWithLimit(filename string, maxRows int) ([][]string, error) {
	ext := strings.ReplaceAll(strings.ToLower(filepath.Ext(filename)), "'", "")

	switch ext {
	case ".csv":
		return readCSVWithLimit(filename, maxRows)
	case ".xlsx", ".xls":
		return readExcelWithLimit(filename, maxRows)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s", ext)
	}
}

// readCSV 读取CSV文件
func readCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 设置LazyQuotes为true，允许裸引号
	reader.LazyQuotes = true
	// 设置TrimLeadingSpace为true，去除前导空格
	reader.TrimLeadingSpace = true
	// 设置FieldsPerRecord为-1，允许变长记录
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV文件失败: %v", err)
	}

	return records, nil
}

// readExcel 读取Excel文件
func readExcel(filename string) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表名称
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	// 读取工作表数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取Excel工作表失败: %v", err)
	}

	return rows, nil
}

// readCSVWithLimit 读取CSV文件的前N行数据
func readCSVWithLimit(filename string, maxRows int) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 设置LazyQuotes为true，允许裸引号
	reader.LazyQuotes = true
	// 设置TrimLeadingSpace为true，去除前导空格
	reader.TrimLeadingSpace = true
	// 设置FieldsPerRecord为-1，允许变长记录
	reader.FieldsPerRecord = -1

	var records [][]string

	for i := 0; i < maxRows; i++ {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break // 文件结束
			}
			return nil, fmt.Errorf("读取CSV文件失败: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}

// readExcelWithLimit 读取Excel文件的前N行数据
func readExcelWithLimit(filename string, maxRows int) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表名称
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("Excel文件中没有工作表")
	}

	var rows [][]string

	// 逐行读取，只读取前maxRows行
	for row := 1; row <= maxRows; row++ {
		var rowData []string
		hasData := false

		// 动态确定列数，从A列开始逐列读取直到遇到空列
		col := 1
		consecutiveEmptyCells := 0
		maxEmptyCells := 10 // 连续10个空单元格就认为行结束

		for {
			cellName, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				break
			}

			cellValue, err := f.GetCellValue(sheetName, cellName)
			if err != nil {
				cellValue = ""
			}

			if cellValue == "" {
				consecutiveEmptyCells++
				if consecutiveEmptyCells >= maxEmptyCells {
					// 移除末尾的空单元格
					for len(rowData) > 0 && rowData[len(rowData)-1] == "" {
						rowData = rowData[:len(rowData)-1]
					}
					break
				}
			} else {
				hasData = true
				consecutiveEmptyCells = 0
			}

			rowData = append(rowData, cellValue)
			col++

			// 防止无限循环，最多读取1000列
			if col > 1000 {
				break
			}
		}

		// 如果这一行没有任何数据，可能已经到达文件末尾
		if !hasData && row > 1 {
			break
		}

		rows = append(rows, rowData)
	}

	return rows, nil
}
