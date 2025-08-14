package main

import "fmt"

// 文件读取器
type FileReader struct{}

// 创建文件读取器
func NewFileReader() *FileReader {
	return &FileReader{}
}

// 读取文件数据
func (fr *FileReader) ReadData(filename string, maxRows int) ([][]string, error) {
	fmt.Printf("正在读取文件: %s \n", filename)

	// 使用csv.go中的ReadFileDataWithLimit函数
	data, err := ReadFileDataWithLimit(filename, maxRows)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return data, nil
}
