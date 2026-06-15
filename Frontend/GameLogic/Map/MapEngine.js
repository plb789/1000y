/**
 * 地图总引擎：镜头、角色、移动、寻路、事件
 */
class MapEngine {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.canvas.width = 1000;
    this.canvas.height = 600;
    this.tileSize = 32;

    // 地图核心
    this.mapParser = new MillenniumMapParser();
    this.mapRenderer = null;

    // 镜头
    this.camera = {
      offsetX: 0,
      offsetY: 0,
      dragStartX: 0,
      dragStartY: 0,
      isDrag: false
    };

    // 玩家数据
    this.player = {
      x: 10,
      y: 10,
      pixelX: 0,
      pixelY: 0,
      speed: 2,
      movePath: []
    };

    // 资源
    this.tilesetImg = null;
    this.roleAnim = null;

    this.bindEvent();
    this.loop();
  }

  async loadMap(mapUrl, tilesetUrl) {
    const mapBuf = await CommonUtil.loadBinary(mapUrl);
    this.mapParser.loadMap(mapBuf);

    // 瓦片图集是可选的
    if (tilesetUrl) {
      this.tilesetImg = await CommonUtil.loadImage(tilesetUrl);
    }
    this.mapRenderer = new MapRenderer(this.canvas, this.tilesetImg);

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
      this.camera.isDrag = true;
      this.camera.dragStartX = e.clientX - this.camera.offsetX;
      this.camera.dragStartY = e.clientY - this.camera.offsetY;
    });

    this.canvas.addEventListener('mousemove', (e) => {
      if (!this.camera.isDrag) return;
      const mapW = this.mapParser.width * this.tileSize;
      const mapH = this.mapParser.height * this.tileSize;
      const canvasW = this.canvas.width;
      const canvasH = this.canvas.height;

      this.camera.offsetX = e.clientX - this.camera.dragStartX;
      this.camera.offsetY = e.clientY - this.camera.dragStartY;

      this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
      this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
    });

    this.canvas.addEventListener('mouseup', () => {
      this.camera.isDrag = false;
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
          window.GameWS.send(3002, { // CMD_MOVE
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
          x: this.player.x,
          y: this.player.y
        });
      }
      
      // 检测传送/事件区域
      this.checkEventArea();
    } else {
      this.player.pixelX += (dx / dist) * this.player.speed;
      this.player.pixelY += (dy / dist) * this.player.speed;
    }
    this.followPlayer();
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

    this.mapRenderer.renderMap(this.mapParser);
    
    // 绘制路径
    this.mapRenderer.drawPath(this.player.movePath);
    
    // 绘制玩家位置小红点
    this.mapRenderer.drawPlayer(this.player.x, this.player.y);

    // 绘制角色动画
    if (this.roleAnim) {
      this.roleAnim.draw(this.ctx, this.player.pixelX, this.player.pixelY);
    }

    this.ctx.restore();
    
    // 渲染完成后调用回调（用于绘制其他玩家）
    if (this.afterRender) {
      this.afterRender();
    }
  }

  loop() {
    this.updatePlayerMove();
    if (this.roleAnim) this.roleAnim.update();
    this.render();
    
    // 玩家位置更新后调用回调（用于更新小地图）
    if (this.onPlayerMove) {
      this.onPlayerMove(this.player.x, this.player.y);
    }
    
    requestAnimationFrame(() => this.loop());
  }
}

window.MapEngine = MapEngine;