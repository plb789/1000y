package model

import (
	"time"
)

// RoleSkill 角色武学表
type RoleSkill struct {
	ID       uint64    `gorm:"primaryKey;column:id" json:"id"`         // 记录ID
	RoleID   uint64    `gorm:"column:role_id;index" json:"role_id"`    // 角色ID
	SkillID  uint32    `gorm:"column:skill_id" json:"skill_id"`       // 武学ID
	Level    uint32    `gorm:"column:level;default:1" json:"level"`    // 当前等级
	Exp      int64     `gorm:"column:exp;default:0" json:"exp"`       // 当前熟练度
	IsEquip  uint8     `gorm:"column:is_equip;default:0" json:"is_equip"` // 是否装备
	LearnTime time.Time `gorm:"column:learn_time" json:"learn_time"`   // 学习时间
}

func (RoleSkill) TableName() string {
	return "role_skill"
}

// MonsterBase 怪物基础表
type MonsterBase struct {
	ID          uint32 `gorm:"primaryKey;column:id" json:"id"`             // 怪物ID
	Name        string `gorm:"column:name;size:50" json:"name"`              // 怪物名称
	Level       uint32 `gorm:"column:level;default:1" json:"level"`         // 等级
	Type        uint8  `gorm:"column:type;default:0" json:"type"`            // 类型: 0=普通, 1=精英, 2=BOSS
	MapID       uint32 `gorm:"column:map_id;index" json:"map_id"`           // 所属地图ID
	Hp          int    `gorm:"column:hp" json:"hp"`                         // 生命值
	Attack      int    `gorm:"column:attack" json:"attack"`                 // 攻击力
	Defense     int    `gorm:"column:defense" json:"defense"`               // 防御力
	Speed       int    `gorm:"column:speed" json:"speed"`                   // 速度
	Hit         int    `gorm:"column:hit;default:50" json:"hit"`             // 命中率
	Dodge       int    `gorm:"column:dodge;default:10" json:"dodge"`       // 闪避率
	Crit        int    `gorm:"column:crit;default:5" json:"crit"`           // 暴击率
	AIType      uint8  `gorm:"column:ai_type;default:0" json:"ai_type"`     // AI类型
	AttackRange int    `gorm:"column:attack_range;default:3" json:"attack_range"`   // 警戒范围
	ChaseRange  int    `gorm:"column:chase_range;default:5" json:"chase_range"`     // 追击范围
	GoldMin     int    `gorm:"column:gold_min;default:0" json:"gold_min"`   // 金币掉落下限
	GoldMax     int    `gorm:"column:gold_max;default:0" json:"gold_max"`   // 金币掉落上限
	Exp         int    `gorm:"column:exp;default:0" json:"exp"`              // 经验掉落
	DropGroupID *uint32 `gorm:"column:drop_group_id" json:"drop_group_id"`  // 掉落组ID
	SpriteID    *int   `gorm:"column:sprite_id" json:"sprite_id"`           // 精灵ID
	RespawnTime int    `gorm:"column:respawn_time;default:60" json:"respawn_time"` // 复活时间(秒)
}

func (MonsterBase) TableName() string {
	return "monster_base"
}
