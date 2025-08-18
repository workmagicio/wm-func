package main

// 字段映射请求
type FieldMappingRequest struct {
	Filename string
	Data     [][]string
}

// 字段映射响应
type FieldMappingResponse struct {
	DateType          string `json:"date_type"`
	DateCode          string `json:"date_code"`
	GeoType           string `json:"geo_type"`
	GeoCode           string `json:"geo_code"`
	GeoName           string `json:"geo_name"`
	SalesPlatform     string `json:"sales_platform"`
	SalesPlatformType string `json:"sales_platform_type"`
	CountryCode       string `json:"country_code"`
	OrderID           string `json:"order_id"`
	Orders            string `json:"orders"`
	Sales             string `json:"sales"`
	Profit            string `json:"profit"`
	NewCustomerOrders string `json:"new_customer_orders"`
	NewCustomerSales  string `json:"new_customer_sales"`
	DataStartRow      int    `json:"data_start_row"`
	HeaderRow         int    `json:"header_row"`
}

// Gemini API请求结构
type GeminiAPIRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

// Gemini API响应结构
type GeminiAPIResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// 内容结构
type Content struct {
	Parts []Part `json:"parts"`
}

// 部分结构
type Part struct {
	Text string `json:"text"`
}

// 候选结构
type Candidate struct {
	Content Content `json:"content"`
}

// 生成配置
type GenerationConfig struct {
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   ResponseSchema `json:"responseSchema"`
}

// 响应Schema结构
type ResponseSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// 属性结构
type Property struct {
	Type        string              `json:"type"`
	Description string              `json:"description,omitempty"`
	Enum        []string            `json:"enum,omitempty"`
	Items       *Property           `json:"items,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
}

type ResponseJson struct {
	CountryCode       string `json:"country_code"`
	DataStartRow      int    `json:"data_start_row"`
	DateCode          string `json:"date_code"`
	DateType          string `json:"date_type"`
	GeoCode           string `json:"geo_code"`
	GeoName           string `json:"geo_name"`
	GeoType           string `json:"geo_type"`
	HeaderRow         int    `json:"header_row"`
	NewCustomerOrders string `json:"new_customer_orders"`
	NewCustomerSales  string `json:"new_customer_sales"`
	OrderID           string `json:"order_id"`
	Orders            string `json:"orders"`
	Profit            string `json:"profit"`
	Sales             string `json:"sales"`
	SalesPlatform     string `json:"sales_platform"`
	SalesPlatformType string `json:"sales_platform_type"`
}

// 字段处理工具配置
type FieldProcessor struct {
	FieldName  string                `json:"field_name"`
	Operations []ProcessingOperation `json:"operations"`
}

// 处理操作
type ProcessingOperation struct {
	Type        string `json:"type"`        // "predefined" 或 "regex"
	Name        string `json:"name"`        // 操作名称
	Pattern     string `json:"pattern"`     // 正则表达式模式
	Replacement string `json:"replacement"` // 替换内容
	Description string `json:"description"` // 操作描述
}

// 全局字段处理器存储
var fieldProcessors = make(map[string]*FieldProcessor)

// Group操作配置
type GroupConfig struct {
	Enabled       bool     `json:"enabled"`
	GroupByFields []string `json:"group_by_fields"` // 用于分组的字段列表
	SumFields     []string `json:"sum_fields"`      // 需要求和的字段
	FirstFields   []string `json:"first_fields"`    // 取首值的字段

	// 兼容性字段（废弃）
	OrderIDField string `json:"order_id_field,omitempty"` // 保留用于向后兼容
}

// Product level检测结果
type ProductLevelDetection struct {
	IsProductLevel  bool     `json:"is_product_level"`
	OrderIDFields   []string `json:"order_id_fields"`
	ProductIDFields []string `json:"product_id_fields"`
	SKUFields       []string `json:"sku_fields"`
	ConfidenceScore float64  `json:"confidence_score"`
}

// 所有可映射的字段列表
var allMappableFields = []string{
	"date_type", "date_code", "geo_type", "geo_code", "geo_name",
	"sales_platform", "sales_platform_type", "country_code",
	"order_id", "orders", "sales", "profit", "new_customer_orders", "new_customer_sales",
	"data_start_row",
}

// 预定义的数据清理操作
var predefinedOperations = map[string][]ProcessingOperation{
	"date_code": {
		{
			Type:        "predefined",
			Name:        "convert_date_format",
			Pattern:     "",
			Replacement: "",
			Description: "Convert various date formats to YYYY-MM-DD",
		},
		{
			Type:        "predefined",
			Name:        "convert_week_format",
			Pattern:     `^\d{6}$`,
			Replacement: "",
			Description: "Convert YYYYWW format (e.g., 202502) to week start date",
		},
		{
			Type:        "predefined",
			Name:        "convert_month_format",
			Pattern:     `^\d{6}$`,
			Replacement: "",
			Description: "Convert YYYYMM format (e.g., 202501) to month start date",
		},
		{
			Type:        "predefined",
			Name:        "extract_date_only",
			Pattern:     `\s+\d{1,2}:\d{2}:\d{2}.*$`,
			Replacement: "",
			Description: "Remove time part, keep date only",
		},
	},
	"geo_code": {
		{
			Type:        "predefined",
			Name:        "format_zip_code",
			Pattern:     `^(\d{5})(\d{4})?$`,
			Replacement: "$1",
			Description: "Format ZIP codes to 5-digit format (12345-1234 → 12345)",
		},
		{
			Type:        "predefined",
			Name:        "clean_geo_code",
			Pattern:     `[^\w\s-]`,
			Replacement: "",
			Description: "Remove special characters, keep letters, numbers, spaces, and hyphens",
		},
		{
			Type:        "predefined",
			Name:        "uppercase_state_code",
			Pattern:     `^([a-z]{2})$`,
			Replacement: "",
			Description: "Convert state codes to uppercase (ca → CA)",
		},
	},
	"sales": {
		{
			Type:        "predefined",
			Name:        "remove_currency_symbols",
			Pattern:     `[$¥€£₹₽₩¢]`,
			Replacement: "",
			Description: "Remove currency symbols ($, ¥, €, £, etc.)",
		},
		{
			Type:        "predefined",
			Name:        "remove_commas",
			Pattern:     `,`,
			Replacement: "",
			Description: "Remove thousand separators (commas)",
		},
		{
			Type:        "predefined",
			Name:        "extract_numbers",
			Pattern:     `[^\d.-]`,
			Replacement: "",
			Description: "Keep only numbers, dots, and minus signs",
		},
	},
	"profit": {
		{
			Type:        "predefined",
			Name:        "remove_currency_symbols",
			Pattern:     `[$¥€£₹₽₩¢]`,
			Replacement: "",
			Description: "Remove currency symbols ($, ¥, €, £, etc.)",
		},
		{
			Type:        "predefined",
			Name:        "remove_commas",
			Pattern:     `,`,
			Replacement: "",
			Description: "Remove thousand separators (commas)",
		},
	},
	"orders": {
		{
			Type:        "predefined",
			Name:        "remove_commas",
			Pattern:     `,`,
			Replacement: "",
			Description: "Remove thousand separators (commas)",
		},
		{
			Type:        "predefined",
			Name:        "extract_numbers",
			Pattern:     `[^\d]`,
			Replacement: "",
			Description: "Keep only numbers",
		},
	},
}

// 全局group配置
var groupConfig = &GroupConfig{
	Enabled:       false,
	GroupByFields: []string{"date_type", "date_code", "geo_type", "geo_code", "sales_platform", "sales_platform_type", "country_code"},
	SumFields:     []string{"sales", "profit", "orders"},
	FirstFields:   []string{}, // 分组字段会自动取首值，不需要额外配置
	OrderIDField:  "",         // 保留用于向后兼容
}
