package battle

import (
	"errors"
	common "game-server/Common"
	"log"
	"sync"
	"time"
)

// Service 战斗服务
type Service struct {
	battles         map[uint64]*BattleState // key=战斗ID
	battleByFighter map[uint64]uint64       // key=角色ID, value=战斗ID
	mu              sync.RWMutex
}

// NewService 创建战斗服务实例
func NewService() *Service {
	return &Service{
		battles:         make(map[uint64]*BattleState),
		battleByFighter: make(map[uint64]uint64),
	}
}

// AttackRequest 攻击请求
type AttackRequest struct {
	AttackerID   uint64 `json:"attacker_id" binding:"required"`
	AttackerType uint8  `json:"attacker_type" binding:"required"` // 1=玩家, 2=怪物
	TargetID     uint64 `json:"target_id" binding:"required"`
	TargetType   uint8  `json:"target_type" binding:"required"` // 1=玩家, 2=怪物
	SkillID      uint32 `json:"skill_id"`                       // 0=普通攻击
	X            int    `json:"x"`                              // 攻击时位置
	Y            int    `json:"y"`
}

// NormalAttack 普通攻击
func (s *Service) NormalAttack(req AttackRequest) (*AttackResult, error) {
	// 检查是否在冷却中(攻击间隔)
	// 简化处理,这里不做间隔检查

	// 获取攻击者和防御者属性
	attacker, err := s.getFighter(req.AttackerID, req.AttackerType)
	if err != nil {
		return nil, err
	}
	defender, err := s.getFighter(req.TargetID, req.TargetType)
	if err != nil {
		return nil, err
	}

	// 计算命中
	if !CalculateHit(attacker, defender) {
		// 闪避
		return &AttackResult{
			AttackerID:   req.AttackerID,
			AttackerType: req.AttackerType,
			TargetID:     req.TargetID,
			TargetType:   req.TargetType,
			Damage:       0,
			DamageType:   DamageTypePhysical,
			IsMiss:       true,
			AttackType:   AttackTypeNormal,
			Timestamp:    time.Now().UnixMilli(),
		}, nil
	}

	// 计算伤害
	damage, isCrit := CalculateDamage(attacker, defender, 0)

	// 应用伤害
	newHP, isDead := defender.TakeDamage(damage, DamageTypePhysical)

	// 更新防御者HP
	if err := s.updateFighterHP(req.TargetID, req.TargetType, newHP); err != nil {
		// log error but don't fail
	}

	result := &AttackResult{
		AttackerID:   req.AttackerID,
		AttackerType: req.AttackerType,
		TargetID:     req.TargetID,
		TargetType:   req.TargetType,
		Damage:       damage,
		DamageType:   DamageTypePhysical,
		IsCrit:       isCrit,
		IsMiss:       false,
		AttackType:   AttackTypeNormal,
		Timestamp:    time.Now().UnixMilli(),
	}

	// 检查是否死亡
	if isDead {
		result.Damage = newHP - defender.GetHP() + damage // 实际扣血
	}

	return result, nil
}

// SkillAttack 技能攻击
func (s *Service) SkillAttack(req AttackRequest, skillID uint32, skillType uint8) (*AttackResult, error) {
	// 检查技能冷却
	if GlobalCoolDownManager.IsCoolingDown(req.AttackerID, skillID) {
		return nil, errors.New("技能冷却中")
	}

	// 获取攻击者和防御者属性
	attacker, err := s.getFighter(req.AttackerID, req.AttackerType)
	if err != nil {
		return nil, err
	}
	defender, err := s.getFighter(req.TargetID, req.TargetType)
	if err != nil {
		return nil, err
	}

	// 获取技能信息
	var skillLevel uint32 = 1
	if req.AttackerType == 1 {
		// 玩家技能,获取技能等级
		skillLevel = s.getRoleSkillLevel(req.AttackerID, skillID)
	}

	// 计算命中
	if !CalculateHit(attacker, defender) {
		return &AttackResult{
			AttackerID:   req.AttackerID,
			AttackerType: req.AttackerType,
			TargetID:     req.TargetID,
			TargetType:   req.TargetType,
			Damage:       0,
			DamageType:   DamageTypePhysical,
			IsMiss:       true,
			SkillID:      skillID,
			AttackType:   AttackTypeSkill,
			Timestamp:    time.Now().UnixMilli(),
		}, nil
	}

	// 计算技能伤害
	damage := CalculateSkillDamage(attacker, defender, skillType, skillLevel)

	// 设置冷却(默认3秒)
	GlobalCoolDownManager.SetCoolDown(req.AttackerID, skillID, 3000)

	// 应用伤害
	newHP, _ := defender.TakeDamage(damage, DamageTypePhysical)

	// 更新防御者HP
	if err := s.updateFighterHP(req.TargetID, req.TargetType, newHP); err != nil {
		// log error
	}

	return &AttackResult{
		AttackerID:   req.AttackerID,
		AttackerType: req.AttackerType,
		TargetID:     req.TargetID,
		TargetType:   req.TargetType,
		Damage:       damage,
		DamageType:   DamageTypePhysical,
		IsCrit:       false,
		IsMiss:       false,
		SkillID:      skillID,
		AttackType:   AttackTypeSkill,
		Timestamp:    time.Now().UnixMilli(),
	}, nil
}

// getFighter 获取战斗者属性
func (s *Service) getFighter(id uint64, fighterType uint8) (*BaseFighter, error) {
	switch fighterType {
	case 1: // 玩家
		return s.getPlayerFighter(id)
	case 2: // 怪物
		return s.getMonsterFighter(id)
	default:
		return nil, errors.New("无效的战斗者类型")
	}
}

// getPlayerFighter 获取玩家战斗属性
func (s *Service) getPlayerFighter(roleID uint64) (*BaseFighter, error) {
	roleInfo, err := common.DBRoleGet(roleID)
	if err != nil || roleInfo == nil {
		return nil, errors.New("角色不存在")
	}

	// 从已装备武学计算加成
	skillSvc := skillService.NewService()
	skillBonus, err := skillSvc.CalculateSkillBonus(roleID)
	if err != nil {
		// 如果获取武学加成失败，使用空加成
		skillBonus = map[string]int{
			"hp": 0, "mp": 0, "attack": 0, "defense": 0,
			"speed": 0, "hit": 0, "dodge": 0, "crit": 0,
		}
	}

	return &BaseFighter{
		ID:         roleInfo.ID,
		Attack:     roleInfo.Attack + skillBonus["attack"],
		Defense:    roleInfo.Defense + skillBonus["defense"],
		Speed:      roleInfo.Speed + skillBonus["speed"],
		Hit:        roleInfo.Hit + skillBonus["hit"],
		Dodge:      roleInfo.Dodge + skillBonus["dodge"],
		Crit:       roleInfo.Crit + skillBonus["crit"],
		CritDamage: roleInfo.CritDamage,
		CurrentHP:  roleInfo.Hp + skillBonus["hp"],
		MaxHP:      roleInfo.MaxHp + skillBonus["hp"],
		SkillBonus: skillBonus,
	}, nil
}

// getMonsterFighter 获取怪物战斗属性
func (s *Service) getMonsterFighter(monsterID uint64) (*BaseFighter, error) {
	config := common.GetMonsterConfig(uint32(monsterID))
	if config == nil {
		return nil, errors.New("怪物不存在")
	}

	return &BaseFighter{
		ID:         uint64(config.ID),
		Attack:     config.Attack,
		Defense:    config.Defense,
		Speed:      config.Speed,
		Hit:        config.Hit,
		Dodge:      config.Dodge,
		Crit:       config.Crit,
		CritDamage: 150, // 怪物默认暴击伤害150%
		CurrentHP:  config.Hp,
		MaxHP:      config.Hp,
		SkillBonus: nil,
	}, nil
}

// updateFighterHP 更新战斗者HP
func (s *Service) updateFighterHP(id uint64, fighterType uint8, newHP int) error {
	switch fighterType {
	case 1: // 玩家
		return common.DBRoleSetHP(id, newHP)
	case 2: // 怪物
		// 怪物不持久化到数据库（只在内存中）
		return nil
	}
	return errors.New("无效的战斗者类型")
}

// getRoleSkillLevel 获取角色技能等级
func (s *Service) getRoleSkillLevel(roleID uint64, skillID uint32) uint32 {
	skills, err := common.DBSkillGetList(roleID)
	if err != nil {
		return 1
	}

	for _, skill := range skills {
		if sid, ok := skill["skill_id"].(float64); ok && uint32(sid) == skillID {
			if level, ok := skill["level"].(float64); ok {
				return uint32(level)
			}
		}
	}
	return 1
}

// StartPVP 开始PVP战斗
func (s *Service) StartPVP(attackerID, defenderID uint64) (*BattleState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已在战斗中
	if _, exists := s.battleByFighter[attackerID]; exists {
		return nil, errors.New("已在战斗中")
	}
	if _, exists := s.battleByFighter[defenderID]; exists {
		return nil, errors.New("对手已在战斗中")
	}

	// 创建战斗
	battleID := uint64(time.Now().UnixNano())
	battle := NewBattleState(battleID, attackerID, defenderID, 1) // 1=PVP

	s.battles[battleID] = battle
	s.battleByFighter[attackerID] = battleID
	s.battleByFighter[defenderID] = battleID

	return battle, nil
}

// EndBattle 结束战斗
func (s *Service) EndBattle(battleID uint64, winner uint8) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	battle, exists := s.battles[battleID]
	if !exists {
		return errors.New("战斗不存在")
	}

	battle.EndBattle(winner)

	// 清理战斗者映射
	delete(s.battleByFighter, battle.AttackerID)
	delete(s.battleByFighter, battle.DefenderID)

	return nil
}

// GetBattle 获取战斗状态
func (s *Service) GetBattle(fighterID uint64) (*BattleState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	battleID, exists := s.battleByFighter[fighterID]
	if !exists {
		return nil, false
	}

	battle, exists := s.battles[battleID]
	return battle, exists
}

// IsInBattle 检查是否在战斗中
func (s *Service) IsInBattle(fighterID uint64) bool {
	_, exists := s.GetBattle(fighterID)
	return exists
}

// GetBattleOpponent 获取战斗对手
func (s *Service) GetBattleOpponent(fighterID uint64) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	battleID, exists := s.battleByFighter[fighterID]
	if !exists {
		return 0
	}

	battle, exists := s.battles[battleID]
	if !exists || battle.Status != 0 {
		return 0
	}

	if battle.AttackerID == fighterID {
		return battle.DefenderID
	}
	return battle.AttackerID
}

// RecordKill 记录击杀
func (s *Service) RecordKill(killerID uint64, killerType uint8, victimID uint64, victimType uint8) error {
	// 根据击杀类型记录
	if killerType == 1 {
		// 玩家击杀,增加杀人数和PK值
		if err := common.DBRoleRecordKill(killerID); err != nil {
			log.Printf("记录击杀失败: %v", err)
		}
	}

	// 记录死亡
	if victimType == 1 {
		if err := common.DBRoleRecordDeath(victimID); err != nil {
			log.Printf("记录死亡失败: %v", err)
		}
	}

	return nil
}

// GetSkillCoolDown 获取技能冷却时间
func (s *Service) GetSkillCoolDown(roleID uint64, skillID uint32) int64 {
	return GlobalCoolDownManager.GetCoolDownRemaining(roleID, skillID)
}
