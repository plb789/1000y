// 千年江湖 - 游戏主逻辑

class Game {
  constructor() {
    // 游戏状态
    this.state = 'loading'; // loading, login, role_select, role_create, playing
    
    // 账号数据
    this.account = {
      id: 0,
      token: ''
    };
    
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
    
    // 角色列表
    this.roles = [];
    
    // 其他玩家
    this.players = new Map(); // key=roleID, value={id, name, x, y, hp, maxHp}
    
    // 地图数据
    this.mapEngine = null;
    this.currentMap = null;
    
    // 地图配置（用于加载界面背景）
    this.mapConfigs = null;
    
    // 技能冷却
    this.skillCooldowns = new Map();
    
    // UI元素
    this.ui = {
      loadingOverlay: document.getElementById('loadingOverlay'),
      loginPanel: document.getElementById('loginPanel'),
      roleSelectPanel: document.getElementById('roleSelectPanel'),
      roleCreatePanel: document.getElementById('roleCreatePanel'),
      gamePanel: document.getElementById('gamePanel'),
      loginBtn: document.getElementById('loginBtn'),
      registerBtn: document.getElementById('registerBtn'),
      guestBtn: document.getElementById('guestBtn'),
      username: document.getElementById('username'),
      password: document.getElementById('password'),
      loginMsg: document.getElementById('loginMsg'),
      roleList: document.getElementById('roleList'),
      noRoleMsg: document.getElementById('noRoleMsg'),
      createRoleBtn: document.getElementById('createRoleBtn'),
      roleSelectMsg: document.getElementById('roleSelectMsg'),
      roleNameInput: document.getElementById('createRoleName'),
      appearanceOptions: document.getElementById('appearanceOptions'),
      confirmCreateBtn: document.getElementById('confirmCreateBtn'),
      backToSelectBtn: document.getElementById('backToSelectBtn'),
      roleCreateMsg: document.getElementById('roleCreateMsg'),
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
  
  async init() {
    console.log('游戏初始化...');
    
    // 初始化UI、特效、音效系统
    this.initSystems();
    
    // 绑定事件
    this.bindEvents();
    
    // 初始化地图引擎
    this.initMapEngine();
    
    // 加载地图配置（用于加载界面背景）
    await this.loadMapConfigs();
    
    // 连接WebSocket
    this.connect();
    
    // 更新状态
    this.state = 'login';
    this.ui.loadingOverlay.classList.add('hidden');
    this.ui.loginPanel.classList.remove('hidden');
  }
  
  /**
   * 初始化UI、特效、音效系统
   */
  initSystems() {
    // 初始化粒子系统
    if (window.ParticleSystem) {
      this.particleSystem = new window.ParticleSystem(this.ui.canvas);
    }
    
    // 初始化特效管理器
    if (window.EffectManager) {
      this.effectManager = window.EffectManager;
      if (this.particleSystem) {
        this.effectManager.setParticleSystem(this.particleSystem);
      }
    }
    
    // 初始化音效系统
    if (window.SoundManager) {
      this.soundManager = window.SoundManager;
    }
    
    // UI管理器已自动初始化
    if (window.UIManager) {
      this.uiManager = window.UIManager;
    }
    
    // 特效设置
    this.effectSettings = {
      enableParticles: true,
      enableSkillEffects: true,
      enableScreenShake: true
    };
    
    // 事件监听器存储（用于销毁时移除）
    this.eventListeners = [];
    
    console.log('UI、特效、音效系统初始化完成');
  }
  
  async loadMapConfigs() {
    try {
      const response = await fetch('/Res/Map/maps.json');
      if (response.ok) {
        const data = await response.json();
        this.mapConfigs = data.maps || [];
        console.log('地图配置加载成功');
      } else {
        console.log('地图配置文件响应异常，使用默认配置');
        this.mapConfigs = [];
      }
    } catch (err) {
      console.log('没有找到地图配置文件，使用默认配置');
      this.mapConfigs = [];
    }
  }
  
  bindEvents() {
    // 登录按钮
    if (this.ui.loginBtn) {
      this.ui.loginBtn.addEventListener('click', () => this.handleLogin());
    }
    if (this.ui.registerBtn) {
      this.ui.registerBtn.addEventListener('click', () => this.handleRegister());
    }
    if (this.ui.guestBtn) {
      this.ui.guestBtn.addEventListener('click', () => this.handleGuestLogin());
    }
    
    // 回车登录
    if (this.ui.password) {
      this.ui.password.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') this.handleLogin();
      });
    }
    
    // 角色选择界面
    if (this.ui.createRoleBtn) {
      this.ui.createRoleBtn.addEventListener('click', () => this.showRoleCreatePanel());
    }
    
    // 创建角色界面
    if (this.ui.confirmCreateBtn) {
      this.ui.confirmCreateBtn.addEventListener('click', () => this.handleRoleCreate());
    }
    if (this.ui.backToSelectBtn) {
      this.ui.backToSelectBtn.addEventListener('click', () => this.showRoleSelectPanel());
    }
    
    // 聊天发送
    if (this.ui.chatInput) {
      this.ui.chatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') this.sendChat();
      });
    }
    
    // 键盘事件
    document.addEventListener('keydown', (e) => this.handleKeyDown(e));
    
    // 技能按钮
    document.querySelectorAll('.skill-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        const skillId = parseInt(btn.dataset.skill);
        this.useSkill(skillId);
      });
    });
    
    // 窗口大小调整
    window.addEventListener('resize', () => {
      if (this.state === 'playing') {
        this.mapEngine.resizeCanvas();
      }
    });
    
    // 页面关闭事件 - 清理资源防止内存泄漏
    window.addEventListener('beforeunload', (e) => {
      this.destroy();
    });
    
    // 页面卸载事件
    window.addEventListener('unload', () => {
      this.destroy();
    });
    
    // 设置面板事件
    this.bindSettingsEvents();
    
    // 鼠标点击移动由MapEngine处理，这里不再绑定
  }
  
  /**
   * 绑定设置面板事件
   */
  bindSettingsEvents() {
    const settingsBtn = document.getElementById('settingsBtn');
    const settingsPanel = document.getElementById('settingsPanel');
    const closeSettingsBtn = document.getElementById('closeSettingsBtn');
    const bgmVolume = document.getElementById('bgmVolume');
    const sfxVolume = document.getElementById('sfxVolume');
    const muteAll = document.getElementById('muteAll');
    const enableParticles = document.getElementById('enableParticles');
    const enableSkillEffects = document.getElementById('enableSkillEffects');
    const enableScreenShake = document.getElementById('enableScreenShake');
    
    // 打开设置面板
    if (settingsBtn) {
      this.addEventListener(settingsBtn, 'click', () => {
        if (settingsPanel) {
          settingsPanel.style.display = 'block';
        }
      });
    }
    
    // 关闭设置面板
    if (closeSettingsBtn) {
      this.addEventListener(closeSettingsBtn, 'click', () => {
        if (settingsPanel) {
          settingsPanel.style.display = 'none';
        }
      });
    }
    
    // BGM音量
    if (bgmVolume) {
      this.addEventListener(bgmVolume, 'input', (e) => {
        const volume = parseInt(e.target.value) / 100;
        if (this.soundManager) {
          this.soundManager.setBGMVolume(volume);
        }
      });
    }
    
    // 音效音量
    if (sfxVolume) {
      this.addEventListener(sfxVolume, 'input', (e) => {
        const volume = parseInt(e.target.value) / 100;
        if (this.soundManager) {
          this.soundManager.setSFXVolume(volume);
        }
      });
    }
    
    // 静音
    if (muteAll) {
      this.addEventListener(muteAll, 'change', (e) => {
        const muted = e.target.checked;
        if (this.soundManager) {
          this.soundManager.setBGMMute(muted);
          this.soundManager.setSFXMute(muted);
        }
      });
    }
    
    // 粒子特效
    if (enableParticles) {
      this.addEventListener(enableParticles, 'change', (e) => {
        this.effectSettings.enableParticles = e.target.checked;
      });
    }
    
    // 技能特效
    if (enableSkillEffects) {
      this.addEventListener(enableSkillEffects, 'change', (e) => {
        this.effectSettings.enableSkillEffects = e.target.checked;
      });
    }
    
    // 屏幕震动
    if (enableScreenShake) {
      this.addEventListener(enableScreenShake, 'change', (e) => {
        this.effectSettings.enableScreenShake = e.target.checked;
      });
    }

    // 同步当前状态到UI控件
    this.syncSettingsToUI();
  }

  /**
   * 添加事件监听器并保存引用（用于销毁时移除）
   * @param {Element} element - DOM元素
   * @param {string} event - 事件类型
   * @param {Function} handler - 事件处理函数
   */
  addEventListener(element, event, handler) {
    element.addEventListener(event, handler);
    this.eventListeners.push({ element, event, handler });
  }

  /**
   * 移除所有事件监听器
   */
  removeAllEventListeners() {
    for (const listener of this.eventListeners) {
      listener.element.removeEventListener(listener.event, listener.handler);
    }
    this.eventListeners = [];
  }

  /**
   * 将当前设置状态同步到UI控件
   */
  syncSettingsToUI() {
    const bgmVolume = document.getElementById('bgmVolume');
    const sfxVolume = document.getElementById('sfxVolume');
    const muteAll = document.getElementById('muteAll');
    const enableParticles = document.getElementById('enableParticles');
    const enableSkillEffects = document.getElementById('enableSkillEffects');
    const enableScreenShake = document.getElementById('enableScreenShake');

    // 同步音量设置
    if (bgmVolume && this.soundManager) {
      bgmVolume.value = Math.round(this.soundManager.getBGMVolume() * 100);
    }

    if (sfxVolume && this.soundManager) {
      sfxVolume.value = Math.round(this.soundManager.getSFXVolume() * 100);
    }

    // 同步静音状态（只有两者都静音时才勾选）
    if (muteAll && this.soundManager) {
      muteAll.checked = this.soundManager.isBGMMuted() && this.soundManager.isSFXMuted();
    }

    // 同步特效设置
    if (enableParticles) {
      enableParticles.checked = this.effectSettings.enableParticles;
    }

    if (enableSkillEffects) {
      enableSkillEffects.checked = this.effectSettings.enableSkillEffects;
    }

    if (enableScreenShake) {
      enableScreenShake.checked = this.effectSettings.enableScreenShake;
    }
  }
  
  initMapEngine() {
    this.mapEngine = new MapEngine(this.ui.canvas);
    // 设置渲染完成后回调，用于更新其他玩家位置并绘制
    // 注意：updateOtherPlayers需要在renderPlayers之前调用，确保位置已更新
    this.mapEngine.afterRender = (deltaTime) => {
      try {
        // 基于时间更新其他玩家位置
        if (this.state === 'playing') {
          this.updateOtherPlayers(deltaTime);
        }
        // 绘制其他玩家
        this.renderPlayers();
      } catch (e) {
        console.error('afterRender错误:', e);
      }
    };
    
    // 设置移动前检查回调，用于玩家间碰撞检测
    this.mapEngine.onBeforeMove = (path) => {
      // 过滤掉路径上被其他玩家占据的格子
      if (!path || path.length === 0) return path;
      
      const filteredPath = path.filter(point => {
        // 检查是否有其他玩家在这个格子
        for (const [id, player] of this.players) {
          if (player.x === point.x && player.y === point.y) {
            return false; // 这个点被其他玩家占据，跳过
          }
        }
        return true;
      });
      
      // 如果过滤后路径变短了，需要重新计算终点
      if (filteredPath.length < path.length) {
        // 发送新的目标位置到服务器
        if (filteredPath.length > 0) {
          const newTarget = filteredPath[filteredPath.length - 1];
          if (window.GameWS && window.GameWS.send) {
            window.GameWS.send(2001, {
              x: newTarget.x,
              y: newTarget.y
            });
          }
        } else {
          // 路径被完全堵住，发送当前位置
          if (window.GameWS && window.GameWS.send) {
            window.GameWS.send(2001, {
              x: this.player.x,
              y: this.player.y
            });
          }
        }
      }
      
      return filteredPath;
    };
    
    // 设置玩家阻挡检查回调
    this.mapEngine.onPlayerBlocked = (x, y) => {
      // 检查是否有其他玩家在这个格子
      for (const [id, player] of this.players) {
        if (player.x === x && player.y === y) {
          return true; // 被阻挡
        }
      }
      return false;
    };
    
    // 设置需要重新寻路时的回调
    this.mapEngine.onRepathNeeded = () => {
      const targetX = this.mapEngine.player.moveTargetX;
      const targetY = this.mapEngine.player.moveTargetY;
      if (targetX === null || targetY === null) return;
      
      // 重新计算到目标的路径
      const newPath = AStar.findPath(
        this.player.x, this.player.y,
        targetX, targetY,
        this.mapEngine.mapParser.collision,
        this.mapEngine.mapParser.width,
        this.mapEngine.mapParser.height
      );
      
      // 应用路径前检查（过滤被玩家占据的点）
      if (newPath.length > 0 && this.mapEngine.onBeforeMove) {
        const filteredPath = this.mapEngine.onBeforeMove(newPath);
        if (filteredPath) {
          this.mapEngine.player.movePath = filteredPath;
        } else {
          this.mapEngine.player.movePath = newPath;
        }
      } else {
        this.mapEngine.player.movePath = newPath;
      }
      
      // 发送新的目标位置到服务器
      if (this.mapEngine.player.movePath.length > 0) {
        const newTarget = this.mapEngine.player.movePath[this.mapEngine.player.movePath.length - 1];
        if (window.GameWS && window.GameWS.send) {
          window.GameWS.send(Protocol.CMD_MOVE, {
            x: newTarget.x,
            y: newTarget.y
          });
        }
      }
    };
    
    // 设置玩家移动回调，用于更新小地图
    // 添加节流：只在玩家位置改变时更新小地图
    this.mapEngine.onPlayerMove = (x, y) => {
      if (this.player.x !== x || this.player.y !== y) {
        this.player.x = x;
        this.player.y = y;
        // 节流：小地图更新频率限制为每200ms一次
        if (!this.minimapUpdateTimer) {
          this.minimapUpdateTimer = setTimeout(() => {
            this.renderMiniMap();
            this.minimapUpdateTimer = null;
          }, 200);
        }
      }
    };
    // FPS 更新回调
    this.mapEngine.onFpsUpdate = (fps) => {
      const fpsElement = document.getElementById('fpsValue');
      const fpsContainer = fpsElement.parentElement;
      if (fpsElement) {
        fpsElement.textContent = fps;
        
        // 根据 FPS 值改变颜色
        fpsContainer.className = 'fps-counter';
        if (fps < 30) {
          fpsContainer.classList.add('low');
        } else if (fps < 50) {
          fpsContainer.classList.add('medium');
        } else {
          fpsContainer.classList.add('high');
        }
      }
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
  
  handleRegister() {
    const username = this.ui.username.value.trim();
    const password = this.ui.password.value;
    
    if (!username || !password) {
      this.ui.loginMsg.textContent = '请输入账号和密码';
      return;
    }
    
    if (username.length < 4 || username.length > 20) {
      this.ui.loginMsg.textContent = '用户名长度需4-20位';
      return;
    }
    
    if (password.length < 6 || password.length > 20) {
      this.ui.loginMsg.textContent = '密码长度需6-20位';
      return;
    }
    
    this.ui.registerBtn.disabled = true;
    this.ui.registerBtn.textContent = '注册中...';
    
    // 发送注册请求
    window.GameWS.send(Protocol.CMD_REGISTER, {
      username: username,
      password: password
    });
  }
  
  handleGuestLogin() {
    this.ui.guestBtn.disabled = true;
    this.ui.guestBtn.textContent = '进入中...';
    
    // 游客登录 - 生成短格式用户名
    const guestId = Math.floor(Math.random() * 90000) + 10000;
    window.GameWS.send(Protocol.CMD_LOGIN, {
      type: 'guest',
      username: 'guest_' + guestId,
      password: ''
    });
  }
  
  handleLoginResponse(data) {
    this.ui.loginBtn.disabled = false;
    this.ui.loginBtn.textContent = '登 录';
    this.ui.registerBtn.disabled = false;
    this.ui.registerBtn.textContent = '注 册';
    this.ui.guestBtn.disabled = false;
    this.ui.guestBtn.textContent = '游客试玩';
    
    if (data.code === 200) {
      // 保存账号信息
      this.account.id = data.account_id || 0;
      this.account.token = data.token || '';
      
      // 如果已经有角色ID（重连或游客），直接进入游戏
      if (data.role_id) {
        this.player.id = data.role_id;
        this.player.name = data.name || '游客';
        this.enterGame();
      } else {
        // 否则显示角色选择界面
        this.showRoleSelectPanel();
        // 自动获取角色列表
        window.GameWS.send(Protocol.CMD_ROLE_LIST, {});
      }
    } else {
      this.ui.loginMsg.textContent = data.msg || '登录失败';
    }
  }
  
  handleRegisterResponse(data) {
    this.ui.registerBtn.disabled = false;
    this.ui.registerBtn.textContent = '注 册';
    
    if (data.code === 200) {
      // 注册成功，自动登录
      this.account.id = data.account_id || 0;
      this.account.token = data.token || '';
      this.ui.loginMsg.textContent = '注册成功！';
      this.ui.loginMsg.style.color = '#4ade80';
      
      // 显示角色选择界面
      this.showRoleSelectPanel();
      // 自动获取角色列表
      window.GameWS.send(Protocol.CMD_ROLE_LIST, {});
    } else {
      this.ui.loginMsg.textContent = data.msg || '注册失败';
      this.ui.loginMsg.style.color = '#ff6b6b';
    }
  }
  
  showRoleSelectPanel() {
    this.state = 'role_select';
    this.ui.loginPanel.style.display = 'none';
    this.ui.roleSelectPanel.style.display = 'block';
    this.ui.roleCreatePanel.style.display = 'none';
    this.ui.roleSelectMsg.textContent = '';
  }
  
  showRoleCreatePanel() {
    this.state = 'role_create';
    this.ui.roleSelectPanel.style.display = 'none';
    this.ui.roleCreatePanel.style.display = 'block';
    this.ui.roleNameInput.value = '';
    this.ui.roleCreateMsg.textContent = '';
    
    // 初始化外观选项
    this.initAppearanceOptions();
  }
  
  initAppearanceOptions() {
    const appearances = [0, 1, 2, 3, 4, 5];
    this.ui.appearanceOptions.innerHTML = '';
    
    appearances.forEach((app, index) => {
      const btn = document.createElement('button');
      btn.className = 'login-btn';
      btn.style.width = '50px';
      btn.style.height = '50px';
      btn.style.padding = '0';
      btn.style.fontSize = '14px';
      btn.textContent = index + 1;
      btn.dataset.appearance = app;
      
      if (index === 0) {
        btn.style.borderColor = '#e94560';
        btn.style.boxShadow = '0 0 10px rgba(233, 69, 96, 0.5)';
        btn.dataset.selected = 'true';
      }
      
      btn.addEventListener('click', () => {
        // 移除其他按钮的选中状态
        this.ui.appearanceOptions.querySelectorAll('button').forEach(b => {
          b.style.borderColor = '#4a5568';
          b.style.boxShadow = 'none';
          b.dataset.selected = 'false';
        });
        // 设置当前按钮选中状态
        btn.style.borderColor = '#e94560';
        btn.style.boxShadow = '0 0 10px rgba(233, 69, 96, 0.5)';
        btn.dataset.selected = 'true';
      });
      
      this.ui.appearanceOptions.appendChild(btn);
    });
  }
  
  handleRoleListResponse(data) {
    if (data.code === 200) {
      this.roles = data.roles || [];
      this.renderRoleList();
    } else {
      this.ui.roleSelectMsg.textContent = data.msg || '获取角色列表失败';
    }
  }
  
  renderRoleList() {
    this.ui.roleList.innerHTML = '';
    
    if (this.roles.length === 0) {
      this.ui.noRoleMsg.style.display = 'block';
      return;
    }
    
    this.ui.noRoleMsg.style.display = 'none';
    
    this.roles.forEach(role => {
      const roleCard = document.createElement('div');
      roleCard.className = 'login-box';
      roleCard.style.width = '180px';
      roleCard.style.padding = '20px';
      roleCard.style.cursor = 'pointer';
      roleCard.style.transition = 'all 0.3s';
      
      roleCard.innerHTML = `
        <div style="text-align: center;">
          <div style="color: #e94560; font-size: 18px; margin-bottom: 10px;">${role.name || '无名'}</div>
          <div style="color: #fff; font-size: 14px;">等级: ${role.level || 1}</div>
          <div style="color: #999; font-size: 12px; margin-top: 5px;">${role.gender === 0 ? '男' : '女'}</div>
        </div>
      `;
      
      roleCard.addEventListener('click', () => {
        this.selectRole(role.id);
      });
      
      roleCard.addEventListener('mouseenter', () => {
        roleCard.style.transform = 'translateY(-5px)';
        roleCard.style.boxShadow = '0 5px 20px rgba(233, 69, 96, 0.4)';
      });
      
      roleCard.addEventListener('mouseleave', () => {
        roleCard.style.transform = 'none';
        roleCard.style.boxShadow = '0 0 30px rgba(233, 69, 96, 0.3)';
      });
      
      this.ui.roleList.appendChild(roleCard);
    });
  }
  
  selectRole(roleId) {
    this.ui.roleSelectMsg.textContent = '正在进入游戏...';
    window.GameWS.send(Protocol.CMD_ROLE_SELECT, { role_id: roleId });
  }
  
  handleRoleSelectResponse(data) {
    if (data.code === 200) {
      // 设置角色信息
      this.player.id = data.role_id;
      this.player.name = data.name;
      this.player.level = data.level || 1;
      this.player.hp = data.hp || 100;
      this.player.maxHp = data.max_hp || 100;
      this.player.mp = data.mp || 100;
      this.player.maxMp = data.max_mp || 100;
      this.player.attack = data.attack || 10;
      this.player.defense = data.defense || 5;
      this.player.speed = data.speed || 10;
      this.player.gold = data.gold || 0;
      this.player.mapId = data.map_id || 1;
      this.player.x = data.x || 5;
      this.player.y = data.y || 5;
      
      // 进入游戏
      this.enterGame();
    } else {
      this.ui.roleSelectMsg.textContent = data.msg || '选择角色失败';
    }
  }
  
  handleRoleCreate() {
    const name = this.ui.roleNameInput.value.trim();
    const genderRadio = document.querySelector('input[name="gender"]:checked');
    const gender = genderRadio ? parseInt(genderRadio.value) : 0;
    const appearanceBtn = this.ui.appearanceOptions.querySelector('button[data-selected="true"]');
    const appearance = appearanceBtn ? parseInt(appearanceBtn.dataset.appearance) : 0;
    
    if (!name || name.length < 2 || name.length > 12) {
      this.ui.roleCreateMsg.textContent = '角色名长度需2-12位';
      return;
    }
    
    this.ui.confirmCreateBtn.disabled = true;
    this.ui.confirmCreateBtn.textContent = '创建中...';
    
    window.GameWS.send(Protocol.CMD_ROLE_CREATE, {
      name: name,
      gender: gender,
      appearance: appearance
    });
  }
  
  handleRoleCreateResponse(data) {
    this.ui.confirmCreateBtn.disabled = false;
    this.ui.confirmCreateBtn.textContent = '确认创建';
    
    if (data.code === 200) {
      // 创建成功，返回角色选择界面
      this.ui.roleCreateMsg.textContent = '创建成功！';
      this.ui.roleCreateMsg.style.color = '#4ade80';
      
      // 重新获取角色列表
      setTimeout(() => {
        this.showRoleSelectPanel();
        window.GameWS.send(Protocol.CMD_ROLE_LIST, {});
      }, 500);
    } else {
      this.ui.roleCreateMsg.textContent = data.msg || '创建失败';
      this.ui.roleCreateMsg.style.color = '#ff6b6b';
    }
  }
  
  enterGame() {
    this.state = 'playing';
    this.ui.loginPanel.style.display = 'none';
    this.ui.roleSelectPanel.style.display = 'none';
    this.ui.roleCreatePanel.style.display = 'none';
    this.ui.gamePanel.classList.add('active');
    
    // 游戏面板显示后，调整画布大小
    this.mapEngine.resizeCanvas();
    
    // 初始化音效系统（需要用户交互后才能启动）
    if (this.soundManager) {
      this.soundManager.ensureContext();
    }
    
    // 加载地图
    this.loadMap(this.player.mapId);
    
    // 更新UI
    this.updatePlayerUI();
    
    // 初始化在线人数显示（自己在线）
    this.updateOnlineCount();
    
    this.addChatMessage('系统', '欢迎来到千年江湖！', 'system');
    
    // 显示欢迎提示
    if (this.uiManager) {
      this.uiManager.toast('欢迎来到千年江湖！', 'success', 3000);
    }
    
    // 开始游戏循环
    this.startGameLoop();
  }
  
  async loadMap(mapId) {
    this.player.mapId = mapId;
    
    // 获取地图配置（用于加载界面背景）
    const mapConfig = this.getMapConfig(mapId);
    
    // 显示加载界面
    this.showLoadingScreen(mapConfig);
    
    // 构建地图文件和瓦片图集路径
    const mapFile = `/Res/Map/${String(mapId).padStart(3, '0')}.map`;
    const tilesetFile = `/Res/Map/${String(mapId).padStart(3, '0')}.png`;
    const animationFile = `/Res/Map/${String(mapId).padStart(3, '0')}_anim.json`;
    
    try {
      // 步骤1: 加载动画数据 (10%)
      this.updateLoadingProgress(10, '加载动画数据...');
      let animationData = null;
      try {
        const response = await fetch(animationFile);
        if (response.ok) {
          const data = await response.json();
          animationData = data.animations || data;
          console.log('动画数据加载成功');
        } else {
          console.log('动画数据文件响应异常');
          animationData = null; // 确保重置为null
        }
      } catch (err) {
        console.log('动画数据文件加载失败:', err.message);
        animationData = null; // 确保重置为null
      }
      
      // 步骤2: 加载地图数据 (30%)
      this.updateLoadingProgress(30, '加载地图数据...');
      await this.mapEngine.loadMapData(mapFile);
      
      // 步骤3: 加载瓦片图集 (60%)
      this.updateLoadingProgress(60, '加载瓦片图集...');
      await this.mapEngine.loadTileset(tilesetFile);
      
      // 步骤4: 初始化地图渲染器 (80%)
      this.updateLoadingProgress(80, '初始化地图...');
      await this.mapEngine.initializeMap(animationData);
      
      console.log('地图加载成功');
      this.currentMap = this.mapEngine.currentMap;
      
      // 步骤5: 同步玩家位置 (90%)
      this.updateLoadingProgress(90, '同步玩家位置...');
      this.syncPlayerPosition();
      
      // 步骤6: 通知服务器 (95%)
      this.updateLoadingProgress(95, '连接服务器...');
      window.GameWS.send(Protocol.CMD_ENTER_MAP, {
        role_id: this.player.id,
        map_id: mapId,
        x: this.player.x,
        y: this.player.y
      });
      
      // 步骤7: 绘制小地图 (100%)
      this.updateLoadingProgress(100, '加载完成');
      this.renderMiniMap();
      
      // 隐藏加载界面（延迟500ms显示完成状态）
      setTimeout(() => {
        this.hideLoadingScreen();
      }, 500);
      
    } catch (err) {
      console.error('地图加载失败:', err);
      
      try {
        // 尝试不加载瓦片图集再次加载
        this.updateLoadingProgress(40, '重试加载地图数据...');
        await this.mapEngine.loadMapData(mapFile);
        
        this.updateLoadingProgress(60, '跳过瓦片图集...');
        await this.mapEngine.loadTileset(null);
        
        this.updateLoadingProgress(80, '初始化地图...');
        await this.mapEngine.initializeMap(animationData);
        
        console.log('地图加载成功（无瓦片图集）');
        this.currentMap = this.mapEngine.currentMap;
        this.syncPlayerPosition();
        window.GameWS.send(Protocol.CMD_ENTER_MAP, {
          role_id: this.player.id,
          map_id: mapId,
          x: this.player.x,
          y: this.player.y
        });
        this.renderMiniMap();
        
        this.updateLoadingProgress(100, '加载完成');
        setTimeout(() => {
          this.hideLoadingScreen();
        }, 500);
        
      } catch {
        // 使用测试地图
        this.updateLoadingProgress(100, '使用测试地图');
        this.createTestMap();
        setTimeout(() => {
          this.hideLoadingScreen();
        }, 500);
      }
    }
  }
  
  getMapConfig(mapId) {
    if (!this.mapConfigs) return null;
    return this.mapConfigs.find(m => m.id === mapId);
  }
  
  showLoadingScreen(mapConfig) {
    const screen = document.getElementById('loadingScreen');
    if (!screen) return;
    
    screen.style.display = 'block';
    
    // 设置背景图
    const bg = document.getElementById('loadingBg');
    if (bg) {
      if (mapConfig && mapConfig.bgImage) {
        bg.style.backgroundImage = `url(${mapConfig.bgImage})`;
      } else {
        bg.style.backgroundImage = 'url(/Res/Map/loading/default.jpg)';
      }
    }
    
    // 设置地图名称
    const mapName = document.getElementById('loadingMapName');
    if (mapName) {
      if (mapConfig && mapConfig.name) {
        mapName.textContent = `正在进入「${mapConfig.name}」`;
      } else {
        mapName.textContent = '';
      }
    }
    
    // 重置进度
    this.updateLoadingProgress(0, '正在进入地图...');
  }
  
  hideLoadingScreen() {
    const screen = document.getElementById('loadingScreen');
    if (!screen) return;
    
    screen.style.opacity = '0';
    setTimeout(() => {
      screen.style.display = 'none';
      screen.style.opacity = '1';
    }, 500);
  }
  
  updateLoadingProgress(percent, message) {
    const progress = document.getElementById('loadingProgress');
    const percentText = document.getElementById('loadingPercent');
    const subtitle = document.getElementById('loadingSubtitle');
    
    if (progress) progress.style.width = `${Math.min(percent, 100)}%`;
    if (percentText) percentText.textContent = `${Math.min(percent, 100)}%`;
    if (subtitle && message) subtitle.textContent = message;
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
    // 获取地图解析器
    const mapParser = this.mapEngine?.mapParser;
    if (!mapParser) return;
    
    // 获取Canvas上下文
    const ctx = this.ui.miniMap.getContext('2d');
    const mapWidth = mapParser.width;
    const mapHeight = mapParser.height;
    
    // 获取玩家位置
    const playerX = this.player.x;
    const playerY = this.player.y;
    
    // 获取实际Canvas尺寸（防止样式尺寸和实际尺寸不一致）
    const canvasWidth = this.ui.miniMap.width;
    const canvasHeight = this.ui.miniMap.height;
    
    // 小地图配置
    const backgroundColor = '#1a1a1a';
    const obstacleColor = '#4a4a4a';
    const walkableColor = '#3a3a5a';
    const playerColor = '#ff4444';
    const otherPlayerColor = '#44ff44';
    
    // 清空整个画布（使用实际canvas尺寸）
    ctx.fillStyle = backgroundColor;
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);
    
    // ===== 关键优化：小地图显示范围与主地图可见范围保持一致 =====
    // 获取主地图画布尺寸
    const mainCanvasWidth = this.ui.canvas.width;
    const mainCanvasHeight = this.ui.canvas.height;
    const tileSize = this.mapEngine?.tileSize || 48;
    
    // 计算主地图可见的瓦片范围
    const mainVisibleTileWidth = mainCanvasWidth / tileSize;
    const mainVisibleTileHeight = mainCanvasHeight / tileSize;
    
    // 计算缩放比例：让小地图显示的瓦片范围等于主地图的可见范围
    // 小地图显示的瓦片宽度 = 小地图画布宽度 / scale
    // 要让两者相等：scale = 小地图画布宽度 / 主地图可见瓦片宽度
    const scaleX = canvasWidth / mainVisibleTileWidth;
    const scaleY = canvasHeight / mainVisibleTileHeight;
    
    // 取较小的缩放比例，确保小地图显示的范围不超过主地图可见范围
    let scale = Math.min(scaleX, scaleY);
    
    // 处理极端情况：地图尺寸为0
    if (mainCanvasWidth === 0 || mainCanvasHeight === 0 || isNaN(scale) || !isFinite(scale)) {
      scale = 1;
    }
    
    // 计算偏移量，让玩家始终显示在小地图中心
    let finalOffsetX = canvasWidth / 2 - playerX * scale;
    let finalOffsetY = canvasHeight / 2 - playerY * scale;
    
    // 限制偏移量，确保画布被地图完全填满（允许裁剪地图边缘）
    const scaledMapWidth = mapWidth * scale;
    const scaledMapHeight = mapHeight * scale;
    
    // 水平方向：确保地图左右填满画布（允许裁剪地图边缘）
    finalOffsetX = Math.max(-scaledMapWidth + canvasWidth, Math.min(0, finalOffsetX));
    
    // 垂直方向：确保地图上下填满画布（允许裁剪地图边缘）
    finalOffsetY = Math.max(-scaledMapHeight + canvasHeight, Math.min(0, finalOffsetY));
    
    // ===== 性能优化：只计算和绘制可见范围的瓦片 =====
    // 计算小地图画布能显示的瓦片范围（向外扩展2格作为边界）
    const visibleTileStartX = Math.max(0, Math.floor(-finalOffsetX / scale) - 2);
    const visibleTileStartY = Math.max(0, Math.floor(-finalOffsetY / scale) - 2);
    const visibleTileEndX = Math.min(mapWidth, Math.ceil((canvasWidth - finalOffsetX) / scale) + 2);
    const visibleTileEndY = Math.min(mapHeight, Math.ceil((canvasHeight - finalOffsetY) / scale) + 2);
    
    // 绘制地图瓦片（只绘制可见范围内的瓦片）
    const tiles = mapParser.tiles || [];
    
    if (tiles.length > 0) {
      // 只遍历可见范围内的瓦片
      for (let y = visibleTileStartY; y < visibleTileEndY; y++) {
        for (let x = visibleTileStartX; x < visibleTileEndX; x++) {
          const index = y * mapWidth + x;
          const tile = tiles[index];
          
          // 根据瓦片属性设置颜色
          if (tile && tile.attr === 1) {
            ctx.fillStyle = obstacleColor;
          } else {
            ctx.fillStyle = walkableColor;
          }
          
          // 绘制瓦片
          const drawX = finalOffsetX + x * scale;
          const drawY = finalOffsetY + y * scale;
          
          ctx.fillRect(drawX, drawY, Math.max(1, scale), Math.max(1, scale));
        }
      }
    } else if (this.currentMap?.tiles) {
      // 备用绘制方式：只遍历可见范围内的瓦片
      for (let y = visibleTileStartY; y < visibleTileEndY; y++) {
        for (let x = visibleTileStartX; x < visibleTileEndX; x++) {
          const tile = this.currentMap.tiles[y]?.[x];
          ctx.fillStyle = tile === 1 ? obstacleColor : walkableColor;
          
          const drawX = finalOffsetX + x * scale;
          const drawY = finalOffsetY + y * scale;
          
          ctx.fillRect(drawX, drawY, Math.max(1, scale), Math.max(1, scale));
        }
      }
    }
    
    // ===== 性能优化：只绘制可见范围内的其他玩家 =====
    // 计算可见区域的瓦片范围（用于玩家筛选）
    const viewRadius = Math.max(canvasWidth, canvasHeight) / scale / 2 + 2;
    
    // 计算玩家自己在小地图上的实际位置（考虑偏移量限制后的真实位置）
    const selfDrawX = canvasWidth / 2;
    const selfDrawY = canvasHeight / 2;
    
    if (this.players && this.players.size > 0) {
      this.players.forEach(otherPlayer => {
        // 先检查玩家是否在可见范围内（距离判断）
        const dx = otherPlayer.x - playerX;
        const dy = otherPlayer.y - playerY;
        
        // 只绘制可见范围内的玩家
        if (Math.abs(dx) <= viewRadius && Math.abs(dy) <= viewRadius) {
          // 基于玩家中心计算其他玩家的位置（相对坐标方式，更准确）
          const px = selfDrawX + dx * scale;
          const py = selfDrawY + dy * scale;
          
          ctx.fillStyle = otherPlayerColor;
          ctx.beginPath();
          ctx.arc(px, py, 2, 0, Math.PI * 2);
          ctx.fill();
        }
      });
    }
    
    // 绘制自己（始终在中心）
    ctx.fillStyle = playerColor;
    ctx.beginPath();
    ctx.arc(selfDrawX, selfDrawY, 3, 0, Math.PI * 2);
    ctx.fill();
  }
  
  handleMessage(cmd, data) {
    switch (cmd) {
      case Protocol.CMD_LOGIN:
        this.handleLoginResponse(data);
        break;
        
      case Protocol.CMD_REGISTER:
        this.handleRegisterResponse(data);
        break;
        
      case Protocol.CMD_HEARTBEAT:
        // 心跳响应
        break;
        
      case Protocol.CMD_ROLE_LIST:
        this.handleRoleListResponse(data);
        break;
        
      case Protocol.CMD_ROLE_CREATE:
        this.handleRoleCreateResponse(data);
        break;
        
      case Protocol.CMD_ROLE_SELECT:
        this.handleRoleSelectResponse(data);
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
        
      case Protocol.CMD_ONLINE_COUNT:
        this.handleOnlineCount(data);
        break;
        
      case Protocol.CMD_SYNC:
        this.handleSyncMessage(data);
        break;
        
      case Protocol.CMD_CHAT:
        this.handleChatMessage(data);
        break;
        
      case Protocol.CMD_MOVE:
        this.handleMoveMessage(data);
        break;
        
      case Protocol.CMD_DAMAGE:
        this.handleDamage(data);
        break;
        
      case Protocol.CMD_DEATH:
        this.handleDeath(data);
        break;
        
      case Protocol.CMD_RESPAWN:
        this.handleRespawn(data);
        break;
        
      case Protocol.CMD_LEVEL_UP:
        this.handleLevelUp(data);
        break;
        
      case Protocol.CMD_BUFF:
        this.handleBuff(data);
        break;
        
      case Protocol.CMD_DEBUFF:
        this.handleDebuff(data);
        break;
        
      case Protocol.CMD_MAP_EVENT:
        this.handleMapEvent(data);
        break;
        
      default:
        console.log('未知消息:', cmd, data);
    }
  }
  
  handleMoveMessage(data) {
    if (data.role_id === this.player.id) {
      // 自己的移动 - 服务器广播回来的当前位置
      this.player.x = data.x;
      this.player.y = data.y;
      // 同步位置到地图引擎
      this.syncPlayerPosition();
    } else {
      const player = this.players.get(data.role_id);
      if (player) {
        const tileSize = this.mapEngine?.tileSize || 48;
        
        // 计算当前位置(显示中的)到新目标的距离
        const currentScreenX = player.lastScreenX !== undefined ? player.lastScreenX : player.x * tileSize;
        const currentScreenY = player.lastScreenY !== undefined ? player.lastScreenY : player.y * tileSize;
        const newTargetX = data.x * tileSize;
        const newTargetY = data.y * tileSize;
        const distToTarget = Math.hypot(newTargetX - currentScreenX, newTargetY - currentScreenY);
        
        // 如果距离超过阈值（3格），说明位置有较大变化（如传送、服务器校正）
        // 直接跳转避免长时间追赶
        if (distToTarget > tileSize * 3) {
          player.lastScreenX = newTargetX;
          player.lastScreenY = newTargetY;
        }
        // 否则继续从当前位置插值到新目标
        
        // 更新到新位置
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
  
  /**
   * 处理伤害消息
   */
  handleDamage(data) {
    const targetId = data.target_id;
    const damage = data.damage || 0;
    const isCritical = data.is_critical || false;
    const isBlocked = data.is_blocked || false;
    const isDodged = data.is_dodged || false;
    
    // 获取目标玩家
    let target = null;
    let isSelf = false;
    
    if (targetId === this.player.id) {
      target = this.player;
      isSelf = true;
    } else {
      target = this.players.get(targetId);
    }
    
    if (!target) return;
    
    // 更新血量
    if (target.hp !== undefined) {
      target.hp = Math.max(0, target.hp - damage);
    }
    
    // 更新UI
    if (isSelf) {
      this.updatePlayerUI();
    }
    
    // 获取目标位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = target.x * tileSize + tileSize / 2;
    const y = target.y * tileSize + tileSize / 2;
    
    // 触发受击特效
    this.triggerHitEffect(isCritical, isBlocked, isDodged, x, y);
    
    // 播放受击音效
    this.playHitSound(isCritical, isBlocked, isDodged);
    
    // 显示伤害数字
    this.showDamageNumber(x, y, damage, isCritical);
  }
  
  /**
   * 触发受击特效
   */
  triggerHitEffect(isCritical, isBlocked, isDodged, x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    let effectName = 'hit_normal';
    
    if (isDodged) {
      effectName = 'dodge';
    } else if (isBlocked) {
      effectName = 'block';
    } else if (isCritical) {
      effectName = 'hit_critical';
    }
    
    this.effectManager.triggerInteractionEffect(effectName, x, y);
  }
  
  /**
   * 播放受击音效
   */
  playHitSound(isCritical, isBlocked, isDodged) {
    if (!this.soundManager) return;
    
    let soundId = 'hit_normal';
    
    if (isDodged) {
      soundId = 'dodge';
    } else if (isBlocked) {
      soundId = 'block';
    } else if (isCritical) {
      soundId = 'hit_critical';
    }
    
    this.soundManager.play(soundId, { volume: 0.5 });
  }
  
  /**
   * 显示伤害数字
   */
  showDamageNumber(x, y, damage, isCritical) {
    if (!this.uiManager) return;
    
    this.uiManager.showDamageNumber({
      x: x,
      y: y - 30,
      damage: damage,
      isCritical: isCritical,
      duration: 1000
    });
  }
  
  /**
   * 处理死亡消息
   */
  handleDeath(data) {
    const targetId = data.target_id;
    
    // 获取目标玩家
    let target = null;
    let isSelf = false;
    
    if (targetId === this.player.id) {
      target = this.player;
      isSelf = true;
    } else {
      target = this.players.get(targetId);
    }
    
    if (!target) return;
    
    // 获取目标位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = target.x * tileSize + tileSize / 2;
    const y = target.y * tileSize + tileSize / 2;
    
    // 触发死亡特效
    this.triggerDeathEffect(x, y);
    
    // 如果是自己死亡，显示死亡提示
    if (isSelf) {
      this.onPlayerDeath();
    }
  }
  
  /**
   * 触发死亡特效
   */
  triggerDeathEffect(x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    this.effectManager.triggerInteractionEffect('death', x, y);
  }
  
  /**
   * 玩家死亡处理
   */
  onPlayerDeath() {
    // 显示死亡UI
    if (this.uiManager) {
      this.uiManager.showDialog({
        title: '你已死亡',
        message: '是否选择复活？',
        buttons: [
          { text: '原地复活', onClick: () => this.respawn() },
          { text: '回城复活', onClick: () => this.respawnToTown() }
        ],
        closeable: false
      });
    }
    
    // 停止玩家移动
    if (this.mapEngine && this.mapEngine.player) {
      this.mapEngine.player.stopMoving();
    }
  }
  
  /**
   * 原地复活
   */
  respawn() {
    // 关闭对话框
    this.uiManager.hideDialog();
    
    // 触发复活特效
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = this.player.x * tileSize + tileSize / 2;
    const y = this.player.y * tileSize + tileSize / 2;
    
    if (this.effectManager) {
      this.effectManager.triggerInteractionEffect('respawn', x, y);
    }
    
    // 发送复活请求
    window.GameWS.send(Protocol.CMD_RESPAWN, {
      type: 'here',
      role_id: this.player.id
    });
  }
  
  /**
   * 回城复活
   */
  respawnToTown() {
    // 关闭对话框
    this.uiManager.hideDialog();
    
    // 发送回城复活请求
    window.GameWS.send(Protocol.CMD_RESPAWN, {
      type: 'town',
      role_id: this.player.id
    });
  }
  
  /**
   * 处理复活消息
   */
  handleRespawn(data) {
    const targetId = data.target_id;
    
    if (targetId === this.player.id) {
      // 自己复活
      this.player.x = data.x || this.player.x;
      this.player.y = data.y || this.player.y;
      this.player.hp = data.hp || this.player.maxHp;
      this.player.mp = data.mp || this.player.maxMp;
      
      // 更新位置到地图引擎
      this.syncPlayerPosition();
      
      // 更新UI
      this.updatePlayerUI();
      
      // 显示复活提示
      if (this.uiManager) {
        this.uiManager.toast('已复活！', 'success', 2000);
      }
    } else {
      // 其他玩家复活
      const player = this.players.get(targetId);
      if (player) {
        player.x = data.x || player.x;
        player.y = data.y || player.y;
        player.hp = data.hp || player.maxHp;
      }
    }
    
    // 触发复活特效
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = (data.x || 0) * tileSize + tileSize / 2;
    const y = (data.y || 0) * tileSize + tileSize / 2;
    
    if (this.effectManager) {
      this.effectManager.triggerInteractionEffect('respawn', x, y);
    }
  }
  
  /**
   * 处理升级消息
   */
  handleLevelUp(data) {
    const targetId = data.target_id;
    
    if (targetId === this.player.id) {
      // 自己升级
      const oldLevel = this.player.level;
      this.player.level = data.level || this.player.level;
      this.player.maxHp = data.max_hp || this.player.maxHp;
      this.player.maxMp = data.max_mp || this.player.maxMp;
      this.player.attack = data.attack || this.player.attack;
      this.player.defense = data.defense || this.player.defense;
      this.player.speed = data.speed || this.player.speed;
      
      // 回满血蓝
      this.player.hp = this.player.maxHp;
      this.player.mp = this.player.maxMp;
      
      // 更新UI
      this.updatePlayerUI();
      
      // 显示升级提示
      if (this.uiManager) {
        this.uiManager.toast(`恭喜升级！Lv.${oldLevel} → Lv.${this.player.level}`, 'success', 3000);
      }
    }
    
    // 获取目标位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = (data.x || this.player.x) * tileSize + tileSize / 2;
    const y = (data.y || this.player.y) * tileSize + tileSize / 2;
    
    // 触发升级特效
    this.triggerLevelUpEffect(x, y);
  }
  
  /**
   * 触发升级特效
   */
  triggerLevelUpEffect(x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    this.effectManager.triggerInteractionEffect('level_up', x, y);
  }
  
  /**
   * 处理增益效果
   */
  handleBuff(data) {
    const targetId = data.target_id;
    const buffType = data.buff_type;
    
    // 获取目标玩家
    let target = null;
    
    if (targetId === this.player.id) {
      target = this.player;
    } else {
      target = this.players.get(targetId);
    }
    
    if (!target) return;
    
    // 获取目标位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = target.x * tileSize + tileSize / 2;
    const y = target.y * tileSize + tileSize / 2;
    
    // 触发增益特效
    this.triggerBuffEffect(buffType, x, y);
    
    // 如果是自己，显示提示
    if (targetId === this.player.id && this.uiManager) {
      const buffNames = {
        attack: '攻击增益',
        defense: '防御增益',
        speed: '速度增益',
        heal: '持续治疗'
      };
      this.uiManager.toast(`获得${buffNames[buffType] || buffType}效果！`, 'success', 2000);
    }
  }
  
  /**
   * 触发增益特效
   */
  triggerBuffEffect(buffType, x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    const buffMap = {
      attack: 'buff_attack',
      defense: 'buff_defense',
      speed: 'buff_speed',
      heal: 'buff_heal'
    };
    
    const effectName = buffMap[buffType] || 'buff_attack';
    this.effectManager.triggerStatusEffect(effectName, { x, y });
  }
  
  /**
   * 处理减益效果
   */
  handleDebuff(data) {
    const targetId = data.target_id;
    const debuffType = data.debuff_type;
    
    // 获取目标玩家
    let target = null;
    
    if (targetId === this.player.id) {
      target = this.player;
    } else {
      target = this.players.get(targetId);
    }
    
    if (!target) return;
    
    // 获取目标位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = target.x * tileSize + tileSize / 2;
    const y = target.y * tileSize + tileSize / 2;
    
    // 触发减益特效
    this.triggerDebuffEffect(debuffType, x, y);
    
    // 如果是自己，显示提示
    if (targetId === this.player.id && this.uiManager) {
      const debuffNames = {
        poison: '中毒',
        burn: '灼烧',
        freeze: '冰冻',
        stun: '眩晕',
        bleed: '流血',
        silence: '沉默',
        fear: '恐惧'
      };
      this.uiManager.toast(`受到${debuffNames[debuffType] || debuffType}效果！`, 'warning', 2000);
    }
  }
  
  /**
   * 触发减益特效
   */
  triggerDebuffEffect(debuffType, x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    const debuffMap = {
      poison: 'poison',
      burn: 'burn',
      freeze: 'freeze',
      stun: 'stun',
      bleed: 'bleed',
      silence: 'silence',
      fear: 'fear'
    };
    
    const effectName = debuffMap[debuffType] || 'poison';
    this.effectManager.triggerStatusEffect(effectName, { x, y });
  }
  
  /**
   * 处理地图事件
   */
  handleMapEvent(data) {
    const eventType = data.event_type;
    const x = data.x || 0;
    const y = data.y || 0;
    
    // 获取位置
    const tileSize = this.mapEngine?.tileSize || 48;
    const screenX = x * tileSize + tileSize / 2;
    const screenY = y * tileSize + tileSize / 2;
    
    // 触发地图事件特效
    this.triggerMapEventEffect(eventType, screenX, screenY);
  }
  
  /**
   * 触发地图事件特效
   */
  triggerMapEventEffect(eventType, x, y) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager) return;
    
    const eventMap = {
      spawn: 'map_event_spawn',
      end: 'map_event_end',
      portal_open: 'portal_open',
      chest_open: 'chest_open'
    };
    
    const effectName = eventMap[eventType] || 'map_event_spawn';
    this.effectManager.triggerInteractionEffect(effectName, x, y);
  }
  
  // 处理地图玩家列表（服务器主动推送）
  handleMapPlayer(data) {
    if (!data.players || !Array.isArray(data.players)) return;
    
    console.log('收到地图玩家列表:', data.players);
    console.log('this.players:', this.players);
    
    data.players.forEach(player => {
      if (player.role_id === this.player.id) return; // 跳过自己
      
      const px = player.x || 10;
      const py = player.y || 10;
      const tileSize = this.mapEngine?.tileSize || 48;
      this.players.set(player.role_id, {
        id: player.role_id,
        name: player.name || `玩家${player.role_id}`,
        x: px,
        y: py,
        lastScreenX: px * tileSize, // 初始屏幕坐标
        lastScreenY: py * tileSize,
        hp: player.hp || 100,
        maxHp: player.maxHp || 100
      });
      console.log(`添加玩家 ${player.name} 到列表，位置 (${player.x}, ${player.y})`);
    });
  }
  
  handleEnterMap(data) {
    if (data.role_id === this.player.id) return; // 忽略自己
    
    console.log('玩家进入地图:', data);
    
    const tileSize = this.mapEngine?.tileSize || 48;
    this.players.set(data.role_id, {
      id: data.role_id,
      name: data.name || `玩家${data.role_id}`,
      x: data.x || 10,
      y: data.y || 10,
      lastScreenX: (data.x || 10) * tileSize, // 初始化屏幕位置
      lastScreenY: (data.y || 10) * tileSize,
      hp: data.hp || 100,
      maxHp: data.maxHp || 100
    });
    
    this.addChatMessage('系统', `${data.name || '某玩家'}进入了地图`, 'system');
  }
  
  handleLeaveMap(data) {
    this.players.delete(data.role_id);
    this.addChatMessage('系统', `${data.name || '玩家'}离开了地图`, 'system');
    
    // 不再调用 updateOnlineCount()，在线人数由服务端的 handleOnlineCount 消息控制
  }
  
  handleOnlineCount(data) {
    // 使用服务端广播的全局在线人数，而不是本地计算的
    if (data.count !== undefined && this.ui.onlineCount) {
      this.ui.onlineCount.textContent = data.count;
      console.log('更新在线人数:', data.count);
    }
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

    // F3: 切换地图瓦片调试模式（显示每个格子的瓦片ID和图集坐标）
    if (e.key === 'F3' && this.mapEngine && this.mapEngine.mapRenderer) {
      this.mapEngine.mapRenderer.debugMode = !this.mapEngine.mapRenderer.debugMode;
      console.log(`🔍 调试模式: ${this.mapEngine.mapRenderer.debugMode ? '开启 (按F3关闭)' : '关闭'}`);
      return;
    }

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
    // 检查地图碰撞
    if (this.currentMap?.tiles?.[newY]?.[newX] === 1) {
      return; // 地图障碍物碰撞
    }
    
    // 检查玩家间碰撞 - 目标位置是否有其他玩家
    for (const [id, player] of this.players) {
      if (player.x === newX && player.y === newY) {
        return; // 玩家碰撞
      }
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
    
    // 重绘（不带deltaTime，手动触发一次渲染）
    this.mapEngine.render(16.67); // 假设16.67ms帧时间
    this.renderMiniMap();
  }
  
  useSkill(skillId) {
    // 检查冷却
    const cooldownEnd = this.skillCooldowns.get(skillId) || 0;
    if (Date.now() < cooldownEnd) {
      // 显示冷却提示
      if (this.uiManager) {
        this.uiManager.toast('技能冷却中', 'warning', 1000);
      }
      return;
    }
    
    // 发送技能请求
    window.GameWS.send(Protocol.CMD_USE_SKILL, {
      skill_id: skillId,
      target_id: 0, // 默认无目标
      x: this.player.x,
      y: this.player.y
    });
    
    // 触发技能特效
    this.triggerSkillEffect(skillId);
    
    // 播放技能音效
    this.playSkillSound(skillId);
    
    // 设置冷却
    const cooldown = skillId === 0 ? 1000 : 3000;
    this.skillCooldowns.set(skillId, Date.now() + cooldown);
    
    // 更新UI
    this.updateSkillUI();
  }
  
  /**
   * 触发技能特效
   */
  triggerSkillEffect(skillId) {
    if (!this.effectSettings.enableSkillEffects) return;
    if (!this.effectManager || !this.particleSystem) return;
    
    const tileSize = this.mapEngine?.tileSize || 48;
    const x = this.player.x * tileSize + tileSize / 2;
    const y = this.player.y * tileSize + tileSize / 2;
    
    // 根据技能ID选择特效
    const effectMap = {
      0: 'sword_slash',   // 普通攻击
      1: 'fire_burst',    // 技能1 - 火焰
      2: 'ice_burst',     // 技能2 - 冰霜
      3: 'magic_circle',  // 技能3 - 魔法阵
      4: 'thunder_strike' // 技能4 - 雷霆
    };
    
    const effectName = effectMap[skillId] || 'hit_spark';
    
    if (this.effectSettings.enableScreenShake && skillId >= 2) {
      // 高级技能触发屏幕震动
      this.effectManager.triggerSkillEffect(effectName, { x, y }, { x, y }, {
        shake: skillId >= 3,
        shakeIntensity: skillId === 4 ? 8 : 3
      });
    } else {
      this.effectManager.triggerSkillEffect(effectName, { x, y }, { x, y }, {
        shake: false
      });
    }
  }
  
  /**
   * 播放技能音效
   */
  playSkillSound(skillId) {
    if (!this.soundManager) return;
    
    const soundMap = {
      0: 'sword_swing',
      1: 'fire_cast',
      2: 'ice_cast',
      3: 'heal',
      4: 'thunder'
    };
    
    const soundId = soundMap[skillId] || 'punch';
    this.soundManager.play(soundId, {
      volume: 0.6,
      pan: 0
    });
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
    let lastTime = performance.now();
    
    const loop = (currentTime) => {
      const deltaTime = currentTime - lastTime;
      lastTime = currentTime;
      
      if (this.state === 'playing') {
        // 更新技能冷却UI
        this.updateSkillUI();
        
        // 更新特效系统
        if (this.effectSettings.enableParticles && this.effectManager) {
          this.effectManager.update(deltaTime);
        }
      }
      
      requestAnimationFrame(loop);
    };
    
    loop(performance.now());
  }
  
  // 更新其他玩家的平滑移动（基于时间）
  updateOtherPlayers(deltaTime) {
    // 确保deltaTime有效
    if (!deltaTime || deltaTime < 0 || deltaTime > 1000) {
      deltaTime = 16.67; // 默认60fps帧时间
    }
    
    const tileSize = this.mapEngine?.tileSize || 48;
    const baseSpeed = 6; // 基准速度：6像素/16.67ms (60fps)
    const moveDistance = (baseSpeed * deltaTime) / (1000 / 60);
    
    // 获取当前玩家位置（用于视野判断）
    const myX = this.player.x;
    const myY = this.player.y;
    
    // 计算视野范围（屏幕能显示的格子数 + 额外缓冲区）
    const canvas = this.mapEngine?.canvas;
    // 以屏幕大小为基准计算视野，加上1倍屏幕大小的缓冲区
    // 这样可以看到屏幕外1倍距离的玩家，实现平滑的进出视野
    const bufferMultiplier = 1.0; // 缓冲区倍数（1.0表示1倍屏幕大小）
    const viewWidth = canvas ? Math.ceil(canvas.width / tileSize) * (1 + bufferMultiplier) : 40;
    const viewHeight = canvas ? Math.ceil(canvas.height / tileSize) * (1 + bufferMultiplier) : 30;
    
    // 视野边界
    const minX = myX - viewWidth / 2;
    const maxX = myX + viewWidth / 2;
    const minY = myY - viewHeight / 2;
    const maxY = myY + viewHeight / 2;
    
    this.players.forEach(player => {
      // 计算玩家是否在视野内
      const inView = player.x >= minX && player.x <= maxX && player.y >= minY && player.y <= maxY;
      
      if (!inView) {
        // 视野外玩家：保持目标位置更新，但不计算插值
        // 下次进入视野时会直接显示在正确位置
        return;
      }
      
      // 视野内玩家：检查是否需要快速追赶（进入视野时位置已偏差）
      if (player.lastScreenX === undefined) {
        player.lastScreenX = player.x * tileSize;
        player.lastScreenY = player.y * tileSize;
      } else {
        // 检查显示位置与实际位置的偏差
        const targetScreenX = player.x * tileSize;
        const targetScreenY = player.y * tileSize;
        const offset = Math.hypot(targetScreenX - player.lastScreenX, targetScreenY - player.lastScreenY);
        
        // 如果偏差超过3个格子，快速跳转避免长时间追赶
        if (offset > tileSize * 3) {
          player.lastScreenX = targetScreenX;
          player.lastScreenY = targetScreenY;
          return;
        }
      }
      
      const targetScreenX = player.x * tileSize;
      const targetScreenY = player.y * tileSize;
      const dx = targetScreenX - player.lastScreenX;
      const dy = targetScreenY - player.lastScreenY;
      const dist = Math.hypot(dx, dy);
      
      if (dist <= moveDistance) {
        // 到达目标
        player.lastScreenX = targetScreenX;
        player.lastScreenY = targetScreenY;
      } else {
        // 继续移动
        player.lastScreenX += (dx / dist) * moveDistance;
        player.lastScreenY += (dy / dist) * moveDistance;
      }
    });
  }
  
  renderPlayers() {
    // 获取地图画布
    const canvas = this.mapEngine?.canvas;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    
    // 渲染特效系统
    if (this.effectSettings.enableParticles && this.effectManager) {
      this.effectManager.render(ctx);
    }
    
    // 获取当前玩家位置和视野范围（用于视野裁剪）
    const tileSize = this.mapEngine.tileSize || 48;
    const myX = this.player.x;
    const myY = this.player.y;
    // 与updateOtherPlayers保持一致的视野计算
    const bufferMultiplier = 1.0;
    const viewWidth = Math.ceil(canvas.width / tileSize) * (1 + bufferMultiplier);
    const viewHeight = Math.ceil(canvas.height / tileSize) * (1 + bufferMultiplier);
    const minX = myX - viewWidth / 2;
    const maxX = myX + viewWidth / 2;
    const minY = myY - viewHeight / 2;
    const maxY = myY + viewHeight / 2;
    
    // 绘制其他玩家（仅视野范围内）
    this.players.forEach(player => {
      // 视野裁剪：跳过视野外的玩家，减少绘制
      if (player.x < minX || player.x > maxX || player.y < minY || player.y > maxY) {
        return;
      }
      
      // 直接使用updateOtherPlayers计算好的屏幕位置
      let screenX, screenY;
      if (player.lastScreenX !== undefined && player.lastScreenY !== undefined) {
        screenX = player.lastScreenX;
        screenY = player.lastScreenY;
      } else {
        screenX = player.x * tileSize;
        screenY = player.y * tileSize;
      }
      
      // 简单绘制为绿色圆形
      ctx.fillStyle = '#4ade80';
      ctx.beginPath();
      ctx.arc(screenX + tileSize / 2, screenY + tileSize / 2, tileSize / 3, 0, Math.PI * 2);
      ctx.fill();
      
      // 名字
      ctx.fillStyle = '#fff';
      ctx.font = '12px Microsoft YaHei';
      ctx.textAlign = 'center';
      ctx.fillText(player.name, screenX + tileSize / 2, screenY - 5);
    });
  }
  
  /**
   * 退出游戏：返回登录界面
   */
  logout() {
    console.log('玩家退出游戏');
    
    // 发送退出消息给服务器
    if (window.GameWS) {
      window.GameWS.send(Protocol.CMD_LOGOUT, {});
    }
    
    // 清理资源
    this.destroy();
    
    // 重置状态
    this.state = 'login';
    
    // 显示登录面板，隐藏游戏面板
    this.ui.gamePanel.classList.add('hidden');
    this.ui.loginPanel.classList.remove('hidden');
    
    // 重置玩家数据
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
    
    // 清空其他玩家列表
    this.players.clear();
    
    console.log('退出游戏完成');
  }
  
  /**
   * 销毁方法：清理定时器和资源，防止内存泄漏
   */
  destroy() {
    // 清除小地图更新定时器
    if (this.minimapUpdateTimer) {
      clearTimeout(this.minimapUpdateTimer);
      this.minimapUpdateTimer = null;
    }
    
    // 清除其他可能存在的定时器
    if (this.fpsUpdateTimer) {
      clearInterval(this.fpsUpdateTimer);
      this.fpsUpdateTimer = null;
    }
    
    // 清理特效系统
    if (this.effectManager) {
      this.effectManager.clearAll();
    }
    
    if (this.particleSystem) {
      this.particleSystem.clearAll();
    }
    
    // 清理音效系统
    if (this.soundManager) {
      this.soundManager.stopAll();
      this.soundManager.stopBGM(0);
    }
    
    // 清理UI系统
    if (this.uiManager) {
      this.uiManager.clearAll();
    }
    
    // 清理地图引擎
    if (this.mapEngine && typeof this.mapEngine.destroy === 'function') {
      this.mapEngine.destroy();
    }
    
    // 移除所有事件监听器
    this.removeAllEventListeners();
    
    // 清理 WebSocket 连接
    if (window.GameWS) {
      window.GameWS.close();
    }
    
    console.log('Game instance destroyed');
  }
}

// 协议定义（暴露到全局供其他模块使用）
const Protocol = window.Protocol = {
  // 登录相关 1001-1010
  CMD_LOGIN: 1001,
  CMD_LOGOUT: 1002,
  CMD_HEARTBEAT: 1003,
  CMD_REGISTER: 1004,
  
  // 游戏相关 2001-2030
  CMD_MOVE: 2001,
  CMD_ATTACK: 2002,
  CMD_USE_SKILL: 2003,
  CMD_CHAT: 2004,
  CMD_PICKUP: 2005,
  CMD_USE_ITEM: 2006,
  CMD_EQUIP: 2007,
  CMD_TRADE: 2008,
  CMD_DAMAGE: 2009,
  CMD_DEATH: 2010,
  CMD_RESPAWN: 2011,
  CMD_LEVEL_UP: 2012,
  CMD_BUFF: 2013,
  CMD_DEBUFF: 2014,
  
  // 地图相关 3001-3050
  CMD_ENTER_MAP: 3001,
  CMD_LEAVE_MAP: 3002,
  CMD_MAP_PLAYER: 3003,
  CMD_ONLINE_COUNT: 3004,
  CMD_NPC_TALK: 3005,
  CMD_NPC_TRADE: 3006,
  CMD_MAP_EVENT: 3007,
  
  // 武学相关 4001-4020
  CMD_SKILL_LEARN: 4001,
  CMD_SKILL_UPGRADE: 4002,
  
  // 角色相关 5001-5020
  CMD_ROLE_INFO: 5001,
  CMD_ROLE_ATTRIB: 5002,
  CMD_SYNC: 5003,
  CMD_ROLE_LIST: 5004,
  CMD_ROLE_CREATE: 5005,
  CMD_ROLE_SELECT: 5006
};

// 游戏单例
window.game = new Game();
