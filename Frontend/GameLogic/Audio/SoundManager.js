/**
 * 音效管理器
 * 统一管理游戏中的所有音效：背景音乐、技能音效、环境音效等
 */
class SoundManager {
  constructor() {
    // 音频上下文
    this.audioContext = null;
    
    // 音效库
    this.sounds = new Map();
    
    // 背景音乐
    this.bgm = {
      current: null,
      volume: 0.5,
      muted: false,
      playlist: [],
      currentIndex: 0,
      loop: true
    };
    
    // 音效设置
    this.sfx = {
      volume: 0.7,
      muted: false,
      maxConcurrent: 10  // 最大同时播放音效数
    };
    
    // 当前播放的音效
    this.playingSounds = new Map();
    
    // 音效预加载队列
    this.loadQueue = [];
    this.loading = false;
    
    // 音效配置
    this.soundConfigs = new Map();
    
    // 初始化
    this.init();
    
    console.log('音效管理器初始化完成');
  }
  
  /**
   * 初始化音频上下文
   */
  async init() {
    try {
      this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
      
      // 创建主音量节点
      this.masterGain = this.audioContext.createGain();
      this.masterGain.connect(this.audioContext.destination);
      
      // 创建BGM音量节点
      this.bgmGain = this.audioContext.createGain();
      this.bgmGain.gain.value = this.bgm.volume;
      this.bgmGain.connect(this.masterGain);
      
      // 创建音效音量节点
      this.sfxGain = this.audioContext.createGain();
      this.sfxGain.gain.value = this.sfx.volume;
      this.sfxGain.connect(this.masterGain);
      
      // 初始化音效配置
      this.initSoundConfigs();
      
      console.log('音频上下文初始化成功');
    } catch (error) {
      console.error('音频上下文初始化失败:', error);
    }
  }
  
  /**
   * 初始化音效配置
   */
  initSoundConfigs() {
    // ==================== 技能音效 ====================
    
    // 剑法音效
    this.soundConfigs.set('sword_swing', {
      category: 'skill',
      volume: 0.6,
      pitchVariation: 0.1,
      cooldown: 100
    });
    
    this.soundConfigs.set('sword_combo', {
      category: 'skill',
      volume: 0.7,
      pitchVariation: 0.15,
      cooldown: 50
    });
    
    this.soundConfigs.set('sword_qi', {
      category: 'skill',
      volume: 0.8,
      pitchVariation: 0.05,
      cooldown: 200
    });
    
    // 拳法音效
    this.soundConfigs.set('punch', {
      category: 'skill',
      volume: 0.5,
      pitchVariation: 0.2,
      cooldown: 100
    });
    
    this.soundConfigs.set('heavy_punch', {
      category: 'skill',
      volume: 0.8,
      pitchVariation: 0.1,
      cooldown: 200
    });
    
    // 内功音效
    this.soundConfigs.set('heal', {
      category: 'skill',
      volume: 0.6,
      loop: false,
      fadeIn: 0.2,
      fadeOut: 0.3
    });
    
    this.soundConfigs.set('shield', {
      category: 'skill',
      volume: 0.5,
      fadeIn: 0.1,
      fadeOut: 0.2
    });
    
    // 火系音效
    this.soundConfigs.set('fire_cast', {
      category: 'skill',
      volume: 0.7,
      pitchVariation: 0.1
    });
    
    this.soundConfigs.set('fire_wave', {
      category: 'skill',
      volume: 0.8,
      duration: 0.5
    });
    
    // 冰系音效
    this.soundConfigs.set('ice_cast', {
      category: 'skill',
      volume: 0.6,
      pitchVariation: 0.05
    });
    
    this.soundConfigs.set('ice_shield', {
      category: 'skill',
      volume: 0.5,
      fadeIn: 0.1
    });
    
    // 雷系音效
    this.soundConfigs.set('thunder', {
      category: 'skill',
      volume: 1.0,
      shake: true
    });
    
    // ==================== 受击音效 ====================
    
    this.soundConfigs.set('hit_normal', {
      category: 'combat',
      volume: 0.4,
      pitchVariation: 0.3,
      cooldown: 50
    });
    
    this.soundConfigs.set('hit_heavy', {
      category: 'combat',
      volume: 0.6,
      pitchVariation: 0.2,
      cooldown: 100
    });
    
    this.soundConfigs.set('hit_critical', {
      category: 'combat',
      volume: 0.8,
      pitchVariation: 0.1
    });
    
    this.soundConfigs.set('block', {
      category: 'combat',
      volume: 0.3,
      pitchVariation: 0.1
    });
    
    this.soundConfigs.set('dodge', {
      category: 'combat',
      volume: 0.2
    });
    
    // ==================== 状态音效 ====================
    
    this.soundConfigs.set('poison', {
      category: 'status',
      volume: 0.3,
      loop: true,
      interval: 1000
    });
    
    this.soundConfigs.set('freeze', {
      category: 'status',
      volume: 0.4
    });
    
    this.soundConfigs.set('stun', {
      category: 'status',
      volume: 0.3
    });
    
    // ==================== 交互音效 ====================
    
    this.soundConfigs.set('pickup', {
      category: 'interaction',
      volume: 0.5,
      pitchVariation: 0.2
    });
    
    this.soundConfigs.set('level_up', {
      category: 'interaction',
      volume: 0.8,
      fadeIn: 0.1
    });
    
    this.soundConfigs.set('death', {
      category: 'interaction',
      volume: 0.6,
      fadeOut: 0.5
    });
    
    this.soundConfigs.set('teleport', {
      category: 'interaction',
      volume: 0.5,
      fadeIn: 0.2,
      fadeOut: 0.2
    });
    
    this.soundConfigs.set('button_click', {
      category: 'ui',
      volume: 0.3
    });
    
    this.soundConfigs.set('menu_open', {
      category: 'ui',
      volume: 0.4
    });
    
    this.soundConfigs.set('menu_close', {
      category: 'ui',
      volume: 0.3
    });
    
    // ==================== 环境音效 ====================
    
    this.soundConfigs.set('ambient_forest', {
      category: 'ambient',
      volume: 0.3,
      loop: true,
      fadeIn: 1.0,
      fadeOut: 1.0
    });
    
    this.soundConfigs.set('ambient_water', {
      category: 'ambient',
      volume: 0.4,
      loop: true,
      fadeIn: 0.5
    });
    
    this.soundConfigs.set('ambient_wind', {
      category: 'ambient',
      volume: 0.2,
      loop: true
    });
    
    this.soundConfigs.set('ambient_rain', {
      category: 'ambient',
      volume: 0.5,
      loop: true,
      fadeIn: 0.5
    });
    
    this.soundConfigs.set('ambient_fire', {
      category: 'ambient',
      volume: 0.3,
      loop: true
    });
    
    // ==================== 背景音乐 ====================
    
    this.soundConfigs.set('bgm_main', {
      category: 'bgm',
      volume: 0.5,
      loop: true,
      fadeIn: 1.0,
      fadeOut: 1.0
    });
    
    this.soundConfigs.set('bgm_combat', {
      category: 'bgm',
      volume: 0.6,
      loop: true,
      fadeIn: 0.5,
      fadeOut: 0.5
    });
    
    this.soundConfigs.set('bgm_town', {
      category: 'bgm',
      volume: 0.4,
      loop: true,
      fadeIn: 1.0
    });
    
    this.soundConfigs.set('bgm_dungeon', {
      category: 'bgm',
      volume: 0.5,
      loop: true,
      fadeIn: 0.5
    });
    
    this.soundConfigs.set('bgm_victory', {
      category: 'bgm',
      volume: 0.7,
      loop: false,
      fadeIn: 0.2
    });
  }
  
  /**
   * 加载音效文件
   */
  async loadSound(soundId, url) {
    await this.ensureContext();
    
    try {
      const response = await fetch(url);
      const arrayBuffer = await response.arrayBuffer();
      const audioBuffer = await this.audioContext.decodeAudioData(arrayBuffer);
      
      this.sounds.set(soundId, {
        id: soundId,
        buffer: audioBuffer,
        url: url,
        duration: audioBuffer.duration,
        config: this.soundConfigs.get(soundId) || {}
      });
      
      console.log(`音效加载成功: ${soundId}`);
      return true;
    } catch (error) {
      console.error(`音效加载失败 ${soundId}:`, error);
      return false;
    }
  }
  
  /**
   * 批量加载音效
   */
  async loadSounds(soundList) {
    const promises = soundList.map(sound => this.loadSound(sound.id, sound.url));
    const results = await Promise.all(promises);
    const successCount = results.filter(r => r).length;
    console.log(`批量加载音效完成: ${successCount}/${soundList.length}`);
    return successCount;
  }
  
  /**
   * 预加载音效队列
   */
  addToLoadQueue(soundId, url) {
    this.loadQueue.push({ id: soundId, url: url });
    this.processLoadQueue();
  }
  
  /**
   * 处理加载队列
   */
  async processLoadQueue() {
    if (this.loading || this.loadQueue.length === 0) return;
    
    this.loading = true;
    
    while (this.loadQueue.length > 0) {
      const item = this.loadQueue.shift();
      await this.loadSound(item.id, item.url);
    }
    
    this.loading = false;
  }
  
  /**
   * 确保音频上下文已初始化
   */
  async ensureContext() {
    if (!this.audioContext) {
      await this.init();
    }
    
    // 恢复暂停的上下文
    if (this.audioContext.state === 'suspended') {
      await this.audioContext.resume();
    }
  }
  
  /**
   * 播放音效
   */
  play(soundId, options = {}) {
    if (this.sfx.muted) return null;
    
    const sound = this.sounds.get(soundId);
    if (!sound) {
      // 尝试使用程序化音效
      return this.playProceduralSound(soundId, options);
    }
    
    const config = Object.assign({}, sound.config, options);
    
    // 检查冷却时间
    if (config.cooldown && this.checkCooldown(soundId, config.cooldown)) {
      return null;
    }
    
    // 检查同时播放数量限制
    if (this.playingSounds.size >= this.sfx.maxConcurrent) {
      this.stopOldestSound();
    }
    
    return this.playBuffer(sound.buffer, config, soundId);
  }
  
  /**
   * 播放音频缓冲
   */
  playBuffer(buffer, config, soundId) {
    const source = this.audioContext.createBufferSource();
    source.buffer = buffer;
    
    // 音量控制
    const gainNode = this.audioContext.createGain();
    const volume = (config.volume || 1.0) * this.sfx.volume;
    
    // 淡入效果
    if (config.fadeIn > 0) {
      gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
      gainNode.gain.linearRampToValueAtTime(volume, this.audioContext.currentTime + config.fadeIn);
    } else {
      gainNode.gain.value = volume;
    }
    
    // 淡出效果
    if (config.fadeOut > 0) {
      const fadeOutTime = this.audioContext.currentTime + buffer.duration - config.fadeOut;
      gainNode.gain.setValueAtTime(volume, fadeOutTime);
      gainNode.gain.linearRampToValueAtTime(0, fadeOutTime + config.fadeOut);
    }
    
    // 声像控制（空间音效）
    const panNode = this.audioContext.createStereoPanner();
    panNode.pan.value = config.pan || 0;
    
    // 播放速度（音调变化）
    if (config.pitchVariation) {
      source.playbackRate.value = 1 + (Math.random() - 0.5) * config.pitchVariation * 2;
    } else if (config.pitch) {
      source.playbackRate.value = config.pitch;
    }
    
    // 循环播放
    source.loop = config.loop || false;
    
    // 连接节点
    source.connect(gainNode);
    gainNode.connect(panNode);
    panNode.connect(this.sfxGain);
    
    // 播放
    source.start(0);
    
    // 记录播放状态
    const playId = `${soundId}_${Date.now()}`;
    this.playingSounds.set(playId, {
      source: source,
      gainNode: gainNode,
      panNode: panNode,
      soundId: soundId,
      startTime: Date.now()
    });
    
    // 播放完成后清理
    source.onended = () => {
      this.playingSounds.delete(playId);
    };
    
    return { source, gainNode, panNode, playId };
  }
  
  /**
   * 播放程序化音效
   */
  playProceduralSound(type, options = {}) {
    if (!this.audioContext) return null;
    
    const buffer = this.createProceduralBuffer(type, options.duration || 0.5);
    if (!buffer) return null;
    
    // 存储程序化音效
    const proceduralId = `procedural_${type}`;
    this.sounds.set(proceduralId, {
      id: proceduralId,
      buffer: buffer,
      duration: buffer.duration,
      config: { volume: options.volume || 0.5 }
    });
    
    return this.play(proceduralId, options);
  }
  
  /**
   * 创建程序化音频缓冲
   */
  createProceduralBuffer(type, duration) {
    if (!this.audioContext) return null;
    
    const sampleRate = this.audioContext.sampleRate;
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate);
    const data = buffer.getChannelData(0);
    
    switch (type) {
      case 'sword_swing':
        this.createSwooshSound(data, sampleRate, 200, 800);
        break;
      case 'punch':
        this.createImpactSound(data, sampleRate, 0.3);
        break;
      case 'hit_normal':
        this.createImpactSound(data, sampleRate, 0.2);
        break;
      case 'hit_heavy':
        this.createImpactSound(data, sampleRate, 0.4);
        break;
      case 'fire_cast':
        this.createFireSound(data, sampleRate);
        break;
      case 'ice_cast':
        this.createIceSound(data, sampleRate);
        break;
      case 'thunder':
        this.createThunderSound(data, sampleRate);
        break;
      case 'heal':
        this.createHealSound(data, sampleRate);
        break;
      case 'pickup':
        this.createPickupSound(data, sampleRate);
        break;
      case 'level_up':
        this.createLevelUpSound(data, sampleRate);
        break;
      case 'button_click':
        this.createClickSound(data, sampleRate);
        break;
      default:
        this.createWhiteNoise(data, sampleRate);
    }
    
    return buffer;
  }
  
  /**
   * 创建挥动音效
   */
  createSwooshSound(data, sampleRate, startFreq, endFreq) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const progress = t / (data.length / sampleRate);
      const freq = startFreq + (endFreq - startFreq) * progress;
      const envelope = Math.exp(-progress * 3);
      data[i] = Math.sin(2 * Math.PI * freq * t) * envelope * 0.5;
    }
  }
  
  /**
   * 创建撞击音效
   */
  createImpactSound(data, sampleRate, intensity) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const envelope = Math.exp(-t * 10);
      const noise = (Math.random() - 0.5) * intensity;
      const tone = Math.sin(2 * Math.PI * 100 * t) * intensity * 0.5;
      data[i] = (noise + tone) * envelope;
    }
  }
  
  /**
   * 创建火焰音效
   */
  createFireSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const crackle = Math.random() > 0.95 ? Math.random() * 0.3 : 0;
      const roar = (Math.random() - 0.5) * 0.2;
      const envelope = Math.exp(-t * 2);
      data[i] = (roar + crackle) * envelope;
    }
  }
  
  /**
   * 创建冰霜音效
   */
  createIceSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const baseFreq = 800;
      const harmonic1 = Math.sin(2 * Math.PI * baseFreq * t) * 0.2;
      const harmonic2 = Math.sin(2 * Math.PI * baseFreq * 2 * t) * 0.1;
      const shimmer = Math.random() > 0.9 ? Math.random() * 0.1 : 0;
      const envelope = Math.exp(-t * 3);
      data[i] = (harmonic1 + harmonic2 + shimmer) * envelope;
    }
  }
  
  /**
   * 创建雷电音效
   */
  createThunderSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const noise = (Math.random() - 0.5) * 0.8;
      const rumble = Math.sin(2 * Math.PI * 50 * t) * 0.5;
      const envelope = Math.exp(-t * 5);
      data[i] = (noise + rumble) * envelope;
    }
  }
  
  /**
   * 创建治愈音效
   */
  createHealSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const baseFreq = 523.25; // C5
      const harmonic1 = Math.sin(2 * Math.PI * baseFreq * t) * 0.3;
      const harmonic2 = Math.sin(2 * Math.PI * baseFreq * 1.5 * t) * 0.2;
      const harmonic3 = Math.sin(2 * Math.PI * baseFreq * 2 * t) * 0.1;
      const envelope = Math.sin(t * Math.PI / (data.length / sampleRate));
      data[i] = (harmonic1 + harmonic2 + harmonic3) * envelope;
    }
  }
  
  /**
   * 创建拾取音效
   */
  createPickupSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const freq = 600 + t * 400;
      const envelope = Math.exp(-t * 4);
      data[i] = Math.sin(2 * Math.PI * freq * t) * envelope * 0.4;
    }
  }
  
  /**
   * 创建升级音效
   */
  createLevelUpSound(data, sampleRate) {
    const notes = [523.25, 659.25, 783.99, 1046.50]; // C5, E5, G5, C6
    const noteDuration = data.length / sampleRate / notes.length;
    
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const noteIndex = Math.floor(t / noteDuration);
      const freq = notes[Math.min(noteIndex, notes.length - 1)];
      const localT = t - noteIndex * noteDuration;
      const envelope = Math.exp(-localT * 2);
      data[i] = Math.sin(2 * Math.PI * freq * t) * envelope * 0.5;
    }
  }
  
  /**
   * 创建点击音效
   */
  createClickSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const envelope = Math.exp(-t * 20);
      data[i] = Math.sin(2 * Math.PI * 1000 * t) * envelope * 0.3;
    }
  }
  
  /**
   * 创建白噪音
   */
  createWhiteNoise(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      data[i] = (Math.random() - 0.5) * 0.3;
    }
  }
  
  /**
   * 检查冷却时间
   */
  checkCooldown(soundId, cooldown) {
    const lastPlay = this.playingSounds.get(soundId);
    if (!lastPlay) return false;
    
    return Date.now() - lastPlay.startTime < cooldown;
  }
  
  /**
   * 停止最旧的音效
   */
  stopOldestSound() {
    let oldest = null;
    let oldestTime = Date.now();
    
    this.playingSounds.forEach((sound, id) => {
      if (sound.startTime < oldestTime) {
        oldest = id;
        oldestTime = sound.startTime;
      }
    });
    
    if (oldest) {
      this.stop(oldest);
    }
  }
  
  /**
   * 停止音效
   */
  stop(playId) {
    const sound = this.playingSounds.get(playId);
    if (sound && sound.source) {
      try {
        sound.source.stop();
        this.playingSounds.delete(playId);
      } catch (error) {
        console.error(`停止音效失败 ${playId}:`, error);
      }
    }
  }
  
  /**
   * ==================== 背景音乐管理 ====================
   */
  
  /**
   * 播放背景音乐
   */
  async playBGM(bgmId, options = {}) {
    if (this.bgm.muted) return;
    
    await this.ensureContext();
    
    const sound = this.sounds.get(bgmId);
    if (!sound) {
      console.warn(`背景音乐未找到: ${bgmId}`);
      return;
    }
    
    const config = Object.assign({}, sound.config, options);
    
    // 停止当前BGM
    if (this.bgm.current) {
      await this.stopBGM(config.fadeOut || 0.5);
    }
    
    // 创建新的BGM源
    const source = this.audioContext.createBufferSource();
    source.buffer = sound.buffer;
    source.loop = config.loop !== false;
    
    // 音量控制
    const gainNode = this.audioContext.createGain();
    const volume = (config.volume || 1.0) * this.bgm.volume;
    
    // 淡入
    if (config.fadeIn > 0) {
      gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
      gainNode.gain.linearRampToValueAtTime(volume, this.audioContext.currentTime + config.fadeIn);
    } else {
      gainNode.gain.value = volume;
    }
    
    // 连接
    source.connect(gainNode);
    gainNode.connect(this.bgmGain);
    
    // 播放
    source.start(0);
    
    this.bgm.current = {
      source: source,
      gainNode: gainNode,
      bgmId: bgmId,
      config: config
    };
    
    console.log(`播放背景音乐: ${bgmId}`);
  }
  
  /**
   * 停止背景音乐
   */
  async stopBGM(fadeOut = 0.5) {
    if (!this.bgm.current) return;
    
    const { source, gainNode } = this.bgm.current;
    
    if (fadeOut > 0) {
      gainNode.gain.linearRampToValueAtTime(0, this.audioContext.currentTime + fadeOut);
      setTimeout(() => {
        try {
          source.stop();
        } catch (e) {}
      }, fadeOut * 1000);
    } else {
      try {
        source.stop();
      } catch (e) {}
    }
    
    this.bgm.current = null;
  }
  
  /**
   * 设置BGM音量
   */
  setBGMVolume(volume) {
    this.bgm.volume = Math.max(0, Math.min(1, volume));
    
    if (this.bgmGain) {
      this.bgmGain.gain.value = this.bgm.volume;
    }
    
    if (this.bgm.current) {
      this.bgm.current.gainNode.gain.value = this.bgm.volume;
    }
  }

  /**
   * 获取BGM音量
   */
  getBGMVolume() {
    return this.bgm.volume;
  }
  
  /**
   * 切换BGM静音
   */
  toggleBGMMute() {
    this.bgm.muted = !this.bgm.muted;
    
    if (this.bgmGain) {
      this.bgmGain.gain.value = this.bgm.muted ? 0 : this.bgm.volume;
    }
    
    return this.bgm.muted;
  }

  /**
   * 获取BGM静音状态
   */
  isBGMMuted() {
    return this.bgm.muted;
  }

  /**
   * 设置BGM静音状态
   * @param {boolean} muted - 是否静音
   */
  setBGMMute(muted) {
    this.bgm.muted = muted;
    
    if (this.bgmGain) {
      this.bgmGain.gain.value = muted ? 0 : this.bgm.volume;
    }
  }
  
  /**
   * ==================== 音效设置 ====================
   */
  
  /**
   * 设置音效音量
   */
  setSFXVolume(volume) {
    this.sfx.volume = Math.max(0, Math.min(1, volume));
    
    if (this.sfxGain) {
      this.sfxGain.gain.value = this.sfx.volume;
    }
  }

  /**
   * 获取音效音量
   */
  getSFXVolume() {
    return this.sfx.volume;
  }
  
  /**
   * 切换音效静音
   */
  toggleSFXMute() {
    this.sfx.muted = !this.sfx.muted;
    
    if (this.sfxGain) {
      this.sfxGain.gain.value = this.sfx.muted ? 0 : this.sfx.volume;
    }
    
    return this.sfx.muted;
  }

  /**
   * 获取音效静音状态
   */
  isSFXMuted() {
    return this.sfx.muted;
  }

  /**
   * 设置音效静音状态
   * @param {boolean} muted - 是否静音
   */
  setSFXMute(muted) {
    this.sfx.muted = muted;
    
    if (this.sfxGain) {
      this.sfxGain.gain.value = muted ? 0 : this.sfx.volume;
    }
  }
  
  /**
   * 设置主音量
   */
  setMasterVolume(volume) {
    if (this.masterGain) {
      this.masterGain.gain.value = Math.max(0, Math.min(1, volume));
    }
  }
  
  /**
   * 停止所有音效
   */
  stopAll() {
    this.playingSounds.forEach((sound, id) => {
      this.stop(id);
    });
    
    this.playingSounds.clear();
  }
  
  /**
   * ==================== 空间音效 ====================
   */
  
  /**
   * 根据位置播放音效
   */
  playAtPosition(soundId, x, y, canvasWidth, canvasHeight, options = {}) {
    // 计算声像位置（-1到1）
    const pan = (x - canvasWidth / 2) / (canvasWidth / 2);
    
    // 计算距离衰减
    const centerX = canvasWidth / 2;
    const centerY = canvasHeight / 2;
    const distance = Math.sqrt((x - centerX) ** 2 + (y - centerY) ** 2);
    const maxDistance = Math.sqrt(centerX ** 2 + centerY ** 2);
    const distanceVolume = 1 - (distance / maxDistance) * 0.5;
    
    return this.play(soundId, {
      ...options,
      pan: pan,
      volume: (options.volume || 1) * distanceVolume
    });
  }
  
  /**
   * ==================== 状态查询 ====================
   */
  
  /**
   * 获取音效状态
   */
  getStatus() {
    return {
      contextState: this.audioContext?.state || 'not initialized',
      loadedSounds: this.sounds.size,
      playingSounds: this.playingSounds.size,
      bgmPlaying: this.bgm.current !== null,
      bgmVolume: this.bgm.volume,
      bgmMuted: this.bgm.muted,
      sfxVolume: this.sfx.volume,
      sfxMuted: this.sfx.muted
    };
  }
  
  /**
   * 检查音效是否已加载
   */
  isLoaded(soundId) {
    return this.sounds.has(soundId);
  }
  
  /**
   * 获取音效时长
   */
  getDuration(soundId) {
    const sound = this.sounds.get(soundId);
    return sound ? sound.duration : 0;
  }
}

// 创建全局单例
window.SoundManager = new SoundManager();