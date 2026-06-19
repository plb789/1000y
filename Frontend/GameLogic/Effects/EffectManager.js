/**
 * 特效管理器
 * 统一管理游戏中的所有特效：技能特效、攻击特效、环境特效等
 */
class EffectManager {
  constructor() {
    // 粒子系统实例
    this.particleSystem = null;
    
    // 特效配置
    this.effects = new Map();
    
    // 活动特效实例
    this.activeEffects = new Map();
    
    // 特效ID计数器
    this.effectIdCounter = 0;
    
    // 特效层级
    this.layers = {
      ground: 0,      // 地面特效（影子、脚印等）
      character: 1,   // 角色特效（光环、状态等）
      skill: 2,       // 技能特效
      weather: 3,     // 天气特效
      ui: 4           // UI特效
    };
    
    // 初始化特效配置
    this.initEffectConfigs();
    
    console.log('特效管理器初始化完成');
  }
  
  /**
   * 设置粒子系统
   */
  setParticleSystem(particleSystem) {
    this.particleSystem = particleSystem;
  }
  
  /**
   * 初始化特效配置
   */
  initEffectConfigs() {
    // ==================== 技能特效 ====================
    
    // 剑法类技能
    this.effects.set('sword_basic', {
      name: '基础剑法',
      type: 'skill',
      animation: 'sword_slash',
      sound: 'sword_swing',
      duration: 300,
      hitEffect: 'hit_spark',
      range: 60,
      damageDelay: 150
    });
    
    this.effects.set('sword_combo', {
      name: '连击剑法',
      type: 'skill',
      animation: 'sword_slash',
      sound: 'sword_combo',
      duration: 500,
      hitEffect: 'heavy_hit',
      range: 80,
      damageDelay: 100,
      comboCount: 3,
      comboInterval: 150
    });
    
    this.effects.set('sword_qi', {
      name: '剑气',
      type: 'skill',
      animation: 'sword_slash',
      sound: 'sword_qi',
      duration: 600,
      hitEffect: 'hit_spark',
      range: 150,
      damageDelay: 200,
      projectile: true,
      projectileSpeed: 400,
      projectileEffect: 'sword_qi_wave'
    });
    
    // 拳法类技能
    this.effects.set('punch_basic', {
      name: '基础拳法',
      type: 'skill',
      animation: 'punch',
      sound: 'punch',
      duration: 200,
      hitEffect: 'hit_spark',
      range: 40,
      damageDelay: 100
    });
    
    this.effects.set('punch_heavy', {
      name: '重拳',
      type: 'skill',
      animation: 'punch',
      sound: 'heavy_punch',
      duration: 400,
      hitEffect: 'heavy_hit',
      range: 50,
      damageDelay: 200,
      shake: true,
      shakeIntensity: 3
    });
    
    // 内功类技能
    this.effects.set('heal', {
      name: '疗伤心法',
      type: 'skill',
      animation: 'heal',
      sound: 'heal',
      duration: 1000,
      selfEffect: true,
      particleEffect: 'heal',
      glowColor: '#4ade80',
      healAmount: true
    });
    
    this.effects.set('shield', {
      name: '护体真气',
      type: 'skill',
      animation: 'shield',
      sound: 'shield',
      duration: 500,
      selfEffect: true,
      particleEffect: 'magic_circle',
      glowColor: '#60a5fa',
      shieldEffect: true
    });
    
    // 火系技能
    this.effects.set('fire_ball', {
      name: '火球术',
      type: 'skill',
      animation: 'fire_burst',
      sound: 'fire_cast',
      duration: 800,
      hitEffect: 'fire_explosion',
      range: 200,
      damageDelay: 300,
      projectile: true,
      projectileSpeed: 300,
      projectileEffect: 'fire_ball',
      trailEffect: 'fire_trail'
    });
    
    this.effects.set('fire_wave', {
      name: '烈焰掌',
      type: 'skill',
      animation: 'fire_burst',
      sound: 'fire_wave',
      duration: 600,
      hitEffect: 'fire_explosion',
      range: 100,
      damageDelay: 200,
      aoeRadius: 80,
      aoeEffect: 'fire_aoe'
    });
    
    // 冰系技能
    this.effects.set('ice_arrow', {
      name: '冰箭术',
      type: 'skill',
      animation: 'ice_burst',
      sound: 'ice_cast',
      duration: 600,
      hitEffect: 'ice_shatter',
      range: 180,
      damageDelay: 250,
      projectile: true,
      projectileSpeed: 350,
      projectileEffect: 'ice_arrow',
      slowEffect: true
    });
    
    this.effects.set('ice_shield', {
      name: '寒冰护体',
      type: 'skill',
      animation: 'ice_burst',
      sound: 'ice_shield',
      duration: 500,
      selfEffect: true,
      particleEffect: 'ice_shield',
      glowColor: '#60a5fa',
      defenseBoost: true
    });
    
    // 雷系技能
    this.effects.set('thunder_strike', {
      name: '雷霆一击',
      type: 'skill',
      animation: 'thunder',
      sound: 'thunder',
      duration: 400,
      hitEffect: 'thunder_impact',
      range: 120,
      damageDelay: 150,
      shake: true,
      shakeIntensity: 8,
      flash: true,
      flashColor: '#fbbf24'
    });
    
    // ==================== 战斗特效 ====================
    
    // 受击特效
    this.effects.set('hit_normal', {
      name: '普通受击',
      type: 'combat',
      particleEffect: 'hit_spark',
      duration: 200,
      sound: 'hit_normal',
      screenShake: false
    });
    
    this.effects.set('hit_heavy', {
      name: '重击受击',
      type: 'combat',
      particleEffect: 'heavy_hit',
      duration: 300,
      sound: 'hit_heavy',
      screenShake: true,
      shakeIntensity: 3
    });
    
    this.effects.set('hit_critical', {
      name: '暴击',
      type: 'combat',
      particleEffect: 'hit_critical',
      duration: 500,
      sound: 'hit_critical',
      screenShake: true,
      shakeIntensity: 5,
      flash: true,
      flashColor: '#fbbf24',
      glowColor: '#fbbf24'
    });
    
    this.effects.set('block', {
      name: '格挡',
      type: 'combat',
      particleEffect: 'block',
      duration: 300,
      sound: 'block',
      glowColor: '#60a5fa'
    });
    
    this.effects.set('dodge', {
      name: '闪避',
      type: 'combat',
      particleEffect: 'dodge',
      duration: 250,
      sound: 'dodge'
    });
    
    this.effects.set('parry', {
      name: '招架',
      type: 'combat',
      particleEffect: 'parry',
      duration: 350,
      sound: 'block',
      glowColor: '#4ade80'
    });
    
    // ==================== 角色状态特效 ====================
    
    this.effects.set('poison', {
      name: '中毒',
      type: 'status',
      particleEffect: 'poison_bubbles',
      interval: 1000,
      duration: 5000,
      color: '#a855f7',
      sound: 'poison'
    });
    
    this.effects.set('burn', {
      name: '灼烧',
      type: 'status',
      particleEffect: 'fire_burst',
      interval: 500,
      duration: 3000,
      color: '#ef4444',
      sound: 'burn'
    });
    
    this.effects.set('freeze', {
      name: '冰冻',
      type: 'status',
      particleEffect: 'ice_burst',
      duration: 2000,
      color: '#60a5fa',
      slowEffect: 0.5,
      sound: 'freeze'
    });
    
    this.effects.set('stun', {
      name: '眩晕',
      type: 'status',
      particleEffect: 'stun_stars',
      duration: 1500,
      color: '#fbbf24',
      sound: 'stun'
    });
    
    this.effects.set('bleed', {
      name: '流血',
      type: 'status',
      particleEffect: 'bleed',
      interval: 800,
      duration: 4000,
      color: '#ef4444'
    });
    
    this.effects.set('silence', {
      name: '沉默',
      type: 'status',
      particleEffect: 'silence',
      duration: 3000,
      color: '#6b7280'
    });
    
    this.effects.set('fear', {
      name: '恐惧',
      type: 'status',
      particleEffect: 'fear',
      duration: 2000,
      color: '#9333ea',
      sound: 'fear'
    });
    
    // ==================== 角色增益特效 ====================
    
    this.effects.set('buff_attack', {
      name: '攻击增益',
      type: 'buff',
      particleEffect: 'buff_attack',
      duration: 10000,
      color: '#ef4444',
      glowColor: '#ef4444'
    });
    
    this.effects.set('buff_defense', {
      name: '防御增益',
      type: 'buff',
      particleEffect: 'buff_defense',
      duration: 10000,
      color: '#60a5fa',
      glowColor: '#60a5fa'
    });
    
    this.effects.set('buff_speed', {
      name: '速度增益',
      type: 'buff',
      particleEffect: 'buff_speed',
      duration: 8000,
      color: '#4ade80',
      glowColor: '#4ade80'
    });
    
    this.effects.set('buff_heal', {
      name: '持续治疗',
      type: 'buff',
      particleEffect: 'heal',
      interval: 2000,
      duration: 10000,
      color: '#4ade80',
      sound: 'heal'
    });
    
    // ==================== 环境特效 ====================
    
    this.effects.set('weather_rain', {
      name: '下雨',
      type: 'weather',
      particleEffect: 'rain',
      continuous: true,
      intensity: 100,
      ambientSound: 'ambient_rain'
    });
    
    this.effects.set('weather_snow', {
      name: '下雪',
      type: 'weather',
      particleEffect: 'snow',
      continuous: true,
      intensity: 50,
      ambientSound: 'ambient_wind'
    });
    
    this.effects.set('weather_leaves', {
      name: '落叶',
      type: 'weather',
      particleEffect: 'falling_leaves',
      continuous: true,
      intensity: 20
    });
    
    this.effects.set('weather_fog', {
      name: '迷雾',
      type: 'weather',
      particleEffect: 'fog',
      continuous: true,
      intensity: 30,
      duration: 0 // 持续直到关闭
    });
    
    // ==================== 交互特效 ====================
    
    this.effects.set('pickup', {
      name: '拾取',
      type: 'interaction',
      particleEffect: 'pickup',
      duration: 500,
      sound: 'pickup'
    });
    
    this.effects.set('level_up', {
      name: '升级',
      type: 'interaction',
      particleEffect: 'level_up',
      duration: 1500,
      sound: 'level_up',
      glowColor: '#fbbf24',
      flash: true,
      screenShake: true,
      shakeIntensity: 3
    });
    
    this.effects.set('death', {
      name: '死亡',
      type: 'interaction',
      particleEffect: 'death',
      duration: 800,
      sound: 'death',
      fadeOut: true
    });
    
    this.effects.set('respawn', {
      name: '复活',
      type: 'interaction',
      particleEffect: 'respawn',
      duration: 1000,
      sound: 'respawn',
      glowColor: '#4ade80',
      fadeIn: true
    });
    
    this.effects.set('teleport', {
      name: '传送',
      type: 'interaction',
      particleEffect: 'magic_circle',
      duration: 600,
      sound: 'teleport',
      fadeOut: true,
      fadeIn: true
    });
    
    this.effects.set('drop_item', {
      name: '掉落物品',
      type: 'interaction',
      particleEffect: 'drop_item',
      duration: 400,
      sound: 'drop_item'
    });
    
    this.effects.set('craft_success', {
      name: '制作成功',
      type: 'interaction',
      particleEffect: 'craft_success',
      duration: 600,
      sound: 'craft_success',
      glowColor: '#fbbf24'
    });
    
    this.effects.set('equip', {
      name: '装备',
      type: 'interaction',
      particleEffect: 'equip',
      duration: 400,
      sound: 'equip',
      glowColor: '#60a5fa'
    });
    
    this.effects.set('unequip', {
      name: '卸下',
      type: 'interaction',
      particleEffect: 'unequip',
      duration: 300,
      sound: 'unequip'
    });
    
    // ==================== 任务/成就特效 ====================
    
    this.effects.set('quest_complete', {
      name: '任务完成',
      type: 'interaction',
      particleEffect: 'quest_complete',
      duration: 1000,
      sound: 'quest_complete',
      glowColor: '#4ade80'
    });
    
    this.effects.set('achievement', {
      name: '成就达成',
      type: 'interaction',
      particleEffect: 'achievement',
      duration: 2000,
      sound: 'achievement',
      glowColor: '#fbbf24',
      flash: true
    });
    
    // ==================== 地图事件特效 ====================
    
    this.effects.set('map_event_spawn', {
      name: '事件出现',
      type: 'world',
      particleEffect: 'event_spawn',
      duration: 800,
      sound: 'event_spawn'
    });
    
    this.effects.set('map_event_end', {
      name: '事件结束',
      type: 'world',
      particleEffect: 'event_end',
      duration: 600,
      sound: 'event_end'
    });
    
    this.effects.set('portal_open', {
      name: '传送门开启',
      type: 'world',
      particleEffect: 'portal_open',
      duration: 500,
      sound: 'portal_open',
      glowColor: '#a855f7'
    });
    
    this.effects.set('chest_open', {
      name: '宝箱开启',
      type: 'world',
      particleEffect: 'chest_open',
      duration: 600,
      sound: 'chest_open',
      glowColor: '#fbbf24'
    });
    
    // ==================== 社交特效 ====================
    
    this.effects.set('friend_add', {
      name: '添加好友',
      type: 'interaction',
      particleEffect: 'friend_add',
      duration: 400,
      sound: 'friend_add',
      glowColor: '#60a5fa'
    });
    
    this.effects.set('guild_join', {
      name: '加入门派',
      type: 'interaction',
      particleEffect: 'guild_join',
      duration: 800,
      sound: 'guild_join',
      glowColor: '#e94560'
    });
    
    this.effects.set('trade_start', {
      name: '交易开始',
      type: 'interaction',
      particleEffect: 'trade_start',
      duration: 400,
      sound: 'trade_start',
      glowColor: '#4ade80'
    });
    
    this.effects.set('trade_end', {
      name: '交易结束',
      type: 'interaction',
      particleEffect: 'trade_end',
      duration: 400,
      sound: 'trade_end'
    });
  }
  
  /**
   * 触发技能特效
   */
  triggerSkillEffect(skillId, caster, target, options = {}) {
    const config = this.effects.get(skillId);
    if (!config || !this.particleSystem) {
      console.warn(`技能特效配置未找到: ${skillId}`);
      return null;
    }
    
    const effectId = this.generateEffectId();
    const effect = {
      id: effectId,
      config: config,
      caster: caster,
      target: target,
      startTime: Date.now(),
      options: options,
      phase: 'start',
      active: true
    };
    
    this.activeEffects.set(effectId, effect);
    
    // 播放音效
    if (config.sound) {
      this.playSound(config.sound, caster);
    }
    
    // 屏幕闪烁
    if (config.flash) {
      this.flashScreen(config.flashColor);
    }
    
    // 屏幕震动
    if (config.shake) {
      this.shakeScreen(config.shakeIntensity || 5);
    }
    
    // 触发粒子特效
    if (config.particleEffect && this.particleSystem) {
      const x = target ? target.x : caster.x;
      const y = target ? target.y : caster.y;
      this.particleSystem.triggerEffect(config.particleEffect, x, y);
    }
    
    // 投射物特效
    if (config.projectile) {
      this.createProjectile(effect, caster, target);
    }
    
    // 连击特效
    if (config.comboCount) {
      this.handleComboEffect(effect);
    }
    
    // 自动结束
    setTimeout(() => {
      this.endEffect(effectId);
    }, config.duration);
    
    return effectId;
  }
  
  /**
   * 触发状态特效
   */
  triggerStatusEffect(statusId, target, options = {}) {
    const config = this.effects.get(statusId);
    if (!config || !this.particleSystem) {
      return null;
    }
    
    const effectId = this.generateEffectId();
    const effect = {
      id: effectId,
      config: config,
      target: target,
      startTime: Date.now(),
      options: options,
      active: true
    };
    
    this.activeEffects.set(effectId, effect);
    
    // 持续性状态特效
    if (config.interval && config.duration) {
      effect.intervalId = setInterval(() => {
        if (effect.active && config.particleEffect) {
          this.particleSystem.triggerEffect(config.particleEffect, target.x, target.y);
        }
      }, config.interval);
      
      setTimeout(() => {
        this.endEffect(effectId);
      }, config.duration);
    }
    
    return effectId;
  }
  
  /**
   * 触发环境特效
   */
  triggerWeatherEffect(weatherId, options = {}) {
    const config = this.effects.get(weatherId);
    if (!config || !this.particleSystem) {
      return null;
    }
    
    const effectId = this.generateEffectId();
    const effect = {
      id: effectId,
      config: config,
      startTime: Date.now(),
      options: options,
      active: true
    };
    
    this.activeEffects.set(effectId, effect);
    
    // 持续性天气特效
    if (config.continuous && config.particleEffect) {
      // 天气特效会持续触发粒子
      effect.weatherInterval = setInterval(() => {
        if (effect.active && this.particleSystem) {
          const canvas = this.particleSystem.canvas;
          if (canvas) {
            const x = Math.random() * canvas.width;
            this.particleSystem.triggerEffect(config.particleEffect, x, 0);
          }
        }
      }, 1000 / (config.intensity || 50));
    }
    
    return effectId;
  }
  
  /**
   * 触发交互特效
   */
  triggerInteractionEffect(interactionId, x, y, options = {}) {
    const config = this.effects.get(interactionId);
    if (!config || !this.particleSystem) {
      return null;
    }
    
    const effectId = this.generateEffectId();
    
    // 播放音效
    if (config.sound) {
      this.playSound(config.sound, { x, y });
    }
    
    // 触发粒子特效
    if (config.particleEffect && this.particleSystem) {
      this.particleSystem.triggerEffect(config.particleEffect, x, y);
    }
    
    // 屏幕闪烁
    if (config.flash) {
      this.flashScreen(config.glowColor || '#ffffff');
    }
    
    // 自动结束
    if (config.duration) {
      setTimeout(() => {
        this.endEffect(effectId);
      }, config.duration);
    }
    
    return effectId;
  }
  
  /**
   * 创建投射物
   */
  createProjectile(effect, caster, target) {
    if (!target) return;
    
    const config = effect.config;
    const dx = target.x - caster.x;
    const dy = target.y - caster.y;
    const distance = Math.sqrt(dx * dx + dy * dy);
    const duration = distance / config.projectileSpeed * 1000;
    
    effect.projectile = {
      x: caster.x,
      y: caster.y,
      targetX: target.x,
      targetY: target.y,
      vx: (dx / distance) * config.projectileSpeed,
      vy: (dy / distance) * config.projectileSpeed,
      startTime: Date.now(),
      duration: duration
    };
    
    // 更新投射物位置
    effect.projectileInterval = setInterval(() => {
      if (!effect.active || !effect.projectile) {
        clearInterval(effect.projectileInterval);
        return;
      }
      
      const elapsed = Date.now() - effect.projectile.startTime;
      const progress = elapsed / duration;
      
      if (progress >= 1) {
        // 投射物到达目标
        clearInterval(effect.projectileInterval);
        this.onProjectileHit(effect);
      } else {
        // 更新位置
        effect.projectile.x = caster.x + (target.x - caster.x) * progress;
        effect.projectile.y = caster.y + (target.y - caster.y) * progress;
        
        // 轨迹特效
        if (config.trailEffect && this.particleSystem) {
          this.particleSystem.triggerEffect(config.trailEffect, effect.projectile.x, effect.projectile.y);
        }
      }
    }, 16);
  }
  
  /**
   * 投射物命中
   */
  onProjectileHit(effect) {
    const config = effect.config;
    
    if (config.hitEffect && this.particleSystem) {
      this.particleSystem.triggerEffect(config.hitEffect, effect.projectile.targetX, effect.projectile.targetY);
    }
    
    // 触发命中回调
    if (effect.options.onHit) {
      effect.options.onHit(effect.target);
    }
  }
  
  /**
   * 处理连击特效
   */
  handleComboEffect(effect) {
    const config = effect.config;
    let comboCount = 0;
    
    effect.comboInterval = setInterval(() => {
      if (!effect.active || comboCount >= config.comboCount) {
        clearInterval(effect.comboInterval);
        return;
      }
      
      comboCount++;
      
      // 触发连击特效
      if (this.particleSystem) {
        this.particleSystem.triggerEffect(config.animation, effect.target.x, effect.target.y);
      }
      
      // 触发连击回调
      if (effect.options.onComboHit) {
        effect.options.onComboHit(comboCount);
      }
    }, config.comboInterval);
  }
  
  /**
   * 播放音效
   */
  playSound(soundId, position) {
    // 调用音效系统播放音效
    if (window.AnimationAudioSystem) {
      const pan = position ? (position.x - (this.particleSystem?.canvas?.width || 800) / 2) / 400 : 0;
      window.AnimationAudioSystem.playSound(soundId, { pan: pan });
    }
  }
  
  /**
   * 屏幕闪烁
   */
  flashScreen(color) {
    const flash = document.createElement('div');
    flash.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: ${color};
      opacity: 0.3;
      pointer-events: none;
      z-index: 9999;
      animation: flashFade 0.3s ease forwards;
    `;
    
    document.body.appendChild(flash);
    
    setTimeout(() => {
      if (flash.parentNode) {
        flash.parentNode.removeChild(flash);
      }
    }, 300);
  }
  
  /**
   * 屏幕震动
   */
  shakeScreen(intensity) {
    if (this.particleSystem) {
      this.particleSystem.shakeScreen(intensity);
    }
  }
  
  /**
   * 结束特效
   */
  endEffect(effectId) {
    const effect = this.activeEffects.get(effectId);
    if (!effect) return;
    
    effect.active = false;
    
    // 清理定时器
    if (effect.intervalId) {
      clearInterval(effect.intervalId);
    }
    if (effect.projectileInterval) {
      clearInterval(effect.projectileInterval);
    }
    if (effect.comboInterval) {
      clearInterval(effect.comboInterval);
    }
    if (effect.weatherInterval) {
      clearInterval(effect.weatherInterval);
    }
    
    this.activeEffects.delete(effectId);
  }
  
  /**
   * 生成特效ID
   */
  generateEffectId() {
    return `effect_${++this.effectIdCounter}_${Date.now()}`;
  }
  
  /**
   * 更新特效
   */
  update(deltaTime) {
    // 更新粒子系统
    if (this.particleSystem) {
      this.particleSystem.update(deltaTime);
    }
  }
  
  /**
   * 渲染特效
   */
  render(ctx) {
    // 渲染粒子系统
    if (this.particleSystem) {
      this.particleSystem.render(ctx);
    }
    
    // 渲染投射物
    this.activeEffects.forEach(effect => {
      if (effect.projectile && effect.active) {
        this.renderProjectile(ctx, effect);
      }
    });
  }
  
  /**
   * 渲染投射物
   */
  renderProjectile(ctx, effect) {
    const proj = effect.projectile;
    if (!proj) return;
    
    ctx.save();
    
    // 绘制投射物
    const gradient = ctx.createRadialGradient(proj.x, proj.y, 0, proj.x, proj.y, 15);
    gradient.addColorStop(0, '#ffffff');
    gradient.addColorStop(0.5, effect.config.glowColor || '#e94560');
    gradient.addColorStop(1, 'transparent');
    
    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.arc(proj.x, proj.y, 15, 0, Math.PI * 2);
    ctx.fill();
    
    ctx.restore();
  }
  
  /**
   * 清除所有特效
   */
  clearAll() {
    this.activeEffects.forEach((effect, id) => {
      this.endEffect(id);
    });
    
    if (this.particleSystem) {
      this.particleSystem.clearAll();
    }
  }
  
  /**
   * 添加自定义特效配置
   */
  addEffectConfig(effectId, config) {
    this.effects.set(effectId, config);
  }
  
  /**
   * 获取特效配置
   */
  getEffectConfig(effectId) {
    return this.effects.get(effectId);
  }
  
  /**
   * 获取活动特效数量
   */
  getActiveEffectCount() {
    return this.activeEffects.size;
  }
}

// 创建全局单例
window.EffectManager = new EffectManager();