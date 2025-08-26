package main

import (
	"log"
	t_pool "wm-func/common/pool"
	"wm-func/wm_account"
)

const Platform = "knocommerce"

func main() {
	log.Printf("[%s] Knocommerce数据同步程序启动", Platform)

	accounts := wm_account.GetAccountsWithPlatform(Platform)
	log.Printf("[%s] 获取到账户数量: %d", Platform, len(accounts))

	pool := t_pool.NewWorkerPool(10)
	pool.Run()

	for _, account := range accounts {
		ac := account
		pool.AddTask(func() {
			run(ac)
		})

	}
	pool.Wait()

	log.Printf("[%s] Knocommerce数据同步程序结束", Platform)
}

func run(account wm_account.Account) {
	traceId := account.GetTraceId()
	log.Printf("[%s] 开始处理账户", traceId)

	token := NewTokenManager(account)
	//token, err := RefreshToken(account)
	//if err != nil {
	//	log.Printf("[%s] RefreshToken失败: %v", traceId, err)
	//	return
	//}
	//log.Printf("[%s] RefreshToken成功", traceId)

	log.Printf("[%s] 开始RequestQuestion", traceId)
	RequestQuestion(account, token)
	log.Printf("[%s] RequestQuestion完成", traceId)

	log.Printf("[%s] 开始RequestSurvey", traceId)
	RequestSurvey(account, token)
	log.Printf("[%s] RequestSurvey完成", traceId)

	log.Printf("[%s] 开始RequestResponseCount", traceId)
	RequestResponseCount(account, token)
	log.Printf("[%s] RequestResponseCount完成", traceId)

	log.Printf("[%s] 开始RequestResponse", traceId)
	RequestResponse(account, token)
	log.Printf("[%s] RequestResponse完成", traceId)

	log.Printf("[%s] 账户处理完成", traceId)
}
