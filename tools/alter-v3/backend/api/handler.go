package api

import (
	"net/http"
	"wm-func/tools/alter-v3/backend/config"
	"wm-func/tools/alter-v3/backend/controller"
	"wm-func/tools/alter-v3/backend/tags"

	"github.com/gin-gonic/gin"
)

func GetAlterData(c *gin.Context) {
	platform := c.Param("name")
	ctl := controller.NewController(platform)
	ctl.Cac()
	ctl.AttachUserTags() // 附加用户自定义标签

	c.JSON(http.StatusOK, ctl.ReturnData())
}

func AddConfig(c *gin.Context) {
	cfg := config.Config{}
	c.ShouldBindJSON(&cfg)
	config.AddConfig(cfg)
	c.JSON(http.StatusOK, "success")
}

func RemoveConfig(c *gin.Context) {
	name := c.Param("name")
	config.RemoveConfig(name)
	c.JSON(http.StatusOK, "success")
}

func GetAllConfig(c *gin.Context) {
	configs := config.GetAllConfig()
	c.JSON(http.StatusOK, configs)
}

// UserTag API handlers

// AddUserTag 添加或更新用户标签
func AddUserTag(c *gin.Context) {
	var userTag tags.UserTags
	if err := c.ShouldBindJSON(&userTag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags.AddUserTag(userTag)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// UpdateUserTag 更新用户标签
func UpdateUserTag(c *gin.Context) {
	var userTag tags.UserTags
	if err := c.ShouldBindJSON(&userTag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags.UpdateUserTag(userTag)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// RemoveUserTag 删除用户标签
func RemoveUserTag(c *gin.Context) {
	key := c.Param("key")
	tags.RemoveUserTag(key)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// GetAllUserTags 获取所有用户标签名称（去重）
func GetAllUserTags(c *gin.Context) {
	userTags := tags.GetAllUserTags()

	// 使用 map 去重
	nameMap := make(map[string]bool)
	for _, tag := range userTags {
		nameMap[tag.Name] = true
	}

	// 转换为数组
	names := make([]string, 0, len(nameMap))
	for name := range nameMap {
		names = append(names, name)
	}

	c.JSON(http.StatusOK, names)
}
