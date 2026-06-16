package Model

import (
	"time"

	"gorm.io/gorm"
)

// Account 账号表
type Account struct {
	gorm.Model
	Username      string     `gorm:"size:32;unique"`
	Password      string     `gorm:"size:128"`  // bcrypt哈希密码
	Salt          string     `gorm:"size:32"`   // 密码盐值
	Status        int        `gorm:"default:0"` // 0正常 1封禁
	LoginIP       string     `gorm:"size:32"`
	LastLoginTime *time.Time `gorm:"size:32"`
	LastLoginIP   string     `gorm:"size:32"`
}

// Role 角色表
type Role struct {
	gorm.Model
	AccountID    uint64 `gorm:"index"`
	Name         string `gorm:"size:32;unique"`
	Level        int    `gorm:"default:1"`
	Exp          int64
	Gold         int64  `gorm:"default:0"`
	BindGold     int64  `gorm:"default:0"`
	Yuanbao      int64  `gorm:"default:0"`
	Gender       uint8  `gorm:"default:0"`
	Appearance   uint32 `gorm:"default:0"`
	Hp           int    `gorm:"default:100"`
	MaxHp        int    `gorm:"default:100"`
	Mp           int    `gorm:"default:100"`
	MaxMp        int    `gorm:"default:100"`
	Stamina      int    `gorm:"default:100"`
	MaxStamina   int    `gorm:"default:100"`
	Attack       int    `gorm:"default:10"`
	Defense      int    `gorm:"default:5"`
	Speed        int    `gorm:"default:10"`
	Hit          int    `gorm:"default:50"`
	Dodge        int    `gorm:"default:10"`
	Crit         int    `gorm:"default:5"`
	CritDamage   int    `gorm:"default:150"`
	MapID        int    `gorm:"default:1"`
	MapX         int    `gorm:"default:100"`
	MapY         int    `gorm:"default:100"`
	PkMode       uint8  `gorm:"default:0"`
	PkValue      int    `gorm:"default:0"`
	Status       uint8  `gorm:"default:0"`
	KillCount    int    `gorm:"default:0"`
	DeathCount   int    `gorm:"default:0"`
	Title        string `gorm:"size:64"`
	CreateTime   time.Time
	LastLogin    time.Time
	LastLoginIP  string `gorm:"size:32"`
	LogoutTime   time.Time
	LastSaveTime time.Time
}

// RoleSkill 角色武学表
type RoleSkill struct {
	gorm.Model
	RoleID  uint64 `gorm:"index"`
	SkillID uint32
	Level   int   `gorm:"default:1"`
	Exp     int64 `gorm:"default:0"`
	IsEquip uint8 `gorm:"default:0"`
}

// SkillBase 武学基础表
type SkillBase struct {
	gorm.Model
	Name        string `gorm:"size:32"`
	Type        uint8  // 1内功 2身法 3护体 4拳法 5剑法 6刀法 7枪法 8斧法
	Level       uint32 `gorm:"default:1"`
	MaxLevel    uint32 `gorm:"default:10"`
	ExpFactor   int    // 升级经验系数
	HpBonus     int    // 生命加成
	MpBonus     int    // 内力加成
	AttackBonus int    // 攻击加成
	DefBonus    int    // 防御加成
	SpeedBonus  int    // 速度加成
	HitBonus    int    // 命中加成
	DodgeBonus  int    // 闪避加成
	CritBonus   int    // 暴击加成
	IsActive    uint8  `gorm:"default:1"`
}

// ItemBase 道具基础表
type ItemBase struct {
	gorm.Model
	Name          string `gorm:"size:32"`
	Type          uint8  // 1药品 2装备 3材料 4任务
	Quality       uint8  // 1白 2绿 3蓝 4紫 5橙
	LevelReq      uint32 `gorm:"default:1"`
	StackMax      uint32 `gorm:"default:1"`
	Price         int
	hpRestore     int    `gorm:"column:hp_restore"` // HP恢复
	mpRestore     int    `gorm:"column:mp_restore"` // MP恢复
	IsDropable    uint8  `gorm:"default:1"`
	IsSellable    uint8  `gorm:"default:1"`
	IsDestroyable uint8  `gorm:"default:1"`
	EquipType     uint8  // 装备类型 1武器 2衣服 3帽子 4鞋子 5戒指 6项链
	Description   string `gorm:"size:256"`
}

// RoleBag 角色背包表
type RoleBag struct {
	gorm.Model
	RoleID     uint64 `gorm:"index"`
	GridIndex  int
	ItemID     uint32
	Count      uint32 `gorm:"default:1"`
	IsBind     uint8  `gorm:"default:0"`
	DurCurrent *int   // 当前耐久
	DurMax     *int   // 最大耐久
	GetTime    time.Time
}

// RoleEquipment 角色装备表
type RoleEquipment struct {
	gorm.Model
	RoleID    uint64 `gorm:"index"`
	EquipType uint8  // 装备位置
	BagItemID *uint64
	EquipTime time.Time
}
