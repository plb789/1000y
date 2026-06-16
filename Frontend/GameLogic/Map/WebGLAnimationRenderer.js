/**
 * WebGL动画渲染器
 * 使用WebGL优化动画渲染性能
 */
class WebGLAnimationRenderer {
  constructor(canvas) {
    this.canvas = canvas;
    this.gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
    
    if (!this.gl) {
      console.error('WebGL不可用，将回退到Canvas 2D');
      this.enabled = false;
      return;
    }
    
    this.enabled = true;
    this.program = null;
    this.buffers = {};
    this.textures = {};
    this.uniforms = {};
    
    this.init();
  }
  
  /**
   * 初始化WebGL
   */
  init() {
    const gl = this.gl;
    
    // 顶点着色器
    const vertexShaderSource = `
      attribute vec2 a_position;
      attribute vec2 a_texCoord;
      attribute vec4 a_color;
      
      uniform vec2 u_resolution;
      uniform vec2 u_translation;
      uniform float u_scale;
      uniform float u_rotation;
      
      varying vec2 v_texCoord;
      varying vec4 v_color;
      
      void main() {
        // 应用旋转
        float cosR = cos(u_rotation);
        float sinR = sin(u_rotation);
        vec2 rotatedPos = vec2(
          a_position.x * cosR - a_position.y * sinR,
          a_position.x * sinR + a_position.y * cosR
        );
        
        // 应用缩放和平移
        vec2 position = rotatedPos * u_scale + u_translation;
        
        // 转换到裁剪空间
        vec2 clipSpace = (position / u_resolution) * 2.0 - 1.0;
        
        gl_Position = vec4(clipSpace * vec2(1, -1), 0, 1);
        v_texCoord = a_texCoord;
        v_color = a_color;
      }
    `;
    
    // 片段着色器
    const fragmentShaderSource = `
      precision mediump float;
      
      varying vec2 v_texCoord;
      varying vec4 v_color;
      
      uniform sampler2D u_texture;
      uniform bool u_useTexture;
      uniform float u_opacity;
      uniform float u_time;
      uniform vec4 u_tintColor;
      
      void main() {
        vec4 finalColor;
        
        if (u_useTexture) {
          finalColor = texture2D(u_texture, v_texCoord) * v_color;
        } else {
          finalColor = v_color;
        }
        
        // 应用着色
        finalColor *= u_tintColor;
        
        // 应用透明度
        finalColor.a *= u_opacity;
        
        gl_FragColor = finalColor;
      }
    `;
    
    // 编译着色器
    const vertexShader = this.compileShader(gl.VERTEX_SHADER, vertexShaderSource);
    const fragmentShader = this.compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource);
    
    // 创建程序
    this.program = gl.createProgram();
    gl.attachShader(this.program, vertexShader);
    gl.attachShader(this.program, fragmentShader);
    gl.linkProgram(this.program);
    
    if (!gl.getProgramParameter(this.program, gl.LINK_STATUS)) {
      console.error('WebGL程序链接失败:', gl.getProgramInfoLog(this.program));
      this.enabled = false;
      return;
    }
    
    // 获取属性和uniform位置
    this.attributes = {
      position: gl.getAttribLocation(this.program, 'a_position'),
      texCoord: gl.getAttribLocation(this.program, 'a_texCoord'),
      color: gl.getAttribLocation(this.program, 'a_color')
    };
    
    this.uniforms = {
      resolution: gl.getUniformLocation(this.program, 'u_resolution'),
      translation: gl.getUniformLocation(this.program, 'u_translation'),
      scale: gl.getUniformLocation(this.program, 'u_scale'),
      rotation: gl.getUniformLocation(this.program, 'u_rotation'),
      texture: gl.getUniformLocation(this.program, 'u_texture'),
      useTexture: gl.getUniformLocation(this.program, 'u_useTexture'),
      opacity: gl.getUniformLocation(this.program, 'u_opacity'),
      time: gl.getUniformLocation(this.program, 'u_time'),
      tintColor: gl.getUniformLocation(this.program, 'u_tintColor')
    };
    
    // 创建缓冲区
    this.createBuffers();
    
    // 设置默认状态
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
    
    console.log('WebGL动画渲染器初始化成功');
  }
  
  /**
   * 编译着色器
   */
  compileShader(type, source) {
    const gl = this.gl;
    const shader = gl.createShader(type);
    gl.shaderSource(shader, source);
    gl.compileShader(shader);
    
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
      console.error('着色器编译失败:', gl.getShaderInfoLog(shader));
      gl.deleteShader(shader);
      return null;
    }
    
    return shader;
  }
  
  /**
   * 创建缓冲区
   */
  createBuffers() {
    const gl = this.gl;
    
    // 顶点缓冲区（单位矩形）
    const vertices = new Float32Array([
      -0.5, -0.5,  0.0, 0.0,  1.0, 1.0, 1.0, 1.0,  // 左下
       0.5, -0.5,  1.0, 0.0,  1.0, 1.0, 1.0, 1.0,  // 右下
       0.5,  0.5,  1.0, 1.0,  1.0, 1.0, 1.0, 1.0,  // 右上
      -0.5,  0.5,  0.0, 1.0,  1.0, 1.0, 1.0, 1.0   // 左上
    ]);
    
    this.buffers.vertex = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.vertex);
    gl.bufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW);
    
    // 索引缓冲区
    const indices = new Uint16Array([
      0, 1, 2,  // 第一个三角形
      0, 2, 3   // 第二个三角形
    ]);
    
    this.buffers.index = gl.createBuffer();
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.buffers.index);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, indices, gl.STATIC_DRAW);
  }
  
  /**
   * 创建纹理
   */
  createTexture(image) {
    const gl = this.gl;
    const texture = gl.createTexture();
    
    gl.bindTexture(gl.TEXTURE_2D, texture);
    
    // 设置纹理参数
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
    
    // 上传纹理数据
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, image);
    
    return texture;
  }
  
  /**
   * 开始渲染
   */
  beginRender(width, height) {
    const gl = this.gl;
    
    gl.viewport(0, 0, width, height);
    gl.clearColor(0, 0, 0, 0);
    gl.clear(gl.COLOR_BUFFER_BIT);
    
    gl.useProgram(this.program);
    
    // 设置uniform
    gl.uniform2f(this.uniforms.resolution, width, height);
    gl.uniform1f(this.uniforms.time, performance.now() / 1000);
    
    // 绑定缓冲区
    gl.bindBuffer(gl.ARRAY_BUFFER, this.buffers.vertex);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, this.buffers.index);
    
    // 设置属性指针
    const stride = 8 * 4; // 8个float，每个4字节
    
    gl.enableVertexAttribArray(this.attributes.position);
    gl.vertexAttribPointer(this.attributes.position, 2, gl.FLOAT, false, stride, 0);
    
    gl.enableVertexAttribArray(this.attributes.texCoord);
    gl.vertexAttribPointer(this.attributes.texCoord, 2, gl.FLOAT, false, stride, 2 * 4);
    
    gl.enableVertexAttribArray(this.attributes.color);
    gl.vertexAttribPointer(this.attributes.color, 4, gl.FLOAT, false, stride, 4 * 4);
  }
  
  /**
   * 渲染动画
   */
  renderAnimation(animation, tileSize, tileset = null) {
    const gl = this.gl;
    
    // 设置变换
    gl.uniform2f(this.uniforms.translation, animation.x + tileSize / 2, animation.y + tileSize / 2);
    gl.uniform1f(this.uniforms.scale, tileSize);
    gl.uniform1f(this.uniforms.rotation, 0);
    gl.uniform1f(this.uniforms.opacity, 1.0);
    gl.uniform4f(this.uniforms.tintColor, 1.0, 1.0, 1.0, 1.0);
    
    // 设置纹理
    if (tileset) {
      gl.activeTexture(gl.TEXTURE0);
      gl.bindTexture(gl.TEXTURE_2D, tileset);
      gl.uniform1i(this.uniforms.texture, 0);
      gl.uniform1i(this.uniforms.useTexture, 1);
    } else {
      gl.uniform1i(this.uniforms.useTexture, 0);
    }
    
    // 绘制
    gl.drawElements(gl.TRIANGLES, 6, gl.UNSIGNED_SHORT, 0);
  }
  
  /**
   * 批量渲染动画
   */
  renderAnimations(animations, tileSize, tileset = null) {
    animations.forEach(animation => {
      this.renderAnimation(animation, tileSize, tileset);
    });
  }
  
  /**
   * 结束渲染
   */
  endRender() {
    // 可以在这里添加后处理效果
  }
  
  /**
   * 清理资源
   */
  dispose() {
    const gl = this.gl;
    
    // 删除缓冲区
    Object.values(this.buffers).forEach(buffer => {
      gl.deleteBuffer(buffer);
    });
    
    // 删除纹理
    Object.values(this.textures).forEach(texture => {
      gl.deleteTexture(texture);
    });
    
    // 删除程序
    if (this.program) {
      gl.deleteProgram(this.program);
    }
    
    this.enabled = false;
  }
  
  /**
   * 检查WebGL支持
   */
  static isSupported() {
    try {
      const canvas = document.createElement('canvas');
      return !!(canvas.getContext('webgl') || canvas.getContext('experimental-webgl'));
    } catch (e) {
      return false;
    }
  }
  
  /**
   * 获取性能信息
   */
  getPerformanceInfo() {
    if (!this.enabled || !this.gl) {
      return { enabled: false };
    }
    
    const gl = this.gl;
    const debugInfo = gl.getExtension('WEBGL_debug_renderer_info');
    
    return {
      enabled: true,
      renderer: debugInfo ? gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL) : 'Unknown',
      vendor: debugInfo ? gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL) : 'Unknown',
      maxTextureSize: gl.getParameter(gl.MAX_TEXTURE_SIZE),
      maxVertexAttribs: gl.getParameter(gl.MAX_VERTEX_ATTRIBS),
      maxFragmentUniforms: gl.getParameter(gl.MAX_FRAGMENT_UNIFORM_VECTORS),
      maxVertexUniforms: gl.getParameter(gl.MAX_VERTEX_UNIFORM_VECTORS)
    };
  }
}

// 创建全局单例
window.WebGLAnimationRenderer = WebGLAnimationRenderer;