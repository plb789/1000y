结合你既定的 **Electron 前端目录结构**，我把之前所有代码按目录归类、拆分文件、规范命名，同时补充引用关系、入口整合，完全匹配你的项目架构，可直接落地使用。

# 一、文件整体分配对照表
| 功能模块 | 对应目录 | 文件名 |
|--------|--------|--------|
| WebSocket 通信框架 | `Frontend/Core/Network/` | `GameWS.js` |
| 通用工具函数 | `Frontend/Core/Utils/` | `CommonUtil.js` |
| A* 寻路算法 | `Frontend/Core/Utils/` | `AStar.js` |
| 地图解析器(.map) | `Frontend/GameLogic/Map/` | `MillenniumMapParser.js` |
| 地图渲染器 | `Frontend/GameLogic/Map/` | `MapRenderer.js` |
| 地图主引擎(镜头/移动/逻辑) | `Frontend/GameLogic/Map/` | `MapEngine.js` |
| SPR帧动画解析器 | `Frontend/ResourceLoad/` | `SprParser.js` |
| SPR动画播放器 | `Frontend/ResourceLoad/` | `SpriteAnimator.js` |
| DDS贴图解析器 | `Frontend/ResourceLoad/` | `DdsParser.js` |
| 通信协议常量 | `Frontend/Protocol/` | `ProtocolDefine.js` |
| 主入口页面 | `Frontend/` | `index.html` |

---

# 二、分目录完整代码实现
## 1. Core 核心底层目录
### 1.1 `Frontend/Core/Network/GameWS.js`（WebSocket 通信）
```javascript
/**
 * WebSocket 长连接封装
 * 协议格式: [cmd(2字节)][len(2字节)][消息体][校验码(1字节)]
 */
class GameWS {
  constructor() {
    this.ws = null;
    this.url = "ws://127.0.0.1:8080/ws";
    this.isConnected = false;
    this.heartTimer = null;
    this.reconnectTimer = null;
    this.reconnectMax = 5;
    this.reconnectCount = 0;
    this.router = new Map(); // 协议路由表
  }

  connect() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
    this.ws = new WebSocket(this.url);
    this.ws.binaryType = "arraybuffer";

    this.ws.onopen = () => {
      this.isConnected = true;
      this.reconnectCount = 0;
      console.log("WebSocket 连接成功");
      this._startHeart();
    };

    this.ws.onmessage = (e) => {
      this._onRecv(e.data);
    };

    this.ws.onclose = () => {
      this.isConnected = false;
      this._stopHeart();
      this._tryReconnect();
    };

    this.ws.onerror = (err) => {
      console.error("WebSocket 异常：", err);
    };
  }

  _startHeart() {
    this._stopHeart();
    this.heartTimer = setInterval(() => {
      this.sendMsg(0x0001, new Uint8Array());
    }, 10000);
  }

  _stopHeart() {
    if (this.heartTimer) clearInterval(this.heartTimer);
  }

  _tryReconnect() {
    if (this.reconnectCount >= this.reconnectMax) return;
    this.reconnectCount++;
    console.log(`断线重连 ${this.reconnectCount}/${this.reconnectMax}`);
    this.reconnectTimer = setTimeout(() => this.connect(), 3000);
  }

  // 注册协议回调
  on(cmd, callback) {
    this.router.set(cmd, callback);
  }

  // 发送二进制消息
  sendMsg(cmd, bodyBuf) {
    if (!this.isConnected) return;
    const body = new Uint8Array(bodyBuf);
    const bodyLen = body.length;
    const totalLen = 4 + bodyLen + 1;
    const pkg = new Uint8Array(totalLen);
    const view = new DataView(pkg.buffer);

    view.setUint16(0, cmd, true);
    view.setUint16(2, bodyLen, true);
    pkg.set(body, 4);

    // 计算校验码
    let check = 0;
    for (let i = 0; i < totalLen - 1; i++) check += pkg[i];
    pkg[totalLen - 1] = check & 0xFF;

    this.ws.send(pkg.buffer);
  }

  // 解析接收数据包
  _onRecv(buffer) {
    const data = new Uint8Array(buffer);
    const totalLen = data.length;
    if (totalLen < 5) return;

    const view = new DataView(buffer);
    const cmd = view.getUint16(0, true);
    const bodyLen = view.getUint16(2, true);
    const check = data[totalLen - 1];

    // 校验码校验
    let sum = 0;
    for (let i = 0; i < totalLen - 1; i++) sum += data[i];
    if ((sum & 0xFF) !== check) {
      console.warn("数据包校验失败");
      return;
    }

    const body = data.slice(4, 4 + bodyLen);
    const cb = this.router.get(cmd);
    if (cb) cb(body);
  }

  close() {
    this._stopHeart();
    if (this.ws) this.ws.close();
  }
}

// 全局单例
window.GameWS = new GameWS();
```

### 1.2 `Frontend/Core/Utils/CommonUtil.js`（通用工具）
```javascript
/**
 * 全局通用工具类
 */
const CommonUtil = {
  // 加载图片资源
  loadImage(url) {
    return new Promise((resolve) => {
      const img = new Image();
      img.onload = () => resolve(img);
      img.src = url;
    });
  },

  // 加载二进制文件
  loadBinary(url) {
    return fetch(url).then(res => res.arrayBuffer());
  }
};

window.CommonUtil = CommonUtil;
```

### 1.3 `Frontend/Core/Utils/AStar.js`（A* 寻路）
```javascript
/**
 * A* 八方向寻路算法
 */
const AStar = (() => {
  const dirs = [
    [-1, 0], [1, 0], [0, -1], [0, 1],
    [-1, -1], [-1, 1], [1, -1], [1, 1]
  ];

  class Node {
    constructor(x, y) {
      this.x = x;
      this.y = y;
      this.g = 0;
      this.h = 0;
      this.f = 0;
      this.parent = null;
    }
  }

  function heuristic(x1, y1, x2, y2) {
    return Math.abs(x1 - x2) + Math.abs(y1 - y2);
  }

  function findPath(startX, startY, endX, endY, collision, mapW, mapH) {
    if (collision[startY * mapW + startX] === 1) return [];
    if (collision[endY * mapW + endX] === 1) return [];
    if (startX === endX && startY === endY) return [];

    const openList = [];
    const closeSet = new Set();
    const startNode = new Node(startX, startY);
    const endNode = new Node(endX, endY);
    openList.push(startNode);

    while (openList.length > 0) {
      let currIdx = 0;
      for (let i = 0; i < openList.length; i++) {
        if (openList[i].f < openList[currIdx].f) currIdx = i;
      }
      const curr = openList[currIdx];
      openList.splice(currIdx, 1);
      closeSet.add(`${curr.x},${curr.y}`);

      if (curr.x === endNode.x && curr.y === endNode.y) {
        const path = [];
        let temp = curr;
        while (temp) {
          path.unshift({ x: temp.x, y: temp.y });
          temp = temp.parent;
        }
        return path;
      }

      for (const [dx, dy] of dirs) {
        const nx = curr.x + dx;
        const ny = curr.y + dy;
        const key = `${nx},${ny}`;

        if (nx < 0 || ny < 0 || nx >= mapW || ny >= mapH) continue;
        if (collision[ny * mapW + nx] === 1) continue;
        if (closeSet.has(key)) continue;

        const neighbor = new Node(nx, ny);
        const g = curr.g + 1;
        let exist = openList.find(n => n.x === nx && n.y === ny);

        if (!exist) {
          neighbor.g = g;
          neighbor.h = heuristic(nx, ny, endNode.x, endNode.y);
          neighbor.f = neighbor.g + neighbor.h;
          neighbor.parent = curr;
          openList.push(neighbor);
        } else if (g < exist.g) {
          exist.g = g;
          exist.f = exist.g + exist.h;
          exist.parent = curr;
        }
      }
    }
    return [];
  }

  return { findPath };
})();

window.AStar = AStar;
```

---

## 2. GameLogic/Map 地图模块
### 2.1 `Frontend/GameLogic/Map/MillenniumMapParser.js`（.map 解析）
```javascript
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
```

### 2.2 `Frontend/GameLogic/Map/MapRenderer.js`（地图渲染）
```javascript
/**
 * 地图渲染器：底层+高层瓦片绘制
 */
class MapRenderer {
  constructor(canvas, tileSetImage) {
    this.ctx = canvas.getContext('2d');
    this.tileSet = tileSetImage;
    this.tileSize = 32;
    this.tileCol = 16; // 瓦片图集一行16个
  }

  renderMap(map) {
    const w = map.width;
    const h = map.height;
    for (let y = 0; y < h; y++) {
      for (let x = 0; x < w; x++) {
        const tile = map.getTile(x, y);
        if (!tile) continue;
        this.drawTile(x, y, tile.low);
        if (tile.high > 0) {
          this.drawTile(x, y, tile.high);
        }
      }
    }
  }

  drawTile(x, y, tileIndex) {
    const ts = this.tileSize;
    const sx = (tileIndex % this.tileCol) * ts;
    const sy = Math.floor(tileIndex / this.tileCol) * ts;
    this.ctx.drawImage(
      this.tileSet,
      sx, sy, ts, ts,
      x * ts, y * ts, ts, ts
    );
  }
}

window.MapRenderer = MapRenderer;
```

### 2.3 `Frontend/GameLogic/Map/MapEngine.js`（地图主引擎）
```javascript
/**
 * 地图总引擎：镜头、角色、移动、寻路、事件
 */
class MapEngine {
  constructor(canvas) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.canvas.width = 1000;
    this.canvas.height = 600;
    this.tileSize = 32;

    // 地图核心
    this.mapParser = new MillenniumMapParser();
    this.mapRenderer = null;

    // 镜头
    this.camera = {
      offsetX: 0,
      offsetY: 0,
      dragStartX: 0,
      dragStartY: 0,
      isDrag: false
    };

    // 玩家数据
    this.player = {
      x: 10,
      y: 10,
      pixelX: 0,
      pixelY: 0,
      speed: 2,
      movePath: []
    };

    // 资源
    this.tilesetImg = null;
    this.roleAnim = null;

    this.bindEvent();
    this.loop();
  }

  async loadMap(mapUrl, tilesetUrl) {
    const mapBuf = await CommonUtil.loadBinary(mapUrl);
    this.mapParser.loadMap(mapBuf);

    this.tilesetImg = await CommonUtil.loadImage(tilesetUrl);
    this.mapRenderer = new MapRenderer(this.canvas, this.tilesetImg);

    this.syncPlayerPixel();
    this.followPlayer();
  }

  tile2Pixel(tileX, tileY) {
    return {
      x: tileX * this.tileSize,
      y: tileY * this.tileSize
    };
  }

  pixel2Tile(px, py) {
    return {
      x: Math.floor(px / this.tileSize),
      y: Math.floor(py / this.tileSize)
    };
  }

  syncPlayerPixel() {
    const pos = this.tile2Pixel(this.player.x, this.player.y);
    this.player.pixelX = pos.x;
    this.player.pixelY = pos.y;
  }

  followPlayer() {
    const mapW = this.mapParser.width * this.tileSize;
    const mapH = this.mapParser.height * this.tileSize;
    const canvasW = this.canvas.width;
    const canvasH = this.canvas.height;

    this.camera.offsetX = this.player.pixelX - canvasW / 2 + this.tileSize / 2;
    this.camera.offsetY = this.player.pixelY - canvasH / 2 + this.tileSize / 2;

    this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
    this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
  }

  bindEvent() {
    // 镜头拖拽
    this.canvas.addEventListener('mousedown', (e) => {
      this.camera.isDrag = true;
      this.camera.dragStartX = e.clientX - this.camera.offsetX;
      this.camera.dragStartY = e.clientY - this.camera.offsetY;
    });

    this.canvas.addEventListener('mousemove', (e) => {
      if (!this.camera.isDrag) return;
      const mapW = this.mapParser.width * this.tileSize;
      const mapH = this.mapParser.height * this.tileSize;
      const canvasW = this.canvas.width;
      const canvasH = this.canvas.height;

      this.camera.offsetX = e.clientX - this.camera.dragStartX;
      this.camera.offsetY = e.clientY - this.camera.dragStartY;

      this.camera.offsetX = Math.max(0, Math.min(this.camera.offsetX, mapW - canvasW));
      this.camera.offsetY = Math.max(0, Math.min(this.camera.offsetY, mapH - canvasH));
    });

    this.canvas.addEventListener('mouseup', () => {
      this.camera.isDrag = false;
    });

    // 点击寻路
    this.canvas.addEventListener('click', (e) => {
      if (this.camera.isDrag) return;
      const rect = this.canvas.getBoundingClientRect();
      const clickPx = e.clientX - rect.left + this.camera.offsetX;
      const clickPy = e.clientY - rect.top + this.camera.offsetY;
      const targetTile = this.pixel2Tile(clickPx, clickPy);

      const tile = this.mapParser.getTile(targetTile.x, targetTile.y);
      if (!tile || tile.attr === 1) return;

      this.player.movePath = AStar.findPath(
        this.player.x, this.player.y,
        targetTile.x, targetTile.y,
        this.mapParser.collision,
        this.mapParser.width,
        this.mapParser.height
      );
    });
  }

  updatePlayerMove() {
    const path = this.player.movePath;
    if (path.length === 0) return;

    const next = path[0];
    const targetPos = this.tile2Pixel(next.x, next.y);
    const dx = targetPos.x - this.player.pixelX;
    const dy = targetPos.y - this.player.pixelY;
    const dist = Math.hypot(dx, dy);

    if (dist < this.player.speed) {
      this.player.pixelX = targetPos.x;
      this.player.pixelY = targetPos.y;
      this.player.x = next.x;
      this.player.y = next.y;
      path.shift();
      // 检测传送/事件区域
      this.checkEventArea();
    } else {
      this.player.pixelX += (dx / dist) * this.player.speed;
      this.player.pixelY += (dy / dist) * this.player.speed;
    }
    this.followPlayer();
  }

  checkEventArea() {
    const tile = this.mapParser.getTile(this.player.x, this.player.y);
    if (!tile) return;
    if (tile.attr === 2) {
      console.log("进入传送区域");
    }
  }

  render() {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    if (!this.mapRenderer) return;

    this.ctx.save();
    this.ctx.translate(-this.camera.offsetX, -this.camera.offsetY);

    this.mapRenderer.renderMap(this.mapParser);

    // 绘制角色动画
    if (this.roleAnim) {
      this.roleAnim.draw(this.ctx, this.player.pixelX, this.player.pixelY);
    }

    this.ctx.restore();
  }

  loop() {
    this.updatePlayerMove();
    if (this.roleAnim) this.roleAnim.update();
    this.render();
    requestAnimationFrame(() => this.loop());
  }
}

window.MapEngine = MapEngine;
```

---

## 3. ResourceLoad 资源加载目录
### 3.1 `Frontend/ResourceLoad/SprParser.js`（SPR 帧动画解析）
```javascript
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
```

### 3.2 `Frontend/ResourceLoad/SpriteAnimator.js`（动画播放器）
```javascript
/**
 * SPR 动画播放控制器
 */
class SpriteAnimator {
  constructor(sprParser) {
    this.parser = sprParser;
    this.curFrame = 0;
    this.frameRate = 10;
    this.frameTimer = 0;
    this.isLoop = true;
    this.actionStart = 0;
    this.actionEnd = 0;
  }

  setAction(start, end, loop = true) {
    this.actionStart = start;
    this.actionEnd = end;
    this.curFrame = start;
    this.isLoop = loop;
    this.frameTimer = 0;
  }

  update(deltaTime = 16) {
    this.frameTimer += deltaTime;
    const interval = 1000 / this.frameRate;
    if (this.frameTimer >= interval) {
      this.frameTimer = 0;
      this.curFrame++;
      if (this.curFrame > this.actionEnd) {
        this.curFrame = this.isLoop ? this.actionStart : this.actionEnd;
      }
    }
  }

  draw(ctx, x, y) {
    const frame = this.parser.getFrame(this.curFrame);
    if (!frame) return;
    ctx.putImageData(frame.imgData, x - frame.offX, y - frame.offY);
  }
}

window.SpriteAnimator = SpriteAnimator;
```

### 3.3 `Frontend/ResourceLoad/DdsParser.js`（DDS 贴图解析）
```javascript
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
```

---

## 4. Protocol 通信协议目录
### 4.1 `Frontend/Protocol/ProtocolDefine.js`（协议常量）
```javascript
/**
 * 全局通信协议号定义
 */
const ProtocolDefine = {
  CMD_HEART: 0x0001,    // 心跳包
  CMD_MOVE:  0x0002,    // 角色移动
  CMD_CHAT:  0x0003     // 聊天消息
};

window.ProtocolDefine = ProtocolDefine;
```

---

## 5. 入口页面 `Frontend/index.html`（统一引入+初始化）
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>千年江湖 - 客户端</title>
  <style>
    body { margin: 0; background: #000; display: flex; justify-content: center; align-items: center; height: 100vh; }
    canvas { border: 2px solid #666; }
  </style>
</head>
<body>
  <canvas id="gameCanvas"></canvas>

  <!-- 1. 工具类 -->
  <script src="Core/Utils/CommonUtil.js"></script>
  <script src="Core/Utils/AStar.js"></script>

  <!-- 2. 网络通信 -->
  <script src="Core/Network/GameWS.js"></script>

  <!-- 3. 通信协议 -->
  <script src="Protocol/ProtocolDefine.js"></script>

  <!-- 4. 资源解析 -->
  <script src="ResourceLoad/SprParser.js"></script>
  <script src="ResourceLoad/SpriteAnimator.js"></script>
  <script src="ResourceLoad/DdsParser.js"></script>

  <!-- 5. 地图模块 -->
  <script src="GameLogic/Map/MillenniumMapParser.js"></script>
  <script src="GameLogic/Map/MapRenderer.js"></script>
  <script src="GameLogic/Map/MapEngine.js"></script>

  <!-- 主逻辑初始化 -->
  <script>
    window.onload = async () => {
      // 1. 连接 WebSocket
      GameWS.connect();

      // 2. 初始化地图引擎
      const canvas = document.getElementById("gameCanvas");
      const gameMap = new MapEngine(canvas);

      // 加载地图与瓦片集（路径根据实际资源修改）
      await gameMap.loadMap("/Res/Map/001.map", "/Res/Texture/tileset.png");

      // 3. 加载角色SPR动画（示例）
      const sprParser = new SprParser();
      const palBuf = await CommonUtil.loadBinary("/Res/Sprite/char.pal");
      const sprBuf = await CommonUtil.loadBinary("/Res/Sprite/role.spr");
      sprParser.loadPalette(palBuf);
      sprParser.loadSpr(sprBuf);

      const roleAnim = new SpriteAnimator(sprParser);
      roleAnim.setAction(0, 15, true);
      gameMap.roleAnim = roleAnim;
    };
  </script>
</body>
</html>
```

---

# 三、Electron 补充配置（简要）
## `Frontend/Electron/main.js`
```javascript
const { app, BrowserWindow } = require('electron');
const path = require('path');

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: true,
      contextIsolation: false
    }
  });

  mainWindow.loadFile(path.join(__dirname, '../index.html'));
  // 开发调试
  mainWindow.webContents.openDevTools();
}

app.whenReady().then(createWindow);
app.on('window-all-closed', () => app.quit());
```

## `Frontend/Electron/package.json`
```json
{
  "name": "qiannian-client",
  "version": "1.0.0",
  "main": "main.js",
  "scripts": {
    "start": "electron .",
    "package": "electron-packager . 千年江湖 --win --out=../Build"
  },
  "dependencies": {},
  "devDependencies": {
    "electron": "^28.0.0",
    "electron-packager": "^17.1.2"
  }
}
```

---

# 四、使用说明
1. 严格按照上面目录存放所有 JS 文件，**引入顺序不可打乱**；
2. 将 `.map`/`.spr`/`.pal`/`.dds` 资源放入项目 `Res` 对应子目录；
3. 先启动 Go 网关 WebSocket 服务，再运行 Electron 客户端；
4. 鼠标**左键拖拽镜头**，**点击地面**自动寻路移动；
5. 所有类均挂载到 `window` 全局，模块之间可互相调用。

目前整套代码已完整落地：**地图解析+渲染+镜头+寻路+角色动画+DDS贴图+WebSocket通信**，完全匹配你既定的文件夹架构。