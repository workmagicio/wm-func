package main

import (
	"fmt"
	"log"
	"os"
	"time"
	lock2 "wm-func/common/lock"
	t_pool "wm-func/common/pool"
	"wm-func/wm_account"
)

const Platform = "knocommerce"

func main() {
	log.Printf("[%s] Knocommerce数据同步程序启动", Platform)

	accounts := wm_account.GetAccountsWithPlatform(Platform)
	log.Printf("[%s] 获取到账户数量: %d", Platform, len(accounts))

	lock := lock2.NewMySQLLocker()
	var kaccounts = []KAccount{}
	for _, account := range accounts {
		kaccounts = append(kaccounts, KAccount{
			account,
			lock,
		})
	}

	pool := t_pool.NewWorkerPool(10)
	pool.Run()

	for _, account := range kaccounts {
		ac := account
		lockKey := fmt.Sprintf("knocommerce:%d:%s", ac.TenantId, ac.AccountId)
		ownerID := fmt.Sprintf("process-%d", os.Getpid())
		// 先写死180分钟
		lockDuration := 180 * time.Minute

		err := ac.TryLock(lockKey, ownerID, lockDuration)
		if err == nil {
			// 成功获取锁
			pool.AddTask(func() {
				run(ac)
				// 任务完成后释放锁
				ac.Unlock(lockKey, ownerID)
			})
		} else {
			// 获取锁失败
			log.Printf("[%s] 无法获取锁，跳过该账户: %v", ac.GetSimpleTraceId(), err)
		}
	}
	pool.Wait()

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

	responseCountTraceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE_COUNT)
	log.Printf("[%s] 开始RequestResponseCount", responseCountTraceId)
	RequestResponseCount(account, token)
	log.Printf("[%s] RequestResponseCount完成", responseCountTraceId)

	responseTraceId := account.GetTraceIdWithSubType(SUBTYPE_RESPONSE)
	log.Printf("[%s] 开始RequestResponse", responseTraceId)
	RequestResponse(account, token)
	log.Printf("[%s] RequestResponse完成", responseTraceId)

	log.Printf("[%s] 账户处理完成", traceId)
}
