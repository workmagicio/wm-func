package cache

import "time"

var read_s3_cache_path = "./data/"

type S3Cache struct {
	// 缓存的内容，格式为：租户ID -> 文件路径 -> 最后修改时间
	Files map[string]map[string]time.Time `json:"files"`
}

func NewS3Cache() *S3Cache {
	return &S3Cache{Files: make(map[string]map[string]time.Time)}
}
