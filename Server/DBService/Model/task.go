package Model

import (
	"time"

	"gorm.io/gorm"
)

// TaskBase 任务基础表
type TaskBase struct {
	ID          uint64         `gorm:"primaryKey;comment:任务ID"`
	CreatedAt   time.Time      `gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Name        string         `gorm:"size:32;comment:任务名称"`
	Type        uint8          `gorm:"default:1;comment:任务类型(1主线 2支线 3日常 4周常)"`
	Desc        string         `gorm:"size:256;comment:任务描述"`
	TargetType  uint8          `gorm:"comment:目标类型(1杀怪 2采集 3对话 4探索)"`
	TargetID    uint32         `gorm:"comment:目标ID"`
	TargetCount uint32         `gorm:"default:1;comment:目标数量"`
	ExpReward   int64          `gorm:"default:0;comment:经验奖励"`
	GoldReward  int64          `gorm:"default:0;comment:金币奖励"`
	ItemReward  uint32         `gorm:"default:0;comment:物品奖励ID"`
	ItemCount   uint32         `gorm:"default:0;comment:物品奖励数量"`
	LevelReq    uint32         `gorm:"default:1;comment:等级要求"`
	PreTaskID   uint32         `gorm:"default:0;comment:前置任务ID"`
}

// RoleTask 角色任务进度表
type RoleTask struct {
	ID           uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt    time.Time      `gorm:"comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID       uint64         `gorm:"index;comment:角色ID"`
	TaskID       uint32         `gorm:"index;comment:任务ID"`
	Status       uint8          `gorm:"default:0;comment:状态(0未接 1进行中 2已完成 3已领奖)"`
	Progress     uint32         `gorm:"default:0;comment:当前进度"`
	AcceptTime   time.Time      `gorm:"comment:接取时间"`
	CompleteTime *time.Time     `gorm:"comment:完成时间"`
}
