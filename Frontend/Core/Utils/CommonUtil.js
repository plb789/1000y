/**
 * 全局通用工具类
 */
const CommonUtil = {
  // 加载图片资源
  loadImage(url) {
    return new Promise((resolve, reject) => {
      const img = new Image();
      img.onload = () => resolve(img);
      img.onerror = () => reject(new Error('图片加载失败: ' + url));
      img.src = url;
    });
  },

  // 加载二进制文件
  loadBinary(url) {
    // 如果是相对路径，确保使用正确的base URL
    if (!url.startsWith('http') && !url.startsWith('//')) {
      // 获取当前页面的origin
      const base = window.location.protocol + '//' + window.location.host;
      // 如果当前页面是file协议，使用localhost作为base
      if (window.location.protocol === 'file:') {
        url = 'http://localhost:8088' + (url.startsWith('/') ? '' : '/') + url;
      } else {
        url = base + (url.startsWith('/') ? '' : '/') + url;
      }
    }
    return fetch(url).then(res => {
      if (!res.ok) throw new Error('地图加载失败: ' + res.status);
      return res.arrayBuffer();
    });
  }
};

window.CommonUtil = CommonUtil;