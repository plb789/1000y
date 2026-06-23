package buff

import (
	"log"
	"sync"
	"time"

	common "game-server/Common"
)

// ActiveBuff 激活的BUFF实例
type ActiveBuff struct {
	BuffID     uint32    // BUFF配置ID
	TargetID   uint64    // 目标角色ID
	TargetType uint8     // 1=玩家, 2=怪物
	Stack      int       // 当前叠加层数
	ExpireAt   time.Time // 过期时间（duration=0表示永久，如打坐）
	LastTickAt time.Time // 上次tick时间（用于持续伤害/恢复）
	SourceID   uint64    // 来源ID（施放者）
}

// IsPermanent 是否永久BUFF（duration=0）
func (b *ActiveBuff) IsPermanent() bool {
	return b.ExpireAt.IsZero()
}

// IsExpired 是否已过期
func (b *ActiveBuff) IsExpired() bool {
	if b.IsPermanent() {
		return false
	}
	return time.Now().After(b.ExpireAt)
}

// BuffEffectSummary BUFF对属性的当前总加成（含叠加层数）
type BuffEffectSummary struct {
	HpChange           int
	MpChange           int
	AttackChange       int
	DefChange          int
	SpeedChange        int
	HitChange          int
	DodgeChange        int
	CritChange         int
	DamageReductionPct int // 总减伤百分比（可叠加，负值=易伤）
	ReflectPct         int // 总反弹百分比
	LifestealPct       int // 总吸血百分比
}

// Manager BUFF管理器（全局单例）
type Manager struct {
	mu sync.RWMutex
	// buffs[targetID] = 该目标身上所有激活的BUFF
	buffs map[uint64][]*ActiveBuff // key: targetID
}

var (
	globalManager *Manager
	once          sync.Once
)

// GetManager 获取全局BUFF管理器
func GetManager() *Manager {
	once.Do(func() {
		globalManager = &Manager{
			buffs: make(map[uint64][]*ActiveBuff),
		}
	})
	return globalManager
}

// AddBuff 给目标添加BUFF
// 返回：实际添加的BUFF实例（nil表示添加失败，如不可叠加且已存在）
func (m *Manager) AddBuff(targetID uint64, targetType uint8, buffID uint32, sourceID uint64) *ActiveBuff {
	config := common.GetBuffConfig(buffID)
	if config == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.buffs[targetID]

	// 检查是否已存在同ID BUFF
	for _, b := range list {
		if b.BuffID == buffID {
			if config.StackMax <= 1 {
				// 不可叠加，刷新持续时间
				if config.Duration > 0 {
					b.ExpireAt = time.Now().Add(time.Duration(config.Duration) * time.Second)
				}
				return b
			}
			// 可叠加
			if b.Stack < config.StackMax {
				b.Stack++
				if config.Duration > 0 {
					b.ExpireAt = time.Now().Add(time.Duration(config.Duration) * time.Second)
				}
				return b
			}
			// 已达最大叠加，刷新时间
			if config.Duration > 0 {
				b.ExpireAt = time.Now().Add(time.Duration(config.Duration) * time.Second)
			}
			return b
		}
	}

	// 新增BUFF
	buff := &ActiveBuff{
		BuffID:     buffID,
		TargetID:   targetID,
		TargetType: targetType,
		Stack:      1,
		SourceID:   sourceID,
	}
	if config.Duration > 0 {
		buff.ExpireAt = time.Now().Add(time.Duration(config.Duration) * time.Second)
	}
	buff.LastTickAt = time.Now()

	m.buffs[targetID] = append(list, buff)
	return buff
}

// RemoveBuff 移除目标身上的指定BUFF
func (m *Manager) RemoveBuff(targetID uint64, buffID uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.buffs[targetID]
	for i, b := range list {
		if b.BuffID == buffID {
			m.buffs[targetID] = append(list[:i], list[i+1:]...)
			if len(m.buffs[targetID]) == 0 {
				delete(m.buffs, targetID)
			}
			return true
		}
	}
	return false
}

// ClearAllBuffs 清除目标所有BUFF（死亡时调用）
func (m *Manager) ClearAllBuffs(targetID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.buffs, targetID)
}

// GetBuffs 获取目标身上所有激活的BUFF（已过滤过期）
func (m *Manager) GetBuffs(targetID uint64) []*ActiveBuff {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.buffs[targetID]
	var result []*ActiveBuff
	for _, b := range list {
		if !b.IsExpired() {
			result = append(result, b)
		}
	}
	return result
}

// CalculateEffect 计算目标身上所有BUFF的属性总加成
func (m *Manager) CalculateEffect(targetID uint64) BuffEffectSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := BuffEffectSummary{}
	list := m.buffs[targetID]
	for _, b := range list {
		if b.IsExpired() {
			continue
		}
		config := common.GetBuffConfig(b.BuffID)
		if config == nil {
			continue
		}
		// 叠加层数影响数值
		stack := b.Stack
		summary.HpChange += config.HpChange * stack
		summary.MpChange += config.MpChange * stack
		summary.AttackChange += config.AttackChange * stack
		summary.DefChange += config.DefChange * stack
		summary.SpeedChange += config.SpeedChange * stack
		summary.HitChange += config.HitChange * stack
		summary.DodgeChange += config.DodgeChange * stack
		summary.CritChange += config.CritChange * stack
		summary.DamageReductionPct += config.DamageReductionPct * stack
		summary.ReflectPct += config.ReflectPct * stack
		summary.LifestealPct += config.LifestealPct * stack
	}
	return summary
}

// HasBuff 是否有指定BUFF（含控制类BUFF检查）
func (m *Manager) HasBuff(targetID uint64, buffID uint32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.buffs[targetID]
	for _, b := range list {
		if b.BuffID == buffID && !b.IsExpired() {
			return true
		}
	}
	return false
}

// HasControlBuff 是否处于控制状态（眩晕/冰冻/石化/混乱，type=3）
func (m *Manager) HasControlBuff(targetID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.buffs[targetID]
	for _, b := range list {
		if b.IsExpired() {
			continue
		}
		config := common.GetBuffConfig(b.BuffID)
		if config != nil && config.Type == 3 {
			return true
		}
	}
	return false
}

// IsSilenced 是否沉默（无法使用技能）
func (m *Manager) IsSilenced(targetID uint64) bool {
	return m.HasBuff(targetID, 8) // buff_id=8 沉默
}

// IsStunned 是否眩晕（无法移动和攻击）
func (m *Manager) IsStunned(targetID uint64) bool {
	return m.HasBuff(targetID, 7) // buff_id=7 眩晕
}

// IsInvincible 是否无敌
func (m *Manager) IsInvincible(targetID uint64) bool {
	return m.HasBuff(targetID, 10) // buff_id=10 无敌
}

// Tick 处理BUFF的持续效果（持续伤害/恢复）
// 返回：本次tick造成的HP变化（正=恢复，负=伤害）、MP变化
func (m *Manager) Tick(targetID uint64, now time.Time) (hpChange, mpChange int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.buffs[targetID]
	for _, b := range list {
		if b.IsExpired() {
			continue
		}
		config := common.GetBuffConfig(b.BuffID)
		if config == nil {
			continue
		}
		// 持续效果按秒tick
		elapsed := now.Sub(b.LastTickAt)
		if elapsed < time.Second {
			continue
		}
		seconds := int(elapsed.Seconds())
		if seconds < 1 {
			seconds = 1
		}
		stack := b.Stack
		hpChange += config.HpChange * stack * seconds
		mpChange += config.MpChange * stack * seconds
		b.LastTickAt = now
	}
	return
}

// CleanupExpired 清理所有目标的过期BUFF（定时调用）
func (m *Manager) CleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for targetID, list := range m.buffs {
		var active []*ActiveBuff
		for _, b := range list {
			if !b.IsExpired() {
				active = append(active, b)
			}
		}
		if len(active) == 0 {
			delete(m.buffs, targetID)
		} else if len(active) != len(list) {
			m.buffs[targetID] = active
		}
	}
}

// GetBuffListInfo 获取目标BUFF列表信息（用于推送给客户端显示）
func (m *Manager) GetBuffListInfo(targetID uint64) []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := m.buffs[targetID]
	var result []map[string]interface{}
	for _, b := range list {
		if b.IsExpired() {
			continue
		}
		config := common.GetBuffConfig(b.BuffID)
		if config == nil {
			continue
		}
		info := map[string]interface{}{
			"buff_id":              b.BuffID,
			"name":                 config.Name,
			"type":                 config.Type,
			"stack":                b.Stack,
			"duration":             config.Duration,
			"can_cancel":           config.CanCancel,
			"desc":                 config.Description,
			"damage_reduction_pct": config.DamageReductionPct,
			"reflect_pct":          config.ReflectPct,
			"lifesteal_pct":        config.LifestealPct,
		}
		if !b.IsPermanent() {
			remaining := time.Until(b.ExpireAt).Seconds()
			if remaining < 0 {
				remaining = 0
			}
			info["remaining"] = int(remaining)
		} else {
			info["remaining"] = 0
		}
		result = append(result, info)
	}
	return result
}

// ========== 战斗效果快捷方法（供战斗系统调用）==========

// GetTotalDamageReduction 获取目标总减伤百分比（含叠加）
// 返回值：0=无减伤, 100=完全免疫(无敌), 负数=易伤(如破甲-20%)
func (m *Manager) GetTotalDamageReduction(targetID uint64) int {
	effect := m.CalculateEffect(targetID)
	return effect.DamageReductionPct
}

// GetTotalReflectPct 获取目标总反弹百分比
func (m *Manager) GetTotalReflectPct(targetID uint64) int {
	effect := m.CalculateEffect(targetID)
	return effect.ReflectPct
}

// GetTotalLifestealPct 获取攻击者总吸血百分比
func (m *Manager) GetTotalLifestealPct(attackerID uint64) int {
	effect := m.CalculateEffect(attackerID)
	return effect.LifestealPct
}

// IsFullyInvincible 是否完全免疫伤害（减伤>=100%）
func (m *Manager) IsFullyInvincible(targetID uint64) bool {
	return m.GetTotalDamageReduction(targetID) >= 100
}

// ========== BUFF定时Tick系统 ==========

// BuffTickApplier BUFF效果应用回调（由外部注入，解耦buff_manager对其他服务的依赖）
type BuffTickApplier interface {
	// ApplyPlayerHPChange 修改玩家HP（正=恢复, 负=伤害）
	ApplyPlayerHPChange(playerID uint64, hpChange int)
	// ApplyPlayerMPChange 修改玩家MP（正=恢复, 负=消耗）
	ApplyPlayerMPChange(playerID uint64, mpChange int)
	// ApplyMonsterHPChange 修改怪物HP（正=恢复, 负=伤害）, 返回(当前HP, 是否死亡)
	ApplyMonsterHPChange(monsterID uint64, hpChange int) (currentHP int, isDead bool)
	// BroadcastBuffTick 推送单个BUFF Tick结果到客户端
	BroadcastBuffTick(targetID uint64, targetType uint8, hpChange int, mpChange int)
	// BroadcastBuffTickBatch 批量推送BUFF Tick结果（性能优化：减少网络请求）
	BroadcastBuffTickBatch(results []BuffTickResult)
}

var (
	globalTicker     *time.Ticker
	globalTickerStop chan struct{}
)

// StartTicker 启动BUFF定时Tick器（每秒执行一次，在main.go中调用）
func (m *Manager) StartTicker(applier BuffTickApplier) {
	if globalTicker != nil {
		log.Println("[BUFF] 定时器已在运行，跳过重复启动")
		return
	}

	globalTicker = time.NewTicker(1 * time.Second)
	globalTickerStop = make(chan struct{})

	go func() {
		log.Println("[BUFF] 定时Tick器已启动 (间隔:1s)")
		for {
			select {
			case <-globalTicker.C:
				m.tickAllTargets(applier)
			case <-globalTickerStop:
				globalTicker.Stop()
				globalTicker = nil
				log.Println("[BUFF] 定时Tick器已停止")
				return
			}
		}
	}()
}

// StopTicker 停止BUFF定时Tick器
func (m *Manager) StopTicker() {
	if globalTickerStop != nil {
		close(globalTickerStop)
		globalTickerStop = nil
	}
}

// BuffTickResult 单次BUFF tick结果
type BuffTickResult struct {
	TargetID   uint64
	TargetType uint8
	HpChange   int
	MpChange   int
}

// tickAllTargets 遍历所有有BUFF的目标，执行Tick并应用效果（优化版：批量处理）
func (m *Manager) tickAllTargets(applier BuffTickApplier) {
	now := time.Now()

	// 快照所有有BUFF的targetID（避免锁竞争）
	m.mu.RLock()
	targetIDs := make([]uint64, 0, len(m.buffs))
	for id := range m.buffs {
		targetIDs = append(targetIDs, id)
	}
	m.mu.RUnlock()

	// 批量收集所有tick结果
	var results []BuffTickResult

	for _, targetID := range targetIDs {
		hpChange, mpChange := m.Tick(targetID, now)

		if hpChange == 0 && mpChange == 0 {
			continue // 无变化则跳过
		}

		// 判断目标类型：玩家ID通常较小且为角色ID格式，怪物ID为实例ID
		// 通过尝试获取怪物信息来区分；更可靠的方式是记录AddBuff时的targetType
		targetType := m.getTargetType(targetID)

		switch targetType {
		case 1: // 玩家
			if hpChange != 0 {
				applier.ApplyPlayerHPChange(targetID, hpChange)
			}
			if mpChange != 0 {
				applier.ApplyPlayerMPChange(targetID, mpChange)
			}
		case 2: // 怪物
			if hpChange != 0 {
				currentHP, isDead := applier.ApplyMonsterHPChange(targetID, hpChange)
				if isDead {
					// 怪物被持续BUFF杀死（如中毒致死）
					log.Printf("[BUFF] 怪物[%d]被BUFF持续效果击杀", targetID)
					m.ClearAllBuffs(targetID)
				}
				_ = currentHP
			}
		}

		// 收集结果用于批量广播
		results = append(results, BuffTickResult{
			TargetID:   targetID,
			TargetType: targetType,
			HpChange:   hpChange,
			MpChange:   mpChange,
		})
	}

	// 批量广播所有tick结果（减少网络请求次数）
	if len(results) > 0 {
		applier.BroadcastBuffTickBatch(results)
	}

	// 每分钟清理一次过期BUFF（轻量操作）
	if now.Unix()%60 == 0 {
		m.CleanupExpired()
	}
}

// getTargetType 根据targetID推断目标类型（简化判断）
func (m *Manager) getTargetType(targetID uint64) uint8 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, ok := m.buffs[targetID]
	if !ok {
		return 0
	}
	// 取第一个BUFF的TargetType（AddBuff时记录的）
	if len(list) > 0 {
		return list[0].TargetType
	}
	return 0
}
