/**
 * UI动态加载器
 * 负责从Frontend/UI目录加载UI配置并动态应用到UI组件
 */
class UIDynamicLoader {
  constructor(game) {
    this.game = game;
    this.configManager = null;
    this.uiElements = {};
    this.initialized = false;
  }

  /**
   * 初始化
   */
  async init() {
    if (this.initialized) return;

    // 创建配置管理器
    this.configManager = new UIConfigManager(this.game);

    // 加载所有配置
    await this.configManager.loadAll();

    this.initialized = true;
    console.log('UI动态加载器初始化完成');
  }

  /**
   * 应用UI配置到元素
   */
  applyConfigToElement(moduleName, componentName, element) {
    if (!this.configManager) return;

    const config = this.configManager.getComponentConfig(moduleName, componentName);
    if (!config || !element) return;

    // 应用基础样式
    this.applyStyles(element, config);

    // 应用位置
    this.applyPosition(element, config);

    // 应用尺寸
    this.applySize(element, config);

    // 存储配置引用
    element.uiConfig = config;
  }

  /**
   * 应用样式
   */
  applyStyles(element, config) {
    if (!config.style) return;

    const style = config.style;

    if (style.background) element.style.background = style.background;
    if (style.border) {
      element.style.border = style.border;
      // 尝试从border字符串中提取颜色
      const colorMatch = style.border.match(/solid\s+(.+)/);
      if (colorMatch) {
        element.dataset.borderColor = colorMatch[1];
      }
    }
    if (style.borderRadius !== undefined) {
      element.style.borderRadius = (typeof style.borderRadius === 'number' ? style.borderRadius : parseInt(style.borderRadius)) + 'px';
    }
    if (style.opacity !== undefined) element.style.opacity = style.opacity;
    if (style.padding) element.style.padding = style.padding;
    if (style.margin) element.style.margin = style.margin;
    if (style.boxShadow) element.style.boxShadow = style.boxShadow;
    if (style.fontSize) element.style.fontSize = (typeof style.fontSize === 'number' ? style.fontSize : parseInt(style.fontSize)) + 'px';
    if (style.fontFamily) element.style.fontFamily = style.fontFamily;
    if (style.color) element.style.color = style.color;
    if (style.textAlign) element.style.textAlign = style.textAlign;
    if (style.zIndex) element.style.zIndex = style.zIndex;
  }

  /**
   * 应用位置
   */
  applyPosition(element, config) {
    if (!config.position) return;

    const pos = config.position;

    // 清除现有位置样式
    element.style.left = '';
    element.style.top = '';
    element.style.right = '';
    element.style.bottom = '';
    element.style.transform = '';

    if (pos.x !== undefined) {
      if (pos.x === 'center') {
        element.style.left = '50%';
        element.style.transform = (element.style.transform || '') + ' translateX(-50%)';
      } else if (pos.x === 'auto') {
        element.style.left = 'auto';
      } else {
        element.style.left = typeof pos.x === 'number' ? pos.x + 'px' : pos.x;
      }
    }

    if (pos.y !== undefined) {
      if (pos.y === 'center') {
        element.style.top = '50%';
        element.style.transform = (element.style.transform || '') + ' translateY(-50%)';
      } else if (pos.y === 'auto') {
        element.style.top = 'auto';
      } else {
        element.style.top = typeof pos.y === 'number' ? pos.y + 'px' : pos.y;
      }
    }

    if (pos.right !== undefined) element.style.right = pos.right + 'px';
    if (pos.bottom !== undefined) element.style.bottom = pos.bottom + 'px';
  }

  /**
   * 应用尺寸
   */
  applySize(element, config) {
    if (!config.size) return;

    const size = config.size;

    if (size.width !== undefined) {
      element.style.width = typeof size.width === 'number' ? size.width + 'px' : size.width;
    }
    if (size.height !== undefined) {
      element.style.height = typeof size.height === 'number' ? size.height + 'px' : size.height;
    }
  }

  /**
   * 创建UI组件
   */
  createComponent(moduleName, componentName, options = {}) {
    const config = this.configManager?.getComponentConfig(moduleName, componentName);
    const element = document.createElement(options.tag || 'div');

    // 应用配置
    if (config) {
      this.applyConfigToElement(moduleName, componentName, element);
    }

    // 应用额外选项
    if (options.id) element.id = options.id;
    if (options.className) element.className = options.className;
    if (options.innerHTML) element.innerHTML = options.innerHTML;
    if (options.textContent) element.textContent = options.textContent;

    // 添加到DOM
    if (options.parent) {
      options.parent.appendChild(element);
    }

    return element;
  }

  /**
   * 注册UI元素
   */
  registerElement(name, element) {
    this.uiElements[name] = element;
  }

  /**
   * 获取UI元素
   */
  getElement(name) {
    return this.uiElements[name];
  }

  /**
   * 应用主题
   */
  applyTheme() {
    if (!this.configManager?.theme) return;

    const root = document.documentElement;
    const theme = this.configManager.theme.theme || this.configManager.theme;

    // CSS变量
    root.style.setProperty('--ui-primary', theme.primary);
    root.style.setProperty('--ui-secondary', theme.secondary);
    root.style.setProperty('--ui-success', theme.success);
    root.style.setProperty('--ui-warning', theme.warning);
    root.style.setProperty('--ui-danger', theme.danger);
    root.style.setProperty('--ui-info', theme.info);
    root.style.setProperty('--ui-dark', theme.dark);
    root.style.setProperty('--ui-light', theme.light);
    root.style.setProperty('--ui-text', theme.text);
    root.style.setProperty('--ui-text-muted', theme.textMuted);

    // 应用动画配置
    if (theme.animations) {
      root.style.setProperty('--ui-duration-fast', theme.animations.duration?.fast || 150 + 'ms');
      root.style.setProperty('--ui-duration-normal', theme.animations.duration?.normal || 300 + 'ms');
      root.style.setProperty('--ui-duration-slow', theme.animations.duration?.slow || 500 + 'ms');
    }
  }

  /**
   * 获取颜色
   */
  getColor(type) {
    const theme = this.configManager?.theme?.theme || {};
    return theme[type] || this.getDefaultColor(type);
  }

  /**
   * 获取默认颜色
   */
  getDefaultColor(type) {
    const colors = {
      primary: '#e94560',
      secondary: '#4a5568',
      success: '#4ade80',
      warning: '#fbbf24',
      danger: '#ef4444',
      info: '#60a5fa',
      text: '#ffffff',
      textMuted: '#999999'
    };
    return colors[type] || '#ffffff';
  }

  /**
   * 获取质量颜色
   */
  getQualityColor(quality) {
    const colors = {
      1: '#ffffff',
      2: '#4ade80',
      3: '#60a5fa',
      4: '#a855f7',
      5: '#fbbf24'
    };
    return colors[quality] || '#ffffff';
  }

  /**
   * 获取动画样式
   */
  getAnimationStyle(type = 'default') {
    const animations = this.configManager?.getAnimationConfig() || {};
    const easing = animations.easing?.[type] || 'ease-out';
    const duration = animations.duration?.normal || 300;

    return `animation: ui-${type} ${duration}ms ${easing}`;
  }

  /**
   * 创建动画关键帧（带缓存复用）
   */
  createAnimation(keyframes, duration = 300, easing = 'ease-out') {
    // 使用缓存避免重复创建相同动画
    if (!this.animationCache) {
      this.animationCache = new Map();
    }

    // 根据关键帧内容生成唯一key
    const cacheKey = `${keyframes}-${duration}-${easing}`;
    
    // 如果缓存中已有相同动画，直接返回
    if (this.animationCache.has(cacheKey)) {
      return this.animationCache.get(cacheKey);
    }

    const styleId = 'ui-dynamic-animation';
    let styleEl = document.getElementById(styleId);

    if (!styleEl) {
      styleEl = document.createElement('style');
      styleEl.id = styleId;
      document.head.appendChild(styleEl);
    }

    const animationName = 'ui-anim-' + Date.now();
    const keyframeCSS = `
      @keyframes ${animationName} {
        ${keyframes}
      }
    `;

    styleEl.textContent += keyframeCSS;

    const result = {
      name: animationName,
      duration: duration + 'ms',
      easing: easing
    };

    // 缓存结果
    this.animationCache.set(cacheKey, result);

    return result;
  }

  /**
   * 清理动画样式（用于页面切换或内存管理）
   */
  clearAnimations() {
    const styleId = 'ui-dynamic-animation';
    const styleEl = document.getElementById(styleId);
    if (styleEl) {
      styleEl.textContent = '';
    }
    if (this.animationCache) {
      this.animationCache.clear();
    }
  }

  /**
   * 刷新所有UI
   */
  refreshAll() {
    for (const name in this.uiElements) {
      const element = this.uiElements[name];
      if (element.uiConfig) {
        this.applyStyles(element, element.uiConfig);
      }
    }
  }

  /**
   * 导出配置
   */
  exportModuleConfig(moduleName) {
    return this.configManager?.exportConfig(moduleName);
  }

  /**
   * 获取模块列表
   */
  getModuleList() {
    return Object.keys(this.configManager?.configs || {});
  }
}

// 导出到全局
window.UIDynamicLoader = UIDynamicLoader;
