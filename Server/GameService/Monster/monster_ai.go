package monster

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	common "game-server/Common"
)

// ========== 接口定义（避免循环依赖）==========

// BattleServiceInterface 战斗服务接口 - 由外部注入
type BattleServiceInterface interface {
	// MonsterAttackPlayer 怪物攻击玩家
	MonsterAttackPlayer(monsterInstanceID uint64, playerID uint64) *common.MonsterAttackResult
}

// MapServiceInterface 地图服务接口（用于碰撞检测）
type MapServiceInterface interface {
	// CanMove 检查瓦片是否可通行
	CanMove(mapID uint32, tileX, tileY int) bool
}

// EntityCollisionChecker 实体碰撞检测接口（用于检查玩家/怪物位置）
type EntityCollisionChecker interface {
	// CheckEntityCollision 检查实体碰撞
	// 返回值: (是否有碰撞, 碰撞的实体ID)
	CheckEntityCollision(mapID uint32, x, y int, excludeID uint64) (bool, uint64)

	// GetEntityPosition 获取实体位置
	GetEntityPosition(entityID uint64) (x, y int, exists bool)
}

// 全局战斗服务实例（通过依赖注入设置）
var globalBattleSvc BattleServiceInterface

// 全局地图服务实例（通过依赖注入设置）
var globalMapSvc MapServiceInterface

// 全局实体碰撞检测器（通过依赖注入设置）
var globalEntityChecker EntityCollisionChecker

// SetBattleService 注入战斗服务实例（在main.go中调用）
func SetBattleService(svc BattleServiceInterface) {
	globalBattleSvc = svc
	log.Printf("✅ AI系统: 战斗服务已注入")
}

// SetMapService 注入地图服务实例（用于碰撞检测）
func SetMapService(svc MapServiceInterface) {
	globalMapSvc = svc
	log.Printf("✅ AI系统: 地图服务已注入（碰撞检测）")
}

// SetEntityChecker 注入实体碰撞检测器（用于玩家/怪物碰撞）
func SetEntityChecker(checker EntityCollisionChecker) {
	globalEntityChecker = checker
	log.Printf("✅ AI系统: 实体碰撞检测器已注入")
}

// AIState AI状态类型
type AIState uint8

const (
	AIStateIdle    AIState = 0 // 空闲
	AIStatePatrol  AIState = 1 // 巡逻
	AIStateAlert   AIState = 2 // 警戒
	AIStateChase   AIState = 3 // 追击
	AIStateCombat  AIState = 4 // 战斗
	AIStateFlee    AIState = 5 // 逃跑
	AIStateDead    AIState = 6 // 死亡
	AIStateRespawn AIState = 7 // 复活中
)

// AIMonster AI怪物实例（扩展MonsterInstance）
type AIMonster struct {
	*MonsterInstance
	State           AIState // 当前状态
	PrevState       AIState // 前一状态
	PatrolPoints    []Point // 巡逻路径点
	CurrentPatrol   int     // 当前巡逻点索引
	HomeX           int     // 出生点X
	HomeY           int     // 出生点Y
	AlertRange      int     // 警戒范围
	ChaseRange      int     // 追击范围
	AttackRange     int     // 攻击范围
	AttackCooldown  int64   // 攻击冷却(毫秒)
	LastAttackTime  int64   // 上次攻击时间
	TargetID        uint64  // 当前目标ID
	FleeThreshold   float64 // 逃跑血量阈值(0-1)
	LastUpdateTime  int64   // 上次更新时间
	RespawnDuration int64   // 复活时间(毫秒)
}

// Point 坐标点
type Point struct {
	X int
	Y int
}

// AIService 怪物AI服务
type AIService struct {
	monsters map[uint64]*AIMonster // key=实例ID
	mu       sync.RWMutex
	ticker   *time.Ticker
	stopCh   chan struct{}

	// 位置同步回调（用于广播怪物位置给客户端）
	positionSyncCallback func(map[uint32][]MonsterPositionInfo)
}

// MonsterPositionInfo 怪物位置信息（用于同步）
type MonsterPositionInfo struct {
	InstanceID uint64 `json:"instance_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	State      int    `json:"state"` // AI状态
	HP         int    `json:"hp"`    // 当前血量
}

// NewAIService 创建AI服务
func NewAIService() *AIService {
	return &AIService{
		monsters: make(map[uint64]*AIMonster),
	}
}

// RegisterMonster 注册怪物到AI系统
func (s *AIService) RegisterMonster(monster *MonsterInstance, config *common.MonsterBaseConfig) *AIMonster {
	s.mu.Lock()
	defer s.mu.Unlock()

	aiMonster := &AIMonster{
		MonsterInstance: monster,
		State:           AIStateIdle,
		HomeX:           monster.X,
		HomeY:           monster.Y,
		AlertRange:      config.AttackRange + 2, // 警戒范围 > 攻击范围
		ChaseRange:      config.ChaseRange,
		AttackRange:     config.AttackRange,
		AttackCooldown:  2000, // 默认2秒攻击冷却
		FleeThreshold:   0.1,  // 血量<10%时逃跑
		RespawnDuration: int64(config.RespawnTime) * 1000,
	}

	// 根据AI类型设置行为参数
	switch config.AIType {
	case 0: // 被动型 - 不主动攻击
		aiMonster.AlertRange = 0
	case 1: // 主动型 - 正常追击
		aiMonster.FleeThreshold = 0.05 // 5%血量才逃跑
	case 2: // 攻击型 - BOSS/精英，不逃跑
		aiMonster.FleeThreshold = 0 // 永不逃跑
		aiMonster.ChaseRange *= 2   // 追击范围翻倍
	}

	// 生成随机巡逻路径
	aiMonster.generatePatrolPath()

	s.monsters[monster.ID] = aiMonster
	return aiMonster
}

// generatePatrolPath 生成巡逻路径
func (m *AIMonster) generatePatrolPath() {
	// 生成3-5个巡逻点
	numPoints := 3 + rand.Intn(3)
	m.PatrolPoints = make([]Point, numPoints)
	m.PatrolPoints[0] = Point{X: m.HomeX, Y: m.HomeY} // 第一个点是出生点

	for i := 1; i < numPoints; i++ {
		// 在出生点周围3-6格范围内随机生成点
		offsetX := rand.Intn(7) - 3
		offsetY := rand.Intn(7) - 3
		m.PatrolPoints[i] = Point{
			X: m.HomeX + offsetX,
			Y: m.HomeY + offsetY,
		}
	}
}

// SetPositionSyncCallback 设置位置同步回调
func (s *AIService) SetPositionSyncCallback(callback func(map[uint32][]MonsterPositionInfo)) {
	s.positionSyncCallback = callback
}

// Start 启动AI主循环
func (s *AIService) Start() {
	s.ticker = time.NewTicker(500 * time.Millisecond) // 每500ms更新一次
	s.stopCh = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.Update()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop 停止AI服务
func (s *AIService) Stop() {
	if s.stopCh != nil {
		close(s.stopCh)
	}
	if s.ticker != nil {
		s.ticker.Stop()
	}
}

// Update 更新所有怪物AI状态
func (s *AIService) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	for _, monster := range s.monsters {
		monster.Update(now)
	}

	// 位置同步：收集所有活跃怪物的位置并广播给客户端
	if s.positionSyncCallback != nil {
		positionMap := make(map[uint32][]MonsterPositionInfo)

		for instanceID, monster := range s.monsters {
			// 只同步非死亡状态的怪物
			if monster.State != AIStateDead && monster.State != AIStateRespawn {
				info := MonsterPositionInfo{
					InstanceID: instanceID,
					X:          monster.X,
					Y:          monster.Y,
					State:      int(monster.State),
					HP:         monster.CurrentHP,
				}

				// 按地图ID分组
				positionMap[monster.MapID] = append(positionMap[monster.MapID], info)
			}
		}

		// 触发回调进行广播（异步执行，避免阻塞AI更新）
		if len(positionMap) > 0 {
			go s.positionSyncCallback(positionMap)

			// 调试日志：每10次打印一次（避免刷屏）
			if now%5000 < 500 {
				totalMonsters := 0
				for _, positions := range positionMap {
					totalMonsters += len(positions)
				}
				log.Printf("📍 AI位置同步: %d个地图, 共%d个活跃怪物",
					len(positionMap), totalMonsters)
			}
		}
	}
}

// Update 更新单个怪物AI状态
func (m *AIMonster) Update(now int64) {
	m.LastUpdateTime = now

	switch m.State {
	case AIStateIdle:
		m.updateIdle(now)
	case AIStatePatrol:
		m.updatePatrol(now)
	case AIStateAlert:
		m.updateAlert(now)
	case AIStateChase:
		m.updateChase(now)
	case AIStateCombat:
		m.updateCombat(now)
	case AIStateFlee:
		m.updateFlee(now)
	case AIStateDead:
		m.updateDead(now)
	case AIStateRespawn:
		m.updateRespawn(now)
	}
}

// updateIdle 状态：空闲 -> 巡逻
func (m *AIMonster) updateIdle(now int64) {
	// 空闲一段时间后开始巡逻
	if rand.Intn(100) < 10 { // 10%概率进入巡逻
		m.changeState(AIStatePatrol)
	}
}

// updatePatrol 状态：沿路径巡逻
func (m *AIMonster) updatePatrol(now int64) {
	// 检查是否有目标进入警戒范围
	if m.TargetID != 0 && m.checkInRange(m.TargetID, m.AlertRange) {
		m.changeState(AIStateAlert)
		return
	}

	// 移动到下一个巡逻点
	if len(m.PatrolPoints) == 0 {
		return
	}

	target := m.PatrolPoints[m.CurrentPatrol]
	distance := m.calculateDistance(target.X, target.Y)

	if distance <= 0 { // 到达目标点
		m.CurrentPatrol = (m.CurrentPatrol + 1) % len(m.PatrolPoints)
		if m.CurrentPatrol == 0 { // 完成一圈巡逻后可能休息
			if rand.Intn(100) < 30 {
				m.changeState(AIStateIdle)
				return
			}
		}
	} else {
		// 向目标点移动一步
		m.moveToward(target.X, target.Y)
	}
}

// updateAlert 状态：警戒 -> 追击 或 空闲
func (m *AIMonster) updateAlert(now int64) {
	if m.TargetID == 0 {
		m.changeState(AIStateIdle)
		return
	}

	// 检查目标是否还在警戒范围
	if !m.checkInRange(m.TargetID, m.AlertRange) {
		m.TargetID = 0
		m.changeState(AIStatePatrol)
		return
	}

	// 目标进入攻击范围 -> 追击
	if m.checkInRange(m.TargetID, m.AttackRange) {
		m.changeState(AIStateChase)
		return
	}

	// 警戒一段时间后开始追击
	if rand.Intn(100) < 30 { // 30%概率开始追击
		m.changeState(AIStateChase)
	}
}

// updateChase 状态：追击目标
func (m *AIMonster) updateChase(now int64) {
	if m.TargetID == 0 {
		m.changeState(AIStatePatrol)
		return
	}

	// 检查是否在攻击范围内
	if m.checkInRange(m.TargetID, m.AttackRange) {
		m.changeState(AIStateCombat)
		return
	}

	// 检查是否超出追击范围
	if !m.checkInRange(m.TargetID, m.ChaseRange) {
		m.TargetID = 0
		m.changeState(AIStatePatrol) // 返回出生点
		return
	}

	// 向目标移动
	targetPos := getTargetPosition(m.TargetID)
	if targetPos != nil {
		m.moveToward(targetPos.X, targetPos.Y)
	}
}

// updateCombat 状态：战斗中
func (m *AIMonster) updateCombat(now int64) {
	if m.TargetID == 0 {
		m.changeState(AIStatePatrol)
		return
	}

	// 检查血量是否低于逃跑阈值
	if m.FleeThreshold > 0 && float64(m.CurrentHP)/float64(m.MaxHP) < m.FleeThreshold {
		m.changeState(AIStateFlee)
		return
	}

	// 检查目标是否离开攻击范围
	if !m.checkInRange(m.TargetID, m.AttackRange+1) {
		m.changeState(AIStateChase)
		return
	}

	// 检查攻击冷却
	if now-m.LastAttackTime < m.AttackCooldown {
		return
	}

	// 执行攻击
	m.performAttack(now)
}

// updateFlee 状态：逃跑（修复整数除法问题）
func (m *AIMonster) updateFlee(now int64) {
	// 向远离目标的方向移动
	targetPos := getTargetPosition(m.TargetID)
	if targetPos != nil {
		// 计算反方向
		dx := m.X - targetPos.X
		dy := m.Y - targetPos.Y

		// 归一化并放大（使用float64计算避免整数除法截断）
		length := sqrt(float64(dx*dx + dy*dy))
		if length > 0 {
			fleeX := m.X + int(math.Round(float64(dx*3)/float64(length)))
			fleeY := m.Y + int(math.Round(float64(dy*3)/float64(length)))
			m.X = fleeX
			m.Y = fleeY
		}
	}

	// 跑出追击范围后恢复巡逻
	if m.TargetID == 0 || !m.checkInRange(m.TargetID, m.ChaseRange) {
		m.TargetID = 0
		m.changeState(AIStatePatrol)
	}
}

// updateDead 状态：死亡处理
func (m *AIMonster) updateDead(now int64) {
	m.Status = 4 // 死亡状态
	m.RespawnAt = now + m.RespawnDuration
	m.changeState(AIStateRespawn)
}

// updateRespawn 状态：等待复活
func (m *AIMonster) updateRespawn(now int64) {
	if now >= m.RespawnAt {
		// 复活
		m.CurrentHP = m.MaxHP
		m.X = m.HomeX
		m.Y = m.HomeY
		m.Status = 0 // 活着
		m.TargetID = 0
		m.changeState(AIStateIdle)
	}
}

// changeState 切换状态
func (m *AIMonster) changeState(newState AIState) {
	if m.State != newState {
		m.PrevState = m.State
		m.State = newState
		// log.Printf("怪物 %d 状态变化: %d -> %d", m.ID, m.PrevState, m.State)
	}
}

// performAttack 执行攻击 - 通过注入的战斗服务接口完成实际伤害计算
func (m *AIMonster) performAttack(now int64) {
	m.LastAttackTime = now

	if m.TargetID == 0 {
		return
	}

	// 检查战斗服务是否已注入
	if globalBattleSvc == nil {
		log.Printf("警告: 怪物%d无法攻击玩家%d: 战斗服务未注入（需在main.go中调用SetBattleService）",
			m.ID, m.TargetID)
		return
	}

	// ✅ 通过接口调用战斗服务（避免循环依赖）
	result := globalBattleSvc.MonsterAttackPlayer(m.ID, m.TargetID)

	if result == nil || result.IsMiss {
		return // 未命中或错误，无需后续处理
	}

	log.Printf("⚔️ 怪物[%d]攻击玩家%d: 伤害=%d, 暴击=%v, 玩家剩余HP=%d/%d",
		m.ID, m.TargetID, result.Damage, result.IsCrit,
		result.PlayerHP, result.PlayerMaxHP)

	// TODO: 推送战斗结果给客户端（通过WebSocket）
	// 需要调用gameService的推送接口通知前端显示伤害数字、更新血条等
}

// moveToward 向目标移动（带完整碰撞检测）
func (m *AIMonster) moveToward(targetX, targetY int) {
	dx := targetX - m.X
	dy := targetY - m.Y
	distance := sqrt(float64(dx*dx + dy*dy))

	if distance <= 0 {
		return
	}

	// 每次移动1格
	speed := 1
	if distance < speed {
		speed = int(distance)
	}

	// 计算目标位置（使用float64避免整数除法截断）
	newX := m.X + int(math.Round(float64(dx*speed)/float64(distance)))
	newY := m.Y + int(math.Round(float64(dy*speed)/float64(distance)))

	// ===== 碰撞检测系统 =====

	// 1. 检查地图瓦片是否可通行
	if globalMapSvc != nil {
		if !globalMapSvc.CanMove(m.MapID, newX, newY) {
			// 地图不可通行，尝试寻找替代方向
			m.tryAlternativePath(targetX, targetY)
			return
		}
	}

	// 2. 检查实体间碰撞（玩家/其他怪物）
	if globalEntityChecker != nil {
		hasCollision, collisionID := globalEntityChecker.CheckEntityCollision(
			m.MapID, newX, newY, m.ID,
		)

		if hasCollision {
			log.Printf("🛡️ 怪物 %d (%s) 碰撞检测: 目标(%d,%d) → 与实体 %d 碰撞，执行避让",
				m.ID, m.Name, newX, newY, collisionID)

			// 有实体碰撞，尝试推开或绕行
			m.handleEntityCollision(newX, newY, collisionID, targetX, targetY)
			return
		}
	}

	// 所有检测通过，执行移动
	m.X = newX
	m.Y = newY
}

// tryAlternativePath 尝试替代路径（当地图不可通行时）
func (m *AIMonster) tryAlternativePath(targetX, targetY int) {
	// 尝试8个方向的替代路径
	directions := []struct {
		dx, dy int
	}{
		{1, 0}, {0, 1}, {-1, 0}, {0, -1}, // 上下左右
		{1, 1}, {1, -1}, {-1, 1}, {-1, -1}, // 对角线
	}

	// 随机打乱方向顺序，避免所有怪物都走同一方向
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		newX := m.X + dir.dx
		newY := m.Y + dir.dy

		// 检查地图碰撞
		canMove := true
		if globalMapSvc != nil {
			canMove = globalMapSvc.CanMove(m.MapID, newX, newY)
		}

		// 检查实体碰撞
		hasEntityCollision := false
		if canMove && globalEntityChecker != nil {
			hasEntityCollision, _ = globalEntityChecker.CheckEntityCollision(
				m.MapID, newX, newY, m.ID,
			)
		}

		if canMove && !hasEntityCollision {
			// 找到可行路径
			m.X = newX
			m.Y = newY
			return
		}
	}

	// 所有方向都不可行，原地不动
	log.Printf("⚠️ 怪物 %d (%s) 无法移动: 被阻挡在(%d,%d)", m.ID, m.Name, m.X, m.Y)
}

// handleEntityCollision 处理实体碰撞（玩家或其他怪物）
func (m *AIMonster) handleEntityCollision(newX, newY int, collisionID uint64, targetX, targetY int) {
	// 获取碰撞实体的位置
	var entityX, entityY int

	if globalEntityChecker != nil {
		var exists bool
		entityX, entityY, exists = globalEntityChecker.GetEntityPosition(collisionID)

		if !exists {
			// 实体不存在，允许移动
			m.X = newX
			m.Y = newY
			return
		}
	}

	// 根据AI状态决定行为
	switch m.State {
	case AIStateChase, AIStateCombat:
		// 追击/战斗状态：已经到达攻击范围或被阻挡，停止移动
		// （让战斗系统处理攻击逻辑）
		return

	case AIStateFlee:
		// 逃跑状态：尝试绕过障碍物
		m.tryAlternativePath(
			m.X+(m.X-entityX)*2, // 向远离实体方向逃跑
			m.Y+(m.Y-entityY)*2,
		)
		return

	default:
		// 巡逻/警戒状态：尝试绕行或停止
		// 有50%概率尝试绕行，50%概率等待
		if rand.Float64() < 0.5 {
			m.tryAlternativePath(targetX, targetY)
		}
		// 否则原地不动
		return
	}
}

// checkInRange 检查目标是否在范围内
func (m *AIMonster) checkInRange(targetID uint64, rangeVal int) bool {
	pos := getTargetPosition(targetID)
	if pos == nil {
		return false
	}

	distance := m.calculateDistance(pos.X, pos.Y)
	return distance <= rangeVal
}

// calculateDistance 计算与目标的距离
func (m *AIMonster) calculateDistance(targetX, targetY int) int {
	dx := m.X - targetX
	dy := m.Y - targetY
	return int(sqrt(float64(dx*dx + dy*dy)))
}

// SetTarget 设置目标（由外部调用）
func (m *AIMonster) SetTarget(targetID uint64) {
	m.TargetID = targetID

	// 根据距离决定状态
	if m.checkInRange(targetID, m.AttackRange) {
		m.changeState(AIStateCombat)
	} else if m.checkInRange(targetID, m.ChaseRange) {
		m.changeState(AIStateChase)
	} else if m.checkInRange(targetID, m.AlertRange) {
		m.changeState(AIStateAlert)
	}
}

// OnHurted 受伤回调（被玩家攻击时调用）
func (m *AIMonster) OnHurted(attackerID uint64) {
	m.SetTarget(attackerID)
}

// 辅助函数
func sqrt(n float64) int {
	return int(mathSqrt(n))
}

func mathSqrt(x float64) float64 {
	// 简单的整数平方根近似
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// getTargetPosition 获取目标位置（需要外部提供实现）
var getTargetPosition func(uint64) *Point

// SetTargetPositionSetter 设置目标位置获取函数
func SetTargetPositionSetter(fn func(uint64) *Point) {
	getTargetPosition = fn
}

// GetMonsterAI 获取怪物的AI实例
func (s *AIService) GetMonsterAI(instanceID uint64) (*AIMonster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ai, ok := s.monsters[instanceID]
	return ai, ok
}

// GetAllMonstersInMap 获取地图中的所有怪物
func (s *AIService) GetAllMonstersInMap(mapID uint32) []*AIMonster {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*AIMonster
	for _, m := range s.monsters {
		if m.MapID == mapID && m.State != AIStateDead && m.State != AIStateRespawn {
			result = append(result, m)
		}
	}
	return result
}

// OnMonsterHurted 实现AIServiceInterface接口 - 通知怪物被攻击
func (s *AIService) OnMonsterHurted(monsterID uint64, attackerID uint64) {
	s.mu.RLock()
	ai, exists := s.monsters[monsterID]
	s.mu.RUnlock()

	if !exists || ai == nil {
		return
	}

	// 调用怪物的受伤处理逻辑
	ai.SetTarget(attackerID)
}
