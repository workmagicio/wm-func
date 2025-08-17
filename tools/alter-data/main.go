package main

import (
	"fmt"
	"log"
	"net/http"
	"wm-func/tools/alter-data/handlers"
	"wm-func/tools/alter-data/platforms"

	"github.com/gorilla/mux"
)

func main() {
	// 注册平台实现
	platforms.RegisterPlatform(&platforms.GooglePlatform{})

	// 创建处理器
	apiHandler := handlers.NewAPIHandler()

	// 创建路由
	router := mux.NewRouter()

	// API 路由
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/platforms", apiHandler.GetPlatforms).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}", apiHandler.GetPlatformData).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}/{tenant_id}", apiHandler.GetTenantData).Methods("GET")

	// 静态文件路由
	router.PathPrefix("/static/").Handler(handlers.ServeStatic())

	// 主页路由
	router.HandleFunc("/", handlers.ServeIndex).Methods("GET")

	// 启动服务器
	port := ":8090"
	fmt.Printf("🚀 数据监控看板服务已启动\n")
	fmt.Printf("📊 访问地址: http://localhost%s\n", port)
	fmt.Printf("🔗 API 接口:\n")
	fmt.Printf("   GET /api/platforms\n")
	fmt.Printf("   GET /api/data/{platform}\n")
	fmt.Printf("   GET /api/data/{platform}/{tenant_id}\n")
	fmt.Printf("📁 已实现平台: %v\n", platforms.GetImplementedPlatformNames())

	log.Fatal(http.ListenAndServe(port, router))
}
