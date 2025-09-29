package main

import (
	"log"
	t_pool "wm-func/common/pool"
	"wm-func/wm_account"
)

const (
	Platform   = "facebookMarketing"
	SubType    = "ad_metrics"
	MaxWorkers = 10
)

func main() {
	log.Println("start run meta ad metrics sync...")
	run()
	log.Println("end run meta ad metrics sync...")
}

func run() {
	accounts := wm_account.GetAccountsWithPlatform(Platform)
	log.Printf("start total accounts: %d", len(accounts))

	if len(accounts) == 0 {
		log.Println("没有找到 Meta 广告账户，程序退出")
		return
	}

	pool := t_pool.NewWorkerPool(MaxWorkers)
	pool.Run()
	defer pool.Close()

	var tenants = map[int64]bool{
		//150133: true,
		150110: true,
		//150198: true,
	}

	for _, account := range accounts {
		if !tenants[account.TenantId] {
			continue
		}
		//150161
		//150198
		//150198
		// 避免闭包问题，复制account变量
		acc := account
		pool.AddTask(func() {
			log.Printf("[%s] start exec account", acc.GetTraceId())
			err := exec(acc)
			if err != nil {
				log.Printf("[%s] exec account failed: %v", acc.GetTraceId(), err)
				return
			}
			log.Printf("[%s] end exec account", acc.GetTraceId())
		})
	}

	pool.Wait()
}

func exec(account wm_account.Account) error {
	// 1. 获取同步状态
	//syncState, err := getState(account)
	//if err != nil {
	//	log.Printf("[%s] 获取同步状态失败: %v", account.GetTraceId(), err)
	//	return err
	//}

	// 2. 检查是否需要同步（这里可以根据你的业务逻辑调整）
	//if syncState.UpdatedAt.Add(time.Hour).After(time.Now().UTC()) &&
	//	syncState.Status == STATUS_SUCCESS {
	//	log.Printf("[%s] 同步时间小于1小时，跳过", account.GetTraceId())
	//	return nil
	//}

	// 3. 同步广告数据（这里你可以添加自己的逻辑）
	err := syncAdMetrics(account, SyncState{})
	if err != nil {
		//syncState.Status = STATUS_FAILED
		//syncState.Message = err.Error()
		//_ = updateSyncState(account, syncState)

	}
	return err
}
