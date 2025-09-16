package api

import (
	"fmt"
	"log"
	"net/http"
	"time"
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
	startTime := time.Now()
	clientIP := c.ClientIP()

	log.Printf("🌐 [AddTag] 请求开始 - IP: %s", clientIP)
	log.Printf("📋 [AddTag] URL: %s, Method: %s", c.Request.URL.String(), c.Request.Method)

	var req tags.AddTagRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [AddTag] 参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	log.Printf("📝 [AddTag] 请求参数 - TenantId: %d, Platform: %s, TagName: %s",
		req.TenantId, req.Platform, req.TagName)

	// 验证tag名称不能为空
	if len(req.TagName) == 0 {
		log.Printf("❌ [AddTag] 标签名称为空")
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "标签名称不能为空",
		})
		return
	}

	// 验证tag名称长度
	if len(req.TagName) > 20 {
		log.Printf("❌ [AddTag] 标签名称过长: %d 字符", len(req.TagName))
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "标签名称不能超过20个字符",
		})
		return
	}

	// 调用业务逻辑添加标签
	log.Printf("🔍 [AddTag] 调用 tags.AddTag")
	if err := tags.AddTag(req); err != nil {
		log.Printf("❌ [AddTag] 添加标签失败: %v", err)
		c.JSON(http.StatusInternalServerError, TagResponse{
			Success: false,
			Message: "添加标签失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的标签列表
	updatedTags := tags.GetPlatformTags(req.Platform)

	duration := time.Since(startTime)
	log.Printf("📊 [AddTag] 标签添加成功 - 更新后标签数量: %d", len(updatedTags))

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

	log.Printf("✅ [AddTag] 请求完成 - 耗时: %v, IP: %s", duration, clientIP)
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
	startTime := time.Now()
	clientIP := c.ClientIP()

	log.Printf("🌐 [RemoveTag] 请求开始 - IP: %s", clientIP)
	log.Printf("📋 [RemoveTag] URL: %s, Method: %s", c.Request.URL.String(), c.Request.Method)

	var req tags.RemoveTagRequest

	// 绑定请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [RemoveTag] 参数绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "参数错误: " + err.Error(),
		})
		return
	}

	log.Printf("📝 [RemoveTag] 请求参数 - TenantId: %d, Platform: %s, TagName: %s",
		req.TenantId, req.Platform, req.TagName)

	// 调用业务逻辑删除标签
	log.Printf("🔍 [RemoveTag] 调用 tags.RemoveTag")
	if err := tags.RemoveTag(req); err != nil {
		log.Printf("❌ [RemoveTag] 删除标签失败: %v", err)
		c.JSON(http.StatusInternalServerError, TagResponse{
			Success: false,
			Message: "删除标签失败: " + err.Error(),
		})
		return
	}

	// 获取更新后的标签列表
	updatedTags := tags.GetPlatformTags(req.Platform)

	duration := time.Since(startTime)
	log.Printf("📊 [RemoveTag] 标签删除成功 - 更新后标签数量: %d", len(updatedTags))

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

	log.Printf("✅ [RemoveTag] 请求完成 - 耗时: %v, IP: %s", duration, clientIP)
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
	startTime := time.Now()
	clientIP := c.ClientIP()

	log.Printf("🌐 [GetTags] 请求开始 - IP: %s", clientIP)
	log.Printf("📋 [GetTags] URL: %s, Method: %s", c.Request.URL.String(), c.Request.Method)

	tenantIdStr := c.Param("tenant_id")
	platform := c.Param("platform")

	log.Printf("📝 [GetTags] 请求参数 - TenantIdStr: %s, Platform: %s", tenantIdStr, platform)

	// 解析tenant_id
	var tenantId int64
	if _, err := fmt.Sscanf(tenantIdStr, "%d", &tenantId); err != nil {
		log.Printf("❌ [GetTags] 租户ID解析失败: %s, 错误: %v", tenantIdStr, err)
		c.JSON(http.StatusBadRequest, TagResponse{
			Success: false,
			Message: "无效的租户ID",
		})
		return
	}

	// 获取所有标签
	log.Printf("🔍 [GetTags] 调用 tags.GetAllTags - TenantId: %d", tenantId)
	allTags := tags.GetAllTags(tenantId, platform)

	duration := time.Since(startTime)
	log.Printf("📊 [GetTags] 获取标签成功 - 标签数量: %d", len(allTags))

	c.JSON(http.StatusOK, TagResponse{
		Success: true,
		Message: "获取标签成功",
		Data: map[string]interface{}{
			"tenant_id": tenantId,
			"platform":  platform,
			"tags":      allTags,
		},
	})

	log.Printf("✅ [GetTags] 请求完成 - 耗时: %v, IP: %s", duration, clientIP)
}
