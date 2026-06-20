/**
 * 物品数据管理器
 * 负责加载和管理游戏中的装备、物品数据
 */
class ItemDataManager {
  constructor() {
    this.items = {};
    this.weapons = {};
    this.armors = {};
    this.accessories = {};
    this.allItems = {};
  }

  /**
   * 加载物品数据
   */
  async loadData() {
    try {
      const response = await fetch('/Data/Items.json');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      
      // 按类型存储
      data.items.forEach(item => {
        this.items[item.id] = item;
        this.allItems[item.id] = item;
      });
      
      data.weapons.forEach(weapon => {
        this.weapons[weapon.id] = weapon;
        this.allItems[weapon.id] = weapon;
      });
      
      data.armors.forEach(armor => {
        this.armors[armor.id] = armor;
        this.allItems[armor.id] = armor;
      });
      
      data.accessories.forEach(accessory => {
        this.accessories[accessory.id] = accessory;
        this.allItems[accessory.id] = accessory;
      });
      
      console.log('物品数据加载完成');
    } catch (error) {
      console.error('加载物品数据失败:', error);
    }
  }

  /**
   * 根据ID获取物品信息
   */
  getItem(id) {
    return this.allItems[id] || null;
  }

  /**
   * 根据ID获取武器信息
   */
  getWeapon(id) {
    return this.weapons[id] || null;
  }

  /**
   * 根据ID获取防具信息
   */
  getArmor(id) {
    return this.armors[id] || null;
  }

  /**
   * 根据ID获取饰品信息
   */
  getAccessory(id) {
    return this.accessories[id] || null;
  }

  /**
   * 获取所有物品列表
   */
  getAllItems() {
    return Object.values(this.allItems);
  }

  /**
   * 获取指定类型的物品列表
   */
  getItemsByType(type) {
    return Object.values(this.allItems).filter(item => item.type === type);
  }

  /**
   * 获取物品品质颜色
   */
  getQualityColor(quality) {
    const colors = {
      1: '#9ca3af', // 普通 - 灰色
      2: '#22c55e', // 优秀 - 绿色
      3: '#3b82f6', // 精良 - 蓝色
      4: '#a855f7', // 史诗 - 紫色
      5: '#f59e0b'  // 传说 - 橙色
    };
    return colors[quality] || '#ffffff';
  }
}

// 创建全局实例
const itemDataManager = new ItemDataManager();

// 数据加载由 Game 统一控制，此处不再自动调用

// 导出到全局
window.ItemDataManager = ItemDataManager;
window.itemDataManager = itemDataManager;