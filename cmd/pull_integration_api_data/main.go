package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	google_ads "wm-func/cmd/pull_integration_api_data/google-ads"
	"wm-func/wm_account"
)

const (
	GoogleAdsPlatform = "googleAds"
	DefaultDays       = 90
	ENV_TEST          = "TEST"
)

// 配置结构
type Config struct {
	Platform      string
	Days          int
	Environment   string
	TestTenantIDs []int64
}

func main() {
	// 解析命令行参数
	config := parseFlags()

	log.Printf("开始拉取集成API数据...")
	log.Printf("平台: %s, 天数: %d, 环境: %s", config.Platform, config.Days, config.Environment)

	// 计算时间范围
	endTime := time.Now().UTC()
	startTime := endTime.AddDate(0, 0, -config.Days)

	log.Printf("时间范围: %s 到 %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	switch config.Platform {
	case GoogleAdsPlatform:
		if err := pullGoogleAdsData(startTime, endTime, config); err != nil {
			log.Fatalf("拉取Google Ads数据失败: %v", err)
		}
	default:
		log.Fatalf("不支持的平台: %s", config.Platform)
	}

	log.Printf("数据拉取完成!")
}

// pullGoogleAdsData 拉取Google Ads数据
func pullGoogleAdsData(startTime, endTime time.Time, config Config) error {
	log.Printf("开始获取Google Ads账户连接信息...")

	// 获取Google Ads账户
	accounts := wm_account.GetAccountsWithPlatform(GoogleAdsPlatform)
	if len(accounts) == 0 {
		return fmt.Errorf("未找到任何Google Ads账户连接")
	}

	log.Printf("找到 %d 个Google Ads账户", len(accounts))

	// 转换为Google Ads所需的连接格式
	var connections []google_ads.Connection
	for _, acc := range accounts {
		if acc.TenantId != 150180 {
			continue
		}
		conn := google_ads.Connection{
			TenantID: acc.TenantId,
			Accounts: []google_ads.Account{
				{
					ID:     acc.AccountId,
					Cipher: acc.Cipher,
				},
			},
			Tokens: google_ads.Tokens{
				RefreshToken: acc.RefreshToken,
			},
		}
		connections = append(connections, conn)
	}

	// 调用Google Ads数据拉取
	return google_ads.GoogleAdsPullData(connections, startTime, endTime, config.Environment, config.TestTenantIDs)
}

// parseFlags 解析命令行参数
func parseFlags() Config {
	var config Config

	// 定义命令行参数
	flag.StringVar(&config.Platform, "platform", GoogleAdsPlatform, "要拉取数据的平台 (google_ads)")
	flag.IntVar(&config.Days, "days", DefaultDays, "拉取最近几天的数据")
	flag.StringVar(&config.Environment, "env", "PROD", "运行环境 (TEST/PROD)")

	var testTenantIDsStr string
	flag.StringVar(&testTenantIDsStr, "test-tenants", "", "测试环境的租户ID列表，逗号分隔")

	flag.Parse()

	// 解析测试租户ID
	if testTenantIDsStr != "" {
		config.TestTenantIDs = parseTestTenantIDs(testTenantIDsStr)
	}

	// 从环境变量读取配置（优先级高于命令行参数）
	if envDays := os.Getenv("PULL_DATA_DAYS"); envDays != "" {
		if days, err := strconv.Atoi(envDays); err == nil {
			config.Days = days
		}
	}

	if envPlatform := os.Getenv("PULL_DATA_PLATFORM"); envPlatform != "" {
		config.Platform = envPlatform
	}

	if envEnvironment := os.Getenv("PULL_DATA_ENV"); envEnvironment != "" {
		config.Environment = envEnvironment
	}

	return config
}

// parseTestTenantIDs 解析测试租户ID字符串
func parseTestTenantIDs(str string) []int64 {
	if str == "" {
		return nil
	}

	var ids []int64
	parts := strings.Split(str, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.ParseInt(part, 10, 64); err == nil {
			ids = append(ids, id)
		} else {
			log.Printf("警告: 无法解析租户ID '%s': %v", part, err)
		}
	}

	return ids
}
