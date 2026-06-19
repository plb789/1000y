package Model

import (
	"time"

	"gorm.io/gorm"
)

// ChatLog 聊天记录表
type ChatLog struct {
	ID           uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt    time.Time      `gorm:"comment:发送时间"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Channel      uint8          `gorm:"index;comment:聊天频道(1世界 2当前 3私聊 4帮会 5系统)"`
	SenderID     uint64         `gorm:"index;comment:发送者ID"`
	SenderName   string         `gorm:"size:32;comment:发送者名称"`
	ReceiverID   uint64         `gorm:"comment:接收者ID(私聊用)"`
	ReceiverName string         `gorm:"size:32;comment:接收者名称"`
	Content      string         `gorm:"size:256;comment:聊天内容"`
}

// ChatChannel 聊天频道常量
const (
	ChatWorld   uint8 = 1 // 世界频道
	ChatCurrent       = 2 // 当前频道
	ChatPrivate       = 3 // 私聊频道
	ChatGuild         = 4 // 帮会频道
	ChatSystem        = 5 // 系统频道
)
