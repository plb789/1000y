/**
 * 资源管理器：优化资源加载性能，支持缓存和异步加载
 */
class ResourceManager {
  constructor() {
    this.cache = new Map();
    this.loading = new Map();
    this.maxCacheSize = 100; // 最大缓存数量
    this.cacheOrder = []; // LRU顺序
  }

  /**
   * 加载图片资源（支持PNG/JPG/WebP）
   */
  async loadImage(url) {
    // 检查缓存
    if (this.cache.has(url)) {
      this.updateCacheOrder(url);
      return this.cache.get(url);
    }

    // 检查是否正在加载
    if (this.loading.has(url)) {
      return this.loading.get(url);
    }

    const promise = new Promise((resolve, reject) => {
      const img = new Image();
      img.crossOrigin = 'anonymous';
      
      img.onload = () => {
        this.cacheResource(url, img);
        this.loading.delete(url);
        resolve(img);
      };
      
      img.onerror = (err) => {
        this.loading.delete(url);
        reject(err);
      };
      
      img.src = url;
    });

    this.loading.set(url, promise);
    return promise;
  }

  /**
   * 加载二进制文件
   */
  async loadBinary(url) {
    if (this.cache.has(url)) {
      this.updateCacheOrder(url);
      return this.cache.get(url);
    }

    if (this.loading.has(url)) {
      return this.loading.get(url);
    }

    const promise = fetch(url)
      .then(response => response.arrayBuffer())
      .then(buffer => {
        this.cacheResource(url, buffer);
        this.loading.delete(url);
        return buffer;
      })
      .catch(err => {
        this.loading.delete(url);
        throw err;
      });

    this.loading.set(url, promise);
    return promise;
  }

  /**
   * 加载DDS贴图
   */
  async loadDDS(url) {
    const buffer = await this.loadBinary(url);
    const parser = new DdsParser();
    const imageData = await parser.load(buffer);
    
    // 转换为Image对象
    const canvas = document.createElement('canvas');
    canvas.width = parser.width;
    canvas.height = parser.height;
    const ctx = canvas.getContext('2d');
    ctx.putImageData(imageData, 0, 0);
    
    return canvas;
  }

  /**
   * 缓存资源
   */
  cacheResource(key, resource) {
    // LRU策略：如果缓存已满，移除最久未使用的
    if (this.cache.size >= this.maxCacheSize) {
      const oldestKey = this.cacheOrder.shift();
      this.cache.delete(oldestKey);
    }

    this.cache.set(key, resource);
    this.cacheOrder.push(key);
  }

  /**
   * 更新缓存顺序（标记为最近使用）
   */
  updateCacheOrder(key) {
    const index = this.cacheOrder.indexOf(key);
    if (index !== -1) {
      this.cacheOrder.splice(index, 1);
      this.cacheOrder.push(key);
    }
  }

  /**
   * 预加载资源列表
   */
  async preload(resources) {
    const promises = resources.map(res => {
      if (res.type === 'image' || res.type === 'tileset') {
        return this.loadImage(res.url);
      } else if (res.type === 'binary') {
        return this.loadBinary(res.url);
      } else if (res.type === 'dds') {
        return this.loadDDS(res.url);
      }
      return Promise.resolve(null);
    });

    return Promise.all(promises);
  }

  /**
   * 清空缓存
   */
  clearCache() {
    this.cache.clear();
    this.cacheOrder = [];
  }

  /**
   * 获取缓存统计
   */
  getStats() {
    return {
      cached: this.cache.size,
      loading: this.loading.size,
      maxSize: this.maxCacheSize
    };
  }
}

// 创建全局单例
window.ResourceManager = new ResourceManager();
