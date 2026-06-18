/**
 * 原版千年游戏素材导入器
 * 支持导入原版千年的地图(.map)、贴图(.dds)、精灵(.spr)等资源
 */
class MillenniumImporter {
  constructor() {
    this.tilesetCache = new Map();
    this.mapCache = new Map();
    this.tileInfo = {}; // 瓦片属性信息
  }

  /**
   * 导入原版地图文件
   * 支持两种格式:
   *   1. ATZMAP2格式 (千年3原版客户端): 56字节头 + uint32宽高 + 每瓦片12字节(u32+u32+u32)
   *   2. MAPFILE格式 (本编辑器导出): 128字节头 + uint16宽高 + 每瓦片3字节(u8+u8+u8)
   */
  async importMap(filePath) {
    const buffer = await this.loadFile(filePath);
    const view = new DataView(buffer);

    // 检测文件格式（通过magic或文件大小特征）
    const magic = String.fromCharCode(view.getUint8(0), view.getUint8(1), view.getUint8(2), view.getUint8(3),
                                   view.getUint8(4), view.getUint8(5), view.getUint8(6), view.getUint8(7));

    if (magic.startsWith('ATZMAP')) {
      // ★ ATZMAP2 格式 (千年3原版 .map 文件)
      return this._importAtzMap(buffer, view);
    } else {
      // ★ MAPFILE 格式 (本编辑器导出的 .map 文件)
      return this._importMapFileFormat(buffer, view);
    }
  }

  /**
   * 解析 ATZMAP2 格式 (千年3原版客户端)
   * 头部结构:
   *   0x00: "ATZMAP2" magic (8字节)
   *   0x08: 填充0 (8字节)
   *   0x10: width (uint32 LE)
   *   0x14: height (uint32 LE)
   *   0x18: tileSize (uint32 LE)
   *   0x1C-0x2E: 填充0 (19字节)
   *   0x2F: 瓦片数据开始，每瓦片12字节:
   *     [0]    low (地面层瓦片ID, u8)
   *     [1-2]  图集引用/组ID (uint16 LE, 通常=总瓦片数如0x0640=1600)
   *     [3]    保留 (u8, 通常=0)
   *     [4]    high (高层/装饰层瓦片ID或变体, u8)
   *     [5-11] 填充0 (7字节)
   */
  _importAtzMap(buffer, view) {
    const width = view.getUint32(0x10, true);
    const height = view.getUint32(0x14, true);
    // 总瓦片数在0x30处(uint32)，但实际验证发现数据从0x2F开始
    const totalTiles = width * height;

    console.log(`📦 ATZMAP2格式解析:`);
    console.log(`   尺寸: ${width} x ${height} = ${totalTiles} 瓦片`);

    if (width <= 0 || width > 10000 || height <= 0 || height > 10000) {
      throw new Error(`ATZMAP2: 无效的地图尺寸 ${width} x ${height}`);
    }

    // 验证文件大小：头部0x2F + 每瓦片12字节
    const expectedDataSize = 0x2F + totalTiles * 12;
    if (buffer.byteLength < expectedDataSize - 12) {
      console.warn(`⚠️ ATZMAP2: 文件大小${buffer.byteLength}与预期${expectedDataSize}不完全匹配，尝试继续解析`);
    }

    const tiles = [];
    let offset = 0x2F; // 瓦片数据起始位置（经二进制分析确认）

    for (let y = 0; y < height; y++) {
      const row = [];
      for (let x = 0; x < width; x++) {
        if (offset + 11 >= buffer.byteLength) {
          throw new Error(`ATZMAP2: 数据不足于读取瓦片(${x},${y})`);
        }
        const low = view.getUint8(offset);         offset += 1;
        // 跳过: 图集引用(2字节) + 保留(1字节)
        offset += 3;
        const high = view.getUint8(offset);        offset += 1;
        // 跳过: 填充(7字节)
        offset += 7;
        // attr在ATZMAP2中没有明确字段，默认为0（可在编辑器中手动设置）
        row.push({ low: low & 0xFF, high: high & 0xFF, attr: 0 });
      }
      tiles.push(row);
    }

    console.log(`   ✅ 成功解析 ${totalTiles} 个瓦片 (low范围待渲染验证)`);

    return { width, height, tiles, source: 'millennium-atzmap2' };
  }

  /**
   * 从 ArrayBuffer 直接导入地图（用于编辑器内部加载）
   * 支持 ATZMAP2 和 MAPFILE 两种格式
   */
  async _importMapFromBuffer(buffer) {
    const view = new DataView(buffer);

    // 检测文件格式（通过magic）
    const magic = String.fromCharCode(view.getUint8(0), view.getUint8(1), view.getUint8(2), view.getUint8(3),
                                   view.getUint8(4), view.getUint8(5), view.getUint8(6), view.getUint8(7));

    if (magic.startsWith('ATZMAP')) {
      // ATZMAP2 格式 (千年3原版 .map 文件)
      return this._importAtzMap(buffer, view);
    } else {
      // MAPFILE 格式 (本编辑器导出的 .map 文件)
      return this._importMapFileFormat(buffer, view);
    }
  }

  /**
   * 解析 MAPFILE 格式 (本编辑器导出)
   * 头部结构:
   *   0x00: "MAPFILE" magic (6字节+填充到128字节)
   *   0x7C: tilesetCols (uint16 LE, 偏移124)
   *   0x80: width (uint16 LE)
   *   0x82: height (uint16 LE)
   *   0x84: 瓦片数据开始，每瓦片5字节: low(u16) + high(u16) + attr(u8)
   *   ★ 支持旧格式(3字节/瓦片)和新格式(5字节/瓦片)的自动检测
   */
  _importMapFileFormat(buffer, view) {
    // ★ 读取文件头中的tilesetCols（偏移124）
    const tilesetCols = view.getUint16(124, true);
    
    let offset = 128; // 跳过128字节文件头

    const width = view.getUint16(offset, true);
    offset += 2;
    const height = view.getUint16(offset, true);
    offset += 2;

    // ★ 自动检测瓦片数据格式：通过计算期望文件大小来判断
    // 新格式: 每瓦片5字节(low:2 + high:2 + attr:1)
    // 旧格式: 每瓦片3字节(low:1 + high:1 + attr:1)
    const expectedSizeNewFormat = 128 + 4 + width * height * 5;
    const expectedSizeOldFormat = 128 + 4 + width * height * 3;
    const isNewFormat = Math.abs(buffer.byteLength - expectedSizeNewFormat) <= Math.abs(buffer.byteLength - expectedSizeOldFormat);
    
    console.log(`📦 MAPFILE格式解析: ${isNewFormat ? '新格式(16位ID)' : '旧格式(8位ID)'}`);

    const tiles = [];
    const tileBytes = isNewFormat ? 5 : 3;

    for (let y = 0; y < height; y++) {
      const row = [];
      for (let x = 0; x < width; x++) {
        if (offset + tileBytes > buffer.byteLength) {
          throw new Error(`MAPFILE: 数据不足于读取瓦片(${x},${y})`);
        }
        
        let low, high;
        if (isNewFormat) {
          // ★ 新格式：瓦片ID为16位(uint16 LE)，支持0-65535范围
          low = view.getUint16(offset, true);
          offset += 2;
          high = view.getUint16(offset, true);
          offset += 2;
        } else {
          // ★ 旧格式：瓦片ID为8位(uint8)，兼容旧版本导出文件
          low = view.getUint8(offset++);
          high = view.getUint8(offset++);
        }
        const attr = view.getUint8(offset++);

        row.push({ low, high, attr });
      }
      tiles.push(row);
    }

    return { width, height, tiles, tilesetCols, source: 'mapfile', format: isNewFormat ? 'new' : 'old' };
  }

  /**
   * 批量导入地图文件
   */
  async importMaps(folderPath, mapIds) {
    const results = [];
    for (const id of mapIds) {
      const fileName = `${String(id).padStart(3, '0')}.map`;
      try {
        const map = await this.importMap(`${folderPath}/${fileName}`);
        map.id = id;
        results.push(map);
        console.log(`成功导入地图: ${fileName}`);
      } catch (err) {
        console.warn(`导入地图失败 ${fileName}:`, err);
      }
    }
    return results;
  }

  /**
   * 导入原版DDS贴图并转换为瓦片图集
   * 原版千年使用多个DDS文件存储瓦片贴图
   */
  async importTilesetFromDDS(ddsPaths, tileSize = 32, cols = 16) {
    const images = [];
    
    for (const path of ddsPaths) {
      try {
        const buffer = await this.loadFile(path);
        const parser = new DdsParser();
        const imageData = await parser.load(buffer);
        images.push(imageData);
      } catch (err) {
        console.warn(`加载DDS失败 ${path}:`, err);
      }
    }

    // 合并所有贴图为一个大图集
    const totalTiles = images.reduce((sum, img) => {
      return sum + Math.floor(img.width / tileSize) * Math.floor(img.height / tileSize);
    }, 0);

    const rows = Math.ceil(totalTiles / cols);
    const canvas = document.createElement('canvas');
    canvas.width = cols * tileSize;
    canvas.height = rows * tileSize;
    const ctx = canvas.getContext('2d');

    // 使用drawImage方式合并图片
    let tileIndex = 0;
    for (const imageData of images) {
      const imgTilesX = Math.floor(imageData.width / tileSize);
      const imgTilesY = Math.floor(imageData.height / tileSize);
      
      // 创建临时canvas显示单个DDS图片
      const ddsCanvas = document.createElement('canvas');
      ddsCanvas.width = imageData.width;
      ddsCanvas.height = imageData.height;
      ddsCanvas.getContext('2d').putImageData(imageData, 0, 0);

      for (let ty = 0; ty < imgTilesY; ty++) {
        for (let tx = 0; tx < imgTilesX; tx++) {
          const atlasX = (tileIndex % cols) * tileSize;
          const atlasY = Math.floor(tileIndex / cols) * tileSize;
          
          ctx.drawImage(ddsCanvas, 
            tx * tileSize, ty * tileSize, tileSize, tileSize,
            atlasX, atlasY, tileSize, tileSize);
          
          tileIndex++;
        }
      }
    }

    return canvas;
  }

  /**
   * 导入瓦片属性定义文件
   */
  async importTileAttributes(csvPath) {
    const text = await this.loadTextFile(csvPath);
    const lines = text.split('\n');
    const attributes = {};

    for (const line of lines) {
      const parts = line.split(',');
      if (parts.length >= 3) {
        const index = parseInt(parts[0].trim());
        attributes[index] = {
          name: parts[1].trim(),
          passable: parts[2].trim() === '1',
          layer: parseInt(parts[3]) || 0,
          description: parts[4] || ''
        };
      }
    }

    this.tileInfo = attributes;
    return attributes;
  }

  /**
   * 导出为JSON格式（用于地图编辑器）
   */
  exportToJson(mapData) {
    return JSON.stringify({
      id: mapData.id || 1,
      name: mapData.name || '未命名地图',
      width: mapData.width,
      height: mapData.height,
      tiles: mapData.tiles,
      source: mapData.source || 'imported',
      exportTime: new Date().toISOString()
    }, null, 2);
  }

  /**
   * 下载导出的文件
   */
  downloadExport(data, filename, type = 'json') {
    let content = data;
    let mimeType = 'application/json';
    
    if (type === 'json') {
      if (typeof data !== 'string') {
        content = JSON.stringify(data, null, 2);
      }
    } else if (type === 'binary') {
      content = new Blob([data]);
      mimeType = 'application/octet-stream';
    }

    const blob = typeof content === 'string' ? new Blob([content], { type: mimeType }) : content;
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  }

  /**
   * 加载文件（支持File对象或路径）
   */
  async loadFile(input) {
    if (input instanceof File) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve(e.target.result);
        reader.onerror = reject;
        reader.readAsArrayBuffer(input);
      });
    } else {
      // 路径方式（浏览器环境）
      const response = await fetch(input);
      if (!response.ok) throw new Error(`加载失败: ${input}`);
      return response.arrayBuffer();
    }
  }

  /**
   * 加载文本文件
   */
  async loadTextFile(input) {
    if (input instanceof File) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve(e.target.result);
        reader.onerror = reject;
        reader.readAsText(input);
      });
    } else {
      const response = await fetch(input);
      if (!response.ok) throw new Error(`加载失败: ${input}`);
      return response.text();
    }
  }

  /**
   * 获取瓦片信息
   */
  getTileInfo(index) {
    return this.tileInfo[index] || { name: `瓦片${index}`, passable: true, layer: 0 };
  }

  /**
   * 获取支持的文件类型
   */
  getSupportedTypes() {
    return {
      maps: ['.map'],
      textures: ['.dds'],
      sprites: ['.spr'],
      attributes: ['.csv', '.txt']
    };
  }

  /**
   * 批量导入多个地图文件（通过FileList）
   */
  async importMultipleMaps(files) {
    const results = [];
    const errors = [];
    
    for (const file of files) {
      if (!file.name.endsWith('.map')) continue;
      
      try {
        const map = await this.importMap(file);
        map.fileName = file.name;
        results.push(map);
      } catch (err) {
        errors.push({ file: file.name, error: err.message });
      }
    }
    
    return { success: results, errors };
  }

  /**
   * 导出为编辑器MAPFILE格式(.map)
   * 格式：128字节头 + 宽(2) + 高(2) + (low+high+attr)*n
   * ★ 瓦片ID使用16位(uint16)，支持0-65535范围
   */
  exportToMillenniumFormat(mapData) {
    const { width, height, tiles, tilesetCols = 16 } = mapData;
    const dataSize = 2 + 2 + (width * height * 5); // ★ 修复：每瓦片5字节(low:2 + high:2 + attr:1)
    const buffer = new ArrayBuffer(128 + dataSize);
    const view = new Uint8Array(buffer);
    
    // 文件头（填充为0）
    view.fill(0);
    
    // ★ 写入文件标识 "MAPFILE"（必须写在前6字节）
    const header = 'MAPFILE';
    for (let i = 0; i < header.length; i++) {
      view[i] = header.charCodeAt(i);
    }
    
    // ★ 写入 tilesetCols（偏移124，2字节）
    // 这确保导出和导入时使用相同的图集列数，避免瓦片排序混乱
    view[124] = tilesetCols & 0xFF;
    view[125] = (tilesetCols >> 8) & 0xFF;
    
    // 地图尺寸
    const offset = 128;
    view[offset] = width & 0xFF;
    view[offset + 1] = (width >> 8) & 0xFF;
    view[offset + 2] = height & 0xFF;
    view[offset + 3] = (height >> 8) & 0xFF;
    
    // 瓦片数据
    let dataOffset = offset + 4;
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        const tile = tiles[y]?.[x] || { low: 0, high: 0, attr: 0 };
        // ★ 修复：瓦片ID改为16位(uint16 LE)，支持0-65535范围
        view[dataOffset++] = tile.low & 0xFF;
        view[dataOffset++] = (tile.low >> 8) & 0xFF;
        view[dataOffset++] = (tile.high || 0) & 0xFF;
        view[dataOffset++] = ((tile.high || 0) >> 8) & 0xFF;
        view[dataOffset++] = tile.attr || 0;
        
        // ★ 调试：打印前5个瓦片的导出信息
        if (y < 2 && x < 5) {
          console.log(`导出位置(${x},${y}): 瓦片ID=${tile.low}`);
        }
      }
    }
    
    return buffer;
  }

  /**
   * 导出碰撞数据为二进制格式
   */
  exportCollisionData(mapData) {
    const { width, height, tiles } = mapData;
    const buffer = new Uint8Array(width * height);
    
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        const tile = tiles[y]?.[x] || { attr: 0 };
        buffer[y * width + x] = tile.attr === 1 ? 1 : 0;
      }
    }
    
    return buffer;
  }

  /**
   * 批量导出多个地图
   */
  async exportMultipleMaps(maps, format = 'millennium') {
    const results = [];
    
    for (const map of maps) {
      try {
        let data, filename, type;
        
        if (format === 'millennium') {
          data = this.exportToMillenniumFormat(map);
          filename = `${String(map.id || 1).padStart(3, '0')}.map`;
          type = 'binary';
        } else {
          data = this.exportToJson(map);
          filename = `${map.name || 'map'}_${map.id || 1}.json`;
          type = 'json';
        }
        
        this.downloadExport(data, filename, type);
        results.push({ success: true, filename });
        
        // 添加延迟避免浏览器阻止多个下载
        await new Promise(resolve => setTimeout(resolve, 500));
      } catch (err) {
        results.push({ success: false, error: err.message });
      }
    }
    
    return results;
  }
}

window.MillenniumImporter = MillenniumImporter;
