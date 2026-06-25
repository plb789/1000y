package Model

import (
	"time"

	"gorm.io/gorm"
)

// TaskBase 任务基础表
type TaskBase struct {
	ID           uint64         `gorm:"primaryKey;comment:任务ID"`
	CreatedAt    time.Time      `gorm:"comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index;comment:删除时间"`
	Name         string         `gorm:"size:32;comment:任务名称"`
	Type         uint8          `gorm:"default:1;comment:任务类型(1主线 2支线 3日常 4周常 5活动)"`
	Desc         string         `gorm:"size:256;comment:任务描述"`
	TargetType   uint8          `gorm:"comment:目标类型(1杀怪 2采集 3对话 4探索)"`
	TargetID     uint32         `gorm:"comment:目标ID"`
	TargetCount  uint32         `gorm:"default:1;comment:目标数量"`
	ExpReward    int64          `gorm:"default:0;comment:经验奖励"`
	GoldReward   int64          `gorm:"default:0;comment:金币奖励"`
	HonorReward  int64          `gorm:"default:0;comment:声望奖励(新增)"`
	ItemReward   uint32         `gorm:"default:0;comment:物品奖励ID"`
	ItemCount    uint32         `gorm:"default:0;comment:物品奖励数量"`
	LevelReq     uint32         `gorm:"default:1;comment:等级要求"`
	PreTaskID    uint32         `gorm:"default:0;comment:前置任务ID"`
	Repeatable   uint8          `gorm:"default:0;comment:是否可重复(新增)"`
	TimeLimit    int            `gorm:"default:0;comment:时间限制(秒,新增)"`
	AutoAccept   uint8          `gorm:"default:0;comment:是否自动接取(新增)"`
	AutoComplete uint8          `gorm:"default:0;comment:是否自动完成(新增)"`
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
	Objectives   string         `gorm:"type:text;comment:多目标进度JSON(新增)"` // 多目标进度JSON
	AcceptTime   time.Time      `gorm:"comment:接取时间"`
	CompleteTime *time.Time     `gorm:"comment:完成时间"`
	FinishTime   *time.Time     `gorm:"comment:领奖时间(新增)"`             // 领奖时间
	DailyCount   int            `gorm:"default:0;comment:今日完成次数(新增)"` // 今日完成次数
	TotalCount   int            `gorm:"default:0;comment:总完成次数(新增)"`  // 总完成次数
}

// TableName 指定表名
func (RoleTask) TableName() string {
	return "role_tasks"
}
