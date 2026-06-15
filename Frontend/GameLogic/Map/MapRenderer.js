/**
 * 地图渲染器：底层+高层瓦片绘制
 */
class MapRenderer {
  constructor(canvas, tileSetImage) {
    this.ctx = canvas.getContext('2d');
    this.tileSet = tileSetImage;
    this.tileSize = 32;
    this.tileCol = 16; // 瓦片图集一行16个
    
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

  renderMap(map) {
    const w = map.width;
    const h = map.height;
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
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
    
    // 如果有瓦片图集，使用图集
    if (this.tileSet) {
      const sx = (tileIndex % this.tileCol) * ts;
      const sy = Math.floor(tileIndex / this.tileCol) * ts;
      this.ctx.drawImage(
        this.tileSet,
        sx, sy, ts, ts,
        px, py, ts, ts
      );
    } else {
      // 没有瓦片图集，使用颜色
      const colorIndex = tileIndex % this.tileColors.length;
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