package api

import "github.com/gin-gonic/gin"

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

	api := r.Group("/api")
	{
		api.GET("/alter-data/:name", GetAlterData)
		api.POST("/config", AddConfig)
		api.DELETE("/config/:name", RemoveConfig)
		api.GET("/config", GetAllConfig)

		// UserTag 路由
		api.POST("/user-tag", AddUserTag)
		api.PUT("/user-tag", UpdateUserTag)
		api.DELETE("/user-tag/:key", RemoveUserTag)
		api.GET("/user-tags", GetAllUserTags)
	}

	return r
}
