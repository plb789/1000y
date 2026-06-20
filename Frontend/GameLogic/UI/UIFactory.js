/**
 * UI组件工厂
 * 根据配置动态创建UI组件
 */
class UIFactory {
  constructor(game) {
    this.game = game;
    this.loader = null;
  }

  /**
   * 设置加载器
   */
  setLoader(loader) {
    this.loader = loader;
  }

  /**
   * 创建按钮
   */
  createButton(options = {}) {
    const {
      text = '',
      icon = '',
      className = '',
      style = {},
      onClick = null,
      parent = null
    } = options;

    const btn = document.createElement('button');
    btn.className = 'ui-btn ' + className;
    btn.innerHTML = icon ? `<span class="icon">${icon}</span>${text}` : text;

    // 应用基础样式
    btn.style.cssText = `
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      padding: 10px 20px;
      background: linear-gradient(135deg, #e94560, #c53030);
      border: none;
      border-radius: 6px;
      color: #fff;
      font-size: 14px;
      font-family: inherit;
      cursor: pointer;
      transition: all 0.2s ease;
    `;

    // 应用自定义样式
    Object.assign(btn.style, style);

    // 事件
    btn.onmouseenter = () => {
      btn.style.transform = 'translateY(-2px)';
      btn.style.boxShadow = '0 5px 15px rgba(233, 69, 96, 0.3)';
    };

    btn.onmouseleave = () => {
      btn.style.transform = '';
      btn.style.boxShadow = '';
    };

    if (onClick) btn.onclick = onClick;
    if (parent) parent.appendChild(btn);

    return btn;
  }

  /**
   * 创建输入框
   */
  createInput(options = {}) {
    const {
      placeholder = '',
      type = 'text',
      value = '',
      className = '',
      style = {},
      onChange = null,
      parent = null
    } = options;

    const input = document.createElement('input');
    input.type = type;
    input.placeholder = placeholder;
    input.value = value;
    input.className = 'ui-input ' + className;

    // 应用样式
    input.style.cssText = `
      padding: 8px 12px;
      background: rgba(45, 55, 72, 0.6);
      border: 1px solid #4a5568;
      border-radius: 6px;
      color: #fff;
      font-size: 14px;
      font-family: inherit;
      outline: none;
      transition: border-color 0.2s;
    `;

    Object.assign(input.style, style);

    input.onfocus = () => {
      input.style.borderColor = '#e94560';
    };

    input.onblur = () => {
      input.style.borderColor = '#4a5568';
    };

    if (onChange) input.onchange = onChange;
    if (parent) parent.appendChild(input);

    return input;
  }

  /**
   * 创建滑块
   */
  createSlider(options = {}) {
    const {
      min = 0,
      max = 100,
      value = 50,
      className = '',
      style = {},
      onChange = null,
      showValue = true,
      parent = null
    } = options;

    const container = document.createElement('div');
    container.className = 'ui-slider-container ' + className;
    container.style.cssText = `
      display: flex;
      align-items: center;
      gap: 10px;
    `;

    const slider = document.createElement('input');
    slider.type = 'range';
    slider.min = min;
    slider.max = max;
    slider.value = value;
    slider.style.cssText = `
      flex: 1;
      cursor: pointer;
      accent-color: #e94560;
    `;

    if (onChange) {
      slider.oninput = () => {
        if (showValue) valueDisplay.textContent = slider.value;
        onChange(parseInt(slider.value));
      };
    }

    const valueDisplay = document.createElement('span');
    valueDisplay.textContent = value;
    valueDisplay.style.cssText = `
      color: #e94560;
      min-width: 30px;
      text-align: right;
    `;

    container.appendChild(slider);
    if (showValue) container.appendChild(valueDisplay);
    if (parent) parent.appendChild(container);

    return { container, slider, valueDisplay };
  }

  /**
   * 创建复选框
   */
  createCheckbox(options = {}) {
    const {
      label = '',
      checked = false,
      className = '',
      style = {},
      onChange = null,
      parent = null
    } = options;

    const container = document.createElement('label');
    container.className = 'ui-checkbox ' + className;
    container.style.cssText = `
      display: inline-flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      font-size: 14px;
      color: #fff;
    `;

    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.checked = checked;
    checkbox.style.cssText = `
      width: 18px;
      height: 18px;
      cursor: pointer;
      accent-color: #e94560;
    `;

    const text = document.createElement('span');
    text.textContent = label;

    container.appendChild(checkbox);
    container.appendChild(text);

    if (onChange) {
      checkbox.onchange = () => onChange(checkbox.checked);
    }

    Object.assign(container.style, style);
    if (parent) parent.appendChild(container);

    return { container, checkbox };
  }

  /**
   * 创建面板
   */
  createPanel(options = {}) {
    const {
      title = '',
      className = '',
      style = {},
      showClose = true,
      onClose = null,
      parent = null
    } = options;

    const panel = document.createElement('div');
    panel.className = 'ui-panel ' + className;

    // 默认样式
    panel.style.cssText = `
      background: rgba(20, 20, 30, 0.95);
      border: 2px solid #e94560;
      border-radius: 15px;
      overflow: hidden;
    `;

    // 标题栏
    const header = document.createElement('div');
    header.style.cssText = `
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 15px 20px;
      border-bottom: 1px solid rgba(233, 69, 96, 0.3);
      background: rgba(0, 0, 0, 0.3);
    `;

    const titleEl = document.createElement('div');
    titleEl.textContent = title;
    titleEl.style.cssText = `
      color: #e94560;
      font-size: 16px;
      font-weight: bold;
    `;

    header.appendChild(titleEl);

    if (showClose) {
      const closeBtn = document.createElement('button');
      closeBtn.innerHTML = '×';
      closeBtn.style.cssText = `
        background: transparent;
        border: none;
        color: #999;
        font-size: 24px;
        cursor: pointer;
        padding: 0;
        line-height: 1;
      `;
      closeBtn.onclick = () => {
        if (onClose) onClose();
        panel.remove();
      };
      header.appendChild(closeBtn);
    }

    const content = document.createElement('div');
    content.className = 'ui-panel-content';
    content.style.cssText = `
      padding: 20px;
    `;

    panel.appendChild(header);
    panel.appendChild(content);

    // 应用自定义样式
    Object.assign(panel.style, style);

    if (parent) parent.appendChild(panel);

    return { panel, header, content };
  }

  /**
   * 创建标签页
   */
  createTabs(options = {}) {
    const {
      tabs = [],
      className = '',
      style = {},
      onChange = null,
      parent = null
    } = options;

    const container = document.createElement('div');
    container.className = 'ui-tabs ' + className;

    const tabBar = document.createElement('div');
    tabBar.style.cssText = `
      display: flex;
      padding: 10px 20px;
      gap: 10px;
      border-bottom: 1px solid rgba(233, 69, 96, 0.2);
    `;

    const contentContainer = document.createElement('div');
    contentContainer.style.cssText = `
      padding: 15px;
    `;

    const tabButtons = [];
    const tabContents = [];
    let activeIndex = 0;

    tabs.forEach((tab, index) => {
      const btn = document.createElement('button');
      btn.className = 'ui-tab-btn';
      btn.textContent = typeof tab === 'string' ? tab : tab.name;
      btn.style.cssText = `
        flex: 1;
        padding: 8px 15px;
        background: ${index === 0 ? 'rgba(233, 69, 96, 0.2)' : 'rgba(74, 85, 104, 0.3)'};
        border: 1px solid ${index === 0 ? '#e94560' : '#4a5568'};
        border-radius: 6px;
        color: ${index === 0 ? '#e94560' : '#999'};
        cursor: pointer;
        font-size: 14px;
        font-family: inherit;
        transition: all 0.2s;
      `;

      btn.onclick = () => {
        // 移除所有active状态
        tabButtons.forEach((b, i) => {
          b.classList.remove('active');
          b.style.background = 'rgba(74, 85, 104, 0.3)';
          b.style.borderColor = '#4a5568';
          b.style.color = '#999';
        });

        // 设置当前active
        btn.classList.add('active');
        btn.style.background = 'rgba(233, 69, 96, 0.2)';
        btn.style.borderColor = '#e94560';
        btn.style.color = '#e94560';

        // 隐藏所有内容
        tabContents.forEach(c => c.style.display = 'none');

        // 显示当前内容
        if (tabContents[index]) {
          tabContents[index].style.display = 'block';
        }

        if (onChange) onChange(index);
      };

      tabButtons.push(btn);
      tabBar.appendChild(btn);

      // 创建内容
      const content = document.createElement('div');
      content.className = 'ui-tab-content';
      content.style.display = index === 0 ? 'block' : 'none';
      contentContainer.appendChild(content);
      tabContents.push(content);
    });

    container.appendChild(tabBar);
    container.appendChild(contentContainer);
    Object.assign(container.style, style);

    if (parent) parent.appendChild(container);

    return { container, tabBar, contentContainer, tabButtons, tabContents };
  }

  /**
   * 创建物品槽
   */
  createItemSlot(options = {}) {
    const {
      size = 60,
      quality = 1,
      icon = '',
      count = 0,
      showBorder = true,
      className = '',
      style = {},
      onClick = null,
      onRightClick = null,
      parent = null
    } = options;

    const slot = document.createElement('div');
    slot.className = 'ui-item-slot ' + className;

    const qualityColors = {
      1: '#ffffff',
      2: '#4ade80',
      3: '#60a5fa',
      4: '#a855f7',
      5: '#fbbf24'
    };

    const borderColor = qualityColors[quality] || '#4a5568';

    slot.style.cssText = `
      width: ${size}px;
      height: ${size}px;
      background: rgba(45, 55, 72, 0.6);
      border: 2px solid ${showBorder ? borderColor : 'transparent'};
      border-radius: 8px;
      display: flex;
      justify-content: center;
      align-items: center;
      position: relative;
      cursor: pointer;
      transition: all 0.2s;
    `;

    // 图标
    const iconEl = document.createElement('div');
    iconEl.className = 'slot-icon';
    iconEl.textContent = icon;
    iconEl.style.cssText = 'font-size: 28px;';
    slot.appendChild(iconEl);

    // 数量
    if (count > 1) {
      const countEl = document.createElement('div');
      countEl.className = 'slot-count';
      countEl.textContent = count;
      countEl.style.cssText = `
        position: absolute;
        bottom: 2px;
        right: 4px;
        font-size: 10px;
        color: #fff;
        text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
      `;
      slot.appendChild(countEl);
    }

    // 悬停效果
    slot.onmouseenter = () => {
      slot.style.borderColor = '#e94560';
      slot.style.transform = 'scale(1.05)';
    };

    slot.onmouseleave = () => {
      slot.style.borderColor = showBorder ? borderColor : 'transparent';
      slot.style.transform = '';
    };

    if (onClick) slot.onclick = onClick;
    if (onRightClick) {
      slot.oncontextmenu = (e) => {
        e.preventDefault();
        onRightClick(e);
      };
    }

    Object.assign(slot.style, style);
    if (parent) parent.appendChild(slot);

    return slot;
  }

  /**
   * 创建提示框
   */
  createTooltip(options = {}) {
    const {
      title = '',
      content = '',
      x = 0,
      y = 0,
      className = '',
      parent = null
    } = options;

    const tooltip = document.createElement('div');
    tooltip.className = 'ui-tooltip ' + className;
    tooltip.innerHTML = `
      ${title ? `<div style="color: #e94560; font-weight: bold; margin-bottom: 8px;">${title}</div>` : ''}
      <div style="color: #ccc; font-size: 12px;">${content}</div>
    `;

    tooltip.style.cssText = `
      position: fixed;
      left: ${x + 15}px;
      top: ${y + 15}px;
      background: rgba(20, 20, 30, 0.95);
      border: 2px solid #e94560;
      border-radius: 10px;
      padding: 12px 15px;
      min-width: 180px;
      max-width: 250px;
      z-index: 1000;
      font-family: 'Microsoft YaHei', sans-serif;
      pointer-events: none;
      box-shadow: 0 0 20px rgba(233, 69, 96, 0.3);
    `;

    if (parent) parent.appendChild(tooltip);

    return tooltip;
  }

  /**
   * 创建进度条
   */
  createProgressBar(options = {}) {
    const {
      value = 0,
      max = 100,
      color = '#e94560',
      height = 8,
      showText = true,
      className = '',
      style = {},
      parent = null
    } = options;

    const container = document.createElement('div');
    container.className = 'ui-progress ' + className;

    const percent = Math.min(100, Math.max(0, (value / max) * 100));

    container.style.cssText = `
      width: 100%;
      height: ${height}px;
      background: rgba(45, 55, 72, 0.6);
      border-radius: ${height / 2}px;
      overflow: hidden;
      position: relative;
    `;

    const fill = document.createElement('div');
    fill.className = 'ui-progress-fill';
    fill.style.cssText = `
      width: ${percent}%;
      height: 100%;
      background: ${color};
      border-radius: ${height / 2}px;
      transition: width 0.3s ease;
    `;

    if (showText) {
      const text = document.createElement('div');
      text.className = 'ui-progress-text';
      text.textContent = `${value}/${max}`;
      text.style.cssText = `
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        font-size: 10px;
        color: #fff;
        text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
      `;
      container.appendChild(text);
    }

    container.appendChild(fill);
    Object.assign(container.style, style);

    if (parent) parent.appendChild(container);

    return {
      container,
      fill,
      setValue: (v) => {
        const p = Math.min(100, Math.max(0, (v / max) * 100));
        fill.style.width = p + '%';
        if (showText) {
          const textEl = container.querySelector('.ui-progress-text');
          if (textEl) textEl.textContent = `${v}/${max}`;
        }
      }
    };
  }

  /**
   * 创建遮罩层
   */
  createOverlay(options = {}) {
    const {
      onClick = null,
      parent = null
    } = options;

    const overlay = document.createElement('div');
    overlay.className = 'ui-overlay';

    overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.5);
      z-index: 899;
    `;

    if (onClick) overlay.onclick = onClick;
    if (parent) parent.appendChild(overlay);

    return overlay;
  }
}

// 导出到全局
window.UIFactory = UIFactory;
