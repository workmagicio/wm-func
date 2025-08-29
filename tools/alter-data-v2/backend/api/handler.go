package api

import (
	"net/http"

	"wm-func/tools/alter-data-v2/backend/controller"

	"github.com/gin-gonic/gin"
)

// GetAlterDataRequest API请求参数
type GetAlterDataRequest struct {
	Platform    string `form:"platform" binding:"required"`
	NeedRefresh bool   `form:"needRefresh"`
}

// GetAlterDataResponse API响应
type GetAlterDataResponse struct {
	Success bool                     `json:"success"`
	Data    controller.AllTenantData `json:"data,omitempty"`
	Message string                   `json:"message,omitempty"`
}

// GetAlterData 获取平台数据差异分析
// @Summary 获取平台数据差异分析
// @Description 根据平台获取租户数据差异分析，包括新老租户分组和最近30天差异统计
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param platform query string true "平台名称" Enums(googleAds,facebookMarketing,tiktokMarketing)
// @Param needRefresh query bool false "是否需要刷新缓存" default(false)
// @Success 200 {object} GetAlterDataResponse "成功"
// @Failure 400 {object} GetAlterDataResponse "参数错误"
// @Failure 500 {object} GetAlterDataResponse "服务器错误"
// @Router /api/alter-data [get]
func GetAlterData(c *gin.Context) {
	var req GetAlterDataRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, GetAlterDataResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 调用业务逻辑
	result := controller.GetAlterDataWithPlatform(req.NeedRefresh, req.Platform)

	// 返回结果
	c.JSON(http.StatusOK, GetAlterDataResponse{
		Success: true,
		Data:    result,
		Message: "获取数据成功",
	})
}

// HealthCheck 健康检查接口
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "alter-data-v2 service is running",
	})
}
