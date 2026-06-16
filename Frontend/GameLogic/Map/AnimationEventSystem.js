/**
 * 动画事件系统
 * 支持动画生命周期中的各种事件处理
 */
class AnimationEventSystem {
  constructor() {
    this.eventListeners = new Map(); // 事件监听器
    this.animationEvents = new Map(); // 动画事件映射
    this.eventQueue = []; // 事件队列
    this.isProcessing = false;
  }
  
  /**
   * 事件类型定义
   */
  static get EventTypes() {
    return {
      // 动画生命周期事件
      ANIMATION_START: 'animation_start',
      ANIMATION_END: 'animation_end',
      ANIMATION_PAUSE: 'animation_pause',
      ANIMATION_RESUME: 'animation_resume',
      ANIMATION_LOOP: 'animation_loop',
      
      // 帧事件
      FRAME_START: 'frame_start',
      FRAME_END: 'frame_end',
      FRAME_CHANGE: 'frame_change',
      
      // 触发器事件
      TRIGGER_ACTIVATED: 'trigger_activated',
      TRIGGER_DEACTIVATED: 'trigger_deactivated',
      
      // 交互事件
      PLAYER_ENTER: 'player_enter',
      PLAYER_LEAVE: 'player_leave',
      PLAYER_INTERACT: 'player_interact',
      
      // 自定义事件
      CUSTOM: 'custom'
    };
  }
  
  /**
   * 添加事件监听器
   */
  addEventListener(eventType, callback, priority = 0) {
    if (!this.eventListeners.has(eventType)) {
      this.eventListeners.set(eventType, []);
    }
    
    this.eventListeners.get(eventType).push({
      callback,
      priority,
      once: false,
      enabled: true
    });
    
    // 按优先级排序
    this.eventListeners.get(eventType).sort((a, b) => b.priority - a.priority);
  }
  
  /**
   * 添加一次性事件监听器
   */
  addOnceEventListener(eventType, callback, priority = 0) {
    if (!this.eventListeners.has(eventType)) {
      this.eventListeners.set(eventType, []);
    }
    
    this.eventListeners.get(eventType).push({
      callback,
      priority,
      once: true,
      enabled: true
    });
    
    this.eventListeners.get(eventType).sort((a, b) => b.priority - a.priority);
  }
  
  /**
   * 移除事件监听器
   */
  removeEventListener(eventType, callback) {
    if (!this.eventListeners.has(eventType)) return;
    
    const listeners = this.eventListeners.get(eventType);
    const index = listeners.findIndex(l => l.callback === callback);
    
    if (index !== -1) {
      listeners.splice(index, 1);
    }
  }
  
  /**
   * 触发事件
   */
  triggerEvent(eventType, eventData = {}) {
    const event = {
      type: eventType,
      data: eventData,
      timestamp: Date.now(),
      propagationStopped: false
    };
    
    // 添加到事件队列
    this.eventQueue.push(event);
    
    // 如果没有在处理，开始处理
    if (!this.isProcessing) {
      this.processEventQueue();
    }
  }
  
  /**
   * 处理事件队列
   */
  async processEventQueue() {
    this.isProcessing = true;
    
    while (this.eventQueue.length > 0) {
      const event = this.eventQueue.shift();
      await this.dispatchEvent(event);
    }
    
    this.isProcessing = false;
  }
  
  /**
   * 分发事件
   */
  async dispatchEvent(event) {
    const listeners = this.eventListeners.get(event.type);
    if (!listeners) return;
    
    for (const listener of listeners) {
      if (!listener.enabled) continue;
      
      try {
        const result = await listener.callback(event);
        
        // 如果返回false，停止事件传播
        if (result === false) {
          event.propagationStopped = true;
          break;
        }
        
        // 如果是一次性监听器，移除它
        if (listener.once) {
          listener.enabled = false;
        }
      } catch (error) {
        console.error(`事件处理错误 [${event.type}]:`, error);
      }
    }
    
    // 清理一次性监听器
    this.cleanupOnceListeners(event.type);
  }
  
  /**
   * 清理一次性监听器
   */
  cleanupOnceListeners(eventType) {
    const listeners = this.eventListeners.get(eventType);
    if (!listeners) return;
    
    const activeListeners = listeners.filter(l => !l.once || l.enabled);
    this.eventListeners.set(eventType, activeListeners);
  }
  
  /**
   * 绑定动画事件
   */
  bindAnimationEvent(animationId, eventType, condition, action) {
    if (!this.animationEvents.has(animationId)) {
      this.animationEvents.set(animationId, []);
    }
    
    this.animationEvents.get(animationId).push({
      eventType,
      condition,
      action
    });
  }
  
  /**
   * 检查并触发动画事件
   */
  checkAnimationEvents(animationId, currentFrame, animationState) {
    const events = this.animationEvents.get(animationId);
    if (!events) return;
    
    events.forEach(eventConfig => {
      if (this.shouldTriggerEvent(eventConfig, currentFrame, animationState)) {
        this.triggerEvent(eventConfig.eventType, {
          animationId,
          frame: currentFrame,
          state: animationState,
          action: eventConfig.action
        });
        
        // 执行关联动作
        if (eventConfig.action) {
          eventConfig.action(animationId, currentFrame, animationState);
        }
      }
    });
  }
  
  /**
   * 检查是否应该触发事件
   */
  shouldTriggerEvent(eventConfig, currentFrame, animationState) {
    const { eventType, condition } = eventConfig;
    
    // 检查条件
    if (condition) {
      if (typeof condition === 'function') {
        return condition(currentFrame, animationState);
      } else if (condition.frame !== undefined) {
        return currentFrame === condition.frame;
      } else if (condition.frameRange) {
        const { start, end } = condition.frameRange;
        return currentFrame >= start && currentFrame <= end;
      }
    }
    
    return true;
  }
  
  /**
   * 创建预设事件配置
   */
  createPresetEvents(animationId, animationType) {
    const presets = this.getPresetEventConfigs(animationType);
    
    presets.forEach(preset => {
      this.bindAnimationEvent(
        animationId,
        preset.eventType,
        preset.condition,
        preset.action
      );
    });
  }
  
  /**
   * 获取预设事件配置
   */
  getPresetEventConfigs(animationType) {
    const configs = [];
    
    switch (animationType) {
      case 'fire':
        configs.push({
          eventType: AnimationEventSystem.EventTypes.FRAME_CHANGE,
          condition: { frame: 0 },
          action: (animId, frame) => {
            console.log(`火焰动画 ${animId} 开始新循环`);
          }
        });
        break;
        
      case 'magic_circle':
        configs.push({
          eventType: AnimationEventSystem.EventTypes.ANIMATION_LOOP,
          condition: null,
          action: (animId, frame) => {
            console.log(`魔法阵 ${animId} 完成一次循环`);
          }
        });
        break;
        
      case 'waterfall':
        configs.push({
          eventType: AnimationEventSystem.EventTypes.FRAME_CHANGE,
          condition: { frameRange: { start: 0, end: 2 } },
          action: (animId, frame) => {
            // 瀑布水花效果
          }
        });
        break;
    }
    
    return configs;
  }
  
  /**
   * 暂停事件处理
   */
  pause() {
    this.isProcessing = false;
  }
  
  /**
   * 恢复事件处理
   */
  resume() {
    if (!this.isProcessing && this.eventQueue.length > 0) {
      this.processEventQueue();
    }
  }
  
  /**
   * 清空事件队列
   */
  clearQueue() {
    this.eventQueue = [];
  }
  
  /**
   * 清除所有监听器
   */
  clearAllListeners() {
    this.eventListeners.clear();
    this.animationEvents.clear();
    this.eventQueue = [];
  }
  
  /**
   * 获取事件统计
   */
  getStats() {
    let totalListeners = 0;
    this.eventListeners.forEach(listeners => {
      totalListeners += listeners.length;
    });
    
    return {
      totalEventTypes: this.eventListeners.size,
      totalListeners,
      queuedEvents: this.eventQueue.length,
      boundAnimations: this.animationEvents.size,
      isProcessing: this.isProcessing
    };
  }
  
  /**
   * 导出事件配置
   */
  exportEvents() {
    const events = [];
    
    this.animationEvents.forEach((eventList, animationId) => {
      eventList.forEach(eventConfig => {
        events.push({
          animationId,
          eventType: eventConfig.eventType,
          condition: eventConfig.condition,
          hasAction: !!eventConfig.action
        });
      });
    });
    
    return events;
  }
  
  /**
   * 导入事件配置
   */
  importEvents(events) {
    events.forEach(eventConfig => {
      this.bindAnimationEvent(
        eventConfig.animationId,
        eventConfig.eventType,
        eventConfig.condition,
        null // 动作函数无法序列化，需要重新绑定
      );
    });
  }
}

// 创建全局单例
window.AnimationEventSystem = new AnimationEventSystem();