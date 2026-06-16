/**
 * Electron 打包脚本
 * 确保所有文件正确复制到 dist 目录
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('🚀 开始打包流程...\n');

// 颜色输出
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  blue: '\x1b[34m'
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

// 清理并创建 dist 目录
function prepareDist() {
  log('📁 准备 dist 目录...', 'blue');
  
  if (fs.existsSync('dist')) {
    fs.rmSync('dist', { recursive: true, force: true });
  }
  
  fs.mkdirSync('dist', { recursive: true });
  log('✅ dist 目录已创建\n', 'green');
}

// 递归复制目录
function copyDir(src, dst, verbose = false) {
  if (!fs.existsSync(src)) {
    log(`⚠️  源目录不存在: ${src}`, 'yellow');
    return;
  }
  
  if (!fs.existsSync(dst)) {
    fs.mkdirSync(dst, { recursive: true });
  }
  
  const files = fs.readdirSync(src);
  let copiedCount = 0;
  
  files.forEach(file => {
    const srcPath = path.join(src, file);
    const dstPath = path.join(dst, file);
    const stat = fs.statSync(srcPath);
    
    if (stat.isDirectory()) {
      const subCount = copyDir(srcPath, dstPath, verbose);
      copiedCount += subCount;
    } else {
      fs.copyFileSync(srcPath, dstPath);
      copiedCount++;
      if (verbose) {
        log(`  ✓ ${file}`, 'green');
      }
    }
  });
  
  return copiedCount;
}

// 复制单个文件
function copyFile(src, dst) {
  if (!fs.existsSync(src)) {
    log(`⚠️  文件不存在: ${src}`, 'yellow');
    return false;
  }
  
  fs.copyFileSync(src, dst);
  return true;
}

// 复制所有必要文件
function copyAllFiles() {
  log('📋 复制文件到 dist 目录...\n', 'blue');
  
  let totalCopied = 0;
  
  // 复制根目录文件
  const rootFiles = ['Game.js', 'main.js', 'preload.js', 'server.js', 'index.html', 'package.json'];
  rootFiles.forEach(file => {
    if (copyFile(file, path.join('dist', file))) {
      log(`  ✓ ${file}`, 'green');
      totalCopied++;
    }
  });
  
  // 修改 dist/package.json，将 main 改为 main.js（生产环境）
  const distPackageJson = path.join('dist', 'package.json');
  if (fs.existsSync(distPackageJson)) {
    const packageData = JSON.parse(fs.readFileSync(distPackageJson, 'utf8'));
    packageData.main = 'main.js'; // 生产环境使用 main.js
    fs.writeFileSync(distPackageJson, JSON.stringify(packageData, null, 2));
    log('  ✓ 已修改 package.json (main: main.js)', 'green');
  }
  
  log('', 'reset');
  
  // 复制目录
  const dirs = [
    { src: '../Res', dst: 'dist/Res', name: '资源文件' },
    { src: '../Core', dst: 'dist/Core', name: '核心模块' },
    { src: '../GameLogic', dst: 'dist/GameLogic', name: '游戏逻辑' },
    { src: '../Protocol', dst: 'dist/Protocol', name: '协议模块' },
    { src: '../ResourceLoad', dst: 'dist/ResourceLoad', name: '资源加载' }
  ];
  
  dirs.forEach(dir => {
    log(`  📂 ${dir.name}:`, 'blue');
    const count = copyDir(dir.src, dir.dst);
    log(`    ✓ 已复制 ${count} 个文件\n`, 'green');
    totalCopied += count;
  });
  
  // 验证关键文件
  log('🔍 验证关键文件...', 'blue');
  const criticalFiles = [
    'dist/GameLogic/Map/MapAnimationSystem.js',
    'dist/GameLogic/Map/AnimationTriggerSystem.js',
    'dist/GameLogic/Map/AnimationAudioSystem.js',
    'dist/GameLogic/Map/AnimationEventSystem.js',
    'dist/GameLogic/Map/WebGLAnimationRenderer.js',
    'dist/Res/Map/MapEditor.html'
  ];
  
  criticalFiles.forEach(file => {
    if (fs.existsSync(file)) {
      log(`  ✓ ${file}`, 'green');
    } else {
      log(`  ✗ ${file} - 缺失！`, 'red');
    }
  });
  
  log(`\n✅ 总共复制了 ${totalCopied} 个文件\n`, 'green');
}

// 压缩代码
function minifyCode() {
  log('🔧 压缩代码...', 'blue');
  
  try {
    execSync('npm run minify', { stdio: 'inherit' });
    log('✅ 代码压缩完成\n', 'green');
  } catch (error) {
    log('⚠️  代码压缩失败，继续打包...\n', 'yellow');
  }
}

// 打包应用
function packageApp() {
  log('📦 打包 Electron 应用...', 'blue');
  
  try {
    // 设置环境变量禁用证书验证（解决网络问题）
    const env = {
      ...process.env,
      NODE_TLS_REJECT_UNAUTHORIZED: '0',
      ELECTRON_MIRROR: 'https://npmmirror.com/mirrors/electron/'
    };
    
    const packager = require('electron-packager');
    
    packager({
      dir: 'dist',
      name: '千年江湖',
      platform: 'win32',
      arch: 'x64',
      out: '../Build',
      asar: true,
      overwrite: true
    }).then(() => {
      log('\n✅ 应用打包完成！', 'green');
      log('📂 输出目录: ../Build\n', 'blue');
      cleanup();
    }).catch((error) => {
      log('\n❌ 打包失败:', 'red');
      log(error.message, 'red');
      process.exit(1);
    });
    
  } catch (error) {
    log('\n❌ 打包失败:', 'red');
    log(error.message, 'red');
    process.exit(1);
  }
}

// 清理临时文件
function cleanup() {
  log('🧹 清理临时文件...', 'blue');
  
  if (fs.existsSync('dist')) {
    fs.rmSync('dist', { recursive: true, force: true });
    log('✅ 临时文件已清理\n', 'green');
  }
}

// 主流程
async function main() {
  try {
    prepareDist();
    copyAllFiles();
    minifyCode();
    await packageApp(); // 等待打包完成
    
    log('🎉 打包流程全部完成！', 'green');
    log('💡 提示: 请检查 ../Build 目录中的打包结果\n', 'blue');
    
  } catch (error) {
    log('\n❌ 打包过程中出现错误:', 'red');
    log(error.message, 'red');
    console.error(error);
    process.exit(1);
  }
}

// 运行
main();