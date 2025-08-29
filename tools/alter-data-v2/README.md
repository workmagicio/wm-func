# Alter Data V2 API

数据差异分析API服务，用于分析不同平台的租户数据差异。

## 功能特性

- 🔍 **数据差异分析**: 对比API数据和概览数据的差异
- 👥 **租户分组**: 按注册时间分为新租户(<30天)和老租户(≥30天)
- 📊 **30天统计**: 每个租户最近30天的数据差异累计
- 🕒 **数据时效性**: 返回数据最后更新时间
- 💾 **缓存机制**: 支持缓存刷新控制

## 启动服务

```bash
# 安装依赖
go mod tidy

# 启动服务(默认端口8080)
go run main.go

# 指定端口启动
go run main.go -port=8081
```

## API接口

### 获取数据差异分析

**GET** `/api/alter-data`

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| platform | string | ✅ | 平台名称 | `googleAds`, `facebookMarketing`, `tiktokMarketing` |
| needRefresh | bool | ❌ | 是否刷新缓存 | `true` / `false` (默认: `false`) |

#### 请求示例

```bash
# 获取Google Ads数据(使用缓存)
curl "http://localhost:8080/api/alter-data?platform=googleAds"

# 获取Google Ads数据(强制刷新)
curl "http://localhost:8080/api/alter-data?platform=googleAds&needRefresh=true"
```

#### 响应格式

```json
{
  "success": true,
  "data": {
    "new_tenants": [
      {
        "tenant_id": 123456,
        "last_30_day_diff": 1500,
        "date_sequence": [
          {
            "date": "2024-01-01",
            "api_data": 1000,
            "data": 1100
          }
        ],
        "tags": ["高价值客户"]
      }
    ],
    "old_tenants": [
      {
        "tenant_id": 789012,
        "last_30_day_diff": -500,
        "date_sequence": [...],
        "tags": ["稳定客户"]
      }
    ],
    "data_last_load_time": "2024-01-15T10:30:00Z"
  },
  "message": "获取数据成功"
}
```

#### 响应字段说明

- `new_tenants`: 新租户列表(注册<30天)
- `old_tenants`: 老租户列表(注册≥30天，按diff逆序排序)
- `tenant_id`: 租户ID
- `last_30_day_diff`: 最近30天的数据差异累计 (Data - ApiData)
- `date_sequence`: 每日数据序列
- `data_last_load_time`: 数据最后加载时间
- `tags`: 租户标签

### 健康检查

**GET** `/health`

```bash
curl "http://localhost:8080/health"
```

```json
{
  "status": "ok", 
  "message": "alter-data-v2 service is running"
}
```

## 数据说明

### 数据差异计算
- **基准数据**: ApiData (从integration_api_data_view获取)
- **对比数据**: Data (从概览数据获取)
- **差异公式**: `diff = Data - ApiData`

### 租户分组规则
- **新租户**: 注册时间 > 当前时间-30天
- **老租户**: 注册时间 ≤ 当前时间-30天
- **排序**: 老租户按30天diff降序排列

### 时间范围
- **数据范围**: 最近90天
- **统计范围**: 最近30天

## 错误处理

### 400 参数错误
```json
{
  "success": false,
  "message": "参数错误: Key: 'GetAlterDataRequest.Platform' Error:Field validation for 'Platform' failed on the 'required' tag"
}
```

### 500 服务器错误
```json
{
  "success": false,
  "message": "服务器内部错误"
}
```

## 开发调试

```bash
# 运行测试
go test ./...

# 查看日志
go run main.go -port=8080
```

## 注意事项

1. **缓存机制**: 默认使用缓存数据，可通过`needRefresh=true`强制刷新
2. **数据一致性**: 以ApiData为基准，计算与概览数据的差异
3. **性能考虑**: 大量数据时建议使用缓存，避免频繁刷新
4. **时区**: 所有时间使用UTC时区
