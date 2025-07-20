package main

import (
	"gorm.io/gorm/clause"
	"log"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/wm_account"
)

const (
	STATUS_SUCCESS = "SUCCESS"
	STATUS_RUNNING = "RUNNING"
	STATUS_FAILED  = "FAILED"
)

type ShopifyCustomerFirstOrder struct {
	TenantId   int64     `json:"tenantId"`
	CustomerId string    `json:"customerId"`
	OrderId    string    `json:"orderId"`
	Data       []byte    `json:"data"`
	CreateTime time.Time `json:"createTime"`
}

func (s *ShopifyCustomerFirstOrder) TableName() string {
	return "airbyte_destination_v2.shopify_customer_first_order"

}

func saveFirstOrderCustomers(account wm_account.ShopifyAccount, customers []ShopifyCustomerFirstOrder) error {
	if len(customers) == 0 {
		return nil
	}

	db := airbyte_db.GetDB()
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(customers, 500).Error; err != nil {
		return err
	}
	log.Printf("[%s] successfully inserted %d", account.GetTraceId(), len(customers))

	return nil
}
