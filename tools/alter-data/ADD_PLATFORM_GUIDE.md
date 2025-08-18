# 新增平台标准流程指南

## 🎯 概述
此文档记录了在数据监控看板中新增平台的完整步骤，确保不遗漏任何配置，避免引入bug。

## ⚠️ 重要提醒
**必须按照顺序完成所有步骤，特别注意第4步的参数数量更新！**

---

## 📋 新增平台清单 (必须全部完成)

### ✅ **步骤1: 平台基础配置**
**文件:** `tools/alter-data/internal/config/platform_config.go`

**操作:** 在 `platformConfigs` 数组中添加或启用平台配置

```go
{
    Name:        "平台标识",          // 小写，无空格，用于API和URL
    DisplayName: "平台显示名称",       // 用户界面显示的名称
    QueryKey:    "平台查询键",        // 对应SQL查询的键名
    Enabled:     true,              // 必须设为true才能启用
    Description: "平台描述",         // 简短描述
},
```

### ✅ **步骤2: 单平台SQL查询配置**
**文件:** `tools/alter-data/internal/config/query_config.go`

**操作:** 在 `queryConfigs` map中添加平台专属查询

```go
"平台查询键": {
    Key:         "平台查询键",
    Name:        "查询名称",
    Description: "查询描述",
    SQL: `您的SQL查询语句`,
},
```

**SQL要求:**
- 必须返回: `tenant_id`, `raw_date`, `api_spend`, `ad_spend`
- 必须包含租户过滤: `join platform_offline.dwd_view_analytics_non_testing_tenants`
- 时间范围建议: `RAW_DATE > utc_date() - interval 90 day`

### ✅ **步骤3: 跨平台查询集成 (可选但推荐)**
**文件:** `tools/alter-data/internal/config/query_config.go`

**操作:** 在 `tenant_cross_platform_query` 中添加平台子查询

在 `all_platforms` UNION 前添加:
```sql
平台名_api as (
    select
        TENANT_ID,
        RAW_DATE,
        round(sum(字段名), 0) as spend
    from
        platform_offline.integration_api_data_view
    where RAW_PLATFORM = '平台标识'
      and RAW_DATE > utc_date() - interval 90 day
      and TENANT_ID = ?
    group by 1, 2
),
平台名_ads as (
    select
        TENANT_ID,
        event_date,
        round(sum(字段名), 0) as ad_spend
    from platform_offline.dws_view_analytics_ads_ad_level_metrics_attrs_latest
    where event_date > utc_date() - interval 90 day
      and ADS_PLATFORM = '平台名'
      and TENANT_ID = ?
    group by 1, 2
),
平台名_merge as (
    select
        平台名_api.TENANT_ID,
        '平台显示名' as platform,
        平台名_api.RAW_DATE,
        平台名_api.spend as api_spend,
        coalesce(平台名_ads.ad_spend, 0) as ad_spend
    from
        平台名_api
        left join 平台名_ads on 平台名_api.TENANT_ID = 平台名_ads.TENANT_ID 
        and 平台名_api.RAW_DATE = 平台名_ads.EVENT_DATE
),
```

然后在 `all_platforms` 的 UNION 中添加:
```sql
UNION ALL
select * from 平台名_merge
```

### ⚠️ **步骤4: 参数数量更新 (关键步骤)**
**文件:** `tools/alter-data/internal/service/dashboard_service.go`

**操作:** 更新跨平台查询的参数数量

1. **统计参数数量:**
   ```bash
   cd tools/alter-data
   grep -o "TENANT_ID = ?" internal/config/query_config.go | wc -l
   ```

2. **更新参数传递:**
   找到 `GetTenantCrossPlatformDataWithRefresh` 方法中的这段代码:
   ```go
   // 跨平台查询需要N个参数，每个平台的API和Ads查询各需要一个tenantID参数
   // Google(2) + Meta(2) + ... = N个参数
   rawData, err := processor.ExecuteQueryWithParams(sql, tenantID, tenantID, ...)
   ```
   
   **重要:** tenantID 的数量必须等于统计出的参数数量！

### ✅ **步骤5: 部署和验证**
```bash
cd tools/alter-data
./scp.sh
```

**验证步骤:**
1. 访问 `http://服务器IP:8090` 确认平台出现在选择器中
2. 测试单平台视图: `/api/data/平台标识`
3. 测试跨平台视图: `/api/tenant/134301` (应该包含新平台数据)

---

## 🚨 常见错误和解决方案

### 错误1: 400 Bad Request
**原因:** 步骤4参数数量不匹配
**解决:** 重新统计参数数量，确保 tenantID 参数个数正确

### 错误2: 平台不显示
**原因:** 步骤1中 Enabled 为 false
**解决:** 将 Enabled 设为 true

### 错误3: SQL查询失败
**原因:** 步骤2中SQL语法错误或字段名错误
**解决:** 检查SQL语句，确保字段名和表名正确

### 错误4: 跨平台数据缺失
**原因:** 步骤3未完成或UNION语句有误
**解决:** 检查跨平台查询中是否正确添加了新平台

---

## 📊 当前支持的平台列表

| 平台 | 标识 | 类型 | 状态 |
|------|------|------|------|
| Google Ads | google | 广告 | ✅ |
| Meta Ads | meta | 广告 | ✅ |
| TikTok Ads | tiktok | 广告 | ✅ |
| Snapchat Ads | snapchat | 广告 | ✅ |
| Pinterest Ads | pinterest | 广告 | ✅ |
| AppLovin | applovin | 广告 | ✅ |
| Shopify | shopify | 电商 | ✅ |
| TikTok Shop | tiktokshop | 电商 | ✅ |

**当前跨平台查询参数数量:** 16个 (8个平台 × 2个查询)

---

## 🔍 检查清单

在新增平台前，请确保:
- [ ] 已准备好平台的SQL查询语句
- [ ] 确认RAW_PLATFORM和ADS_PLATFORM的正确值
- [ ] 了解数据字段名称(AD_SPEND, ORDERS等)

在新增平台后，请验证:
- [ ] 平台显示在前端选择器中
- [ ] 单平台API正常返回数据
- [ ] 跨平台API包含新平台数据
- [ ] 无400/500错误

---

**最后更新:** 2025-08-18  
**当前版本:** 支持8个平台  
**下次新增平台时请更新此文档**
