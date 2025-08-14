package main

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Result struct {
	TenantId int64     `gorm:"column:tenant_id"`
	RawDate  time.Time `gorm:"column:event_date"` // 2025-07-15
	AdSpend  float64   `gorm:"column:ad_spend"`
}

func query(sql string, db *gorm.DB) []Result {
	result := make([]Result, 0)
	fmt.Println(sql)
	if err := db.Raw(sql).Scan(&result).Limit(-1).Error; err != nil {
		panic(err)
	}
	return result
}
