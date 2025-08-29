package main

import (
	"flag"
	"fmt"
	"log"
	"wm-func/tools/alter-data-v2/backend/api"
)

func main() {
	// 命令行参数
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	// 设置路由
	router := api.SetupRouter()

	// 启动服务器
	serverAddr := fmt.Sprintf(":%s", *port)
	log.Printf("alter-data-v2 服务器启动在端口 %s", *port)
	log.Printf("API文档地址: http://localhost:%s/api/alter-data", *port)
	log.Printf("健康检查: http://localhost:%s/health", *port)

	// 启动HTTP服务器
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
