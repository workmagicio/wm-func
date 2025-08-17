



# 数据一致性监控看板

## 项目概述
这是一个监控看板，主要用于对比不同 tenant 的不同 platform 的数据一致性，检测 API 数据与实际广告数据之间的差异。

## 功能特性
- **多租户监控**：每个 tenant 独立显示一个图表
- **数据对比**：对比 API 消费数据 vs 广告实际消费数据
- **智能排序**：按差异值大小排序，差异最大的租户排在前面
- **可视化展示**：使用 ECharts 进行数据可视化
- **响应式布局**：一行展示 3 个图表，自适应不同屏幕尺寸
- **一体化部署**：前后端集成，运行 `main.go` 即可启动完整服务

## 技术架构

### 后端架构
```
├── main.go                 # 主程序入口，启动 HTTP 服务
├── handlers/               # HTTP 处理器
│   ├── api.go             # REST API 路由和处理逻辑
│   └── static.go          # 静态文件服务
├── platforms/             # 数据平台模块
│   ├── model.go           # 数据模型定义
│   ├── google.go          # Google Ads 数据处理
│   ├── meta.go            # Meta 数据处理 (预留)
│   ├── interface.go       # 平台接口定义
│   └── registry.go        # 平台注册管理
├── query_sql/             # SQL 查询模块
│   ├── google.go          # Google Ads 查询语句
│   └── meta.go            # Meta 查询语句 (预留)
└── static/                # 前端静态文件
    ├── index.html         # 主页面
    ├── js/
    │   ├── app.js         # 主应用逻辑
    │   ├── platform.js    # 平台切换逻辑
    │   ├── chart.js       # ECharts 图表管理
    │   └── echarts.min.js # ECharts 库
    └── css/
        └── style.css      # 样式文件
```

### 前端架构
```
Dashboard Layout:
┌─────────────────────────────────────────────────────────┐
│                    数据监控看板                           │
├─────────────────────────────────────────────────────────┤
│  Platform: [Google Ads ▼] [Meta ▼] [TikTok ▼] ...     │
├─────────────────────────────────────────────────────────┤
│  [Tenant 1 Chart]  [Tenant 2 Chart]  [Tenant 3 Chart] │
│  [Tenant 4 Chart]  [Tenant 5 Chart]  [Tenant 6 Chart] │
│  ...                                                    │
└─────────────────────────────────────────────────────────┘
```

**交互流程：**
1. 页面加载时显示默认平台 (Google Ads) 的所有租户图表
2. 用户点击平台选择器切换平台
3. 前端发送 API 请求获取新平台数据
4. 图表区域动态更新显示新平台的租户数据

**前端功能模块：**
```javascript
// platform.js - 平台管理
class PlatformManager {
  loadPlatforms()     // 加载可用平台列表
  switchPlatform()    // 切换平台时的处理逻辑
  getCurrentPlatform() // 获取当前选中平台
}

// chart.js - 图表管理  
class ChartManager {
  initChart()         // 初始化图表容器
  updateChartData()   // 更新图表数据
  createTenantChart() // 创建单个租户图表
  destroyCharts()     // 销毁现有图表
}

// app.js - 主应用逻辑
class Dashboard {
  init()              // 应用初始化
  loadPlatformData()  // 加载平台数据
  renderCharts()      // 渲染图表网格
  handlePlatformChange() // 处理平台切换事件
}
```

### API 接口设计
```
GET /api/platforms        # 获取所有可用平台列表
GET /api/tenants          # 获取所有租户列表  
GET /api/data/{platform}  # 获取指定平台所有租户的数据
GET /api/data/{platform}/{tenant_id}  # 获取指定租户的平台数据
```

**API 响应示例：**
```json
// GET /api/platforms
{
  "success": true,
  "data": [
    {"name": "google", "display_name": "Google Ads"},
    {"name": "meta", "display_name": "Meta Ads"},
    {"name": "tiktok", "display_name": "TikTok Ads"}
  ]
}

// GET /api/data/google
{
  "success": true,
  "platform": "google",
  "data": [
    {
      "tenant_id": 1001,
      "tenant_name": "Tenant A", 
      "date_range": ["2024-01-01", "2024-01-02", "..."],
      "api_spend": [1000, 1200, ...],
      "ad_spend": [980, 1180, ...],
      "difference": [20, 20, ...]
    }
  ]
}
```

### 智能排序算法
系统会自动计算每个租户的总差异值，并按以下规则排序：

1. **差异值计算**：对每个租户的所有日期差异值取绝对值求和
2. **排序规则**：差异值越大的租户排在越前面
3. **实时更新**：每次数据刷新时自动重新排序

**示例排序结果**：
- Tenant 150181: 总差异 4,441,802 (排第1)
- Tenant 150179: 总差异 4,441,705 (排第2)  
- Tenant 133857: 总差异 4,365,577 (排第3)

### 数据模型
```go
// 平台信息
type PlatformInfo struct {
    Name        string `json:"name"`         // 平台标识: google, meta, tiktok
    DisplayName string `json:"display_name"` // 显示名称: Google Ads, Meta Ads
}

// 租户数据
type TenantData struct {
    TenantID   int64    `json:"tenant_id"`
    TenantName string   `json:"tenant_name"`
    Platform   string   `json:"platform"`
    DateRange  []string `json:"date_range"`
    APISpend   []int64  `json:"api_spend"`
    AdSpend    []int64  `json:"ad_spend"`
    Difference []int64  `json:"difference"`
}

// API 响应格式
type PlatformResponse struct {
    Success bool           `json:"success"`
    Data    []PlatformInfo `json:"data"`
    Message string         `json:"message"`
}

type DashboardResponse struct {
    Success  bool         `json:"success"`
    Platform string       `json:"platform"`
    Data     []TenantData `json:"data"`
    Message  string       `json:"message"`
}
```

### ECharts 图表配置
- **图表类型**：折线图，显示时间趋势
- **时间排序**：前端自动按日期递增排序，确保时间轴正确显示
- **时间标记**：自动显示30天前标记线（橙色虚线，带"30天前"标签）
- **图表尺寸**：桌面端420px高度，移动端自适应，确保日期标签完整显示
- **数据系列**：
  - API 消费数据 (蓝色线)
  - 广告实际消费 (红色线)  
  - 差异值 (橙色柱状图)
- **交互功能**：
  - 鼠标悬停显示详细数值
  - 图例点击控制系列显示/隐藏
  - 数据缩放功能

## 扩展设计

### 平台接口 (Interface)
```go
type Platform interface {
    GetName() string
    GetTenantData(tenantID int64, days int) ([]AlterData, error)
    GetAllTenantsData(days int) (map[int64][]AlterData, error)
}
```

### 平台注册机制
```go
// 支持动态注册新平台
func RegisterPlatform(name string, platform Platform)
func GetPlatform(name string) (Platform, error)
func GetAllPlatforms() map[string]Platform
```

## 部署方式
```bash
# 启动服务
go run main.go

# 服务将在 http://localhost:8080 启动
# 前端页面：http://localhost:8080
# API 端点：http://localhost:8080/api/*
```

## 开发计划
1. ✅ 数据模型设计
2. ✅ Google Ads 数据查询
3. 🚧 HTTP 服务和 API 接口
4. 🚧 平台接口和注册机制
5. 🚧 前端页面和 ECharts 集成  
6. 🚧 平台选择器功能
7. 🚧 多租户数据展示
8. ⏳ Meta 平台预留接口
9. ⏳ 其他平台扩展机制

**开发优先级：**
- **Phase 1**: 后端 API 服务 (platforms接口、HTTP handlers)
- **Phase 2**: 前端基础框架 (HTML、平台选择器、图表容器)  
- **Phase 3**: 数据可视化 (ECharts 集成、动态更新)
- **Phase 4**: 功能完善 (错误处理、加载状态、响应式布局)
