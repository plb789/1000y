/**
 * 地图动画系统
 * 支持火焰、水流、树木摆动等动态效果
 */
class MapAnimationSystem {
  constructor() {
    this.animations = new Map(); // 存储所有动画实例
    this.animationTypes = new Map(); // 动画类型定义
    this.isRunning = false;
    this.lastTime = 0;
    
    // 注册内置动画类型
    this.registerAnimationTypes();
  }
  
  /**
   * 注册内置动画类型
   */
  registerAnimationTypes() {
    // 火焰动画
    this.animationTypes.set('fire', {
      name: '火焰',
      frameCount: 4,
      duration: 400, // 毫秒
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 火焰效果：通过颜色变化和抖动实现
        const colors = ['#ff4500', '#ff6347', '#ff7f50', '#ffa500'];
        const color = colors[frame % colors.length];
        
        // 基础火焰
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.moveTo(x + tileSize * 0.3, y + tileSize * 0.8);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.2, x + tileSize * 0.7, y + tileSize * 0.8);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.5, x + tileSize * 0.3, y + tileSize * 0.8);
        ctx.fill();
        
        // 内焰
        ctx.fillStyle = '#ffff00';
        ctx.beginPath();
        ctx.moveTo(x + tileSize * 0.4, y + tileSize * 0.7);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.3, x + tileSize * 0.6, y + tileSize * 0.7);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.5, x + tileSize * 0.4, y + tileSize * 0.7);
        ctx.fill();
        
        // 火花
        if (frame % 2 === 0) {
          ctx.fillStyle = '#ff0000';
          for (let i = 0; i < 3; i++) {
            const sparkX = x + tileSize * (0.3 + Math.random() * 0.4);
            const sparkY = y + tileSize * (0.2 + Math.random() * 0.3);
            ctx.beginPath();
            ctx.arc(sparkX, sparkY, 2, 0, Math.PI * 2);
            ctx.fill();
          }
        }
      }
    });
    
    // 水流动画
    this.animationTypes.set('water', {
      name: '水流',
      frameCount: 8,
      duration: 200,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 水流波纹效果
        const offset = (frame * 2) % tileSize;
        
        // 基础水色
        ctx.fillStyle = '#1e4d8c';
        ctx.fillRect(x, y, tileSize, tileSize);
        
        // 波纹
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.3)';
        ctx.lineWidth = 1;
        
        for (let i = 0; i < 3; i++) {
          const waveY = y + (offset + i * tileSize / 3) % tileSize;
          ctx.beginPath();
          ctx.moveTo(x, waveY);
          for (let wx = 0; wx <= tileSize; wx += 8) {
            const waveOffset = Math.sin((wx + frame * 5) * 0.1) * 3;
            ctx.lineTo(x + wx, waveY + waveOffset);
          }
          ctx.stroke();
        }
        
        // 水面反光
        ctx.fillStyle = 'rgba(255, 255, 255, 0.1)';
        ctx.beginPath();
        ctx.arc(x + tileSize * 0.3, y + tileSize * 0.4, 4, 0, Math.PI * 2);
        ctx.fill();
        ctx.beginPath();
        ctx.arc(x + tileSize * 0.7, y + tileSize * 0.6, 3, 0, Math.PI * 2);
        ctx.fill();
      }
    });
    
    // 树木摆动动画
    this.animationTypes.set('tree', {
      name: '树木摆动',
      frameCount: 12,
      duration: 500,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 树干
        ctx.fillStyle = '#8b4513';
        ctx.fillRect(x + tileSize * 0.4, y + tileSize * 0.6, tileSize * 0.2, tileSize * 0.4);
        
        // 树冠摆动
        const swayAngle = Math.sin(frame * 0.5) * 0.1; // 摆动角度
        const swayOffset = Math.sin(frame * 0.5) * 3; // 摆动偏移
        
        ctx.save();
        ctx.translate(x + tileSize * 0.5, y + tileSize * 0.6);
        ctx.rotate(swayAngle);
        
        // 树冠
        ctx.fillStyle = '#228b22';
        ctx.beginPath();
        ctx.moveTo(0, -tileSize * 0.5);
        ctx.lineTo(-tileSize * 0.4, 0);
        ctx.lineTo(tileSize * 0.4, 0);
        ctx.closePath();
        ctx.fill();
        
        // 树冠阴影
        ctx.fillStyle = '#1e7b1e';
        ctx.beginPath();
        ctx.moveTo(0, -tileSize * 0.4);
        ctx.lineTo(-tileSize * 0.3, 0);
        ctx.lineTo(tileSize * 0.3, 0);
        ctx.closePath();
        ctx.fill();
        
        ctx.restore();
        
        // 树叶飘落效果
        if (frame % 6 === 0) {
          ctx.fillStyle = '#32cd32';
          const leafX = x + tileSize * (0.2 + Math.random() * 0.6);
          const leafY = y + tileSize * (0.8 + Math.random() * 0.2);
          ctx.beginPath();
          ctx.ellipse(leafX, leafY, 3, 2, Math.random() * Math.PI, 0, Math.PI * 2);
          ctx.fill();
        }
      }
    });
    
    // 瀑布动画
    this.animationTypes.set('waterfall', {
      name: '瀑布',
      frameCount: 6,
      duration: 150,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 瀑布背景
        ctx.fillStyle = '#4a90c2';
        ctx.fillRect(x, y, tileSize, tileSize);
        
        // 水流线条
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
        ctx.lineWidth = 2;
        
        for (let i = 0; i < 5; i++) {
          const lineX = x + tileSize * (0.2 + i * 0.15);
          const flowOffset = (frame * 4 + i * 10) % tileSize;
          
          ctx.beginPath();
          ctx.moveTo(lineX, y);
          ctx.lineTo(lineX, y + flowOffset);
          ctx.stroke();
        }
        
        // 水花
        ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
        for (let i = 0; i < 3; i++) {
          const splashX = x + tileSize * (0.3 + Math.random() * 0.4);
          const splashY = y + tileSize * (0.8 + Math.random() * 0.2);
          ctx.beginPath();
          ctx.arc(splashX, splashY, 2, 0, Math.PI * 2);
          ctx.fill();
        }
      }
    });
    
    // 火把动画
    this.animationTypes.set('torch', {
      name: '火把',
      frameCount: 4,
      duration: 300,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 火把杆
        ctx.fillStyle = '#8b4513';
        ctx.fillRect(x + tileSize * 0.45, y + tileSize * 0.5, tileSize * 0.1, tileSize * 0.5);
        
        // 火焰
        const colors = ['#ff4500', '#ff6347', '#ff7f50', '#ffa500'];
        const color = colors[frame % colors.length];
        
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.moveTo(x + tileSize * 0.4, y + tileSize * 0.5);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.2, x + tileSize * 0.6, y + tileSize * 0.5);
        ctx.quadraticCurveTo(x + tileSize * 0.5, y + tileSize * 0.35, x + tileSize * 0.4, y + tileSize * 0.5);
        ctx.fill();
        
        // 光晕效果
        const gradient = ctx.createRadialGradient(
          x + tileSize * 0.5, y + tileSize * 0.35, 0,
          x + tileSize * 0.5, y + tileSize * 0.35, tileSize * 0.4
        );
        gradient.addColorStop(0, 'rgba(255, 200, 0, 0.3)');
        gradient.addColorStop(1, 'rgba(255, 100, 0, 0)');
        ctx.fillStyle = gradient;
        ctx.fillRect(x, y, tileSize, tileSize);
      }
    });
    
    // 烟雾动画
    this.animationTypes.set('smoke', {
      name: '烟雾',
      frameCount: 8,
      duration: 400,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 烟雾粒子
        ctx.fillStyle = 'rgba(128, 128, 128, 0.3)';
        
        for (let i = 0; i < 4; i++) {
          const smokeX = x + tileSize * (0.3 + (i * 0.15 + frame * 0.02) % 0.4);
          const smokeY = y + tileSize * (0.8 - (frame * 0.05 + i * 0.1) % 0.8);
          const size = 3 + (frame + i * 2) % 5;
          
          ctx.beginPath();
          ctx.arc(smokeX, smokeY, size, 0, Math.PI * 2);
          ctx.fill();
        }
      }
    });
    
    // 花朵摇摆动画
    this.animationTypes.set('flower', {
      name: '花朵摇摆',
      frameCount: 16,
      duration: 600,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 花茎
        const swayAngle = Math.sin(frame * 0.4) * 0.15;
        
        ctx.save();
        ctx.translate(x + tileSize * 0.5, y + tileSize * 0.8);
        ctx.rotate(swayAngle);
        
        ctx.strokeStyle = '#228b22';
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.moveTo(0, 0);
        ctx.lineTo(0, -tileSize * 0.5);
        ctx.stroke();
        
        // 花瓣
        const petalColors = ['#ff69b4', '#ff1493', '#ff00ff'];
        const petalColor = petalColors[frame % petalColors.length];
        
        ctx.fillStyle = petalColor;
        for (let i = 0; i < 5; i++) {
          ctx.save();
          ctx.rotate((i * 72) * Math.PI / 180);
          ctx.beginPath();
          ctx.ellipse(0, -tileSize * 0.5, 4, 8, 0, 0, Math.PI * 2);
          ctx.fill();
          ctx.restore();
        }
        
        // 花蕊
        ctx.fillStyle = '#ffff00';
        ctx.beginPath();
        ctx.arc(0, -tileSize * 0.5, 3, 0, Math.PI * 2);
        ctx.fill();
        
        ctx.restore();
      }
    });
    
    // 旗帜飘动动画
    this.animationTypes.set('flag', {
      name: '旗帜飘动',
      frameCount: 12,
      duration: 200,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 旗杆
        ctx.fillStyle = '#8b4513';
        ctx.fillRect(x + tileSize * 0.45, y, tileSize * 0.1, tileSize * 0.9);
        
        // 旗帜
        const waveOffset = Math.sin(frame * 0.5) * 3;
        ctx.fillStyle = '#ff0000';
        ctx.beginPath();
        ctx.moveTo(x + tileSize * 0.55, y + tileSize * 0.1);
        ctx.lineTo(x + tileSize * 0.9 + waveOffset, y + tileSize * 0.25);
        ctx.lineTo(x + tileSize * 0.85 + waveOffset, y + tileSize * 0.45);
        ctx.lineTo(x + tileSize * 0.55, y + tileSize * 0.5);
        ctx.closePath();
        ctx.fill();
        
        // 旗帜图案
        ctx.fillStyle = '#ffff00';
        ctx.beginPath();
        ctx.arc(x + tileSize * 0.7 + waveOffset * 0.5, y + tileSize * 0.3, 4, 0, Math.PI * 2);
        ctx.fill();
      }
    });
    
    // 闪电动画
    this.animationTypes.set('lightning', {
      name: '闪电',
      frameCount: 4,
      duration: 150,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 只有特定帧显示闪电
        if (frame === 0 || frame === 2) {
          ctx.strokeStyle = '#00bfff';
          ctx.lineWidth = 2;
          ctx.beginPath();
          ctx.moveTo(x + tileSize * 0.5, y);
          ctx.lineTo(x + tileSize * 0.3, y + tileSize * 0.3);
          ctx.lineTo(x + tileSize * 0.6, y + tileSize * 0.3);
          ctx.lineTo(x + tileSize * 0.4, y + tileSize * 0.6);
          ctx.lineTo(x + tileSize * 0.7, y + tileSize * 0.6);
          ctx.lineTo(x + tileSize * 0.5, y + tileSize);
          ctx.stroke();
          
          // 光晕
          const gradient = ctx.createRadialGradient(
            x + tileSize * 0.5, y + tileSize * 0.5, 0,
            x + tileSize * 0.5, y + tileSize * 0.5, tileSize * 0.5
          );
          gradient.addColorStop(0, 'rgba(0, 191, 255, 0.3)');
          gradient.addColorStop(1, 'rgba(0, 191, 255, 0)');
          ctx.fillStyle = gradient;
          ctx.fillRect(x, y, tileSize, tileSize);
        }
      }
    });
    
    // 魔法阵动画
    this.animationTypes.set('magic_circle', {
      name: '魔法阵',
      frameCount: 16,
      duration: 300,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        const centerX = x + tileSize * 0.5;
        const centerY = y + tileSize * 0.5;
        const radius = tileSize * 0.4;
        
        // 外圈
        ctx.strokeStyle = `hsla(${frame * 22.5}, 70%, 60%, 0.8)`;
        ctx.lineWidth = 2;
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, 0, Math.PI * 2);
        ctx.stroke();
        
        // 内圈
        ctx.strokeStyle = `hsla(${(frame * 22.5 + 180) % 360}, 70%, 60%, 0.6)`;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius * 0.6, 0, Math.PI * 2);
        ctx.stroke();
        
        // 旋转的符文
        const runeAngle = (frame * 22.5) * Math.PI / 180;
        ctx.save();
        ctx.translate(centerX, centerY);
        ctx.rotate(runeAngle);
        
        ctx.fillStyle = '#ff00ff';
        for (let i = 0; i < 4; i++) {
          ctx.save();
          ctx.rotate((i * 90) * Math.PI / 180);
          ctx.beginPath();
          ctx.moveTo(radius * 0.3, 0);
          ctx.lineTo(radius * 0.5, -3);
          ctx.lineTo(radius * 0.7, 0);
          ctx.lineTo(radius * 0.5, 3);
          ctx.closePath();
          ctx.fill();
          ctx.restore();
        }
        
        ctx.restore();
        
        // 中心光点
        ctx.fillStyle = `rgba(255, 0, 255, ${0.5 + Math.sin(frame * 0.5) * 0.3})`;
        ctx.beginPath();
        ctx.arc(centerX, centerY, 4, 0, Math.PI * 2);
        ctx.fill();
      }
    });
    
    // 草地风吹动画
    this.animationTypes.set('grass_wind', {
      name: '草地风吹',
      frameCount: 20,
      duration: 400,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 基础草地
        ctx.fillStyle = '#3a5f0b';
        ctx.fillRect(x, y, tileSize, tileSize);
        
        // 草叶摆动
        const swayAngle = Math.sin(frame * 0.3) * 0.2;
        
        for (let i = 0; i < 8; i++) {
          const grassX = x + tileSize * (0.1 + i * 0.12);
          const grassHeight = tileSize * (0.3 + (i % 3) * 0.1);
          const individualSway = swayAngle * (1 + (i % 2) * 0.5);
          
          ctx.save();
          ctx.translate(grassX, y + tileSize);
          ctx.rotate(individualSway);
          
          ctx.strokeStyle = '#4a7f0b';
          ctx.lineWidth = 2;
          ctx.beginPath();
          ctx.moveTo(0, 0);
          ctx.quadraticCurveTo(3, -grassHeight * 0.5, 0, -grassHeight);
          ctx.stroke();
          
          ctx.restore();
        }
      }
    });
    
    // 雪花飘落动画
    this.animationTypes.set('snow', {
      name: '雪花飘落',
      frameCount: 30,
      duration: 500,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 雪花粒子
        ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
        
        for (let i = 0; i < 5; i++) {
          const snowX = x + tileSize * ((i * 0.2 + frame * 0.01) % 1);
          const snowY = y + tileSize * ((i * 0.15 + frame * 0.02) % 1);
          const size = 2 + (i % 2);
          
          ctx.beginPath();
          ctx.arc(snowX, snowY, size, 0, Math.PI * 2);
          ctx.fill();
        }
        
        // 雪花形状
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
        ctx.lineWidth = 1;
        for (let i = 0; i < 2; i++) {
          const flakeX = x + tileSize * ((i * 0.4 + frame * 0.015) % 1);
          const flakeY = y + tileSize * ((i * 0.3 + frame * 0.025) % 1);
          
          ctx.beginPath();
          ctx.moveTo(flakeX - 3, flakeY);
          ctx.lineTo(flakeX + 3, flakeY);
          ctx.moveTo(flakeX, flakeY - 3);
          ctx.lineTo(flakeX, flakeY + 3);
          ctx.stroke();
        }
      }
    });
    
    // 虫子爬行动画
    this.animationTypes.set('insect', {
      name: '虫子爬行',
      frameCount: 8,
      duration: 300,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 虫子位置
        const progress = (frame % 8) / 8;
        const insectX = x + tileSize * progress;
        const insectY = y + tileSize * 0.5;
        
        // 身体
        ctx.fillStyle = '#8b4513';
        ctx.beginPath();
        ctx.ellipse(insectX, insectY, 6, 3, 0, 0, Math.PI * 2);
        ctx.fill();
        
        // 腿
        ctx.strokeStyle = '#8b4513';
        ctx.lineWidth = 1;
        for (let i = 0; i < 3; i++) {
          const legOffset = (i - 1) * 3;
          const legAngle = Math.sin(frame * 0.8) * 0.3;
          
          ctx.beginPath();
          ctx.moveTo(insectX + legOffset, insectY);
          ctx.lineTo(insectX + legOffset + Math.cos(legAngle) * 4, insectY + Math.sin(legAngle) * 4);
          ctx.stroke();
        }
        
        // 触角
        ctx.beginPath();
        ctx.moveTo(insectX + 4, insectY - 1);
        ctx.lineTo(insectX + 6, insectY - 4);
        ctx.moveTo(insectX + 4, insectY + 1);
        ctx.lineTo(insectX + 6, insectY + 4);
        ctx.stroke();
      }
    });
    
    // 蘑菇呼吸动画
    this.animationTypes.set('mushroom', {
      name: '蘑菇呼吸',
      frameCount: 16,
      duration: 500,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        const centerX = x + tileSize * 0.5;
        const breatheScale = 1 + Math.sin(frame * 0.4) * 0.05;
        
        // 菌柄
        ctx.fillStyle = '#f5f5dc';
        ctx.fillRect(centerX - 3, y + tileSize * 0.5, 6, tileSize * 0.4);
        
        // 菌盖
        ctx.save();
        ctx.translate(centerX, y + tileSize * 0.5);
        ctx.scale(breatheScale, breatheScale);
        
        ctx.fillStyle = '#ff4500';
        ctx.beginPath();
        ctx.arc(0, 0, tileSize * 0.35, Math.PI, 0);
        ctx.closePath();
        ctx.fill();
        
        // 菌盖斑点
        ctx.fillStyle = '#ffffff';
        ctx.beginPath();
        ctx.arc(-5, -5, 3, 0, Math.PI * 2);
        ctx.arc(5, -7, 2, 0, Math.PI * 2);
        ctx.arc(0, -8, 2, 0, Math.PI * 2);
        ctx.fill();
        
        ctx.restore();
      }
    });
    
    // 宝石闪烁动画
    this.animationTypes.set('gem', {
      name: '宝石闪烁',
      frameCount: 12,
      duration: 400,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        const centerX = x + tileSize * 0.5;
        const centerY = y + tileSize * 0.5;
        const sparkleIntensity = 0.5 + Math.sin(frame * 0.5) * 0.5;
        
        // 宝石主体
        ctx.fillStyle = '#9932cc';
        ctx.beginPath();
        ctx.moveTo(centerX, centerY - 8);
        ctx.lineTo(centerX + 6, centerY);
        ctx.lineTo(centerX, centerY + 8);
        ctx.lineTo(centerX - 6, centerY);
        ctx.closePath();
        ctx.fill();
        
        // 宝石切面
        ctx.fillStyle = '#ba55d3';
        ctx.beginPath();
        ctx.moveTo(centerX, centerY - 8);
        ctx.lineTo(centerX + 3, centerY - 3);
        ctx.lineTo(centerX, centerY);
        ctx.lineTo(centerX - 3, centerY - 3);
        ctx.closePath();
        ctx.fill();
        
        // 闪光效果
        if (sparkleIntensity > 0.7) {
          ctx.strokeStyle = `rgba(255, 255, 255, ${sparkleIntensity})`;
          ctx.lineWidth = 2;
          ctx.beginPath();
          ctx.moveTo(centerX - 10, centerY - 10);
          ctx.lineTo(centerX - 5, centerY - 5);
          ctx.moveTo(centerX + 10, centerY - 10);
          ctx.lineTo(centerX + 5, centerY - 5);
          ctx.stroke();
        }
        
        // 光晕
        const gradient = ctx.createRadialGradient(
          centerX, centerY, 0,
          centerX, centerY, tileSize * 0.4
        );
        gradient.addColorStop(0, `rgba(153, 50, 204, ${sparkleIntensity * 0.3})`);
        gradient.addColorStop(1, 'rgba(153, 50, 204, 0)');
        ctx.fillStyle = gradient;
        ctx.fillRect(x, y, tileSize, tileSize);
      }
    });
    
    // 蜡烛燃烧动画
    this.animationTypes.set('candle', {
      name: '蜡烛燃烧',
      frameCount: 8,
      duration: 250,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        const centerX = x + tileSize * 0.5;
        
        // 蜡烛体
        ctx.fillStyle = '#ffe4c4';
        ctx.fillRect(centerX - 4, y + tileSize * 0.4, 8, tileSize * 0.6);
        
        // 烛芯
        ctx.strokeStyle = '#1a1a1a';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(centerX, y + tileSize * 0.4);
        ctx.lineTo(centerX, y + tileSize * 0.35);
        ctx.stroke();
        
        // 火焰
        const flameColors = ['#ff4500', '#ff6347', '#ffa500', '#ffff00'];
        const flameColor = flameColors[frame % flameColors.length];
        
        ctx.fillStyle = flameColor;
        ctx.beginPath();
        ctx.moveTo(centerX - 3, y + tileSize * 0.35);
        ctx.quadraticCurveTo(centerX + Math.sin(frame * 0.5) * 2, y + tileSize * 0.2, centerX + 3, y + tileSize * 0.35);
        ctx.quadraticCurveTo(centerX, y + tileSize * 0.28, centerX - 3, y + tileSize * 0.35);
        ctx.fill();
        
        // 烛光
        const gradient = ctx.createRadialGradient(
          centerX, y + tileSize * 0.3, 0,
          centerX, y + tileSize * 0.3, tileSize * 0.3
        );
        gradient.addColorStop(0, 'rgba(255, 200, 0, 0.4)');
        gradient.addColorStop(1, 'rgba(255, 100, 0, 0)');
        ctx.fillStyle = gradient;
        ctx.fillRect(x, y, tileSize, tileSize);
      }
    });
    
    // 蒸汽动画
    this.animationTypes.set('steam', {
      name: '蒸汽',
      frameCount: 10,
      duration: 350,
      loop: true,
      render: (ctx, x, y, tileSize, frame, tileset) => {
        // 蒸汽粒子
        for (let i = 0; i < 6; i++) {
          const steamX = x + tileSize * (0.2 + (i * 0.13 + frame * 0.015) % 0.6);
          const steamY = y + tileSize * (0.8 - (frame * 0.03 + i * 0.08) % 0.8);
          const size = 4 + (frame + i) % 4;
          const opacity = 0.3 + Math.sin(frame * 0.3 + i) * 0.2;
          
          ctx.fillStyle = `rgba(200, 200, 200, ${opacity})`;
          ctx.beginPath();
          ctx.arc(steamX, steamY, size, 0, Math.PI * 2);
          ctx.fill();
        }
      }
    });
  }
  
  /**
   * 添加动画实例
   */
  addAnimation(id, type, x, y, options = {}) {
    const animationType = this.animationTypes.get(type);
    if (!animationType) {
      console.warn(`未知的动画类型: ${type}`);
      return null;
    }
    
    const animation = {
      id,
      type,
      x,
      y,
      frame: 0,
      lastFrameTime: 0,
      options: { ...options },
      ...animationType
    };
    
    this.animations.set(id, animation);
    return animation;
  }
  
  /**
   * 移除动画实例
   */
  removeAnimation(id) {
    this.animations.delete(id);
  }
  
  /**
   * 清除所有动画
   */
  clearAll() {
    this.animations.clear();
  }
  
  /**
   * 启动动画系统
   */
  start() {
    if (this.isRunning) return;
    
    this.isRunning = true;
    this.lastTime = performance.now();
    this.animate();
  }
  
  /**
   * 停止动画系统
   */
  stop() {
    this.isRunning = false;
  }
  
  /**
   * 动画主循环
   */
  animate(currentTime = performance.now()) {
    if (!this.isRunning) return;
    
    const deltaTime = currentTime - this.lastTime;
    this.lastTime = currentTime;
    
    // 更新所有动画
    this.animations.forEach((animation) => {
      animation.lastFrameTime += deltaTime;
      
      if (animation.lastFrameTime >= animation.duration) {
        animation.lastFrameTime = 0;
        animation.frame++;
        
        if (animation.loop) {
          animation.frame %= animation.frameCount;
        } else if (animation.frame >= animation.frameCount) {
          // 非循环动画，完成后移除
          this.removeAnimation(animation.id);
          return;
        }
      }
    });
    
    requestAnimationFrame((time) => this.animate(time));
  }
  
  /**
   * 渲染所有动画
   */
  render(ctx, tileSize, tileset = null) {
    this.animations.forEach((animation) => {
      if (animation.render) {
        animation.render(ctx, animation.x, animation.y, tileSize, animation.frame, tileset);
      }
    });
  }
  
  /**
   * 获取指定位置的动画
   */
  getAnimationAt(x, y) {
    for (const animation of this.animations.values()) {
      if (animation.x === x && animation.y === y) {
        return animation;
      }
    }
    return null;
  }
  
  /**
   * 获取所有动画类型
   */
  getAnimationTypes() {
    return Array.from(this.animationTypes.entries()).map(([key, value]) => ({
      key,
      name: value.name,
      frameCount: value.frameCount,
      duration: value.duration
    }));
  }
  
  /**
   * 注册自定义动画类型
   */
  registerAnimationType(key, config) {
    this.animationTypes.set(key, config);
  }
  
  /**
   * 获取动画统计信息
   */
  getStats() {
    return {
      totalAnimations: this.animations.size,
      isRunning: this.isRunning,
      animationTypes: this.animationTypes.size
    };
  }
}

// 创建全局单例
window.MapAnimationSystem = new MapAnimationSystem();