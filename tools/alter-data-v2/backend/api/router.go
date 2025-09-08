package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建gin引擎
	r := gin.Default()

	// 设置CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置静态文件服务 - 服务前端构建的文件
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/favicon.ico", "./dist/favicon.ico")
	r.StaticFile("/vite.svg", "./dist/vite.svg")

	// 服务前端应用的入口页面
	r.GET("/", func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	// 处理前端路由 - SPA应用的路由回退
	r.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if filepath.HasPrefix(c.Request.URL.Path, "/api") || filepath.HasPrefix(c.Request.URL.Path, "/health") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "接口不存在",
			})
			return
		}

		// 否则服务前端应用
		c.File("./dist/index.html")
	})

	// 健康检查接口
	r.GET("/health", HealthCheck)

	// API路由组
	api := r.Group("/api")
	{
		// 获取数据差异分析
		api.GET("/alter-data", GetAlterData)

		// 归因数据分析接口
		api.GET("/attribution", GetAttributionData)
		api.GET("/attribution/all", GetAllAttributionData)
		api.GET("/attribution-data/grouped", GetAttributionDataGrouped)
		api.GET("/attribution/:tenantId", GetAttributionDataByPath)

		// 标签管理接口
		api.POST("/tags", AddTag)
		api.DELETE("/tags", RemoveTag)
		api.GET("/tags/:tenant_id/:platform", GetTags)
	}

	return r
}
