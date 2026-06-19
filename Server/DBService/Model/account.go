package Model

import (
	"time"

	"gorm.io/gorm"
)

// Account 账号表
type Account struct {
	ID            uint64         `gorm:"primaryKey;comment:账号ID"`
	CreatedAt     time.Time      `gorm:"comment:创建时间"`
	UpdatedAt     time.Time      `gorm:"comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Username      string         `gorm:"size:32;unique;comment:用户名"`
	Password      string         `gorm:"size:128;comment:密码(bcrypt哈希)"`
	Salt          string         `gorm:"size:32;comment:密码盐值"`
	Status        int            `gorm:"default:0;comment:状态(0正常 1封禁)"`
	LoginIP       string         `gorm:"size:32;comment:登录IP"`
	LastLoginTime *time.Time     `gorm:"comment:最后登录时间"`
	LastLoginIP   string         `gorm:"size:32;comment:最后登录IP"`
}

// Role 角色表
type Role struct {
	ID           uint64         `gorm:"primaryKey;comment:角色ID"`
	CreatedAt    time.Time      `gorm:"comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间"`
	AccountID    uint64         `gorm:"index;comment:账号ID"`
	Name         string         `gorm:"size:32;unique;comment:角色名"`
	Level        int            `gorm:"default:1;comment:等级"`
	Exp          int64          `gorm:"default:0;comment:经验值"`
	Gold         int64          `gorm:"default:0;comment:金币"`
	BindGold     int64          `gorm:"default:0;comment:绑定金币"`
	Yuanbao      int64          `gorm:"default:0;comment:元宝"`
	Gender       uint8          `gorm:"default:0;comment:性别(0男 1女)"`
	Appearance   uint32         `gorm:"default:0;comment:外观造型"`
	Hp           int            `gorm:"default:100;comment:当前生命"`
	MaxHp        int            `gorm:"default:100;comment:最大生命"`
	Mp           int            `gorm:"default:100;comment:当前内力"`
	MaxMp        int            `gorm:"default:100;comment:最大内力"`
	Stamina      int            `gorm:"default:100;comment:当前体力"`
	MaxStamina   int            `gorm:"default:100;comment:最大体力"`
	Attack       int            `gorm:"default:10;comment:攻击"`
	Defense      int            `gorm:"default:5;comment:防御"`
	Speed        int            `gorm:"default:10;comment:速度"`
	Hit          int            `gorm:"default:50;comment:命中"`
	Dodge        int            `gorm:"default:10;comment:闪避"`
	Crit         int            `gorm:"default:5;comment:暴击率"`
	CritDamage   int            `gorm:"default:150;comment:暴击伤害"`
	MapID        int            `gorm:"default:1;comment:所在地图ID"`
	MapX         int            `gorm:"default:100;comment:地图X坐标"`
	MapY         int            `gorm:"default:100;comment:地图Y坐标"`
	PkMode       uint8          `gorm:"default:0;comment:PK模式(0和平 1队伍 2帮派 3全体)"`
	PkValue      int            `gorm:"default:0;comment:惩罚值"`
	Status       uint8          `gorm:"default:0;comment:状态(0正常)"`
	KillCount    int            `gorm:"default:0;comment:击杀数"`
	DeathCount   int            `gorm:"default:0;comment:死亡数"`
	Title        string         `gorm:"size:64;comment:称号"`
	CreateTime   time.Time      `gorm:"comment:创建时间"`
	LastLogin    time.Time      `gorm:"comment:最后登录时间"`
	LastLoginIP  string         `gorm:"size:32;comment:最后登录IP"`
	LogoutTime   time.Time      `gorm:"comment:最后登出时间"`
	LastSaveTime time.Time      `gorm:"comment:最后保存时间"`
}

// RoleSkill 角色武学表
type RoleSkill struct {
	ID      uint64 `gorm:"primaryKey;comment:记录ID"`
	RoleID  uint64 `gorm:"index;comment:角色ID"`
	SkillID uint32 `gorm:"comment:武学ID"`
	Level   int    `gorm:"default:1;comment:武学等级"`
	Exp     int64  `gorm:"default:0;comment:武学经验"`
	IsEquip uint8  `gorm:"default:0;comment:是否装备(0否 1是)"`
}

// SkillBase 武学基础表
type SkillBase struct {
	ID          uint64 `gorm:"primaryKey;comment:武学ID"`
	Name        string `gorm:"size:32;comment:武学名称"`
	Type        uint8  `gorm:"comment:武学类型(1内功 2身法 3护体 4拳法 5剑法 6刀法 7枪法 8斧法)"`
	Level       uint32 `gorm:"default:1;comment:武学等级"`
	MaxLevel    uint32 `gorm:"default:10;comment:最大等级"`
	ExpFactor   int    `gorm:"comment:升级经验系数"`
	HpBonus     int    `gorm:"comment:生命加成"`
	MpBonus     int    `gorm:"comment:内力加成"`
	AttackBonus int    `gorm:"comment:攻击加成"`
	DefBonus    int    `gorm:"comment:防御加成"`
	SpeedBonus  int    `gorm:"comment:速度加成"`
	HitBonus    int    `gorm:"comment:命中加成"`
	DodgeBonus  int    `gorm:"comment:闪避加成"`
	CritBonus   int    `gorm:"comment:暴击加成"`
	IsActive    uint8  `gorm:"default:1;comment:是否主动(0被动 1主动)"`
}

// ItemBase 道具基础表
type ItemBase struct {
	ID            uint64 `gorm:"primaryKey;comment:道具ID"`
	Name          string `gorm:"size:32;comment:道具名称"`
	Type          uint8  `gorm:"comment:道具类型(1药品 2装备 3材料 4任务)"`
	Quality       uint8  `gorm:"comment:品质(1白 2绿 3蓝 4紫 5橙)"`
	LevelReq      uint32 `gorm:"default:1;comment:等级要求"`
	StackMax      uint32 `gorm:"default:1;comment:最大堆叠"`
	Price         int    `gorm:"comment:售价"`
	HpRestore     int    `gorm:"column:hp_restore;comment:HP恢复值"`
	MpRestore     int    `gorm:"column:mp_restore;comment:MP恢复值"`
	IsDropable    uint8  `gorm:"default:1;comment:是否可丢弃(0否 1是)"`
	IsSellable    uint8  `gorm:"default:1;comment:是否可出售(0否 1是)"`
	IsDestroyable uint8  `gorm:"default:1;comment:是否可销毁(0否 1是)"`
	EquipType     uint8  `gorm:"comment:装备类型(1武器 2衣服 3帽子 4鞋子 5戒指 6项链)"`
	Description   string `gorm:"size:256;comment:道具描述"`
}

// RoleBag 角色背包表
type RoleBag struct {
	ID         uint64         `gorm:"primaryKey;comment:背包记录ID"`
	CreatedAt  time.Time      `gorm:"comment:创建时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID     uint64         `gorm:"index;comment:角色ID"`
	GridIndex  int            `gorm:"comment:背包格子索引"`
	ItemID     uint32         `gorm:"comment:道具ID"`
	Count      uint32         `gorm:"default:1;comment:道具数量"`
	IsBind     uint8          `gorm:"default:0;comment:是否绑定(0否 1是)"`
	DurCurrent *int           `gorm:"comment:当前耐久度"`
	DurMax     *int           `gorm:"comment:最大耐久度"`
	GetTime    time.Time      `gorm:"comment:获得时间"`
}

// RoleEquipment 角色装备表
type RoleEquipment struct {
	ID        uint64         `gorm:"primaryKey;comment:装备记录ID"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID    uint64         `gorm:"index;comment:角色ID"`
	EquipType uint8          `gorm:"comment:装备位置(1武器 2衣服 3帽子 4鞋子 5戒指 6项链)"`
	BagItemID *uint64        `gorm:"comment:背包物品ID"`
	EquipTime time.Time      `gorm:"comment:装备时间"`
}
