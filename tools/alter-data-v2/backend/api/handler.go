package api

import (
	"net/http"

	"wm-func/tools/alter-data-v2/backend/controller"
	"wm-func/tools/alter-data-v2/backend/tags"

	"github.com/gin-gonic/gin"
)

// GetAlterDataRequest API请求参数
type GetAlterDataRequest struct {
	Platform    string `form:"platform" binding:"required"`
	NeedRefresh bool   `form:"needRefresh"`
	TenantId    *int64 `form:"tenantId"` // 可选参数，用于缓存更新
}

// GetAlterDataResponse API响应
type GetAlterDataResponse struct {
	Success    bool                     `json:"success"`
	Data       controller.AllTenantData `json:"data,omitempty"`
	Message    string                   `json:"message,omitempty"`
	GlobalTags []string                 `json:"global_tags,omitempty"`
}

// GetAlterData 获取平台数据差异分析
// @Summary 获取平台数据差异分析
// @Description 根据平台获取租户数据差异分析，包括新老租户分组和最近30天差异统计
// @Tags 数据分析
// @Accept json
// @Produce json
// @Param platform query string true "平台名称" Enums(googleAds,facebookMarketing,tiktokMarketing)
// @Param needRefresh query bool false "是否需要刷新缓存" default(false)
// @Param tenantId query int false "租户ID，用于补齐缓存数据" default()
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
	var result controller.AllTenantData
	if req.TenantId != nil {
		// 有 tenantId 参数，调用带 tenantId 的函数
		result = controller.GetAlterDataWithPlatformWithTenantId(req.NeedRefresh, req.Platform, *req.TenantId)
	} else {
		// 没有 tenantId 参数，调用普通函数（等同于 tenantId = -1）
		result = controller.GetAlterDataWithPlatformWithTenantId(req.NeedRefresh, req.Platform, -1)
	}

	// 获取全局标签列表
	globalTags := tags.GetPlatformTags(req.Platform)

	// 返回结果
	c.JSON(http.StatusOK, GetAlterDataResponse{
		Success:    true,
		Data:       result,
		Message:    "获取数据成功",
		GlobalTags: globalTags,
	})
}

// HealthCheck 健康检查接口
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "alter-data-v2 service is running",
	})
}
