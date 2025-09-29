package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Name           string `json:"name"`
	BasePlatform   string `json:"base_platform"`
	ApiSql         string `json:"api_data_query"`
	WmSql          string `json:"wm_data_query"`
	TotalDataCount int    `json:"total_data_count"`
}

func GetConfit() map[string]Config {
	b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/tools/alter-v3/config.json")
	if err != nil {
		panic(err)
	}
	configs := []Config{}
	if err = json.Unmarshal(b, &configs); err != nil {
		panic(err)
	}

	var result = map[string]Config{}
	for _, config := range configs {
		if config.TotalDataCount == 0 {
			config.TotalDataCount = 75
		}

		result[config.Name] = config
	}

	return result
}
