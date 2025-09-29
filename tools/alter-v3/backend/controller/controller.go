package controller

import (
	"fmt"
	"sort"
	"wm-func/tools/alter-v3/backend/alter_common"
	"wm-func/tools/alter-v3/backend/cac"
	"wm-func/tools/alter-v3/backend/config"
	"wm-func/tools/alter-v3/backend/data"
)

type Controller struct {
	Name        string
	Cfg         config.Config
	ApiData     map[int64]map[string]alter_common.Data
	WmData      map[int64]map[string]alter_common.Data
	streamSlice []string
	tenants     []int64
	allData     []*data.Analytics
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

	c.initData()

	return c
}

func (c *Controller) Cac() {
	cac.Ca(c.Cfg.BasePlatform, c.allData)

	for _, v := range c.allData {
		if len(v.Tags) > 0 {
			fmt.Println(v.TenantId, v.Tags)
		}
	}
}

type ReturnData struct {
	NewTenantData []*data.Analytics
	TenantData    []*data.Analytics
}

func (c *Controller) ReturnData() ReturnData {
	newTenant := []*data.Analytics{}
	allTenant := []*data.Analytics{}

	for i, v := range c.allData {
		if v.IsNewTenant {
			newTenant = append(newTenant, c.allData[i])
		} else {
			allTenant = append(allTenant, c.allData[i])
		}
	}
	sort.Slice(newTenant, func(i, j int) bool {
		return len(newTenant[i].Tags) > len(newTenant[j].Tags)
	})

	sort.Slice(allTenant, func(i, j int) bool {
		return len(allTenant[i].Tags) > len(allTenant[j].Tags)
	})

	allTenantErr := []*data.Analytics{}
	allTenantTrue := []*data.Analytics{}
	for i, v := range allTenant {
		if len(v.Tags) > 0 {
			allTenantErr = append(allTenantErr, allTenant[i])
		} else {
			allTenantTrue = append(allTenantTrue, allTenant[i])
		}
	}

	sort.Slice(allTenantTrue, func(i, j int) bool {
		return allTenantTrue[i].GetAvg() > allTenantTrue[j].GetAvg()
	})

	return ReturnData{
		NewTenantData: newTenant,
		TenantData:    append(allTenantErr, allTenantTrue...),
	}
}
