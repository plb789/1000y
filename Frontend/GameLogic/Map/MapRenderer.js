/**
 * 地图渲染器：底层+高层瓦片绘制
 */
class MapRenderer {
  constructor(canvas, tileSetImage, tileSize = 48) {
    this.ctx = canvas.getContext('2d');
    this.tileSet = tileSetImage;
    this.tileSize = tileSize;
    this.tileCol = 16; // 瓦片图集一行16个
    
    // 动画系统
    this.animationSystem = null;
    this.mapAnimations = new Map();
    
    // 瓦片颜色表（当没有瓦片图集时使用）
    this.tileColors = [
      '#3a5f0b', // 0: 草地
      '#1e4d8c', // 1: 水域
      '#5c5c5c', // 2: 墙壁
      '#8b7355', // 3: 地板
      '#228b22', // 4: 树木
      '#c2b280', // 5: 沙地
      '#8b4513', // 6: 泥土
      '#696969', // 7: 石板路
      '#a0522d', // 8: 楼梯
      '#654321', // 9: 门
      '#8b7765', // 10: 桥梁
      '#ff69b4', // 11: 花丛
      '#4a5568', // 12: 石头
      '#2d3748', // 13: 深墙
      '#1a202c', // 14: 洞穴
      '#e94560', // 15: 终点
    ];
  }
  
  /**
   * 设置动画系统
   */
  setAnimationSystem(animationSystem) {
    this.animationSystem = animationSystem;
  }
  
  /**
   * 加载地图动画数据
   */
  loadAnimations(animations) {
    if (!this.animationSystem) return;
    
    animations.forEach(anim => {
      const key = `${anim.x},${anim.y}`;
      const animId = `anim_${key}`;
      this.animationSystem.addAnimation(animId, anim.type, anim.x * this.tileSize, anim.y * this.tileSize, { speed: anim.speed });
      this.mapAnimations.set(key, anim);
    });
  }
  
  /**
   * 清除所有动画
   */
  clearAnimations() {
    if (this.animationSystem) {
      this.animationSystem.clearAll();
    }
    this.mapAnimations.clear();
  }

  renderMap(map, cameraX, cameraY, canvasWidth, canvasHeight) {
    const w = map.width;
    const h = map.height;
    const ts = this.tileSize;
    
    // 计算可见瓦片范围
    const startX = Math.max(0, Math.floor(cameraX / ts) - 1);
    const startY = Math.max(0, Math.floor(cameraY / ts) - 1);
    const endX = Math.min(w, Math.ceil((cameraX + canvasWidth) / ts) + 1);
    const endY = Math.min(h, Math.ceil((cameraY + canvasHeight) / ts) + 1);
    
    // 只渲染可见区域
    for (let y = startY; y < endY; y++) {
      for (let x = startX; x < endX; x++) {
        const tile = map.getTile(x, y);
        if (!tile) continue;
        this.drawTile(x, y, tile.low, tile.attr);
        if (tile.high > 0) {
          this.drawTile(x, y, tile.high, tile.attr);
        }
      }
    }
  }

  drawTile(x, y, tileIndex, attr) {
    const ts = this.tileSize;
    const px = x * ts;
    const py = y * ts;
    
    // 瓦片索引从1开始，图集索引从0开始，需要减1
    const atlasIndex = tileIndex - 1;
    
    // tileIndex = 0 表示空瓦片，跳过绘制
    if (tileIndex === 0) {
      return;
    }
    
    // 如果有瓦片图集，使用图集
    if (this.tileSet && atlasIndex >= 0) {
      const sx = (atlasIndex % this.tileCol) * ts;
      const sy = Math.floor(atlasIndex / this.tileCol) * ts;
      this.ctx.drawImage(
        this.tileSet,
        sx, sy, ts, ts,
        px, py, ts, ts
      );
    } else {
      // 没有瓦片图集，使用颜色
      const colorIndex = atlasIndex >= 0 ? atlasIndex % this.tileColors.length : 0;
      this.ctx.fillStyle = this.tileColors[colorIndex] || '#ff00ff';
      this.ctx.fillRect(px, py, ts, ts);
      
      // 如果是阻挡，显示X标记
      if (attr === 1) {
        this.ctx.strokeStyle = '#ff0000';
        this.ctx.lineWidth = 2;
        this.ctx.beginPath();
        this.ctx.moveTo(px + 4, py + 4);
        this.ctx.lineTo(px + ts - 4, py + ts - 4);
        this.ctx.moveTo(px + ts - 4, py + 4);
        this.ctx.lineTo(px + 4, py + ts - 4);
        this.ctx.stroke();
      }
      
      // 绘制边框
      this.ctx.strokeStyle = 'rgba(0,0,0,0.3)';
      this.ctx.lineWidth = 1;
      this.ctx.strokeRect(px, py, ts, ts);
    }
  }
  
  // 绘制玩家位置小红点
  drawPlayer(x, y) {
    const ts = this.tileSize;
    const px = x * ts + ts / 2;
    const py = y * ts + ts / 2;
    
    this.ctx.fillStyle = '#e94560';
    this.ctx.beginPath();
    this.ctx.arc(px, py, ts / 3, 0, Math.PI * 2);
    this.ctx.fill();
    
    // 白色边框
    this.ctx.strokeStyle = '#fff';
    this.ctx.lineWidth = 2;
    this.ctx.stroke();
  }
  
  // 绘制路径点
  drawPath(path) {
    if (!path || path.length === 0) return;
    const ts = this.tileSize;
    
    this.ctx.fillStyle = 'rgba(233, 69, 96, 0.5)';
    path.forEach((p, i) => {
      const px = p.x * ts + ts / 2;
      const py = p.y * ts + ts / 2;
      this.ctx.beginPath();
      this.ctx.arc(px, py, 4, 0, Math.PI * 2);
      this.ctx.fill();
    });
  }
}

window.MapRenderer = MapRenderer;