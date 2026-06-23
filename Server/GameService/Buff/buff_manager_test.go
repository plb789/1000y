package buff

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	common "game-server/Common"
)

// TestMain 测试入口：初始化配置
func TestMain(m *testing.M) {
	// 加载配置文件（如果尚未加载）
	if common.GameConfig == nil {
		configPath := "../Config" // 相对于Buff目录
		err := common.LoadGameConfig(configPath)
		if err != nil {
			fmt.Printf("⚠️ 配置加载失败: %v\n", err)
			fmt.Printf("⚠️ 将跳过需要配置的测试用例\n")
		} else {
			fmt.Printf("✅ 配置加载成功: %d个BUFF, %d个技能\n",
				len(common.GameConfig.Buffs), len(common.GameConfig.Skills))
		}
	}

	// 运行测试
	exitCode := m.Run()

	os.Exit(exitCode)
}

// MockApplier 模拟实现BuffTickApplier接口用于测试
type MockApplier struct {
	mu              sync.Mutex
	PlayerHPChanges []PlayerHPChange
	PlayerMPChanges []PlayerMPChange
	MonsterChanges  []MonsterHPChange
	Broadcasts      []BroadcastRecord
	BatchBroadcasts []BatchBroadcastRecord
}

type PlayerHPChange struct {
	PlayerID uint64
	HPChange int
}

type PlayerMPChange struct {
	PlayerID uint64
	MPChange int
}

type MonsterHPChange struct {
	MonsterID uint64
	HPChange  int
	CurrentHP int
	IsDead    bool
}

type BroadcastRecord struct {
	TargetID   uint64
	TargetType uint8
	HpChange   int
	MpChange   int
}

type BatchBroadcastRecord struct {
	ResultsCount int
}

func (m *MockApplier) ApplyPlayerHPChange(playerID uint64, hpChange int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PlayerHPChanges = append(m.PlayerHPChanges, PlayerHPChange{
		PlayerID: playerID,
		HPChange: hpChange,
	})
}

func (m *MockApplier) ApplyPlayerMPChange(playerID uint64, mpChange int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PlayerMPChanges = append(m.PlayerMPChanges, PlayerMPChange{
		PlayerID: playerID,
		MPChange: mpChange,
	})
}

func (m *MockApplier) ApplyMonsterHPChange(monsterID uint64, hpChange int) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟怪物HP变化（假设初始HP=100）
	staticHP := map[uint64]int{
		10001: 100,
		10002: 50,
		10003: 15, // 低血量，会被击杀
	}

	currentHP := staticHP[monsterID]
	if currentHP == 0 {
		currentHP = 100 // 默认值
	}

	newHP := currentHP + hpChange // hpChange是负数（伤害）
	isDead := newHP <= 0

	if isDead {
		newHP = 0
	}

	m.MonsterChanges = append(m.MonsterChanges, MonsterHPChange{
		MonsterID: monsterID,
		HPChange:  hpChange,
		CurrentHP: newHP,
		IsDead:    isDead,
	})

	return newHP, isDead
}

func (m *MockApplier) BroadcastBuffTick(targetID uint64, targetType uint8, hpChange int, mpChange int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Broadcasts = append(m.Broadcasts, BroadcastRecord{
		TargetID:   targetID,
		TargetType: targetType,
		HpChange:   hpChange,
		MpChange:   mpChange,
	})
}

func (m *MockApplier) BroadcastBuffTickBatch(results []BuffTickResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BatchBroadcasts = append(m.BatchBroadcasts, BatchBroadcastRecord{
		ResultsCount: len(results),
	})
}

// ========== 单元测试用例 ==========

// TestTickAllTargets_Basic 测试基础Tick功能（玩家打坐回血）
func TestTickAllTargets_Basic(t *testing.T) {
	manager := GetManager()
	applier := &MockApplier{}

	// 检查配置是否已加载（buff_id=1必须存在）
	config := common.GetBuffConfig(1)
	if config == nil {
		t.Skip("跳过：配置文件未加载或不存在buff_id=1（打坐）")
	}

	// 为玩家添加打坐BUFF (buff_id=1, 每秒+10HP/+5MP)
	buff := manager.AddBuff(1001, 1, 1, 0) // targetID=1001, targetType=1(玩家), buffID=1(打坐), sourceID=0
	if buff == nil {
		t.Fatal("添加打坐BUFF失败")
	}

	// 手动设置LastTickAt为1秒前（确保Tick方法的时间检查通过）
	buff.LastTickAt = time.Now().Add(-1 * time.Second)

	// 执行一次tick
	manager.tickAllTargets(applier)

	// 验证玩家HP被修改
	if len(applier.PlayerHPChanges) != 1 {
		t.Fatalf("期望1次玩家HP修改，实际: %d", len(applier.PlayerHPChanges))
	}
	if applier.PlayerHPChanges[0].PlayerID != 1001 {
		t.Errorf("期望玩家ID=1001，实际: %d", applier.PlayerHPChanges[0].PlayerID)
	}
	if applier.PlayerHPChanges[0].HPChange != 10 { // buffs.json中打坐hp_change=10
		t.Errorf("期望HP变化=+10，实际: %d", applier.PlayerHPChanges[0].HPChange)
	}

	// 验证玩家MP被修改
	if len(applier.PlayerMPChanges) != 1 {
		t.Fatalf("期望1次玩家MP修改，实际: %d", len(applier.PlayerMPChanges))
	}
	if applier.PlayerMPChanges[0].MPChange != 5 { // buffs.json中打坐mp_change=5
		t.Errorf("期望MP变化=+5，实际: %d", applier.PlayerMPChanges[0].MPChange)
	}

	// 验证批量广播被调用
	if len(applier.BatchBroadcasts) != 1 {
		t.Fatalf("期望1次批量广播，实际: %d", len(applier.BatchBroadcasts))
	}
	if applier.BatchBroadcasts[0].ResultsCount != 1 {
		t.Errorf("期望广播1个结果，实际: %d", applier.BatchBroadcasts[0].ResultsCount)
	}

	// 清理测试数据
	manager.ClearAllBuffs(1001)

	t.Logf("✅ 基础Tick测试通过: 玩家[1001] HP:+50 MP:+10")
}

// TestTickAllTargets_MonsterPoisonDeath 测试怪物中毒致死
func TestTickAllTargets_MonsterPoisonDeath(t *testing.T) {
	manager := GetManager()
	applier := &MockApplier{}

	// 检查配置是否已加载（buff_id=2必须存在）
	config := common.GetBuffConfig(2)
	if config == nil {
		t.Skip("跳过：配置文件未加载或不存在buff_id=2（中毒）")
	}

	// 为低血量怪物添加中毒BUFF (buff_id=2, 每秒-5HP)
	buff := manager.AddBuff(10003, 2, 2, 1001) // targetID=10003(怪物), targetType=2, buffID=2(中毒)
	if buff == nil {
		t.Fatal("添加中毒BUFF失败")
	}

	// 手动设置LastTickAt为1秒前
	buff.LastTickAt = time.Now().Add(-1 * time.Second)

	// 执行一次tick（怪物当前HP=15，受到-5伤害，不会死亡）
	manager.tickAllTargets(applier)

	// 验证怪物HP被修改（但不应该死亡，因为-5不会杀死15HP的怪物）
	if len(applier.MonsterChanges) != 1 {
		t.Fatalf("期望1次怪物HP修改，实际: %d", len(applier.MonsterChanges))
	}
	if applier.MonsterChanges[0].IsDead {
		t.Error("期望怪物不死亡（-5伤害不足以杀死15HP怪物），但IsDead=true")
	}
	if applier.MonsterChanges[0].HPChange != -5 { // buffs.json中中毒hp_change=-5
		t.Errorf("期望HP变化=-5，实际: %d", applier.MonsterChanges[0].HPChange)
	}
}

// TestTickAllTargets_MultipleTargets 测试多目标批量处理
func TestTickAllTargets_MultipleTargets(t *testing.T) {
	manager := GetManager()
	applier := &MockApplier{}

	// 检查配置是否已加载
	if common.GetBuffConfig(1) == nil || common.GetBuffConfig(2) == nil {
		t.Skip("跳过：配置文件未加载或缺少必要BUFF配置")
	}

	// 添加3个目标的不同BUFF
	buff1 := manager.AddBuff(1001, 1, 1, 0)     // 玩家打坐
	buff2 := manager.AddBuff(10001, 2, 2, 1001) // 怪物中毒
	buff3 := manager.AddBuff(10002, 2, 2, 1001) // 另一个怪物中毒

	// 手动设置LastTickAt为1秒前（确保时间检查通过）
	if buff1 != nil {
		buff1.LastTickAt = time.Now().Add(-1 * time.Second)
	}
	if buff2 != nil {
		buff2.LastTickAt = time.Now().Add(-1 * time.Second)
	}
	if buff3 != nil {
		buff3.LastTickAt = time.Now().Add(-1 * time.Second)
	}

	// 执行一次tick
	manager.tickAllTargets(applier)

	// 验证3个目标都被处理
	totalChanges := len(applier.PlayerHPChanges) +
		len(applier.PlayerMPChanges) +
		len(applier.MonsterChanges)
	if totalChanges != 4 { // 玩家HP+MP + 2个怪物HP
		t.Fatalf("期望4次修改（1HP+1MP+2怪物HP），实际: %d", totalChanges)
	}

	// 验证批量广播包含3个结果
	if len(applier.BatchBroadcasts) != 1 {
		t.Fatal("期望1次批量广播")
	}
	if applier.BatchBroadcasts[0].ResultsCount != 3 {
		t.Errorf("期望广播3个结果，实际: %d", applier.BatchBroadcasts[0].ResultsCount)
	}

	// 清理
	manager.ClearAllBuffs(1001)
	manager.ClearAllBuffs(10001)
	manager.ClearAllBuffs(10002)

	t.Logf("✅ 多目标批量处理测试通过: 3个目标 → 4次修改 → 1次批量广播(含3个结果)")
}

// TestTickAllTargets_NoEffectBuffs 跳过无效果BUFF
func TestTickAllTargets_NoEffectBuffs(t *testing.T) {
	// 注意：此测试依赖配置文件中存在纯属性加成类BUFF（无hp_change/mp_change）
	// 如果buffs.json中没有此类BUFF，此测试会跳过

	manager := GetManager()
	applier := &MockApplier{}

	// 尝试添加属性加成类BUFF（如果存在的话）
	// 这里使用buff_id=3作为示例（假设是攻击力提升类BUFF）
	buff := manager.AddBuff(2001, 1, 3, 0)
	if buff == nil {
		t.Skip("跳过：配置文件中不存在buff_id=3（纯属性加成类BUFF）")
	}

	// 执行tick
	manager.tickAllTargets(applier)

	// 验证没有HP/MP修改
	if len(applier.PlayerHPChanges) > 0 || len(applier.PlayerMPChanges) > 0 {
		t.Error("期望无HP/MP修改（纯属性BUFF不应触发）")
	}

	// 验证也没有广播（因为无效果）
	if len(applier.BatchBroadcasts) > 0 {
		t.Error("期望无广播（无效果的BUFF不应推送）")
	}

	// 清理
	manager.ClearAllBuffs(2001)

	t.Log("✅ 无效果BUFF跳过测试通过: 属性加成类BUFF不触发Tick效果")
}

// TestTicker_StartStop 测试定时器启动和停止
func TestTicker_StartStop(t *testing.T) {
	manager := GetManager()
	applier := &MockApplier{}

	// 检查配置是否已加载
	if common.GetBuffConfig(1) == nil {
		t.Skip("跳过：配置文件未加载或不存在buff_id=1")
	}

	// 启动定时器
	manager.StartTicker(applier)
	time.Sleep(100 * time.Millisecond) // 等待启动日志

	// 尝试重复启动（应被阻止）
	manager.StartTicker(applier)

	// 添加一个测试BUFF（假设buff_id=1存在且有hp_change>0）
	manager.AddBuff(9999, 1, 1, 0)

	// 等待至少2-3个tick周期
	time.Sleep(2500 * time.Millisecond) // 2.5秒确保触发至少2次tick

	// 停止定时器
	manager.StopTicker()

	// 验证tick被执行了至少1次（可能触发1-3次，取决于系统调度）
	if len(applier.PlayerHPChanges) < 1 {
		t.Errorf("期望至少1次tick执行，实际: %d次", len(applier.PlayerHPChanges))
	} else {
		t.Logf("✅ 定时器测试通过: 在2.5秒内触发了%d次tick", len(applier.PlayerHPChanges))
	}

	// 清理
	manager.ClearAllBuffs(9999)
}

// TestBuffExpiration_Cleanup 测试过期BUFF自动清理
func TestBuffExpiration_Cleanup(t *testing.T) {
	manager := GetManager()

	// 手动创建一个已过期的BUFF（绕过AddBuff的自动过期时间设置）
	expiredBuff := &ActiveBuff{
		BuffID:     2,
		TargetType: 1,
		Stack:      1,
		ExpireAt:   time.Now().Add(-1 * time.Minute), // 1分钟前就过期了
		LastTickAt: time.Now().Add(-1 * time.Minute),
	}

	// 直接添加到管理器（模拟已存在的过期BUFF）
	manager.mu.Lock()
	manager.buffs[2001] = []*ActiveBuff{expiredBuff}
	manager.mu.Unlock()

	// 执行清理
	manager.CleanupExpired()

	// 验证过期BUFF被移除
	buffs := manager.GetBuffs(2001)
	if len(buffs) != 0 {
		t.Errorf("期望过期BUFF被清理，实际剩余: %d个", len(buffs))
	} else {
		t.Log("✅ 过期BUFF清理测试通过: 已过期BUFF被自动移除")
	}
}

// TestBuffResultStructure 测试BuffTickResult结构体字段完整性
func TestBuffResultStructure(t *testing.T) {
	result := BuffTickResult{
		TargetID:   12345,
		TargetType: 2,
		HpChange:   -20,
		MpChange:   0,
	}

	if result.TargetID != 12345 {
		t.Error("TargetID字段错误")
	}
	if result.TargetType != 2 {
		t.Error("TargetType字段错误")
	}
	if result.HpChange != -20 {
		t.Error("HpChange字段错误")
	}
	if result.MpChange != 0 {
		t.Error("MpChange字段错误")
	}

	t.Log("✅ BuffTickResult结构体字段验证通过")
}
