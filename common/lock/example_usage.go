package lock

import (
	"fmt"
	"log"
	"time"
)

// 使用示例：如何使用分布式锁
func ExampleUsage() {
	// 2. 创建分布式锁实例
	locker := NewMySQLLocker()

	// 3. 基本使用示例
	basicExample(locker)

	// 4. 高级使用示例
	advancedExample(locker)

	// 5. 并发安全示例
	concurrencyExample(locker)
}

// basicExample 基本使用示例
func basicExample(locker Locker) {
	lockKey := "order:process:123"
	ownerID := "worker-001"
	lockDuration := 30 * time.Second

	fmt.Println("=== 基本使用示例 ===")

	// 获取锁
	err := locker.TryLock(lockKey, ownerID, lockDuration)
	if err != nil {
		if err == ErrLockExists {
			fmt.Println("锁已被其他进程持有")
			return
		}
		log.Printf("获取锁失败: %v", err)
		return
	}

	fmt.Println("✅ 成功获取锁，开始处理业务...")

	// 模拟业务处理
	time.Sleep(5 * time.Second)

	// 续期锁（如果需要更多时间）
	err = locker.Renew(lockKey, ownerID, 30*time.Second)
	if err != nil {
		log.Printf("续期失败: %v", err)
	} else {
		fmt.Println("🔄 锁续期成功")
	}

	// 继续处理业务
	time.Sleep(3 * time.Second)

	// 释放锁
	err = locker.Unlock(lockKey, ownerID)
	if err != nil {
		log.Printf("释放锁失败: %v", err)
	} else {
		fmt.Println("🔓 锁已释放")
	}
}

// advancedExample 高级使用示例
func advancedExample(locker Locker) {
	lockKey := "critical:section:456"
	ownerID := "worker-002"

	fmt.Println("\n=== 高级使用示例 ===")

	// 阻塞式获取锁
	err := locker.Lock(lockKey, ownerID, 5*time.Minute)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	fmt.Println("✅ 阻塞式获取锁成功")

	// 检查锁状态
	isLocked, err := locker.IsLocked(lockKey)
	if err != nil {
		log.Printf("检查锁状态失败: %v", err)
	} else {
		fmt.Printf("🔍 锁状态: %v\n", isLocked)
	}

	// 获取锁持有者
	owner, err := locker.GetOwner(lockKey)
	if err != nil {
		log.Printf("获取锁持有者失败: %v", err)
	} else {
		fmt.Printf("👤 锁持有者: %s\n", owner)
	}

	// 释放锁
	locker.Unlock(lockKey, ownerID)
	fmt.Println("🔓 锁已释放")
}

// concurrencyExample 并发安全示例
func concurrencyExample(locker Locker) {
	fmt.Println("\n=== 并发安全示例 ===")

	lockKey := "concurrent:test:789"

	// 启动3个goroutine同时竞争锁
	for i := 0; i < 3; i++ {
		go func(workerID int) {
			ownerID := fmt.Sprintf("worker-%d", workerID)

			err := locker.TryLock(lockKey, ownerID, 3*time.Second)
			if err != nil {
				if err == ErrLockExists {
					fmt.Printf("Worker %d: 锁被占用，等待...\n", workerID)

					// 等待并重试
					time.Sleep(100 * time.Millisecond)
					err = locker.Lock(lockKey, ownerID, 3*time.Second)
					if err != nil {
						fmt.Printf("Worker %d: 最终获取锁失败: %v\n", workerID, err)
						return
					}
				} else {
					fmt.Printf("Worker %d: 获取锁失败: %v\n", workerID, err)
					return
				}
			}

			fmt.Printf("✅ Worker %d: 获得锁，开始工作\n", workerID)

			// 模拟工作
			time.Sleep(1 * time.Second)

			// 释放锁
			err = locker.Unlock(lockKey, ownerID)
			if err != nil {
				fmt.Printf("Worker %d: 释放锁失败: %v\n", workerID, err)
			} else {
				fmt.Printf("🔓 Worker %d: 工作完成，锁已释放\n", workerID)
			}
		}(i + 1)
	}

	// 等待所有goroutine完成
	time.Sleep(8 * time.Second)

	// 清理过期锁
	count, err := locker.CleanExpiredLocks()
	if err != nil {
		log.Printf("清理过期锁失败: %v", err)
	} else {
		fmt.Printf("🧹 清理了 %d 个过期锁\n", count)
	}
}

// 自动续期示例：适用于长时间运行的任务
func autoRenewExample(locker Locker) {
	lockKey := "long:running:task"
	ownerID := "long-worker"
	initialDuration := 10 * time.Second

	// 获取初始锁
	err := locker.TryLock(lockKey, ownerID, initialDuration)
	if err != nil {
		log.Printf("获取锁失败: %v", err)
		return
	}

	fmt.Println("✅ 开始长时间任务...")

	// 启动自动续期goroutine
	renewTicker := time.NewTicker(5 * time.Second) // 每5秒续期一次
	defer renewTicker.Stop()

	done := make(chan bool)

	// 自动续期
	go func() {
		for {
			select {
			case <-renewTicker.C:
				err := locker.Renew(lockKey, ownerID, 10*time.Second)
				if err != nil {
					log.Printf("⚠️  续期失败: %v", err)
				} else {
					fmt.Println("🔄 锁自动续期成功")
				}
			case <-done:
				return
			}
		}
	}()

	// 模拟长时间任务（20秒）
	time.Sleep(20 * time.Second)

	// 停止自动续期
	done <- true

	// 释放锁
	err = locker.Unlock(lockKey, ownerID)
	if err != nil {
		log.Printf("释放锁失败: %v", err)
	} else {
		fmt.Println("🔓 长时间任务完成，锁已释放")
	}
}
