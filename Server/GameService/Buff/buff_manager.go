package buff

import (
	"sync"
	"time"

	common "game-server/Common"
)

// ActiveBuff 激活的BUFF实例
type ActiveBuff struct {
	BuffID      uint32    // BUFF配置ID
	TargetID    uint64    // 目标角色ID
	TargetType  uint8     // 1=玩家, 2=怪物
	Stack       int       // 当前叠加层数
	ExpireAt    time.Time // 过期时间（duration=0表示永久，如打坐）
	LastTickAt  time.Time // 上次tick时间（用于持续伤害/恢复）
	SourceID    uint64    // 来源ID（施放者）
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
	HpChange     int
	MpChange     int
	AttackChange int
	DefChange    int
	SpeedChange  int
	HitChange    int
	DodgeChange  int
	CritChange   int
}

// Manager BUFF管理器（全局单例）
type Manager struct {
	mu    sync.RWMutex
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
			"buff_id":    b.BuffID,
			"name":       config.Name,
			"type":       config.Type,
			"stack":      b.Stack,
			"duration":   config.Duration,
			"can_cancel": config.CanCancel,
			"desc":       config.Description,
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
