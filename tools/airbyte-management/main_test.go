package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestUpdateConnectionStatus(t *testing.T) {
	connections := GetAllConnections()
	prefix := "tiktok_marketing_"
	name := "GMV-Max"

	for _, conn := range connections {
		if conn.Prefix != prefix || !strings.Contains(conn.Name, name) {
			continue
		}

		if conn.Status != "inactive" {
			continue
		}

		fmt.Println(conn)
		UpdateConnectionStatus(conn.ConnectionId, "active")
	}
}

func TestUpdateConnectionStatusActive(t *testing.T) {
	// 示例：激活一个 connection
	connectionId := "your-connection-id-here"
	UpdateConnectionStatus(connectionId, "active")
}

func TestUpdateConnectionStatusInactive(t *testing.T) {
	// 示例：停用一个 connection
	connectionId := "your-connection-id-here"
	UpdateConnectionStatus(connectionId, "inactive")
}

func TestUpdateConnectionCronExpression(t *testing.T) {
	connections := GetAllConnections()
	prefix := "tiktok_marketing_"
	name := "GMV-Max"

	for _, conn := range connections {
		if conn.Prefix != prefix || !strings.Contains(conn.Name, name) {
			continue
		}
		if conn.Status != "inactive" {
			continue
		}

		fmt.Println(conn)
		UpdateSetting(conn.ConnectionId, fmt.Sprintf("0 %d %d * * ?", rand.Intn(59), rand.Intn(23)))

	}
}

//curl 'http://internal-airbytes.workmagic.io/api/v1/web_backend/connections/update' \
//-H 'Accept: */*' \
//-H 'Accept-Language: zh-CN,zh;q=0.9' \
//-H 'Cache-Control: no-cache' \
//-b 'intercom-id-oe2kqapl=a4e37469-24f4-4f73-9e18-dc179a8a60f0; intercom-device-id-oe2kqapl=a5b4fc31-1c2b-48de-ae9c-7413b2e3bfbf; wm_client_id=wm.6ejfa4madub.1719805732; _ga=GA1.1.wm.6ejfa4madub.1719805732; _ga_HDBMVFQGBH=GS1.1.1720529831.2.0.1720529831.0.0.0; hubspotutk=ce588d9f8ef9f115719855aaa17d1644; _gcl_au=1.1.1966978318.1747636076.1120573105.1748275619.1748275621; __hstc=266511057.ce588d9f8ef9f115719855aaa17d1644.1747636080850.1748351580323.1748936291268.13; _ga_QXWRYC1ZJY=GS2.1.s1750404697$o1005$g0$t1750404697$j60$l0$h0; ajs_user_id=1bcb346c-c895-480a-9bcf-b77ecd57e4ff; ajs_anonymous_id=1281cbb5-d836-4234-a5cc-24d1f005c760' \
//-H 'Origin: http://internal-airbytes.workmagic.io' \
//-H 'Pragma: no-cache' \
//-H 'Proxy-Connection: keep-alive' \
//-H 'Referer: http://internal-airbytes.workmagic.io/workspaces/1bcb346c-c895-480a-9bcf-b77ecd57e4ff/connections?search=gmv' \
//-H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36' \
//-H 'content-type: application/json' \
//-H 'x-airbyte-analytic-source: webapp' \
//-H 'x-api-key: C7986gw4eZp8sr34QuBd' \
//--data-raw '{"connectionId":"{connection_id}","status":"active"}' \
//--insecure
