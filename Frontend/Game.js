// 千年江湖 - 游戏主逻辑

class Game {
  constructor() {
    // 游戏状态
    this.state = 'loading'; // loading, login, playing
    
    // 玩家数据
    this.player = {
      id: 0,
      name: '',
      level: 1,
      exp: 0,
      gold: 0,
      hp: 100,
      maxHp: 100,
      mp: 100,
      maxMp: 100,
      attack: 10,
      defense: 5,
      speed: 10,
      hit: 50,
      dodge: 10,
      crit: 5,
      mapId: 1,
      x: 5,
      y: 5
    };
    
    // 其他玩家
    this.players = new Map(); // key=roleID, value={id, name, x, y, hp, maxHp}
    
    // 地图数据
    this.mapEngine = null;
    this.currentMap = null;
    
    // 技能冷却
    this.skillCooldowns = new Map();
    
    // UI元素
    this.ui = {
      loadingOverlay: document.getElementById('loadingOverlay'),
      loginPanel: document.getElementById('loginPanel'),
      gamePanel: document.getElementById('gamePanel'),
      loginBtn: document.getElementById('loginBtn'),
      guestBtn: document.getElementById('guestBtn'),
      username: document.getElementById('username'),
      password: document.getElementById('password'),
      loginMsg: document.getElementById('loginMsg'),
      chatMessages: document.getElementById('chatMessages'),
      chatInput: document.getElementById('chatInput'),
      roleName: document.getElementById('roleName'),
      roleLevel: document.getElementById('roleLevel'),
      roleGold: document.getElementById('roleGold'),
      roleHP: document.getElementById('roleHP'),
      roleMP: document.getElementById('roleMP'),
      onlineCount: document.getElementById('onlineCount'),
      canvas: document.getElementById('gameCanvas'),
      miniMap: document.getElementById('miniMap')
    };
    
    // 初始化
    this.init();
  }
  
  init() {
    console.log('游戏初始化...');
    
    // 绑定事件
    this.bindEvents();
    
    // 初始化地图引擎
    this.initMapEngine();
    
    // 连接WebSocket
    this.connect();
    
    // 更新状态
    this.state = 'login';
    this.ui.loadingOverlay.classList.add('hidden');
    this.ui.loginPanel.classList.remove('hidden');
  }
  
  bindEvents() {
    // 登录按钮
    this.ui.loginBtn.addEventListener('click', () => this.handleLogin());
    this.ui.guestBtn.addEventListener('click', () => this.handleGuestLogin());
    
    // 回车登录
    this.ui.password.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.handleLogin();
    });
    
    // 聊天发送
    this.ui.chatInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.sendChat();
    });
    
    // 键盘事件
    document.addEventListener('keydown', (e) => this.handleKeyDown(e));
    
    // 技能按钮
    document.querySelectorAll('.skill-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const skillId = parseInt(btn.dataset.skill);
        this.useSkill(skillId);
      });
    });
    
    // 鼠标点击移动由MapEngine处理，这里不再绑定
  }
  
  initMapEngine() {
    this.mapEngine = new MapEngine(this.ui.canvas);
    // 设置渲染完成后回调，用于绘制其他玩家
    this.mapEngine.afterRender = () => {
      this.renderPlayers();
    };
    // 设置玩家移动回调，用于更新小地图
    this.mapEngine.onPlayerMove = (x, y) => {
      this.player.x = x;
      this.player.y = y;
      this.renderMiniMap();
    };
  }
  
  connect() {
    window.window.GameWS.connect('ws://localhost:8080/ws', {
      onOpen: () => {
        console.log('WebSocket连接成功');
        this.addChatMessage('系统', '连接服务器成功', 'system');
      },
      onClose: () => {
        console.log('WebSocket连接断开');
        this.addChatMessage('系统', '连接断开，正在重连...', 'system');
        setTimeout(() => this.connect(), 3000);
      },
      onError: (err) => {
        console.error('WebSocket错误:', err);
      },
      onMessage: (cmd, data) => this.handleMessage(cmd, data)
    });
  }
  
  handleLogin() {
    const username = this.ui.username.value.trim();
    const password = this.ui.password.value;
    
    if (!username || !password) {
      this.ui.loginMsg.textContent = '请输入账号和密码';
      return;
    }
    
    this.ui.loginBtn.disabled = true;
    this.ui.loginBtn.textContent = '登录中...';
    
    // 发送登录请求
    window.GameWS.send(Protocol.CMD_LOGIN, {
      type: 'account',
      username: username,
      password: password
    });
  }
  
  handleGuestLogin() {
    this.ui.guestBtn.disabled = true;
    this.ui.guestBtn.textContent = '进入中...';
    
    // 游客登录
    window.GameWS.send(Protocol.CMD_LOGIN, {
      type: 'guest',
      username: 'guest_' + Date.now(),
      password: ''
    });
  }
  
  handleLoginResponse(data) {
    this.ui.loginBtn.disabled = false;
    this.ui.loginBtn.textContent = '登 录';
    this.ui.guestBtn.disabled = false;
    this.ui.guestBtn.textContent = '游客试玩';
    
    if (data.code === 200) {
      this.player.id = data.role_id;
      this.player.name = this.ui.username.value || '游客';
      
      // 进入游戏
      this.enterGame();
    } else {
      this.ui.loginMsg.textContent = data.msg || '登录失败';
    }
  }
  
  enterGame() {
    this.state = 'playing';
    this.ui.loginPanel.classList.add('hidden');
    this.ui.gamePanel.classList.add('active');
    
    // 加载地图
    this.loadMap(this.player.mapId);
    
    // 更新UI
    this.updatePlayerUI();
    
    // 初始化在线人数显示（自己在线）
    this.updateOnlineCount();
    
    this.addChatMessage('系统', '欢迎来到千年江湖！', 'system');
    
    // 开始游戏循环
    this.startGameLoop();
  }
  
  loadMap(mapId) {
    this.player.mapId = mapId;
    
    // 加载地图数据
    this.mapEngine.loadMap(`/Res/Map/${String(mapId).padStart(3, '0')}.map`)
      .then(() => {
        console.log('地图加载成功');
        this.currentMap = this.mapEngine.currentMap;
        
        // 同步玩家位置到地图引擎
        this.syncPlayerPosition();
        
        // 通知服务器玩家进入地图
        window.GameWS.send(Protocol.CMD_ENTER_MAP, {
          map_id: mapId,
          x: this.player.x,
          y: this.player.y
        });
        
        // 绘制小地图
        this.renderMiniMap();
      })
      .catch(err => {
        console.error('地图加载失败:', err);
        // 使用测试地图
        this.createTestMap();
      });
  }
  
  syncPlayerPosition() {
    // 将Game.player的位置同步到MapEngine.player
    this.mapEngine.player.x = this.player.x;
    this.mapEngine.player.y = this.player.y;
    this.mapEngine.syncPlayerPixel();
    this.mapEngine.followPlayer();
  }
  
  createTestMap() {
    // 创建一个简单的测试地图
    const testMap = {
      width: 50,
      height: 50,
      tileWidth: 32,
      tileHeight: 32,
      tiles: []
    };
    
    for (let y = 0; y < testMap.height; y++) {
      testMap.tiles[y] = [];
      for (let x = 0; x < testMap.width; x++) {
        // 边缘为墙
        if (x === 0 || y === 0 || x === testMap.width - 1 || y === testMap.height - 1) {
          testMap.tiles[y][x] = 1;
        } else {
          testMap.tiles[y][x] = 0;
        }
      }
    }
    
    this.currentMap = testMap;
    this.mapEngine.currentMap = testMap;
    this.mapEngine.render();
    this.renderMiniMap();
  }
  
  renderMiniMap() {
    // 使用地图引擎的地图数据
    const mapParser = this.mapEngine?.mapParser;
    if (!mapParser) return;
    
    const ctx = this.ui.miniMap.getContext('2d');
    const width = mapParser.width;
    const height = mapParser.height;
    const scale = 180 / Math.max(width, height);
    
    // 清空背景
    ctx.fillStyle = '#000';
    ctx.fillRect(0, 0, 180, 180);
    
    // 绘制地图（地图数据是一维数组）
    if (mapParser.tiles && mapParser.tiles.length > 0) {
      for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
          const idx = y * width + x;
          const tile = mapParser.tiles[idx];
          if (tile && tile.attr === 1) {
            ctx.fillStyle = '#444';
          } else {
            ctx.fillStyle = '#2d3748';
          }
          ctx.fillRect(x * scale, y * scale, scale, scale);
        }
      }
    } else if (this.currentMap && this.currentMap.tiles) {
      // 备用：使用currentMap的数据（二维数组）
      for (let y = 0; y < this.currentMap.height; y++) {
        for (let x = 0; x < this.currentMap.width; x++) {
          const tile = this.currentMap.tiles[y]?.[x];
          if (tile === 1) {
            ctx.fillStyle = '#444';
          } else {
            ctx.fillStyle = '#2d3748';
          }
          ctx.fillRect(x * scale, y * scale, scale, scale);
        }
      }
    }
    
    // 绘制其他玩家（绿色点）
    this.players.forEach(player => {
      ctx.fillStyle = '#4ade80';
      ctx.beginPath();
      ctx.arc(player.x * scale, player.y * scale, 2.5, 0, Math.PI * 2);
      ctx.fill();
    });
    
    // 绘制自己（红色点）
    ctx.fillStyle = '#e94560';
    ctx.beginPath();
    ctx.arc(this.player.x * scale, this.player.y * scale, 3, 0, Math.PI * 2);
    ctx.fill();
  }
  
  handleMessage(cmd, data) {
    switch (cmd) {
      case Protocol.CMD_LOGIN:
        this.handleLoginResponse(data);
        break;
        
      case Protocol.CMD_HEARTBEAT:
        // 心跳响应
        break;
        
      case Protocol.CMD_MAP_PLAYER:
        this.handleMapPlayer(data);
        break;
        
      case Protocol.CMD_ENTER_MAP:
        this.handleEnterMap(data);
        break;
        
      case Protocol.CMD_LEAVE_MAP:
        this.handleLeaveMap(data);
        break;
        
      case Protocol.CMD_SYNC:
        this.handleSyncMessage(data);
        break;
        
      case Protocol.CMD_MOVE:
        this.handleMoveMessage(data);
        break;
        
      default:
        console.log('未知消息:', cmd, data);
    }
  }
  
  handleMoveMessage(data) {
    console.log('收到移动消息:', data);
    
    if (data.role_id === this.player.id) {
      // 自己的移动 - 地图引擎已经在处理，不需要重复同步
      // 但需要更新Game.player的位置以保持同步
      this.player.x = data.x;
      this.player.y = data.y;
    } else {
      // 其他玩家移动
      console.log('其他玩家移动:', data.role_id, '->', data.x, data.y);
      const player = this.players.get(data.role_id);
      if (player) {
        player.x = data.x;
        player.y = data.y;
      } else {
        console.log('未找到玩家:', data.role_id);
      }
    }
    
    // 更新小地图
    this.renderMiniMap();
  }
  
  handleChatMessage(data) {
    const channelNames = { 0: '世界', 1: '地图', 2: '门派', 3: '私聊' };
    this.addChatMessage(`[${channelNames[data.channel] || '系统'}]${data.from_name}`, data.content, data.channel === 0 ? 'world' : 'map');
  }
  
  // 处理地图玩家列表（服务器主动推送）
  handleMapPlayer(data) {
    if (!data.players || !Array.isArray(data.players)) return;
    
    console.log('收到地图玩家列表:', data.players);
    console.log('this.players:', this.players);
    
    data.players.forEach(player => {
      if (player.role_id === this.player.id) return; // 跳过自己
      
      this.players.set(player.role_id, {
        id: player.role_id,
        name: player.name || `玩家${player.role_id}`,
        x: player.x || 10,
        y: player.y || 10,
        hp: player.hp || 100,
        maxHp: player.maxHp || 100
      });
      console.log(`添加玩家 ${player.name} 到列表，位置 (${player.x}, ${player.y})`);
    });
    
    // 确保renderPlayers存在
    if (typeof this.renderPlayers === 'function') {
      this.renderPlayers();
    } else {
      console.error('renderPlayers不是函数');
    }
  }
  
  handleEnterMap(data) {
    if (data.role_id === this.player.id) return; // 忽略自己
    
    console.log('玩家进入地图:', data);
    
    this.players.set(data.role_id, {
      id: data.role_id,
      name: data.name || `玩家${data.role_id}`,
      x: data.x || 10,
      y: data.y || 10,
      hp: data.hp || 100,
      maxHp: data.maxHp || 100
    });
    
    // 更新在线人数显示
    this.updateOnlineCount();
    
    this.addChatMessage('系统', `${data.name || '某玩家'}进入了地图`, 'system');
    
    // 立即渲染新玩家
    if (typeof this.renderPlayers === 'function') {
      this.renderPlayers();
    }
  }
  
  handleLeaveMap(data) {
    this.players.delete(data.role_id);
    this.addChatMessage('系统', `${data.name || '玩家'}离开了地图`, 'system');
    
    // 更新在线人数显示
    this.updateOnlineCount();
  }
  
  updateOnlineCount() {
    // 在线人数 = 自己 + 其他玩家数量
    const onlineCount = this.players.size + 1;
    if (this.ui.onlineCount) {
      this.ui.onlineCount.textContent = onlineCount;
    }
  }
  
  handleSyncMessage(data) {
    // 同步属性
    if (data.hp !== undefined) this.player.hp = data.hp;
    if (data.maxHp !== undefined) this.player.maxHp = data.maxHp;
    if (data.mp !== undefined) this.player.mp = data.mp;
    if (data.maxMp !== undefined) this.player.maxMp = data.maxMp;
    if (data.exp !== undefined) this.player.exp = data.exp;
    if (data.level !== undefined) this.player.level = data.level;
    if (data.gold !== undefined) this.player.gold = data.gold;
    
    this.updatePlayerUI();
  }
  
  addChatMessage(from, content, type = 'system') {
    const div = document.createElement('div');
    div.className = type;
    div.innerHTML = `<strong>${from}:</strong> ${content}`;
    this.ui.chatMessages.appendChild(div);
    this.ui.chatMessages.scrollTop = this.ui.chatMessages.scrollHeight;
  }
  
  sendChat() {
    const content = this.ui.chatInput.value.trim();
    if (!content) return;
    
    window.GameWS.send(Protocol.CMD_CHAT, {
      channel: 1, // 当前地图
      content: content
    });
    
    this.ui.chatInput.value = '';
  }
  
  handleKeyDown(e) {
    if (this.state !== 'playing') return;
    
    const step = 1;
    let newX = this.player.x;
    let newY = this.player.y;
    
    switch (e.key) {
      case 'ArrowUp':
      case 'w':
      case 'W':
        newY -= step;
        break;
      case 'ArrowDown':
      case 's':
      case 'S':
        newY += step;
        break;
      case 'ArrowLeft':
      case 'a':
      case 'A':
        newX -= step;
        break;
      case 'ArrowRight':
      case 'd':
      case 'D':
        newX += step;
        break;
      case ' ': // 空格攻击
        this.useSkill(0);
        e.preventDefault();
        return;
      default:
        return;
    }
    
    e.preventDefault();
    this.tryMove(newX, newY);
  }
  
  handleCanvasClick(e) {
    if (this.state !== 'playing') return;
    
    const rect = this.ui.canvas.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const clickY = e.clientY - rect.top;
    
    // 计算点击的地图坐标（考虑摄像机偏移）
    const tileX = Math.floor((clickX + this.mapEngine.camera.offsetX) / (this.currentMap?.tileWidth || 32));
    const tileY = Math.floor((clickY + this.mapEngine.camera.offsetY) / (this.currentMap?.tileHeight || 32));
    
    // 简单移动到点击位置
    this.tryMove(tileX, tileY);
  }

  tryMove(newX, newY) {
    // 检查碰撞
    if (this.currentMap?.tiles?.[newY]?.[newX] === 1) {
      return; // 碰撞
    }
    
    // 发送移动请求
    window.GameWS.send(Protocol.CMD_MOVE, {
      x: newX,
      y: newY
    });
    
    // 立即更新位置
    this.player.x = newX;
    this.player.y = newY;
    
    // 同步位置到地图引擎（摄像机跟随）
    this.syncPlayerPosition();
    
    // 重绘
    this.mapEngine.render();
    this.renderMiniMap();
  }
  
  useSkill(skillId) {
    // 检查冷却
    const cooldownEnd = this.skillCooldowns.get(skillId) || 0;
    if (Date.now() < cooldownEnd) {
      return;
    }
    
    // 发送技能请求
    window.GameWS.send(Protocol.CMD_USE_SKILL, {
      skill_id: skillId,
      target_id: 0, // 默认无目标
      x: this.player.x,
      y: this.player.y
    });
    
    // 设置冷却
    const cooldown = skillId === 0 ? 1000 : 3000;
    this.skillCooldowns.set(skillId, Date.now() + cooldown);
    
    // 更新UI
    this.updateSkillUI();
  }
  
  updateSkillUI() {
    document.querySelectorAll('.skill-btn').forEach(btn => {
      const skillId = parseInt(btn.dataset.skill);
      const cooldownEnd = this.skillCooldowns.get(skillId) || 0;
      
      if (Date.now() < cooldownEnd) {
        btn.classList.add('cooling');
      } else {
        btn.classList.remove('cooling');
      }
    });
  }
  
  updatePlayerUI() {
    this.ui.roleName.textContent = this.player.name || '游客';
    this.ui.roleLevel.textContent = `Lv.${this.player.level}`;
    this.ui.roleGold.textContent = `💰 ${this.player.gold}`;
    this.ui.roleHP.textContent = `❤️ ${this.player.hp}/${this.player.maxHp}`;
    this.ui.roleMP.textContent = `💧 ${this.player.mp}/${this.player.maxMp}`;
    
    // 属性面板
    document.getElementById('attrHP').textContent = `${this.player.hp}/${this.player.maxHp}`;
    document.getElementById('attrMP').textContent = `${this.player.mp}/${this.player.maxMp}`;
    document.getElementById('attrAttack').textContent = this.player.attack;
    document.getElementById('attrDefense').textContent = this.player.defense;
    document.getElementById('attrSpeed').textContent = this.player.speed;
    document.getElementById('attrHit').textContent = this.player.hit;
    document.getElementById('attrDodge').textContent = this.player.dodge;
    document.getElementById('attrCrit').textContent = this.player.crit + '%';
  }
  
  startGameLoop() {
    const loop = () => {
      if (this.state === 'playing') {
        // 更新技能冷却UI
        this.updateSkillUI();
      }
      
      requestAnimationFrame(loop);
    };
    
    loop();
  }
  
  renderPlayers() {
    // 获取地图画布
    const canvas = this.mapEngine?.canvas;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    
    // 保存当前变换
    ctx.save();
    
    // 应用摄像机偏移（与地图渲染一致）
    ctx.translate(-this.mapEngine.camera.offsetX, -this.mapEngine.camera.offsetY);
    
    // 绘制其他玩家
    this.players.forEach(player => {
      const screenX = player.x * 32;
      const screenY = player.y * 32;
      
      // 简单绘制为绿色圆形
      ctx.fillStyle = '#4ade80';
      ctx.beginPath();
      ctx.arc(screenX + 16, screenY + 16, 12, 0, Math.PI * 2);
      ctx.fill();
      
      // 名字
      ctx.fillStyle = '#fff';
      ctx.font = '12px Microsoft YaHei';
      ctx.textAlign = 'center';
      ctx.fillText(player.name, screenX + 16, screenY - 5);
    });
    
    // 恢复变换
    ctx.restore();
  }
}

// 协议定义
const Protocol = {
  // 登录相关 1001-1010
  CMD_LOGIN: 1001,
  CMD_LOGOUT: 1002,
  CMD_HEARTBEAT: 1003,
  
  // 游戏相关 2001-2030
  CMD_MOVE: 2001,
  CMD_ATTACK: 2002,
  CMD_USE_SKILL: 2003,
  CMD_CHAT: 2004,
  CMD_PICKUP: 2005,
  CMD_USE_ITEM: 2006,
  CMD_EQUIP: 2007,
  CMD_TRADE: 2008,
  
  // 地图相关 3001-3050
  CMD_ENTER_MAP: 3001,
  CMD_LEAVE_MAP: 3002,
  CMD_MAP_PLAYER: 3003,
  CMD_NPC_TALK: 3004,
  CMD_NPC_TRADE: 3005,
  
  // 武学相关 4001-4020
  CMD_SKILL_LEARN: 4001,
  CMD_SKILL_UPGRADE: 4002,
  
  // 角色相关 5001-5020
  CMD_ROLE_INFO: 5001,
  CMD_ROLE_ATTRIB: 5002,
  CMD_SYNC: 5003
};

// 游戏单例
window.game = new Game();
