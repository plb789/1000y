package role

import (
	"errors"
	"game-server/DBService/mysql"
	"game-server/GameService/Role/model"
	"sync"
	"time"
)

// OnlinePlayer 在线玩家信息
type OnlinePlayer struct {
	UID       uint64
	RoleID    uint64
	TileX     int
	TileY     int
	Status    uint8
	LoginTime time.Time
	LastSave  time.Time
}

// Service 角色服务
type Service struct {
	onlinePlayers sync.Map // map[uint64]*OnlinePlayer (key=roleID)
}

// NewService 创建角色服务实例
func NewService() *Service {
	return &Service{}
}

// CreateRole 创建角色
func (s *Service) CreateRole(req model.RoleCreateRequest) (*model.Role, error) {
	// 检查角色名是否已存在
	var count int64
	mysql.DB.Model(&model.Role{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("角色名已被占用")
	}

	// 检查账号是否已有角色(每个账号最多创建3个角色)
	mysql.DB.Model(&model.Role{}).Where("account_id = ?", req.AccountID).Count(&count)
	if count >= 3 {
		return nil, errors.New("每个账号最多创建3个角色")
	}

	// 创建角色
	role := &model.Role{
		AccountID:  req.AccountID,
		Name:       req.Name,
		Level:      1,
		Exp:        0,
		Gold:       100, // 初始金币
		BindGold:   0,
		Yuanbao:    0,
		Gender:     req.Gender,
		Appearance: req.Appearance,
		Hp:         100,
		MaxHp:      100,
		Mp:         100,
		MaxMp:      100,
		Stamina:    100,
		MaxStamina: 100,
		Attack:     10,
		Defense:    5,
		Speed:      10,
		Hit:        50,
		Dodge:      10,
		Crit:       5,
		CritDamage: 150,
		MapID:      1,
		MapX:       100,
		MapY:       100,
		PkMode:     0,
		PkValue:    0,
		Status:     0,
		CreateTime: time.Now(),
	}

	if err := mysql.DB.Create(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}

// GetRoleByID 根据ID获取角色
func (s *Service) GetRoleByID(roleID uint64) (*model.Role, error) {
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (s *Service) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	if err := mysql.DB.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRolesByAccount 获取账号下所有角色
func (s *Service) GetRolesByAccount(accountID uint64) ([]model.RoleBrief, error) {
	var roles []model.RoleBrief
	err := mysql.DB.Table("role").
		Select("id, name, level, gender, appearance, map_id, title").
		Where("account_id = ?", accountID).
		Order("create_time DESC").
		Find(&roles).Error
	return roles, err
}

// UpdateRole 更新角色信息
func (s *Service) UpdateRole(roleID uint64, req model.RoleUpdateRequest) error {
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Gender != 0 && req.Gender <= 1 {
		updates["gender"] = req.Gender
	}
	if req.Exp > 0 {
		updates["exp"] = req.Exp
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}

	if len(updates) == 0 {
		return errors.New("没有需要更新的字段")
	}

	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

// UpdateRoleAttributes 批量更新角色属性
func (s *Service) UpdateRoleAttributes(roleID uint64, req model.RoleAttributeRequest) error {
	updates := make(map[string]interface{})
	
	if req.Hp != nil {
		updates["hp"] = *req.Hp
	}
	if req.MaxHp != nil {
		updates["max_hp"] = *req.MaxHp
	}
	if req.Mp != nil {
		updates["mp"] = *req.Mp
	}
	if req.MaxMp != nil {
		updates["max_mp"] = *req.MaxMp
	}
	if req.Attack != nil {
		updates["attack"] = *req.Attack
	}
	if req.Defense != nil {
		updates["defense"] = *req.Defense
	}
	if req.Speed != nil {
		updates["speed"] = *req.Speed
	}
	if req.Hit != nil {
		updates["hit"] = *req.Hit
	}
	if req.Dodge != nil {
		updates["dodge"] = *req.Dodge
	}
	if req.Crit != nil {
		updates["crit"] = *req.Crit
	}
	if req.Gold != nil {
		updates["gold"] = *req.Gold
	}
	if req.BindGold != nil {
		updates["bind_gold"] = *req.BindGold
	}
	if req.Yuanbao != nil {
		updates["yuanbao"] = *req.Yuanbao
	}

	if len(updates) == 0 {
		return errors.New("没有需要更新的属性")
	}

	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

// DeleteRole 删除角色
func (s *Service) DeleteRole(roleID uint64, accountID uint64) error {
	result := mysql.DB.Where("id = ? AND account_id = ?", roleID, accountID).Delete(&model.Role{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("角色不存在或无权删除")
	}
	return nil
}

// AddExp 增加经验值(自动处理升级)
func (s *Service) AddExp(roleID uint64, exp int64) (bool, uint32, int64, error) {
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return false, 0, 0, err
	}

	leveledUp := false
	currentLevel := role.Level

	role.Exp += exp

	// 计算升级所需经验: level * 100 * level
	for {
		expNeeded := int64(currentLevel) * 100 * currentLevel
		if role.Exp >= expNeeded && currentLevel < 200 { // 最高200级
			role.Exp -= expNeeded
			currentLevel++
			leveledUp = true
		} else {
			break
		}
	}

	// 更新角色
	role.Level = currentLevel
	if err := mysql.DB.Save(&role).Error; err != nil {
		return false, 0, 0, err
	}

	return leveledUp, currentLevel, role.Exp, nil
}

// AddGold 增加金币
func (s *Service) AddGold(roleID uint64, gold int64) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Update("gold", mysql.DB.Raw("gold + ?", gold)).Error
}

// ConsumeGold 消耗金币
func (s *Service) ConsumeGold(roleID uint64, gold int64) error {
	// 先检查金币是否足够
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return err
	}
	if role.Gold < gold {
		return errors.New("金币不足")
	}
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Update("gold", mysql.DB.Raw("gold - ?", gold)).Error
}

// ChangeHP 改变生命值
func (s *Service) ChangeHP(roleID uint64, change int) (int, error) {
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return 0, err
	}

	role.Hp += change
	if role.Hp > role.MaxHp {
		role.Hp = role.MaxHp
	}
	if role.Hp < 0 {
		role.Hp = 0
	}

	if err := mysql.DB.Save(&role).Error; err != nil {
		return 0, err
	}
	return role.Hp, nil
}

// ChangeMP 改变内力值
func (s *Service) ChangeMP(roleID uint64, change int) (int, error) {
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return 0, err
	}

	role.Mp += change
	if role.Mp > role.MaxMp {
		role.Mp = role.MaxMp
	}
	if role.Mp < 0 {
		role.Mp = 0
	}

	if err := mysql.DB.Save(&role).Error; err != nil {
		return 0, err
	}
	return role.Mp, nil
}

// ChangeStamina 改变体力值
func (s *Service) ChangeStamina(roleID uint64, change int) (int, error) {
	var role model.Role
	if err := mysql.DB.Where("id = ?", roleID).First(&role).Error; err != nil {
		return 0, err
	}

	role.Stamina += change
	if role.Stamina > role.MaxStamina {
		role.Stamina = role.MaxStamina
	}
	if role.Stamina < 0 {
		role.Stamina = 0
	}

	if err := mysql.DB.Save(&role).Error; err != nil {
		return 0, err
	}
	return role.Stamina, nil
}

// ChangeMap 切换地图
func (s *Service) ChangeMap(roleID uint64, mapID int, x int, y int) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"map_id": mapID,
			"map_x":  x,
			"map_y":  y,
		}).Error
}

// UpdatePosition 更新位置
func (s *Service) UpdatePosition(roleID uint64, x int, y int) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"map_x": x,
			"map_y": y,
		}).Error
}

// SetStatus 设置角色状态
func (s *Service) SetStatus(roleID uint64, status uint8) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).Update("status", status).Error
}

// SetPKMode 设置PK模式
func (s *Service) SetPKMode(roleID uint64, mode uint8) error {
	if mode > 3 {
		return errors.New("无效的PK模式")
	}
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).Update("pk_mode", mode).Error
}

// UpdatePkValue 更新善恶值
func (s *Service) UpdatePkValue(roleID uint64, change int) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Update("pk_value", mysql.DB.Raw("pk_value + ?", change)).Error
}

// RecordKill 记录击杀
func (s *Service) RecordKill(roleID uint64) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"kill_count": mysql.DB.Raw("kill_count + 1"),
			"pk_value":   mysql.DB.Raw("pk_value + 50"), // 增加50点PK值
		}).Error
}

// RecordDeath 记录死亡
func (s *Service) RecordDeath(roleID uint64) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Update("death_count", mysql.DB.Raw("death_count + 1")).Error
}

// FullRecovery 完全恢复(满血满蓝)
func (s *Service) FullRecovery(roleID uint64) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"hp":      mysql.DB.Raw("max_hp"),
			"mp":      mysql.DB.Raw("max_mp"),
			"stamina": mysql.DB.Raw("max_stamina"),
		}).Error
}

// SaveRole 保存角色(手动存档)
func (s *Service) SaveRole(roleID uint64) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Update("last_save_time", time.Now()).Error
}

// LoginRecord 记录登录
func (s *Service) LoginRecord(roleID uint64, ip string) error {
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"status":       3, // 在线状态
			"last_login":   time.Now(),
			"last_login_ip": ip,
		}).Error
}

// LogoutRecord 记录登出
func (s *Service) LogoutRecord(roleID uint64) error {
	now := time.Now()
	return mysql.DB.Model(&model.Role{}).Where("id = ?", roleID).
		Updates(map[string]interface{}{
			"status":       0,
			"logout_time":  now,
			"last_save_time": now,
		}).Error
}

// OnlinePlayerMgr 在线玩家管理

// PlayerLogin 玩家上线
func (s *Service) PlayerLogin(roleID uint64) *OnlinePlayer {
	player := &OnlinePlayer{
		RoleID:    roleID,
		LoginTime: time.Now(),
		LastSave:  time.Now(),
	}
	s.onlinePlayers.Store(roleID, player)
	return player
}

// PlayerLogout 玩家下线
func (s *Service) PlayerLogout(roleID uint64) {
	s.onlinePlayers.Delete(roleID)
}

// GetOnlinePlayer 获取在线玩家
func (s *Service) GetOnlinePlayer(roleID uint64) (*OnlinePlayer, bool) {
	val, ok := s.onlinePlayers.Load(roleID)
	if !ok {
		return nil, false
	}
	return val.(*OnlinePlayer), true
}

// GetAllOnlinePlayers 获取所有在线玩家
func (s *Service) GetAllOnlinePlayers() map[uint64]*OnlinePlayer {
	result := make(map[uint64]*OnlinePlayer)
	s.onlinePlayers.Range(func(key, value interface{}) bool {
		result[key.(uint64)] = value.(*OnlinePlayer)
		return true
	})
	return result
}

// GetOnlineCount 获取在线人数
func (s *Service) GetOnlineCount() int {
	count := 0
	s.onlinePlayers.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// UpdatePlayerPosition 更新在线玩家位置
func (s *Service) UpdatePlayerPosition(roleID uint64, x, y int) {
	if player, ok := s.GetOnlinePlayer(roleID); ok {
		player.TileX = x
		player.TileY = y
	}
}

// UpdatePlayerStatus 更新在线玩家状态
func (s *Service) UpdatePlayerStatus(roleID uint64, status uint8) {
	if player, ok := s.GetOnlinePlayer(roleID); ok {
		player.Status = status
	}
}
