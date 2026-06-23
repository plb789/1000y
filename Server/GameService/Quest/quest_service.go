package quest

import (
	"errors"
	"log"
	"sync"
	"time"

	common "game-server/Common"
)

// Service 任务服务
type Service struct {
	mu sync.RWMutex
	// 内存缓存：角色ID -> 角色的所有任务记录（运行时加速查询）
	roleQuestCache map[uint64][]*RoleQuest
}

// NewService 创建任务服务实例
func NewService() *Service {
	return &Service{
		roleQuestCache: make(map[uint64][]*RoleQuest),
	}
}

// GetQuestConfig 获取任务配置
func GetQuestConfig(questID uint32) *common.QuestBaseConfig {
	return common.GetQuestConfig(questID)
}

// GetQuestList 获取角色的任务列表（核心方法）
func (s *Service) GetQuestList(roleID uint64, playerLevel uint32) (*QuestListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeQuests []QuestInfo
	var completedQuests []QuestInfo
	var availableQuests []QuestInfo
	var finishedQuests []QuestInfo

	// 1. 获取该角色已有的任务记录
	existingQuests := s.roleQuestCache[roleID]

	// 2. 遍历所有任务配置，分类处理
	for i := range common.GameConfig.Quests {
		config := &common.GameConfig.Quests[i]

		// 检查等级要求
		if config.LevelReq > playerLevel {
			continue // 等级不足，不显示
		}

		// 查找该任务在角色记录中的状态
		record := s.findQuestRecord(existingQuests, config.ID)

		if record == nil {
			// 无记录：可接取（前提是满足前置条件）
			if s.canAcceptQuest(roleID, config.ID) {
				availableQuests = append(availableQuests, QuestInfo{
					QuestBaseConfig: *config,
					Status:          QuestStatusAvailable,
					Progress:        0,
				})
			}
		} else {
			// 有记录：根据状态分类
			info := QuestInfo{
				QuestBaseConfig: *config,
				Status:          record.Status,
				Progress:        record.Progress,
				DailyCount:      record.DailyCount,
				TotalCount:      record.TotalCount,
			}

			switch record.Status {
			case QuestStatusActive:
				activeQuests = append(activeQuests, info)
			case QuestStatusCompleted:
				completedQuests = append(completedQuests, info)
			case QuestStatusFinished:
				finishedQuests = append(finishedQuests, info)
			default:
				// 状态异常，当作可接取处理
				availableQuests = append(availableQuests, info)
			}
		}
	}

	return &QuestListResponse{
		ActiveQuests:    activeQuests,
		CompletedQuests: completedQuests,
		AvailableQuests: availableQuests,
		FinishedQuests:  finishedQuests,
	}, nil
}

// AcceptQuest 接取任务
func (s *Service) AcceptQuest(roleID uint64, questID uint32, playerLevel uint32) (*QuestInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 检查任务配置是否存在
	config := GetQuestConfig(questID)
	if config == nil {
		return nil, errors.New("任务不存在")
	}

	// 2. 检查等级要求
	if config.LevelReq > playerLevel {
		return nil, errors.New("等级不足")
	}

	// 3. 检查是否已经接取或完成
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record != nil {
		switch record.Status {
		case QuestStatusActive:
			return nil, errors.New("任务已在进行中")
		case QuestStatusCompleted:
			return nil, errors.New("任务已完成，请领取奖励")
		case QuestStatusFinished:
			// 已完成的可重复任务，可以重新接取
			if config.Repeatable == 1 {
				// 重置进度和状态
				record.Status = QuestStatusActive
				record.Progress = 0
				record.AcceptTime = record.AcceptTime // 更新接取时间
				log.Printf("[QUEST] 玩家[%d]重新接取重复任务[%d]: %s", roleID, questID, config.Name)

				return &QuestInfo{
					QuestBaseConfig: *config,
					Status:          QuestStatusActive,
					Progress:        0,
					DailyCount:      record.DailyCount,
					TotalCount:      record.TotalCount,
				}, nil
			}
			return nil, errors.New("任务已完成且不可重复")
		}
	}

	// 4. 检查是否可接取（前置条件等）
	if !s.canAcceptQuestUnlocked(roleID, questID) {
		return nil, errors.New("前置条件未满足")
	}

	// 5. 创建新的任务记录
	newRecord := &RoleQuest{
		RoleID:     roleID,
		QuestID:    questID,
		Status:     QuestStatusActive,
		Progress:   0,
		AcceptTime: time.Now(), // 使用当前时间
		DailyCount: 0,
		TotalCount: 0,
	}

	// 6. 添加到缓存
	s.roleQuestCache[roleID] = append(s.roleQuestCache[roleID], newRecord)

	log.Printf("[QUEST] 玩家[%d]接取任务[%d]: %s", roleID, questID, config.Name)

	return &QuestInfo{
		QuestBaseConfig: *config,
		Status:          QuestStatusActive,
		Progress:        0,
	}, nil
}

// CompleteQuest 完成任务（领取奖励）
func (s *Service) CompleteQuest(roleID uint64, questID uint32) (*QuestInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 查找任务记录
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record == nil {
		return nil, errors.New("未找到任务记录")
	}

	// 2. 检查状态必须是"已完成未领奖"(status=2)
	if record.Status != QuestStatusCompleted {
		return nil, errors.New("任务尚未完成")
	}

	config := GetQuestConfig(questID)
	if config == nil {
		return nil, errors.New("任务配置不存在")
	}

	// 3. 更新状态为已领奖
	now := time.Now()
	record.Status = QuestStatusFinished
	record.FinishTime = &now
	record.DailyCount++
	record.TotalCount++

	log.Printf("[QUEST] 玩家[%d]完成任务[%d]领取奖励: %s (+%d经验, +%d金币)",
		roleID, questID, config.Name, config.RewardExp, config.RewardGold)

	return &QuestInfo{
		QuestBaseConfig: *config,
		Status:          QuestStatusFinished,
		Progress:        record.Progress,
		DailyCount:      record.DailyCount,
		TotalCount:      record.TotalCount,
	}, nil
}

// AbandonQuest 放弃任务
func (s *Service) AbandonQuest(roleID uint64, questID uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingQuests := s.roleQuestCache[roleID]
	idx := -1
	for i, q := range existingQuests {
		if q.QuestID == questID && q.Status == QuestStatusActive {
			idx = i
			break
		}
	}

	if idx == -1 {
		return errors.New("未找到进行中的任务")
	}

	// 从缓存中移除（或标记为放弃）
	s.roleQuestCache[roleID] = append(existingQuests[:idx], existingQuests[idx+1:]...)

	config := GetQuestConfig(questID)
	name := "未知任务"
	if config != nil {
		name = config.Name
	}

	log.Printf("[QUEST] 玩家[%d]放弃任务[%d]: %s", roleID, questID, name)
	return nil
}

// UpdateProgress 更新任务进度（击杀/采集/对话/探索）
func (s *Service) UpdateProgress(roleID uint64, questID uint32, targetType uint8, targetID uint32, count uint32) (*QuestProgressUpdate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 查找进行中的任务
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record == nil || record.Status != QuestStatusActive {
		return nil, errors.New("未找到进行中的任务")
	}

	// 2. 检查目标类型和ID是否匹配
	config := GetQuestConfig(questID)
	if config == nil {
		return nil, errors.New("任务配置不存在")
	}

	if config.TargetType != targetType || config.TargetID != targetID {
		return nil, errors.New("目标类型或ID不匹配")
	}

	// 3. 更新进度
	oldProgress := record.Progress
	record.Progress += count

	// 4. 检查是否完成
	if record.Progress >= uint32(config.TargetCount) {
		record.Progress = uint32(config.TargetCount) // 不超过目标值
		record.Status = QuestStatusCompleted
		now := time.Now()
		record.CompleteTime = &now

		log.Printf("[QUEST] 玩家[%d]任务[%d]%s已完成! 进度: %d/%d",
			roleID, questID, config.Name, record.Progress, config.TargetCount)
	} else {
		log.Printf("[QUEST] 玩家[%d]任务[%d]%s进度更新: %d→%d/%d",
			roleID, questID, config.Name, oldProgress, record.Progress, config.TargetCount)
	}

	return &QuestProgressUpdate{
		QuestID:  questID,
		Progress: record.Progress,
		Status:   record.Status,
	}, nil
}

// OnMonsterKilled 怪物被击杀时调用（自动更新所有匹配的击杀类任务）
func (s *Service) OnMonsterKilled(roleID uint64, monsterID uint32) []*QuestProgressUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	var updates []*QuestProgressUpdate

	existingQuests := s.roleQuestCache[roleID]
	for _, record := range existingQuests {
		if record.Status != QuestStatusActive {
			continue
		}

		config := GetQuestConfig(record.QuestID)
		if config == nil || config.TargetType != 1 || config.TargetID != monsterID {
			continue // 不是击杀类任务或不匹配怪物ID
		}

		// 更新进度
		record.Progress++
		if record.Progress >= uint32(config.TargetCount) {
			record.Progress = uint32(config.TargetCount)
			record.Status = QuestStatusCompleted
			now := time.Now()
			record.CompleteTime = &now

			log.Printf("[QUEST] 自动更新-玩家[%d]任务[%d]%s完成! 击杀怪物[%d]",
				roleID, record.QuestID, config.Name, monsterID)
		}

		updates = append(updates, &QuestProgressUpdate{
			QuestID:  record.QuestID,
			Progress: record.Progress,
			Status:   record.Status,
		})
	}

	return updates
}

// OnItemGathered 物品采集时调用（自动更新所有匹配的采集类任务）
func (s *Service) OnItemGathered(roleID uint64, itemID uint32) []*QuestProgressUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	var updates []*QuestProgressUpdate

	existingQuests := s.roleQuestCache[roleID]
	for _, record := range existingQuests {
		if record.Status != QuestStatusActive {
			continue
		}

		config := GetQuestConfig(record.QuestID)
		if config == nil || config.TargetType != 2 || config.TargetID != itemID {
			continue
		}

		record.Progress++
		if record.Progress >= uint32(config.TargetCount) {
			record.Progress = uint32(config.TargetCount)
			record.Status = QuestStatusCompleted
			now := time.Now()
			record.CompleteTime = &now

			log.Printf("[QUEST] 自动更新-玩家[%d]任务[%d]%s完成! 采集物品[%d]",
				roleID, record.QuestID, config.Name, itemID)
		}

		updates = append(updates, &QuestProgressUpdate{
			QuestID:  record.QuestID,
			Progress: record.Progress,
			Status:   record.Status,
		})
	}

	return updates
}

// canAcceptQuest 检查是否可接取任务（带锁版本）
func (s *Service) canAcceptQuest(roleID uint64, questID uint32) bool {
	return s.canAcceptQuestUnlocked(roleID, questID)
}

// canAcceptQuestUnlocked 检查是否可接取任务（无锁版本，内部使用）
func (s *Service) canAcceptQuestUnlocked(roleID uint64, questID uint32) bool {
	// TODO: 后续可添加前置任务检查、每日次数限制等
	// 目前简单实现：只要没有进行中或已完成的记录就可以接取
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record == nil {
		return true // 无记录，可接取
	}

	// 已完成的可重复任务可以重新接取
	if record.Status == QuestStatusFinished {
		config := GetQuestConfig(questID)
		if config != nil && config.Repeatable == 1 {
			return true
		}
	}

	return false
}

// findQuestRecord 在任务列表中查找指定任务的记录
func (s *Service) findQuestRecord(quests []*RoleQuest, questID uint32) *RoleQuest {
	for _, q := range quests {
		if q.QuestID == questID {
			return q
		}
	}
	return nil
}

// ResetDailyQuests 重置日常任务（每天凌晨调用）
func (s *Service) ResetDailyQuests() {
	s.mu.Lock()
	defer s.mu.Unlock()

	resetCount := 0
	for _, quests := range s.roleQuestCache {
		for _, quest := range quests {
			config := GetQuestConfig(quest.QuestID)
			if config != nil && config.Type == 3 && quest.DailyCount > 0 { // 日常任务
				if quest.Status == QuestStatusFinished {
					quest.Status = QuestStatusAvailable
					quest.Progress = 0
					quest.DailyCount = 0
					resetCount++
				}
			}
		}
	}

	if resetCount > 0 {
		log.Printf("[QUEST] 重置了%d个日常任务", resetCount)
	}
}
