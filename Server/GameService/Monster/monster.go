package monster

import (
	"errors"
	common "game-server/Common"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// MonsterInstance 怪物实例
type MonsterInstance struct {
	ID        uint64 // 实例ID
	BaseID    uint32 // 基础ID
	Name      string // 名称
	Level     uint32 // 等级
	Type      uint8  // 类型: 0=普通, 1=精英, 2=BOSS
	MapID     uint32 // 地图ID
	X         int    // X坐标
	Y         int    // Y坐标
	CurrentHP int    // 当前HP
	MaxHP     int    // 最大HP
	Attack    int    // 攻击力
	Defense   int    // 防御力
	Speed     int    // 速度
	Status    uint8  // 状态: 0=空闲, 1=巡逻, 2=追击, 3=战斗, 4=死亡
	TargetID  uint64 // 攻击目标
	RespawnAt int64  // 复活时间戳
}

// NPCInstance NPC实例
type NPCInstance struct {
	ID     uint64
	BaseID uint32
	Name   string
	Type   uint8 // 类型: 1=普通NPC, 2=商店NPC, 3=任务NPC, 4=仓库NPC, 5=传送NPC
	MapID  uint32
	X      int
	Y      int
	Facing uint8   // 朝向
	Dialog string  // 对话文本
	ShopID *uint32 // 商店ID(商店NPC)
}

// Service 怪物/NPC服务
type Service struct {
	monsters  map[uint64]*MonsterInstance // key=实例ID
	npcs      map[uint64]*NPCInstance     // key=实例ID
	monsterID uint64                      // 自增ID
	npcID     uint64                      // 自增ID
	mu        sync.RWMutex
	battleSvc BattleServiceInterface // 使用接口避免循环依赖
}

// NewService 创建服务
func NewService(battleSvc BattleServiceInterface) *Service {
	return &Service{
		monsters:  make(map[uint64]*MonsterInstance),
		npcs:      make(map[uint64]*NPCInstance),
		monsterID: 0,
		npcID:     0,
		battleSvc: battleSvc,
	}
}

// SpawnMonster 生成怪物
func (s *Service) SpawnMonster(baseID uint32, mapID uint32, x, y int) (*MonsterInstance, error) {
	config := common.GetMonsterConfig(baseID)
	if config == nil || config.MapID != mapID {
		return nil, errors.New("怪物不存在")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.monsterID++
	monster := &MonsterInstance{
		ID:        s.monsterID,
		BaseID:    config.ID,
		Name:      config.Name,
		Level:     config.Level,
		Type:      config.Type,
		MapID:     mapID,
		X:         x,
		Y:         y,
		CurrentHP: config.Hp,
		MaxHP:     config.Hp,
		Attack:    config.Attack,
		Defense:   config.Defense,
		Speed:     config.Speed,
		Status:    0,
	}

	s.monsters[s.monsterID] = monster
	return monster, nil
}

// GetMonster 获取怪物实例
func (s *Service) GetMonster(instanceID uint64) (*MonsterInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.monsters[instanceID]
	return m, ok
}

// GetMonstersByMap 获取地图上的所有怪物
func (s *Service) GetMonstersByMap(mapID uint32) []*MonsterInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*MonsterInstance
	for _, m := range s.monsters {
		if m.MapID == mapID && m.Status != 4 {
			result = append(result, m)
		}
	}
	return result
}

// GetAllMonsters 获取所有怪物实例（包括死亡）
func (s *Service) GetAllMonsters() []*MonsterInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*MonsterInstance, 0, len(s.monsters))
	for _, m := range s.monsters {
		result = append(result, m)
	}
	return result
}

// MonsterTakeDamage 怪物受伤
func (s *Service) MonsterTakeDamage(instanceID uint64, damage int) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	monster, ok := s.monsters[instanceID]
	if !ok {
		return 0, false
	}

	monster.CurrentHP -= damage
	if monster.CurrentHP <= 0 {
		monster.CurrentHP = 0
		monster.Status = 4 // 死亡
		return 0, true
	}

	return monster.CurrentHP, false
}

// MonsterDie 怪物死亡处理
func (s *Service) MonsterDie(instanceID uint64) (exp int, gold int, drops []uint32, err error) {
	s.mu.Lock()
	monster, ok := s.monsters[instanceID]
	if !ok {
		s.mu.Unlock()
		return 0, 0, nil, errors.New("怪物不存在")
	}
	s.mu.Unlock()

	// 获取怪物基础数据
	config := common.GetMonsterConfig(monster.BaseID)
	if config == nil {
		return 0, 0, nil, errors.New("怪物基础数据不存在")
	}

	// 计算经验
	exp = config.Exp

	// 计算金币（本地实现，避免依赖Battle包）
	gold = calculateGoldDrop(config.GoldMin, config.GoldMax)

	// 计算掉落
	if config.DropGroupID != nil {
		drops, err = s.RollDrops(*config.DropGroupID)
		if err != nil {
			drops = nil
		}
	}

	// 设置复活时间
	s.mu.Lock()
	monster.Status = 4
	monster.RespawnAt = time.Now().Unix() + int64(config.RespawnTime)
	s.mu.Unlock()

	return exp, gold, drops, nil
}

// RollDrops 掉落roll
func (s *Service) RollDrops(monsterID uint32) ([]uint32, error) {
	dropsConfig := common.GetDropsByMonsterID(monsterID)
	if len(dropsConfig) == 0 {
		return nil, nil
	}

	var result []uint32
	for _, drop := range dropsConfig {
		if rand.Float64()*10000 < float64(drop.DropRate) {
			// 掉落
			count := drop.DropMin
			if drop.DropMax > drop.DropMin {
				count = drop.DropMin + uint32(rand.Intn(int(drop.DropMax-drop.DropMin+1)))
			}
			for i := uint32(0); i < count; i++ {
				result = append(result, drop.ItemID)
			}
		}
	}

	return result, nil
}

// RespawnMonster 复活怪物
func (s *Service) RespawnMonster(instanceID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	monster, ok := s.monsters[instanceID]
	if !ok {
		return errors.New("怪物不存在")
	}

	// 检查是否到复活时间
	if time.Now().Unix() < monster.RespawnAt {
		return errors.New("还没到复活时间")
	}

	// 获取基础数据重置
	config := common.GetMonsterConfig(monster.BaseID)
	if config == nil {
		return errors.New("怪物基础数据不存在")
	}

	monster.CurrentHP = config.Hp
	monster.Status = 0
	monster.TargetID = 0

	return nil
}

// MonsterSetTarget 设置怪物目标
func (s *Service) MonsterSetTarget(instanceID uint64, targetID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	monster, ok := s.monsters[instanceID]
	if !ok {
		return errors.New("怪物不存在")
	}

	monster.TargetID = targetID
	monster.Status = 2 // 追击状态
	return nil
}

// MonsterClearTarget 清除怪物目标
func (s *Service) MonsterClearTarget(instanceID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	monster, ok := s.monsters[instanceID]
	if !ok {
		return errors.New("怪物不存在")
	}

	monster.TargetID = 0
	monster.Status = 0 // 空闲状态
	return nil
}

// SpawnNPC 生成NPC
func (s *Service) SpawnNPC(baseID uint32, mapID uint32, x, y int) (*NPCInstance, error) {
	config := common.GetNPCConfig(baseID)
	if config == nil || config.MapID != mapID {
		return nil, errors.New("NPC不存在")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.npcID++
	npc := &NPCInstance{
		ID:     s.npcID,
		BaseID: config.ID,
		Name:   config.Name,
		Type:   config.Type,
		MapID:  mapID,
		X:      x,
		Y:      y,
		Facing: config.Face,
		Dialog: config.DialogText,
		ShopID: config.ShopID,
	}

	s.npcs[s.npcID] = npc
	return npc, nil
}

// GetNPC 获取NPC实例
func (s *Service) GetNPC(instanceID uint64) (*NPCInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.npcs[instanceID]
	return n, ok
}

// GetNPCsByMap 获取地图上的NPC
func (s *Service) GetNPCsByMap(mapID uint32) []*NPCInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*NPCInstance
	for _, n := range s.npcs {
		if n.MapID == mapID {
			result = append(result, n)
		}
	}
	return result
}

// NPCInteract NPC交互(对话)
func (s *Service) NPCInteract(npcID uint64) (string, *uint32, error) {
	npc, ok := s.GetNPC(npcID)
	if !ok {
		return "", nil, errors.New("NPC不存在")
	}

	return npc.Dialog, npc.ShopID, nil
}

// IsMonsterAlive 检查怪物是否存活
func (s *Service) IsMonsterAlive(instanceID uint64) bool {
	monster, ok := s.GetMonster(instanceID)
	if !ok {
		return false
	}
	return monster.Status != 4 && monster.CurrentHP > 0
}

// GetMonsterPosition 获取怪物位置
func (s *Service) GetMonsterPosition(instanceID uint64) (int, int, bool) {
	monster, ok := s.GetMonster(instanceID)
	if !ok {
		return 0, 0, false
	}
	return monster.X, monster.Y, true
}

// MonsterMove 怪物移动
func (s *Service) MonsterMove(instanceID uint64, x, y int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	monster, ok := s.monsters[instanceID]
	if !ok {
		return errors.New("怪物不存在")
	}

	monster.X = x
	monster.Y = y
	return nil
}

// DropGroup 掉落组
type DropGroup struct {
	ID        uint32 `gorm:"primaryKey;column:id" json:"id"`
	MonsterID uint32 `gorm:"column:monster_id;index" json:"monster_id"`
	ItemID    uint32 `gorm:"column:item_id" json:"item_id"`
	DropRate  uint32 `gorm:"column:drop_rate" json:"drop_rate"` // 万分比
	DropMin   uint32 `gorm:"column:drop_min" json:"drop_min"`
	DropMax   uint32 `gorm:"column:drop_max" json:"drop_max"`
}

func (DropGroup) TableName() string {
	return "drop_group"
}

// ==================== 全局服务实例 ====================

var (
	globalService   *Service
	globalAIService *AIService

	// 位置广播回调（由外部设置，用于通过Gateway广播怪物位置）
	positionBroadcastFunc func(map[uint32][]MonsterPositionInfo)

	// 玩家受击推送回调（由外部设置，用于通过Gateway推送怪物攻击玩家的结果给被攻击玩家）
	// 参数: targetRoleID(被攻击玩家), monsterName(怪物名), result(攻击结果)
	playerDamagePushFunc func(targetRoleID uint64, monsterName string, result *common.MonsterAttackResult)
)

// SetPlayerDamagePushFunc 设置玩家受击推送函数（在main.go中调用）
// 用于将怪物攻击玩家的结果通过Gateway WebSocket推送给被攻击的客户端
func SetPlayerDamagePushFunc(fn func(targetRoleID uint64, monsterName string, result *common.MonsterAttackResult)) {
	playerDamagePushFunc = fn
}

// InitService 初始化怪物服务（在main.go中调用）
func InitService(battleSvc BattleServiceInterface) {
	globalService = NewService(battleSvc)
	globalAIService = NewAIService()

	// 设置AI目标位置获取函数
	SetTargetPositionSetter(func(targetID uint64) *Point {
		if globalService == nil {
			return nil
		}
		// 尝试从玩家位置获取
		if x, y, ok := GetPlayerPositionFromMap(targetID); ok {
			return &Point{X: x, Y: y}
		}
		return nil
	})

	// 设置位置同步回调（将广播给所有客户端）
	globalAIService.SetPositionSyncCallback(func(positionMap map[uint32][]MonsterPositionInfo) {
		if positionBroadcastFunc != nil && len(positionMap) > 0 {
			positionBroadcastFunc(positionMap)
		}
	})

	// 启动AI系统
	globalAIService.Start()
}

// SetPositionBroadcastFunc 设置位置广播函数（在main.go中调用）
func SetPositionBroadcastFunc(fn func(map[uint32][]MonsterPositionInfo)) {
	positionBroadcastFunc = fn
}

// InitMapMonsters 初始化指定地图的怪物（在loadManagedMaps中调用）
// 优先级：配置文件 > 算法生成
func InitMapMonsters(mapID uint32) {
	if globalService == nil {
		log.Printf("警告: 怪物服务未初始化，无法生成地图%d的怪物", mapID)
		return
	}

	// ========== 方式1：尝试从配置文件加载 ==========
	spawnConfig := common.GetMapSpawnConfig(mapID)
	if spawnConfig != nil && len(spawnConfig.SpawnPoints) > 0 {
		log.Printf("📍 地图%d[%s]: 使用配置文件生成怪物 (%d个生成点)",
			mapID, spawnConfig.MapName, len(spawnConfig.SpawnPoints))
		initMonstersFromConfig(mapID, *spawnConfig)
		return
	}

	// ========== 方式2：Fallback到算法生成 ==========
	log.Printf("⚠️ 地图%d: 未找到生成点配置，使用默认算法生成", mapID)
	initMonstersFromAlgorithm(mapID)
}

// initMonstersFromConfig 从配置文件生成怪物
func initMonstersFromConfig(mapID uint32, config common.MapSpawnConfig) {
	spawnedCount := 0
	skippedCount := 0

	for _, spawnPoint := range config.SpawnPoints {
		// 跳过未激活的生成点
		if !spawnPoint.IsActive {
			skippedCount++
			continue
		}

		// 验证怪物模板是否存在
		monsterConfig := common.GetMonsterConfig(spawnPoint.BaseMonsterID)
		if monsterConfig == nil {
			log.Printf("❌ 生成点[%d-%s]: 怪物模板%d不存在，跳过",
				spawnPoint.ID, spawnPoint.Name, spawnPoint.BaseMonsterID)
			skippedCount++
			continue
		}

		// 根据配置数量生成怪物
		for i := 0; i < spawnPoint.Count; i++ {
			// 计算生成位置（支持随机偏移）
			x, y := spawnPoint.X, spawnPoint.Y
			if spawnPoint.SpawnRadius > 0 {
				x, y = generatePositionWithRadius(
					spawnPoint.X, spawnPoint.Y,
					spawnPoint.SpawnRadius,
				)
			}

			// 生成怪物实例
			monster, err := globalService.SpawnMonster(
				spawnPoint.BaseMonsterID, mapID, x, y,
			)
			if err != nil {
				log.Printf("❌ 生成点[%d]第%d个怪物失败: %v",
					spawnPoint.ID, i+1, err)
				continue
			}

			// 如果有AI类型覆盖，临时修改配置
			finalConfig := *monsterConfig
			if spawnPoint.AITypeOverride != nil {
				finalConfig.AIType = *spawnPoint.AITypeOverride
			}

			// 注册到AI系统
			if globalAIService != nil {
				globalAIService.RegisterMonster(monster, &finalConfig)
			}

			spawnedCount++

			// 记录详细日志（仅第一个）
			if i == 0 {
				log.Printf("  ✅ [%s] %s #%d → (%d,%d) | 数量:%d | 半径:%d格",
					spawnPoint.Name,
					finalConfig.Name,
					monster.ID,
					x, y,
					spawnPoint.Count,
					spawnPoint.SpawnRadius,
				)
			}
		}
	}

	// 输出统计信息
	globalSettings := config.GlobalSettings
	log.Printf("📊 地图%d生成完成: 成功=%d, 跳过=%d | 最大上限=%d | 自动复活=%v",
		mapID, spawnedCount, skippedCount,
		globalSettings.MaxMonstersPerMap,
		globalSettings.AutoRespawn,
	)
}

// initMonstersFromAlgorithm 使用算法自动生成（备用方案）
func initMonstersFromAlgorithm(mapID uint32) {
	// 获取该地图的所有怪物配置
	var mapMonsters []common.MonsterBaseConfig
	for _, m := range common.GameConfig.Monsters {
		if m.MapID == mapID {
			mapMonsters = append(mapMonsters, m)
		}
	}

	if len(mapMonsters) == 0 {
		log.Printf("地图%d没有配置怪物", mapID)
		return
	}

	// 获取地图配置信息（用于智能分布）
	mapConfig := getMapConfig(mapID)
	if mapConfig == nil {
		log.Printf("警告: 找不到地图%d配置，使用默认参数", mapID)
		defaultX, defaultY := 500, 500
		mapConfig = &common.MapBaseConfig{
			ID:      mapID,
			Width:   1000,
			Height:  1000,
			ReviveX: &defaultX,
			ReviveY: &defaultY,
		}
	}

	// 计算安全区域参数（处理指针类型）
	centerX := 500
	centerY := 500
	if mapConfig.ReviveX != nil {
		centerX = *mapConfig.ReviveX
	}
	if mapConfig.ReviveY != nil {
		centerY = *mapConfig.ReviveY
	}
	mapWidth := mapConfig.Width
	mapHeight := mapConfig.Height

	margin := int(float64(mapWidth) * 0.05)
	if margin < 20 {
		margin = 20
	}

	safeRadius := int(float64(mapWidth) * 0.08)
	if safeRadius < 30 {
		safeRadius = 30
	}

	spawnedCount := 0
	for _, monsterConfig := range mapMonsters {
		count := 3 + rand.Intn(6)
		if monsterConfig.Type == 1 {
			count = 1 + rand.Intn(2)
		} else if monsterConfig.Type == 2 {
			count = 1
		}

		var spawnRadius int
		switch {
		case monsterConfig.Level <= 2:
			spawnRadius = 80 + rand.Intn(120)
		case monsterConfig.Level <= 5:
			spawnRadius = 150 + rand.Intn(150)
		default:
			spawnRadius = 250 + rand.Intn(200)
		}

		maxRadius := int(math.Min(
			float64(mapWidth/2-margin),
			float64(mapHeight/2-margin),
		))
		if spawnRadius > maxRadius {
			spawnRadius = maxRadius
		}

		for i := 0; i < count; i++ {
			x, y := generatePositionAroundCenter(centerX, centerY, safeRadius, spawnRadius, margin, mapWidth, mapHeight)

			monster, err := globalService.SpawnMonster(monsterConfig.ID, mapID, x, y)
			if err != nil {
				log.Printf("生成怪物失败[%s]: %v", monsterConfig.Name, err)
				continue
			}

			if globalAIService != nil {
				globalAIService.RegisterMonster(monster, &monsterConfig)
			}

			spawnedCount++
		}
	}

	log.Printf("✅ 地图%d算法生成完成: 共%d个怪物 (配置了%d种)",
		mapID, spawnedCount, len(mapMonsters))
}

// getMapConfig 获取地图配置
func getMapConfig(mapID uint32) *common.MapBaseConfig {
	for _, m := range common.GameConfig.Maps {
		if m.ID == mapID {
			return &m
		}
	}
	return nil
}

// calculateGoldDrop 计算金币掉落（本地实现，避免循环依赖）
func calculateGoldDrop(minGold, maxGold int) int {
	if minGold >= maxGold {
		return minGold
	}
	return minGold + rand.Intn(maxGold-minGold+1)
}

// generatePositionAroundCenter 围绕中心点生成位置（避开中心保护区）
func generatePositionAroundCenter(centerX, centerY, minRadius, maxRadius, margin, mapWidth, mapHeight int) (int, int) {
	maxAttempts := 10 // 最大重试次数

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 随机角度和距离
		angle := rand.Float64() * 2 * math.Pi
		distance := float64(minRadius) + rand.Float64()*float64(maxRadius-minRadius)

		// 极坐标转直角坐标
		x := centerX + int(distance*math.Cos(angle))
		y := centerY + int(distance*math.Sin(angle))

		// 边界检查
		if x < margin {
			x = margin + rand.Intn(20)
		}
		if x > mapWidth-margin {
			x = mapWidth - margin - rand.Intn(20)
		}
		if y < margin {
			y = margin + rand.Intn(20)
		}
		if y > mapHeight-margin {
			y = mapHeight - margin - rand.Intn(20)
		}

		// 再次确认不在保护区内
		distToCenter := math.Sqrt(math.Pow(float64(x-centerX), 2) + math.Pow(float64(y-centerY), 2))
		if distToCenter >= float64(minRadius) {
			return x, y
		}
	}

	// 如果多次失败，使用安全的默认位置
	fallbackX := centerX + maxRadius/2 + rand.Intn(50)
	fallbackY := centerY + rand.Intn(100) - 50

	// 最终边界检查
	if fallbackX < margin {
		fallbackX = margin + 50
	}
	if fallbackX > mapWidth-margin {
		fallbackX = mapWidth - margin - 50
	}
	if fallbackY < margin {
		fallbackY = margin + 50
	}
	if fallbackY > mapHeight-margin {
		fallbackY = mapHeight - margin - 50
	}

	return fallbackX, fallbackY
}

// generatePositionWithRadius 在指定点周围radius范围内随机生成位置（用于配置文件的spawn_radius）
func generatePositionWithRadius(centerX, centerY, radius int) (int, int) {
	if radius <= 0 {
		return centerX, centerY
	}

	// 随机角度和距离（在圆内均匀分布）
	angle := rand.Float64() * 2 * math.Pi
	distance := rand.Float64() * float64(radius)

	x := centerX + int(distance*math.Cos(angle))
	y := centerY + int(distance*math.Sin(angle))

	return x, y
}

// GetService 获取全局怪物服务实例
func GetService() *Service {
	return globalService
}

// GetAIService 获取全局AI服务实例
func GetAIService() *AIService {
	return globalAIService
}

// ========== 接口实现（供Battle包调用）==========

// GetMonsterInfo 实现MonsterServiceInterface接口 - 获取怪物信息
func (s *Service) GetMonsterInfo(monsterID uint64) (*common.MonsterInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, exists := s.monsters[monsterID]
	if !exists {
		return nil, false
	}

	// 转换为标准化的MonsterInfo结构（避免暴露内部实现）
	info := &common.MonsterInfo{
		ID:        m.ID,
		BaseID:    m.BaseID,
		Name:      m.Name,
		Level:     m.Level,
		Type:      m.Type,
		MapID:     m.MapID,
		X:         m.X,
		Y:         m.Y,
		CurrentHP: m.CurrentHP,
		MaxHP:     m.MaxHP,
		Attack:    m.Attack,
		Defense:   m.Defense,
		Speed:     m.Speed,
		Status:    m.Status,
	}
	return info, true
}

// GetPlayerPositionFromMap 从地图服务获取玩家位置（需要外部注入）
var GetPlayerPositionFromMap func(uint64) (int, int, bool)

// SetPlayerPositionFunc 设置玩家位置获取函数
func SetPlayerPositionFunc(fn func(uint64) (int, int, bool)) {
	GetPlayerPositionFromMap = fn
}

// GMSpawnMonster GM命令：手动生成怪物（用于测试）
func GMSpawnMonster(baseID uint32, mapID uint32, x, y int, count int) ([]*MonsterInstance, error) {
	if globalService == nil {
		return nil, errors.New("怪物服务未初始化")
	}

	var monsters []*MonsterInstance
	for i := 0; i < count; i++ {
		// 如果指定坐标为0，则随机生成
		spawnX, spawnY := x, y
		if x == 0 || y == 0 {
			spawnX = rand.Intn(900) + 50
			spawnY = rand.Intn(900) + 50
		}

		monster, err := globalService.SpawnMonster(baseID, mapID, spawnX, spawnY)
		if err != nil {
			return monsters, err
		}

		// 注册到AI系统
		if globalAIService != nil {
			config := common.GetMonsterConfig(baseID)
			if config != nil {
				globalAIService.RegisterMonster(monster, config)
			}
		}

		monsters = append(monsters, monster)
		log.Printf("🎮 GM生成怪物: %s (实例ID=%d) 在位置(%d,%d)",
			monster.Name, monster.ID, spawnX, spawnY)
	}

	return monsters, nil
}
