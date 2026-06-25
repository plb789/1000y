/**
 * HUD 系统 - 血条/状态栏/角色信息面板
 * 管理玩家头部血条、状态栏、角色面板、目标信息面板
 * 与 UIManager 协同工作，提供游戏内 HUD 显示
 */
class HUDSystem {
  constructor(game) {
    this.game = game;
    this.initialized = false;

    // 状态栏配置
    this.barConfig = {
      hp: { color: '#ef4444', bgColor: '#3a1a1a', label: 'HP', icon: '❤' },
      mp: { color: '#60a5fa', bgColor: '#1a2a3a', label: 'MP', icon: '💧' },
      exp: { color: '#a855f7', bgColor: '#2a1a3a', label: 'EXP', icon: '✦' }
    };

    // 角色面板显示状态
    this.rolePanelVisible = false;

    // 头顶血条缓存
    this.playerHeadBar = null;

    this.init();
  }

  /**
   * 初始化 HUD 系统
   */
  init() {
    if (this.initialized) return;
    this.createStyles();
    this.enhanceTopBar();
    this.enhanceRolePanel();
    this.bindEvents();
    this.initialized = true;
    console.log('HUD系统初始化完成');
  }

  /**
   * 创建 HUD 专用样式
   */
  createStyles() {
    if (document.getElementById('hud-system-styles')) return;
    const style = document.createElement('style');
    style.id = 'hud-system-styles';
    style.textContent = `
      /* 增强顶部状态栏 */
      #topBar {
        background: linear-gradient(180deg, rgba(0,0,0,0.92) 0%, rgba(0,0,0,0.6) 70%, rgba(0,0,0,0) 100%) !important;
        height: 60px !important;
        padding: 8px 20px !important;
      }

      .player-info {
        display: flex;
        align-items: center;
        gap: 12px !important;
      }

      .player-info span {
        background: rgba(0,0,0,0.5);
        padding: 4px 12px;
        border-radius: 6px;
        border: 1px solid rgba(233, 69, 96, 0.4);
        font-size: 13px;
        color: #fff;
      }

      /* 状态条容器 */
      .hud-status-bars {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 180px;
      }

      .hud-bar {
        position: relative;
        height: 14px;
        background: rgba(0,0,0,0.6);
        border: 1px solid rgba(255,255,255,0.2);
        border-radius: 7px;
        overflow: hidden;
        display: flex;
        align-items: center;
      }

      .hud-bar-fill {
        height: 100%;
        transition: width 0.3s ease;
        border-radius: 6px;
        position: relative;
        overflow: hidden;
      }

      .hud-bar-fill::after {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 50%;
        background: linear-gradient(180deg, rgba(255,255,255,0.3) 0%, rgba(255,255,255,0) 100%);
      }

      .hud-bar-fill.hp { background: linear-gradient(90deg, #dc2626 0%, #ef4444 50%, #f87171 100%); }
      .hud-bar-fill.mp { background: linear-gradient(90deg, #2563eb 0%, #60a5fa 50%, #93c5fd 100%); }
      .hud-bar-fill.exp { background: linear-gradient(90deg, #7c3aed 0%, #a855f7 50%, #c084fc 100%); }

      .hud-bar-text {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        color: #fff;
        font-size: 10px;
        font-weight: bold;
        text-shadow: 1px 1px 2px rgba(0,0,0,0.9);
        white-space: nowrap;
        z-index: 2;
      }

      .hud-bar-label {
        position: absolute;
        left: 6px;
        top: 50%;
        transform: translateY(-50%);
        color: rgba(255,255,255,0.7);
        font-size: 10px;
        font-weight: bold;
        z-index: 2;
      }

      /* 玩家信息块 */
      .hud-player-block {
        display: flex;
        align-items: center;
        gap: 10px;
        background: rgba(0,0,0,0.5);
        padding: 6px 12px;
        border-radius: 8px;
        border: 1px solid rgba(233, 69, 96, 0.4);
      }

      .hud-player-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        background: radial-gradient(circle, #4a5568 0%, #1a1a2e 100%);
        border: 2px solid #e94560;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 20px;
        flex-shrink: 0;
      }

      .hud-player-meta {
        display: flex;
        flex-direction: column;
        gap: 2px;
      }

      .hud-player-name {
        color: #fff;
        font-size: 13px;
        font-weight: bold;
        text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
      }

      .hud-player-level {
        color: #FFD700;
        font-size: 11px;
      }

      /* 增强角色面板 */
      #rolePanel {
        position: absolute;
        top: 70px;
        left: 20px;
        width: 260px;
        background: rgba(10, 10, 20, 0.92);
        border: 2px solid #e94560;
        border-radius: 12px;
        padding: 15px;
        color: #fff;
        z-index: 101;
        display: none;
        box-shadow: 0 0 20px rgba(233, 69, 96, 0.3);
        font-family: 'Microsoft YaHei', sans-serif;
      }

      #rolePanel.show { display: block; animation: hudPanelSlideIn 0.3s ease; }

      @keyframes hudPanelSlideIn {
        from { opacity: 0; transform: translateX(-20px); }
        to { opacity: 1; transform: translateX(0); }
      }

      .role-panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding-bottom: 10px;
        margin-bottom: 10px;
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }

      .role-panel-title {
        color: #e94560;
        font-size: 16px;
        font-weight: bold;
      }

      .role-panel-close {
        background: transparent;
        border: none;
        color: #999;
        font-size: 18px;
        cursor: pointer;
        padding: 0;
        line-height: 1;
      }

      .role-panel-close:hover { color: #e94560; }

      .role-panel-section {
        margin-bottom: 12px;
      }

      .role-panel-section-title {
        color: #fbbf24;
        font-size: 12px;
        margin-bottom: 6px;
        padding-bottom: 3px;
        border-bottom: 1px solid rgba(255,255,255,0.1);
      }

      .attr-row {
        display: flex;
        justify-content: space-between;
        margin: 4px 0;
        font-size: 12px;
      }

      .attr-row .label { color: #999; }
      .attr-row .value { color: #4ade80; font-weight: bold; }
      .attr-row .value.bonus { color: #fbbf24; }

      /* 装备摘要 */
      .equip-summary {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 4px;
        margin-top: 6px;
      }

      .equip-slot-mini {
        aspect-ratio: 1;
        background: rgba(45, 55, 72, 0.6);
        border: 1px solid #4a5568;
        border-radius: 4px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        position: relative;
        cursor: pointer;
      }

      .equip-slot-mini.equipped {
        border-color: #e94560;
        background: rgba(233, 69, 96, 0.15);
      }

      .equip-slot-mini:hover {
        border-color: #fbbf24;
      }

      /* 状态栏按钮 */
      .hud-toggle-btn {
        background: rgba(0,0,0,0.6);
        border: 1px solid #e94560;
        color: #fff;
        padding: 4px 10px;
        border-radius: 6px;
        cursor: pointer;
        font-size: 12px;
        transition: all 0.2s;
      }

      .hud-toggle-btn:hover {
        background: rgba(233, 69, 96, 0.3);
        transform: translateY(-1px);
      }

      .hud-toggle-btn.active {
        background: rgba(233, 69, 96, 0.5);
      }

      /* 经验条独立样式 */
      .hud-exp-bar {
        width: 200px;
        height: 10px;
        background: rgba(0,0,0,0.6);
        border: 1px solid rgba(168, 85, 247, 0.4);
        border-radius: 5px;
        overflow: hidden;
        position: relative;
      }

      .hud-exp-fill {
        height: 100%;
        background: linear-gradient(90deg, #7c3aed 0%, #a855f7 100%);
        transition: width 0.5s ease;
      }

      .hud-exp-text {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        color: #fff;
        font-size: 9px;
        text-shadow: 1px 1px 2px rgba(0,0,0,0.9);
      }
    `;
    document.head.appendChild(style);
  }

  /**
   * 增强顶部状态栏
   */
  enhanceTopBar() {
    const topBar = document.getElementById('topBar');
    if (!topBar) return;

    // 保留右侧的 FPS 和在线人数
    const fpsCounter = topBar.querySelector('.fps-counter');
    const onlineCount = topBar.querySelector('.online-count');

    // 清空左侧 player-info
    const oldPlayerInfo = topBar.querySelector('.player-info');
    if (oldPlayerInfo) oldPlayerInfo.remove();

    // 创建新的玩家信息块
    const playerBlock = document.createElement('div');
    playerBlock.className = 'hud-player-block';
    playerBlock.innerHTML = `
      <div class="hud-player-avatar" id="hudPlayerAvatar">🧙</div>
      <div class="hud-player-meta">
        <div class="hud-player-name" id="hudPlayerName">未登录</div>
        <div class="hud-player-level" id="hudPlayerLevel">Lv.1</div>
      </div>
      <div class="hud-status-bars">
        <div class="hud-bar">
          <span class="hud-bar-label">HP</span>
          <div class="hud-bar-fill hp" id="hudHpFill" style="width:100%;"></div>
          <span class="hud-bar-text" id="hudHpText">100/100</span>
        </div>
        <div class="hud-bar">
          <span class="hud-bar-label">MP</span>
          <div class="hud-bar-fill mp" id="hudMpFill" style="width:100%;"></div>
          <span class="hud-bar-text" id="hudMpText">100/100</span>
        </div>
      </div>
      <div style="display:flex; flex-direction:column; gap:4px; align-items:center;">
        <div class="hud-exp-bar">
          <div class="hud-exp-fill" id="hudExpFill" style="width:0%;"></div>
          <span class="hud-exp-text" id="hudExpText">0%</span>
        </div>
        <div style="display:flex; gap:6px; align-items:center;">
          <span style="color:#FFD700; font-size:12px;">💰<span id="hudGold">0</span></span>
          <button class="hud-toggle-btn" id="hudRoleBtn" title="角色面板(C)">角色</button>
          <button class="hud-toggle-btn" id="hudBagBtn" title="背包(B)">背包</button>
          <button class="hud-toggle-btn" id="hudSkillBtn" title="技能(K)">技能</button>
          <button class="hud-toggle-btn" id="hudQuestBtn" title="任务(Q)">任务</button>
        </div>
      </div>
    `;

    // 插入到最前面
    topBar.insertBefore(playerBlock, topBar.firstChild);
  }

  /**
   * 增强角色面板
   */
  enhanceRolePanel() {
    const rolePanel = document.getElementById('rolePanel');
    if (!rolePanel) return;

    // 重建角色面板内容
    rolePanel.innerHTML = `
      <div class="role-panel-header">
        <div class="role-panel-title">角色信息</div>
        <button class="role-panel-close" id="rolePanelClose">×</button>
      </div>

      <div class="role-panel-section">
        <div class="role-panel-section-title">基础属性</div>
        <div class="attr-row"><span class="label">等级</span><span class="value" id="attrLevel">1</span></div>
        <div class="attr-row"><span class="label">经验</span><span class="value" id="attrExp">0/100</span></div>
        <div class="attr-row"><span class="label">金币</span><span class="value" id="attrGold">0</span></div>
      </div>

      <div class="role-panel-section">
        <div class="role-panel-section-title">生命/内力</div>
        <div class="attr-row"><span class="label">生命</span><span class="value" id="attrHP">100/100</span></div>
        <div class="attr-row"><span class="label">内力</span><span class="value" id="attrMP">100/100</span></div>
      </div>

      <div class="role-panel-section">
        <div class="role-panel-section-title">战斗属性</div>
        <div class="attr-row"><span class="label">攻击</span><span class="value" id="attrAttack">10</span></div>
        <div class="attr-row"><span class="label">防御</span><span class="value" id="attrDefense">5</span></div>
        <div class="attr-row"><span class="label">速度</span><span class="value" id="attrSpeed">10</span></div>
        <div class="attr-row"><span class="label">命中</span><span class="value" id="attrHit">50</span></div>
        <div class="attr-row"><span class="label">闪避</span><span class="value" id="attrDodge">10</span></div>
        <div class="attr-row"><span class="label">暴击</span><span class="value" id="attrCrit">5%</span></div>
      </div>

      <div class="role-panel-section">
        <div class="role-panel-section-title">装备摘要</div>
        <div class="equip-summary" id="equipSummary">
          <div class="equip-slot-mini" data-pos="1" title="武器">⚔</div>
          <div class="equip-slot-mini" data-pos="2" title="衣服">👕</div>
          <div class="equip-slot-mini" data-pos="3" title="头盔">🧢</div>
          <div class="equip-slot-mini" data-pos="4" title="护腕">🧤</div>
          <div class="equip-slot-mini" data-pos="5" title="腰带">🎽</div>
          <div class="equip-slot-mini" data-pos="6" title="鞋子">👢</div>
          <div class="equip-slot-mini" data-pos="7" title="戒指">💍</div>
          <div class="equip-slot-mini" data-pos="8" title="项链">📿</div>
        </div>
      </div>

      <div class="role-panel-section">
        <div class="role-panel-section-title">PK状态</div>
        <div class="attr-row"><span class="label">PK模式</span><span class="value" id="attrPkMode">和平</span></div>
        <div class="attr-row"><span class="label">PK值</span><span class="value" id="attrPkValue">0</span></div>
      </div>
    `;
  }

  /**
   * 绑定事件
   */
  bindEvents() {
    // 角色面板开关
    const roleBtn = document.getElementById('hudRoleBtn');
    if (roleBtn) {
      roleBtn.addEventListener('click', () => this.toggleRolePanel());
    }

    // 角色面板关闭按钮
    const closeBtn = document.getElementById('rolePanelClose');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.hideRolePanel());
    }

    // 背包按钮
    const bagBtn = document.getElementById('hudBagBtn');
    if (bagBtn) {
      bagBtn.addEventListener('click', () => {
        if (this.game.inventory) {
          this.game.inventory.toggle();
        }
      });
    }

    // 技能按钮（显示"我的技能"- 已学/已装备）
    const skillBtn = document.getElementById('hudSkillBtn');
    if (skillBtn) {
      skillBtn.addEventListener('click', (event) => {
        // 优先使用SkillBar的简单技能列表（显示已学/已装备）
        if (this.game.skillBar && typeof this.game.skillBar.toggleSkillPanel === 'function') {
          this.game.skillBar.toggleSkillPanel();
        } else if (this.game.skillPanel) {
          // 回退到完整武学图谱
          this.game.skillPanel.toggle(event); // ★ 传递event用于阻止冒泡
        }
      });
    }

    // 任务按钮
    const questBtn = document.getElementById('hudQuestBtn');
    if (questBtn) {
      questBtn.addEventListener('click', () => {
        if (this.game.questSystem) {
          this.game.questSystem.toggle();
        }
      });
    }

    // 装备槽点击（打开背包装备页）
    const equipSummary = document.getElementById('equipSummary');
    if (equipSummary) {
      equipSummary.addEventListener('click', (e) => {
        const slot = e.target.closest('.equip-slot-mini');
        if (slot) {
          const pos = parseInt(slot.dataset.pos);
          if (this.game.inventory) {
            this.game.inventory.show();
            this.game.inventory.switchTab(1); // 切换到装备页
          }
        }
      });
    }

    // 键盘快捷键
    this.keydownHandler = (e) => {
      if (this.game.state !== 'playing') return;
      // 只在输入框外响应
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

      switch (e.key.toLowerCase()) {
        case 'c':
          this.toggleRolePanel();
          break;
        case 'q':
          if (this.game.questSystem) {
            this.game.questSystem.toggle();
          }
          break;
        case 'k':
          // 优先使用完整的SkillPanel(武学图谱)，回退到SkillBar
          if (this.game.skillPanel) {
            this.game.skillPanel.toggle();
          } else if (this.game.skillBar) {
            this.game.skillBar.toggleSkillPanel?.();
          }
          break;
      }
    };
    document.addEventListener('keydown', this.keydownHandler);
  }

  /**
   * 切换角色面板显示
   */
  toggleRolePanel() {
    const panel = document.getElementById('rolePanel');
    if (!panel) return;
    if (panel.classList.contains('show')) {
      this.hideRolePanel();
    } else {
      this.showRolePanel();
    }
  }

  /**
   * 显示角色面板
   */
  showRolePanel() {
    const panel = document.getElementById('rolePanel');
    if (panel) {
      panel.classList.add('show');
      this.rolePanelVisible = true;
      this.updateRolePanel();
    }
  }

  /**
   * 隐藏角色面板
   */
  hideRolePanel() {
    const panel = document.getElementById('rolePanel');
    if (panel) {
      panel.classList.remove('show');
      this.rolePanelVisible = false;
    }
  }

  /**
   * 更新所有 HUD 显示
   */
  update() {
    this.updateTopBar();
    if (this.rolePanelVisible) {
      this.updateRolePanel();
    }
  }

  /**
   * 更新顶部状态栏
   */
  updateTopBar() {
    const player = this.game.player;
    if (!player) return;

    // 玩家名
    const nameEl = document.getElementById('hudPlayerName');
    if (nameEl) nameEl.textContent = player.name || '游客';

    // 等级
    const levelEl = document.getElementById('hudPlayerLevel');
    if (levelEl) levelEl.textContent = `Lv.${player.level || 1}`;

    // HP
    const maxHp = player.maxHp || 100;
    const hp = Math.max(0, player.hp || 0);
    const hpPercent = (hp / maxHp) * 100;
    this.setBar('hudHpFill', 'hudHpText', hpPercent, `${hp}/${maxHp}`);

    // MP
    const maxMp = player.maxMp || 100;
    const mp = Math.max(0, player.mp || 0);
    const mpPercent = (mp / maxMp) * 100;
    this.setBar('hudMpFill', 'hudMpText', mpPercent, `${mp}/${maxMp}`);

    // 经验
    const expPercent = this.calculateExpPercent();
    this.setBar('hudExpFill', 'hudExpText', expPercent, `${expPercent.toFixed(1)}%`);

    // 金币
    const goldEl = document.getElementById('hudGold');
    if (goldEl) goldEl.textContent = player.gold || 0;
  }

  /**
   * 计算经验百分比
   */
  calculateExpPercent() {
    const player = this.game.player;
    if (!player) return 0;
    const exp = player.exp || 0;
    const maxExp = player.maxExp || (player.level || 1) * 100;
    return Math.min(100, (exp / maxExp) * 100);
  }

  /**
   * 设置进度条
   */
  setBar(fillId, textId, percent, text) {
    const fill = document.getElementById(fillId);
    const textEl = document.getElementById(textId);
    if (fill) fill.style.width = `${Math.max(0, Math.min(100, percent))}%`;
    if (textEl) textEl.textContent = text;
  }

  /**
   * 更新角色面板
   */
  updateRolePanel() {
    const player = this.game.player;
    if (!player) return;

    const setText = (id, val) => {
      const el = document.getElementById(id);
      if (el) el.textContent = val;
    };

    setText('attrLevel', player.level || 1);
    setText('attrExp', `${player.exp || 0}/${player.maxExp || (player.level || 1) * 100}`);
    setText('attrGold', player.gold || 0);
    setText('attrHP', `${player.hp}/${player.maxHp}`);
    setText('attrMP', `${player.mp}/${player.maxMp}`);
    setText('attrAttack', player.attack || 0);
    setText('attrDefense', player.defense || 0);
    setText('attrSpeed', player.speed || 0);
    setText('attrHit', player.hit || 0);
    setText('attrDodge', player.dodge || 0);
    setText('attrCrit', `${player.crit || 0}%`);

    // PK状态
    const pkModes = ['和平', '队伍', '帮派', '全体'];
    setText('attrPkMode', pkModes[player.pkMode] || '和平');
    setText('attrPkValue', player.pkValue || 0);

    // 更新装备摘要
    this.updateEquipSummary();
  }

  /**
   * 更新装备摘要
   */
  updateEquipSummary() {
    const summary = document.getElementById('equipSummary');
    if (!summary) return;

    const slots = summary.querySelectorAll('.equip-slot-mini');
    const equipped = this.game.player?.equippedItems || [];

    slots.forEach(slot => {
      const pos = parseInt(slot.dataset.pos);
      const item = equipped.find(e => e.equip_type === pos);
      if (item) {
        slot.classList.add('equipped');
        slot.title = item.name || '已装备';
      } else {
        slot.classList.remove('equipped');
        // 恢复默认图标
        const defaultIcons = {1:'⚔',2:'👕',3:'🧢',4:'🧤',5:'🎽',6:'👢',7:'💍',8:'📿'};
        slot.title = defaultIcons[pos] || '';
      }
    });
  }

  /**
   * 显示玩家头顶血条（在Canvas上渲染）
   */
  renderPlayerHeadBar(ctx, tileSize) {
    const player = this.game.player;
    if (!player) return;

    const camera = this.game.mapEngine?.camera;
    if (!camera) return;

    const x = player.x * tileSize + tileSize / 2;
    const y = player.y * tileSize;

    const barWidth = tileSize * 0.9;
    const barHeight = 4;
    const barX = x - barWidth / 2;
    const barY = y - 8;

    const maxHp = player.maxHp || 100;
    const hp = Math.max(0, player.hp || 0);
    const hpPercent = hp / maxHp;

    ctx.save();
    // 背景
    ctx.fillStyle = 'rgba(0,0,0,0.7)';
    ctx.fillRect(barX - 1, barY - 1, barWidth + 2, barHeight + 2);
    // 血量
    ctx.fillStyle = hpPercent > 0.5 ? '#4ade80' : (hpPercent > 0.25 ? '#fbbf24' : '#ef4444');
    ctx.fillRect(barX, barY, barWidth * hpPercent, barHeight);
    ctx.restore();
  }

  /**
   * 销毁
   */
  destroy() {
    if (this.keydownHandler) {
      document.removeEventListener('keydown', this.keydownHandler);
      this.keydownHandler = null;
    }
  }
}

// 导出到全局
window.HUDSystem = HUDSystem;
