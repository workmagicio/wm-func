import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// 生产环境配置 - 用于构建静态文件
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../dist',
    emptyOutDir: true,
  },
  // 生产环境不需要代理，Nginx 会处理
})
