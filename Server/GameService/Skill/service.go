package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	common "game-server/Common"
	attribute "game-server/GameService/Attribute"
	"math"
)

// Service 武学服务
type Service struct{}

// NewService 创建武学服务实例
func NewService() *Service {
	return &Service{}
}

// LearnSkill 学习武学
// 返回: 错误信息
func (s *Service) LearnSkill(roleID uint64, skillID uint32) error {
	return common.DBSkillLearn(roleID, skillID)
}

// GetRoleSkills 获取角色所有武学
func (s *Service) GetRoleSkills(roleID uint64) ([]map[string]interface{}, error) {
	return common.DBSkillGetList(roleID)
}

// GetRoleSkillsByType 获取角色指定类型的武学
func (s *Service) GetRoleSkillsByType(roleID uint64, skillType uint8) ([]map[string]interface{}, error) {
	allSkills, err := common.DBSkillGetList(roleID)
	if err != nil {
		return nil, err
	}

	// 过滤指定类型
	var filtered []map[string]interface{}
	for _, skill := range allSkills {
		if t, ok := skill["type"].(float64); ok && uint8(t) == skillType {
			filtered = append(filtered, skill)
		}
	}
	return filtered, nil
}

// GetSkillBase 获取武学基础信息
func (s *Service) GetSkillBase(skillID uint32) (map[string]interface{}, error) {
	config := common.GetSkillConfig(skillID)
	if config == nil {
		return nil, errors.New("武学不存在")
	}
	return map[string]interface{}{
		"id":            config.ID,
		"name":          config.Name,
		"type":          config.Type,
		"sub_type":      config.SubType,
		"level":         config.Level,
		"max_level":     config.MaxLevel,
		"exp_factor":    config.ExpFactor,
		"description":   config.Description,
		"hp_bonus":      config.HpBonus,
		"mp_bonus":      config.MpBonus,
		"attack_bonus":  config.AttackBonus,
		"defense_bonus": config.DefBonus,
		"speed_bonus":   config.SpeedBonus,
		"hit_bonus":     config.HitBonus,
		"dodge_bonus":   config.DodgeBonus,
		"crit_bonus":    config.CritBonus,
		"buff_id":       config.BuffID,
		"skill_effect":  config.SkillEffect,
		"is_active":     config.IsActive,
		"weapon_type":   config.WeaponType,
		// 战斗属性
		"damage":       config.Damage,
		"mp_cost":      config.MpCost,
		"cooldown":     config.Cooldown,
		"range":        config.Range,
		"cast_time":    config.CastTime,
		"aoe_radius":   config.AoeRadius,
		"attack_speed": config.AttackSpeed,
	}, nil
}

// GetAllSkillBase 获取所有武学基础信息
func (s *Service) GetAllSkillBase() ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, config := range common.GetAllSkillConfig() {
		result = append(result, map[string]interface{}{
			"id":            config.ID,
			"name":          config.Name,
			"type":          config.Type,
			"sub_type":      config.SubType,
			"level":         config.Level,
			"max_level":     config.MaxLevel,
			"exp_factor":    config.ExpFactor,
			"description":   config.Description,
			"hp_bonus":      config.HpBonus,
			"mp_bonus":      config.MpBonus,
			"attack_bonus":  config.AttackBonus,
			"defense_bonus": config.DefBonus,
			"speed_bonus":   config.SpeedBonus,
			"hit_bonus":     config.HitBonus,
			"dodge_bonus":   config.DodgeBonus,
			"crit_bonus":    config.CritBonus,
			"buff_id":       config.BuffID,
			"skill_effect":  config.SkillEffect,
			"is_active":     config.IsActive,
			"weapon_type":   config.WeaponType,
			// 战斗属性
			"damage":       config.Damage,
			"mp_cost":      config.MpCost,
			"cooldown":     config.Cooldown,
			"range":        config.Range,
			"cast_time":    config.CastTime,
			"aoe_radius":   config.AoeRadius,
			"attack_speed": config.AttackSpeed,
		})
	}
	return result, nil
}

// GetSkillBaseByType 获取指定类型的所有武学
func (s *Service) GetSkillBaseByType(skillType uint8) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, config := range common.GetAllSkillConfig() {
		if config.Type == skillType {
			result = append(result, map[string]interface{}{
				"id":            config.ID,
				"name":          config.Name,
				"type":          config.Type,
				"sub_type":      config.SubType,
				"level":         config.Level,
				"max_level":     config.MaxLevel,
				"exp_factor":    config.ExpFactor,
				"description":   config.Description,
				"hp_bonus":      config.HpBonus,
				"mp_bonus":      config.MpBonus,
				"attack_bonus":  config.AttackBonus,
				"defense_bonus": config.DefBonus,
				"speed_bonus":   config.SpeedBonus,
				"hit_bonus":     config.HitBonus,
				"dodge_bonus":   config.DodgeBonus,
				"crit_bonus":    config.CritBonus,
				"buff_id":       config.BuffID,
				"skill_effect":  config.SkillEffect,
				"is_active":     config.IsActive,
				"weapon_type":   config.WeaponType,
				// 战斗属性
				"damage":       config.Damage,
				"mp_cost":      config.MpCost,
				"cooldown":     config.Cooldown,
				"range":        config.Range,
				"cast_time":    config.CastTime,
				"aoe_radius":   config.AoeRadius,
				"attack_speed": config.AttackSpeed,
			})
		}
	}
	return result, nil
}

// AddExp 增加武学熟练度
// roleID: 角色ID
// skillID: 武学ID
// exp: 增加的熟练度
// 返回: 是否升级及错误信息
func (s *Service) AddExp(roleID uint64, skillID uint32, exp int64) (bool, uint32, error) {
	// 获取武学配置
	skillConfig := common.GetSkillConfig(skillID)
	if skillConfig == nil {
		return false, 0, errors.New("武学不存在")
	}

	// 获取角色当前武学数据
	skills, err := common.DBSkillGetList(roleID)
	if err != nil {
		return false, 0, err
	}

	var currentLevel int = 1
	var currentExp int64 = 0
	for _, skill := range skills {
		if sid, ok := skill["skill_id"].(float64); ok && uint32(sid) == skillID {
			if level, ok := skill["level"].(float64); ok {
				currentLevel = int(level)
			}
			if e, ok := skill["exp"].(float64); ok {
				currentExp = int64(e)
			}
			break
		}
	}

	// 计算升级
	leveledUp := false
	newExp := currentExp + exp
	newLevel := currentLevel

	for newExp >= int64(skillConfig.ExpFactor)*int64(newLevel)*int64(newLevel) && uint32(newLevel) < skillConfig.MaxLevel {
		newExp -= int64(skillConfig.ExpFactor) * int64(newLevel) * int64(newLevel)
		newLevel++
		leveledUp = true
	}

	// 调用DBService更新
	_, _, _, err = common.DBSkillAddExpWithLevel(roleID, skillID, newExp, newLevel, leveledUp)
	if err != nil {
		return false, 0, err
	}

	return leveledUp, uint32(newLevel), nil
}

// UpgradeSkill 手动升级武学（使用道具或其他方式）
// 返回: 新等级及错误信息
func (s *Service) UpgradeSkill(roleID uint64, skillID uint32) (uint32, error) {
	// 获取武学信息
	skillInfo, err := common.DBSkillGetBase(skillID)
	if err != nil {
		return 0, errors.New("武学不存在")
	}

	// 检查是否已达最高等级
	skills, err := common.DBSkillGetList(roleID)
	if err != nil {
		return 0, err
	}

	for _, s := range skills {
		if sid, ok := s["skill_id"].(float64); ok && uint32(sid) == skillID {
			if level, ok := s["level"].(float64); ok {
				maxLevel := 10
				if ml, ok := skillInfo["max_level"].(float64); ok {
					maxLevel = int(ml)
				}
				if uint32(level) >= uint32(maxLevel) {
					return uint32(level), errors.New("已达最高等级")
				}
				// 手动升级需要消耗经验，这里简化处理
				return uint32(level) + 1, nil
			}
		}
	}

	return 0, errors.New("未学习该武学")
}

// EquipSkill 装备武学
// 千年游戏中,外功/拳法/剑法/刀法/枪法/斧法只能装备一个
// 内功/身法/护体可以各装备一个
func (s *Service) EquipSkill(roleID uint64, skillID uint32) error {
	// 获取要装备的武学配置
	config := common.GetSkillConfig(skillID)
	if config == nil {
		return errors.New("武学不存在")
	}

	// 获取角色已装备的武学
	equipped, err := s.GetEquippedSkills(roleID)
	if err != nil {
		return err
	}

	// 需要互斥的武学类型：外功(2)/拳法(5)/剑法(6)/刀法(7)/枪法(8)/斧法(9)
	exclusiveTypes := map[uint8]bool{2: true, 5: true, 6: true, 7: true, 8: true, 9: true}

	// 如果当前武学类型需要互斥，先卸下同类型的已装备武学
	if exclusiveTypes[config.Type] {
		for _, skill := range equipped {
			skillType, _ := skill["type"].(float64)
			if uint8(skillType) == config.Type {
				skillID, _ := skill["skill_id"].(float64)
				if err := common.DBSkillUnequip(roleID, uint32(skillID)); err != nil {
					return err
				}
			}
		}
	}

	return common.DBSkillEquip(roleID, skillID)
}

// UnequipSkill 卸下武学
func (s *Service) UnequipSkill(roleID uint64, skillID uint32) error {
	return common.DBSkillUnequip(roleID, skillID)
}

// GetEquippedSkills 获取角色已装备的武学
func (s *Service) GetEquippedSkills(roleID uint64) ([]map[string]interface{}, error) {
	return common.DBSkillGetEquipped(roleID)
}

// CalculateSkillBonus 计算武学加成
// 根据已装备的武学计算角色属性加成
func (s *Service) CalculateSkillBonus(roleID uint64) (map[string]int, error) {
	equippedSkills, err := s.GetEquippedSkills(roleID)
	if err != nil {
		return nil, err
	}

	bonus := map[string]int{
		"hp":      0,
		"mp":      0,
		"attack":  0,
		"defense": 0,
		"speed":   0,
		"hit":     0,
		"dodge":   0,
		"crit":    0,
	}

	for _, skill := range equippedSkills {
		level := 1.0
		if l, ok := skill["level"].(float64); ok {
			level = l
		}

		addHp := 0
		if v, ok := skill["hp_bonus"].(float64); ok {
			addHp = int(v)
		}
		bonus["hp"] += int(math.Ceil(float64(addHp) * level))

		addMp := 0
		if v, ok := skill["mp_bonus"].(float64); ok {
			addMp = int(v)
		}
		bonus["mp"] += int(math.Ceil(float64(addMp) * level))

		addAttack := 0
		if v, ok := skill["attack_bonus"].(float64); ok {
			addAttack = int(v)
		}
		bonus["attack"] += int(math.Ceil(float64(addAttack) * level))

		addDef := 0
		if v, ok := skill["defense_bonus"].(float64); ok {
			addDef = int(v)
		}
		bonus["defense"] += int(math.Ceil(float64(addDef) * level))

		addSpeed := 0
		if v, ok := skill["speed_bonus"].(float64); ok {
			addSpeed = int(v)
		}
		bonus["speed"] += int(math.Ceil(float64(addSpeed) * level))

		addHit := 0
		if v, ok := skill["hit_bonus"].(float64); ok {
			addHit = int(v)
		}
		bonus["hit"] += int(math.Ceil(float64(addHit) * level))

		addDodge := 0
		if v, ok := skill["dodge_bonus"].(float64); ok {
			addDodge = int(v)
		}
		bonus["dodge"] += int(math.Ceil(float64(addDodge) * level))

		addCrit := 0
		if v, ok := skill["crit_bonus"].(float64); ok {
			addCrit = int(v)
		}
		bonus["crit"] += int(math.Ceil(float64(addCrit) * level))
	}

	return bonus, nil
}

// ★ 实现 attribute.SkillBonusProvider 接口（打破循环依赖）
// CalculateSkillBonusWithCtx 带上下文的技能加成计算（供CalcEngine调用）
func (s *Service) CalculateSkillBonusWithCtx(ctx context.Context, roleID uint64) (*attribute.SkillBonus, error) {
	bonusMap, err := s.CalculateSkillBonus(roleID)
	if err != nil {
		return &attribute.SkillBonus{}, err
	}

	return &attribute.SkillBonus{
		Hp:      bonusMap["hp"],
		Mp:      bonusMap["mp"],
		Attack:  bonusMap["attack"],
		Defense: bonusMap["defense"],
		Speed:   bonusMap["speed"],
		Hit:     bonusMap["hit"],
		Dodge:   bonusMap["dodge"],
		Crit:    bonusMap["crit"],
	}, nil
}

// GetEquippedSkillsWithCtx 带上下文的获取已装备武学（供CalcEngine调用）
func (s *Service) GetEquippedSkillsWithCtx(ctx context.Context, roleID uint64) ([]map[string]interface{}, error) {
	return s.GetEquippedSkills(roleID)
}

// ForgetSkill 遗忘武学（需谨慎使用）
func (s *Service) ForgetSkill(roleID uint64, skillID uint32) error {
	return common.DBSkillForget(roleID, skillID)
}

// GetSkillExpProgress 获取武学熟练度进度
func (s *Service) GetSkillExpProgress(roleID uint64, skillID uint32) (currentExp, expNeeded, level, maxLevel int64, err error) {
	skills, err := common.DBSkillGetList(roleID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	for _, skill := range skills {
		if sid, ok := skill["skill_id"].(float64); ok && uint32(sid) == skillID {
			level = int64(getFloatValue(skill, "level"))
			currentExp = int64(getFloatValue(skill, "exp"))
			expFactor := int64(getFloatValue(skill, "exp_factor"))
			maxLevel = int64(getFloatValue(skill, "max_level"))
			expNeeded = expFactor * level * level
			return currentExp, expNeeded, level, maxLevel, nil
		}
	}

	err = errors.New("未学习该武学")
	return 0, 0, 0, 0, err
}

// CanLearnSkillByLevel 检查角色等级是否满足武学学习条件
func (s *Service) CanLearnSkillByLevel(roleLevel uint32, skillID uint32) (bool, error) {
	// 使用本地配置（与 /api/skill/base/list 同源）
	config := common.GetSkillConfig(skillID)
	if config == nil {
		return false, errors.New("武学不存在")
	}

	if roleLevel < config.Level {
		return false, fmt.Errorf("需要等级%d, 当前等级%d", config.Level, roleLevel)
	}

	return true, nil
}

// getFloatValue 安全获取float64值
func getFloatValue(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

// SkillBaseInfo 武学基础信息(用于内部转换)
type SkillBaseInfo struct {
	ID          uint32 `json:"id"`
	Name        string `json:"name"`
	Type        uint8  `json:"type"`
	Level       uint32 `json:"level"`
	ExpFactor   int    `json:"exp_factor"`
	MaxLevel    uint32 `json:"max_level"`
	LevelReq    uint32 `json:"level_req"`
	HpBonus     int    `json:"hp_bonus"`
	MpBonus     int    `json:"mp_bonus"`
	AttackBonus int    `json:"attack_bonus"`
	DefBonus    int    `json:"defense_bonus"`
	SpeedBonus  int    `json:"speed_bonus"`
	HitBonus    int    `json:"hit_bonus"`
	DodgeBonus  int    `json:"dodge_bonus"`
	CritBonus   int    `json:"crit_bonus"`
	IsActive    bool   `json:"is_active"`
}

// RoleSkillInfo 角色武学信息(用于内部转换)
type RoleSkillInfo struct {
	ID      uint64 `json:"id"`
	RoleID  uint64 `json:"role_id"`
	SkillID uint32 `json:"skill_id"`
	Level   int    `json:"level"`
	Exp     int64  `json:"exp"`
	IsEquip uint8  `json:"is_equip"`
}

// parseSkillBase 解析武学基础信息
func parseSkillBase(data map[string]interface{}) SkillBaseInfo {
	info := SkillBaseInfo{}
	if v, ok := data["id"].(float64); ok {
		info.ID = uint32(v)
	}
	if v, ok := data["name"].(string); ok {
		info.Name = v
	}
	if v, ok := data["type"].(float64); ok {
		info.Type = uint8(v)
	}
	if v, ok := data["level"].(float64); ok {
		info.Level = uint32(v)
	}
	if v, ok := data["exp_factor"].(float64); ok {
		info.ExpFactor = int(v)
	}
	if v, ok := data["max_level"].(float64); ok {
		info.MaxLevel = uint32(v)
	}
	if v, ok := data["level_req"].(float64); ok {
		info.LevelReq = uint32(v)
	}
	if v, ok := data["hp_bonus"].(float64); ok {
		info.HpBonus = int(v)
	}
	if v, ok := data["mp_bonus"].(float64); ok {
		info.MpBonus = int(v)
	}
	if v, ok := data["attack_bonus"].(float64); ok {
		info.AttackBonus = int(v)
	}
	if v, ok := data["defense_bonus"].(float64); ok {
		info.DefBonus = int(v)
	}
	if v, ok := data["speed_bonus"].(float64); ok {
		info.SpeedBonus = int(v)
	}
	if v, ok := data["hit_bonus"].(float64); ok {
		info.HitBonus = int(v)
	}
	if v, ok := data["dodge_bonus"].(float64); ok {
		info.DodgeBonus = int(v)
	}
	if v, ok := data["crit_bonus"].(float64); ok {
		info.CritBonus = int(v)
	}
	return info
}

// parseRoleSkill 解析角色武学信息
func parseRoleSkill(data map[string]interface{}) RoleSkillInfo {
	info := RoleSkillInfo{}
	if v, ok := data["id"].(float64); ok {
		info.ID = uint64(v)
	}
	if v, ok := data["role_id"].(float64); ok {
		info.RoleID = uint64(v)
	}
	if v, ok := data["skill_id"].(float64); ok {
		info.SkillID = uint32(v)
	}
	if v, ok := data["level"].(float64); ok {
		info.Level = int(v)
	}
	if v, ok := data["exp"].(float64); ok {
		info.Exp = int64(v)
	}
	if v, ok := data["is_equip"].(float64); ok {
		info.IsEquip = uint8(v)
	}
	return info
}

// 避免导入未使用的encoding/json
var _ = json.Marshal
