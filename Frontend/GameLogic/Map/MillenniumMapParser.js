/**
 * 千年 .map 二进制瓦片地图解析器
 */
class MillenniumMapParser {
  constructor() {
    this.width = 0;
    this.height = 0;
    this.tiles = [];
    this.collision = []; // 碰撞矩阵 0=可通行 1=阻挡
    this.tilesetCols = 0; // 瓦片图集列数（从文件头读取，0表示未指定）
  }

  loadMap(arrayBuffer) {
    const view = new DataView(arrayBuffer);

    // 从文件头偏移124读取瓦片图集列数（uint16 LE）
    // 新版地图编辑器会将此值写入文件头，用于游戏渲染器正确定位瓦片
    this.tilesetCols = view.getUint16(124, true);
    if (this.tilesetCols === 0 || this.tilesetCols > 256) {
      this.tilesetCols = 0; // 无效值，标记为未指定
    }

    let offset = 128; // 跳过128字节文件头

    this.width = view.getUint16(offset, true);
    offset += 2;
    this.height = view.getUint16(offset, true);
    offset += 2;

    const total = this.width * this.height;
    this.tiles = [];
    this.collision = [];

    console.log(`📦 地图解析完成:`);
    console.log(`   尺寸: ${this.width} x ${this.height} = ${total} 瓦片`);
    console.log(`   文件头tilesetCols[124]: ${this.tilesetCols || '(未指定/旧格式)'}`);

    for (let i = 0; i < total; i++) {
      const low = view.getUint8(offset++);
      const high = view.getUint8(offset++);
      const attr = view.getUint8(offset++);
      this.tiles.push({ low, high, attr });
      this.collision.push(attr === 1 ? 1 : 0);
    }

    if (this.tiles.length > 0) {
      console.log(`   前10个瓦片: ${this.tiles.slice(0, 10).map(t => t.low).join(',')}`);
    }
  }

  getTile(x, y) {
    if (x < 0 || y < 0 || x >= this.width || y >= this.height) return null;
    const idx = y * this.width + x;
    return this.tiles[idx];
  }
}

window.MillenniumMapParser = MillenniumMapParser;