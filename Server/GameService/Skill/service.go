package skill

import (
	"errors"
	"game-server/DBService/mysql"
	"game-server/GameService/Skill/model"
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
	// 检查武学是否存在
	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ? AND is_active = 1", skillID).First(&skillBase).Error; err != nil {
		return errors.New("武学不存在或已下架")
	}

	// 检查角色是否已学习该武学
	var existCount int64
	mysql.DB.Model(&model.RoleSkill{}).Where("role_id = ? AND skill_id = ?", roleID, skillID).Count(&existCount)
	if existCount > 0 {
		return errors.New("已学习该武学")
	}

	// 创建角色武学记录
	roleSkill := model.RoleSkill{
		RoleID:   roleID,
		SkillID:  skillID,
		Level:    1,
		Exp:      0,
		IsEquip:  0,
	}
	return mysql.DB.Create(&roleSkill).Error
}

// GetRoleSkills 获取角色所有武学
func (s *Service) GetRoleSkills(roleID uint64) ([]model.RoleSkillWithBase, error) {
	var skills []model.RoleSkillWithBase
	err := mysql.DB.Table("role_skill").
		Select("role_skill.*, skill_base.*").
		Joins("LEFT JOIN skill_base ON role_skill.skill_id = skill_base.id").
		Where("role_skill.role_id = ?", roleID).
		Order("skill_base.type ASC, role_skill.level DESC").
		Find(&skills).Error
	return skills, err
}

// GetRoleSkillsByType 获取角色指定类型的武学
func (s *Service) GetRoleSkillsByType(roleID uint64, skillType uint8) ([]model.RoleSkillWithBase, error) {
	var skills []model.RoleSkillWithBase
	err := mysql.DB.Table("role_skill").
		Select("role_skill.*, skill_base.*").
		Joins("LEFT JOIN skill_base ON role_skill.skill_id = skill_base.id").
		Where("role_skill.role_id = ? AND skill_base.type = ?", roleID, skillType).
		Order("role_skill.level DESC").
		Find(&skills).Error
	return skills, err
}

// GetSkillBase 获取武学基础信息
func (s *Service) GetSkillBase(skillID uint32) (*model.SkillBase, error) {
	var skill model.SkillBase
	err := mysql.DB.Where("id = ? AND is_active = 1", skillID).First(&skill).Error
	if err != nil {
		return nil, err
	}
	return &skill, nil
}

// GetAllSkillBase 获取所有武学基础信息
func (s *Service) GetAllSkillBase() ([]model.SkillBase, error) {
	var skills []model.SkillBase
	err := mysql.DB.Where("is_active = 1").Order("type ASC, level ASC").Find(&skills).Error
	return skills, err
}

// GetSkillBaseByType 获取指定类型的所有武学
func (s *Service) GetSkillBaseByType(skillType uint8) ([]model.SkillBase, error) {
	var skills []model.SkillBase
	err := mysql.DB.Where("type = ? AND is_active = 1", skillType).Order("level ASC").Find(&skills).Error
	return skills, err
}

// AddExp 增加武学熟练度
// roleID: 角色ID
// skillID: 武学ID
// exp: 增加的熟练度
// 返回: 是否升级及错误信息
func (s *Service) AddExp(roleID uint64, skillID uint32, exp int64) (bool, int, error) {
	// 获取角色武学信息
	var roleSkill model.RoleSkill
	if err := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).First(&roleSkill).Error; err != nil {
		return false, 0, errors.New("未学习该武学")
	}

	// 获取武学基础信息
	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ?", skillID).First(&skillBase).Error; err != nil {
		return false, 0, errors.New("武学不存在")
	}

	// 检查是否已达最高等级
	if roleSkill.Level >= skillBase.MaxLevel {
		return false, roleSkill.Level, nil
	}

	// 增加熟练度
	roleSkill.Exp += exp

	// 计算升级所需熟练度
	// 公式: 升级经验 = exp_factor * level * level
	expNeeded := int64(skillBase.ExpFactor) * int64(roleSkill.Level) * int64(roleSkill.Level)

	// 检查是否升级
	leveledUp := false
	currentLevel := roleSkill.Level
	for roleSkill.Exp >= expNeeded && roleSkill.Level < skillBase.MaxLevel {
		roleSkill.Exp -= expNeeded
		roleSkill.Level++
		leveledUp = true
		currentLevel = roleSkill.Level
		// 下一级所需经验
		expNeeded = int64(skillBase.ExpFactor) * int64(roleSkill.Level) * int64(roleSkill.Level)
	}

	// 保存更新
	if err := mysql.DB.Save(&roleSkill).Error; err != nil {
		return false, 0, err
	}

	return leveledUp, currentLevel, nil
}

// UpgradeSkill 手动升级武学（使用道具或其他方式）
// 返回: 新等级及错误信息
func (s *Service) UpgradeSkill(roleID uint64, skillID uint32) (int, error) {
	// 获取角色武学信息
	var roleSkill model.RoleSkill
	if err := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).First(&roleSkill).Error; err != nil {
		return 0, errors.New("未学习该武学")
	}

	// 获取武学基础信息
	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ?", skillID).First(&skillBase).Error; err != nil {
		return 0, errors.New("武学不存在")
	}

	// 检查是否已达最高等级
	if roleSkill.Level >= skillBase.MaxLevel {
		return roleSkill.Level, errors.New("已达最高等级")
	}

	// 手动升级需要消耗经验,扣减当前熟练度的50%作为升级代价
	expCost := roleSkill.Exp / 2
	roleSkill.Exp -= expCost
	roleSkill.Level++

	if err := mysql.DB.Save(&roleSkill).Error; err != nil {
		return 0, err
	}

	return roleSkill.Level, nil
}

// EquipSkill 装备武学
// 千年游戏中,外功/拳法/剑法/刀法/枪法/斧法只能装备一个
// 内功/身法/护体可以各装备一个
func (s *Service) EquipSkill(roleID uint64, skillID uint32) error {
	// 获取武学基础信息
	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ?", skillID).First(&skillBase).Error; err != nil {
		return errors.New("武学不存在")
	}

	// 检查角色是否拥有该武学
	var roleSkill model.RoleSkill
	if err := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).First(&roleSkill).Error; err != nil {
		return errors.New("未学习该武学")
	}

	// 对于外功类武学(拳法/剑法/刀法/枪法/斧法),需要先卸下同类型已装备的武学
	if skillBase.Type >= 5 && skillBase.Type <= 9 {
		// 先将同类型武学全部设为未装备
		if err := mysql.DB.Model(&model.RoleSkill{}).
			Joins("LEFT JOIN skill_base ON role_skill.skill_id = skill_base.id").
			Where("role_skill.role_id = ? AND skill_base.type = ? AND role_skill.is_equip = 1", roleID, skillBase.Type).
			Update("is_equip", 0).Error; err != nil {
			return err
		}
	}

	// 装备该武学
	roleSkill.IsEquip = 1
	return mysql.DB.Save(&roleSkill).Error
}

// UnequipSkill 卸下武学
func (s *Service) UnequipSkill(roleID uint64, skillID uint32) error {
	var roleSkill model.RoleSkill
	if err := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).First(&roleSkill).Error; err != nil {
		return errors.New("未学习该武学")
	}

	roleSkill.IsEquip = 0
	return mysql.DB.Save(&roleSkill).Error
}

// GetEquippedSkills 获取角色已装备的武学
func (s *Service) GetEquippedSkills(roleID uint64) ([]model.RoleSkillWithBase, error) {
	var skills []model.RoleSkillWithBase
	err := mysql.DB.Table("role_skill").
		Select("role_skill.*, skill_base.*").
		Joins("LEFT JOIN skill_base ON role_skill.skill_id = skill_base.id").
		Where("role_skill.role_id = ? AND role_skill.is_equip = 1", roleID).
		Order("skill_base.type ASC").
		Find(&skills).Error
	return skills, err
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
		level := float64(skill.Level)
		bonus["hp"] += int(math.Ceil(float64(skill.SkillBase.HpBonus) * level))
		bonus["mp"] += int(math.Ceil(float64(skill.SkillBase.MpBonus) * level))
		bonus["attack"] += int(math.Ceil(float64(skill.SkillBase.AttackBonus) * level))
		bonus["defense"] += int(math.Ceil(float64(skill.SkillBase.DefBonus) * level))
		bonus["speed"] += int(math.Ceil(float64(skill.SkillBase.SpeedBonus) * level))
		bonus["hit"] += int(math.Ceil(float64(skill.SkillBase.HitBonus) * level))
		bonus["dodge"] += int(math.Ceil(float64(skill.SkillBase.DodgeBonus) * level))
		bonus["crit"] += int(math.Ceil(float64(skill.SkillBase.CritBonus) * level))
	}

	return bonus, nil
}

// ForgetSkill 遗忘武学（需谨慎使用）
func (s *Service) ForgetSkill(roleID uint64, skillID uint32) error {
	result := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).Delete(&model.RoleSkill{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未学习该武学")
	}
	return nil
}

// GetSkillExpProgress 获取武学熟练度进度
func (s *Service) GetSkillExpProgress(roleID uint64, skillID uint32) (currentExp, expNeeded, level, maxLevel int64, err error) {
	var roleSkill model.RoleSkill
	if err := mysql.DB.Where("role_id = ? AND skill_id = ?", roleID, skillID).First(&roleSkill).Error; err != nil {
		err = errors.New("未学习该武学")
		return
	}

	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ?", skillID).First(&skillBase).Error; err != nil {
		err = errors.New("武学不存在")
		return
	}

	level = int64(roleSkill.Level)
	maxLevel = int64(skillBase.MaxLevel)
	currentExp = roleSkill.Exp
	expNeeded = int64(skillBase.ExpFactor) * int64(roleSkill.Level) * int64(roleSkill.Level)

	return
}

// CanLearnSkillByLevel 检查角色等级是否满足武学学习条件
func (s *Service) CanLearnSkillByLevel(roleLevel uint32, skillID uint32) (bool, error) {
	var skillBase model.SkillBase
	if err := mysql.DB.Where("id = ?", skillID).First(&skillBase).Error; err != nil {
		return false, errors.New("武学不存在")
	}

	if roleLevel < skillBase.Level {
		return false, errors.New("角色等级不足")
	}

	return true, nil
}
