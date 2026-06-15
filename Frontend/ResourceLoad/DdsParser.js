/**
 * DDS DXT1 贴图解析器
 */
class DdsParser {
  constructor() {
    this.width = 0;
    this.height = 0;
    this.imageData = null;
  }

  async load(ddsBuffer) {
    const view = new DataView(ddsBuffer);
    let offset = 0;
    const magic = String.fromCharCode(
      view.getUint8(offset), view.getUint8(offset+1),
      view.getUint8(offset+2), view.getUint8(offset+3)
    );
    if (magic !== "DDS ") throw new Error("非标准DDS文件");
    offset += 4;
    offset += 124;

    this.height = view.getUint32(offset, true); offset +=4;
    this.width = view.getUint32(offset, true); offset +=4;
    offset += 8;
    offset += 44;

    const fourCC = String.fromCharCode(
      view.getUint8(offset), view.getUint8(offset+1),
      view.getUint8(offset+2), view.getUint8(offset+3)
    );
    offset += 4;
    const dataBuf = new Uint8Array(ddsBuffer, offset);

    if (fourCC === "DXT1") {
      this.imageData = this._decodeDXT1(dataBuf, this.width, this.height);
    } else {
      throw new Error(`不支持格式:${fourCC}`);
    }
    return this.imageData;
  }

  _decodeDXT1(data, w, h) {
    const img = new ImageData(w, h);
    const out = img.data;
    const blockW = Math.ceil(w / 4);
    const blockH = Math.ceil(h / 4);

    for (let by = 0; by < blockH; by++) {
      for (let bx = 0; bx < blockW; bx++) {
        const bOff = (by * blockW + bx) * 8;
        const c0 = data[bOff] | (data[bOff+1] << 8);
        const c1 = data[bOff+2] | (data[bOff+3] << 8);
        const idx = [data[bOff+4], data[bOff+5], data[bOff+6], data[bOff+7]];
        const col = this._unpackDxtColor(c0, c1);
        const useAlpha = c0 <= c1;

        for (let y = 0; y < 4; y++) {
          for (let x = 0; x < 4; x++) {
            const px = bx * 4 + x;
            const py = by * 4 + y;
            if (px >= w || py >= h) continue;
            const i = (idx[y] >> (x * 2)) & 3;
            const p = (py * w + px) * 4;
            if (useAlpha && i === 3) {
              out[p+3] = 0;
            } else {
              out[p] = col[i][0];
              out[p+1] = col[i][1];
              out[p+2] = col[i][2];
              out[p+3] = 255;
            }
          }
        }
      }
    }
    return img;
  }

  _unpackDxtColor(c0, c1) {
    const r0 = (c0 >> 11) & 0x1F;
    const g0 = (c0 >> 5) & 0x3F;
    const b0 = c0 & 0x1F;
    const r1 = (c1 >> 11) & 0x1F;
    const g1 = (c1 >> 5) & 0x3F;
    const b1 = c1 & 0x1F;

    const col = [
      [r0 << 3, g0 << 2, b0 << 3],
      [r1 << 3, g1 << 2, b1 << 3]
    ];
    col[2] = [Math.round((col[0][0]*2+col[1][0])/3), Math.round((col[0][1]*2+col[1][1])/3), Math.round((col[0][2]*2+col[1][2])/3)];
    col[3] = [Math.round((col[0][0]+col[1][0]*2)/3), Math.round((col[0][1]+col[1][1]*2)/3), Math.round((col[0][2]+col[1][2]*2)/3)];
    return col;
  }
}

window.DdsParser = DdsParser;