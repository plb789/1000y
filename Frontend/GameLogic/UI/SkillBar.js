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

    // 技能配置缓存（从服务端加载的技能基础数据）
    this.skillConfigCache = new Map(); // skillId -> skillConfig

    // ★ 新增：角色技能数据缓存（供SkillPanel复用，避免重复请求）
    this.learnedSkills = [];  // 已学技能列表
    this.equippedSkills = []; // 已装备技能列表

    // ★ 新增：加载状态标志（供SkillPanel等待）
    this.isLoading = false;   // 是否正在从服务端加载数据

    // 初始化
    this.init();
  }
  
  init() {
    this.createSkillBar();
    this.bindEvents();
    // 不再使用硬编码延迟，改为由游戏登录完成事件触发
  }

  /**
   * 登录完成后触发技能加载
   * 由 Game.js 的 enterGame 方法调用
   */
  onGameEnter() {
    if (this.game && this.game.player && this.game.player.id > 0) {
      this.loadFromServer();
    }
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

    // 武学图谱按钮
    if (!this.container.querySelector('#skillPanelBtn')) {
      const skillBtn = document.createElement('button');
      skillBtn.id = 'skillPanelBtn';
      skillBtn.className = 'skill-btn';
      skillBtn.textContent = '☯';
      skillBtn.title = '武学图谱 (K)';
      skillBtn.style.cssText = `width:36px; height:36px; border-radius:8px; cursor:pointer;
        background:linear-gradient(135deg,#1a1a2e,#16213e); border:1px solid #e94560; color:#e94560;
        font-size:18px; display:flex; align-items:center; justify-content:center;
        transition:all 0.2s; margin-left:4px;`;
      skillBtn.onmouseover = () => { skillBtn.style.background = '#e94560'; skillBtn.style.color = '#fff'; };
      skillBtn.onmouseout = () => { skillBtn.style.background = 'linear-gradient(135deg,#1a1a2e,#16213e)'; skillBtn.style.color = '#e94560'; };
      skillBtn.onclick = (event) => {
        if (this.game && this.game.skillPanel) {
          this.game.skillPanel.toggle(event); // ★ 传递event用于阻止冒泡
        }
      };
      this.container.appendChild(skillBtn);
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
    
    // 技能等级标签
    const level = document.createElement('div');
    level.className = 'skill-level';
    level.style.cssText = `
      position: absolute;
      top: 2px;
      left: 3px;
      font-size: 9px;
      color: #fbbf24;
      font-weight: bold;
      font-family: 'Microsoft YaHei', sans-serif;
      background: rgba(0, 0, 0, 0.7);
      padding: 1px 4px;
      border-radius: 4px;
    `;
    
    // 熟练度进度条
    const expBar = document.createElement('div');
    expBar.className = 'skill-exp-bar';
    expBar.style.cssText = `
      position: absolute;
      bottom: 0;
      left: 0;
      width: 100%;
      height: 2px;
      background: rgba(0, 0, 0, 0.5);
    `;
    
    const expFill = document.createElement('div');
    expFill.className = 'skill-exp-fill';
    expFill.style.cssText = `
      width: 0%;
      height: 100%;
      background: linear-gradient(90deg, #4ade80, #22c55e);
      transition: width 0.3s ease;
    `;
    expBar.appendChild(expFill);
    
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
      bottom: 4px;
      right: 3px;
      font-size: 9px;
      color: #60a5fa;
      font-family: 'Microsoft YaHei', sans-serif;
    `;
    
    // 悬浮提示框
    const tooltip = document.createElement('div');
    tooltip.className = 'skill-tooltip';
    tooltip.style.cssText = `
      position: absolute;
      bottom: calc(100% + 8px);
      left: 50%;
      transform: translateX(-50%);
      background: rgba(20, 20, 30, 0.98);
      border: 1px solid #4a5568;
      border-radius: 8px;
      padding: 10px 14px;
      min-width: 180px;
      max-width: 250px;
      display: none;
      z-index: 100;
      pointer-events: none;
      font-family: 'Microsoft YaHei', sans-serif;
      box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
    `;
    tooltip.innerHTML = `
      <div style="font-weight:bold; color:#fff; font-size:13px; margin-bottom:6px;"></div>
      <div style="font-size:11px; color:#888; margin-bottom:6px;"></div>
      <div style="display:flex; gap:12px; font-size:11px;">
        <span style="color:#ef4444;">伤害:</span>
        <span style="color:#fff;">--</span>
      </div>
      <div style="display:flex; gap:12px; font-size:11px;">
        <span style="color:#60a5fa;">MP:</span>
        <span style="color:#fff;">--</span>
      </div>
      <div style="display:flex; gap:12px; font-size:11px;">
        <span style="color:#fbbf24;">冷却:</span>
        <span style="color:#fff;">--</span>
      </div>
      <div style="display:flex; gap:12px; font-size:11px;">
        <span style="color:#888;">等级:</span>
        <span style="color:#fbbf24;">--</span>
      </div>
      <div style="display:flex; gap:12px; font-size:11px;">
        <span style="color:#4ade80;">熟练度:</span>
        <span style="color:#fff;">--</span>
      </div>
    `;
    
    // 组装
    slot.appendChild(icon);
    slot.appendChild(name);
    slot.appendChild(level);
    slot.appendChild(expBar);
    slot.appendChild(shortcut);
    slot.appendChild(mpCost);
    slot.appendChild(cooldownOverlay);
    slot.appendChild(tooltip);
    
    // 鼠标事件
    slot.addEventListener('click', () => this.onSlotClick(index));
    slot.addEventListener('contextmenu', (e) => {
      e.preventDefault();
      this.onSlotRightClick(index);
    });
    
    // 悬浮显示详情
    slot.addEventListener('mouseenter', () => this.showTooltip(index));
    slot.addEventListener('mouseleave', () => this.hideTooltip(index));
    
    return {
      element: slot,
      icon,
      name,
      level,
      expBar,
      expFill,
      shortcut,
      mpCost,
      cooldownOverlay,
      cooldownText,
      tooltip,
      skillId: null,
      cooldownEnd: 0,
      skillConfig: null,
      skillLevel: 1,
      skillExp: 0,
      skillMaxExp: 100
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
    
    const skillId = skillData.id || skillData.skill_id || 0;
    slot.skillId = skillId;
    slot.icon.textContent = this.getSkillIcon(skillData.type);
    slot.name.textContent = skillData.name || '';
    
    // 设置技能等级
    const level = skillData.level || skillData.skill_level || 1;
    slot.skillLevel = level;
    slot.level.textContent = `Lv.${level}`;
    slot.level.style.display = level > 0 ? 'block' : 'none';
    
    // 设置熟练度
    const exp = skillData.exp || skillData.experience || 0;
    const maxExp = skillData.max_exp || this.getMaxExp(level);
    slot.skillExp = exp;
    slot.skillMaxExp = maxExp;
    const expPercent = Math.min(100, (exp / maxExp) * 100);
    slot.expFill.style.width = `${expPercent}%`;
    slot.expBar.style.display = skillId > 0 ? 'block' : 'none';
    
    // MP消耗显示
    const mpCost = skillData.mp_cost || 0;
    slot.mpCost.textContent = mpCost > 0 ? `-${mpCost}` : '';
    slot.mpCost.style.display = mpCost > 0 ? 'block' : 'none';
    
    // 缓存技能配置数据到slot（供后续冷却计算使用）
    slot.skillConfig = { ...skillData };
    
    // 清除冷却状态
    this.clearSlotCooldown(index);
  }
  
  /**
   * 获取技能最大熟练度（根据等级）
   */
  getMaxExp(level) {
    return level * 100; // 每级需要100熟练度
  }
  
  /**
   * 显示悬浮提示
   */
  showTooltip(index) {
    const slot = this.skillSlots[index];
    if (!slot || !slot.skillId) return;
    
    const config = slot.skillConfig;
    if (!config) return;
    
    const tooltip = slot.tooltip;
    const parts = tooltip.querySelectorAll('div');
    
    // 技能名称
    parts[0].textContent = config.name || '未知技能';
    
    // 技能描述
    parts[1].textContent = config.description || '';
    
    // 伤害
    parts[2].querySelectorAll('span')[1].textContent = config.damage || '--';
    
    // MP消耗
    parts[3].querySelectorAll('span')[1].textContent = config.mp_cost || '0';
    
    // 冷却
    const cooldown = config.cooldown || 0;
    if (cooldown > 0) {
      parts[4].querySelectorAll('span')[1].textContent = `${cooldown}s`;
    } else if (config.attack_speed) {
      parts[4].querySelectorAll('span')[1].textContent = `攻速驱动(${config.attack_speed})`;
    } else {
      parts[4].querySelectorAll('span')[1].textContent = '--';
    }
    
    // 等级
    parts[5].querySelectorAll('span')[1].textContent = `Lv.${slot.skillLevel}`;
    
    // 熟练度
    parts[6].querySelectorAll('span')[1].textContent = `${slot.skillExp}/${slot.skillMaxExp}`;
    
    tooltip.style.display = 'block';
  }
  
  /**
   * 隐藏悬浮提示
   */
  hideTooltip(index) {
    const slot = this.skillSlots[index];
    if (!slot) return;
    slot.tooltip.style.display = 'none';
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
   * 更新冷却状态（含攻速CD显示）
   */
  updateCooldowns() {
    const now = Date.now();
    
    this.skillSlots.forEach((slot, index) => {
      if (!slot.skillId) return;
      
      const cooldownEnd = this.game.skillCooldowns.get(slot.skillId) || 0;
      const fixedCooldown = this.getSkillCooldown(slot.skillId);
      
      // 判断当前是否处于CD中
      let isCooling = false;
      let cdEndTime = 0;
      let totalCdTime = 0;

      if (fixedCooldown > 0) {
        // 有固定冷却 → 走冷却机制
        isCooling = now < cooldownEnd;
        cdEndTime = cooldownEnd;
        totalCdTime = fixedCooldown;
      } else {
        // 无固定冷却 → 走攻速机制
        const lastUse = this.game.lastSkillUseTime?.get(slot.skillId) || 0;
        const interval = this.getAttackInterval(slot.skillId);
        isCooling = (now - lastUse) < interval;
        cdEndTime = lastUse + interval;
        totalCdTime = interval;
      }

      if (isCooling) {
        const remaining = Math.ceil((cdEndTime - now) / 1000);
        slot.cooldownText.textContent = (remaining > 0 ? remaining : '') + (fixedCooldown > 0 ? 's' : '');
        slot.cooldownOverlay.style.display = 'flex';
        slot.element.classList.add('cooling');
        slot.element.style.borderColor = fixedCooldown > 0 ? '#e94560' : '#f59e0b'; // 冷却=红，攻速=橙
        
        const progress = Math.max(0, Math.min(1, (cdEndTime - now) / totalCdTime));
        slot.cooldownOverlay.style.clipPath = `polygon(0 0, 100% 0, 100% ${progress * 100}%, 0 ${progress * 100}%)`;
      } else {
        this.clearSlotCooldown(index);
        slot.element.style.borderColor = '#4a5568';
      }
    });
  }
  
  /**
   * 获取技能冷却时间（从配置读取，毫秒）
   * 规则：cooldown=0 表示无冷却，走攻速；cooldown>0 走固定冷却
   */
  getSkillCooldown(skillId) {
    if (skillId === 0) return 0; // 普通攻击无固定冷却，走攻速
    const config = this.skillConfigCache.get(skillId);
    if (config) {
      // cooldown=0 → 无冷却（走攻速）；cooldown>0 → 固定冷却(毫秒)
      if (config.cooldown === 0 || config.cooldown === undefined) return 0;
      return config.cooldown * 1000; // 秒转毫秒
    }
    return 0; // 默认无冷却
  }

  /**
   * 基于攻速计算攻击间隔（毫秒）
   * 攻速范围1-100，数值越低攻击越快
   * 优先使用技能配置中的attack_speed(武器类)，其次用角色基础攻速
   * 公式：攻速值 * 100ms，如attack_speed=6 → 600ms间隔(很快)
   */
  getAttackInterval(skillId) {
    // 优先从技能配置读取该武功的独立攻速（仅武器类有）
    if (skillId && this.skillConfigCache.has(skillId)) {
      const config = this.skillConfigCache.get(skillId);
      if (config.attack_speed && config.attack_speed > 0) return config.attack_speed * 100;
    }
    // 回退到角色基础攻速
    const attackSpeed = this.game?.player?.attackSpeed || 10;
    return Math.max(300, attackSpeed * 100); // 最低300ms保护
  }

  /**
   * 检查技能是否处于攻速CD中（仅cooldown=0的技能使用此方法）
   */
  isInAttackSpeedCooldown(skillId) {
    if (!this.game) return false;
    const lastUse = this.game.lastSkillUseTime?.get(skillId) || 0;
    const interval = this.getAttackInterval(skillId);
    return (Date.now() - lastUse) < interval;
  }

  /**
   * 获取技能MP消耗（从配置读取）
   */
  getSkillMpCost(skillId) {
    if (skillId === 0) return 0;
    const config = this.skillConfigCache.get(skillId);
    if (config && config.mp_cost !== undefined) return config.mp_cost;
    return 0; // 默认不消耗MP
  }

  /**
   * 获取技能基础伤害（从配置读取）
   */
  getSkillDamage(skillId) {
    if (skillId === 0) return 0;
    const config = this.skillConfigCache.get(skillId);
    if (config && config.damage !== undefined) return config.damage;
    return 0;
  }

  /**
   * 获取技能施法时间(毫秒)（从配置读取）
   */
  getSkillCastTime(skillId) {
    if (skillId === 0) return 0;
    const config = this.skillConfigCache.get(skillId);
    if (config && config.cast_time !== undefined) return config.cast_time;
    return 0; // 默认无施法前摇
  }

  /**
   * 获取技能类型（从配置读取）
   */
  getSkillType(skillId) {
    if (skillId === 0) return 0;
    const config = this.skillConfigCache.get(skillId);
    if (config && config.type !== undefined) return config.type;
    return 0;
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
    
    // 设置技能到槽位（合并角色技能数据与配置数据）
    skills.forEach((skill, index) => {
      if (index < this.maxSlots) {
        const config = this.skillConfigCache.get(skill.skill_id || skill.id);
        const mergedData = { ...config, ...skill };
        this.setSkill(index, mergedData);
      }
    });
    
    // 默认设置普通攻击到第一个槽位
    if (this.maxSlots > 0 && !this.skillSlots[0].skillId) {
      this.setSkill(0, { id: 0, name: '攻击', type: 0, mp_cost: 0 });
    }
  }

  /**
   * 从服务端加载技能配置（skills.json基础数据）
   * ★ 优化：通过网关代理获取（支持分布式架构）
   */
  async loadSkillConfigs() {
    try {
      // ★ 动态获取网关地址（从WebSocket连接URL转换）
      const gatewayBaseURL = this._getGatewayBaseURL();

      // 使用AbortController实现5秒超时
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(`${gatewayBaseURL}/api/skill/base/list`, {
        method: 'GET',
        signal: controller.signal
      });
      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const result = await response.json();
      if (result.code === 200 && Array.isArray(result.data)) {
        result.data.forEach(skill => {
          this.skillConfigCache.set(skill.id, skill);
        });
        console.log(`[SkillBar] 加载 ${this.skillConfigCache.size} 个技能配置`);
      } else {
        console.warn('[SkillBar] 技能配置返回格式异常:', result);
      }
    } catch (error) {
      if (error.name === 'AbortError') {
        console.warn('[SkillBar] 加载技能配置失败: 请求超时 (5s)');
      } else {
        console.warn('[SkillBar] 加载技能配置失败:', error.message);
      }
    }
  }

  /**
   * 获取网关基础URL（HTTP格式）
   * ★ 从WebSocket连接URL自动提取，支持分布式架构
   */
  _getGatewayBaseURL() {
    try {
      // 方案1：从GameWS连接URL提取（推荐）
      if (window.GameWS && window.GameWS.url) {
        const wsURL = window.GameWS.url;
        // ws://127.0.0.1:8080/ws → http://127.0.0.1:8080
        return wsURL.replace(/^ws(s?):\/\//, 'http$1://').replace(/\/ws$/, '');
      }
    } catch (e) {
      console.warn('[SkillBar] 无法从GameWS获取URL:', e.message);
    }

    // 方案2：降级为默认网关地址
    return 'http://127.0.0.1:8080';
  }

  /**
   * 从服务端加载角色已学武学
   */
  async loadFromServer(retryCount = 3) {
    // ★ 防止重复加载
    if (this.isLoading) {
      console.log('[SkillBar] 正在加载中，跳过重复请求');
      return;
    }

    this.isLoading = true; // ★ 设置加载状态
    try {
      const roleId = this.game?.player?.id;
      console.log('[SkillBar] loadFromServer - retryCount:', retryCount, 'roleId:', roleId, 'player:', this.game?.player);
      if (!roleId) {
        if (retryCount > 0) {
          // 等待登录完成后重试
          setTimeout(() => this.loadFromServer(retryCount - 1), 1500);
        }
        console.warn('[SkillBar] 角色ID不存在，等待登录...');
        return;
      }

      // 先确保配置已加载
      if (this.skillConfigCache.size === 0) {
        console.log('[SkillBar] 加载技能配置...');
        await this.loadSkillConfigs();
        console.log('[SkillBar] 技能配置加载完成，数量:', this.skillConfigCache.size);
      }

      // 通过WebSocket并行请求：已装备武学 + 所有已学武学
      console.log('[SkillBar] 请求技能列表, roleId:', roleId);
      const [equippedData, allData] = await Promise.all([
        window.GameWS.request(window.Protocol.CMD_SKILL_LIST, { role_id: roleId, type: 'equipped' }, 5000).catch(() => ({ code: 0, data: [] })),
        window.GameWS.request(window.Protocol.CMD_SKILL_LIST, { role_id: roleId, type: 'learned' }, 5000).catch(() => ({ code: 0, data: [] }))
      ]);
      console.log('[SkillBar] 已装备:', equippedData, '全部:', allData);

      // 用配置补全所有已学武学的详细信息（用于技能面板展示）
      // ★ 改为实例属性（供SkillPanel复用）
      this.learnedSkills = [];
      if (allData.code === 200 && Array.isArray(allData.data)) {
        this.learnedSkills = allData.data.map(s => {
          const skillId = s.skill_id || s.id;
          const config = this.skillConfigCache.get(skillId) || {};
          return {
            id: skillId,
            skill_id: skillId,
            name: config.name || '未知武学',
            type: config.type || 0,
            level: s.level || 1,
            exp: s.exp || 0,
            is_equip: s.is_equip || 0,
            mp_cost: config.mp_cost || 0,
            cooldown: config.cooldown || 0,
            damage: config.damage || 0,
            range: config.range || 1,
            cast_time: config.cast_time || 0,
            attack_speed: config.attack_speed || 0,
          };
        });
        // 存储到player对象供技能面板使用
        this.game.player.skills = this.learnedSkills;
        console.log('[SkillBar] 已加载', this.learnedSkills.length, '个已学技能');
      }

      // ★ 存储已装备技能数据（供SkillPanel复用）
      this.equippedSkills = [];
      if (equippedData.code === 200 && equippedData.data) {
        this.equippedSkills = equippedData.data;
      }

      // 合并数据：优先显示已装备的技能到快捷栏
      let skillsToShow = [];

      if (equippedData.code === 200 && equippedData.data) {
        // 已装备的技能 - 补全信息（包含熟练度）
        skillsToShow = equippedData.data.map(s => {
          const skillId = s.skill_id || s.id;
          const config = this.skillConfigCache.get(skillId) || {};
          const level = s.level || s.skill_level || 1;
          const exp = s.exp || s.experience || 0;
          const maxExp = s.max_exp || level * 100;
          return {
            ...s,
            id: skillId,
            skill_id: skillId,
            name: config.name || '未知',
            type: config.type || 0,
            mp_cost: config.mp_cost || 0,
            cooldown: config.cooldown || 0,
            damage: config.damage || 0,
            attack_speed: config.attack_speed || 0,
            level: level,
            exp: exp,
            max_exp: maxExp,
            is_equipped: true
          };
        });
      } else {
        // ★ 无已装备技能时：显示空槽位（由setSkills默认添加普通攻击）
        skillsToShow = [];
      }

      // 更新技能栏（无论是否有技能都需要更新）
      this.setSkills(skillsToShow);
    } catch (error) {
      console.warn('[SkillBar] 从服务端加载技能失败:', error);
    } finally {
      this.isLoading = false; // ★ 重置加载状态
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
