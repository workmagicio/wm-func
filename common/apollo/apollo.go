package apollo

import (
	"encoding/json"
	"github.com/philchia/agollo/v4"
	"log"
	"os"
)

func init() {
	Init()
}

func Init() {
	errr := agollo.Start(&agollo.Conf{
		AppID:           "platform-api",
		Cluster:         "UAT",
		NameSpaceNames:  []string{"application", "datasource"},
		MetaAddr:        "http://internal-apollo-meta-server-preview.workmagic.io",
		AccesskeySecret: "88b67fe6bcde46e59359f41ea9f3cd07",
	}, agollo.WithLogger(&logger{log: log.New(os.Stdout, "[agollo] ", log.LstdFlags)}))
	if errr != nil {
		panic(errr)
	}
}

func GetS3Config() S3Config {
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

func GetDevelopS3Config() S3Config {
	val := agollo.GetString("application.service.aws.iam.develop")
	res := S3Config{}
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		panic(err)
	}
	return res
}

func GetAirbyteMysqlConfig() DBConfig {
	res := agollo.GetString("application.service.integration.airbyte.cluster.destination.mysql")

	cfg := DBConfig{}
	err := json.Unmarshal([]byte(res), &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func GetPinterestSourceSetting() string {
	return agollo.GetString("application.service.integration.airbyte.source.pinterest")
}

func GetXkMysqlConfig() DBConfig {
	res := agollo.GetString("application.service.integration.xk.mysql.conf")

	cfg := DBConfig{}
	err := json.Unmarshal([]byte(res), &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
