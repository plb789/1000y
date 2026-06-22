/**
 * 地图渲染器：支持三层地图系统的Z轴深度排序渲染
 * Layer 0: 地面层 | Layer 1: 物体层 | Layer 2: 覆盖层（桥梁/屋顶）
 * 使用Y-Sort算法确保正确的遮挡关系（如角色在桥下被遮挡）
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
    this.tileSet = tileSetImage || null;
    this.tileSize = tileSize;
    this.debugMode = false;

    // 图集列数优先级：外部指定(.map文件头) > 图片宽度计算 > 默认16
    if (tileColOverride > 0) {
      this.tileCol = tileColOverride;
    } else if (tileSetImage && tileSetImage.width > 0) {
      this.tileCol = Math.floor(tileSetImage.width / tileSize);
    } else {
      this.tileCol = 16;
    }

    if (tileColOverride > 0 && tileSetImage && tileSetImage.width > 0) {
      const calcCols = Math.floor(tileSetImage.width / tileSize);
      if (tileColOverride !== calcCols) {
        console.warn(`⚠️ 列数不一致！文件头tilesetCols=${tileColOverride} vs 图片宽度计算=${calcCols}`);
      }
    }

    console.log(`🎨 MapRenderer 初始化:`);
    console.log(`   最终tileCol=${this.tileCol} (来源: ${tileColOverride > 0 ? '.map文件头' : (tileSetImage ? '图片宽度计算' : '默认')})`);
    if (tileSetImage) {
      console.log(`   图集尺寸: ${tileSetImage.width}x${tileSetImage.height}px`);
    }

    // 动画系统
    this.animationSystem = null;
    this.mapAnimations = new Map();

    // ★ 新增：分层渲染配置
    this.layerRendering = {
      ground: true,     // 渲染地面层
      object: true,     // 渲染物体层
      overlay: true,    // 渲染覆盖层（桥梁等）
      player: true      // 渲染玩家
    };

    // 瓦片颜色表（当没有瓦片图集时使用）
    this.tileColors = [
      '#3a5f0b', '#1e4d8c', '#5c5c5c', '#8b7355',
      '#228b22', '#c2b280', '#8b4513', '#696969',
      '#a0522d', '#654321', '#8b7765', '#ff69b4',
      '#4a5568', '#2d3748', '#1a202c', '#e94560'
    ];
  }

  setAnimationSystem(animationSystem) {
    this.animationSystem = animationSystem;
  }

  loadAnimations(animations) {
    if (!this.animationSystem) return;
    animations.forEach(anim => {
      const key = `${anim.x},${anim.y}`;
      const animId = `anim_${key}`;
      this.animationSystem.addAnimation(animId, anim.type, anim.x * this.tileSize, anim.y * this.tileSize, { speed: anim.speed });
      this.mapAnimations.set(key, anim);
    });
  }

  clearAnimations() {
    if (this.animationSystem) {
      this.animationSystem.clearAll();
    }
    this.mapAnimations.clear();
  }

  /**
   * ★ 核心方法：使用Z轴深度排序的地图渲染
   * 实现Y-Sort算法，正确处理遮挡关系（角色在桥下被遮挡效果）
   *
   * @param {Object} map - MillenniumMapParser实例（包含layerData和objects）
   * @param {number} cameraX - 相机X偏移
   * @param {number} cameraY - 相机Y偏移
   * @param {number} canvasWidth - 画布宽度
   * @param {number} canvasHeight - 画布高度
   * @param {Object} [playerPos] - 可选的玩家位置 {x, y}（瓦片坐标），用于Z排序
   */
  renderMap(map, cameraX, cameraY, canvasWidth, canvasHeight, playerPos = null) {
    const w = map.width;
    const h = map.height;
    const ts = this.tileSize;

    // 计算可见瓦片范围（优化性能）
    const startX = Math.max(0, Math.floor(cameraX / ts) - 1);
    const startY = Math.max(0, Math.floor(cameraY / ts) - 1);
    const endX = Math.min(w, Math.ceil((cameraX + canvasWidth) / ts) + 1);
    const endY = Math.min(h, Math.ceil((cameraY + canvasHeight) / ts) + 1);

    // ★ 检查是否支持新的分层系统（v2.0+）
    if (map.layerData && map.getZSortedRenderList) {
      // 使用新的Z轴深度排序渲染
      this.renderWithZSort(map, startX, startY, endX, endY, playerPos);
    } else {
      // 兼容旧版格式：简单的双层渲染
      this.renderLegacy(map, startX, startY, endX, endY);
    }
  }

  /**
   * ★ 新版渲染器：Z轴深度排序渲染（支持三层地图+物体对象）
   * 渲染顺序：地面层 → 物体层 → [按Y排序的角色] → 覆盖层（桥梁）
   *
   * 性能优化：将视口范围传递给 getZSortedRenderList，从源头只收集视口内瓦片
   *          避免遍历整张地图（3000×3000地图从2700万次遍历降到约2000次）
   */
  renderWithZSort(map, startX, startY, endX, endY, playerPos) {
    const ts = this.tileSize;

    // 获取玩家位置用于Z排序
    const px = playerPos ? playerPos.x : (map.playerX || 0);
    const py = playerPos ? playerPos.y : (map.playerY || 0);

    // ★ 传入视口范围，启用视口裁剪
    //    这样 getZSortedRenderList 只收集视口内瓦片，大幅降低遍历和排序开销
    const viewport = { startX, startY, endX, endY };
    const renderList = map.getZSortedRenderList(px, py, viewport);

    // 按排序顺序逐个渲染
    renderList.forEach(item => {
      // 视野裁剪：只渲染可见区域（双重保险，防止 viewport 边界误差）
      if (item.x < startX || item.x >= endX || item.y < startY || item.y >= endY) {
        return;
      }

      switch (item.type) {
        case 'tile':           // 地面层瓦片
          if (this.layerRendering.ground) {
            this.drawTile(item.x, item.y, item.tileId, 0);
          }
          break;

        case 'tile_object':    // 物体层瓦片
          if (this.layerRendering.object) {
            this.drawTile(item.x, item.y, item.tileId, 0);
          }
          break;

        case 'object':         // 独立物体对象
          if (this.layerRendering.object) {
            this.drawTile(item.x, item.y, item.tileId, 0);
            // 如果有动画效果，绘制动画覆盖
            if (item.animType && this.animationSystem) {
              this.renderObjectAnimation(item);
            }
          }
          break;

        case 'player':         // 玩家角色（由外部调用drawPlayer处理）
          // 不在这里绘制玩家，仅占位保证Z顺序正确
          // 实际玩家绘制在MapEngine.render()中单独调用
          break;

        case 'tile_overlay':   // 覆盖层（桥梁、屋顶）- 关键！在角色之后渲染
          if (this.layerRendering.overlay) {
            this.drawTile(item.x, item.y, item.tileId, 0);

            // 调试模式：显示覆盖层标记
            if (this.debugMode && item.blocksView) {
              this.ctx.strokeStyle = 'rgba(255, 215, 0, 0.7)';
              this.ctx.lineWidth = 2;
              this.ctx.strokeRect(
                item.x * ts + 2,
                item.y * ts + 2,
                ts - 4,
                ts - 4
              );
              this.ctx.fillStyle = 'rgba(255, 215, 0, 0.3)';
              this.ctx.font = 'bold 10px monospace';
              this.ctx.fillText('桥', item.x * ts + ts/2 - 6, item.y * ts + 12);
            }
          }
          break;
      }
    });
  }

  /**
   * 旧版兼容渲染器（不支持分层的旧地图格式）
   */
  renderLegacy(map, startX, startY, endX, endY) {
    for (let y = startY; y < endY; y++) {
      for (let x = startX; x < endX; x++) {
        const tile = map.getTile(x, y);
        if (!tile) continue;

        // 绘制底层瓦片
        this.drawTile(x, y, tile.low, tile.attr);

        // 绘制高层瓦片
        if (tile.high > 0) {
          this.drawTile(x, y, tile.high, tile.attr);
        }
      }
    }
  }

  /**
   * ★ 分层渲染：分别渲染各层（供调试或特殊需求使用）
   * @param {string} layer - 层级名称 ('ground' | 'object' | 'overlay')
   */
  renderLayer(map, layer, cameraX, cameraY, canvasWidth, canvasHeight) {
    const w = map.width;
    const h = map.height;
    const ts = this.tileSize;

    const startX = Math.max(0, Math.floor(cameraX / ts) - 1);
    const startY = Math.max(0, Math.floor(cameraY / ts) - 1);
    const endX = Math.min(w, Math.ceil((cameraX + canvasWidth) / ts) + 1);
    const endY = Math.min(h, Math.ceil((cameraY + canvasHeight) / ts) + 1);

    let layerData;
    switch (layer) {
      case 'ground':
        layerData = map.layerData?.ground;
        break;
      case 'object':
        layerData = map.layerData?.object;
        break;
      case 'overlay':
        layerData = map.layerData?.overlay;
        break;
      default:
        console.warn(`未知层级: ${layer}`);
        return;
    }

    if (!layerData) return;

    for (let y = startY; y < endY; y++) {
      for (let x = startX; x < endX; x++) {
        const idx = y * w + x;
        const tile = layerData[idx];
        if (tile && tile.tileId > 0) {
          this.drawTile(x, y, tile.tileId, 0);
        }
      }
    }
  }

  /**
   * 渲染物体动画效果
   */
  renderObjectAnimation(obj) {
    if (!this.animationSystem || !obj.animType) return;

    const ts = this.tileSize;
    const px = obj.x * ts;
    const py = obj.y * ts;

    // 根据动画类型添加视觉效果
    switch (obj.animType) {
      case 'fire':
        this.drawFireEffect(px, py, ts);
        break;
      case 'water':
      case 'waterfall':
        this.drawWaterEffect(px, py, ts);
        break;
      case 'torch':
      case 'candle':
        this.drawLightEffect(px, py, ts, obj.animType);
        break;
      case 'magic_circle':
        this.drawMagicCircle(px, py, ts);
        break;
      default:
        // 其他动画类型由animationSystem统一处理
        break;
    }
  }

  /**
   * 火焰效果
   */
  drawFireEffect(px, py, ts) {
    const time = Date.now() / 1000;
    this.ctx.save();

    for (let i = 0; i < 3; i++) {
      const offsetX = Math.sin(time * 3 + i) * 3;
      const offsetY = -Math.abs(Math.cos(time * 4 + i * 0.5)) * 8 - i * 3;
      const alpha = 0.6 - i * 0.15;
      const size = ts * (0.3 + i * 0.15);

      const gradient = this.ctx.createRadialGradient(
        px + ts/2 + offsetX, py + ts/2 + offsetY, 0,
        px + ts/2 + offsetX, py + ts/2 + offsetY, size
      );
      gradient.addColorStop(0, `rgba(255, ${200 - i*50}, 0, ${alpha})`);
      gradient.addColorStop(0.5, `rgba(255, ${100 - i*30}, 0, ${alpha * 0.5})`);
      gradient.addColorStop(1, 'rgba(255, 0, 0, 0)');

      this.ctx.fillStyle = gradient;
      this.ctx.fillRect(px, py, ts, ts);
    }

    this.ctx.restore();
  }

  /**
   * 水流/瀑布效果
   */
  drawWaterEffect(px, py, ts) {
    const time = Date.now() / 1000;
    this.ctx.save();

    this.ctx.globalAlpha = 0.3 + Math.sin(time * 2) * 0.1;

    // 波纹线条
    for (let i = 0; i < 3; i++) {
      const yOffset = ((time * 20 + i * ts/3) % ts);
      this.ctx.strokeStyle = `rgba(100, 180, 255, ${0.4 - i * 0.1})`;
      this.ctx.lineWidth = 1;
      this.ctx.beginPath();

      for (let x = 0; x < ts; x += 4) {
        const waveY = yOffset + Math.sin(x * 0.1 + time * 3) * 3;
        if (x === 0) {
          this.ctx.moveTo(px + x, py + waveY);
        } else {
          this.ctx.lineTo(px + x, py + waveY);
        }
      }
      this.ctx.stroke();
    }

    this.ctx.restore();
  }

  /**
   * 光源效果（火把、蜡烛）
   */
  drawLightEffect(px, py, ts, type) {
    const time = Date.now() / 1000;
    const intensity = type === 'torch' ? 1.2 : 0.8;
    const flicker = Math.sin(time * 8) * 0.1 + Math.sin(time * 13) * 0.05;

    this.ctx.save();

    const centerX = px + ts / 2;
    const centerY = py + ts / 3;
    const radius = ts * (1.2 + flicker) * intensity;

    const gradient = this.ctx.createRadialGradient(
      centerX, centerY, 0,
      centerX, centerY, radius
    );
    gradient.addColorStop(0, `rgba(255, 200, 50, ${0.4 + flicker})`);
    gradient.addColorStop(0.3, `rgba(255, 150, 0, ${0.2 + flicker * 0.5})`);
    gradient.addColorStop(0.7, 'rgba(255, 100, 0, 0.05)');
    gradient.addColorStop(1, 'rgba(255, 50, 0, 0)');

    this.ctx.fillStyle = gradient;
    this.ctx.fillRect(px - radius/2, py - radius/2, ts + radius, ts + radius);

    this.ctx.restore();
  }

  /**
   * 魔法阵效果
   */
  drawMagicCircle(px, py, ts) {
    const time = Date.now() / 1000;
    this.ctx.save();

    const centerX = px + ts / 2;
    const centerY = py + ts / 2;
    const radius = ts * 0.4;

    // 旋转的外圈
    this.ctx.translate(centerX, centerY);
    this.ctx.rotate(time * 0.5);

    this.ctx.strokeStyle = `rgba(150, 100, 255, ${0.5 + Math.sin(time * 2) * 0.2})`;
    this.ctx.lineWidth = 2;
    this.ctx.beginPath();
    this.ctx.arc(0, 0, radius, 0, Math.PI * 2);
    this.ctx.stroke();

    // 内圈符文
    this.ctx.rotate(-time * 1.2);
    this.ctx.strokeStyle = `rgba(200, 150, 255, 0.4)`;
    this.ctx.lineWidth = 1;
    this.ctx.beginPath();
    this.ctx.arc(0, 0, radius * 0.6, 0, Math.PI * 2);
    this.ctx.stroke();

    this.ctx.restore();
  }

  drawTile(x, y, tileIndex, attr) {
    const ts = this.tileSize;
    const px = x * ts;
    const py = y * ts;

    const atlasIndex = tileIndex - 1;

    if (tileIndex === 0) return;

    if (this.tileSet && atlasIndex >= 0) {
      const sx = (atlasIndex % this.tileCol) * ts;
      const sy = Math.floor(atlasIndex / this.tileCol) * ts;

      // 检查瓦片是否超出图集范围（即使在非调试模式下也要检查）
      const outOfBounds = (sx + ts > this.tileSet.width) || (sy + ts > this.tileSet.height);
      if (outOfBounds) {
        if (this.debugMode) {
          this.ctx.fillStyle = '#ff0000';
          this.ctx.fillRect(px, py, ts, ts);
          this.ctx.fillStyle = '#ffffff';
          this.ctx.font = 'bold 10px monospace';
          this.ctx.textAlign = 'center';
          this.ctx.fillText('!' + tileIndex, px + ts / 2, py + ts / 2);
        } else {
          // 非调试模式下，绘制一个占位符而不是崩溃
          this.ctx.fillStyle = '#ff6b6b';
          this.ctx.fillRect(px, py, ts, ts);
          this.ctx.strokeStyle = '#fff';
          this.ctx.lineWidth = 1;
          this.ctx.strokeRect(px, py, ts, ts);
        }
        return;
      }

      this.ctx.drawImage(this.tileSet, sx, sy, ts, ts, px, py, ts, ts);

      if (this.debugMode) {
        this.ctx.fillStyle = 'rgba(0,0,0,0.6)';
        this.ctx.fillRect(px, py + ts - 14, ts, 14);
        this.ctx.fillStyle = '#00ff00';
        this.ctx.font = 'bold 9px monospace';
        this.ctx.textAlign = 'left';
        this.ctx.fillText(`${tileIndex}(${sx},${sy})`, px + 1, py + ts - 3);
      }
    } else {
      const colorIndex = atlasIndex >= 0 ? atlasIndex % this.tileColors.length : 0;
      this.ctx.fillStyle = this.tileColors[colorIndex] || '#ff00ff';
      this.ctx.fillRect(px, py, ts, ts);

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

      this.ctx.strokeStyle = 'rgba(0,0,0,0.3)';
      this.ctx.lineWidth = 1;
      this.ctx.strokeRect(px, py, ts, ts);
    }
  }

  drawPlayer(x, y) {
    const ts = this.tileSize;
    const px = x * ts + ts / 2;
    const py = y * ts + ts / 2;

    this.ctx.fillStyle = '#e94560';
    this.ctx.beginPath();
    this.ctx.arc(px, py, ts / 3, 0, Math.PI * 2);
    this.ctx.fill();

    this.ctx.strokeStyle = '#fff';
    this.ctx.lineWidth = 2;
    this.ctx.stroke();
  }

  drawPlayerByPixel(px, py) {
    const ts = this.tileSize;

    this.ctx.fillStyle = '#e94560';
    this.ctx.beginPath();
    this.ctx.arc(px + ts / 2, py + ts / 2, ts / 3, 0, Math.PI * 2);
    this.ctx.fill();

    this.ctx.strokeStyle = '#fff';
    this.ctx.lineWidth = 2;
    this.ctx.stroke();
  }

  drawPath(path) {
    if (!path || path.length === 0) return;
    const ts = this.tileSize;

    this.ctx.fillStyle = 'rgba(233, 69, 96, 0.5)';
    path.forEach(p => {
      const px = p.x * ts + ts / 2;
      const py = p.y * ts + ts / 2;
      this.ctx.beginPath();
      this.ctx.arc(px, py, 4, 0, Math.PI * 2);
      this.ctx.fill();
    });
  }

  /**
   * 设置图层可见性
   * @param {string} layer - 图层名
   * @param {boolean} visible - 是否可见
   */
  setLayerVisible(layer, visible) {
    if (layer in this.layerRendering) {
      this.layerRendering[layer] = visible;
    }
  }
}

window.MapRenderer = MapRenderer;