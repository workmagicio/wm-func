package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"wm-func/common/http_request"
)

func TestMain2(t *testing.T) {
	requestData := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}

	headers := map[string]string{
		"X-Shopify-Access-Token": "",
		"Content-Type":           "application/json",
	}

	url := buildShopifyURL("")

	response, err := http_request.Post(url, headers, nil, jsonData)
	if err != nil {
		panic(err)
	}

	var gqlResponse GraphQLResponse
	if err := json.Unmarshal(response, &gqlResponse); err != nil {
		panic(err)
	}

	fmt.Println(gqlResponse)

}

var query = `
        query {
          order(id: "gid://shopify/Order/5687608803363") {
            customerJourneySummary {
              firstVisit {
				referrerUrl
landingPage
occurredAt
source
sourceDescription
}
                  }
                }
        }`
