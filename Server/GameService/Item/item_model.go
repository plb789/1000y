package item

import (
	"time"
)

// ItemBase 道具基础表
type ItemBase struct {
	ID         uint32  `gorm:"primaryKey;column:id" json:"id"`                   // 道具ID
	Name       string  `gorm:"column:name;size:50" json:"name"`                  // 道具名称
	Type       uint8   `gorm:"column:type" json:"type"`                           // 类型: 1=药品, 2=装备, 3=材料, 4=任务, 5=秘籍, 6=时装, 7=货币
	SubType    uint8   `gorm:"column:sub_type" json:"sub_type"`                  // 子类型
	Quality    uint8   `gorm:"column:quality;default:1" json:"quality"`          // 品质: 1=白, 2=绿, 3=蓝, 4=紫, 5=橙
	LevelReq   uint32  `gorm:"column:level_req;default:0" json:"level_req"`      // 等级需求
	StackMax   uint32  `gorm:"column:stack_max;default:99" json:"stack_max"`    // 最大堆叠数量
	Price      int     `gorm:"column:price;default:0" json:"price"`             // 售价
	Description string `gorm:"column:description;size:255" json:"description"`     // 描述

	// 装备属性
	EquipType    uint8 `gorm:"column:equip_type" json:"equip_type"`       // 装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链
	HpBonus      int   `gorm:"column:hp_bonus;default:0" json:"hp_bonus"`     // 生命加成
	MpBonus      int   `gorm:"column:mp_bonus;default:0" json:"mp_bonus"`     // 内力加成
	AttackBonus  int   `gorm:"column:attack_bonus;default:0" json:"attack_bonus"`  // 攻击加成
	DefenseBonus int   `gorm:"column:defense_bonus;default:0" json:"defense_bonus"` // 防御加成
	SpeedBonus   int   `gorm:"column:speed_bonus;default:0" json:"speed_bonus"`   // 速度加成

	// 药品效果
	HpRestore int `gorm:"column:hp_restore;default:0" json:"hp_restore"`   // 恢复生命
	MpRestore int `gorm:"column:mp_restore;default:0" json:"mp_restore"`   // 恢复内力
	BuffID   uint32 `gorm:"column:buff_id" json:"buff_id"`               // 使用后获得BUFF

	Icon      string `gorm:"column:icon;size:100" json:"icon"`      // 图标资源路径
	Model     string `gorm:"column:model;size:100" json:"model"`   // 模型资源路径
	IsDropable uint8 `gorm:"column:is_dropable;default:1" json:"is_dropable"`   // 是否可丢弃
	IsSellable uint8 `gorm:"column:is_sellable;default:1" json:"is_sellable"`   // 是否可出售
	IsDestroyable uint8 `gorm:"column:is_destroyable;default:1" json:"is_destroyable"` // 是否可销毁
	IsBind    uint8 `gorm:"column:is_bind;default:0" json:"is_bind"`   // 是否绑定
}

func (ItemBase) TableName() string {
	return "item_base"
}

// RoleBag 角色背包表
type RoleBag struct {
	ID           uint64    `gorm:"primaryKey;column:id" json:"id"`               // 记录ID
	RoleID       uint64    `gorm:"column:role_id;index" json:"role_id"`         // 角色ID
	GridIndex    int       `gorm:"column:grid_index" json:"grid_index"`         // 背包格子索引(0-79)
	ItemID       uint32    `gorm:"column:item_id" json:"item_id"`               // 道具ID
	Count        uint32    `gorm:"column:count;default:1" json:"count"`         // 数量
	IsBind       uint8     `gorm:"column:is_bind;default:0" json:"is_bind"`     // 是否绑定: 0=未绑定, 1=绑定
	EnhanceLevel int       `gorm:"column:enhance_level;default:0" json:"enhance_level"` // 强化等级
	DurMax       *int      `gorm:"column:dur_max" json:"dur_max"`                // 最大耐久(装备)
	DurCurrent   *int      `gorm:"column:dur_current" json:"dur_current"`       // 当前耐久(装备)
	GetTime      time.Time `gorm:"column:get_time" json:"get_time"`               // 获得时间
}

func (RoleBag) TableName() string {
	return "role_bag"
}

// RoleEquipment 角色装备表
type RoleEquipment struct {
	ID         uint64    `gorm:"primaryKey;column:id" json:"id"`               // 记录ID
	RoleID     uint64    `gorm:"column:role_id;index" json:"role_id"`           // 角色ID
	EquipType  uint8     `gorm:"column:equip_type" json:"equip_type"`           // 装备位置: 1=武器, 2=衣服, 3=头盔, 4=护腕, 5=腰带, 6=鞋子, 7=戒指, 8=项链
	BagItemID  *uint64   `gorm:"column:bag_item_id" json:"bag_item_id"`         // 背包物品ID(NULL表示空位)
	EquipTime  time.Time `gorm:"column:equip_time" json:"equip_time"`           // 穿戴时间
}

func (RoleEquipment) TableName() string {
	return "role_equipment"
}

// BagItemWithBase 背包物品详情(关联道具基础信息)
type BagItemWithBase struct {
	RoleBag
	ItemBase
}

// EquipmentWithBase 装备详情(关联道具基础信息)
type EquipmentWithBase struct {
	RoleEquipment
	ItemBase
}

// ItemType 道具类型常量
var ItemTypeName = map[uint8]string{
	1: "药品",
	2: "装备",
	3: "材料",
	4: "任务",
	5: "秘籍",
	6: "时装",
	7: "货币",
}

// ItemQuality 道具品质常量
var ItemQualityName = map[uint8]string{
	1: "白色",
	2: "绿色",
	3: "蓝色",
	4: "紫色",
	5: "橙色",
}

// EquipTypeName 装备位置常量
var EquipTypeName = map[uint8]string{
	1: "武器",
	2: "衣服",
	3: "头盔",
	4: "护腕",
	5: "腰带",
	6: "鞋子",
	7: "戒指",
	8: "项链",
}

// AddItemRequest 添加道具请求
type AddItemRequest struct {
	RoleID  uint64 `json:"role_id" binding:"required"`
	ItemID  uint32 `json:"item_id" binding:"required"`
	Count   uint32 `json:"count"`
	IsBind  uint8  `json:"is_bind"`
}

// MoveItemRequest 移动道具请求
type MoveItemRequest struct {
	FromGrid int `json:"from_grid" binding:"required"`
	ToGrid   int `json:"to_grid" binding:"required"`
}

// SplitItemRequest 拆分道具请求
type SplitItemRequest struct {
	GridIndex int    `json:"grid_index" binding:"required"`
	Count     uint32 `json:"count" binding:"required"`
}

// UseItemRequest 使用道具请求
type UseItemRequest struct {
	GridIndex int `json:"grid_index" binding:"required"`
}

// EquipItemRequest 穿戴装备请求
type EquipItemRequest struct {
	BagItemID uint64 `json:"bag_item_id" binding:"required"`
}

// TradeRequest 交易请求
type TradeRequest struct {
	TargetRoleID uint64          `json:"target_role_id" binding:"required"`
	Items        []TradeItemInfo `json:"items" binding:"required"`
	Gold         int64          `json:"gold"` // 我方出金币
}

type TradeItemInfo struct {
	GridIndex int    `json:"grid_index"`
	ItemID    uint32 `json:"item_id"`
	Count     uint32 `json:"count"`
}
