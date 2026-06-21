/**
 * 卡通人形角色渲染器
 * 支持两种渲染模式，自动降级：
 *   1. 资源模式（优先）：从 assets/characters/ 加载精灵图/序列帧
 *   2. 矢量模式（降级）：找不到资源时使用 Canvas 矢量绘制
 *
 * 资源约定：
 *   - 清单文件：assets/characters/manifest.json
 *   - 图片路径：assets/characters/{gender}/{state}_{dir}_{frame}.png
 *     gender:  male / female
 *     state:   idle / walk / attack / cast / dead
 *     dir:     8方向: down / down_left / left / up_left / up / up_right / right / down_right
 *              4方向: down / left / right / up （兼容模式，对角方向自动降级）
 *     frame:   0, 1, 2 ...
 *
 * 性别: 0=男, 1=女
 * 方向: 0=下, 1=左下, 2=左, 3=左上, 4=上, 5=右上, 6=右, 7=右下（顺时针）
 */
class CharacterRenderer {
  constructor() {
    // 动画状态
    this.state = 'idle';
    this.direction = 0;
    this.animTime = 0;
    this.stateTime = 0;

    // 动画时长
    this.attackDuration = 350;
    this.castDuration = 500;
    this.walkSpeed = 0.012;

    // 资源系统
    this.resourceMode = false;          // 是否启用资源模式
    this.resourceLoaded = false;        // 资源是否已加载完成
    this.resourceLoadFailed = false;    // 资源是否加载失败（永久降级）
    this.sprites = {};                  // { 'male_idle_down_0': HTMLImageElement, ... }
    this.frameCounts = {};              // { 'male_idle_down': 4, ... } 每个动画的帧数
    this.frameDuration = 120;           // 序列帧间隔(ms)

    // 性别外观配置（矢量模式使用）
    this.presets = {
      0: {
        name: '男',
        bodyColor: '#3b82f6', bodyShadow: '#1e40af',
        pantsColor: '#1f2937', hairColor: '#4b2e1a', hairStyle: 'short',
        skinColor: '#fde0b5', skinShadow: '#e6c79c', eyeColor: '#1f2937'
      },
      1: {
        name: '女',
        bodyColor: '#ec4899', bodyShadow: '#be185d',
        pantsColor: '#831843', hairColor: '#7c2d12', hairStyle: 'long',
        skinColor: '#fde0b5', skinShadow: '#e6c79c', eyeColor: '#1f2937'
      }
    };

    // 8方向名称映射（顺时针，从下开始）
    // 0=下, 1=左下, 2=左, 3=左上, 4=上, 5=右上, 6=右, 7=右下
    this.dirNames = [
      'down',        // 0
      'down_left',   // 1
      'left',        // 2
      'up_left',     // 3
      'up',          // 4
      'up_right',    // 5
      'right',       // 6
      'down_right'   // 7
    ];

    // 对角方向降级到正方向（4方向资源兼容）
    // 当对角方向资源缺失时，降级到对应的正方向
    this.diagonalFallback = {
      1: 0,   // down_left -> down
      3: 2,   // up_left   -> left
      5: 6,   // up_right  -> right
      7: 0    // down_right-> down
    };

    this.stateNames = ['idle', 'walk', 'attack', 'cast', 'dead'];
  }

  /* ==================== 状态控制 ==================== */

  setState(state) {
    if (this.state !== state) {
      this.state = state;
      this.stateTime = 0;
    }
  }

  /**
   * 根据移动向量设置方向（支持8方向）
   * 使用角度判定，将360度分成8个区间，每个45度
   * 0=下, 1=左下, 2=左, 3=左上, 4=上, 5=右上, 6=右, 7=右下
   */
  setDirection(dx, dy) {
    if (dx === 0 && dy === 0) return;

    // 计算角度（atan2 返回 -PI~PI，0=右方向，PI/2=下方向）
    // 注意：Canvas Y轴向下为正，所以 dy>0 表示向下
    const angle = Math.atan2(dy, dx);
    // 转换为 0~2PI，并以"下"为0度（顺时针）
    // atan2: 0=右, PI/2=下, PI/-PI=左, -PI/2=上
    // 我们要: 0=下, PI/2=左, PI=上, -PI/2=右 (顺时针从下开始)
    // 所以需要旋转 -PI/2 (即顺时针旋转90度)
    let normalizedAngle = angle + Math.PI / 2;
    if (normalizedAngle < 0) normalizedAngle += Math.PI * 2;

    // 将角度映射到 0-7 八个方向
    // 每个方向占 45度（PI/4），中心点在 0, 45, 90, 135, 180, 225, 270, 315
    // 加上 22.5度偏移使区间对齐到 [0,45), [45,90), ...
    const sector = Math.floor((normalizedAngle + Math.PI / 8) / (Math.PI / 4)) % 8;
    this.direction = sector;
  }

  playAttack() { this.setState('attack'); }
  playCast() { this.setState('cast'); }
  playDead() { this.setState('dead'); }

  update(deltaTime) {
    this.animTime += deltaTime;
    this.stateTime += deltaTime;
    if (this.state === 'attack' && this.stateTime >= this.attackDuration) {
      this.setState('idle');
    } else if (this.state === 'cast' && this.stateTime >= this.castDuration) {
      this.setState('idle');
    }
  }

  /* ==================== 资源加载 ==================== */

  /**
   * 异步加载角色资源清单
   * 加载失败会自动降级到矢量模式
   * @param {string} manifestUrl - 清单文件URL，默认 'assets/characters/manifest.json'
   */
  async loadResources(manifestUrl = 'assets/characters/manifest.json') {
    if (this.resourceLoaded || this.resourceLoadFailed) return;

    try {
      const response = await fetch(manifestUrl, { cache: 'no-cache' });
      if (!response.ok) {
        console.log(`[CharacterRenderer] 清单不存在，使用矢量模式: ${manifestUrl}`);
        this.resourceLoadFailed = true;
        return;
      }
      const manifest = await response.json();
      await this._loadManifest(manifest);
      this.resourceMode = true;
      this.resourceLoaded = true;
      console.log('[CharacterRenderer] 资源模式已启用', Object.keys(this.sprites).length, '帧');
    } catch (err) {
      console.warn('[CharacterRenderer] 资源加载失败，降级到矢量模式:', err);
      this.resourceLoadFailed = true;
    }
  }

  /**
   * 解析清单并加载所有图片
   * 清单格式：
   * {
   *   "version": 1,
   *   "frame_duration": 120,
   *   "characters": {
   *     "male": {
   *       "idle_down": ["male/idle_down_0.png", "male/idle_down_1.png"],
   *       "walk_down": ["male/walk_down_0.png", "male/walk_down_1.png", "male/walk_down_2.png"],
   *       "attack_down": ["male/attack_down_0.png"]
   *     },
   *     "female": { ... }
   *   }
   * }
   */
  async _loadManifest(manifest) {
    if (!manifest || !manifest.characters) {
      throw new Error('清单格式错误');
    }
    if (manifest.frame_duration) {
      this.frameDuration = manifest.frame_duration;
    }

    const baseDir = manifest.base_dir || 'assets/characters/';
    const tasks = [];

    for (const [genderKey, anims] of Object.entries(manifest.characters)) {
      for (const [animKey, frames] of Object.entries(anims)) {
        if (!Array.isArray(frames) || frames.length === 0) continue;
        // 记录帧数
        this.frameCounts[`${genderKey}_${animKey}`] = frames.length;
        // 加载每一帧
        frames.forEach((frameUrl, idx) => {
          const fullUrl = frameUrl.startsWith('http') || frameUrl.startsWith('/')
            ? frameUrl
            : baseDir + frameUrl;
          const cacheKey = `${genderKey}_${animKey}_${idx}`;
          tasks.push(
            window.ResourceManager.loadImage(fullUrl)
              .then(img => { this.sprites[cacheKey] = img; })
              .catch(err => {
                console.warn(`[CharacterRenderer] 帧加载失败: ${fullUrl}`, err);
              })
          );
        });
      }
    }

    await Promise.all(tasks);

    // 如果一张图都没加载成功，视为失败
    if (Object.keys(this.sprites).length === 0) {
      throw new Error('所有帧加载失败');
    }
  }

  /**
   * 获取当前帧图片
   * 降级策略：
   *   1. 精确匹配 8方向 (如 attack_down_left)
   *   2. 对角方向降级到正方向 (down_left -> down)
   *   3. 方向无关动画 (如 dead)
   *   4. idle_down 兜底
   */
  _getCurrentFrame(gender) {
    const genderKey = gender === 1 ? 'female' : 'male';
    const stateName = this.state;
    const dirName = this.dirNames[this.direction];
    const animKey = `${stateName}_${dirName}`;
    const prefix = `${genderKey}_${animKey}`;

    // 1. 精确匹配当前方向
    let actualPrefix = prefix;
    if (!this.frameCounts[actualPrefix]) {
      // 2. 对角方向降级到正方向
      const fallbackDir = this.diagonalFallback[this.direction];
      if (fallbackDir !== undefined) {
        const fallbackDirName = this.dirNames[fallbackDir];
        const fallbackKey = `${genderKey}_${stateName}_${fallbackDirName}`;
        if (this.frameCounts[fallbackKey]) {
          actualPrefix = fallbackKey;
        }
      }
      // 3. 方向无关动画（如 dead）
      if (!this.frameCounts[actualPrefix]) {
        const dirLessPrefix = `${genderKey}_${stateName}`;
        if (this.frameCounts[dirLessPrefix]) {
          actualPrefix = dirLessPrefix;
        } else {
          // 4. idle_down 兜底
          actualPrefix = `${genderKey}_idle_down`;
          if (!this.frameCounts[actualPrefix]) return null;
        }
      }
    }

    const count = this.frameCounts[actualPrefix];
    let frameIdx;
    if (this.state === 'attack' || this.state === 'cast') {
      // 攻击/施法按 stateTime 比例取帧
      const duration = this.state === 'attack' ? this.attackDuration : this.castDuration;
      frameIdx = Math.min(count - 1, Math.floor((this.stateTime / duration) * count));
    } else {
      // 待机/行走按 animTime 循环
      frameIdx = Math.floor(this.animTime / this.frameDuration) % count;
    }

    return this.sprites[`${actualPrefix}_${frameIdx}`] || null;
  }

  /* ==================== 主绘制入口 ==================== */

  /**
   * 绘制角色
   * @param {CanvasRenderingContext2D} ctx
   * @param {number} cx - 角色中心X(像素)
   * @param {number} cy - 角色底部Y(像素)
   * @param {number} tileSize - 格子大小
   * @param {object} options - { gender, name, isSelf, isMoving, moveDx, moveDy }
   */
  draw(ctx, cx, cy, tileSize, options = {}) {
    const gender = options.gender === 1 ? 1 : 0;

    // 更新朝向
    if (options.moveDx !== undefined && options.moveDy !== undefined) {
      if (options.moveDx !== 0 || options.moveDy !== 0) {
        this.setDirection(options.moveDx, options.moveDy);
      }
    }

    // 行走状态判定
    if (options.isMoving && this.state === 'idle') {
      this.setState('walk');
    } else if (!options.isMoving && this.state === 'walk') {
      this.setState('idle');
    }

    // 优先使用资源模式
    if (this.resourceMode && this.resourceLoaded) {
      const frame = this._getCurrentFrame(gender);
      if (frame) {
        this._drawSprite(ctx, frame, cx, cy, tileSize);
        this._drawNameTag(ctx, cx, cy - tileSize * 0.7, options.name, options.isSelf);
        return;
      }
      // 帧缺失，降级到矢量
    }

    // 矢量模式
    this._drawVector(ctx, cx, cy, tileSize, gender, options);
  }

  /* ==================== 资源模式绘制 ==================== */

  _drawSprite(ctx, img, cx, cy, tileSize) {
    // 按格子大小等比缩放绘制
    // 图片设计基准为 48x48，按 tileSize 自适应
    const designSize = 48;
    const scale = tileSize / designSize;
    const w = img.width * scale;
    const h = img.height * scale;
    // 底部居中对齐
    ctx.drawImage(img, cx - w / 2, cy - h, w, h);
  }

  /* ==================== 矢量模式绘制 ==================== */

  _drawVector(ctx, cx, cy, tileSize, gender, options) {
    const preset = this.presets[gender];
    const ts = tileSize;

    // 矢量模式只绘制4正方向，对角方向通过水平翻转近似
    // 将8方向映射到4方向用于矢量绘制
    const vectorDir = this._getVectorDirection();
    const isFlipped = this._isDirectionFlipped();
    const savedDir = this.direction;
    this.direction = vectorDir;

    ctx.save();
    ctx.translate(cx, cy);

    // 对角方向应用水平翻转（左对角 -> 翻转右方向）
    if (isFlipped) {
      ctx.scale(-1, 1);
    }

    // 死亡状态：旋转倒地
    if (this.state === 'dead') {
      ctx.rotate(Math.PI / 2);
      ctx.translate(0, -ts * 0.1);
    }

    // 攻击动画：向前突进
    let attackOffset = 0;
    if (this.state === 'attack') {
      const t = this.stateTime / this.attackDuration;
      if (t < 0.4) attackOffset = -ts * 0.08 * (t / 0.4);
      else if (t < 0.7) attackOffset = ts * 0.15 * ((t - 0.4) / 0.3);
      else attackOffset = ts * 0.15 * (1 - (t - 0.7) / 0.3);
      const dirVec = this._getDirectionVector();
      ctx.translate(dirVec.x * attackOffset, dirVec.y * attackOffset);
    }

    // 施法动画：身体上抬
    let castRaise = 0;
    if (this.state === 'cast') {
      const t = this.stateTime / this.castDuration;
      castRaise = Math.sin(t * Math.PI) * ts * 0.08;
    }

    // 阴影
    this._drawShadow(ctx, ts);

    // 缩放到设计基准
    const scale = ts / 48;
    ctx.scale(scale, scale);

    // 行走摆动
    let walkBob = 0, armSwing = 0;
    if (this.state === 'walk') {
      walkBob = Math.sin(this.animTime * this.walkSpeed) * 1.5;
      armSwing = Math.sin(this.animTime * this.walkSpeed) * 0.3;
    }

    this._drawLegs(ctx, preset, walkBob);
    this._drawBody(ctx, preset, walkBob, castRaise);
    this._drawArms(ctx, preset, armSwing, castRaise);
    this._drawHead(ctx, preset, walkBob);

    ctx.restore();

    // 恢复原始方向
    this.direction = savedDir;

    // 名字标签
    this._drawNameTag(ctx, cx, cy - ts * 0.7, options.name, options.isSelf);
  }

  /**
   * 获取矢量模式使用的4方向（对角方向映射到正方向）
   * 0=下, 1=左下->下, 2=左, 3=左上->上, 4=上, 5=右上->上, 6=右, 7=右下->下
   * 注意：这里返回的是4方向编号体系（0=下,1=左,2=右,3=上）
   */
  _getVectorDirection() {
    switch (this.direction) {
      case 0: return 0;  // down
      case 1: return 0;  // down_left -> down
      case 2: return 1;  // left
      case 3: return 3;  // up_left -> up
      case 4: return 3;  // up
      case 5: return 3;  // up_right -> up
      case 6: return 2;  // right
      case 7: return 0;  // down_right -> down
      default: return 0;
    }
  }

  /**
   * 判断当前方向是否需要水平翻转
   * 左侧方向（左下、左、左上）使用右侧资源翻转
   */
  _isDirectionFlipped() {
    // 矢量模式中，左侧方向（1=左下, 2=左, 3=左上）通过翻转右侧绘制
    // 但矢量绘制本身已经处理了左右眼睛，所以只有对角方向才需要翻转
    // 这里我们让矢量模式直接使用对应方向，不翻转
    return false;
  }

  _getDirectionVector() {
    switch (this.direction) {
      case 0: return { x: 0, y: 1 };      // down
      case 1: return { x: -0.7, y: 0.7 }; // down_left
      case 2: return { x: -1, y: 0 };     // left
      case 3: return { x: -0.7, y: -0.7 };// up_left
      case 4: return { x: 0, y: -1 };     // up
      case 5: return { x: 0.7, y: -0.7 }; // up_right
      case 6: return { x: 1, y: 0 };      // right
      case 7: return { x: 0.7, y: 0.7 };  // down_right
      default: return { x: 0, y: 1 };
    }
  }

  _drawShadow(ctx, ts) {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.35)';
    ctx.beginPath();
    ctx.ellipse(0, 2, ts * 0.28, ts * 0.08, 0, 0, Math.PI * 2);
    ctx.fill();
  }

  _drawLegs(ctx, p, walkBob) {
    const legOffset = walkBob * 0.5;
    ctx.fillStyle = p.pantsColor;
    ctx.fillRect(-7, -6 + legOffset, 5, 12);
    ctx.fillRect(2, -6 - legOffset, 5, 12);
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(-7, 4 + legOffset, 5, 3);
    ctx.fillRect(2, 4 - legOffset, 5, 3);
  }

  _drawBody(ctx, p, walkBob, castRaise) {
    const bob = walkBob * 0.3;
    ctx.fillStyle = p.bodyColor;
    ctx.fillRect(-9, -22 - bob, 18, 18);
    ctx.fillStyle = p.bodyShadow;
    ctx.fillRect(3, -22 - bob, 6, 18);
    ctx.fillStyle = '#fbbf24';
    ctx.fillRect(-9, -8 - bob, 18, 2);

    if (p.hairStyle === 'long') {
      ctx.fillStyle = p.pantsColor;
      ctx.beginPath();
      ctx.moveTo(-9, -8 - bob);
      ctx.lineTo(-12, 2 - bob);
      ctx.lineTo(12, 2 - bob);
      ctx.lineTo(9, -8 - bob);
      ctx.closePath();
      ctx.fill();
    }

    if (castRaise > 0) ctx.translate(0, -castRaise);
  }

  _drawArms(ctx, p, armSwing, castRaise) {
    const isAttack = this.state === 'attack';
    const isCast = this.state === 'cast';
    let leftArmAngle = armSwing;
    let rightArmAngle = -armSwing;

    if (isAttack) {
      const t = this.stateTime / this.attackDuration;
      if (t < 0.4) rightArmAngle = -0.8 * (t / 0.4);
      else if (t < 0.7) rightArmAngle = 1.2 * ((t - 0.4) / 0.3);
      else rightArmAngle = 1.2 * (1 - (t - 0.7) / 0.3);
    } else if (isCast) {
      const t = this.stateTime / this.castDuration;
      const raise = Math.sin(t * Math.PI);
      leftArmAngle = -1.2 * raise;
      rightArmAngle = 1.2 * raise;
    }

    // 左臂
    ctx.save();
    ctx.translate(-9, -20);
    ctx.rotate(leftArmAngle);
    ctx.fillStyle = p.bodyColor;
    ctx.fillRect(-3, 0, 4, 12);
    ctx.fillStyle = p.skinColor;
    ctx.fillRect(-3, 11, 4, 3);
    ctx.restore();

    // 右臂
    ctx.save();
    ctx.translate(9, -20);
    ctx.rotate(rightArmAngle);
    ctx.fillStyle = p.bodyColor;
    ctx.fillRect(-1, 0, 4, 12);
    ctx.fillStyle = p.skinColor;
    ctx.fillRect(-1, 11, 4, 3);
    ctx.restore();

    // 攻击时绘制武器
    if (isAttack) {
      ctx.save();
      ctx.translate(9, -20);
      ctx.rotate(rightArmAngle);
      ctx.translate(0, 14);
      ctx.fillStyle = '#e5e7eb';
      ctx.fillRect(-1, 0, 2, 14);
      ctx.fillStyle = '#92400e';
      ctx.fillRect(-2, -2, 4, 3);
      ctx.fillStyle = 'rgba(255, 255, 255, 0.6)';
      ctx.fillRect(-2, 2, 4, 10);
      ctx.restore();
    }

    // 施法时绘制法球
    if (isCast) {
      const t = this.stateTime / this.castDuration;
      const radius = 3 + Math.sin(t * Math.PI * 4) * 1.5;
      ctx.save();
      ctx.translate(0, -32);
      const grad = ctx.createRadialGradient(0, 0, 0, 0, 0, radius * 2);
      grad.addColorStop(0, 'rgba(96, 165, 250, 0.9)');
      grad.addColorStop(0.5, 'rgba(96, 165, 250, 0.4)');
      grad.addColorStop(1, 'rgba(96, 165, 250, 0)');
      ctx.fillStyle = grad;
      ctx.beginPath();
      ctx.arc(0, 0, radius * 2, 0, Math.PI * 2);
      ctx.fill();
      ctx.fillStyle = '#fff';
      ctx.beginPath();
      ctx.arc(0, 0, radius * 0.6, 0, Math.PI * 2);
      ctx.fill();
      ctx.restore();
    }
  }

  _drawHead(ctx, p, walkBob) {
    const bob = walkBob * 0.3;
    const headY = -32 - bob;

    ctx.fillStyle = p.skinShadow;
    ctx.fillRect(-2, -24 - bob, 4, 3);

    ctx.fillStyle = p.skinColor;
    ctx.beginPath();
    ctx.arc(0, headY, 8, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = p.skinShadow;
    ctx.beginPath();
    ctx.arc(3, headY, 8, -Math.PI / 4, Math.PI / 4);
    ctx.fill();

    ctx.fillStyle = p.hairColor;
    if (p.hairStyle === 'long') {
      ctx.beginPath();
      ctx.arc(0, headY - 1, 9, Math.PI, 0);
      ctx.fill();
      ctx.fillRect(-9, headY - 2, 3, 14);
      ctx.fillRect(6, headY - 2, 3, 14);
      ctx.fillRect(-7, headY - 2, 14, 3);
    } else {
      ctx.beginPath();
      ctx.arc(0, headY - 1, 9, Math.PI, 0);
      ctx.fill();
      ctx.fillRect(-8, headY - 3, 16, 4);
    }

    if (this.direction !== 3) {
      ctx.fillStyle = p.eyeColor;
      if (this.direction === 0) {
        ctx.fillRect(-4, headY - 1, 2, 2);
        ctx.fillRect(2, headY - 1, 2, 2);
        ctx.fillStyle = '#fff';
        ctx.fillRect(-3, headY - 1, 1, 1);
        ctx.fillRect(3, headY - 1, 1, 1);
      } else if (this.direction === 1) {
        ctx.fillRect(-4, headY - 1, 2, 2);
        ctx.fillStyle = '#fff';
        ctx.fillRect(-3, headY - 1, 1, 1);
      } else if (this.direction === 2) {
        ctx.fillRect(2, headY - 1, 2, 2);
        ctx.fillStyle = '#fff';
        ctx.fillRect(3, headY - 1, 1, 1);
      }
    }

    ctx.fillStyle = '#92400e';
    if (this.direction === 0) {
      ctx.fillRect(-1, headY + 3, 2, 1);
    }
  }

  _drawNameTag(ctx, x, y, name, isSelf) {
    if (!name) return;
    ctx.save();
    ctx.font = 'bold 12px Microsoft YaHei';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';

    const padding = 4;
    const metrics = ctx.measureText(name);
    const w = metrics.width + padding * 2;
    const h = 16;

    ctx.fillStyle = isSelf ? 'rgba(233, 69, 96, 0.8)' : 'rgba(0, 0, 0, 0.6)';
    ctx.fillRect(x - w / 2, y - h / 2, w, h);

    ctx.strokeStyle = isSelf ? '#fff' : 'rgba(255,255,255,0.3)';
    ctx.lineWidth = 1;
    ctx.strokeRect(x - w / 2, y - h / 2, w, h);

    ctx.fillStyle = '#fff';
    ctx.fillText(name, x, y);
    ctx.restore();
  }
}

window.CharacterRenderer = CharacterRenderer;
