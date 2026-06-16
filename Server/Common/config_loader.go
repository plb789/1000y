package common

import (
	"encoding/json"
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
	Shops         []ShopConfig         // 商店配置
	ShopGoods     []ShopGoodsConfig    // 商品配置
	Announcements []AnnouncementConfig // 公告配置
	ServerConfigs []ServerConfigItem   // 服务器配置
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
	EquipType     uint8  `json:"equip_type"`  // 装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链
	WeaponType    uint8  `json:"weapon_type"` // 武器类型: 0=徒手, 1=剑, 2=刀, 3=枪, 4=斧, 5=拳 (仅武器有效)
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
	Type         uint8  `json:"type"`
	Duration     int    `json:"duration"`
	HpChange     int    `json:"hp_change"`
	MpChange     int    `json:"mp_change"`
	AttackChange int    `json:"attack_change"`
	DefChange    int    `json:"defense_change"`
	SpeedChange  int    `json:"speed_change"`
	HitChange    int    `json:"hit_change"`
	DodgeChange  int    `json:"dodge_change"`
	CritChange   int    `json:"crit_change"`
	CanCancel    uint8  `json:"can_cancel"`
	StackMax     int    `json:"stack_max"`
	Description  string `json:"description"`
}

// QuestBaseConfig 任务配置
type QuestBaseConfig struct {
	ID              uint32  `json:"id"`
	Name            string  `json:"name"`
	Type            uint8   `json:"type"`
	LevelReq        uint32  `json:"level_req"`
	Repeatable      uint8   `json:"repeatable"`
	TargetType      uint8   `json:"target_type"`
	TargetID        uint32  `json:"target_id"`
	TargetCount     int     `json:"target_count"`
	RewardExp       uint64  `json:"reward_exp"`
	RewardGold      uint64  `json:"reward_gold"`
	RewardItemID    *uint32 `json:"reward_item_id"`
	RewardItemCount *int    `json:"reward_item_count"`
	Description     string  `json:"description"`
	NPCAcceptID     *uint32 `json:"npc_accept_id"`
	NPCCompleteID   *uint32 `json:"npc_complete_id"`
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
