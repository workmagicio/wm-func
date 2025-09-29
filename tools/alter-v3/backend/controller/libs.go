package controller

import (
	"time"
	"wm-func/tools/alter-v3/backend/alter_common"
	"wm-func/tools/alter-v3/backend/data"
)

func (c *Controller) initData() {
	c.ApiData = c.getDataWithSql(c.Cfg.ApiSql)
	c.WmData = c.getDataWithSql(c.Cfg.WmSql)
	allTenant := alter_common.GetAllTenantWithPlatform(c.Cfg.BasePlatform)
	streams := c.getStreamSlice()
	allData := []data.Analytics{}

	for _, tenantId := range allTenant {
		t := data.Analytics{
			TenantId: tenantId,
		}
		dataSlice := []data.DataSlice{}

		for _, stream := range streams {

		}
	}
}

func (c *Controller) getDataWithSql(exec string) []alter_common.Data {
	if exec == "" {
		return nil
	}
	return alter_common.GetData(exec)
}

func (c *Controller) getStreamSlice() []string {
	now := time.Now().UTC()
	res := []string{}
	for i := 0; i < c.Cfg.TotalDataCount; i++ {
		res = append(res, now.Format("2006-01-02"))
		now = now.Add(time.Hour * 24 * -1)
	}

	return res
}
