package bdao

import (
	"fmt"
	"testing"
)

func TestGetApiDataByPlatform(t *testing.T) {
	platform := "googleAds"

	// 第一次调用 - 从DB获取并缓存
	fmt.Println("=== 第一次调用 (isNeedRefresh=false) ===")
	data1 := GetApiDataByPlatform(false, platform)
	fmt.Printf("获取到 %d 条记录\n", len(data1))

	// 第二次调用 - 从缓存获取
	fmt.Println("\n=== 第二次调用 (isNeedRefresh=false) - 应该使用缓存 ===")
	data2 := GetApiDataByPlatform(false, platform)
	fmt.Printf("获取到 %d 条记录\n", len(data2))

	// 第三次调用 - 强制刷新
	fmt.Println("\n=== 第三次调用 (isNeedRefresh=true) - 强制刷新 ===")
	data3 := GetApiDataByPlatform(true, platform)
	fmt.Printf("获取到 %d 条记录\n", len(data3))

	if len(data1) != len(data2) || len(data2) != len(data3) {
		t.Errorf("数据长度不一致: %d, %d, %d", len(data1), len(data2), len(data3))
	}
}

func TestGetOverviewDataByPlatform(t *testing.T) {
	platform := "googleAds"

	// 第一次调用 - 从DB获取并缓存
	fmt.Println("=== Overview数据第一次调用 (isNeedRefresh=false) ===")
	data1 := GetOverviewDataByPlatform(false, platform)
	fmt.Printf("获取到 %d 条记录\n", len(data1))

	// 第二次调用 - 从缓存获取
	fmt.Println("\n=== Overview数据第二次调用 (isNeedRefresh=false) - 应该使用缓存 ===")
	data2 := GetOverviewDataByPlatform(false, platform)
	fmt.Printf("获取到 %d 条记录\n", len(data2))

	if len(data1) != len(data2) {
		t.Errorf("Overview数据长度不一致: %d, %d", len(data1), len(data2))
	}
}
