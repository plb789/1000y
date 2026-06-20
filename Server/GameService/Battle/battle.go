package battle

import (
	"math"
	"math/rand"
	"time"
)

// AttackType 攻击类型
type AttackType uint8

const (
	AttackTypeNormal AttackType = iota // 普通攻击
	AttackTypeSkill                    // 技能攻击
)

// DamageType 伤害类型
type DamageType uint8

const (
	DamageTypePhysical DamageType = iota // 物理伤害
	DamageTypeMagic                      // 魔法伤害
	DamageTypeTrue                       // 真实伤害
)

// AttackResult 攻击结果
type AttackResult struct {
	AttackerID   uint64     // 攻击者ID
	AttackerType uint8      // 攻击者类型: 1=玩家, 2=怪物
	TargetID     uint64     // 目标ID
	TargetType   uint8      // 目标类型: 1=玩家, 2=怪物
	Damage       int        // 伤害值
	DamageType   DamageType // 伤害类型
	IsCrit       bool       // 是否暴击
	IsMiss       bool       // 是否闪避
	IsBlocked    bool       // 是否格挡
	SkillID      uint32     // 技能ID(如果是技能攻击)
	AttackType   AttackType // 攻击类型
	Timestamp    int64      // 时间戳
}

// Fighter 战斗者接口
type Fighter interface {
	GetID() uint64
	GetAttack() int
	GetDefense() int
	GetSpeed() int
	GetHit() int
	GetDodge() int
	GetCrit() int
	GetCritDamage() int
	TakeDamage(damage int, dtype DamageType) (newHP int, dead bool)
	GetHP() int
	GetMaxHP() int
	IsAlive() bool
}

// BaseFighter 基础战斗者属性
type BaseFighter struct {
	ID         uint64
	Attack     int
	Defense    int
	Speed      int
	Hit        int
	Dodge      int
	Crit       int
	CritDamage int
	CurrentHP  int
	MaxHP      int
	SkillBonus map[string]int // 来自武学的加成
}

// TakeDamage 受到伤害
func (f *BaseFighter) TakeDamage(damage int, dtype DamageType) (newHP int, dead bool) {
	f.CurrentHP -= damage
	if f.CurrentHP <= 0 {
		f.CurrentHP = 0
		return f.CurrentHP, true
	}
	return f.CurrentHP, false
}

// GetHP 获取当前HP
func (f *BaseFighter) GetHP() int {
	return f.CurrentHP
}

// GetMaxHP 获取最大HP
func (f *BaseFighter) GetMaxHP() int {
	return f.MaxHP
}

// IsAlive 检查是否存活
func (f *BaseFighter) IsAlive() bool {
	return f.CurrentHP > 0
}

// GetID 获取ID
func (f *BaseFighter) GetID() uint64 {
	return f.ID
}

// GetAttack 获取攻击力
func (f *BaseFighter) GetAttack() int {
	return f.Attack
}

// GetDefense 获取防御力
func (f *BaseFighter) GetDefense() int {
	return f.Defense
}

// GetSpeed 获取速度
func (f *BaseFighter) GetSpeed() int {
	return f.Speed
}

// GetHit 获取命中率
func (f *BaseFighter) GetHit() int {
	return f.Hit
}

// GetDodge 获取闪避率
func (f *BaseFighter) GetDodge() int {
	return f.Dodge
}

// GetCrit 获取暴击率
func (f *BaseFighter) GetCrit() int {
	return f.Crit
}

// GetCritDamage 获取暴击伤害
func (f *BaseFighter) GetCritDamage() int {
	return f.CritDamage
}

// CalculateDamage 计算伤害
// attacker: 攻击者
// defender: 防御者
// skillBonus: 技能加成(外功/拳法等额外伤害)
// 返回: 伤害值, 是否暴击
func CalculateDamage(attacker, defender *BaseFighter, skillBonus int) (int, bool) {
	// 基础伤害 = 攻击 - 防御 * 0.5
	baseDamage := attacker.Attack - defender.Defense/2
	if baseDamage < 1 {
		baseDamage = 1
	}

	// 加上技能加成
	baseDamage += skillBonus

	// 暴击计算
	isCrit := false
	critRate := attacker.Crit + attacker.SkillBonus["crit"]
	if rand.Float64()*100 < float64(critRate) {
		isCrit = true
		baseDamage = int(float64(baseDamage) * float64(attacker.CritDamage) / 100)
	}

	// 伤害波动 ±10%
	variance := float64(baseDamage) * 0.1
	baseDamage = int(float64(baseDamage) + (rand.Float64()*2-1)*variance)

	// 确保最小伤害
	if baseDamage < 1 {
		baseDamage = 1
	}

	return baseDamage, isCrit
}

// CalculateHit 计算命中率
// attacker: 攻击方
// defender: 防御方
// 返回: 是否命中
func CalculateHit(attacker, defender *BaseFighter) bool {
	// 基础命中率 = 攻击方命中 - 防御方闪避
	hitRate := float64(attacker.Hit - defender.Dodge)
	if hitRate < 30 {
		hitRate = 30 // 最低30%命中率
	}
	if hitRate > 95 {
		hitRate = 95 // 最高95%命中率
	}

	return rand.Float64()*100 < hitRate
}

// CalculateDodge 计算闪避
// 这个函数在CalculateHit返回false时使用
func CalculateDodge(attacker, defender *BaseFighter) bool {
	// 闪避率 = 防御方闪避 - 攻击方命中(负面修正)
	dodgeRate := float64(defender.Dodge - attacker.Hit/2)
	if dodgeRate < 5 {
		dodgeRate = 5 // 最低5%闪避率
	}
	if dodgeRate > 50 {
		dodgeRate = 50 // 最高50%闪避率
	}
	return rand.Float64()*100 < dodgeRate
}

// CalculateBlock 计算格挡
// attacker: 攻击方
// defender: 防御方
// 返回: 是否格挡, 格挡减伤比例(0-1)
func CalculateBlock(attacker, defender *BaseFighter) (bool, float64) {
	// 格挡率基于防御方的防御力
	blockRate := float64(defender.Defense) / (float64(defender.Defense) + float64(attacker.Attack)) * 30
	if blockRate > 40 {
		blockRate = 40 // 最高40%格挡率
	}

	if rand.Float64()*100 < blockRate {
		// 格挡成功，减少30%-50%伤害
		reduction := 0.3 + rand.Float64()*0.2
		return true, reduction
	}
	return false, 0
}

// CheckAttackRange 检查攻击范围
// attackerX, attackerY: 攻击者坐标
// targetX, targetY: 目标坐标
// attackRange: 攻击范围（格子数）
// 返回: 是否在范围内, 实际距离
func CheckAttackRange(attackerX, attackerY, targetX, targetY int, attackRange int) (bool, float64) {
	dx := float64(attackerX - targetX)
	dy := float64(attackerY - targetY)
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= float64(attackRange), distance
}

// CalculateFinalDamage 计算最终伤害（包含所有修正）
// attacker: 攻击者
// defender: 防御者
// skillBonus: 技能加成
// 返回: 最终伤害, 是否暴击, 是否闪避, 是否格挡, 格挡减伤比例
func CalculateFinalDamage(attacker, defender *BaseFighter, skillBonus int) (int, bool, bool, bool, float64) {
	// 1. 命中检测
	if !CalculateHit(attacker, defender) {
		return 0, false, true, false, 0 // 闪避
	}

	// 2. 格挡检测
	isBlocked, blockReduction := CalculateBlock(attacker, defender)

	// 3. 计算基础伤害
	damage, isCrit := CalculateDamage(attacker, defender, skillBonus)

	// 4. 应用格挡减伤
	if isBlocked {
		damage = int(float64(damage) * (1 - blockReduction))
		if damage < 1 {
			damage = 1
		}
	}

	return damage, isCrit, false, isBlocked, blockReduction
}

// SkillDamageFactor 技能伤害系数
// 不同类型的武学有不同的伤害系数
var SkillDamageFactor = map[uint8]float64{
	1: 1.0, // 内功 - 无伤害
	2: 1.5, // 外功 - 高伤害
	3: 1.0, // 身法 - 无伤害
	4: 1.0, // 护体 - 无伤害
	5: 1.2, // 拳法 - 中高伤害
	6: 1.3, // 剑法 - 中高伤害
	7: 1.4, // 刀法 - 高伤害
	8: 1.5, // 枪法 - 高伤害
	9: 1.6, // 斧法 - 最高伤害
}

// CalculateSkillDamage 计算技能伤害
func CalculateSkillDamage(attacker, defender *BaseFighter, skillType uint8, skillLevel uint32) int {
	// 基础伤害 = 攻击 - 防御 * 0.5
	baseDamage := attacker.Attack - defender.Defense/2
	if baseDamage < 1 {
		baseDamage = 1
	}

	// 技能伤害系数
	factor := SkillDamageFactor[skillType]
	if factor == 0 {
		factor = 1.0
	}

	// 技能等级加成: 每级 +5%
	levelBonus := 1.0 + float64(skillLevel-1)*0.05

	// 最终伤害
	damage := int(float64(baseDamage) * factor * levelBonus)

	// 伤害波动
	variance := float64(damage) * 0.1
	damage = int(float64(damage) + (rand.Float64()*2-1)*variance)

	if damage < 1 {
		damage = 1
	}

	return damage
}

// CalculateExpGain 计算经验获取
// 击杀怪物时,根据怪物等级和角色等级差计算经验
func CalculateExpGain(monsterLevel, roleLevel uint32, baseExp int) int {
	// 等级差修正
	levelDiff := int(monsterLevel) - int(roleLevel)
	var multiplier float64 = 1.0

	if levelDiff >= 10 {
		multiplier = 1.5 // 越10级击杀,经验+50%
	} else if levelDiff >= 5 {
		multiplier = 1.2 // 越5级击杀,经验+20%
	} else if levelDiff <= -10 {
		multiplier = 0.3 // 低于10级击杀,经验-70%
	} else if levelDiff <= -5 {
		multiplier = 0.5 // 低于5级击杀,经验-50%
	}

	return int(float64(baseExp) * multiplier)
}

// CalculateGoldDrop 计算金币掉落
func CalculateGoldDrop(minGold, maxGold int) int {
	if minGold >= maxGold {
		return minGold
	}
	return minGold + rand.Intn(maxGold-minGold+1)
}

// BattleState 战斗状态
type BattleState struct {
	ID         uint64
	Type       uint8 // 1=玩家vs玩家, 2=玩家vs怪物, 3=怪物vs玩家
	AttackerID uint64
	DefenderID uint64
	AttackerHP int
	DefenderHP int
	StartTime  time.Time
	EndTime    time.Time
	Status     uint8 // 0=进行中, 1=结束
	Winner     uint8 // 0=无, 1=攻击方, 2=防御方
}

// NewBattleState 创建战斗状态
func NewBattleState(id, attackerID, defenderID uint64, battleType uint8) *BattleState {
	return &BattleState{
		ID:         id,
		Type:       battleType,
		AttackerID: attackerID,
		DefenderID: defenderID,
		StartTime:  time.Now(),
		Status:     0,
	}
}

// EndBattle 结束战斗
func (bs *BattleState) EndBattle(winner uint8) {
	bs.Status = 1
	bs.Winner = winner
	bs.EndTime = time.Now()
}

// IsInBattle 检查是否在战斗中
func IsInBattle(fighterID uint64, battles map[uint64]*BattleState) bool {
	if battle, ok := battles[fighterID]; ok && battle.Status == 0 {
		return true
	}
	return false
}

// GetBattleOpponent 获取战斗对手
func GetBattleOpponent(fighterID uint64, battles map[uint64]*BattleState) uint64 {
	if battle, ok := battles[fighterID]; ok && battle.Status == 0 {
		if battle.AttackerID == fighterID {
			return battle.DefenderID
		}
		return battle.AttackerID
	}
	return 0
}

// PVPExpFactor PVP经验系数(较低,防止刷经验)
var PVPExpFactor = 0.3

// CalculatePVPExp 计算PVP经验
func CalculatePVPExp(winnerLevel, loserLevel uint32) int {
	// 基础经验 = (对手等级 * 10)
	baseExp := int(loserLevel) * 10

	// 等级差修正
	levelDiff := int(winnerLevel) - int(loserLevel)
	if levelDiff > 0 {
		baseExp = int(float64(baseExp) * (1.0 - float64(levelDiff)*0.05))
	}

	if baseExp < 1 {
		baseExp = 1
	}

	return int(float64(baseExp) * PVPExpFactor)
}

// CalculatePenaltyExp 死亡惩罚经验
func CalculatePenaltyExp(roleLevel uint32, currentExp int64) int {
	// 死亡损失经验 = 当前等级 * 50
	penalty := int(roleLevel) * 50

	// 已有经验不足以扣除时,扣到0
	if int64(penalty) > currentExp {
		penalty = int(currentExp)
	}

	return penalty
}

// CoolDownManager 冷却管理器
type CoolDownManager struct {
	cooldowns map[uint64]map[uint32]int64 // key=角色ID, key=技能ID, value=过期时间戳
}

// NewCoolDownManager 创建冷却管理器
func NewCoolDownManager() *CoolDownManager {
	return &CoolDownManager{
		cooldowns: make(map[uint64]map[uint32]int64),
	}
}

// IsCoolingDown 检查是否在冷却中
func (cdm *CoolDownManager) IsCoolingDown(roleID uint64, skillID uint32) bool {
	if skillCooldowns, ok := cdm.cooldowns[roleID]; ok {
		if expireTime, ok := skillCooldowns[skillID]; ok {
			return time.Now().UnixMilli() < expireTime
		}
	}
	return false
}

// SetCoolDown 设置冷却
func (cdm *CoolDownManager) SetCoolDown(roleID uint64, skillID uint32, durationMs int64) {
	if _, ok := cdm.cooldowns[roleID]; !ok {
		cdm.cooldowns[roleID] = make(map[uint32]int64)
	}
	cdm.cooldowns[roleID][skillID] = time.Now().UnixMilli() + durationMs
}

// GetCoolDownRemaining 获取剩余冷却时间
func (cdm *CoolDownManager) GetCoolDownRemaining(roleID uint64, skillID uint32) int64 {
	if skillCooldowns, ok := cdm.cooldowns[roleID]; ok {
		if expireTime, ok := skillCooldowns[skillID]; ok {
			remaining := expireTime - time.Now().UnixMilli()
			if remaining > 0 {
				return remaining
			}
		}
	}
	return 0
}

// ClearCoolDown 清除冷却
func (cdm *CoolDownManager) ClearCoolDown(roleID uint64, skillID uint32) {
	if skillCooldowns, ok := cdm.cooldowns[roleID]; ok {
		delete(skillCooldowns, skillID)
	}
}

// ClearAllCoolDowns 清除所有冷却
func (cdm *CoolDownManager) ClearAllCoolDowns(roleID uint64) {
	delete(cdm.cooldowns, roleID)
}

// CalculateDistance 计算两点距离(格子)
// func CalculateDistance(x1, y1, x2, y2 int) int {
// 	dx := math.Abs(x1 - x2)
// 	dy := math.Abs(y1 - y2)
// 	return int(math.Sqrt(float64(dx*dx + dy*dy)))
// }

// CalculateDistanceSqrt 计算两点距离平方(用于比较)
func CalculateDistanceSqrt(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy
}

// IsInAttackRange 检查是否在攻击范围内
func IsInAttackRange(x1, y1, x2, y2, rangeTiles int) bool {
	distSq := CalculateDistanceSqrt(x1, y1, x2, y2)
	return distSq <= rangeTiles*rangeTiles
}

// GlobalCoolDownManager 全局冷却管理器
var GlobalCoolDownManager = NewCoolDownManager()
