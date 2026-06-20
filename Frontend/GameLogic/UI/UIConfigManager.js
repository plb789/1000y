/**
 * UI配置管理器
 * 统一管理UI配置的加载、应用和个性化设置
 */
class UIConfigManager {
  constructor(game) {
    this.game = game;
    this.configs = {};
    this.theme = {};
    this.loaded = false;
    this.moduleOrder = [
      'Common',
      'RoleUI',
      'SkillBarUI',
      'MiniMapUI',
      'ChatUI',
      'InventoryUI',
      'ShopUI',
      'SocialUI',
      'SettingsUI'
    ];
  }

  /**
   * 加载所有UI配置
   */
  async loadAll() {
    if (this.loaded) return;

    console.log('开始加载UI配置...');

    // 加载主题配置
    await this.loadTheme();

    // 加载UI模块配置
    for (const moduleName of this.moduleOrder) {
      if (moduleName === 'Common') continue;
      await this.loadModule(moduleName);
    }

    this.loaded = true;
    console.log('UI配置加载完成');
  }

  /**
   * 加载主题配置
   */
  async loadTheme() {
    try {
      const response = await fetch('/UI/Common/theme.json');
      if (response.ok) {
        this.theme = await response.json();
        this.applyTheme();
      }
    } catch (e) {
      console.log('加载主题失败，使用默认主题');
      this.theme = this.getDefaultTheme();
      this.applyTheme();
    }
  }

  /**
   * 加载单个UI模块配置
   */
  async loadModule(moduleName) {
    try {
      const response = await fetch(`/UI/${moduleName}/config.json`);
      if (response.ok) {
        this.configs[moduleName] = await response.json();
        console.log(`加载 ${moduleName} 配置成功`);
      } else {
        console.log(`${moduleName} 配置文件不存在，使用默认配置`);
        this.configs[moduleName] = this.getDefaultConfig(moduleName);
      }
    } catch (e) {
      console.log(`加载 ${moduleName} 失败:`, e);
      this.configs[moduleName] = this.getDefaultConfig(moduleName);
    }
  }

  /**
   * 获取UI配置
   */
  getConfig(moduleName) {
    return this.configs[moduleName] || this.getDefaultConfig(moduleName);
  }

  /**
   * 获取组件配置
   */
  getComponentConfig(moduleName, componentName) {
    const config = this.getConfig(moduleName);
    return config.components?.[componentName] || this.getDefaultComponentConfig(componentName);
  }

  /**
   * 应用主题到CSS变量
   */
  applyTheme() {
    if (!this.theme || !this.theme.theme) return;

    const root = document.documentElement;
    const t = this.theme.theme;

    root.style.setProperty('--ui-primary', t.primary || '#e94560');
    root.style.setProperty('--ui-secondary', t.secondary || '#4a5568');
    root.style.setProperty('--ui-success', t.success || '#4ade80');
    root.style.setProperty('--ui-warning', t.warning || '#fbbf24');
    root.style.setProperty('--ui-danger', t.danger || '#ef4444');
    root.style.setProperty('--ui-info', t.info || '#60a5fa');
    root.style.setProperty('--ui-dark', t.dark || '#1a202c');
    root.style.setProperty('--ui-light', t.light || '#f7fafc');
    root.style.setProperty('--ui-background', t.background || 'rgba(0, 0, 0, 0.8)');
    root.style.setProperty('--ui-text', t.text || '#ffffff');
    root.style.setProperty('--ui-text-muted', t.textMuted || '#999999');
  }

  /**
   * 应用UI组件样式
   */
  applyComponentStyle(moduleName, componentName, element) {
    const config = this.getComponentConfig(moduleName, componentName);
    if (!config || !element) return;

    // 应用位置
    if (config.position) {
      const pos = config.position;
      if (pos.x !== undefined) {
        if (pos.x === 'center') {
          element.style.left = '50%';
          element.style.transform = 'translateX(-50%)';
        } else if (pos.x === 'auto') {
          element.style.left = 'auto';
        } else {
          element.style.left = typeof pos.x === 'number' ? pos.x + 'px' : pos.x;
        }
      }
      if (pos.y !== undefined) {
        if (pos.y === 'center') {
          element.style.top = '50%';
          element.style.transform = 'translateY(-50%)';
        } else if (pos.y === 'auto') {
          element.style.top = 'auto';
        } else {
          element.style.top = typeof pos.y === 'number' ? pos.y + 'px' : pos.y;
        }
      }
      if (pos.right !== undefined) element.style.right = pos.right + 'px';
      if (pos.bottom !== undefined) element.style.bottom = pos.bottom + 'px';
    }

    // 应用尺寸
    if (config.size) {
      if (config.size.width !== undefined) {
        element.style.width = typeof config.size.width === 'number' ? config.size.width + 'px' : config.size.width;
      }
      if (config.size.height !== undefined) {
        element.style.height = typeof config.size.height === 'number' ? config.size.height + 'px' : config.size.height;
      }
    }

    // 应用样式
    if (config.style) {
      if (config.style.background) element.style.background = config.style.background;
      if (config.style.border) element.style.border = config.style.border;
      if (config.style.borderRadius !== undefined) element.style.borderRadius = config.style.borderRadius + 'px';
      if (config.style.opacity !== undefined) element.style.opacity = config.style.opacity;
    }

    // 应用z-index（使用括号明确运算符优先级）
    if (config.zIndex !== undefined) {
      element.style.zIndex = (this.theme.zIndex?.base || 0) + config.zIndex;
    }
  }

  /**
   * 获取动画配置
   */
  getAnimationConfig() {
    return this.theme.animations || {
      duration: { fast: 150, normal: 300, slow: 500 },
      easing: { default: 'ease-out', bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)' }
    };
  }

  /**
   * 创建样式字符串
   */
  createStyleString(moduleName, componentName, baseStyle = '') {
    const config = this.getComponentConfig(moduleName, componentName);
    if (!config) return baseStyle;

    let style = baseStyle;

    if (config.style) {
      if (config.style.background) style += `background: ${config.style.background};`;
      if (config.style.border) style += `border: ${config.style.border};`;
      if (config.style.borderRadius !== undefined) style += `border-radius: ${config.style.borderRadius}px;`;
      if (config.style.opacity !== undefined) style += `opacity: ${config.style.opacity};`;
    }

    return style;
  }

  /**
   * 获取默认主题
   */
  getDefaultTheme() {
    return {
      primary: '#e94560',
      secondary: '#4a5568',
      success: '#4ade80',
      warning: '#fbbf24',
      danger: '#ef4444',
      info: '#60a5fa',
      dark: '#1a202c',
      light: '#f7fafc',
      background: 'rgba(0, 0, 0, 0.8)',
      text: '#ffffff',
      textMuted: '#999999'
    };
  }

  /**
   * 获取默认模块配置
   */
  getDefaultConfig(moduleName) {
    const defaults = {
      RoleUI: {
        name: 'RoleUI',
        version: '1.0',
        components: {
          rolePanel: {
            enabled: true,
            position: { x: 20, y: 60 },
            size: { width: 220, height: 'auto' },
            style: {
              background: 'rgba(0, 0, 0, 0.8)',
              border: '2px solid #e94560',
              borderRadius: 10,
              padding: '15px'
            }
          }
        }
      },
      ChatUI: {
        name: 'ChatUI',
        version: '1.0',
        components: {
          chatBox: {
            enabled: true,
            position: { x: 20, y: 'auto', bottom: 90 },
            size: { width: 350, height: 200 },
            style: {
              background: 'rgba(0, 0, 0, 0.7)',
              border: '1px solid #444',
              borderRadius: 10
            }
          }
        }
      },
      SkillBarUI: {
        name: 'SkillBarUI',
        version: '1.0',
        components: {
          skillBar: {
            enabled: true,
            position: { x: 'center', y: 'auto', bottom: 0 },
            size: { height: 80 },
            style: {
              background: 'linear-gradient(0deg, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0))'
            }
          }
        }
      },
      InventoryUI: {
        name: 'InventoryUI',
        version: '1.0',
        components: {
          inventoryPanel: {
            enabled: true,
            position: { x: 'center', y: 'center' },
            size: { width: 480, height: 'auto' },
            style: {
              background: 'rgba(20, 20, 30, 0.95)',
              border: '2px solid #e94560',
              borderRadius: 15
            }
          }
        }
      }
    };

    return defaults[moduleName] || { name: moduleName, version: '1.0', components: {} };
  }

  /**
   * 获取默认组件配置
   */
  getDefaultComponentConfig(componentName) {
    return {
      enabled: true,
      position: { x: 0, y: 0 },
      size: { width: 100, height: 100 },
      style: {
        background: 'rgba(0, 0, 0, 0.8)',
        border: '1px solid #444',
        borderRadius: 8
      }
    };
  }

  /**
   * 导出配置到JSON
   */
  exportConfig(moduleName) {
    const config = this.configs[moduleName];
    if (!config) return null;
    return JSON.stringify(config, null, 2);
  }

  /**
   * 更新模块配置
   */
  updateConfig(moduleName, newConfig) {
    this.configs[moduleName] = newConfig;
    console.log(`更新 ${moduleName} 配置`);
  }

  /**
   * 合并自定义配置
   */
  mergeConfig(moduleName, customConfig) {
    const base = this.getConfig(moduleName);
    return this.deepMerge(base, customConfig);
  }

  /**
   * 深度合并对象
   */
  deepMerge(target, source) {
    if (!source) return target;
    const result = { ...target };
    
    for (const key in source) {
      if (source[key] && typeof source[key] === 'object' && !Array.isArray(source[key])) {
        result[key] = this.deepMerge(target[key] || {}, source[key]);
      } else {
        result[key] = source[key];
      }
    }
    
    return result;
  }
}

// 导出到全局
window.UIConfigManager = UIConfigManager;
