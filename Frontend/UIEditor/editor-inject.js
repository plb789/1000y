/**
 * UI编辑器注入脚本
 * 此脚本会被注入到客户端页面中，用于实现元素选择、拖拽、调整大小等编辑功能
 */

(function() {
  'use strict';
  
  // 编辑器状态
  const editorState = {
    enabled: false,
    selectedElement: null,
    hoveredElement: null,
    isDragging: false,
    isResizing: false,
    dragStart: { x: 0, y: 0 },
    elementStart: { x: 0, y: 0, width: 0, height: 0 },
    resizeHandle: null,
    changes: {},
    undoStack: [],
    redoStack: [],
    clipboard: null
  };
  
  // 高亮层
  let highlightOverlay = null;
  let selectionBox = null;
  let resizeHandles = [];
  
  // 初始化
  function init() {
    createOverlay();
    setupEventListeners();
    setupMessageListener();
    
    // 通知父窗口编辑器已加载
    window.parent.postMessage({ type: 'editor-ready' }, '*');
  }
  
  // 创建高亮层
  function createOverlay() {
    // 主高亮层
    highlightOverlay = document.createElement('div');
    highlightOverlay.id = 'editor-highlight-overlay';
    highlightOverlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      pointer-events: none;
      z-index: 999999;
    `;
    document.body.appendChild(highlightOverlay);
    
    // 选择框
    selectionBox = document.createElement('div');
    selectionBox.id = 'editor-selection-box';
    selectionBox.style.cssText = `
      position: absolute;
      border: 2px solid #e94560;
      background: rgba(233, 69, 96, 0.1);
      pointer-events: none;
      display: none;
    `;
    highlightOverlay.appendChild(selectionBox);
    
    // 选择标签
    const label = document.createElement('div');
    label.id = 'editor-selection-label';
    label.style.cssText = `
      position: absolute;
      top: -24px;
      left: 0;
      background: #e94560;
      color: #fff;
      padding: 2px 8px;
      font-size: 11px;
      font-family: 'Microsoft YaHei', sans-serif;
      border-radius: 3px;
      white-space: nowrap;
    `;
    selectionBox.appendChild(label);
    
    // 创建调整大小手柄
    const handles = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w'];
    handles.forEach(pos => {
      const handle = document.createElement('div');
      handle.className = 'editor-resize-handle';
      handle.dataset.position = pos;
      handle.style.cssText = `
        position: absolute;
        width: 10px;
        height: 10px;
        background: #e94560;
        border: 1px solid #fff;
        border-radius: 2px;
        pointer-events: auto;
        cursor: ${pos}-resize;
        display: none;
      `;
      
      // 设置位置
      switch (pos) {
        case 'nw': handle.style.cssText += 'top: -5px; left: -5px;'; break;
        case 'n': handle.style.cssText += 'top: -5px; left: 50%; transform: translateX(-50%);'; break;
        case 'ne': handle.style.cssText += 'top: -5px; right: -5px;'; break;
        case 'e': handle.style.cssText += 'top: 50%; right: -5px; transform: translateY(-50%);'; break;
        case 'se': handle.style.cssText += 'bottom: -5px; right: -5px;'; break;
        case 's': handle.style.cssText += 'bottom: -5px; left: 50%; transform: translateX(-50%);'; break;
        case 'sw': handle.style.cssText += 'bottom: -5px; left: -5px;'; break;
        case 'w': handle.style.cssText += 'top: 50%; left: -5px; transform: translateY(-50%);'; break;
      }
      
      selectionBox.appendChild(handle);
      resizeHandles.push(handle);
    });
  }
  
  // 设置事件监听
  function setupEventListeners() {
    // 鼠标移动 - 高亮元素
    document.addEventListener('mousemove', (e) => {
      if (!editorState.enabled) return;
      
      if (editorState.isDragging) {
        handleDrag(e);
      } else if (editorState.isResizing) {
        handleResize(e);
      } else {
        highlightElementAt(e.clientX, e.clientY);
      }
    });
    
    // 鼠标点击 - 选择元素
    document.addEventListener('click', (e) => {
      if (!editorState.enabled) return;
      e.preventDefault();
      e.stopPropagation();
      
      selectElementAt(e.clientX, e.clientY);
    }, true);
    
    // 鼠标按下 - 开始拖拽
    document.addEventListener('mousedown', (e) => {
      if (!editorState.enabled || !editorState.selectedElement) return;
      
      // 检查是否点击了调整手柄
      const handle = e.target.closest('.editor-resize-handle');
      if (handle) {
        e.preventDefault();
        startResize(e, handle.dataset.position);
        return;
      }
      
      // 检查是否点击了选中元素
      if (editorState.selectedElement.contains(e.target)) {
        e.preventDefault();
        startDrag(e);
      }
    }, true);
    
    // 鼠标释放 - 结束拖拽
    document.addEventListener('mouseup', (e) => {
      if (editorState.isDragging) {
        endDrag(e);
      } else if (editorState.isResizing) {
        endResize(e);
      }
    });
    
    // 键盘事件
    document.addEventListener('keydown', (e) => {
      if (!editorState.enabled) return;
      
      // Delete - 删除元素
      if (e.key === 'Delete' && editorState.selectedElement) {
        e.preventDefault();
        deleteSelectedElement();
      }
      
      // Ctrl+C - 复制
      if (e.ctrlKey && e.key === 'c' && editorState.selectedElement) {
        e.preventDefault();
        copyElement();
      }
      
      // Ctrl+V - 粘贴
      if (e.ctrlKey && e.key === 'v' && editorState.clipboard) {
        e.preventDefault();
        pasteElement();
      }
      
      // Ctrl+Z - 撤销
      if (e.ctrlKey && e.key === 'z') {
        e.preventDefault();
        undo();
      }
      
      // Ctrl+Y - 重做
      if (e.ctrlKey && e.key === 'y') {
        e.preventDefault();
        redo();
      }
      
      // Escape - 取消选择
      if (e.key === 'Escape') {
        clearSelection();
      }
    });
  }
  
  // 设置消息监听
  function setupMessageListener() {
    window.addEventListener('message', (e) => {
      const { type, data } = e.data;
      
      switch (type) {
        case 'enable-editor':
          editorState.enabled = true;
          highlightOverlay.style.display = 'block';
          break;
          
        case 'disable-editor':
          editorState.enabled = false;
          highlightOverlay.style.display = 'none';
          clearSelection();
          break;
          
        case 'select-element':
          const el = document.getElementById(data.id);
          if (el) selectElement(el);
          break;
          
        case 'update-style':
          if (editorState.selectedElement) {
            editorState.selectedElement.style[data.property] = data.value;
            updateSelectionBox();
            saveChange('style', data.property, data.value);
          }
          break;
          
        case 'update-attribute':
          if (editorState.selectedElement) {
            editorState.selectedElement.setAttribute(data.attribute, data.value);
            saveChange('attribute', data.attribute, data.value);
          }
          break;
          
        case 'add-element':
          addNewElement(data);
          break;
          
        case 'delete-element':
          deleteSelectedElement();
          break;
          
        case 'apply-image':
          if (editorState.selectedElement) {
            if (editorState.selectedElement.tagName === 'IMG') {
              editorState.selectedElement.src = data.imageData;
            } else {
              editorState.selectedElement.style.backgroundImage = `url(${data.imageData})`;
              editorState.selectedElement.style.backgroundSize = 'cover';
              editorState.selectedElement.style.backgroundPosition = 'center';
            }
            saveChange('style', 'background-image', data.imageData);
          }
          break;
          
        case 'get-html':
          window.parent.postMessage({
            type: 'html-content',
            data: {
              html: document.documentElement.outerHTML,
              scene: data.scene
            }
          }, '*');
          break;
          
        case 'switch-scene':
          switchScene(data.scene);
          break;
          
        case 'get-element-info':
          if (editorState.selectedElement) {
            sendElementInfo();
          }
          break;
      }
    });
  }
  
  // 高亮鼠标位置的元素
  function highlightElementAt(x, y) {
    // 移除之前的高亮
    const oldHover = document.querySelector('.editor-hover-highlight');
    if (oldHover) oldHover.remove();
    
    // 获取鼠标位置的元素
    const element = getElementAtPoint(x, y);
    if (!element || element === editorState.selectedElement) return;
    
    // 创建高亮
    const rect = element.getBoundingClientRect();
    const hover = document.createElement('div');
    hover.className = 'editor-hover-highlight';
    hover.style.cssText = `
      position: fixed;
      left: ${rect.left}px;
      top: ${rect.top}px;
      width: ${rect.width}px;
      height: ${rect.height}px;
      border: 1px dashed #60a5fa;
      background: rgba(96, 165, 250, 0.05);
      pointer-events: none;
      z-index: 999998;
    `;
    highlightOverlay.appendChild(hover);
    
    editorState.hoveredElement = element;
  }
  
  // 选择元素
  function selectElementAt(x, y) {
    const element = getElementAtPoint(x, y);
    if (element) {
      selectElement(element);
    }
  }
  
  // 选择指定元素
  function selectElement(element) {
    editorState.selectedElement = element;
    
    // 更新选择框
    updateSelectionBox();
    selectionBox.style.display = 'block';
    
    // 显示调整手柄
    resizeHandles.forEach(h => h.style.display = 'block');
    
    // 发送元素信息到父窗口
    sendElementInfo();
  }
  
  // 更新选择框位置
  function updateSelectionBox() {
    if (!editorState.selectedElement) return;
    
    const rect = editorState.selectedElement.getBoundingClientRect();
    selectionBox.style.left = rect.left + 'px';
    selectionBox.style.top = rect.top + 'px';
    selectionBox.style.width = rect.width + 'px';
    selectionBox.style.height = rect.height + 'px';
    
    // 更新标签
    const label = document.getElementById('editor-selection-label');
    label.textContent = editorState.selectedElement.id || 
                        editorState.selectedElement.className || 
                        editorState.selectedElement.tagName.toLowerCase();
  }
  
  // 获取鼠标位置的元素
  function getElementAtPoint(x, y) {
    // 临时隐藏选择框
    selectionBox.style.display = 'none';
    
    const element = document.elementFromPoint(x, y);
    
    // 恢复选择框
    if (editorState.selectedElement) {
      selectionBox.style.display = 'block';
    }
    
    // 排除编辑器元素
    if (element && (element.closest('#editor-highlight-overlay') || 
        element.id === 'editor-highlight-overlay')) {
      return null;
    }
    
    return element;
  }
  
  // 清除选择
  function clearSelection() {
    editorState.selectedElement = null;
    selectionBox.style.display = 'none';
    resizeHandles.forEach(h => h.style.display = 'none');
    
    window.parent.postMessage({ type: 'selection-cleared' }, '*');
  }
  
  // 发送元素信息到父窗口
  function sendElementInfo() {
    if (!editorState.selectedElement) return;
    
    const element = editorState.selectedElement;
    const style = window.getComputedStyle(element);
    
    window.parent.postMessage({
      type: 'element-selected',
      data: {
        id: element.id,
        className: element.className,
        tagName: element.tagName.toLowerCase(),
        textContent: element.textContent?.substring(0, 100),
        style: {
          backgroundColor: style.backgroundColor,
          color: style.color,
          fontSize: style.fontSize,
          fontFamily: style.fontFamily,
          width: style.width,
          height: style.height,
          padding: style.padding,
          margin: style.margin,
          border: style.border,
          borderRadius: style.borderRadius,
          position: style.position,
          left: style.left,
          top: style.top,
          right: style.right,
          bottom: style.bottom,
          display: style.display,
          opacity: style.opacity,
          zIndex: style.zIndex,
          backgroundImage: style.backgroundImage,
          boxShadow: style.boxShadow,
          textAlign: style.textAlign
        },
        attributes: Array.from(element.attributes).map(attr => ({
          name: attr.name,
          value: attr.value
        }))
      }
    }, '*');
  }
  
  // 开始拖拽
  function startDrag(e) {
    editorState.isDragging = true;
    editorState.dragStart = { x: e.clientX, y: e.clientY };
    
    const rect = editorState.selectedElement.getBoundingClientRect();
    editorState.elementStart = {
      x: rect.left,
      y: rect.top,
      width: rect.width,
      height: rect.height
    };
    
    // 保存初始状态用于撤销
    saveUndoState();
  }
  
  // 处理拖拽
  function handleDrag(e) {
    if (!editorState.isDragging || !editorState.selectedElement) return;
    
    const deltaX = e.clientX - editorState.dragStart.x;
    const deltaY = e.clientY - editorState.dragStart.y;
    
    const newX = editorState.elementStart.x + deltaX;
    const newY = editorState.elementStart.y + deltaY;
    
    // 设置位置
    editorState.selectedElement.style.position = 'absolute';
    editorState.selectedElement.style.left = newX + 'px';
    editorState.selectedElement.style.top = newY + 'px';
    
    updateSelectionBox();
  }
  
  // 结束拖拽
  function endDrag(e) {
    editorState.isDragging = false;
    
    // 保存修改
    saveChange('position', 'left', editorState.selectedElement.style.left);
    saveChange('position', 'top', editorState.selectedElement.style.top);
    
    // 通知父窗口
    sendElementInfo();
  }
  
  // 开始调整大小
  function startResize(e, position) {
    editorState.isResizing = true;
    editorState.resizeHandle = position;
    editorState.dragStart = { x: e.clientX, y: e.clientY };
    
    const rect = editorState.selectedElement.getBoundingClientRect();
    editorState.elementStart = {
      x: rect.left,
      y: rect.top,
      width: rect.width,
      height: rect.height
    };
    
    saveUndoState();
  }
  
  // 处理调整大小
  function handleResize(e) {
    if (!editorState.isResizing || !editorState.selectedElement) return;
    
    const deltaX = e.clientX - editorState.dragStart.x;
    const deltaY = e.clientY - editorState.dragStart.y;
    
    let newWidth = editorState.elementStart.width;
    let newHeight = editorState.elementStart.height;
    let newLeft = editorState.elementStart.x;
    let newTop = editorState.elementStart.y;
    
    const handle = editorState.resizeHandle;
    
    // 根据手柄位置调整大小
    if (handle.includes('e')) newWidth = editorState.elementStart.width + deltaX;
    if (handle.includes('w')) {
      newWidth = editorState.elementStart.width - deltaX;
      newLeft = editorState.elementStart.x + deltaX;
    }
    if (handle.includes('s')) newHeight = editorState.elementStart.height + deltaY;
    if (handle.includes('n')) {
      newHeight = editorState.elementStart.height - deltaY;
      newTop = editorState.elementStart.y + deltaY;
    }
    
    // 最小尺寸限制
    newWidth = Math.max(20, newWidth);
    newHeight = Math.max(20, newHeight);
    
    // 应用新尺寸
    editorState.selectedElement.style.width = newWidth + 'px';
    editorState.selectedElement.style.height = newHeight + 'px';
    
    if (handle.includes('w') || handle.includes('n')) {
      editorState.selectedElement.style.left = newLeft + 'px';
      editorState.selectedElement.style.top = newTop + 'px';
    }
    
    updateSelectionBox();
  }
  
  // 结束调整大小
  function endResize(e) {
    editorState.isResizing = false;
    editorState.resizeHandle = null;
    
    // 保存修改
    saveChange('size', 'width', editorState.selectedElement.style.width);
    saveChange('size', 'height', editorState.selectedElement.style.height);
    
    sendElementInfo();
  }
  
  // 添加新元素
  function addNewElement(data) {
    const parent = document.getElementById(data.parentId) || document.body;
    const element = document.createElement(data.tagName);
    
    // 设置属性
    if (data.id) element.id = data.id;
    if (data.className) element.className = data.className;
    if (data.textContent) element.textContent = data.textContent;
    if (data.innerHTML) element.innerHTML = data.innerHTML;
    
    // 设置样式
    if (data.style) {
      Object.assign(element.style, data.style);
    }
    
    parent.appendChild(element);
    
    // 选择新元素
    selectElement(element);
    
    // 发送更新通知
    window.parent.postMessage({ type: 'element-added', data: { id: element.id } }, '*');
    
    saveUndoState();
  }
  
  // 删除选中元素
  function deleteSelectedElement() {
    if (!editorState.selectedElement) return;
    
    // 不允许删除关键元素
    const criticalIds = ['loginPanel', 'roleSelectPanel', 'roleCreatePanel', 'gamePanel', 'gameCanvas', 'topBar', 'bottomBar'];
    if (criticalIds.includes(editorState.selectedElement.id)) {
      window.parent.postMessage({ type: 'error', message: '不能删除关键组件' }, '*');
      return;
    }
    
    saveUndoState();
    
    const parent = editorState.selectedElement.parentElement;
    editorState.selectedElement.remove();
    
    clearSelection();
    
    window.parent.postMessage({ type: 'element-deleted' }, '*');
  }
  
  // 复制元素
  function copyElement() {
    if (!editorState.selectedElement) return;
    
    editorState.clipboard = {
      tagName: editorState.selectedElement.tagName,
      className: editorState.selectedElement.className,
      style: editorState.selectedElement.style.cssText,
      innerHTML: editorState.selectedElement.innerHTML,
      textContent: editorState.selectedElement.textContent
    };
    
    window.parent.postMessage({ type: 'element-copied' }, '*');
  }
  
  // 粘贴元素
  function pasteElement() {
    if (!editorState.clipboard) return;
    
    const element = document.createElement(editorState.clipboard.tagName);
    element.className = editorState.clipboard.className + ' copy-' + Date.now();
    element.style.cssText = editorState.clipboard.style;
    element.innerHTML = editorState.clipboard.innerHTML;
    
    // 偏移位置
    element.style.position = 'absolute';
    element.style.left = (parseInt(element.style.left) || 100) + 20 + 'px';
    element.style.top = (parseInt(element.style.top) || 100) + 20 + 'px';
    
    document.body.appendChild(element);
    selectElement(element);
    
    saveUndoState();
    window.parent.postMessage({ type: 'element-pasted' }, '*');
  }
  
  // 保存修改
  function saveChange(category, property, value) {
    const id = editorState.selectedElement?.id || 'unknown';
    if (!editorState.changes[id]) {
      editorState.changes[id] = {};
    }
    if (!editorState.changes[id][category]) {
      editorState.changes[id][category] = {};
    }
    editorState.changes[id][category][property] = value;
  }
  
  // 保存撤销状态
  function saveUndoState() {
    const html = document.documentElement.outerHTML;
    editorState.undoStack.push(html);
    if (editorState.undoStack.length > 50) {
      editorState.undoStack.shift();
    }
    editorState.redoStack = [];
  }
  
  // 撤销
  function undo() {
    if (editorState.undoStack.length === 0) return;
    
    const currentHtml = document.documentElement.outerHTML;
    editorState.redoStack.push(currentHtml);
    
    const previousHtml = editorState.undoStack.pop();
    document.documentElement.outerHTML = previousHtml;
    
    window.parent.postMessage({ type: 'undo-done' }, '*');
  }
  
  // 重做
  function redo() {
    if (editorState.redoStack.length === 0) return;
    
    const currentHtml = document.documentElement.outerHTML;
    editorState.undoStack.push(currentHtml);
    
    const nextHtml = editorState.redoStack.pop();
    document.documentElement.outerHTML = nextHtml;
    
    window.parent.postMessage({ type: 'redo-done' }, '*');
  }
  
  // 切换场景
  function switchScene(scene) {
    // 隐藏所有面板
    const panels = ['loginPanel', 'roleSelectPanel', 'roleCreatePanel', 'gamePanel', 'loadingOverlay', 'loadingScreen'];
    panels.forEach(id => {
      const el = document.getElementById(id);
      if (el) {
        if (id === 'gamePanel') {
          el.classList.remove('active');
        } else {
          el.style.display = 'none';
        }
      }
    });
    
    // 显示对应场景
    switch (scene) {
      case 'login':
        const loginPanel = document.getElementById('loginPanel');
        if (loginPanel) loginPanel.style.display = 'flex';
        break;
      case 'roleSelect':
        const roleSelectPanel = document.getElementById('roleSelectPanel');
        if (roleSelectPanel) roleSelectPanel.style.display = 'block';
        break;
      case 'roleCreate':
        const roleCreatePanel = document.getElementById('roleCreatePanel');
        if (roleCreatePanel) roleCreatePanel.style.display = 'block';
        break;
      case 'game':
        const gamePanel = document.getElementById('gamePanel');
        if (gamePanel) {
          gamePanel.classList.add('active');
          gamePanel.style.display = 'block';
        }
        break;
    }
    
    clearSelection();
    window.parent.postMessage({ type: 'scene-switched', scene }, '*');
  }
  
  // 页面加载完成后初始化
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();