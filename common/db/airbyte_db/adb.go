package airbyte_db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
	"wm-func/common/apollo"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB 初始化 MySQL 数据库连接池
func InitDB() {
	//application.service.integration.airbyte.default.destination.mysql
	once.Do(func() {
		cfg := apollo.GetAirbyteMysqlConfig()

		var err error
		// 使用 gorm.Open() 和 mysql.Open() 连接 MySQL 数据库
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Name, cfg.Password, cfg.Host, 3306, "airbyte_destination_v2"),
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent), // 禁用所有日志
		})
		if err != nil {
			log.Fatalf("MySQL 连接失败: %v", err)
		}

		// 配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("获取数据库连接失败: %v", err)
		}
		sqlDB.SetMaxOpenConns(100)                 // 最大连接数
		sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
		sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接的最大存活时间
	})
}

// GetDB 返回数据库实例
func GetDB() *gorm.DB {
	InitDB()
	return db
}
