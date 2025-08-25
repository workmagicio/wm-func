package lock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql" // 引入 MySQL 驱动
)

func setupTestLocker(t *testing.T) Locker {
	// 这是一个关键的辅助函数。为了让测试可靠，你需要确保：
	// 1. NewMySQLLocker() 能够成功创建一个实例。
	// 2. 在每次测试运行前，相关的数据库表（如 distributed_locks）是干净的。

	// 这里我假设你可以通过某种方式获取到底层的 *sql.DB 来执行清理操作。
	// 如果 NewMySQLLocker() 隐藏了数据库连接，你可能需要在你的 Locker 实现中
	// 添加一个仅用于测试的 `Reset()` 或 `ClearAll()` 方法。

	// dsn := os.Getenv("TEST_DB_DSN") // 建议从环境变量读取
	// db, _ := sql.Open("mysql", dsn)
	// _, err := db.Exec("DELETE FROM distributed_locks")
	// if err != nil {
	// 	t.Fatalf("清理锁表失败: %v", err)
	// }

	// 根据你的示例，我们直接调用 NewMySQLLocker()
	locker := NewMySQLLocker()
	return locker
}

// --- Test Cases ---

// TestLockerSuite 运行所有与 Locker 相关的测试
func TestLockerSuite(t *testing.T) {
	// 注意：下面的每个 t.Run 都是一个独立的子测试。
	// 理想情况下，每个子测试都应该在一个干净的环境下运行。
	// setupTestLocker 函数就是为了实现这一点。

	t.Run("TestTryLock_SuccessAndFail", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "trylock-key-1"
		owner1 := "owner-1"
		owner2 := "owner-2"
		duration := 10 * time.Second

		// 1. owner1 成功获取锁
		err := locker.TryLock(key, owner1, duration)
		if err != nil {
			t.Fatalf("owner1 应该能成功获取锁，但失败了: %v", err)
		}

		// 2. owner2 尝试获取同一个锁，应该立即失败
		err = locker.TryLock(key, owner2, duration)
		if err == nil {
			t.Fatal("owner2 不应该能获取已被 owner1 持有的锁")
		}

		// 3. owner1 释放锁
		err = locker.Unlock(key, owner1)
		if err != nil {
			t.Fatalf("owner1 释放锁失败: %v", err)
		}

		// 4. owner2 再次尝试，这次应该成功
		err = locker.TryLock(key, owner2, duration)
		if err != nil {
			t.Fatalf("owner2 在锁被释放后应该能成功获取，但失败了: %v", err)
		}
		locker.Unlock(key, owner2) // 清理
	})

	t.Run("TestUnlock_WrongOwner", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "unlock-key-2"
		owner1 := "owner-real"
		owner2 := "owner-fake"
		duration := 10 * time.Second

		if err := locker.TryLock(key, owner1, duration); err != nil {
			t.Fatalf("获取锁失败: %v", err)
		}

		// 错误的持有者尝试解锁
		err := locker.Unlock(key, owner2)
		if err == nil {
			t.Fatal("错误的持有者不应该能成功解锁")
		}

		// 确认锁依然被原始持有者持有
		owner, err := locker.GetOwner(key)
		if err != nil || owner != owner1 {
			t.Fatalf("锁应仍被 owner1 持有，但 GetOwner() 返回 '%s' (err: %v)", owner, err)
		}
		locker.Unlock(key, owner1) // 清理
	})

	t.Run("TestLock_Expiration", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "expire-key-1"
		owner := "owner-expire"
		duration := 1 * time.Second

		if err := locker.TryLock(key, owner, duration); err != nil {
			t.Fatalf("获取锁失败: %v", err)
		}

		// 等待时间超过锁的有效期
		time.Sleep(duration + 500*time.Millisecond)

		locked, err := locker.IsLocked(key)
		if err != nil {
			t.Fatalf("检查锁状态失败: %v", err)
		}
		if locked {
			t.Fatal("锁在过期后状态依然是 locked")
		}
	})

	t.Run("TestRenew_SuccessAndFail", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "renew-key"
		owner1 := "owner-renew"
		owner2 := "imposter"
		duration := 2 * time.Second

		if err := locker.TryLock(key, owner1, duration); err != nil {
			t.Fatalf("获取锁失败: %v", err)
		}

		time.Sleep(duration / 2) // 等待一半时间
		if err := locker.Renew(key, owner1, 3*time.Second); err != nil {
			t.Fatalf("持有者续期失败: %v", err)
		}

		// 在原过期时间之后检查锁是否还存在
		time.Sleep(duration) // 等待时间超过原始 duration
		if locked, _ := locker.IsLocked(key); !locked {
			t.Fatal("锁在续期后、但在新过期时间前就消失了")
		}

		// 非持有者续期失败
		if err := locker.Renew(key, owner2, duration); err == nil {
			t.Fatal("非持有者不应该能续期成功")
		}
	})

	t.Run("TestLock_Blocking", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "blocking-lock-key"
		owner1 := "owner-blocking-1"
		owner2 := "owner-blocking-2"
		duration := 2 * time.Second

		wg := sync.WaitGroup{}
		wg.Add(1)

		// Goroutine 1: 获取锁并持有
		go func() {
			locker.Lock(key, owner1, duration)
			time.Sleep(duration / 2)
			locker.Unlock(key, owner1)
			wg.Done()
		}()

		time.Sleep(100 * time.Millisecond) // 确保 Goroutine 1 先拿到锁

		// 主 Goroutine (作为 owner2) 尝试获取锁，应该被阻塞
		startTime := time.Now()
		locker.Lock(key, owner2, duration)
		elapsed := time.Since(startTime)

		// 检查阻塞时间是否符合预期
		// 应该大于 (duration / 2) - 100ms
		if elapsed < (duration/2 - 200*time.Millisecond) {
			t.Errorf("Lock() 方法似乎没有阻塞足够长的时间, 仅阻塞了 %v", elapsed)
		}

		locker.Unlock(key, owner2)
		wg.Wait()
	})

	t.Run("TestConcurrency_MutualExclusion", func(t *testing.T) {
		locker := setupTestLocker(t)
		key := "concurrency-key"
		numGoroutines := 20
		incrementsPerGoroutine := 50

		var counter int64 // 使用原子操作来避免测试代码自身的数据竞争
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				ownerID := fmt.Sprintf("worker-%d", id)
				for j := 0; j < incrementsPerGoroutine; j++ {
					// 尝试获取锁，这里使用阻塞的 Lock
					err := locker.Lock(key, ownerID, 5*time.Second)
					if err != nil {
						// 在并发测试中，我们不希望有错误发生
						t.Errorf("Worker %d 获取锁失败: %v", id, err)
						return
					}

					// --- 临界区 ---
					// 读取、增加、写回，这是一个典型的需要锁保护的操作
					atomic.AddInt64(&counter, 1)
					// --- 临界区结束 ---

					err = locker.Unlock(key, ownerID)
					if err != nil {
						t.Errorf("Worker %d 释放锁失败: %v", id, err)
					}
				}
			}(i)
		}

		wg.Wait()

		expected := int64(numGoroutines * incrementsPerGoroutine)
		if counter != expected {
			t.Errorf("并发测试失败：期望的计数值是 %d, 但实际得到 %d。这表明锁未能保证互斥性。", expected, counter)
		}
	})

	t.Run("TestCleanExpiredLocks", func(t *testing.T) {
		locker := setupTestLocker(t)

		// 1. 创建一个很快就会过期的锁
		err := locker.TryLock("expired-key", "owner-expired", 1*time.Second)
		if err != nil {
			t.Fatalf("创建 expired-key 失败: %v", err)
		}

		// 2. 创建一个不会过期的锁
		err = locker.TryLock("active-key", "owner-active", 1*time.Minute)
		if err != nil {
			t.Fatalf("创建 active-key 失败: %v", err)
		}

		// 3. 等待第一个锁过期
		time.Sleep(1500 * time.Millisecond)

		// 4. 执行清理
		cleanedCount, err := locker.CleanExpiredLocks()
		if err != nil {
			t.Fatalf("CleanExpiredLocks 执行失败: %v", err)
		}

		// 5. 验证结果
		if cleanedCount != 1 {
			t.Errorf("期望清理 1 个过期锁, 实际清理了 %d 个", cleanedCount)
		}

		// 确认过期的锁确实被删了
		if locked, _ := locker.IsLocked("expired-key"); locked {
			t.Error("CleanExpiredLocks 未能成功清理已过期的锁")
		}

		// 确认未过期的锁还在
		if locked, _ := locker.IsLocked("active-key"); !locked {
			t.Error("CleanExpiredLocks 错误地清理了未过期的锁")
		}
	})
}
