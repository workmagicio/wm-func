package airbyte_db

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"sync"
	"time"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB 初始化 MySQL 数据库连接池
func InitDB() {
	//application.service.integration.airbyte.default.destination.mysql
	once.Do(func() {
		type DBConfig struct {
			Name          string `json:"name"`
			WorkspaceId   string `json:"workspaceId"`
			Configuration struct {
				DestinationType   string `json:"destinationType"`
				Host              string `json:"host"`
				Port              int    `json:"port"`
				Username          string `json:"username"`
				Password          string `json:"password"`
				Database          string `json:"database"`
				RawDataSchema     string `json:"raw_data_schema"`
				WmTenantId        string `json:"wm_tenant_id"`
				Ssl               bool   `json:"ssl"`
				DisableTypeDedupe bool   `json:"disable_type_dedupe"`
			} `json:"configuration"`
		}

		//cfg := apollo.GetAirbyteMysqlConfig()
		cfg := DBConfig{}
		if err := json.Unmarshal([]byte(`{
    "name": "MySQL {tenantId}",
    "workspaceId": "{workspaceId}",
    "definitionId": "7ff16f4f-ff86-4330-b6b0-7a28f1223570",
    "configuration": {
        "destinationType": "mysql",
        "host": "internal-adb.workmagic.io",
        "port": 3306,
        "username": "airbyte_06x",
        "password": "fAtnYwwPugw2gpq3",
        "database": "airbyte_destination_v2",
        "raw_data_schema": "airbyte_destination_v2",
        "wm_tenant_id": "{tenantId}",
        "ssl": false,
        "disable_type_dedupe": true
    }
}`), &cfg); err != nil {
			panic(err)
		}

		var err error
		// 使用 gorm.Open() 和 mysql.Open() 连接 MySQL 数据库
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Configuration.Username, cfg.Configuration.Password, cfg.Configuration.Host, cfg.Configuration.Port, cfg.Configuration.Database),
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
