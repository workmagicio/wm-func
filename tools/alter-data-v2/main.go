package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"wm-func/tools/alter-data-v2/backend/api"
)

func main() {
	// 设置日志输出格式，确保在Docker中可见
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 命令行参数
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// 打印启动信息
	fmt.Println("===========================================")
	fmt.Println("🚀 ALTER DATA V2 服务启动中...")
	fmt.Printf("📅 启动时间: %s\n", log.Prefix())
	fmt.Printf("🌐 服务端口: %s\n", *port)
	fmt.Printf("🔧 GIN模式: %s\n", os.Getenv("GIN_MODE"))
	fmt.Println("===========================================")

	// 设置路由
	router := api.SetupRouter()

	// 启动服务器
	serverAddr := fmt.Sprintf(":%s", *port)
	log.Printf("🟢 alter-data-v2 服务器启动在端口 %s", *port)
	log.Printf("📋 API文档地址: http://localhost:%s/api/alter-data", *port)
	log.Printf("💊 健康检查: http://localhost:%s/health", *port)

	// 启动HTTP服务器
	fmt.Printf("🎯 开始监听端口 %s...\n", *port)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("❌ 启动服务器失败: %v", err)
	}
}
