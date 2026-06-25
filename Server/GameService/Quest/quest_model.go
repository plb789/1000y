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

// QuestTypeIcon 任务类型图标映射
var QuestTypeIcon = map[uint8]string{
	1: "📜",
	2: "📋",
	3: "⭐",
	4: "🏆",
	5: "🎉",
}

// TargetTypeName 目标类型名称映射
var TargetTypeName = map[uint8]string{
	1: "击杀",
	2: "采集",
	3: "对话",
	4: "探索",
}

// TargetTypeIcon 目标类型图标映射
var TargetTypeIcon = map[uint8]string{
	1: "⚔️",
	2: "🌿",
	3: "💬",
	4: "🚶",
}

// RoleQuestObjective 角色任务目标进度（多目标支持）
type RoleQuestObjective struct {
	ObjectiveID uint32 `gorm:"column:objective_id" json:"objective_id"`   // 目标ID
	Progress    uint32 `gorm:"column:progress;default:0" json:"progress"` // 当前进度
}

// RoleQuest 角色任务表（数据库持久化）
type RoleQuest struct {
	ID           uint64     `gorm:"primaryKey;column:id" json:"id"`                  // 记录ID
	RoleID       uint64     `gorm:"column:role_id;index" json:"role_id"`             // 角色ID
	QuestID      uint32     `gorm:"column:quest_id" json:"quest_id"`                 // 任务ID
	Status       uint8      `gorm:"column:status;default:0" json:"status"`           // 状态: 0=可接取, 1=进行中, 2=已完成, 3=已领奖
	Progress     uint32     `gorm:"column:progress;default:0" json:"progress"`       // 主目标进度（兼容单目标）
	Objectives   string     `gorm:"column:objectives;type:text" json:"objectives"`   // 多目标进度JSON（新增）
	AcceptTime   time.Time  `gorm:"column:accept_time" json:"accept_time"`           // 接取时间
	CompleteTime *time.Time `gorm:"column:complete_time" json:"complete_time"`       // 完成时间
	FinishTime   *time.Time `gorm:"column:finish_time" json:"finish_time"`           // 领奖时间
	DailyCount   int        `gorm:"column:daily_count;default:0" json:"daily_count"` // 今日完成次数（日常/周常用）
	TotalCount   int        `gorm:"column:total_count;default:0" json:"total_count"` // 总完成次数
}

func (RoleQuest) TableName() string {
	return "role_tasks"
}

// QuestObjectiveInfo 任务目标信息（返回给前端）
type QuestObjectiveInfo struct {
	ID          uint32 `json:"id"`           // 目标ID
	TargetType  uint8  `json:"target_type"`  // 目标类型: 1=击杀, 2=采集, 3=对话, 4=探索
	TargetID    uint32 `json:"target_id"`    // 目标ID
	TargetName  string `json:"target_name"`  // 目标名称
	TargetCount int    `json:"target_count"` // 目标数量
	Progress    uint32 `json:"progress"`     // 当前进度
}

// QuestInfo 任务详情（返回给前端的完整信息，包含配置+进度）
type QuestInfo struct {
	common.QuestBaseConfig
	Status         uint8                `json:"status"`          // 当前状态
	Progress       uint32               `json:"progress"`        // 主目标进度
	Objectives     []QuestObjectiveInfo `json:"objectives"`      // 多目标进度列表（新增）
	DailyCount     int                  `json:"daily_count"`     // 今日完成次数
	TotalCount     int                  `json:"total_count"`     // 总完成次数
	ChainCompleted bool                 `json:"chain_completed"` // 任务链已完成（触发链奖励时返回）
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
	RoleID      uint64 `json:"role_id" binding:"required"`
	QuestID     uint32 `json:"quest_id" binding:"required"`
	TargetType  uint8  `json:"target_type"`            // 目标类型
	TargetID    uint32 `json:"target_id"`              // 目标ID
	ObjectiveID uint32 `json:"objective_id,omitempty"` // 目标ID（多目标任务）
	Count       uint32 `json:"count,omitempty"`        // 本次增加的进度数量（默认1）
}

// QuestProgressUpdate 进度更新通知（推送给前端）
type QuestProgressUpdate struct {
	QuestID    uint32               `json:"quest_id"`   // 任务ID
	Progress   uint32               `json:"progress"`   // 主目标进度
	Objectives []QuestObjectiveInfo `json:"objectives"` // 多目标进度列表（新增）
	Status     uint8                `json:"status"`     // 新状态（可选，如果完成了会变成2）
}
