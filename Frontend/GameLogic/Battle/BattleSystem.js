/**
 * 战斗系统 - 管理怪物、攻击、伤害显示
 * 负责怪物渲染、血条显示、点击攻击交互
 */
class BattleSystem {
  constructor(game) {
    this.game = game;
    
    // 怪物实例列表
    this.monsters = new Map(); // key=instanceID, value=MonsterData
    
    // 伤害飘字列表
    this.damageNumbers = [];
    
    // 掉落物品列表
    this.droppedItems = []; // {x, y, itemID, itemName, quantity, expireTime}
    
    // 攻击冷却（毫秒）
    this.attackCooldown = 1500;
    this.lastAttackTime = 0;
    
    // 选中的目标
    this.selectedTarget = null;
    
    // 事件回调
    this.onMonsterClick = null; // 点击怪物回调
    
    // ===== 平滑移动系统 =====
    this.smoothMoveEnabled = true; // 是否启用平滑移动
    this.lastFrameTime = 0; // 上一帧时间戳

    // ===== 战斗动画增强系统 =====
    // 受击动画：怪物被击中时闪烁、震动
    // 格式: { monsterId: { startTime, duration, intensity, isCrit } }
    this.hitAnimations = new Map();

    // 屏幕震动效果
    this.screenShake = { intensity: 0, duration: 0, startTime: 0 };

    // 死亡动画：怪物死亡时淡出、放大
    // 格式: { monsterId: { startTime, duration, originalAlpha } }
    this.deathAnimations = new Map();

    // 攻击者动作动画：玩家攻击时的挥砍动作
    // 格式: { startTime, duration, direction, attackerId }
    this.attackerAnimations = [];

    // 连击数追踪（用于显示连击）
    this.comboTracker = { count: 0, lastHitTime: 0, timeout: 2000 };

    // ===== 施法动画系统（前摇读条）=====
    // 格式: { skillId, skillName, startTime, duration, x, y, targetX, targetY }
    this.castAnimation = null;

    // ===== 技能释放特效系统（后摇）=====
    // 格式: { startTime, duration, type, x, y, skillType, skillName }
    this.skillEffects = [];

    // ===== 闪避/格挡特效增强 =====
    // 格式: { targetId, type, startTime, duration, x, y }
    this.dodgeEffects = new Map();

    // ===== 玩家BUFF状态栏 =====
    // 格式: { buffId: { name, type, stack, duration, remaining, startTime } }
    this.playerBuffs = new Map();
    // BUFF名称映射（用于显示，后续可从服务端获取）
    this.buffNameMap = {
      1: '打坐', 2: '中毒', 3: '虚弱', 4: '力量祝福', 5: '防御强化',
      6: '疾风术', 7: '眩晕', 8: '沉默', 9: '减速', 10: '无敌',
      11: '生命恢复', 12: '内力恢复', 13: '破甲', 14: '混乱', 15: '隐身',
      16: '灼烧', 17: '冰冻', 18: '石化', 19: '吸血', 20: '反弹伤害'
    };
    // BUFF颜色映射（按类型：1=增益绿, 2=减益红, 3=控制紫）
    this.buffColorMap = {
      1: '#44FF44', 2: '#FF4444', 3: '#AA44FF', 4: '#44FF44', 5: '#44FF44',
      6: '#44FF44', 7: '#AA44FF', 8: '#AA44FF', 9: '#FF4444', 10: '#00FFFF',
      11: '#44FF44', 12: '#44FF44', 13: '#FF4444', 14: '#AA44FF', 15: '#00FFFF',
      16: '#FF4444', 17: '#AA44FF', 18: '#AA44FF', 19: '#44FF44', 20: '#44FF44'
    };

    console.log('战斗系统初始化完成');
  }
  
  /**
   * 更新怪物数据（从服务端同步）
   * @param {Array} monsters - 怪物数组 [{id, baseId, name, x, y, hp, maxHp, ...}]
   */
  updateMonsters(monsters) {
    if (!monsters || !Array.isArray(monsters)) return;
    
    console.log(`🎮 BattleSystem.updateMonsters: 收到 ${monsters.length} 个怪物`);
    
    const now = Date.now();
    
    monsters.forEach(data => {
      const monster = this.monsters.get(data.id);
      if (monster) {
        // 更新已有怪物（使用平滑移动，带碰撞检测）
        const playerX = this.game?.player?.x;
        const playerY = this.game?.player?.y;
        if (playerX !== undefined && playerY !== undefined &&
            data.x === playerX && data.y === playerY) {
          // 目标位置与玩家重叠，跳过位置更新
          console.log(`🛡️ updateMonsters: 怪物${data.id}目标位置与玩家重叠，跳过`);
        } else {
          this.updateMonsterPosition(monster, data.x, data.y, now);
        }
        
        // 更新其他属性
        if (data.hp !== undefined) {
          monster.hp = data.hp;
        }
        if (data.maxHp !== undefined) monster.maxHp = data.maxHp;
        if (data.name !== undefined) monster.name = data.name;
        if (data.level !== undefined) monster.level = data.level;
        if (data.status !== undefined) monster.status = data.status;
        if (data.aiState !== undefined) monster.aiState = data.aiState;
      } else {
        // 添加新怪物（初始化平滑移动状态）
        this.monsters.set(data.id, {
          id: data.id,
          baseId: data.base_id || data.baseId,
          name: data.name || '未知怪物',
          x: data.x,
          y: data.y,
          hp: data.hp,
          maxHp: data.maxHp || data.max_hp || data.hp,
          level: data.level || 1,
          type: data.type || 0, // 0=普通, 1=精英, 2=BOSS
          status: data.status || 0, // 0=空闲, 4=死亡
          spriteId: data.sprite_id || data.spriteId || 2001,
          
          // ===== 平滑移动状态 =====
          displayX: data.x, // 渲染用位置（平滑插值）
          displayY: data.y,
          targetX: data.x,   // 目标位置（服务端真实位置）
          targetY: data.y,
          lastMoveTime: now, // 上次收到位置更新的时间
          moveSpeed: 3.0,    // 移动速度（格/秒），可根据AI状态调整
        });
      }
    });
    
    console.log(`📊 当前怪物总数: ${this.monsters.size}`);
  }
  
  /**
   * 移除怪物
   */
  removeMonster(instanceID) {
    this.monsters.delete(instanceID);
    if (this.selectedTarget === instanceID) {
      this.selectedTarget = null;
    }
  }
  
  // ===== 平滑移动系统核心方法 =====
  
  /**
   * 更新怪物目标位置（收到服务端同步时调用）
   * @param {Object} monster - 怪物对象
   * @param {number} newX - 新的X坐标
   * @param {number} newY - 新的Y坐标
   * @param {number} now - 当前时间戳
   */
  updateMonsterPosition(monster, newX, newY, now = Date.now()) {
    if (!this.smoothMoveEnabled) {
      // 如果禁用平滑移动，直接跳转
      monster.x = newX;
      monster.y = newY;
      monster.displayX = newX;
      monster.displayY = newY;
      return;
    }
    
    const oldTargetX = monster.targetX || monster.x;
    const oldTargetY = monster.targetY || monster.y;
    
    // 只有位置真正改变时才更新
    if (oldTargetX !== newX || oldTargetY !== newY) {
      // 计算实际移动距离（用于调整速度）
      const distance = Math.hypot(newX - oldTargetX, newY - oldTargetY);
      
      // 根据AI状态调整移动速度（可选）
      let speed = 3.0; // 默认速度：3格/秒
      if (monster.aiState === 4) {
        speed = 5.0; // 追击状态：更快
      } else if (monster.aiState === 5) {
        speed = 6.0; // 战斗状态：最快
      } else if (monster.aiState === 6) {
        speed = 5.5; // 逃跑状态：很快
      } else if (monster.aiState === 1 || monster.aiState === 2) {
        speed = 2.0; // 巡逻/警戒状态：较慢
      }
      
      // 更新目标位置和速度
      monster.targetX = newX;
      monster.targetY = newY;
      monster.lastMoveTime = now;
      monster.moveSpeed = speed;
      
      // 更新逻辑位置（用于碰撞检测等）
      monster.x = newX;
      monster.y = newY;
    }
  }
  
  /**
   * 每帧更新所有怪物的平滑位置（在游戏主循环中调用）
   * @param {number} deltaTime - 距上一帧的时间（毫秒）
   */
  update(deltaTime) {
    if (!this.smoothMoveEnabled) return;
    
    const now = Date.now();
    const dt = deltaTime / 1000.0; // 转换为秒
    
    this.monsters.forEach((monster, id) => {
      if (monster.status === 4) return; // 死亡怪物不移动

      // 确保有平滑移动状态字段
      if (monster.displayX === undefined) monster.displayX = monster.x;
      if (monster.displayY === undefined) monster.displayY = monster.y;
      if (monster.targetX === undefined) monster.targetX = monster.x;
      if (monster.targetY === undefined) monster.targetY = monster.y;

      // 获取玩家位置（用于碰撞检测）
      const playerX = this.game?.player?.x;
      const playerY = this.game?.player?.y;

      // 碰撞检测：如果目标位置与玩家重叠，停止插值，保持当前位置
      if (playerX !== undefined && playerY !== undefined &&
          monster.targetX === playerX && monster.targetY === playerY) {
        monster.displayX = monster.x;
        monster.displayY = monster.y;
        monster.targetX = monster.x;
        monster.targetY = monster.y;
        return;
      }

      // 计算到目标的距离
      const dx = monster.targetX - monster.displayX;
      const dy = monster.targetY - monster.displayY;
      const distance = Math.hypot(dx, dy);

      if (distance < 0.01) {
        // 已经到达目标，直接对齐（再次校验碰撞）
        if (playerX !== undefined && playerY !== undefined &&
            monster.targetX === playerX && monster.targetY === playerY) {
          // 目标与玩家重叠，回退到逻辑位置
          monster.displayX = monster.x;
          monster.displayY = monster.y;
        } else {
          monster.displayX = monster.targetX;
          monster.displayY = monster.targetY;
        }
        return;
      }

      // 计算本帧应该移动的距离
      const moveDistance = (monster.moveSpeed || 3.0) * dt;

      if (moveDistance >= distance) {
        // 本帧可以到达目标（再次校验碰撞）
        if (playerX !== undefined && playerY !== undefined &&
            monster.targetX === playerX && monster.targetY === playerY) {
          // 目标与玩家重叠，不移动
          return;
        }
        monster.displayX = monster.targetX;
        monster.displayY = monster.targetY;
      } else {
        // 线性插值移动
        const ratio = moveDistance / distance;
        let newDisplayX = monster.displayX + dx * ratio;
        let newDisplayY = monster.displayY + dy * ratio;

        // 插值过程中也检测是否进入玩家格子
        if (playerX !== undefined && playerY !== undefined) {
          const roundedX = Math.round(newDisplayX);
          const roundedY = Math.round(newDisplayY);
          if (roundedX === playerX && roundedY === playerY) {
            // 插值会进入玩家格子，停止在当前格子的边缘
            return;
          }
        }

        monster.displayX = newDisplayX;
        monster.displayY = newDisplayY;
      }
    });
  }
  
  /**
   * 获取怪物的渲染位置（使用平滑后的display坐标）
   * @param {Object} monster - 怪物对象
   * @returns {{x: number, y: number}} 渲染用的坐标
   */
  getMonsterRenderPos(monster) {
    if (this.smoothMoveEnabled && monster.displayX !== undefined) {
      return {
        x: monster.displayX,
        y: monster.displayY
      };
    }
    return {
      x: monster.x,
      y: monster.y
    };
  }
  
  /**
   * 获取指定位置的怪物（使用渲染位置进行检测）
   */
  getMonsterAt(x, y, tolerance = 0.5) {
    for (let [id, monster] of this.monsters) {
      if (monster.status === 4) continue; // 死亡的怪物
      
      // 使用平滑后的渲染位置
      const renderPos = this.getMonsterRenderPos(monster);
      
      const dx = Math.abs(renderPos.x - x);
      const dy = Math.abs(renderPos.y - y);
      
      if (dx <= tolerance && dy <= tolerance) {
        return monster;
      }
    }
    return null;
  }
  
  /**
   * 渲染所有怪物和血条
   * @param {CanvasRenderingContext2D} ctx - Canvas上下文
   * @param {number} tileSize - 格子大小
   */
  render(ctx, tileSize) {
    const camera = this.game.mapEngine?.camera;
    if (!camera) {
      console.warn('⚠️ BattleSystem.render: camera不存在');
      return;
    }
    
    // 遍历所有怪物进行渲染
    this.monsters.forEach((monster, id) => {
      if (monster.status === 4) return; // 不渲染死亡怪物
      
      // 使用平滑后的渲染位置
      const renderPos = this.getMonsterRenderPos(monster);
      
      // 计算世界坐标（MapEngine 已通过 ctx.translate(-camera.offsetX, -camera.offsetY) 处理相机）
      const worldX = renderPos.x * tileSize + tileSize / 2;
      const worldY = renderPos.y * tileSize + tileSize / 2;
      
      // 视野裁剪：转换为屏幕坐标进行判断
      const screenX = worldX - camera.offsetX;
      const screenY = worldY - camera.offsetY;
      
      if (screenX < -tileSize || screenX > ctx.canvas.width + tileSize ||
          screenY < -tileSize || screenY > ctx.canvas.height + tileSize) {
        return;
      }
      
      // 绘制怪物本体（使用平滑后的世界坐标）
      this.drawMonster(ctx, monster, worldX, worldY, tileSize);
      
      // 绘制血条（跟随平滑位置）
      this.drawHealthBar(ctx, monster, worldX, worldY, tileSize);
      
      // 绘制选中框
      if (this.selectedTarget === id) {
        this.drawSelectionBox(ctx, worldX, worldY, tileSize);
      }
    });
    
    // 渲染伤害飘字
    this.renderDamageNumbers(ctx);
    
    // 渲染掉落物品
    this.renderDroppedItems(ctx, tileSize);

    // 渲染屏幕边缘闪烁效果（BUFF持续伤害提示）
    this.renderScreenEdgeFlashes(ctx);
  }
  
  /**
   * 绘制怪物
   */
  drawMonster(ctx, monster, x, y, tileSize) {
    const size = tileSize * 0.8;
    const halfSize = size / 2;

    // 根据怪物类型选择颜色
    let color = '#FF6B6B'; // 普通 - 红色
    if (monster.type === 1) color = '#9B59B6'; // 精英 - 紫色
    if (monster.type === 2) color = '#E74C3C'; // BOSS - 深红

    // ===== 受击动画：闪烁与震动 =====
    let shakeX = 0, shakeY = 0;
    let flashColor = null;
    let alpha = 1.0;
    let scale = 1.0;

    const hitAnim = this.hitAnimations.get(monster.id);
    if (hitAnim) {
      const elapsed = Date.now() - hitAnim.startTime;
      if (elapsed < hitAnim.duration) {
        const progress = elapsed / hitAnim.duration;
        const decay = 1 - progress;
        const shakeAmount = hitAnim.intensity * decay * 4;
        shakeX = (Math.random() - 0.5) * shakeAmount * 2;
        shakeY = (Math.random() - 0.5) * shakeAmount * 2;
        if (Math.floor(elapsed / 60) % 2 === 0) {
          flashColor = hitAnim.isCrit ? '#FFEB3B' : '#FFFFFF';
        }
        if (hitAnim.isCrit) {
          scale = 1 + 0.15 * decay;
        }
      } else {
        this.hitAnimations.delete(monster.id);
      }
    }

    // ===== 死亡动画：淡出与放大 =====
    const deathAnim = this.deathAnimations.get(monster.id);
    if (deathAnim) {
      const elapsed = Date.now() - deathAnim.startTime;
      if (elapsed < deathAnim.duration) {
        const progress = elapsed / deathAnim.duration;
        alpha = 1 - progress;
        scale = 1 + progress * 0.5;
      } else {
        this.deathAnimations.delete(monster.id);
      }
    }

    ctx.save();
    ctx.globalAlpha = alpha;

    const drawX = x + shakeX;
    const drawY = y + shakeY;
    ctx.translate(drawX, drawY);
    ctx.scale(scale, scale);
    ctx.translate(-drawX, -drawY);

    // 绘制怪物身体（圆形）
    ctx.fillStyle = flashColor || color;
    ctx.strokeStyle = '#000';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(drawX, drawY, halfSize * 0.7, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();

    // 绘制眼睛
    ctx.fillStyle = '#FFF';
    ctx.beginPath();
    ctx.arc(drawX - size * 0.15, drawY - size * 0.1, size * 0.12, 0, Math.PI * 2);
    ctx.arc(drawX + size * 0.15, drawY - size * 0.1, size * 0.12, 0, Math.PI * 2);
    ctx.fill();

    // 绘制瞳孔
    ctx.fillStyle = '#000';
    ctx.beginPath();
    ctx.arc(drawX - size * 0.15, drawY - size * 0.08, size * 0.06, 0, Math.PI * 2);
    ctx.arc(drawX + size * 0.15, drawY - size * 0.08, size * 0.06, 0, Math.PI * 2);
    ctx.fill();

    // 绘制名称
    ctx.fillStyle = '#FFF';
    ctx.font = 'bold 10px Arial';
    ctx.textAlign = 'center';
    ctx.fillText(monster.name, drawX, drawY - halfSize - 5);

    ctx.restore();
  }
  
  /**
   * 绘制血条
   */
  drawHealthBar(ctx, monster, x, y, tileSize) {
    const barWidth = tileSize * 0.8;
    const barHeight = 6;
    const barX = x - barWidth / 2;
    const barY = y - tileSize * 0.5 - 18;
    
    const hpPercent = Math.max(0, monster.hp / monster.maxHp);
    
    ctx.save();
    
    // 背景（黑色半透明）
    ctx.fillStyle = 'rgba(0, 0, 0, 0.6)';
    ctx.fillRect(barX - 1, barY - 1, barWidth + 2, barHeight + 2);
    
    // 血量背景（灰色）
    ctx.fillStyle = '#555';
    ctx.fillRect(barX, barY, barWidth, barHeight);
    
    // 当前血量（根据HP百分比变色）
    let hpColor = '#27AE60'; // 绿色 > 50%
    if (hpPercent <= 0.3) {
      hpColor = '#E74C3C'; // 红色 < 30%
    } else if (hpPercent <= 0.5) {
      hpColor = '#F39C12'; // 黄色 < 50%
    }
    
    ctx.fillStyle = hpColor;
    ctx.fillRect(barX, barY, barWidth * hpPercent, barHeight);
    
    // 边框
    ctx.strokeStyle = '#000';
    ctx.lineWidth = 1;
    ctx.strokeRect(barX, barY, barWidth, barHeight);
    
    ctx.restore();
  }
  
  /**
   * 绘制选中框
   */
  drawSelectionBox(ctx, x, y, tileSize) {
    const size = tileSize * 0.85;
    
    ctx.save();
    ctx.strokeStyle = '#00FFFF';
    ctx.lineWidth = 2;
    ctx.setLineDash([5, 3]);
    ctx.strokeRect(x - size/2, y - size/2, size, size);
    ctx.restore();
  }
  
  /**
   * 处理点击事件（检查是否点击到怪物）
   * @param {number} clickX - 点击的格子X坐标
   * @param {number} clickY - 点击的格子Y坐标
   * @returns {object|null} 点击到的怪物或null
   */
  handleClick(clickX, clickY) {
    const monster = this.getMonsterAt(clickX, clickY);
    
    if (monster && monster.status !== 4) {
      // 选中目标
      this.selectedTarget = monster.id;
      
      // 触发回调
      if (this.onMonsterClick) {
        this.onMonsterClick(monster);
      }
      
      return monster;
    } else {
      // 取消选中
      this.selectedTarget = null;
    }
    
    return null;
  }
  
  /**
   * 发起攻击请求
   * @param {number} targetId - 目标怪物实例ID
   */
  async attack(targetId) {
    // 死亡状态下禁止攻击
    if (this.game.player.isDead) {
      console.log('已死亡，无法攻击');
      return false;
    }

    // 检查冷却
    const now = Date.now();
    if (now - this.lastAttackTime < this.attackCooldown) {
      console.log('攻击冷却中...');
      this.game.showFloatingText('攻击冷却中...', this.game.player.x, this.game.player.y, '#FFAA00');
      return false;
    }

    // 检查目标存在性
    const target = this.monsters.get(targetId);
    if (!target) {
      console.log('目标不存在');
      return false;
    }

    // 距离校验交由服务端处理，前端直接发送攻击请求
    // 服务端返回error_code=1(距离过远)时，Game.handleDamage会自动触发追击移动

    const player = this.game.player;

    // 设置冷却
    this.lastAttackTime = now;

    console.log(`⚔️ 发起攻击 -> 怪物 ${target.name} (${targetId})`);

    // 通过WebSocket发送攻击请求（cmd=2002 CMD_ATTACK）
    // 服务端Gateway转发到GameService，处理结果通过/internal/push异步推送回来
    if (window.GameWS && window.GameWS.isConnected) {
      window.GameWS.send(2002, {
        target_id: targetId,
        target_type: 2,    // 2=怪物
        skill_id: 0,       // 0=普通攻击
        x: Math.floor(player.x),
        y: Math.floor(player.y)
      });
    } else {
      console.warn('WebSocket未连接，无法发起攻击');
      this.game.showFloatingText('网络未连接', this.game.player.x, this.game.player.y, '#FF0000');
      return false;
    }

    // 播放攻击动画（不等待服务端响应，提升手感）
    this.playAttackAnimation(player.x, player.y, target.x, target.y);

    return true;
  }
  
  /**
   * 处理攻击结果
   */
  handleAttackResult(result, target) {
    if (!result || !result.Success) {
      return;
    }
    
    const player = this.game.player;
    
    // 更新玩家MP（使用服务端返回的值同步）
    if (result.CurrentMP !== undefined) {
      player.mp = result.CurrentMP;
      if (result.MaxMP !== undefined) {
        player.maxMp = result.MaxMP;
      }
    }
    
    // 更新怪物血量
    if (target && result.CurrentHP !== undefined) {
      target.hp = result.CurrentHP;
      if (target.hp <= 0) {
        target.status = 4; // 标记为死亡
      }
    }
    
    // 显示伤害飘字
    if (result.IsMiss) {
      this.addDamageNumber(target.x, target.y, '闪避', '#FFFFFF', false); // 白色-闪避
      this.game.showFloatingText('闪避!', target.x, target.y, '#FFFFFF');
    } else if (result.Damage > 0) {
      let color = '#FF4444'; // 红色-普通伤害
      let text = `${result.Damage}`;
      
      if (result.IsCrit) {
        color = '#FFD700'; // 金色-暴击
        text = `暴击 ${result.Damage}!`;
      } else if (result.IsBlocked) {
        color = '#888888'; // 灰色-格挡
        text = `${result.Damage} (格挡)`;
      }
      
      // 显示伤害数字
      this.addDamageNumber(target.x, target.y, text, color, result.IsCrit);
      
      // 显示浮动文字（备用）
      this.game.showFloatingText(text, target.x, target.y, color);
    }
    
    // 处理怪物死亡
    if (result.IsDead) {
      console.log(`💀 怪物 ${target.name} 被击杀!`);
      
      // 显示击杀奖励
      if (result.ExpGain > 0 || result.GoldGain > 0) {
        setTimeout(() => {
          let rewardText = '';
          if (result.ExpGain > 0) rewardText += `+${result.ExpGain} EXP `;
          if (result.GoldGain > 0) rewardText += `+${result.GoldGain} 金币`;
          this.game.showFloatingText(rewardText, player.x, player.y, '#00FF00');
          
          // 检查是否升级
          if (result.LeveledUp) {
            this.showLevelUpEffect(player.x, player.y, result.NewLevel);
            this.game.showFloatingText(
              `🎉 升级! Lv.${result.NewLevel}`, 
              player.x, 
              player.y - 1.5, 
              '#FFD700'
            );
            
            // 播放升级音效（如果有）
            if (this.game.playLevelUpSound) {
              this.game.playLevelUpSound();
            }
          }
          
          // 显示掉落物品提示
          if (result.Drops && result.Drops.length > 0) {
            this.game.showFloatingText(`获得 ${result.Drops.length} 个物品!`, 
              player.x, player.y + 1, '#FFD700');
            
            // 添加掉落物品到地面（在怪物死亡位置）
            this.addDroppedItems(target.x, target.y, result.Drops);
          }
        }, 500); // 延迟显示奖励信息
      }
      
      // 从怪物列表移除（延迟执行，让死亡动画播放完）
      setTimeout(() => {
        this.monsters.delete(target.id);
        if (this.selectedTarget === target.id) {
          this.selectedTarget = null;
        }
      }, 2000);
    }
    
    console.log(`✅ 攻击结果: 伤害=${result.Damage}, 暴击=${result.IsCrit}, 闪避=${result.IsMiss}, 死亡=${result.IsDead}`);
  }
  
  /**
   * 播放攻击动画
   */
  playAttackAnimation(fromX, fromY, toX, toY) {
    const tileSize = this.game.mapEngine?.tileSize || 48;

    // 触发角色攻击动画（卡通人形挥砍动作）
    if (this.game.characterRenderer) {
      // 根据目标位置设置朝向
      const dx = toX - fromX;
      const dy = toY - fromY;
      this.game.characterRenderer.setDirection(dx, dy);
      this.game.characterRenderer.playAttack();
    }

    // 触发技能特效（普通攻击）
    if (this.game.triggerSkillEffect) {
      this.game.triggerSkillEffect(0); // 0=普通攻击
    }

    // 播放攻击音效
    if (this.game.playSkillSound) {
      this.game.playSkillSound(0);
    }
  }
  
  /**
   * 处理服务端返回的伤害结果
   * @param {object} data - 伤害数据
   */
  handleDamageResult(data) {
    const targetId = data.target_id;
    const damage = data.damage || 0;
    const isCrit = data.is_crit || false;
    const isMiss = data.is_miss || false;
    const isBlocked = data.is_blocked || false;
    const currentHp = data.current_hp;
    const isDead = data.is_dead || false;

    // 更新玩家MP（使用服务端返回的值同步）
    if (data.current_mp !== undefined) {
      this.game.player.mp = data.current_mp;
      if (data.max_mp !== undefined) {
        this.game.player.maxMp = data.max_mp;
      }
    }

    // 更新怪物HP
    const monster = this.monsters.get(targetId);
    if (monster) {
      monster.hp = currentHp;

      if (isDead) {
        monster.status = 4; // 死亡状态

        // 延迟移除死亡怪物（3秒后）
        // 怪物复活时服务端会通过 monster_spawn 消息通知前端重新创建
        setTimeout(() => {
          this.removeMonster(targetId);
        }, 3000);

        // 仅击杀者显示掉落信息（广播消息中其他玩家也会收到，但只有击杀者应看到）
        const isKiller = data.attacker_id === this.game.player.id;
        if (isKiller) {
          if (data.exp_gain) {
            this.showExpGain(data.exp_gain, monster.x, monster.y);
          }
          if (data.drops && data.drops.length > 0) {
            this.showDrops(data.drops, monster.x, monster.y);
          }
        }
      }
    }
    
    // 显示伤害数字
    const targetPos = monster ? { x: monster.x, y: monster.y } : { x: data.target_x, y: data.target_y };
    this.addDamageNumber(targetPos.x, targetPos.y, damage, isCrit, isMiss, isBlocked);

    // ===== 闪避/格挡特效增强 =====
    if (isMiss) {
      this.triggerDodgeEffect(targetId, targetPos.x, targetPos.y);
    } else if (isBlocked && !isCrit) {
      this.triggerBlockEffect(targetId, targetPos.x, targetPos.y);
    }

    // 触发受击动画（怪物被击中时闪烁、震动）
    if (monster && !isMiss) {
      this.triggerHitAnimation(targetId, isCrit, isBlocked);
    }

    // 暴击时触发屏幕震动
    if (isCrit && this.game.effectSettings?.enableScreenShake) {
      this.triggerScreenShake(isCrit ? 8 : 3, 300);
    }

    // 死亡动画
    if (isDead && monster) {
      this.triggerDeathAnimation(targetId);
    }

    // 连击数追踪（玩家攻击时）
    if (data.attacker_id === this.game.player.id && !isMiss && damage > 0) {
      this.updateComboTracker();
    }

    // 技能攻击：显示技能名飘字 + 触发技能释放特效（后摇）
    if (data.is_skill_attack && data.skill_name) {
      const skillColor = isCrit ? '#FFD700' : '#00BFFF';
      if (this.game.showFloatingText) {
        this.game.showFloatingText(data.skill_name, targetPos.x, targetPos.y - 0.5, skillColor);
      }
      // 技能释放弹道特效
      const player = this.game.player;
      const fromX = player ? player.x : (data.attacker_x || targetPos.x);
      const fromY = player ? player.y : (data.attacker_y || targetPos.y);
      this.triggerSkillEffect(
        data.skill_type || 0,
        data.skill_name,
        fromX, fromY,
        targetPos.x, targetPos.y,
        isCrit
      );
    }

    // ===== BUFF效果飘字反馈 =====

    // 无敌免疫
    if (data.invulnerable && this.game.showFloatingText) {
      this.game.showFloatingText('免疫!', targetPos.x, targetPos.y - 1, '#00FFFF');
      this.addDamageNumber(targetPos.x, targetPos.y, 'IMMUNE', '#00FFFF', false, false);
    }

    // 反弹伤害
    if (data.reflect_damage > 0 && this.game.player) {
      const px = this.game.player.x;
      const py = this.game.player.y;
      this.addDamageNumber(px, py, `-${data.reflect_damage}`, '#FF4444', false, false);
      if (this.game.showFloatingText) {
        this.game.showFloatingText(`反弹 ${data.reflect_damage}`, px, py - 1, '#FF4444');
      }
    }

    // 吸血回复
    if (data.lifesteal_heal > 0 && this.game.player) {
      const px = this.game.player.x;
      const py = this.game.player.y;
      this.addDamageNumber(px, py, `+${data.lifesteal_heal}`, '#00FF00', false, false);
      if (this.game.showFloatingText) {
        this.game.showFloatingText(`吸血 +${data.lifesteal_heal}HP`, px, py - 1.2, '#00FF00');
      }
    }

    // 自身BUFF附加提示
    if (data.self_buff_applied > 0 && this.game.player) {
      const px = this.game.player.x;
      const py = this.game.player.y;
      if (this.game.showFloatingText) {
        this.game.showFloatingText('增益生效!', px, py - 1.3, '#00FF88');
      }
      // 记录到玩家BUFF列表（用于状态栏显示）
      this.addPlayerBuff(data.self_buff_applied);
    }

    // 目标被附加DEBUFF提示
    if (data.buff_applied > 0 && this.game.showFloatingText) {
      this.game.showFloatingText(`DEBUFF!`, targetPos.x, targetPos.y + 1, '#FF6600');
    }

    // AOE多目标伤害显示（aoe_targets包含被波及的怪物列表）
    if (data.aoe_targets && Array.isArray(data.aoe_targets) && data.aoe_targets.length > 0) {
      data.aoe_targets.forEach(aoeTarget => {
        const aoeMonster = this.monsters.get(aoeTarget.target_id);
        const aoeX = aoeMonster ? aoeMonster.x : (aoeTarget.x || targetPos.x);
        const aoeY = aoeMonster ? aoeMonster.y : (aoeTarget.y || targetPos.y);

        // 更新AOE怪物血量
        if (aoeMonster) {
          aoeMonster.hp = aoeTarget.current_hp;
          if (aoeTarget.is_dead) {
            aoeMonster.status = 4;
            setTimeout(() => this.removeMonster(aoeTarget.target_id), 3000);
            this.triggerDeathAnimation(aoeTarget.target_id);
          }
        }

        // AOE溅射伤害飘字（橙色，比主伤害小）
        this.addDamageNumber(aoeX, aoeY, String(aoeTarget.damage), '#FF8C00', false);

        // AOE受击动画
        if (aoeMonster && !aoeTarget.is_dead) {
          this.triggerHitAnimation(aoeTarget.target_id, false, false);
        }

        // AOE击杀提示
        if (aoeTarget.is_dead && this.game.showFloatingText) {
          this.game.showFloatingText(`AOE击杀 ${aoeTarget.name}`, aoeX, aoeY - 1, '#FF4444');
        }
      });

      // AOE范围特效提示
      if (this.game.showFloatingText) {
        this.game.showFloatingText(
          `AOE! x${data.aoe_targets.length}`, targetPos.x, targetPos.y + 1, '#FF6600'
        );
      }
    }

    // 触发受击特效
    if (this.game.triggerHitEffect) {
      this.game.triggerHitEffect(isCrit, isBlocked, isMiss,
        targetPos.x * (this.game.mapEngine?.tileSize || 48),
        targetPos.y * (this.game.mapEngine?.tileSize || 48)
      );
    }
    
    // 播放受击音效
    if (this.game.playHitSound) {
      this.game.playHitSound(isCrit, isBlocked, isMiss);
    }
  }

  /**
   * 触发受击动画
   */
  triggerHitAnimation(monsterId, isCrit, isBlocked) {
    const intensity = isCrit ? 1.0 : (isBlocked ? 0.3 : 0.6);
    this.hitAnimations.set(monsterId, {
      startTime: Date.now(),
      duration: isCrit ? 400 : 250,
      intensity: intensity,
      isCrit: isCrit
    });
  }

  /**
   * 触发屏幕震动
   */
  triggerScreenShake(intensity, duration) {
    this.screenShake = {
      intensity: intensity,
      duration: duration,
      startTime: Date.now()
    };
  }

  /**
   * 触发死亡动画
   */
  // ========== BUFF Tick效果处理 ==========

  /**
   * 处理BUFF定时Tick消息（持续伤害/恢复效果）
   * @param {Object} data - 服务端推送的buff_tick数据
   * {
   *   target_id: 目标ID,
   *   target_type: 1=玩家, 2=怪物,
   *   hp_change: HP变化（正=恢复, 负=伤害）,
   *   mp_change: MP变化（正=恢复, 负=消耗）,
   *   tick_type: 'buff'
   * }
   */
  handleBuffTick(data) {
    const targetId = data.target_id;
    const targetType = data.target_type;
    const hpChange = data.hp_change || 0;
    const mpChange = data.mp_change || 0;

    if (hpChange === 0 && mpChange === 0) return;

    let x = 0, y = 0;
    let targetName = '未知';

    if (targetType === 2) { // 怪物
      const monster = this.monsters.get(targetId);
      if (monster) {
        x = monster.x;
        y = monster.y;
        targetName = monster.name || '怪物';
        // 更新怪物HP
        if (hpChange !== 0) {
          monster.hp = Math.max(0, (monster.hp || 0) + hpChange);
          // 检查是否被BUFF致死
          if (monster.hp <= 0) {
            monster.status = 4; // 死亡状态
            setTimeout(() => this.removeMonster(targetId), 3000);
            this.triggerDeathAnimation(targetId);
            // 显示击杀提示
            if (this.game.showFloatingText) {
              this.game.showFloatingText(`BUFF击杀 ${targetName}`, x, y - 1.5, '#FF00FF');
            }
          } else {
            // 触发受击动画
            this.triggerHitAnimation(targetId, false, false);
          }
        }
      }
    } else if (targetType === 1) { // 玩家
      const player = this.game.player;
      if (player && player.id === targetId) {
        x = player.x;
        y = player.y;
        targetName = player.roleName || '玩家';
      }
    }

    // HP飘字显示
    if (hpChange !== 0 && (x !== 0 || y !== 0)) {
      const text = hpChange > 0 ? `+${hpChange} HP` : `${hpChange} HP`;
      const color = hpChange > 0 ? '#00FF00' : '#FF00FF'; // 绿色=恢复，紫色=持续伤害
      
      this.addDamageNumber(x, y, text, color, false);
      
      if (this.game.showFloatingText) {
        const effectText = hpChange > 0 ? '回血' : '持续伤害';
        this.game.showFloatingText(`${effectText} ${hpChange}HP`, x, y - 1.2, color);
      }

      // 持续伤害时添加屏幕边缘红色闪烁提示（仅对自身）
      if (hpChange < 0 && data.target_id === this.game?.player?.id) {
        this.addScreenEdgeFlash('#FF0000', 300); // 红色闪烁300ms
      }
    }

    // MP飘字显示
    if (mpChange !== 0 && (x !== 0 || y !== 0)) {
      const text = mpChange > 0 ? `+${mpChange} MP` : `${mpChange} MP`;
      const color = mpChange > 0 ? '#00BFFF' : '#FF6600'; // 蓝色=回蓝，橙色=耗蓝
      
      this.addDamageNumber(x, y + 0.5, text, color, false); // MP飘字稍微偏下，避免重叠
      
      if (this.game.showFloatingText) {
        const effectText = mpChange > 0 ? '回蓝' : '消耗MP';
        this.game.showFloatingText(`${effectText} ${mpChange}MP`, x, y - 0.8, color);
      }
    }

    console.log(`[BUFF-TICK] ${targetName}(ID:${targetId}) HP:${hpChange} MP:${mpChange}`);
  }

  /**
   * 屏幕边缘闪烁效果（用于中毒等持续伤害提示）
   * @param {string} color - 闪烁颜色
   * @param {number} duration - 持续时间(毫秒)
   */
  addScreenEdgeFlash(color = '#FF0000', duration = 300) {
    if (!this.screenFlashes) {
      this.screenFlashes = [];
    }
    
    this.screenFlashes.push({
      color,
      startTime: Date.now(),
      duration,
      alpha: 0.3
    });
  }

  /**
   * 渲染屏幕边缘闪烁效果
   */
  renderScreenEdgeFlashes(ctx) {
    if (!this.screenFlashes || this.screenFlashes.length === 0) return;

    const now = Date.now();
    const width = ctx.canvas.width;
    const height = ctx.canvas.height;

    this.screenFlashes = this.screenFlashes.filter(flash => {
      const elapsed = now - flash.startTime;
      if (elapsed > flash.duration) return false;

      const progress = elapsed / flash.duration;
      const alpha = flash.alpha * (1 - progress); // 渐隐

      // 绘制四边半透明边框
      ctx.save();
      ctx.globalAlpha = alpha;
      ctx.fillStyle = flash.color;
      
      const borderSize = 20; // 边框宽度

      // 上边框
      ctx.fillRect(0, 0, width, borderSize);
      // 下边框
      ctx.fillRect(0, height - borderSize, width, borderSize);
      // 左边框
      ctx.fillRect(0, 0, borderSize, height);
      // 右边框
      ctx.fillRect(width - borderSize, 0, borderSize, height);

      ctx.restore();
      return true;
    });
  }

  triggerDeathAnimation(monsterId) {
    this.deathAnimations.set(monsterId, {
      startTime: Date.now(),
      duration: 1500,
      originalAlpha: 1.0
    });
  }

  // ========== 施法前摇动画 ==========

  /**
   * 开始施法动画（前摇读条）
   * @param {number} skillId - 技能ID
   * @param {string} skillName - 技能名称
   * @param {number} castTime - 施法时间(毫秒)
   * @param {number} x - 施法者X
   * @param {number} y - 施法者Y
   * @param {number} targetX - 目标X
   * @param {number} targetY - 目标Y
   */
  startCastAnimation(skillId, skillName, castTime, x, y, targetX, targetY) {
    if (castTime <= 0) return; // 无施法时间则不显示

    // 取消之前的施法（同一时间只能施放一个技能）
    if (this.castAnimation) {
      this.finishCastEarly();
    }

    this.castAnimation = {
      skillId,
      skillName,
      startTime: Date.now(),
      duration: castTime,
      x, y,
      targetX, targetY,
      cancelled: false
    };

    // 显示施法提示飘字
    if (this.game.showFloatingText) {
      this.game.showFloatingText(`施法: ${skillName}`, x, y - 1.2, '#00BFFF');
    }
  }

  /**
   * 检查施法是否完成
   * @returns {boolean} true=施法完成/无施法, false=正在施法中
   */
  isCasting() {
    if (!this.castAnimation) return false;
    const elapsed = Date.now() - this.castAnimation.startTime;
    return elapsed < this.castAnimation.duration && !this.castAnimation.cancelled;
  }

  /**
   * 提前取消施法（移动时调用）
   */
  cancelCast() {
    if (this.castAnimation) {
      this.castAnimation.cancelled = true;
      if (this.game.showFloatingText) {
        this.game.showFloatingText('施法中断', this.castAnimation.x, this.castAnimation.y - 1, '#FF8800');
      }
      this.castAnimation = null;
    }
  }

  /**
   * 施法完成（正常结束）
   */
  finishCastEarly() {
    this.castAnimation = null;
  }

  // ========== 技能释放特效（后摇）==========

  /**
   * 触发技能释放特效
   * @param {number} skillType - 技能类型(1-9)
   * @param {string} skillName - 技能名称
   * @param {number} fromX - 起点X
   * @param {number} fromY - 起点Y
   * @param {number} toX - 终点X
   * @param {number} toY - 终点Y
   * @param {boolean} isCrit - 是否暴击
   */
  triggerSkillEffect(skillType, skillName, fromX, fromY, toX, toY, isCrit = false) {
    const effectColors = {
      1: '#4CAF50', // 内功 - 绿色
      2: '#FF5722', // 外功 - 橙红
      3: '#00BCD4', // 身法 - 青色
      4: '#9C27B0', // 护体 - 紫色
      5: '#FF9800', // 拳法 - 金橙
      6: '#E91E63', // 剑法 - 粉红
      7: '#F44336', // 刀法 - 红色
      8: '#3F51B5', // 枪法 - 靛蓝
      9: '#795548', // 斧法 - 棕色
    };
    const color = effectColors[skillType] || '#FFFFFF';

    this.skillEffects.push({
      startTime: Date.now(),
      duration: isCrit ? 500 : 350,
      type: skillType,
      fromX, fromY, toX, toY,
      skillName,
      color,
      isCrit
    });
  }

  // ========== 闪避/格挡特效增强 ==========

  /**
   * 触发闪避特效
   */
  triggerDodgeEffect(targetId, x, y) {
    this.dodgeEffects.set(targetId, {
      type: 'dodge',
      startTime: Date.now(),
      duration: 600,
      x, y
    });
  }

  /**
   * 触发格挡特效
   */
  triggerBlockEffect(targetId, x, y) {
    this.dodgeEffects.set(targetId, {
      type: 'block',
      startTime: Date.now(),
      duration: 400,
      x, y
    });
  }

  /**
   * 更新连击追踪
   */
  updateComboTracker() {
    const now = Date.now();
    if (now - this.comboTracker.lastHitTime < this.comboTracker.timeout) {
      this.comboTracker.count++;
    } else {
      this.comboTracker.count = 1;
    }
    this.comboTracker.lastHitTime = now;

    // 连击数>=3时显示连击提示
    if (this.comboTracker.count >= 3) {
      const player = this.game.player;
      const color = this.comboTracker.count >= 10 ? '#FF1493' : '#FFD700';
      this.addDamageNumber(player.x, player.y - 1.5, `${this.comboTracker.count} 连击!`, color, true);
    }
  }

  /**
   * 获取屏幕震动偏移
   */
  getScreenShakeOffset() {
    if (this.screenShake.intensity === 0) return { x: 0, y: 0 };

    const elapsed = Date.now() - this.screenShake.startTime;
    if (elapsed > this.screenShake.duration) {
      this.screenShake.intensity = 0;
      return { x: 0, y: 0 };
    }

    const progress = elapsed / this.screenShake.duration;
    const decay = 1 - progress;
    const intensity = this.screenShake.intensity * decay;

    return {
      x: (Math.random() - 0.5) * intensity * 2,
      y: (Math.random() - 0.5) * intensity * 2
    };
  }
  
  /**
   * 添加伤害飘字
   * @param {number} x - 世界坐标X
   * @param {number} y - 世界坐标Y
   * @param {string|number} text - 显示的文本或伤害数值
   * @param {string} color - 文本颜色
   * @param {boolean} isCrit - 是否暴击（放大显示）
   */
  addDamageNumber(x, y, text, color = '#FF4444', isCrit = false) {
    // 兼容旧接口：如果传入的是布尔值，转换为新的参数格式
    if (typeof text === 'number' || (typeof text === 'string' && !isNaN(text))) {
      // 旧接口格式: addDamageNumber(x, y, damage, isCrit, isMiss, isBlocked)
      const damage = text;
      const oldIsCrit = color;
      const oldIsMiss = isCrit;
      
      let finalColor = '#FF4444'; // 默认红色
      let finalText = `${damage}`;
      
      if (oldIsMiss) {
        finalColor = '#888888';
        finalText = '闪避';
      } else if (oldIsCrit) {
        finalColor = '#FFD700';
        finalText = `暴击 ${damage}!`;
      } else if (isCrit !== false && arguments[5]) { // isBlocked
        finalColor = '#AAAAAA';
        finalText = `${damage} (格挡)`;
      }
      
      this.damageNumbers.push({
        x: x,
        y: y,
        text: finalText,
        color: finalColor,
        isCrit: oldIsCrit,
        startTime: Date.now(),
        duration: 1200, // 显示时长1.2秒
        offsetY: 0,
        alpha: 1,
        scale: oldIsCrit ? 1.5 : 1.0 // 暴击时放大
      });
    } else {
      // 新接口格式: addDamageNumber(x, y, text, color, isCrit)
      this.damageNumbers.push({
        x: x,
        y: y,
        text: String(text),
        color: color,
        isCrit: isCrit,
        startTime: Date.now(),
        duration: 1200,
        offsetY: 0,
        alpha: 1,
        scale: isCrit ? 1.5 : 1.0
      });
    }
  }
  
  /**
   * 渲染所有伤害飘字
   */
  renderDamageNumbers(ctx) {
    const now = Date.now();
    const tileSize = this.game.mapEngine?.tileSize || 48;
    
    // 过滤过期的飘字并更新位置
    this.damageNumbers = this.damageNumbers.filter(d => {
      const elapsed = now - d.startTime;
      if (elapsed > d.duration) return false;
      
      // 计算动画进度
      const progress = elapsed / d.duration;
      d.offsetY = -progress * 40; // 向上飘动40像素
      d.alpha = 1 - progress;      // 渐隐
      
      return true;
    });
    
    // 绘制每个飘字（使用世界坐标，与玩家/怪物对齐）
    this.damageNumbers.forEach(d => {
      const screenX = d.x * tileSize + tileSize/2;
      const screenY = d.y * tileSize + tileSize/2 + d.offsetY;
      
      ctx.save();
      ctx.globalAlpha = d.alpha;
      
      // 应用缩放（暴击时放大）
      if (d.scale && d.scale !== 1) {
        ctx.translate(screenX, screenY);
        ctx.scale(d.scale, d.scale);
        ctx.translate(-screenX, -screenY);
      }
      
      // 根据类型设置字体样式
      const fontSize = d.isCrit ? 22 : 18;
      ctx.font = `bold ${fontSize}px Arial`;
      ctx.fillStyle = d.color;
      ctx.textAlign = 'center';
      
      // 添加文字阴影/描边效果
      ctx.strokeStyle = 'rgba(0, 0, 0, 0.8)';
      ctx.lineWidth = 3;
      ctx.strokeText(d.text, screenX, screenY);
      ctx.fillText(d.text, screenX, screenY);
      
      ctx.restore();
    });
  }
  
  /**
   * 显示经验获得
   */
  showExpGain(exp, x, y) {
    this.addDamageNumber(x, y - 1, `+${exp} EXP`, false, false, false);
  }
  
  /**
   * 显示物品掉落
   */
  showDrops(drops, x, y) {
    // 显示掉落飘字提示
    drops.forEach((itemId, index) => {
      setTimeout(() => {
        this.addDamageNumber(x, y - 1 - index * 0.5, `+物品#${itemId}`, '#FFD700', false);
      }, index * 200);
    });

    // 添加掉落物到地面（可拾取）
    this.addDroppedItems(x, y, drops);
  }
  
  /**
   * 添加掉落物品到地面
   */
  addDroppedItems(x, y, itemIDs) {
    // 简单的物品名称映射（实际应该从配置文件读取）
    const itemNames = {
      1001: '铁剑',
      1002: '皮甲',
      1003: '生命药水',
      1004: '魔法药水',
      2001: '狼牙',
      2002: '兔肉',
      2003: '鸡毛'
    };
    
    itemIDs.forEach((itemID, index) => {
      // 在怪物周围随机散落（稍微偏移位置）
      const offsetX = (Math.random() - 0.5) * 2;
      const offsetY = (Math.random() - 0.5) * 2;
      
      this.droppedItems.push({
        x: x + offsetX,
        y: y + offsetY,
        itemID: itemID,
        itemName: itemNames[itemID] || `物品#${itemID}`,
        quantity: 1,
        expireTime: Date.now() + 60000, // 60秒后消失
        spawnTime: Date.now()
      });
    });
    
    console.log(`💎 地面新增 ${itemIDs.length} 个掉落物品`);
  }
  
  /**
   * 渲染掉落物品
   */
  renderDroppedItems(ctx, tileSize) {
    const camera = this.game.mapEngine?.camera;
    const now = Date.now();
    
    // 过滤过期物品并更新动画
    this.droppedItems = this.droppedItems.filter(item => {
      if (now > item.expireTime) {
        return false; // 过期移除
      }
      return true;
    });
    
    // 绘制每个掉落物品（使用世界坐标，与玩家/怪物对齐）
    this.droppedItems.forEach(item => {
      const worldX = item.x * tileSize + tileSize/2;
      const worldY = item.y * tileSize + tileSize/2;
      
      // 视野裁剪：转换为屏幕坐标进行判断
      const screenX = worldX - (camera?.offsetX || 0);
      const screenY = worldY - (camera?.offsetY || 0);
      
      if (screenX < -20 || screenX > ctx.canvas.width + 20 ||
          screenY < -20 || screenY > ctx.canvas.height + 20) {
        return;
      }
      
      ctx.save();
      
      // 物品发光效果（呼吸动画）
      const age = now - item.spawnTime;
      const glow = Math.sin(age / 300) * 0.3 + 0.7; // 呼吸效果
      
      // 绘制光晕（使用世界坐标）
      ctx.beginPath();
      ctx.arc(worldX, worldY, 12, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(255, 215, 0, ${0.3 * glow})`;
      ctx.fill();
      
      // 绘制物品图标（简单的宝箱形状）
      ctx.fillStyle = '#FFD700'; // 金色
      ctx.strokeStyle = '#B8860B';
      ctx.lineWidth = 2;
      
      // 宝箱主体
      ctx.fillRect(worldX - 8, worldY - 6, 16, 12);
      ctx.strokeRect(worldX - 8, worldY - 6, 16, 12);
      
      // 宝箱装饰线
      ctx.beginPath();
      ctx.moveTo(worldX - 8, worldY);
      ctx.lineTo(worldX + 8, worldY);
      ctx.stroke();
      
      // 锁扣
      ctx.fillStyle = '#B8860B';
      ctx.fillRect(worldX - 3, worldY - 3, 6, 6);
      
      // 显示物品名称（鼠标悬停时显示，这里简化为始终显示）
      if (glow > 0.9) { // 闪烁时显示名称
        ctx.font = '10px Arial';
        ctx.fillStyle = '#FFFFFF';
        ctx.textAlign = 'center';
        ctx.fillText(item.itemName, worldX, worldY - 15);
      }

      ctx.restore();
    });

    // ===== 渲染施法前摇读条 =====
    this.renderCastBar(ctx, tileSize);

    // ===== 渲染技能释放特效 =====
    this.renderSkillEffects(ctx, tileSize);

    // ===== 渲染闪避/格挡特效 =====
    this.renderDodgeEffects(ctx, tileSize);

    // ===== 渲染玩家BUFF状态栏 =====
    this.renderPlayerBuffBar(ctx, tileSize);
  }

  /**
   * 渲染施法前摇读条（显示在玩家头顶）
   */
  renderCastBar(ctx, tileSize) {
    if (!this.castAnimation) return;

    const cast = this.castAnimation;
    const elapsed = Date.now() - cast.startTime;

    // 施法完成或取消后清除
    if (elapsed >= cast.duration || cast.cancelled) {
      this.castAnimation = null;
      return;
    }

    const progress = Math.min(1, elapsed / cast.duration);
    const worldX = cast.x * tileSize + tileSize / 2;
    const worldY = cast.y * tileSize; // 玩家头顶位置

    ctx.save();

    // 背景条
    const barWidth = tileSize * 1.2;
    const barHeight = 6;
    const barX = worldX - barWidth / 2;
    const barY = worldY - tileSize * 0.9 - 20;

    // 外框（深色背景）
    ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
    ctx.fillRect(barX - 1, barY - 1, barWidth + 2, barHeight + 2);

    // 进度条（渐变色：蓝→青→白）
    const gradient = ctx.createLinearGradient(barX, barY, barX + barWidth * progress, barY);
    gradient.addColorStop(0, '#0066FF');
    gradient.addColorStop(0.5, '#00CCFF');
    gradient.addColorStop(1, '#00FFFF');
    ctx.fillStyle = gradient;
    ctx.fillRect(barX, barY, barWidth * progress, barHeight);

    // 边框
    ctx.strokeStyle = '#88CCFF';
    ctx.lineWidth = 1;
    ctx.strokeRect(barX, barY, barWidth, barHeight);

    // 技能名称
    ctx.font = 'bold 10px Arial';
    ctx.fillStyle = '#FFFFFF';
    ctx.textAlign = 'center';
    ctx.shadowColor = '#000';
    ctx.shadowBlur = 3;
    ctx.fillText(cast.skillName || '施法中...', worldX, barY - 4);
    ctx.shadowBlur = 0;

    // 剩余时间提示
    const remainingMs = cast.duration - elapsed;
    if (remainingMs > 500) {
      const remainingSec = (remainingMs / 1000).toFixed(1);
      ctx.font = '9px Arial';
      ctx.fillStyle = '#AAEEFF';
      ctx.fillText(`${remainingSec}s`, worldX, barY + barHeight + 12);
    }

    // 连接线到目标（虚线指示方向）
    if (cast.targetX !== undefined && cast.targetY !== undefined) {
      const targetScreenX = cast.targetX * tileSize + tileSize / 2;
      const targetScreenY = cast.targetY * tileSize + tileSize / 2;
      ctx.setLineDash([4, 4]);
      ctx.strokeStyle = `rgba(0, 180, 255, ${0.3 * (1 - progress)})`;
      ctx.lineWidth = 1;
      ctx.beginPath();
      ctx.moveTo(worldX, worldY);
      ctx.lineTo(targetScreenX, targetScreenY);
      ctx.stroke();
      ctx.setLineDash([]);
    }

    ctx.restore();
  }

  /**
   * 渲染技能释放特效（从施法者到目标的轨迹/冲击波）
   */
  renderSkillEffects(ctx, tileSize) {
    if (this.skillEffects.length === 0) return;

    const now = Date.now();
    this.skillEffects = this.skillEffects.filter(effect => {
      const elapsed = now - effect.startTime;
      if (elapsed >= effect.duration) return false;

      const progress = elapsed / effect.duration;
      const alpha = 1 - progress;

      ctx.save();
      ctx.globalAlpha = alpha;

      // 计算当前插值位置（从起点到终点）
      const currentX = effect.fromX + (effect.toX - effect.fromX) * progress;
      const currentY = effect.fromY + (effect.toY - effect.fromY) * progress;

      const screenX = currentX * tileSize + tileSize / 2;
      const screenY = currentY * tileSize + tileSize / 2;

      if (effect.isCrit) {
        // 暴击：大范围爆炸光环+星芒
        const radius = progress * tileSize * 1.5;
        const grad = ctx.createRadialGradient(screenX, screenY, 0, screenX, screenY, radius);
        grad.addColorStop(0, `${effect.color}FF`);
        grad.addColorStop(0.5, `${effect.color}88`);
        grad.addColorStop(1, `${effect.color}00`);
        ctx.fillStyle = grad;
        ctx.beginPath();
        ctx.arc(screenX, screenY, radius, 0, Math.PI * 2);
        ctx.fill();

        // 暴击星芒射线
        ctx.strokeStyle = '#FFD700';
        ctx.lineWidth = 2 * alpha;
        for (let i = 0; i < 6; i++) {
          const angle = (i / 6) * Math.PI * 2 + progress * Math.PI;
          const len = radius * 0.8;
          ctx.beginPath();
          ctx.moveTo(screenX, screenY);
          ctx.lineTo(screenX + Math.cos(angle) * len, screenY + Math.sin(angle) * len);
          ctx.stroke();
        }
      } else {
        // 普通技能：根据类型绘制不同形状的弹道/冲击
        const size = effect.type >= 5 ? tileSize * 0.4 : tileSize * 0.25;

        if (effect.type === 5 || effect.type === 9) {
          // 拳法/斧法：圆形冲击波
          ctx.fillStyle = effect.color;
          ctx.globalAlpha = alpha * 0.5;
          ctx.beginPath();
          ctx.arc(screenX, screenY, size * progress, 0, Math.PI * 2);
          ctx.fill();
        } else if (effect.type === 6 || effect.type === 7) {
          // 剑法/刀法：新月形斩击弧线
          ctx.strokeStyle = effect.color;
          ctx.lineWidth = 3 * alpha;
          ctx.globalAlpha = alpha;
          ctx.beginPath();
          ctx.arc(screenX, screenY, size, -Math.PI * 0.6, Math.PI * 0.6, false);
          ctx.stroke();
        } else if (effect.type === 8) {
          // 枪法：直线穿刺
          ctx.strokeStyle = effect.color;
          ctx.lineWidth = 2 * alpha;
          ctx.globalAlpha = alpha;
          const dx = effect.toX - effect.fromX;
          const dy = effect.toY - effect.fromY;
          const dist = Math.hypot(dx, dy) || 1;
          ctx.beginPath();
          ctx.moveTo(screenX - (dx / dist) * size, screenY - (dy / dist) * size);
          ctx.lineTo(screenX + (dx / dist) * size, screenY + (dy / dist) * size);
          ctx.stroke();
        } else {
          // 内功/外功等：光球
          const grad = ctx.createRadialGradient(screenX, screenY, 0, screenX, screenY, size);
          grad.addColorStop(0, `${effect.color}FF`);
          grad.addColorStop(1, `${effect.color}00`);
          ctx.fillStyle = grad;
          ctx.beginPath();
          ctx.arc(screenX, screenY, size, 0, Math.PI * 2);
          ctx.fill();
        }
      }

      ctx.restore();
      return true;
    });
  }

  /**
   * 渲染闪避/格挡特效增强
   */
  renderDodgeEffects(ctx, tileSize) {
    if (this.dodgeEffects.size === 0) return;

    const now = Date.now();

    this.dodgeEffects.forEach((effect, targetId) => {
      const elapsed = now - effect.startTime;
      if (elapsed >= effect.duration) {
        this.dodgeEffects.delete(targetId);
        return;
      }

      const progress = elapsed / effect.duration;
      const alpha = 1 - progress;
      const worldX = effect.x * tileSize + tileSize / 2;
      const worldY = effect.y * tileSize + tileSize / 2;

      ctx.save();
      ctx.globalAlpha = alpha;

      if (effect.type === 'dodge') {
        // 闪避：残影分身效果
        const offsetAngle = ((targetId * 137.5) % 360) * Math.PI / 180;
        const offsetDist = 15 * Math.sin(progress * Math.PI);

        for (let i = 1; i <= 2; i++) {
          const ghostX = worldX + Math.cos(offsetAngle + i * 0.8) * offsetDist * i * 0.6;
          const ghostY = worldY + Math.sin(offsetAngle + i * 0.8) * offsetDist * i * 0.6;
          ctx.globalAlpha = alpha * (0.3 / i);
          ctx.font = 'bold 14px Arial';
          ctx.textAlign = 'center';
          ctx.fillText('MISS', ghostX, ghostY);
        }

        // 主飘字（向上飘+轻微摇摆）
        ctx.globalAlpha = alpha;
        ctx.translate(worldX, worldY - 10 - progress * 20);
        ctx.rotate(Math.sin(progress * Math.PI * 2) * 0.2);
        ctx.font = 'bold 16px Arial';
        ctx.fillStyle = '#EEEEEE';
        ctx.strokeStyle = '#666666';
        ctx.lineWidth = 2;
        ctx.strokeText('闪避!', 0, 0);
        ctx.fillText('闪避!', 0, 0);

      } else if (effect.type === 'block') {
        // 格挡：盾牌光盾效果
        const shieldSize = tileSize * 0.5 * (1 + progress * 0.3);

        ctx.strokeStyle = '#AAAAAA';
        ctx.lineWidth = 3 * alpha;
        ctx.fillStyle = `rgba(200, 200, 200, ${alpha * 0.2})`;
        ctx.beginPath();
        ctx.arc(worldX, worldY, shieldSize, Math.PI * 0.75, Math.PI * 2.25, false);
        ctx.closePath();
        ctx.fill();
        ctx.stroke();

        ctx.font = 'bold 14px Arial';
        ctx.fillStyle = '#CCCCCC';
        ctx.textAlign = 'center';
        ctx.fillText('格挡', worldX, worldY - shieldSize - 5);
      }

      ctx.restore();
    });
  }

  // ========== 玩家BUFF状态栏 ==========

  /**
   * 添加/刷新玩家自身BUFF（技能释放时触发）
   */
  addPlayerBuff(buffId) {
    const name = this.buffNameMap[buffId] || `BUFF${buffId}`;
    const color = this.buffColorMap[buffId] || '#FFFFFF';
    const existing = this.playerBuffs.get(buffId);
    if (existing) {
      // 已存在则刷新时间
      existing.startTime = Date.now();
      existing.stack = Math.min((existing.stack || 1) + 1, 5);
    } else {
      // 新增BUFF
      this.playerBuffs.set(buffId, {
        id: buffId,
        name: name,
        type: buffId <= 4 ? 1 : (buffId <= 9 ? 2 : (buffId <= 15 ? 3 : 1)),
        stack: 1,
        duration: 300,
        startTime: Date.now(),
        color: color
      });
    }
  }

  /**
   * 移除玩家BUFF
   */
  removePlayerBuff(buffId) {
    this.playerBuffs.delete(buffId);
  }

  /**
   * 清除所有过期BUFF
   */
  cleanupExpiredPlayerBuffs() {
    const now = Date.now();
    for (const [buffId, buff] of this.playerBuffs) {
      if (buff.duration > 0) {
        const elapsed = (now - buff.startTime) / 1000;
        if (elapsed >= buff.duration) {
          this.playerBuffs.delete(buffId);
        }
      }
    }
  }

  /**
   * 渲染玩家BUFF状态栏（屏幕左下角）
   */
  renderPlayerBuffBar(ctx, tileSize) {
    this.cleanupExpiredPlayerBuffs();

    if (this.playerBuffs.size === 0) return;

    const now = Date.now();
    const buffs = Array.from(this.playerBuffs.values());
    const iconSize = 28;
    const gap = 4;
    const startX = 10;
    const startY = ctx.canvas.height - 120;

    ctx.save();

    for (let i = 0; i < buffs.length; i++) {
      const buff = buffs[i];
      const x = startX + i * (iconSize + gap);
      const y = startY;

      // 剩余时间计算
      let remainingSec = '';
      let progress = 1;
      if (buff.duration > 0) {
        const elapsed = (now - buff.startTime) / 1000;
        const remaining = buff.duration - elapsed;
        progress = Math.max(0, remaining / buff.duration);
        remainingSec = remaining > 0 ? Math.ceil(remaining) + 's' : '0s';
      } else {
        remainingSec = '∞';
      }

      // 进度环背景
      ctx.beginPath();
      ctx.arc(x + iconSize / 2, y + iconSize / 2, iconSize / 2 + 1, 0, Math.PI * 2);
      ctx.strokeStyle = 'rgba(0,0,0,0.6)';
      ctx.lineWidth = 2;
      ctx.stroke();

      // 进度环（剩余时间）
      if (progress < 1 && progress > 0) {
        ctx.beginPath();
        ctx.arc(x + iconSize / 2, y + iconSize / 2, iconSize / 2 + 1, -Math.PI / 2, -Math.PI / 2 + Math.PI * 2 * progress);
        ctx.strokeStyle = buff.color;
        ctx.lineWidth = 2;
        ctx.stroke();
      }

      // BUFF图标背景（带呼吸效果）
      const bgAlpha = 0.3 + Math.sin(now / 500 + i) * 0.08;
      ctx.globalAlpha = bgAlpha;
      ctx.fillStyle = buff.color;
      ctx.fillRect(x, y, iconSize, iconSize);
      ctx.globalAlpha = 1;

      // 边框
      ctx.strokeStyle = buff.color;
      ctx.lineWidth = 1.5;
      ctx.strokeRect(x, y, iconSize, iconSize);

      // 名称缩写
      ctx.font = 'bold 9px Arial';
      ctx.fillStyle = '#FFFFFF';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      const shortName = buff.name.length > 2 ? buff.name.substring(0, 2) : buff.name;
      ctx.fillText(shortName, x + iconSize / 2, y + iconSize / 2 - 3);

      // 叠加层数
      if (buff.stack > 1) {
        ctx.font = 'bold 8px Arial';
        ctx.fillStyle = '#FFD700';
        ctx.fillText(`${buff.stack}`, x + iconSize - 6, y + 8);
      }

      // 剩余时间
      ctx.font = '8px Arial';
      
      // 低时间警告：剩余时间<5秒时红色闪烁
      if (buff.duration > 0 && remaining > 0 && remaining <= 5) {
        const flashAlpha = 0.5 + Math.sin(now / 150) * 0.3;
        ctx.fillStyle = `rgba(255, 68, 68, ${flashAlpha})`;
        
        // 红色闪烁边框
        ctx.strokeStyle = `rgba(255, 68, 68, ${flashAlpha})`;
        ctx.lineWidth = 2;
        ctx.strokeRect(x - 1, y - 1, iconSize + 2, iconSize + 2);
      } else {
        ctx.fillStyle = '#CCCCCC';
      }
      
      ctx.textBaseline = 'bottom';
      ctx.fillText(remainingSec, x + iconSize / 2, y + iconSize - 1);
    }

    // 底部标签（根据BUFF类型显示不同颜色）
    const buffTypeCount = {
      positive: buffs.filter(b => b.type === 1).length,
      negative: buffs.filter(b => b.type === 2).length,
      control: buffs.filter(b => b.type === 3).length
    };
    
    let labelText = `[增益 ${buffTypeCount.positive}]`;
    if (buffTypeCount.negative > 0) labelText += ` | 减益 ${buffTypeCount.negative}`;
    if (buffTypeCount.control > 0) labelText += ` | 控制 ${buffTypeCount.control}`;
    
    ctx.font = '10px Arial';
    
    // 如果有减益或控制效果，标签颜色变红/紫警示
    if (buffTypeCount.control > 0) {
      ctx.fillStyle = '#AA44FF'; // 控制效果紫色警示
    } else if (buffTypeCount.negative > 0) {
      ctx.fillStyle = '#FF6644'; // 减益效果橙红色警示
    } else {
      ctx.fillStyle = '#888888';
    }
    
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
    ctx.fillText(labelText, startX, startY + iconSize + 3);

    ctx.restore();
  }

  /**
   * 检查是否可以拾取物品
   */
  checkPickupRange(playerX, playerY) {
    const pickupRange = 1.5; // 拾取范围1.5格
    
    for (let i = this.droppedItems.length - 1; i >= 0; i--) {
      const item = this.droppedItems[i];
      const distance = Math.hypot(playerX - item.x, playerY - item.y);
      
      if (distance <= pickupRange) {
        return item; // 返回可拾取的物品
      }
    }
    return null;
  }
  
  /**
   * 拾取物品
   */
  async pickupItem(item) {
    if (!item) return false;

    try {
      console.log(`🎒 拾取物品: ${item.itemName} x${item.quantity}`);

      // 发送拾取请求到服务端（CMD_PICKUP=2005）
      if (window.GameWS && window.GameWS.isConnected) {
        window.GameWS.send(window.Protocol.CMD_PICKUP || 2005, {
          item_id: item.itemID,
          x: Math.floor(item.x),
          y: Math.floor(item.y),
          quantity: item.quantity
        });
      }

      // 从地面移除（乐观更新，服务端会推送权威背包数据）
      const index = this.droppedItems.indexOf(item);
      if (index > -1) {
        this.droppedItems.splice(index, 1);
      }

      // 显示拾取成功提示
      this.game.showFloatingText(
        `获得 ${item.itemName} x${item.quantity}`,
        this.game.player.x,
        this.game.player.y,
        '#00FF00'
      );

      // 显示飘字
      this.addDamageNumber(this.game.player.x, this.game.player.y,
        `+${item.itemName}`, '#00FF00', false);

      return true;
    } catch (error) {
      console.error('拾取失败:', error);
      return false;
    }
  }
  
  /**
   * 显示升级特效
   */
  showLevelUpEffect(x, y, newLevel) {
    console.log(`🎉 升级特效: 等级 ${newLevel}`);
    
    // 添加升级飘字（大号金色）
    this.addDamageNumber(x, y - 2, `LEVEL UP!`, '#FFD700', true);
    
    // 添加等级数字
    setTimeout(() => {
      this.addDamageNumber(x, y - 1.5, `Lv.${newLevel}`, '#FFFFFF', true);
    }, 300);
    
    // 触发升级特效（如果有特效系统）
    if (this.game.triggerLevelUpEffect) {
      this.game.triggerLevelUpEffect(x, y);
    }
  }
  
  /**
   * 清除所有数据
   */
  clear() {
    this.monsters.clear();
    this.damageNumbers = [];
    this.selectedTarget = null;
  }
}

// 导出供其他模块使用
window.BattleSystem = BattleSystem;
