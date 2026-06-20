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
        if (data.hp !== undefined) monster.hp = data.hp;
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
    
    // 绘制怪物身体（圆形）
    ctx.save();
    ctx.fillStyle = color;
    ctx.strokeStyle = '#000';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(x, y, halfSize * 0.7, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
    
    // 绘制眼睛
    ctx.fillStyle = '#FFF';
    ctx.beginPath();
    ctx.arc(x - size * 0.15, y - size * 0.1, size * 0.12, 0, Math.PI * 2);
    ctx.arc(x + size * 0.15, y - size * 0.1, size * 0.12, 0, Math.PI * 2);
    ctx.fill();
    
    // 绘制瞳孔
    ctx.fillStyle = '#000';
    ctx.beginPath();
    ctx.arc(x - size * 0.15, y - size * 0.08, size * 0.06, 0, Math.PI * 2);
    ctx.arc(x + size * 0.15, y - size * 0.08, size * 0.06, 0, Math.PI * 2);
    ctx.fill();
    
    // 绘制名称
    ctx.fillStyle = '#FFF';
    ctx.font = 'bold 10px Arial';
    ctx.textAlign = 'center';
    ctx.fillText(monster.name, x, y - halfSize - 5);
    
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
    // 检查冷却
    const now = Date.now();
    if (now - this.lastAttackTime < this.attackCooldown) {
      console.log('攻击冷却中...');
      this.game.showFloatingText('攻击冷却中...', this.game.player.x, this.game.player.y, '#FFAA00');
      return false;
    }
    
    // 检查距离
    const target = this.monsters.get(targetId);
    if (!target) {
      console.log('目标不存在');
      return false;
    }
    
    const player = this.game.player;
    const distance = Math.hypot(player.x - target.x, player.y - target.y);
    
    if (distance > 1.5) { // 近战攻击范围1.5格
      console.log('距离过远');
      this.game.showFloatingText('距离过远!', player.x, player.y, '#FF0000');
      return false;
    }
    
    // 设置冷却
    this.lastAttackTime = now;
    
    try {
      console.log(`⚔️ 发起攻击 -> 怪物 ${target.name} (${targetId})`);
      
      // 通过HTTP发送攻击请求到服务端
            const response = await fetch('http://localhost:8082/api/battle/attack', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                role_id: this.game.roleID || 1,
                monster_id: targetId,
                skill_id: 0,  // 0=普通攻击
                x: Math.floor(player.x),
                y: Math.floor(player.y)
              })
            });;
      
      const result = await response.json();
      
      if (result.code !== 200 && result.code !== 0) {
        // 攻击失败（服务端返回错误）
        console.warn(`❌ 攻击失败: ${result.msg}`);
        this.game.showFloatingText(result.msg || '攻击失败', player.x, player.y, '#FF0000');
        return false;
      }
      
      // 处理攻击结果
      const data = result.data;
      this.handleAttackResult(data, target);
      
      // 播放攻击动画
      this.playAttackAnimation(player.x, player.y, target.x, target.y);
      
      return true;
      
    } catch (error) {
      console.error('攻击请求失败:', error);
      this.game.showFloatingText('网络错误', player.x, player.y, '#FF0000');
      return false;
    }
  }
  
  /**
   * 处理攻击结果
   */
  handleAttackResult(result, target) {
    if (!result || !result.Success) {
      return;
    }
    
    const player = this.game.player;
    
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
    
    // 更新怪物HP
    const monster = this.monsters.get(targetId);
    if (monster) {
      monster.hp = currentHp;
      
      if (isDead) {
        monster.status = 4; // 死亡状态
        
        // 延迟移除死亡怪物（3秒后）
        setTimeout(() => {
          this.removeMonster(targetId);
        }, 3000);
        
        // 显示掉落信息
        if (data.exp_gain) {
          this.showExpGain(data.exp_gain, monster.x, monster.y);
        }
        if (data.drops && data.drops.length > 0) {
          this.showDrops(data.drops, monster.x, monster.y);
        }
      }
    }
    
    // 显示伤害数字
    const targetPos = monster ? { x: monster.x, y: monster.y } : { x: data.target_x, y: data.target_y };
    this.addDamageNumber(targetPos.x, targetPos.y, damage, isCrit, isMiss, isBlocked);
    
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
    drops.forEach((itemId, index) => {
      setTimeout(() => {
        this.addDamageNumber(x, y - 1 - index * 0.5, `+物品`, false, false, false);
      }, index * 200);
    });
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
      
      // 从地面移除
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
