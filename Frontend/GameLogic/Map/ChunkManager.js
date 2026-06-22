/**
 * 地图分块加载管理器
 *
 * ★ 设计目标：
 *   1. 支持超大地图（10000×10000+）按需加载，避免一次性占用过多内存
 *   2. 玩家周围 3×3 区块（9块）常驻内存，远端区块自动卸载
 *   3. 对上层透明：MapEngine/MapRenderer 通过统一接口访问瓦片数据
 *   4. 兼容小地图：地图较小时自动退化为全量加载模式
 *
 * ★ 工作原理：
 *   - 将整张地图切成 CHUNK_SIZE × CHUNK_SIZE 的区块
 *   - 每个区块独立存储瓦片数据（ground/object/overlay/collision）
 *   - 玩家移动时动态加载新进入视野的区块，卸载远离的区块
 *   - 使用 LRU 策略管理区块缓存，限制最大内存占用
 *
 * ★ 两种模式：
 *   1. 全量模式（地图 ≤ THRESHOLD）：一次性加载所有数据，不分块
 *   2. 分块模式（地图 > THRESHOLD）：按需加载区块
 *
 * ★ 数据来源：
 *   - 方式A（默认）：从完整 .map 文件切片（前端处理，服务端无需改动）
 *   - 方式B（可选）：从服务端按区块请求（需要服务端支持 ?chunk=x,y 接口）
 */
class ChunkManager {
  /**
   * @param {object} options
   * @param {number} [options.chunkSize=64] - 区块尺寸（瓦片数）
   * @param {number} [options.loadRadius=1] - 加载半径（1=3×3区块，2=5×5区块）
   * @param {number} [options.maxChunks=25] - 最大缓存区块数（LRU 淘汰）
   * @param {number} [options.threshold=200000] - 启用分块模式的阈值（总瓦片数）
   * @param {function} [options.chunkLoader] - 自定义区块加载器 (chunkX, chunkY) => Promise<ChunkData>
   */
  constructor(options = {}) {
    this.chunkSize = options.chunkSize || 64;
    this.loadRadius = options.loadRadius || 1;
    this.maxChunks = options.maxChunks || 25;
    this.threshold = options.threshold || 200000; // 约 450×450 地图
    this.chunkLoader = options.chunkLoader || null;

    // 地图元信息
    this.mapWidth = 0;
    this.mapHeight = 0;
    this.chunkCols = 0;  // 区块列数
    this.chunkRows = 0;  // 区块行数
    this.enabled = false; // 是否启用分块模式

    // 区块缓存：Map<chunkKey, {data, lastAccess}>
    // chunkKey = chunkY * chunkCols + chunkX
    this.chunks = new Map();

    // 加载中的区块：Set<chunkKey>，防止重复加载
    this.loadingChunks = new Set();

    // 完整地图数据（全量模式时使用）
    this.fullMapData = null;

    // ★ 区块卸载回调：(chunk) => void
    //    MapEngine 注册此回调，在区块卸载时重置对应区域的 collision 数据
    //    避免 A* 寻路将已卸载区块误判为可通行
    this.onChunkUnload = null;

    // 统计信息
    this.stats = {
      chunksLoaded: 0,
      chunksUnloaded: 0,
      cacheHits: 0,
      cacheMisses: 0
    };
  }

  /**
   * 初始化分块管理器
   * 根据地图大小自动决定是否启用分块模式
   */
  initialize(mapWidth, mapHeight) {
    this.mapWidth = mapWidth;
    this.mapHeight = mapHeight;
    this.chunkCols = Math.ceil(mapWidth / this.chunkSize);
    this.chunkRows = Math.ceil(mapHeight / this.chunkSize);

    const totalTiles = mapWidth * mapHeight;
    this.enabled = totalTiles > this.threshold;

    console.log(`📦 ChunkManager 初始化:`);
    console.log(`   地图尺寸: ${mapWidth} × ${mapHeight} = ${totalTiles} 瓦片`);
    console.log(`   区块尺寸: ${this.chunkSize} × ${this.chunkSize}`);
    console.log(`   区块数量: ${this.chunkCols} × ${this.chunkRows} = ${this.chunkCols * this.chunkRows}`);
    console.log(`   模式: ${this.enabled ? '分块加载' : '全量加载'}`);
    if (this.enabled) {
      console.log(`   加载半径: ${this.loadRadius} (视野 ${(this.loadRadius * 2 + 1) ** 2} 块)`);
      console.log(`   最大缓存: ${this.maxChunks} 块`);
    }

    return this.enabled;
  }

  /**
   * 从完整地图数据切分区块（方式A）
   * 在全量模式下直接保存引用，在分块模式下按需切片
   * @param {object} mapData - 完整地图数据 {tiles, collision, layerData, objects}
   */
  setFullMapData(mapData) {
    if (!this.enabled) {
      // 全量模式：直接保存引用
      this.fullMapData = mapData;
      return;
    }

    // 分块模式：保存完整数据用于切片，但按需访问
    // 注意：这里仍保存完整数据，但通过 getTile 访问时会按区块管理
    // 真正的内存优化需要服务端按区块提供数据（方式B）
    this.fullMapData = mapData;
    console.log(`   ⚠️ 分块模式使用前端切片（方式A），内存优化有限`);
    console.log(`   ⚠️ 如需真正内存优化，请配置 chunkLoader 使用服务端按区块加载（方式B）`);
  }

  /**
   * 获取指定位置的瓦片数据
   * @param {number} x - 瓦片X坐标
   * @param {number} y - 瓦片Y坐标
   * @param {string} layer - 层级 'ground' | 'object' | 'overlay'
   * @returns {object|null} 瓦片数据
   */
  getTile(x, y, layer) {
    if (x < 0 || y < 0 || x >= this.mapWidth || y >= this.mapHeight) {
      return null;
    }

    if (!this.enabled) {
      // 全量模式：直接访问
      if (!this.fullMapData) return null;
      const idx = y * this.mapWidth + x;
      return this.fullMapData.layerData[layer][idx] || null;
    }

    // 分块模式：按区块访问
    const chunkX = Math.floor(x / this.chunkSize);
    const chunkY = Math.floor(y / this.chunkSize);
    const chunk = this._getChunk(chunkX, chunkY);

    if (!chunk) {
      this.stats.cacheMisses++;
      return null;
    }

    this.stats.cacheHits++;
    // 区块内局部坐标
    const localX = x - chunkX * this.chunkSize;
    const localY = y - chunkY * this.chunkSize;
    const localIdx = localY * this.chunkSize + localX;
    return chunk.data[layer][localIdx] || null;
  }

  /**
   * 获取指定位置的碰撞值
   */
  getCollision(x, y) {
    if (x < 0 || y < 0 || x >= this.mapWidth || y >= this.mapHeight) {
      return 1; // 越界视为阻挡
    }

    if (!this.enabled) {
      if (!this.fullMapData) return 1;
      return this.fullMapData.collision[y * this.mapWidth + x] || 0;
    }

    const chunkX = Math.floor(x / this.chunkSize);
    const chunkY = Math.floor(y / this.chunkSize);
    const chunk = this._getChunk(chunkX, chunkY);
    if (!chunk) return 1;

    const localX = x - chunkX * this.chunkSize;
    const localY = y - chunkY * this.chunkSize;
    return chunk.data.collision[localY * this.chunkSize + localX] || 0;
  }

  /**
   * 获取区块（带缓存）
   */
  _getChunk(chunkX, chunkY) {
    const key = this._chunkKey(chunkX, chunkY);
    const chunk = this.chunks.get(key);

    if (chunk) {
      chunk.lastAccess = Date.now();
      return chunk;
    }

    // 区块未加载，尝试同步加载（仅前端切片模式可用）
    if (this.fullMapData && !this.chunkLoader) {
      this._loadChunkFromFullData(chunkX, chunkY);
      return this.chunks.get(key);
    }

    // 异步加载模式：返回 null，由调用方触发预加载
    return null;
  }

  /**
   * 从完整地图数据切分区块（同步）
   */
  _loadChunkFromFullData(chunkX, chunkY) {
    const key = this._chunkKey(chunkX, chunkY);
    if (this.chunks.has(key) || this.loadingChunks.has(key)) return;

    const startTileX = chunkX * this.chunkSize;
    const startTileY = chunkY * this.chunkSize;
    const endTileX = Math.min(startTileX + this.chunkSize, this.mapWidth);
    const endTileY = Math.min(startTileY + this.chunkSize, this.mapHeight);
    const chunkW = endTileX - startTileX;
    const chunkH = endTileY - startTileY;

    // 切分各层数据
    const ground = new Array(chunkW * chunkH);
    const object = new Array(chunkW * chunkH);
    const overlay = new Array(chunkW * chunkH);
    const collision = new Uint8Array(chunkW * chunkH);

    for (let y = 0; y < chunkH; y++) {
      for (let x = 0; x < chunkW; x++) {
        const srcIdx = (startTileY + y) * this.mapWidth + (startTileX + x);
        const dstIdx = y * chunkW + x;
        ground[dstIdx] = this.fullMapData.layerData.ground[srcIdx] || null;
        object[dstIdx] = this.fullMapData.layerData.object[srcIdx] || null;
        overlay[dstIdx] = this.fullMapData.layerData.overlay[srcIdx] || null;
        collision[dstIdx] = this.fullMapData.collision[srcIdx] || 0;
      }
    }

    const chunk = {
      chunkX, chunkY,
      startTileX, startTileY,
      width: chunkW, height: chunkH,
      data: { ground, object, overlay, collision },
      lastAccess: Date.now()
    };

    this._addChunk(key, chunk);
  }

  /**
   * 异步加载区块（服务端按区块加载模式）
   */
  async _loadChunkAsync(chunkX, chunkY) {
    const key = this._chunkKey(chunkX, chunkY);
    if (this.chunks.has(key) || this.loadingChunks.has(key)) return;

    if (!this.chunkLoader) return;

    this.loadingChunks.add(key);
    try {
      const data = await this.chunkLoader(chunkX, chunkY);
      if (!data) return;

      const chunk = {
        chunkX, chunkY,
        startTileX: chunkX * this.chunkSize,
        startTileY: chunkY * this.chunkSize,
        width: data.width || this.chunkSize,
        height: data.height || this.chunkSize,
        data: data,
        lastAccess: Date.now()
      };
      this._addChunk(key, chunk);
    } catch (err) {
      console.error(`区块 (${chunkX},${chunkY}) 加载失败:`, err);
    } finally {
      this.loadingChunks.delete(key);
    }
  }

  /**
   * 添加区块到缓存，执行 LRU 淘汰
   */
  _addChunk(key, chunk) {
    // 缓存已满，淘汰最久未访问的区块
    while (this.chunks.size >= this.maxChunks) {
      this._evictLRU();
    }

    this.chunks.set(key, chunk);
    this.stats.chunksLoaded++;
  }

  /**
   * LRU 淘汰：移除最久未访问的区块
   */
  _evictLRU() {
    let oldestKey = null;
    let oldestTime = Infinity;
    for (const [key, chunk] of this.chunks) {
      if (chunk.lastAccess < oldestTime) {
        oldestTime = chunk.lastAccess;
        oldestKey = key;
      }
    }
    if (oldestKey !== null) {
      this._removeChunk(oldestKey);
    }
  }

  /**
   * ★ 统一的区块移除方法
   *    从缓存中删除区块，并触发 onChunkUnload 回调
   *    让上层（MapEngine）有机会清理与该区块相关的派生数据（如 collision 数组）
   *
   * @param {string|number} key - 区块缓存键
   */
  _removeChunk(key) {
    const chunk = this.chunks.get(key);
    if (!chunk) return;
    this.chunks.delete(key);
    this.stats.chunksUnloaded++;
    // 触发回调，让 MapEngine 重置对应区域的 collision 数据
    if (typeof this.onChunkUnload === 'function') {
      try {
        this.onChunkUnload(chunk);
      } catch (err) {
        console.error('onChunkUnload 回调执行失败:', err);
      }
    }
  }

  /**
   * 区块坐标转缓存键
   */
  _chunkKey(chunkX, chunkY) {
    return chunkY * this.chunkCols + chunkX;
  }

  /**
   * 根据玩家位置预加载周围区块
   * 应在玩家移动时调用
   * @param {number} playerTileX - 玩家瓦片X坐标
   * @param {number} playerTileY - 玩家瓦片Y坐标
   */
  async preloadChunks(playerTileX, playerTileY) {
    if (!this.enabled) return 0;

    const centerChunkX = Math.floor(playerTileX / this.chunkSize);
    const centerChunkY = Math.floor(playerTileY / this.chunkSize);

    const tasks = [];
    let syncLoaded = 0;  // 同步切片加载的区块数
    for (let dy = -this.loadRadius; dy <= this.loadRadius; dy++) {
      for (let dx = -this.loadRadius; dx <= this.loadRadius; dx++) {
        const cx = centerChunkX + dx;
        const cy = centerChunkY + dy;
        if (cx < 0 || cy < 0 || cx >= this.chunkCols || cy >= this.chunkRows) continue;

        const key = this._chunkKey(cx, cy);
        if (!this.chunks.has(key) && !this.loadingChunks.has(key)) {
          if (this.chunkLoader) {
            // 异步加载
            tasks.push(this._loadChunkAsync(cx, cy));
          } else if (this.fullMapData) {
            // 同步切片
            this._loadChunkFromFullData(cx, cy);
            syncLoaded++;
          }
        }
      }
    }

    if (tasks.length > 0) {
      await Promise.all(tasks);
      return tasks.length;  // 异步加载的区块数
    }

    return syncLoaded;  // 同步加载的区块数（无新加载则为0）
  }

  /**
   * 卸载远离玩家的区块
   * 应定期调用（如每秒一次）释放内存
   * @param {number} playerTileX - 玩家瓦片X坐标
   * @param {number} playerTileY - 玩家瓦片Y坐标
   */
  unloadDistantChunks(playerTileX, playerTileY) {
    if (!this.enabled) return;

    const centerChunkX = Math.floor(playerTileX / this.chunkSize);
    const centerChunkY = Math.floor(playerTileY / this.chunkSize);
    const keepRadius = this.loadRadius + 1; // 保留一圈缓冲

    const toRemove = [];
    for (const [key, chunk] of this.chunks) {
      const dx = Math.abs(chunk.chunkX - centerChunkX);
      const dy = Math.abs(chunk.chunkY - centerChunkY);
      if (dx > keepRadius || dy > keepRadius) {
        toRemove.push(key);
      }
    }

    for (const key of toRemove) {
      this._removeChunk(key);
    }

    if (toRemove.length > 0) {
      console.log(`🗑️ 卸载 ${toRemove.length} 个远端区块，当前缓存 ${this.chunks.size} 块`);
    }
  }

  /**
   * 检查区块是否已加载
   */
  isChunkLoaded(chunkX, chunkY) {
    return this.chunks.has(this._chunkKey(chunkX, chunkY));
  }

  /**
   * 获取分块统计信息
   */
  getStats() {
    return {
      ...this.stats,
      cacheSize: this.chunks.size,
      maxChunks: this.maxChunks,
      enabled: this.enabled
    };
  }

  /**
   * 清空所有缓存
   */
  clear() {
    this.chunks.clear();
    this.loadingChunks.clear();
    this.fullMapData = null;
  }

  /**
   * ★ 解析服务端返回的二进制区块数据
   * 服务端格式：[width:u16 LE][height:u16 LE][每瓦片5字节: low:u16 LE, high:u16 LE, attr:u8]
   *
   * overlay 层无需额外编码：信息已包含在 high/attr 中
   *   - attr === 2 或 high > 128 → 覆盖层（与 MillenniumMapParser.initLayerData 保持一致）
   *   - 覆盖层 tileId = high > 128 ? high - 128 : high
   *
   * @param {ArrayBuffer} buffer - 服务端返回的二进制数据
   * @returns {object} 解码后的区块数据 {width, height, ground, object, overlay, collision}
   */
  static decodeChunkBuffer(buffer) {
    const view = new DataView(buffer);
    const chunkW = view.getUint16(0, true); // little-endian
    const chunkH = view.getUint16(2, true);
    const total = chunkW * chunkH;

    const ground = new Array(total);
    const object = new Array(total);
    const overlay = new Array(total);
    const collision = new Uint8Array(total);

    let pos = 4;
    for (let i = 0; i < total; i++) {
      const low = view.getUint16(pos, true);
      const high = view.getUint16(pos + 2, true);
      const attr = view.getUint8(pos + 4);
      pos += 5;

      const x = i % chunkW;
      const y = Math.floor(i / chunkW);

      // ★ 与 MillenniumMapParser.initLayerData 保持一致的分层逻辑
      // 地面层：始终使用 low 值
      if (low > 0) {
        ground[i] = {
          tileId: low,
          x, y,
          zIndex: 0,
          blocksView: false
        };
      } else {
        ground[i] = null;
      }

      // 物体层：high > 0 且 high <= 128 且 attr !== 2（非覆盖层）
      if (high > 0 && high <= 128 && attr !== 2) {
        object[i] = {
          tileId: high,
          x, y,
          zIndex: 1,
          blocksView: false
        };
      } else {
        object[i] = null;
      }

      // ★ 覆盖层：attr === 2 或 high > 128（桥梁、屋顶等遮挡角色）
      if (attr === 2 || high > 128) {
        overlay[i] = {
          tileId: high > 128 ? high - 128 : high,
          x, y,
          zIndex: 2,
          isOverlay: true,
          blocksView: true  // 标记为可遮挡下方角色
        };
      } else {
        overlay[i] = null;
      }

      collision[i] = (attr === 1) ? 1 : 0;
    }

    return { width: chunkW, height: chunkH, ground, object, overlay, collision };
  }

  /**
   * ★ 创建服务端模式的 ChunkManager
   * 使用 HTTP 接口按需加载区块，无需下载完整 .map 文件
   *
   * @param {object} options
   * @param {number} options.mapId - 地图ID
   * @param {string} [options.apiBase] - 服务端API基础路径（默认 '/api/map'）
   * @param {number} [options.chunkSize=64] - 区块尺寸
   * @param {number} [options.loadRadius=1] - 加载半径
   * @param {number} [options.maxChunks=25] - 最大缓存区块数
   * @returns {ChunkManager} 配置好的 ChunkManager 实例
   */
  static createServerMode(options) {
    const {
      mapId,
      apiBase = '/api/map',
      chunkSize = 64,
      loadRadius = 1,
      maxChunks = 25
    } = options;

    const cm = new ChunkManager({
      chunkSize,
      loadRadius,
      maxChunks,
      threshold: 0 // 强制启用分块模式
    });

    // 配置 chunkLoader：通过 HTTP 获取二进制区块数据
    cm.chunkLoader = async (chunkX, chunkY) => {
      const url = `${apiBase}/${mapId}/chunk/${chunkX}/${chunkY}`;
      try {
        const resp = await fetch(url);
        if (!resp.ok) {
          console.warn(`区块 (${chunkX},${chunkY}) 加载失败: HTTP ${resp.status}`);
          return null;
        }
        const buffer = await resp.arrayBuffer();
        return ChunkManager.decodeChunkBuffer(buffer);
      } catch (err) {
        console.error(`区块 (${chunkX},${chunkY}) 网络请求失败:`, err);
        return null;
      }
    };

    return cm;
  }

  /**
   * ★ 查询服务端获取区块划分信息
   * 用于在加载地图前决定是否启用分块模式
   *
   * @param {number} mapId - 地图ID
   * @param {string} [apiBase] - 服务端API基础路径
   * @returns {Promise<object|null>} 区块信息
   */
  static async queryChunkInfo(mapId, apiBase = '/api/map') {
    try {
      const resp = await fetch(`${apiBase}/${mapId}/chunk_info`);
      if (!resp.ok) return null;
      const json = await resp.json();
      if (json.code !== 200) return null;
      return json.data;
    } catch (err) {
      console.warn('查询区块信息失败:', err);
      return null;
    }
  }
}

window.ChunkManager = ChunkManager;
