package common

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ConfigLoader 配置加载器
type ConfigLoader struct {
	Maps          []MapBaseConfig      // 地图配置
	NPCs          []NPCBaseConfig      // NPC配置
	Monsters      []MonsterBaseConfig  // 怪物配置
	Skills        []SkillBaseConfig    // 武学配置
	Items         []ItemBaseConfig     // 道具配置
	DropGroups    []DropGroupConfig    // 掉落组配置
	Buffs         []BuffBaseConfig     // BUFF配置
	Quests        []QuestBaseConfig    // 任务配置
	Achievements  []AchievementConfig  // 成就配置（新增）
	Shops         []ShopConfig         // 商店配置
	ShopGoods     []ShopGoodsConfig    // 商品配置
	Announcements []AnnouncementConfig // 公告配置
	ServerConfigs []ServerConfigItem   // 服务器配置
	SpawnPoints   SpawnPointsConfig    // 怪物生成点配置（新增）
}

// DropGroupConfig 掉落组配置
type DropGroupConfig struct {
	ID        uint32 `json:"id"`
	MonsterID uint32 `json:"monster_id"`
	ItemID    uint32 `json:"item_id"`
	DropRate  uint32 `json:"drop_rate"` // 掉落概率(万分比)
	DropMin   uint32 `json:"drop_min"`
	DropMax   uint32 `json:"drop_max"`
}

// MapBaseConfig 地图配置
type MapBaseConfig struct {
	ID          uint32  `json:"id"`
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	TileWidth   int     `json:"tile_width"`
	TileHeight  int     `json:"tile_height"`
	MapFile     string  `json:"map_file"`
	TilesetFile string  `json:"tileset_file"` // 瓦片图集文件
	Music       string  `json:"music"`
	PkAllowed   uint8   `json:"pk_allowed"`
	ReviveMapID *uint32 `json:"revive_map_id"`
	ReviveX     *int    `json:"revive_x"`
	ReviveY     *int    `json:"revive_y"`
	LevelReq    uint32  `json:"level_req"`
	MinimapFile string  `json:"minimap_file"`
}

// NPCBaseConfig NPC配置
type NPCBaseConfig struct {
	ID         uint32  `json:"id"`
	Name       string  `json:"name"`
	Type       uint8   `json:"type"`
	MapID      uint32  `json:"map_id"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Face       uint8   `json:"face"`
	SpriteID   *int    `json:"sprite_id"`
	DialogText string  `json:"dialog_text"`
	ShopID     *uint32 `json:"shop_id"`
}

// MonsterBaseConfig 怪物配置
type MonsterBaseConfig struct {
	ID          uint32  `json:"id"`
	Name        string  `json:"name"`
	Level       uint32  `json:"level"`
	Type        uint8   `json:"type"`
	MapID       uint32  `json:"map_id"`
	Hp          int     `json:"hp"`
	Attack      int     `json:"attack"`
	Defense     int     `json:"defense"`
	Speed       int     `json:"speed"`
	Hit         int     `json:"hit"`
	Dodge       int     `json:"dodge"`
	Crit        int     `json:"crit"`
	AIType      uint8   `json:"ai_type"`
	AttackRange int     `json:"attack_range"`
	ChaseRange  int     `json:"chase_range"`
	GoldMin     int     `json:"gold_min"`
	GoldMax     int     `json:"gold_max"`
	Exp         int     `json:"exp"`
	DropGroupID *uint32 `json:"drop_group_id"`
	SpriteID    *int    `json:"sprite_id"`
	RespawnTime int     `json:"respawn_time"`
}

// SkillBaseConfig 武学配置
type SkillBaseConfig struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	Type        uint8  `json:"type"`     // 武功类型: 1=内功, 2=外功, 3=轻功, 4=护体, 5=拳法, 6=剑法, 7=刀法, 8=枪法, 9=斧法
	SubType     uint8  `json:"sub_type"` // 子类型: 1=基础, 2=进阶, 3=高级
	Level       uint32 `json:"level"`
	MaxLevel    uint32 `json:"max_level"`
	ExpFactor   uint32 `json:"exp_factor"`
	Description string `json:"description"`
	HpBonus     int    `json:"hp_bonus"`
	MpBonus     int    `json:"mp_bonus"`
	AttackBonus int    `json:"attack_bonus"`
	DefBonus    int    `json:"defense_bonus"`
	SpeedBonus  int    `json:"speed_bonus"`
	HitBonus    int    `json:"hit_bonus"`
	DodgeBonus  int    `json:"dodge_bonus"`
	CritBonus   int    `json:"crit_bonus"`
	BuffID      uint32 `json:"buff_id"`
	SkillEffect string `json:"skill_effect"`
	IsActive    uint8  `json:"is_active"`
	WeaponType  uint8  `json:"weapon_type"` // 武器类型: 0=徒手, 1=剑, 2=刀, 3=枪, 4=斧, 5=拳
	// 战斗属性（从skills.json加载）
	Damage      int `json:"damage"`       // 基础伤害
	MpCost      int `json:"mp_cost"`      // 内力消耗
	Cooldown    int `json:"cooldown"`     // 冷却时间(秒)，0=无冷却走攻速
	Range       int `json:"range"`        // 攻击范围(1=近战, 2-5=远程)
	CastTime    int `json:"cast_time"`    // 施法时间(毫秒)
	AoeRadius   int `json:"aoe_radius"`   // AOE范围半径(0=单体)
	AttackSpeed int `json:"attack_speed"` // 攻速值(1-100，越低越快，0=不适用)
}

// ItemBaseConfig 道具配置
type ItemBaseConfig struct {
	ID            uint32 `json:"id"`
	Name          string `json:"name"`
	Type          uint8  `json:"type"`
	SubType       uint8  `json:"sub_type"` // 武器子类型: 1=剑, 2=刀, 3=棍, 4=枪, 5=斧
	Quality       uint8  `json:"quality"`
	LevelReq      uint32 `json:"level_req"`
	StackMax      uint32 `json:"stack_max"`
	Price         int    `json:"price"`
	Description   string `json:"description"`
	EquipType     uint8  `json:"equip_type"`    // 装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链
	WeaponType    uint8  `json:"weapon_type"`   // 武器类型: 0=徒手, 1=剑, 2=刀, 3=枪, 4=斧, 5=拳 (仅武器有效)
	HpBonus       int    `json:"hp_bonus"`      // 生命加成(装备)
	MpBonus       int    `json:"mp_bonus"`      // 内力加成(装备)
	AttackBonus   int    `json:"attack_bonus"`  // 攻击加成(装备)
	DefenseBonus  int    `json:"defense_bonus"` // 防御加成(装备)
	SpeedBonus    int    `json:"speed_bonus"`   // 速度加成(装备)
	HpRestore     int    `json:"hp_restore"`
	MpRestore     int    `json:"mp_restore"`
	BuffID        uint32 `json:"buff_id"`
	Icon          string `json:"icon"`
	Model         string `json:"model"`
	IsDropable    uint8  `json:"is_dropable"`
	IsSellable    uint8  `json:"is_sellable"`
	IsDestroyable uint8  `json:"is_destroyable"`
	IsBind        uint8  `json:"is_bind"`
}

// BuffBaseConfig BUFF配置
type BuffBaseConfig struct {
	ID           uint32 `json:"id"`
	Name         string `json:"name"`
	Type         uint8  `json:"type"`           // 1=增益, 2=减益, 3=控制
	Duration     int    `json:"duration"`       // 持续时间(秒), 0=永久
	HpChange     int    `json:"hp_change"`      // 每秒HP变化(正=恢复, 负=持续伤害)
	MpChange     int    `json:"mp_change"`      // 每秒MP变化
	AttackChange int    `json:"attack_change"`  // 攻击力加成
	DefChange    int    `json:"defense_change"` // 防御力加成
	SpeedChange  int    `json:"speed_change"`   // 速度加成
	HitChange    int    `json:"hit_change"`     // 命中率加成
	DodgeChange  int    `json:"dodge_change"`   // 闪避率加成
	CritChange   int    `json:"crit_change"`    // 暴击率加成
	CanCancel    uint8  `json:"can_cancel"`     // 是否可手动取消(1=可)
	StackMax     int    `json:"stack_max"`      // 最大叠加层数
	Description  string `json:"description"`    // 描述
	// 高级战斗效果（新增）
	DamageReductionPct int `json:"damage_reduction_pct"` // 减伤百分比(0-100), 如无敌=100
	ReflectPct         int `json:"reflect_pct"`          // 反弹伤害百分比(0-100)
	LifestealPct       int `json:"lifesteal_pct"`        // 吸血百分比(0-100), 攻击时回血
}

// QuestObjectiveConfig 任务目标配置（支持多目标）
type QuestObjectiveConfig struct {
	ID          uint32 `json:"id"`           // 目标ID
	TargetType  uint8  `json:"target_type"`  // 目标类型: 1=击杀, 2=采集, 3=对话, 4=探索
	TargetID    uint32 `json:"target_id"`    // 目标ID（怪物ID/物品ID/NPC ID等）
	TargetName  string `json:"target_name"`  // 目标名称（用于前端显示）
	TargetCount int    `json:"target_count"` // 目标数量
}

// QuestBaseConfig 任务配置
type QuestBaseConfig struct {
	ID              uint32                 `json:"id"`
	Name            string                 `json:"name"`
	Type            uint8                  `json:"type"` // 任务类型: 1=主线, 2=支线, 3=日常, 4=周常, 5=活动
	LevelReq        uint32                 `json:"level_req"`
	Repeatable      uint8                  `json:"repeatable"`
	TargetType      uint8                  `json:"target_type"`  // 主目标类型（兼容单目标任务）
	TargetID        uint32                 `json:"target_id"`    // 主目标ID（兼容单目标任务）
	TargetName      string                 `json:"target_name"`  // 主目标名称（用于前端显示）
	TargetCount     int                    `json:"target_count"` // 主目标数量
	Objectives      []QuestObjectiveConfig `json:"objectives"`   // 多目标列表（支持复杂任务）
	RewardExp       uint64                 `json:"reward_exp"`
	RewardGold      uint64                 `json:"reward_gold"`
	RewardHonor     uint64                 `json:"reward_honor"` // 声望奖励（新增）
	RewardItemID    *uint32                `json:"reward_item_id"`
	RewardItemCount *int                   `json:"reward_item_count"`
	Description     string                 `json:"description"`
	NPCAcceptID     *uint32                `json:"npc_accept_id"`
	NPCCompleteID   *uint32                `json:"npc_complete_id"`
	PrevQuestID     *uint32                `json:"prev_quest_id"`     // 前置任务ID（新增）
	TimeLimit       *int                   `json:"time_limit"`        // 时间限制（秒，新增）
	AutoAccept      uint8                  `json:"auto_accept"`       // 是否自动接取（新增）
	AutoComplete    uint8                  `json:"auto_complete"`     // 是否自动完成（新增）
	QuestChainID    *uint32                `json:"quest_chain_id"`    // 任务链ID（同一链的任务共享）
	ChainRewardExp  uint64                 `json:"chain_reward_exp"`  // 任务链额外经验奖励
	ChainRewardGold uint64                 `json:"chain_reward_gold"` // 任务链额外金币奖励
}

// AchievementConfig 成就配置
type AchievementConfig struct {
	ID              uint32  `json:"id"`                // 成就ID
	Name            string  `json:"name"`              // 成就名称
	Description     string  `json:"description"`       // 成就描述
	Type            uint8   `json:"type"`              // 成就类型: 1=任务, 2=战斗, 3=收集, 4=探索, 5=社交
	Condition       string  `json:"condition"`         // 条件类型: quest_complete, monster_kill, item_collect, etc.
	TargetID        uint32  `json:"target_id"`         // 目标ID (如任务ID、怪物ID等)
	TargetCount     int     `json:"target_count"`      // 目标数量
	RewardExp       uint64  `json:"reward_exp"`        // 奖励经验
	RewardGold      uint64  `json:"reward_gold"`       // 奖励金币
	RewardHonor     uint64  `json:"reward_honor"`      // 奖励声望
	RewardItemID    *uint32 `json:"reward_item_id"`    // 奖励物品ID
	RewardItemCount *int    `json:"reward_item_count"` // 奖励物品数量
	Icon            string  `json:"icon"`              // 图标
	Point           uint32  `json:"point"`             // 成就点数
}

// ShopConfig 商店配置
type ShopConfig struct {
	ID    uint32  `json:"id"`
	Name  string  `json:"name"`
	Type  uint8   `json:"type"`
	NPCID *uint32 `json:"npc_id"`
}

// ShopGoodsConfig 商品配置
type ShopGoodsConfig struct {
	ID        uint32 `json:"id"`
	ShopID    uint32 `json:"shop_id"`
	ItemID    uint32 `json:"item_id"`
	Price     int64  `json:"price"`
	PriceType uint8  `json:"price_type"`
	Stock     int    `json:"stock"`
}

// AnnouncementConfig 公告配置
type AnnouncementConfig struct {
	ID        uint32  `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      uint8   `json:"type"`
	Priority  uint8   `json:"priority"`
	StartTime string  `json:"start_time"`
	EndTime   *string `json:"end_time"`
}

// ServerConfigItem 服务器配置项
type ServerConfigItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ========== 怪物生成点配置（新增）==========

// SpawnPointsConfig 怪物生成点配置（顶层）
type SpawnPointsConfig struct {
	Version     string                    `json:"version"`      // 配置版本号
	Description string                    `json:"description"`  // 配置描述
	LastUpdated string                    `json:"last_updated"` // 最后更新时间
	Maps        map[string]MapSpawnConfig `json:"maps"`         // 地图生成点配置（key=地图ID）
}

// MapSpawnConfig 单个地图的怪物生成配置
type MapSpawnConfig struct {
	MapID          uint32                 `json:"map_id"`          // 地图ID
	MapName        string                 `json:"map_name"`        // 地图名称
	SpawnPoints    []SpawnPointConfig     `json:"spawn_points"`    // 生成点列表
	GlobalSettings MapSpawnGlobalSettings `json:"global_settings"` // 全局设置
}

// SpawnPointConfig 单个生成点配置
type SpawnPointConfig struct {
	ID             uint32 `json:"id"`               // 生成点唯一ID
	Name           string `json:"name"`             // 生成点名称
	BaseMonsterID  uint32 `json:"base_monster_id"`  // 基础怪物模板ID
	MonsterName    string `json:"monster_name"`     // 怪物名称（可选，用于显示）
	X              int    `json:"x"`                // 中心X坐标
	Y              int    `json:"y"`                // 中心Y坐标
	Count          int    `json:"count"`            // 同时存在的最大数量
	RespawnTime    int    `json:"respawn_time"`     // 复活时间（秒）
	SpawnRadius    int    `json:"spawn_radius"`     // 生成半径（格），0=精确位置
	LevelRange     [2]int `json:"level_range"`      // 适用玩家等级范围 [min, max]
	AITypeOverride *uint8 `json:"ai_type_override"` // AI类型覆盖（null=使用默认值）
	IsActive       bool   `json:"is_active"`        // 是否激活
	Description    string `json:"description"`      // 描述说明
}

// MapSpawnGlobalSettings 地图生成全局设置
type MapSpawnGlobalSettings struct {
	MaxMonstersPerMap    int  `json:"max_monsters_per_map"`   // 地图最大怪物数量
	AutoRespawn          bool `json:"auto_respawn"`           // 是否自动复活
	RespawnCheckInterval int  `json:"respawn_check_interval"` // 复活检查间隔（秒）
	DespawnDistance      int  `json:"despawn_distance"`       // 玩家离开多远后消失（格）
}

// 全局配置实例
var GameConfig *ConfigLoader

// LoadGameConfig 加载所有游戏配置
func LoadGameConfig(configPath string) error {
	GameConfig = &ConfigLoader{}

	// 加载地图配置
	if err := loadJSONFile(configPath+"/maps.json", &GameConfig.Maps); err != nil {
		return err
	}

	// 加载NPC配置
	if err := loadJSONFile(configPath+"/npcs.json", &GameConfig.NPCs); err != nil {
		return err
	}

	// 加载怪物配置
	if err := loadJSONFile(configPath+"/monsters.json", &GameConfig.Monsters); err != nil {
		return err
	}

	// 加载武学配置
	if err := loadJSONFile(configPath+"/skills.json", &GameConfig.Skills); err != nil {
		return err
	}

	// 加载道具配置
	if err := loadJSONFile(configPath+"/items.json", &GameConfig.Items); err != nil {
		return err
	}

	// 加载掉落组配置
	if err := loadJSONFile(configPath+"/drop_groups.json", &GameConfig.DropGroups); err != nil {
		return err
	}

	// 加载BUFF配置
	if err := loadJSONFile(configPath+"/buffs.json", &GameConfig.Buffs); err != nil {
		return err
	}

	// 加载任务配置
	if err := loadJSONFile(configPath+"/quests.json", &GameConfig.Quests); err != nil {
		return err
	}

	// 加载成就配置
	if err := loadJSONFile(configPath+"/achievements.json", &GameConfig.Achievements); err != nil {
		log.Printf("⚠️ 成就配置加载失败: %v (可选)", err)
		GameConfig.Achievements = []AchievementConfig{} // 允许为空
	} else {
		log.Printf("✅ 成就配置加载成功: %d个成就", len(GameConfig.Achievements))
	}

	// 加载商店配置(包含商店和商品)
	var shopsData struct {
		Shops     []ShopConfig      `json:"shops"`
		ShopGoods []ShopGoodsConfig `json:"shop_goods"`
	}
	if err := loadJSONFile(configPath+"/shops.json", &shopsData); err != nil {
		return err
	}
	GameConfig.Shops = shopsData.Shops
	GameConfig.ShopGoods = shopsData.ShopGoods

	// 加载公告配置
	if err := loadJSONFile(configPath+"/announcements.json", &GameConfig.Announcements); err != nil {
		return err
	}

	// 加载服务器配置
	if err := loadJSONFile(configPath+"/server_config.json", &GameConfig.ServerConfigs); err != nil {
		return err
	}

	// 加载怪物生成点配置（新增）
	if err := loadJSONFile(configPath+"/monster_spawns.json", &GameConfig.SpawnPoints); err != nil {
		log.Printf("警告: 加载怪物生成配置失败: %v (将使用默认算法生成)", err)
		// 不返回错误，使用空配置（后续会fallback到算法生成）
		GameConfig.SpawnPoints = SpawnPointsConfig{
			Maps: make(map[string]MapSpawnConfig),
		}
	} else {
		log.Printf("✅ 怪物生成配置加载成功: %d个地图", len(GameConfig.SpawnPoints.Maps))
	}

	return nil
}

func loadJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// GetMapConfig 获取地图配置
func GetMapConfig(mapID uint32) *MapBaseConfig {
	for i := range GameConfig.Maps {
		if GameConfig.Maps[i].ID == mapID {
			return &GameConfig.Maps[i]
		}
	}
	return nil
}

// GetNPCConfig 获取NPC配置
func GetNPCConfig(npcID uint32) *NPCBaseConfig {
	for i := range GameConfig.NPCs {
		if GameConfig.NPCs[i].ID == npcID {
			return &GameConfig.NPCs[i]
		}
	}
	return nil
}

// GetNPCsByMap 获取地图所有NPC
func GetNPCsByMap(mapID uint32) []NPCBaseConfig {
	var npcs []NPCBaseConfig
	for i := range GameConfig.NPCs {
		if GameConfig.NPCs[i].MapID == mapID {
			npcs = append(npcs, GameConfig.NPCs[i])
		}
	}
	return npcs
}

// GetMonsterConfig 获取怪物配置
func GetMonsterConfig(monsterID uint32) *MonsterBaseConfig {
	for i := range GameConfig.Monsters {
		if GameConfig.Monsters[i].ID == monsterID {
			return &GameConfig.Monsters[i]
		}
	}
	return nil
}

// GetSkillConfig 获取武学配置
func GetSkillConfig(skillID uint32) *SkillBaseConfig {
	for i := range GameConfig.Skills {
		if GameConfig.Skills[i].ID == skillID {
			return &GameConfig.Skills[i]
		}
	}
	return nil
}

// GetAllSkillConfig 获取所有武学配置
func GetAllSkillConfig() []SkillBaseConfig {
	return GameConfig.Skills
}

// GetItemConfig 获取道具配置
func GetItemConfig(itemID uint32) *ItemBaseConfig {
	for i := range GameConfig.Items {
		if GameConfig.Items[i].ID == itemID {
			return &GameConfig.Items[i]
		}
	}
	return nil
}

// GetItemConfigByEquipType 根据装备类型获取道具配置（用于装备槽位显示）
func GetItemConfigByEquipType(equipType uint8) *ItemBaseConfig {
	// 返回该类型的第一个物品配置（简化处理）
	// 实际应用中应该通过 bag_item_id 查询具体物品
	for i := range GameConfig.Items {
		if GameConfig.Items[i].Type == 2 && GameConfig.Items[i].EquipType == equipType { // Type=2 表示装备
			return &GameConfig.Items[i]
		}
	}
	return nil
}

// GetAllItemConfig 获取所有道具配置
func GetAllItemConfig() []ItemBaseConfig {
	return GameConfig.Items
}

// GetDropsByMonsterID 获取怪物的掉落配置
func GetDropsByMonsterID(monsterID uint32) []DropGroupConfig {
	var drops []DropGroupConfig
	for i := range GameConfig.DropGroups {
		if GameConfig.DropGroups[i].MonsterID == monsterID {
			drops = append(drops, GameConfig.DropGroups[i])
		}
	}
	return drops
}

// GetBuffConfig 获取BUFF配置
func GetBuffConfig(buffID uint32) *BuffBaseConfig {
	for i := range GameConfig.Buffs {
		if GameConfig.Buffs[i].ID == buffID {
			return &GameConfig.Buffs[i]
		}
	}
	return nil
}

// GetQuestConfig 获取任务配置
func GetQuestConfig(questID uint32) *QuestBaseConfig {
	for i := range GameConfig.Quests {
		if GameConfig.Quests[i].ID == questID {
			return &GameConfig.Quests[i]
		}
	}
	return nil
}

// GetQuestsByType 获取指定类型任务
func GetQuestsByType(questType uint8) []QuestBaseConfig {
	var quests []QuestBaseConfig
	for i := range GameConfig.Quests {
		if GameConfig.Quests[i].Type == questType {
			quests = append(quests, GameConfig.Quests[i])
		}
	}
	return quests
}

// GetAllQuestConfigs 获取所有任务配置
func GetAllQuestConfigs() []*QuestBaseConfig {
	configs := make([]*QuestBaseConfig, 0, len(GameConfig.Quests))
	for i := range GameConfig.Quests {
		configs = append(configs, &GameConfig.Quests[i])
	}
	return configs
}

// GetAchievementConfig 获取成就配置
func GetAchievementConfig(achievementID uint32) *AchievementConfig {
	for i := range GameConfig.Achievements {
		if GameConfig.Achievements[i].ID == achievementID {
			return &GameConfig.Achievements[i]
		}
	}
	return nil
}

// GetAllAchievementConfigs 获取所有成就配置
func GetAllAchievementConfigs() []*AchievementConfig {
	configs := make([]*AchievementConfig, 0, len(GameConfig.Achievements))
	for i := range GameConfig.Achievements {
		configs = append(configs, &GameConfig.Achievements[i])
	}
	return configs
}

// GetShopConfig 获取商店配置
func GetShopConfig(shopID uint32) *ShopConfig {
	for i := range GameConfig.Shops {
		if GameConfig.Shops[i].ID == shopID {
			return &GameConfig.Shops[i]
		}
	}
	return nil
}

// GetShopGoods 获取商店商品
func GetShopGoods(shopID uint32) []ShopGoodsConfig {
	var goods []ShopGoodsConfig
	for i := range GameConfig.ShopGoods {
		if GameConfig.ShopGoods[i].ShopID == shopID {
			goods = append(goods, GameConfig.ShopGoods[i])
		}
	}
	return goods
}

// GetActiveAnnouncements 获取当前有效公告
func GetActiveAnnouncements() []AnnouncementConfig {
	var announcements []AnnouncementConfig
	for i := range GameConfig.Announcements {
		announcements = append(announcements, GameConfig.Announcements[i])
	}
	return announcements
}

// GetServerConfig 获取服务器配置
func GetServerConfig(key string) *ServerConfigItem {
	for i := range GameConfig.ServerConfigs {
		if GameConfig.ServerConfigs[i].Key == key {
			return &GameConfig.ServerConfigs[i]
		}
	}
	return nil
}

// GetServerConfigValue 获取服务器配置值(字符串)
func GetServerConfigValue(key string, defaultVal string) string {
	config := GetServerConfig(key)
	if config == nil {
		return defaultVal
	}
	return config.Value
}

// ========== 怪物生成点配置访问函数（新增）==========

// GetMapSpawnConfig 获取指定地图的怪物生成配置
func GetMapSpawnConfig(mapID uint32) *MapSpawnConfig {
	if GameConfig == nil || GameConfig.SpawnPoints.Maps == nil {
		return nil
	}

	mapKey := fmt.Sprintf("%d", mapID)
	if config, exists := GameConfig.SpawnPoints.Maps[mapKey]; exists {
		return &config
	}
	return nil
}

// GetActiveSpawnPointsByMap 获取指定地图的所有激活的生成点
func GetActiveSpawnPointsByMap(mapID uint32) []SpawnPointConfig {
	config := GetMapSpawnConfig(mapID)
	if config == nil {
		return []SpawnPointConfig{}
	}

	var activePoints []SpawnPointConfig
	for _, point := range config.SpawnPoints {
		if point.IsActive {
			activePoints = append(activePoints, point)
		}
	}
	return activePoints
}

// GetAllSpawnPoints 获取所有地图的生成点配置（调试用）
func GetAllSpawnPoints() map[string]MapSpawnConfig {
	if GameConfig == nil || GameConfig.SpawnPoints.Maps == nil {
		return make(map[string]MapSpawnConfig)
	}
	return GameConfig.SpawnPoints.Maps
}

// ========== 共享数据结构（避免循环依赖）==========

// MonsterInfo 怪物运行时信息（供Battle和Monster包共享使用）
type MonsterInfo struct {
	ID        uint64 // 实例ID
	BaseID    uint32 // 基础ID
	Name      string // 名称
	Level     uint32 // 等级
	Type      uint8  // 类型: 0=普通, 1=精英, 2=BOSS
	MapID     uint32 // 地图ID
	X         int    // X坐标
	Y         int    // Y坐标
	CurrentHP int    // 当前HP
	MaxHP     int    // 最大HP
	Attack    int    // 攻击力
	Defense   int    // 防御力
	Speed     int    // 速度
	Status    uint8  // 状态: 0=空闲, 1=巡逻, 2=追击, 3=战斗, 4=死亡
}

// MonsterAttackResult 怪物攻击结果（供Battle和Monster包共享使用）
type MonsterAttackResult struct {
	MonsterID   uint64 `json:"monster_id"`
	TargetID    uint64 `json:"target_id"`
	Damage      int    `json:"damage"`
	IsCrit      bool   `json:"is_crit"`
	IsMiss      bool   `json:"is_miss"`
	PlayerHP    int    `json:"player_hp"`
	PlayerMaxHP int    `json:"player_max_hp"`
	IsDead      bool   `json:"is_dead"`
}

// AOETargetInfo AOE范围内的目标信息（用于范围伤害计算）
type AOETargetInfo struct {
	InstanceID uint64 `json:"instance_id"` // 怪物实例ID
	Name       string `json:"name"`        // 怪物名称
	X          int    `json:"x"`           // X坐标
	Y          int    `json:"y"`           // Y坐标
}
