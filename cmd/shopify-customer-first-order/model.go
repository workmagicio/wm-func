package main

import "time"

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
