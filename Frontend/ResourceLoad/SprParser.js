/**
 * 千年 .spr 帧动画 + .pal 调色板解析
 */
class SprParser {
  constructor() {
    this.frameCount = 0;
    this.frames = [];
    this.palette = [];
  }

  loadPalette(palBuffer) {
    const view = new DataView(palBuffer);
    this.palette = [];
    for (let i = 0; i < 256; i++) {
      const r = view.getUint8(i * 3);
      const g = view.getUint8(i * 3 + 1);
      const b = view.getUint8(i * 3 + 2);
      this.palette.push([r, g, b, 255]);
    }
  }

  loadSpr(sprBuffer) {
    const view = new DataView(sprBuffer);
    let offset = 32; // 跳过32字节文件头
    this.frameCount = view.getUint16(offset, true);
    offset += 2;
    this.frames = [];

    for (let f = 0; f < this.frameCount; f++) {
      const fw = view.getUint16(offset, true); offset += 2;
      const fh = view.getUint16(offset, true); offset += 2;
      const offX = view.getUint16(offset, true); offset += 2;
      const offY = view.getUint16(offset, true); offset += 2;

      const pixelLen = fw * fh;
      const pixelIdx = new Uint8Array(sprBuffer, offset, pixelLen);
      offset += pixelLen;

      const imgData = this._indexToImageData(pixelIdx, fw, fh);
      this.frames.push({ w: fw, h: fh, offX, offY, imgData });
    }
  }

  _indexToImageData(idxData, w, h) {
    const imgData = new ImageData(w, h);
    const data = imgData.data;
    for (let i = 0; i < idxData.length; i++) {
      const color = this.palette[idxData[i]];
      const p = i * 4;
      data[p] = color[0];
      data[p + 1] = color[1];
      data[p + 2] = color[2];
      data[p + 3] = idxData[i] === 0 ? 0 : color[3];
    }
    return imgData;
  }

  getFrame(index) {
    if (index < 0 || index >= this.frameCount) return null;
    return this.frames[index];
  }
}

window.SprParser = SprParser;