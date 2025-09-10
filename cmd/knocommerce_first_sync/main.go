package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

const Platform = "knocommerce"

func main() {
	// 健康检查端点
	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("knocommerce-first-sync service is running"))
	})

	// 主业务端点
	http.HandleFunc("/run/", func(writer http.ResponseWriter, r *http.Request) {
		tenantId := r.URL.Path[len("/run/"):]
		tenantIdInt64, err := strconv.ParseInt(tenantId, 10, 64)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		RunWithTenantId(tenantIdInt64)
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("success"))
	})

	// 从环境变量获取端口，Cloud Run 默认使用 PORT=8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // 本地开发时的默认端口
	}

	log.Printf("服务器启动，监听端口 :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func RunWithTenantId(tenantId int64) {
	log.Printf("[%s] Knocommerce数据同步程序启动", Platform)

	accounts := wm_account.GetAccountsWithPlatform(Platform)
	log.Printf("[%s] 获取到账户数量: %d", Platform, len(accounts))

	lock := lock2.NewMySQLLocker()
	var kaccounts = []KAccount{}
	for _, account := range accounts {
		if account.TenantId != tenantId {
			continue
		}
		kaccounts = append(kaccounts, KAccount{
			account,
			lock,
		})
	}

	for _, account := range kaccounts {
		ac := account
		run(ac)
	}

	log.Printf("[%s] Knocommerce数据同步程序结束", Platform)
}

func run(account KAccount) {
	traceId := account.GetSimpleTraceId()
	log.Printf("[%s] 开始处理账户", traceId)

	token := NewTokenManager(account)
	//token, err := RefreshToken(account)
	//if err != nil {
	//	log.Printf("[%s] RefreshToken失败: %v", traceId, err)
	//	return
	//}
	//log.Printf("[%s] RefreshToken成功", traceId)

	questionTraceId := account.GetTraceIdWithSubType(SUBTYPE_QUESTION)
	log.Printf("[%s] 开始RequestQuestion", questionTraceId)
	RequestQuestion(account, token)
	log.Printf("[%s] RequestQuestion完成", questionTraceId)

	surveyTraceId := account.GetTraceIdWithSubType(SUBTYPE_SURVEY)
	log.Printf("[%s] 开始RequestSurvey", surveyTraceId)
	RequestSurvey(account, token)
	log.Printf("[%s] RequestSurvey完成", surveyTraceId)

	//responseCountTraceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE_COUNT)
	//log.Printf("[%s] 开始RequestResponseCount", responseCountTraceId)
	//RequestResponseCount(account, token)
	//log.Printf("[%s] RequestResponseCount完成", responseCountTraceId)
	//
	//responseTraceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE)
	//log.Printf("[%s] 开始RequestResponse", responseTraceId)
	//RequestResponse(account, token)
	//log.Printf("[%s] RequestResponse完成", responseTraceId)

	log.Printf("[%s] 账户处理完成", traceId)
}
