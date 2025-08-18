package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite数据处理服务
type SQLiteService struct {
	db        *sql.DB
	dbPath    string
	tableName string
}

// 创建SQLite服务
func NewSQLiteService() (*SQLiteService, error) {
	// 创建临时数据库文件
	timestamp := time.Now().Unix()
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("field_mapping_%d.db", timestamp))

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("无法创建数据处理环境: %v", err)
	}

	service := &SQLiteService{
		db:        db,
		dbPath:    dbPath,
		tableName: "data_table",
	}

	return service, nil
}

// 关闭服务并清理临时文件
func (s *SQLiteService) Close() error {
	if s.db != nil {
		s.db.Close()
	}

	// 删除临时数据库文件
	if s.dbPath != "" {
		os.Remove(s.dbPath)
	}

	return nil
}

// 导入数据到SQLite
func (s *SQLiteService) ImportData(data [][]string, responseMap map[string]interface{}) error {
	if len(data) < 2 {
		return fmt.Errorf("数据不足，至少需要表头和一行数据")
	}

	// 获取数据起始行
	dataStartRow := getDataStartRow(responseMap)
	if dataStartRow < 1 || dataStartRow > len(data) {
		return fmt.Errorf("无效的数据起始行: %d", dataStartRow)
	}

	// 获取表头行
	headerRowIndex := getHeaderRowIndex(responseMap)
	if headerRowIndex >= len(data) {
		headerRowIndex = 0
	}
	headers := data[headerRowIndex]

	// 创建表结构
	err := s.createTable(headers, responseMap)
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 插入数据
	err = s.insertData(data, headers, responseMap, dataStartRow)
	if err != nil {
		return fmt.Errorf("插入数据失败: %v", err)
	}

	return nil
}

// 创建表结构
func (s *SQLiteService) createTable(headers []string, responseMap map[string]interface{}) error {
	var columns []string

	// 为每个映射的字段创建列
	for _, fieldName := range allMappableFields {
		if fieldName == "data_start_row" {
			continue
		}

		displayName := getFieldDisplayName(fieldName)
		columnName := s.sanitizeColumnName(displayName)

		// 根据字段类型确定SQL数据类型
		sqlType := s.getSQLType(fieldName)
		columns = append(columns, fmt.Sprintf("%s %s", columnName, sqlType))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", s.tableName, strings.Join(columns, ", "))

	fmt.Printf("📊 Creating data structure with %d columns...\n", len(columns))
	_, err := s.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("执行CREATE TABLE失败: %v", err)
	}

	return nil
}

// 插入数据
func (s *SQLiteService) insertData(data [][]string, headers []string, responseMap map[string]interface{}, dataStartRow int) error {
	// 构建INSERT语句
	var columnNames []string
	var placeholders []string

	for _, fieldName := range allMappableFields {
		if fieldName == "data_start_row" {
			continue
		}
		displayName := getFieldDisplayName(fieldName)
		columnName := s.sanitizeColumnName(displayName)
		columnNames = append(columnNames, columnName)
		placeholders = append(placeholders, "?")
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		s.tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "))

	// 准备语句
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("准备INSERT语句失败: %v", err)
	}
	defer stmt.Close()

	// 插入每一行数据
	insertedRows := 0
	for i := dataStartRow - 1; i < len(data); i++ {
		row := data[i]
		values := s.mapRowToValues(row, headers, responseMap)

		_, err := stmt.Exec(values...)
		if err != nil {
			fmt.Printf("⚠️ 跳过第%d行，插入失败: %v\n", i+1, err)
			continue
		}
		insertedRows++
	}

	fmt.Printf("✅ Successfully processed %d rows of data\n", insertedRows)
	return nil
}

// 将Excel行映射为数据库值
func (s *SQLiteService) mapRowToValues(row []string, headers []string, responseMap map[string]interface{}) []interface{} {
	var values []interface{}

	for _, fieldName := range allMappableFields {
		if fieldName == "data_start_row" {
			continue
		}

		mappedValue := getMapValue(responseMap, fieldName)
		var cellValue interface{}

		if mappedValue == "" {
			cellValue = nil // 未映射的字段为NULL
		} else {
			// 查找对应的列索引
			cleanMappedValue := cleanInferredValue(mappedValue)
			colIndex := findColumnIndex(headers, cleanMappedValue)
			if colIndex >= 0 && colIndex < len(row) {
				rawValue := row[colIndex]
				// 应用字段处理操作
				processedValue := applyFieldProcessing(rawValue, fieldName)
				cellValue = s.convertValue(processedValue, fieldName)
			} else {
				// 如果不是列映射，可能是固定值
				cellValue = s.convertValue(cleanMappedValue, fieldName)
			}
		}

		values = append(values, cellValue)
	}

	return values
}

// 转换值为适当的数据类型
func (s *SQLiteService) convertValue(value string, fieldName string) interface{} {
	if value == "" {
		return nil
	}

	// 数值字段转换为数字
	if fieldName == "sales" || fieldName == "profit" || fieldName == "orders" ||
		fieldName == "new_customer_orders" || fieldName == "new_customer_sales" {
		if num, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
			return num
		}
	}

	return strings.TrimSpace(value)
}

// 获取SQL数据类型
func (s *SQLiteService) getSQLType(fieldName string) string {
	switch fieldName {
	case "sales", "profit", "new_customer_sales":
		return "REAL"
	case "orders", "new_customer_orders", "data_start_row":
		return "INTEGER"
	default:
		return "TEXT"
	}
}

// 清理列名，确保符合SQL标准
func (s *SQLiteService) sanitizeColumnName(name string) string {
	// 替换空格和特殊字符为下划线
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ToLower(name)

	// 确保以字母开头
	if len(name) > 0 && !isLetter(name[0]) {
		name = "col_" + name
	}

	return name
}

// 检查字符是否为字母
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// 执行分组查询
func (s *SQLiteService) ExecuteGroupQuery(groupByFields []string, sumFields []string, responseMap map[string]interface{}) ([][]string, error) {
	// 构建GROUP BY查询，按照标准字段顺序
	var selectFields []string
	var groupFields []string

	// 创建字段类型映射
	groupFieldsMap := make(map[string]bool)
	sumFieldsMap := make(map[string]bool)

	for _, field := range groupByFields {
		groupFieldsMap[field] = true
	}
	for _, field := range sumFields {
		sumFieldsMap[field] = true
	}

	// 按照allMappableFields的顺序添加所有字段
	for _, fieldName := range allMappableFields {
		if fieldName == "data_start_row" {
			continue
		}

		displayName := getFieldDisplayName(fieldName)
		columnName := s.sanitizeColumnName(displayName)

		if groupFieldsMap[fieldName] {
			// 分组字段
			selectFields = append(selectFields, columnName)
			groupFields = append(groupFields, columnName)
		} else if sumFieldsMap[fieldName] {
			// 聚合字段
			if fieldName == "orders" {
				// 检查是否配置了order_id字段，如果配置了则使用COUNT(DISTINCT)
				orderIDValue := getMapValue(responseMap, "order_id")
				if orderIDValue != "" && orderIDValue != "VIRTUAL_COUNT" {
					// 获取order_id列名
					orderIDColumnName := s.sanitizeColumnName(getFieldDisplayName("order_id"))
					selectFields = append(selectFields, fmt.Sprintf("COUNT(DISTINCT %s) as %s", orderIDColumnName, columnName))
				} else {
					// 使用默认的行数统计
					selectFields = append(selectFields, fmt.Sprintf("COUNT(*) as %s", columnName))
				}
			} else {
				selectFields = append(selectFields, fmt.Sprintf("SUM(%s) as %s", columnName, columnName))
			}
		} else {
			// 其他字段，取第一个非空值
			selectFields = append(selectFields, fmt.Sprintf("MAX(%s) as %s", columnName, columnName))
		}
	}

	querySQL := fmt.Sprintf("SELECT %s FROM %s GROUP BY %s",
		strings.Join(selectFields, ", "),
		s.tableName,
		strings.Join(groupFields, ", "))

	fmt.Printf("🔍 Executing data aggregation...\n")

	rows, err := s.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("执行查询失败: %v", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列名失败: %v", err)
	}

	var result [][]string
	result = append(result, columns) // 添加表头

	// 读取数据
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("扫描行数据失败: %v", err)
		}

		var row []string
		for _, value := range values {
			if value == nil {
				row = append(row, "")
			} else {
				row = append(row, fmt.Sprintf("%v", value))
			}
		}
		result = append(result, row)
	}

	fmt.Printf("✅ Data aggregation returned %d groups (including header)\n", len(result))
	return result, nil
}

// 获取表的行数
func (s *SQLiteService) GetRowCount() (int, error) {
	var count int
	err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", s.tableName)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取行数失败: %v", err)
	}
	return count, nil
}
