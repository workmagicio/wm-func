package controller

import (
	"wm-func/tools/alter-v3/backend/alter_common"
	"wm-func/tools/alter-v3/backend/config"
)

type Controller struct {
	Name        string
	Cfg         config.Config
	ApiData     []alter_common.Data
	WmData      []alter_common.Data
	streamSlice []string
	tenants     []int64
}

func NewController(name string) *Controller {
	cfg := config.GetConfit()
	if _, ok := cfg[name]; !ok {
		panic("not exists " + name)
	}

	c := &Controller{
		Name: name,
		Cfg:  cfg[name],
	}

	c.initStreamSlice()
	c.initData()

	return c
}
