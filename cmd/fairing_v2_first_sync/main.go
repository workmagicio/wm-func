package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	lock2 "wm-func/common/lock"
	"wm-func/wm_account"
)

func main() {
	// 健康检查端点
	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("fairing-v2-first-sync service is running"))
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
	log.Printf("[%s] Fairing数据同步程序启动", Platform)

	accounts := wm_account.GetFairingAccounts()
	log.Printf("[%s] 获取到账户数量: %d", Platform, len(accounts))

	lock := lock2.NewMySQLLocker()
	var fAccounts = []FAccount{}
	for _, account := range accounts {
		if account.TenantId != tenantId {
			continue
		}
		fAccounts = append(fAccounts, FAccount{
			Account: account,
			Locker:  lock,
		})
	}

	for _, account := range fAccounts {
		ac := account
		run(ac)
	}

	log.Printf("[%s] Fairing数据同步程序结束", Platform)
}

func run(account FAccount) {
	traceId := account.GetSimpleTraceId()
	log.Printf("[%s] 开始处理账户", traceId)

	RequestQuestion(account)

	log.Printf("[%s] 账户处理完成", traceId)
}
