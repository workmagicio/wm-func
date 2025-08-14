package account

import (
	"fmt"
	"wm-func/common/db/platform_db"
)

type Account struct {
	TenantId     int64  `gorm:"tenant_id"`
	Platform     string `gorm:"platform"`
	AccountId    string `gorm:"account_id"`
	AccessToken  string `gorm:"access_token"`
	RefreshToken string `gorm:"refresh_token"`
	Cipher       string `gorm:"cipher"`
}

func GetAccountsWithPlatform(platform string) []Account {
	sq := fmt.Sprintf(query_need_sync_account, platform, platform)
	res := []Account{}
	client := platform_db.GetDB()
	if err := client.Raw(sq).Scan(&res).Error; err != nil {
		panic(err)
	}
	return res
}
