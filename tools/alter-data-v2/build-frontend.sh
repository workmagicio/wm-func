#!/bin/bash

# 前端编译脚本
# 用于自动化编译 alter-data-v2 前端项目

set -e  # 遇到错误时退出

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

echo "🚀 开始编译前端项目..."

# 检查frontend目录是否存在
if [ ! -d "$FRONTEND_DIR" ]; then
    echo "❌ 错误: frontend 目录不存在 ($FRONTEND_DIR)"
    exit 1
fi

# 进入frontend目录
cd "$FRONTEND_DIR"
echo "📁 当前目录: $(pwd)"

# 检查是否有package.json
if [ ! -f "package.json" ]; then
    echo "❌ 错误: package.json 文件不存在"
    exit 1
fi

# 设置代理（根据用户的记忆配置）
echo "🔧 配置网络代理..."
if command -v xp >/dev/null 2>&1; then
    echo "使用 xp 命令设置代理..."
    xp
else
    echo "⚠️  警告: xp 命令不可用，跳过代理设置"
fi

# 检查node_modules是否存在，如果不存在或过期则安装依赖
if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
    echo "📦 安装依赖包..."
    if command -v yarn >/dev/null 2>&1; then
        echo "使用 yarn 安装依赖..."
        yarn install
    else
        echo "使用 npm 安装依赖..."
        npm install
    fi
else
    echo "✅ 依赖包已是最新，跳过安装"
fi

# 获取实际的构建输出目录（从vite.config.ts读取outDir）
BUILD_OUTPUT_DIR="../dist"
if [ -f "vite.config.ts" ]; then
    # 尝试从vite配置中提取outDir
    VITE_OUTDIR=$(grep -o "outDir: *['\"][^'\"]*['\"]" vite.config.ts | sed "s/outDir: *['\"]//g" | sed "s/['\"].*//g" || echo "")
    if [ ! -z "$VITE_OUTDIR" ]; then
        BUILD_OUTPUT_DIR="$VITE_OUTDIR"
    fi
fi

# 清理之前的构建
if [ -d "$BUILD_OUTPUT_DIR" ]; then
    echo "🧹 清理之前的构建..."
    rm -rf "$BUILD_OUTPUT_DIR"
fi

# 构建项目
echo "🔨 开始构建项目..."
if command -v yarn >/dev/null 2>&1; then
    yarn build
else
    npm run build
fi

# 检查构建结果
if [ -d "$BUILD_OUTPUT_DIR" ]; then
    echo "✅ 前端构建成功!"
    echo "📊 构建统计:"
    echo "   构建目录: $(cd "$BUILD_OUTPUT_DIR" && pwd)"
    echo "   文件数量: $(find "$BUILD_OUTPUT_DIR" -type f | wc -l)"
    echo "   总大小: $(du -sh "$BUILD_OUTPUT_DIR" | cut -f1)"
    
    echo ""
    echo "📋 主要文件:"
    find "$BUILD_OUTPUT_DIR" -name "*.html" -o -name "*.js" -o -name "*.css" | head -10 | while read file; do
        size=$(du -sh "$file" | cut -f1)
        echo "   $file ($size)"
    done
else
    echo "❌ 构建失败: $BUILD_OUTPUT_DIR 目录未生成"
    exit 1
fi

echo ""
echo "🎉 前端编译完成!"
echo "💡 你可以运行以下命令预览构建结果:"
echo "   cd $FRONTEND_DIR && npm run preview"
