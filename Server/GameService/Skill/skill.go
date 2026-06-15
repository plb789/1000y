package model

import (
	"time"
)

// SkillBase 武学基础表
type SkillBase struct {
	ID          uint32 `gorm:"primaryKey;column:id" json:"id"`                  // 武学ID
	Name        string `gorm:"column:name;size:50" json:"name"`                  // 武学名称
	Type        uint8  `gorm:"column:type" json:"type"`                         // 类型: 1=内功, 2=外功, 3=身法, 4=护体, 5=拳法, 6=剑法, 7=刀法, 8=枪法, 9=斧法
	SubType     uint8  `gorm:"column:sub_type" json:"sub_type"`                 // 子类型: 1=无名武功(初级), 2=进阶武功, 3=高级武功
	Level       uint32 `gorm:"column:level" json:"level"`                       // 等级要求
	MaxLevel    uint32 `gorm:"column:max_level" json:"max_level"`               // 武学最高等级
	ExpFactor   uint32 `gorm:"column:exp_factor" json:"exp_factor"`             // 经验系数(每级所需经验)
	Description string `gorm:"column:description;size:255" json:"description"`   // 武学描述
	HpBonus     int    `gorm:"column:hp_bonus" json:"hp_bonus"`                // 生命加成/级
	MpBonus     int    `gorm:"column:mp_bonus" json:"mp_bonus"`                // 内力加成/级
	AttackBonus int    `gorm:"column:attack_bonus" json:"attack_bonus"`         // 攻击加成/级
	DefBonus    int    `gorm:"column:defense_bonus" json:"defense_bonus"`       // 防御加成/级
	SpeedBonus  int    `gorm:"column:speed_bonus" json:"speed_bonus"`           // 速度加成/级
	HitBonus    int    `gorm:"column:hit_bonus" json:"hit_bonus"`               // 命中加成/级
	DodgeBonus  int    `gorm:"column:dodge_bonus" json:"dodge_bonus"`            // 闪避加成/级
	CritBonus   int    `gorm:"column:crit_bonus" json:"crit_bonus"`              // 暴击加成/级
	BuffID      uint32 `gorm:"column:buff_id" json:"buff_id"`                   // 被动BUFF ID
	SkillEffect string `gorm:"column:skill_effect;size:100" json:"skill_effect"` // 技能特效
	IsActive    uint8  `gorm:"column:is_active" json:"is_active"`                // 是否激活
}

func (SkillBase) TableName() string {
	return "skill_base"
}

// RoleSkill 角色武学表
type RoleSkill struct {
	ID       uint64    `gorm:"primaryKey;column:id" json:"id"`         // 记录ID
	RoleID   uint64    `gorm:"column:role_id;index" json:"role_id"`    // 角色ID
	SkillID  uint32    `gorm:"column:skill_id" json:"skill_id"`       // 武学ID
	Level    uint32    `gorm:"column:level;default:1" json:"level"`    // 当前等级
	Exp      int64     `gorm:"column:exp;default:0" json:"exp"`        // 当前熟练度
	IsEquip  uint8     `gorm:"column:is_equip;default:0" json:"is_equip"` // 是否装备: 0=未装备, 1=已装备
	LearnTime time.Time `gorm:"column:learn_time" json:"learn_time"`    // 学习时间
}

func (RoleSkill) TableName() string {
	return "role_skill"
}

// RoleSkillWithBase 角色武学详情(关联武学基础信息)
type RoleSkillWithBase struct {
	RoleSkill
	SkillBase
}

// SkillType 武学类型常量
type SkillType uint8

const (
	SkillTypeNeigong  SkillType = 1 // 内功
	SkillTypeWaigong  SkillType = 2 // 外功
	SkillTypeShenfa   SkillType = 3 // 身法
	SkillTypeHuti     SkillType = 4 // 护体
	SkillTypeQuanfa   SkillType = 5 // 拳法
	SkillTypeJianfa   SkillType = 6 // 剑法
	SkillTypeDaofa    SkillType = 7 // 刀法
	SkillTypeQiangfa  SkillType = 8 // 枪法
	SkillTypeFufa     SkillType = 9 // 斧法
)

// SkillSubType 武学子类型常量
type SkillSubType uint8

const (
	SkillSubTypeWuming SkillSubType = 1 // 无名武功(初级)
	SkillSubTypeJincheng SkillSubType = 2 // 进阶武功
	SkillSubTypeGaoji SkillSubType = 3 // 高级武功
)

// SkillTypeName 武学类型名称映射
var SkillTypeName = map[uint8]string{
	1: "内功",
	2: "外功",
	3: "身法",
	4: "护体",
	5: "拳法",
	6: "剑法",
	7: "刀法",
	8: "枪法",
	9: "斧法",
}

// SkillSubTypeName 武学子类型名称映射
var SkillSubTypeName = map[uint8]string{
	1: "初级",
	2: "进阶",
	3: "高级",
}
