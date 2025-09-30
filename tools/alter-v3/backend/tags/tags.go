package tags

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

var user_tag_path = "./user_tag.json"

var default_filter_tenants = []int64{
	133822, 133849, 134531, 150076, 150075, 150078, 150079, 150080, 150081, 150082, 150083,
}

// 内存缓存
var (
	userTagsCache     []UserTags
	userTagsCacheLock sync.RWMutex
	cacheInitialized  bool
)

func GetDefaultTags() map[int64]string {
	res := map[int64]string{}
	for i := 0; i < len(default_filter_tenants); i++ {
		res[default_filter_tenants[i]] = "code_filter_region"
	}
	return res
}

type UserTags struct {
	Key       string // tenantId|platform|name
	TenantId  string
	Platform  string
	Name      string
	ValidTime time.Time
}

func (userTag UserTags) GetKey() string {
	return userTag.TenantId + "|" + userTag.Platform + "|" + userTag.Name
}

// 初始化缓存
func initCache() {
	if cacheInitialized {
		return
	}

	userTagsCacheLock.Lock()
	defer userTagsCacheLock.Unlock()

	if cacheInitialized {
		return
	}

	userTagsCache = readUserTagsFromFile()
	cacheInitialized = true
}

func AddUserTag(userTag UserTags) {
	initCache()

	userTagsCacheLock.Lock()
	defer userTagsCacheLock.Unlock()

	// 设置Key
	userTag.Key = userTag.GetKey()

	// 需要判断是否存在，如果存在，则更新
	found := false
	for i, tag := range userTagsCache {
		if tag.Key == userTag.Key {
			userTagsCache[i] = userTag
			found = true
			break
		}
	}

	if !found {
		userTagsCache = append(userTagsCache, userTag)
	}

	// 写入文件
	writeUserTagsToFile(userTagsCache)
}

func RemoveUserTag(key string) {
	initCache()

	userTagsCacheLock.Lock()
	defer userTagsCacheLock.Unlock()

	for i, userTag := range userTagsCache {
		if userTag.Key == key {
			userTagsCache = append(userTagsCache[:i], userTagsCache[i+1:]...)
			break
		}
	}

	// 写入文件
	writeUserTagsToFile(userTagsCache)
}

func UpdateUserTag(userTag UserTags) {
	// 更新操作与添加操作相同
	AddUserTag(userTag)
}

func GetAllUserTags() []UserTags {
	initCache()

	userTagsCacheLock.RLock()
	defer userTagsCacheLock.RUnlock()

	// 过滤掉已过期的标签
	validTags := []UserTags{}
	for _, userTag := range userTagsCache {
		if userTag.ValidTime.IsZero() || userTag.ValidTime.After(time.Now()) {
			validTags = append(validTags, userTag)
		}
	}

	// 异步清理过期标签
	go cleanExpiredUserTags()

	return validTags
}

// 根据 TenantId 和 Platform 获取标签
func GetUserTagsByTenantAndPlatform(tenantId, platform string) []string {
	initCache()

	userTagsCacheLock.RLock()
	defer userTagsCacheLock.RUnlock()

	tags := []string{}
	for _, userTag := range userTagsCache {
		if userTag.TenantId == tenantId && userTag.Platform == platform {
			if userTag.ValidTime.IsZero() || userTag.ValidTime.After(time.Now()) {
				tags = append(tags, userTag.Name)
			}
		}
	}

	return tags
}

// 从文件读取（仅在初始化时使用）
func readUserTagsFromFile() []UserTags {
	b, err := os.ReadFile(user_tag_path)
	if err != nil {
		// 如果文件不存在，返回空数组
		if os.IsNotExist(err) {
			return []UserTags{}
		}
		// 其他读取错误也返回空数组，不影响系统运行
		return []UserTags{}
	}

	// 如果文件为空，返回空数组
	if len(b) == 0 {
		return []UserTags{}
	}

	userTags := []UserTags{}
	if err = json.Unmarshal(b, &userTags); err != nil {
		// JSON 解析失败，返回空数组
		return []UserTags{}
	}

	return userTags
}

// 写入文件（在修改缓存后调用）
func writeUserTagsToFile(userTags []UserTags) {
	bt, err := json.Marshal(userTags)
	if err != nil {
		panic(err)
	}

	os.WriteFile(user_tag_path, bt, os.ModePerm)
}

// 清理过期的标签（异步执行）
func cleanExpiredUserTags() {
	userTagsCacheLock.Lock()
	defer userTagsCacheLock.Unlock()

	validTags := []UserTags{}
	hasExpired := false

	for _, userTag := range userTagsCache {
		if userTag.ValidTime.IsZero() || userTag.ValidTime.After(time.Now()) {
			validTags = append(validTags, userTag)
		} else {
			hasExpired = true
		}
	}

	// 如果有过期的标签，更新缓存和文件
	if hasExpired {
		userTagsCache = validTags
		writeUserTagsToFile(userTagsCache)
	}
}

// 兼容旧方法（已废弃，但保留以防其他地方调用）
func readUserTags() []UserTags {
	return readUserTagsFromFile()
}

func writeUserTags(userTags []UserTags) {
	writeUserTagsToFile(userTags)
}
