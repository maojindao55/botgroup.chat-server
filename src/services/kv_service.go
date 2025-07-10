package services

import (
	"context"
	"fmt"
	"project/src/config"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// KVItem KV存储项（内存版本使用）
type KVItem struct {
	Value     string
	ExpiresAt time.Time
}

// KVService KV存储服务接口
type KVService interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
	Close() error
}

// redisKVService Redis版本的KV存储服务实现
type redisKVService struct {
	client *redis.Client
	ctx    context.Context
}

// memoryKVService 内存版本的KV存储服务实现（作为fallback）
type memoryKVService struct {
	storage map[string]*KVItem
	mutex   sync.RWMutex
}

// NewKVService 创建KV存储服务实例
func NewKVService(redisConfig config.RedisConfig) KVService {
	// 尝试连接Redis
	if redisService, err := newRedisKVService(redisConfig); err == nil {
		return redisService
	}

	// Redis连接失败，使用内存版本
	fmt.Println("Redis连接失败，使用内存KV存储")
	return newMemoryKVService()
}

// newRedisKVService 创建Redis KV服务
func newRedisKVService(redisConfig config.RedisConfig) (KVService, error) {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %v", err)
	}

	fmt.Println("Redis连接成功，使用Redis KV存储")
	return &redisKVService{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// Set Redis版本设置键值对
func (r *redisKVService) Set(key, value string, ttl time.Duration) error {
	return r.client.Set(r.ctx, key, value, ttl).Err()
}

// Get Redis版本获取键值
func (r *redisKVService) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 键不存在返回空字符串
	}
	return val, err
}

// Delete Redis版本删除键
func (r *redisKVService) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Close Redis版本关闭连接
func (r *redisKVService) Close() error {
	return r.client.Close()
}

// newMemoryKVService 创建内存KV服务
func newMemoryKVService() KVService {
	kv := &memoryKVService{
		storage: make(map[string]*KVItem),
	}

	// 启动清理过期键的协程
	go kv.cleanup()

	return kv
}

// Set 内存版本设置键值对
func (kv *memoryKVService) Set(key, value string, ttl time.Duration) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	expiresAt := time.Now().Add(ttl)
	kv.storage[key] = &KVItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

// Get 内存版本获取键值
func (kv *memoryKVService) Get(key string) (string, error) {
	kv.mutex.RLock()
	defer kv.mutex.RUnlock()

	item, exists := kv.storage[key]
	if !exists {
		return "", nil // 键不存在返回空字符串
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(kv.storage, key)
		return "", nil
	}

	return item.Value, nil
}

// Delete 内存版本删除键
func (kv *memoryKVService) Delete(key string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	delete(kv.storage, key)
	return nil
}

// Close 内存版本关闭（无操作）
func (kv *memoryKVService) Close() error {
	return nil
}

// cleanup 内存版本清理过期的键
func (kv *memoryKVService) cleanup() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kv.cleanupExpired()
		}
	}
}

// cleanupExpired 内存版本清理过期的键
func (kv *memoryKVService) cleanupExpired() {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()

	now := time.Now()
	for key, item := range kv.storage {
		if now.After(item.ExpiresAt) {
			delete(kv.storage, key)
		}
	}
}
