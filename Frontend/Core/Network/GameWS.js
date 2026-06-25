/**
 * WebSocket 长连接封装
 * 协议格式: [cmd(2字节)][body(N字节)]
 */

// 二进制协议集合：这些cmd的body为紧凑二进制格式，不走JSON解析
// 怪物位置更新(3101): [map_id:4B][count:2B][timestamp:8B] + 每怪物[instance_id:4B][x:2B][y:2B][state:1B][hp:4B]
const BINARY_PROTOCOLS = new Set([
  3101 // CMD_MONSTER_POSITION_UPDATE
]);

class GameWS {
  constructor() {
    this.ws = null;
    this.url = "ws://127.0.0.1:8080/ws";
    this.isConnected = false;
    this.heartTimer = null;
    this.reconnectTimer = null;
    this.reconnectMax = 10;
    this.reconnectCount = 0;
    this.router = new Map(); // 协议路由表
    this.pendingRequests = new Map(); // 待响应的请求 (cmd -> {resolve, reject, timer})
    this._msgId = 0; // 消息ID生成器
    
    // 事件回调
    this.callbacks = {
      onOpen: null,
      onClose: null,
      onError: null,
      onMessage: null
    };
  }

  /**
   * 连接服务器
   * @param {string} url - WebSocket地址
   * @param {object} callbacks - 回调函数 {onOpen, onClose, onError, onMessage}
   */
  connect(url, callbacks = {}) {
    if (url) this.url = url;
    if (callbacks) {
      this.callbacks.onOpen = callbacks.onOpen || null;
      this.callbacks.onClose = callbacks.onClose || null;
      this.callbacks.onError = callbacks.onError || null;
      this.callbacks.onMessage = callbacks.onMessage || null;
    }
    
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
    
    try {
      this.ws = new WebSocket(this.url);
      this.ws.binaryType = "arraybuffer";

      this.ws.onopen = () => {
        this.isConnected = true;
        this.reconnectCount = 0;
        console.log("WebSocket 连接成功");
        this._startHeart();
        if (this.callbacks.onOpen) this.callbacks.onOpen();
      };

      this.ws.onmessage = (e) => {
        this._onRecv(e.data);
      };

      this.ws.onclose = () => {
        this.isConnected = false;
        this._stopHeart();
        this._tryReconnect();
        if (this.callbacks.onClose) this.callbacks.onClose();
      };

      this.ws.onerror = (err) => {
        console.error("WebSocket 异常：", err);
        if (this.callbacks.onError) this.callbacks.onError(err);
      };
    } catch (err) {
      console.error('WebSocket创建失败:', err);
    }
  }

  _startHeart() {
    this._stopHeart();
    this.heartTimer = setInterval(() => {
      // 使用心跳协议
      this.send(1003, {});
    }, 30000);
  }

  _stopHeart() {
    if (this.heartTimer) {
      clearInterval(this.heartTimer);
      this.heartTimer = null;
    }
  }

  _tryReconnect() {
    if (this.reconnectCount >= this.reconnectMax) {
      console.log('已达到最大重连次数');
      return;
    }
    this.reconnectCount++;
    console.log(`断线重连 ${this.reconnectCount}/${this.reconnectMax}`);
    this.reconnectTimer = setTimeout(() => this.connect(), 3000);
  }

  /**
   * 注册协议回调
   * @param {number} cmd - 协议号
   * @param {function} callback - 回调函数
   */
  on(cmd, callback) {
    this.router.set(cmd, callback);
  }

  /**
   * 发送JSON格式消息
   * @param {number} cmd - 协议号
   * @param {object} data - 数据对象
   */
  send(cmd, data) {
    if (!this.isConnected) {
      console.warn('WebSocket未连接');
      return;
    }
    
    // 确保cmd是有效数字
    if (typeof cmd !== 'number' || !Number.isFinite(cmd)) {
      console.error('无效的cmd:', cmd);
      return;
    }
    
    try {
      const jsonStr = JSON.stringify(data);
      const bodyBuf = new TextEncoder().encode(jsonStr);
      this.sendMsg(cmd, bodyBuf);
    } catch (err) {
      console.error('发送消息失败:', err);
    }
  }

  /**
   * 发送请求并等待响应（RPC模式）
   * @param {number} cmd - 协议号
   * @param {object} data - 请求数据
   * @param {number} timeout - 超时时间(ms)，默认5000
   * @returns {Promise} 响应数据
   */
  request(cmd, data, timeout = 5000) {
    return new Promise((resolve, reject) => {
      if (!this.isConnected) {
        reject(new Error('WebSocket未连接'));
        return;
      }

      // 生成唯一消息ID
      const msgId = ++this._msgId;
      const requestData = { ...data, msg_id: msgId };

      // 设置超时
      const timer = setTimeout(() => {
        if (this.pendingRequests.has(msgId)) {
          this.pendingRequests.delete(msgId);
          reject(new Error('请求超时'));
        }
      }, timeout);

      // 存储待响应的请求
      this.pendingRequests.set(msgId, { resolve, reject, timer });

      // 发送请求
      this.send(cmd, requestData);
    });
  }

  /**
   * 发送二进制消息
   * @param {number} cmd - 协议号
   * @param {Uint8Array} bodyBuf - 二进制数据
   */
  sendMsg(cmd, bodyBuf) {
    if (!this.isConnected || !this.ws) return;
    
    // 调试日志
    console.log('sendMsg called:', {cmd: cmd, bodyLen: bodyBuf ? bodyBuf.length : 'null'});
    
    // 确保bodyBuf是有效的Uint8Array
    if (!(bodyBuf instanceof Uint8Array)) {
      console.error('bodyBuf不是Uint8Array:', typeof bodyBuf);
      bodyBuf = new Uint8Array(0);
    }
    
    const body = new Uint8Array(bodyBuf);
    const bodyLen = body.length;
    const totalLen = 5 + bodyLen;
    
    console.log('preparing packet:', {bodyLen: bodyLen, totalLen: totalLen});
    
    // 简化协议: [cmd(2字节)][body(N字节)]
    const pkg = new Uint8Array(2 + bodyLen);
    const view = new DataView(pkg.buffer);

    // 命令 (小端序)
    view.setUint16(0, cmd, true);
    // 数据
    if (bodyLen > 0) {
      pkg.set(body, 2);
    }

    this.ws.send(pkg.buffer);
  }

  /**
   * 解析接收数据包
   * 简化协议: [cmd(2字节)][body(N字节)]
   * 二进制协议cmd集合: body为紧凑二进制，不走JSON解析
   */
  _onRecv(buffer) {
    try {
      const data = new Uint8Array(buffer);
      const totalLen = data.length;
      if (totalLen < 2) return;

      const view = new DataView(buffer);
      // 命令 (小端序)
      const cmd = view.getUint16(0, true);
      // 数据
      const bodyLen = totalLen - 2;
      const bodyData = data.slice(2, 2 + bodyLen);
      
      let body = {};
      if (bodyLen > 0) {
        // 二进制协议集合：直接传Uint8Array，不做JSON解析
        if (BINARY_PROTOCOLS.has(cmd)) {
          body = bodyData;
        } else {
          const jsonStr = new TextDecoder().decode(bodyData);
          try {
            body = JSON.parse(jsonStr);
          } catch (e) {
            console.warn('JSON解析失败:', jsonStr);
            body = { raw: jsonStr };
          }
        }
      }

      // 检查是否有待响应的请求（RPC模式）
      if (body.msg_id && this.pendingRequests.has(body.msg_id)) {
        const pending = this.pendingRequests.get(body.msg_id);
        clearTimeout(pending.timer);
        this.pendingRequests.delete(body.msg_id);
        if (body.code === 200 || body.code === 0) {
          pending.resolve(body);
        } else {
          pending.reject(new Error(body.msg || '请求失败'));
        }
        return; // 不再触发普通路由回调
      }

      // 触发路由回调
      const cb = this.router.get(cmd);
      if (cb) {
        cb(body);
      }

      // 触发全局消息回调
      if (this.callbacks.onMessage) {
        this.callbacks.onMessage(cmd, body);
      }
    } catch (err) {
      console.error('消息处理错误:', err);
    }
  }

  /**
   * 关闭连接
   */
  close() {
    this._stopHeart();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
  }
  
  /**
   * 获取连接状态
   */
  getConnected() {
    return this.isConnected;
  }
}

// 全局单例
if (!window.GameWS) {
  window.GameWS = new GameWS();
}
