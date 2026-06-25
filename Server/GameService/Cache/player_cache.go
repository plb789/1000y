package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"game-server/GameService/Attribute"
	common "game-server/Common"

	"github.com/redis/go-redis/v9"
)

// PlayerState 玩家完整状态（缓存在GameService的Redis中）
// ★ 设计原则：这是GameService的本地缓存，不与DBService的Redis冲突
type PlayerState struct {
	RoleID           uint64                `json:"role_id"`            // 角色ID
	FinalAttributes  *attribute.Attribute   `json:"final_attributes"`   // 最终属性（计算结果）
	BonusDetail      *attribute.AttributeBonus `json:"bonus_detail"`   // 加成明细
	EquippedSkills   []map[string]interface{} `json:"equipped_skills"` // 已装备技能列表
	EquippedItems    []map[string]interface{} `json:"equipped_items"`  // 已装备物品列表
	LastUpdated      time.Time              `json:"last_updated"`      // 最后更新时间
	Dirty            bool                   `json:"dirty"`             // 是否有未保存的修改
}

// PlayerCache 玩家状态缓存管理器
// ★ 核心职责：
//   - 缓存在线玩家的完整状态（避免频繁查询DBService）
//   - 提供快速读取接口（<1ms响应）
//   - 管理缓存生命周期（TTL自动过期 + 手动失效）
type PlayerCache struct {
	rdb         *redis.Client // Redis客户端（使用配置文件中的Redis连接）
	defaultTTL  time.Duration // 默认缓存过期时间
	calcEngine  *attribute.CalcEngine // 属性计算引擎引用
}

// NewPlayerCache 创建玩家缓存实例
func NewPlayerCache(rdb *redis.Client, calcEngine *attribute.CalcEngine) *PlayerCache {
	return &PlayerCache{
		rdb:        rdb,
		defaultTTL: 30 * time.Minute, // 默认30分钟过期（平衡内存占用和数据新鲜度）
		calcEngine: calcEngine,
	}
}

// GetOrLoad 获取玩家状态（优先从Redis缓存读取，未命中则重新计算并缓存）
// ★ 这是核心方法：所有需要玩家属性的Handler都应该调用此方法
func (pc *PlayerCache) GetOrLoad(ctx context.Context, roleID uint64) (*PlayerState, error) {
	cacheKey := fmt.Sprintf("gs:player:%d:state", roleID)

	// 1. 先尝试从Redis获取
	cached, err := pc.getFromRedis(ctx, cacheKey)
	if err == nil && cached != nil {
		fmt.Printf("[PlayerCache] 缓存命中: roleID=%d\n", roleID)
		return cached, nil
	}

	// 2. 缓存未命中，调用CalcEngine重新计算
	fmt.Printf("[PlayerCache] 缓存未命中，重新计算: roleID=%d\n", roleID)
	state, err := pc.calculateAndCache(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("计算并缓存玩家状态失败: %v", err)
	}

	return state, nil
}

// Invalidate 使指定玩家的缓存失效
// ★ 在以下场景调用：装备/卸载技能、装备/卸载物品、使用道具、BUFF变化等
func (pc *PlayerCache) Invalidate(ctx context.Context, roleID uint64) error {
	cacheKey := fmt.Sprintf("gs:player:%d:state", roleID)

	err := pc.rdb.Del(ctx, cacheKey).Err()
	if err != nil {
		return fmt.Errorf("使缓存失败失败: %v", err)
	}

	fmt.Printf("[PlayerCache] 缓存已失效: roleID=%d\n", roleID)
	return nil
}

// InvalidateBatch 批量使缓存失效（用于系统维护）
func (pc *PlayerCache) InvalidateBatch(ctx context.Context, roleIDs []uint64) error {
	if len(roleIDs) == 0 {
		return nil
	}

	pipe := pc.rdb.Pipeline()
	for _, roleID := range roleIDs {
		cacheKey := fmt.Sprintf("gs:player:%d:state", roleID)
		pipe.Del(ctx, cacheKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量使缓存失效失败: %v", err)
	}

	fmt.Printf("[PlayerCache] 批量缓存已失效: count=%d\n", len(roleIDs))
	return nil
}

// GetFinalAttributes 快捷方法：直接获取最终属性（供战斗系统等高频场景使用）
func (pc *PlayerCache) GetFinalAttributes(ctx context.Context, roleID uint64) (*attribute.Attribute, error) {
	state, err := pc.GetOrLoad(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return state.FinalAttributes, nil
}

// getFromRedis 从Redis读取缓存数据
func (pc *PlayerCache) getFromRedis(ctx context.Context, cacheKey string) (*PlayerState, error) {
	data, err := pc.rdb.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, err
	}

	var state PlayerState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

// calculateAndCache 调用CalcEngine计算属性并写入Redis缓存
func (pc *PlayerCache) calculateAndCache(ctx context.Context, roleID uint64) (*PlayerState, error) {
	// 1. 调用CalcEngine计算最终属性
	finalAttrs, bonusDetail, err := pc.calcEngine.CalculateFinalAttributes(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 2. 获取已装备技能列表（供前端显示）
	equippedSkills, _ := common.DBSkillGetEquipped(roleID)

	// 3. 获取已装备物品列表（供前端显示）
	equippedItems := []map[string]interface{}{}
	// TODO: 这里可以调用Item服务获取，暂时为空

	// 4. 组装完整状态
	state := &PlayerState{
		RoleID:          roleID,
		FinalAttributes: finalAttrs,
		BonusDetail:     bonusDetail,
		EquippedSkills:  equippedSkills,
		EquippedItems:   equippedItems,
		LastUpdated:     time.Now(),
		Dirty:           false,
	}

	// 5. 序列化并写入Redis
	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("序列化玩家状态失败: %v", err)
	}

	cacheKey := fmt.Sprintf("gs:player:%d:state", roleID)
	err = pc.rdb.Set(ctx, cacheKey, data, pc.defaultTTL).Err()
	if err != nil {
		// 写入缓存失败不影响主流程，仅记录日志
		fmt.Printf("⚠️ [PlayerCache] 写入Redis缓存失败: roleID=%d, err=%v\n", roleID, err)
		// 仍然返回state，只是下次需要重新计算
	}

	fmt.Printf("[PlayerCache] 状态已缓存: roleID=%d, TTL=%v\n", roleID, pc.defaultTTL)
	return state, nil
}

// CleanupExpired 清理过期的缓存条目（可选，通常由Redis自动处理）
func (pc *PlayerCache) CleanupExpired(ctx context.Context) error {
	pattern := "gs:player:*:state"
	iter := pc.rdb.Scan(ctx, 0, pattern, 100).Iterator()

	count := 0
	for iter.Next(ctx) {
		key := iter.Val()
		ttl := pc.rdb.TTL(ctx, key).Val()

		// 如果TTL < 0说明key不存在或无过期时间
		if ttl < 0 {
			continue
		}

		// 统计即将过期的key数量（< 1分钟）
		if ttl < time.Minute {
			count++
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if count > 0 {
		fmt.Printf("[PlayerCache] 即将过期的缓存条目: %d个\n", count)
	}

	return nil
}

// GetStats 获取缓存统计信息（监控用）
func (pc *PlayerCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	pattern := "gs:player:*:state"
	iter := pc.rdb.Scan(ctx, 0, pattern, 10000).Iterator()

	totalKeys := 0
	for iter.Next(ctx) {
		totalKeys++
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_cached_players": totalKeys,
		"default_ttl_minutes":  pc.defaultTTL.Minutes(),
		"timestamp":            time.Now().Unix(),
	}

	return stats, nil
}
