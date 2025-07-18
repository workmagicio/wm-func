package account

import (
	"fmt"
	"wm-func/common/db/platform_db"
)

type ShopifyAccount struct {
	TenantId    int64  `json:"tenant_id"`
	ShopDomain  string `json:"shop_domain"`
	AccessToken string `json:"access_token"`
}

func GetShopifyAccount() []ShopifyAccount {
	sql := fmt.Sprintf(query_shopify_accounts)

	var result []ShopifyAccount
	client := platform_db.GetDB()
	if err := client.Raw(sql).Scan(&result).Error; err != nil {
		panic(err)
	}
	return result
}
