package main

import (
	"log"
	"time"
	t_pool "wm-func/common/pool"
	"wm-func/wm_account"
)

func main() {
	//checkApollo()

	log.Println("start run...")
	run()
	log.Println("end run...")

	//ticker := time.NewTicker(SyncInterval * time.Second)
	//defer ticker.Stop()
	//
	//for {
	//	select {
	//	case <-ticker.C:
	//		run()
	//	}
	//}
}

func run() {
	accounts := wm_account.GetShopifyAccount()
	pool := t_pool.NewWorkerPool(MaxWorkers)
	pool.Run()
	defer pool.Close()

	for _, account := range accounts {
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

func exec(account wm_account.ShopifyAccount) error {
	// 1. 获取同步状态
	syncState, err := getState(account)
	if err != nil {
		log.Printf("[%s] 获取同步状态失败: %v", account.GetTraceId(), err)
		return err
	}

	if syncState.UpdatedAt.Add(time.Hour).After(time.Now().UTC()) &&
		syncState.Status == STATUS_SUCCESS {
		log.Printf("[%s] 同步时间小于1小时，跳过", account.GetTraceId())
		return nil
	}

	// 2. 同步客户数据
	err = syncCustomers(account, syncState)
	if err != nil {
		syncState.Status = STATUS_FAILED
		syncState.Message = err.Error()
		_ = updateSyncState(account, syncState)
	}
	return err
}
