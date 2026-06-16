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
   * 原版地图格式：128字节头 + 宽(2) + 高(2) + (low+high+attr)*n
   */
  async importMap(filePath) {
    const buffer = await this.loadFile(filePath);
    const view = new DataView(buffer);
    
    let offset = 128; // 跳过文件头
    
    const width = view.getUint16(offset, true);
    offset += 2;
    const height = view.getUint16(offset, true);
    offset += 2;

    const tiles = [];
    const collision = [];

    for (let y = 0; y < height; y++) {
      const row = [];
      for (let x = 0; x < width; x++) {
        const low = view.getUint8(offset++);
        const high = view.getUint8(offset++);
        const attr = view.getUint8(offset++);
        
        row.push({ low, high, attr });
        collision.push(attr === 1 ? 1 : 0);
      }
      tiles.push(row);
    }

    return {
      width,
      height,
      tiles,
      collision,
      source: 'millennium'
    };
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
   * 导出为原版千年地图格式(.map)
   * 格式：128字节头 + 宽(2) + 高(2) + (low+high+attr)*n
   */
  exportToMillenniumFormat(mapData) {
    const { width, height, tiles } = mapData;
    const dataSize = 2 + 2 + (width * height * 3); // 宽+高+瓦片数据
    const buffer = new ArrayBuffer(128 + dataSize);
    const view = new Uint8Array(buffer);
    
    // 文件头（填充为0）
    view.fill(0);
    
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
        const tile = tiles[y]?.[x] || { low: 1, high: 0, attr: 0 };
        view[dataOffset++] = tile.low || 1;
        view[dataOffset++] = tile.high || 0;
        view[dataOffset++] = tile.attr || 0;
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
