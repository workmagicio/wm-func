package api

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	// 创建gin引擎
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/alter-data/:platform", GetAlterData)
	}
	return r
}
