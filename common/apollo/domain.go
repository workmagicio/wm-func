package apollo

type S3Config struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
}

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

type MysqlConfig struct {
	Host     string
	Name     string
	Password string
}

type LLMConfig struct {
	Key     string
	BaseUrl string
}
