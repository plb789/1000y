package Model

import (
	"time"

	"gorm.io/gorm"
)

// Mail 邮件表
type Mail struct {
	ID          uint64         `gorm:"primaryKey;comment:邮件ID"`
	CreatedAt   time.Time      `gorm:"comment:发送时间"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Title       string         `gorm:"size:64;comment:邮件标题"`
	Content     string         `gorm:"size:1024;comment:邮件内容"`
	FromRoleID  uint64         `gorm:"comment:发送者ID(0为系统)"`
	FromName    string         `gorm:"size:32;comment:发送者名称"`
	ToRoleID    uint64         `gorm:"index;comment:接收者ID"`
	ItemID      uint32         `gorm:"default:0;comment:附件物品ID"`
	ItemCount   uint32         `gorm:"default:0;comment:附件物品数量"`
	Gold        int64          `gorm:"default:0;comment:附件金币"`
	Yuanbao     int64          `gorm:"default:0;comment:附件元宝"`
	IsRead      uint8          `gorm:"default:0;comment:是否已读(0未读 1已读)"`
	IsGetAttach uint8          `gorm:"default:0;comment:是否领取附件(0未领取 1已领取)"`
	ExpiredAt   time.Time      `gorm:"comment:过期时间"`
}
