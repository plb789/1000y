/**
 * 千年 .map 二进制瓦片地图解析器
 * 支持三层地图系统：地面层(0) / 物体层(1) / 顶层/覆盖层(2)
 */
class MillenniumMapParser {
  constructor() {
    this.width = 0;
    this.height = 0;
    this.tiles = [];
    this.collision = []; // 碰撞矩阵 0=可通行 1=阻挡
    this.tilesetCols = 0; // 瓦片图集列数（从文件头读取，0表示未指定）

    // ★ 新增：分层系统支持
    this.objects = [];        // 地图上的物体对象 [{id, x, y, tileId, layer, zIndex, animType, ...}]
    this.topLayer = [];       // 顶层瓦片数据（覆盖层，用于桥梁等遮挡效果）
    this.layerData = {        // 分层瓦片数据
      ground: [],             // Layer 0: 地面层
      object: [],             // Layer 1: 物体层
      overlay: []             // Layer 2: 覆盖层（桥梁、屋顶）
    };
  }

  loadMap(arrayBuffer) {
    const view = new DataView(arrayBuffer);

    // 从文件头偏移124读取瓦片图集列数（uint16 LE）
    this.tilesetCols = view.getUint16(124, true);
    if (this.tilesetCols === 0 || this.tilesetCols > 256) {
      console.warn(`⚠️ 文件头tilesetCols[124]=${this.tilesetCols}无效，将使用图片宽度计算或默认值`);
      this.tilesetCols = 0;
    }

    let offset = 128; // 跳过128字节文件头

    this.width = view.getUint16(offset, true);
    offset += 2;
    this.height = view.getUint16(offset, true);
    offset += 2;

    const total = this.width * this.height;
    this.tiles = [];
    this.collision = [];

    // ★ 自动检测文件格式：通过计算期望文件大小来判断
    // 新格式: 每瓦片5字节(low:2 + high:2 + attr:1)
    // 旧格式: 每瓦片3字节(low:1 + high:1 + attr:1)
    const expectedSizeNewFormat = 128 + 4 + total * 5;
    const expectedSizeOldFormat = 128 + 4 + total * 3;
    const isNewFormat = Math.abs(arrayBuffer.byteLength - expectedSizeNewFormat) <= Math.abs(arrayBuffer.byteLength - expectedSizeOldFormat);
    
    console.log(`📦 地图解析完成:`);
    console.log(`   尺寸: ${this.width} x ${this.height} = ${total} 瓦片`);
    console.log(`   文件头tilesetCols[124]: ${this.tilesetCols || '(未指定/旧格式)'}`);
    console.log(`   格式: ${isNewFormat ? '新格式(16位ID)' : '旧格式(8位ID)'}`);

    for (let i = 0; i < total; i++) {
      let low, high;
      if (isNewFormat) {
        // 新格式：16位瓦片ID
        low = view.getUint16(offset, true);
        offset += 2;
        high = view.getUint16(offset, true);
        offset += 2;
      } else {
        // 旧格式：8位瓦片ID
        low = view.getUint8(offset++);
        high = view.getUint8(offset++);
      }
      const attr = view.getUint8(offset++);

      const tileData = { low, high, attr };

      // ★ 新增：根据high值自动分层
      // high=0: 仅地面层 | high>0且attr!=2: 物体层 | attr==2或特殊标记: 覆盖层
      if (high > 0 && attr !== 2) {
        tileData.objectLayer = high;  // 物体层瓦片ID
      }
      if (attr === 2 || (high > 128)) {  // attr=2传送点或high>128标记为覆盖层
        tileData.overlayLayer = high;
      }

      this.tiles.push(tileData);
      this.collision.push(attr === 1 ? 1 : 0);
    }

    // ★ 初始化分层数据
    this.initLayerData();

    if (this.tiles.length > 0) {
      console.log(`   前10个瓦片: ${this.tiles.slice(0, 10).map(t => t.low).join(',')}`);
    }

    // 尝试从二进制尾部加载物体数据（新版格式）
    this.loadObjectsFromBinary(view, offset);
  }

  /**
   * 初始化分层瓦片数据（用于Z轴排序渲染）
   */
  initLayerData() {
    const total = this.width * this.height;
    this.layerData.ground = new Array(total);
    this.layerData.object = new Array(total);
    this.layerData.overlay = new Array(total);

    for (let i = 0; i < total; i++) {
      const tile = this.tiles[i];
      const x = i % this.width;
      const y = Math.floor(i / this.width);

      // 地面层：始终有low值
      this.layerData.ground[i] = {
        x, y,
        tileId: tile.low,
        layer: 0,
        zIndex: this.calculateZIndex(x, y, 0)  // 使用统一方法计算
      };

      // 物体层：high值且非覆盖层
      if (tile.high > 0 && tile.high <= 128 && tile.attr !== 2) {
        this.layerData.object[i] = {
          x, y,
          tileId: tile.high,
          layer: 1,
          zIndex: this.calculateZIndex(x, y, 1),  // 使用统一方法计算
          isObject: true
        };
      }

      // 覆盖层：桥梁、屋顶等需要遮挡角色的物体
      if (tile.attr === 2 || tile.high > 128) {
        this.layerData.overlay[i] = {
          x, y,
          tileId: tile.high > 128 ? tile.high - 128 : tile.high,
          layer: 2,
          zIndex: this.calculateZIndex(x, y, 2),  // 使用统一方法计算
          isOverlay: true,
          blocksView: true  // 标记为可遮挡下方角色
        };
      }
    }
  }

  /**
   * 从二进制数据加载物体对象（新版.map格式扩展）
   */
  loadObjectsFromBinary(view, offset) {
    try {
      // 检查是否有额外数据（物体列表）
      if (offset + 4 <= view.byteLength) {
        const objCount = view.getUint32(offset, true);
        offset += 4;

        if (objCount > 0 && objCount < 10000) {  // 合理范围检查
          this.objects = [];

          for (let i = 0; i < objCount; i++) {
            if (offset + 12 > view.byteLength) break;

            const obj = {
              id: view.getUint32(offset, true)
            };
            offset += 4;
            obj.x = view.getUint16(offset, true);
            offset += 2;
            obj.y = view.getUint16(offset, true);
            offset += 2;
            obj.tileId = view.getUint16(offset, true);
            offset += 2;
            obj.layer = view.getUint8(offset);       // 0=地面 1=物体 2=覆盖
            offset += 1;
            obj.zIndex = view.getFloat32(offset, true);
            offset += 4;
            obj.animType = '';                          // 动画类型
            obj.properties = {};                         // 自定义属性

            this.objects.push(obj);
          }

          console.log(`🎯 加载物体对象: ${this.objects.length} 个`);
        }
      }
    } catch (e) {
      console.log('ℹ️ 无物体数据或格式不兼容（旧版地图）');
      this.objects = [];
    }
  }

  /**
   * 添加物体到地图
   */
  addObject(obj) {
    const newObj = {
      id: Date.now() + Math.random(),  // 唯一ID
      x: obj.x,
      y: obj.y,
      tileId: obj.tileId,
      layer: obj.layer || 1,           // 默认物体层
      zIndex: this.calculateZIndex(obj.x, obj.y, obj.layer),
      animType: obj.animType || '',
      properties: obj.properties || {},
      width: obj.width || 1,
      height: obj.height || 1
    };

    this.objects.push(newObj);

    // 如果是覆盖层物体，更新overlay数据
    if (newObj.layer === 2) {
      const idx = newObj.y * this.width + newObj.x;
      if (idx >= 0 && idx < this.layerData.overlay.length) {
        this.layerData.overlay[idx] = {
          x: newObj.x,
          y: newObj.y,
          tileId: newObj.tileId,
          layer: 2,
          zIndex: newObj.zIndex,
          isOverlay: true,
          blocksView: true,
          objectId: newObj.id
        };
      }
    }

    return newObj;
  }

  /**
   * 移除物体
   */
  removeObject(objectId) {
    const idx = this.objects.findIndex(o => o.id === objectId);
    if (idx >= 0) {
      const obj = this.objects[idx];
      this.objects.splice(idx, 1);

      // 清理对应的layerData
      if (obj.layer === 2) {
        const dataIdx = obj.y * this.width + obj.x;
        if (this.layerData.overlay[dataIdx] && this.layerData.overlay[dataIdx].objectId === objectId) {
          this.layerData.overlay[dataIdx] = null;
        }
      }

      return true;
    }
    return false;
  }

  /**
   * 移动物体位置
   */
  moveObject(objectId, newX, newY) {
    const obj = this.objects.find(o => o.id === objectId);
    if (obj) {
      obj.x = newX;
      obj.y = newY;
      obj.zIndex = this.calculateZIndex(newX, newY, obj.layer);
      return true;
    }
    return false;
  }

  /**
   * 计算Z轴索引（用于Y-Sort深度排序）
   * @param {number} x - 瓦片X坐标
   * @param {number} y - 瓦片Y坐标
   * @param {number} layer - 层级 (0=地面, 1=物体, 2=覆盖)
   */
  calculateZIndex(x, y, layer = 0) {
    // Y-Sort算法：主要按Y坐标排序，同Y时按X排序
    // 层级作为微调偏移：地面<物体<覆盖<角色
    // 使用整数计算避免浮点精度问题
    const baseZ = y * this.width + x;
    const layerOffset = Math.floor(layer * 1000);  // 使用整数偏移避免精度问题

    return baseZ + layerOffset;
  }

  /**
   * 获取指定位置的Z排序渲染列表（包含瓦片和物体）
   * 用于游戏渲染器按正确顺序绘制
   * @param {number} playerX - 角色X坐标（瓦片坐标）
   * @param {number} playerY - 角色Y坐标（瓦片坐标）
   * @returns {Array} 排序后的渲染项列表
   */
  getZSortedRenderList(playerX, playerY) {
    const renderList = [];

    // 1. 添加所有地面层瓦片
    this.layerData.ground.forEach((tile, idx) => {
      if (tile && tile.tileId > 0) {
        renderList.push({
          type: 'tile',
          layer: 0,
          x: tile.x,
          y: tile.y,
          tileId: tile.tileId,
          zIndex: tile.zIndex
        });
      }
    });

    // 2. 添加物体层瓦片
    this.layerData.object.forEach((tile, idx) => {
      if (tile && tile.tileId > 0) {
        renderList.push({
          type: 'tile_object',
          layer: 1,
          x: tile.x,
          y: tile.y,
          tileId: tile.tileId,
          zIndex: tile.zIndex
        });
      }
    });

    // 3. 添加独立物体对象
    this.objects.forEach(obj => {
      renderList.push({
        type: 'object',
        layer: obj.layer,
        x: obj.x,
        y: obj.y,
        tileId: obj.tileId,
        zIndex: obj.zIndex,
        objectId: obj.id,
        animType: obj.animType,
        properties: obj.properties
      });
    });

    // 4. 添加角色（使用动态Z索引）
    if (playerX !== undefined && playerY !== undefined) {
      renderList.push({
        type: 'player',
        layer: 1.5,  // 角色在物体层和覆盖层之间
        x: playerX,
        y: playerY,
        zIndex: this.calculateZIndex(playerX, playerY, 1.5)  // 使用统一方法计算
      });
    }

    // 5. 添加覆盖层瓦片（桥梁等）- 这些在角色之后渲染以产生遮挡效果
    this.layerData.overlay.forEach((tile, idx) => {
      if (tile && tile.tileId > 0) {
        renderList.push({
          type: 'tile_overlay',
          layer: 2,
          x: tile.x,
          y: tile.y,
          tileId: tile.tileId,
          zIndex: tile.zIndex,
          blocksView: tile.blocksView
        });
      }
    });

    // ★ 关键：按zIndex排序实现正确的遮挡关系
    renderList.sort((a, b) => a.zIndex - b.zIndex);

    return renderList;
  }

  /**
   * 检查某位置是否被覆盖层遮挡（用于判断角色是否在桥下）
   * @param {number} x - 瓦片X坐标
   * @param {number} y - 瓦片Y坐标
   * @returns {boolean} 是否被遮挡
   */
  isBlockedByOverlay(x, y) {
    if (x < 0 || y < 0 || x >= this.width || y >= this.height) return false;

    const idx = y * this.width + x;
    const overlay = this.layerData.overlay[idx];

    return overlay && overlay.blocksView === true;
  }

  getTile(x, y) {
    if (x < 0 || y < 0 || x >= this.width || y >= this.height) return null;
    const idx = y * this.width + x;
    return this.tiles[idx];
  }

  /**
   * 导出包含物体的完整地图数据（用于保存为新格式）
   */
  exportWithObjects() {
    return {
      width: this.width,
      height: this.height,
      tiles: this.tiles,
      objects: this.objects,
      layerData: this.layerData,
      tilesetCols: this.tilesetCols,
      version: '2.0'  // 标记为支持分层的新版本
    };
  }
}

window.MillenniumMapParser = MillenniumMapParser;