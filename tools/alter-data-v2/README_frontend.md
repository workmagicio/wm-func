# Alter Data V2 前端集成说明

## 项目结构

```
alter-data-v2/
├── main.go                 # Go后端入口
├── backend/               # 后端代码
├── frontend/              # React前端代码
│   ├── src/              # 源码目录
│   │   ├── components/   # React组件
│   │   ├── pages/        # 页面组件
│   │   ├── App.tsx       # 主应用组件
│   │   └── main.tsx      # 入口文件
│   ├── package.json      # 前端依赖
│   └── vite.config.ts    # Vite构建配置
├── dist/                  # 前端构建输出(自动生成)
├── build.sh              # 构建脚本
└── start.sh              # 启动脚本
```

## 快速开始

### 1. 构建应用

```bash
# 赋予脚本执行权限
chmod +x build.sh start.sh

# 构建前端和后端
./build.sh
```

### 2. 启动服务

```bash
# 默认端口8080启动
./start.sh

# 指定端口启动
./start.sh 8081
```

### 3. 访问应用

- **前端应用**: http://localhost:8080
- **API接口**: http://localhost:8080/api/alter-data
- **健康检查**: http://localhost:8080/health

## 开发模式

### 前端开发

如果需要前端热重载开发：

```bash
cd frontend
npm run dev
```

前端开发服务器将在 http://localhost:5173 启动

### 后端开发

```bash
go run main.go
```

## 前端功能

### 当前功能
- ✅ 响应式UI设计
- ✅ 平台选择导航 (Google Ads、Facebook Marketing、TikTok Marketing)
- ✅ 数据差异分析面板
- ✅ 租户分类展示 (新租户 vs 老租户)
- ✅ 模拟数据展示

### 待添加功能
- 🔄 连接真实API接口
- 📊 数据可视化图表
- 🔄 实时数据刷新
- 📱 移动端适配优化

## 技术栈

### 前端
- **React 18** - UI框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **React Router** - 路由管理
- **CSS3** - 样式设计

### 后端
- **Go** - 后端语言
- **Gin** - Web框架
- **静态文件服务** - 集成前端

## 部署说明

1. 运行 `./build.sh` 构建前端
2. 构建的文件将输出到 `dist/` 目录
3. Go后端会自动服务这些静态文件
4. 单个二进制文件即可部署

## 注意事项

1. **构建顺序**: 必须先构建前端，后端才能正确服务静态文件
2. **路由处理**: 前端使用 React Router，后端配置了 SPA 路由回退
3. **API代理**: 开发模式下前端会代理API请求到后端
4. **跨域处理**: 后端已配置CORS支持

## 故障排查

### 前端构建失败
```bash
cd frontend
npm install  # 重新安装依赖
npm run build
```

### 静态文件404
检查 `dist/` 目录是否存在且包含 `index.html`

### API接口访问失败
确认后端服务已启动，检查 `/health` 端点
