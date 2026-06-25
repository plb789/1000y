package battle

import (
	"errors"
	"fmt"
	common "game-server/Common"
	"log"
	"math/rand"
	"sync"
	"time"

	buff "game-server/GameService/Buff"
	item "game-server/GameService/Item"
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

// SkillSelfBuffId 技能类型→自身BUFF ID映射（使用技能时100%给自身附加增益BUFF）
// 内功(type=1) → 力量祝福(4): 攻击+20,暴击+5
// 护体(type=4) → 防御强化(5): 防御+30,减伤15%
// 身法(type=3) → 疾风术(6): 速度+15,闪避+10
var SkillSelfBuffId = map[uint8]uint32{
	1: 4, // 内功→力量祝福
	3: 6, // 身法→疾风术
	4: 5, // 护体→防御强化
}

// ========== 接口定义（避免循环依赖）==========

// MonsterServiceInterface 怪物服务接口 - 由外部注入
type MonsterServiceInterface interface {
	GetMonsterInfo(monsterID uint64) (*common.MonsterInfo, bool)
	MonsterTakeDamage(instanceID uint64, damage int) (newHP int, isDead bool)
	MonsterDie(instanceID uint64) (exp int, gold int, drops []uint32, err error)
	GetMonstersInArea(centerX, centerY int, radius int, excludeID uint64) []common.AOETargetInfo // 获取范围内怪物
}

// AIServiceInterface AI服务接口 - 由外部注入
type AIServiceInterface interface {
	// OnMonsterHurted 通知AI系统怪物受伤
	OnMonsterHurted(monsterID uint64, attackerID uint64)
	// OnMonsterDeath 通知AI系统怪物死亡，触发复活倒计时
	OnMonsterDeath(monsterID uint64)
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
		log.Printf("getPlayerFighter: 获取角色数据失败 roleID=%d, err=%v", roleID, err)
		return nil, fmt.Errorf("角色数据获取失败: %v", err)
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
			result.HP = role.Hp
			result.MP = role.Mp
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
	Success    bool                     `json:"success"`     // 攻击是否成功
	ErrorCode  int                      `json:"error_code"`  // 错误代码: 0=成功, 1=距离过远, 2=冷却中, 3=目标不存在, 4=目标已死, 5=玩家死亡, 6=技能未学习, 7=MP不足, 8=武器不符, 9=沉默状态
	ErrorMsg   string                   `json:"error_msg"`   // 错误信息
	AttackerID uint64                   `json:"attacker_id"` // 攻击者ID
	TargetID   uint64                   `json:"target_id"`   // 目标ID
	Damage     int                      `json:"damage"`      // 伤害值
	IsCrit     bool                     `json:"is_crit"`     // 是否暴击
	IsMiss     bool                     `json:"is_miss"`     // 是否闪避
	IsBlocked  bool                     `json:"is_blocked"`  // 是否格挡
	CurrentHP  int                      `json:"current_hp"`  // 目标当前血量
	MaxHP      int                      `json:"max_hp"`      // 目标最大血量
	IsDead     bool                     `json:"is_dead"`     // 目标是否死亡
	ExpGain    int                      `json:"exp_gain"`    // 获得经验
	GoldGain   int                      `json:"gold_gain"`   // 获得金币
	Drops      []uint32                 `json:"drops"`       // 掉落物品ID列表
	DropItems  []map[string]interface{} `json:"drop_items"`  // 实际放入背包的物品详情
	LeveledUp  bool                     `json:"leveled_up"`  // 是否升级
	NewLevel   int                      `json:"new_level"`   // 新等级（如果升级了）
	// 技能相关字段
	SkillID           uint32 `json:"skill_id"`            // 使用的技能ID（0=普通攻击）
	SkillName         string `json:"skill_name"`          // 技能名称
	SkillType         uint8  `json:"skill_type"`          // 技能类型(1-9)
	IsSkillAttack     bool   `json:"is_skill_attack"`     // 是否技能攻击
	MpCost            int    `json:"mp_cost"`             // MP消耗
	CurrentMP         int    `json:"current_mp"`          // 玩家剩余MP
	MaxMP             int    `json:"max_mp"`              // 玩家最大MP
	BuffApplied       uint32 `json:"buff_applied"`        // 附加的BUFF ID（0=无）
	SelfBuffApplied   uint32 `json:"self_buff_applied"`   // 自身附加的BUFF ID（0=无）
	Invulnerable      bool   `json:"invulnerable"`        // 目标是否无敌免疫
	ReflectDamage     int    `json:"reflect_damage"`      // 反弹伤害值
	ReflectTargetName string `json:"reflect_target_name"` // 反弹目标名
	LifestealHeal     int    `json:"lifesteal_heal"`      // 吸血回复量
	// AOE多目标
	AOETargets []map[string]interface{} `json:"aoe_targets"` // AOE波及的目标 [{target_id, name, damage, is_dead}, ...]
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

	// 1.0 获取玩家完整信息（用于返回MP等）
	roleInfo, _ := common.DBRoleGet(roleID)

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

	// 3. 检查攻击距离（技能有独立范围配置，普通攻击默认1.5格）
	attackRange := 1.5 // 普通攻击近战范围
	if skillID > 0 {
		// 预读技能配置获取范围（仅用于距离校验，不消耗MP等）
		if tmpCfg := common.GetSkillConfig(skillID); tmpCfg != nil && tmpCfg.Range > 1 {
			attackRange = float64(tmpCfg.Range)
		}
	}
	playerPos := getPlayerPosition(roleID)
	inRange, distance := CheckAttackRange(playerPos.X, playerPos.Y, monster.X, monster.Y, attackRange)
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

		// 校验冷却时间（cooldown=0表示无冷却，走攻速机制由前端控制）
		if skillConfig.Cooldown > 0 {
			if GlobalCoolDownManager.IsCoolingDown(roleID, skillID) {
				result.ErrorCode = 2
				result.ErrorMsg = "技能冷却中"
				return result
			}
		}

		// 校验武器类型限制（skills.json的weapon_type字段）
		if roleInfo != nil && skillConfig != nil && skillConfig.WeaponType > 0 {
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

		// 使用技能配置的MP消耗（基础值 + 等级加成）
		mpCost := skillConfig.MpCost + int(skillLevel)*2
		if roleInfo != nil {
			if roleInfo.Mp < mpCost {
				result.ErrorCode = 7
				result.ErrorMsg = "内力不足"
				return result
			}
			newMP, _ := common.DBRoleChangeMP(roleID, -mpCost)
			result.MpCost = mpCost
			result.CurrentMP = newMP
			result.MaxMP = roleInfo.MaxMp
		}

		// 设置冷却（cooldown>0时服务端记录固定冷却；cooldown=0时不设置，由前端攻速控制）
		if skillConfig.Cooldown > 0 {
			GlobalCoolDownManager.SetCoolDown(roleID, skillID, int64(skillConfig.Cooldown)*1000)
		}

		result.SkillID = skillID
		result.SkillName = skillConfig.Name
		result.SkillType = skillConfig.Type
		result.IsSkillAttack = true

		// 6.1.1 自身BUFF：内功/身法/护体技能使用时自动给自身附加增益BUFF
		if selfBuffID, ok := SkillSelfBuffId[skillConfig.Type]; ok {
			appliedBuff := buff.GetManager().AddBuff(roleID, 1, selfBuffID, roleID)
			if appliedBuff != nil {
				result.SelfBuffApplied = selfBuffID
				buffConfig := common.GetBuffConfig(selfBuffID)
				buffName := "未知"
				if buffConfig != nil {
					buffName = buffConfig.Name
				}
				log.Printf("✨ 技能[%s]释放，自身附加BUFF[%d:%s]", skillConfig.Name, selfBuffID, buffName)
			}
		}
	}

	// 6.2 计算伤害
	var damage int
	var isCrit, isMiss, isBlocked bool
	if isSkillAttack {
		// 技能伤害：以skills.json的damage为基数，结合等级/攻击力/暴击计算
		if !CalculateHit(playerFighter, monsterFighter) {
			damage = 0
			isMiss = true
		} else {
			isBlocked, blockReduction := CalculateBlock(playerFighter, monsterFighter)

			// 基础伤害 = 技能配置damage + 攻击力加成(比例) + 等级成长
			// 公式：baseDamage = skill.Damage + (playerAttack * 0.3) + (level-1) * skill.Damage * 0.05
			skillBaseDmg := skillConfig.Damage
			attackRatio := int(float64(playerFighter.Attack) * 0.3)                  // 攻击力30%转化
			levelGrowth := int(float64(skillBaseDmg) * float64(skillLevel-1) * 0.05) // 每级+5%基础伤害

			baseDmg := skillBaseDmg + attackRatio + levelGrowth
			if baseDmg < 1 {
				baseDmg = 1
			}

			// 暴击判定
			critRate := playerFighter.Crit
			if rand.Float64()*100 < float64(critRate) {
				isCrit = true
				baseDmg = int(float64(baseDmg) * float64(playerFighter.CritDamage) / 100)
			}

			// 格挡减伤
			if isBlocked {
				baseDmg = int(float64(baseDmg) * (1 - blockReduction))
			}

			// 伤害波动 ±10%
			variance := float64(baseDmg) * 0.1
			baseDmg = int(float64(baseDmg) + (rand.Float64()*2-1)*variance)
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
		// 返回玩家当前MP（避免客户端误用零值）
		if roleInfo != nil {
			result.CurrentMP = roleInfo.Mp
			result.MaxMP = roleInfo.MaxMp
		}
		// 通知怪物AI被攻击
		s.notifyMonsterHurted(monsterInstanceID, roleID)
		return result
	}

	// 8. 应用伤害到怪物（集成BUFF效果：无敌/减伤/反弹/吸血）

	// 8.0 无敌判定（减伤>=100%则完全免疫）
	if buff.GetManager().IsFullyInvincible(monsterInstanceID) {
		damage = 0
		result.Damage = 0
		result.IsMiss = false // 不是闪避，是免疫
		result.CurrentHP = monster.CurrentHP
		result.MaxHP = monster.MaxHP
		result.Success = true
		result.Invulnerable = true // 新增字段标记无敌状态
		// 返回玩家当前MP（避免客户端误用零值）
		if roleInfo != nil {
			result.CurrentMP = roleInfo.Mp
			result.MaxMP = roleInfo.MaxMp
		}
		log.Printf("🛡️ 怪物[%d]处于无敌状态，免疫伤害", monsterInstanceID)
		return result
	}

	// 8.1 BUFF减伤百分比计算（负值=易伤，如破甲-20%则伤害×1.2）
	reductionPct := buff.GetManager().GetTotalDamageReduction(monsterInstanceID)
	if reductionPct != 0 {
		reductionFactor := 1.0 - float64(reductionPct)/100.0
		damage = int(float64(damage) * reductionFactor)
		if damage < 0 {
			damage = 0 // 减伤不会造成负伤害
		}
		if reductionPct > 0 {
			log.Printf("🛡️ 怪物[%d]BUFF减伤%d%%，实际伤害=%d", monsterInstanceID, reductionPct, damage)
		} else if reductionPct < 0 {
			log.Printf("⚠️ 怪物[%d]易伤%d%%，实际伤害=%d", monsterInstanceID, -reductionPct, damage)
		}
	}

	newHP, isDead := globalMonsterSvc.MonsterTakeDamage(monsterInstanceID, damage)
	result.CurrentHP = newHP
	result.MaxHP = monster.MaxHP
	result.IsDead = isDead
	result.Success = true

	// 普通攻击也返回玩家当前MP（避免客户端误用零值）
	if roleInfo != nil {
		result.CurrentMP = roleInfo.Mp
		result.MaxMP = roleInfo.MaxMp
	}

	// 8.2 反弹伤害（目标身上有反弹类BUFF时，将部分伤害反弹给攻击者）
	reflectPct := buff.GetManager().GetTotalReflectPct(monsterInstanceID)
	if reflectPct > 0 && damage > 0 {
		reflectDmg := int(float64(damage) * float64(reflectPct) / 100.0)
		if reflectDmg > 0 {
			// 反弹伤害给玩家（通过扣血实现）
			roleInfo, err := common.DBRoleGet(roleID)
			if err == nil && roleInfo != nil {
				newPlayerHP := roleInfo.Hp - reflectDmg
				if newPlayerHP < 0 {
					newPlayerHP = 0
				}
				common.DBRoleSetHP(roleID, newPlayerHP)
				result.ReflectDamage = reflectDmg
				result.ReflectTargetName = "自身"
				log.Printf("💥 反弹! 怪物[%d]反弹%d%%伤害=%d给玩家", monsterInstanceID, reflectPct, reflectDmg)
			}
		}
	}

	// 8.3 吸血效果（攻击者身上有吸血类BUFF时，按伤害比例回血）
	lifestealPct := buff.GetManager().GetTotalLifestealPct(roleID)
	if lifestealPct > 0 && damage > 0 {
		healAmount := int(float64(damage) * float64(lifestealPct) / 100.0)
		if healAmount > 0 {
			roleInfo, err := common.DBRoleGet(roleID)
			if err == nil && roleInfo != nil {
				newPlayerHP := roleInfo.Hp + healAmount
				maxHP := int(roleInfo.MaxHp)
				if newPlayerHP > maxHP {
					newPlayerHP = maxHP
				}
				common.DBRoleSetHP(roleID, newPlayerHP)
				result.LifestealHeal = healAmount
				log.Printf("🩸 吸血! 玩家回复%dHP(伤害%d×%d%%)", healAmount, damage, lifestealPct)
			}
		}
	}

	// 8.4 技能命中后附加BUFF（非闪避、非死亡时）
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

	// 8.2 AOE范围伤害（技能配置aoe_radius > 0时，对目标周围怪物造成溅射伤害）
	if isSkillAttack && !isMiss && skillConfig != nil && skillConfig.AoeRadius > 0 {
		aoeTargets := s.calculateAOEDamage(roleID, monsterInstanceID, monster.X, monster.Y,
			skillConfig.AoeRadius, damage, skillConfig)
		if len(aoeTargets) > 0 {
			result.AOETargets = aoeTargets
			log.Printf("💥 技能[%s]AOE波及 %d 个目标", skillConfig.Name, len(aoeTargets))
		}
	}

	// 9. 通知怪物AI被攻击（触发追击/反击）
	s.notifyMonsterHurted(monsterInstanceID, roleID)

	// 10. 如果怪物死亡，处理掉落
	if isDead {
		// 通知AI服务怪物死亡，触发复活倒计时
		if globalAIService != nil {
			globalAIService.OnMonsterDeath(monsterInstanceID)
		}

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

			// 将掉落物品添加到玩家背包（★ 优化：先本地验证，减少DB压力）
			if len(drops) > 0 {
				itemSvc := item.NewService()

				// ★ 先检查背包容量（避免逐个添加后发现背包满）
				bagItems, bagErr := itemSvc.GetBagItems(roleID)
				currentBagCount := 0
				if bagErr == nil && bagItems != nil {
					currentBagCount = len(bagItems)
				}

				for _, itemID := range drops {
					// ★ 先从本地配置验证物品是否存在（避免无效DB查询）
					itemConfig := common.GetItemConfig(itemID)
					if itemConfig == nil {
						log.Printf("⚠️ 掉落物品[%d]不存在于items.json配置中，跳过", itemID)
						continue // 跳过无效物品，不调用DB
					}

					// ★ 背包容量检查
					if currentBagCount >= item.BagMaxSlots {
						log.Printf("⚠️ 玩家 %d 背包已满(%d/%d)，掉落物品[%d-%s]丢失",
							roleID, currentBagCount, item.BagMaxSlots, itemID, itemConfig.Name)
						result.DropItems = append(result.DropItems, map[string]interface{}{
							"item_id":   itemID,
							"slot":      -1, // -1表示背包满未添加
							"item_name": itemConfig.Name,
							"quality":   itemConfig.Quality,
							"icon":      itemConfig.Icon,
							"type":      itemConfig.Type,
							"type_name": getTypeName(itemConfig.Type),
							"reason":    "bag_full",
						})
						continue
					}

					slot, addErr := itemSvc.AddItem(roleID, itemID, 1, 0) // 0=不绑定
					if addErr != nil {
						log.Printf("添加掉落物品[%d-%s]到背包失败: %v", itemID, itemConfig.Name, addErr)
						result.DropItems = append(result.DropItems, map[string]interface{}{
							"item_id":   itemID,
							"slot":      -2, // -2表示添加失败
							"item_name": itemConfig.Name,
							"reason":    "add_failed",
						})
					} else {
						currentBagCount++ // ★ 成功添加后递增计数
						log.Printf("🎁 玩家 %d 获得掉落物品[%d-%s], 放入格子 %d (背包%d/%d)",
							roleID, itemID, itemConfig.Name, slot, currentBagCount, item.BagMaxSlots)

						// ★ 构建完整的物品信息供前端显示（包含icon、品质、属性等）
						dropInfo := map[string]interface{}{
							"item_id":     itemID,
							"slot":        slot,
							"item_name":   itemConfig.Name,
							"quality":     itemConfig.Quality,
							"icon":        itemConfig.Icon,
							"type":        itemConfig.Type,
							"type_name":   getTypeName(itemConfig.Type),
							"description": itemConfig.Description,
							"count":       1,
						}

						// ★ 装备类物品附加属性信息
						if itemConfig.Type == 2 { // 装备类型
							dropInfo["can_equip"] = true
							dropInfo["equip_type"] = itemConfig.EquipType
							dropInfo["attrs"] = map[string]int{
								"attack":  itemConfig.AttackBonus,
								"defense": itemConfig.DefenseBonus,
								"speed":   itemConfig.SpeedBonus,
								"hp":      itemConfig.HpBonus,
								"mp":      itemConfig.MpBonus,
							}
							// 过滤掉0值属性
							cleanAttrs := make(map[string]int)
							for k, v := range dropInfo["attrs"].(map[string]int) {
								if v != 0 {
									cleanAttrs[k] = v
								}
							}
							dropInfo["attrs"] = cleanAttrs
						}

						// ★ 消耗品类物品标记可使用
						if itemConfig.Type == 1 { // 消耗品类型
							dropInfo["can_use"] = true
							dropInfo["hp_restore"] = itemConfig.HpRestore
							dropInfo["mp_restore"] = itemConfig.MpRestore
						}

						result.DropItems = append(result.DropItems, dropInfo)
					}
				}

				// ★ 统计掉落结果
				successCount := 0
				for _, di := range result.DropItems {
					if slot, ok := di["slot"].(int); ok && slot >= 0 {
						successCount++
					}
				}
				if len(drops) > 0 {
					log.Printf("📦 掉落汇总: 总计%d个物品, 成功拾取%d个, 背包容量%d/%d",
						len(drops), successCount, currentBagCount, item.BagMaxSlots)
				}
			}
		}
	}

	return result
}

// calculateAOEDamage 计算AOE范围伤害
// 对目标点周围aoeRadius格内的其他怪物造成溅射伤害（主目标伤害的40%-60%）
func (s *Service) calculateAOEDamage(roleID uint64, mainTargetID uint64, centerX, centerY int,
	aoeRadius int, mainDamage int, skillConfig *common.SkillBaseConfig) []map[string]interface{} {

	var aoeResults []map[string]interface{}

	// 遍历所有怪物，找出范围内的其他目标
	if globalMonsterSvc == nil {
		return aoeResults
	}

	// AOE溅射比例：主伤害的 40%~60%
	splashRatio := 0.4 + rand.Float64()*0.2

	// 通过MonsterService获取范围内所有怪物实例
	// 这里需要遍历地图上的所有怪物，使用接口方式获取
	// 由于当前接口限制，我们通过全局怪物列表来查找
	// （实际生产环境应通过MapService或SceneService获取区域内的实体）

	// 获取玩家所在地图ID
	playerPos := getPlayerPosition(roleID)
	if playerPos == nil {
		return aoeResults
	}

	// 暴击不传递到AOE目标（避免伤害过高）
	aoeDamage := int(float64(mainDamage) * splashRatio)
	if aoeDamage < 1 {
		aoeDamage = 1
	}

	// 遍历场景中所有存活怪物，计算AOE范围
	// 注意：这里需要访问怪物管理器的全部实例列表
	// 当前架构下通过MonsterService接口逐个检查距离
	// 为性能考虑，后续可优化为空间索引

	// 使用common包获取地图上的怪物实例
	aoeTargets := s.getMonstersInRange(centerX, centerY, aoeRadius, mainTargetID)

	for _, target := range aoeTargets {
		// 对每个AOE目标造成溅射伤害
		targetNewHP, targetDead := globalMonsterSvc.MonsterTakeDamage(target.InstanceID, aoeDamage)

		aoeInfo := map[string]interface{}{
			"target_id":    target.InstanceID,
			"name":         target.Name,
			"damage":       aoeDamage,
			"current_hp":   targetNewHP,
			"is_dead":      targetDead,
			"splash_ratio": splashRatio,
		}
		aoeResults = append(aoeResults, aoeInfo)

		// 通知AI被攻击
		s.notifyMonsterHurted(target.InstanceID, roleID)

		// AOE目标死亡处理
		if targetDead {
			if globalAIService != nil {
				globalAIService.OnMonsterDeath(target.InstanceID)
			}
			exp, gold, drops, _ := globalMonsterSvc.MonsterDie(target.InstanceID)
			s.rewardPlayer(roleID, exp, gold)
			if len(drops) > 0 {
				// AOE击杀掉落也加入结果
				if _, ok := aoeInfo["drops"]; !ok {
					aoeInfo["drops"] = drops
				}
			}

			log.Printf("💀 AOE击杀: 怪物[%s](%d), 溅射伤害%d", target.Name, target.InstanceID, aoeDamage)
		} else {
			log.Printf("💥 AOE命中: 怪物[%s](%d), 溅射伤害%d, 剩余HP=%d",
				target.Name, target.InstanceID, aoeDamage, targetNewHP)
		}

		// AOE目标也附加BUFF（概率减半）
		if skillConfig != nil && rand.Float64() < 0.15 {
			buffID := skillConfig.BuffID
			if buffID == 0 {
				buffID = SkillTypeToBuffId[skillConfig.Type]
			}
			if buffID > 0 {
				buff.GetManager().AddBuff(target.InstanceID, 2, buffID, roleID)
			}
		}
	}

	return aoeResults
}

// aoeTargetInfo AOE目标信息（别名，使用common包定义）
// type aoeTargetInfo = common.AOETargetInfo

// getMonstersInRange 获取指定范围内的怪物（排除主目标）
func (s *Service) getMonstersInRange(centerX, centerY int, radius int, excludeID uint64) []common.AOETargetInfo {
	if globalMonsterSvc == nil {
		return nil
	}
	return globalMonsterSvc.GetMonstersInArea(centerX, centerY, radius, excludeID)
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
		// 未命中时也需返回玩家当前血量，避免前端显示 0/0
		result.PlayerHP = playerFighter.CurrentHP
		result.PlayerMaxHP = playerFighter.MaxHP
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
	SkillType     uint8  `json:"skill_type"`
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
		if skillConfig != nil && skillConfig.WeaponType > 0 {
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
		result.SkillType = skillConfig.Type
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

// getTypeName 根据物品类型ID返回中文名称
func getTypeName(itemType uint8) string {
	typeNames := map[uint8]string{
		1: "消耗品",
		2: "装备",
		3: "材料",
		4: "任务物品",
		5: "其他",
	}
	if name, ok := typeNames[itemType]; ok {
		return name
	}
	return "未知"
}
