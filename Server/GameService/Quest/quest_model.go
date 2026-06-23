package quest

import (
	"time"

	common "game-server/Common"
)

// QuestStatus 任务状态常量
const (
	QuestStatusAvailable = 0 // 可接取（未接取）
	QuestStatusActive    = 1 // 进行中
	QuestStatusCompleted = 2 // 已完成未领奖
	QuestStatusFinished  = 3 // 已完成已领奖
)

// QuestTypeName 任务类型名称映射
var QuestTypeName = map[uint8]string{
	1: "主线",
	2: "支线",
	3: "日常",
	4: "周常",
	5: "活动",
}

// TargetTypeName 目标类型名称映射
var TargetTypeName = map[uint8]string{
	1: "击杀",
	2: "采集",
	3: "对话",
	4: "探索",
}

// RoleQuest 角色任务表（数据库持久化）
type RoleQuest struct {
	ID           uint64     `gorm:"primaryKey;column:id" json:"id"`                  // 记录ID
	RoleID       uint64     `gorm:"column:role_id;index" json:"role_id"`             // 角色ID
	QuestID      uint32     `gorm:"column:quest_id" json:"quest_id"`                 // 任务ID
	Status       uint8      `gorm:"column:status;default:0" json:"status"`           // 状态: 0=可接取, 1=进行中, 2=已完成, 3=已领奖
	Progress     uint32     `gorm:"column:progress;default:0" json:"progress"`       // 当前进度
	AcceptTime   time.Time  `gorm:"column:accept_time" json:"accept_time"`           // 接取时间
	CompleteTime *time.Time `gorm:"column:complete_time" json:"complete_time"`       // 完成时间
	FinishTime   *time.Time `gorm:"column:finish_time" json:"finish_time"`           // 领奖时间
	DailyCount   int        `gorm:"column:daily_count;default:0" json:"daily_count"` // 今日完成次数（日常/周常用）
	TotalCount   int        `gorm:"column:total_count;default:0" json:"total_count"` // 总完成次数
}

func (RoleQuest) TableName() string {
	return "role_quest"
}

// QuestInfo 任务详情（返回给前端的完整信息，包含配置+进度）
type QuestInfo struct {
	common.QuestBaseConfig
	Status     uint8  `json:"status"`      // 当前状态
	Progress   uint32 `json:"progress"`    // 当前进度
	DailyCount int    `json:"daily_count"` // 今日完成次数
	TotalCount int    `json:"total_count"` // 总完成次数
}

// QuestListResponse 任务列表响应
type QuestListResponse struct {
	ActiveQuests    []QuestInfo `json:"active_quests"`    // 进行中的任务
	CompletedQuests []QuestInfo `json:"completed_quests"` // 已完成未领奖的任务
	AvailableQuests []QuestInfo `json:"available_quests"` // 可接取的任务
	FinishedQuests  []QuestInfo `json:"finished_quests"`  // 已完成已领奖的任务
}

// AcceptQuestRequest 接取任务请求
type AcceptQuestRequest struct {
	RoleID  uint64 `json:"role_id" binding:"required"`
	QuestID uint32 `json:"quest_id" binding:"required"`
}

// CompleteQuestRequest 完成任务请求
type CompleteQuestRequest struct {
	RoleID  uint64 `json:"role_id" binding:"required"`
	QuestID uint32 `json:"quest_id" binding:"required"`
}

// AbandonQuestRequest 放弃任务请求
type AbandonQuestRequest struct {
	RoleID  uint64 `json:"role_id" binding:"required"`
	QuestID uint32 `json:"quest_id" binding:"required"`
}

// UpdateProgressRequest 更新进度请求
type UpdateProgressRequest struct {
	RoleID     uint64 `json:"role_id" binding:"required"`
	QuestID    uint32 `json:"quest_id" binding:"required"`
	TargetType uint8  `json:"target_type"`     // 目标类型
	TargetID   uint32 `json:"target_id"`       // 目标ID
	Count      uint32 `json:"count,omitempty"` // 本次增加的进度数量（默认1）
}

// QuestProgressUpdate 进度更新通知（推送给前端）
type QuestProgressUpdate struct {
	QuestID  uint32 `json:"quest_id"` // 任务ID
	Progress uint32 `json:"progress"` // 新进度值
	Status   uint8  `json:"status"`   // 新状态（可选，如果完成了会变成2）
}
