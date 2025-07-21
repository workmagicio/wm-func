# Fairing Question & Responses 接口使用说明

## 🎯 **功能完成情况**

✅ **已完成的功能**：
- **Fairing Questions API** - 完整集成（全量同步）
- **Fairing Responses API** - 完整集成（基于 Stream Slice 的批量增量同步）  
- Airbyte 数据库表结构适配
- **双模式同步逻辑**：Questions全量 + Responses Stream Slice批量增量
- 智能变化检测机制（Questions）
- **🆕 基于 Stream Slice 的批量同步**（Responses）
- **🆕 一次任务执行处理多个时间片段**
- **🆕 可配置的 slice 大小和处理数量**
- **🆕 智能进度保存和恢复机制**
- **🆕 首次同步可拉取一年数据**（一次任务完成）
- 多实例并发支持
- 完整的测试模式

## 🔄 **同步模式详解**

### Questions API - 全量同步
- **端点**: `GET https://app.fairing.co/api/questions`
- **特性**: 一次返回所有问题配置
- **同步方式**: 全量获取，变化检测
- **触发条件**: 数据量变化时才更新
- **适用场景**: 配置类数据，更新频率低

### Responses API - 基于 Stream Slice 的批量增量同步 🆕
- **端点**: `GET https://app.fairing.co/api/responses`
- **特性**: 支持分页和时间过滤（`since`/`until`参数）
- **同步方式**: Stream Slice 批量处理，一次任务处理多个时间片段
- **分页机制**: `after` cursor + `limit` 参数
- **Stream Slice 策略**: 
  - 任务开始时创建所有需要的时间片段（类似 Airbyte 设计）
  - 一次任务循环处理多个 slice（默认最多50个）
  - 首次同步：一次任务可处理整年数据（365个slice）
  - 增量同步：智能创建缺失的时间片段
  - 进度保存：每个slice立即保存状态
- **数据一致性保障**: 
  - **每个slice立即保存状态**：避免数据重复问题
  - **原子性操作**：数据保存成功后立即更新同步状态
  - **断点续传**：任务中断后精确从上次位置继续
- **性能优化**: 
  - 减少锁竞争（一次获取锁处理多个slice）
  - 批量处理提高效率
  - 可配置的处理数量限制
- **适用场景**: 交易数据，持续增长

## 🚀 **快速开始**

### 1. **测试 Questions API（全量）**
```bash
# 进入项目目录
cd cmd/fairing

# 仅测试 Questions API
./deploy.sh test-questions

# 测试并保存数据到数据库
FAIRING_TEST_SAVE=true ./deploy.sh test-questions
```

### 2. **测试 Responses API（增量分页）**
```bash
# 完整API测试（包含增量分页演示）
./deploy.sh test

# 测试并保存数据
FAIRING_TEST_SAVE=true ./deploy.sh test
```

### 3. **生产环境运行**
```bash
# 启动多实例同步
./deploy.sh start

# 查看运行状态
./deploy.sh status

# 查看增量同步日志
./deploy.sh logs 1
```

## 📊 **API 特性对比**

| 特性 | Questions API | Responses API |
|------|---------------|---------------|
| **同步方式** | 全量同步 | 增量同步 |
| **分页支持** | ❌ 无分页 | ✅ cursor分页 |
| **时间过滤** | ❌ 不支持 | ✅ since/until |
| **响应格式** | `[]Question` | `{data: [], prev, next}` |
| **更新频率** | 检测到变化时 | 持续增量 |
| **数据特点** | 配置数据 | 交易数据 |

## 🔧 **增量同步参数**

### Responses API 查询参数
```bash
GET /api/responses?since=2024-01-01T00:00:00Z&limit=100&after=ENz45Kvbxru0B5X6kDN7Y
```

- `since`: ISO8601时间戳，获取指定时间后的数据
- `limit`: 每页大小（默认2，最大100）
- `after`: 分页游标，获取指定响应之后的数据
- `before`: 分页游标，获取指定响应之前的数据

## 📝 **配置说明**

### 环境变量配置 🆕
```bash
# 首次同步天数配置（默认365天）
export FAIRING_INITIAL_DAYS=365

# 近期数据同步天数（默认7天）
export FAIRING_RECENT_DAYS=7

# 每个 slice 的天数（默认1天）
export FAIRING_SLICE_DAYS=1

# Stream Slice 运行时配置
export FAIRING_MAX_SLICES_PER_RUN=800   # 每次最多处理800个slice
```

### 新增配置项 🆕
```go
type FairingConfig struct {
    BatchSize         int // 数据库批量插入大小: 500
    ResponsesPageSize int // Responses API每页大小: 100
    RateLimit         int // 请求频率限制: 10 req/s
    MaxPages          int // 最大页数保护: 1000
    TimeoutSecs       int // 请求超时时间: 30s
    
    // Stream Slice 相关配置
    MaxSlicesPerRun   int // 每次运行最多处理的slice数量: 800
    SliceDays         int // 每个slice的天数: 1
}
```

### 同步状态字段 🆕
```go
type FairingSyncState struct {
    Status       string     `json:"status"`        // 同步状态
    Message      string     `json:"message"`       // 状态消息
    UpdatedAt    time.Time  `json:"updated_at"`    // 更新时间
    RecordCount  int64      `json:"record_count"`  // 总记录数量
    LastSyncTime *time.Time `json:"last_sync_time"` // 最后成功同步时间点
    
    // Stream Slice 进度追踪（轻量级，不存储详细slice）
    SyncStartDate     *time.Time `json:"sync_start_date"`     // 本轮同步起始日期
    SyncEndDate       *time.Time `json:"sync_end_date"`       // 本轮同步结束日期  
    CurrentSliceDate  *time.Time `json:"current_slice_date"`  // 当前处理到的日期
    CompletedSlices   int        `json:"completed_slices"`    // 已完成片段数
    TotalSlices       int        `json:"total_slices"`        // 总片段数
    
    // 同步配置
    InitialDays       int  `json:"initial_days"`        // 首次同步天数
    IsInitialSync     bool `json:"is_initial_sync"`     // 是否为首次同步
    RecentSyncDays    int  `json:"recent_sync_days"`    // 近期数据同步天数
    SliceDays         int  `json:"slice_days"`          // 每个slice天数
}
```

**💡 轻量级设计说明**：
- ✅ **不存储详细的 slice 数组**：避免 JSON 字段过大
- ✅ **只记录关键进度信息**：起始/结束日期、当前处理位置、完成数量
- ✅ **运行时动态计算**：根据当前状态动态生成需要处理的时间范围
- ✅ **数据库友好**：状态信息小，序列化快，适合频繁保存

## 🧪 **测试流程详解**

### 1. **Questions测试输出**
```
[tenant-account-question] 开始测试 Questions API
[tenant-account-question] Questions API响应成功，获取数量: 15
[tenant-account-question] 成功处理 15 条 question 数据
  - 问题ID: 13
  - 问题内容: How did you hear about us?
  - 问题类型: single_response
```

### 2. **Responses测试输出**
```
[tenant-account-response] 开始测试 Responses API
[tenant-account-response] Responses API响应成功，本页数量: 100
[tenant-account-response] 成功获取第一页数据，共 100 条
[tenant-account-response] 有下一页数据: https://app.fairing.co/api/responses?after=ENz45Kvbxru0B5X6kDN7Y
  - 响应ID: ENz45Kvbxru0B5X6kDN7Y
  - 问题: How did you hear about us?
  - 回答: Podcast
  - 客户ID: 3445695217710
  - 订单总额: 57.71
```

## 📈 **生产环境监控**

### Questions同步日志
```
[tenant-account-question] 开始全量同步Questions数据
[tenant-account-question] Questions数据无变化，跳过保存。记录数: 15
[tenant-account-question] Questions全量同步完成，共15条记录
```

### Responses增量同步日志
```
[tenant-account-response] 开始增量同步Responses数据
[tenant-account-response] 增量同步起始时间: 2024-01-15T10:30:00Z
[tenant-account-response] 第1页处理完成，本页100条，累计100条
[tenant-account-response] 第2页处理完成，本页85条，累计185条
[tenant-account-response] 已获取所有页面，共2页
[tenant-account-response] Responses增量同步完成，共185条新记录
```

## 🔍 **故障排查**

### 1. **Responses分页问题**
```
错误: 解析next URL失败
解决: 检查API返回的next字段格式
```

### 2. **增量同步时间问题**
```
错误: since参数格式错误
解决: 确保时间格式为 RFC3339 (2006-01-02T15:04:05Z07:00)
```

### 3. **数据重复问题**
```
问题: Responses数据重复插入
解决: 检查 _airbyte_raw_id 唯一键约束
```

## ⚡ **性能优化**

### 增量同步优化
1. **合理的分页大小**: 100条/页（API最大限制）
2. **时间窗口控制**: 基于LastSyncTime精确增量
3. **分页保护**: 最大1000页防止无限循环
4. **速率限制**: 10 req/s避免API限制

### 数据库优化
1. **批量插入**: 500条/批次
2. **冲突处理**: 基于复合主键的 ON CONFLICT
3. **索引优化**: wm_tenant_id + _airbyte_raw_id

## ✅ **验收标准**

- [x] Questions API 全量同步正常
- [x] Responses API 增量分页同步正常
- [x] 时间戳过滤功能正确
- [x] 分页机制工作正常
- [x] 数据能成功保存到对应表
- [x] 多实例并发无冲突
- [x] 完整的错误处理和重试机制
- [x] 通过测试模式验证所有功能

---

**🎉 Fairing Questions & Responses 接口完全可用！支持不同数据特性的最优同步策略！** 