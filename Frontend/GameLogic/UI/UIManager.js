/**
 * UI管理器
 * 管理所有UI组件的创建、显示、隐藏和交互
 */
class UIManager {
  constructor() {
    // UI组件容器
    this.components = new Map();
    
    // 当前显示的UI层级
    this.layers = {
      background: [],  // 背景层（小地图等）
      normal: [],      // 正常层（角色面板、聊天框等）
      popup: [],       // 弹窗层（对话框、提示框等）
      top: []          // 顶层（loading、系统提示等）
    };
    
    // UI配置
    this.config = {
      animationDuration: 300,  // 动画持续时间(ms)
      toastDuration: 3000,     // 提示框显示时间(ms)
      zIndexBase: 100          // z-index基数
    };
    
    // 主题配置
    this.theme = {
      primary: '#e94560',
      secondary: '#4a5568',
      success: '#4ade80',
      warning: '#fbbf24',
      danger: '#ef4444',
      info: '#60a5fa',
      dark: '#1a1a2e',
      light: '#ffffff',
      textPrimary: '#ffffff',
      textSecondary: '#999999',
      borderRadius: '8px',
      shadowColor: 'rgba(233, 69, 96, 0.3)'
    };
    
    // 初始化
    this.init();
  }
  
  /**
   * 初始化UI管理器
   */
  init() {
    // 创建UI容器
    this.createContainer();
    
    // 创建样式
    this.createStyles();
    
    // 创建常用UI组件
    this.createCommonComponents();
    
    console.log('UI管理器初始化完成');
  }
  
  /**
   * 创建UI容器
   */
  createContainer() {
    // 确保body存在
    if (!document.body) return;
    
    // 创建各层级容器
    Object.keys(this.layers).forEach(layerName => {
      const container = document.createElement('div');
      container.id = `ui-layer-${layerName}`;
      container.className = 'ui-layer';
      container.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        pointer-events: none;
        z-index: ${this.config.zIndexBase + this.getLayerIndex(layerName)};
      `;
      document.body.appendChild(container);
    });
  }
  
  /**
   * 获取层级索引
   */
  getLayerIndex(layerName) {
    const indices = { background: 0, normal: 100, popup: 200, top: 300 };
    return indices[layerName] || 0;
  }
  
  /**
   * 创建全局样式
   */
  createStyles() {
    const style = document.createElement('style');
    style.id = 'ui-global-styles';
    style.textContent = `
      /* UI组件基础样式 */
      .ui-component {
        position: absolute;
        pointer-events: auto;
        transition: all ${this.config.animationDuration}ms ease;
      }
      
      .ui-component.hidden {
        opacity: 0;
        transform: scale(0.95);
        pointer-events: none;
      }
      
      /* 按钮样式 */
      .ui-button {
        padding: 10px 20px;
        border: none;
        border-radius: ${this.theme.borderRadius};
        cursor: pointer;
        font-size: 14px;
        font-weight: 500;
        transition: all 0.2s ease;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
      }
      
      .ui-button:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px ${this.theme.shadowColor};
      }
      
      .ui-button:active {
        transform: translateY(0);
      }
      
      .ui-button.primary {
        background: linear-gradient(135deg, ${this.theme.primary}, #c23a51);
        color: ${this.theme.light};
      }
      
      .ui-button.secondary {
        background: linear-gradient(135deg, ${this.theme.secondary}, #2d3748);
        color: ${this.theme.light};
      }
      
      .ui-button.success {
        background: linear-gradient(135deg, ${this.theme.success}, #22c55e);
        color: ${this.theme.dark};
      }
      
      .ui-button.danger {
        background: linear-gradient(135deg, ${this.theme.danger}, #dc2626);
        color: ${this.theme.light};
      }
      
      .ui-button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
        transform: none;
      }
      
      /* 面板样式 */
      .ui-panel {
        background: rgba(0, 0, 0, 0.85);
        border: 2px solid ${this.theme.primary};
        border-radius: 12px;
        padding: 15px;
        box-shadow: 0 0 20px ${this.theme.shadowColor};
      }
      
      .ui-panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding-bottom: 10px;
        margin-bottom: 10px;
        border-bottom: 1px solid rgba(255,255,255,0.1);
      }
      
      .ui-panel-title {
        color: ${this.theme.primary};
        font-size: 16px;
        font-weight: bold;
      }
      
      .ui-panel-close {
        width: 24px;
        height: 24px;
        border: none;
        background: transparent;
        color: ${this.theme.textSecondary};
        cursor: pointer;
        font-size: 18px;
        transition: color 0.2s;
      }
      
      .ui-panel-close:hover {
        color: ${this.theme.primary};
      }
      
      /* 进度条样式 */
      .ui-progress {
        width: 100%;
        height: 8px;
        background: rgba(255,255,255,0.1);
        border-radius: 4px;
        overflow: hidden;
      }
      
      .ui-progress-bar {
        height: 100%;
        border-radius: 4px;
        transition: width 0.3s ease;
      }
      
      .ui-progress-bar.hp {
        background: linear-gradient(90deg, ${this.theme.danger}, #f87171);
      }
      
      .ui-progress-bar.mp {
        background: linear-gradient(90deg, ${this.theme.info}, #93c5fd);
      }
      
      .ui-progress-bar.exp {
        background: linear-gradient(90deg, ${this.theme.success}, #86efac);
      }
      
      /* Toast提示样式 */
      .ui-toast {
        position: fixed;
        top: 20px;
        left: 50%;
        transform: translateX(-50%);
        padding: 12px 24px;
        border-radius: ${this.theme.borderRadius};
        color: ${this.theme.light};
        font-size: 14px;
        z-index: ${this.config.zIndexBase + 400};
        animation: toastSlideIn 0.3s ease;
      }
      
      .ui-toast.success {
        background: rgba(74, 222, 128, 0.9);
        border: 1px solid ${this.theme.success};
      }
      
      .ui-toast.warning {
        background: rgba(251, 191, 36, 0.9);
        border: 1px solid ${this.theme.warning};
      }
      
      .ui-toast.danger {
        background: rgba(239, 68, 68, 0.9);
        border: 1px solid ${this.theme.danger};
      }
      
      .ui-toast.info {
        background: rgba(96, 165, 250, 0.9);
        border: 1px solid ${this.theme.info};
      }
      
      @keyframes toastSlideIn {
        from {
          opacity: 0;
          transform: translateX(-50%) translateY(-20px);
        }
        to {
          opacity: 1;
          transform: translateX(-50%) translateY(0);
        }
      }
      
      /* 对话框样式 */
      .ui-dialog {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        min-width: 300px;
        max-width: 500px;
        background: rgba(0, 0, 0, 0.9);
        border: 2px solid ${this.theme.primary};
        border-radius: 12px;
        padding: 20px;
        z-index: ${this.config.zIndexBase + 250};
        box-shadow: 0 0 30px ${this.theme.shadowColor};
      }
      
      .ui-dialog-overlay {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.5);
        z-index: ${this.config.zIndexBase + 240};
      }
      
      .ui-dialog-title {
        color: ${this.theme.primary};
        font-size: 18px;
        font-weight: bold;
        margin-bottom: 15px;
        text-align: center;
      }
      
      .ui-dialog-content {
        color: ${this.theme.textPrimary};
        font-size: 14px;
        margin-bottom: 20px;
        line-height: 1.6;
      }
      
      .ui-dialog-buttons {
        display: flex;
        justify-content: center;
        gap: 10px;
      }
      
      /* 输入框样式 */
      .ui-input {
        width: 100%;
        padding: 10px 15px;
        background: rgba(255, 255, 255, 0.1);
        border: 1px solid rgba(255, 255, 255, 0.2);
        border-radius: ${this.theme.borderRadius};
        color: ${this.theme.textPrimary};
        font-size: 14px;
        outline: none;
        transition: all 0.2s ease;
      }
      
      .ui-input:focus {
        border-color: ${this.theme.primary};
        box-shadow: 0 0 10px ${this.theme.shadowColor};
      }
      
      .ui-input::placeholder {
        color: ${this.theme.textSecondary};
      }
      
      /* 标签样式 */
      .ui-label {
        color: ${this.theme.textSecondary};
        font-size: 12px;
        margin-bottom: 5px;
        display: block;
      }
      
      /* 数值显示样式 */
      .ui-value {
        color: ${this.theme.success};
        font-size: 14px;
        font-weight: 500;
      }
      
      /* 技能冷却遮罩 */
      .ui-skill-cooldown {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.6);
        border-radius: inherit;
        display: flex;
        justify-content: center;
        align-items: center;
        color: ${this.theme.light};
        font-size: 14px;
        font-weight: bold;
      }
      
      /* 物品提示框 */
      .ui-tooltip {
        position: fixed;
        background: rgba(0, 0, 0, 0.9);
        border: 1px solid ${this.theme.primary};
        border-radius: ${this.theme.borderRadius};
        padding: 10px 15px;
        color: ${this.theme.textPrimary};
        font-size: 12px;
        max-width: 200px;
        z-index: ${this.config.zIndexBase + 350};
        pointer-events: none;
      }
      
      .ui-tooltip-title {
        color: ${this.theme.primary};
        font-size: 14px;
        font-weight: bold;
        margin-bottom: 5px;
      }
      
      .ui-tooltip-desc {
        color: ${this.theme.textSecondary};
        line-height: 1.4;
      }
      
      /* 滚动条样式 */
      .ui-scrollable::-webkit-scrollbar {
        width: 6px;
      }
      
      .ui-scrollable::-webkit-scrollbar-track {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 3px;
      }
      
      .ui-scrollable::-webkit-scrollbar-thumb {
        background: ${this.theme.primary};
        border-radius: 3px;
      }
      
      /* 动画效果 */
      .ui-fade-in {
        animation: fadeIn ${this.config.animationDuration}ms ease;
      }
      
      .ui-fade-out {
        animation: fadeOut ${this.config.animationDuration}ms ease;
      }
      
      .ui-scale-in {
        animation: scaleIn ${this.config.animationDuration}ms ease;
      }
      
      .ui-scale-out {
        animation: scaleOut ${this.config.animationDuration}ms ease;
      }
      
      @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
      }
      
      @keyframes fadeOut {
        from { opacity: 1; }
        to { opacity: 0; }
      }
      
      @keyframes scaleIn {
        from { opacity: 0; transform: scale(0.9); }
        to { opacity: 1; transform: scale(1); }
      }
      
      @keyframes scaleOut {
        from { opacity: 1; transform: scale(1); }
        to { opacity: 0; transform: scale(0.9); }
      }
      
      /* 闪烁效果 */
      .ui-flash {
        animation: flash 0.5s ease;
      }
      
      @keyframes flash {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
      }
      
      /* 抖动效果 */
      .ui-shake {
        animation: shake 0.5s ease;
      }
      
      @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-5px); }
        75% { transform: translateX(5px); }
      }
    `;
    document.head.appendChild(style);
  }
  
  /**
   * 创建常用UI组件
   */
  createCommonComponents() {
    // Toast容器
    this.toastContainer = document.createElement('div');
    this.toastContainer.id = 'ui-toast-container';
    this.toastContainer.style.cssText = `
      position: fixed;
      top: 20px;
      left: 50%;
      transform: translateX(-50%);
      z-index: ${this.config.zIndexBase + 400};
      display: flex;
      flex-direction: column;
      gap: 10px;
    `;
    document.body.appendChild(this.toastContainer);
  }
  
  /**
   * 创建按钮
   */
  createButton(options = {}) {
    const button = document.createElement('button');
    button.className = `ui-button ui-component ${options.type || 'primary'}`;
    button.textContent = options.text || '';
    button.style.cssText = options.style || '';
    
    if (options.onClick) {
      button.addEventListener('click', options.onClick);
    }
    
    if (options.disabled) {
      button.disabled = true;
    }
    
    if (options.icon) {
      const icon = document.createElement('span');
      icon.innerHTML = options.icon;
      button.insertBefore(icon, button.firstChild);
    }
    
    return button;
  }
  
  /**
   * 创建面板
   */
  createPanel(options = {}) {
    const panel = document.createElement('div');
    panel.className = 'ui-panel ui-component';
    panel.id = options.id || `panel-${Date.now()}`;
    
    // 标题栏
    if (options.title) {
      const header = document.createElement('div');
      header.className = 'ui-panel-header';
      
      const title = document.createElement('div');
      title.className = 'ui-panel-title';
      title.textContent = options.title;
      header.appendChild(title);
      
      if (options.closable) {
        const closeBtn = document.createElement('button');
        closeBtn.className = 'ui-panel-close';
        closeBtn.innerHTML = '×';
        closeBtn.addEventListener('click', () => this.hideComponent(panel.id));
        header.appendChild(closeBtn);
      }
      
      panel.appendChild(header);
    }
    
    // 内容区域
    const content = document.createElement('div');
    content.className = 'ui-panel-content';
    if (options.content) {
      content.innerHTML = options.content;
    }
    panel.appendChild(content);
    
    // 位置和大小
    if (options.x) panel.style.left = options.x + 'px';
    if (options.y) panel.style.top = options.y + 'px';
    if (options.width) panel.style.width = options.width + 'px';
    if (options.height) panel.style.minHeight = options.height + 'px';
    
    // 注册组件
    this.components.set(panel.id, {
      element: panel,
      layer: options.layer || 'normal',
      visible: false
    });
    
    return panel;
  }
  
  /**
   * 创建进度条
   */
  createProgressBar(options = {}) {
    const container = document.createElement('div');
    container.className = 'ui-progress ui-component';
    
    const bar = document.createElement('div');
    bar.className = `ui-progress-bar ${options.type || 'hp'}`;
    bar.style.width = `${options.value || 0}%`;
    
    container.appendChild(bar);
    
    // 文字显示
    if (options.showText) {
      const text = document.createElement('div');
      text.className = 'ui-progress-text';
      text.style.cssText = `
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        color: ${this.theme.light};
        font-size: 12px;
        font-weight: bold;
      `;
      text.textContent = options.text || '';
      container.appendChild(text);
      container.style.position = 'relative';
    }
    
    return { container, bar };
  }
  
  /**
   * 显示Toast提示
   */
  toast(message, type = 'info', duration = this.config.toastDuration) {
    const toast = document.createElement('div');
    toast.className = `ui-toast ${type}`;
    toast.textContent = message;
    
    this.toastContainer.appendChild(toast);
    
    // 自动移除
    setTimeout(() => {
      toast.style.animation = 'fadeOut 0.3s ease forwards';
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
      }, 300);
    }, duration);
    
    return toast;
  }
  
  /**
   * 显示对话框
   */
  showDialog(options = {}) {
    // 创建遮罩层
    const overlay = document.createElement('div');
    overlay.className = 'ui-dialog-overlay';
    overlay.id = `dialog-overlay-${Date.now()}`;
    
    // 创建对话框
    const dialog = document.createElement('div');
    dialog.className = 'ui-dialog ui-component ui-scale-in';
    
    // 标题
    if (options.title) {
      const title = document.createElement('div');
      title.className = 'ui-dialog-title';
      title.textContent = options.title;
      dialog.appendChild(title);
    }
    
    // 内容
    const content = document.createElement('div');
    content.className = 'ui-dialog-content';
    content.innerHTML = options.content || '';
    dialog.appendChild(content);
    
    // 按钮
    const buttons = document.createElement('div');
    buttons.className = 'ui-dialog-buttons';
    
    if (options.buttons) {
      options.buttons.forEach(btn => {
        const button = this.createButton({
          text: btn.text,
          type: btn.type || 'secondary',
          onClick: () => {
            if (btn.onClick) btn.onClick();
            this.closeDialog(overlay.id);
          }
        });
        buttons.appendChild(button);
      });
    } else {
      // 默认确认按钮
      const confirmBtn = this.createButton({
        text: '确定',
        type: 'primary',
        onClick: () => this.closeDialog(overlay.id)
      });
      buttons.appendChild(confirmBtn);
    }
    
    dialog.appendChild(buttons);
    
    // 添加到DOM
    document.body.appendChild(overlay);
    document.body.appendChild(dialog);
    
    // 点击遮罩关闭
    overlay.addEventListener('click', () => {
      if (options.closeOnClickOverlay !== false) {
        this.closeDialog(overlay.id);
      }
    });
    
    // 注册组件
    const dialogId = `dialog-${Date.now()}`;
    this.components.set(dialogId, {
      element: dialog,
      overlay: overlay,
      layer: 'popup',
      visible: true
    });
    
    return { overlay, dialog, id: dialogId };
  }
  
  /**
   * 关闭对话框
   */
  closeDialog(overlayId) {
    const overlay = document.getElementById(overlayId);
    if (overlay) {
      // 找到对应的dialog
      const dialog = overlay.nextElementSibling;
      if (dialog && dialog.classList.contains('ui-dialog')) {
        dialog.classList.remove('ui-scale-in');
        dialog.classList.add('ui-scale-out');
        overlay.style.animation = 'fadeOut 0.3s ease forwards';
        
        setTimeout(() => {
          if (overlay.parentNode) overlay.parentNode.removeChild(overlay);
          if (dialog.parentNode) dialog.parentNode.removeChild(dialog);
        }, 300);
      }
    }
  }
  
  /**
   * 显示组件
   */
  showComponent(componentId, layer = 'normal') {
    const component = this.components.get(componentId);
    if (!component) return;
    
    const container = document.getElementById(`ui-layer-${layer}`);
    if (container && !component.element.parentNode) {
      container.appendChild(component.element);
    }
    
    component.element.classList.remove('hidden');
    component.element.classList.add('ui-fade-in');
    component.visible = true;
    component.layer = layer;
  }
  
  /**
   * 隐藏组件
   */
  hideComponent(componentId) {
    const component = this.components.get(componentId);
    if (!component) return;
    
    component.element.classList.remove('ui-fade-in');
    component.element.classList.add('ui-fade-out', 'hidden');
    component.visible = false;
    
    setTimeout(() => {
      if (component.element.parentNode) {
        component.element.parentNode.removeChild(component.element);
      }
    }, this.config.animationDuration);
  }
  
  /**
   * 创建物品提示框
   */
  showTooltip(options = {}, x, y) {
    // 移除已存在的tooltip
    this.hideTooltip();
    
    const tooltip = document.createElement('div');
    tooltip.className = 'ui-tooltip';
    tooltip.id = 'ui-active-tooltip';
    
    if (options.title) {
      const title = document.createElement('div');
      title.className = 'ui-tooltip-title';
      title.textContent = options.title;
      tooltip.appendChild(title);
    }
    
    if (options.desc) {
      const desc = document.createElement('div');
      desc.className = 'ui-tooltip-desc';
      desc.innerHTML = options.desc;
      tooltip.appendChild(desc);
    }
    
    // 位置
    tooltip.style.left = (x + 10) + 'px';
    tooltip.style.top = (y + 10) + 'px';
    
    document.body.appendChild(tooltip);
    
    return tooltip;
  }
  
  /**
   * 隐藏物品提示框
   */
  hideTooltip() {
    const tooltip = document.getElementById('ui-active-tooltip');
    if (tooltip && tooltip.parentNode) {
      tooltip.parentNode.removeChild(tooltip);
    }
  }
  
  /**
   * 创建输入框
   */
  createInput(options = {}) {
    const container = document.createElement('div');
    container.className = 'ui-input-container';
    
    if (options.label) {
      const label = document.createElement('label');
      label.className = 'ui-label';
      label.textContent = options.label;
      container.appendChild(label);
    }
    
    const input = document.createElement('input');
    input.className = 'ui-input';
    input.type = options.type || 'text';
    input.placeholder = options.placeholder || '';
    input.value = options.value || '';
    
    if (options.maxLength) {
      input.maxLength = options.maxLength;
    }
    
    if (options.onChange) {
      input.addEventListener('change', options.onChange);
    }
    
    if (options.onInput) {
      input.addEventListener('input', options.onInput);
    }
    
    container.appendChild(input);
    
    return { container, input };
  }
  
  /**
   * 更新进度条
   */
  updateProgressBar(bar, value, text = null) {
    bar.style.width = `${Math.min(100, Math.max(0, value))}%`;
    
    const textEl = bar.parentNode.querySelector('.ui-progress-text');
    if (textEl && text !== null) {
      textEl.textContent = text;
    }
  }
  
  /**
   * 添加动画效果
   */
  addAnimation(element, animationName) {
    element.classList.add(`ui-${animationName}`);
    setTimeout(() => {
      element.classList.remove(`ui-${animationName}`);
    }, 500);
  }
  
  /**
   * 设置主题
   */
  setTheme(newTheme) {
    Object.assign(this.theme, newTheme);
    // 重新创建样式
    const oldStyle = document.getElementById('ui-global-styles');
    if (oldStyle) {
      oldStyle.parentNode.removeChild(oldStyle);
    }
    this.createStyles();
  }
  
  /**
   * 获取组件
   */
  getComponent(componentId) {
    return this.components.get(componentId);
  }
  
  /**
   * 销毁组件
   */
  destroyComponent(componentId) {
    const component = this.components.get(componentId);
    if (component) {
      if (component.element.parentNode) {
        component.element.parentNode.removeChild(component.element);
      }
      if (component.overlay && component.overlay.parentNode) {
        component.overlay.parentNode.removeChild(component.overlay);
      }
      this.components.delete(componentId);
    }
  }
  
  /**
   * 清理所有组件
   */
  clearAll() {
    this.components.forEach((component, id) => {
      this.destroyComponent(id);
    });
    
    // 清理Toast
    while (this.toastContainer.firstChild) {
      this.toastContainer.removeChild(this.toastContainer.firstChild);
    }
  }
}

// 创建全局单例
window.UIManager = new UIManager();