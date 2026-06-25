package common

import (
	"log"
	"sync"
	"time"
)

// MonsterSyncManager 怪物同步管理器
type MonsterSyncManager struct {
	monsterStates   map[uint32]*MonsterState    // monsterID -> state
	gatewayCallback func(msg *BroadcastMessage) // 回调函数：发送到网关
	batchBuffer     []*MonsterUpdate            // 批量更新缓冲
	batchMutex      sync.Mutex
	batchTimer      *time.Timer
	batchInterval   time.Duration // 批量发送间隔
	maxBatchSize    int           // 最大批量大小
	syncStats       *MonsterSyncMetrics
	mutex           sync.RWMutex
}

// MonsterState 怪物状态
type MonsterState struct {
	MonsterID  uint32 `json:"monster_id"`
	MapID      uint32 `json:"map_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"max_hp"`
	Status     uint8  `json:"status"`      // 0=空闲, 1=移动, 2=战斗, 3=死亡, 4=复活
	TargetID   uint64 `json:"target_id"`   // 当前目标ID
	LastUpdate int64  `json:"last_update"` // 最后更新时间戳
}

// MonsterUpdate 怪物更新信息
type MonsterUpdate struct {
	MonsterID  uint32                 `json:"monster_id"`
	MapID      uint32                 `json:"map_id"`
	UpdateType string                 `json:"update_type"` // "position", "hp", "status", "spawn", "death"
	Data       map[string]interface{} `json:"data"`
	Timestamp  int64                  `json:"timestamp"`
}

// MonsterSyncMetrics 怪物同步指标
type MonsterSyncMetrics struct {
	TotalUpdates    int64
	BatchedUpdates  int64
	PositionUpdates int64
	HPUpdates       int64
	StatusUpdates   int64
	SpawnEvents     int64
	DeathEvents     int64
	AvgBatchSize    float64
	SyncLatency     time.Duration
	mutex           sync.RWMutex
}

var monsterSyncMgr *MonsterSyncManager

// InitMonsterSyncManager 初始化怪物同步管理器
func InitMonsterSyncManager(gatewayCallback func(msg *BroadcastMessage), batchInterval time.Duration, maxBatchSize int) {
	mgr := &MonsterSyncManager{
		monsterStates:   make(map[uint32]*MonsterState),
		gatewayCallback: gatewayCallback,
		batchBuffer:     make([]*MonsterUpdate, 0, maxBatchSize),
		batchInterval:   batchInterval,
		maxBatchSize:    maxBatchSize,
		syncStats:       &MonsterSyncMetrics{},
	}

	// 启动批量发送定时器
	mgr.batchTimer = time.AfterFunc(batchInterval, mgr.flushBatch)

	monsterSyncMgr = mgr

	log.Printf(`✅ 怪物同步管理器初始化完成: batchInterval=%v, maxBatchSize=%d`,
		batchInterval, maxBatchSize)
}

// UpdateMonsterPosition 更新怪物位置
func UpdateMonsterPosition(monsterID, mapID uint32, x, y int) {
	if monsterSyncMgr == nil {
		return
	}

	update := &MonsterUpdate{
		MonsterID:  monsterID,
		MapID:      mapID,
		UpdateType: "position",
		Data: map[string]interface{}{
			"x": x,
			"y": y,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	monsterSyncMgr.addToBatch(update)
	monsterSyncMgr.updateLocalState(monsterID, mapID, x, y)

	monsterSyncMgr.syncStats.mutex.Lock()
	monsterSyncMgr.syncStats.PositionUpdates++
	monsterSyncMgr.syncStats.TotalUpdates++
	monsterSyncMgr.syncStats.mutex.Unlock()
}

// UpdateMonsterHP 更新怪物血量
func UpdateMonsterHP(monsterID uint32, hp, maxHP int) {
	if monsterSyncMgr == nil {
		return
	}

	update := &MonsterUpdate{
		MonsterID:  monsterID,
		UpdateType: "hp",
		Data: map[string]interface{}{
			"hp":     hp,
			"max_hp": maxHP,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	monsterSyncMgr.addToBatch(update)

	// 更新本地状态
	monsterSyncMgr.mutex.Lock()
	if state, ok := monsterSyncMgr.monsterStates[monsterID]; ok {
		state.HP = hp
		state.MaxHP = maxHP
	}
	monsterSyncMgr.mutex.Unlock()

	monsterSyncMgr.syncStats.mutex.Lock()
	monsterSyncMgr.syncStats.HPUpdates++
	monsterSyncMgr.syncStats.TotalUpdates++
	monsterSyncMgr.syncStats.mutex.Unlock()
}

// UpdateMonsterStatus 更新怪物状态
func UpdateMonsterStatus(monsterID uint32, status uint8, targetID uint64) {
	if monsterSyncMgr == nil {
		return
	}

	update := &MonsterUpdate{
		MonsterID:  monsterID,
		UpdateType: "status",
		Data: map[string]interface{}{
			"status":    status,
			"target_id": targetID,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	monsterSyncMgr.addToBatch(update)

	// 更新本地状态
	monsterSyncMgr.mutex.Lock()
	if state, ok := monsterSyncMgr.monsterStates[monsterID]; ok {
		state.Status = status
		state.TargetID = targetID
	}
	monsterSyncMgr.mutex.Unlock()

	monsterSyncMgr.syncStats.mutex.Lock()
	monsterSyncMgr.syncStats.StatusUpdates++
	monsterSyncMgr.syncStats.TotalUpdates++
	monsterSyncMgr.syncStats.mutex.Unlock()

	// 特殊状态处理
	if status == 3 { // 死亡
		monsterSyncMgr.syncStats.DeathEvents++
	} else if status == 0 { // 复活/生成
		monsterSyncMgr.syncStats.SpawnEvents++
	}
}

// SpawnMonster 怪物生成
func SpawnMonster(monsterID, mapID uint32, x, y int, hp, maxHP int) {
	if monsterSyncMgr == nil {
		return
	}

	update := &MonsterUpdate{
		MonsterID:  monsterID,
		MapID:      mapID,
		UpdateType: "spawn",
		Data: map[string]interface{}{
			"x":      x,
			"y":      y,
			"hp":     hp,
			"max_hp": maxHP,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	monsterSyncMgr.addToBatch(update)
	monsterSyncMgr.updateLocalState(monsterID, mapID, x, y)

	// 设置初始HP
	monsterSyncMgr.mutex.Lock()
	if state, ok := monsterSyncMgr.monsterStates[monsterID]; ok {
		state.HP = hp
		state.MaxHP = maxHP
		state.Status = 0 // 空闲状态
	}
	monsterSyncMgr.mutex.Unlock()

	monsterSyncMgr.syncStats.mutex.Lock()
	monsterSyncMgr.syncStats.SpawnEvents++
	monsterSyncMgr.syncStats.TotalUpdates++
	monsterSyncMgr.syncStats.mutex.Unlock()
}

// DeathMonster 怪物死亡
func DeathMonster(monsterID uint32) {
	if monsterSyncMgr == nil {
		return
	}

	update := &MonsterUpdate{
		MonsterID:  monsterID,
		UpdateType: "death",
		Data:       map[string]interface{}{},
		Timestamp:  time.Now().UnixMilli(),
	}

	monsterSyncMgr.addToBatch(update)

	// 更新本地状态为死亡
	monsterSyncMgr.mutex.Lock()
	if state, ok := monsterSyncMgr.monsterStates[monsterID]; ok {
		state.Status = 3 // 死亡状态
		state.HP = 0
	}
	monsterSyncMgr.mutex.Unlock()

	monsterSyncMgr.syncStats.mutex.Lock()
	monsterSyncMgr.syncStats.DeathEvents++
	monsterSyncMgr.syncStats.TotalUpdates++
	monsterSyncMgr.syncStats.mutex.Unlock()
}

// addToBatch 添加到批量缓冲
func (m *MonsterSyncManager) addToBatch(update *MonsterUpdate) {
	m.batchMutex.Lock()
	defer m.batchMutex.Unlock()

	m.batchBuffer = append(m.batchBuffer, update)

	// 如果达到最大批量大小，立即刷新
	if len(m.batchBuffer) >= m.maxBatchSize {
		m.flushBatchLocked()
		m.batchTimer.Reset(m.batchInterval)
	}
}

// flushBatch 刷新批量缓冲
func (m *MonsterSyncManager) flushBatch() {
	m.batchMutex.Lock()
	defer m.batchMutex.Unlock()
	m.flushBatchLocked()
}

// flushBatchLocked 刷新批量缓冲（需要已持有锁）
func (m *MonsterSyncManager) flushBatchLocked() {
	if len(m.batchBuffer) == 0 {
		return
	}

	startTime := time.Now()

	// 按地图ID分组更新
	updatesByMap := make(map[uint32][]*MonsterUpdate)
	for _, update := range m.batchBuffer {
		if update.MapID > 0 {
			updatesByMap[update.MapID] = append(updatesByMap[update.MapID], update)
		}
	}

	// 为每个地图创建广播消息
	for mapID, updates := range updatesByMap {
		msg := &BroadcastMessage{
			MessageType: CmdMonsterPositionUpdate,
			MapID:       mapID,
			Data: map[string]interface{}{
				"updates": updates,
				"count":   len(updates),
			},
			Priority: 5, // 中等优先级
		}

		// 通过回调发送到网关
		if m.gatewayCallback != nil {
			m.gatewayCallback(msg)
		}
	}

	// 更新统计
	duration := time.Since(startTime)
	m.syncStats.mutex.Lock()
	m.syncStats.BatchedUpdates += int64(len(m.batchBuffer))
	m.syncStats.SyncLatency = duration
	if len(m.batchBuffer) > 0 {
		m.syncStats.AvgBatchSize = (m.syncStats.AvgBatchSize*9 + float64(len(m.batchBuffer))) / 10
	}
	m.syncStats.mutex.Unlock()

	log.Printf(`📤 批量同步怪物状态: %d个更新, %d个地图, 耗时%v`,
		len(m.batchBuffer), len(updatesByMap), duration)

	// 清空缓冲
	m.batchBuffer = m.batchBuffer[:0]

	// 重置定时器
	if m.batchTimer != nil {
		m.batchTimer.Reset(m.batchInterval)
	}
}

// updateLocalState 更新本地状态缓存
func (m *MonsterSyncManager) updateLocalState(monsterID, mapID uint32, x, y int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state, ok := m.monsterStates[monsterID]
	if !ok {
		state = &MonsterState{
			MonsterID:  monsterID,
			MapID:      mapID,
			LastUpdate: time.Now().Unix(),
		}
		m.monsterStates[monsterID] = state
	}

	state.X = x
	state.Y = y
	state.LastUpdate = time.Now().Unix()
}

// GetMonsterState 获取怪物当前状态
func GetMonsterState(monsterID uint32) (*MonsterState, bool) {
	if monsterSyncMgr == nil {
		return nil, false
	}

	monsterSyncMgr.mutex.RLock()
	defer monsterSyncMgr.mutex.RUnlock()

	state, ok := monsterSyncMgr.monsterStates[monsterID]
	return state, ok
}

// GetMonstersInMap 获取地图中的所有怪物
func GetMonstersInMap(mapID uint32) []*MonsterState {
	if monsterSyncMgr == nil {
		return nil
	}

	monsterSyncMgr.mutex.RLock()
	defer monsterSyncMgr.mutex.RUnlock()

	var monsters []*MonsterState
	for _, state := range monsterSyncMgr.monsterStates {
		if state.MapID == mapID && state.Status != 3 { // 排除死亡怪物
			monsters = append(monsters, state)
		}
	}
	return monsters
}

// ForceFlush 强制刷新所有待处理的更新（用于重要事件）
func ForceFlush() {
	if monsterSyncMgr != nil {
		monsterSyncMgr.flushBatch()
	}
}

// GetMonsterSyncMetrics 获取怪物同步指标
func GetMonsterSyncMetrics() *MonsterSyncMetrics {
	if monsterSyncMgr == nil {
		return nil
	}
	return monsterSyncMgr.syncStats
}

// ShutdownMonsterSync 关闭怪物同步管理器
func ShutdownMonsterSync() {
	if monsterSyncMgr == nil {
		return
	}

	// 刷新所有剩余的更新
	monsterSyncMgr.flushBatch()

	// 停止定时器
	if monsterSyncMgr.batchTimer != nil {
		monsterSyncMgr.batchTimer.Stop()
	}

	log.Println("🔒 怪物同步管理器已关闭")
}
