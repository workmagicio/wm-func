package main

type ShopifyCustomerFirstOrder struct {
	TenantId   int64  `json:"tenantId"`
	CustomerId string `json:"customerId"`
	OrderId    string `json:"orderId"`
}
