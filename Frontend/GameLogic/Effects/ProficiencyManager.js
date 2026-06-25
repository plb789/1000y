/**
 * 熟练度管理器
 * 负责管理技能熟练度的获取、增加和升级逻辑
 */
class ProficiencyManager {
  constructor(game) {
    this.game = game;
    this.expGainConfig = {
      // 技能使用获得的基础熟练度
      baseGain: 1,
      // 命中额外加成
      hitBonus: 0.5,
      // 暴击额外加成
      critBonus: 1.5,
      // 击杀额外加成
      killBonus: 5,
      // 连击加成（每增加1连击）
      comboBonus: 0.2
    };
  }

  /**
   * 使用技能时增加熟练度
   * @param {number} skillId - 技能ID
   * @param {object} options - 额外参数
   * @param {boolean} options.isHit - 是否命中
   * @param {boolean} options.isCrit - 是否暴击
   * @param {boolean} options.isKill - 是否击杀
   * @param {number} options.combo - 连击数
   */
  async addExp(skillId, options = {}) {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId || !skillId || skillId === 0) return;

      // 计算熟练度增量
      let expGain = this.expGainConfig.baseGain;
      
      if (options.isHit) {
        expGain += this.expGainConfig.hitBonus;
      }
      if (options.isCrit) {
        expGain += this.expGainConfig.critBonus;
      }
      if (options.isKill) {
        expGain += this.expGainConfig.killBonus;
      }
      if (options.combo && options.combo > 1) {
        expGain += (options.combo - 1) * this.expGainConfig.comboBonus;
      }

      // 通过WebSocket发送请求到服务器
      const data = await window.GameWS.request(window.Protocol.CMD_SKILL_EXP, {
        role_id: roleId,
        skill_id: skillId,
        exp: Math.floor(expGain)
      }, 5000).catch(() => null);

      if (data && data.code === 200) {
        const result = data.data;
        // 显示升级提示
        if (result.level_up) {
          this.showLevelUpNotification(result.skill_name, result.new_level);
        }
        // 更新本地缓存
        this.updateLocalSkillExp(skillId, result.current_exp, result.current_level);
        // 刷新UI
        this.refreshUI();
      }

      return data;
    } catch (error) {
      console.error('[ProficiencyManager] 增加熟练度失败:', error);
      return null;
    }
  }

  /**
   * 更新本地技能熟练度数据
   */
  updateLocalSkillExp(skillId, exp, level) {
    if (!this.game.player?.skills) return;
    
    const skill = this.game.player.skills.find(s => s.id === skillId);
    if (skill) {
      skill.exp = exp;
      skill.level = level;
      skill.max_exp = level * 100;
      skill.exp_percent = Math.min(100, (exp / (level * 100)) * 100);
    }
  }

  /**
   * 显示升级通知
   */
  showLevelUpNotification(skillName, newLevel) {
    const notification = document.createElement('div');
    notification.style.cssText = `
      position: fixed;
      top: 40%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
      border: 2px solid #fbbf24;
      border-radius: 12px;
      padding: 20px 30px;
      text-align: center;
      z-index: 2000;
      animation: levelUpAnim 2s ease-out forwards;
      box-shadow: 0 0 30px rgba(251, 191, 36, 0.5);
    `;
    notification.innerHTML = `
      <div style="font-size:24px; margin-bottom:10px;">🎉</div>
      <div style="font-size:18px; font-weight:bold; color:#fbbf24; margin-bottom:5px;">武学升级</div>
      <div style="font-size:14px; color:#fff;">${skillName}</div>
      <div style="font-size:20px; font-weight:bold; color:#4ade80;">Lv.${newLevel}</div>
    `;
    
    document.body.appendChild(notification);
    
    // 添加动画样式
    const style = document.createElement('style');
    style.textContent = `
      @keyframes levelUpAnim {
        0% { opacity: 0; transform: translate(-50%, -50%) scale(0.5); }
        20% { opacity: 1; transform: translate(-50%, -50%) scale(1.1); }
        30% { transform: translate(-50%, -50%) scale(0.95); }
        40% { transform: translate(-50%, -50%) scale(1.02); }
        100% { opacity: 0; transform: translate(-50%, -70%) scale(1); }
      }
    `;
    document.head.appendChild(style);
    
    // 2秒后移除
    setTimeout(() => {
      notification.remove();
      style.remove();
    }, 2000);
  }

  /**
   * 刷新UI显示
   */
  refreshUI() {
    // 刷新技能栏
    if (this.game.skillBar) {
      this.game.skillBar.loadFromServer();
    }
    // 刷新武学面板
    if (this.game.skillPanel) {
      this.game.skillPanel.loadData().then(() => {
        this.game.skillPanel.renderSkillList();
      });
    }
  }

  /**
   * 获取技能当前熟练度
   */
  getSkillExp(skillId) {
    const skill = this.game.player?.skills?.find(s => s.id === skillId);
    if (skill) {
      return {
        exp: skill.exp || 0,
        level: skill.level || 1,
        maxExp: skill.max_exp || 100,
        percent: skill.exp_percent || 0
      };
    }
    return null;
  }

  /**
   * 获取熟练度等级名称
   */
  getLevelName(level) {
    const names = [
      '初学', '入门', '熟练', '精通', '炉火纯青', 
      '登堂入室', '出神入化', '返璞归真', '天人合一', '传说'
    ];
    return names[Math.min(level - 1, names.length - 1)] || '传说';
  }

  /**
   * 获取熟练度等级颜色
   */
  getLevelColor(level) {
    const colors = [
      '#9ca3af', '#4ade80', '#22c55e', '#84cc16', '#eab308',
      '#f97316', '#ef4444', '#ec4899', '#a855f7', '#6366f1'
    ];
    return colors[Math.min(level - 1, colors.length - 1)] || '#6366f1';
  }
}

export default ProficiencyManager;