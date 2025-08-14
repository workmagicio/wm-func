package main

type DailyReport struct {
	TenantId              int64   `gorm:"column:tenant_id"`
	EntityId              string  `gorm:"column:entity_id"`
	Asin                  string  `gorm:"column:asin"`
	ShipToZipCode         string  `gorm:"column:ship_to_zip_code"`
	ShipToCountryCode     string  `gorm:"column:ship_to_country_code"`
	ShipToCity            string  `gorm:"column:ship_to_city"`
	ShipToStateOrProvince string  `gorm:"column:ship_to_state_or_province"`
	StatDate              string  `gorm:"column:stat_date"`
	ShippedRevenue        float64 `gorm:"column:shipped_revenue"`
	ShippedUnits          int64   `gorm:"column:shipped_units"`
	SalesDiscount         float64 `gorm:"column:sales_discount"`
	ShippedCogs           float64 `gorm:"column:shipped_cogs"`
	ContraCogs            float64 `gorm:"column:contra_cogs"`
}

func (r *DailyReport) TableName() string {
	return "platform_offline.amazon_vendor_zip_code_daily_report"
}

// 新增：公共金额结构体及其衍生结构体，后续 Response 将引用

type Value struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}

type ShippedUnitsWithRevenue struct {
	Units int   `json:"units"`
	Value Value `json:"value"`
}

type ShippedOrdersTotals struct {
	ShippedUnitsWithRevenue ShippedUnitsWithRevenue `json:"shippedUnitsWithRevenue"`
	AverageSellingPrice     Value                   `json:"averageSellingPrice"`
}

type Totals struct {
	ShippedOrders ShippedOrdersTotals `json:"shippedOrders"`
}

type ShippedOrdersMetrics struct {
	ShippedUnitsWithRevenue ShippedUnitsWithRevenue `json:"shippedUnitsWithRevenue"`
}

type Costs struct {
	SalesDiscount Value `json:"salesDiscount"`
	ShippedCogs   Value `json:"shippedCogs"`
	ContraCogs    Value `json:"contraCogs"`
}

type MetricsData struct {
	ShippedOrders ShippedOrdersMetrics `json:"shippedOrders"`
	Costs         Costs                `json:"costs"`
}

type GroupByKey struct {
	ShipToZipCode         string `json:"shipToZipCode"`
	Asin                  string `json:"asin"`
	ShipToCountryCode     string `json:"shipToCountryCode"`
	ShipToCity            string `json:"shipToCity"`
	ShipToStateOrProvince string `json:"shipToStateOrProvince"`
}

type Metric struct {
	GroupByKey GroupByKey  `json:"groupByKey"`
	Metrics    MetricsData `json:"metrics"`
}

type DocumentResponse struct {
	StartDate     string   `json:"startDate"`
	EndDate       string   `json:"endDate"`
	MarketplaceId string   `json:"marketplaceId"`
	Totals        Totals   `json:"totals"`
	Metrics       []Metric `json:"metrics"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type CreateReportResponse struct {
	QueryId string `json:"queryId"`
}

type QueryStatusResponse struct {
	ProcessingStatus string `json:"processingStatus"`
	DataDocumentId   string `json:"dataDocumentId"`
}

type QueryDocumentResponse struct {
	DocumentUrl string `json:"documentUrl"`
}
