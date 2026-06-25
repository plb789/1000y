package achievement

import common "game-server/Common"

// RoleAchievement 玩家成就进度记录
type RoleAchievement struct {
	ID            uint64 `json:"id"`             // 记录ID
	RoleID        uint64 `json:"role_id"`        // 玩家ID
	AchievementID uint32 `json:"achievement_id"` // 成就ID
	Progress      int    `json:"progress"`       // 当前进度
	Unlocked      bool   `json:"unlocked"`       // 是否已解锁
	UnlockTime    int64  `json:"unlock_time"`    // 解锁时间戳
}

// AchievementInfo 成就详情（返回给前端）
type AchievementInfo struct {
	common.AchievementConfig
	Progress int    `json:"progress"` // 当前进度
	Unlocked bool   `json:"unlocked"` // 是否已解锁
	Point    uint32 `json:"point"`    // 成就点数
}

// AchievementTypeInfo 成就类型统计
type AchievementTypeInfo struct {
	Type     uint8  `json:"type"`     // 成就类型
	Name     string `json:"name"`     // 类型名称
	Total    int    `json:"total"`    // 该类型总数
	Unlocked int    `json:"unlocked"` // 已解锁数
	Progress int    `json:"progress"` // 总进度
}

// AchievementStats 成就系统统计
type AchievementStats struct {
	TotalAchievements int                   `json:"total_achievements"` // 成就总数
	UnlockedCount     int                   `json:"unlocked_count"`     // 已解锁数
	TotalPoints       uint32                `json:"total_points"`       // 成就总点数
	ByType            []AchievementTypeInfo `json:"by_type"`            // 按类型统计
}
