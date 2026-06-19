package Model

import (
	"time"

	"gorm.io/gorm"
)

// Friend 好友表
type Friend struct {
	ID         uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt  time.Time      `gorm:"comment:创建时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID     uint64         `gorm:"index;comment:角色ID"`
	FriendID   uint64         `gorm:"index;comment:好友ID"`
	FriendName string         `gorm:"size:32;comment:好友名称"`
	Status     uint8          `gorm:"default:0;comment:状态(0待确认 1已添加 2黑名单)"`
}

// FriendRequest 好友申请表
type FriendRequest struct {
	ID           uint64         `gorm:"primaryKey;comment:申请ID"`
	CreatedAt    time.Time      `gorm:"comment:申请时间"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间"`
	FromRoleID   uint64         `gorm:"index;comment:申请人ID"`
	FromRoleName string         `gorm:"size:32;comment:申请人名称"`
	ToRoleID     uint64         `gorm:"index;comment:被申请人ID"`
	Message      string         `gorm:"size:64;comment:申请留言"`
	Status       uint8          `gorm:"default:0;comment:状态(0待处理 1已同意 2已拒绝)"`
	HandledAt    *time.Time     `gorm:"comment:处理时间"`
}
