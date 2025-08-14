# Pinterest 广告报告 API

这个模块实现了Pinterest广告报告API的完整功能，包括创建报告、查询报告状态和下载报告数据。

## 主要功能

### 1. 创建报告 (CreateReport)
- 创建Pinterest广告报告请求
- 支持指定时间范围
- 支持CAMPAIGN级别的数据
- 返回报告token用于后续查询

### 2. 查询报告状态 (QueryReport)
- 轮询查询报告生成状态
- 自动等待报告完成
- 返回下载链接和文件大小
- 支持超时处理

### 3. 下载报告数据 (DownloadReport)
- 从URL下载报告文件
- 直接解析JSON格式数据
- 返回结构化的ReportData数组

### 4. 完整流程 (ProcessReportData)
- 一站式报告处理流程
- 自动完成创建->查询->下载的完整流程
- 包含完整的错误处理和日志记录

## 数据结构

### ReportRequest
```go
type ReportRequest struct {
    StartDate    string   `json:"start_date"`    // 开始日期 (YYYY-MM-DD)
    EndDate      string   `json:"end_date"`      // 结束日期 (YYYY-MM-DD)
    Granularity  string   `json:"granularity"`   // 数据粒度 (TOTAL/DAY)
    Columns      []string `json:"columns"`       // 数据列 (SPEND_IN_DOLLAR等)
    Level        string   `json:"level"`         // 数据级别 (CAMPAIGN/AD_GROUP等)
    ReportFormat string   `json:"report_format"` // 报告格式 (JSON/CSV)
}
```

### ReportResponse
```go
type ReportResponse struct {
    Token         string `json:"token"`          // 报告token
    Message       string `json:"message"`        // 响应消息
    ReportStatus  string `json:"report_status"`  // 报告状态
    URL           string `json:"url,omitempty"`  // 下载链接
    Size          int64  `json:"size,omitempty"` // 文件大小
}
```

### ReportData
```go
type ReportData struct {
    SpendInDollar float64 `json:"SPEND_IN_DOLLAR"` // 花费金额
    CampaignID    int64   `json:"CAMPAIGN_ID"`     // 广告系列ID
    Date          string  `json:"DATE"`            // 日期
}
```

## 使用示例

```go
// 创建Pinterest实例
pinterest := NewPinterest(account)

// 方式1: 使用完整流程（时间范围已固定为最近180天）
reportData, err := pinterest.ProcessReportData()
if err != nil {
    log.Printf("处理报告失败: %v", err)
    return
}

// 方式2: 分步骤处理
// 1. 创建报告（时间范围已固定为最近180天）
createResp, err := pinterest.CreateReport()
if err != nil {
    log.Printf("创建报告失败: %v", err)
    return
}

// 2. 查询报告状态
queryResp, err := pinterest.QueryReport(createResp.Token)
if err != nil {
    log.Printf("查询报告失败: %v", err)
    return
}

// 3. 下载报告数据
reportData, err := pinterest.DownloadReport(queryResp.URL)
if err != nil {
    log.Printf("下载报告失败: %v", err)
    return
}
```

## 特性

### 错误处理
- 完整的错误处理机制
- HTTP请求重试机制（最多3次）
- 详细的错误日志记录

### 日志记录
- 基于trace_id的日志追踪
- 详细的操作日志记录
- 支持调试和监控

### 自动token管理
- 自动检查token过期
- 自动刷新访问令牌
- 无需手动处理token生命周期

### 数据格式支持
- 支持JSON格式数据解析
- 直接unmarshal到结构化数据
- 简化的数据处理流程

## 配置参数

```go
const (
    PinterestAPIBase = "https://api.pinterest.com/v5"  // API基础URL
    DefaultTimeout   = 30 * time.Second                // 默认超时时间
    MaxRetries       = 3                               // 最大重试次数
    RetryDelay       = 5 * time.Second                 // 重试延迟
)
```

## 注意事项

1. **报告生成时间**: Pinterest报告生成通常需要几分钟，程序会自动轮询等待
2. **API限制**: 请注意Pinterest API的调用频率限制
3. **数据范围**: 当前配置为查询最近180天的数据（从180天前到今天的UTC时间）
4. **时间格式**: 使用UTC时间确保时区一致性，格式为YYYY-MM-DD
5. **错误处理**: 程序包含完整的错误处理，但仍需要根据具体业务需求调整
6. **日志记录**: 所有操作都有详细日志，便于调试和监控
