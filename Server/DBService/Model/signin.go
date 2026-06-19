package Model

import (
	"time"

	"gorm.io/gorm"
)

// Ranking 排行榜表
type Ranking struct {
	ID        uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Type      uint8          `gorm:"index;comment:排行类型(1等级榜 2战力榜 3金币榜 4公会榜)"`
	RoleID    uint64         `gorm:"index;unique;comment:角色ID"`
	Name      string         `gorm:"size:32;comment:角色名称"`
	Value     int64          `gorm:"comment:排行值"`
	GuildID   uint64         `gorm:"comment:公会ID"`
	GuildName string         `gorm:"size:32;comment:公会名称"`
}

// RoleSignIn 角色签到表
type RoleSignIn struct {
	ID             uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt      time.Time      `gorm:"comment:创建时间"`
	UpdatedAt      time.Time      `gorm:"comment:更新时间"`
	DeletedAt      gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID         uint64         `gorm:"unique;comment:角色ID"`
	TotalDays      uint32         `gorm:"default:0;comment:累计签到天数"`
	ContinuousDays uint32         `gorm:"default:0;comment:连续签到天数"`
	LastSignIn     time.Time      `gorm:"comment:最后签到日期"`
	Month          uint32         `gorm:"comment:当前签到月份"`
}

// SignInReward 签到奖励配置表
type SignInReward struct {
	ID        uint64         `gorm:"primaryKey;comment:奖励ID"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Day       uint32         `gorm:"index;comment:第几天"`
	Type      uint8          `gorm:"comment:奖励类型(1道具 2金币 3元宝)"`
	ItemID    uint32         `gorm:"default:0;comment:物品ID"`
	Count     uint32         `gorm:"default:0;comment:物品数量"`
	IsDouble  uint8          `gorm:"default:0;comment:是否双倍奖励日"`
}
