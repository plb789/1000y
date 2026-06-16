const { app, BrowserWindow } = require('electron');
const path = require('path');
const http = require('http');
const fs = require('fs');

const HTTP_PORT = 8088;
let mainWindow;

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
    console.log(`HTTP Server running at http://localhost:${HTTP_PORT}/`);
  });
  
  return server;
}

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

  // 加载本地HTTP服务器上的页面
  mainWindow.loadURL(`http://localhost:${HTTP_PORT}/`);
}

app.whenReady().then(() => {
  startHttpServer();
  createWindow();
});

app.on('window-all-closed', () => app.quit());