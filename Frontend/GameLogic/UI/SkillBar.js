/**
 * 技能栏管理器
 * 管理技能快捷栏的UI、冷却、快捷键绑定等功能
 */
class SkillBar {
  constructor(game) {
    this.game = game;
    this.container = null;
    this.skillSlots = [];
    this.maxSlots = 8;
    this.shortcutKeys = ['1', '2', '3', '4', '5', '6', '7', '8']; // 快捷键映射（8个槽位）
    
    // 初始化
    this.init();
  }
  
  init() {
    this.createSkillBar();
    this.bindEvents();
  }
  
  /**
   * 创建技能栏
   */
  createSkillBar() {
    // 获取现有的bottomBar
    this.container = document.getElementById('bottomBar');
    if (!this.container) {
      console.warn('bottomBar not found');
      return;
    }
    
    // 保留的按钮列表
    const preserveButtons = ['settingsBtn', 'inventoryBtn'];
    
    // 逐个移除子节点，保留指定按钮
    Array.from(this.container.children).forEach(child => {
      if (!preserveButtons.includes(child.id)) {
        child.remove();
      }
    });
    
    // 如果没有找到settingsBtn，则重新创建
    if (!this.container.querySelector('#settingsBtn')) {
      const btn = document.createElement('button');
      btn.id = 'settingsBtn';
      btn.className = 'settings-btn';
      btn.textContent = '⚙️';
      btn.onclick = () => this.openSettings();
      this.container.appendChild(btn);
    }
    
    // 如果没有找到inventoryBtn，则重新创建
    if (!this.container.querySelector('#inventoryBtn')) {
      const btn = document.createElement('button');
      btn.id = 'inventoryBtn';
      btn.className = 'skill-btn inventory-btn';
      btn.textContent = '🎒';
      btn.onclick = () => {
        if (this.game && this.game.inventory) {
          this.game.inventory.toggle();
        }
      };
      this.container.appendChild(btn);
    }
    
    // 创建技能栏容器
    const barContainer = document.createElement('div');
    barContainer.className = 'skill-bar-container';
    barContainer.style.cssText = `
      display: flex;
      gap: 8px;
      align-items: center;
      padding: 0 10px;
    `;
    
    // 创建技能槽位
    for (let i = 0; i < this.maxSlots; i++) {
      const slot = this.createSkillSlot(i);
      this.skillSlots.push(slot);
      barContainer.appendChild(slot.element);
    }
    
    // 添加到容器
    this.container.insertBefore(barContainer, this.container.firstChild);
    
    // 重新绑定设置按钮事件
    this.bindSettingsEvent();
  }
  
  /**
   * 创建单个技能槽
   */
  createSkillSlot(index) {
    const slot = document.createElement('div');
    slot.className = 'skill-slot';
    slot.dataset.index = index;
    
    // 技能槽样式
    slot.style.cssText = `
      width: 56px;
      height: 56px;
      background: linear-gradient(135deg, rgba(45, 55, 72, 0.9), rgba(26, 32, 44, 0.9));
      border: 2px solid #4a5568;
      border-radius: 8px;
      position: relative;
      cursor: pointer;
      transition: all 0.2s ease;
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: center;
      overflow: hidden;
    `;
    
    // 冷却遮罩
    const cooldownOverlay = document.createElement('div');
    cooldownOverlay.className = 'skill-cooldown-overlay';
    cooldownOverlay.style.cssText = `
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.6);
      border-radius: 6px;
      display: none;
      justify-content: center;
      align-items: center;
      z-index: 2;
      pointer-events: none;
    `;
    
    // 冷却时间文字
    const cooldownText = document.createElement('span');
    cooldownText.className = 'skill-cooldown-text';
    cooldownText.style.cssText = `
      color: #fff;
      font-size: 16px;
      font-weight: bold;
      font-family: 'Microsoft YaHei', sans-serif;
      text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
    `;
    cooldownOverlay.appendChild(cooldownText);
    
    // 技能图标
    const icon = document.createElement('div');
    icon.className = 'skill-icon';
    icon.style.cssText = `
      font-size: 24px;
      margin-bottom: 2px;
    `;
    
    // 技能名称
    const name = document.createElement('div');
    name.className = 'skill-name';
    name.style.cssText = `
      font-size: 10px;
      color: #ccc;
      text-align: center;
      max-width: 50px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    `;
    
    // 快捷键提示
    const shortcut = document.createElement('div');
    shortcut.className = 'skill-shortcut';
    shortcut.style.cssText = `
      position: absolute;
      top: 2px;
      right: 3px;
      font-size: 10px;
      color: rgba(255, 255, 255, 0.6);
      font-family: 'Microsoft YaHei', sans-serif;
    `;
    shortcut.textContent = this.shortcutKeys[index] || '';
    
    // MP消耗提示
    const mpCost = document.createElement('div');
    mpCost.className = 'skill-mp-cost';
    mpCost.style.cssText = `
      position: absolute;
      bottom: 2px;
      right: 3px;
      font-size: 9px;
      color: #60a5fa;
      font-family: 'Microsoft YaHei', sans-serif;
    `;
    
    // 组装
    slot.appendChild(icon);
    slot.appendChild(name);
    slot.appendChild(shortcut);
    slot.appendChild(mpCost);
    slot.appendChild(cooldownOverlay);
    
    // 鼠标事件
    slot.addEventListener('click', () => this.onSlotClick(index));
    slot.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      this.onSlotRightClick(index);
    });
    
    return {
      element: slot,
      icon,
      name,
      shortcut,
      mpCost,
      cooldownOverlay,
      cooldownText,
      skillId: null,
      cooldownEnd: 0
    };
  }
  
  /**
   * 绑定事件
   */
  bindEvents() {
    // 键盘快捷键 - 将处理函数保存为实例方法，以便后续移除
    this.keydownHandler = (e) => {
      if (this.game.state !== 'playing') return;
      
      const keyIndex = this.shortcutKeys.indexOf(e.key);
      if (keyIndex !== -1) {
        this.useSkillBySlot(keyIndex);
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
   * 打开/关闭设置面板
   */
  openSettings() {
    const settingsPanel = document.getElementById('settingsPanel');
    if (settingsPanel) {
      settingsPanel.style.display = settingsPanel.style.display === 'none' ? 'block' : 'none';
    }
  }
  
  /**
   * 绑定设置按钮事件
   * 注意：settingsBtn 的 onclick 已在 init() 中通过 openSettings() 绑定，此处无需重复绑定
   */
  bindSettingsEvent() {
    // 预留扩展：未来如需额外事件可在此添加
  }
  
  /**
   * 技能槽点击
   */
  onSlotClick(index) {
    if (index >= this.maxSlots) return;
    
    const slot = this.skillSlots[index];
    if (!slot || !slot.skillId) return;
    
    this.game.useSkill(slot.skillId);
  }
  
  /**
   * 技能槽右键点击（打开技能详情/配置）
   */
  onSlotRightClick(index) {
    const slot = this.skillSlots[index];
    if (!slot || !slot.skillId) return;
    
    // 可以在这里打开技能详情面板
    console.log('技能详情:', slot.skillId);
  }
  
  /**
   * 使用技能（通过槽位索引）
   */
  useSkillBySlot(index) {
    if (index >= this.maxSlots) return;
    
    const slot = this.skillSlots[index];
    if (!slot || !slot.skillId) {
      // 空槽位，尝试使用普通攻击
      this.game.useSkill(0);
      return;
    }
    
    this.game.useSkill(slot.skillId);
  }
  
  /**
   * 设置技能到槽位
   */
  setSkill(index, skillData) {
    if (index >= this.maxSlots) return;
    
    const slot = this.skillSlots[index];
    if (!slot) return;
    
    slot.skillId = skillData.id;
    slot.icon.textContent = this.getSkillIcon(skillData.type);
    slot.name.textContent = skillData.name;
    slot.mpCost.textContent = skillData.mp_cost ? `${skillData.mp_cost}MP` : '';
    
    // 清除冷却状态
    this.clearSlotCooldown(index);
  }
  
  /**
   * 清空技能槽
   */
  clearSlot(index) {
    if (index >= this.maxSlots) return;
    
    const slot = this.skillSlots[index];
    if (!slot) return;
    
    slot.skillId = null;
    slot.icon.textContent = '';
    slot.name.textContent = '';
    slot.mpCost.textContent = '';
    this.clearSlotCooldown(index);
  }
  
  /**
   * 清除槽位冷却状态
   */
  clearSlotCooldown(index) {
    const slot = this.skillSlots[index];
    if (!slot) return;
    
    slot.cooldownOverlay.style.display = 'none';
    slot.element.classList.remove('cooling');
  }
  
  /**
   * 更新冷却状态
   */
  updateCooldowns() {
    const now = Date.now();
    
    this.skillSlots.forEach((slot, index) => {
      if (!slot.skillId) return;
      
      const cooldownEnd = this.game.skillCooldowns.get(slot.skillId) || 0;
      
      if (now < cooldownEnd) {
        // 显示冷却
        const remaining = Math.ceil((cooldownEnd - now) / 1000);
        slot.cooldownText.textContent = remaining + 's';
        slot.cooldownOverlay.style.display = 'flex';
        slot.element.classList.add('cooling');
        slot.element.style.borderColor = '#e94560';
        
        // 更新冷却进度
        const totalCooldown = this.getSkillCooldown(slot.skillId);
        const progress = (cooldownEnd - now) / totalCooldown;
        slot.cooldownOverlay.style.clipPath = `polygon(0 0, 100% 0, 100% ${progress * 100}%, 0 ${progress * 100}%)`;
      } else {
        // 清除冷却
        this.clearSlotCooldown(index);
        slot.element.style.borderColor = '#4a5568';
      }
    });
  }
  
  /**
   * 获取技能冷却时间
   */
  getSkillCooldown(skillId) {
    if (skillId === 0) return 1000; // 普通攻击
    return 3000; // 默认技能冷却
  }
  
  /**
   * 获取技能图标
   */
  getSkillIcon(type) {
    const icons = {
      0: '⚔️',  // 无类型/普通攻击
      1: '🧘',  // 内功
      2: '👊',  // 外功
      3: '💨',  // 身法
      4: '🛡️',  // 护体
      5: '🥋',  // 拳法
      6: '⚔️',  // 剑法
      7: '🔪',  // 刀法
      8: '🔫',  // 枪法
      9: '🪓'   // 斧法
    };
    return icons[type] || '✨';
  }
  
  /**
   * 设置玩家技能列表
   */
  setSkills(skills) {
    // 清空所有槽位
    for (let i = 0; i < this.maxSlots; i++) {
      this.clearSlot(i);
    }
    
    // 设置技能到槽位
    skills.forEach((skill, index) => {
      if (index < this.maxSlots) {
        this.setSkill(index, skill);
      }
    });
    
    // 默认设置普通攻击到第一个槽位
    if (this.maxSlots > 0 && !this.skillSlots[0].skillId) {
      this.setSkill(0, { id: 0, name: '攻击', type: 0, mp_cost: 0 });
    }
  }
  
  /**
   * 更新UI
   */
  update() {
    this.updateCooldowns();
    this.updateMpStatus();
  }

  /**
   * 更新MP状态（MP不足时技能图标变灰）
   */
  updateMpStatus() {
    const playerMp = this.game.player?.mp || 0;
    this.skillSlots.forEach((slot) => {
      if (!slot.skillId) return;
      const mpCost = parseInt(slot.mpCost.textContent) || 0;
      if (mpCost > playerMp) {
        slot.element.classList.add('mp-insufficient');
      } else {
        slot.element.classList.remove('mp-insufficient');
      }
    });
  }

  /**
   * 切换技能面板显示（供HUD按钮调用）
   */
  toggleSkillPanel() {
    let panel = document.getElementById('skillPanel');
    if (!panel) {
      panel = this.createSkillPanel();
    }
    if (panel.style.display === 'none' || !panel.style.display) {
      panel.style.display = 'block';
      this.refreshSkillPanel();
    } else {
      panel.style.display = 'none';
    }
  }

  /**
   * 创建技能面板（显示所有已学技能）
   */
  createSkillPanel() {
    const panel = document.createElement('div');
    panel.id = 'skillPanel';
    panel.style.cssText = `
      position: absolute;
      top: 70px;
      right: 20px;
      width: 320px;
      max-height: 400px;
      overflow-y: auto;
      background: rgba(10, 10, 20, 0.95);
      border: 2px solid #e94560;
      border-radius: 12px;
      padding: 15px;
      color: #fff;
      z-index: 101;
      display: none;
      box-shadow: 0 0 20px rgba(233, 69, 96, 0.3);
    `;
    panel.innerHTML = `
      <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:10px; padding-bottom:8px; border-bottom:1px solid rgba(233,69,96,0.3);">
        <div style="color:#e94560; font-size:16px; font-weight:bold;">技能列表</div>
        <button id="skillPanelClose" style="background:transparent; border:none; color:#999; font-size:18px; cursor:pointer;">×</button>
      </div>
      <div id="skillListContainer"></div>
      <div style="margin-top:10px; padding-top:8px; border-top:1px solid rgba(255,255,255,0.1); font-size:11px; color:#999;">
        提示：拖动技能到快捷栏可设置快捷键
      </div>
    `;
    document.body.appendChild(panel);

    panel.querySelector('#skillPanelClose').addEventListener('click', () => {
      panel.style.display = 'none';
    });

    return panel;
  }

  /**
   * 刷新技能面板内容
   */
  refreshSkillPanel() {
    const container = document.getElementById('skillListContainer');
    if (!container) return;

    const skills = this.game.player?.skills || [];
    if (skills.length === 0) {
      container.innerHTML = '<div style="text-align:center; color:#999; padding:20px;">暂未学习技能</div>';
      return;
    }

    container.innerHTML = skills.map(skill => `
      <div class="skill-item" data-skill-id="${skill.id}" style="
        display:flex; align-items:center; gap:10px; padding:8px; margin:4px 0;
        background:rgba(45,55,72,0.6); border:1px solid #4a5568; border-radius:6px;
        cursor:grab;
      ">
        <div style="font-size:24px;">${this.getSkillIcon(skill.type)}</div>
        <div style="flex:1;">
          <div style="color:#fff; font-weight:bold;">${skill.name} ${skill.level ? 'Lv.' + skill.level : ''}</div>
          <div style="color:#999; font-size:11px;">
            ${skill.mp_cost ? `MP:${skill.mp_cost} ` : ''}
            ${skill.cooldown ? `CD:${skill.cooldown}s ` : ''}
            ${skill.damage ? `伤害:${skill.damage}` : ''}
          </div>
        </div>
        <div style="color:#fbbf24; font-size:11px;">拖拽</div>
      </div>
    `).join('');

    // 添加拖拽事件
    container.querySelectorAll('.skill-item').forEach(item => {
      item.draggable = true;
      item.addEventListener('dragstart', (e) => {
        e.dataTransfer.setData('skillId', item.dataset.skillId);
      });
    });
  }
}

// 导出到全局
window.SkillBar = SkillBar;
