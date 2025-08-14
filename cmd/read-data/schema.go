package main

// 创建响应Schema
func createResponseSchema() ResponseSchema {
	return ResponseSchema{
		Type: "object",
		Properties: map[string]Property{
			"date_type":           {Type: "string", Description: "Excel中对应的日期类型列标题"},
			"date_code":           {Type: "string", Description: "Excel中对应的日期代码列标题"},
			"geo_type":            {Type: "string", Description: "Excel中对应的地理类型列标题"},
			"geo_code":            {Type: "string", Description: "Excel中对应的地理代码列标题"},
			"geo_name":            {Type: "string", Description: "Excel中对应的地理名称列标题"},
			"sales_platform":      {Type: "string", Description: "Excel中对应的销售平台列标题"},
			"sales_platform_type": {Type: "string", Description: "Excel中对应的销售平台类型列标题"},
			"country_code":        {Type: "string", Description: "Excel中对应的国家代码列标题"},
			"orders":              {Type: "string", Description: "Excel中对应的订单数列标题"},
			"sales":               {Type: "string", Description: "Excel中对应的销售额列标题"},
			"profit":              {Type: "string", Description: "Excel中对应的利润列标题"},
			"new_customer_orders": {Type: "string", Description: "Excel中对应的新客户订单数列标题"},
			"new_customer_sales":  {Type: "string", Description: "Excel中对应的新客户销售额列标题"},
			"data_start_row":      {Type: "integer", Description: "数据开始的行号下标"},
			"header_row":          {Type: "integer", Description: "表头所在的行号下标"},
			"date_format_issues":  {Type: "boolean", Description: "日期格式是否需要转换"},
			"week_format_issues":  {Type: "boolean", Description: "是否包含YYYYWW格式需要转换"},
			"month_format_issues": {Type: "boolean", Description: "是否包含YYYYMM格式需要转换"},
			"currency_symbols":    {Type: "boolean", Description: "是否包含货币符号需要清理"},
			"number_formatting":   {Type: "boolean", Description: "数字格式是否需要清理"},
			"suggested_operations": {
				Type:        "array",
				Description: "建议自动应用的清理操作",
				Items: &Property{
					Type: "string",
				},
			},
		},
		Required: []string{},
	}
}
