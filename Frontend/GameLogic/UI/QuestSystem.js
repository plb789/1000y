/**
 * 任务系统 - 管理任务接取、追踪、进度、奖励
 * 与服务端通过 CMD_QUEST_* 协议交互
 */
class QuestSystem {
  constructor(game) {
    this.game = game;
    this.initialized = false;

    // 任务数据
    this.activeQuests = [];      // 进行中的任务
    this.completedQuests = [];   // 已完成未领奖的任务
    this.availableQuests = [];   // 可接取的任务
    this.finishedQuests = [];    // 已完成已领奖的任务

    // UI状态
    this.isOpen = false;
    this.currentTab = 'active'; // active, available, finished
    this.container = null;
    this.trackerEl = null;

    // 任务类型映射
    this.questTypes = {
      1: { name: '主线', color: '#e94560' },
      2: { name: '支线', color: '#60a5fa' },
      3: { name: '日常', color: '#4ade80' },
      4: { name: '周常', color: '#a855f7' },
      5: { name: '活动', color: '#fbbf24' }
    };

    // 目标类型映射
    this.targetTypes = {
      1: { name: '击杀', icon: '⚔' },
      2: { name: '采集', icon: '🌿' },
      3: { name: '对话', icon: '💬' },
      4: { name: '探索', icon: '🗺' }
    };

    this.init();
  }

  /**
   * 初始化
   */
  init() {
    if (this.initialized) return;
    this.createStyles();
    this.createPanel();
    this.createTracker();
    this.bindEvents();
    this.initialized = true;
    console.log('任务系统初始化完成');
  }

  /**
   * 创建样式
   */
  createStyles() {
    if (document.getElementById('quest-system-styles')) return;
    const style = document.createElement('style');
    style.id = 'quest-system-styles';
    style.textContent = `
      .quest-panel {
        position: absolute;
        top: 70px;
        right: 20px;
        width: 380px;
        max-height: 500px;
        background: rgba(10, 10, 20, 0.95);
        border: 2px solid #e94560;
        border-radius: 12px;
        padding: 15px;
        color: #fff;
        z-index: 101;
        display: none;
        box-shadow: 0 0 20px rgba(233, 69, 96, 0.3);
        font-family: 'Microsoft YaHei', sans-serif;
        flex-direction: column;
      }

      .quest-panel.show { display: flex; animation: questPanelSlideIn 0.3s ease; }

      @keyframes questPanelSlideIn {
        from { opacity: 0; transform: translateX(20px); }
        to { opacity: 1; transform: translateX(0); }
      }

      .quest-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding-bottom: 10px;
        margin-bottom: 10px;
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }

      .quest-title {
        color: #e94560;
        font-size: 16px;
        font-weight: bold;
      }

      .quest-close {
        background: transparent;
        border: none;
        color: #999;
        font-size: 18px;
        cursor: pointer;
        padding: 0;
        line-height: 1;
      }

      .quest-close:hover { color: #e94560; }

      .quest-tabs {
        display: flex;
        gap: 4px;
        margin-bottom: 10px;
      }

      .quest-tab {
        flex: 1;
        padding: 6px;
        text-align: center;
        background: rgba(45, 55, 72, 0.6);
        border: 1px solid #4a5568;
        border-radius: 6px;
        color: #999;
        cursor: pointer;
        font-size: 12px;
        transition: all 0.2s;
      }

      .quest-tab:hover { background: rgba(233, 69, 96, 0.2); }

      .quest-tab.active {
        background: rgba(233, 69, 96, 0.4);
        color: #fff;
        border-color: #e94560;
      }

      .quest-list {
        flex: 1;
        overflow-y: auto;
        max-height: 380px;
        padding-right: 4px;
      }

      .quest-list::-webkit-scrollbar { width: 6px; }
      .quest-list::-webkit-scrollbar-track { background: rgba(0,0,0,0.3); }
      .quest-list::-webkit-scrollbar-thumb { background: #e94560; border-radius: 3px; }

      .quest-item {
        background: rgba(45, 55, 72, 0.6);
        border: 1px solid #4a5568;
        border-radius: 8px;
        padding: 10px;
        margin-bottom: 8px;
        transition: all 0.2s;
      }

      .quest-item:hover { border-color: #e94560; }

      .quest-item.completed {
        border-color: #4ade80;
        background: rgba(74, 222, 128, 0.1);
      }

      .quest-item-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 6px;
      }

      .quest-item-name {
        color: #fff;
        font-weight: bold;
        font-size: 13px;
      }

      .quest-type-badge {
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 10px;
        font-weight: bold;
      }

      .quest-desc {
        color: #ccc;
        font-size: 11px;
        margin-bottom: 6px;
        line-height: 1.4;
      }

      .quest-progress {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 11px;
        color: #fbbf24;
        margin-bottom: 6px;
      }

      .quest-progress-bar {
        flex: 1;
        height: 6px;
        background: rgba(0,0,0,0.5);
        border-radius: 3px;
        margin: 0 8px;
        overflow: hidden;
      }

      .quest-progress-fill {
        height: 100%;
        background: linear-gradient(90deg, #4ade80 0%, #22c55e 100%);
        transition: width 0.3s;
      }

      .quest-rewards {
        display: flex;
        gap: 8px;
        font-size: 10px;
        color: #999;
        margin-bottom: 6px;
      }

      .quest-reward-item {
        background: rgba(0,0,0,0.4);
        padding: 2px 6px;
        border-radius: 3px;
      }

      .quest-actions {
        display: flex;
        gap: 6px;
      }

      .quest-btn {
        flex: 1;
        padding: 5px;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 11px;
        font-weight: bold;
        transition: all 0.2s;
      }

      .quest-btn-accept {
        background: #4ade80;
        color: #000;
      }

      .quest-btn-accept:hover { background: #22c55e; }

      .quest-btn-complete {
        background: #fbbf24;
        color: #000;
      }

      .quest-btn-complete:hover { background: #f59e0b; }

      .quest-btn-abandon {
        background: #6b7280;
        color: #fff;
      }

      .quest-btn-abandon:hover { background: #4b5563; }

      .quest-empty {
        text-align: center;
        color: #999;
        padding: 30px 10px;
        font-size: 12px;
      }

      /* 任务追踪器（屏幕右侧） */
      .quest-tracker {
        position: absolute;
        top: 70px;
        right: 20px;
        width: 240px;
        background: rgba(0, 0, 0, 0.7);
        border: 1px solid rgba(233, 69, 96, 0.4);
        border-radius: 8px;
        padding: 8px;
        color: #fff;
        z-index: 100;
        font-size: 11px;
        pointer-events: none;
      }

      .quest-tracker-title {
        color: #e94560;
        font-size: 12px;
        font-weight: bold;
        margin-bottom: 6px;
        padding-bottom: 4px;
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }

      .quest-tracker-item {
        margin-bottom: 6px;
        padding: 4px;
        background: rgba(0,0,0,0.4);
        border-radius: 4px;
      }

      .quest-tracker-name {
        color: #fff;
        font-weight: bold;
        margin-bottom: 2px;
      }

      .quest-tracker-progress {
        color: #fbbf24;
        font-size: 10px;
      }
    `;
    document.head.appendChild(style);
  }

  /**
   * 创建任务面板
   */
  createPanel() {
    const panel = document.createElement('div');
    panel.id = 'questPanel';
    panel.className = 'quest-panel';
    panel.innerHTML = `
      <div class="quest-header">
        <div class="quest-title">任务列表</div>
        <button class="quest-close" id="questCloseBtn">×</button>
      </div>
      <div class="quest-tabs">
        <div class="quest-tab active" data-tab="active">进行中 (<span id="activeCount">0</span>)</div>
        <div class="quest-tab" data-tab="available">可接取 (<span id="availableCount">0</span>)</div>
        <div class="quest-tab" data-tab="finished">已完成 (<span id="finishedCount">0</span>)</div>
      </div>
      <div class="quest-list" id="questList"></div>
    `;
    document.body.appendChild(panel);
    this.container = panel;
  }

  /**
   * 创建任务追踪器（屏幕右侧实时显示）
   */
  createTracker() {
    const tracker = document.createElement('div');
    tracker.id = 'questTracker';
    tracker.className = 'quest-tracker';
    tracker.style.display = 'none';
    tracker.innerHTML = `
      <div class="quest-tracker-title">任务追踪</div>
      <div id="questTrackerList"></div>
    `;
    document.body.appendChild(tracker);
    this.trackerEl = tracker;
  }

  /**
   * 绑定事件
   */
  bindEvents() {
    // 关闭按钮
    const closeBtn = document.getElementById('questCloseBtn');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.hide());
    }

    // 标签切换
    const tabs = this.container.querySelectorAll('.quest-tab');
    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        this.currentTab = tab.dataset.tab;
        this.renderQuestList();
      });
    });
  }

  /**
   * 切换显示
   */
  toggle() {
    if (this.isOpen) {
      this.hide();
    } else {
      this.show();
    }
  }

  /**
   * 显示任务面板
   */
  show() {
    this.isOpen = true;
    this.container.classList.add('show');
    // 隐藏追踪器（避免重叠）
    this.trackerEl.style.display = 'none';
    // 请求任务列表
    this.requestQuestList();
    this.renderQuestList();
  }

  /**
   * 隐藏任务面板
   */
  hide() {
    this.isOpen = false;
    this.container.classList.remove('show');
    // 显示追踪器
    this.updateTracker();
  }

  /**
   * 请求任务列表
   */
  requestQuestList() {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_LIST || 6001, {});
    }
  }

  /**
   * 接取任务
   */
  acceptQuest(questId) {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_ACCEPT || 6002, {
        quest_id: questId
      });
    }
    // 乐观更新：从可接取列表移除，添加到进行中
    const idx = this.availableQuests.findIndex(q => q.id === questId);
    if (idx !== -1) {
      const quest = this.availableQuests.splice(idx, 1)[0];
      quest.status = 1;
      quest.progress = 0;
      this.activeQuests.push(quest);
      this.renderQuestList();
      this.updateTracker();
    }
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('接取任务成功', 'success', 1500);
    }
  }

  /**
   * 完成任务（领取奖励）
   */
  completeQuest(questId) {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_COMPLETE || 6003, {
        quest_id: questId
      });
    }
    // 乐观更新
    const idx = this.completedQuests.findIndex(q => q.id === questId);
    if (idx !== -1) {
      const quest = this.completedQuests.splice(idx, 1)[0];
      quest.status = 3;
      this.finishedQuests.push(quest);
      this.renderQuestList();
      this.updateTracker();
    }
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('任务完成，已领取奖励', 'success', 2000);
    }
  }

  /**
   * 放弃任务
   */
  abandonQuest(questId) {
    if (!window.confirm('确认放弃该任务？')) return;
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_ABANDON || 6004, {
        quest_id: questId
      });
    }
    const idx = this.activeQuests.findIndex(q => q.id === questId);
    if (idx !== -1) {
      const quest = this.activeQuests.splice(idx, 1)[0];
      quest.status = 0;
      quest.progress = 0;
      this.availableQuests.push(quest);
      this.renderQuestList();
      this.updateTracker();
    }
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('已放弃任务', 'info', 1500);
    }
  }

  /**
   * 处理服务端推送的任务列表
   */
  handleQuestList(data) {
    if (!data) return;
    this.activeQuests = data.active_quests || [];
    this.completedQuests = data.completed_quests || [];
    this.availableQuests = data.available_quests || [];
    this.finishedQuests = data.finished_quests || [];

    // 更新计数
    this.updateTabCounts();

    if (this.isOpen) {
      this.renderQuestList();
    }
    this.updateTracker();
  }

  /**
   * 处理任务进度更新
   */
  handleQuestProgress(data) {
    if (!data || !data.quest_id) return;
    const quest = this.activeQuests.find(q => q.id === data.quest_id);
    if (quest) {
      quest.progress = data.progress || quest.progress;
      if (data.status !== undefined) {
        quest.status = data.status;
        // 如果任务完成（status=2），移到已完成列表
        if (quest.status === 2) {
          this.activeQuests = this.activeQuests.filter(q => q.id !== quest.id);
          this.completedQuests.push(quest);
          if (this.game.uiManager?.toast) {
            this.game.uiManager.toast(`任务完成：${quest.name}`, 'success', 2000);
          }
        }
      }
      if (this.isOpen) this.renderQuestList();
      this.updateTracker();
    }
  }

  /**
   * 更新标签计数
   */
  updateTabCounts() {
    const activeCount = document.getElementById('activeCount');
    const availableCount = document.getElementById('availableCount');
    const finishedCount = document.getElementById('finishedCount');
    if (activeCount) activeCount.textContent = this.activeQuests.length;
    if (availableCount) availableCount.textContent = this.availableQuests.length;
    if (finishedCount) finishedCount.textContent = this.finishedQuests.length;
  }

  /**
   * 渲染任务列表
   */
  renderQuestList() {
    const listEl = document.getElementById('questList');
    if (!listEl) return;

    this.updateTabCounts();

    let quests = [];
    switch (this.currentTab) {
      case 'active':
        quests = this.activeQuests;
        break;
      case 'available':
        quests = this.availableQuests;
        break;
      case 'finished':
        quests = this.finishedQuests;
        break;
    }

    if (quests.length === 0) {
      const emptyText = {
        active: '暂无进行中的任务',
        available: '没有可接取的任务',
        finished: '尚未完成任何任务'
      };
      listEl.innerHTML = `<div class="quest-empty">${emptyText[this.currentTab]}</div>`;
      return;
    }

    listEl.innerHTML = quests.map(quest => this.renderQuestItem(quest)).join('');

    // 绑定按钮事件
    listEl.querySelectorAll('.quest-btn-accept').forEach(btn => {
      btn.addEventListener('click', () => this.acceptQuest(parseInt(btn.dataset.id)));
    });
    listEl.querySelectorAll('.quest-btn-complete').forEach(btn => {
      btn.addEventListener('click', () => this.completeQuest(parseInt(btn.dataset.id)));
    });
    listEl.querySelectorAll('.quest-btn-abandon').forEach(btn => {
      btn.addEventListener('click', () => this.abandonQuest(parseInt(btn.dataset.id)));
    });
  }

  /**
   * 渲染单个任务项
   */
  renderQuestItem(quest) {
    const typeInfo = this.questTypes[quest.type] || { name: '未知', color: '#999' };
    const targetInfo = this.targetTypes[quest.target_type] || { name: '目标', icon: '◆' };
    const isCompleted = quest.status === 2;
    const progress = quest.progress || 0;
    const targetCount = quest.target_count || 1;
    const progressPercent = Math.min(100, (progress / targetCount) * 100);

    // 奖励显示
    const rewards = [];
    if (quest.reward_exp) rewards.push(`<span class="quest-reward-item">经验 +${quest.reward_exp}</span>`);
    if (quest.reward_gold) rewards.push(`<span class="quest-reward-item">金币 +${quest.reward_gold}</span>`);
    if (quest.reward_item_id) rewards.push(`<span class="quest-reward-item">物品 x${quest.reward_item_count || 1}</span>`);

    // 操作按钮
    let actions = '';
    if (this.currentTab === 'available') {
      actions = `<button class="quest-btn quest-btn-accept" data-id="${quest.id}">接取任务</button>`;
    } else if (this.currentTab === 'active') {
      if (isCompleted) {
        actions = `<button class="quest-btn quest-btn-complete" data-id="${quest.id}">领取奖励</button>`;
      } else {
        actions = `<button class="quest-btn quest-btn-abandon" data-id="${quest.id}">放弃</button>`;
      }
    } else if (this.currentTab === 'finished') {
      actions = `<div style="text-align:center; color:#4ade80; font-size:11px;">✓ 已完成</div>`;
    }

    return `
      <div class="quest-item ${isCompleted ? 'completed' : ''}">
        <div class="quest-item-header">
          <div class="quest-item-name">${quest.name}</div>
          <div class="quest-type-badge" style="background:${typeInfo.color}; color:#fff;">${typeInfo.name}</div>
        </div>
        <div class="quest-desc">${quest.description || ''}</div>
        <div class="quest-progress">
          <span>${targetInfo.icon} ${targetInfo.name}</span>
          <div class="quest-progress-bar">
            <div class="quest-progress-fill" style="width:${progressPercent}%;"></div>
          </div>
          <span>${progress}/${targetCount}</span>
        </div>
        ${rewards.length > 0 ? `<div class="quest-rewards">${rewards.join('')}</div>` : ''}
        <div class="quest-actions">${actions}</div>
      </div>
    `;
  }

  /**
   * 更新任务追踪器
   */
  updateTracker() {
    if (!this.trackerEl) return;
    const listEl = document.getElementById('questTrackerList');
    if (!listEl) return;

    // 面板打开时不显示追踪器
    if (this.isOpen) {
      this.trackerEl.style.display = 'none';
      return;
    }

    if (this.activeQuests.length === 0) {
      this.trackerEl.style.display = 'none';
      return;
    }

    this.trackerEl.style.display = 'block';

    // 只显示前3个任务
    const showQuests = this.activeQuests.slice(0, 3);
    listEl.innerHTML = showQuests.map(quest => {
      const progress = quest.progress || 0;
      const targetCount = quest.target_count || 1;
      const targetInfo = this.targetTypes[quest.target_type] || { name: '目标' };
      return `
        <div class="quest-tracker-item">
          <div class="quest-tracker-name">${quest.name}</div>
          <div class="quest-tracker-progress">${targetInfo.icon} ${targetInfo.name}: ${progress}/${targetCount}</div>
        </div>
      `;
    }).join('');
  }

  /**
   * 通知怪物击杀（用于击杀类任务进度）
   */
  onMonsterKilled(monsterId) {
    // 检查是否有击杀类任务与此怪物相关
    this.activeQuests.forEach(quest => {
      if (quest.target_type === 1 && quest.target_id === monsterId) {
        // 发送进度更新请求（服务端会验证）
        if (window.GameWS && window.GameWS.isConnected) {
          window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
            quest_id: quest.id,
            target_type: 1,
            target_id: monsterId
          });
        }
      }
    });
  }

  /**
   * 通知物品采集（用于采集类任务进度）
   */
  onItemGathered(itemId) {
    this.activeQuests.forEach(quest => {
      if (quest.target_type === 2 && quest.target_id === itemId) {
        if (window.GameWS && window.GameWS.isConnected) {
          window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
            quest_id: quest.id,
            target_type: 2,
            target_id: itemId
          });
        }
      }
    });
  }
}

// 导出到全局
window.QuestSystem = QuestSystem;
