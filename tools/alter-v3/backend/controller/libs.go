package controller

import (
	"strconv"
	"strings"
	"time"
	"wm-func/tools/alter-v3/backend/alter_common"
	"wm-func/tools/alter-v3/backend/data"
)

func (c *Controller) initData() {
	c.ApiData = c.getDataWithSql(c.Cfg.ApiSql)
	c.WmData = c.getDataWithSql(c.Cfg.WmSql)

	// 使用配置中的 tenants 字段或查询所有租户
	allTenant := c.parseTenants()
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
	now := time.Now().Add(time.Hour * 10 * -1).UTC()
	res := []string{}
	for i := 0; i < c.Cfg.TotalDataCount; i++ {
		res = append(res, now.Format("2006-01-02"))
		now = now.Add(time.Hour * 24 * -1)
	}

	return res
}

// parseTenants 解析 tenants 字符串，返回租户ID列表
func (c *Controller) parseTenants() []int64 {
	if c.Cfg.Tenants == "" || c.Cfg.Tenants == "all" {
		// 如果是 "all" 或空字符串，返回所有租户
		return alter_common.GetAllTenantWithPlatform(c.Cfg.BasePlatform)
	}

	// 解析逗号分隔的租户ID字符串
	tenantStrs := strings.Split(c.Cfg.Tenants, ",")
	var tenants []int64

	for _, tenantStr := range tenantStrs {
		tenantStr = strings.TrimSpace(tenantStr)
		if tenantStr == "" {
			continue
		}

		tenantId, err := strconv.ParseInt(tenantStr, 10, 64)
		if err != nil {
			// 如果解析失败，跳过这个租户ID
			continue
		}

		tenants = append(tenants, tenantId)
	}

	return tenants
}
