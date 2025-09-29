package api

import (
	"net/http"
	"wm-func/tools/alter-v3/backend/config"
	"wm-func/tools/alter-v3/backend/controller"

	"github.com/gin-gonic/gin"
)

func GetAlterData(c *gin.Context) {
	platform := c.Param("name")
	ctl := controller.NewController(platform)
	ctl.Cac()

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
