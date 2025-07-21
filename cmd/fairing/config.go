package main

const (
	Platform   = "fairing"
	MaxWorkers = 3 // 减少worker数量，为多实例留出资源
)

var subTypes = []string{"question", "response"}

type FairingConfig struct {
	BatchSize         int // 数据库批量插入大小
	ResponsesPageSize int // Responses API每页大小
	RateLimit         int // requests per second
	MaxPages          int // 最大页数保护，避免无限循环
	TimeoutSecs       int // API请求超时时间

	// Stream Slice 相关配置
	MaxSlicesPerRun int // 每次运行最多处理的slice数量
	SliceDays       int // 每个slice的天数
}

func getFairingConfig() FairingConfig {
	return FairingConfig{
		BatchSize:         500,
		ResponsesPageSize: 100, // Responses API最大每页100条
		RateLimit:         10,
		MaxPages:          1000, // 防止无限分页循环
		TimeoutSecs:       30,

		// Stream Slice 配置
		MaxSlicesPerRun: 800, // 每次最多处理800个slice
		SliceDays:       1,   // 每个slice为1天
	}
}

// 实例配置
type InstanceConfig struct {
	InstanceId    string // 实例标识，用于日志区分
	MaxRetries    int    // 最大重试次数
	RetryInterval int    // 重试间隔（秒）
}

func getInstanceConfig() InstanceConfig {
	return InstanceConfig{
		InstanceId:    generateInstanceId(),
		MaxRetries:    3,
		RetryInterval: 5,
	}
}

// 生成实例ID - 实际实现在libs.go中
