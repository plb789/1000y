// AchievementPanel.js - 成就系统UI
class AchievementPanel {
  constructor(game) {
    this.game = game;
    this.panel = null;
    this.isVisible = false;
    this.achievements = [];
    this.stats = null;
    this.initialized = false;

    // 成就类型名称映射
    this.typeNames = {
      1: '任务成就',
      2: '战斗成就',
      3: '收集成就',
      4: '探索成就',
      5: '社交成就'
    };

    // 成就图标映射
    this.typeIcons = {
      1: '📜',  // 任务
      2: '⚔️',  // 战斗
      3: '🎒',  // 收集
      4: '🗺️',  // 探索
      5: '👥'   // 社交
    };
  }

  init() {
    if (this.initialized) return;
    this.createStyles();
    this.createPanel();
    this.bindEvents();
    this.loadAchievements();
    this.initialized = true;
    console.log('[Achievement] 成就系统初始化完成');
  }

  createStyles() {
    const style = document.createElement('style');
    style.textContent = `
      /* 成就面板样式 */
      .achievement-panel {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 580px;
        max-height: 80vh;
        background: rgba(10, 15, 30, 0.95);
        border: 2px solid #fbbf24;
        border-radius: 12px;
        color: #fff;
        font-family: 'Microsoft YaHei', sans-serif;
        z-index: 1001;
        display: none;
        flex-direction: column;
        box-shadow: 0 0 30px rgba(251, 191, 36, 0.3);
      }

      .achievement-panel.show {
        display: flex;
      }

      /* 成就面板头部 */
      .achievement-header {
        padding: 15px 20px;
        background: linear-gradient(135deg, #92400e, #78350f);
        border-bottom: 1px solid #fbbf24;
        display: flex;
        justify-content: space-between;
        align-items: center;
        border-radius: 10px 10px 0 0;
      }

      .achievement-header h2 {
        margin: 0;
        font-size: 18px;
        color: #fbbf24;
      }

      .achievement-stats {
        display: flex;
        gap: 15px;
        font-size: 12px;
      }

      .achievement-stat {
        display: flex;
        align-items: center;
        gap: 4px;
      }

      .achievement-stat-value {
        color: #fbbf24;
        font-weight: bold;
      }

      .achievement-close {
        background: none;
        border: none;
        color: #fff;
        font-size: 20px;
        cursor: pointer;
        padding: 5px;
        line-height: 1;
      }

      .achievement-close:hover {
        color: #fbbf24;
      }

      /* 成就标签页 */
      .achievement-tabs {
        display: flex;
        padding: 10px 15px;
        background: rgba(0, 0, 0, 0.3);
        gap: 8px;
      }

      .achievement-tab {
        padding: 6px 14px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid transparent;
        border-radius: 15px;
        color: #9ca3af;
        cursor: pointer;
        font-size: 12px;
        transition: all 0.2s;
      }

      .achievement-tab:hover {
        background: rgba(255, 255, 255, 0.15);
        color: #fff;
      }

      .achievement-tab.active {
        background: rgba(251, 191, 36, 0.2);
        border-color: #fbbf24;
        color: #fbbf24;
      }

      /* 成就内容区 */
      .achievement-content {
        flex: 1;
        overflow-y: auto;
        padding: 15px;
      }

      .achievement-list {
        display: grid;
        gap: 12px;
      }

      /* 成就项 */
      .achievement-item {
        background: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
        border-radius: 8px;
        padding: 12px 15px;
        display: flex;
        gap: 12px;
        transition: all 0.2s;
      }

      .achievement-item:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(251, 191, 36, 0.3);
      }

      .achievement-item.unlocked {
        border-color: rgba(251, 191, 36, 0.5);
        background: rgba(251, 191, 36, 0.1);
      }

      .achievement-item.locked {
        opacity: 0.6;
      }

      /* 成就图标 */
      .achievement-icon {
        width: 48px;
        height: 48px;
        background: rgba(251, 191, 36, 0.2);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 22px;
        flex-shrink: 0;
      }

      .achievement-item.unlocked .achievement-icon {
        background: linear-gradient(135deg, #fbbf24, #f59e0b);
        box-shadow: 0 0 15px rgba(251, 191, 36, 0.5);
      }

      .achievement-item.locked .achievement-icon {
        background: rgba(255, 255, 255, 0.1);
      }

      /* 成就信息 */
      .achievement-info {
        flex: 1;
        min-width: 0;
      }

      .achievement-name {
        font-size: 14px;
        font-weight: bold;
        color: #fff;
        margin-bottom: 4px;
      }

      .achievement-item.locked .achievement-name {
        color: #9ca3af;
      }

      .achievement-desc {
        font-size: 12px;
        color: #9ca3af;
        margin-bottom: 6px;
      }

      .achievement-progress {
        display: flex;
        align-items: center;
        gap: 8px;
      }

      .achievement-progress-bar {
        flex: 1;
        height: 6px;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
        overflow: hidden;
      }

      .achievement-progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #fbbf24, #f59e0b);
        border-radius: 3px;
        transition: width 0.3s;
      }

      .achievement-item.unlocked .achievement-progress-fill {
        width: 100% !important;
        background: linear-gradient(90deg, #4ade80, #22c55e);
      }

      .achievement-progress-text {
        font-size: 11px;
        color: #9ca3af;
        min-width: 40px;
        text-align: right;
      }

      /* 成就奖励 */
      .achievement-rewards {
        display: flex;
        gap: 10px;
        margin-top: 6px;
      }

      .achievement-reward {
        font-size: 10px;
        padding: 2px 6px;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
        color: #d1d5db;
      }

      .achievement-point {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 11px;
        color: #fbbf24;
      }

      /* 成就解锁动画 */
      @keyframes achievementUnlock {
        0% { transform: scale(0.5); opacity: 0; }
        50% { transform: scale(1.1); }
        100% { transform: scale(1); opacity: 1; }
      }

      .achievement-unlock-effect {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        background: linear-gradient(135deg, rgba(30, 20, 10, 0.95), rgba(50, 30, 10, 0.95));
        border: 3px solid #fbbf24;
        border-radius: 15px;
        padding: 30px 40px;
        text-align: center;
        z-index: 9999;
        animation: achievementUnlock 0.5s ease-out;
        box-shadow: 0 0 50px rgba(251, 191, 36, 0.6);
      }

      .achievement-unlock-effect .unlock-title {
        color: #fbbf24;
        font-size: 14px;
        margin-bottom: 10px;
      }

      .achievement-unlock-effect .unlock-icon {
        font-size: 50px;
        margin: 15px 0;
      }

      .achievement-unlock-effect .unlock-name {
        color: #fff;
        font-size: 18px;
        font-weight: bold;
        margin-bottom: 8px;
      }

      .achievement-unlock-effect .unlock-desc {
        color: #9ca3af;
        font-size: 12px;
        margin-bottom: 15px;
      }

      .achievement-unlock-effect .unlock-rewards {
        display: flex;
        justify-content: center;
        gap: 15px;
        color: #fbbf24;
        font-size: 13px;
      }

      /* 空状态 */
      .achievement-empty {
        text-align: center;
        padding: 40px;
        color: #9ca3af;
      }

      .achievement-empty-icon {
        font-size: 40px;
        margin-bottom: 10px;
      }

      /* 滚动条 */
      .achievement-content::-webkit-scrollbar {
        width: 6px;
      }

      .achievement-content::-webkit-scrollbar-track {
        background: rgba(255, 255, 255, 0.05);
      }

      .achievement-content::-webkit-scrollbar-thumb {
        background: rgba(251, 191, 36, 0.3);
        border-radius: 3px;
      }

      .achievement-content::-webkit-scrollbar-thumb:hover {
        background: rgba(251, 191, 36, 0.5);
      }
    `;
    document.head.appendChild(style);
  }

  createPanel() {
    this.panel = document.createElement('div');
    this.panel.className = 'achievement-panel';
    this.panel.innerHTML = `
      <div class="achievement-header">
        <h2>🏆 成就系统</h2>
        <div class="achievement-stats">
          <div class="achievement-stat">
            <span>成就:</span>
            <span class="achievement-stat-value" id="achievementCount">0/0</span>
          </div>
          <div class="achievement-stat">
            <span>点数:</span>
            <span class="achievement-stat-value" id="achievementPoints">0</span>
          </div>
        </div>
        <button class="achievement-close" id="achievementClose">×</button>
      </div>
      <div class="achievement-tabs">
        <div class="achievement-tab active" data-type="0">全部</div>
        <div class="achievement-tab" data-type="1">任务</div>
        <div class="achievement-tab" data-type="2">战斗</div>
        <div class="achievement-tab" data-type="3">收集</div>
        <div class="achievement-tab" data-type="4">探索</div>
        <div class="achievement-tab" data-type="5">社交</div>
      </div>
      <div class="achievement-content">
        <div class="achievement-list" id="achievementList">
          <div class="achievement-empty">
            <div class="achievement-empty-icon">📜</div>
            <div>正在加载成就...</div>
          </div>
        </div>
      </div>
    `;
    document.body.appendChild(this.panel);
  }

  bindEvents() {
    // 关闭按钮
    document.getElementById('achievementClose').addEventListener('click', () => this.hide());

    // 标签切换
    document.querySelectorAll('.achievement-tab').forEach(tab => {
      tab.addEventListener('click', (e) => {
        document.querySelectorAll('.achievement-tab').forEach(t => t.classList.remove('active'));
        e.target.classList.add('active');
        this.renderAchievements(parseInt(e.target.dataset.type));
      });
    });

    // ESC关闭
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.isVisible) {
        this.hide();
      }
    });
  }

  show() {
    this.panel.classList.add('show');
    this.isVisible = true;
    this.loadAchievements();
  }

  hide() {
    this.panel.classList.remove('show');
    this.isVisible = false;
  }

  toggle() {
    if (this.isVisible) {
      this.hide();
    } else {
      this.show();
    }
  }

  /**
   * 通过WebSocket加载成就列表
   */
  loadAchievements() {
    const roleId = this.game?.player?.id;
    if (!roleId) {
      this.loadMockData();
      return;
    }

    // 通过WebSocket发送成就列表请求
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_ACHIEVEMENT_LIST || 6101, {
        role_id: roleId
      });
    } else {
      console.log('[Achievement] WebSocket未连接，使用测试数据');
      this.loadMockData();
    }
  }

  /**
   * 处理服务端推送的成就列表
   */
  handleAchievementList(data) {
    console.log('[Achievement] 收到成就列表:', data);
    if (data.code === 200 && data.data) {
      this.achievements = data.data;
      this.updateStats();
      this.renderAchievements(0);
    } else {
      this.loadMockData();
    }
  }

  /**
   * 处理服务端推送的成就统计
   */
  handleAchievementStats(data) {
    console.log('[Achievement] 收到成就统计:', data);
    if (data.code === 200 && data.data) {
      this.stats = data.data;
      this.updateStats();
    }
  }

  /**
   * 处理服务端推送的成就解锁消息
   */
  handleAchievementUnlocked(data) {
    console.log('[Achievement] 收到成就解锁:', data);

    if (data.type === 'achievement_unlocked') {
      // 更新本地成就数据
      const achievement = this.achievements.find(a => a.id === data.data.achievement_id);
      if (achievement) {
        achievement.unlocked = true;
        achievement.progress = achievement.target_count;
        this.updateStats();
        this.renderAchievements(0);
      }

      // 显示解锁动画
      this.showUnlockEffect(data.data);
    }
  }

  loadMockData() {
    // 测试数据
    this.achievements = [
      {
        id: 1,
        name: '初出茅庐',
        description: '完成第一个任务',
        type: 1,
        condition: 'quest_complete',
        target_id: 1,
        target_count: 1,
        progress: 1,
        unlocked: true,
        reward_exp: 100,
        reward_gold: 50,
        point: 10,
        icon: '📜'
      },
      {
        id: 2,
        name: '除暴安良',
        description: '击杀10只野猪',
        type: 2,
        condition: 'monster_kill',
        target_id: 101,
        target_count: 10,
        progress: 6,
        unlocked: false,
        reward_exp: 200,
        reward_gold: 100,
        point: 20,
        icon: '⚔️'
      },
      {
        id: 3,
        name: '收藏家',
        description: '收集10个道具',
        type: 3,
        condition: 'item_collect',
        target_id: 201,
        target_count: 10,
        progress: 3,
        unlocked: false,
        reward_exp: 150,
        reward_gold: 75,
        point: 15,
        icon: '🎒'
      }
    ];
    this.updateStats();
    this.renderAchievements(0);
  }

  updateStats() {
    const total = this.achievements.length;
    const unlocked = this.achievements.filter(a => a.unlocked).length;
    const points = this.achievements.filter(a => a.unlocked).reduce((sum, a) => sum + (a.point || 0), 0);

    const countEl = document.getElementById('achievementCount');
    const pointsEl = document.getElementById('achievementPoints');
    if (countEl) countEl.textContent = `${unlocked}/${total}`;
    if (pointsEl) pointsEl.textContent = points;
  }

  renderAchievements(type) {
    const listEl = document.getElementById('achievementList');
    if (!listEl) return;

    let filtered = this.achievements;
    if (type > 0) {
      filtered = this.achievements.filter(a => a.type === type);
    }

    if (filtered.length === 0) {
      listEl.innerHTML = `
        <div class="achievement-empty">
          <div class="achievement-empty-icon">📭</div>
          <div>暂无成就</div>
        </div>
      `;
      return;
    }

    listEl.innerHTML = filtered.map(achievement => {
      const icon = this.typeIcons[achievement.type] || '🏆';
      const progress = achievement.progress || 0;
      const target = achievement.target_count || 1;
      const percent = Math.min(100, (progress / target) * 100);

      return `
        <div class="achievement-item ${achievement.unlocked ? 'unlocked' : 'locked'}">
          <div class="achievement-icon">${icon}</div>
          <div class="achievement-info">
            <div class="achievement-name">${achievement.name}</div>
            <div class="achievement-desc">${achievement.description}</div>
            <div class="achievement-progress">
              <div class="achievement-progress-bar">
                <div class="achievement-progress-fill" style="width: ${percent}%"></div>
              </div>
              <div class="achievement-progress-text">${progress}/${target}</div>
            </div>
            <div class="achievement-rewards">
              ${achievement.reward_exp ? `<span class="achievement-reward">经验+${achievement.reward_exp}</span>` : ''}
              ${achievement.reward_gold ? `<span class="achievement-reward">金币+${achievement.reward_gold}</span>` : ''}
              ${achievement.point ? `<span class="achievement-point">★ ${achievement.point}</span>` : ''}
            </div>
          </div>
        </div>
      `;
    }).join('');
  }

  // 显示成就解锁动画
  showUnlockEffect(achievement) {
    const effect = document.createElement('div');
    effect.className = 'achievement-unlock-effect';
    effect.innerHTML = `
      <div class="unlock-title">🏆 成就解锁</div>
      <div class="unlock-icon">${this.typeIcons[achievement.type] || '🏆'}</div>
      <div class="unlock-name">${achievement.name}</div>
      <div class="unlock-desc">${achievement.description}</div>
      <div class="unlock-rewards">
        ${achievement.reward_exp ? `<span>经验 +${achievement.reward_exp}</span>` : ''}
        ${achievement.reward_gold ? `<span>金币 +${achievement.reward_gold}</span>` : ''}
        ${achievement.reward_honor ? `<span>声望 +${achievement.reward_honor}</span>` : ''}
      </div>
    `;
    document.body.appendChild(effect);

    // 创建粒子效果
    this.createParticles(window.innerWidth / 2, window.innerHeight / 2);

    // 3秒后移除
    setTimeout(() => {
      effect.remove();
    }, 3000);
  }

  createParticles(x, y) {
    const colors = ['#fbbf24', '#f59e0b', '#4ade80', '#22c55e'];
    for (let i = 0; i < 30; i++) {
      const particle = document.createElement('div');
      particle.style.cssText = `
        position: fixed;
        left: ${x}px;
        top: ${y}px;
        width: 8px;
        height: 8px;
        background: ${colors[Math.floor(Math.random() * colors.length)]};
        border-radius: 50%;
        pointer-events: none;
        z-index: 10000;
      `;
      document.body.appendChild(particle);

      const angle = (Math.PI * 2 * i) / 30;
      const velocity = 100 + Math.random() * 150;
      const vx = Math.cos(angle) * velocity;
      const vy = Math.sin(angle) * velocity;

      let posX = x;
      let posY = y;
      let opacity = 1;

      const animate = () => {
        posX += vx * 0.02;
        posY += vy * 0.02;
        opacity -= 0.02;
        particle.style.left = posX + 'px';
        particle.style.top = posY + 'px';
        particle.style.opacity = opacity;
        if (opacity > 0) {
          requestAnimationFrame(animate);
        } else {
          particle.remove();
        }
      };
      requestAnimationFrame(animate);
    }
  }
}

// 导出成就面板实例
window.AchievementPanel = AchievementPanel;
