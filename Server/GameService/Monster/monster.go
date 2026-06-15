package monster

import (
	"errors"
	"game-server/DBService/mysql"
	"game-server/GameService/Battle"
	"game-server/GameService/Map/model"
	"math/rand"
	"sync"
	"time"
)

// MonsterInstance 怪物实例
type MonsterInstance struct {
	ID        uint64    // 实例ID
	BaseID    uint32    // 基础ID
	Name      string    // 名称
	Level     uint32    // 等级
	Type      uint8     // 类型: 0=普通, 1=精英, 2=BOSS
	MapID     uint32    // 地图ID
	X         int       // X坐标
	Y         int       // Y坐标
	CurrentHP int       // 当前HP
	MaxHP     int       // 最大HP
	Attack    int       // 攻击力
	Defense   int       // 防御力
	Speed     int       // 速度
	Status    uint8     // 状态: 0=空闲, 1=巡逻, 2=追击, 3=战斗, 4=死亡
	TargetID  uint64    // 攻击目标
	RespawnAt int64     // 复活时间戳
}

// NPCInstance NPC实例
type NPCInstance struct {
	ID       uint64
	BaseID   uint32
	Name     string
	Type     uint8     // 类型: 1=普通NPC, 2=商店NPC, 3=任务NPC, 4=仓库NPC, 5=传送NPC
	MapID    uint32
	X        int
	Y        int
	Facing   uint8     // 朝向
	Dialog   string    // 对话文本
	ShopID   *uint32   // 商店ID(商店NPC)
}

// Service 怪物/NPC服务
type Service struct {
	monsters   map[uint64]*MonsterInstance // key=实例ID
	npcs       map[uint64]*NPCInstance     // key=实例ID
	monsterID  uint64                      // 自增ID
	npcID      uint64                      // 自增ID
	mu         sync.RWMutex
	battleSvc  *Battle.Service
}

// NewService 创建服务
func NewService(battleSvc *Battle.Service) *Service {
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
	var base model.MonsterBase
	if err := mysql.DB.Where("id = ? AND map_id = ?", baseID, mapID).First(&base).Error; err != nil {
		return nil, errors.New("怪物不存在")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.monsterID++
	monster := &MonsterInstance{
		ID:        s.monsterID,
		BaseID:    base.ID,
		Name:      base.Name,
		Level:     base.Level,
		Type:      base.Type,
		MapID:     mapID,
		X:         x,
		Y:         y,
		CurrentHP: base.Hp,
		MaxHP:     base.Hp,
		Attack:    base.Attack,
		Defense:   base.Defense,
		Speed:     base.Speed,
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
	var base model.MonsterBase
	if err := mysql.DB.Where("id = ?", monster.BaseID).First(&base).Error; err != nil {
		return 0, 0, nil, err
	}

	// 计算经验
	exp = base.Exp

	// 计算金币
	gold = Battle.CalculateGoldDrop(base.GoldMin, base.GoldMax)

	// 计算掉落
	if base.DropGroupID != nil {
		drops, err = s.RollDrops(*base.DropGroupID)
		if err != nil {
			drops = nil
		}
	}

	// 设置复活时间
	s.mu.Lock()
	monster.Status = 4
	monster.RespawnAt = time.Now().Unix() + int64(base.RespawnTime)
	s.mu.Unlock()

	return exp, gold, drops, nil
}

// RollDrops 掉落roll
func (s *Service) RollDrops(groupID uint32) ([]uint32, error) {
	var drops []model.DropGroup
	if err := mysql.DB.Where("monster_id = ?", groupID).Find(&drops).Error; err != nil {
		return nil, err
	}

	var result []uint32
	for _, drop := range drops {
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
	var base model.MonsterBase
	if err := mysql.DB.Where("id = ?", monster.BaseID).First(&base).Error; err != nil {
		return err
	}

	monster.CurrentHP = base.Hp
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
	var base model.NPCBase
	if err := mysql.DB.Where("id = ? AND map_id = ?", baseID, mapID).First(&base).Error; err != nil {
		return nil, errors.New("NPC不存在")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.npcID++
	npc := &NPCInstance{
		ID:     s.npcID,
		BaseID: base.ID,
		Name:   base.Name,
		Type:   base.Type,
		MapID:  mapID,
		X:      x,
		Y:      y,
		Facing: base.Face,
		Dialog: base.DialogText,
		ShopID: base.ShopID,
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
