package api

import (
	"fmt"
	"net/http"
	"wm-func/tools/alter-data-v2/backend/tags"

	"github.com/gin-gonic/gin"
)

// TagResponse API响应格式
type TagResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AddTag 添加标签接口
// @Summary 添加标签
// @Description 为指定租户和平台添加标签，标签30天后自动过期
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param tag body tags.AddTagRequest true "添加标签请求"
// @Success 200 {object} TagResponse "成功"
// @Failure 400 {object} TagResponse "参数错误"
// @Failure 500 {object} TagResponse "服务器错误"
// @Router /api/tags [post]
func AddTag(c *gin.Context) {
	var req tags.AddTagRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 验证tag名称不能为空
	if len(req.TagName) == 0 {
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "标签名称不能为空",
		})
		return
	}

	// 验证tag名称长度
	if len(req.TagName) > 20 {
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "标签名称不能超过20个字符",
		})
		return
	}

	// 调用业务逻辑添加标签
	if err := tags.AddTag(req); err != nil {
		c.JSON(http.StatusInternalServerError, TagResponse{
			Success: false,
			Message: "添加标签失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的标签列表
	updatedTags := tags.GetPlatformTags(req.Platform)

	// 返回成功响应
	c.JSON(http.StatusOK, TagResponse{
		Success: true,
		Message: "标签添加成功",
		Data: map[string]interface{}{
			"tenant_id": req.TenantId,
			"platform":  req.Platform,
			"tag_name":  req.TagName,
			"tags":      updatedTags,
		},
	})
}

// RemoveTag 删除标签接口
// @Summary 删除标签
// @Description 删除指定租户和平台的标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param tag body tags.RemoveTagRequest true "删除标签请求"
// @Success 200 {object} TagResponse "成功"
// @Failure 400 {object} TagResponse "参数错误"
// @Failure 500 {object} TagResponse "服务器错误"
// @Router /api/tags [delete]
func RemoveTag(c *gin.Context) {
	var req tags.RemoveTagRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	// 调用业务逻辑删除标签
	if err := tags.RemoveTag(req); err != nil {
		c.JSON(http.StatusInternalServerError, TagResponse{
			Success: false,
			Message: "删除标签失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的标签列表
	updatedTags := tags.GetPlatformTags(req.Platform)

	// 返回成功响应
	c.JSON(http.StatusOK, TagResponse{
		Success: true,
		Message: "标签删除成功",
		Data: map[string]interface{}{
			"tenant_id": req.TenantId,
			"platform":  req.Platform,
			"tag_name":  req.TagName,
			"tags":      updatedTags,
		},
	})
}

// GetTags 获取标签接口 (可选，用于调试)
// @Summary 获取标签
// @Description 获取指定租户和平台的所有有效标签
// @Tags 标签管理
// @Produce json
// @Param tenant_id path int true "租户ID"
// @Param platform path string true "平台名称"
// @Success 200 {object} TagResponse "成功"
// @Router /api/tags/{tenant_id}/{platform} [get]
func GetTags(c *gin.Context) {
	tenantIdStr := c.Param("tenant_id")
	platform := c.Param("platform")

	// 解析tenant_id
	var tenantId int64
	if _, err := fmt.Sscanf(tenantIdStr, "%d", &tenantId); err != nil {
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "无效的租户ID",
		})
		return
	}

	// 获取所有标签
	allTags := tags.GetAllTags(tenantId, platform)

	c.JSON(http.StatusOK, TagResponse{
		Success: true,
		Message: "获取标签成功",
		Data: map[string]interface{}{
			"tenant_id": tenantId,
			"platform":  platform,
			"tags":      allTags,
		},
	})
}
