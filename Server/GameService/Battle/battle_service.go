package battle

import (
	"errors"
	common "game-server/Common"
	"log"
	"sync"
	"time"

	skillService "game-server/GameService/Skill"
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
func (s *Service) GetCoolDownRemaining(roleID uint64, skillID uint32) int64 {
	return GlobalCoolDownManager.GetCoolDownRemaining(roleID, skillID)
}

// DamageResult 伤害结果
type DamageResult struct {
	TargetID   uint64 `json:"target_id"`
	Damage     int    `json:"damage"`
	CurrentHP  int    `json:"current_hp"`
	MaxHP      int    `json:"max_hp"`
	IsCritical bool   `json:"is_critical"`
	IsBlocked  bool   `json:"is_blocked"`
	IsDodged   bool   `json:"is_dodged"`
	IsDead     bool   `json:"is_dead"`
}

// ProcessDamage 处理伤害
func (s *Service) ProcessDamage(req DamageRequest) *DamageResult {
	result := &DamageResult{
		TargetID:   req.TargetID,
		Damage:     req.Damage,
		IsCritical: req.IsCritical,
		IsBlocked:  req.IsBlocked,
		IsDodged:   req.IsDodged,
	}

	// 如果是闪避，不造成伤害
	if req.IsDodged {
		result.Damage = 0
		return result
	}

	// 更新目标血量
	currentHP, err := common.DBRoleChangeHP(req.TargetID, -req.Damage)
	if err != nil {
		log.Printf("更新角色HP失败: %v", err)
	}
	result.CurrentHP = currentHP
	result.IsDead = currentHP <= 0

	// 获取MaxHP
	if roleInfo, err := common.DBRoleGet(req.TargetID); err == nil && roleInfo != nil {
		result.MaxHP = roleInfo.MaxHp
	}

	return result
}

// DeathResult 死亡结果
type DeathResult struct {
	TargetID uint64 `json:"target_id"`
	KillerID uint64 `json:"killer_id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

// ProcessDeath 处理死亡
func (s *Service) ProcessDeath(req DeathRequest) *DeathResult {
	result := &DeathResult{
		TargetID: req.TargetID,
		KillerID: req.KillerID,
	}

	// 记录击杀
	s.RecordKill(req.KillerID, 1, req.TargetID, 1)

	return result
}

// RespawnResult 复活结果
type RespawnResult struct {
	RoleID uint64 `json:"role_id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	HP     int    `json:"hp"`
	MP     int    `json:"mp"`
}

// ProcessRespawn 处理复活
func (s *Service) ProcessRespawn(req RespawnRequest) *RespawnResult {
	result := &RespawnResult{
		RoleID: req.RoleID,
	}

	// 根据复活类型处理
	switch req.Type {
	case "here":
		// 原地复活，消耗一定资源
		hp, _ := common.DBRoleChangeHP(req.RoleID, 100)
		mp, _ := common.DBRoleChangeMP(req.RoleID, 100)
		result.HP = hp
		result.MP = mp
		// 位置不变
		if role, err := common.DBRoleGet(req.RoleID); err == nil && role != nil {
			result.X = role.MapX
			result.Y = role.MapY
		}

	case "town":
		// 回城复活，满血满蓝，回到主城
		common.DBRoleFullRecovery(req.RoleID)
		common.DBRoleChangeMap(req.RoleID, 1, 100, 100) // 主城地图ID为1，复活点坐标
		if role, err := common.DBRoleGet(req.RoleID); err == nil && role != nil {
			result.X = role.MapX
			result.Y = role.MapY
		}
	}

	return result
}

// LevelUpResult 升级结果
type LevelUpResult struct {
	RoleID  uint64 `json:"role_id"`
	Level   int    `json:"level"`
	MaxHP   int    `json:"max_hp"`
	MaxMP   int    `json:"max_mp"`
	Attack  int    `json:"attack"`
	Defense int    `json:"defense"`
	Speed   int    `json:"speed"`
}

// ProcessLevelUp 处理升级
func (s *Service) ProcessLevelUp(req LevelUpRequest) *LevelUpResult {
	result := &LevelUpResult{
		RoleID: req.RoleID,
		Level:  req.Level,
	}

	// 获取更新后的角色属性
	role, _ := common.DBRoleGet(req.RoleID)
	if role != nil {
		result.MaxHP = role.MaxHp
		result.MaxMP = role.MaxMp
		result.Attack = role.Attack
		result.Defense = role.Defense
		result.Speed = role.Speed
	}

	return result
}

// BuffResult 增益结果
type BuffResult struct {
	TargetID    uint64 `json:"target_id"`
	BuffType    string `json:"buff_type"`
	Duration    int    `json:"duration"`
	EffectValue int    `json:"effect_value"`
}

// ProcessBuff 处理增益
func (s *Service) ProcessBuff(req BuffRequest) *BuffResult {
	result := &BuffResult{
		TargetID: req.TargetID,
		BuffType: req.BuffType,
	}

	// 根据增益类型设置持续时间和效果值
	switch req.BuffType {
	case "attack":
		result.Duration = 10000 // 10秒
		result.EffectValue = 20 // 攻击+20
	case "defense":
		result.Duration = 10000
		result.EffectValue = 20 // 防御+20
	case "speed":
		result.Duration = 8000
		result.EffectValue = 30 // 速度+30
	case "heal":
		result.Duration = 10000
		result.EffectValue = 5 // 每秒恢复5点
	}

	return result
}

// DeBuffResult 减益结果
type DeBuffResult struct {
	TargetID    uint64 `json:"target_id"`
	DeBuffType  string `json:"debuff_type"`
	Duration    int    `json:"duration"`
	EffectValue int    `json:"effect_value"`
}

// ProcessDeBuff 处理减益
func (s *Service) ProcessDeBuff(req DeBuffRequest) *DeBuffResult {
	result := &DeBuffResult{
		TargetID:   req.TargetID,
		DeBuffType: req.DeBuffType,
	}

	// 根据减益类型设置持续时间和效果值
	switch req.DeBuffType {
	case "poison":
		result.Duration = 5000 // 5秒
		result.EffectValue = 3 // 每秒3点伤害
	case "burn":
		result.Duration = 3000
		result.EffectValue = 5
	case "freeze":
		result.Duration = 2000
		result.EffectValue = 50 // 减速50%
	case "stun":
		result.Duration = 1500
		result.EffectValue = 0
	case "bleed":
		result.Duration = 4000
		result.EffectValue = 2
	case "silence":
		result.Duration = 3000
		result.EffectValue = 0
	case "fear":
		result.Duration = 2000
		result.EffectValue = 0
	}

	return result
}

// MapEventResult 地图事件结果
type MapEventResult struct {
	EventType string `json:"event_type"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
}

// ProcessMapEvent 处理地图事件
func (s *Service) ProcessMapEvent(req MapEventRequest) *MapEventResult {
	result := &MapEventResult{
		EventType: req.EventType,
		X:         req.X,
		Y:         req.Y,
	}

	return result
}

// DamageRequest 伤害请求（用于handler）
type DamageRequest struct {
	TargetID   uint64 `json:"target_id"`
	AttackerID uint64 `json:"attacker_id"`
	Damage     int    `json:"damage"`
	IsCritical bool   `json:"is_critical"`
	IsBlocked  bool   `json:"is_blocked"`
	IsDodged   bool   `json:"is_dodged"`
}

// DeathRequest 死亡请求（用于handler）
type DeathRequest struct {
	TargetID uint64 `json:"target_id"`
	KillerID uint64 `json:"killer_id"`
}

// RespawnRequest 复活请求（用于handler）
type RespawnRequest struct {
	RoleID uint64 `json:"role_id"`
	Type   string `json:"type"`
}

// LevelUpRequest 升级请求（用于handler）
type LevelUpRequest struct {
	RoleID uint64 `json:"role_id"`
	Level  int    `json:"level"`
}

// BuffRequest 增益请求（用于handler）
type BuffRequest struct {
	TargetID uint64 `json:"target_id"`
	BuffType string `json:"buff_type"`
}

// DeBuffRequest 减益请求（用于handler）
type DeBuffRequest struct {
	TargetID   uint64 `json:"target_id"`
	DeBuffType string `json:"debuff_type"`
}

// MapEventRequest 地图事件请求（用于handler）
type MapEventRequest struct {
	EventType string `json:"event_type"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	MapID     uint32 `json:"map_id"`
}
