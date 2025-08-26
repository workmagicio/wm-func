package backend

import (
	"fmt"
	"testing"
	"wm-func/tools/alter-data-v2/backend/bcache"
	"wm-func/tools/alter-data-v2/backend/bmodel"
)

func TestRun(t *testing.T) {
	res := bmodel.GetDataWithPlatform("googleAds")
	fmt.Println("Original data:", len(res), "records")

	// 保存缓存
	if err := bcache.SaveCache(PLATFORM_GOOGLE, res); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}
	fmt.Println("✅ 缓存已保存")

	// 加载缓存
	c, err := bcache.LoadCache(PLATFORM_GOOGLE)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	fmt.Printf("缓存创建时间: %v\n", c.CreateTime)

	// 类型安全的加载（推荐）
	typedData, err := bcache.LoadTyped[[]bmodel.ApiData](PLATFORM_GOOGLE)
	if err != nil {
		t.Fatalf("Failed to load typed cache: %v", err)
	}
	fmt.Println("✅ 加载成功，记录数:", len(typedData))
}
