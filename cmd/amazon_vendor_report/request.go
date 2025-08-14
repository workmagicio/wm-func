package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type AmazonVendorReportRequestClient struct {
	refreshToken string
	accessToken  string
	clientId     string
	clientSecret string
	expiresAt    time.Time
	baseUrl      string
}

func NewAmazonVendorReportRequestClient(
	refreshToken string,
) *AmazonVendorReportRequestClient {

	return &AmazonVendorReportRequestClient{
		baseUrl:      "https://api.amazon.com/auth/o2/token",
		refreshToken: refreshToken,
		clientId:     getID(),
		clientSecret: getKey(),
	}
}

func (c *AmazonVendorReportRequestClient) getAccessToken() string {
	if c.accessToken != "" && c.expiresAt.After(time.Now()) {
		return c.accessToken
	}

	requestBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": c.refreshToken,
		"client_id":     c.clientId,
		"client_secret": c.clientSecret,
	}

	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(c.baseUrl, "application/json", bytes.NewBuffer(requestBodyJson))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	res := RefreshTokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		panic(err)
	}

	c.accessToken = res.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)

	log.Printf("Obtained access token, expires in %d seconds", res.ExpiresIn)

	return c.accessToken
}

func (c *AmazonVendorReportRequestClient) getHeaders() map[string]string {
	return map[string]string{
		"x-amz-access-token": c.getAccessToken(),
	}
}

func (c *AmazonVendorReportRequestClient) createQuery(startDate string, endDate string) string {
	log.Printf("Creating query for date range %s - %s", startDate, endDate)
	queryParams := map[string]string{
		"query": fmt.Sprintf("query MyQuery {analytics_vendorAnalytics_2024_09_30 {sourcingView(startDate:\"%s\",endDate:\"%s\",aggregateBy:DAY,currencyCode:\"USD\") {startDate endDate marketplaceId totals {shippedOrders {shippedUnitsWithRevenue {units value {amount currencyCode}} averageSellingPrice {amount currencyCode}}} metrics {groupByKey {shipToCountryCode shipToCity shipToStateOrProvince shipToZipCode} metrics {shippedOrders {shippedUnitsWithRevenue {units value {amount currencyCode}}} costs {salesDiscount {amount currencyCode} shippedCogs {amount currencyCode} contraCogs {amount currencyCode}}}}}}}", startDate, endDate),
	}

	queryParamsJson, err := json.Marshal(queryParams)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://sellingpartnerapi-na.amazon.com/dataKiosk/2023-11-15/queries", bytes.NewBuffer(queryParamsJson))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-amz-access-token", c.getAccessToken())

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	code := res.StatusCode
	fmt.Println(code)
	result := CreateReportResponse{}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		panic(err)
	}

	log.Printf("Query created, id: %s", result.QueryId)
	return result.QueryId
}

func (c *AmazonVendorReportRequestClient) queryReportStatus(queryId string) QueryStatusResponse {
	log.Printf("Querying report status for id: %s", queryId)
	url := fmt.Sprintf("https://sellingpartnerapi-na.amazon.com/dataKiosk/2023-11-15/queries/%s", queryId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-amz-access-token", c.getAccessToken())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	result := QueryStatusResponse{}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		panic(err)
	}

	log.Printf("Report %s status: %s", queryId, result.ProcessingStatus)
	return result
}

func (c *AmazonVendorReportRequestClient) queryDocument(documentId string) QueryDocumentResponse {
	log.Printf("Querying document for id: %s", documentId)
	url := fmt.Sprintf("https://sellingpartnerapi-na.amazon.com/dataKiosk/2023-11-15/documents/%s", documentId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-amz-access-token", c.getAccessToken())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	result := QueryDocumentResponse{}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		panic(err)
	}

	log.Printf("Got document URL for id %s: %s", documentId, result.DocumentUrl)
	return result
}
