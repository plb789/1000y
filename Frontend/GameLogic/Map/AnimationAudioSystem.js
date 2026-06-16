/**
 * 动画音效同步系统
 * 支持动画帧与音效的精确同步
 */
class AnimationAudioSystem {
  constructor() {
    this.audioContext = null;
    this.soundEffects = new Map(); // 存储音效
    this.animationSounds = new Map(); // 动画与音效的映射
    this.isPlaying = new Map(); // 当前播放的音效 {playId: {source, gainNode, panNode}}
    this.volume = 0.7; // 主音量
    this.muted = false; // 是否静音
  }
  
  /**
   * 初始化音频上下文
   */
  async init() {
    if (this.audioContext) return;
    
    try {
      this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
      console.log('音频系统初始化成功');
    } catch (error) {
      console.error('音频系统初始化失败:', error);
    }
  }
  
  /**
   * 加载音效文件
   */
  async loadSound(soundId, url) {
    await this.init();
    
    try {
      const response = await fetch(url);
      const arrayBuffer = await response.arrayBuffer();
      const audioBuffer = await this.audioContext.decodeAudioData(arrayBuffer);
      
      this.soundEffects.set(soundId, {
        id: soundId,
        buffer: audioBuffer,
        url: url,
        duration: audioBuffer.duration
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
    const promises = soundList.map(sound => 
      this.loadSound(sound.id, sound.url)
    );
    
    const results = await Promise.all(promises);
    const successCount = results.filter(r => r).length;
    
    console.log(`批量加载音效完成: ${successCount}/${soundList.length}`);
    return successCount;
  }
  
  /**
   * 绑定动画与音效
   */
  bindAnimationSound(animationId, soundConfig) {
    this.animationSounds.set(animationId, {
      animationId,
      soundId: soundConfig.soundId,
      triggerFrame: soundConfig.triggerFrame || 0, // 触发帧
      triggerInterval: soundConfig.triggerInterval || null, // 触发间隔（帧）
      loop: soundConfig.loop || false,
      volume: soundConfig.volume || 1.0,
      pitch: soundConfig.pitch || 1.0,
      pan: soundConfig.pan || 0, // 声像 (-1 到 1)
      fadeOut: soundConfig.fadeOut || 0,
      fadeIn: soundConfig.fadeIn || 0
    });
  }
  
  /**
   * 播放音效
   */
  playSound(soundId, options = {}) {
    if (this.muted) return null;
    
    const sound = this.soundEffects.get(soundId);
    if (!sound || !this.audioContext) {
      console.warn(`音效未找到或音频系统未初始化: ${soundId}`);
      return null;
    }
    
    const source = this.audioContext.createBufferSource();
    source.buffer = sound.buffer;
    
    // 音量控制
    const gainNode = this.audioContext.createGain();
    const volume = (options.volume !== undefined ? options.volume : 1.0) * this.volume;
    gainNode.gain.value = volume;
    
    // 声像控制
    const panNode = this.audioContext.createStereoPanner();
    panNode.pan.value = options.pan !== undefined ? options.pan : 0;
    
    // 播放速度（音调）
    source.playbackRate.value = options.pitch !== undefined ? options.pitch : 1.0;
    
    // 连接节点
    source.connect(gainNode);
    gainNode.connect(panNode);
    panNode.connect(this.audioContext.destination);
    
    // 淡入效果
    if (options.fadeIn > 0) {
      gainNode.gain.setValueAtTime(0, this.audioContext.currentTime);
      gainNode.gain.linearRampToValueAtTime(volume, this.audioContext.currentTime + options.fadeIn);
    }
    
    // 淡出效果
    if (options.fadeOut > 0) {
      const fadeOutTime = this.audioContext.currentTime + sound.duration - options.fadeOut;
      gainNode.gain.setValueAtTime(volume, fadeOutTime);
      gainNode.gain.linearRampToValueAtTime(0, fadeOutTime + options.fadeOut);
    }
    
    // 播放
    source.start(0);
    
    // 记录播放状态
    const playId = `${soundId}_${Date.now()}`;
    this.isPlaying.set(playId, { source, gainNode, panNode });
    
    // 播放完成后清理
    source.onended = () => {
      this.isPlaying.delete(playId);
    };
    
    return { source, gainNode, panNode, playId };
  }
  
  /**
   * 停止音效
   */
  stopSound(playId) {
    const playing = this.isPlaying.get(playId);
    if (playing && playing.source) {
      try {
        playing.source.stop();
        this.isPlaying.delete(playId);
      } catch (error) {
        console.error(`停止音效失败 ${playId}:`, error);
      }
    }
  }
  
  /**
   * 同步动画帧与音效
   */
  syncAnimationFrame(animationId, currentFrame, animationSystem) {
    const soundConfig = this.animationSounds.get(animationId);
    if (!soundConfig) return;
    
    const animation = animationSystem.animations.get(animationId);
    if (!animation) return;
    
    // 检查是否应该触发音效
    const shouldTrigger = this.shouldTriggerSound(soundConfig, currentFrame);
    
    if (shouldTrigger) {
      this.playSound(soundConfig.soundId, {
        volume: soundConfig.volume,
        pitch: soundConfig.pitch,
        pan: soundConfig.pan,
        fadeIn: soundConfig.fadeIn,
        fadeOut: soundConfig.fadeOut
      });
    }
  }
  
  /**
   * 检查是否应该触发音效
   */
  shouldTriggerSound(soundConfig, currentFrame) {
    const { triggerFrame, triggerInterval } = soundConfig;
    
    if (triggerInterval !== null) {
      // 周期性触发
      return currentFrame % triggerInterval === 0;
    }
    
    // 单次触发
    return currentFrame === triggerFrame;
  }
  
  /**
   * 设置主音量
   */
  setMasterVolume(volume) {
    this.volume = Math.max(0, Math.min(1, volume));
  }
  
  /**
   * 静音/取消静音
   */
  toggleMute() {
    this.muted = !this.muted;
    return this.muted;
  }
  
  /**
   * 停止所有音效
   */
  stopAll() {
    // 需要维护所有播放中的source引用
    this.isPlaying.clear();
  }
  
  /**
   * 创建程序化音效（火焰、水流等）
   */
  createProceduralSound(type, duration = 1.0) {
    if (!this.audioContext) return null;
    
    const sampleRate = this.audioContext.sampleRate;
    const buffer = this.audioContext.createBuffer(1, sampleRate * duration, sampleRate);
    const data = buffer.getChannelData(0);
    
    switch (type) {
      case 'fire':
        this.createFireNoise(data, sampleRate);
        break;
      case 'water':
        this.createWaterNoise(data, sampleRate);
        break;
      case 'wind':
        this.createWindNoise(data, sampleRate);
        break;
      case 'magic':
        this.createMagicSound(data, sampleRate);
        break;
      default:
        this.createWhiteNoise(data, sampleRate);
    }
    
    return buffer;
  }
  
  /**
   * 创建火焰噪音
   */
  createFireNoise(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const crackle = Math.random() > 0.98 ? Math.random() * 0.5 : 0;
      const roar = (Math.random() - 0.5) * 0.3;
      data[i] = roar + crackle;
    }
  }
  
  /**
   * 创建水流噪音
   */
  createWaterNoise(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const base = (Math.random() - 0.5) * 0.2;
      const ripple = Math.sin(t * 10) * 0.1 * Math.random();
      data[i] = base + ripple;
    }
  }
  
  /**
   * 创建风噪音
   */
  createWindNoise(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const gust = Math.sin(t * 2 + Math.random()) * 0.3;
      const whisper = (Math.random() - 0.5) * 0.1;
      data[i] = gust + whisper;
    }
  }
  
  /**
   * 创建魔法音效
   */
  createMagicSound(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      const t = i / sampleRate;
      const baseFreq = 440;
      const harmonic1 = Math.sin(2 * Math.PI * baseFreq * t) * 0.3;
      const harmonic2 = Math.sin(2 * Math.PI * baseFreq * 2 * t) * 0.2;
      const harmonic3 = Math.sin(2 * Math.PI * baseFreq * 3 * t) * 0.1;
      const sparkle = Math.random() > 0.95 ? Math.random() * 0.2 : 0;
      data[i] = harmonic1 + harmonic2 + harmonic3 + sparkle;
    }
  }
  
  /**
   * 创建白噪音
   */
  createWhiteNoise(data, sampleRate) {
    for (let i = 0; i < data.length; i++) {
      data[i] = (Math.random() - 0.5) * 0.5;
    }
  }
  
  /**
   * 预设音效配置
   */
  getPresetSoundConfigs() {
    return [
      {
        animationType: 'fire',
        soundType: 'fire',
        triggerInterval: 4,
        volume: 0.6,
        loop: true
      },
      {
        animationType: 'water',
        soundType: 'water',
        triggerInterval: 2,
        volume: 0.5,
        loop: true
      },
      {
        animationType: 'waterfall',
        soundType: 'water',
        triggerInterval: 1,
        volume: 0.7,
        loop: true
      },
      {
        animationType: 'torch',
        soundType: 'fire',
        triggerInterval: 6,
        volume: 0.4,
        loop: true
      },
      {
        animationType: 'magic_circle',
        soundType: 'magic',
        triggerInterval: 8,
        volume: 0.5,
        loop: true
      }
    ];
  }
  
  /**
   * 自动绑定预设音效
   */
  autoBindPresetSounds(animationSystem) {
    const presets = this.getPresetSoundConfigs();
    
    animationSystem.animations.forEach((animation, animId) => {
      const preset = presets.find(p => p.animationType === animation.type);
      if (preset) {
        // 创建程序化音效
        const soundBuffer = this.createProceduralSound(preset.soundType, 2.0);
        const soundId = `procedural_${preset.soundType}`;
        
        // 临时存储程序化音效
        this.soundEffects.set(soundId, {
          id: soundId,
          buffer: soundBuffer,
          duration: soundBuffer.duration
        });
        
        // 绑定音效
        this.bindAnimationSound(animId, {
          soundId: soundId,
          triggerInterval: preset.triggerInterval,
          volume: preset.volume,
          loop: preset.loop
        });
        
        console.log(`自动绑定音效: ${animation.type} -> ${soundId}`);
      }
    });
  }
  
  /**
   * 获取系统状态
   */
  getStatus() {
    return {
      initialized: this.audioContext !== null,
      soundCount: this.soundEffects.size,
      playingCount: this.isPlaying.size,
      boundAnimations: this.animationSounds.size,
      volume: this.volume,
      muted: this.muted
    };
  }
}

// 创建全局单例
window.AnimationAudioSystem = new AnimationAudioSystem();