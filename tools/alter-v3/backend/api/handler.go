package api

import (
	"net/http"
	"wm-func/tools/alter-v3/backend/config"
	"wm-func/tools/alter-v3/backend/controller"

	"github.com/gin-gonic/gin"
)

func GetAlterData(c *gin.Context) {
	platform := c.Param("platform")
	ctl := controller.NewController(platform)
	ctl.Cac()

	c.JSON(http.StatusOK, ctl.ReturnData())
}

func AddConifg(c *gin.Context) {
	cfg := config.Config{}
	c.ShouldBindJSON(&cfg)
	config.AddConfig(cfg)
	c.JSON(http.StatusOK, "success")
}

func RemoveConifg(c *gin.Context) {
	name := c.Param("name")
	config.RemoveConfig(name)
	c.JSON(http.StatusOK, "success")
}
