/**
 * 武学技能管理面板
 * 功能：查看武学图鉴、已学武学管理、装备/卸下武学到快捷栏
 */
class SkillPanel {
  constructor(game) {
    this.game = game;
    this.container = null;
    this.visible = false;

    // 数据
    this.allSkills = [];      // 所有武学（从skills.json）
    this.mySkills = [];       // 已学武学
    this.equippedSkills = []; // 已装备武学
    this.isLoading = false;   // 🔒 加载锁：防止异步期间被干扰

    // 类型映射
    this.typeNames = {
      1: '内功', 2: '外功', 3: '身法', 4: '护体',
      5: '拳法', 6: '剑法', 7: '刀法', 8: '枪法', 9: '斧法'
    };
    this.subTypeNames = { 1: '初级', 2: '进阶', 3: '高级' };

    // 当前筛选
    this.currentFilter = 0; // 0=全部

    this.init();
  }

  init() {
    this.createPanel();
    this.bindEvents();
  }

  /**
   * 创建面板DOM
   */
  createPanel() {
    this.container = document.createElement('div');
    this.container.id = 'skill-panel';
    this.container.style.cssText = `
      position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%);
      width: 720px; max-height: 520px;
      background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
      border: 1px solid #4a5568; border-radius: 12px;
      display: none; flex-direction: column; z-index: 1000;
      box-shadow: 0 8px 32px rgba(0,0,0,0.5); font-family: "Microsoft YaHei", sans-serif;
    `;
    document.body.appendChild(this.container);

    // 标题栏
    const header = document.createElement('div');
    header.style.cssText = `display:flex; justify-content:space-between; align-items:center;
      padding:12px 16px; background:rgba(255,255,255,0.06); border-bottom:1px solid #2d3748; border-radius:10px 10px 0 0;`;
    header.innerHTML = `
      <span style="font-size:16px; font-weight:bold; color:#e2e8f0;">武学图谱</span>
      <button id="skill-panel-close" style="background:none; border:none; color:#aaa; font-size:22px; cursor:pointer;">&times;</button>
    `;
    this.container.appendChild(header);

    // 筛选栏
    const filterBar = document.createElement('div');
    filterBar.id = 'skill-filter-bar';
    filterBar.style.cssText = `display:flex; gap:6px; padding:10px 16px; flex-wrap:wrap;`;

    const filters = [
      { key: 0, label: '全部' },
      { key: 1, label: '内功' }, { key: 2, label: '外功' }, { key: 3, label: '身法' }, { key: 4, label: '护体' },
      { key: 5, label: '拳法' }, { key: 6, label: '剑法' }, { key: 7, label: '刀法' }, { key: 8, label: '枪法' }, { key: 9, label: '斧法' }
    ];
    filters.forEach(f => {
      const btn = document.createElement('button');
      btn.dataset.type = f.key;
      btn.textContent = f.label;
      btn.style.cssText = `padding:4px 12px; border-radius:14px; cursor:pointer; font-size:12px;
        background:${f.key === 0 ? '#4a5568' : '#2d3748'}; color:#fff; border:1px solid ${f.key === 0 ? '#718096' : '#4a5568'};`;
      filterBar.appendChild(btn);
    });
    this.container.appendChild(filterBar);

    // 主内容区（自定义滚动条，固定高度）
    const content = document.createElement('div');
    content.id = 'skill-content';
    content.style.cssText = `height:360px; overflow-y:auto; padding:10px 16px;
      scrollbar-width: thin;
      scrollbar-color: #4a5568 #0a0a14;
    `;
    content.innerHTML = `<div style="color:#888; text-align:center; padding:40px;">加载中...</div>`;
    this.container.appendChild(content);

    // 详情区（底部）
    const detailArea = document.createElement('div');
    detailArea.id = 'skill-detail';
    detailArea.style.cssText = `padding:12px 16px; border-top:1px solid #2d3748; background:rgba(0,0,0,0.2); display:none;`;
    this.container.appendChild(detailArea);

    // 注入自定义滚动条样式（WebKit浏览器）
    if (!document.getElementById('skill-panel-scrollbar-style')) {
      const style = document.createElement('style');
      style.id = 'skill-panel-scrollbar-style';
      style.textContent = `
        #skill-content::-webkit-scrollbar { width: 8px; }
        #skill-content::-webkit-scrollbar-track {
          background: #0a0a14;
          border-radius: 4px;
          margin: 4px 0;
        }
        #skill-content::-webkit-scrollbar-thumb {
          background: linear-gradient(180deg, #4a5568, #2d3748);
          border-radius: 4px;
          border: 2px solid #0a0a14;
          transition: background 0.3s ease;
        }
        #skill-content::-webkit-scrollbar-thumb:hover {
          background: linear-gradient(180deg, #718096, #4a5568);
        }
        #skill-content::-webkit-scrollbar-corner { background: #0a0a14; }
      `;
      document.head.appendChild(style);
    }
  }

  bindEvents() {
    // 关闭按钮
    document.getElementById('skill-panel-close').onclick = () => this.hide();

    // 筛选按钮
    document.querySelectorAll('#skill-filter-bar button').forEach(btn => {
      btn.onclick = () => {
        this.currentFilter = parseInt(btn.dataset.type);
        // 更新样式
        document.querySelectorAll('#skill-filter-bar button').forEach(b => {
          b.style.background = parseInt(b.dataset.type) === this.currentFilter ? '#4a5568' : '#2d3748';
          b.style.borderColor = parseInt(b.dataset.type) === this.currentFilter ? '#718096' : '#4a5568';
        });
        this.renderSkillList();
      };
    });

    // 点击外部关闭（保存为实例方法以便移除）
    this.handleClickOutside = e => {
      if (this.visible && !this.container.contains(e.target)) {
        this.hide();
      }
    };
    document.addEventListener('click', this.handleClickOutside);
  }

  /**
   * 显示面板并加载数据
   */
  async show() {
    if (this.visible || this.isLoading) return;

    this.visible = true;
    this.isLoading = true;

    // 恢复正常的深色游戏风格样式
    Object.assign(this.container.style, {
      display: 'flex',
      position: 'fixed',
      top: '50%',
      left: '50%',
      transform: 'translate(-50%, -50%)',  // 居中显示
      zIndex: '900',                        // 在UI层之上但不是最高
      border: '1px solid #4a5568',          // 灰色边框
      background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)',  // 深色背景
      color: '#ffffff',                     // 白色文字
      opacity: '1',
      visibility: 'visible',
      pointerEvents: 'auto'
    });

    try {
      await this.loadData();
      this.renderSkillList();
    } finally {
      this.isLoading = false; // 🔓 解锁
    }
  }

  hide() {
    if (this.isLoading || !this.visible) return; // 🔒 加载期间或已隐藏则跳过

    this.visible = false;
    this.container.style.display = 'none';
    document.getElementById('skill-detail').style.display = 'none';
    // 移除点击外部关闭的事件监听
    document.removeEventListener('click', this.handleClickOutside);
  }

  /**
   * 销毁面板（清理资源）
   */
  destroy() {
    this.hide();
    // 确保移除所有事件监听
    document.removeEventListener('click', this.handleClickOutside);
  }

  toggle() {
    this.visible ? this.hide() : this.show();
  }

  /**
   * 从服务端加载数据
   */
  async loadData() {
    try {
      const roleId = this.game?.player?.id;

      // 并行请求：所有武学 + 已学武学 + 已装备武学
      const [allRes, myRes, equippedRes] = await Promise.all([
        fetch('http://localhost:8082/api/skill/base/list'),
        roleId ? fetch(`http://localhost:8082/api/skill/role/${roleId}/list`) : Promise.resolve({ json: () => ({ code: 0, data: [] }) }),
        roleId ? fetch(`http://localhost:8082/api/skill/role/${roleId}/equipped`) : Promise.resolve({ json: () => ({ code: 0, data: [] }) })
      ]);

      const allData = await allRes.json();
      const myData = await myRes.json();
      const equippedData = await equippedRes.json();

      if (allData.code === 200) this.allSkills = allData.data || [];
      if (myData.code === 200) this.mySkills = myData.data || [];
      if (equippedData.code === 200) this.equippedSkills = equippedData.data || [];

      // 同步到game.player（使用技能配置补全名称）
      if (this.game.player) {
        this.game.player.skills = this.mySkills.map(s => {
          const skillId = s.skill_id || s.id;
          const config = this.allSkills.find(skill => skill.id === skillId) || {};
          return {
            ...s,
            id: skillId,
            skill_id: skillId,
            name: config.name || '未知武学',
            type: config.type || 0,
            mp_cost: config.mp_cost || 0,
            cooldown: config.cooldown || 0,
            damage: config.damage || 0,
            range: config.range || 1,
            cast_time: config.cast_time || 0,
            attack_speed: config.attack_speed || 0,
          };
        });
      }
    } catch (error) {
      console.error('[SkillPanel] 加载数据失败:', error);
    }
  }

  /**
   * 渲染武学列表
   */
  renderSkillList() {
    const content = document.getElementById('skill-content');
    if (!content) {
      console.error('[SkillPanel] 找不到skill-content元素!');
      return;
    }

    let skills = [...this.allSkills];
    if (this.currentFilter > 0) {
      skills = skills.filter(s => s.type === this.currentFilter);
    }

    if (skills.length === 0) {
      content.innerHTML = `<div style="color:#888; text-align:center; padding:40px;">暂无此类武学</div>`;
      return;
    }

    try {
      content.innerHTML = skills.map(skill => {
      const mySkill = this.mySkills.find(s => (s.skill_id || s.id) === skill.id);
      const isEquipped = this.equippedSkills.some(s => (s.skill_id || s.id) === skill.id);
      const canLearn = !mySkill && (this.game?.player?.level || 1) >= skill.level;
      const typeColor = this.getTypeColor(skill.type);
      const subLabel = this.subTypeNames[skill.sub_type] || '';

      return `
        <div class="skill-item" data-id="${skill.id}"
          style="display:flex; align-items:center; padding:8px 12px; margin:4px 0; border-radius:8px;
            background:${mySkill ? 'rgba(74,222,128,0.08)' : 'rgba(255,255,255,0.03)'};
            border:1px solid ${isEquipped ? '#fbbf24' : (mySkill ? '#4ade80' : '#2d3748')}; cursor:pointer; transition:all 0.2s;"
          onmouseover="this.style.background=\x27${mySkill ? 'rgba(74,222,128,0.15)' : 'rgba(255,255,255,0.08)'}\x27"
          onmouseout="this.style.background=\x27${mySkill ? 'rgba(74,222,128,0.08)' : 'rgba(255,255,255,0.03)'}\x27">
          
          <!-- 图标 -->
          <div style="width:44px; height:44px; border-radius:8px; background:${typeColor}20; 
            border:1px solid ${typeColor}; display:flex; align-items:center; justify-content:center; margin-right:12px;">
            <span style="font-size:20px;">${this.getSkillIcon(skill.type)}</span>
          </div>
          
          <!-- 信息 -->
          <div style="flex:1;">
            <div style="display:flex; align-items:center; gap:8px;">
              <span style="font-weight:bold; color:#fff; font-size:14px;">${skill.name}</span>
              <span style="font-size:10px; color:${typeColor}; background:${typeColor}20; padding:1px 6px; border-radius:8px;">
                ${this.typeNames[skill.type] || ''}
              </span>
              ${subLabel ? `<span style="font-size:10px; color:#888;">${subLabel}</span>` : ''}
              ${isEquipped ? '<span style="font-size:10px; color:#fbbf24;">[已装备]</span>' : ''}
            </div>
            <div style="font-size:11px; color:#888; margin-top:3px;">
              Lv.${skill.level}需求 | ${skill.description}
              ${mySkill ? `<span style="color:#4ade80;"> | 已学 Lv.${mySkill.level || 1}</span>` : ''}
            </div>
            ${skill.damage > 0 ? `<div style="font-size:10px; color:#ef4444; margin-top:2px;">
              伤害:${skill.damage} MP:${skill.mp_cost}
              ${skill.cooldown > 0 ? '冷却:' + skill.cooldown + 's' : '攻速驱动'}
              ${skill.attack_speed > 0 ? ' 攻速:' + skill.attack_speed : ''}
              范围:${skill.range > 1 ? skill.range + '格(远程)' : '近战'}
            </div>` : ''}
          </div>

          <!-- 操作 -->
          <div style="display:flex; flex-direction:column; gap:4px; align-items:flex-end;">
            ${mySkill ? `
              ${!isEquipped ? `<button data-action="equip" data-id="${skill.id}"
                style="padding:3px 10px; font-size:11px; border-radius:10px; cursor:pointer;
                background:#fbbf24; color:#000; border:none;" onclick="event.stopPropagation();">装备</button>`
                : `<button data-action="unequip" data-id="${skill.id}"
                style="padding:3px 10px; font-size:11px; border-radius:10px; cursor:pointer;
                background:#666; color:#fff; border:none;" onclick="event.stopPropagation();">卸下</button>`}
            ` : canLearn ? `
              <button data-action="learn" data-id="${skill.id}"
                style="padding:3px 10px; font-size:11px; border-radius:10px; cursor:pointer;
                background:#4ade80; color:#000; border:none;" onclick="event.stopPropagation();">学习</button>
            ` : `
              <span style="font-size:10px; color:#666;">Lv.${skill.level}</span>
            `}
          </div>
        </div>
      `;
    }).join('');

    // 绑定点击事件显示详情
    content.querySelectorAll('.skill-item').forEach(item => {
      item.onclick = () => this.showDetail(parseInt(item.dataset.id));
    });

    // 绑定操作按钮事件
    content.querySelectorAll('[data-action]').forEach(btn => {
      btn.onclick = (e) => {
        e.stopPropagation();
        const action = btn.dataset.action;
        const id = parseInt(btn.dataset.id);
        if (action === 'equip') this.equipSkill(id);
        else if (action === 'unequip') this.unequipSkill(id);
        else if (action === 'learn') this.learnSkill(id);
      };
    });

    } catch (error) {
      console.error('[SkillPanel] 渲染失败:', error);
      content.innerHTML = `<div style="color:#ef4444; text-align:center; padding:40px;">渲染错误: ${error.message}</div>`;
    }
  }

  /**
   * 显示技能详情
   */
  showDetail(skillId) {
    const detail = document.getElementById('skill-detail');
    const skill = this.allSkills.find(s => s.id === skillId);
    if (!skill || !detail) return;

    const mySkill = this.mySkills.find(s => (s.skill_id || s.id) === skillId);
    const isEquipped = this.equippedSkills.some(s => (s.skill_id || s.id) === skillId);
    const typeColor = this.getTypeColor(skill.type);

    detail.style.display = 'block';
    detail.innerHTML = `
      <div style="display:flex; gap:16px; align-items:flex-start;">
        <div style="width:56px; height:56px; border-radius:10px; background:${typeColor}20;
          border:2px solid ${typeColor}; display:flex; align-items:center; justify-content:center;">
          <span style="font-size:26px;">${this.getSkillIcon(skill.type)}</span>
        </div>
        <div style="flex:1;">
          <div style="display:flex; align-items:center; gap:10px; margin-bottom:6px;">
            <span style="font-size:18px; font-weight:bold; color:#fff;">${skill.name}</span>
            <span style="font-size:12px; color:${typeColor};">${this.typeNames[skill.type]}</span>
            ${mySkill ? `<span style="color:#4ade80; font-size:13px;">Lv.${mySkill.level || 1}</span>` : `<span style="color:#888; font-size:12px;">未学习</span>`}
          </div>
          <p style="color:#aaa; font-size:12px; margin:0 0 8px 0;">${skill.description}</p>
          
          <div style="display:grid; grid-template-columns:repeat(4,1fr); gap:8px; font-size:11px;">
            <div><span style="color:#888;">等级要求:</span> <span style="color:#fff;">${skill.level}</span></div>
            <div><span style="color:#888;">最高等级:</span> <span style="color:#fff;">${skill.max_level}</span></div>
            ${skill.damage > 0 ? `
              <div><span style="color:#ef4444;">基础伤害:</span> <span style="color:#fff;">${skill.damage}</span></div>
              <div><span style="color:#60a5fa;">MP消耗:</span> <span style="color:#fff;">${skill.mp_cost}</span></div>
              <div><span style="color:${skill.cooldown > 0 ? '#fbbf24' : '#f59e0b'};">冷却:</span> <span style="color:#fff;">${skill.cooldown > 0 ? skill.cooldown + 's(固定)' : '无(攻速驱动)'}</span></div>
              ${skill.attack_speed > 0 ? `<div><span style="color:#c084fc;">攻速:</span> <span style="color:#fff;">${skill.attack_speed}(${skill.attack_speed <= 7 ? '很快' : skill.attack_speed <= 10 ? '中等' : '较慢'})</span></div>` : ''}
              <div><span style="color:#888;">范围:</span> <span style="color:#fff;">${skill.range > 1 ? skill.range + '格(远程)' : '近战'}</span></div>
              <div><span style="color:#888;">施法:</span> <span style="color:#fff;">${skill.cast_time}ms</span></div>
              ${skill.aoe_radius > 0 ? `<div><span style="color:#a78bfa;">AOE:</span> <span style="color:#fff;">${skill.aoe_radius}格</span></div>` : ''}
            ` : ''}
          </div>
          
          ${skill.attack_bonus > 0 || skill.defense_bonus > 0 || skill.hp_bonus > 0 ? `
            <div style="margin-top:8px; padding-top:8px; border-top:1px solid #333; font-size:11px; color:#888;">
              每级加成: 
              ${skill.hp_bonus > 0 ? `<span style="color:#ef4444;">生命+${skill.hp_bonus}</span> ` : ''}
              ${skill.mp_bonus > 0 ? `<span style="color:#60a5fa;">内力+${skill.mp_bonus}</span> ` : ''}
              ${skill.attack_bonus > 0 ? `<span style="color:#4ade80;">攻击+${skill.attack_bonus}</span> ` : ''}
              ${skill.defense_bonus > 0 ? `<span style="color:#60a5fa;">防御+${skill.defense_bonus}</span> ` : ''}
              ${skill.speed_bonus > 0 ? `<span style="color:#fbbf24;">速度+${skill.speed_bonus}</span> ` : ''}
              ${skill.hit_bonus > 0 ? `<span style="color:#fff;">命中+${skill.hit_bonus}</span> ` : ''}
              ${skill.dodge_bonus > 0 ? `<span style="color:#c084fc;">闪避+${skill.dodge_bonus}</span> ` : ''}
              ${skill.crit_bonus > 0 ? `<span style="color:#ef4444;">暴击+${skill.crit_bonus}</span>` : ''}
            </div>
          ` : ''}

          <div style="margin-top:8px; display:flex; gap:8px;">
            ${!mySkill ? `
              <button onclick="window.__skillPanel.learnSkill(${skill.id})"
                style="padding:6px 18px; border-radius:16px; cursor:pointer; background:#4ade80; color:#000; border:none; font-weight:bold;">学习此武学</button>
            ` : !isEquipped ? `
              <button onclick="window.__skillPanel.equipSkill(${skill.id})"
                style="padding:6px 18px; border-radius:16px; cursor:pointer; background:#fbbf24; color:#000; border:none; font-weight:bold;">装备到快捷栏</button>
            ` : `
              <button onclick="window.__skillPanel.unequipSkill(${skill.id})"
                style="padding:6px 18px; border-radius:16px; cursor:pointer; background:#666; color:#fff; border:none;">卸下装备</button>
            `}
          </div>
        </div>
      </div>
    `;
  }

  /**
   * 学习武学
   */
  async learnSkill(skillId) {
    try {
      const roleId = this.game?.player?.id;
      const roleLevel = this.game?.player?.level || 1;
      console.log('[SkillPanel] 学习武学 - roleId:', roleId, 'skillId:', skillId, 'roleLevel:', roleLevel, '完整player:', this.game?.player);
      if (!roleId) return;

      const res = await fetch(`http://localhost:8082/api/skill/role/${roleId}/learn`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          skill_id: skillId,
          role_level: roleLevel
        })
      });
      const data = await res.json();

      if (data.code === 200) {
        if (this.game.uiManager) this.game.uiManager.toast('学习成功！', 'success', 1500);
        await this.loadData();
        this.renderSkillList();
        // 刷新技能栏
        if (this.game.skillBar) this.game.skillBar.loadFromServer();
        // 刷新属性
        if (this.game.inventory) this.game.inventory.updateAttributes();
      } else {
        if (this.game.uiManager) this.game.uiManager.toast(data.msg || '学习失败', 'error', 2000);
      }
    } catch (error) {
      console.error('[SkillPanel] 学习失败:', error);
      if (this.game.uiManager) this.game.uiManager.toast('网络错误', 'error', 1500);
    }
  }

  /**
   * 装备武学到快捷栏
   */
  async equipSkill(skillId) {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) return;

      const res = await fetch(`http://localhost:8082/api/skill/role/${roleId}/equip`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ skill_id: skillId })
      });
      const data = await res.json();

      if (data.code === 200) {
        if (this.game.uiManager) this.game.uiManager.toast('装备成功！', 'success', 1000);
        await this.loadData();
        this.renderSkillList();
        this.showDetail(skillId); // 刷新详情
        if (this.game.skillBar) this.game.skillBar.loadFromServer();
      } else {
        if (this.game.uiManager) this.game.uiManager.toast(data.msg || '装备失败', 'error', 1500);
      }
    } catch (error) {
      console.error('[SkillPanel] 装备失败:', error);
      if (this.game.uiManager) this.game.uiManager.toast('网络错误', 'error', 1500);
    }
  }

  /**
   * 卸下武学
   */
  async unequipSkill(skillId) {
    try {
      const roleId = this.game?.player?.id;
      if (!roleId) return;

      const res = await fetch(`http://localhost:8082/api/skill/role/${roleId}/unequip`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ skill_id: skillId })
      });
      const data = await res.json();

      if (data.code === 200) {
        if (this.game.uiManager) this.game.uiManager.toast('已卸下', 'success', 1000);
        await this.loadData();
        this.renderSkillList();
        this.showDetail(skillId);
        if (this.game.skillBar) this.game.skillBar.loadFromServer();
      } else {
        if (this.game.uiManager) this.game.uiManager.toast(data.msg || '操作失败', 'error', 1500);
      }
    } catch (error) {
      console.error('[SkillPanel] 卸下失败:', error);
      if (this.game.uiManager) this.game.uiManager.toast('网络错误', 'error', 1500);
    }
  }

  getSkillIcon(type) {
    const icons = { 1: '☯', 2: '💪', 3: '🏃', 4: '🛡', 5: '👊', 6: '🗡', 7: '⚔', 8: '🔱', 9: '🪓' };
    return icons[type] || '✨';
  }

  getTypeColor(type) {
    const colors = { 1: '#60a5fa', 2: '#ef4444', 3: '#c084fc', 4: '#f59e0b',
                     5: '#f97316', 6: '#06b6d4', 7: '#dc2626', 8: '#84cc16', 9: '#a78bfa' };
    return colors[type] || '#888';
  }
}

// 全局暴露（供onclick调用）
if (typeof window !== 'undefined') {
  window.SkillPanel = SkillPanel;
}
