package main

import (
	"math/rand"
	"time"
)

func GetCheckTenantIds() []string {
	all := GetTenantInfo()

	// 按类型分组
	oldTenants := []TenantInfo{}
	newTenants := []TenantInfo{}

	for _, tenant := range all {
		if tenant.TenantType == "old" {
			oldTenants = append(oldTenants, tenant)
		} else {
			newTenants = append(newTenants, tenant)
		}
	}

	// 初始化随机种子
	rand.Seed(time.Now().UnixNano())

	result := []string{}

	// 老客户随机取3个
	selectedOld := randomSelectTenants(oldTenants, 3)
	for _, tenant := range selectedOld {
		result = append(result, tenant.TenantID)
	}

	// 新客户随机取5个
	selectedNew := randomSelectTenants(newTenants, 5)
	for _, tenant := range selectedNew {
		result = append(result, tenant.TenantID)
	}

	return result
}

// 随机选择指定数量的租户
func randomSelectTenants(tenants []TenantInfo, count int) []TenantInfo {
	if len(tenants) == 0 {
		return []TenantInfo{}
	}

	// 如果租户数量小于等于需要的数量，直接返回全部
	if len(tenants) <= count {
		return tenants
	}

	// 随机选择
	selected := make([]TenantInfo, 0, count)
	indices := rand.Perm(len(tenants))

	for i := 0; i < count; i++ {
		selected = append(selected, tenants[indices[i]])
	}

	return selected
}

// xie
