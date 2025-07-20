package main

import (
	"encoding/json"
	"strings"
	"time"
)

type GraphQLResponse struct {
	Data struct {
		Customers struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Edges []Edge `json:"edges"`
		} `json:"customers"`
	} `json:"data"`
	Extensions struct {
		Cost struct {
			RequestedQueryCost int `json:"requestedQueryCost"`
			ActualQueryCost    int `json:"actualQueryCost"`
			ThrottleStatus     struct {
				MaximumAvailable   float64 `json:"maximumAvailable"`
				CurrentlyAvailable int     `json:"currentlyAvailable"`
				RestoreRate        float64 `json:"restoreRate"`
			} `json:"throttleStatus"`
		} `json:"cost"`
	} `json:"extensions"`
}

type Edge struct {
	Cursor string `json:"cursor"`
	Node   Node   `json:"node"`
}

type Node struct {
	Id             string  `json:"id"`
	Email          *string `json:"email"`
	NumberOfOrders string  `json:"numberOfOrders"`
	FirstOrder     struct {
		Edges []struct {
			Node struct {
				Id          string    `json:"id"`
				ProcessedAt time.Time `json:"processedAt"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"firstOrder"`
}

func (e Node) TransForCustomerFirstOrder(tenantId int64) ShopifyCustomerFirstOrder {
	cusSplit := strings.Split(e.Id, "/")
	customerId := cusSplit[len(cusSplit)-1]
	orderId := ""
	if len(e.FirstOrder.Edges) > 0 {
		odSplit := strings.Split(e.FirstOrder.Edges[0].Node.Id, "/")
		orderId = odSplit[len(odSplit)-1]
	}

	b, _ := json.Marshal(e)

	return ShopifyCustomerFirstOrder{
		TenantId:   tenantId,
		CustomerId: customerId,
		OrderId:    orderId,
		Data:       b,
		CreateTime: time.Now().Truncate(time.Microsecond),
	}

}

// 首次订单客户信息
type FirstOrderCustomer struct {
	TenantID    int64     `json:"tenant_id"`
	ShopDomain  string    `json:"shop_domain"`
	CustomerID  string    `json:"customer_id"`
	Email       string    `json:"email"`
	OrderID     string    `json:"order_id"`
	ProcessedAt time.Time `json:"processed_at"`
	CollectedAt time.Time `json:"collected_at"`
}

// 同步状态信息
type SyncState struct {
	LastCursor string    `json:"last_cursor"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GraphQL 请求数据结构
type GraphQLRequest struct {
	Query string `json:"query"`
}
