/**
 * 千年 .map 二进制瓦片地图解析器
 */
class MillenniumMapParser {
  constructor() {
    this.width = 0;
    this.height = 0;
    this.tiles = [];
    this.collision = []; // 碰撞矩阵 0=可通行 1=阻挡
  }

  loadMap(arrayBuffer) {
    const view = new DataView(arrayBuffer);
    let offset = 128; // 跳过128字节文件头

    this.width = view.getUint16(offset, true);
    offset += 2;
    this.height = view.getUint16(offset, true);
    offset += 2;

    const total = this.width * this.height;
    this.tiles = [];
    this.collision = [];

    for (let i = 0; i < total; i++) {
      const low = view.getUint8(offset++);
      const high = view.getUint8(offset++);
      const attr = view.getUint8(offset++);
      this.tiles.push({ low, high, attr });
      this.collision.push(attr === 1 ? 1 : 0);
    }
    console.log(`地图加载完成：${this.width} x ${this.height}`);
  }

  getTile(x, y) {
    if (x < 0 || y < 0 || x >= this.width || y >= this.height) return null;
    const idx = y * this.width + x;
    return this.tiles[idx];
  }
}

window.MillenniumMapParser = MillenniumMapParser;