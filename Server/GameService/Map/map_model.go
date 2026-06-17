package gamemap

// MapBase 地图基础表
type MapBase struct {
	ID          uint32  `gorm:"primaryKey;column:id" json:"id"`                   // 地图ID
	Name        string  `gorm:"column:name;size:50" json:"name"`                  // 地图名称
	Width       int     `gorm:"column:width" json:"width"`                        // 地图宽度(像素)
	Height      int     `gorm:"column:height" json:"height"`                      // 地图高度(像素)
	TileWidth   int     `gorm:"column:tile_width;default:48" json:"tile_width"`   // 瓦片宽度
	TileHeight  int     `gorm:"column:tile_height;default:48" json:"tile_height"` // 瓦片高度
	MapFile     string  `gorm:"column:map_file;size:100" json:"map_file"`         // 地图文件路径(.map)
	TilesetFile string  `gorm:"column:tileset_file;size:100" json:"tileset_file"` // 瓦片图集文件路径(.png)
	Music       string  `gorm:"column:music;size:100" json:"music"`               // 背景音乐
	PkAllowed   uint8   `gorm:"column:pk_allowed;default:1" json:"pk_allowed"`    // 是否允许PK
	ReviveMapID *uint32 `gorm:"column:revive_map_id" json:"revive_map_id"`        // 复活地图ID
	ReviveX     *int    `gorm:"column:revive_x" json:"revive_x"`                  // 复活X坐标
	ReviveY     *int    `gorm:"column:revive_y" json:"revive_y"`                  // 复活Y坐标
	LevelReq    uint32  `gorm:"column:level_req;default:0" json:"level_req"`      // 进入等级要求
	MinimapFile string  `gorm:"column:minimap_file;size:100" json:"minimap_file"` // 小地图文件
}

func (MapBase) TableName() string {
	return "map_base"
}

// MapBrief 地图简要信息
type MapBrief struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	LevelReq  uint32 `json:"level_req"`
	PkAllowed uint8  `json:"pk_allowed"`
}

// NPCBase NPC基础表
type NPCBase struct {
	ID         uint32  `gorm:"primaryKey;column:id" json:"id"`             // NPC ID
	Name       string  `gorm:"column:name;size:50" json:"name"`            // NPC名称
	Type       uint8   `gorm:"column:type" json:"type"`                    // 类型: 1=普通NPC, 2=商店NPC, 3=任务NPC, 4=仓库NPC, 5=传送NPC
	MapID      uint32  `gorm:"column:map_id;index" json:"map_id"`          // 所属地图ID
	X          int     `gorm:"column:x" json:"x"`                          // X坐标
	Y          int     `gorm:"column:y" json:"y"`                          // Y坐标
	Face       uint8   `gorm:"column:face;default:0" json:"face"`          // 朝向: 0=下, 1=左, 2=右, 3=上
	SpriteID   *int    `gorm:"column:sprite_id" json:"sprite_id"`          // 精灵ID
	DialogText string  `gorm:"column:dialog_text;text" json:"dialog_text"` // 对话文本
	ShopID     *uint32 `gorm:"column:shop_id" json:"shop_id"`              // 商店ID
}

func (NPCBase) TableName() string {
	return "npc_base"
}

// MonsterBase 怪物基础表
type MonsterBase struct {
	ID          uint32  `gorm:"primaryKey;column:id" json:"id"`                     // 怪物ID
	Name        string  `gorm:"column:name;size:50" json:"name"`                    // 怪物名称
	Level       uint32  `gorm:"column:level;default:1" json:"level"`                // 等级
	Type        uint8   `gorm:"column:type;default:0" json:"type"`                  // 类型: 0=普通, 1=精英, 2=BOSS
	MapID       uint32  `gorm:"column:map_id;index" json:"map_id"`                  // 所属地图ID
	Hp          int     `gorm:"column:hp" json:"hp"`                                // 生命值
	Attack      int     `gorm:"column:attack" json:"attack"`                        // 攻击力
	Defense     int     `gorm:"column:defense" json:"defense"`                      // 防御力
	Speed       int     `gorm:"column:speed" json:"speed"`                          // 速度
	Hit         int     `gorm:"column:hit;default:50" json:"hit"`                   // 命中率
	Dodge       int     `gorm:"column:dodge;default:10" json:"dodge"`               // 闪避率
	Crit        int     `gorm:"column:crit;default:5" json:"crit"`                  // 暴击率
	AIType      uint8   `gorm:"column:ai_type;default:0" json:"ai_type"`            // AI类型: 0=被动, 1=主动攻击, 2=巡逻
	AttackRange int     `gorm:"column:attack_range;default:3" json:"attack_range"`  // 警戒范围(格)
	ChaseRange  int     `gorm:"column:chase_range;default:5" json:"chase_range"`    // 追击范围(格)
	GoldMin     int     `gorm:"column:gold_min;default:0" json:"gold_min"`          // 金币掉落下限
	GoldMax     int     `gorm:"column:gold_max;default:0" json:"gold_max"`          // 金币掉落上限
	Exp         int     `gorm:"column:exp;default:0" json:"exp"`                    // 经验掉落
	DropGroupID *uint32 `gorm:"column:drop_group_id" json:"drop_group_id"`          // 掉落组ID
	SpriteID    *int    `gorm:"column:sprite_id" json:"sprite_id"`                  // 精灵ID
	RespawnTime int     `gorm:"column:respawn_time;default:60" json:"respawn_time"` // 复活时间(秒)
}

func (MonsterBase) TableName() string {
	return "monster_base"
}

// NPCType NPC类型常量
var NPCTypeName = map[uint8]string{
	1: "普通NPC",
	2: "商店NPC",
	3: "任务NPC",
	4: "仓库NPC",
	5: "传送NPC",
}

// MonsterType 怪物类型常量
var MonsterTypeName = map[uint8]string{
	0: "普通",
	1: "精英",
	2: "BOSS",
}

// DropGroup 掉落组表
type DropGroup struct {
	ID         uint32 `gorm:"primaryKey;column:id" json:"id"`             // 掉落组ID
	MonsterID  uint32 `gorm:"column:monster_id;index" json:"monster_id"` // 怪物ID
	ItemID     uint32 `gorm:"column:item_id" json:"item_id"`             // 道具ID
	DropRate   uint32 `gorm:"column:drop_rate" json:"drop_rate"`         // 掉落概率(万分比)
	DropMin    uint32 `gorm:"column:drop_min;default:1" json:"drop_min"` // 最小掉落数量
	DropMax    uint32 `gorm:"column:drop_max;default:1" json:"drop_max"` // 最大掉落数量
}

func (DropGroup) TableName() string {
	return "drop_group"
}
