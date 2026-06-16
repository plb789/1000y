/**
 * Electron 开发服务器
 * 支持热重载和实时预览
 */

const { app, BrowserWindow } = require('electron');
const path = require('path');
const http = require('http');
const fs = require('fs');

// 尝试加载 chokidar（如果可用）
let chokidar;
try {
  chokidar = require('chokidar');
} catch (e) {
  console.log('⚠️  chokidar 未安装，热重载功能不可用');
  console.log('💡 运行 npm install 安装依赖');
}

const HTTP_PORT = 8088;
const DEV_MODE = true; // 默认启用开发模式

let mainWindow;
let httpServer;

// MIME类型映射
const mimeTypes = {
  '.html': 'text/html',
  '.js': 'application/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.map': 'application/octet-stream',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml'
};

// 启动本地HTTP服务器
function startHttpServer() {
  const server = http.createServer((req, res) => {
    let filePath = path.join(__dirname, req.url === '/' ? 'index.html' : req.url);
    
    // 如果文件不存在，尝试从 Frontend 目录加载
    if (!fs.existsSync(filePath)) {
      filePath = path.join(__dirname, '..', req.url);
    }
    
    // 安全检查：防止路径遍历
    if (!filePath.startsWith(__dirname) && !filePath.startsWith(path.join(__dirname, '..'))) {
      res.writeHead(403);
      res.end('Forbidden');
      return;
    }
    
    const ext = path.extname(filePath).toLowerCase();
    const contentType = mimeTypes[ext] || 'application/octet-stream';
    
    fs.readFile(filePath, (err, content) => {
      if (err) {
        if (err.code === 'ENOENT') {
          res.writeHead(404);
          res.end('404 Not Found: ' + req.url);
        } else {
          res.writeHead(500);
          res.end('500 Internal Server Error');
        }
      } else {
        res.writeHead(200, { 'Content-Type': contentType });
        res.end(content);
      }
    });
  });
  
  server.listen(HTTP_PORT, () => {
    console.log(`🌐 HTTP Server running at http://localhost:${HTTP_PORT}/`);
  });
  
  return server;
}

// 创建窗口
function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: true,
      contextIsolation: false
    }
  });

  // 加载本地HTTP服务器上的页面
  mainWindow.loadURL(`http://localhost:${HTTP_PORT}/`);
  
  // 开发模式下打开开发者工具
  if (DEV_MODE) {
    mainWindow.webContents.openDevTools();
  }
  
  // 监听窗口关闭
  mainWindow.on('closed', () => {
    mainWindow = null;
  });
}

// 设置文件监听（热重载）
function setupFileWatcher() {
  if (!chokidar) {
    console.log('⚠️  文件监听功能不可用（chokidar 未安装）\n');
    return;
  }
  
  console.log('👀 启动文件监听...\n');
  
  // 监听 Frontend 目录下的所有文件变化
  const watcher = chokidar.watch([
    '../Res/**/*.html',
    '../Res/**/*.js',
    '../Res/**/*.css',
    '../Res/**/*.json',
    '../GameLogic/**/*.js',
    '../Core/**/*.js',
    '../Protocol/**/*.js',
    '../ResourceLoad/**/*.js',
    '*.js',
    '*.html'
  ], {
    ignored: /node_modules/,
    persistent: true
  });
  
  watcher.on('change', (filePath) => {
    console.log(`📝 文件已修改: ${filePath}`);
    
    // 重新加载页面
    if (mainWindow && !mainWindow.isDestroyed()) {
      console.log('🔄 重新加载页面...\n');
      mainWindow.reload();
    }
  });
  
  watcher.on('error', (error) => {
    console.error('❌ 文件监听错误:', error);
  });
}

// 应用启动
app.whenReady().then(() => {
  console.log('🚀 启动 Electron 开发服务器...\n');
  
  httpServer = startHttpServer();
  createWindow();
  
  // 开发模式下启用热重载
  if (DEV_MODE) {
    setupFileWatcher();
  }
});

// 所有窗口关闭时退出应用
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// 激活应用（macOS）
app.on('activate', () => {
  if (mainWindow === null) {
    createWindow();
  }
});

// 应用退出前清理
app.on('before-quit', () => {
  if (httpServer) {
    httpServer.close();
  }
});