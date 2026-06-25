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
    this.questLog = [];          // 任务日志记录

    // UI状态
    this.isOpen = false;
    this.currentTab = 'active'; // active, available, finished
    this.container = null;
    this.trackerEl = null;
    this.detailPanel = null;
    this.selectedQuest = null;

    // 任务类型映射
    this.questTypes = {
      1: { name: '主线', color: '#e94560', icon: '⭐' },
      2: { name: '支线', color: '#60a5fa', icon: '📖' },
      3: { name: '日常', color: '#4ade80', icon: '☀️' },
      4: { name: '周常', color: '#a855f7', icon: '🌙' },
      5: { name: '活动', color: '#fbbf24', icon: '🎉' }
    };

    // 目标类型映射
    this.targetTypes = {
      1: { name: '击杀', icon: '⚔', color: '#ef4444' },
      2: { name: '采集', icon: '🌿', color: '#22c55e' },
      3: { name: '对话', icon: '💬', color: '#60a5fa' },
      4: { name: '探索', icon: '🗺', color: '#fbbf24' },
      5: { name: '收集', icon: '📦', color: '#a855f7' },
      6: { name: '护送', icon: '🛡', color: '#3b82f6' }
    };

    // 追踪设置
    this.trackEnabled = true;
    this.autoTrack = true; // 自动追踪最近接取的任务

    this.init();
  }

  /**
   * 初始化
   */
  init() {
    if (this.initialized) return;
    this.createStyles();
    this.createPanel();
    this.createDetailPanel();
    this.createTracker();
    this.bindEvents();
    this.loadQuestListFromServer();
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

      .quest-btn-batch-accept {
        margin-left: auto;
        padding: 4px 10px;
        background: #4ade80;
        color: #000;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 10px;
        font-weight: bold;
        transition: all 0.2s;
      }

      .quest-btn-batch-accept:hover { background: #22c55e; }

      .quest-btn-share {
        padding: 3px 8px;
        background: #60a5fa;
        color: #fff;
        border: none;
        border-radius: 3px;
        cursor: pointer;
        font-size: 10px;
        margin-right: 4px;
        transition: all 0.2s;
      }

      .quest-btn-share:hover { background: #3b82f6; }

      .quest-empty {
        text-align: center;
        color: #999;
        padding: 30px 10px;
        font-size: 12px;
      }

      /* 任务详情面板 */
      .quest-detail-panel {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        width: 420px;
        background: rgba(10, 10, 20, 0.98);
        border: 2px solid #e94560;
        border-radius: 12px;
        padding: 20px;
        color: #fff;
        z-index: 200;
        display: none;
        box-shadow: 0 0 30px rgba(233, 69, 96, 0.4);
        font-family: 'Microsoft YaHei', sans-serif;
      }

      .quest-detail-panel.show { display: block; animation: questDetailFadeIn 0.3s ease; }

      @keyframes questDetailFadeIn {
        from { opacity: 0; transform: translate(-50%, -50%) scale(0.95); }
        to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
      }

      .quest-detail-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 15px;
        padding-bottom: 10px;
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }

      .quest-detail-title {
        font-size: 18px;
        font-weight: bold;
        color: #fff;
      }

      .quest-detail-close {
        background: transparent;
        border: none;
        color: #999;
        font-size: 20px;
        cursor: pointer;
      }

      .quest-detail-close:hover { color: #e94560; }

      .quest-detail-type {
        display: inline-block;
        padding: 3px 10px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: bold;
        margin-bottom: 10px;
      }

      .quest-detail-desc {
        color: #ccc;
        font-size: 13px;
        line-height: 1.5;
        margin-bottom: 15px;
      }

      .quest-detail-objectives {
        margin-bottom: 15px;
      }

      .quest-detail-objectives-title {
        color: #e94560;
        font-size: 13px;
        font-weight: bold;
        margin-bottom: 8px;
      }

      .quest-detail-objective {
        display: flex;
        align-items: center;
        padding: 8px;
        background: rgba(0,0,0,0.3);
        border-radius: 6px;
        margin-bottom: 6px;
      }

      .quest-detail-objective-icon {
        margin-right: 10px;
        font-size: 16px;
      }

      .quest-detail-objective-info {
        flex: 1;
      }

      .quest-detail-objective-name {
        color: #fff;
        font-size: 12px;
      }

      .quest-detail-objective-progress {
        color: #fbbf24;
        font-size: 11px;
        margin-top: 2px;
      }

      .quest-detail-objective-bar {
        width: 100%;
        height: 4px;
        background: rgba(0,0,0,0.5);
        border-radius: 2px;
        margin-top: 6px;
        overflow: hidden;
      }

      .quest-detail-objective-fill {
        height: 100%;
        background: linear-gradient(90deg, #4ade80, #22c55e);
        transition: width 0.3s;
      }

      .quest-detail-rewards {
        background: rgba(0,0,0,0.3);
        border-radius: 8px;
        padding: 12px;
        margin-bottom: 15px;
      }

      .quest-detail-rewards-title {
        color: #4ade80;
        font-size: 13px;
        font-weight: bold;
        margin-bottom: 8px;
      }

      .quest-detail-reward-list {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
      }

      .quest-detail-reward-item {
        display: flex;
        align-items: center;
        background: rgba(0,0,0,0.4);
        padding: 4px 10px;
        border-radius: 4px;
        font-size: 12px;
      }

      .quest-detail-reward-icon {
        margin-right: 5px;
      }

      .quest-detail-actions {
        display: flex;
        gap: 10px;
      }

      .quest-detail-btn {
        flex: 1;
        padding: 10px;
        border: none;
        border-radius: 6px;
        cursor: pointer;
        font-size: 13px;
        font-weight: bold;
        transition: all 0.2s;
      }

      .quest-detail-btn-accept {
        background: #4ade80;
        color: #000;
      }

      .quest-detail-btn-accept:hover { background: #22c55e; }

      .quest-detail-btn-complete {
        background: #fbbf24;
        color: #000;
      }

      .quest-detail-btn-complete:hover { background: #f59e0b; }

      .quest-detail-btn-abandon {
        background: #6b7280;
        color: #fff;
      }

      .quest-detail-btn-abandon:hover { background: #4b5563; }

      .quest-detail-btn-share {
        background: #60a5fa;
        color: #fff;
      }

      .quest-detail-btn-share:hover { background: #3b82f6; }

      .quest-detail-btn-close {
        background: #4a5568;
        color: #fff;
      }

      .quest-detail-btn-close:hover { background: #2d3748; }

      /* 类型筛选按钮 */
      .quest-filter-bar {
        display: flex;
        gap: 4px;
        margin-bottom: 10px;
        overflow-x: auto;
        padding-bottom: 4px;
      }

      .quest-filter-bar::-webkit-scrollbar { height: 4px; }
      .quest-filter-bar::-webkit-scrollbar-track { background: rgba(0,0,0,0.2); }
      .quest-filter-bar::-webkit-scrollbar-thumb { background: #4a5568; border-radius: 2px; }

      .quest-filter-btn {
        padding: 4px 10px;
        background: rgba(45, 55, 72, 0.6);
        border: 1px solid #4a5568;
        border-radius: 4px;
        color: #999;
        cursor: pointer;
        font-size: 11px;
        white-space: nowrap;
        transition: all 0.2s;
      }

      .quest-filter-btn:hover { background: rgba(233, 69, 96, 0.2); }

      .quest-filter-btn.active {
        background: rgba(233, 69, 96, 0.4);
        color: #fff;
        border-color: #e94560;
      }

      /* 任务追踪器（屏幕右侧） */
      .quest-tracker {
        position: absolute;
        top: 70px;
        right: 20px;
        width: 240px;
        background: rgba(0, 0, 0, 0.8);
        border: 1px solid rgba(233, 69, 96, 0.4);
        border-radius: 8px;
        padding: 10px;
        color: #fff;
        z-index: 100;
        font-size: 11px;
        pointer-events: auto;
        cursor: pointer;
      }

      .quest-tracker:hover {
        border-color: #e94560;
        background: rgba(0, 0, 0, 0.9);
      }

      .quest-tracker-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
        padding-bottom: 5px;
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }

      .quest-tracker-title {
        color: #e94560;
        font-size: 12px;
        font-weight: bold;
      }

      .quest-tracker-toggle {
        font-size: 10px;
        color: #666;
        cursor: pointer;
      }

      .quest-tracker-toggle:hover { color: #e94560; }

      .quest-tracker-item {
        margin-bottom: 8px;
        padding: 6px;
        background: rgba(0,0,0,0.4);
        border-radius: 4px;
        transition: all 0.2s;
      }

      .quest-tracker-item:hover {
        background: rgba(233, 69, 96, 0.15);
        border-left: 2px solid #e94560;
      }

      .quest-tracker-item.completed {
        opacity: 0.7;
      }

      .quest-tracker-name {
        color: #fff;
        font-weight: bold;
        margin-bottom: 3px;
        font-size: 11px;
      }

      .quest-tracker-objectives {
        margin-left: 4px;
      }

      .quest-tracker-objective {
        display: flex;
        align-items: center;
        font-size: 10px;
        color: #fbbf24;
        margin-bottom: 2px;
      }

      .quest-tracker-objective-icon {
        margin-right: 4px;
      }

      .quest-tracker-objective.completed {
        color: #4ade80;
        text-decoration: line-through;
      }

      .quest-tracker-expand {
        text-align: center;
        color: #666;
        font-size: 10px;
        padding: 4px;
        margin-top: 4px;
        cursor: pointer;
      }

      .quest-tracker-expand:hover { color: #e94560; }

      /* 任务日志 */
      .quest-log {
        margin-top: 10px;
        padding-top: 10px;
        border-top: 1px solid rgba(233, 69, 96, 0.2);
      }

      .quest-log-title {
        color: #999;
        font-size: 11px;
        margin-bottom: 6px;
      }

      .quest-log-item {
        color: #666;
        font-size: 10px;
        margin-bottom: 3px;
        padding-left: 10px;
        position: relative;
      }

      .quest-log-item::before {
        content: '•';
        position: absolute;
        left: 0;
        color: #e94560;
      }

      /* 任务追踪器样式 */
      .quest-tracker {
        position: absolute;
        top: 120px;
        right: 20px;
        width: 260px;
        background: rgba(10, 10, 20, 0.9);
        border: 2px solid #60a5fa;
        border-radius: 10px;
        padding: 12px;
        color: #fff;
        z-index: 100;
        font-family: 'Microsoft YaHei', sans-serif;
        box-shadow: 0 0 15px rgba(96, 165, 250, 0.3);
      }

      .quest-tracker-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(96, 165, 250, 0.3);
      }

      .quest-tracker-title {
        color: #60a5fa;
        font-size: 14px;
        font-weight: bold;
      }

      .quest-tracker-toggle {
        color: #999;
        font-size: 11px;
        cursor: pointer;
      }

      .quest-tracker-toggle:hover { color: #60a5fa; }

      .quest-tracker-item {
        padding: 8px;
        background: rgba(0,0,0,0.3);
        border-radius: 6px;
        margin-bottom: 8px;
      }

      .quest-tracker-item:last-child { margin-bottom: 0; }

      .quest-tracker-item.completed {
        border-left: 3px solid #4ade80;
      }

      .quest-tracker-name {
        font-size: 12px;
        font-weight: bold;
        margin-bottom: 4px;
        color: #fbbf24;
      }

      .quest-tracker-objectives { font-size: 11px; }

      .quest-tracker-objective {
        display: flex;
        align-items: center;
        gap: 4px;
        color: #ccc;
        padding: 2px 0;
      }

      .quest-tracker-objective.completed {
        color: #4ade80;
        text-decoration: line-through;
      }

      .quest-tracker-objective-icon { font-size: 10px; }

      .quest-tracker-expand {
        text-align: center;
        color: #60a5fa;
        font-size: 11px;
        margin-top: 10px;
        padding-top: 8px;
        border-top: 1px solid rgba(96, 165, 250, 0.3);
        cursor: pointer;
      }

      .quest-tracker-expand:hover { color: #93c5fd; }

      /* 任务完成动画 */
      .quest-complete-effect {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        z-index: 9999;
        pointer-events: none;
        animation: questCompleteAnim 1.5s ease-out forwards;
      }

      @keyframes questCompleteAnim {
        0% { opacity: 1; transform: translate(-50%, -50%) scale(0.5); }
        50% { opacity: 1; transform: translate(-50%, -50%) scale(1.2); }
        100% { opacity: 0; transform: translate(-50%, -50%) scale(2); }
      }

      .quest-complete-text {
        font-size: 36px;
        font-weight: bold;
        color: #4ade80;
        text-shadow: 0 0 20px #4ade80, 0 0 40px #4ade80;
        white-space: nowrap;
        animation: questCompleteTextAnim 1.5s ease-out forwards;
      }

      @keyframes questCompleteTextAnim {
        0% { opacity: 0; transform: scale(0.5); }
        30% { opacity: 1; transform: scale(1.1); }
        50% { transform: scale(1); }
        100% { opacity: 0; transform: translateY(-30px); }
      }

      /* 奖励发放动画 */
      .quest-reward-effect {
        position: fixed;
        bottom: 100px;
        left: 50%;
        transform: translateX(-50%);
        z-index: 9999;
        pointer-events: none;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
        animation: rewardPopAnim 2s ease-out forwards;
      }

      @keyframes rewardPopAnim {
        0% { opacity: 0; transform: translateX(-50%) translateY(50px); }
        20% { opacity: 1; transform: translateX(-50%) translateY(0); }
        80% { opacity: 1; transform: translateX(-50%) translateY(0); }
        100% { opacity: 0; transform: translateX(-50%) translateY(-30px); }
      }

      .quest-reward-item-anim {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 10px 20px;
        background: rgba(0, 0, 0, 0.8);
        border-radius: 25px;
        border: 2px solid;
        font-size: 16px;
        font-weight: bold;
      }

      .quest-reward-exp { border-color: #fbbf24; color: #fbbf24; }
      .quest-reward-gold { border-color: #f59e0b; color: #f59e0b; }
      .quest-reward-honor { border-color: #a855f7; color: #a855f7; }
      .quest-reward-item-anim .reward-icon { font-size: 20px; }

      /* 任务进度更新动画 */
      .quest-progress-flash {
        animation: progressFlash 0.5s ease-out;
      }

      @keyframes progressFlash {
        0% { background-color: rgba(74, 222, 128, 0.5); }
        100% { background-color: transparent; }
      }

      /* 追踪器任务项点击效果 */
      .quest-tracker-item {
        cursor: pointer;
        transition: all 0.2s;
      }

      .quest-tracker-item:hover {
        background: rgba(96, 165, 250, 0.2);
        transform: translateX(-3px);
      }

      /* 任务指引高亮 */
      .quest-guide-highlight {
        position: fixed;
        border: 3px solid #fbbf24;
        border-radius: 10px;
        box-shadow: 0 0 20px rgba(251, 191, 36, 0.5);
        z-index: 9998;
        pointer-events: none;
        animation: guidePulse 1.5s infinite;
      }

      @keyframes guidePulse {
        0%, 100% { opacity: 0.6; transform: scale(1); }
        50% { opacity: 1; transform: scale(1.02); }
      }

      /* 粒子效果容器 */
      .quest-particles {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        pointer-events: none;
        z-index: 9997;
      }

      .quest-particle {
        position: absolute;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        animation: particleFade 1s ease-out forwards;
      }

      @keyframes particleFade {
        0% { opacity: 1; transform: scale(1); }
        100% { opacity: 0; transform: scale(0) translateY(-100px); }
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
        <button class="quest-btn-batch-accept" id="batchAcceptBtn" title="一键接取所有可接任务">一键接取</button>
      </div>
      <div class="quest-filter-bar" id="questFilterBar">
        <div class="quest-filter-btn active" data-filter="all">全部</div>
        <div class="quest-filter-btn" data-filter="1">⭐ 主线</div>
        <div class="quest-filter-btn" data-filter="2">📖 支线</div>
        <div class="quest-filter-btn" data-filter="3">☀️ 日常</div>
        <div class="quest-filter-btn" data-filter="4">🌙 周常</div>
        <div class="quest-filter-btn" data-filter="5">🎉 活动</div>
      </div>
      <div class="quest-list" id="questList"></div>
    `;
    document.body.appendChild(panel);
    this.container = panel;
  }

  /**
   * 创建任务详情面板
   */
  createDetailPanel() {
    const panel = document.createElement('div');
    panel.id = 'questDetailPanel';
    panel.className = 'quest-detail-panel';
    panel.innerHTML = `
      <div class="quest-detail-header">
        <div class="quest-detail-title" id="questDetailTitle">任务详情</div>
        <button class="quest-detail-close" id="questDetailCloseBtn">×</button>
      </div>
      <div class="quest-detail-type" id="questDetailType"></div>
      <div class="quest-detail-desc" id="questDetailDesc"></div>
      <div class="quest-detail-objectives">
        <div class="quest-detail-objectives-title">任务目标</div>
        <div id="questDetailObjectives"></div>
      </div>
      <div class="quest-detail-rewards">
        <div class="quest-detail-rewards-title">任务奖励</div>
        <div class="quest-detail-reward-list" id="questDetailRewards"></div>
      </div>
      <div class="quest-detail-actions" id="questDetailActions"></div>
    `;
    document.body.appendChild(panel);
    this.detailPanel = panel;
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
      <div class="quest-tracker-header">
        <div class="quest-tracker-title">任务追踪</div>
        <div class="quest-tracker-toggle" id="questTrackerToggle">[隐藏]</div>
      </div>
      <div id="questTrackerList"></div>
      <div class="quest-tracker-expand" id="questTrackerExpand">查看全部任务 →</div>
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
        this.currentFilter = 'all';
        this.updateFilterButtons();
        this.renderQuestList();
      });
    });

    // 类型筛选按钮
    this.currentFilter = 'all';
    const filterBtns = document.querySelectorAll('.quest-filter-btn');
    filterBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        filterBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        this.currentFilter = btn.dataset.filter;
        this.renderQuestList();
      });
    });

    // 详情面板关闭按钮
    const detailCloseBtn = document.getElementById('questDetailCloseBtn');
    if (detailCloseBtn) {
      detailCloseBtn.addEventListener('click', () => this.hideDetailPanel());
    }

    // 追踪器切换按钮
    const trackerToggle = document.getElementById('questTrackerToggle');
    if (trackerToggle) {
      trackerToggle.addEventListener('click', (e) => {
        e.stopPropagation();
        this.toggleTracker();
      });
    }

    // 追踪器展开按钮
    const trackerExpand = document.getElementById('questTrackerExpand');
    if (trackerExpand) {
      trackerExpand.addEventListener('click', (e) => {
        e.stopPropagation();
        this.show();
      });
    }

    // 批量接取按钮
    const batchAcceptBtn = document.getElementById('batchAcceptBtn');
    if (batchAcceptBtn) {
      batchAcceptBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.batchAcceptQuests();
      });
    }

    // 追踪器点击打开任务面板
    if (this.trackerEl) {
      this.trackerEl.addEventListener('click', () => {
        this.show();
      });
    }

    // 键盘快捷键 L 打开任务面板
    this.keydownHandler = (e) => {
      if (e.key === 'l' || e.key === 'L') {
        if (this.game.state === 'playing') {
          this.toggle();
        }
      }
    };
    document.addEventListener('keydown', this.keydownHandler);
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
   * 批量接取任务
   */
  batchAcceptQuests() {
    const roleId = this.game.player?.id;
    if (!roleId) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('玩家未登录', 'error', 2000);
      }
      return;
    }

    // 获取所有可接取的任务ID
    const questIds = this.availableQuests.map(q => q.id);
    if (questIds.length === 0) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('没有可接取的任务', 'info', 2000);
      }
      return;
    }

    // 调用批量接取API
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_BATCH_ACCEPT || 6099, {
        role_id: roleId,
        quest_ids: questIds
      });
    }

    // 乐观更新：移动所有可接任务到进行中
    const acceptedQuests = this.availableQuests.splice(0, this.availableQuests.length);
    acceptedQuests.forEach(quest => {
      quest.status = 1;
      quest.progress = 0;
      this.activeQuests.push(quest);
    });

    this.renderQuestList();
    this.updateTracker();
    this.updateTabCounts();

    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast(`已接取${acceptedQuests.length}个任务`, 'success', 2000);
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
   * 分享任务到剪贴板
   */
  shareQuest(questId) {
    // 查找任务
    let quest = null;
    let questSource = '';
    const allQuests = [...this.activeQuests, ...this.completedQuests, ...this.availableQuests, ...this.finishedQuests];
    quest = allQuests.find(q => q.id === questId);
    if (this.activeQuests.find(q => q.id === questId)) questSource = '进行中';
    else if (this.completedQuests.find(q => q.id === questId)) questSource = '已完成(待领奖)';
    else if (this.availableQuests.find(q => q.id === questId)) questSource = '可接取';
    else if (this.finishedQuests.find(q => q.id === questId)) questSource = '已完成';

    if (!quest) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('未找到任务', 'error', 2000);
      }
      return;
    }

    const typeInfo = this.questTypes[quest.type] || { name: '未知' };
    const targetInfo = this.targetTypes[quest.target_type] || { name: quest.target_name || '任务目标' };

    // 构建分享文本
    const shareText = `
【${typeInfo.name}任务】${quest.name}
类型：${typeInfo.name}
进度：${quest.progress || 0}/${quest.target_count || 1}
描述：${quest.description || '无'}
奖励：${quest.reward_exp ? `经验+${quest.reward_exp}` : ''} ${quest.reward_gold ? `金币+${quest.reward_gold}` : ''} ${quest.reward_honor ? `声望+${quest.reward_honor}` : ''}
状态：${questSource}
    `.trim();

    // 复制到剪贴板
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(shareText).then(() => {
        if (this.game.uiManager?.toast) {
          this.game.uiManager.toast('任务信息已复制到剪贴板', 'success', 2000);
        }
      }).catch(() => {
        // 降级方案：使用传统方法
        this.fallbackCopyText(shareText);
      });
    } else {
      this.fallbackCopyText(shareText);
    }
  }

  /**
   * 降级版复制方法
   */
  fallbackCopyText(text) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.appendChild(textarea);
    textarea.select();
    try {
      document.execCommand('copy');
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('任务信息已复制到剪贴板', 'success', 2000);
      }
    } catch (err) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('复制失败，请手动复制', 'error', 2000);
      }
    }
    document.body.removeChild(textarea);
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
          // 显示任务完成动画
          this.showQuestCompleteEffect(quest.name);
          // 显示奖励动画
          this.showRewardEffect({
            exp: quest.reward_exp || 0,
            gold: quest.reward_gold || 0,
            honor: quest.reward_honor || 0,
            items: quest.reward_item_id ? [{ name: '奖励物品', count: quest.reward_item_count || 1 }] : []
          });
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

    // 应用类型筛选
    if (this.currentFilter !== 'all') {
      quests = quests.filter(q => q.type === parseInt(this.currentFilter));
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
    listEl.querySelectorAll('.quest-btn-share').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.shareQuest(parseInt(btn.dataset.id));
      });
    });

    // 绑定任务项点击事件（查看详情）
    listEl.querySelectorAll('.quest-item').forEach(item => {
      item.addEventListener('click', (e) => {
        if (!e.target.classList.contains('quest-btn')) {
          const questId = parseInt(item.dataset.questId);
          if (questId) {
            this.showDetailPanel(questId);
          }
        }
      });
    });
  }

  /**
   * 更新筛选按钮状态
   */
  updateFilterButtons() {
    const filterBtns = document.querySelectorAll('.quest-filter-btn');
    filterBtns.forEach(btn => {
      btn.classList.toggle('active', btn.dataset.filter === this.currentFilter);
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
    if (quest.reward_honor) rewards.push(`<span class="quest-reward-item">声望 +${quest.reward_honor}</span>`);
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
    // 添加分享按钮（所有状态都可以分享）
    actions += `<button class="quest-btn-share" data-id="${quest.id}">分享</button>`;

    return `
      <div class="quest-item ${isCompleted ? 'completed' : ''}" data-quest-id="${quest.id}">
        <div class="quest-item-header">
          <div class="quest-item-name">${typeInfo.icon} ${quest.name}</div>
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

    // 如果追踪器被隐藏
    if (!this.trackEnabled) {
      this.trackerEl.style.display = 'none';
      return;
    }

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

    // 更新切换按钮文字
    const toggleBtn = document.getElementById('questTrackerToggle');
    if (toggleBtn) {
      toggleBtn.textContent = this.trackEnabled ? '[隐藏]' : '[显示]';
    }

    // 只显示前3个任务
    const showQuests = this.activeQuests.slice(0, 3);
    listEl.innerHTML = showQuests.map(quest => {
      const isCompleted = quest.status === 2;
      const objectives = this.getQuestObjectives(quest);
      return `
        <div class="quest-tracker-item ${isCompleted ? 'completed' : ''}" data-quest-id="${quest.id}">
          <div class="quest-tracker-name">${quest.name}</div>
          <div class="quest-tracker-objectives">
            ${objectives.map(obj => `
              <div class="quest-tracker-objective ${obj.isCompleted ? 'completed' : ''}">
                <span class="quest-tracker-objective-icon">${obj.icon}</span>
                ${obj.name}: ${obj.current}/${obj.total}
              </div>
            `).join('')}
          </div>
        </div>
      `;
    }).join('');

    // 绑定追踪器任务项点击事件
    listEl.querySelectorAll('.quest-tracker-item').forEach(item => {
      item.addEventListener('click', () => {
        const questId = parseInt(item.dataset.questId);
        const quest = this.activeQuests.find(q => q.id === questId);
        if (quest) {
          this.onTrackerItemClick(quest);
        }
      });
    });
  }

  /**
   * 获取任务目标列表
   */
  getQuestObjectives(quest) {
    const objectives = [];
    
    // 支持多目标任务
    if (quest.objectives && Array.isArray(quest.objectives)) {
      quest.objectives.forEach(obj => {
        const targetInfo = this.targetTypes[obj.target_type] || { name: '目标', icon: '◆', color: '#999' };
        objectives.push({
          id: obj.id,
          name: obj.target_name || targetInfo.name,
          icon: targetInfo.icon,
          targetType: obj.target_type,
          targetId: obj.target_id,
          current: obj.progress || 0,
          total: obj.target_count || 1,
          isCompleted: (obj.progress || 0) >= (obj.target_count || 1)
        });
      });
    } else {
      // 单目标任务
      const targetInfo = this.targetTypes[quest.target_type] || { name: '目标', icon: '◆', color: '#999' };
      objectives.push({
        id: 1,
        name: quest.target_name || targetInfo.name,
        icon: targetInfo.icon,
        targetType: quest.target_type,
        targetId: quest.target_id,
        current: quest.progress || 0,
        total: quest.target_count || 1,
        isCompleted: (quest.progress || 0) >= (quest.target_count || 1)
      });
    }
    
    return objectives;
  }

  /**
   * 显示任务完成动画
   * @param {string} questName - 任务名称
   */
  showQuestCompleteEffect(questName) {
    // 创建完成效果容器
    const effect = document.createElement('div');
    effect.className = 'quest-complete-effect';
    effect.innerHTML = `<div class="quest-complete-text">✦ ${questName} 完成！✦</div>`;
    document.body.appendChild(effect);

    // 创建粒子效果
    this.createParticles(window.innerWidth / 2, window.innerHeight / 2, '#4ade80', 30);

    // 2秒后移除效果
    setTimeout(() => {
      effect.remove();
    }, 2000);
  }

  /**
   * 显示奖励发放动画
   * @param {Object} rewards - 奖励信息 {exp, gold, honor, items}
   */
  showRewardEffect(rewards) {
    const container = document.createElement('div');
    container.className = 'quest-reward-effect';

    let html = '';
    if (rewards.exp > 0) {
      html += `<div class="quest-reward-item-anim quest-reward-exp"><span class="reward-icon">✨</span> 经验 +${this.formatNumber(rewards.exp)}</div>`;
    }
    if (rewards.gold > 0) {
      html += `<div class="quest-reward-item-anim quest-reward-gold"><span class="reward-icon">💰</span> 金币 +${this.formatNumber(rewards.gold)}</div>`;
    }
    if (rewards.honor > 0) {
      html += `<div class="quest-reward-item-anim quest-reward-honor"><span class="reward-icon">🏅</span> 声望 +${this.formatNumber(rewards.honor)}</div>`;
    }
    if (rewards.items && rewards.items.length > 0) {
      rewards.items.forEach(item => {
        html += `<div class="quest-reward-item-anim" style="border-color: #60a5fa; color: #60a5fa;"><span class="reward-icon">📦</span> ${item.name} x${item.count}</div>`;
      });
    }

    container.innerHTML = html;
    document.body.appendChild(container);

    // 创建奖励粒子效果
    this.createParticles(window.innerWidth / 2, window.innerHeight - 150, '#fbbf24', 20);

    // 3秒后移除效果
    setTimeout(() => {
      container.remove();
    }, 3000);
  }

  /**
   * 创建粒子效果
   * @param {number} x - 中心X坐标
   * @param {number} y - 中心Y坐标
   * @param {string} color - 粒子颜色
   * @param {number} count - 粒子数量
   */
  createParticles(x, y, color, count) {
    const container = document.createElement('div');
    container.className = 'quest-particles';
    document.body.appendChild(container);

    for (let i = 0; i < count; i++) {
      const particle = document.createElement('div');
      particle.className = 'quest-particle';
      particle.style.backgroundColor = color;
      particle.style.left = `${x + (Math.random() - 0.5) * 100}px`;
      particle.style.top = `${y + (Math.random() - 0.5) * 100}px`;
      particle.style.animationDelay = `${Math.random() * 0.3}s`;
      container.appendChild(particle);
    }

    // 1.5秒后移除粒子容器
    setTimeout(() => {
      container.remove();
    }, 1500);
  }

  /**
   * 格式化数字（添加千分位）
   */
  formatNumber(num) {
    if (num >= 10000) {
      return (num / 10000).toFixed(1) + '万';
    }
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  }

  /**
   * 高亮任务指引目标
   * @param {Object} guide - 指引信息
   */
  highlightGuideTarget(guide) {
    // 移除现有高亮
    this.removeGuideHighlight();

    // 创建高亮框
    const highlight = document.createElement('div');
    highlight.id = 'questGuideHighlight';
    highlight.className = 'quest-guide-highlight';

    // 根据目标类型计算高亮位置
    let targetX, targetY, width, height;

    if (guide.target_type === 3) { // NPC对话
      // NPC位置
      const npcEl = document.querySelector(`[data-npc-id="${guide.target_id}"]`);
      if (npcEl) {
        const rect = npcEl.getBoundingClientRect();
        targetX = rect.left - 20;
        targetY = rect.top - 20;
        width = rect.width + 40;
        height = rect.height + 40;
      }
    } else if (guide.target_type === 1) { // 怪物
      // 怪物位置
      const monsterEl = document.querySelector(`[data-monster-id="${guide.target_id}"]`);
      if (monsterEl) {
        const rect = monsterEl.getBoundingClientRect();
        targetX = rect.left - 20;
        targetY = rect.top - 20;
        width = rect.width + 40;
        height = rect.height + 40;
      }
    } else {
      // 默认中央显示
      targetX = window.innerWidth / 2 - 100;
      targetY = window.innerHeight / 2 - 50;
      width = 200;
      height = 100;
    }

    highlight.style.left = `${targetX}px`;
    highlight.style.top = `${targetY}px`;
    highlight.style.width = `${width}px`;
    highlight.style.height = `${height}px`;

    document.body.appendChild(highlight);

    // 10秒后自动移除
    setTimeout(() => {
      this.removeGuideHighlight();
    }, 10000);
  }

  /**
   * 移除任务指引高亮
   */
  removeGuideHighlight() {
    const highlight = document.getElementById('questGuideHighlight');
    if (highlight) {
      highlight.remove();
    }
  }

  /**
   * 点击追踪器任务项（触发自动寻路）
   */
  onTrackerItemClick(quest) {
    // 获取任务指引信息
    const guide = this.getQuestGuide(quest);
    if (guide) {
      // 高亮目标
      this.highlightGuideTarget(guide);

      // 如果需要切换地图，先传送
      if (guide.map_id && guide.map_id !== this.game.player?.mapId) {
        this.game.switchMap(guide.map_id);
      }

      // 触发自动寻路
      if (this.game.pathfinding) {
        this.game.pathfinding.startPathfind(guide.x, guide.y);
      }

      // 显示指引信息
      this.showGuideMessage(guide);
    }
  }

  /**
   * 获取任务指引信息
   */
  getQuestGuide(quest) {
    // 根据任务类型构建指引
    const objectives = this.getQuestObjectives(quest);
    const incompleteObj = objectives.find(obj => !obj.isCompleted);

    if (!incompleteObj) return null;

    const typeInfo = this.targetTypes[incompleteObj.targetType] || { name: '目标', icon: '◆' };

    return {
      quest_id: quest.id,
      quest_name: quest.name,
      target_type: incompleteObj.targetType,
      target_id: incompleteObj.targetId,
      target_name: incompleteObj.name,
      map_id: quest.map_id || 0,
      map_name: quest.map_name || '',
      x: quest.x || 0,
      y: quest.y || 0,
      description: `${typeInfo.icon} ${incompleteObj.name}: ${incompleteObj.current}/${incompleteObj.total}`
    };
  }

  /**
   * 显示指引消息
   */
  showGuideMessage(guide) {
    // 创建临时消息提示
    const msg = document.createElement('div');
    msg.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: rgba(0, 0, 0, 0.9);
      border: 2px solid #fbbf24;
      border-radius: 10px;
      padding: 15px 25px;
      color: #fff;
      font-size: 14px;
      z-index: 10000;
      text-align: center;
      animation: fadeInOut 3s ease-out forwards;
    `;
    msg.innerHTML = `
      <div style="color: #fbbf24; font-weight: bold; margin-bottom: 8px;">任务指引</div>
      <div>${guide.description}</div>
      ${guide.map_name ? `<div style="color: #999; font-size: 12px; margin-top: 5px;">📍 ${guide.map_name}</div>` : ''}
    `;
    document.body.appendChild(msg);

    // 添加动画样式
    const style = document.createElement('style');
    style.textContent = `
      @keyframes fadeInOut {
        0% { opacity: 0; transform: translate(-50%, -50%) scale(0.8); }
        15% { opacity: 1; transform: translate(-50%, -50%) scale(1); }
        85% { opacity: 1; transform: translate(-50%, -50%) scale(1); }
        100% { opacity: 0; transform: translate(-50%, -50%) scale(0.8); }
      }
    `;
    document.head.appendChild(style);

    setTimeout(() => {
      msg.remove();
      style.remove();
    }, 3000);
  }

  /**
   * 显示任务详情面板
   */
  showDetailPanel(questId) {
    // 查找任务（在所有列表中查找）
    let quest = this.activeQuests.find(q => q.id === questId);
    if (!quest) quest = this.availableQuests.find(q => q.id === questId);
    if (!quest) quest = this.completedQuests.find(q => q.id === questId);
    if (!quest) quest = this.finishedQuests.find(q => q.id === questId);
    
    if (!quest) return;
    
    this.selectedQuest = quest;
    
    const typeInfo = this.questTypes[quest.type] || { name: '未知', color: '#999', icon: '◆' };
    const isCompleted = quest.status === 2;
    const isFinished = quest.status === 3;
    
    // 更新标题
    document.getElementById('questDetailTitle').textContent = `${typeInfo.icon} ${quest.name}`;
    
    // 更新类型标签
    document.getElementById('questDetailType').textContent = typeInfo.name;
    document.getElementById('questDetailType').style.background = typeInfo.color;
    
    // 更新描述
    document.getElementById('questDetailDesc').textContent = quest.description || '暂无描述';
    
    // 更新目标列表
    const objectives = this.getQuestObjectives(quest);
    document.getElementById('questDetailObjectives').innerHTML = objectives.map(obj => {
      const progressPercent = Math.min(100, (obj.current / obj.total) * 100);
      return `
        <div class="quest-detail-objective">
          <div class="quest-detail-objective-icon" style="color: ${obj.isCompleted ? '#4ade80' : '#fbbf24'}">${obj.icon}</div>
          <div class="quest-detail-objective-info">
            <div class="quest-detail-objective-name">${obj.name}</div>
            <div class="quest-detail-objective-progress">${obj.current}/${obj.total}</div>
            <div class="quest-detail-objective-bar">
              <div class="quest-detail-objective-fill" style="width: ${progressPercent}%;"></div>
            </div>
          </div>
        </div>
      `;
    }).join('');
    
    // 更新奖励列表
    const rewards = [];
    if (quest.reward_exp) rewards.push(`<div class="quest-detail-reward-item"><span class="quest-detail-reward-icon">⭐</span>经验 +${quest.reward_exp}</div>`);
    if (quest.reward_gold) rewards.push(`<div class="quest-detail-reward-item"><span class="quest-detail-reward-icon">💰</span>金币 +${quest.reward_gold}</div>`);
    if (quest.reward_item_id) rewards.push(`<div class="quest-detail-reward-item"><span class="quest-detail-reward-icon">📦</span>物品 x${quest.reward_item_count || 1}</div>`);
    if (quest.reward_honor) rewards.push(`<div class="quest-detail-reward-item"><span class="quest-detail-reward-icon">🏆</span>声望 +${quest.reward_honor}</div>`);
    document.getElementById('questDetailRewards').innerHTML = rewards.join('');
    
    // 更新操作按钮
    let actions = '';
    if (this.currentTab === 'available') {
      actions = '<button class="quest-detail-btn quest-detail-btn-accept" id="detailAcceptBtn">接取任务</button>';
    } else if (this.currentTab === 'active') {
      if (isCompleted) {
        actions = '<button class="quest-detail-btn quest-detail-btn-complete" id="detailCompleteBtn">领取奖励</button>';
      } else {
        actions = '<button class="quest-detail-btn quest-detail-btn-abandon" id="detailAbandonBtn">放弃任务</button>';
      }
    } else if (this.currentTab === 'finished') {
      actions = '<div style="text-align:center; color:#4ade80; padding: 10px;">✓ 任务已完成</div>';
    }
    actions += '<button class="quest-detail-btn quest-detail-btn-share" id="detailShareBtn">分享</button>';
    actions += '<button class="quest-detail-btn quest-detail-btn-close" id="detailCloseBtn">关闭</button>';
    document.getElementById('questDetailActions').innerHTML = actions;
    
    // 显示面板
    this.detailPanel.classList.add('show');
    
    // 绑定按钮事件
    setTimeout(() => {
      const acceptBtn = document.getElementById('detailAcceptBtn');
      const completeBtn = document.getElementById('detailCompleteBtn');
      const abandonBtn = document.getElementById('detailAbandonBtn');
      const shareBtn = document.getElementById('detailShareBtn');
      const closeBtn = document.getElementById('detailCloseBtn');
      
      if (acceptBtn) acceptBtn.addEventListener('click', () => { this.acceptQuest(questId); this.hideDetailPanel(); });
      if (completeBtn) completeBtn.addEventListener('click', () => { this.completeQuest(questId); this.hideDetailPanel(); });
      if (abandonBtn) abandonBtn.addEventListener('click', () => { this.abandonQuest(questId); this.hideDetailPanel(); });
      if (shareBtn) shareBtn.addEventListener('click', () => { this.shareQuest(questId); });
      if (closeBtn) closeBtn.addEventListener('click', () => this.hideDetailPanel());
    }, 50);
  }

  /**
   * 隐藏任务详情面板
   */
  hideDetailPanel() {
    if (this.detailPanel) {
      this.detailPanel.classList.remove('show');
    }
    this.selectedQuest = null;
  }

  /**
   * 切换追踪器显示
   */
  toggleTracker() {
    this.trackEnabled = !this.trackEnabled;
    this.updateTracker();
  }

  /**
   * 添加任务日志记录
   */
  addQuestLog(message) {
    this.questLog.unshift({
      id: Date.now(),
      message,
      time: new Date().toLocaleTimeString()
    });
    // 保留最近20条日志
    if (this.questLog.length > 20) {
      this.questLog.pop();
    }
  }

  /**
   * 从服务器加载任务列表
   */
  loadQuestListFromServer() {
    const roleId = this.game.player?.id;
    if (!roleId) {
      console.log('[Quest] 玩家未登录，加载测试数据');
      this.loadTestData();
      return;
    }

    // 通过WebSocket发送任务列表请求
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_QUEST_LIST || 6001, {
        role_id: roleId
      });
    } else {
      console.log('[Quest] WebSocket未连接，加载测试数据');
      this.loadTestData();
    }
  }

  /**
   * 加载测试数据（开发/无服务器时使用）
   */
  loadTestData() {
    // 创建测试任务数据
    this.availableQuests = [
      {
        id: 1,
        name: '初入江湖',
        type: 1,
        description: '前往新手村找村长对话，开启你的江湖之旅。',
        target_type: 3,
        target_id: 1001,
        target_name: '村长',
        target_count: 1,
        progress: 0,
        status: 0,
        reward_exp: 100,
        reward_gold: 50,
        objectives: [
          { id: 1, target_type: 3, target_id: 1001, target_name: '村长', progress: 0, target_count: 1 }
        ]
      },
      {
        id: 2,
        name: '除暴安良',
        type: 2,
        description: '新手村附近出现了一些山贼，村民们深受其害。请你去消灭他们。',
        target_type: 1,
        target_id: 2001,
        target_name: '山贼',
        target_count: 10,
        progress: 0,
        status: 0,
        reward_exp: 200,
        reward_gold: 100,
        reward_item_id: 1,
        reward_item_count: 5,
        objectives: [
          { id: 1, target_type: 1, target_id: 2001, target_name: '山贼', progress: 0, target_count: 10 }
        ]
      },
      {
        id: 3,
        name: '每日修行',
        type: 3,
        description: '每日修行，提升自身实力。完成3次战斗即可完成任务。',
        target_type: 1,
        target_id: 0,
        target_name: '任意怪物',
        target_count: 3,
        progress: 0,
        status: 0,
        reward_exp: 150,
        reward_gold: 80,
        objectives: [
          { id: 1, target_type: 1, target_id: 0, target_name: '任意怪物', progress: 0, target_count: 3 }
        ]
      },
      {
        id: 4,
        name: '采集草药',
        type: 2,
        description: '药师需要一些草药来制作药品，请帮忙采集5株草药。',
        target_type: 2,
        target_id: 3001,
        target_name: '草药',
        target_count: 5,
        progress: 0,
        status: 0,
        reward_exp: 120,
        reward_gold: 60,
        objectives: [
          { id: 1, target_type: 2, target_id: 3001, target_name: '草药', progress: 0, target_count: 5 }
        ]
      }
    ];

    this.activeQuests = [
      {
        id: 5,
        name: '主线：江湖危机',
        type: 1,
        description: '江湖中出现了一股神秘势力，四处作恶。武林盟主希望你能调查此事。首先前往洛阳城，找到盟主特使了解详情。',
        target_type: 3,
        target_id: 1002,
        target_name: '盟主特使',
        target_count: 1,
        progress: 0,
        status: 1,
        reward_exp: 500,
        reward_gold: 300,
        reward_honor: 50,
        objectives: [
          { id: 1, target_type: 3, target_id: 1002, target_name: '盟主特使', progress: 0, target_count: 1 }
        ]
      },
      {
        id: 6,
        name: '围剿山匪',
        type: 2,
        description: '黑风寨的山匪越来越猖獗，官府希望你能协助剿灭他们。需要击杀5名山匪头领和20名山匪小喽啰。',
        target_type: 1,
        target_id: 2002,
        target_name: '山匪头领',
        target_count: 5,
        progress: 3,
        status: 1,
        reward_exp: 800,
        reward_gold: 500,
        objectives: [
          { id: 1, target_type: 1, target_id: 2002, target_name: '山匪头领', progress: 3, target_count: 5 },
          { id: 2, target_type: 1, target_id: 2003, target_name: '山匪小喽啰', progress: 15, target_count: 20 }
        ]
      }
    ];

    this.completedQuests = [
      {
        id: 7,
        name: '新手训练',
        type: 1,
        description: '完成基础战斗训练，学习基本的战斗技巧。',
        target_type: 4,
        target_id: 0,
        target_name: '完成训练',
        target_count: 1,
        progress: 1,
        status: 2,
        reward_exp: 50,
        reward_gold: 20,
        objectives: [
          { id: 1, target_type: 4, target_id: 0, target_name: '完成训练', progress: 1, target_count: 1 }
        ]
      }
    ];

    this.finishedQuests = [
      {
        id: 8,
        name: '初识江湖',
        type: 1,
        description: '完成你的第一次战斗，证明你有资格闯荡江湖。',
        target_type: 1,
        target_id: 2000,
        target_name: '野猪',
        target_count: 5,
        progress: 5,
        status: 3,
        reward_exp: 30,
        reward_gold: 10,
        objectives: [
          { id: 1, target_type: 1, target_id: 2000, target_name: '野猪', progress: 5, target_count: 5 }
        ]
      }
    ];

    // 更新UI
    this.updateTabCounts();
    this.updateTracker();
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
      
      // 检查多目标任务
      if (quest.objectives && Array.isArray(quest.objectives)) {
        quest.objectives.forEach(obj => {
          if (obj.target_type === 1 && obj.target_id === monsterId) {
            if (window.GameWS && window.GameWS.isConnected) {
              window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
                quest_id: quest.id,
                target_type: 1,
                target_id: monsterId,
                objective_id: obj.id
              });
            }
          }
        });
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
      
      // 检查多目标任务
      if (quest.objectives && Array.isArray(quest.objectives)) {
        quest.objectives.forEach(obj => {
          if (obj.target_type === 2 && obj.target_id === itemId) {
            if (window.GameWS && window.GameWS.isConnected) {
              window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
                quest_id: quest.id,
                target_type: 2,
                target_id: itemId,
                objective_id: obj.id
              });
            }
          }
        });
      }
    });
  }

  /**
   * 通知对话完成（用于对话类任务进度）
   */
  onDialogComplete(npcId) {
    this.activeQuests.forEach(quest => {
      if (quest.target_type === 3 && quest.target_id === npcId) {
        if (window.GameWS && window.GameWS.isConnected) {
          window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
            quest_id: quest.id,
            target_type: 3,
            target_id: npcId
          });
        }
      }
      
      // 检查多目标任务
      if (quest.objectives && Array.isArray(quest.objectives)) {
        quest.objectives.forEach(obj => {
          if (obj.target_type === 3 && obj.target_id === npcId) {
            if (window.GameWS && window.GameWS.isConnected) {
              window.GameWS.send(window.Protocol.CMD_QUEST_PROGRESS || 6005, {
                quest_id: quest.id,
                target_type: 3,
                target_id: npcId,
                objective_id: obj.id
              });
            }
          }
        });
      }
    });
  }

  /**
   * 获取任务进度百分比
   */
  getQuestProgressPercent(quest) {
    const objectives = this.getQuestObjectives(quest);
    if (objectives.length === 0) return 0;
    
    const totalProgress = objectives.reduce((sum, obj) => sum + obj.current, 0);
    const totalTarget = objectives.reduce((sum, obj) => sum + obj.total, 0);
    
    return Math.min(100, Math.round((totalProgress / totalTarget) * 100));
  }
}

// 导出到全局
window.QuestSystem = QuestSystem;
