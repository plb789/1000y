package achievement

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	Common "game-server/Common"
)

// Service 成就服务
type Service struct {
	mu sync.RWMutex
	// 玩家成就缓存: roleID -> []RoleAchievement
	roleAchievements map[uint64][]RoleAchievement
	// 已加载的玩家ID集合
	loadedRoles map[uint64]bool
}

// 全局成就服务实例
var (
	globalAchievementService *Service
	achievementOnce          sync.Once
)

// GetAchievementService 获取成就服务单例
func GetAchievementService() *Service {
	achievementOnce.Do(func() {
		globalAchievementService = &Service{
			roleAchievements: make(map[uint64][]RoleAchievement),
			loadedRoles:      make(map[uint64]bool),
		}
	})
	return globalAchievementService
}

// InitAchievementService 初始化成就服务（启动定时保存）
func InitAchievementService() {
	svc := GetAchievementService()

	// 启动定时保存（每5分钟保存一次脏数据）
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			svc.saveAllDirty()
		}
	}()

	log.Printf("[ACHIEVEMENT] 成就服务初始化完成")
}

// 状态常量
const (
	AchievementStatusLocked   = false // 未解锁
	AchievementStatusUnlocked = true  // 已解锁
)

// OnQuestCompleted 任务完成时调用，检查并解锁相关成就
func (s *Service) OnQuestCompleted(roleID uint64, questID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保玩家数据已加载
	if !s.loadedRoles[roleID] {
		s.loadRoleAchievements(roleID)
	}

	// 查找所有与该任务相关的成就
	configs := Common.GetAllAchievementConfigs()
	for _, cfg := range configs {
		if cfg.Condition != "quest_complete" {
			continue
		}
		if cfg.TargetID != questID {
			continue
		}

		// 更新进度
		s.updateProgressLocked(roleID, cfg.ID, cfg.TargetCount)
	}
}

// OnMonsterKilled 怪物击杀时调用
func (s *Service) OnMonsterKilled(roleID uint64, monsterID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.loadedRoles[roleID] {
		s.loadRoleAchievements(roleID)
	}

	configs := Common.GetAllAchievementConfigs()
	for _, cfg := range configs {
		if cfg.Condition != "monster_kill" {
			continue
		}
		if cfg.TargetID != monsterID {
			continue
		}

		s.updateProgressLocked(roleID, cfg.ID, cfg.TargetCount)
	}
}

// OnItemCollected 物品采集时调用
func (s *Service) OnItemCollected(roleID uint64, itemID uint32, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.loadedRoles[roleID] {
		s.loadRoleAchievements(roleID)
	}

	configs := Common.GetAllAchievementConfigs()
	for _, cfg := range configs {
		if cfg.Condition != "item_collect" {
			continue
		}
		if cfg.TargetID != itemID {
			continue
		}

		s.updateProgressLocked(roleID, cfg.ID, cfg.TargetCount)
	}
}

// updateProgressLocked 更新进度（需要在持有锁时调用）
func (s *Service) updateProgressLocked(roleID uint64, achievementID uint32, targetCount int) {
	records := s.roleAchievements[roleID]
	var record *RoleAchievement
	var idx int = -1

	for i, r := range records {
		if r.AchievementID == achievementID {
			record = &records[i]
			idx = i
			break
		}
	}

	config := Common.GetAchievementConfig(achievementID)
	if config == nil {
		return
	}

	if record == nil {
		// 创建新记录
		record = &RoleAchievement{
			RoleID:        roleID,
			AchievementID: achievementID,
			Progress:      0,
			Unlocked:      false,
		}
		s.roleAchievements[roleID] = append(records, *record)
		idx = len(s.roleAchievements[roleID]) - 1
	}

	if record.Unlocked {
		return // 已解锁的不再处理
	}

	// 更新进度
	record.Progress++
	if record.Progress >= targetCount {
		record.Unlocked = true
		record.UnlockTime = time.Now().Unix()
		log.Printf("[ACHIEVEMENT] 玩家[%d]解锁成就[%d]: %s", roleID, achievementID, config.Name)

		// 发放成就奖励
		s.grantReward(roleID, config)

		// 推送成就解锁通知
		s.pushAchievementUnlocked(roleID, config)
	}

	// 更新缓存
	s.roleAchievements[roleID][idx] = *record
}

// grantReward 发放成就奖励
func (s *Service) grantReward(roleID uint64, config *Common.AchievementConfig) {
	if config.RewardExp > 0 {
		Common.DBRoleAddExp(roleID, int64(config.RewardExp))
		log.Printf("[ACHIEVEMENT] 发放经验: roleID=%d, exp=%d", roleID, config.RewardExp)
	}
	if config.RewardGold > 0 {
		Common.DBRoleAddGold(roleID, int64(config.RewardGold))
		log.Printf("[ACHIEVEMENT] 发放金币: roleID=%d, gold=%d", roleID, config.RewardGold)
	}
	if config.RewardHonor > 0 {
		Common.DBRoleAddHonor(roleID, int64(config.RewardHonor))
		log.Printf("[ACHIEVEMENT] 发放声望: roleID=%d, honor=%d", roleID, config.RewardHonor)
	}
}

// pushAchievementUnlocked 推送成就解锁通知
func (s *Service) pushAchievementUnlocked(roleID uint64, config *Common.AchievementConfig) {
	pushData := AchievementPushData{
		Type:   "achievement_unlocked",
		RoleID: roleID,
		Data: map[string]interface{}{
			"achievement_id": config.ID,
			"name":           config.Name,
			"description":    config.Description,
			"icon":           config.Icon,
			"point":          config.Point,
			"reward_exp":     config.RewardExp,
			"reward_gold":    config.RewardGold,
			"reward_honor":   config.RewardHonor,
		},
	}
	Common.GlobalMessageBus.Publish("achievement.push", pushData)
}

// loadRoleAchievements 从数据库加载玩家成就数据
func (s *Service) loadRoleAchievements(roleID uint64) error {
	s.loadedRoles[roleID] = true

	jsonStr, err := Common.DBGetRoleAchievements(roleID)
	if err != nil {
		log.Printf("[ACHIEVEMENT] 加载玩家成就数据失败: %v", err)
		return err
	}

	if jsonStr == "" {
		s.roleAchievements[roleID] = []RoleAchievement{}
		return nil
	}

	var records []RoleAchievement
	if err := json.Unmarshal([]byte(jsonStr), &records); err != nil {
		log.Printf("[ACHIEVEMENT] 解析玩家成就数据失败: %v", err)
		s.roleAchievements[roleID] = []RoleAchievement{}
		return err
	}

	s.roleAchievements[roleID] = records
	return nil
}

// saveAllDirty 保存所有脏数据
func (s *Service) saveAllDirty() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for roleID, records := range s.roleAchievements {
		jsonStr, err := json.Marshal(records)
		if err != nil {
			log.Printf("[ACHIEVEMENT] 序列化成就数据失败: roleID=%d, err=%v", roleID, err)
			continue
		}

		if err := Common.DBSaveRoleAchievements(roleID, string(jsonStr)); err != nil {
			log.Printf("[ACHIEVEMENT] 保存成就数据失败: roleID=%d, err=%v", roleID, err)
		}
	}
}

// SaveRoleAchievements 保存指定玩家的成就数据
func (s *Service) SaveRoleAchievements(roleID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.roleAchievements[roleID]
	jsonStr, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("序列化成就数据失败: %w", err)
	}

	return Common.DBSaveRoleAchievements(roleID, string(jsonStr))
}

// GetRoleAchievements 获取玩家所有成就信息
func (s *Service) GetRoleAchievements(roleID uint64) ([]AchievementInfo, error) {
	s.mu.Lock()
	if !s.loadedRoles[roleID] {
		s.loadRoleAchievements(roleID)
	}
	s.mu.Unlock()

	records := s.roleAchievements[roleID]
	allConfigs := Common.GetAllAchievementConfigs()

	// 构建已解锁ID集合
	unlockedMap := make(map[uint32]RoleAchievement)
	for _, r := range records {
		unlockedMap[r.AchievementID] = r
	}

	var infos []AchievementInfo
	for _, cfg := range allConfigs {
		info := AchievementInfo{
			AchievementConfig: *cfg,
			Unlocked:          false,
			Progress:          0,
		}

		if r, ok := unlockedMap[cfg.ID]; ok {
			info.Unlocked = r.Unlocked
			info.Progress = r.Progress
		}

		infos = append(infos, info)
	}

	return infos, nil
}

// GetAchievementStats 获取玩家成就统计
func (s *Service) GetAchievementStats(roleID uint64) (*AchievementStats, error) {
	infos, err := s.GetRoleAchievements(roleID)
	if err != nil {
		return nil, err
	}

	stats := &AchievementStats{
		TotalAchievements: len(infos),
		ByType:            []AchievementTypeInfo{},
	}

	typeNameMap := map[uint8]string{
		1: "任务成就",
		2: "战斗成就",
		3: "收集成就",
		4: "探索成就",
		5: "社交成就",
	}

	typeStats := make(map[uint8]*AchievementTypeInfo)
	for _, info := range infos {
		stats.TotalPoints += info.Point
		if info.Unlocked {
			stats.UnlockedCount++
		}

		if _, ok := typeStats[info.Type]; !ok {
			typeStats[info.Type] = &AchievementTypeInfo{
				Type:     info.Type,
				Name:     typeNameMap[info.Type],
				Total:    0,
				Unlocked: 0,
				Progress: 0,
			}
		}
		typeStats[info.Type].Total++
		if info.Unlocked {
			typeStats[info.Type].Unlocked++
		}
	}

	for _, ts := range typeStats {
		stats.ByType = append(stats.ByType, *ts)
	}

	return stats, nil
}

// AchievementPushData 成就推送数据
type AchievementPushData struct {
	Type   string                 `json:"type"`
	RoleID uint64                 `json:"role_id"`
	Data   map[string]interface{} `json:"data"`
}
