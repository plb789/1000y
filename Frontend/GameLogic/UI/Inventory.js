/**
 * 背包管理器
 * 管理背包UI、物品拖拽、物品使用、装备穿戴等功能
 */
class Inventory {
  constructor(game) {
    this.game = game;
    this.isOpen = false;
    this.container = null;
    this.tabs = ['物品', '装备', '任务']; // 标签页
    this.currentTab = 0;
    this.gridSize = { cols: 6, rows: 5 }; // 背包网格扩展为6x5
    this.itemSlots = [];
    this.equipSlots = {}; // 装备栏位
    this.isDragging = false;
    this.draggedItem = null;
    this.draggedSlot = null;
    this.draggedType = null; // 'item' or 'equip'
    this.dragGhost = null; // 拖拽虚影

    // 物品数据
    this.items = [];
    this.equipments = [];
    this.questItems = [];

    // 排序模式
    this.sortMode = 'default'; // default, quality, type, name, count
    
    // 筛选模式
    this.filterType = 'all'; // all, potion, material, equipment, quest
    
    // 搜索关键词
    this.searchKeyword = '';
    
    // 装备栏定义
    this.equipPositions = {
      weapon: { name: '武器', icon: '⚔️', slot: 0 },
      helmet: { name: '头盔', icon: '🧢', slot: 1 },
      armor: { name: '衣服', icon: '👕', slot: 2 },
      necklace: { name: '项链', icon: '📿', slot: 3 },
      ring: { name: '戒指', icon: '💍', slot: 4 },
      boots: { name: '鞋子', icon: '👢', slot: 5 }
    };
    
    // 初始化
    this.create();
  }
  
  /**
   * 创建背包面板
   */
  create() {
    // 创建遮罩层
    this.overlay = document.createElement('div');
    this.overlay.id = 'inventory-overlay';
    this.overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.5);
      z-index: 900;
      display: none;
    `;
    this.overlay.addEventListener('click', (e) => {
      if (e.target === this.overlay) {
        this.hide();
      }
    });
    document.body.appendChild(this.overlay);
    
    // 创建背包容器
    this.container = document.createElement('div');
    this.container.id = 'inventory-panel';
    this.container.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      width: 480px;
      background: rgba(20, 20, 30, 0.95);
      border: 2px solid #e94560;
      border-radius: 15px;
      z-index: 901;
      display: none;
      font-family: 'Microsoft YaHei', sans-serif;
      box-shadow: 0 0 30px rgba(233, 69, 96, 0.3);
    `;
    
    this.container.innerHTML = this.createHTML();
    document.body.appendChild(this.container);
    
    this.initSlots();
    this.bindEvents();
  }
  
  /**
   * 创建背包HTML结构
   */
  createHTML() {
    return `
      <!-- 标题栏 -->
      <div style="display: flex; justify-content: space-between; align-items: center; padding: 15px 20px; border-bottom: 1px solid rgba(233, 69, 96, 0.3);">
        <div style="color: #e94560; font-size: 18px; font-weight: bold;">背 包</div>
        <div style="display: flex; gap: 15px; align-items: center;">
          <span style="color: #999; font-size: 12px;">💰 <span id="inventory-gold">0</span></span>
          <button id="inventory-close" style="
            background: transparent;
            border: none;
            color: #999;
            font-size: 24px;
            cursor: pointer;
            padding: 0;
            line-height: 1;
          ">×</button>
        </div>
      </div>
      
      <!-- 标签页 -->
      <div style="display: flex; padding: 10px 20px; gap: 10px; border-bottom: 1px solid rgba(233, 69, 96, 0.2);">
        <button class="inv-tab active" data-tab="0" style="
          flex: 1;
          padding: 8px 15px;
          background: rgba(233, 69, 96, 0.2);
          border: 1px solid #e94560;
          border-radius: 6px;
          color: #e94560;
          cursor: pointer;
          font-size: 14px;
        ">物品</button>
        <button class="inv-tab" data-tab="1" style="
          flex: 1;
          padding: 8px 15px;
          background: rgba(74, 85, 104, 0.3);
          border: 1px solid #4a5568;
          border-radius: 6px;
          color: #999;
          cursor: pointer;
          font-size: 14px;
        ">装备</button>
        <button class="inv-tab" data-tab="2" style="
          flex: 1;
          padding: 8px 15px;
          background: rgba(74, 85, 104, 0.3);
          border: 1px solid #4a5568;
          border-radius: 6px;
          color: #999;
          cursor: pointer;
          font-size: 14px;
        ">任务</button>
      </div>
      
      <!-- 搜索和筛选栏 -->
      <div id="inv-search-bar" style="display: flex; padding: 10px 20px; gap: 10px; border-bottom: 1px solid rgba(233, 69, 96, 0.1);">
        <input type="text" id="inv-search-input" placeholder="搜索物品..." style="
          flex: 1;
          padding: 6px 12px;
          background: rgba(45, 55, 72, 0.6);
          border: 1px solid #4a5568;
          border-radius: 6px;
          color: #fff;
          font-size: 12px;
          outline: none;
        ">
        <select id="inv-filter-select" style="
          padding: 6px 10px;
          background: rgba(45, 55, 72, 0.6);
          border: 1px solid #4a5568;
          border-radius: 6px;
          color: #fff;
          font-size: 12px;
          cursor: pointer;
        ">
          <option value="all">全部类型</option>
          <option value="potion">药品</option>
          <option value="material">材料</option>
          <option value="equipment">装备</option>
          <option value="quest">任务物品</option>
        </select>
        <select id="inv-sort-select" style="
          padding: 6px 10px;
          background: rgba(45, 55, 72, 0.6);
          border: 1px solid #4a5568;
          border-radius: 6px;
          color: #fff;
          font-size: 12px;
          cursor: pointer;
        ">
          <option value="default">默认排序</option>
          <option value="quality">品质优先</option>
          <option value="type">按类型</option>
          <option value="name">按名称</option>
          <option value="count">数量优先</option>
        </select>
      </div>
      
      <!-- 背包内容区 -->
      <div id="inventory-content" style="padding: 15px;">
        <!-- 物品页 -->
        <div id="inv-tab-0" class="inv-tab-content">
          <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; padding: 0 5px;">
            <span style="font-size: 12px; color: #999;">背包容量</span>
            <span style="font-size: 12px; color: ${this.items.length > this.gridSize.cols * this.gridSize.rows * 0.8 ? '#ef4444' : '#4ade80'};">
              ${this.items.length} / ${this.gridSize.cols * this.gridSize.rows}
            </span>
          </div>
          <div id="inv-item-grid" style="
            display: grid;
            grid-template-columns: repeat(6, 56px);
            gap: 8px;
            justify-content: center;
          ">
          </div>
        </div>
        
        <!-- 装备页 -->
        <div id="inv-tab-1" class="inv-tab-content" style="display: none;">
          <div style="display: flex; gap: 20px; justify-content: center;">
            <!-- 装备栏 -->
            <div id="inv-equip-grid" style="
              display: grid;
              grid-template-columns: repeat(2, 75px);
              gap: 12px;
            ">
            </div>
            
            <!-- 角色属性 -->
            <div style="
              background: rgba(0, 0, 0, 0.3);
              border-radius: 10px;
              padding: 15px;
              min-width: 200px;
            ">
              <div style="color: #e94560; font-size: 14px; margin-bottom: 10px; text-align: center;">角色属性</div>
              <div id="inv-attr-list" style="font-size: 12px; color: #ccc;">
                <div style="display: flex; justify-content: space-between; margin: 5px 0;">
                  <span>生命</span><span style="color: #ef4444;">100/100</span>
                </div>
                <div style="display: flex; justify-content: space-between; margin: 5px 0;">
                  <span>内力</span><span style="color: #60a5fa;">100/100</span>
                </div>
                <div style="display: flex; justify-content: space-between; margin: 5px 0;">
                  <span>攻击</span><span style="color: #4ade80;">10</span>
                </div>
                <div style="display: flex; justify-content: space-between; margin: 5px 0;">
                  <span>防御</span><span style="color: #4ade80;">5</span>
                </div>
                <div style="display: flex; justify-content: space-between; margin: 5px 0;">
                  <span>速度</span><span style="color: #4ade80;">10</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 任务页 -->
        <div id="inv-tab-2" class="inv-tab-content" style="display: none;">
          <div id="inv-quest-list" style="text-align: center; color: #999; padding: 30px;">
            暂无任务物品
          </div>
        </div>
      </div>
    `;
  }
  
  /**
   * 初始化物品槽位
   */
  initSlots() {
    // 初始化物品网格
    const itemGrid = document.getElementById('inv-item-grid');
    if (itemGrid) {
      for (let i = 0; i < this.gridSize.cols * this.gridSize.rows; i++) {
        const slot = this.createItemSlot(i);
        this.itemSlots.push(slot);
        itemGrid.appendChild(slot.element);
      }
    }
    
    // 初始化装备栏
    const equipGrid = document.getElementById('inv-equip-grid');
    if (equipGrid) {
      Object.keys(this.equipPositions).forEach(pos => {
        const equip = this.createEquipSlot(pos);
        this.equipSlots[pos] = equip;
        equipGrid.appendChild(equip.element);
      });
    }
  }
  
  /**
   * 创建物品槽
   */
  createItemSlot(index) {
    const slot = document.createElement('div');
    slot.className = 'inv-item-slot';
    slot.dataset.index = index;
    
    slot.style.cssText = `
      width: 56px;
      height: 56px;
      background: rgba(45, 55, 72, 0.6);
      border: 2px solid #4a5568;
      border-radius: 8px;
      display: flex;
      justify-content: center;
      align-items: center;
      cursor: pointer;
      position: relative;
      transition: all 0.2s ease;
    `;
    
    slot.addEventListener('mouseenter', (e) => this.onSlotHover(e, index, 'item'));
    slot.addEventListener('mouseleave', () => this.hideTooltip());
    slot.addEventListener('click', () => this.onSlotClick(index, 'item'));
    slot.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      this.onSlotRightClick(index, 'item');
    });
    slot.addEventListener('mousedown', (e) => this.onDragStart(e, index, 'item'));
    
    return {
      element: slot,
      item: null
    };
  }
  
  /**
   * 创建装备槽
   */
  createEquipSlot(position) {
    const config = this.equipPositions[position];
    const slot = document.createElement('div');
    slot.className = 'inv-equip-slot';
    slot.dataset.position = position;
    
    slot.style.cssText = `
      width: 70px;
      height: 70px;
      background: rgba(45, 55, 72, 0.6);
      border: 2px solid #4a5568;
      border-radius: 8px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      cursor: pointer;
      position: relative;
      transition: all 0.2s ease;
    `;
    
    // 装备图标（使用createIconElement支持emoji和图片两种格式）
    const icon = this.createIconElement(config.icon, 28);
    icon.className = 'equip-icon';
    
    // 装备名称
    const name = document.createElement('div');
    name.className = 'equip-name';
    name.style.cssText = 'font-size: 10px; color: #999; margin-top: 4px;';
    name.textContent = config.name;
    
    slot.appendChild(icon);
    slot.appendChild(name);
    
    slot.addEventListener('mouseenter', (e) => this.onSlotHover(e, position, 'equip'));
    slot.addEventListener('mouseleave', () => this.hideTooltip());
    slot.addEventListener('click', () => this.onSlotClick(position, 'equip'));
    slot.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      this.onSlotRightClick(position, 'equip');
    });
    slot.addEventListener('mousedown', (e) => this.onDragStart(e, position, 'equip'));
    
    return {
      element: slot,
      position,
      item: null
    };
  }
  
  /**
   * 绑定事件
   */
  bindEvents() {
    // 关闭按钮
    const closeBtn = document.getElementById('inventory-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.hide());
    }
    
    // 标签页切换
    document.querySelectorAll('.inv-tab').forEach(tab => {
      tab.addEventListener('click', (e) => {
        const tabIndex = parseInt(e.target.dataset.tab);
        this.switchTab(tabIndex);
      });
    });
    
    // 搜索输入
    const searchInput = document.getElementById('inv-search-input');
    if (searchInput) {
      searchInput.addEventListener('input', (e) => {
        this.searchKeyword = e.target.value.toLowerCase();
        this.refreshItems();
      });
    }
    
    // 类型筛选
    const filterSelect = document.getElementById('inv-filter-select');
    if (filterSelect) {
      filterSelect.addEventListener('change', (e) => {
        this.filterType = e.target.value;
        this.refreshItems();
      });
    }
    
    // 排序选择
    const sortSelect = document.getElementById('inv-sort-select');
    if (sortSelect) {
      sortSelect.addEventListener('change', (e) => {
        this.sortMode = e.target.value;
        this.refreshItems();
      });
    }
    
    // 键盘B键打开/关闭背包 - 将处理函数保存为实例方法，以便后续移除
    this.keydownHandler = (e) => {
      if (e.key === 'b' || e.key === 'B') {
        if (this.game.state === 'playing') {
          this.toggle();
        }
      } else if (e.key === 'Escape' && this.isOpen) {
        this.hide();
      }
    };
    document.addEventListener('keydown', this.keydownHandler);
  }
  
  /**
   * 销毁实例，清理事件监听器
   */
  destroy() {
    // 移除键盘事件
    if (this.keydownHandler) {
      document.removeEventListener('keydown', this.keydownHandler);
      this.keydownHandler = null;
    }
  }
  
  /**
   * 切换标签页
   */
  switchTab(index) {
    this.currentTab = index;
    
    // 更新标签样式
    document.querySelectorAll('.inv-tab').forEach((tab, i) => {
      if (i === index) {
        tab.classList.add('active');
        tab.style.background = 'rgba(233, 69, 96, 0.2)';
        tab.style.borderColor = '#e94560';
        tab.style.color = '#e94560';
      } else {
        tab.classList.remove('active');
        tab.style.background = 'rgba(74, 85, 104, 0.3)';
        tab.style.borderColor = '#4a5568';
        tab.style.color = '#999';
      }
    });
    
    // 更新内容显示
    document.querySelectorAll('.inv-tab-content').forEach((content, i) => {
      content.style.display = i === index ? 'block' : 'none';
    });
  }
  
  /**
   * 显示/隐藏背包
   */
  toggle() {
    if (this.isOpen) {
      this.hide();
    } else {
      this.show();
    }
  }
  
  /**
   * 显示背包
   */
  show() {
    this.isOpen = true;
    this.overlay.style.display = 'block';
    this.container.style.display = 'block';
    
    // 刷新数据
    this.refreshItems();
    this.refreshEquipments();
    this.updateGold();
    this.updateAttributes();
  }
  
  /**
   * 隐藏背包
   */
  hide() {
    this.isOpen = false;
    this.overlay.style.display = 'none';
    this.container.style.display = 'none';
    this.hideTooltip();
  }
  
  /**
   * 刷新物品列表
   * ★ 按 grid_index 位置放置物品（保持用户手动调整的位置）
   */
  refreshItems() {
    // 清空所有槽位
    this.itemSlots.forEach(slot => {
      slot.item = null;
      slot.element.innerHTML = '';
      slot.element.style.borderColor = '#4a5568';
    });

    // ★ 按 grid_index 放置物品（保持位置不变）
    this.items.forEach((item) => {
      // 获取物品的目标格子位置
      const targetIndex = item.grid_index !== undefined ? item.grid_index : this.items.indexOf(item);
      
      // 确保索引在有效范围内
      if (targetIndex >= 0 && targetIndex < this.itemSlots.length) {
        this.setItemToSlot(this.itemSlots[targetIndex], item);
      } else {
        // 如果位置无效，放到第一个空位
        const firstEmptySlot = this.itemSlots.findIndex(s => s.item === null);
        if (firstEmptySlot !== -1) {
          item.grid_index = firstEmptySlot;
          this.setItemToSlot(this.itemSlots[firstEmptySlot], item);
        }
      }
    });

    // 更新容量显示
    this.updateCapacityDisplay();
  }

  /**
   * 获取筛选后的物品列表
   */
  getFilteredItems() {
    let items = [...this.items];
    
    // 类型筛选
    if (this.filterType !== 'all') {
      items = items.filter(item => {
        switch (this.filterType) {
          case 'potion':
            return item.type === 'potion' || item.type === 1;
          case 'material':
            return item.type === 'material' || item.type === 3;
          case 'equipment':
            return item.type === 'weapon' || item.type === 'armor' || item.type === 'helmet' || 
                   item.type === 'necklace' || item.type === 'ring' || item.type === 'boots' || item.type === 2;
          case 'quest':
            return item.type === 'quest' || item.type === 4;
          default:
            return true;
        }
      });
    }
    
    // 关键词搜索
    if (this.searchKeyword) {
      items = items.filter(item => {
        const name = (item.name || '').toLowerCase();
        const desc = (item.description || '').toLowerCase();
        return name.includes(this.searchKeyword) || desc.includes(this.searchKeyword);
      });
    }
    
    return items;
  }

  /**
   * 更新容量显示
   */
  updateCapacityDisplay() {
    const content = document.getElementById('inv-tab-0');
    if (!content) return;
    
    const capacityEl = content.querySelector('span:last-child');
    if (capacityEl) {
      const filteredCount = this.getFilteredItems().length;
      const maxCapacity = this.gridSize.cols * this.gridSize.rows;
      const isFull = filteredCount > maxCapacity * 0.8;
      capacityEl.textContent = `${filteredCount} / ${maxCapacity}`;
      capacityEl.style.color = isFull ? '#ef4444' : '#4ade80';
    }
  }

  /**
   * 应用排序
   * @param {Array} items - 要排序的物品数组（可选，默认使用this.items）
   */
  applySort(items = null) {
    const targetArray = items || this.items;
    if (!this.sortMode || this.sortMode === 'default') return;
    targetArray.sort((a, b) => {
      switch (this.sortMode) {
        case 'quality':
          return (b.quality || 0) - (a.quality || 0);
        case 'type':
          return (a.type || 0) - (b.type || 0);
        case 'name':
          return (a.name || '').localeCompare(b.name || '');
        case 'count':
          return (b.count || 1) - (a.count || 1);
        default:
          return 0;
      }
    });
  }

  /**
   * 设置排序模式
   */
  setSortMode(mode) {
    this.sortMode = mode;
    this.refreshItems();
  }
  
  /**
   * 刷新装备列表
   */
  refreshEquipments() {
    Object.values(this.equipSlots).forEach(slot => {
      slot.item = null;
      slot.element.innerHTML = '';
      const config = this.equipPositions[slot.position];
      slot.element.innerHTML = `
        <div style="font-size: 28px;">${config.icon}</div>
        <div style="font-size: 10px; color: #999; margin-top: 4px;">${config.name}</div>
      `;
      slot.element.style.borderColor = '#4a5568';
    });
    
    // 填充装备
    this.equipments.forEach(equip => {
      if (equip.equip_pos && this.equipSlots[equip.equip_pos]) {
        this.setItemToSlot(this.equipSlots[equip.equip_pos], equip);
      }
    });
  }
  
  /**
   * 设置物品到槽位
   */
  setItemToSlot(slot, item) {
    slot.item = item;
    slot.element.innerHTML = '';

    // 物品图标（使用createIconElement支持emoji和图片两种格式）
    const iconValue = this.getItemIcon(item);
    const icon = this.createIconElement(iconValue, 28);

    // 数量
    if (item.count > 1) {
      const count = document.createElement('div');
      count.style.cssText = `
        position: absolute;
        bottom: 2px;
        right: 4px;
        font-size: 10px;
        color: #fff;
        text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
      `;
      count.textContent = item.count;
      slot.element.appendChild(count);
    }

    // 品质边框
    const borderColor = this.getQualityColor(item.quality);
    slot.element.style.borderColor = borderColor;

    // ★ 品质发光效果（高品质物品添加光晕）
    if (item.quality >= 4) { // 史诗及以上
      slot.element.style.boxShadow = `0 0 8px ${borderColor}80`;
    } else if (item.quality === 3) { // 精良
      slot.element.style.boxShadow = `0 0 4px ${borderColor}40`;
    }

    slot.element.appendChild(icon);
  }
  
  /**
   * 槽位悬停
   */
  onSlotHover(e, index, type) {
    let slot, item;
    
    if (type === 'item') {
      slot = this.itemSlots[index];
      item = slot?.item;
    } else {
      slot = this.equipSlots[index];
      item = slot?.item;
    }
    
    if (!item) return;
    
    // 显示物品提示
    this.showItemTooltip(item, e.clientX, e.clientY);
  }
  
  /**
   * 显示物品提示
   */
  showItemTooltip(item, x, y) {
    this.hideTooltip();
    
    const tooltip = document.createElement('div');
    tooltip.id = 'inv-item-tooltip';
    tooltip.style.cssText = `
      position: fixed;
      left: ${x + 15}px;
      top: ${y + 15}px;
      background: rgba(20, 20, 30, 0.95);
      border: 2px solid ${this.getQualityColor(item.quality)};
      border-radius: 10px;
      padding: 12px 15px;
      min-width: 180px;
      max-width: 250px;
      z-index: 1000;
      font-family: 'Microsoft YaHei', sans-serif;
      pointer-events: none;
    `;
    
    const qualityName = this.getQualityName(item.quality);
    
    tooltip.innerHTML = `
      <div style="color: ${this.getQualityColor(item.quality)}; font-size: 14px; font-weight: bold; margin-bottom: 8px;">
        ${item.name}
      </div>
      <div style="color: #999; font-size: 11px; margin-bottom: 8px;">
        ${item.type_name} · ${qualityName}
      </div>
      ${item.description ? `<div style="color: #ccc; font-size: 12px; margin-bottom: 8px; line-height: 1.4;">${item.description}</div>` : ''}
      ${item.attrs ? `<div style="color: #4ade80; font-size: 12px; margin-top: 8px;">${this.formatAttrs(item.attrs)}</div>` : ''}
      <div style="color: #60a5fa; font-size: 11px; margin-top: 8px;">
        ${item.count > 1 ? `数量: ${item.count}` : '右键使用/穿戴'}
      </div>
    `;
    
    document.body.appendChild(tooltip);
  }
  
  /**
   * 隐藏提示框
   */
  hideTooltip() {
    const tooltip = document.getElementById('inv-item-tooltip');
    if (tooltip) {
      tooltip.parentNode.removeChild(tooltip);
    }
  }
  
  /**
   * 槽位点击
   */
  onSlotClick(index, type) {
    // 如果正在拖拽，执行交换
    if (this.isDragging && this.draggedItem) {
      this.handleDrop(index, type);
      return;
    }
    
    let item;
    
    if (type === 'item') {
      item = this.itemSlots[index]?.item;
    } else {
      item = this.equipSlots[index]?.item;
    }
    
    if (!item) return;
    
    // 使用物品
    this.useItem(item);
  }

  /**
   * 槽位右键点击
   */
  onSlotRightClick(index, type) {
    let item;
    
    if (type === 'item') {
      item = this.itemSlots[index]?.item;
    } else {
      item = this.equipSlots[index]?.item;
    }
    
    if (!item) return;
    
    if (type === 'equip' || item.can_equip) {
      // 穿戴装备
      this.equipItem(item);
    } else {
      // 使用物品
      this.useItem(item);
    }
  }

  /**
   * 开始拖拽
   */
  onDragStart(e, index, type) {
    let slot, item;
    
    if (type === 'item') {
      slot = this.itemSlots[index];
      item = slot?.item;
    } else {
      slot = this.equipSlots[index];
      item = slot?.item;
    }
    
    if (!item) return;
    
    e.preventDefault();
    
    this.isDragging = true;
    this.draggedItem = item;
    this.draggedSlot = slot;
    this.draggedType = type;
    
    // 创建拖拽虚影
    this.createDragGhost(e, item);
    
    // 隐藏原槽位物品显示
    slot.element.style.opacity = '0.5';
    
    // 添加全局事件监听
    document.addEventListener('mousemove', this.onDragMoveHandler);
    document.addEventListener('mouseup', this.onDragEndHandler);
  }

  /**
   * 创建拖拽虚影
   */
  createDragGhost(e, item) {
    if (this.dragGhost) {
      document.body.removeChild(this.dragGhost);
    }
    
    this.dragGhost = document.createElement('div');
    this.dragGhost.style.cssText = `
      position: fixed;
      width: 50px;
      height: 50px;
      background: rgba(45, 55, 72, 0.9);
      border: 2px solid ${this.getQualityColor(item.quality)};
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 24px;
      z-index: 9999;
      pointer-events: none;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    `;
    
    this.dragGhost.innerHTML = `
      <div>${this.getItemIcon(item.type)}</div>
      ${item.count > 1 ? `<div style="position: absolute; bottom: 2px; right: 4px; font-size: 10px; color: #fff;">${item.count}</div>` : ''}
    `;
    
    document.body.appendChild(this.dragGhost);
    this.updateDragGhostPosition(e);
  }

  /**
   * 更新拖拽虚影位置
   */
  updateDragGhostPosition(e) {
    if (!this.dragGhost) return;
    this.dragGhost.style.left = `${e.clientX - 25}px`;
    this.dragGhost.style.top = `${e.clientY - 25}px`;
  }

  /**
   * 拖拽移动处理函数
   */
  onDragMoveHandler = (e) => {
    if (!this.isDragging) return;
    this.updateDragGhostPosition(e);
  }

  /**
   * 拖拽结束处理函数
   */
  onDragEndHandler = (e) => {
    this.endDrag(e); // ★ 传入鼠标事件，用于检测放置位置
  }

  /**
   * 结束拖拽
   */
  endDrag(e) {
    if (!this.isDragging) return;

    // ★ 检测鼠标释放位置，执行放置操作
    if (e && this.draggedItem && this.draggedSlot) {
      this.detectAndHandleDrop(e);
    }

    this.isDragging = false;

    // 恢复原槽位显示
    if (this.draggedSlot) {
      this.draggedSlot.element.style.opacity = '1';
    }

    // 移除拖拽虚影
    if (this.dragGhost) {
      document.body.removeChild(this.dragGhost);
      this.dragGhost = null;
    }

    // 移除全局事件监听
    document.removeEventListener('mousemove', this.onDragMoveHandler);
    document.removeEventListener('mouseup', this.onDragEndHandler);

    this.draggedItem = null;
    this.draggedSlot = null;
    this.draggedType = null;
  }

  /**
   * ★ 检测鼠标释放位置并执行放置操作
   */
  detectAndHandleDrop(e) {
    // 获取鼠标释放位置的坐标
    const x = e.clientX;
    const y = e.clientY;

    // 检测是否在背包槽位上释放
    for (let i = 0; i < this.itemSlots.length; i++) {
      const slot = this.itemSlots[i];
      const rect = slot.element.getBoundingClientRect();

      // 判断鼠标是否在槽位范围内
      if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
        // ★ 执行放置操作
        this.handleDrop(i, 'item');
        return;
      }
    }

    // 检测是否在装备槽位上释放
    for (const position in this.equipSlots) {
      const slot = this.equipSlots[position];
      const rect = slot.element.getBoundingClientRect();

      if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
        // ★ 执行放置操作
        this.handleDrop(position, 'equip');
        return;
      }
    }

    // 如果不在任何槽位上释放，物品回到原位置（自动刷新）
    console.log('[Inventory] 拖拽取消，物品回到原位置');
    this.refreshItems();
  }

  /**
   * 处理放置
   */
  handleDrop(index, type) {
    if (!this.draggedItem || !this.draggedSlot) return;

    let targetSlot, targetItem;

    if (type === 'item') {
      targetSlot = this.itemSlots[index];
      targetItem = targetSlot?.item;
    } else {
      targetSlot = this.equipSlots[index];
      targetItem = targetSlot?.item;
    }

    // ★ 拖拽到空槽位 - 移动物品到新位置
    if (!targetItem) {
      if (this.draggedType === 'equip') {
        // 装备拖到背包空槽位 → 卸下装备
        this.unequipItemBySlot(this.draggedSlot.position);
      } else if (this.draggedType === 'item') {
        // ★ 背包物品拖到空槽位 → 移动到该槽位
        this.moveItemToSlot(this.draggedSlot, targetSlot);
      }
      return; // ★ 不调用 endDrag()，外层已处理
    }

    // 拖拽到有物品的槽位
    if (this.draggedType === 'item' && type === 'item') {
      // ★ 背包内操作：先尝试堆叠，再交换
      if (this.canStackItems(this.draggedItem, targetItem)) {
        // 相同物品 → 堆叠
        this.stackItems(this.draggedItem, targetItem, this.draggedSlot);
      } else {
        // 不同物品 → 交换位置
        this.swapItemsInBag(this.draggedSlot, targetSlot);
      }
    } else if (this.draggedType === 'equip' && type === 'item') {
      // 装备拖到背包物品上
      this.unequipItemBySlot(this.draggedSlot.position);
    } else if (this.draggedType === 'item' && type === 'equip') {
      // 背包物品拖到装备栏
      if (this.draggedItem.can_equip && this.draggedItem.equip_pos === targetSlot.position) {
        this.equipItem(this.draggedItem);
      } else if (targetItem) {
        this.swapEquipments(this.draggedItem, targetItem);
      }
    } else if (this.draggedType === 'equip' && type === 'equip') {
      // 装备栏内交换
      if (this.draggedSlot.position !== targetSlot.position) {
        this.swapEquipmentsInSlots(this.draggedSlot, targetSlot);
      }
    }
  }

  /**
   * ★ 判断两个物品是否可以堆叠
   * 条件：相同 item_id + 可堆叠类型（消耗品/材料）
   */
  canStackItems(item1, item2) {
    if (!item1 || !item2) return false;

    // ★ 必须是相同的物品类型ID（item_id，而非背包记录ID）
    const itemID1 = item1.item_id || item1.id;
    const itemID2 = item2.item_id || item2.id;
    const sameType = itemID1 === itemID2;

    // ★ 检查类型是否可堆叠（支持数字和字符串两种格式）
    // 数字类型（服务端）: 1=消耗品, 3=材料
    // 字符串类型（Items.json）: 'potion', 'consumable', 'material', '药水', '消耗品', '材料'
    const isStackable = this.isItemStackable(item1.type) || this.isItemStackable(item2.type);

    // 检查是否超过最大堆叠数
    const maxStack = item1.stack_max || item2.stack_max || 99;
    const totalCount = (item1.count || 1) + (item2.count || 1);

    console.log(`[Inventory] 堆叠检查: ${item1.name}(${itemID1}) + ${item2.name}(${itemID2}) = ${totalCount}/${maxStack}`);
    console.log(`[Inventory]   - sameType: ${sameType}, type1: ${item1.type}, type2: ${item2.type}, isStackable: ${isStackable}`);

    return sameType && isStackable && (totalCount <= maxStack);
  }

  /**
   * ★ 判断物品类型是否可堆叠
   * @param {number|string} itemType - 物品类型（数字或字符串）
   * @returns {boolean}
   */
  isItemStackable(itemType) {
    if (itemType === undefined || itemType === null) return false;

    // 规范化类型
    const normalize = (t) => {
      if (typeof t === 'number') return t;
      const lower = String(t).toLowerCase();
      const typeMap = {
        'potion': 1,
        'consumable': 1,
        'consumables': 1,
        '药水': 1,
        '消耗品': 1,
        '消耗': 1,
        'material': 3,
        'materials': 3,
        '材料': 3,
      };
      return typeMap[lower] || null;
    };

    const normalized = normalize(itemType);
    return normalized === 1 || normalized === 3;
  }

  /**
   * ★ 堆叠两个相同物品（带服务端同步）
   */
  stackItems(sourceItem, targetItem, sourceSlot) {
    // 累加数量到目标物品
    const sourceCount = sourceItem.count || 1;
    targetItem.count = (targetItem.count || 1) + sourceCount;

    // 从源位置移除物品
    this.removeItemFromSlot(sourceSlot);

    // 更新显示
    this.refreshItems();

    // 显示提示
    this.showToast(`${targetItem.name || '物品'} 已堆叠 (${targetItem.count})`);

    // ★ 同步到服务端（异步）
    this.syncStackToServer(sourceItem.id, targetItem.id, sourceCount);
  }

  /**
   * ★ 异步同步堆叠操作到服务端
   */
  async syncStackToServer(sourceItemId, targetItemId, count) {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) {
        console.warn('[Inventory] 无法同步堆叠：角色ID不存在');
        return;
      }

      console.log(`[Inventory] 正在同步堆叠: item ${sourceItemId} → ${targetItemId}, count: ${count}`);

      // 通过网关代理访问
      const gatewayUrl = window.GameConfig?.gatewayUrl || 'http://localhost:8080';
      
      const response = await fetch(`${gatewayUrl}/api/item/bag/${roleId}/merge`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          role_id: roleId,
          source_item_id: sourceItemId,
          target_item_id: targetItemId,
          count: count,
        }),
      });

      const result = await response.json();
      if (result.code === 200) {
        console.log('[Inventory] 堆叠同步成功（通过网关）');
      } else {
        console.warn('[Inventory] 堆叠同步失败:', result.msg || '未知错误');
      }
    } catch (error) {
      console.error('[Inventory] 堆叠同步异常:', error);
    }
  }

  /**
   * ★ 移动物品到指定空槽位（带服务端同步）
   */
  moveItemToSlot(sourceSlot, targetSlot) {
    const item = sourceSlot.item;
    if (!item || targetSlot.item) return; // 目标必须为空

    // ★ 获取槽位索引（从 DOM dataset 或数组位置）
    const sourceIndex = parseInt(sourceSlot.element.dataset.index) || this.itemSlots.indexOf(sourceSlot);
    const targetIndex = parseInt(targetSlot.element.dataset.index) || this.itemSlots.indexOf(targetSlot);

    // 找到 items 数组中的索引
    const idx = this.items.findIndex(i => i.id === item.id);
    if (idx === -1) {
      console.warn('[Inventory] 找不到物品:', item);
      return;
    }

    // ★ 更新物品的 grid_index（格子位置）
    this.items[idx].grid_index = targetIndex;

    // 刷新显示
    this.refreshItems();

    console.log(`[Inventory] 物品 "${item.name}" 从格子 ${sourceIndex} 移动到 ${targetIndex}`);

    // ★ 同步到服务端（异步，不阻塞UI）
    this.syncMoveToServer(sourceIndex, targetIndex);
  }

  /**
   * ★ 异步同步物品移动到服务端（通过网关代理，支持分布式架构）
   */
  async syncMoveToServer(fromGrid, toGrid) {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) {
        console.warn('[Inventory] 无法同步：角色ID不存在');
        return;
      }

      console.log(`[Inventory] 正在同步移动: ${fromGrid} → ${toGrid}`);

      // ★ 通过网关代理访问，遵循分布式架构
      // 网关统一入口：http://localhost:8080
      // 网关会根据配置路由到对应的GameService实例
      const gatewayUrl = window.GameConfig?.gatewayUrl || 'http://localhost:8080';
      
      const response = await fetch(`${gatewayUrl}/api/item/bag/${roleId}/move`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          from_grid: fromGrid,
          to_grid: toGrid,
        }),
      });

      const result = await response.json();
      if (result.code === 200) {
        console.log('[Inventory] 移动同步成功（通过网关）');
      } else {
        console.warn('[Inventory] 移动同步失败:', result.msg || '未知错误');
      }
    } catch (error) {
      console.error('[Inventory] 移动同步异常:', error);
      // 网络错误时继续使用客户端状态，下次登录时可能恢复
    }
  }

  /**
   * ★ 从指定槽位移除物品（内部方法）
   */
  removeItemFromSlot(slot) {
    if (!slot?.item) return;

    const idx = this.items.findIndex(i => i.id === slot.item.id);
    if (idx !== -1) {
      this.items.splice(idx, 1);
    }
  }

  /**
   * 背包内交换物品（带服务端同步）
   */
  swapItemsInBag(slot1, slot2) {
    const item1 = slot1.item;
    const item2 = slot2.item;

    // ★ 获取槽位索引
    const index1 = parseInt(slot1.element.dataset.index) || this.itemSlots.indexOf(slot1);
    const index2 = parseInt(slot2.element.dataset.index) || this.itemSlots.indexOf(slot2);

    // 找到在items数组中的位置
    const idx1 = this.items.findIndex(i => i.id === item1.id);
    const idx2 = this.items.findIndex(i => i.id === item2.id);

    if (idx1 !== -1 && idx2 !== -1) {
      // ★ 交换物品位置
      [this.items[idx1], this.items[idx2]] = [this.items[idx2], this.items[idx1]];

      // ★ 更新 grid_index（格子位置）
      this.items[idx1].grid_index = index2;
      this.items[idx2].grid_index = index1;

      this.refreshItems();
      console.log(`[Inventory] 交换物品: ${item1.name}(${index1}) ↔ ${item2.name}(${index2})`);

      // ★ 同步到服务端（异步）
      this.syncMoveToServer(index1, index2);
    }
  }

  /**
   * 交换装备
   */
  swapEquipments(item1, item2) {
    // 卸下旧装备
    this.unequipItemByEquip(item2);
    // 装备新装备
    this.equipItem(item1);
  }

  /**
   * 装备栏内交换
   */
  swapEquipmentsInSlots(slot1, slot2) {
    const item1 = slot1.item;
    const item2 = slot2.item;
    
    // 交换位置
    const pos1 = item1.equip_pos;
    const pos2 = item2.equip_pos;
    
    item1.equip_pos = pos2;
    item2.equip_pos = pos1;
    
    // 更新数组
    this.equipments = this.equipments.map(e => {
      if (e.id === item1.id) return item1;
      if (e.id === item2.id) return item2;
      return e;
    });
    
    this.refreshEquipments();
    this.updateAttributes();
  }

  /**
   * 通过槽位卸下装备
   */
  unequipItemBySlot(position) {
    const slot = this.equipSlots[position];
    if (slot && slot.item) {
      this.unequipItem(position);
    }
  }

  /**
   * 通过装备对象卸下装备
   */
  unequipItemByEquip(equip) {
    const slot = this.equipSlots[equip.equip_pos];
    if (slot) {
      this.unequipItem(equip.equip_pos);
    }
  }
  
  /**
   * 使用物品
   */
  useItem(item) {
    if (!item.can_use) {
      // 不能直接使用的物品（如装备）
      return;
    }
    
    // 发送使用物品协议
    if (window.GameWS) {
      window.GameWS.send(window.Protocol.CMD_USE_ITEM, {
        item_id: item.id,
        target_id: 0
      });
    }
    
    // 从背包移除（如果是消耗品）
    if (item.count > 1) {
      item.count--;
    } else {
      const index = this.items.findIndex(i => i.id === item.id);
      if (index !== -1) {
        this.items.splice(index, 1);
      }
    }
    
    this.refreshItems();
    this.updateAttributes();
    this.hide();
  }
  
  /**
   * 穿戴装备
   */
  equipItem(item) {
    if (!item.can_equip) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('该物品无法装备', 'warning', 1500);
      }
      return;
    }

    // 检查等级要求
    if (item.level_req && (this.game.player.level || 1) < item.level_req) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast(`需要等级 ${item.level_req} 才能装备`, 'warning', 1500);
      }
      return;
    }

    // 发送装备协议到服务端
    if (window.GameWS) {
      window.GameWS.send(window.Protocol.CMD_EQUIP, {
        item_id: item.id,
        equip_pos: item.equip_pos
      });
    }

    // 更新本地装备数据（乐观更新，服务端会推送权威数据）
    const oldEquip = this.equipments.find(e => e.equip_pos === item.equip_pos);

    // 卸下旧装备到背包
    if (oldEquip) {
      const idx = this.items.findIndex(i => i.id === oldEquip.id);
      if (idx === -1) {
        this.items.push(oldEquip);
      }
    }

    // 移除已装备物品从背包
    const itemIdx = this.items.findIndex(i => i.id === item.id);
    if (itemIdx !== -1) {
      this.items.splice(itemIdx, 1);
    }

    // 添加到装备栏
    const existEquipIdx = this.equipments.findIndex(e => e.equip_pos === item.equip_pos);
    if (existEquipIdx !== -1) {
      this.equipments.splice(existEquipIdx, 1);
    }
    this.equipments.push(item);

    // 同步到 player.equippedItems（供 HUD 角色面板使用）
    this.game.player.equippedItems = this.equipments;

    this.refreshItems();
    this.refreshEquipments();
    this.updateAttributes();

    // 提示装备成功
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast(`装备：${item.name}`, 'success', 1500);
    }

    // 更新 HUD 装备摘要
    if (this.game.hudSystem) {
      this.game.hudSystem.updateEquipSummary();
    }
  }
  
  /**
   * 卸下装备
   */
  unequipItem(position) {
    const slot = this.equipSlots[position];
    if (!slot || !slot.item) return;

    const item = slot.item;

    // 发送到服务器
    if (window.GameWS) {
      window.GameWS.send(window.Protocol.CMD_EQUIP, {
        item_id: item.id,
        equip_pos: '',
        action: 'unequip'
      });
    }

    // 移到背包
    this.equipments = this.equipments.filter(e => e.id !== item.id);
    this.items.push(item);

    // 同步到 player.equippedItems
    this.game.player.equippedItems = this.equipments;

    this.refreshItems();
    this.refreshEquipments();
    this.updateAttributes();

    // 提示卸下成功
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast(`卸下：${item.name}`, 'info', 1500);
    }

    // 更新 HUD 装备摘要
    if (this.game.hudSystem) {
      this.game.hudSystem.updateEquipSummary();
    }
  }
  
  /**
   * 更新金币显示
   */
  updateGold() {
    const goldEl = document.getElementById('inventory-gold');
    if (goldEl) {
      goldEl.textContent = this.game.player.gold || 0;
    }
  }
  
  /**
   * 更新角色属性
   */
  /**
   * 更新角色属性显示（包含装备加成+武学加成）
   * 同时将计算后的属性同步到game.player对象
   */
  updateAttributes() {
    if (!this.game) return;

    const player = this.game.player;
    
    // 基础属性（使用登录时保存的基础值，而非当前可能包含加成的值）
    let baseHp = player.baseMaxHp || player.maxHp || 100;
    let baseMp = player.baseMaxMp || player.maxMp || 100;
    let baseAttack = player.baseAttack || player.attack || 10;
    let baseDefense = player.baseDefense || player.defense || 5;
    let baseSpeed = player.baseSpeed || player.speed || 10;
    
    // === 装备加成累计 ===
    let equipHp = 0, equipMp = 0, equipAttack = 0, equipDefense = 0, equipSpeed = 0;
    
    this.equipments.forEach(equip => {
      equipHp += (equip.hp_bonus || 0);
      equipMp += (equip.mp_bonus || 0);
      equipAttack += (equip.attack_bonus || 0);
      equipDefense += (equip.defense_bonus || 0);
      equipSpeed += (equip.speed_bonus || 0);
      
      if (equip.attrs && typeof equip.attrs === 'object') {
        equipAttack += (equip.attrs.attack || 0);
        equipDefense += (equip.attrs.defense || 0);
        equipSpeed += (equip.attrs.speed || 0);
      }
    });
    
    // === 武学加成累计（从已装备武学计算） ===
    let skillHp = 0, skillMp = 0, skillAttack = 0, skillDefense = 0, skillSpeed = 0;
    if (this.game.skillBar && this.game.skillBar.skillConfigCache.size > 0) {
      const skills = player.skills || [];
      skills.forEach(skill => {
        const skillId = skill.skill_id || skill.id;
        const level = skill.level || 1;
        const config = this.game.skillBar.skillConfigCache.get(skillId);
        if (config) {
          // 武学加成 = 每级加成 * 当前等级
          skillHp += (config.hp_bonus || 0) * level;
          skillMp += (config.mp_bonus || 0) * level;
          skillAttack += (config.attack_bonus || 0) * level;
          skillDefense += (config.defense_bonus || 0) * level;
          skillSpeed += (config.speed_bonus || 0) * level;
        }
      });
    }
    
    // 计算最终属性
    const totalHp = baseHp + equipHp + skillHp;
    const totalMp = baseMp + equipMp + skillMp;
    const totalAttack = baseAttack + equipAttack + skillAttack;
    const totalDefense = baseDefense + equipDefense + skillDefense;
    const totalSpeed = baseSpeed + equipSpeed + skillSpeed;
    
    // 同步到player对象（供战斗系统和其他模块使用）
    player.equipBonus = { hp: equipHp, mp: equipMp, attack: equipAttack, defense: equipDefense, speed: equipSpeed };
    player.skillBonus = { hp: skillHp, mp: skillMp, attack: skillAttack, defense: skillDefense, speed: skillSpeed };
    
    // 更新显示的最大生命/内力（始终更新为计算后的值，包括卸下装备/技能后恢复基础值）
    player.maxHp = totalHp;
    player.maxMp = totalMp;
    
    // 更新UI显示
    const attrList = document.getElementById('inv-attr-list');
    if (attrList) {
      attrList.innerHTML = `
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>生命</span><span style="color: #ef4444;">${player.hp}/${totalHp}
            ${equipHp > 0 ? `<small style="color:#60a5fa">[装+${equipHp}]</small>` : ''}
            ${skillHp > 0 ? `<small style="color:#fbbf24">[武+${skillHp}]</small>` : ''}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>内力</span><span style="color: #60a5fa;">${player.mp}/${totalMp}
            ${equipMp > 0 ? `<small style="color:#60a5fa">[装+${equipMp}]</small>` : ''}
            ${skillMp > 0 ? `<small style="color:#fbbf24">[武+${skillMp}]</small>` : ''}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>攻击</span><span style="color: #4ade80;">${totalAttack}
            ${equipAttack > 0 ? `<small style="color:#60a5fa">[装+${equipAttack}]</small>` : ''}
            ${skillAttack > 0 ? `<small style="color:#fbbf24">[武+${skillAttack}]</small>` : ''}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>防御</span><span style="color: #4ade80;">${totalDefense}
            ${equipDefense > 0 ? `<small style="color:#60a5fa">[装+${equipDefense}]</small>` : ''}
            ${skillDefense > 0 ? `<small style="color:#fbbf24">[武+${skillDefense}]</small>` : ''}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>速度</span><span style="color: #4ade80;">${totalSpeed}
            ${equipSpeed > 0 ? `<small style="color:#60a5fa">[装+${equipSpeed}]</small>` : ''}
            ${skillSpeed > 0 ? `<small style="color:#fbbf24">[武+${skillSpeed}]</small>` : ''}
          </span>
        </div>
      `;
    }
    
    // 触发属性更新事件（供其他系统监听）
    if (this.game.onAttributesUpdate) {
      this.game.onAttributesUpdate({
        hp: totalHp,
        mp: totalMp,
        attack: totalAttack,
        defense: totalDefense,
        speed: totalSpeed
      });
    }
  }

  /**
   * ★ 新增：使用服务端返回的bonusDetail更新属性显示（完美适配GameService缓存系统）
   * 直接使用服务端计算好的加成数据，100%准确，无需前端重复计算
   */
  updateAttributesWithBonus(bonusDetail) {
    if (!this.game || !this.game.player) return;

    const player = this.game.player;

    // ★ 如果没有bonusDetail，降级调用旧方法（兼容模式）
    if (!bonusDetail) {
      console.log('[Inventory] 无bonusDetail，降级为标准更新');
      this.updateAttributes();
      return;
    }

    console.log('[Inventory] 使用服务端加成明细更新属性:', bonusDetail);

    // 从bonusDetail中提取装备和技能加成
    const itemBonus = bonusDetail.item_bonus || {};
    const skillBonus = bonusDetail.skill_bonus || {};
    const buffBonus = bonusDetail.buff_bonus || {};

    // 构建加成标记字符串
    const formatBonus = (value, label, color) => {
      if (value > 0) return `<small style="color:${color}">[${label}+${value}]</small>`;
      return '';
    };

    // 更新UI显示（直接使用player对象上的final_attrs值，这些值已被SkillPanel._applyFinalAttributes更新过）
    const currentMaxHp = player.maxHp || 100;
    const currentMaxMp = player.maxMp || 100;
    const currentAttack = player.attack || 10;
    const currentDefense = player.defense || 5;
    const currentSpeed = player.speed || 10;

    const attrList = document.getElementById('inv-attr-list');
    if (attrList) {
      attrList.innerHTML = `
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>生命</span><span style="color: #ef4444;">${player.hp}/${currentMaxHp}
            ${formatBonus(itemBonus.hp || itemBonus.max_hp || 0, '装', '#60a5fa')}
            ${formatBonus(skillBonus.hp || 0, '武', '#fbbf24')}
            ${formatBonus(buffBonus.hp || buffBonus.max_hp || 0, 'BUFF', '#a78bfa')}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>内力</span><span style="color: #60a5fa;">${player.mp}/${currentMaxMp}
            ${formatBonus(itemBonus.mp || itemBonus.max_mp || 0, '装', '#60a5fa')}
            ${formatBonus(skillBonus.mp || 0, '武', '#fbbf24')}
            ${formatBonus(buffBonus.mp || buffBonus.max_mp || 0, 'BUFF', '#a78bfa')}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>攻击</span><span style="color: #4ade80;">${currentAttack}
            ${formatBonus(itemBonus.attack || 0, '装', '#60a5fa')}
            ${formatBonus(skillBonus.attack || 0, '武', '#fbbf24')}
            ${formatBonus(buffBonus.attack || 0, 'BUFF', '#a78bfa')}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>防御</span><span style="color: #4ade80;">${currentDefense}
            ${formatBonus(itemBonus.defense || 0, '装', '#60a5fa')}
            ${formatBonus(skillBonus.defense || 0, '武', '#fbbf24')}
            ${formatBonus(buffBonus.defense || 0, 'BUFF', '#a78bfa')}
          </span>
        </div>
        <div style="display: flex; justify-content: space-between; margin: 5px 0;">
          <span>速度</span><span style="color: #4ade80;">${currentSpeed}
            ${formatBonus(itemBonus.speed || 0, '装', '#60a5fa')}
            ${formatBonus(skillBonus.speed || 0, '武', '#fbbf24')}
            ${formatBonus(buffBonus.speed || 0, 'BUFF', '#a78bfa')}
          </span>
        </div>
      `;
    }

    // 触发属性更新事件（供其他系统监听）
    if (this.game.onAttributesUpdate) {
      this.game.onAttributesUpdate({
        hp: currentMaxHp,
        mp: currentMaxMp,
        attack: currentAttack,
        defense: currentDefense,
        speed: currentSpeed,
        bonusDetail: bonusDetail  // ★ 传递完整的加成明细
      });
    }

    console.log(`[Inventory] 属性面板已更新 - HP:${player.hp}/${currentMaxHp} MP:${player.mp}/${currentMaxMp}`);
  }
  
  /**
   * 添加物品到背包（支持完整物品对象或简化格式）
   */
  addItem(item) {
    // 支持两种调用方式：
    // 方式1: addItem({id, name, icon, quality, type, count}) - 完整对象
    // 方式2: addItem({id: itemId, count: count}) - 简化格式

    // ★ 防御性检查：确保 this.items 已初始化
    if (!this.items) {
      console.warn('[Inventory] this.items 未初始化，尝试重新初始化');
      this.items = [];
    }

    let itemId = item.id;
    let count = item.count || 1;

    // 检查是否可堆叠
    const existing = this.items.find(i =>
      i.id === itemId && i.can_stack !== false
    );

    if (existing) {
      existing.count += count;

      // 更新槽位显示
      const slot = this.itemSlots.find(s => s.item?.id === itemId);
      if (slot) {
        this.setItemToSlot(slot, existing);
      }

      // ★ 持久化掉落物品（累加数量）
      this.saveLocalDroppedItem({ ...existing, count });
    } else {
      // ★ 构建完整的物品对象（合并传入数据和配置数据）
      let newItem;
      if (item.name && item.icon) {
        // 传入的是完整物品数据，直接使用
        newItem = { ...item, count };
      } else {
        // 从 ItemDataManager 获取基础数据补充
        newItem = this.getItemBaseData(itemId);
        newItem.count = count;
        // 合并传入的额外属性
        Object.assign(newItem, item);
      }

      // 确保必要字段存在
      if (!newItem.icon) newItem.icon = this.getItemIcon(newItem);
      if (!newItem.quality) newItem.quality = 1;
      if (!newItem.type_name) newItem.type_name = this.getTypeName(newItem.type);

      this.items.push(newItem);

      // ★ 持久化掉落物品（新增物品）
      this.saveLocalDroppedItem(newItem);
    }

    // 显示提示（使用真实物品名称）
    const itemName = item.name || `物品#${itemId}`;
    this.showToast(`${itemName} +${count}`);

    if (this.isOpen) {
      this.refreshItems();
    }
  }
  
  /**
   * 从背包移除物品
   */
  removeItem(itemId, count = 1) {
    const index = this.items.findIndex(i => i.id === itemId);
    if (index === -1) return false;
    
    const item = this.items[index];
    if (item.count > count) {
      item.count -= count;
    } else {
      this.items.splice(index, 1);
    }
    
    if (this.isOpen) {
      this.refreshItems();
    }
    
    return true;
  }
  
  /**
   * 获取物品图标（优先使用物品自带icon，其次从ItemDataManager获取，最后使用类型默认图标）
   * 支持两种格式：
   * 1. Emoji: "🧪", "⚔️" 等 - 直接显示
   * 2. 图片路径: "item_hp_potion.png", "weapon_iron_sword.png" 等 - 显示为图片
   */
  getItemIcon(item) {
    // 如果传入的是完整物品对象，尝试获取其icon
    if (item && item.icon) {
      return item.icon;
    }

    // 尝试从ItemDataManager根据ID获取
    const itemId = item?.id;
    if (itemId && window.itemDataManager) {
      const itemData = window.itemDataManager.getItem(itemId);
      if (itemData && itemData.icon) {
        return itemData.icon;
      }
    }

    // 根据类型返回默认emoji图标（降级方案）
    const type = item?.type;
    const icons = {
      weapon: '⚔️',
      armor: '👕',
      helmet: '🧢',
      necklace: '📿',
      ring: '💍',
      boots: '👢',
      potion: '🧪',
      material: '💎',
      quest: '📜',
      scroll: '📜',
      // 数字类型的映射（兼容服务端数据格式）
      1: '🧪',   // 消耗品
      2: '⚔️',   // 装备
      3: '💎',   // 材料
      4: '📜'    // 任务物品
    };
    return icons[type] || '📦';
  }

  /**
   * 判断图标是否为图片路径（以 .png/.jpg/.gif/.svg 结尾）
   */
  isImagePath(icon) {
    if (!icon || typeof icon !== 'string') return false;
    return /\.(png|jpg|jpeg|gif|svg|webp)$/i.test(icon);
  }

  /**
   * 创建图标DOM元素（支持emoji和图片两种格式）
   * 图片加载失败时自动降级为智能匹配的emoji
   *
   * 优化：使用静态缓存避免重复404请求
   * @param {string} icon - 图标值（emoji或图片路径）
   * @param {number} size - 图标大小（像素）
   * @returns {HTMLElement} 图标元素
   */
  createIconElement(icon, size = 28) {
    const element = document.createElement('div');

    if (this.isImagePath(icon)) {
      // ★ 检查全局失败缓存（避免重复请求不存在的图片）
      if (Inventory._failedIcons && Inventory._failedIcons.has(icon)) {
        // 已知失败的图标，直接使用降级emoji
        element.style.cssText = `font-size: ${size}px;`;
        element.textContent = this.getFallbackEmoji(icon);
        return element;
      }

      // 图片路径：创建img元素
      element.style.cssText = `
        width: ${size}px; height: ${size}px;
        display: flex; align-items: center; justify-content: center;
      `;
      const img = document.createElement('img');
      img.src = this.getIconImageUrl(icon);
      img.alt = 'item';
      img.style.cssText = `
        max-width: ${size}px; max-height: ${size}px;
        object-fit: contain;
        image-rendering: pixelated;
        filter: drop-shadow(0 0 2px rgba(0,0,0,0.3));
      `;

      // ★ 图片加载失败时：标记到全局缓存 + 智能降级emoji
      const self = this;
      img.onerror = function() {
        this.style.display = 'none';

        // 标记为失败（全局缓存，所有实例共享）
        if (!Inventory._failedIcons) {
          Inventory._failedIcons = new Set();
        }
        Inventory._failedIcons.add(icon);

        // 使用智能匹配的降级emoji
        element.textContent = self.getFallbackEmoji(icon);
        element.style.fontSize = size + 'px';
      };
      element.appendChild(img);
    } else {
      // Emoji：直接显示文本
      element.style.cssText = `font-size: ${size}px;`;
      element.textContent = icon || '📦';
    }

    return element;
  }

  /**
   * 全局失败图标缓存（静态属性，所有Inventory实例共享）
   */
  static _failedIcons = null;

  /**
   * 获取降级emoji（根据文件名智能匹配）
   */
  getFallbackEmoji(fileName) {
    if (!fileName) return '📦';
    const name = (fileName || '').toLowerCase();

    // 消耗品类
    if (name.includes('hp_potion') || name.includes('hp_elixir')) return '🧪';
    if (name.includes('mp_potion') || name.includes('mp_elixir')) return '💧';

    // 武器类
    if (name.includes('sword') || name.includes('blade')) return '⚔️';
    if (name.includes('axe')) return '🪓';
    if (name.includes('bow')) return '🏹';
    if (name.includes('weapon_wood')) return '🪵';
    if (name.includes('weapon_iron') || name.includes('weapon_steel')) return '⚔️';

    // 防具类
    if (name.includes('armor') || name.includes('cloth')) return '👕';
    if (name.includes('helmet') || name.includes('hat')) return '🧢';
    if (name.includes('boots')) return '👢';

    // 饰品类
    if (name.includes('necklace')) return '📿';
    if (name.includes('ring')) return '💍';

    // 材料类
    if (name.includes('iron')) return '�ite';
    if (name.includes('stone') || name.includes('ore')) return '🪨';
    if (name.includes('herb') || name.includes('grass')) return '🌿';
    if (name.includes('crystal') || name.includes('gem')) return '💎';
    if (name.includes('material_')) return '📦';

    return '📦'; // 默认
  }

  /**
   * 获取图标图片的完整URL
   * 支持相对路径和绝对路径
   */
  getIconImageUrl(iconPath) {
    if (!iconPath) return '';

    // 如果是完整的URL，直接返回
    if (iconPath.startsWith('http://') || iconPath.startsWith('https://') || iconPath.startsWith('/')) {
      return iconPath;
    }

    // 相对路径：假设图标在 assets/icons/ 目录下
    return `assets/icons/${iconPath}`;
  }
  
  /**
   * 获取品质颜色
   */
  getQualityColor(quality) {
    // 使用ItemDataManager的方法，避免重复实现
    if (window.itemDataManager) {
      return window.itemDataManager.getQualityColor(quality);
    }
    
    // 降级方案：如果ItemDataManager不存在，使用本地定义
    const colors = {
      1: '#9ca3af', // 普通 - 灰色
      2: '#22c55e', // 优秀 - 绿色
      3: '#3b82f6', // 精良 - 蓝色
      4: '#a855f7', // 史诗 - 紫色
      5: '#f59e0b'  // 传说 - 橙色
    };
    return colors[quality] || '#ffffff';
  }
  
  /**
   * 获取品质名称
   */
  getQualityName(quality) {
    const names = {
      1: '普通',
      2: '优秀',
      3: '精良',
      4: '史诗',
      5: '传说'
    };
    return names[quality] || '普通';
  }
  
  /**
   * 格式化属性
   */
  formatAttrs(attrs) {
    if (!attrs) return '';
    const parts = [];
    if (attrs.attack) parts.push(`攻击 +${attrs.attack}`);
    if (attrs.defense) parts.push(`防御 +${attrs.defense}`);
    if (attrs.speed) parts.push(`速度 +${attrs.speed}`);
    if (attrs.hp) parts.push(`生命 +${attrs.hp}`);
    if (attrs.mp) parts.push(`内力 +${attrs.mp}`);
    return parts.join(' | ');
  }

  /**
   * 根据物品ID从ItemDataManager获取基础数据
   */
  getItemBaseData(itemId) {
    const baseItem = { id: itemId, count: 1 };

    // 从 ItemDataManager 获取配置数据
    if (window.itemDataManager) {
      const config = window.itemDataManager.getItem(itemId);
      if (config) {
        Object.assign(baseItem, {
          name: config.name,
          icon: config.icon,
          type: config.type,
          type_name: config.type_name || this.getTypeName(config.type),
          quality: config.quality || 1,
          description: config.description || ''
        });

        // 装备类物品附加属性
        if (config.attrs) {
          baseItem.attrs = config.attrs;
        }
      }
    }

    // 降级：如果无法获取配置，使用默认值
    if (!baseItem.name) {
      baseItem.name = `物品#${itemId}`;
      baseItem.icon = '📦';
      baseItem.quality = 1;
    }

    return baseItem;
  }

  /**
   * 获取物品类型中文名称
   */
  getTypeName(type) {
    // 支持数字和字符串两种格式
    const typeNames = {
      // 数字类型（服务端返回格式）
      1: '消耗品',
      2: '装备',
      3: '材料',
      4: '任务物品',
      5: '其他',

      // 字符串类型（前端本地格式）
      weapon: '武器',
      armor: '防具',
      helmet: '头盔',
      necklace: '项链',
      ring: '戒指',
      boots: '靴子',
      potion: '消耗品',
      material: '材料',
      quest: '任务物品'
    };
    return typeNames[type] || '其他';
  }

  /**
   * 获取装备位置名称（英文标识）
   */
  getEquipPosName(equipType) {
    const posNames = {
      // 数字类型（服务端格式）
      1: 'weapon',    // 武器
      2: 'armor',     // 防具/衣服
      3: 'helmet',    // 头盔
      4: 'boots',     // 靴子
      5: 'ring',      // 戒指
      6: 'necklace',  // 项链

      // 字符串类型（前端格式，兼容处理）
      weapon: 'weapon',
      armor: 'armor',
      helmet: 'helmet',
      necklace: 'necklace',
      ring: 'ring',
      boots: 'boots'
    };
    return posNames[equipType] || 'unknown';
  }

  /**
   * 初始化背包（使用真实数据源）
   * 不再使用硬编码模拟数据，改为：
   * 1. 从 ItemDataManager 获取真实物品配置
   * 2. 调试模式可加载新手礼包（基于真实配置）
   * 3. 服务端失败时：加载基础生存物品（确保玩家可用）
   */
  initInventory() {
    const itemManager = window.itemDataManager;

    // 清空现有数据
    this.items = [];
    this.equipments = [];

    // 检查 ItemDataManager 是否已加载
    if (!itemManager || !itemManager.allItems || Object.keys(itemManager.allItems).length === 0) {
      console.warn('[Inventory] ItemDataManager 未就绪，延迟初始化');
      return false;
    }

    console.log('[Inventory] 使用真实数据源初始化背包');

    // ★ 策略1：调试模式加载新手礼包
    if (window.DEBUG_MODE && this.shouldLoadStarterKit()) {
      this.loadStarterKit();
    }
    // ★ 策略2：服务端超时/失败时，加载基础生存物品（确保可用性）
    else if (this.shouldLoadBaseKit()) {
      this.loadBaseSurvivalKit();
    }
    // 正式环境且有服务端数据：保持空背包，等待 loadFromServer 填充

    this.refreshItems();
    this.refreshEquipments();
    return true;
  }

  /**
   * 是否应该加载基础生存包（当服务端不可用时）
   */
  shouldLoadBaseKit() {
    // 如果 sessionStorage 中有标记说明已从服务端成功加载过，则不加载
    const hasServerData = sessionStorage.getItem('inventory_server_loaded');
    if (hasServerData) {
      return false;
    }

    // 检查是否有本地掉落物品（战斗获得但未同步的）
    const localDrops = this.getLocalDroppedItems();
    if (localDrops && localDrops.length > 0) {
      console.log(`[Inventory] 发现${localDrops.length}个本地掉落物品，跳过基础包`);
      return false;
    }

    return true; // 首次进入且无服务端数据，加载基础包
  }

  /**
   * 加载基础生存物品（确保玩家在服务端不可用时也能正常游戏）
   * 包含：少量药水、新手武器防具
   */
  loadBaseSurvivalKit() {
    const itemManager = window.itemDataManager;

    // 基础生存物品（比新手礼包少，仅保证基本可玩性）
    const baseItems = [
      { id: 1, count: 5 },   // 金创药 x5
      { id: 2, count: 3 },   // 回蓝丹 x3
    ];

    // 基础装备（直接穿戴）
    const baseEquips = [
      { id: 102, pos: 'weapon' },  // 桃木剑
      { id: 202, pos: 'armor' },   // 布衣
    ];

    console.log('[Inventory] 加载基础生存物品（服务端不可用时的降级方案）');

    // 添加背包物品
    baseItems.forEach(({ id, count }) => {
      const config = itemManager.getItem(id);
      if (config) {
        const item = this.getItemBaseData(id);
        item.count = count;
        this.items.push(item);
        console.log(`  + ${item.name} x${count}`);
      } else {
        console.warn(`  ⚠️ 物品ID ${id} 不存在于 Items.json`);
      }
    });

    // 添加装备
    baseEquips.forEach(({ id, pos }) => {
      const config = itemManager.getItem(id);
      if (config) {
        const equip = this.mergeEquipWithConfig({ id, equip_pos: pos });
        this.equipments.push(equip);
        console.log(`  + [装备] ${equip.name}`);
      }
    });

    // 标记已加载（避免重复加载）
    sessionStorage.setItem('inventory_base_loaded', 'true');
    console.log(`[Inventory] 基础生存包加载完成: ${this.items.length}个物品, ${this.equipments.length}件装备`);
  }

  /**
   * 是否应该加载新手礼包（仅调试模式首次进入时）
   */
  shouldLoadStarterKit() {
    // 检查是否已经加载过新手礼包
    const hasLoaded = sessionStorage.getItem('inventory_starter_loaded');
    if (hasLoaded) {
      console.log('[Inventory] 新手礼包已加载过，跳过');
      return false;
    }
    return true;
  }

  /**
   * 加载新手礼包（基于 Items.json 真实配置）
   * 仅用于开发调试，帮助测试背包UI和装备系统
   */
  loadStarterKit() {
    const itemManager = window.itemDataManager;

    // 新手礼包物品列表（ID来自 items.json 真实配置）
    const starterItems = [
      { id: 1, count: 10 },   // 疗伤药
      { id: 2, count: 5 },    // 回蓝药
      { id: 4, count: 20 },   // 草药
      { id: 101, count: 1 },  // 玄铁剑
      { id: 201, count: 1 },  // 皮甲
      { id: 302, count: 1 },  // 玉佩
    ];

    // 新手装备列表（直接穿戴）
    const starterEquips = [
      { id: 102, pos: 'weapon' },  // 桃木剑（新手武器）
      { id: 202, pos: 'armor' },   // 布衣（新手防具）
    ];

    // 添加背包物品（从 ItemDataManager 获取真实配置）
    starterItems.forEach(({ id, count }) => {
      const config = itemManager.getItem(id);
      if (config) {
        this.items.push({
          ...config,
          count,
          can_use: config.can_use || config.type === 'potion',
          can_equip: config.type === 'weapon' || config.type === 'armor' ||
                    config.type === 'helmet' || config.type === 'necklace' ||
                    config.type === 'ring' || config.type === 'boots'
        });
        console.log(`[Inventory] 新手礼包添加: ${config.name} x${count}`);
      } else {
        console.warn(`[Inventory] 新手礼包物品不存在: ID=${id}`);
      }
    });

    // 添加初始装备（从 ItemDataManager 获取真实配置）
    starterEquips.forEach(({ id, pos }) => {
      const config = itemManager.getItem(id);
      if (config) {
        this.equipments.push({
          ...config,
          can_equip: true,
          equip_pos: pos
        });
        console.log(`[Inventory] 新手装备穿戴: ${config.name}`);
      }
    });

    // 标记已加载
    sessionStorage.setItem('inventory_starter_loaded', 'true');
    console.log(`[Inventory] 新手礼包加载完成: ${this.items.length}个物品, ${this.equipments.length}件装备`);
  }

  /**
   * @deprecated 已废弃，请使用 initInventory() 或 loadFromServer()
   * 保留此方法仅用于向后兼容，内部重定向到 initInventory()
   */
  loadTestData() {
    console.warn('[Inventory] ⚠️ loadTestData() 已废弃，自动重定向到 initInventory()');
    return this.initInventory();
  }
  
  /**
   * 从服务器加载背包数据（生产环境使用）
   * 调用GameService的WebSocket API获取背包和装备数据
   * 自动合并 ItemDataManager 的真实物品配置（icon、品质、属性等）
   *
   * ★ 优化：保护已获得的掉落物品不被清空
   * - 请求前备份当前物品
   * - 失败时恢复或合并本地掉落物品
   * - 使用 sessionStorage 持久化防止丢失
   */
  async loadFromServer() {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) {
        console.warn('[Inventory] 角色ID不存在，尝试本地初始化');
        this.initInventory();
        return;
      }

      console.log(`[Inventory] 正在从服务器加载背包数据 (role_id=${roleId})...`);

      // ★ 关键优化：请求前备份当前物品（保护战斗获得的掉落物）
      const backupItems = [...this.items];
      const backupEquipments = [...this.equipments];
      const localDrops = this.getLocalDroppedItems(); // 从 sessionStorage 读取

      // 通过WebSocket并行请求背包数据和装备数据（★ 增加超时到8秒）
      const [bagData, equipData] = await Promise.all([
        window.GameWS.request(window.Protocol.CMD_ITEM_LIST, { role_id: roleId }, 8000).catch((err) => {
          console.warn('[Inventory] 请求背包列表失败:', err.message);
          return { code: 0, data: [] };
        }),
        window.GameWS.request(window.Protocol.CMD_EQUIP_LIST, { role_id: roleId }, 8000).catch((err) => {
          console.warn('[Inventory] 请求装备列表失败:', err.message);
          return { code: 0, data: [] };
        })
      ]);

      // 判断是否成功获取到任一数据
      const bagSuccess = bagData.code === 200 && bagData.data && Array.isArray(bagData.data);
      const equipSuccess = equipData.code === 200 && equipData.data && Array.isArray(equipData.data);

      if (bagSuccess || equipSuccess) {
        // ★ 成功场景：使用服务端数据，但合并本地掉落物品

        // 处理背包数据
        if (bagSuccess) {
          const serverItems = bagData.data.map(item => this.mergeWithItemConfig(item));
          // ★ 合并策略：服务端数据 + 本地掉落物品（去重）
          this.items = this.mergeDroppedItems(serverItems, [...localDrops, ...backupItems]);
          console.log(`[Inventory] 背包加载完成: ${this.items.length}个物品 (含${localDrops.length}个本地掉落)`);
        } else {
          // 背包请求失败但装备成功：保留当前背包 + 合并掉落物
          this.items = this.mergeDroppedItems(backupItems, localDrops);
          console.log(`[Inventory] 背包请求失败，保留本地数据: ${this.items.length}个物品`);
        }

        // 处理装备数据
        if (equipSuccess) {
          this.equipments = equipData.data.map(equip => this.mergeEquipWithConfig(equip));
          console.log(`[Inventory] 装备加载完成: ${this.equipments.length}件装备`);

          if (this.game && this.game.player) {
            this.game.player.equippedItems = this.equipments;
          }
        } else {
          // 装备请求失败：保留原有装备
          this.equipments = backupEquipments;
          console.log(`[Inventory] 装备请求失败，保留本地装备: ${this.equipments.length}件`);
        }

        // ★ 清除已合并的本地掉落物品（避免重复添加）
        this.clearLocalDroppedItems();

        // ★ 标记服务端数据已成功加载（后续不再加载基础包）
        sessionStorage.setItem('inventory_server_loaded', 'true');

      } else {
        // ★ 失败场景：完全使用本地数据（不清空！）
        console.warn('[Inventory] 服务端数据获取失败，保留本地数据');

        // 恢复备份数据 + 合并掉落物品
        this.items = this.mergeDroppedItems(backupItems, localDrops);
        this.equipments = backupEquipments;

        // 如果完全没有数据，才初始化新手礼包
        if (this.items.length === 0 && this.equipments.length === 0) {
          console.log('[Inventory] 本地数据为空，初始化默认数据');
          this.initInventory();
        }
      }

      // 刷新UI
      this.refreshItems();
      this.refreshEquipments();
      this.updateAttributes();

    } catch (error) {
      console.error('[Inventory] 从服务器加载背包数据失败:', error);
      // 异常情况：保持当前数据不变（不清空！）
      // 仅在完全为空时才初始化
      if (this.items.length === 0 && this.equipments.length === 0) {
        this.initInventory();
      }
    }
  }

  /**
   * 合并服务端数据和本地掉落物品（去重）
   * @param {Array} serverItems - 服务端返回的物品列表
   * @param {Array} localItems - 本地物品列表（掉落物+备份）
   * @returns {Array} 合并后的物品列表
   */
  mergeDroppedItems(serverItems, localItems) {
    if (!localItems || localItems.length === 0) {
      return serverItems || [];
    }

    // 收集服务端已有的物品ID
    const serverItemIds = new Set((serverItems || []).map(item => item.id));

    // 筛选出服务端没有的本地掉落物品
    const newDrops = localItems.filter(item =>
      item.id && !serverItemIds.has(item.id)
    );

    if (newDrops.length > 0) {
      console.log(`[Inventory] 合并${newDrops.length}个本地掉落物品到背包`);
      return [...(serverItems || []), ...newDrops];
    }

    return serverItems || [];
  }

  // ========== 本地掉落物品持久化（sessionStorage）==========

  /**
   * 保存掉落物品到 sessionStorage（防止刷新/超时丢失）
   */
  saveLocalDroppedItem(dropInfo) {
    try {
      const drops = this.getLocalDroppedItems();

      // 查找是否已存在该物品（累加数量）
      const existing = drops.find(d => d.id === dropInfo.id);
      if (existing) {
        existing.count = (existing.count || 1) + (dropInfo.count || 1);
      } else {
        drops.push({
          id: dropInfo.id,
          name: dropInfo.name,
          icon: dropInfo.icon,
          quality: dropInfo.quality,
          type: dropInfo.type,
          type_name: dropInfo.type_name,
          count: dropInfo.count || 1,
          slot: dropInfo.slot,
          timestamp: Date.now() // 用于清理过期数据
        });
      }

      sessionStorage.setItem('inventory_dropped_items', JSON.stringify(drops));
    } catch (e) {
      console.warn('[Inventory] 保存掉落物品失败:', e);
    }
  }

  /**
   * 获取本地保存的掉落物品
   */
  getLocalDroppedItems() {
    try {
      const data = sessionStorage.getItem('inventory_dropped_items');
      if (!data) return [];

      const drops = JSON.parse(data);
      const now = Date.now();

      // 过滤超过10分钟的过期数据（防止无限堆积）
      const validDrops = drops.filter(d =>
        !d.timestamp || (now - d.timestamp) < 10 * 60 * 1000
      );

      if (validDrops.length !== drops.length) {
        sessionStorage.setItem('inventory_dropped_items', JSON.stringify(validDrops));
      }

      return validDrops;
    } catch (e) {
      return [];
    }
  }

  /**
   * 清除本地保存的掉落物品（服务端同步成功后调用）
   */
  clearLocalDroppedItems() {
    try {
      sessionStorage.removeItem('inventory_dropped_items');
    } catch (e) {
      // 忽略
    }
  }

  /**
   * 合并服务端背包数据与 ItemDataManager 配置
   * 确保物品显示正确的 icon、品质、属性等信息
   */
  mergeWithItemConfig(serverItem) {
    // 基础字段：从服务端数据获取
    // ★ 重要：id 必须是背包记录的唯一标识（bagItemId），不能是 item_id！
    const merged = {
      id: serverItem.id,           // ✅ 背包记录唯一ID（用于区分同一物品的多条记录）
      item_id: serverItem.item_id, // 物品类型ID
      bagItemId: serverItem.id,    // 保留引用（向后兼容）
      count: serverItem.count || 1,
      grid_index: serverItem.grid_index,
      level_req: serverItem.level_req || 0,
      price: serverItem.price || 0,
      is_bind: serverItem.is_bind || 0,
    };

    // ★ 从 ItemDataManager 获取真实配置并合并
    const config = window.itemDataManager?.getItem(serverItem.item_id);

    if (config) {
      // 使用 Items.json 的真实配置作为主要数据源
      Object.assign(merged, {
        name: config.name,
        icon: config.icon,
        type: config.type,
        type_name: config.type_name || this.getTypeName(config.type),
        quality: config.quality || serverItem.quality || 1,
        description: config.description || '',
        can_use: config.can_use || config.type === 'potion',
        can_equip: ['weapon', 'armor', 'helmet', 'necklace', 'ring', 'boots'].includes(config.type),
        equip_pos: config.equip_pos || this.getEquipPosName(serverItem.equip_type),
        equip_type: serverItem.equip_type || this.getEquipTypeCode(config.equip_pos),
        // 使用配置中的属性（更完整）
        attrs: config.attrs || this.formatServerAttrs(serverItem),
      });
    } else {
      // 降级：仅使用服务端数据
      console.warn(`[Inventory] 物品配置未找到: ID=${serverItem.item_id}`);
      Object.assign(merged, {
        name: serverItem.name || `物品#${serverItem.item_id}`,
        type: serverItem.type,
        type_name: this.getTypeName(serverItem.type),
        quality: serverItem.quality || 1,
        description: serverItem.description || '',
        can_use: serverItem.type === 1,
        can_equip: serverItem.type === 2,
        equip_pos: this.getEquipPosName(serverItem.equip_type),
        equip_type: serverItem.equip_type,
        attrs: this.formatServerAttrs(serverItem),
        icon: this.getDefaultIcon(serverItem.type)
      });
    }

    return merged;
  }

  /**
   * 合并服务端装备数据与 ItemDataManager 配置
   */
  mergeEquipWithConfig(serverEquip) {
    const merged = {
      id: serverEquip.item_id,
      bagItemId: serverEquip.bag_item_id,
      can_equip: true,
      level_req: serverEquip.level_req || 0,
    };

    // 从 ItemDataManager 获取真实配置
    const config = window.itemDataManager?.getItem(serverEquip.item_id);

    if (config) {
      Object.assign(merged, {
        name: config.name,
        icon: config.icon,
        type: config.type,
        type_name: config.type_name || this.getTypeName(config.type),
        quality: config.quality || serverEquip.quality || 1,
        description: config.description || '',
        equip_pos: config.equip_pos || this.getEquipPosName(serverEquip.equip_type),
        equip_type: serverEquip.equip_type || this.getEquipTypeCode(config.equip_pos),
        attrs: config.attrs || this.formatServerAttrs(serverEquip),
      });
    } else {
      Object.assign(merged, {
        name: serverEquip.name || `装备#${serverEquip.item_id}`,
        type: serverEquip.type,
        type_name: this.getTypeName(serverEquip.type),
        quality: serverEquip.quality || 1,
        description: serverEquip.description || '',
        equip_pos: this.getEquipPosName(serverEquip.equip_type),
        equip_type: serverEquip.equip_type,
        attrs: this.formatServerAttrs(serverEquip),
        icon: this.getDefaultIcon(serverEquip.type)
      });
    }

    return merged;
  }

  /**
   * 格式化服务端返回的属性数据
   */
  formatServerAttrs(item) {
    const attrs = {};
    if (item.attack_bonus > 0) attrs.attack = item.attack_bonus;
    if (item.defense_bonus > 0) attrs.defense = item.defense_bonus;
    if (item.speed_bonus > 0) attrs.speed = item.speed_bonus;
    if (item.hp_bonus > 0) attrs.hp = item.hp_bonus;
    if (item.mp_bonus > 0) attrs.mp = item.mp_bonus;
    return Object.keys(attrs).length > 0 ? attrs : null;
  }

  /**
   * 根据装备位置字符串获取类型代码
   */
  getEquipTypeCode(equipPos) {
    const posToCode = {
      'weapon': 1,
      'helmet': 2,
      'armor': 3,
      'necklace': 4,
      'ring': 5,
      'boots': 6
    };
    return posToCode[equipPos] || 0;
  }

  /**
   * 根据类型获取默认图标（当配置缺失时使用）
   */
  getDefaultIcon(type) {
    const icons = {
      1: '🧪',   // 消耗品
      2: '⚔️',   // 装备
      3: '💎',   // 材料
      4: '📜',   // 任务物品
      'potion': '🧪',
      'weapon': '⚔️',
      'armor': '👕',
      'material': '💎'
    };
    return icons[type] || '📦';
  }

  /**
   * 格式化物品属性（向后兼容，保留供外部调用）
   */
  formatItemAttrs(item) {
    const attrs = [];
    if (item.hp_bonus > 0) attrs.push(`生命+${item.hp_bonus}`);
    if (item.mp_bonus > 0) attrs.push(`内力+${item.mp_bonus}`);
    if (item.attack_bonus > 0) attrs.push(`攻击+${item.attack_bonus}`);
    if (item.defense_bonus > 0) attrs.push(`防御+${item.defense_bonus}`);
    if (item.speed_bonus > 0) attrs.push(`速度+${item.speed_bonus}`);
    if (item.hp_restore > 0) attrs.push(`恢复生命${item.hp_restore}`);
    if (item.mp_restore > 0) attrs.push(`恢复内力${item.mp_restore}`);
    return attrs;
  }

  /**
   * 显示提示消息
   */
  showToast(message, type = 'success') {
    // 优先使用 game.uiManager.toast
    if (this.game?.uiManager?.toast) {
      this.game.uiManager.toast(message, type, 2000);
      return;
    }

    // 降级：console.log + 简单浮动提示
    console.log(`[Inventory] ${message}`);

    const toast = document.createElement('div');
    toast.textContent = message;
    toast.style.cssText = `
      position: fixed; top: 50%; left: 50%;
      transform: translate(-50%, -50%);
      padding: 12px 24px; background: rgba(233, 69, 96, 0.95);
      color: white; border-radius: 8px;
      font-size: 14px; z-index: 9999;
      animation: fadeInOut 2s ease-in-out forwards;
    `;
    document.body.appendChild(toast);
    setTimeout(() => { if (toast.parentNode) toast.parentNode.removeChild(toast); }, 2000);
  }
}

// 导出到全局
window.Inventory = Inventory;
