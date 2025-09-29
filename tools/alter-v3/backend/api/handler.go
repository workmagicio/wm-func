package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wm-func/tools/alter-v3/backend/controller"
)

func GetAlterData(c *gin.Context) {
	platform := c.Param("platform")
	ctl := controller.NewController(platform)
	ctl.Cac()

	c.JSON(http.StatusOK, ctl.ReturnData())
}
