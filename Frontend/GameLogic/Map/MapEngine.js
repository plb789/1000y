/**
 * 地图总引擎：镜头、角色、移动、寻路、事件
 */
class MapEngine {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.tileSize = 48; // 与服务端配置一致
    // 延迟初始化画布大小，等待容器显示后再调用

    // 地图核心
    this.mapParser = new MillenniumMapParser();
    this.mapRenderer = null;

    // 动画系统
    if (typeof MapAnimationSystem !== 'undefined' && typeof MapAnimationSystem === 'function') {
      this.animationSystem = new MapAnimationSystem();
      this.animationSystem.start();
    } else {
      console.warn('MapAnimationSystem is not loaded or not a constructor');
      this.animationSystem = null;
    }

    // 镜头
    this.camera = {
      offsetX: 0,
      offsetY: 0,
      dragStartX: 0,
      dragStartY: 0,
      mouseStartX: 0,
      mouseStartY: 0,
      isDrag: false
    };

    // 玩家数据
    this.player = {
      x: 10,
      y: 10,
      pixelX: 0,
      pixelY: 0,
      speed: 6, // 提高移动速度，使移动更流畅
      movePath: []
    };

    // 资源
    this.tilesetImg = null;
    this.roleAnim = null;
    
    // FPS 计算
    this.fps = 0;
    this.frameCount = 0;
    this.lastFpsTime = performance.now();
    this.onFpsUpdate = null; // FPS 更新回调

    this.bindEvent();
    this.loop();
  }

  resizeCanvas() {
    const container = this.canvas.parentElement;
    if (container) {
      this.canvas.width = container.clientWidth;
      this.canvas.height = container.clientHeight;
      
      // 如果地图已加载，重新调整相机位置以保持玩家在视野中心
      if (this.mapRenderer && this.mapParser) {
        this.followPlayer();
      }
    }
  }

  async loadMap(mapUrl, tilesetUrl, animationData = null) {
    const mapBuf = await CommonUtil.loadBinary(mapUrl);
    this.mapParser.loadMap(mapBuf);

    // 瓦片图集是可选的
    if (tilesetUrl) {
      this.tilesetImg = await CommonUtil.loadImage(tilesetUrl);
    }
    this.mapRenderer = new MapRenderer(this.canvas, this.tilesetImg, this.tileSize);
    
    // 设置动画系统
    if (this.animationSystem) {
      this.mapRenderer.setAnimationSystem(this.animationSystem);
      
      // 加载动画数据
      if (animationData && animationData.length > 0) {
        this.mapRenderer.loadAnimations(animationData);
      }
    }

    this.syncPlayerPixel();
    this.followPlayer();
  }

  tile2Pixel(tileX, tileY) {
    return {
      x: tileX * this.tileSize,
      y: tileY * this.tileSize
    };
  }

  pixel2Tile(px, py) {
    return {
      x: Math.floor(px / this.tileSize),
      y: Math.floor(py / this.tileSize)
    };
  }

  syncPlayerPixel() {
    const pos = this.tile2Pixel(this.player.x, this.player.y);
    this.player.pixelX = pos.x;
    this.player.pixelY = pos.y;
  }

  followPlayer() {
    const mapW = this.mapParser.width * this.tileSize;
    const mapH = this.mapParser.height * this.tileSize;
    const canvasW = this.canvas.width;
    const canvasH = this.canvas.height;

    this.camera.offsetX = this.player.pixelX - canvasW / 2 + this.tileSize / 2;
    this.camera.offsetY = this.player.pixelY - canvasH / 2 + this.tileSize / 2;

    this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
    this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
  }

  bindEvent() {
    // 镜头拖拽
    this.canvas.addEventListener('mousedown', (e) => {
      // 只响应右键拖拽
      if (e.button !== 2) return;
      this.camera.isDrag = true;
      this.camera.dragStartX = this.camera.offsetX; // 记录当前相机位置
      this.camera.dragStartY = this.camera.offsetY;
      this.camera.mouseStartX = e.clientX; // 记录鼠标起始位置
      this.camera.mouseStartY = e.clientY;
    });

    this.canvas.addEventListener('mousemove', (e) => {
      if (!this.camera.isDrag) return;
      const mapW = this.mapParser.width * this.tileSize;
      const mapH = this.mapParser.height * this.tileSize;
      const canvasW = this.canvas.width;
      const canvasH = this.canvas.height;

      // 计算鼠标移动距离，反向应用到相机位置
      const dx = this.camera.mouseStartX - e.clientX;
      const dy = this.camera.mouseStartY - e.clientY;
      
      this.camera.offsetX = this.camera.dragStartX + dx;
      this.camera.offsetY = this.camera.dragStartY + dy;

      this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
      this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
    });

    this.canvas.addEventListener('mouseup', (e) => {
      if (e.button === 2) {
        this.camera.isDrag = false;
      }
    });
    
    // 禁用右键菜单
    this.canvas.addEventListener('contextmenu', (e) => {
      e.preventDefault();
    });

    // 点击寻路
    this.canvas.addEventListener('click', (e) => {
      if (this.camera.isDrag) return;
      const rect = this.canvas.getBoundingClientRect();
      const clickPx = e.clientX - rect.left + this.camera.offsetX;
      const clickPy = e.clientY - rect.top + this.camera.offsetY;
      const targetTile = this.pixel2Tile(clickPx, clickPy);

      const tile = this.mapParser.getTile(targetTile.x, targetTile.y);
      if (!tile || tile.attr === 1) return;

      this.player.movePath = AStar.findPath(
        this.player.x, this.player.y,
        targetTile.x, targetTile.y,
        this.mapParser.collision,
        this.mapParser.width,
        this.mapParser.height
      );
      
      // 如果有路径，发送移动消息到服务器
      if (this.player.movePath.length > 0) {
        const target = this.player.movePath[this.player.movePath.length - 1];
        if (window.GameWS && window.GameWS.send) {
          window.GameWS.send(2001, { // CMD_MOVE
            x: target.x,
            y: target.y
          });
        }
      }
    });
  }

  updatePlayerMove() {
    const path = this.player.movePath;
    if (path.length === 0) return;

    const next = path[0];
    const targetPos = this.tile2Pixel(next.x, next.y);
    const dx = targetPos.x - this.player.pixelX;
    const dy = targetPos.y - this.player.pixelY;
    const dist = Math.hypot(dx, dy);

    if (dist < this.player.speed) {
      this.player.pixelX = targetPos.x;
      this.player.pixelY = targetPos.y;
      this.player.x = next.x;
      this.player.y = next.y;
      path.shift();
      
      // 移动到新格子时发送消息到服务器
      if (window.GameWS && window.GameWS.send) {
        window.GameWS.send(2001, { // CMD_MOVE
          role_id: this.player.id,
          x: this.player.x,
          y: this.player.y
        });
      }
      
      // 检测传送/事件区域
      this.checkEventArea();
      
      // 到达目标格子后更新相机
      this.followPlayer();
    } else {
      this.player.pixelX += (dx / dist) * this.player.speed;
      this.player.pixelY += (dy / dist) * this.player.speed;
      // 平滑移动时不每帧更新相机，使用插值让相机更平滑
      this.updateCameraSmooth();
    }
  }
  
  // 平滑更新相机位置，减少卡顿感
  updateCameraSmooth() {
    const mapW = this.mapParser.width * this.tileSize;
    const mapH = this.mapParser.height * this.tileSize;
    const canvasW = this.canvas.width;
    const canvasH = this.canvas.height;
    
    // 目标相机位置
    const targetOffsetX = this.player.pixelX - canvasW / 2 + this.tileSize / 2;
    const targetOffsetY = this.player.pixelY - canvasH / 2 + this.tileSize / 2;
    
    // 平滑插值
    this.camera.offsetX += (targetOffsetX - this.camera.offsetX) * 0.2;
    this.camera.offsetY += (targetOffsetY - this.camera.offsetY) * 0.2;
    
    // 限制边界
    this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
    this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
  }

  checkEventArea() {
    const tile = this.mapParser.getTile(this.player.x, this.player.y);
    if (!tile) return;
    if (tile.attr === 2) {
      console.log("进入传送区域");
    }
  }

  render() {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    if (!this.mapRenderer) return;

    this.ctx.save();
    this.ctx.translate(-this.camera.offsetX, -this.camera.offsetY);

    // 只渲染可见区域的瓦片
    this.mapRenderer.renderMap(
      this.mapParser,
      this.camera.offsetX,
      this.camera.offsetY,
      this.canvas.width,
      this.canvas.height
    );
    
    // 渲染动画效果
    if (this.animationSystem) {
      this.animationSystem.render(this.ctx, this.tileSize, this.tilesetImg);
    }
    
    // 绘制路径
    this.mapRenderer.drawPath(this.player.movePath);
    
    // 绘制玩家位置小红点（使用像素坐标实现平滑移动）
    this.mapRenderer.drawPlayerByPixel(this.player.pixelX, this.player.pixelY);

    // 绘制角色动画
    if (this.roleAnim) {
      this.roleAnim.draw(this.ctx, this.player.pixelX, this.player.pixelY);
    }
    
    // 渲染完成后调用回调（用于绘制其他玩家）
    // 注意：在 ctx.restore() 之前调用，确保摄像机偏移仍然有效
    if (this.afterRender) {
      this.afterRender();
    }

    this.ctx.restore();
  }

  loop() {
    this.updatePlayerMove();
    if (this.roleAnim) this.roleAnim.update();
    this.render();
    
    // FPS 计算：每秒更新一次
    this.frameCount++;
    const currentTime = performance.now();
    const elapsed = currentTime - this.lastFpsTime;
    if (elapsed >= 1000) {
      this.fps = Math.round((this.frameCount * 1000) / elapsed);
      this.frameCount = 0;
      this.lastFpsTime = currentTime;
      
      // 调用 FPS 更新回调
      if (this.onFpsUpdate) {
        this.onFpsUpdate(this.fps);
      }
    }
    
    // 玩家位置更新后调用回调（用于更新小地图）
    if (this.onPlayerMove) {
      this.onPlayerMove(this.player.x, this.player.y);
    }
    
    requestAnimationFrame(() => this.loop());
  }
}

window.MapEngine = MapEngine;