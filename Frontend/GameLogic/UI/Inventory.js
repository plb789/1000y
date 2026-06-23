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
    this.gridSize = { cols: 5, rows: 4 }; // 背包网格
    this.itemSlots = [];
    this.equipSlots = {}; // 装备栏位
    this.isDragging = false;
    this.draggedItem = null;
    this.draggedSlot = null;

    // 物品数据
    this.items = [];
    this.equipments = [];
    this.questItems = [];

    // 排序模式
    this.sortMode = 'default'; // default, quality, type, name, count
    
    // 装备栏定义
    this.equipPositions = {
      weapon: { name: '武器', icon: '⚔️' },
      armor: { name: '衣服', icon: '👕' },
      helmet: { name: '头盔', icon: '🧢' },
      necklace: { name: '项链', icon: '📿' },
      ring: { name: '戒指', icon: '💍' },
      boots: { name: '鞋子', icon: '👢' }
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
      
      <!-- 背包内容区 -->
      <div id="inventory-content" style="padding: 15px;">
        <!-- 物品页 -->
        <div id="inv-tab-0" class="inv-tab-content">
          <div id="inv-item-grid" style="
            display: grid;
            grid-template-columns: repeat(5, 60px);
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
              grid-template-columns: repeat(2, 70px);
              gap: 10px;
            ">
            </div>
            
            <!-- 角色属性 -->
            <div style="
              background: rgba(0, 0, 0, 0.3);
              border-radius: 10px;
              padding: 15px;
              min-width: 180px;
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
      width: 60px;
      height: 60px;
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
    
    // 装备图标
    const icon = document.createElement('div');
    icon.className = 'equip-icon';
    icon.style.cssText = 'font-size: 28px;';
    icon.textContent = config.icon;
    
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
   */
  refreshItems() {
    // 应用当前排序
    this.applySort();

    // 清空所有槽位
    this.itemSlots.forEach(slot => {
      slot.item = null;
      slot.element.innerHTML = '';
      slot.element.style.borderColor = '#4a5568';
    });

    // 填充物品
    this.items.forEach((item, index) => {
      if (index < this.itemSlots.length) {
        this.setItemToSlot(this.itemSlots[index], item);
      }
    });
  }

  /**
   * 应用排序
   */
  applySort() {
    if (!this.sortMode || this.sortMode === 'default') return;
    this.items.sort((a, b) => {
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
    
    // 物品图标
    const icon = document.createElement('div');
    icon.style.cssText = 'font-size: 28px;';
    icon.textContent = this.getItemIcon(item.type);
    
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
   * 添加物品到背包
   */
  addItem(item) {
    // 检查是否可堆叠
    const existing = this.items.find(i => 
      i.id === item.id && i.can_stack
    );
    
    if (existing) {
      existing.count += item.count || 1;
    } else {
      this.items.push({
        ...item,
        count: item.count || 1
      });
    }
    
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
   * 获取物品图标
   */
  getItemIcon(type) {
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
      scroll: '📜'
    };
    return icons[type] || '📦';
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
   * 加载测试数据（开发环境使用）
   */
  loadTestData() {
    // 使用ItemDataManager获取物品数据
    const itemManager = window.itemDataManager;
    
    if (itemManager && itemManager.allItems) {
      // 从数据管理器获取测试数据
      this.items = [
        { ...itemManager.getItem(1), count: 10 },   // 疗伤药
        { ...itemManager.getItem(2), count: 5 },    // 回蓝药
        { ...itemManager.getItem(101), count: 1 },  // 玄铁剑
        { ...itemManager.getItem(201), count: 1 },  // 皮甲
        { ...itemManager.getItem(3), count: 3 },    // 神秘矿石
        { ...itemManager.getItem(103), count: 1 },  // 屠龙刀
        { ...itemManager.getItem(204), count: 1 },  // 金丝甲
        { ...itemManager.getItem(302), count: 1 }   // 玉佩
      ];
    } else {
      // 降级方案：使用硬编码数据
      this.items = [
        { id: 1, name: '疗伤药', icon: '🧪', type: 'potion', type_name: '药品', quality: 1, count: 10, can_use: true, description: '恢复50点生命' },
        { id: 2, name: '回蓝药', icon: '💧', type: 'potion', type_name: '药品', quality: 1, count: 5, can_use: true, description: '恢复30点内力' },
        { id: 101, name: '玄铁剑', icon: '⚔️', type: 'weapon', type_name: '武器', quality: 3, can_equip: true, equip_pos: 'weapon', attrs: { attack: 15 }, description: '玄铁打造的长剑' },
        { id: 201, name: '皮甲', icon: '🥋', type: 'armor', type_name: '防具', quality: 2, can_equip: true, equip_pos: 'armor', attrs: { defense: 8 }, description: '普通的皮甲' },
        { id: 3, name: '神秘矿石', icon: '💎', type: 'material', type_name: '材料', quality: 4, count: 3, description: '蕴含神秘力量的矿石' }
      ];
    }
    
    this.equipments = [];
    this.refreshItems();
    this.refreshEquipments();
  }
  
  /**
   * 从服务器加载背包数据（生产环境使用）
   * 调用GameService的RESTful API获取背包和装备数据
   */
  async loadFromServer() {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) {
        console.warn('角色ID不存在，无法加载背包');
        this.loadTestData();
        return;
      }

      // 并行请求背包数据和装备数据
      const [bagRes, equipRes] = await Promise.all([
        fetch(`http://localhost:8082/api/item/bag/${roleId}/list`),
        fetch(`http://localhost:8082/api/item/equip/${roleId}/list`)
      ]);

      // 处理背包数据
      if (bagRes.ok) {
        const bagData = await bagRes.json();
        if (bagData.code === 200 && bagData.data) {
          // 转换服务端数据格式为前端格式
          this.items = bagData.data.map(item => ({
            id: item.item_id,
            bagItemId: item.id,
            name: item.name || `物品${item.item_id}`,
            type: item.type,
            type_name: this.getTypeName(item.type),
            quality: item.quality || 1,
            count: item.count || 1,
            can_use: item.type === 1, // 药品类可使用
            can_equip: item.type === 2, // 装备类可穿戴
            equip_pos: this.getEquipPosName(item.equip_type),
            equip_type: item.equip_type,
            description: item.description || '',
            attrs: this.formatItemAttrs(item),
            level_req: item.level_req || 0,
            price: item.price || 0,
            is_bind: item.is_bind || 0,
            grid_index: item.grid_index
          }));
          console.log(`背包加载完成: ${this.items.length}个物品`);
        }
      } else {
        console.warn('背包API调用失败:', bagRes.status);
      }

      // 处理装备数据
      if (equipRes.ok) {
        const equipData = await equipRes.json();
        if (equipData.code === 200 && equipData.data) {
          // 转换服务端装备数据格式
          this.equipments = equipData.data.map(equip => ({
            id: equip.item_id,
            bagItemId: equip.bag_item_id,
            name: equip.name || `装备${equip.item_id}`,
            type: equip.type,
            type_name: this.getTypeName(equip.type),
            quality: equip.quality || 1,
            can_equip: true,
            equip_pos: this.getEquipPosName(equip.equip_type),
            equip_type: equip.equip_type,
            description: equip.description || '',
            attrs: this.formatItemAttrs(equip),
            level_req: equip.level_req || 0
          }));
          console.log(`装备加载完成: ${this.equipments.length}件装备`);
          
          // 同步到player对象
          if (this.game) {
            this.game.player.equippedItems = this.equipments;
          }
        }
      } else {
        console.warn('装备API调用失败:', equipRes.status);
      }

      // 刷新UI
      this.refreshItems();
      this.refreshEquipments();
      this.updateAttributes();

      // 如果都没有数据，使用测试数据
      if (this.items.length === 0 && this.equipments.length === 0) {
        console.log('背包为空，加载测试数据');
        this.loadTestData();
      }

    } catch (error) {
      console.error('从服务器加载背包数据失败:', error);
      // 降级到测试数据，避免背包为空
      this.loadTestData();
    }
  }

  /**
   * 获取类型名称
   */
  getTypeName(type) {
    const typeNames = {
      1: '药品', 2: '装备', 3: '材料', 4: '任务', 5: '秘籍', 6: '时装', 7: '货币'
    };
    return typeNames[type] || '未知';
  }

  /**
   * 获取装备位置名称
   */
  getEquipPosName(equipType) {
    const posNames = {
      1: 'weapon', 2: 'armor', 3: 'helmet', 4: 'necklace',
      5: 'ring', 6: 'boots', 7: 'ring', 8: 'necklace'
    };
    return posNames[equipType] || 'weapon';
  }

  /**
   * 格式化物品属性
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
}

// 导出到全局
window.Inventory = Inventory;
