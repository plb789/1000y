package attribute

import (
	"context"
	"fmt"
	common "game-server/Common"
	"math"

	"github.com/redis/go-redis/v9"
)

// Attribute 角色属性（最终计算结果）
type Attribute struct {
	Hp         int `json:"hp"`          // 当前生命值
	MaxHp      int `json:"max_hp"`      // 最大生命值
	Mp         int `json:"mp"`          // 当前内力值
	MaxMp      int `json:"max_mp"`      // 最大内力值
	Stamina    int `json:"stamina"`     // 体力值
	MaxStamina int `json:"max_stamina"` // 最大体力值
	Attack     int `json:"attack"`      // 攻击力
	Defense    int `json:"defense"`     // 防御力
	Speed      int `json:"speed"`       // 速度
	Hit        int `json:"hit"`         // 命中率
	Dodge      int `json:"dodge"`       // 闪避率
	Crit       int `json:"crit"`        // 暴击率
	CritDamage int `json:"crit_damage"` // 暴击伤害
}

// AttributeBonus 属性加成（用于UI显示明细）
type AttributeBonus struct {
	ItemBonus  *ItemBonus  `json:"item_bonus,omitempty"`  // 装备加成
	SkillBonus *SkillBonus `json:"skill_bonus,omitempty"` // 武学加成
	BuffBonus  *BuffBonus  `json:"buff_bonus,omitempty"`  // BUFF加成
}

// ItemBonus 装备属性加成
type ItemBonus struct {
	Hp         int `json:"hp"`
	MaxHp      int `json:"max_hp"`
	Mp         int `json:"mp"`
	MaxMp      int `json:"max_mp"`
	Attack     int `json:"attack"`
	Defense    int `json:"defense"`
	Speed      int `json:"speed"`
	Hit        int `json:"hit"`
	Dodge      int `json:"dodge"`
	Crit       int `json:"crit"`
	CritDamage int `json:"crit_damage"`
}

// SkillBonus 武学属性加成
type SkillBonus struct {
	Hp      int `json:"hp"`
	Mp      int `json:"mp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
	Hit     int `json:"hit"`
	Dodge   int `json:"dodge"`
	Crit    int `json:"crit"`
}

// BuffBonus BUFF属性加成
type BuffBonus struct {
	Hp      int `json:"hp"`
	MaxHp   int `json:"max_hp"`
	Mp      int `json:"mp"`
	MaxMp   int `json:"max_mp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Speed   int `json:"speed"`
	Hit     int `json:"hit"`
	Dodge   int `json:"dodge"`
	Crit    int `json:"crit"`
}

// ★ 接口定义：技能加成提供者（打破循环依赖的关键）
// Skill包将实现此接口，通过依赖注入传入CalcEngine
type SkillBonusProvider interface {
	// CalculateSkillBonus 计算指定角色的已装备武学加成
	CalculateSkillBonus(ctx context.Context, roleID uint64) (*SkillBonus, error)

	// GetEquippedSkills 获取角色已装备的武学列表（用于返回给前端）
	GetEquippedSkills(ctx context.Context, roleID uint64) ([]map[string]interface{}, error)
}

// ★ 接口定义：装备加成提供者
type ItemBonusProvider interface {
	// GetEquippedItems 获取角色已装备的物品列表
	GetEquippedItems(roleID uint64) ([]map[string]interface{}, error)
}

// CalcEngine 属性计算引擎
// ★ 核心职责：整合基础属性 + 装备加成 + 技能加成 + BUFF加成 = 最终属性
// ★ 通过接口依赖注入，避免直接import skill/item包，打破循环依赖
type CalcEngine struct {
	rdb           *redis.Client      // Redis客户端（缓存中间结果）
	itemProvider  ItemBonusProvider  // 装备加成提供者（接口）
	skillProvider SkillBonusProvider // 技能加成提供者（接口）
}

// NewCalcEngine 创建属性计算引擎实例
// ★ 通过接口注入依赖，而不是具体类型，实现解耦
func NewCalcEngine(rdb *redis.Client, itemProvider ItemBonusProvider, skillProvider SkillBonusProvider) *CalcEngine {
	return &CalcEngine{
		rdb:           rdb,
		itemProvider:  itemProvider,
		skillProvider: skillProvider,
	}
}

// CalculateFinalAttributes 计算角色的最终完整属性
// 这是核心方法：整合所有数据源，返回最终可用的属性
func (ce *CalcEngine) CalculateFinalAttributes(ctx context.Context, roleID uint64) (*Attribute, *AttributeBonus, error) {
	// 1. 获取角色基础属性（从DBService）
	baseAttrs, err := ce.getBaseAttributes(ctx, roleID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取基础属性失败: %v", err)
	}

	// 2. 计算装备加成（通过接口调用）
	itemBonus, err := ce.calculateItemBonus(ctx, roleID)
	if err != nil {
		fmt.Printf("⚠️ 计算装备加成失败: %v\n", err)
		itemBonus = &ItemBonus{}
	}

	// 3. 计算技能加成（通过接口调用，不再直接import skill包）
	skillBonus, err := ce.calculateSkillBonus(ctx, roleID)
	if err != nil {
		fmt.Printf("⚠️ 计算技能加成失败: %v\n", err)
		skillBonus = &SkillBonus{}
	}

	// 4. 计算BUFF加成（暂时为空，后续扩展）
	buffBonus := &BuffBonus{}

	// 5. 合并所有加成为最终属性
	finalAttrs := &Attribute{
		Hp:         baseAttrs.Hp,
		MaxHp:      baseAttrs.MaxHp + itemBonus.MaxHp + skillBonus.Hp + buffBonus.MaxHp,
		Mp:         baseAttrs.Mp,
		MaxMp:      baseAttrs.Mp + itemBonus.Mp + skillBonus.Mp + buffBonus.Mp,
		Stamina:    baseAttrs.Stamina,
		MaxStamina: baseAttrs.MaxStamina,
		Attack:     baseAttrs.Attack + itemBonus.Attack + skillBonus.Attack + buffBonus.Attack,
		Defense:    baseAttrs.Defense + itemBonus.Defense + skillBonus.Defense + buffBonus.Defense,
		Speed:      baseAttrs.Speed + itemBonus.Speed + skillBonus.Speed + buffBonus.Speed,
		Hit:        baseAttrs.Hit + itemBonus.Hit + skillBonus.Hit + buffBonus.Hit,
		Dodge:      baseAttrs.Dodge + itemBonus.Dodge + skillBonus.Dodge + buffBonus.Dodge,
		Crit:       baseAttrs.Crit + itemBonus.Crit + skillBonus.Crit + buffBonus.Crit,
		CritDamage: baseAttrs.CritDamage + itemBonus.CritDamage,
	}

	// 6. 组装加成明细（供前端显示）
	bonusDetail := &AttributeBonus{
		ItemBonus:  itemBonus,
		SkillBonus: skillBonus,
		BuffBonus:  buffBonus,
	}

	return finalAttrs, bonusDetail, nil
}

// getBaseAttributes 从DBService获取角色基础属性
func (ce *CalcEngine) getBaseAttributes(ctx context.Context, roleID uint64) (*Attribute, error) {
	roleInfo, err := common.DBRoleGet(roleID)
	if err != nil {
		return nil, err
	}

	return &Attribute{
		Hp:         roleInfo.Hp,
		MaxHp:      roleInfo.MaxHp,
		Mp:         roleInfo.Mp,
		MaxMp:      roleInfo.MaxMp,
		Stamina:    0,  // RoleInfo暂无此字段，默认为0
		MaxStamina: 0, // RoleInfo暂无此字段，默认为0
		Attack:     roleInfo.Attack,
		Defense:    roleInfo.Defense,
		Speed:      roleInfo.Speed,
		Hit:        roleInfo.Hit,
		Dodge:      roleInfo.Dodge,
		Crit:       roleInfo.Crit,
		CritDamage: roleInfo.CritDamage, // 需要确认RoleInfo是否有此字段
	}, nil
}

// calculateItemBonus 计算已装备物品的属性加成（通过itemProvider接口调用）
func (ce *CalcEngine) calculateItemBonus(ctx context.Context, roleID uint64) (*ItemBonus, error) {
	if ce.itemProvider == nil {
		return &ItemBonus{}, nil
	}

	equippedItems, err := ce.itemProvider.GetEquippedItems(roleID)
	if err != nil {
		return nil, err
	}

	bonus := &ItemBonus{}
	for _, item := range equippedItems {
		if itemConfig, ok := item["config"].(map[string]interface{}); ok {
			if hp, ok := itemConfig["hp_bonus"].(float64); ok {
				bonus.Hp += int(hp)
				bonus.MaxHp += int(hp)
			}
			if mp, ok := itemConfig["mp_bonus"].(float64); ok {
				bonus.Mp += int(mp)
				bonus.MaxMp += int(mp)
			}
			if attack, ok := itemConfig["attack"].(float64); ok {
				bonus.Attack += int(attack)
			}
			if defense, ok := itemConfig["defense"].(float64); ok {
				bonus.Defense += int(defense)
			}
			if speed, ok := itemConfig["speed"].(float64); ok {
				bonus.Speed += int(speed)
			}
			if hit, ok := itemConfig["hit"].(float64); ok {
				bonus.Hit += int(hit)
			}
			if dodge, ok := itemConfig["dodge"].(float64); ok {
				bonus.Dodge += int(dodge)
			}
			if crit, ok := itemConfig["crit"].(float64); ok {
				bonus.Crit += int(crit)
			}
			if critDmg, ok := itemConfig["crit_damage"].(float64); ok {
				bonus.CritDamage += int(critDmg)
			}
		}
	}

	return bonus, nil
}

// calculateSkillBonus 计算已装备武学的属性加成（通过skillProvider接口调用）
// ★ 关键改进：不再直接import skill包，而是通过接口调用
func (ce *CalcEngine) calculateSkillBonus(ctx context.Context, roleID uint64) (*SkillBonus, error) {
	if ce.skillProvider == nil {
		return &SkillBonus{}, nil
	}

	// ★ 通过接口调用，由Skill包的实现来处理具体逻辑
	return ce.skillProvider.CalculateSkillBonus(ctx, roleID)
}

// ceil 向上取整辅助函数
func ceil(x float64) float64 {
	return math.Ceil(x*100) / 100
}
