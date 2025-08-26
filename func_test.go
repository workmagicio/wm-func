package main

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestBase64(t *testing.T) {
	type TenantConfig struct {
		SfccClientID     string
		SfccClientSecret string
	}

	tenantConfig := TenantConfig{
		SfccClientID:     "eeefa2d0-44a2-451e-964a-f1dd58e90b56",
		SfccClientSecret: "Aas@X8XMJFQ2uEi",
	}

	// Python: auth_string = f"{tenant_config.sfcc_client_id}:{tenant_config.sfcc_client_secret}"
	// Go: 使用 fmt.Sprintf 格式化字符串
	authString := fmt.Sprintf("%s:%s", tenantConfig.SfccClientID, tenantConfig.SfccClientSecret)

	// Python:
	// auth_bytes = auth_string.encode('ascii')
	// auth_b64 = base64.b64encode(auth_bytes).decode('ascii')
	// Go: 使用 base64.StdEncoding.EncodeToString 一步完成
	// 它接收一个字节切片 []byte(authString)，并返回编码后的字符串
	authB64 := base64.StdEncoding.EncodeToString([]byte(authString))

	// 打印结果进行验证
	fmt.Println("原始字符串:", authString)
	fmt.Println("Base64 编码后:", authB64)
}
