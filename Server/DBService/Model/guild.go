package Model

import (
	"time"

	"gorm.io/gorm"
)

// Guild 公会表
type Guild struct {
	ID          uint64         `gorm:"primaryKey;comment:公会ID"`
	CreatedAt   time.Time      `gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Name        string         `gorm:"size:32;unique;comment:公会名称"`
	LeaderID    uint64         `gorm:"comment:会长ID"`
	LeaderName  string         `gorm:"size:32;comment:会长名称"`
	Level       uint32         `gorm:"default:1;comment:公会等级"`
	Exp         int64          `gorm:"default:0;comment:公会经验"`
	Notice      string         `gorm:"size:256;comment:公会公告"`
	Gold        int64          `gorm:"default:0;comment:公会资金"`
	MemberCount uint32         `gorm:"default:0;comment:成员数量"`
	MaxMembers  uint32         `gorm:"default:50;comment:最大成员数"`
}

// GuildMember 公会成员表
type GuildMember struct {
	ID         uint64         `gorm:"primaryKey;comment:成员ID"`
	CreatedAt  time.Time      `gorm:"comment:加入时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	GuildID    uint64         `gorm:"index;comment:公会ID"`
	RoleID     uint64         `gorm:"unique;comment:角色ID"`
	RoleName   string         `gorm:"size:32;comment:角色名称"`
	Level      uint32         `gorm:"comment:角色等级"`
	Title      uint8          `gorm:"default:1;comment:职位(1成员 2长老 3副族长 4族长)"`
	Contribute int64          `gorm:"default:0;comment:个人贡献"`
	JoinTime   time.Time      `gorm:"comment:加入时间"`
	LastOnline time.Time      `gorm:"comment:最后在线时间"`
}

// GuildApply 公会申请表
type GuildApply struct {
	ID         uint64         `gorm:"primaryKey;comment:申请ID"`
	CreatedAt  time.Time      `gorm:"comment:申请时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	GuildID    uint64         `gorm:"index;comment:公会ID"`
	RoleID     uint64         `gorm:"index;comment:申请人ID"`
	RoleName   string         `gorm:"size:32;comment:申请人名称"`
	Level      uint32         `gorm:"comment:申请人等级"`
	Message    string         `gorm:"size:64;comment:申请留言"`
	Status     uint8          `gorm:"default:0;comment:状态(0待处理 1已同意 2已拒绝)"`
	HandleTime *time.Time     `gorm:"comment:处理时间"`
}
