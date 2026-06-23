package quest

import (
	common "game-server/Common"
	"testing"
)

// TestMain 测试入口：初始化配置
func TestMain(m *testing.M) {
	// 加载配置文件（如果尚未加载）
	if common.GameConfig == nil {
		configPath := "../Config" // 相对于Quest目录
		err := common.LoadGameConfig(configPath)
		if err != nil {
			panic("配置加载失败: " + err.Error())
		}
	}

	m.Run()
}

// TestGetQuestList_Empty 获取空任务列表（新玩家）
func TestGetQuestList_Empty(t *testing.T) {
	svc := NewService()
	roleID := uint64(9999)
	level := uint32(1)

	list, err := svc.GetQuestList(roleID, level)
	if err != nil {
		t.Fatalf("获取任务列表失败: %v", err)
	}

	// 新玩家应该有可接取的任务（至少任务1"初入江湖"等级要求为1）
	if len(list.AvailableQuests) == 0 {
		t.Errorf("期望有可接取的任务，实际: %d", len(list.AvailableQuests))
	}

	// 进行中和已完成应该为空
	if len(list.ActiveQuests) != 0 {
		t.Errorf("进行中任务应该为0，实际: %d", len(list.ActiveQuests))
	}
	if len(list.CompletedQuests) != 0 {
		t.Errorf("已完成任务应该为0，实际: %d", len(list.CompletedQuests))
	}

	t.Logf("✅ 空列表测试通过: 可接取%d个任务", len(list.AvailableQuests))
}

// TestAcceptQuest_Quest1 接取任务1（初入江湖-对话任务）
func TestAcceptQuest_Quest1(t *testing.T) {
	svc := NewService()
	roleID := uint64(10001)
	level := uint32(1)
	questID := uint32(1) // 初入江湖

	// 接取任务
	info, err := svc.AcceptQuest(roleID, questID, level)
	if err != nil {
		t.Fatalf("接取任务失败: %v", err)
	}

	// 验证返回信息
	if info.Status != QuestStatusActive {
		t.Errorf("期望状态=进行中(%d)，实际: %d", QuestStatusActive, info.Status)
	}
	if info.Progress != 0 {
		t.Errorf("期望进度=0，实际: %d", info.Progress)
	}
	if info.Name != "初入江湖" {
		t.Errorf("期望任务名=初入江湖，实际: %s", info.Name)
	}

	// 验证任务列表更新
	list, _ := svc.GetQuestList(roleID, level)
	found := false
	for _, q := range list.ActiveQuests {
		if q.ID == questID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("任务%d未出现在进行中列表", questID)
	}

	t.Logf("✅ 接取任务测试通过: 任务[%d]%s", questID, info.Name)
}

// TestAcceptQuest_LevelNotEnough 等级不足无法接取
func TestAcceptQuest_LevelNotEnough(t *testing.T) {
	svc := NewService()
	roleID := uint64(10002)
	level := uint32(1)
	questID := uint32(2) // 初试身手（需要等级3）

	_, err := svc.AcceptQuest(roleID, questID, level)
	if err == nil {
		t.Fatal("期望等级不足错误，但接取成功")
	}

	t.Logf("✅ 等级不足测试通过: %v", err)
}

// TestUpdateProgress_KillTask 击杀类任务进度更新
func TestUpdateProgress_KillTask(t *testing.T) {
	svc := NewService()
	roleID := uint64(10003)
	level := uint32(5) // 任务2需要等级3，我们用等级5

	// 先接取击杀任务（任务2: 击杀5只野鸡）
	questID := uint32(2)
	svc.AcceptQuest(roleID, questID, level)

	// 更新进度（击杀1只）
	update, err := svc.UpdateProgress(roleID, questID, 1, 101, 1) // target_type=1(击杀), target_id=101(野鸡)
	if err != nil {
		t.Fatalf("更新进度失败: %v", err)
	}

	if update.Progress != 1 {
		t.Errorf("期望进度=1，实际: %d", update.Progress)
	}
	if update.Status != QuestStatusActive {
		t.Errorf("期望状态=进行中，实际: %d", update.Status)
	}

	// 再击杀4只（总共5只）
	for i := 0; i < 4; i++ {
		update, _ = svc.UpdateProgress(roleID, questID, 1, 101, 1)
	}

	// 应该完成了
	if update.Status != QuestStatusCompleted {
		t.Errorf("期望状态=已完成(%d)，实际: %d", QuestStatusCompleted, update.Status)
	}
	if update.Progress != 5 {
		t.Errorf("期望进度=5，实际: %d", update.Progress)
	}

	t.Logf("✅ 击杀任务进度测试通过: 进度=%d/%d, 状态=%d",
		update.Progress, 5, update.Status)
}

// TestCompleteQuest 领取奖励
func TestCompleteQuest(t *testing.T) {
	svc := NewService()
	roleID := uint64(10004)
	level := uint32(1)

	// 接取对话任务（任务1: 找村长了解情况，目标数=1）
	questID := uint32(1)
	svc.AcceptQuest(roleID, questID, level)

	// 更新进度完成（对话类型，直接完成）
	svc.UpdateProgress(roleID, questID, 3, 1, 1) // target_type=3(对话), target_id=1(村长)

	// 领取奖励
	info, err := svc.CompleteQuest(roleID, questID)
	if err != nil {
		t.Fatalf("领取奖励失败: %v", err)
	}

	if info.Status != QuestStatusFinished {
		t.Errorf("期望状态=已领奖(%d)，实际: %d", QuestStatusFinished, info.Status)
	}
	if info.TotalCount != 1 {
		t.Errorf("期望总次数=1，实际: %d", info.TotalCount)
	}

	t.Logf("✅ 领取奖励测试通过: 获得+%d经验, +%d金币",
		info.RewardExp, info.RewardGold)
}

// TestAbandonQuest 放弃任务
func TestAbandonQuest(t *testing.T) {
	svc := NewService()
	roleID := uint64(10005)
	level := uint32(5)

	// 接取任务
	questID := uint32(9) // 灭鼠行动（可重复）
	svc.AcceptQuest(roleID, questID, level)

	// 放弃任务
	err := svc.AbandonQuest(roleID, questID)
	if err != nil {
		t.Fatalf("放弃任务失败: %v", err)
	}

	// 验证任务回到可接取列表
	list, _ := svc.GetQuestList(roleID, level)
	foundInAvailable := false
	for _, q := range list.AvailableQuests {
		if q.ID == questID {
			foundInAvailable = true
			break
		}
	}
	if !foundInAvailable {
		t.Errorf("放弃后任务应出现在可接取列表")
	}

	// 验证不在进行中列表
	for _, q := range list.ActiveQuests {
		if q.ID == questID {
			t.Errorf("放弃后任务不应在进行中列表")
			break
		}
	}

	t.Logf("✅ 放弃任务测试通过")
}

// TestOnMonsterKilled_AutoUpdate 怪物死亡自动更新任务
func TestOnMonsterKilled_AutoUpdate(t *testing.T) {
	svc := NewService()
	roleID := uint64(10006)
	level := uint32(5)

	// 接取击杀任务（任务2: 击杀5只野鸡）
	questID := uint32(2)
	svc.AcceptQuest(roleID, questID, level)

	// 模拟击杀3只野鸡（monsterID=101）
	updates := svc.OnMonsterKilled(roleID, 101)
	if len(updates) != 1 {
		t.Fatalf("期望1次任务更新，实际: %d", len(updates))
	}

	updates = svc.OnMonsterKilled(roleID, 101)
	updates = svc.OnMonsterKilled(roleID, 101)

	// 验证进度
	list, _ := svc.GetQuestList(roleID, level)
	var progress uint32
	for _, q := range list.ActiveQuests {
		if q.ID == questID {
			progress = q.Progress
			break
		}
	}

	if progress != 3 {
		t.Errorf("期望进度=3，实际: %d", progress)
	}

	t.Logf("✅ 自动更新测试通过: 击杀3次后进度=%d/5", progress)
}

// TestRepeatableQuest 可重复任务
func TestRepeatableQuest(t *testing.T) {
	svc := NewService()
	roleID := uint64(10007)
	level := uint32(5)

	// 接取每日任务（任务5: 每日任务，repeatable=1）
	questID := uint32(5)
	svc.AcceptQuest(roleID, questID, level)

	// 完成并领取
	svc.UpdateProgress(roleID, questID, 1, 202, 5) // 目标是5只山贼
	svc.CompleteQuest(roleID, questID)

	// 再次接取（应该成功，因为是可重复的）
	info, err := svc.AcceptQuest(roleID, questID, level)
	if err != nil {
		t.Fatalf("重复任务重新接取失败: %v", err)
	}

	if info.Status != QuestStatusActive {
		t.Errorf("重新接取后状态应该是进行中，实际: %d", info.Status)
	}
	if info.DailyCount != 1 { // 已完成1次
		t.Errorf("期望今日完成次数=1，实际: %d", info.DailyCount)
	}

	t.Logf("✅ 可重复任务测试通过: 今日第%d次接取", info.DailyCount+1)
}

// TestQuestList_Classification 任务分类正确性
func TestQuestList_Classification(t *testing.T) {
	svc := NewService()
	roleID := uint64(10008)
	level := uint32(30) // 等级30可以看到所有任务

	// 接取几个不同类型的任务
	svc.AcceptQuest(roleID, 1, 1) // 主线-对话
	svc.AcceptQuest(roleID, 2, 5) // 主线-击杀
	svc.AcceptQuest(roleID, 5, 5) // 日常-击杀

	// 完成一个
	svc.UpdateProgress(roleID, 1, 3, 1, 1)
	svc.CompleteQuest(roleID, 1)

	// 获取列表
	list, err := svc.GetQuestList(roleID, level)
	if err != nil {
		t.Fatalf("获取任务列表失败: %v", err)
	}

	// 验证分类
	hasActive := len(list.ActiveQuests) > 0       // 应该有进行中的
	_ = len(list.CompletedQuests) > 0 // 已完成（可能为空）
	hasFinished := len(list.FinishedQuests) > 0   // 应该有已领奖的
	hasAvailable := len(list.AvailableQuests) > 0 // 应该有可接取的

	if !hasActive || !hasFinished || !hasAvailable {
		t.Errorf("任务分类不完整: active=%d, completed=%d, finished=%d, available=%d",
			len(list.ActiveQuests), len(list.CompletedQuests),
			len(list.FinishedQuests), len(list.AvailableQuests))
	}

	t.Logf("✅ 分类测试通过: 进行中=%d, 已完成=%d, 已领奖=%d, 可接取=%d",
		len(list.ActiveQuests), len(list.CompletedQuests),
		len(list.FinishedQuests), len(list.AvailableQuests))
}
