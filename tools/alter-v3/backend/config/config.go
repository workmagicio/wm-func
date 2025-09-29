package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var config_path = "./config.json"

type Config struct {
	Name           string `json:"name"`
	BasePlatform   string `json:"base_platform"`
	ApiSql         string `json:"api_data_query"`
	WmSql          string `json:"wm_data_query"`
	Icon           string `json:"icon"`
	TotalDataCount int    `json:"total_data_count"`
}

func GetConfit() map[string]Config {
	b, err := os.ReadFile(config_path)
	if err != nil {
		panic(err)
	}
	configs := []Config{}
	if err = json.Unmarshal(b, &configs); err != nil {
		panic(err)
	}

	// sql check - 过滤掉包含不安全SQL的配置
	configs = filterSafeConfigs(configs)

	var result = map[string]Config{}
	for _, config := range configs {
		if config.TotalDataCount < 45 {
			config.TotalDataCount = 75
		}

		result[config.Name] = config
	}

	return result
}

// filterSafeConfigs 过滤掉包含不安全SQL的配置项
func filterSafeConfigs(configs []Config) []Config {
	// 定义不允许的SQL关键词（修改操作）
	forbiddenKeywords := []string{
		"DELETE", "UPDATE", "INSERT", "DROP", "CREATE", "ALTER",
		"TRUNCATE", "REPLACE", "MERGE", "UPSERT", "EXEC", "EXECUTE",
		"CALL", "SET", "GRANT", "REVOKE", "COMMIT", "ROLLBACK",
	}

	var safeConfigs []Config

	for _, config := range configs {
		// 检查 ApiSql 是否安全
		if !isSafeSQLStatement(config.ApiSql, forbiddenKeywords) {
			fmt.Printf("警告: 配置 '%s' 的 api_data_query 包含不安全的SQL语句，已跳过\n", config.Name)
			continue
		}

		// 检查 WmSql 是否安全
		if !isSafeSQLStatement(config.WmSql, forbiddenKeywords) {
			fmt.Printf("警告: 配置 '%s' 的 wm_data_query 包含不安全的SQL语句，已跳过\n", config.Name)
			continue
		}

		// 两个SQL都安全，保留此配置
		safeConfigs = append(safeConfigs, config)
	}

	return safeConfigs
}

// isSafeSQLStatement 检查SQL语句是否安全（只允许查询操作）
func isSafeSQLStatement(sql string, forbiddenKeywords []string) bool {
	if sql == "" {
		return true // 空SQL认为是安全的
	}

	// 将SQL转换为大写进行检查
	upperSQL := strings.ToUpper(sql)

	// 移除注释和多余空格
	cleanSQL := cleanSQLStatement(upperSQL)

	// 检查是否包含禁止的关键词
	for _, keyword := range forbiddenKeywords {
		// 使用正则表达式确保关键词是完整的单词，而不是其他单词的一部分
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(keyword))
		matched, err := regexp.MatchString(pattern, cleanSQL)
		if err != nil {
			// 正则表达式错误，为安全起见返回false
			return false
		}

		if matched {
			return false // 包含禁止关键词
		}
	}

	// 额外检查：确保SQL以SELECT开头（忽略空格和注释）
	if !isSelectStatement(cleanSQL) {
		return false // 不是SELECT语句
	}

	return true // 通过所有检查，认为是安全的
}

// cleanSQLStatement 清理SQL语句，移除注释和多余空格
func cleanSQLStatement(sql string) string {
	// 移除单行注释 --
	lines := strings.Split(sql, "\n")
	var cleanLines []string
	for _, line := range lines {
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	cleanSQL := strings.Join(cleanLines, " ")

	// 移除多行注释 /* */
	re := regexp.MustCompile(`/\*.*?\*/`)
	cleanSQL = re.ReplaceAllString(cleanSQL, "")

	// 移除多余空格
	re = regexp.MustCompile(`\s+`)
	cleanSQL = re.ReplaceAllString(cleanSQL, " ")

	return strings.TrimSpace(cleanSQL)
}

// isSelectStatement 检查SQL是否以SELECT开头
func isSelectStatement(sql string) bool {
	if sql == "" {
		return false
	}

	// 检查是否以SELECT开头（可能前面有WITH子句）
	return regexp.MustCompile(`^(WITH\s+.*?\s+)?SELECT\s+`).MatchString(sql)
}

func AddConfig(cfg Config) {
	b, err := os.ReadFile(config_path)
	if err != nil {
		panic(err)
	}

	configs := []Config{}
	if err = json.Unmarshal(b, &configs); err != nil {
		panic(err)
	}

	// 查找是否存在相同name的配置
	found := false
	for i, config := range configs {
		if config.Name == cfg.Name {
			// 如果找到相同name，替换原配置
			configs[i] = cfg
			found = true
			break
		}
	}

	// 如果没有找到相同name，追加新配置
	if !found {
		configs = append(configs, cfg)
	}

	configs = filterSafeConfigs(configs)

	bt, err := json.Marshal(configs)
	if err != nil {
		panic(err)
	}

	os.WriteFile(config_path, bt, os.ModePerm)
}

// RemoveConfig 根据配置名称移除配置项
func RemoveConfig(name string) {
	b, err := os.ReadFile(config_path)
	if err != nil {
		panic(err)
	}

	configs := []Config{}
	if err = json.Unmarshal(b, &configs); err != nil {
		panic(err)
	}

	// 过滤掉指定名称的配置
	var filteredConfigs []Config
	for _, config := range configs {
		if config.Name != name {
			filteredConfigs = append(filteredConfigs, config)
		}
	}

	bt, err := json.Marshal(filteredConfigs)
	if err != nil {
		panic(err)
	}

	os.WriteFile(config_path, bt, os.ModePerm)
}

func GetAllConfig() []Config {
	b, err := os.ReadFile(config_path)
	if err != nil {
		panic(err)
	}

	configs := []Config{}
	if err = json.Unmarshal(b, &configs); err != nil {
		panic(err)
	}

	return configs
}
