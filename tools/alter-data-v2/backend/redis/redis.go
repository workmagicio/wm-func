package redis

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	client *redis.Client
	once   sync.Once
	ctx    = context.Background()
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() RedisConfig {
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
}

// getEnv 获取环境变量，如果不存在返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetClient 获取Redis客户端单例
func GetClient() *redis.Client {
	once.Do(func() {
		config := GetDefaultConfig()
		client = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
			Password:     config.Password,
			DB:           config.DB,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
			MinIdleConns: 2,
		})

		// 测试连接
		_, err := client.Ping(ctx).Result()
		if err != nil {
			panic(fmt.Sprintf("Redis连接失败: %v", err))
		}

		fmt.Println("Redis连接成功")
	})
	return client
}

// Close 关闭Redis连接
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// GetContext 获取上下文
func GetContext() context.Context {
	return ctx
}
