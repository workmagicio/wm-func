package api

import (
	"net/http"
	"strconv"

	"wm-func/tools/alter-data-v2/backend/cac"
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

// GetAttributionDataRequest 归因数据API请求参数
type GetAttributionDataRequest struct {
	TenantId    int64 `form:"tenantId" binding:"required"`
	NeedRefresh bool  `form:"needRefresh"`
}

// GetAttributionDataResponse 归因数据API响应
type GetAttributionDataResponse struct {
	Success bool                      `json:"success"`
	Data    cac.AttributionTenantData `json:"data,omitempty"`
	Message string                    `json:"message,omitempty"`
}

// GetAttributionData 获取归因数据分析
// @Summary 获取归因数据分析
// @Description 根据租户ID获取归因数据分析，包括各平台归因数据、Shopify API数据对比等
// @Tags 归因分析
// @Accept json
// @Produce json
// @Param tenantId query int true "租户ID"
// @Param needRefresh query bool false "是否需要刷新缓存" default(false)
// @Success 200 {object} GetAttributionDataResponse "成功"
// @Failure 400 {object} GetAttributionDataResponse "参数错误"
// @Failure 500 {object} GetAttributionDataResponse "服务器错误"
// @Router /api/attribution [get]
func GetAttributionData(c *gin.Context) {
	var req GetAttributionDataRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, GetAttributionDataResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 调用业务逻辑
	result := cac.GetAttributionDataWithTenantId(req.TenantId, req.NeedRefresh)

	// 返回结果
	c.JSON(http.StatusOK, GetAttributionDataResponse{
		Success: true,
		Data:    result,
		Message: "获取归因数据成功",
	})
}

// GetAttributionDataByPath 通过路径参数获取归因数据分析
// @Summary 通过路径参数获取归因数据分析
// @Description 根据租户ID获取归因数据分析，包括各平台归因数据、Shopify API数据对比等
// @Tags 归因分析
// @Accept json
// @Produce json
// @Param tenantId path int true "租户ID"
// @Param needRefresh query bool false "是否需要刷新缓存" default(false)
// @Success 200 {object} GetAttributionDataResponse "成功"
// @Failure 400 {object} GetAttributionDataResponse "参数错误"
// @Failure 500 {object} GetAttributionDataResponse "服务器错误"
// @Router /api/attribution/{tenantId} [get]
func GetAttributionDataByPath(c *gin.Context) {
	// 获取路径参数
	tenantIdStr := c.Param("tenantId")
	tenantId, err := strconv.ParseInt(tenantIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetAttributionDataResponse{
			Success: false,
			Message: "租户ID参数错误: " + err.Error(),
		})
		return
	}

	// 获取查询参数
	needRefresh := c.Query("needRefresh") == "true"

	// 调用业务逻辑
	result := cac.GetAttributionDataWithTenantId(tenantId, needRefresh)

	// 返回结果
	c.JSON(http.StatusOK, GetAttributionDataResponse{
		Success: true,
		Data:    result,
		Message: "获取归因数据成功",
	})
}

// GetAllAttributionDataResponse 所有归因数据API响应
type GetAllAttributionDataResponse struct {
	Success bool                        `json:"success"`
	Data    []cac.AttributionTenantData `json:"data,omitempty"`
	Message string                      `json:"message,omitempty"`
}

// GetAllAttributionData 获取所有租户的归因数据分析
// @Summary 获取所有租户的归因数据分析
// @Description 获取所有租户的归因数据分析，用于归因分析页面展示
// @Tags 归因分析
// @Accept json
// @Produce json
// @Param needRefresh query bool false "是否需要刷新缓存" default(false)
// @Success 200 {object} GetAllAttributionDataResponse "成功"
// @Failure 500 {object} GetAllAttributionDataResponse "服务器错误"
// @Router /api/attribution/all [get]
func GetAllAttributionData(c *gin.Context) {
	// 获取查询参数
	needRefresh := c.Query("needRefresh") == "true"

	// 调用业务逻辑
	result := cac.GetAllAttributionData(needRefresh)

	// 返回结果
	c.JSON(http.StatusOK, GetAllAttributionDataResponse{
		Success: true,
		Data:    result,
		Message: "获取所有归因数据成功",
	})
}

// GetAttributionDataGroupedResponse 按客户类型分组的归因数据API响应
type GetAttributionDataGroupedResponse struct {
	Success      bool                        `json:"success"`
	NewCustomers []cac.AttributionTenantData `json:"new_customers,omitempty"`
	OldCustomers []cac.AttributionTenantData `json:"old_customers,omitempty"`
	Message      string                      `json:"message,omitempty"`
}

// GetAttributionDataGrouped 获取按新老客户分组的归因数据分析
// @Summary 获取按新老客户分组的归因数据分析
// @Description 获取归因数据并按新客户（注册30天内）和老客户（注册30天以上）分组
// @Tags 归因分析
// @Accept json
// @Produce json
// @Param needRefresh query bool false "是否需要刷新缓存"
// @Success 200 {object} GetAttributionDataGroupedResponse "成功"
// @Failure 500 {object} GetAttributionDataGroupedResponse "服务器错误"
// @Router /api/attribution/grouped [get]
func GetAttributionDataGrouped(c *gin.Context) {
	needRefresh := c.DefaultQuery("needRefresh", "false") == "true"

	newCustomers, oldCustomers := cac.GetAttributionDataGroupedByCustomerType(needRefresh)

	c.JSON(http.StatusOK, GetAttributionDataGroupedResponse{
		Success:      true,
		NewCustomers: newCustomers,
		OldCustomers: oldCustomers,
		Message:      "获取分组归因数据成功",
	})
}

// HealthCheck 健康检查接口
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "alter-data-v2 service is running",
	})
}
