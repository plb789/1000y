package main

import (
	"context"
	"encoding/json"
	"fmt"
	"game-server/DBService/Redis"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

// GameServiceInstance 游戏服务实例信息
type GameServiceInstance struct {
	InstanceID  uint32   `json:"instance_id"`
	URL         string   `json:"url"`
	HandledMaps []uint32 `json:"handled_maps"`
	StartTime   int64    `json:"start_time"`
	Heartbeat   int64    `json:"heartbeat"`
}

// RegistryStore 注册中心存储
type RegistryStore struct {
	mu        sync.RWMutex
	instances map[uint32]*GameServiceInstance
	mapToInst map[uint32]uint32
}

var registryStore = &RegistryStore{
	instances: make(map[uint32]*GameServiceInstance),
	mapToInst: make(map[uint32]uint32),
}

// Redis键
const (
	RegistryKeyPrefix = "registry:gameservice:"
	RegistryListKey   = "registry:gameservice:list"
	RegistryMapPrefix = "registry:map:"
)

// initRegistry 初始化注册中心
func initRegistry() {
	// 从Redis加载已有实例
	loadRegistryFromRedis()

	// 启动过期检测
	go registryCleanExpired()

	// 启动状态监控（每分钟打印一次当前实例状态）
	go registryStatusMonitor()
}

// loadRegistryFromRedis 从Redis加载注册信息
func loadRegistryFromRedis() {
	if Redis.RDB == nil {
		log.Println("Redis未连接，注册中心使用内存存储")
		return
	}

	// 获取所有实例ID
	ids, err := Redis.RDB.SMembers(ctx, RegistryListKey).Result()
	if err != nil {
		log.Printf("从Redis加载注册信息失败: %v", err)
		return
	}

	registryStore.mu.Lock()
	defer registryStore.mu.Unlock()

	for _, idStr := range ids {
		var id uint32
		fmt.Sscanf(idStr, "%d", &id)

		key := fmt.Sprintf("%s%d", RegistryKeyPrefix, id)
		data, err := Redis.RDB.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var inst GameServiceInstance
		if err := json.Unmarshal([]byte(data), &inst); err != nil {
			continue
		}

		// 检查心跳是否过期
		if time.Now().Unix()-inst.Heartbeat > 120 {
			// 过期，删除
			Redis.RDB.Del(ctx, key)
			Redis.RDB.SRem(ctx, RegistryListKey, idStr)
			continue
		}

		registryStore.instances[inst.InstanceID] = &inst
		for _, mapID := range inst.HandledMaps {
			registryStore.mapToInst[mapID] = inst.InstanceID
		}
	}

	log.Printf("从Redis加载了 %d 个游戏服务实例", len(registryStore.instances))
}

// registryCleanExpired 清理过期的实例
func registryCleanExpired() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if Redis.RDB == nil {
			continue
		}

		registryStore.mu.Lock()
		now := time.Now().Unix()
		for id, inst := range registryStore.instances {
			if now-inst.Heartbeat > 120 {
				// 删除过期实例
				key := fmt.Sprintf("%s%d", RegistryKeyPrefix, id)
				Redis.RDB.Del(ctx, key)
				Redis.RDB.SRem(ctx, RegistryListKey, fmt.Sprintf("%d", id))
				for _, mapID := range inst.HandledMaps {
					delete(registryStore.mapToInst, mapID)
				}
				delete(registryStore.instances, id)
				log.Printf("删除过期游戏服务实例: %d", id)
			}
		}
		registryStore.mu.Unlock()
	}
}

// handleRegistryRegister 注册游戏服务实例
func handleRegistryRegister(c *gin.Context) {
	var req struct {
		InstanceID  uint32   `json:"instance_id"`
		URL         string   `json:"url"`
		HandledMaps []uint32 `json:"handled_maps"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	inst := &GameServiceInstance{
		InstanceID:  req.InstanceID,
		URL:         req.URL,
		HandledMaps: req.HandledMaps,
		StartTime:   time.Now().Unix(),
		Heartbeat:   time.Now().Unix(),
	}

	// 保存到内存
	registryStore.mu.Lock()
	registryStore.instances[inst.InstanceID] = inst
	for _, mapID := range inst.HandledMaps {
		registryStore.mapToInst[mapID] = inst.InstanceID
	}
	registryStore.mu.Unlock()

	// 同步到Redis
	if Redis.RDB != nil {
		key := fmt.Sprintf("%s%d", RegistryKeyPrefix, inst.InstanceID)
		data, _ := json.Marshal(inst)
		Redis.RDB.Set(ctx, key, string(data), 0)
		Redis.RDB.SAdd(ctx, RegistryListKey, fmt.Sprintf("%d", inst.InstanceID))

		// 地图映射
		for _, mapID := range inst.HandledMaps {
			mapKey := fmt.Sprintf("%s%d", RegistryMapPrefix, mapID)
			Redis.RDB.Set(ctx, mapKey, inst.InstanceID, 0)
		}
	}

	log.Printf("注册游戏服务实例: instanceID=%d, url=%s, maps=%v", inst.InstanceID, inst.URL, inst.HandledMaps)
	c.JSON(200, gin.H{"code": 0, "msg": "注册成功"})
}

// handleRegistryUnregister 注销游戏服务实例
func handleRegistryUnregister(c *gin.Context) {
	var req struct {
		InstanceID uint32 `json:"instance_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	registryStore.mu.Lock()
	if inst, ok := registryStore.instances[req.InstanceID]; ok {
		for _, mapID := range inst.HandledMaps {
			delete(registryStore.mapToInst, mapID)
			if Redis.RDB != nil {
				mapKey := fmt.Sprintf("%s%d", RegistryMapPrefix, mapID)
				Redis.RDB.Del(ctx, mapKey)
			}
		}
		delete(registryStore.instances, req.InstanceID)
	}
	registryStore.mu.Unlock()

	// 从Redis删除
	if Redis.RDB != nil {
		key := fmt.Sprintf("%s%d", RegistryKeyPrefix, req.InstanceID)
		Redis.RDB.Del(ctx, key)
		Redis.RDB.SRem(ctx, RegistryListKey, fmt.Sprintf("%d", req.InstanceID))
	}

	c.JSON(200, gin.H{"code": 0, "msg": "注销成功"})
}

// handleRegistryHeartbeat 更新心跳
func handleRegistryHeartbeat(c *gin.Context) {
	var req struct {
		InstanceID uint32 `json:"instance_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	registryStore.mu.Lock()
	if inst, ok := registryStore.instances[req.InstanceID]; ok {
		inst.Heartbeat = time.Now().Unix()
	}
	registryStore.mu.Unlock()

	// 更新Redis
	if Redis.RDB != nil {
		key := fmt.Sprintf("%s%d", RegistryKeyPrefix, req.InstanceID)
		if data, err := json.Marshal(registryStore.instances[req.InstanceID]); err == nil {
			Redis.RDB.Set(ctx, key, string(data), 0)
		}
	}

	c.JSON(200, gin.H{"code": 0, "msg": "心跳更新成功"})
}

// handleRegistryList 获取所有实例
func handleRegistryList(c *gin.Context) {
	registryStore.mu.RLock()
	instances := make([]*GameServiceInstance, 0, len(registryStore.instances))
	for _, inst := range registryStore.instances {
		instances = append(instances, inst)
	}
	registryStore.mu.RUnlock()

	c.JSON(200, gin.H{"code": 0, "data": instances})
}

// handleRegistryGetByMap 根据地图ID获取实例
func handleRegistryGetByMap(c *gin.Context) {
	var req struct {
		MapID uint32 `json:"map_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	registryStore.mu.RLock()
	instanceID, ok := registryStore.mapToInst[req.MapID]
	inst := registryStore.instances[instanceID]
	registryStore.mu.RUnlock()

	if !ok || inst == nil {
		c.JSON(200, gin.H{"code": -1, "msg": "地图未被任何实例处理"})
		return
	}

	c.JSON(200, gin.H{"code": 0, "data": inst})
}

// registryStatusMonitor 状态监控 - 定期打印当前注册中心状态
func registryStatusMonitor() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	lastCount := 0

	for range ticker.C {
		registryStore.mu.RLock()
		currentCount := len(registryStore.instances)
		registryStore.mu.RUnlock()

		if currentCount != lastCount {
			// 实例数量发生变化，打印详细信息
			registryStore.mu.RLock()
			log.Printf("=================== 游戏服务实例状态变化 ===================")
			log.Printf("当前在线实例数: %d (变化: %+d)", currentCount, currentCount-lastCount)

			if currentCount > 0 {
				log.Printf("------------------- 实例详情 -------------------")
				for id, inst := range registryStore.instances {
					uptime := time.Since(time.Unix(inst.StartTime, 0)).Round(time.Second)
					lastHeartbeat := time.Since(time.Unix(inst.Heartbeat, 0)).Round(time.Second)
					log.Printf("实例[%d]: URL=%s, 处理地图=%v, 运行时间=%v, 最后心跳=%v前",
						id, inst.URL, inst.HandledMaps, uptime, lastHeartbeat)
				}
				log.Printf("-------------------------------------------------")
			}
			log.Printf("=================================================\n")
			registryStore.mu.RUnlock()
		}

		lastCount = currentCount
	}
}
