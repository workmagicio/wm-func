# Alter Data 项目架构说明

## 项目概述

`alter-data` 是一个数据监控看板服务，用于对比不同广告平台的API数据和实际广告花费数据。

## 重构后的架构

### 目录结构

```
tools/alter-data/
├── main.go                              # 主入口文件
├── ARCHITECTURE.md                      # 架构说明文档
├── internal/                           # 内部实现（不对外暴露）
│   ├── config/                         # 配置层
│   │   ├── platform_config.go          # 平台配置管理
│   │   └── query_config.go             # SQL查询配置
│   ├── data/                           # 数据处理层
│   │   ├── processor.go                # 数据处理器（SQL执行、数据分组）
│   │   └── transformer.go              # 数据转换器（格式转换、排序）
│   ├── platform/                       # 平台抽象层
│   │   ├── interface.go                # 平台接口定义
│   │   ├── base_platform.go            # 通用平台实现
│   │   └── registry.go                 # 平台注册器
│   ├── service/                        # 业务逻辑层
│   │   └── dashboard_service.go        # 仪表板服务
│   └── handlers/                       # HTTP处理层
│       ├── api_handler.go              # API请求处理
│       └── static_handler.go           # 静态文件处理
├── models/                             # 数据模型（对外接口）
│   └── types.go                        # 数据类型定义
└── static/                             # 静态资源
    ├── css/
    ├── js/
    └── index.html
```

### 架构层次

#### 1. 配置层 (Config Layer)
- **platform_config.go**: 管理所有平台的基本信息（名称、显示名、是否启用等）
- **query_config.go**: 管理各平台的SQL查询语句

#### 2. 数据处理层 (Data Layer)
- **processor.go**: 负责SQL查询执行、数据获取和分组
- **transformer.go**: 负责数据格式转换、排序等业务逻辑

#### 3. 平台抽象层 (Platform Layer)
- **interface.go**: 定义平台接口规范
- **base_platform.go**: 提供通用的平台实现，通过配置驱动
- **registry.go**: 管理平台注册和获取

#### 4. 业务逻辑层 (Service Layer)
- **dashboard_service.go**: 核心业务逻辑，协调各层完成数据获取和处理

#### 5. HTTP处理层 (Handler Layer)
- **api_handler.go**: 处理API请求，数据验证和响应格式化
- **static_handler.go**: 处理静态文件服务

## 重构优势

### 1. 消除重复代码
- 原来每个平台都有独立的实现文件，现在统一使用 `BasePlatform`
- SQL查询集中管理，避免重复定义

### 2. 配置驱动
- 新增平台只需要在配置文件中添加配置，无需修改代码
- SQL查询配置化，便于维护和修改

### 3. 分层清晰
- 每一层职责明确，便于理解和维护
- 依赖关系清晰，易于测试

### 4. 扩展性强
- 添加新平台：只需在配置中新增配置项
- 修改业务逻辑：只需修改对应层的实现
- 支持自定义平台实现：可以注册自定义的Platform实现

### 5. 统一错误处理
- 集中的错误处理机制
- 统一的日志记录

## 使用方式

### 添加新平台

1. 在 `platform_config.go` 中添加平台配置
2. 在 `query_config.go` 中添加对应的SQL查询
3. 重启服务即可

### 修改平台逻辑

如果需要特殊的平台逻辑，可以：
1. 实现 `Platform` 接口
2. 使用 `platform.RegisterPlatform()` 注册自定义实现

## API 接口

- `GET /api/platforms` - 获取平台列表
- `GET /api/data/{platform}` - 获取平台数据
- `GET /api/data/{platform}/{tenant_id}` - 获取租户数据

## 配置说明

平台配置支持以下字段：
- `name`: 平台标识
- `display_name`: 显示名称  
- `query_key`: SQL查询键
- `enabled`: 是否启用
- `description`: 平台描述
