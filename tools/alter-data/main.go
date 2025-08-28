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

	// 平台相关API
	apiRouter.HandleFunc("/platforms", apiHandler.GetPlatforms).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}", apiHandler.GetPlatformData).Methods("GET")
	apiRouter.HandleFunc("/data/{platform}/{tenant_id}", apiHandler.GetTenantData).Methods("GET")
	apiRouter.HandleFunc("/refresh/{platform}", apiHandler.RefreshPlatformData).Methods("POST")

	// 租户相关API
	apiRouter.HandleFunc("/tenants", apiHandler.GetTenants).Methods("GET")
	apiRouter.HandleFunc("/tenants/recent", apiHandler.GetRecentTenants).Methods("GET")
	apiRouter.HandleFunc("/tenants/frequent", apiHandler.GetFrequentTenants).Methods("GET")
	apiRouter.HandleFunc("/tenant/{tenant_id}", apiHandler.GetTenantCrossPlatformData).Methods("GET")
	apiRouter.HandleFunc("/tenant/{tenant_id}/refresh", apiHandler.RefreshTenantData).Methods("POST")

	// 系统API
	apiRouter.HandleFunc("/cache/stats", apiHandler.GetCacheStats).Methods("GET")

	// 归因订单分析API
	apiRouter.HandleFunc("/attribution-orders", apiHandler.GetAttributionOrders).Methods("GET")
	apiRouter.HandleFunc("/attribution-orders/{tenant_id}", apiHandler.GetTenantAttributionOrders).Methods("GET")
	apiRouter.HandleFunc("/attribution-orders/refresh", apiHandler.RefreshAttributionOrders).Methods("POST")

	// Amazon订单分析API
	apiRouter.HandleFunc("/amazon-orders", apiHandler.GetAllAmazonOrders).Methods("GET")
	apiRouter.HandleFunc("/amazon-orders/{tenant_id}", apiHandler.GetTenantAmazonOrders).Methods("GET")
	apiRouter.HandleFunc("/amazon-orders/refresh", apiHandler.RefreshAmazonOrders).Methods("POST")

	// Fairing分析API
	apiRouter.HandleFunc("/fairing", apiHandler.GetAllFairing).Methods("GET")
	apiRouter.HandleFunc("/fairing/{tenant_id}", apiHandler.GetTenantFairing).Methods("GET")
	apiRouter.HandleFunc("/fairing/refresh", apiHandler.RefreshFairing).Methods("POST")

	// 静态文件路由
	router.PathPrefix("/static/").Handler(handlers.ServeStatic())

	// 主页路由
	router.HandleFunc("/", handlers.ServeIndex).Methods("GET")
	// 归因订单分析页面
	router.HandleFunc("/attribution", handlers.ServeAttributionPage).Methods("GET")
	// Amazon订单分析页面
	router.HandleFunc("/amazon-orders", handlers.ServeAmazonOrdersPage).Methods("GET")
	// Fairing分析页面
	router.HandleFunc("/fairing", handlers.ServeFairingPage).Methods("GET")

	// 启动服务器
	port := ":8090"
	fmt.Printf("🚀 数据监控看板服务已启动\n")
	fmt.Printf("📊 访问地址: http://localhost%s\n", port)
	fmt.Printf("🔗 API 接口:\n")
	fmt.Printf("📊 平台视图:\n")
	fmt.Printf("   GET  /api/platforms - 获取平台列表\n")
	fmt.Printf("   GET  /api/data/{platform} - 获取平台数据\n")
	fmt.Printf("   GET  /api/data/{platform}?refresh=true - 强制刷新平台数据\n")
	fmt.Printf("   GET  /api/data/{platform}/{tenant_id} - 获取指定平台的租户数据\n")
	fmt.Printf("   POST /api/refresh/{platform} - 刷新平台缓存\n")
	fmt.Printf("👤 租户视图:\n")
	fmt.Printf("   GET  /api/tenants - 获取租户列表\n")
	fmt.Printf("   GET  /api/tenants/recent - 获取最近15天注册租户\n")
	fmt.Printf("   GET  /api/tenant/{tenant_id} - 获取租户跨平台数据\n")
	fmt.Printf("   GET  /api/tenant/{tenant_id}?refresh=true - 强制刷新租户数据\n")
	fmt.Printf("   POST /api/tenant/{tenant_id}/refresh - 刷新租户缓存\n")
	fmt.Printf("🔧 系统:\n")
	fmt.Printf("   GET  /api/cache/stats - 获取缓存统计\n")
	fmt.Printf("📈 归因订单分析:\n")
	fmt.Printf("   GET  /attribution - 归因订单分析页面\n")
	fmt.Printf("   GET  /api/attribution-orders - 获取所有租户归因订单数据\n")
	fmt.Printf("   GET  /api/attribution-orders/{tenant_id} - 获取指定租户归因订单数据\n")
	fmt.Printf("   POST /api/attribution-orders/refresh - 刷新归因订单缓存\n")
	fmt.Printf("🛒 Amazon Vendor分析:\n")
	fmt.Printf("   GET  /amazon-orders - Amazon Vendor分析页面\n")
	fmt.Printf("   GET  /api/amazon-orders - 获取所有租户Amazon Vendor订单数据\n")
	fmt.Printf("   GET  /api/amazon-orders/{tenant_id} - 获取指定租户Amazon Vendor订单数据\n")
	fmt.Printf("   POST /api/amazon-orders/refresh - 刷新Amazon Vendor订单缓存\n")
	fmt.Printf("📋 Fairing分析:\n")
	fmt.Printf("   GET  /fairing - Fairing分析页面\n")
	fmt.Printf("   GET  /api/fairing - 获取所有租户Fairing数据\n")
	fmt.Printf("   GET  /api/fairing/{tenant_id} - 获取指定租户Fairing数据\n")
	fmt.Printf("   POST /api/fairing/refresh - 刷新Fairing缓存\n")
	fmt.Printf("📁 已实现平台: %v\n", platform.GetImplementedPlatformNames())

	log.Fatal(http.ListenAndServe(port, router))
}
