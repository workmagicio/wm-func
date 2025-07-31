package wm_account

import (
	"fmt"
	"wm-func/common/db/platform_db"
)

// Account 通用账户结构体
type Account struct {
	TenantId     int64  `json:"tenant_id"`
	AccountId    string `json:"account_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SecretToken  string `json:"secret_token"`
	Platform     string `json:"platform"` // 平台标识
}

// GetTraceId 获取账户跟踪ID
func (a Account) GetTraceId() string {
	return fmt.Sprintf("%d-%s", a.TenantId, a.AccountId)
}

// GetAccountsWithPlatform 根据平台获取账户列表
func GetAccountsWithPlatform(platform string) []Account {
	sql := fmt.Sprintf(query_account_with_platform, platform)

	var result []Account
	client := platform_db.GetDB()
	if err := client.Raw(sql).Scan(&result).Error; err != nil {
		panic(err)
	}

	return result
}

func GetAccountsWithPlatformNotNull(platform string) []Account {
	sql := fmt.Sprintf(query_account_with_platform_not_null, platform)

	var result []Account
	client := platform_db.GetDB()
	if err := client.Raw(sql).Scan(&result).Error; err != nil {
		panic(err)
	}

	return result
}

// GetFairingAccounts 获取Fairing平台账户
func GetFairingAccounts() []Account {
	return GetAccountsWithPlatform("fairing")
}
