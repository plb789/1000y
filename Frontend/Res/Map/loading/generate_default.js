<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>默认加载背景生成器</title>
  <script>
    // 使用Canvas生成默认背景图片
    function generateDefaultBackground() {
      const canvas = document.createElement('canvas');
      canvas.width = 1920;
      canvas.height = 1080;
      const ctx = canvas.getContext('2d');
      
      // 创建渐变背景
      const gradient = ctx.createLinearGradient(0, 0, 1920, 1080);
      gradient.addColorStop(0, '#1a1a2e');
      gradient.addColorStop(0.3, '#16213e');
      gradient.addColorStop(0.6, '#0f3460');
      gradient.addColorStop(1, '#1a1a2e');
      ctx.fillStyle = gradient;
      ctx.fillRect(0, 0, 1920, 1080);
      
      // 添加装饰性图案
      ctx.fillStyle = 'rgba(233, 69, 96, 0.05)';
      for (let i = 0; i < 50; i++) {
        const x = Math.random() * 1920;
        const y = Math.random() * 1080;
        const radius = Math.random() * 50 + 10;
        ctx.beginPath();
        ctx.arc(x, y, radius, 0, Math.PI * 2);
        ctx.fill();
      }
      
      // 添加标题文字效果（模拟）
      ctx.fillStyle = 'rgba(233, 69, 96, 0.1)';
      ctx.font = 'bold 120px Microsoft YaHei';
      ctx.textAlign = 'center';
      ctx.fillText('千年江湖', 960, 540);
      
      // 输出base64
      const dataUrl = canvas.toDataURL('image/jpeg', 0.8);
      console.log('Default background generated!');
      
      // 如果是Node.js环境，保存文件
      if (typeof module !== 'undefined' && module.exports) {
        const fs = require('fs');
        const base64Data = dataUrl.replace(/^data:image\/jpeg;base64,/, '');
        fs.writeFileSync('default.jpg', Buffer.from(base64Data, 'base64'));
        console.log('Saved default.jpg');
      }
    }
    
    // 浏览器环境直接显示
    if (typeof window !== 'undefined') {
      window.onload = generateDefaultBackground;
    }
  </script>
</head>
<body>
  <h1>默认加载背景生成器</h1>
  <p>打开开发者控制台查看生成的图片数据。</p>
</body>
</html>