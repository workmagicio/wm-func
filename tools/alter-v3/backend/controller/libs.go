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
	newTenant := alter_common.GetLast15DayRegisterTenant()
	streams := c.getStreamSlice()

	allData := []*data.Analytics{}

	for _, tenantId := range allTenant {
		t := &data.Analytics{
			TenantId:     tenantId,
			RegisterTime: newTenant[tenantId],
			IsNewTenant:  len(newTenant[tenantId]) > 0,
			HaveApiData:  c.ApiData != nil,
			Tags:         []string{},
		}

		for _, stream := range streams {
			tmp := data.DataSlice{
				WMData: c.WmData[tenantId][stream].Data,
				Date:   stream,
			}

			if c.ApiData != nil {
				tmp.APiData = c.ApiData[tenantId][stream].Data
			}

			t.Data = append(t.Data, tmp)
		}

		allData = append(allData, t)
	}

	c.allData = allData
}

func (c *Controller) getDataWithSql(exec string) map[int64]map[string]alter_common.Data {
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
