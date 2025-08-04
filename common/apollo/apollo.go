package apollo

import (
	"encoding/json"
	"github.com/philchia/agollo/v4"
	"log"
	"os"
	"sync"
)

// ApolloClient 单例结构体
type ApolloClient struct {
	// 可以添加其他需要的字段
}

var (
	instance *ApolloClient
	once     sync.Once
)

// GetInstance 获取 Apollo 单例实例
func GetInstance() *ApolloClient {
	once.Do(func() {
		instance = &ApolloClient{}
		instance.init()
	})
	return instance
}

// init 初始化 Apollo 连接
func (c *ApolloClient) init() {
	errr := agollo.Start(&agollo.Conf{
		AppID:           "platform-api",
		Cluster:         "UAT",
		NameSpaceNames:  []string{"application", "datasource"},
		MetaAddr:        "http://internal-apollo-meta-server-preview.workmagic.io",
		AccesskeySecret: "88b67fe6bcde46e59359f41ea9f3cd07",
		CacheDir:        "/tmp",
	}, agollo.WithLogger(&logger{log: log.New(os.Stdout, "[agollo] ", log.LstdFlags)}))
	if errr != nil {
		panic(errr)
	}
}

// GetS3Config 获取 S3 配置
func (c *ApolloClient) GetS3Config() S3Config {
	key := agollo.GetString("application.service.aws.workmagicTatariS3Key")
	secret := agollo.GetString("application.service.aws.workmagicTatariS3Secret")

	//val := agollo.GetString("application.service.aws.iam.develop")
	res := S3Config{
		AccessKeyID:     key,
		SecretAccessKey: secret,
		Region:          "us-east-1",
	}

	return res
}

// GetDevelopS3Config 获取开发环境 S3 配置
func (c *ApolloClient) GetDevelopS3Config() S3Config {
	val := agollo.GetString("application.service.aws.iam.develop")
	res := S3Config{}
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		panic(err)
	}
	return res
}

// GetAirbyteMysqlConfig 获取 Airbyte MySQL 配置
func (c *ApolloClient) GetAirbyteMysqlConfig() MysqlConfig {
	host := agollo.GetString("airbyte.datasource.api.url", withDataSource())
	name := agollo.GetString("airbyte.datasource.api.name", withDataSource())
	password := agollo.GetString("airbyte.datasource.api.password", withDataSource())
	return MysqlConfig{
		Host:     host,
		Name:     name,
		Password: password,
	}
}

// GetPinterestSourceSetting 获取 Pinterest 源设置
func (c *ApolloClient) GetPinterestSourceSetting() string {
	return agollo.GetString("application.service.integration.airbyte.source.pinterest")
}

// GetXkMysqlConfig 获取 XK MySQL 配置
func (c *ApolloClient) GetXkMysqlConfig() DBConfig {
	res := agollo.GetString("application.service.integration.xk.mysql.conf")

	cfg := DBConfig{}
	err := json.Unmarshal([]byte(res), &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
func withDataSource() agollo.OpOption {
	return agollo.WithNamespace("datasource")
}

func (c *ApolloClient) GetMysqlConfig() MysqlConfig {
	host := agollo.GetString("gcs_rw.datasource.api.url", withDataSource())
	name := agollo.GetString("gcs_rw.datasource.api.name", withDataSource())
	password := agollo.GetString("gcs_rw.datasource.api.password", withDataSource())
	return MysqlConfig{
		Host:     host,
		Name:     name,
		Password: password,
	}
}

// 为了保持向后兼容，提供全局函数
func GetS3Config() S3Config {
	return GetInstance().GetS3Config()
}

func GetDevelopS3Config() S3Config {
	return GetInstance().GetDevelopS3Config()
}

func GetAirbyteMysqlConfig() MysqlConfig {
	return GetInstance().GetAirbyteMysqlConfig()
}

func GetPinterestSourceSetting() string {
	return GetInstance().GetPinterestSourceSetting()
}

func GetXkMysqlConfig() DBConfig {
	return GetInstance().GetXkMysqlConfig()
}

func GetMysqlConfig() MysqlConfig {
	return GetInstance().GetMysqlConfig()
}
