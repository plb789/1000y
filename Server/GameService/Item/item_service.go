package item

import (
	"errors"
	"game-server/DBService/mysql"
	"game-server/GameService/Item/model"
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
func (s *Service) GetItemBase(itemID uint32) (*model.ItemBase, error) {
	var item model.ItemBase
	if err := mysql.DB.Where("id = ?", itemID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetAllItems 获取所有道具基础信息
func (s *Service) GetAllItems() ([]model.ItemBase, error) {
	var items []model.ItemBase
	err := mysql.DB.Order("type ASC, quality ASC, level_req ASC").Find(&items).Error
	return items, err
}

// GetItemsByType 获取指定类型道具
func (s *Service) GetItemsByType(itemType uint8) ([]model.ItemBase, error) {
	var items []model.ItemBase
	err := mysql.DB.Where("type = ?", itemType).Order("quality ASC, level_req ASC").Find(&items).Error
	return items, err
}

// AddItem 添加道具到背包
// 返回: 成功添加的格子索引, 错误信息
func (s *Service) AddItem(roleID uint64, itemID uint32, count uint32, isBind uint8) (int, error) {
	// 获取道具基础信息
	itemBase, err := s.GetItemBase(itemID)
	if err != nil {
		return -1, errors.New("道具不存在")
	}

	// 检查是否可堆叠
	if itemBase.StackMax > 1 {
		// 尝试合并到已有物品
		err = s.mergeItem(roleID, itemID, count, itemBase.StackMax, isBind)
		if err == nil {
			return -2, nil // -2表示合并成功
		}
	}

	// 查找空格子
	emptySlot, err := s.findEmptySlot(roleID)
	if err != nil {
		return -1, errors.New("背包已满")
	}

	// 创建新物品
	bagItem := model.RoleBag{
		RoleID:    roleID,
		GridIndex: emptySlot,
		ItemID:    itemID,
		Count:     count,
		IsBind:    isBind,
		GetTime:   time.Now(),
	}

	if itemBase.Type == 2 { // 装备
		durMax := 100
		bagItem.DurMax = &durMax
		bagItem.DurCurrent = &durMax
	}

	if err := mysql.DB.Create(&bagItem).Error; err != nil {
		return -1, err
	}

	return emptySlot, nil
}

// mergeItem 合并物品
func (s *Service) mergeItem(roleID uint64, itemID uint32, count uint32, stackMax uint32, isBind uint8) error {
	var bagItems []model.RoleBag
	mysql.DB.Where("role_id = ? AND item_id = ? AND is_bind = ?", roleID, itemID, isBind).
		Order("count ASC").
		Find(&bagItems)

	remaining := count
	for i := range bagItems {
		if remaining <= 0 {
			break
		}
		canAdd := stackMax - bagItems[i].Count
		if canAdd <= 0 {
			continue
		}
		addCount := canAdd
		if addCount > remaining {
			addCount = remaining
		}
		bagItems[i].Count += addCount
		remaining -= addCount
		mysql.DB.Save(&bagItems[i])
	}

	return nil
}

// findEmptySlot 查找空格子
func (s *Service) findEmptySlot(roleID uint64) (int, error) {
	// 获取所有已使用的格子
	var usedSlots []int
	mysql.DB.Model(&model.RoleBag{}).
		Where("role_id = ?", roleID).
		Pluck("grid_index", &usedSlots)

	// 找出空格子
	slotMap := make(map[int]bool)
	for _, slot := range usedSlots {
		slotMap[slot] = true
	}

	for i := 0; i < BagMaxSlots; i++ {
		if !slotMap[i] {
			return i, nil
		}
	}

	return -1, errors.New("背包已满")
}

// GetBagItems 获取角色背包所有物品
func (s *Service) GetBagItems(roleID uint64) ([]model.BagItemWithBase, error) {
	var items []model.BagItemWithBase
	err := mysql.DB.Table("role_bag").
		Select("role_bag.*, item_base.*").
		Joins("LEFT JOIN item_base ON role_bag.item_id = item_base.id").
		Where("role_bag.role_id = ?", roleID).
		Order("role_bag.grid_index ASC").
		Find(&items).Error
	return items, err
}

// GetBagItemByGrid 获取指定格子的物品
func (s *Service) GetBagItemByGrid(roleID uint64, gridIndex int) (*model.BagItemWithBase, error) {
	var item model.BagItemWithBase
	err := mysql.DB.Table("role_bag").
		Select("role_bag.*, item_base.*").
		Joins("LEFT JOIN item_base ON role_bag.item_id = item_base.id").
		Where("role_bag.role_id = ? AND role_bag.grid_index = ?", roleID, gridIndex).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// MoveItem 移动物品(整理背包)
func (s *Service) MoveItem(roleID uint64, fromGrid, toGrid int) error {
	// 检查目标格子是否为空
	var targetItem model.RoleBag
	err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, toGrid).First(&targetItem).Error
	if err == nil {
		return errors.New("目标格子不为空")
	}

	// 获取源物品
	var sourceItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, fromGrid).First(&sourceItem).Error; err != nil {
		return errors.New("源格子没有物品")
	}

	// 移动
	sourceItem.GridIndex = toGrid
	return mysql.DB.Save(&sourceItem).Error
}

// SplitItem 拆分物品
func (s *Service) SplitItem(roleID uint64, gridIndex int, count uint32) error {
	// 获取物品
	var bagItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, gridIndex).First(&bagItem).Error; err != nil {
		return errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return errors.New("道具不存在")
	}

	if itemBase.StackMax <= 1 {
		return errors.New("该物品不可拆分")
	}

	if bagItem.Count < count {
		return errors.New("物品数量不足")
	}

	// 查找空格子
	emptySlot, err := s.findEmptySlot(roleID)
	if err != nil {
		return errors.New("背包已满")
	}

	// 减少原物品数量
	bagItem.Count -= count
	mysql.DB.Save(&bagItem)

	// 创建新物品
	newItem := model.RoleBag{
		RoleID:    roleID,
		GridIndex: emptySlot,
		ItemID:    bagItem.ItemID,
		Count:     count,
		IsBind:    bagItem.IsBind,
		GetTime:   time.Now(),
	}
	return mysql.DB.Create(&newItem).Error
}

// UseItem 使用道具
func (s *Service) UseItem(roleID uint64, gridIndex int) error {
	// 获取物品
	var bagItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, gridIndex).First(&bagItem).Error; err != nil {
		return errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return errors.New("道具不存在")
	}

	// 检查是否为药品
	if itemBase.Type != 1 {
		return errors.New("该物品不可使用")
	}

	// 使用效果
	if itemBase.HpRestore > 0 || itemBase.MpRestore > 0 {
		// 这里应该调用角色服务来修改HP/MP
		// 由于避免循环引用,实际使用时通过API调用角色服务
	}

	// 消耗物品
	if bagItem.Count <= 1 {
		// 删除物品
		return mysql.DB.Delete(&bagItem).Error
	}

	// 减少数量
	bagItem.Count--
	return mysql.DB.Save(&bagItem).Error
}

// DiscardItem 丢弃物品
func (s *Service) DiscardItem(roleID uint64, gridIndex int) error {
	var bagItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, gridIndex).First(&bagItem).Error; err != nil {
		return errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return errors.New("道具不存在")
	}

	if itemBase.IsDropable != 1 {
		return errors.New("该物品不可丢弃")
	}

	return mysql.DB.Delete(&bagItem).Error
}

// SellItem 出售物品
func (s *Service) SellItem(roleID uint64, gridIndex int) (int, error) {
	var bagItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, gridIndex).First(&bagItem).Error; err != nil {
		return 0, errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return 0, errors.New("道具不存在")
	}

	if itemBase.IsSellable != 1 {
		return 0, errors.New("该物品不可出售")
	}

	// 计算售价
	sellPrice := itemBase.Price * int(bagItem.Count) / 2

	// 删除物品
	if err := mysql.DB.Delete(&bagItem).Error; err != nil {
		return 0, err
	}

	// 增加金币(这里应该调用角色服务)
	// roleService.AddGold(roleID, int64(sellPrice))

	return sellPrice, nil
}

// DestroyItem 销毁物品
func (s *Service) DestroyItem(roleID uint64, gridIndex int) error {
	var bagItem model.RoleBag
	if err := mysql.DB.Where("role_id = ? AND grid_index = ?", roleID, gridIndex).First(&bagItem).Error; err != nil {
		return errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return errors.New("道具不存在")
	}

	if itemBase.IsDestroyable != 1 {
		return errors.New("该物品不可销毁")
	}

	return mysql.DB.Delete(&bagItem).Error
}

// EquipItem 穿戴装备
func (s *Service) EquipItem(roleID uint64, bagItemID uint64) error {
	// 获取背包物品
	var bagItem model.RoleBag
	if err := mysql.DB.Where("id = ? AND role_id = ?", bagItemID, roleID).First(&bagItem).Error; err != nil {
		return errors.New("物品不存在")
	}

	// 获取道具基础信息
	itemBase, err := s.GetItemBase(bagItem.ItemID)
	if err != nil {
		return errors.New("道具不存在")
	}

	if itemBase.Type != 2 {
		return errors.New("该物品不是装备")
	}

	// 检查是否已有该类型装备
	var existingEquip model.RoleEquipment
	err = mysql.DB.Where("role_id = ? AND equip_type = ?", roleID, itemBase.EquipType).First(&existingEquip).Error

	if err == nil && existingEquip.BagItemID != nil {
		// 卸下原装备到背包
		var oldBagItem model.RoleBag
		mysql.DB.Where("id = ?", *existingEquip.BagItemID).First(&oldBagItem)
		
		// 放到原装备位置
		existingEquip.BagItemID = nil
		mysql.DB.Save(&existingEquip)
		
		// 如果原背包物品还在,放回背包
		if oldBagItem.ID > 0 {
			oldBagItem.GridIndex = bagItem.GridIndex
			mysql.DB.Save(&oldBagItem)
		}
	}

	// 创建新装备记录
	equip := model.RoleEquipment{
		RoleID:    roleID,
		EquipType: itemBase.EquipType,
		BagItemID: &bagItemID,
		EquipTime: time.Now(),
	}

	// 删除背包物品
	mysql.DB.Delete(&bagItem)

	// 如果已有装备记录,更新;否则创建
	if err == nil {
		equip.ID = existingEquip.ID
		return mysql.DB.Save(&equip).Error
	}

	return mysql.DB.Create(&equip).Error
}

// UnequipItem 卸下装备
func (s *Service) UnequipItem(roleID uint64, equipType uint8) error {
	// 获取装备记录
	var equip model.RoleEquipment
	if err := mysql.DB.Where("role_id = ? AND equip_type = ?", roleID, equipType).First(&equip).Error; err != nil {
		return errors.New("该装备位为空")
	}

	if equip.BagItemID == nil {
		return errors.New("该装备位为空")
	}

	// 查找空格子
	emptySlot, err := s.findEmptySlot(roleID)
	if err != nil {
		return errors.New("背包已满")
	}

	// 获取背包物品
	var bagItem model.RoleBag
	if err := mysql.DB.Where("id = ?", *equip.BagItemID).First(&bagItem).Error; err != nil {
		return errors.New("装备物品不存在")
	}

	// 移动到空格子
	bagItem.GridIndex = emptySlot
	if err := mysql.DB.Save(&bagItem).Error; err != nil {
		return err
	}

	// 清空装备记录
	equip.BagItemID = nil
	return mysql.DB.Save(&equip).Error
}

// GetEquippedItems 获取已穿戴装备
func (s *Service) GetEquippedItems(roleID uint64) ([]model.EquipmentWithBase, error) {
	var equips []model.EquipmentWithBase
	err := mysql.DB.Table("role_equipment").
		Select("role_equipment.*, item_base.*").
		Joins("LEFT JOIN item_base ON role_equipment.bag_item_id = role_bag.id AND role_bag.item_id = item_base.id").
		Joins("LEFT JOIN role_bag ON role_equipment.bag_item_id = role_bag.id").
		Where("role_equipment.role_id = ? AND role_equipment.bag_item_id IS NOT NULL", roleID).
		Order("role_equipment.equip_type ASC").
		Find(&equips).Error
	return equips, err
}

// GetEquipmentByType 获取指定位置的装备
func (s *Service) GetEquipmentByType(roleID uint64, equipType uint8) (*model.EquipmentWithBase, error) {
	var equip model.EquipmentWithBase
	err := mysql.DB.Table("role_equipment").
		Select("role_equipment.*, item_base.*").
		Joins("LEFT JOIN role_bag ON role_equipment.bag_item_id = role_bag.id").
		Joins("LEFT JOIN item_base ON role_bag.item_id = item_base.id").
		Where("role_equipment.role_id = ? AND role_equipment.equip_type = ?", roleID, equipType).
		First(&equip).Error
	if err != nil {
		return nil, err
	}
	return &equip, nil
}

// GetEmptySlotCount 获取背包空位数
func (s *Service) GetEmptySlotCount(roleID uint64) (int, error) {
	var count int64
	mysql.DB.Model(&model.RoleBag{}).Where("role_id = ?", roleID).Count(&count)
	return BagMaxSlots - int(count), nil
}

// ClearBag 清空背包
func (s *Service) ClearBag(roleID uint64) error {
	return mysql.DB.Where("role_id = ?", roleID).Delete(&model.RoleBag{}).Error
}
