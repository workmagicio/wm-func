package lock

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"wm-func/common/db/platform_db"

	"gorm.io/gorm"
)

/*
CREATE TABLE `distributed_locks` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	`lock_key` VARCHAR(255) NOT NULL COMMENT '锁的唯一键',
	`owner_id` VARCHAR(128) NOT NULL COMMENT '锁持有者的唯一标识',
	`expires_at` TIMESTAMP NOT NULL COMMENT '锁的过期时间',
	`created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	`updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_lock_key` (`lock_key`),
	KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';
*/

// DistributedLock 分布式锁模型
type DistributedLock struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	LockKey   string    `gorm:"uniqueIndex:uk_lock_key;size:255;not null" json:"lock_key"`
	OwnerID   string    `gorm:"size:128;not null" json:"owner_id"`
	ExpiresAt time.Time `gorm:"index:idx_expires_at;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (DistributedLock) TableName() string {
	return "platform_offline.distributed_locks"
}

// Locker 分布式锁接口
type Locker interface {
	// Lock 获取锁，如果获取成功返回 nil，否则返回错误
	Lock(key, ownerID string, duration time.Duration) error

	// TryLock 尝试获取锁，不阻塞，立即返回结果
	TryLock(key, ownerID string, duration time.Duration) error

	// Renew 续期锁，延长锁的过期时间
	Renew(key, ownerID string, duration time.Duration) error

	// Unlock 释放锁
	Unlock(key, ownerID string) error

	// IsLocked 检查锁是否存在且未过期
	IsLocked(key string) (bool, error)

	// GetOwner 获取锁的持有者
	GetOwner(key string) (string, error)

	// CleanExpiredLocks 清理过期的锁
	CleanExpiredLocks() (int64, error)
}

// MySQLLocker MySQL实现的分布式锁
type MySQLLocker struct {
	db *gorm.DB
}

// NewMySQLLocker 创建新的MySQL分布式锁实例
func NewMySQLLocker() Locker {
	db := platform_db.GetDB()
	// 自动迁移表结构
	//db.AutoMigrate(&DistributedLock{})

	return &MySQLLocker{
		db: db,
	}
}

var (
	ErrLockNotFound    = errors.New("lock not found")
	ErrLockExists      = errors.New("lock already exists")
	ErrLockExpired     = errors.New("lock has expired")
	ErrNotLockOwner    = errors.New("not the owner of this lock")
	ErrLockFailed      = errors.New("failed to acquire lock")
	ErrInvalidDuration = errors.New("invalid lock duration")
)

// Lock 获取锁（阻塞式，会重试）
func (m *MySQLLocker) Lock(key, ownerID string, duration time.Duration) error {
	if duration <= 0 {
		return ErrInvalidDuration
	}

	for {
		err := m.TryLock(key, ownerID, duration)
		if err == nil {
			return nil // 成功获取锁
		}
		if err != ErrLockExists {
			return err // 其他错误直接返回
		}
		// 锁被占用，等待重试
		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock 尝试获取锁（非阻塞）
func (m *MySQLLocker) TryLock(key, ownerID string, duration time.Duration) error {
	if duration <= 0 {
		return ErrInvalidDuration
	}

	now := time.Now()
	expiresAt := now.Add(duration)

	// 先尝试清理过期的锁
	m.db.Where("lock_key = ? AND expires_at < ?", key, now).Delete(&DistributedLock{})

	// 检查锁是否已存在
	var existingLock DistributedLock
	err := m.db.Where("lock_key = ?", key).First(&existingLock).Error
	if err == nil {
		// 锁已存在，检查是否是同一个owner
		if existingLock.OwnerID == ownerID && existingLock.ExpiresAt.After(now) {
			// 是同一个owner且锁未过期，则更新过期时间
			return m.Renew(key, ownerID, duration)
		}
		// 不是同一个owner或者锁已过期但还没清理
		return ErrLockExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 数据库错误
		return fmt.Errorf("database error: %w", err)
	}

	// 锁不存在，尝试创建新锁
	lock := &DistributedLock{
		LockKey:   key,
		OwnerID:   ownerID,
		ExpiresAt: expiresAt,
	}

	result := m.db.Create(lock)
	if result.Error != nil {
		// 检查是否是唯一键冲突（可能在并发情况下其他进程插入了同样的key）
		if isUniqueConstraintError(result.Error) {
			return ErrLockExists
		}
		return fmt.Errorf("database error: %w", result.Error)
	}

	return nil
}

// Renew 续期锁
func (m *MySQLLocker) Renew(key, ownerID string, duration time.Duration) error {
	if duration <= 0 {
		return ErrInvalidDuration
	}

	now := time.Now()
	newExpiresAt := now.Add(duration)

	result := m.db.
		Where("lock_key = ? AND owner_id = ? AND expires_at > ?", key, ownerID, now).
		Updates(&DistributedLock{
			ExpiresAt: newExpiresAt,
		})

	if result.Error != nil {
		return fmt.Errorf("database error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		// 检查锁是否存在
		var lock DistributedLock
		err := m.db.Where("lock_key = ?", key).First(&lock).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrLockNotFound
			}
			return err
		}

		// 锁存在但owner不匹配或已过期
		if lock.OwnerID != ownerID {
			return ErrNotLockOwner
		}
		if lock.ExpiresAt.Before(now) {
			return ErrLockExpired
		}

		return ErrLockFailed
	}

	return nil
}

// Unlock 释放锁
func (m *MySQLLocker) Unlock(key, ownerID string) error {
	result := m.db.
		Where("lock_key = ? AND owner_id = ?", key, ownerID).
		Delete(&DistributedLock{})

	if result.Error != nil {
		return fmt.Errorf("database error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrLockNotFound
	}

	return nil
}

// IsLocked 检查锁是否存在且未过期
func (m *MySQLLocker) IsLocked(key string) (bool, error) {
	var count int64
	err := m.db.
		Model(&DistributedLock{}).
		Where("lock_key = ? AND expires_at > ?", key, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetOwner 获取锁的持有者
func (m *MySQLLocker) GetOwner(key string) (string, error) {
	var lock DistributedLock
	err := m.db.
		Where("lock_key = ? AND expires_at > ?", key, time.Now()).
		First(&lock).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrLockNotFound
		}
		return "", err
	}

	return lock.OwnerID, nil
}

// CleanExpiredLocks 清理过期的锁
func (m *MySQLLocker) CleanExpiredLocks() (int64, error) {
	result := m.db.
		Where("expires_at < ?", time.Now()).
		Delete(&DistributedLock{})

	if result.Error != nil {
		return 0, fmt.Errorf("database error: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// isUniqueConstraintError 判断是否是唯一键约束错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// MySQL唯一键约束错误的常见关键词
	return strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "error 1062")
}
