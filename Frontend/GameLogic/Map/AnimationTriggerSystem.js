/**
 * 动画触发器系统
 * 支持基于玩家位置、时间、事件等条件触发动画
 */
class AnimationTriggerSystem {
  constructor() {
    this.triggers = new Map(); // 存储所有触发器
    this.activeTriggers = new Set(); // 当前激活的触发器
    this.pendingEvents = []; // 待处理事件列表
    this.playerPosition = { x: 0, y: 0 };
    this.lastCheckTime = 0;
    this.checkInterval = 100; // 检查间隔（毫秒）
  }
  
  /**
   * 添加触发器
   */
  addTrigger(trigger) {
    const triggerId = trigger.id || `trigger_${Date.now()}_${Math.random()}`;
    
    const triggerData = {
      id: triggerId,
      type: trigger.type || 'proximity', // proximity, time, event, custom
      animationId: trigger.animationId,
      condition: trigger.condition || {},
      action: trigger.action || 'play', // play, stop, pause, toggle
      cooldown: trigger.cooldown || 0,
      lastTriggered: 0,
      repeat: trigger.repeat !== false, // 是否可重复触发
      priority: trigger.priority || 0,
      enabled: trigger.enabled !== false,
      ...trigger
    };
    
    this.triggers.set(triggerId, triggerData);
    return triggerId;
  }
  
  /**
   * 移除触发器
   */
  removeTrigger(triggerId) {
    this.triggers.delete(triggerId);
    this.activeTriggers.delete(triggerId);
  }
  
  /**
   * 更新玩家位置
   */
  updatePlayerPosition(x, y) {
    this.playerPosition = { x, y };
  }
  
  /**
   * 检查并触发动画
   */
  checkTriggers(animationSystem, currentTime = Date.now()) {
    // 限制检查频率
    if (currentTime - this.lastCheckTime < this.checkInterval) {
      return;
    }
    this.lastCheckTime = currentTime;
    
    this.triggers.forEach((trigger, triggerId) => {
      if (!trigger.enabled) return;
      
      // 检查冷却时间
      if (currentTime - trigger.lastTriggered < trigger.cooldown) {
        return;
      }
      
      // 检查触发条件
      if (this.shouldTrigger(trigger)) {
        this.executeTrigger(trigger, animationSystem);
        trigger.lastTriggered = currentTime;
        
        if (!trigger.repeat) {
          trigger.enabled = false;
        }
      }
    });
  }
  
  /**
   * 检查是否应该触发
   */
  shouldTrigger(trigger) {
    switch (trigger.type) {
      case 'proximity':
        return this.checkProximity(trigger);
      case 'time':
        return this.checkTime(trigger);
      case 'event':
        return this.checkEvent(trigger);
      case 'custom':
        return this.checkCustom(trigger);
      default:
        return false;
    }
  }
  
  /**
   * 检查距离触发
   */
  checkProximity(trigger) {
    const { x, y, radius } = trigger.condition;
    const distance = Math.hypot(
      this.playerPosition.x - x,
      this.playerPosition.y - y
    );
    
    return distance <= (radius || 3);
  }
  
  /**
   * 检查时间触发
   */
  checkTime(trigger) {
    const { startTime, endTime, interval } = trigger.condition;
    const currentTime = Date.now();
    
    if (interval) {
      // 周期性触发
      return currentTime % interval < 100;
    }
    
    if (startTime && endTime) {
      // 时间范围触发
      return currentTime >= startTime && currentTime <= endTime;
    }
    
    return false;
  }
  
  /**
   * 检查事件触发
   */
  checkEvent(trigger) {
    const { eventType, eventData } = trigger.condition;
    
    // 检查是否有匹配的事件
    return this.pendingEvents?.some(event => {
      if (event.type !== eventType) return false;
      
      // 检查事件数据
      if (eventData) {
        for (const key in eventData) {
          if (event.data[key] !== eventData[key]) {
            return false;
          }
        }
      }
      
      return true;
    }) || false;
  }
  
  /**
   * 检查自定义条件
   */
  checkCustom(trigger) {
    if (typeof trigger.condition.check === 'function') {
      return trigger.condition.check(this.playerPosition);
    }
    return false;
  }
  
  /**
   * 执行触发器
   */
  executeTrigger(trigger, animationSystem) {
    const animation = animationSystem.animations.get(trigger.animationId);
    if (!animation) return;
    
    switch (trigger.action) {
      case 'play':
        animationSystem.start();
        this.activeTriggers.add(trigger.id);
        break;
      case 'stop':
        animationSystem.stop();
        this.activeTriggers.delete(trigger.id);
        break;
      case 'pause':
        animationSystem.isRunning = false;
        break;
      case 'toggle':
        if (animationSystem.isRunning) {
          animationSystem.stop();
          this.activeTriggers.delete(trigger.id);
        } else {
          animationSystem.start();
          this.activeTriggers.add(trigger.id);
        }
        break;
      case 'restart':
        animationSystem.stop();
        animation.frame = 0;
        animationSystem.start();
        this.activeTriggers.add(trigger.id);
        break;
    }
    
    // 触发回调
    if (trigger.onTrigger) {
      trigger.onTrigger(trigger, animation);
    }
    
    console.log(`触发器 ${trigger.id} 已执行: ${trigger.action}`);
  }
  
  /**
   * 触发事件
   */
  triggerEvent(eventType, eventData = {}) {
    if (!this.pendingEvents) {
      this.pendingEvents = [];
    }
    
    this.pendingEvents.push({
      type: eventType,
      data: eventData,
      timestamp: Date.now()
    });
    
    // 清理过期事件
    this.pendingEvents = this.pendingEvents.filter(
      event => Date.now() - event.timestamp < 1000
    );
  }
  
  /**
   * 启用触发器
   */
  enableTrigger(triggerId) {
    const trigger = this.triggers.get(triggerId);
    if (trigger) {
      trigger.enabled = true;
    }
  }
  
  /**
   * 禁用触发器
   */
  disableTrigger(triggerId) {
    const trigger = this.triggers.get(triggerId);
    if (trigger) {
      trigger.enabled = false;
      this.activeTriggers.delete(triggerId);
    }
  }
  
  /**
   * 获取触发器信息
   */
  getTrigger(triggerId) {
    return this.triggers.get(triggerId);
  }
  
  /**
   * 获取所有触发器
   */
  getAllTriggers() {
    return Array.from(this.triggers.values());
  }
  
  /**
   * 获取激活的触发器
   */
  getActiveTriggers() {
    return Array.from(this.activeTriggers).map(id => this.triggers.get(id));
  }
  
  /**
   * 清除所有触发器
   */
  clearAll() {
    this.triggers.clear();
    this.activeTriggers.clear();
    this.pendingEvents = [];
  }
  
  /**
   * 导出触发器数据
   */
  exportTriggers() {
    const triggers = [];
    this.triggers.forEach((trigger) => {
      triggers.push({
        id: trigger.id,
        type: trigger.type,
        animationId: trigger.animationId,
        condition: trigger.condition,
        action: trigger.action,
        cooldown: trigger.cooldown,
        repeat: trigger.repeat,
        priority: trigger.priority,
        enabled: trigger.enabled
      });
    });
    return triggers;
  }
  
  /**
   * 导入触发器数据
   */
  importTriggers(triggers) {
    triggers.forEach(triggerData => {
      this.addTrigger(triggerData);
    });
  }
  
  /**
   * 创建预设触发器
   */
  createPresetTriggers() {
    const presets = [];
    
    // 靠近时播放动画
    presets.push({
      type: 'proximity',
      condition: { x: 10, y: 10, radius: 3 },
      action: 'play',
      cooldown: 1000,
      repeat: true
    });
    
    // 离开时停止动画
    presets.push({
      type: 'proximity',
      condition: { x: 10, y: 10, radius: 5 },
      action: 'stop',
      cooldown: 500,
      repeat: true
    });
    
    // 定时触发
    presets.push({
      type: 'time',
      condition: { interval: 5000 },
      action: 'restart',
      cooldown: 0,
      repeat: true
    });
    
    return presets;
  }
}

// 创建全局单例
window.AnimationTriggerSystem = new AnimationTriggerSystem();