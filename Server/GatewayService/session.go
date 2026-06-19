package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisSessionClient *redis.Client
	redisSessionCtx    = context.Background()
	sessionTimeout     = 5 * time.Minute
)

// InitRedisSession 初始化 Redis Session 客户端（复用广播的 Redis 配置）
func InitRedisSession(addr, password string, db int, poolSize, minIdle int) {
	if poolSize <= 0 {
		poolSize = 50
	}
	if minIdle <= 0 {
		minIdle = 5
	}

	redisSessionClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
	})

	_, err := redisSessionClient.Ping(redisSessionCtx).Result()
	if err != nil {
		log.Fatalf("Redis Session 客户端连接失败: %v", err)
	}
	log.Printf("Redis Session 客户端连接成功")
}

// sessionKey 生成 session 的 Redis key
func sessionKey(token string) string {
	return fmt.Sprintf("gateway:session:%s", token)
}

// roleSessionKey 生成 role->session 映射的 Redis key
func roleSessionKey(roleID uint64) string {
	return fmt.Sprintf("gateway:role_session:%d", roleID)
}

// saveSession 保存会话到 Redis
func saveSession(token string, accountID, roleID uint64, name string, mapID uint32, x, y int) error {
	session := &SessionData{
		AccountID: accountID,
		RoleID:    roleID,
		Name:      name,
		MapID:     mapID,
		X:         x,
		Y:         y,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// 存储 session 数据，设置 TTL
	err = redisSessionClient.Set(redisSessionCtx, sessionKey(token), data, sessionTimeout).Err()
	if err != nil {
		return err
	}

	// 存储 role -> token 映射，设置相同 TTL
	err = redisSessionClient.Set(redisSessionCtx, roleSessionKey(roleID), token, sessionTimeout).Err()
	if err != nil {
		return err
	}

	return nil
}

// getSession 从 Redis 获取会话
func getSession(token string) (*SessionData, bool) {
	data, err := redisSessionClient.Get(redisSessionCtx, sessionKey(token)).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Redis getSession 错误: %v", err)
		}
		return nil, false
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		log.Printf("Redis session JSON 解析错误: %v", err)
		return nil, false
	}

	return &session, true
}

// removeSession 删除会话
func removeSession(token string) {
	// 先获取 roleID 用于删除 role_session 映射
	data, err := redisSessionClient.Get(redisSessionCtx, sessionKey(token)).Bytes()
	if err == nil {
		var session SessionData
		if err := json.Unmarshal(data, &session); err == nil {
			if err := redisSessionClient.Del(redisSessionCtx, roleSessionKey(session.RoleID)).Err(); err != nil {
				log.Printf("removeSession 删除 role_session key 失败: %v", err)
			}
		}
	}

	if err := redisSessionClient.Del(redisSessionCtx, sessionKey(token)).Err(); err != nil {
		log.Printf("removeSession 删除 session key 失败: %v", err)
	}
}

// getSessionByRoleID 通过角色ID获取会话
func getSessionByRoleID(roleID uint64) (*SessionData, bool) {
	// 先获取 token
	token, err := redisSessionClient.Get(redisSessionCtx, roleSessionKey(roleID)).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Redis getSessionByRoleID 错误: %v", err)
		}
		return nil, false
	}

	return getSession(token)
}

// refreshSession 刷新会话过期时间（Redis TTL 自动处理）
func refreshSession(token string) error {
	session, ok := getSession(token)
	if !ok {
		return fmt.Errorf("session 不存在")
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// 更新 session 数据（TTL 会自动刷新）
	err = redisSessionClient.Set(redisSessionCtx, sessionKey(token), data, sessionTimeout).Err()
	if err != nil {
		return err
	}

	// 刷新 role_session 的 TTL
	if err := redisSessionClient.Expire(redisSessionCtx, roleSessionKey(session.RoleID), sessionTimeout).Err(); err != nil {
		log.Printf("刷新 roleSession 过期时间失败: %v", err)
	}

	return nil
}

// getOnlineCount 获取在线人数（遍历方式，适合小规模）
func getOnlineCount() int64 {
	// 使用 SCAN 遍历所有 session key
	var count int64
	iter := redisSessionClient.Scan(redisSessionCtx, 0, "gateway:session:*", 100).Iterator()
	for iter.Next(redisSessionCtx) {
		count++
	}
	return count
}

// cleanupExpiredSessions Redis 自动处理过期，此函数仅做日志记录
func cleanupExpiredSessions() {
	// Redis TTL 自动清理，无需手动处理
	count := getOnlineCount()
	log.Printf("当前在线会话数: %d", count)
}
