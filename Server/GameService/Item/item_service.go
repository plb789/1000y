package item

import (
	"errors"
	"fmt"
	common "game-server/Common"
	"time"
)

const (
	BagMaxSlots = 80 // 背包最大格子数
)

// Service 道具服务
type Service struct{}

// NewService 创建道具服务实例
func NewService() *Service {
	return &Service{}
}

// GetItemBase 获取道具基础信息
func (s *Service) GetItemBase(itemID uint32) (map[string]interface{}, error) {
	config := common.GetItemConfig(itemID)
	if config == nil {
		return nil, errors.New("道具不存在")
	}
	return map[string]interface{}{
		"id":             config.ID,
		"name":           config.Name,
		"type":           config.Type,
		"sub_type":       config.SubType,
		"quality":        config.Quality,
		"level_req":      config.LevelReq,
		"stack_max":      config.StackMax,
		"price":          config.Price,
		"description":    config.Description,
		"equip_type":     config.EquipType,
		"hp_bonus":       config.HpBonus,
		"mp_bonus":       config.MpBonus,
		"attack_bonus":   config.AttackBonus,
		"defense_bonus":  config.DefenseBonus,
		"speed_bonus":    config.SpeedBonus,
		"hp_restore":     config.HpRestore,
		"mp_restore":     config.MpRestore,
		"buff_id":        config.BuffID,
		"icon":           config.Icon,
		"model":          config.Model,
		"is_dropable":    config.IsDropable,
		"is_sellable":    config.IsSellable,
		"is_destroyable": config.IsDestroyable,
		"is_bind":        config.IsBind,
	}, nil
}

// GetAllItems 获取所有道具基础信息
func (s *Service) GetAllItems() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, config := range common.GetAllItemConfig() {
		result = append(result, map[string]interface{}{
			"id":             config.ID,
			"name":           config.Name,
			"type":           config.Type,
			"sub_type":       config.SubType,
			"quality":        config.Quality,
			"level_req":      config.LevelReq,
			"stack_max":      config.StackMax,
			"price":          config.Price,
			"description":    config.Description,
			"equip_type":     config.EquipType,
			"hp_bonus":       config.HpBonus,
			"mp_bonus":       config.MpBonus,
			"attack_bonus":   config.AttackBonus,
			"defense_bonus":  config.DefenseBonus,
			"speed_bonus":    config.SpeedBonus,
			"hp_restore":     config.HpRestore,
			"mp_restore":     config.MpRestore,
			"buff_id":        config.BuffID,
			"icon":           config.Icon,
			"model":          config.Model,
			"is_dropable":    config.IsDropable,
			"is_sellable":    config.IsSellable,
			"is_destroyable": config.IsDestroyable,
			"is_bind":        config.IsBind,
		})
	}
	return result, nil
}

// GetItemsByType 获取指定类型道具
func (s *Service) GetItemsByType(itemType uint8) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, config := range common.GetAllItemConfig() {
		if config.Type == itemType {
			result = append(result, map[string]interface{}{
				"id":             config.ID,
				"name":           config.Name,
				"type":           config.Type,
				"sub_type":       config.SubType,
				"quality":        config.Quality,
				"level_req":      config.LevelReq,
				"stack_max":      config.StackMax,
				"price":          config.Price,
				"description":    config.Description,
				"equip_type":     config.EquipType,
				"hp_bonus":       config.HpBonus,
				"mp_bonus":       config.MpBonus,
				"attack_bonus":   config.AttackBonus,
				"defense_bonus":  config.DefenseBonus,
				"speed_bonus":    config.SpeedBonus,
				"hp_restore":     config.HpRestore,
				"mp_restore":     config.MpRestore,
				"buff_id":        config.BuffID,
				"icon":           config.Icon,
				"model":          config.Model,
				"is_dropable":    config.IsDropable,
				"is_sellable":    config.IsSellable,
				"is_destroyable": config.IsDestroyable,
				"is_bind":        config.IsBind,
			})
		}
	}
	return result, nil
}

// AddItem 添加道具到背包
// 返回: 成功添加的格子索引, 错误信息
func (s *Service) AddItem(roleID uint64, itemID uint32, count uint32, isBind uint8) (int, error) {
	return common.DBItemAdd(roleID, itemID, count, isBind)
}

// GetBagItems 获取角色背包所有物品
// 返回格式：合并数据库数据 + Items.json 配置（icon、name、quality等）
func (s *Service) GetBagItems(roleID uint64) ([]map[string]interface{}, error) {
	// 从DBService获取原始背包数据
	rawItems, err := common.DBItemGetBag(roleID)
	if err != nil {
		return nil, err
	}

	// ★ 合并 Items.json 配置信息（确保前端能正确显示）
	var result []map[string]interface{}
	for _, item := range rawItems {
		enhancedItem := s.enhanceBagItem(item)
		result = append(result, enhancedItem)
	}

	return result, nil
}

// enhanceBagItem 增强背包物品数据（合并配置信息）
func (s *Service) enhanceBagItem(rawItem map[string]interface{}) map[string]interface{} {
	itemID := uint32(0)
	if id, ok := rawItem["item_id"].(float64); ok {
		itemID = uint32(id)
	}

	// 从 Items.json 获取物品配置
	config := common.GetItemConfig(itemID)

	// 构建增强后的物品对象
	result := map[string]interface{}{
		// 数据库字段
		"id":         rawItem["id"],
		"role_id":    rawItem["role_id"],
		"grid_index": rawItem["grid_index"],
		"item_id":    itemID,
		"count":      rawItem["count"],
		"is_bind":    rawItem["is_bind"],
	}

	// 合并配置信息（如果存在）
	if config != nil {
		result["name"] = config.Name
		result["icon"] = config.Icon
		result["quality"] = config.Quality
		result["type"] = config.Type
		result["type_name"] = s.getTypeName(config.Type)
		result["description"] = config.Description
		result["level_req"] = config.LevelReq
		result["stack_max"] = config.StackMax
		result["price"] = config.Price

		// 装备属性
		if config.EquipType > 0 {
			result["can_equip"] = true
			result["equip_type"] = config.EquipType
			result["equip_pos"] = s.getEquipPosName(config.EquipType)
			result["attrs"] = map[string]interface{}{
				"hp_bonus":      config.HpBonus,
				"mp_bonus":      config.MpBonus,
				"attack_bonus":  config.AttackBonus,
				"defense_bonus": config.DefenseBonus,
				"speed_bonus":   config.SpeedBonus,
			}
		}

		// 消耗品属性
		if config.HpRestore > 0 || config.MpRestore > 0 {
			result["can_use"] = true
			result["hp_restore"] = config.HpRestore
			result["mp_restore"] = config.MpRestore
		}
	} else {
		// 配置不存在时的降级处理
		result["name"] = fmt.Sprintf("未知物品#%d", itemID)
		result["icon"] = "📦"
		result["quality"] = 1
		result["type"] = 0
		result["type_name"] = "未知"
	}

	return result
}

// GetEquippedItems 获取已穿戴装备
// 返回格式：合并数据库数据 + Items.json 配置
func (s *Service) GetEquippedItems(roleID uint64) ([]map[string]interface{}, error) {
	// 从DBService获取原始装备数据
	rawEquips, err := common.DBItemGetEquipped(roleID)
	if err != nil {
		return nil, err
	}

	// ★ 合并 Items.json 配置信息
	var result []map[string]interface{}
	for _, equip := range rawEquips {
		enhancedEquip := s.enhanceEquipItem(equip)
		result = append(result, enhancedEquip)
	}

	return result, nil
}

// enhanceEquipItem 增强装备数据（合并配置信息）
func (s *Service) enhanceEquipItem(rawEquip map[string]interface{}) map[string]interface{} {
	equipType := uint8(0)
	if et, ok := rawEquip["equip_type"].(float64); ok {
		equipType = uint8(et)
	}

	config := common.GetItemConfigByEquipType(equipType)

	result := map[string]interface{}{
		"id":          rawEquip["id"],
		"role_id":     rawEquip["role_id"],
		"equip_type":  equipType,
		"equip_pos":   s.getEquipPosName(equipType),
		"bag_item_id": rawEquip["bag_item_id"],
	}

	if config != nil {
		result["name"] = config.Name
		result["icon"] = config.Icon
		result["quality"] = config.Quality
		result["item_id"] = config.ID
		result["attrs"] = map[string]interface{}{
			"hp_bonus":      config.HpBonus,
			"mp_bonus":      config.MpBonus,
			"attack_bonus":  config.AttackBonus,
			"defense_bonus": config.DefenseBonus,
			"speed_bonus":   config.SpeedBonus,
		}
	} else {
		result["name"] = fmt.Sprintf("空装备槽#%d", equipType)
		result["icon"] = "⬜"
		result["quality"] = 1
	}

	return result
}

// getTypeName 获取类型名称
func (s *Service) getTypeName(itemType uint8) string {
	names := map[uint8]string{
		1: "消耗品",
		2: "装备",
		3: "材料",
		4: "任务物品",
	}
	if name, ok := names[itemType]; ok {
		return name
	}
	return "其他"
}

// getEquipPosName 获取装备位置名称
func (s *Service) getEquipPosName(equipType uint8) string {
	names := map[uint8]string{
		1: "weapon",
		2: "armor",
		3: "helmet",
		4: "boots",
		5: "ring",
		6: "necklace",
	}
	if name, ok := names[equipType]; ok {
		return name
	}
	return "unknown"
}

// GetBagItemByGrid 获取指定格子的物品
func (s *Service) GetBagItemByGrid(roleID uint64, gridIndex int) (*map[string]interface{}, error) {
	bagItems, err := common.DBItemGetBag(roleID)
	if err != nil {
		return nil, err
	}

	for _, item := range bagItems {
		if g, ok := item["grid_index"].(float64); ok && int(g) == gridIndex {
			return &item, nil
		}
	}
	return nil, errors.New("物品不存在")
}

// MoveItem 移动物品(整理背包)
func (s *Service) MoveItem(roleID uint64, fromGrid, toGrid int) error {
	return common.DBItemMove(roleID, fromGrid, toGrid)
}

// MergeItem 合并/堆叠物品
func (s *Service) MergeItem(roleID uint64, sourceItemId, targetItemId uint64, count uint32) error {
	return common.DBItemMerge(roleID, sourceItemId, targetItemId, count)
}

// SplitItem 拆分物品
func (s *Service) SplitItem(roleID uint64, gridIndex int, count uint32) error {
	return common.DBItemSplit(roleID, gridIndex, count)
}

// UseItem 使用道具
func (s *Service) UseItem(roleID uint64, gridIndex int) error {
	return common.DBItemUse(roleID, gridIndex)
}

// DiscardItem 丢弃物品
func (s *Service) DiscardItem(roleID uint64, gridIndex int) error {
	return common.DBItemDiscard(roleID, gridIndex)
}

// DestroyItem 销毁物品(不可恢复，与丢弃语义一致，由DBService处理)
func (s *Service) DestroyItem(roleID uint64, gridIndex int) error {
	return common.DBItemDiscard(roleID, gridIndex)
}

// SellItem 出售物品
func (s *Service) SellItem(roleID uint64, gridIndex int) (int, error) {
	return common.DBItemSell(roleID, gridIndex)
}

// EquipItem 穿戴装备
func (s *Service) EquipItem(roleID uint64, bagItemID uint64) error {
	return common.DBItemEquip(roleID, bagItemID)
}

// UnequipItem 卸下装备
func (s *Service) UnequipItem(roleID uint64, equipType uint8) error {
	return common.DBItemUnequip(roleID, equipType)
}

// GetEquipmentByType 获取指定位置的装备
func (s *Service) GetEquipmentByType(roleID uint64, equipType uint8) (*map[string]interface{}, error) {
	equips, err := common.DBItemGetEquipped(roleID)
	if err != nil {
		return nil, err
	}

	for _, equip := range equips {
		if e, ok := equip["equip_type"].(float64); ok && uint8(e) == equipType {
			return &equip, nil
		}
	}
	return nil, errors.New("该装备位为空")
}

// GetEmptySlotCount 获取背包空位数
func (s *Service) GetEmptySlotCount(roleID uint64) (int, error) {
	return common.DBItemGetEmptyCount(roleID)
}

// ClearBag 清空背包(通过DBService API)
func (s *Service) ClearBag(roleID uint64) error {
	bagItems, err := common.DBItemGetBag(roleID)
	if err != nil {
		return err
	}

	for _, item := range bagItems {
		if gridIndex, ok := item["grid_index"].(float64); ok {
			common.DBItemDiscard(roleID, int(gridIndex))
		}
	}
	return nil
}

// findEmptySlot 查找空格子
func (s *Service) findEmptySlot(roleID uint64) (int, error) {
	bagItems, err := common.DBItemGetBag(roleID)
	if err != nil {
		return -1, err
	}

	// 收集已使用的格子
	usedSlots := make(map[int]bool)
	for _, item := range bagItems {
		if gridIndex, ok := item["grid_index"].(float64); ok {
			usedSlots[int(gridIndex)] = true
		}
	}

	// 找出空格子
	for i := 0; i < BagMaxSlots; i++ {
		if !usedSlots[i] {
			return i, nil
		}
	}

	return -1, errors.New("背包已满")
}

// GetItemBaseInfo 获取道具基础信息(内部转换用)
type ItemBaseInfo struct {
	ID            uint32 `json:"id"`
	Name          string `json:"name"`
	Type          uint8  `json:"type"`
	Quality       uint8  `json:"quality"`
	LevelReq      uint32 `json:"level_req"`
	StackMax      uint32 `json:"stack_max"`
	Price         int    `json:"price"`
	HpRestore     int    `json:"hp_restore"`
	MpRestore     int    `json:"mp_restore"`
	IsDropable    bool   `json:"is_dropable"`
	IsSellable    bool   `json:"is_sellable"`
	IsDestroyable bool   `json:"is_destroyable"`
	EquipType     uint8  `json:"equip_type"`
	Description   string `json:"description"`
}

// GetBagItemInfo 获取背包物品信息(内部转换用)
type BagItemInfo struct {
	ID         uint64    `json:"id"`
	RoleID     uint64    `json:"role_id"`
	GridIndex  int       `json:"grid_index"`
	ItemID     uint32    `json:"item_id"`
	Count      uint32    `json:"count"`
	IsBind     uint8     `json:"is_bind"`
	GetTime    time.Time `json:"get_time"`
	DurMax     *int      `json:"dur_max,omitempty"`
	DurCurrent *int      `json:"dur_current,omitempty"`
}

// getFloatValue 安全获取float64值
func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// parseItemBase 解析道具基础信息
func parseItemBase(data map[string]interface{}) ItemBaseInfo {
	info := ItemBaseInfo{}
	if v, ok := data["id"].(float64); ok {
		info.ID = uint32(v)
	}
	if v, ok := data["name"].(string); ok {
		info.Name = v
	}
	if v, ok := data["type"].(float64); ok {
		info.Type = uint8(v)
	}
	if v, ok := data["quality"].(float64); ok {
		info.Quality = uint8(v)
	}
	if v, ok := data["level_req"].(float64); ok {
		info.LevelReq = uint32(v)
	}
	if v, ok := data["stack_max"].(float64); ok {
		info.StackMax = uint32(v)
	}
	if v, ok := data["price"].(float64); ok {
		info.Price = int(v)
	}
	if v, ok := data["hp_restore"].(float64); ok {
		info.HpRestore = int(v)
	}
	if v, ok := data["mp_restore"].(float64); ok {
		info.MpRestore = int(v)
	}
	if v, ok := data["is_dropable"].(float64); ok {
		info.IsDropable = v == 1
	}
	if v, ok := data["is_sellable"].(float64); ok {
		info.IsSellable = v == 1
	}
	if v, ok := data["is_destroyable"].(float64); ok {
		info.IsDestroyable = v == 1
	}
	if v, ok := data["equip_type"].(float64); ok {
		info.EquipType = uint8(v)
	}
	if v, ok := data["description"].(string); ok {
		info.Description = v
	}
	return info
}

// parseBagItem 解析背包物品信息
func parseBagItem(data map[string]interface{}) BagItemInfo {
	info := BagItemInfo{}
	if v, ok := data["id"].(float64); ok {
		info.ID = uint64(v)
	}
	if v, ok := data["role_id"].(float64); ok {
		info.RoleID = uint64(v)
	}
	if v, ok := data["grid_index"].(float64); ok {
		info.GridIndex = int(v)
	}
	if v, ok := data["item_id"].(float64); ok {
		info.ItemID = uint32(v)
	}
	if v, ok := data["count"].(float64); ok {
		info.Count = uint32(v)
	}
	if v, ok := data["is_bind"].(float64); ok {
		info.IsBind = uint8(v)
	}
	return info
}
