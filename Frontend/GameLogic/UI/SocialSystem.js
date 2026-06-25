/**
 * 社交系统
 * 管理好友列表、组队等社交功能
 */
class SocialSystem {
  constructor(game) {
    this.game = game;
    this.initialized = false;
    this.isOpen = false;
    this.currentTab = 'friends'; // friends, team
    this.currentFriendTab = 'online'; // online, offline, requests
    
    // 好友数据
    this.friends = {
      online: [],
      offline: [],
      requests: []
    };
    
    // 队伍数据
    this.team = null;
    this.teamMembers = [];
    this.availableTeams = [];
    
    // UI元素
    this.container = null;
    this.friendPanel = null;
    this.teamPanel = null;
    
    // 初始化
    this.init();
  }

  /**
   * 初始化
   */
  init() {
    if (this.initialized) return;
    this.createStyles();
    this.createPanel();
    this.bindEvents();
    this.loadTestData();
    this.initialized = true;
    console.log('社交系统初始化完成');
  }

  /**
   * 创建样式
   */
  createStyles() {
    if (document.getElementById('social-styles')) return;
    
    const style = document.createElement('style');
    style.id = 'social-styles';
    style.textContent = `
      /* 社交面板 */
      .social-panel {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%) scale(0.9);
        width: 800px;
        height: 550px;
        background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
        border: 2px solid #e94560;
        border-radius: 12px;
        box-shadow: 0 0 30px rgba(233, 69, 96, 0.3);
        display: none;
        z-index: 1000;
        overflow: hidden;
        opacity: 0;
        transition: all 0.3s ease;
      }
      
      .social-panel.show {
        display: block;
        transform: translate(-50%, -50%) scale(1);
        opacity: 1;
      }
      
      /* 头部 */
      .social-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 15px 20px;
        background: rgba(233, 69, 96, 0.2);
        border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      }
      
      .social-title {
        color: #e94560;
        font-size: 20px;
        font-weight: bold;
      }
      
      .social-close-btn {
        width: 30px;
        height: 30px;
        border: none;
        background: rgba(255, 255, 255, 0.1);
        color: #fff;
        font-size: 20px;
        cursor: pointer;
        border-radius: 50%;
        transition: all 0.2s;
      }
      
      .social-close-btn:hover {
        background: rgba(233, 69, 96, 0.5);
        transform: scale(1.1);
      }
      
      /* 标签切换 */
      .social-tabs {
        display: flex;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);
      }
      
      .social-tab {
        flex: 1;
        padding: 12px;
        text-align: center;
        color: #999;
        cursor: pointer;
        transition: all 0.2s;
        font-size: 14px;
      }
      
      .social-tab.active {
        color: #e94560;
        background: rgba(233, 69, 96, 0.1);
        border-bottom: 2px solid #e94560;
      }
      
      .social-tab:hover {
        color: #fff;
      }
      
      /* 内容区域 */
      .social-content {
        height: calc(100% - 130px);
        padding: 10px;
        overflow: hidden;
      }
      
      /* 好友面板 */
      .friends-panel {
        height: 100%;
        display: flex;
        flex-direction: column;
      }
      
      .friends-search {
        padding: 8px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: 6px;
        color: #fff;
        font-size: 13px;
        margin-bottom: 10px;
      }
      
      .friends-search::placeholder {
        color: #666;
      }
      
      .friends-subtabs {
        display: flex;
        margin-bottom: 10px;
        gap: 5px;
      }
      
      .friends-subtab {
        padding: 5px 12px;
        background: rgba(255, 255, 255, 0.1);
        border-radius: 15px;
        color: #999;
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .friends-subtab.active {
        background: #e94560;
        color: #fff;
      }
      
      .friends-list {
        flex: 1;
        overflow-y: auto;
        padding-right: 5px;
      }
      
      .friends-list::-webkit-scrollbar {
        width: 4px;
      }
      
      .friends-list::-webkit-scrollbar-track {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 2px;
      }
      
      .friends-list::-webkit-scrollbar-thumb {
        background: #e94560;
        border-radius: 2px;
      }
      
      .friend-item {
        display: flex;
        align-items: center;
        padding: 10px;
        background: rgba(255, 255, 255, 0.05);
        border-radius: 8px;
        margin-bottom: 8px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .friend-item:hover {
        background: rgba(255, 255, 255, 0.1);
      }
      
      .friend-avatar {
        width: 45px;
        height: 45px;
        border-radius: 50%;
        background: linear-gradient(135deg, #e94560 0%, #ff6b6b 100%);
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
        font-size: 18px;
        font-weight: bold;
        margin-right: 12px;
        border: 2px solid rgba(233, 69, 96, 0.5);
      }
      
      .friend-info {
        flex: 1;
        min-width: 0;
      }
      
      .friend-name {
        color: #fff;
        font-size: 14px;
        font-weight: 500;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      
      .friend-status {
        color: #666;
        font-size: 12px;
        margin-top: 2px;
      }
      
      .friend-status.online {
        color: #4ade80;
      }
      
      .friend-status.offline {
        color: #666;
      }
      
      .friend-level {
        display: inline-block;
        background: rgba(233, 69, 96, 0.3);
        padding: 2px 6px;
        border-radius: 10px;
        font-size: 11px;
        color: #e94560;
        margin-left: 5px;
      }
      
      .friend-actions {
        display: flex;
        gap: 8px;
      }
      
      .friend-action-btn {
        padding: 6px 12px;
        border: none;
        border-radius: 4px;
        font-size: 12px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .friend-action-btn.chat {
        background: #3b82f6;
        color: #fff;
      }
      
      .friend-action-btn.invite {
        background: #22c55e;
        color: #fff;
      }
      
      .friend-action-btn.accept {
        background: #22c55e;
        color: #fff;
      }
      
      .friend-action-btn.reject {
        background: #ef4444;
        color: #fff;
      }
      
      .friend-action-btn.remove {
        background: rgba(233, 69, 96, 0.3);
        color: #e94560;
      }
      
      .friend-action-btn:hover {
        transform: scale(1.05);
      }
      
      /* 组队面板 */
      .team-panel {
        height: 100%;
        display: flex;
        flex-direction: column;
      }
      
      .team-info {
        background: rgba(233, 69, 96, 0.1);
        border-radius: 8px;
        padding: 15px;
        margin-bottom: 10px;
      }
      
      .team-info-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;
      }
      
      .team-name {
        color: #e94560;
        font-size: 16px;
        font-weight: bold;
      }
      
      .team-leader-badge {
        background: #e94560;
        padding: 2px 8px;
        border-radius: 10px;
        font-size: 11px;
        color: #fff;
      }
      
      .team-members-list {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 10px;
      }
      
      .team-member-item {
        display: flex;
        align-items: center;
        padding: 10px;
        background: rgba(255, 255, 255, 0.05);
        border-radius: 8px;
      }
      
      .team-member-avatar {
        width: 40px;
        height: 40px;
        border-radius: 50%;
        background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
        font-size: 16px;
        font-weight: bold;
        margin-right: 10px;
        border: 2px solid rgba(59, 130, 246, 0.5);
      }
      
      .team-member-info {
        flex: 1;
      }
      
      .team-member-name {
        color: #fff;
        font-size: 13px;
        font-weight: 500;
      }
      
      .team-member-role {
        color: #666;
        font-size: 11px;
        margin-top: 2px;
      }
      
      .team-member-role.leader {
        color: #e94560;
      }
      
      .team-member-level {
        color: #4ade80;
        font-size: 11px;
      }
      
      .team-actions {
        display: flex;
        gap: 10px;
        margin-top: 10px;
      }
      
      .team-action-btn {
        flex: 1;
        padding: 10px;
        border: none;
        border-radius: 6px;
        font-size: 13px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .team-action-btn.create {
        background: linear-gradient(135deg, #e94560 0%, #ff6b6b 100%);
        color: #fff;
      }
      
      .team-action-btn.join {
        background: #3b82f6;
        color: #fff;
      }
      
      .team-action-btn.leave {
        background: rgba(233, 69, 96, 0.3);
        color: #e94560;
      }
      
      .team-action-btn.kick {
        background: #ef4444;
        color: #fff;
      }
      
      .team-action-btn:hover {
        transform: scale(1.02);
        box-shadow: 0 0 10px rgba(233, 69, 96, 0.3);
      }
      
      /* 可用队伍列表 */
      .available-teams-title {
        color: #fff;
        font-size: 14px;
        margin-bottom: 10px;
        font-weight: 500;
      }
      
      .available-teams-list {
        flex: 1;
        overflow-y: auto;
        padding-right: 5px;
      }
      
      .available-team-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px;
        background: rgba(255, 255, 255, 0.05);
        border-radius: 8px;
        margin-bottom: 8px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .available-team-item:hover {
        background: rgba(255, 255, 255, 0.1);
      }
      
      .available-team-info {
        flex: 1;
      }
      
      .available-team-name {
        color: #fff;
        font-size: 14px;
        font-weight: 500;
      }
      
      .available-team-count {
        color: #666;
        font-size: 12px;
        margin-top: 2px;
      }
      
      /* 添加好友弹窗 */
      .add-friend-modal {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%) scale(0.9);
        width: 400px;
        background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
        border: 2px solid #e94560;
        border-radius: 12px;
        box-shadow: 0 0 30px rgba(233, 69, 96, 0.3);
        padding: 20px;
        z-index: 1010;
        display: none;
        opacity: 0;
        transition: all 0.3s ease;
      }
      
      .add-friend-modal.show {
        display: block;
        transform: translate(-50%, -50%) scale(1);
        opacity: 1;
      }
      
      .add-friend-title {
        color: #e94560;
        font-size: 18px;
        font-weight: bold;
        margin-bottom: 20px;
        text-align: center;
      }
      
      .add-friend-input {
        width: calc(100% - 20px);
        padding: 10px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: 6px;
        color: #fff;
        font-size: 14px;
        margin-bottom: 15px;
      }
      
      .add-friend-input::placeholder {
        color: #666;
      }
      
      .add-friend-actions {
        display: flex;
        gap: 10px;
      }
      
      .add-friend-btn {
        flex: 1;
        padding: 10px;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        cursor: pointer;
        transition: all 0.2s;
      }
      
      .add-friend-btn.confirm {
        background: #e94560;
        color: #fff;
      }
      
      .add-friend-btn.cancel {
        background: rgba(255, 255, 255, 0.1);
        color: #fff;
      }
      
      .add-friend-btn:hover {
        transform: scale(1.02);
      }
      
      /* 空状态 */
      .social-empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 200px;
        color: #666;
        font-size: 14px;
      }
      
      .social-empty-icon {
        font-size: 48px;
        margin-bottom: 10px;
        opacity: 0.5;
      }
      
      /* 在线状态指示器 */
      .online-indicator {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #4ade80;
        margin-right: 5px;
        box-shadow: 0 0 8px #4ade80;
      }
      
      .offline-indicator {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #666;
        margin-right: 5px;
      }
      
      /* 队伍成员状态 */
      .team-member-status {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: #4ade80;
        margin-left: auto;
      }
      
      .team-member-status.offline {
        background: #666;
      }
    `;
    
    document.head.appendChild(style);
  }

  /**
   * 创建面板
   */
  createPanel() {
    const panel = document.createElement('div');
    panel.id = 'socialPanel';
    panel.className = 'social-panel';
    panel.innerHTML = `
      <div class="social-header">
        <div class="social-title">👥 社交系统</div>
        <button class="social-close-btn" id="socialCloseBtn">×</button>
      </div>
      
      <div class="social-tabs">
        <div class="social-tab active" data-tab="friends">
          <span>好友</span>
          <span id="friendRequestCount" class="friend-request-count"></span>
        </div>
        <div class="social-tab" data-tab="team">
          <span>队伍</span>
        </div>
      </div>
      
      <div class="social-content">
        <!-- 好友面板 -->
        <div id="friendsPanel" class="friends-panel">
          <input type="text" class="friends-search" id="friendsSearch" placeholder="搜索好友...">
          <div class="friends-subtabs">
            <div class="friends-subtab active" data-subtab="online">在线</div>
            <div class="friends-subtab" data-subtab="offline">离线</div>
            <div class="friends-subtab" data-subtab="requests">好友请求</div>
          </div>
          <div class="friends-list" id="friendsList">
            <div class="social-empty">
              <div class="social-empty-icon">👤</div>
              <div>暂无好友</div>
            </div>
          </div>
          <button class="team-action-btn create" style="margin-top: 10px;" id="addFriendBtn">+ 添加好友</button>
        </div>
        
        <!-- 组队面板 -->
        <div id="teamPanel" class="team-panel" style="display: none;">
          <div id="currentTeam" class="team-info">
            <div class="social-empty">
              <div class="social-empty-icon">👥</div>
              <div>暂无队伍</div>
            </div>
          </div>
          
          <div class="team-actions" id="teamActions">
            <button class="team-action-btn create" id="createTeamBtn">创建队伍</button>
          </div>
          
          <div class="available-teams-title">附近队伍</div>
          <div class="available-teams-list" id="availableTeamsList">
            <div class="social-empty">
              <div class="social-empty-icon">🔍</div>
              <div>暂无可用队伍</div>
            </div>
          </div>
        </div>
      </div>
    `;
    
    document.body.appendChild(panel);
    this.container = panel;
    
    // 获取子面板引用
    this.friendPanel = document.getElementById('friendsPanel');
    this.teamPanel = document.getElementById('teamPanel');
  }

  /**
   * 创建添加好友弹窗
   */
  createAddFriendModal() {
    const modal = document.createElement('div');
    modal.id = 'addFriendModal';
    modal.className = 'add-friend-modal';
    modal.innerHTML = `
      <div class="add-friend-title">添加好友</div>
      <input type="text" class="add-friend-input" id="addFriendInput" placeholder="输入玩家昵称或ID">
      <div class="add-friend-actions">
        <button class="add-friend-btn confirm" id="confirmAddFriend">添加</button>
        <button class="add-friend-btn cancel" id="cancelAddFriend">取消</button>
      </div>
    `;
    
    document.body.appendChild(modal);
    this.addFriendModal = modal;
  }

  /**
   * 绑定事件
   */
  bindEvents() {
    // 关闭按钮
    const closeBtn = document.getElementById('socialCloseBtn');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.hide());
    }

    // 标签切换
    const tabs = this.container.querySelectorAll('.social-tab');
    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        this.currentTab = tab.dataset.tab;
        this.showTab(this.currentTab);
      });
    });

    // 好友子标签切换
    const subTabs = this.container.querySelectorAll('.friends-subtab');
    subTabs.forEach(tab => {
      tab.addEventListener('click', () => {
        subTabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        this.currentFriendTab = tab.dataset.subtab;
        this.renderFriendList();
      });
    });

    // 搜索好友
    const searchInput = document.getElementById('friendsSearch');
    if (searchInput) {
      searchInput.addEventListener('input', () => {
        this.renderFriendList(searchInput.value);
      });
    }

    // 添加好友按钮
    const addFriendBtn = document.getElementById('addFriendBtn');
    if (addFriendBtn) {
      addFriendBtn.addEventListener('click', () => this.showAddFriendModal());
    }

    // 创建队伍按钮
    const createTeamBtn = document.getElementById('createTeamBtn');
    if (createTeamBtn) {
      createTeamBtn.addEventListener('click', () => this.createTeam());
    }

    // 键盘快捷键 F 打开社交面板
    this.keydownHandler = (e) => {
      if (e.key === 'f' || e.key === 'F') {
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
   * 显示社交面板
   */
  show() {
    this.isOpen = true;
    this.container.classList.add('show');
    this.showTab(this.currentTab);
  }

  /**
   * 隐藏社交面板
   */
  hide() {
    this.isOpen = false;
    this.container.classList.remove('show');
    this.hideAddFriendModal();
  }

  /**
   * 显示指定标签
   */
  showTab(tab) {
    if (tab === 'friends') {
      this.friendPanel.style.display = 'flex';
      this.teamPanel.style.display = 'none';
      this.renderFriendList();
    } else if (tab === 'team') {
      this.friendPanel.style.display = 'none';
      this.teamPanel.style.display = 'flex';
      this.renderTeamPanel();
    }
  }

  /**
   * 显示添加好友弹窗
   */
  showAddFriendModal() {
    if (!this.addFriendModal) {
      this.createAddFriendModal();
    }
    
    // 绑定弹窗事件
    const confirmBtn = document.getElementById('confirmAddFriend');
    const cancelBtn = document.getElementById('cancelAddFriend');
    
    if (confirmBtn) {
      confirmBtn.addEventListener('click', () => this.sendFriendRequest());
    }
    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.hideAddFriendModal());
    }
    
    this.addFriendModal.classList.add('show');
  }

  /**
   * 隐藏添加好友弹窗
   */
  hideAddFriendModal() {
    if (this.addFriendModal) {
      this.addFriendModal.classList.remove('show');
      document.getElementById('addFriendInput').value = '';
    }
  }

  /**
   * 发送好友请求
   */
  sendFriendRequest() {
    const input = document.getElementById('addFriendInput');
    const targetName = input.value.trim();
    
    if (!targetName) {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('请输入玩家昵称或ID', 'warning', 2000);
      }
      return;
    }
    
    // 模拟发送好友请求
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_FRIEND_REQUEST || 7002, {
        target_name: targetName
      });
    }
    
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast(`已向 ${targetName} 发送好友请求`, 'success', 2000);
    }
    
    this.hideAddFriendModal();
  }

  /**
   * 接受好友请求
   */
  acceptFriendRequest(friendId) {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_FRIEND_ACCEPT || 7003, {
        friend_id: friendId
      });
    }
    
    // 从请求列表移除并添加到好友列表
    const requestIndex = this.friends.requests.findIndex(f => f.id === friendId);
    if (requestIndex !== -1) {
      const friend = this.friends.requests[requestIndex];
      this.friends.requests.splice(requestIndex, 1);
      this.friends.online.push(friend);
      
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast(`已添加 ${friend.name} 为好友`, 'success', 2000);
      }
      
      this.updateFriendRequestCount();
      this.renderFriendList();
    }
  }

  /**
   * 拒绝好友请求
   */
  rejectFriendRequest(friendId) {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_FRIEND_REJECT || 7004, {
        friend_id: friendId
      });
    }
    
    // 从请求列表移除
    this.friends.requests = this.friends.requests.filter(f => f.id !== friendId);
    
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('已拒绝好友请求', 'info', 2000);
    }
    
    this.updateFriendRequestCount();
    this.renderFriendList();
  }

  /**
   * 删除好友
   */
  removeFriend(friendId) {
    if (confirm('确定要删除这位好友吗？')) {
      if (window.GameWS && window.GameWS.isConnected) {
        window.GameWS.send(window.Protocol.CMD_FRIEND_REMOVE || 7005, {
          friend_id: friendId
        });
      }
      
      // 从好友列表移除
      this.friends.online = this.friends.online.filter(f => f.id !== friendId);
      this.friends.offline = this.friends.offline.filter(f => f.id !== friendId);
      
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('已删除好友', 'info', 2000);
      }
      
      this.renderFriendList();
    }
  }

  /**
   * 邀请好友组队
   */
  inviteFriend(friendId) {
    const friend = this.friends.online.find(f => f.id === friendId);
    if (!friend) return;
    
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_TEAM_INVITE || 8003, {
        target_id: friendId
      });
    }
    
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast(`已邀请 ${friend.name} 加入队伍`, 'success', 2000);
    }
  }

  /**
   * 创建队伍
   */
  createTeam() {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_TEAM_CREATE || 8001, {});
    }
    
    // 模拟创建队伍
    this.team = {
      id: 1,
      name: '我的队伍',
      leader_id: 1,
      leader_name: '玩家角色',
      member_count: 1
    };
    
    this.teamMembers = [
      {
        id: 1,
        name: '玩家角色',
        level: 30,
        role: 'leader',
        online: true,
        hp: 100,
        max_hp: 100
      }
    ];
    
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('队伍创建成功', 'success', 2000);
    }
    
    this.renderTeamPanel();
  }

  /**
   * 加入队伍
   */
  joinTeam(teamId) {
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(window.Protocol.CMD_TEAM_JOIN || 8002, {
        team_id: teamId
      });
    }
    
    // 模拟加入队伍
    const team = this.availableTeams.find(t => t.id === teamId);
    if (team) {
      this.team = team;
      this.teamMembers = team.members;
      
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast(`已加入队伍 ${team.name}`, 'success', 2000);
      }
      
      this.renderTeamPanel();
    }
  }

  /**
   * 离开队伍
   */
  leaveTeam() {
    if (confirm('确定要离开队伍吗？')) {
      if (window.GameWS && window.GameWS.isConnected) {
        window.GameWS.send(window.Protocol.CMD_TEAM_LEAVE || 8004, {});
      }
      
      this.team = null;
      this.teamMembers = [];
      
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('已离开队伍', 'info', 2000);
      }
      
      this.renderTeamPanel();
    }
  }

  /**
   * 踢出队员
   */
  kickMember(memberId) {
    if (confirm('确定要踢出这位队员吗？')) {
      if (window.GameWS && window.GameWS.isConnected) {
        window.GameWS.send(window.Protocol.CMD_TEAM_KICK || 8005, {
          member_id: memberId
        });
      }
      
      this.teamMembers = this.teamMembers.filter(m => m.id !== memberId);
      this.team.member_count = this.teamMembers.length;
      
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('已踢出队员', 'info', 2000);
      }
      
      this.renderTeamPanel();
    }
  }

  /**
   * 更新好友请求数量
   */
  updateFriendRequestCount() {
    const countEl = document.getElementById('friendRequestCount');
    if (countEl) {
      if (this.friends.requests.length > 0) {
        countEl.textContent = `(${this.friends.requests.length})`;
        countEl.style.display = 'inline';
        countEl.style.background = '#ef4444';
        countEl.style.color = '#fff';
        countEl.style.fontSize = '11px';
        countEl.style.padding = '2px 5px';
        countEl.style.borderRadius = '10px';
        countEl.style.marginLeft = '5px';
      } else {
        countEl.textContent = '';
      }
    }
  }

  /**
   * 渲染好友列表
   */
  renderFriendList(searchText = '') {
    const listEl = document.getElementById('friendsList');
    if (!listEl) return;
    
    let friends = [];
    switch (this.currentFriendTab) {
      case 'online':
        friends = this.friends.online;
        break;
      case 'offline':
        friends = this.friends.offline;
        break;
      case 'requests':
        friends = this.friends.requests;
        break;
    }
    
    // 搜索过滤
    if (searchText) {
      friends = friends.filter(f => f.name.toLowerCase().includes(searchText.toLowerCase()));
    }
    
    if (friends.length === 0) {
      const emptyTexts = {
        online: '暂无在线好友',
        offline: '暂无离线好友',
        requests: '暂无好友请求'
      };
      listEl.innerHTML = `
        <div class="social-empty">
          <div class="social-empty-icon">👤</div>
          <div>${emptyTexts[this.currentFriendTab]}</div>
        </div>
      `;
      return;
    }
    
    listEl.innerHTML = friends.map(friend => this.renderFriendItem(friend)).join('');
    
    // 绑定事件
    setTimeout(() => {
      // 聊天按钮
      listEl.querySelectorAll('.friend-action-btn.chat').forEach(btn => {
        btn.addEventListener('click', () => this.openChat(parseInt(btn.dataset.id)));
      });
      
      // 邀请按钮
      listEl.querySelectorAll('.friend-action-btn.invite').forEach(btn => {
        btn.addEventListener('click', () => this.inviteFriend(parseInt(btn.dataset.id)));
      });
      
      // 接受按钮
      listEl.querySelectorAll('.friend-action-btn.accept').forEach(btn => {
        btn.addEventListener('click', () => this.acceptFriendRequest(parseInt(btn.dataset.id)));
      });
      
      // 拒绝按钮
      listEl.querySelectorAll('.friend-action-btn.reject').forEach(btn => {
        btn.addEventListener('click', () => this.rejectFriendRequest(parseInt(btn.dataset.id)));
      });
      
      // 删除按钮
      listEl.querySelectorAll('.friend-action-btn.remove').forEach(btn => {
        btn.addEventListener('click', () => this.removeFriend(parseInt(btn.dataset.id)));
      });
    }, 50);
  }

  /**
   * 渲染单个好友项
   */
  renderFriendItem(friend) {
    const isOnline = this.currentFriendTab === 'online' || friend.online;
    const isRequest = this.currentFriendTab === 'requests';
    
    let actions = '';
    if (isRequest) {
      actions = `
        <div class="friend-actions">
          <button class="friend-action-btn accept" data-id="${friend.id}">接受</button>
          <button class="friend-action-btn reject" data-id="${friend.id}">拒绝</button>
        </div>
      `;
    } else {
      actions = `
        <div class="friend-actions">
          <button class="friend-action-btn chat" data-id="${friend.id}">聊天</button>
          <button class="friend-action-btn invite" data-id="${friend.id}">邀请</button>
          <button class="friend-action-btn remove" data-id="${friend.id}">删除</button>
        </div>
      `;
    }
    
    return `
      <div class="friend-item">
        <div class="friend-avatar">${friend.name.charAt(0)}</div>
        <div class="friend-info">
          <div class="friend-name">
            <span class="${isOnline ? 'online-indicator' : 'offline-indicator'}"></span>
            ${friend.name}
            <span class="friend-level">Lv.${friend.level}</span>
          </div>
          <div class="friend-status ${isOnline ? 'online' : 'offline'}">
            ${isOnline ? '在线' : '离线'}
          </div>
        </div>
        ${actions}
      </div>
    `;
  }

  /**
   * 渲染队伍面板
   */
  renderTeamPanel() {
    const teamInfoEl = document.getElementById('currentTeam');
    const teamActionsEl = document.getElementById('teamActions');
    const availableTeamsEl = document.getElementById('availableTeamsList');
    
    if (!teamInfoEl || !teamActionsEl || !availableTeamsEl) return;
    
    // 渲染当前队伍
    if (this.team && this.teamMembers.length > 0) {
      teamInfoEl.innerHTML = `
        <div class="team-info-header">
          <div class="team-name">${this.team.name}</div>
          <div class="team-leader-badge">队长</div>
        </div>
        <div class="team-members-list">
          ${this.teamMembers.map(member => this.renderTeamMember(member)).join('')}
        </div>
      `;
      
      // 渲染队伍操作按钮
      const isLeader = this.teamMembers.some(m => m.id === 1 && m.role === 'leader');
      teamActionsEl.innerHTML = `
        <button class="team-action-btn leave" id="leaveTeamBtn">离开队伍</button>
        ${isLeader ? '<button class="team-action-btn invite" id="inviteTeamBtn">邀请队员</button>' : ''}
      `;
      
      // 绑定队伍操作事件
      setTimeout(() => {
        const leaveBtn = document.getElementById('leaveTeamBtn');
        const inviteBtn = document.getElementById('inviteTeamBtn');
        
        if (leaveBtn) leaveBtn.addEventListener('click', () => this.leaveTeam());
        if (inviteBtn) inviteBtn.addEventListener('click', () => this.showInvitePanel());
      }, 50);
    } else {
      teamInfoEl.innerHTML = `
        <div class="social-empty">
          <div class="social-empty-icon">👥</div>
          <div>暂无队伍</div>
          <div style="font-size: 12px; margin-top: 5px;">点击下方按钮创建或加入队伍</div>
        </div>
      `;
      
      teamActionsEl.innerHTML = `
        <button class="team-action-btn create" id="createTeamBtn">创建队伍</button>
        <button class="team-action-btn join" id="joinTeamBtn">快速组队</button>
      `;
      
      // 绑定创建/加入按钮事件
      setTimeout(() => {
        const createBtn = document.getElementById('createTeamBtn');
        const joinBtn = document.getElementById('joinTeamBtn');
        
        if (createBtn) createBtn.addEventListener('click', () => this.createTeam());
        if (joinBtn) joinBtn.addEventListener('click', () => this.joinRandomTeam());
      }, 50);
    }
    
    // 渲染可用队伍列表
    if (this.availableTeams.length > 0) {
      availableTeamsEl.innerHTML = this.availableTeams.map(team => this.renderAvailableTeam(team)).join('');
      
      // 绑定加入按钮事件
      setTimeout(() => {
        availableTeamsEl.querySelectorAll('.team-action-btn.join').forEach(btn => {
          btn.addEventListener('click', () => this.joinTeam(parseInt(btn.dataset.id)));
        });
      }, 50);
    } else {
      availableTeamsEl.innerHTML = `
        <div class="social-empty">
          <div class="social-empty-icon">🔍</div>
          <div>暂无可用队伍</div>
        </div>
      `;
    }
  }

  /**
   * 渲染队伍成员
   */
  renderTeamMember(member) {
    const isLeader = member.role === 'leader';
    const isSelf = member.id === 1;
    
    return `
      <div class="team-member-item">
        <div class="team-member-avatar">${member.name.charAt(0)}</div>
        <div class="team-member-info">
          <div class="team-member-name">${member.name}${isLeader ? ' 🛡️' : ''}</div>
          <div class="team-member-role ${isLeader ? 'leader' : ''}">${isLeader ? '队长' : '队员'}</div>
        </div>
        <div class="team-member-level">Lv.${member.level}</div>
        <div class="team-member-status ${member.online ? '' : 'offline'}"></div>
      </div>
    `;
  }

  /**
   * 渲染可用队伍
   */
  renderAvailableTeam(team) {
    return `
      <div class="available-team-item">
        <div class="available-team-info">
          <div class="available-team-name">${team.name}</div>
          <div class="available-team-count">成员: ${team.member_count}/5 | 队长: ${team.leader_name}</div>
        </div>
        <button class="team-action-btn join" data-id="${team.id}">加入</button>
      </div>
    `;
  }

  /**
   * 打开聊天窗口
   */
  openChat(friendId) {
    const friend = this.friends.online.find(f => f.id === friendId) || 
                  this.friends.offline.find(f => f.id === friendId);
    
    if (friend) {
      if (this.game.chatSystem) {
        this.game.chatSystem.openPrivateChat(friend.id, friend.name);
      } else {
        if (this.game.uiManager?.toast) {
          this.game.uiManager.toast(`打开与 ${friend.name} 的聊天`, 'info', 2000);
        }
      }
    }
  }

  /**
   * 显示邀请面板
   */
  showInvitePanel() {
    // 切换到好友面板并选择在线好友
    this.currentTab = 'friends';
    this.currentFriendTab = 'online';
    this.showTab('friends');
    
    const tabs = this.container.querySelectorAll('.social-tab');
    tabs.forEach(t => t.classList.remove('active'));
    this.container.querySelector('.social-tab[data-tab="friends"]').classList.add('active');
    
    if (this.game.uiManager?.toast) {
      this.game.uiManager.toast('请选择要邀请的好友', 'info', 2000);
    }
  }

  /**
   * 随机加入队伍
   */
  joinRandomTeam() {
    if (this.availableTeams.length > 0) {
      this.joinTeam(this.availableTeams[0].id);
    } else {
      if (this.game.uiManager?.toast) {
        this.game.uiManager.toast('暂无可用队伍', 'warning', 2000);
      }
    }
  }

  /**
   * 加载测试数据
   */
  loadTestData() {
    // 在线好友
    this.friends.online = [
      { id: 2, name: '江湖侠客', level: 28, online: true, status: '正在练功' },
      { id: 3, name: '逍遥浪子', level: 32, online: true, status: '组队中' },
      { id: 4, name: '武林盟主', level: 45, online: true, status: '在线' },
      { id: 5, name: '白衣书生', level: 25, online: true, status: '挂机中' }
    ];
    
    // 离线好友
    this.friends.offline = [
      { id: 6, name: '铁血战士', level: 30, online: false },
      { id: 7, name: '神秘剑客', level: 35, online: false },
      { id: 8, name: '天山童姥', level: 50, online: false }
    ];
    
    // 好友请求
    this.friends.requests = [
      { id: 9, name: '新来的', level: 15, online: true },
      { id: 10, name: '路人甲', level: 20, online: false }
    ];
    
    // 可用队伍
    this.availableTeams = [
      {
        id: 2,
        name: '天下第一队',
        leader_id: 11,
        leader_name: '战神',
        member_count: 3,
        members: [
          { id: 11, name: '战神', level: 35, role: 'leader', online: true },
          { id: 12, name: '法师', level: 32, role: 'member', online: true },
          { id: 13, name: '刺客', level: 30, role: 'member', online: true }
        ]
      },
      {
        id: 3,
        name: '新手小分队',
        leader_id: 14,
        leader_name: '新手村村长',
        member_count: 2,
        members: [
          { id: 14, name: '新手村村长', level: 20, role: 'leader', online: true },
          { id: 15, name: '菜鸟玩家', level: 18, role: 'member', online: false }
        ]
      },
      {
        id: 4,
        name: '夕阳红战队',
        leader_id: 16,
        leader_name: '老玩家',
        member_count: 4,
        members: [
          { id: 16, name: '老玩家', level: 40, role: 'leader', online: true },
          { id: 17, name: '退休大侠', level: 38, role: 'member', online: true },
          { id: 18, name: '隐士高人', level: 36, role: 'member', online: true },
          { id: 19, name: '江湖前辈', level: 35, role: 'member', online: false }
        ]
      }
    ];
    
    // 更新UI
    this.updateFriendRequestCount();
  }

  /**
   * 处理好友列表更新
   */
  handleFriendList(data) {
    if (!data) return;
    
    this.friends.online = data.online || [];
    this.friends.offline = data.offline || [];
    this.friends.requests = data.requests || [];
    
    this.updateFriendRequestCount();
    
    if (this.isOpen && this.currentTab === 'friends') {
      this.renderFriendList();
    }
  }

  /**
   * 处理队伍信息更新
   */
  handleTeamInfo(data) {
    if (!data) return;
    
    this.team = data.team || null;
    this.teamMembers = data.members || [];
    this.availableTeams = data.available_teams || [];
    
    if (this.isOpen && this.currentTab === 'team') {
      this.renderTeamPanel();
    }
  }

  /**
   * 处理组队邀请
   */
  handleTeamInvite(data) {
    if (!data || !data.leader_name) return;
    
    if (confirm(`${data.leader_name} 邀请你加入队伍，是否接受？`)) {
      if (window.GameWS && window.GameWS.isConnected) {
        window.GameWS.send(window.Protocol.CMD_TEAM_ACCEPT || 8006, {
          team_id: data.team_id
        });
      }
    }
  }
}

// 导出到全局
window.SocialSystem = SocialSystem;