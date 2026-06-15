package model

import (
	"time"
)

// Role 角色表
type Role struct {
	ID         uint64 `gorm:"primaryKey;column:id" json:"id"`                // 角色ID
	AccountID  uint64 `gorm:"column:account_id;index" json:"account_id"`     // 账号ID
	Name       string `gorm:"column:name;size:20" json:"name"`               // 角色名
	Level      uint32 `gorm:"column:level;default:1" json:"level"`           // 等级
	Exp        int64  `gorm:"column:exp;default:0" json:"exp"`               // 经验值
	Gold       int64  `gorm:"column:gold;default:0" json:"gold"`             // 金币
	BindGold   int64  `gorm:"column:bind_gold;default:0" json:"bind_gold"`   // 绑定金币
	Yuanbao    int    `gorm:"column:yuanbao;default:0" json:"yuanbao"`       // 元宝
	Gender     uint8  `gorm:"column:gender" json:"gender"`                   // 性别: 0=男, 1=女
	Appearance uint32 `gorm:"column:appearance;default:0" json:"appearance"` // 形象ID
	Title      string `gorm:"column:title;size:50" json:"title"`             // 称号

	// 基础属性
	Hp         int `gorm:"column:hp;default:100" json:"hp"`                   // 生命值
	MaxHp      int `gorm:"column:max_hp;default:100" json:"max_hp"`           // 最大生命值
	Mp         int `gorm:"column:mp;default:100" json:"mp"`                   // 内力值
	MaxMp      int `gorm:"column:max_mp;default:100" json:"max_mp"`           // 最大内力值
	Stamina    int `gorm:"column:stamina;default:100" json:"stamina"`         // 体力值
	MaxStamina int `gorm:"column:max_stamina;default:100" json:"max_stamina"` // 最大体力值

	// 战斗属性
	Attack     int `gorm:"column:attack;default:10" json:"attack"`            // 攻击
	Defense    int `gorm:"column:defense;default:5" json:"defense"`           // 防御
	Speed      int `gorm:"column:speed;default:10" json:"speed"`              // 速度
	Hit        int `gorm:"column:hit;default:50" json:"hit"`                  // 命中
	Dodge      int `gorm:"column:dodge;default:10" json:"dodge"`              // 闪避
	Crit       int `gorm:"column:crit;default:5" json:"crit"`                 // 暴击率
	CritDamage int `gorm:"column:crit_damage;default:150" json:"crit_damage"` // 暴击伤害

	// 位置信息
	MapID int `gorm:"column:map_id;default:1" json:"map_id"` // 当前地图ID
	MapX  int `gorm:"column:map_x;default:100" json:"map_x"` // 地图X坐标
	MapY  int `gorm:"column:map_y;default:100" json:"map_y"` // 地图Y坐标

	// PK相关
	PkMode     uint8 `gorm:"column:pk_mode;default:0" json:"pk_mode"`         // PK模式: 0=和平, 1=全体, 2=门派, 3=组队
	PkValue    int   `gorm:"column:pk_value;default:0" json:"pk_value"`       // PK值(善恶值)
	KillCount  int   `gorm:"column:kill_count;default:0" json:"kill_count"`   // 杀人数量
	DeathCount int   `gorm:"column:death_count;default:0" json:"death_count"` // 死亡次数

	// 状态
	Status  uint8 `gorm:"column:status;default:0" json:"status"`     // 状态: 0=正常, 1=打坐, 2=死亡, 3=在线
	HpRegen int   `gorm:"column:hp_regen;default:1" json:"hp_regen"` // 生命回复
	MpRegen int   `gorm:"column:mp_regen;default:1" json:"mp_regen"` // 内力回复

	// 时间戳
	CreateTime   time.Time `gorm:"column:create_time" json:"create_time"`           // 创建时间
	LastSaveTime time.Time `gorm:"column:last_save_time" json:"last_save_time"`     // 最后保存时间
	OnlineTime   int64     `gorm:"column:online_time;default:0" json:"online_time"` // 累计在线时间(秒)
	LogoutTime   time.Time `gorm:"column:logout_time" json:"logout_time"`           // 最后下线时间
}

func (Role) TableName() string {
	return "role"
}

// RoleBrief 角色简要信息(用于列表展示)
type RoleBrief struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Level      uint32 `json:"level"`
	Gender     uint8  `json:"gender"`
	Appearance uint32 `json:"appearance"`
	MapID      int    `json:"map_id"`
	Title      string `json:"title"`
}

// RoleDetail 角色详细信息
type RoleDetail struct {
	Role
	AccountName string `json:"account_name,omitempty"` // 账号名(仅GM可见)
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	AccountID  uint64 `json:"account_id" binding:"required"`
	Name       string `json:"name" binding:"required,min=2,max=20"`
	Gender     uint8  `json:"gender" binding:"required,oneof=0 1"`
	Appearance uint32 `json:"appearance"`
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	Title  string `json:"title"`
	Gender uint8  `json:"gender"`
	Exp    int64  `json:"exp"`
	Level  uint32 `json:"level"`
}

// RoleAttributeRequest 属性修改请求
type RoleAttributeRequest struct {
	Hp       *int   `json:"hp"`
	MaxHp    *int   `json:"max_hp"`
	Mp       *int   `json:"mp"`
	MaxMp    *int   `json:"max_mp"`
	Attack   *int   `json:"attack"`
	Defense  *int   `json:"defense"`
	Speed    *int   `json:"speed"`
	Hit      *int   `json:"hit"`
	Dodge    *int   `json:"dodge"`
	Crit     *int   `json:"crit"`
	Gold     *int64 `json:"gold"`
	BindGold *int64 `json:"bind_gold"`
	Yuanbao  *int   `json:"yuanbao"`
}

// PKMode PK模式常量
var PKModeName = map[uint8]string{
	0: "和平",
	1: "全体",
	2: "门派",
	3: "组队",
}

// RoleStatus 角色状态常量
var RoleStatusName = map[uint8]string{
	0: "正常",
	1: "打坐",
	2: "死亡",
	3: "在线",
}
