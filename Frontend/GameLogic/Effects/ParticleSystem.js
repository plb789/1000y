/**
 * 粒子系统
 * 用于创建各种视觉效果：技能特效、攻击特效、环境特效等
 */
class ParticleSystem {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    
    // 粒子池
    this.particles = [];
    this.maxParticles = 1000;
    
    // 特效模板
    this.templates = new Map();
    
    // 活动特效
    this.activeEffects = new Map();
    
    // 初始化预设模板
    this.initTemplates();
    
    console.log('粒子系统初始化完成');
  }
  
  /**
   * 初始化预设特效模板
   */
  initTemplates() {
    // 攻击特效 - 剑气
    this.templates.set('sword_slash', {
      type: 'burst',
      count: 20,
      lifetime: 500,
      shape: 'line',
      color: { start: '#ffffff', end: '#e94560' },
      size: { start: 3, end: 0 },
      speed: { min: 200, max: 400 },
      direction: { spread: 60, base: 0 },
      gravity: 0,
      fadeOut: true,
      trail: true,
      trailLength: 5
    });
    
    // 攻击特效 - 重击
    this.templates.set('heavy_hit', {
      type: 'explosion',
      count: 30,
      lifetime: 400,
      shape: 'circle',
      color: { start: '#fbbf24', end: '#ef4444' },
      size: { start: 5, end: 2 },
      speed: { min: 150, max: 300 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      shake: true,
      shakeIntensity: 5
    });
    
    // 技能特效 - 火焰
    this.templates.set('fire_burst', {
      type: 'stream',
      count: 50,
      lifetime: 800,
      shape: 'circle',
      color: { start: '#ef4444', end: '#fbbf24' },
      size: { start: 8, end: 2 },
      speed: { min: 50, max: 100 },
      direction: { spread: 30, base: -90 },
      gravity: -50,
      fadeOut: true,
      glow: true,
      glowColor: '#ef4444'
    });
    
    // 技能特效 - 冰霜
    this.templates.set('ice_burst', {
      type: 'burst',
      count: 40,
      lifetime: 600,
      shape: 'diamond',
      color: { start: '#60a5fa', end: '#ffffff' },
      size: { start: 6, end: 1 },
      speed: { min: 100, max: 200 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#60a5fa'
    });
    
    // 技能特效 - 魔法阵
    this.templates.set('magic_circle', {
      type: 'ring',
      count: 60,
      lifetime: 1000,
      shape: 'circle',
      color: { start: '#a855f7', end: '#e94560' },
      size: { start: 3, end: 1 },
      speed: { min: 80, max: 80 },
      direction: { spread: 0, base: 0, rotate: true },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#a855f7',
      ringRadius: 50
    });
    
    // 技能特效 - 治愈
    this.templates.set('heal', {
      type: 'rise',
      count: 30,
      lifetime: 1000,
      shape: 'star',
      color: { start: '#4ade80', end: '#ffffff' },
      size: { start: 5, end: 2 },
      speed: { min: 30, max: 60 },
      direction: { spread: 360, base: -90 },
      gravity: -20,
      fadeOut: true,
      glow: true,
      glowColor: '#4ade80'
    });
    
    // 环境特效 - 落叶
    this.templates.set('falling_leaves', {
      type: 'fall',
      count: 20,
      lifetime: 3000,
      shape: 'leaf',
      color: { start: '#fbbf24', end: '#a3e635' },
      size: { start: 4, end: 4 },
      speed: { min: 20, max: 40 },
      direction: { spread: 30, base: 90 },
      gravity: 30,
      fadeOut: false,
      rotate: true,
      rotateSpeed: 2
    });
    
    // 环境特效 - 雪花
    this.templates.set('snow', {
      type: 'fall',
      count: 50,
      lifetime: 4000,
      shape: 'circle',
      color: { start: '#ffffff', end: '#e0f2fe' },
      size: { start: 2, end: 2 },
      speed: { min: 10, max: 30 },
      direction: { spread: 20, base: 90 },
      gravity: 10,
      fadeOut: false,
      sway: true,
      swayAmount: 20
    });
    
    // 环境特效 - 雨滴
    this.templates.set('rain', {
      type: 'fall',
      count: 100,
      lifetime: 1000,
      shape: 'line',
      color: { start: '#60a5fa', end: '#93c5fd' },
      size: { start: 1, end: 1 },
      speed: { min: 300, max: 500 },
      direction: { spread: 5, base: 90 },
      gravity: 0,
      fadeOut: true
    });
    
    // 受击特效
    this.templates.set('hit_spark', {
      type: 'burst',
      count: 15,
      lifetime: 200,
      shape: 'circle',
      color: { start: '#ffffff', end: '#fbbf24' },
      size: { start: 3, end: 0 },
      speed: { min: 100, max: 200 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true
    });
    
    // 普通受击特效
    this.templates.set('hit_normal', {
      type: 'burst',
      count: 12,
      lifetime: 250,
      shape: 'circle',
      color: { start: '#ffffff', end: '#ef4444' },
      size: { start: 4, end: 1 },
      speed: { min: 80, max: 150 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true
    });
    
    // 暴击受击特效
    this.templates.set('hit_critical', {
      type: 'explosion',
      count: 30,
      lifetime: 400,
      shape: 'star',
      color: { start: '#fbbf24', end: '#ef4444' },
      size: { start: 8, end: 2 },
      speed: { min: 150, max: 300 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#fbbf24',
      shake: true,
      shakeIntensity: 6
    });
    
    // 格挡特效
    this.templates.set('block', {
      type: 'ring',
      count: 20,
      lifetime: 350,
      shape: 'circle',
      color: { start: '#60a5fa', end: '#ffffff' },
      size: { start: 4, end: 1 },
      speed: { min: 100, max: 100 },
      direction: { spread: 0, base: 0, rotate: true },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#60a5fa',
      ringRadius: 40
    });
    
    // 闪避特效
    this.templates.set('dodge', {
      type: 'burst',
      count: 15,
      lifetime: 300,
      shape: 'diamond',
      color: { start: '#a855f7', end: '#ffffff' },
      size: { start: 3, end: 0 },
      speed: { min: 120, max: 200 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#a855f7'
    });
    
    // 护盾特效
    this.templates.set('shield', {
      type: 'ring',
      count: 30,
      lifetime: 600,
      shape: 'circle',
      color: { start: '#60a5fa', end: '#a855f7' },
      size: { start: 5, end: 1 },
      speed: { min: 80, max: 80 },
      direction: { spread: 0, base: 0, rotate: true },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#60a5fa',
      ringRadius: 50
    });
    
    // 雷电特效
    this.templates.set('thunder_impact', {
      type: 'explosion',
      count: 40,
      lifetime: 500,
      shape: 'star',
      color: { start: '#fbbf24', end: '#3b82f6' },
      size: { start: 6, end: 1 },
      speed: { min: 200, max: 400 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#fbbf24',
      shake: true,
      shakeIntensity: 8
    });
    
    // 死亡特效
    this.templates.set('death', {
      type: 'explosion',
      count: 50,
      lifetime: 800,
      shape: 'circle',
      color: { start: '#e94560', end: '#1a1a2e' },
      size: { start: 10, end: 0 },
      speed: { min: 100, max: 200 },
      direction: { spread: 360, base: 0 },
      gravity: 0,
      fadeOut: true,
      glow: true,
      glowColor: '#e94560'
    });
    
    // 升级特效
    this.templates.set('level_up', {
      type: 'rise',
      count: 40,
      lifetime: 1500,
      shape: 'star',
      color: { start: '#fbbf24', end: '#4ade80' },
      size: { start: 8, end: 2 },
      speed: { min: 50, max: 100 },
      direction: { spread: 360, base: -90 },
      gravity: -30,
      fadeOut: true,
      glow: true,
      glowColor: '#fbbf24',
      ring: true,
      ringRadius: 60
    });
    
    // 拾取特效
    this.templates.set('pickup', {
      type: 'rise',
      count: 10,
      lifetime: 500,
      shape: 'circle',
      color: { start: '#fbbf24', end: '#ffffff' },
      size: { start: 4, end: 0 },
      speed: { min: 80, max: 120 },
      direction: { spread: 360, base: -90 },
      gravity: -50,
      fadeOut: true
    });
  }
  
  /**
   * 创建粒子
   */
  createParticle(x, y, template, options = {}) {
    if (this.particles.length >= this.maxParticles) {
      // 移除最旧的粒子
      this.particles.shift();
    }
    
    const config = Object.assign({}, template, options);
    
    const particle = {
      x: x,
      y: y,
      startX: x,
      startY: y,
      vx: 0,
      vy: 0,
      lifetime: config.lifetime,
      age: 0,
      size: config.size.start,
      startSize: config.size.start,
      endSize: config.size.end,
      color: config.color.start,
      startColor: config.color.start,
      endColor: config.color.end,
      alpha: 1,
      shape: config.shape,
      gravity: config.gravity || 0,
      fadeOut: config.fadeOut,
      glow: config.glow,
      glowColor: config.glowColor,
      trail: config.trail,
      trailLength: config.trailLength || 0,
      trailPositions: [],
      rotate: config.rotate,
      rotation: 0,
      rotateSpeed: config.rotateSpeed || 0,
      sway: config.sway,
      swayAmount: config.swayAmount || 0,
      swayOffset: Math.random() * Math.PI * 2,
      active: true
    };
    
    // 计算初始速度
    const speed = config.speed.min + Math.random() * (config.speed.max - config.speed.min);
    const spreadRad = (config.direction.spread / 180) * Math.PI;
    const baseRad = (config.direction.base / 180) * Math.PI;
    const angle = baseRad + (Math.random() - 0.5) * spreadRad;
    
    particle.vx = Math.cos(angle) * speed;
    particle.vy = Math.sin(angle) * speed;
    
    this.particles.push(particle);
    return particle;
  }
  
  /**
   * 触发特效
   */
  triggerEffect(templateName, x, y, options = {}) {
    const template = this.templates.get(templateName);
    if (!template) {
      console.warn(`特效模板未找到: ${templateName}`);
      return null;
    }
    
    const effectId = `${templateName}_${Date.now()}`;
    const effect = {
      id: effectId,
      template: templateName,
      x: x,
      y: y,
      startTime: Date.now(),
      options: options,
      particles: [],
      active: true
    };
    
    // 根据类型创建粒子
    switch (template.type) {
      case 'burst':
      case 'explosion':
        this.createBurstParticles(x, y, template, effect);
        break;
      case 'stream':
        this.createStreamParticles(x, y, template, effect);
        break;
      case 'rise':
        this.createRiseParticles(x, y, template, effect);
        break;
      case 'fall':
        this.createFallParticles(x, y, template, effect);
        break;
      case 'ring':
        this.createRingParticles(x, y, template, effect);
        break;
    }
    
    this.activeEffects.set(effectId, effect);
    
    // 屏幕震动效果
    if (template.shake && this.canvas) {
      this.shakeScreen(template.shakeIntensity || 5);
    }
    
    return effectId;
  }
  
  /**
   * 创建爆发粒子
   */
  createBurstParticles(x, y, template, effect) {
    const count = template.count;
    for (let i = 0; i < count; i++) {
      const particle = this.createParticle(x, y, template, effect.options);
      effect.particles.push(particle);
    }
  }
  
  /**
   * 创建流式粒子（持续发射）
   */
  createStreamParticles(x, y, template, effect) {
    // 流式特效需要持续发射
    effect.emitInterval = setInterval(() => {
      if (!effect.active) {
        clearInterval(effect.emitInterval);
        return;
      }
      const count = Math.ceil(template.count / 10);
      for (let i = 0; i < count; i++) {
        const particle = this.createParticle(x, y, template, effect.options);
        effect.particles.push(particle);
      }
    }, 100);
    
    // 设置停止时间
    setTimeout(() => {
      effect.active = false;
      clearInterval(effect.emitInterval);
    }, template.lifetime);
  }
  
  /**
   * 创建上升粒子
   */
  createRiseParticles(x, y, template, effect) {
    const count = template.count;
    for (let i = 0; i < count; i++) {
      // 随机偏移位置
      const offsetX = (Math.random() - 0.5) * (template.ringRadius || 40);
      const offsetY = Math.random() * 20;
      const particle = this.createParticle(x + offsetX, y + offsetY, template, effect.options);
      effect.particles.push(particle);
    }
  }
  
  /**
   * 创建下落粒子
   */
  createFallParticles(x, y, template, effect) {
    const count = template.count;
    const canvasWidth = this.canvas?.width || 800;
    
    for (let i = 0; i < count; i++) {
      // 在屏幕顶部随机位置生成
      const startX = Math.random() * canvasWidth;
      const particle = this.createParticle(startX, 0, template, effect.options);
      effect.particles.push(particle);
    }
  }
  
  /**
   * 创建环形粒子
   */
  createRingParticles(x, y, template, effect) {
    const count = template.count;
    const radius = template.ringRadius || 50;
    
    for (let i = 0; i < count; i++) {
      const angle = (i / count) * Math.PI * 2;
      const px = x + Math.cos(angle) * radius;
      const py = y + Math.sin(angle) * radius;
      
      const particle = this.createParticle(px, py, template, effect.options);
      // 环形粒子沿圆周运动
      particle.ringAngle = angle;
      particle.ringRadius = radius;
      particle.ringCenterX = x;
      particle.ringCenterY = y;
      particle.ringSpeed = template.speed.min / 100;
      
      effect.particles.push(particle);
    }
  }
  
  /**
   * 更新粒子
   */
  update(deltaTime) {
    const dt = deltaTime / 1000;
    
    this.particles.forEach(particle => {
      if (!particle.active) return;
      
      // 更新年龄
      particle.age += deltaTime;
      
      // 检查生命周期
      if (particle.age >= particle.lifetime) {
        particle.active = false;
        return;
      }
      
      // 计算生命周期进度
      const progress = particle.age / particle.lifetime;
      
      // 更新大小
      particle.size = particle.startSize + (particle.endSize - particle.startSize) * progress;
      
      // 更新颜色
      particle.color = this.lerpColor(particle.startColor, particle.endColor, progress);
      
      // 更新透明度
      if (particle.fadeOut) {
        particle.alpha = 1 - progress;
      }
      
      // 应用重力
      particle.vy += particle.gravity * dt;
      
      // 应用摇摆
      if (particle.sway) {
        const swayTime = particle.age / 500;
        particle.vx += Math.sin(swayTime + particle.swayOffset) * particle.swayAmount * dt;
      }
      
      // 应用旋转
      if (particle.rotate) {
        particle.rotation += particle.rotateSpeed * dt;
      }
      
      // 环形运动
      if (particle.ringAngle !== undefined) {
        particle.ringAngle += particle.ringSpeed * dt;
        particle.x = particle.ringCenterX + Math.cos(particle.ringAngle) * particle.ringRadius;
        particle.y = particle.ringCenterY + Math.sin(particle.ringAngle) * particle.ringRadius;
      } else {
        // 更新位置
        particle.x += particle.vx * dt;
        particle.y += particle.vy * dt;
      }
      
      // 记录轨迹
      if (particle.trail && particle.trailLength > 0) {
        particle.trailPositions.unshift({ x: particle.x, y: particle.y });
        if (particle.trailPositions.length > particle.trailLength) {
          particle.trailPositions.pop();
        }
      }
    });
    
    // 清理失效粒子
    this.particles = this.particles.filter(p => p.active);
    
    // 清理失效特效
    this.activeEffects.forEach((effect, id) => {
      if (effect.particles.every(p => !p.active)) {
        effect.active = false;
        this.activeEffects.delete(id);
      }
    });
  }
  
  /**
   * 渲染粒子
   */
  render(ctx) {
    this.particles.forEach(particle => {
      if (!particle.active || particle.alpha <= 0) return;
      
      ctx.save();
      ctx.globalAlpha = particle.alpha;
      
      // 发光效果
      if (particle.glow && particle.glowColor) {
        ctx.shadowColor = particle.glowColor;
        ctx.shadowBlur = particle.size * 2;
      }
      
      // 绘制轨迹
      if (particle.trail && particle.trailPositions.length > 0) {
        ctx.strokeStyle = particle.color;
        ctx.lineWidth = particle.size;
        ctx.beginPath();
        ctx.moveTo(particle.x, particle.y);
        particle.trailPositions.forEach(pos => {
          ctx.lineTo(pos.x, pos.y);
        });
        ctx.stroke();
      }
      
      // 应用旋转
      if (particle.rotate) {
        ctx.translate(particle.x, particle.y);
        ctx.rotate(particle.rotation);
        ctx.translate(-particle.x, -particle.y);
      }
      
      // 设置颜色
      ctx.fillStyle = particle.color;
      
      // 根据形状绘制
      switch (particle.shape) {
        case 'circle':
          ctx.beginPath();
          ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
          ctx.fill();
          break;
          
        case 'line':
          ctx.strokeStyle = particle.color;
          ctx.lineWidth = particle.size;
          ctx.beginPath();
          ctx.moveTo(particle.x, particle.y);
          ctx.lineTo(particle.x - particle.vx * 0.02, particle.y - particle.vy * 0.02);
          ctx.stroke();
          break;
          
        case 'diamond':
          this.drawDiamond(ctx, particle.x, particle.y, particle.size);
          break;
          
        case 'star':
          this.drawStar(ctx, particle.x, particle.y, particle.size, 5);
          break;
          
        case 'leaf':
          this.drawLeaf(ctx, particle.x, particle.y, particle.size);
          break;
          
        default:
          ctx.fillRect(particle.x - particle.size/2, particle.y - particle.size/2, particle.size, particle.size);
      }
      
      ctx.restore();
    });
  }
  
  /**
   * 绘制菱形
   */
  drawDiamond(ctx, x, y, size) {
    ctx.beginPath();
    ctx.moveTo(x, y - size);
    ctx.lineTo(x + size, y);
    ctx.lineTo(x, y + size);
    ctx.lineTo(x - size, y);
    ctx.closePath();
    ctx.fill();
  }
  
  /**
   * 绘制星形
   */
  drawStar(ctx, x, y, size, points) {
    ctx.beginPath();
    for (let i = 0; i < points * 2; i++) {
      const angle = (i * Math.PI) / points - Math.PI / 2;
      const radius = i % 2 === 0 ? size : size / 2;
      const px = x + Math.cos(angle) * radius;
      const py = y + Math.sin(angle) * radius;
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
    ctx.fill();
  }
  
  /**
   * 绘制叶子形状
   */
  drawLeaf(ctx, x, y, size) {
    ctx.beginPath();
    ctx.moveTo(x, y - size);
    ctx.quadraticCurveTo(x + size, y - size/2, x + size/2, y);
    ctx.quadraticCurveTo(x + size, y + size/2, x, y + size);
    ctx.quadraticCurveTo(x - size, y + size/2, x - size/2, y);
    ctx.quadraticCurveTo(x - size, y - size/2, x, y - size);
    ctx.fill();
  }
  
  /**
   * 颜色插值
   */
  lerpColor(color1, color2, progress) {
    const c1 = this.parseColor(color1);
    const c2 = this.parseColor(color2);
    
    const r = Math.round(c1.r + (c2.r - c1.r) * progress);
    const g = Math.round(c1.g + (c2.g - c1.g) * progress);
    const b = Math.round(c1.b + (c2.b - c1.b) * progress);
    
    return `rgb(${r},${g},${b})`;
  }
  
  /**
   * 解析颜色
   */
  parseColor(color) {
    if (color.startsWith('#')) {
      const hex = color.slice(1);
      return {
        r: parseInt(hex.slice(0, 2), 16),
        g: parseInt(hex.slice(2, 4), 16),
        b: parseInt(hex.slice(4, 6), 16)
      };
    } else if (color.startsWith('rgb')) {
      const match = color.match(/(\d+),(\d+),(\d+)/);
      if (match) {
        return { r: parseInt(match[1]), g: parseInt(match[2]), b: parseInt(match[3]) };
      }
    }
    return { r: 255, g: 255, b: 255 };
  }
  
  /**
   * 屏幕震动
   */
  shakeScreen(intensity) {
    if (!this.canvas) return;
    
    const originalTransform = this.canvas.style.transform || '';
    let shakeCount = 0;
    const maxShakes = 5;
    
    const shake = () => {
      if (shakeCount >= maxShakes) {
        this.canvas.style.transform = originalTransform;
        return;
      }
      
      const offsetX = (Math.random() - 0.5) * intensity;
      const offsetY = (Math.random() - 0.5) * intensity;
      this.canvas.style.transform = `${originalTransform} translate(${offsetX}px, ${offsetY}px)`;
      
      shakeCount++;
      setTimeout(shake, 50);
    };
    
    shake();
  }
  
  /**
   * 停止特效
   */
  stopEffect(effectId) {
    const effect = this.activeEffects.get(effectId);
    if (effect) {
      effect.active = false;
      if (effect.emitInterval) {
        clearInterval(effect.emitInterval);
      }
      // 立即清除粒子
      effect.particles.forEach(p => p.active = false);
      this.activeEffects.delete(effectId);
    }
  }
  
  /**
   * 清除所有特效
   */
  clearAll() {
    this.activeEffects.forEach((effect, id) => {
      this.stopEffect(id);
    });
    this.particles = [];
  }
  
  /**
   * 添加自定义特效模板
   */
  addTemplate(name, config) {
    this.templates.set(name, config);
  }
  
  /**
   * 获取活动特效数量
   */
  getActiveCount() {
    return this.activeEffects.size;
  }
  
  /**
   * 获取粒子数量
   */
  getParticleCount() {
    return this.particles.length;
  }
}

// 创建全局单例（需要传入canvas）
window.ParticleSystem = ParticleSystem;