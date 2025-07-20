package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm/clause"
	"log"
	"time"
	"wm-func/common/db/airbyte_db"
	"wm-func/common/http_request"
	"wm-func/common/state"
	"wm-func/wm_account"
)

// buildGraphQLQuery 构建 GraphQL 查询
func buildGraphQLQuery(syncState SyncState) string {
	config := getShopifyConfig()
	var queryParams string
	if syncState.LastCursor != "" {
		queryParams = fmt.Sprintf(`first: %d, after: "%s"`, config.PageSize, syncState.LastCursor)
	} else {
		queryParams = fmt.Sprintf("first: %d", config.PageSize)
	}
	return fmt.Sprintf(base_query, queryParams)
}

// callShopifyAPI 调用 Shopify GraphQL API
func callShopifyAPI(account wm_account.ShopifyAccount, query string) (*GraphQLResponse, error) {
	requestData := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	headers := map[string]string{
		"X-Shopify-Access-Token": account.AccessToken,
		"Content-Type":           "application/json",
	}

	url := buildShopifyURL(account.ShopDomain)

	response, err := http_request.Post(url, headers, nil, jsonData)
	if err != nil {
		return nil, fmt.Errorf("请求 Shopify API 失败: %w", err)
	}

	var gqlResponse GraphQLResponse
	if err := json.Unmarshal(response, &gqlResponse); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &gqlResponse, nil
}

func processCustomerData(account wm_account.ShopifyAccount, gqlResponse *GraphQLResponse) ([]ShopifyCustomerFirstOrder, string, error) {
	var data []ShopifyCustomerFirstOrder
	var lastCursor string

	for _, edge := range gqlResponse.Data.Customers.Edges {
		lastCursor = edge.Cursor
		data = append(data, edge.Node.TransForCustomerFirstOrder(account.TenantId))
	}

	return data, lastCursor, nil
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

// updateSyncState 更新同步状态
func updateSyncState(account wm_account.ShopifyAccount, syncInfo SyncState) error {

	stateData, err := json.Marshal(syncInfo)
	if err != nil {
		return fmt.Errorf("[%s] 序列化同步状态失败: %w", account.GetTraceId(), err)
	}

	state.SaveSyncInfo(account.TenantId, account.ShopDomain, Platform, SubType, stateData)
	log.Printf("[%s] 同步状态已更新, 最后游标: %s", account.GetTraceId(), syncInfo.LastCursor)

	return nil
}

func syncCustomers(account wm_account.ShopifyAccount, syncState SyncState) error {
	// 1. 构建 GraphQL 查询
	query := buildGraphQLQuery(syncState)

	// 2. 调用 Shopify API
	gqlResponse, err := callShopifyAPI(account, query)
	if err != nil {
		return err
	}

	// 3. 处理客户数据
	customers, lastCursor, err := processCustomerData(account, gqlResponse)
	if err != nil {
		return err
	}

	if len(customers) > 0 {
		// 4. 保存数据
		if err = saveFirstOrderCustomers(account, customers); err != nil {
			return err
		}
		syncState.LastCursor = lastCursor
	}

	hasNextPage := gqlResponse.Data.Customers.PageInfo.HasNextPage
	if hasNextPage {
		syncState.Status = STATUS_RUNNING
	} else {
		syncState.Status = STATUS_SUCCESS
	}

	syncState.UpdatedAt = time.Now().UTC().Truncate(time.Millisecond)

	// 5. 更新同步状态
	if err := updateSyncState(account, syncState); err != nil {
		return err
	}

	// 6. 如果还有更多页面，递归处理
	if hasNextPage {
		log.Printf("[%s] 还有更多页面，继续处理... %s", account.GetTraceId(), lastCursor)

		time.Sleep(time.Second * 3)

		return syncCustomers(account, syncState)
	}

	log.Printf("[%s] 所有页面处理完成", account.GetTraceId())
	return nil
}
