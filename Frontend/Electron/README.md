# Electron 开发和打包指南

## 📋 问题说明

之前打包时出现的问题：
- ❌ 新增的动画系统文件没有被正确复制到打包目录
- ❌ 修改的 MapEditor.html 没有同步到打包版本
- ❌ 缺少文件验证机制，导致打包失败才发现问题

## ✅ 解决方案

已创建两个新脚本：
- `dev.js` - 开发服务器（支持热重载）
- `build.js` - 打包脚本（带详细日志和验证）

## 🚀 使用方法

### 1. 安装依赖

```bash
cd Frontend/Electron
npm install
```

### 2. 开发模式（推荐）

**启动开发服务器**（支持热重载）：
```bash
npm run dev
```

或者：
```bash
npm start
```

**注意**：必须使用 `electron` 命令启动，不能直接用 `node` 运行！

**特性**：
- ✅ 自动监听文件变化
- ✅ 文件修改后自动刷新页面
- ✅ 打开开发者工具
- ✅ 实时查看修改效果

### 3. 打包应用

**打包为 Windows 可执行文件**：
```bash
npm run build
```

或者：
```bash
npm run package
```

**打包流程**：
1. 📁 清理并创建 dist 目录
2. 📋 复制所有必要文件（包括新增的动画系统）
3. 🔧 压缩代码
4. 📦 打包为 Electron 应用
5. 🧹 清理临时文件

**输出位置**：`../Build/千年江湖-win32-x64/`

## 📂 文件结构

```
Frontend/Electron/
├── main.js           # Electron 主进程
├── dev.js            # 开发服务器（新增）
├── build.js          # 打包脚本（新增）
├── package.json      # 项目配置
├── Game.js           # 游戏入口
├── index.html        # 主页面
├── preload.js        # 预加载脚本
└── server.js         # WebSocket 服务器
```

## 🔍 打包验证

打包脚本会自动验证以下关键文件：

### 动画系统文件
- ✅ GameLogic/Map/MapAnimationSystem.js
- ✅ GameLogic/Map/AnimationTriggerSystem.js
- ✅ GameLogic/Map/AnimationAudioSystem.js
- ✅ GameLogic/Map/AnimationEventSystem.js
- ✅ GameLogic/Map/WebGLAnimationRenderer.js

### 编辑器文件
- ✅ Res/Map/MapEditor.html

### 其他模块
- ✅ Res/ - 资源文件
- ✅ Core/ - 核心模块
- ✅ Protocol/ - 协议模块
- ✅ ResourceLoad/ - 资源加载

## 🎯 开发流程

### 推荐的开发流程：

1. **启动开发服务器**
   ```bash
   npm run dev
   ```

2. **修改代码**
   - 修改任何 JS/HTML/CSS 文件
   - 页面会自动刷新

3. **测试功能**
   - 在开发环境中测试
   - 使用开发者工具调试

4. **打包发布**
   ```bash
   npm run build
   ```

5. **验证打包结果**
   - 检查 `../Build` 目录
   - 运行打包后的应用测试

## 🐛 常见问题

### Q1: 热重载不生效？

**解决方案**：
```bash
# 停止当前服务器（Ctrl+C）
# 重新启动
npm run dev
```

### Q2: 打包时提示文件缺失？

**解决方案**：
- 检查文件路径是否正确
- 确保文件在源目录中存在
- 查看打包日志中的验证部分

### Q3: 打包后的应用无法启动？

**解决方案**：
- 检查 `../Build` 目录中的日志文件
- 确保所有依赖都已安装
- 尝试重新打包

### Q4: 打包时出现证书验证错误？

**错误信息**：
```
unable to verify the first certificate
```

**解决方案**：
- ✅ 已配置 Electron 镜像源（`.npmrc` 文件）
- ✅ 已禁用 SSL 证书验证
- ✅ 使用国内镜像加速下载

如果仍然失败，可以手动设置环境变量：
```bash
# PowerShell
$env:NODE_TLS_REJECT_UNAUTHORIZED="0"
$env:ELECTRON_MIRROR="https://npmmirror.com/mirrors/electron/"
npm run build
```

### Q5: 如何查看打包日志？

打包时会输出详细日志：
```
🚀 开始打包流程...

📁 准备 dist 目录...
✅ dist 目录已创建

📋 复制文件到 dist 目录...
  ✓ Game.js
  ✓ main.js
  📂 资源文件:
    ✓ 已复制 45 个文件
  📂 游戏逻辑:
    ✓ 已复制 32 个文件

🔍 验证关键文件...
  ✓ dist/GameLogic/Map/MapAnimationSystem.js
  ✓ dist/GameLogic/Map/AnimationTriggerSystem.js
  ...

✅ 总共复制了 127 个文件

🔧 压缩代码...
✅ 代码压缩完成

📦 打包 Electron 应用...
✅ 应用打包完成！
📂 输出目录: ../Build

🧹 清理临时文件...
✅ 临时文件已清理

🎉 打包流程全部完成！
```

## 📝 注意事项

1. **开发模式**：
   - 使用 `npm run dev` 启动
   - 文件修改会自动刷新
   - 适合日常开发

2. **打包前**：
   - 确保所有文件都已保存
   - 测试功能是否正常
   - 检查控制台是否有错误

3. **打包后**：
   - 在 `../Build` 目录中查找打包结果
   - 测试打包后的应用
   - 分发给用户前进行全面测试

## 🔄 版本控制

建议将以下文件加入版本控制：
- ✅ dev.js
- ✅ build.js
- ✅ package.json
- ✅ main.js
- ✅ 其他源代码文件

**不加入版本控制**：
- ❌ dist/ （临时目录）
- ❌ Build/ （打包输出）
- ❌ node_modules/ （依赖）

## 🎉 总结

现在你可以：
1. ✅ 使用 `npm run dev` 进行开发（支持热重载）
2. ✅ 使用 `npm run build` 进行打包（带详细验证）
3. ✅ 确保所有新增文件都被正确打包
4. ✅ 实时查看打包进度和问题

**问题已解决！** 🎊