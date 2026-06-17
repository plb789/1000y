/**
 * 地图渲染器：底层+高层瓦片绘制
 */
class MapRenderer {
  /**
   * @param {HTMLCanvasElement} canvas - 渲染画布
   * @param {HTMLImageElement} tileSetImage - 瓦片图集图片
   * @param {number} tileSize - 瓦片像素尺寸
   * @param {number} [tileColOverride] - 外部指定的图集列数（优先于图片宽度计算）
   */
  constructor(canvas, tileSetImage, tileSize = 48, tileColOverride = 0) {
    this.ctx = canvas.getContext('2d');
    this.tileSet = tileSetImage || null; // 确保正确处理 null
    this.tileSize = tileSize;
    this.debugMode = false; // 调试模式：显示瓦片ID

    // 图集列数优先级：外部指定 > 图片宽度计算 > 默认16
    if (tileColOverride > 0) {
      this.tileCol = tileColOverride;
    } else if (tileSetImage && tileSetImage.width > 0) {
      this.tileCol = Math.floor(tileSetImage.width / tileSize);
    } else {
      this.tileCol = 16;
    }

    // 验证日志：输出渲染器的关键参数
    console.log(`🎨 MapRenderer 初始化:`);
    console.log(`   最终tileCol=${this.tileCol} (来源: ${tileColOverride > 0 ? '.map文件头' : (tileSetImage ? '图片宽度计算' : '默认')})`);
    if (tileSetImage) {
      console.log(`   图集尺寸: ${tileSetImage.width}x${tileSetImage.height}px`);
      console.log(`   图集实际列数(宽/瓦片): ${Math.floor(tileSetImage.width / tileSize)}, 行数: ${Math.floor(tileSetImage.height / tileSize)}`);
      // 验证前5个瓦片在图集中的位置
      for (let i = 1; i <= 5; i++) {
        const idx = i - 1;
        const sx = (idx % this.tileCol) * tileSize;
        const sy = Math.floor(idx / this.tileCol) * tileSize;
        console.log(`   瓦片${i} → 图集位置(${sx},${sy}) 第${Math.floor(idx / this.tileCol)+1}行第${(idx % this.tileCol)+1}列`);
      }
      // 关键验证：检查图集最后一行的几个瓦片是否越界
      const lastTileIdx = Math.floor(tileSetImage.width / tileSize) * Math.floor(tileSetImage.height / tileSize) - 1;
      if (lastTileIdx > 0) {
        const lsx = (lastTileIdx % this.tileCol) * tileSize;
        const lsy = Math.floor(lastTileIdx / this.tileCol) * tileSize;
        const inBounds = (lsx + tileSize <= tileSetImage.width) && (lsy + tileSize <= tileSetImage.height);
        console.log(`   最后瓦片#${lastTileIdx+1} → 图集位置(${lsx},${lsy}) 越界=${!inBounds}`);
      }
    }
    
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

      // 调试模式：检测越界并标记
      if (this.debugMode) {
        const outOfBounds = (sx + ts > this.tileSet.width) || (sy + ts > this.tileSet.height);
        if (outOfBounds) {
          // 越界：用红色填充 + 显示警告
          this.ctx.fillStyle = '#ff0000';
          this.ctx.fillRect(px, py, ts, ts);
          this.ctx.fillStyle = '#ffffff';
          this.ctx.font = 'bold 10px monospace';
          this.ctx.textAlign = 'center';
          this.ctx.fillText('!' + tileIndex, px + ts / 2, py + ts / 2);
          return; // 不再绘制图集内容
        }
      }

      this.ctx.drawImage(
        this.tileSet,
        sx, sy, ts, ts,
        px, py, ts, ts
      );

      // 调试模式：显示瓦片ID和坐标信息
      if (this.debugMode) {
        this.ctx.fillStyle = 'rgba(0,0,0,0.6)';
        this.ctx.fillRect(px, py + ts - 14, ts, 14);
        this.ctx.fillStyle = '#00ff00';
        this.ctx.font = 'bold 9px monospace';
        this.ctx.textAlign = 'left';
        this.ctx.fillText(`${tileIndex}(${sx},${sy})`, px + 1, py + ts - 3);
      }
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
  
  // 绘制玩家位置小红点（使用瓦片坐标）
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
  
  // 绘制玩家位置小红点（使用像素坐标，实现平滑移动）
  drawPlayerByPixel(px, py) {
    const ts = this.tileSize;
    
    this.ctx.fillStyle = '#e94560';
    this.ctx.beginPath();
    this.ctx.arc(px + ts / 2, py + ts / 2, ts / 3, 0, Math.PI * 2);
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