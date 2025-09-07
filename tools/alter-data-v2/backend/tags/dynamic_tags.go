package tags

import (
	"fmt"
	"strconv"
	"time"
	"wm-func/tools/alter-data-v2/backend/redis"
)

// TagInfo 标签信息结构
type TagInfo struct {
	TagName    string    `json:"tag_name"`
	TenantId   int64     `json:"tenant_id"`
	Platform   string    `json:"platform"`
	CreateTime time.Time `json:"create_time"`
	ExpireTime time.Time `json:"expire_time"`
}

// AddTagRequest 添加标签请求
type AddTagRequest struct {
	TenantId int64  `json:"tenant_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	TagName  string `json:"tag_name" binding:"required"`
}

// RemoveTagRequest 删除标签请求
type RemoveTagRequest struct {
	TenantId int64  `json:"tenant_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
	TagName  string `json:"tag_name" binding:"required"`
}

const (
	TAG_EXPIRE_DAYS      = 30               // 标签过期天数
	TAG_KEY_PREFIX       = "tags:"          // Redis key前缀
	PLATFORM_TAGS_PREFIX = "platform_tags:" // 平台标签缓存前缀
	PLATFORM_TAGS_EXPIRE = 24 * time.Hour   // 平台标签缓存过期时间
)

// getTagKey 获取Redis中tag的key
func getTagKey(tenantId int64, platform string) string {
	return fmt.Sprintf("%s%d:%s", TAG_KEY_PREFIX, tenantId, platform)
}

// AddTag 添加标签
func AddTag(req AddTagRequest) error {
	client := redis.GetClient()
	ctx := redis.GetContext()

	redisKey := getTagKey(req.TenantId, req.Platform)
	expireTime := time.Now().Add(TAG_EXPIRE_DAYS * 24 * time.Hour)

	// 使用Hash存储：tag_name -> expire_timestamp
	err := client.HSet(ctx, redisKey, req.TagName, expireTime.Unix()).Err()
	if err != nil {
		return fmt.Errorf("添加标签失败: %w", err)
	}

	// 更新平台标签缓存
	go UpdatePlatformTagsCache(req.Platform)

	return nil
}

// RemoveTag 删除标签
func RemoveTag(req RemoveTagRequest) error {
	client := redis.GetClient()
	ctx := redis.GetContext()

	redisKey := getTagKey(req.TenantId, req.Platform)

	err := client.HDel(ctx, redisKey, req.TagName).Err()
	if err != nil {
		return fmt.Errorf("删除标签失败: %w", err)
	}

	// 更新平台标签缓存
	go UpdatePlatformTagsCache(req.Platform)

	return nil
}

// GetDynamicTags 获取指定tenant和platform的动态标签
func GetDynamicTags(tenantId int64, platform string) []string {
	client := redis.GetClient()
	ctx := redis.GetContext()

	redisKey := getTagKey(tenantId, platform)

	// 获取所有标签数据
	tagData, err := client.HGetAll(ctx, redisKey).Result()
	if err != nil {
		fmt.Printf("获取动态标签失败: %v\n", err)
		return []string{}
	}

	currentTime := time.Now().Unix()
	validTags := []string{}
	expiredTags := []string{}

	// 检查每个标签是否过期
	for tagName, expireTimeStr := range tagData {
		expireTime, err := strconv.ParseInt(expireTimeStr, 10, 64)
		if err != nil {
			fmt.Printf("解析过期时间失败 %s: %v\n", tagName, err)
			expiredTags = append(expiredTags, tagName)
			continue
		}

		if expireTime > currentTime {
			validTags = append(validTags, tagName)
		} else {
			expiredTags = append(expiredTags, tagName)
		}
	}

	// 清理过期标签
	if len(expiredTags) > 0 {
		go func() {
			for _, expiredTag := range expiredTags {
				client.HDel(ctx, redisKey, expiredTag)
			}
		}()
	}

	return validTags
}

// GetAllTags 获取tenant的所有标签（默认标签+动态标签）
func GetAllTags(tenantId int64, platform string) []string {
	// 获取默认标签
	defaultTags := GetDefaultTags()
	allTags := []string{}

	if defaultTag, exists := defaultTags[tenantId]; exists {
		allTags = append(allTags, defaultTag)
	}

	// 获取动态标签
	dynamicTags := GetDynamicTags(tenantId, platform)
	allTags = append(allTags, dynamicTags...)

	return allTags
}

// HasAnyTags 检查tenant是否有任何标签
func HasAnyTags(tenantId int64, platform string) bool {
	tags := GetAllTags(tenantId, platform)
	return len(tags) > 0
}

// getPlatformTagsKey 获取平台标签缓存的key
func getPlatformTagsKey(platform string) string {
	return fmt.Sprintf("%s%s", PLATFORM_TAGS_PREFIX, platform)
}

// GetPlatformTags 获取指定平台的所有标签列表（去重、排序）
func GetPlatformTags(platform string) []string {
	client := redis.GetClient()
	ctx := redis.GetContext()

	cacheKey := getPlatformTagsKey(platform)

	// 尝试从缓存获取
	cachedTags, err := client.SMembers(ctx, cacheKey).Result()
	if err == nil && len(cachedTags) > 0 {
		// 缓存命中，排序后返回
		tags := make([]string, len(cachedTags))
		copy(tags, cachedTags)
		return sortTags(tags)
	}

	// 缓存未命中，重新计算
	return UpdatePlatformTagsCache(platform)
}

// UpdatePlatformTagsCache 更新平台标签缓存
func UpdatePlatformTagsCache(platform string) []string {
	client := redis.GetClient()
	ctx := redis.GetContext()

	tagSet := make(map[string]bool)

	// 1. 添加默认标签
	defaultTags := GetDefaultTags()
	for _, defaultTag := range defaultTags {
		if defaultTag != "" {
			tagSet[defaultTag] = true
		}
	}

	// 2. 扫描所有动态标签
	pattern := fmt.Sprintf("%s*:%s", TAG_KEY_PREFIX, platform)
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("扫描标签keys失败: %v\n", err)
	} else {
		currentTime := time.Now().Unix()

		for _, key := range keys {
			// 获取该key下的所有标签
			tagData, err := client.HGetAll(ctx, key).Result()
			if err != nil {
				continue
			}

			// 检查标签是否过期
			for tagName, expireTimeStr := range tagData {
				expireTime, err := strconv.ParseInt(expireTimeStr, 10, 64)
				if err != nil {
					continue
				}

				if expireTime > currentTime && tagName != "" {
					tagSet[tagName] = true
				}
			}
		}
	}

	// 3. 过滤掉 code_filter_region 标签（不在添加标签时显示）
	delete(tagSet, "code_filter_region")

	// 4. 转换为切片并排序
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sortedTags := sortTags(tags)

	// 5. 更新缓存
	cacheKey := getPlatformTagsKey(platform)
	if len(sortedTags) > 0 {
		// 清空旧缓存
		client.Del(ctx, cacheKey)
		// 添加新标签
		args := make([]interface{}, len(sortedTags))
		for i, tag := range sortedTags {
			args[i] = tag
		}
		client.SAdd(ctx, cacheKey, args...)
		// 设置过期时间
		client.Expire(ctx, cacheKey, PLATFORM_TAGS_EXPIRE)
	}

	return sortedTags
}

// sortTags 对标签进行排序（错误标签在前，普通标签在后，各自按字母排序）
func sortTags(tags []string) []string {
	if len(tags) <= 1 {
		return tags
	}

	errorTags := []string{}
	normalTags := []string{}

	for _, tag := range tags {
		if len(tag) > 4 && tag[:4] == "err_" {
			errorTags = append(errorTags, tag)
		} else {
			normalTags = append(normalTags, tag)
		}
	}

	// 简单排序（Go的sort包在这里不方便导入，使用简单的冒泡排序）
	bubbleSort := func(arr []string) {
		n := len(arr)
		for i := 0; i < n-1; i++ {
			for j := 0; j < n-i-1; j++ {
				if arr[j] > arr[j+1] {
					arr[j], arr[j+1] = arr[j+1], arr[j]
				}
			}
		}
	}

	bubbleSort(errorTags)
	bubbleSort(normalTags)

	// 错误标签在前，普通标签在后
	result := make([]string, 0, len(tags))
	result = append(result, errorTags...)
	result = append(result, normalTags...)

	return result
}
