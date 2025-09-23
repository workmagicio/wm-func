package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func loadCache(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	return data
}

func saveCache(data []byte, tenantId int64) {
	path := GetFilePathWithTenantId(tenantId)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		panic(err)
	}
}

func IsNeedUpdate(cache *S3Cache, tenantId int64, key string, lastModified time.Time) bool {
	if cache.Files == nil {
		return true
	}

	tenantFiles, exists := cache.Files[fmt.Sprintf("%d", tenantId)]
	if !exists {
		return true
	}

	// 获取缓存中的最后修改时间
	cachedTime, exists := tenantFiles[key]
	if !exists {
		return true
	}

	// 如果S3中的文件比缓存中的新，需要下载
	return lastModified.After(cachedTime)
}

func GetFilePathWithTenantId(tenantId int64) string {
	return filepath.Join(read_s3_cache_path, fmt.Sprintf("s3_cache_%d.json", tenantId))

}

func LoadS3Cache(tenantId int64) *S3Cache {
	cache := &S3Cache{
		Files: make(map[string]map[string]time.Time),
	}

	data := loadCache(GetFilePathWithTenantId(tenantId))

	if err := json.Unmarshal(data, &cache); err != nil {
		// 如果解析失败，返回空缓存
		return cache
	}

	return cache
}

func SaveS3Cache(cache *S3Cache, tenantID int64, key string, lastModified time.Time) {
	if cache.Files == nil {
		cache.Files = make(map[string]map[string]time.Time)
	}

	tenantIdStr := fmt.Sprintf("%d", tenantID)

	if cache.Files[tenantIdStr] == nil {
		cache.Files[tenantIdStr] = make(map[string]time.Time)
	}

	cache.Files[tenantIdStr][key] = lastModified

	b, _ := json.Marshal(cache)
	saveCache(b, tenantID)
}

type Cache struct {
	Key          string
	LastModified time.Time
}

func SaveS3CacheWithArr(cache *S3Cache, tenantId int64, arr []Cache) {
	if cache.Files == nil {
		cache.Files = make(map[string]map[string]time.Time)
	}

	for _, ch := range arr {
		tenantIdStr := fmt.Sprintf("%d", tenantId)

		if cache.Files[tenantIdStr] == nil {
			cache.Files[tenantIdStr] = make(map[string]time.Time)
		}

		cache.Files[tenantIdStr][ch.Key] = ch.LastModified
	}

	b, _ := json.Marshal(cache)
	saveCache(b, tenantId)
}
