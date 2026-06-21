package battle

import (
	"errors"
	common "game-server/Common"
	"log"
	"math/rand"
	"sync"
	"time"

	buff "game-server/GameService/Buff"
	skillService "game-server/GameService/Skill"
)

// SkillTypeToBuffId 技能类型→BUFF ID映射
// skills.json的buff_id全为0，按技能type语义映射DEBUFF：
//
//	type=2 外功 → 13 破甲（外功刚猛，削弱防御）
//	type=5 拳法 → 9 减速（拳法重击，减缓移动）
//	type=6 剑法 → 2 中毒（剑气伤人，持续流血）
//	type=7 刀法 → 13 破甲（刀势威猛，破甲）
//	type=8 枪法 → 9 减速（枪法穿透，减速）
//	type=9 斧法 → 7 眩晕（斧法力沉，概率眩晕）
//	type=1/3/4 内功/身法/护体 → 0 不附加（被动类）
var SkillTypeToBuffId = map[uint8]uint32{
	2: 13, // 外功→破甲
	5: 9,  // 拳法→减速
	6: 2,  // 剑法→中毒
	7: 13, // 刀法→破甲
	8: 9,  // 枪法→减速
	9: 7,  // 斧法→眩晕
}

// SkillBuffTriggerRate 技能附加BUFF的触发概率（按技能type）
var SkillBuffTriggerRate = map[uint8]float64{
	2: 0.30, // 外功30%
	5: 0.25, // 拳法25%
	6: 0.35, // 剑法35%
	7: 0.30, // 刀法30%
	8: 0.25, // 枪法25%
	9: 0.15, // 斧法15%（眩晕较强，概率低）
}

// ========== 接口定义（避免循环依赖）==========

// MonsterServiceInterface 怪物服务接口 - 由外部注入
type MonsterServiceInterface interface {
	GetMonsterInfo(monsterID uint64) (*common.MonsterInfo, bool)
	MonsterTakeDamage(instanceID uint64, damage int) (newHP int, isDead bool)
	MonsterDie(instanceID uint64) (exp int, gold int, drops []uint32, err error)
}

// AIServiceInterface AI服务接口 - 由外部注入
type AIServiceInterface interface {
	// OnMonsterHurted 通知AI系统怪物受伤
	OnMonsterHurted(monsterID uint64, attackerID uint64)
}

// 全局服务实例（通过依赖注入设置）
var globalMonsterSvc MonsterServiceInterface
var globalAIService AIServiceInterface

// SetMonsterService 注入怪物服务实例（在main.go中调用）
func SetMonsterService(svc MonsterServiceInterface) {
	globalMonsterSvc = svc
	log.Printf("✅ Battle系统: 怪物服务已注入")
}

// SetAIService 注入AI服务实例（在main.go中调用）
func SetAIService(svc AIServiceInterface) {
	globalAIService = svc
	log.Printf("✅ Battle系统: AI服务已注入")
}

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

// PVPAttackRequest PVP攻击请求
type PVPAttackRequest struct {
	AttackerID uint64 `json:"attacker_id" binding:"required"`
	TargetID   uint64 `json:"target_id" binding:"required"`
	SkillID    uint32 `json:"skill_id"` // 0=普通攻击
	X          int    `json:"x"`
	Y          int    `json:"y"`
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

	// 从BUFF管理器计算BUFF属性加成
	buffEffect := buff.GetManager().CalculateEffect(roleID)

	return &BaseFighter{
		ID:         roleInfo.ID,
		Attack:     roleInfo.Attack + skillBonus["attack"] + buffEffect.AttackChange,
		Defense:    roleInfo.Defense + skillBonus["defense"] + buffEffect.DefChange,
		Speed:      roleInfo.Speed + skillBonus["speed"] + buffEffect.SpeedChange,
		Hit:        roleInfo.Hit + skillBonus["hit"] + buffEffect.HitChange,
		Dodge:      roleInfo.Dodge + skillBonus["dodge"] + buffEffect.DodgeChange,
		Crit:       roleInfo.Crit + skillBonus["crit"] + buffEffect.CritChange,
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

	battleState, exists := s.battles[battleID]
	if !exists {
		return errors.New("战斗不存在")
	}

	battleState.EndBattle(winner)

	// 清理战斗者映射
	delete(s.battleByFighter, battleState.AttackerID)
	delete(s.battleByFighter, battleState.DefenderID)

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

	battleState, exists := s.battles[battleID]
	if !exists || battleState.Status != 0 {
		return 0
	}

	if battleState.AttackerID == fighterID {
		return battleState.DefenderID
	}
	return battleState.AttackerID
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

// ==================== 完整战斗系统：玩家攻击怪物 ====================

// PlayerAttackResult 玩家攻击结果
type PlayerAttackResult struct {
	Success    bool     `json:"success"`     // 攻击是否成功
	ErrorCode  int      `json:"error_code"`  // 错误代码: 0=成功, 1=距离过远, 2=冷却中, 3=目标不存在, 4=目标已死, 5=玩家死亡, 6=技能未学习, 7=MP不足, 8=武器不符, 9=沉默状态
	ErrorMsg   string   `json:"error_msg"`   // 错误信息
	AttackerID uint64   `json:"attacker_id"` // 攻击者ID
	TargetID   uint64   `json:"target_id"`   // 目标ID
	Damage     int      `json:"damage"`      // 伤害值
	IsCrit     bool     `json:"is_crit"`     // 是否暴击
	IsMiss     bool     `json:"is_miss"`     // 是否闪避
	IsBlocked  bool     `json:"is_blocked"`  // 是否格挡
	CurrentHP  int      `json:"current_hp"`  // 目标当前血量
	MaxHP      int      `json:"max_hp"`      // 目标最大血量
	IsDead     bool     `json:"is_dead"`     // 目标是否死亡
	ExpGain    int      `json:"exp_gain"`    // 获得经验
	GoldGain   int      `json:"gold_gain"`   // 获得金币
	Drops      []uint32 `json:"drops"`       // 掉落物品ID列表
	LeveledUp  bool     `json:"leveled_up"`  // 是否升级
	NewLevel   int      `json:"new_level"`   // 新等级（如果升级了）
	// 技能相关字段
	SkillID       uint32 `json:"skill_id"`        // 使用的技能ID（0=普通攻击）
	SkillName     string `json:"skill_name"`      // 技能名称
	IsSkillAttack bool   `json:"is_skill_attack"` // 是否技能攻击
	MpCost        int    `json:"mp_cost"`         // MP消耗
	CurrentMP     int    `json:"current_mp"`      // 玩家剩余MP
	MaxMP         int    `json:"max_mp"`          // 玩家最大MP
	BuffApplied   uint32 `json:"buff_applied"`    // 附加的BUFF ID（0=无）
}

// PlayerAttackMonster 玩家攻击怪物（完整流程）
func (s *Service) PlayerAttackMonster(roleID uint64, monsterInstanceID uint64, skillID uint32) *PlayerAttackResult {
	result := &PlayerAttackResult{
		AttackerID: roleID,
		TargetID:   monsterInstanceID,
	}

	// 1. 获取玩家属性
	playerFighter, err := s.getPlayerFighter(roleID)
	if err != nil {
		result.ErrorCode = 3
		result.ErrorMsg = "玩家数据获取失败"
		return result
	}

	// 1.1 检查玩家是否死亡（死亡玩家不能发起攻击）
	if playerFighter.CurrentHP <= 0 {
		result.ErrorCode = 5
		result.ErrorMsg = "您已死亡，无法攻击"
		result.Success = false
		return result
	}

	// 2. 获取怪物实例和属性（通过注入的接口）
	if globalMonsterSvc == nil {
		result.ErrorCode = 3
		result.ErrorMsg = "怪物服务未初始化"
		return result
	}
	monster, exists := globalMonsterSvc.GetMonsterInfo(monsterInstanceID)
	if !exists || monster.Status == 4 {
		result.ErrorCode = 3
		result.ErrorMsg = "目标不存在或已死亡"
		return result
	}

	// 3. 检查攻击距离（与前端BattleSystem.js的1.5格保持一致）
	playerPos := getPlayerPosition(roleID)
	inRange, distance := CheckAttackRange(playerPos.X, playerPos.Y, monster.X, monster.Y, 1.5) // 近战1.5格范围
	if !inRange {
		result.ErrorCode = 1
		result.ErrorMsg = "距离过远"
		result.Success = false
		return result
	}
	_ = distance

	// 4. 检查普通攻击冷却（默认1.5秒）
	attackCooldown := int64(1500)
	if GlobalCoolDownManager.IsCoolingDown(roleID, 0) { // skillID=0 表示普通攻击
		result.ErrorCode = 2
		result.ErrorMsg = "攻击冷却中"
		return result
	}
	GlobalCoolDownManager.SetCoolDown(roleID, 0, attackCooldown)

	// 5. 构建怪物战斗属性
	monsterFighter := &BaseFighter{
		ID:         monster.ID,
		Attack:     monster.Attack,
		Defense:    monster.Defense,
		Speed:      monster.Speed,
		Hit:        50, // 从配置读取，这里简化处理
		Dodge:      10,
		Crit:       5,
		CritDamage: 150,
		CurrentHP:  monster.CurrentHP,
		MaxHP:      monster.MaxHP,
	}

	// 6. 计算最终伤害（命中→格挡→暴击→伤害计算）
	// 6.1 技能处理：skillID > 0 时使用技能伤害
	var skillConfig *common.SkillBaseConfig
	var skillLevel uint32 = 1
	isSkillAttack := skillID > 0

	if isSkillAttack {
		// 检查沉默状态（沉默时无法使用技能）
		if buff.GetManager().IsSilenced(roleID) {
			result.ErrorCode = 9
			result.ErrorMsg = "沉默状态，无法使用技能"
			return result
		}

		// 获取技能配置（严格使用skills.json）
		skillConfig = common.GetSkillConfig(skillID)
		if skillConfig == nil {
			result.ErrorCode = 6
			result.ErrorMsg = "技能不存在"
			return result
		}

		// 校验玩家是否学习了该技能
		skillLevel = s.getRoleSkillLevel(roleID, skillID)
		if skillLevel == 0 {
			result.ErrorCode = 6
			result.ErrorMsg = "未学习该技能"
			return result
		}

		// 校验MP消耗（技能等级越高消耗越大，基础10 + 等级*2）
		mpCost := 10 + int(skillLevel)*2
		roleInfo, _ := common.DBRoleGet(roleID)
		if roleInfo != nil {
			// 校验武器类型限制（skills.json的weapon_type字段）
			// weapon_type=0表示徒手即可，其他值要求装备对应类型武器
			if skillConfig.WeaponType > 0 {
				playerWeaponType := uint8(0) // 默认徒手
				if roleInfo.WeaponID > 0 {
					if itemCfg := common.GetItemConfig(uint32(roleInfo.WeaponID)); itemCfg != nil {
						playerWeaponType = itemCfg.WeaponType
					}
				}
				if playerWeaponType != skillConfig.WeaponType {
					result.ErrorCode = 8
					result.ErrorMsg = "武器类型不符，无法施展该武学"
					return result
				}
			}

			if roleInfo.Mp < mpCost {
				result.ErrorCode = 7
				result.ErrorMsg = "内力不足"
				return result
			}
			// 扣除MP
			newMP, _ := common.DBRoleChangeMP(roleID, -mpCost)
			result.MpCost = mpCost
			result.CurrentMP = newMP
			result.MaxMP = roleInfo.MaxMp
		}

		result.SkillID = skillID
		result.SkillName = skillConfig.Name
		result.IsSkillAttack = true
	}

	// 6.2 计算伤害
	var damage int
	var isCrit, isMiss, isBlocked bool
	if isSkillAttack {
		// 技能伤害：使用CalculateSkillDamage（含技能系数和等级加成）
		// 先走命中/格挡判定
		if !CalculateHit(playerFighter, monsterFighter) {
			damage = 0
			isMiss = true
		} else {
			isBlocked, _ = CalculateBlock(playerFighter, monsterFighter)
			baseDmg := CalculateSkillDamage(playerFighter, monsterFighter, skillConfig.Type, skillLevel)
			// 技能也参与暴击判定（内联暴击计算，与CalculateDamage一致）
			critRate := playerFighter.Crit
			if rand.Float64()*100 < float64(critRate) {
				isCrit = true
				baseDmg = int(float64(baseDmg) * float64(playerFighter.CritDamage) / 100)
			}
			if isBlocked {
				baseDmg = baseDmg / 2
			}
			if baseDmg < 1 {
				baseDmg = 1
			}
			damage = baseDmg
		}
	} else {
		// 普通攻击
		skillBonus := 0
		damage, isCrit, isMiss, isBlocked, _ = CalculateFinalDamage(playerFighter, monsterFighter, skillBonus)
	}

	// 7. 填充结果
	result.Damage = damage
	result.IsCrit = isCrit
	result.IsMiss = isMiss
	result.IsBlocked = isBlocked

	if isMiss {
		// 闪避，不造成伤害
		result.Damage = 0
		result.CurrentHP = monster.CurrentHP
		result.MaxHP = monster.MaxHP
		result.Success = true
		// 通知怪物AI被攻击
		s.notifyMonsterHurted(monsterInstanceID, roleID)
		return result
	}

	// 8. 应用伤害到怪物
	newHP, isDead := globalMonsterSvc.MonsterTakeDamage(monsterInstanceID, damage)
	result.CurrentHP = newHP
	result.MaxHP = monster.MaxHP
	result.IsDead = isDead
	result.Success = true

	// 8.1 技能命中后附加BUFF（非闪避、非死亡时）
	if isSkillAttack && !isMiss && !isDead && skillConfig != nil {
		// 优先使用技能配置的buff_id，为0则按技能type映射
		buffID := skillConfig.BuffID
		if buffID == 0 {
			buffID = SkillTypeToBuffId[skillConfig.Type]
		}
		if buffID > 0 {
			// 按技能type的触发概率
			triggerRate, ok := SkillBuffTriggerRate[skillConfig.Type]
			if !ok {
				triggerRate = 0.25 // 默认25%
			}
			if rand.Float64() < triggerRate {
				buff.GetManager().AddBuff(monsterInstanceID, 2, buffID, roleID)
				result.BuffApplied = buffID
				log.Printf("✨ 技能[%s]命中怪物[%d]，附加BUFF[%d]",
					skillConfig.Name, monsterInstanceID, buffID)
			}
		}
	}

	// 9. 通知怪物AI被攻击（触发追击/反击）
	s.notifyMonsterHurted(monsterInstanceID, roleID)

	// 10. 如果怪物死亡，处理掉落
	if isDead {
		exp, gold, drops, err := globalMonsterSvc.MonsterDie(monsterInstanceID)
		if err == nil {
			result.ExpGain = exp
			result.GoldGain = gold
			result.Drops = drops

			// 给玩家增加经验和金币（返回是否升级）
			leveledUp, newLevel := s.rewardPlayer(roleID, exp, gold)

			if leveledUp {
				result.LeveledUp = true
				result.NewLevel = newLevel
			}
		}
	}

	return result
}

// notifyMonsterHurted 通知怪物被攻击（通过注入的AI接口）
func (s *Service) notifyMonsterHurted(monsterID uint64, attackerID uint64) {
	if globalAIService == nil {
		return // AI服务未注入，跳过通知
	}
	globalAIService.OnMonsterHurted(monsterID, attackerID)
}

// rewardPlayer 奖励玩家（经验+金币）
func (s *Service) rewardPlayer(roleID uint64, exp int, gold int) (leveledUp bool, newLevel int) {
	// 增加经验（返回是否升级）
	leveledUp, newLevel, _, err := common.DBRoleAddExp(roleID, int64(exp))
	if err != nil {
		log.Printf("增加经验失败: %v", err)
	}

	// 增加金币
	err = common.DBRoleAddGold(roleID, int64(gold))
	if err != nil {
		log.Printf("增加金币失败: %v", err)
	}

	if leveledUp {
		log.Printf("🎉 玩家 %d 升级到 %d 级! (+%d EXP, +%d 金币)", roleID, newLevel, exp, gold)
	}

	return leveledUp, newLevel
}

// MonsterAttackPlayer 怪物反击玩家
func (s *Service) MonsterAttackPlayer(monsterInstanceID uint64, playerID uint64) *common.MonsterAttackResult {
	result := &common.MonsterAttackResult{
		MonsterID: monsterInstanceID,
		TargetID:  playerID,
	}

	// 获取怪物属性（通过注入的接口）
	if globalMonsterSvc == nil {
		return result
	}
	monster, exists := globalMonsterSvc.GetMonsterInfo(monsterInstanceID)
	if !exists {
		return result
	}

	monsterFighter := &BaseFighter{
		ID:         monster.ID,
		Attack:     monster.Attack,
		Defense:    monster.Defense,
		Speed:      monster.Speed,
		Hit:        50,
		Dodge:      10,
		Crit:       5,
		CritDamage: 150,
		CurrentHP:  monster.CurrentHP,
		MaxHP:      monster.MaxHP,
	}

	// 获取玩家属性
	playerFighter, err := s.getPlayerFighter(playerID)
	if err != nil {
		return result
	}

	// 死亡玩家不被攻击（避免怪物继续攻击尸体）
	if playerFighter.CurrentHP <= 0 {
		result.IsDead = true
		result.PlayerHP = 0
		result.PlayerMaxHP = playerFighter.MaxHP
		return result
	}

	// 计算伤害
	damage, isCrit, isMiss, _, _ := CalculateFinalDamage(monsterFighter, playerFighter, 0)

	result.Damage = damage
	result.IsCrit = isCrit
	result.IsMiss = isMiss

	if isMiss {
		return result
	}

	// 应用伤害到玩家
	newHP, isDead := playerFighter.TakeDamage(damage, DamageTypePhysical)
	if err := common.DBRoleSetHP(playerID, newHP); err != nil {
		log.Printf("更新玩家HP失败: %v", err)
	}

	result.PlayerHP = newHP
	result.PlayerMaxHP = playerFighter.MaxHP
	result.IsDead = isDead

	return result
}

// ==================== 辅助函数 ====================

// PlayerPosition 玩家位置
type PlayerPosition struct {
	X int
	Y int
}

// getPlayerPosition 获取玩家位置（需要外部注入）
var getPlayerPosition func(uint64) *PlayerPosition

// SetPlayerPositionSetter 设置玩家位置获取函数
func SetPlayerPositionSetter(fn func(uint64) *PlayerPosition) {
	getPlayerPosition = fn
}

// ==================== PVP 战斗系统 ====================

// PVPAttackResult PVP攻击结果
type PVPAttackResult struct {
	Success     bool   `json:"success"`
	ErrorCode   int    `json:"error_code"` // 0=成功, 1=距离过远, 2=冷却中, 5=攻击者死亡, 6=技能未学习, 7=MP不足, 9=沉默, 10=安全区禁止PK, 11=目标死亡, 12=和平模式, 13=目标不存在
	ErrorMsg    string `json:"error_msg"`
	AttackerID  uint64 `json:"attacker_id"`
	TargetID    uint64 `json:"target_id"`
	Damage      int    `json:"damage"`
	IsCrit      bool   `json:"is_crit"`
	IsMiss      bool   `json:"is_miss"`
	IsBlocked   bool   `json:"is_blocked"`
	TargetHP    int    `json:"target_hp"`     // 目标当前HP
	TargetMaxHP int    `json:"target_max_hp"` // 目标最大HP
	IsDead      bool   `json:"is_dead"`       // 目标是否死亡
	// 技能相关
	SkillID       uint32 `json:"skill_id"`
	SkillName     string `json:"skill_name"`
	IsSkillAttack bool   `json:"is_skill_attack"`
	MpCost        int    `json:"mp_cost"`
	AttackerMP    int    `json:"attacker_mp"` // 攻击者剩余MP
	AttackerMaxMP int    `json:"attacker_max_mp"`
	BuffApplied   uint32 `json:"buff_applied"`
	// PVP专属
	ExpGain     int `json:"exp_gain"`      // 胜利者获得经验
	PkValueGain int `json:"pk_value_gain"` // 攻击者增加的PK值
}

// PlayerAttackPlayer 玩家攻击玩家（PVP完整流程）
func (s *Service) PlayerAttackPlayer(attackerID, targetID uint64, skillID uint32) *PVPAttackResult {
	result := &PVPAttackResult{
		AttackerID: attackerID,
		TargetID:   targetID,
	}

	// 1. 获取攻击者和目标属性
	attackerInfo, _ := common.DBRoleGet(attackerID)
	if attackerInfo == nil {
		result.ErrorCode = 13
		result.ErrorMsg = "攻击者不存在"
		return result
	}
	targetInfo, _ := common.DBRoleGet(targetID)
	if targetInfo == nil {
		result.ErrorCode = 13
		result.ErrorMsg = "目标不存在"
		return result
	}

	// 2. 安全区判定（pk_allowed=0 的地图禁止PVP）
	mapConfig := common.GetMapConfig(attackerInfo.MapID)
	if mapConfig == nil || mapConfig.PkAllowed == 0 {
		result.ErrorCode = 10
		result.ErrorMsg = "安全区禁止PK"
		return result
	}

	// 3. 攻击者死亡检查
	if attackerInfo.Hp <= 0 {
		result.ErrorCode = 5
		result.ErrorMsg = "您已死亡，无法攻击"
		return result
	}

	// 4. 目标死亡检查
	if targetInfo.Hp <= 0 {
		result.ErrorCode = 11
		result.ErrorMsg = "目标已死亡"
		return result
	}

	// 5. 和平模式检查（攻击者PkMode=0时不能攻击玩家）
	if attackerInfo.PkMode == 0 {
		result.ErrorCode = 12
		result.ErrorMsg = "和平模式，无法攻击玩家"
		return result
	}

	// 6. 距离检查
	attackerPos := getPlayerPosition(attackerID)
	targetPos := getPlayerPosition(targetID)
	if attackerPos == nil || targetPos == nil {
		result.ErrorCode = 13
		result.ErrorMsg = "位置信息获取失败"
		return result
	}
	inRange, _ := CheckAttackRange(attackerPos.X, attackerPos.Y, targetPos.X, targetPos.Y, 1.5)
	if !inRange {
		result.ErrorCode = 1
		result.ErrorMsg = "距离过远"
		return result
	}

	// 7. 获取战斗属性（含武学加成和BUFF加成）
	attackerFighter, err := s.getPlayerFighter(attackerID)
	if err != nil {
		result.ErrorCode = 13
		result.ErrorMsg = "攻击者数据获取失败"
		return result
	}
	targetFighter, err := s.getPlayerFighter(targetID)
	if err != nil {
		result.ErrorCode = 13
		result.ErrorMsg = "目标数据获取失败"
		return result
	}

	// 8. 技能处理
	var skillConfig *common.SkillBaseConfig
	var skillLevel uint32 = 1
	isSkillAttack := skillID > 0

	if isSkillAttack {
		// 沉默检查
		if buff.GetManager().IsSilenced(attackerID) {
			result.ErrorCode = 9
			result.ErrorMsg = "沉默状态，无法使用技能"
			return result
		}

		skillConfig = common.GetSkillConfig(skillID)
		if skillConfig == nil {
			result.ErrorCode = 6
			result.ErrorMsg = "技能不存在"
			return result
		}
		skillLevel = s.getRoleSkillLevel(attackerID, skillID)
		if skillLevel == 0 {
			result.ErrorCode = 6
			result.ErrorMsg = "未学习该技能"
			return result
		}

		// 校验武器类型限制（PVP同样适用）
		if skillConfig.WeaponType > 0 {
			playerWeaponType := uint8(0)
			if attackerInfo.WeaponID > 0 {
				if itemCfg := common.GetItemConfig(uint32(attackerInfo.WeaponID)); itemCfg != nil {
					playerWeaponType = itemCfg.WeaponType
				}
			}
			if playerWeaponType != skillConfig.WeaponType {
				result.ErrorCode = 8
				result.ErrorMsg = "武器类型不符，无法施展该武学"
				return result
			}
		}

		// MP消耗
		mpCost := 10 + int(skillLevel)*2
		if attackerInfo.Mp < mpCost {
			result.ErrorCode = 7
			result.ErrorMsg = "内力不足"
			return result
		}
		newMP, _ := common.DBRoleChangeMP(attackerID, -mpCost)
		result.MpCost = mpCost
		result.AttackerMP = newMP
		result.AttackerMaxMP = attackerInfo.MaxMp

		result.SkillID = skillID
		result.SkillName = skillConfig.Name
		result.IsSkillAttack = true
	}

	// 9. 计算伤害
	var damage int
	var isCrit, isMiss, isBlocked bool
	if isSkillAttack {
		if !CalculateHit(attackerFighter, targetFighter) {
			damage = 0
			isMiss = true
		} else {
			isBlocked, _ = CalculateBlock(attackerFighter, targetFighter)
			baseDmg := CalculateSkillDamage(attackerFighter, targetFighter, skillConfig.Type, skillLevel)
			critRate := attackerFighter.Crit
			if rand.Float64()*100 < float64(critRate) {
				isCrit = true
				baseDmg = int(float64(baseDmg) * float64(attackerFighter.CritDamage) / 100)
			}
			if isBlocked {
				baseDmg = baseDmg / 2
			}
			if baseDmg < 1 {
				baseDmg = 1
			}
			damage = baseDmg
		}
	} else {
		skillBonus := 0
		damage, isCrit, isMiss, isBlocked, _ = CalculateFinalDamage(attackerFighter, targetFighter, skillBonus)
	}

	result.Damage = damage
	result.IsCrit = isCrit
	result.IsMiss = isMiss
	result.IsBlocked = isBlocked

	// 无敌状态检查
	if !isMiss && buff.GetManager().IsInvincible(targetID) {
		damage = 0
		result.Damage = 0
		result.TargetHP = targetInfo.Hp
		result.TargetMaxHP = targetInfo.MaxHp
		result.Success = true
		return result
	}

	if isMiss {
		result.Damage = 0
		result.TargetHP = targetInfo.Hp
		result.TargetMaxHP = targetInfo.MaxHp
		result.Success = true
		return result
	}

	// 10. 应用伤害
	newHP := targetInfo.Hp - damage
	if newHP < 0 {
		newHP = 0
	}
	common.DBRoleSetHP(targetID, newHP)
	result.TargetHP = newHP
	result.TargetMaxHP = targetInfo.MaxHp
	result.Success = true

	// 11. 技能附加BUFF
	if isSkillAttack && !isMiss && skillConfig != nil {
		buffID := skillConfig.BuffID
		if buffID == 0 {
			buffID = SkillTypeToBuffId[skillConfig.Type]
		}
		if buffID > 0 {
			triggerRate, ok := SkillBuffTriggerRate[skillConfig.Type]
			if !ok {
				triggerRate = 0.25
			}
			if rand.Float64() < triggerRate {
				buff.GetManager().AddBuff(targetID, 1, buffID, attackerID)
				result.BuffApplied = buffID
			}
		}
	}

	// 12. 处理目标死亡
	if newHP <= 0 {
		result.IsDead = true
		// 设置目标死亡状态
		common.DBRoleSetStatus(targetID, 2)
		// 清除目标所有BUFF
		buff.GetManager().ClearAllBuffs(targetID)

		// 记录击杀和死亡
		common.DBRoleRecordKill(attackerID)
		common.DBRoleRecordDeath(targetID)

		// PVP经验奖励（使用CalculatePVPExp）
		expGain := CalculatePVPExp(attackerInfo.Level, targetInfo.Level)
		result.ExpGain = expGain

		// 给攻击者增加经验
		s.rewardPlayer(attackerID, expGain, 0)

		// PK值增加（主动攻击方增加PK值，红名惩罚）
		// 和平模式攻击者不增加（但和平模式已被前面拦截），队伍/帮派模式+10，全体模式+20
		pkGain := 10
		if attackerInfo.PkMode == 3 {
			pkGain = 20
		}
		common.DBRoleUpdatePKValue(attackerID, pkGain)
		result.PkValueGain = pkGain

		log.Printf("⚔️ PVP击杀: 玩家[%d]击杀玩家[%d], 经验+%d, PK值+%d",
			attackerID, targetID, expGain, pkGain)
	}

	return result
}
