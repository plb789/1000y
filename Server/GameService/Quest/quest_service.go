package quest

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	common "game-server/Common"
	achievement "game-server/GameService/Achievement"
)

// Service 任务服务
type Service struct {
	mu sync.RWMutex
	// 内存缓存：角色ID -> 角色的所有任务记录（运行时加速查询）
	roleQuestCache map[uint64][]*RoleQuest
	stopCh         chan struct{} // 停止定时器的通道
}

// NewService 创建任务服务实例
func NewService() *Service {
	s := &Service{
		roleQuestCache: make(map[uint64][]*RoleQuest),
		stopCh:         make(chan struct{}),
	}
	// 启动任务重置定时器
	go s.startResetTicker()
	return s
}

// Stop 停止服务（用于优雅关闭）
func (s *Service) Stop() {
	close(s.stopCh)
}

// GetQuestConfig 获取任务配置
func GetQuestConfig(questID uint32) *common.QuestBaseConfig {
	return common.GetQuestConfig(questID)
}

// buildObjectiveInfos 构建目标进度信息列表
func buildObjectiveInfos(config *common.QuestBaseConfig, objectiveProgress map[uint32]uint32) []QuestObjectiveInfo {
	var infos []QuestObjectiveInfo

	if len(config.Objectives) > 0 {
		// 多目标任务
		for _, obj := range config.Objectives {
			progress := objectiveProgress[obj.ID]
			infos = append(infos, QuestObjectiveInfo{
				ID:          obj.ID,
				TargetType:  obj.TargetType,
				TargetID:    obj.TargetID,
				TargetName:  obj.TargetName,
				TargetCount: obj.TargetCount,
				Progress:    progress,
			})
		}
	} else {
		// 单目标任务（兼容）
		progress := objectiveProgress[1]
		if progress == 0 {
			progress = objectiveProgress[0] // 可能用0作为默认目标ID
		}
		infos = append(infos, QuestObjectiveInfo{
			ID:          1,
			TargetType:  config.TargetType,
			TargetID:    config.TargetID,
			TargetName:  config.TargetName,
			TargetCount: config.TargetCount,
			Progress:    progress,
		})
	}

	return infos
}

// parseObjectivesJSON 解析目标进度JSON
func parseObjectivesJSON(jsonStr string) map[uint32]uint32 {
	if jsonStr == "" {
		return make(map[uint32]uint32)
	}

	var objectives []RoleQuestObjective
	if err := json.Unmarshal([]byte(jsonStr), &objectives); err != nil {
		log.Printf("[QUEST] 解析目标进度JSON失败: %v", err)
		return make(map[uint32]uint32)
	}

	result := make(map[uint32]uint32)
	for _, obj := range objectives {
		result[obj.ObjectiveID] = obj.Progress
	}
	return result
}

// serializeObjectivesJSON 序列化目标进度JSON
func serializeObjectivesJSON(objectives map[uint32]uint32) string {
	if len(objectives) == 0 {
		return ""
	}

	var list []RoleQuestObjective
	for id, progress := range objectives {
		list = append(list, RoleQuestObjective{
			ObjectiveID: id,
			Progress:    progress,
		})
	}

	data, err := json.Marshal(list)
	if err != nil {
		log.Printf("[QUEST] 序列化目标进度JSON失败: %v", err)
		return ""
	}
	return string(data)
}

// GetQuestList 获取角色的任务列表（核心方法）
func (s *Service) GetQuestList(roleID uint64, playerLevel uint32) (*QuestListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var activeQuests []QuestInfo
	var completedQuests []QuestInfo
	var availableQuests []QuestInfo
	var finishedQuests []QuestInfo

	// 1. 尝试从数据库加载任务数据
	if dbTasks, err := common.DBTaskGetList(roleID); err == nil && len(dbTasks) > 0 {
		// 数据库有数据，更新缓存
		for _, dbTask := range dbTasks {
			record := &RoleQuest{
				RoleID:     dbTask.RoleID,
				QuestID:    dbTask.TaskID,
				Status:     dbTask.Status,
				Progress:   dbTask.Progress,
				Objectives: dbTask.Objectives,
				DailyCount: dbTask.DailyCount,
				TotalCount: dbTask.TotalCount,
			}
			// 解析时间字符串
			if dbTask.AcceptTime != "" {
				if t, err := time.Parse("2006-01-02T15:04:05Z", dbTask.AcceptTime); err == nil {
					record.AcceptTime = t
				}
			}
			if dbTask.CompleteTime != nil && *dbTask.CompleteTime != "" {
				if t, err := time.Parse("2006-01-02T15:04:05Z", *dbTask.CompleteTime); err == nil {
					record.CompleteTime = &t
				}
			}
			if dbTask.FinishTime != nil && *dbTask.FinishTime != "" {
				if t, err := time.Parse("2006-01-02T15:04:05Z", *dbTask.FinishTime); err == nil {
					record.FinishTime = &t
				}
			}
			s.roleQuestCache[roleID] = append(s.roleQuestCache[roleID], record)
		}
		log.Printf("[QUEST] 从数据库加载玩家[%d]任务数据: %d条", roleID, len(dbTasks))
	}

	// 2. 获取该角色已有的任务记录
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
			// 无记录：检查是否可接取
			if s.canAcceptQuest(roleID, config.ID, playerLevel) {
				// 构建初始目标进度（全0）
				objectiveInfos := buildObjectiveInfos(config, make(map[uint32]uint32))
				availableQuests = append(availableQuests, QuestInfo{
					QuestBaseConfig: *config,
					Status:          QuestStatusAvailable,
					Progress:        0,
					Objectives:      objectiveInfos,
				})
			}
		} else {
			// 有记录：根据状态分类
			objectiveProgress := parseObjectivesJSON(record.Objectives)
			objectiveInfos := buildObjectiveInfos(config, objectiveProgress)

			info := QuestInfo{
				QuestBaseConfig: *config,
				Status:          record.Status,
				Progress:        record.Progress,
				Objectives:      objectiveInfos,
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
				record.Objectives = ""
				record.AcceptTime = time.Now()

				// 保存到数据库
				common.DBTaskSave(common.RoleTaskInfo{
					RoleID:     record.RoleID,
					TaskID:     record.QuestID,
					Status:     record.Status,
					Progress:   record.Progress,
					Objectives: record.Objectives,
					AcceptTime: record.AcceptTime.Format(time.RFC3339),
				})

				objectiveInfos := buildObjectiveInfos(config, make(map[uint32]uint32))
				log.Printf("[QUEST] 玩家[%d]重新接取重复任务[%d]: %s", roleID, questID, config.Name)

				return &QuestInfo{
					QuestBaseConfig: *config,
					Status:          QuestStatusActive,
					Progress:        0,
					Objectives:      objectiveInfos,
					DailyCount:      record.DailyCount,
					TotalCount:      record.TotalCount,
				}, nil
			}
			return nil, errors.New("任务已完成且不可重复")
		}
	}

	// 4. 检查是否可接取（前置条件等）
	if !s.canAcceptQuestUnlocked(roleID, questID, playerLevel) {
		return nil, errors.New("前置条件未满足")
	}

	// 5. 创建新的任务记录
	now := time.Now()
	newRecord := &RoleQuest{
		RoleID:     roleID,
		QuestID:    questID,
		Status:     QuestStatusActive,
		Progress:   0,
		Objectives: "",
		AcceptTime: now,
		DailyCount: 0,
		TotalCount: 0,
	}

	// 6. 添加到缓存
	s.roleQuestCache[roleID] = append(s.roleQuestCache[roleID], newRecord)

	// 7. 保存到数据库
	common.DBTaskSave(common.RoleTaskInfo{
		RoleID:     newRecord.RoleID,
		TaskID:     newRecord.QuestID,
		Status:     newRecord.Status,
		Progress:   newRecord.Progress,
		Objectives: newRecord.Objectives,
		AcceptTime: newRecord.AcceptTime.Format(time.RFC3339),
	})

	objectiveInfos := buildObjectiveInfos(config, make(map[uint32]uint32))
	log.Printf("[QUEST] 玩家[%d]接取任务[%d]: %s", roleID, questID, config.Name)

	return &QuestInfo{
		QuestBaseConfig: *config,
		Status:          QuestStatusActive,
		Progress:        0,
		Objectives:      objectiveInfos,
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

	// 保存到数据库
	finishTimeStr := record.FinishTime.Format(time.RFC3339)
	common.DBTaskSave(common.RoleTaskInfo{
		RoleID:     record.RoleID,
		TaskID:     record.QuestID,
		Status:     record.Status,
		Progress:   record.Progress,
		Objectives: record.Objectives,
		AcceptTime: record.AcceptTime.Format(time.RFC3339),
		FinishTime: &finishTimeStr,
		DailyCount: record.DailyCount,
		TotalCount: record.TotalCount,
	})

	// 4. 发放奖励（经验、金币、声望）
	if config.RewardExp > 0 {
		if _, level, _, err := common.DBRoleAddExp(roleID, int64(config.RewardExp)); err != nil {
			log.Printf("[QUEST] 发放经验奖励失败: %v", err)
		} else {
			log.Printf("[QUEST] 发放经验奖励: roleID=%d, exp=%d, level=%d", roleID, config.RewardExp, level)
		}
	}
	if config.RewardGold > 0 {
		if err := common.DBRoleAddGold(roleID, int64(config.RewardGold)); err != nil {
			log.Printf("[QUEST] 发放金币奖励失败: %v", err)
		} else {
			log.Printf("[QUEST] 发放金币奖励: roleID=%d, gold=%d", roleID, config.RewardGold)
		}
	}
	if config.RewardHonor > 0 {
		if err := common.DBRoleAddHonor(roleID, int64(config.RewardHonor)); err != nil {
			log.Printf("[QUEST] 发放声望奖励失败: %v", err)
		} else {
			log.Printf("[QUEST] 发放声望奖励: roleID=%d, honor=%d", roleID, config.RewardHonor)
		}
	}

	// 5. 检查任务链奖励
	s.checkAndGrantChainReward(roleID, config, record)

	// 6. 检查并解锁相关成就
	achievement.GetAchievementService().OnQuestCompleted(roleID, questID)

	objectiveProgress := parseObjectivesJSON(record.Objectives)
	objectiveInfos := buildObjectiveInfos(config, objectiveProgress)

	log.Printf("[QUEST] 玩家[%d]完成任务[%d]领取奖励: %s (+%d经验, +%d金币, +%d声望)",
		roleID, questID, config.Name, config.RewardExp, config.RewardGold, config.RewardHonor)

	return &QuestInfo{
		QuestBaseConfig: *config,
		Status:          QuestStatusFinished,
		Progress:        record.Progress,
		Objectives:      objectiveInfos,
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

	// 从缓存中移除
	s.roleQuestCache[roleID] = append(existingQuests[:idx], existingQuests[idx+1:]...)

	// 从数据库中删除
	if err := common.DBTaskAbandon(roleID, questID); err != nil {
		log.Printf("[QUEST] 玩家[%d]放弃任务[%d]数据库删除失败: %v", roleID, questID, err)
	}

	config := GetQuestConfig(questID)
	name := "未知任务"
	if config != nil {
		name = config.Name
	}

	log.Printf("[QUEST] 玩家[%d]放弃任务[%d]: %s", roleID, questID, name)
	return nil
}

// UpdateProgress 更新任务进度（支持多目标）
func (s *Service) UpdateProgress(roleID uint64, questID uint32, targetType uint8, targetID uint32, objectiveID uint32, count uint32) (*QuestProgressUpdate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 查找进行中的任务
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record == nil || record.Status != QuestStatusActive {
		return nil, errors.New("未找到进行中的任务")
	}

	// 2. 检查任务配置
	config := GetQuestConfig(questID)
	if config == nil {
		return nil, errors.New("任务配置不存在")
	}

	// 3. 更新进度
	objectiveProgress := parseObjectivesJSON(record.Objectives)

	if len(config.Objectives) > 0 {
		// 多目标任务
		if objectiveID > 0 {
			// 指定了目标ID，更新特定目标
			for _, obj := range config.Objectives {
				if obj.ID == objectiveID && obj.TargetType == targetType && (obj.TargetID == targetID || obj.TargetID == 0) {
					objectiveProgress[obj.ID] += count
					if objectiveProgress[obj.ID] > uint32(obj.TargetCount) {
						objectiveProgress[obj.ID] = uint32(obj.TargetCount)
					}
					break
				}
			}
		} else {
			// 未指定目标ID，匹配所有符合条件的目标
			for _, obj := range config.Objectives {
				if obj.TargetType == targetType && (obj.TargetID == targetID || obj.TargetID == 0 || targetID == 0) {
					objectiveProgress[obj.ID] += count
					if objectiveProgress[obj.ID] > uint32(obj.TargetCount) {
						objectiveProgress[obj.ID] = uint32(obj.TargetCount)
					}
				}
			}
		}
	} else {
		// 单目标任务（兼容）
		if config.TargetType == targetType && (config.TargetID == targetID || config.TargetID == 0 || targetID == 0) {
			record.Progress += count
			if record.Progress > uint32(config.TargetCount) {
				record.Progress = uint32(config.TargetCount)
			}
			objectiveProgress[1] = record.Progress
		}
	}

	// 序列化目标进度
	record.Objectives = serializeObjectivesJSON(objectiveProgress)

	// 计算总进度（用于显示）
	totalProgress := s.calculateTotalProgress(config, objectiveProgress)
	record.Progress = totalProgress

	// 4. 检查是否完成
	if s.isQuestCompleted(config, objectiveProgress, record.Progress) {
		record.Status = QuestStatusCompleted
		now := time.Now()
		record.CompleteTime = &now

		log.Printf("[QUEST] 玩家[%d]任务[%d]%s已完成!", roleID, questID, config.Name)
	} else {
		log.Printf("[QUEST] 玩家[%d]任务[%d]%s进度更新", roleID, questID, config.Name)
	}

	// 5. 保存到数据库
	var completeTimeStr *string
	if record.CompleteTime != nil {
		s := record.CompleteTime.Format(time.RFC3339)
		completeTimeStr = &s
	}
	common.DBTaskSave(common.RoleTaskInfo{
		RoleID:       record.RoleID,
		TaskID:       record.QuestID,
		Status:       record.Status,
		Progress:     record.Progress,
		Objectives:   record.Objectives,
		AcceptTime:   record.AcceptTime.Format(time.RFC3339),
		CompleteTime: completeTimeStr,
	})

	objectiveInfos := buildObjectiveInfos(config, objectiveProgress)
	return &QuestProgressUpdate{
		QuestID:    questID,
		Progress:   record.Progress,
		Objectives: objectiveInfos,
		Status:     record.Status,
	}, nil
}

// calculateTotalProgress 计算总进度百分比
func (s *Service) calculateTotalProgress(config *common.QuestBaseConfig, objectiveProgress map[uint32]uint32) uint32 {
	if len(config.Objectives) > 0 {
		// 多目标任务：计算总进度
		var totalCurrent, totalTarget uint32
		for _, obj := range config.Objectives {
			totalCurrent += objectiveProgress[obj.ID]
			totalTarget += uint32(obj.TargetCount)
		}
		if totalTarget == 0 {
			return 0
		}
		return (totalCurrent * 100) / totalTarget
	}

	// 单目标任务
	return objectiveProgress[1]
}

// isQuestCompleted 检查任务是否完成
func (s *Service) isQuestCompleted(config *common.QuestBaseConfig, objectiveProgress map[uint32]uint32, mainProgress uint32) bool {
	if len(config.Objectives) > 0 {
		// 多目标任务：所有目标都必须完成
		for _, obj := range config.Objectives {
			if objectiveProgress[obj.ID] < uint32(obj.TargetCount) {
				return false
			}
		}
		return true
	}

	// 单目标任务
	return mainProgress >= uint32(config.TargetCount)
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
		if config == nil {
			continue
		}

		// 检查是否匹配击杀目标
		matched := false
		objectiveProgress := parseObjectivesJSON(record.Objectives)

		if len(config.Objectives) > 0 {
			// 多目标任务
			for _, obj := range config.Objectives {
				if obj.TargetType == 1 && (obj.TargetID == monsterID || obj.TargetID == 0 || monsterID == 0) {
					objectiveProgress[obj.ID]++
					if objectiveProgress[obj.ID] > uint32(obj.TargetCount) {
						objectiveProgress[obj.ID] = uint32(obj.TargetCount)
					}
					matched = true
				}
			}
		} else {
			// 单目标任务
			if config.TargetType == 1 && (config.TargetID == monsterID || config.TargetID == 0 || monsterID == 0) {
				record.Progress++
				if record.Progress > uint32(config.TargetCount) {
					record.Progress = uint32(config.TargetCount)
				}
				objectiveProgress[1] = record.Progress
				matched = true
			}
		}

		if matched {
			record.Objectives = serializeObjectivesJSON(objectiveProgress)
			record.Progress = s.calculateTotalProgress(config, objectiveProgress)

			// 检查是否完成
			if s.isQuestCompleted(config, objectiveProgress, record.Progress) {
				record.Status = QuestStatusCompleted
				now := time.Now()
				record.CompleteTime = &now
				log.Printf("[QUEST] 自动更新-玩家[%d]任务[%d]%s完成! 击杀怪物[%d]",
					roleID, record.QuestID, config.Name, monsterID)
			}

			// 保存到数据库
			var completeTimeStr *string
			if record.CompleteTime != nil {
				t := record.CompleteTime.Format(time.RFC3339)
				completeTimeStr = &t
			}
			common.DBTaskSave(common.RoleTaskInfo{
				RoleID:       record.RoleID,
				TaskID:       record.QuestID,
				Status:       record.Status,
				Progress:     record.Progress,
				Objectives:   record.Objectives,
				AcceptTime:   record.AcceptTime.Format(time.RFC3339),
				CompleteTime: completeTimeStr,
			})

			objectiveInfos := buildObjectiveInfos(config, objectiveProgress)
			updates = append(updates, &QuestProgressUpdate{
				QuestID:    record.QuestID,
				Progress:   record.Progress,
				Objectives: objectiveInfos,
				Status:     record.Status,
			})
		}
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
		if config == nil {
			continue
		}

		matched := false
		objectiveProgress := parseObjectivesJSON(record.Objectives)

		if len(config.Objectives) > 0 {
			for _, obj := range config.Objectives {
				if obj.TargetType == 2 && (obj.TargetID == itemID || obj.TargetID == 0 || itemID == 0) {
					objectiveProgress[obj.ID]++
					if objectiveProgress[obj.ID] > uint32(obj.TargetCount) {
						objectiveProgress[obj.ID] = uint32(obj.TargetCount)
					}
					matched = true
				}
			}
		} else {
			if config.TargetType == 2 && (config.TargetID == itemID || config.TargetID == 0 || itemID == 0) {
				record.Progress++
				if record.Progress > uint32(config.TargetCount) {
					record.Progress = uint32(config.TargetCount)
				}
				objectiveProgress[1] = record.Progress
				matched = true
			}
		}

		if matched {
			record.Objectives = serializeObjectivesJSON(objectiveProgress)
			record.Progress = s.calculateTotalProgress(config, objectiveProgress)

			if s.isQuestCompleted(config, objectiveProgress, record.Progress) {
				record.Status = QuestStatusCompleted
				now := time.Now()
				record.CompleteTime = &now
				log.Printf("[QUEST] 自动更新-玩家[%d]任务[%d]%s完成! 采集物品[%d]",
					roleID, record.QuestID, config.Name, itemID)
			}

			// 保存到数据库
			var completeTimeStr *string
			if record.CompleteTime != nil {
				t := record.CompleteTime.Format(time.RFC3339)
				completeTimeStr = &t
			}
			common.DBTaskSave(common.RoleTaskInfo{
				RoleID:       record.RoleID,
				TaskID:       record.QuestID,
				Status:       record.Status,
				Progress:     record.Progress,
				Objectives:   record.Objectives,
				AcceptTime:   record.AcceptTime.Format(time.RFC3339),
				CompleteTime: completeTimeStr,
			})

			objectiveInfos := buildObjectiveInfos(config, objectiveProgress)
			updates = append(updates, &QuestProgressUpdate{
				QuestID:    record.QuestID,
				Progress:   record.Progress,
				Objectives: objectiveInfos,
				Status:     record.Status,
			})
		}
	}

	return updates
}

// OnNPCDialog NPC对话时调用（自动更新所有匹配的对话类任务）
func (s *Service) OnNPCDialog(roleID uint64, npcID uint32) []*QuestProgressUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()

	var updates []*QuestProgressUpdate

	existingQuests := s.roleQuestCache[roleID]
	for _, record := range existingQuests {
		if record.Status != QuestStatusActive {
			continue
		}

		config := GetQuestConfig(record.QuestID)
		if config == nil {
			continue
		}

		matched := false
		objectiveProgress := parseObjectivesJSON(record.Objectives)

		if len(config.Objectives) > 0 {
			for _, obj := range config.Objectives {
				if obj.TargetType == 3 && obj.TargetID == npcID {
					objectiveProgress[obj.ID] = uint32(obj.TargetCount) // 对话任务一次性完成
					matched = true
				}
			}
		} else {
			if config.TargetType == 3 && config.TargetID == npcID {
				record.Progress = uint32(config.TargetCount)
				objectiveProgress[1] = record.Progress
				matched = true
			}
		}

		if matched {
			record.Objectives = serializeObjectivesJSON(objectiveProgress)
			record.Progress = s.calculateTotalProgress(config, objectiveProgress)

			if s.isQuestCompleted(config, objectiveProgress, record.Progress) {
				record.Status = QuestStatusCompleted
				now := time.Now()
				record.CompleteTime = &now
				log.Printf("[QUEST] 自动更新-玩家[%d]任务[%d]%s完成! 对话NPC[%d]",
					roleID, record.QuestID, config.Name, npcID)
			}

			// 保存到数据库
			var completeTimeStr *string
			if record.CompleteTime != nil {
				t := record.CompleteTime.Format(time.RFC3339)
				completeTimeStr = &t
			}
			common.DBTaskSave(common.RoleTaskInfo{
				RoleID:       record.RoleID,
				TaskID:       record.QuestID,
				Status:       record.Status,
				Progress:     record.Progress,
				Objectives:   record.Objectives,
				AcceptTime:   record.AcceptTime.Format(time.RFC3339),
				CompleteTime: completeTimeStr,
			})

			objectiveInfos := buildObjectiveInfos(config, objectiveProgress)
			updates = append(updates, &QuestProgressUpdate{
				QuestID:    record.QuestID,
				Progress:   record.Progress,
				Objectives: objectiveInfos,
				Status:     record.Status,
			})
		}
	}

	return updates
}

// AutoAcceptQuests 自动接取符合条件的任务
func (s *Service) AutoAcceptQuests(roleID uint64, playerLevel uint32) []*QuestInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var accepted []*QuestInfo

	for i := range common.GameConfig.Quests {
		config := &common.GameConfig.Quests[i]

		// 检查是否需要自动接取
		if config.AutoAccept != 1 {
			continue
		}

		// 检查等级要求
		if config.LevelReq > playerLevel {
			continue
		}

		// 检查是否已经接取或完成
		existingQuests := s.roleQuestCache[roleID]
		record := s.findQuestRecord(existingQuests, config.ID)

		if record != nil {
			// 已有记录，检查是否可以重新接取
			if record.Status == QuestStatusFinished && config.Repeatable == 1 {
				// 可重复任务，重新接取
				record.Status = QuestStatusActive
				record.Progress = 0
				record.Objectives = ""
				record.AcceptTime = time.Now()

				objectiveInfos := buildObjectiveInfos(config, make(map[uint32]uint32))
				accepted = append(accepted, &QuestInfo{
					QuestBaseConfig: *config,
					Status:          QuestStatusActive,
					Progress:        0,
					Objectives:      objectiveInfos,
				})

				log.Printf("[QUEST] 自动接取-玩家[%d]任务[%d]: %s", roleID, config.ID, config.Name)
			}
			continue
		}

		// 检查前置任务
		if config.PrevQuestID != nil {
			prevRecord := s.findQuestRecord(existingQuests, *config.PrevQuestID)
			if prevRecord == nil || prevRecord.Status != QuestStatusFinished {
				continue // 前置任务未完成
			}
		}

		// 创建新的任务记录
		now := time.Now()
		newRecord := &RoleQuest{
			RoleID:     roleID,
			QuestID:    config.ID,
			Status:     QuestStatusActive,
			Progress:   0,
			Objectives: "",
			AcceptTime: now,
			DailyCount: 0,
			TotalCount: 0,
		}

		s.roleQuestCache[roleID] = append(s.roleQuestCache[roleID], newRecord)

		// 保存到数据库
		common.DBTaskSave(common.RoleTaskInfo{
			RoleID:     newRecord.RoleID,
			TaskID:     newRecord.QuestID,
			Status:     newRecord.Status,
			Progress:   newRecord.Progress,
			Objectives: newRecord.Objectives,
			AcceptTime: newRecord.AcceptTime.Format(time.RFC3339),
		})

		objectiveInfos := buildObjectiveInfos(config, make(map[uint32]uint32))
		accepted = append(accepted, &QuestInfo{
			QuestBaseConfig: *config,
			Status:          QuestStatusActive,
			Progress:        0,
			Objectives:      objectiveInfos,
		})

		log.Printf("[QUEST] 自动接取-玩家[%d]任务[%d]: %s", roleID, config.ID, config.Name)
	}

	return accepted
}

// canAcceptQuest 检查是否可接取任务（带锁版本）
func (s *Service) canAcceptQuest(roleID uint64, questID uint32, playerLevel uint32) bool {
	return s.canAcceptQuestUnlocked(roleID, questID, playerLevel)
}

// canAcceptQuestUnlocked 检查是否可接取任务（无锁版本，内部使用）
func (s *Service) canAcceptQuestUnlocked(roleID uint64, questID uint32, playerLevel uint32) bool {
	config := GetQuestConfig(questID)
	if config == nil {
		return false
	}

	// 检查等级要求
	if config.LevelReq > playerLevel {
		return false
	}

	// 检查前置任务
	if config.PrevQuestID != nil {
		existingQuests := s.roleQuestCache[roleID]
		prevRecord := s.findQuestRecord(existingQuests, *config.PrevQuestID)
		if prevRecord == nil || prevRecord.Status != QuestStatusFinished {
			return false // 前置任务未完成
		}
	}

	// 检查是否已有记录
	existingQuests := s.roleQuestCache[roleID]
	record := s.findQuestRecord(existingQuests, questID)

	if record == nil {
		return true // 无记录，可接取
	}

	// 已完成的可重复任务可以重新接取
	if record.Status == QuestStatusFinished && config.Repeatable == 1 {
		return true
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
	for roleID, quests := range s.roleQuestCache {
		for _, quest := range quests {
			config := GetQuestConfig(quest.QuestID)
			if config != nil && config.Type == 3 && quest.DailyCount > 0 { // 日常任务
				if quest.Status == QuestStatusFinished {
					quest.Status = QuestStatusAvailable
					quest.Progress = 0
					quest.Objectives = ""
					quest.DailyCount = 0
					resetCount++

					// 如果是自动接取的任务，重新接取
					if config.AutoAccept == 1 {
						quest.Status = QuestStatusActive
						quest.AcceptTime = time.Now()
						log.Printf("[QUEST] 重置日常任务并自动接取-玩家[%d]任务[%d]: %s",
							roleID, quest.QuestID, config.Name)
					}
				}
			}
		}
	}

	if resetCount > 0 {
		log.Printf("[QUEST] 重置了%d个日常任务", resetCount)
	}
}

// ResetWeeklyQuests 重置周常任务（每周一凌晨调用）
func (s *Service) ResetWeeklyQuests() {
	s.mu.Lock()
	defer s.mu.Unlock()

	resetCount := 0
	for roleID, quests := range s.roleQuestCache {
		for _, quest := range quests {
			config := GetQuestConfig(quest.QuestID)
			if config != nil && config.Type == 4 && quest.DailyCount > 0 { // 周常任务
				if quest.Status == QuestStatusFinished {
					quest.Status = QuestStatusAvailable
					quest.Progress = 0
					quest.Objectives = ""
					quest.DailyCount = 0
					resetCount++

					if config.AutoAccept == 1 {
						quest.Status = QuestStatusActive
						quest.AcceptTime = time.Now()
						log.Printf("[QUEST] 重置周常任务并自动接取-玩家[%d]任务[%d]: %s",
							roleID, quest.QuestID, config.Name)
					}
				}
			}
		}
	}

	if resetCount > 0 {
		log.Printf("[QUEST] 重置了%d个周常任务", resetCount)
	}
}

// startResetTicker 启动任务重置定时器
// - 每分钟检查是否需要重置日常任务（检查是否跨天）
// - 每周一检查是否需要重置周常任务（检查是否跨周）
func (s *Service) startResetTicker() {
	// 计算距离下一个凌晨0点的 Duration
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	dailyTicker := time.NewTicker(time.Until(nextMidnight))
	defer dailyTicker.Stop()

	// 每分钟检查一次
	minuteTicker := time.NewTicker(1 * time.Minute)
	defer minuteTicker.Stop()

	lastDailyReset := now.Truncate(24 * time.Hour)
	lastWeeklyReset := getWeekStart(now)

	for {
		select {
		case <-s.stopCh:
			log.Println("[QUEST] 任务重置定时器已停止")
			return
		case <-minuteTicker.C:
			currentTime := time.Now()

			// 检查是否需要重置日常任务（跨天）
			currentDay := currentTime.Truncate(24 * time.Hour)
			if currentDay.After(lastDailyReset) {
				log.Println("[QUEST] 执行日常任务重置...")
				s.ResetDailyQuests()
				lastDailyReset = currentDay
				// 更新下次凌晨的定时器
				nextMidnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, 0, 0, 0, 0, currentTime.Location())
				dailyTicker.Reset(time.Until(nextMidnight))
			}

			// 检查是否需要重置周常任务（每周一凌晨）
			currentWeekStart := getWeekStart(currentTime)
			if currentWeekStart.After(lastWeeklyReset) {
				log.Println("[QUEST] 执行周常任务重置...")
				s.ResetWeeklyQuests()
				lastWeeklyReset = currentWeekStart
			}
		}
	}
}

// getWeekStart 获取指定时间的本周一凌晨时间
func getWeekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday is 7 in ISO week
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}

// checkAndGrantChainReward 检查并发放任务链奖励
// 当玩家完成一个任务时，检查是否完成了整个任务链，如果是则发放额外奖励
func (s *Service) checkAndGrantChainReward(roleID uint64, config *common.QuestBaseConfig, completedRecord *RoleQuest) {
	// 如果任务不属于任何任务链，不检查
	if config.QuestChainID == nil || *config.QuestChainID == 0 {
		return
	}

	chainID := *config.QuestChainID

	// 获取该任务链的所有任务配置
	allConfigs := common.GetAllQuestConfigs()
	var chainConfigs []*common.QuestBaseConfig
	for _, cfg := range allConfigs {
		if cfg.QuestChainID != nil && *cfg.QuestChainID == chainID {
			chainConfigs = append(chainConfigs, cfg)
		}
	}

	if len(chainConfigs) == 0 {
		return
	}

	// 检查玩家是否已完成该任务链的所有任务
	existingQuests := s.roleQuestCache[roleID]
	allChainFinished := true
	for _, chainCfg := range chainConfigs {
		// 跳过刚完成的任务（已在 completedRecord）
		if chainCfg.ID == completedRecord.QuestID {
			continue
		}
		found := false
		for _, record := range existingQuests {
			if record.QuestID == chainCfg.ID && record.Status == QuestStatusFinished {
				found = true
				break
			}
		}
		if !found {
			allChainFinished = false
			break
		}
	}

	// 如果整个任务链完成，发放链奖励
	if allChainFinished && len(chainConfigs) > 1 {
		var totalChainExp, totalChainGold uint64
		for _, cfg := range chainConfigs {
			totalChainExp += cfg.ChainRewardExp
			totalChainGold += cfg.ChainRewardGold
		}

		if totalChainExp > 0 {
			common.DBRoleAddExp(roleID, int64(totalChainExp))
			log.Printf("[QUEST] 发放任务链[%d]额外经验奖励: roleID=%d, exp=%d", chainID, roleID, totalChainExp)
		}
		if totalChainGold > 0 {
			common.DBRoleAddGold(roleID, int64(totalChainGold))
			log.Printf("[QUEST] 发放任务链[%d]额外金币奖励: roleID=%d, gold=%d", chainID, roleID, totalChainGold)
		}

		// 通过MessageBus通知客户端任务链完成
		pushData := QuestPushData{
			Type:   QuestPushComplete,
			RoleID: roleID,
			Data: map[string]interface{}{
				"quest_id":        config.ID,
				"chain_id":        chainID,
				"chain_completed": true,
				"chain_exp":       totalChainExp,
				"chain_gold":      totalChainGold,
			},
		}
		common.GlobalMessageBus.Publish("quest.push", pushData)
	}
}
