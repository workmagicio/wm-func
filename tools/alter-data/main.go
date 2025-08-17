package main

import (
	"fmt"
	"log"
	"net/http"
	"wm-func/tools/alter-data/internal/handlers"
	"wm-func/tools/alter-data/internal/platform"

	"github.com/gorilla/mux"
)

func main() {
	// 初始化平台注册
	if err := platform.InitializePlatforms(); err != nil {
		log.Fatalf("Failed to initialize platforms: %v", err)
	}

	// 创建处理器
	apiHandler := handlers.NewAPIHandler()

	// 创建路由
	router := mux.NewRouter()

	// API 路由
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/platforms", apiHandler.GetPlatforms).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}", apiHandler.GetPlatformData).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}/{tenant_id}", apiHandler.GetTenantData).Methods("GET")
	apiRouter.HandleFunc("/refresh/{platform}", apiHandler.RefreshPlatformData).Methods("POST")
	apiRouter.HandleFunc("/cache/stats", apiHandler.GetCacheStats).Methods("GET")

	// 静态文件路由
	router.PathPrefix("/static/").Handler(handlers.ServeStatic())

	// 主页路由
	router.HandleFunc("/", handlers.ServeIndex).Methods("GET")

	// 启动服务器
	port := ":8090"
	fmt.Printf("🚀 数据监控看板服务已启动\n")
	fmt.Printf("📊 访问地址: http://localhost%s\n", port)
	fmt.Printf("🔗 API 接口:\n")
	fmt.Printf("   GET  /api/platforms - 获取平台列表\n")
	fmt.Printf("   GET  /api/data/{platform} - 获取平台数据\n")
	fmt.Printf("   GET  /api/data/{platform}?refresh=true - 强制刷新平台数据\n")
	fmt.Printf("   GET  /api/data/{platform}/{tenant_id} - 获取租户数据\n")
	fmt.Printf("   POST /api/refresh/{platform} - 刷新平台缓存\n")
	fmt.Printf("   GET  /api/cache/stats - 获取缓存统计\n")
	fmt.Printf("📁 已实现平台: %v\n", platform.GetImplementedPlatformNames())

	log.Fatal(http.ListenAndServe(port, router))
}
