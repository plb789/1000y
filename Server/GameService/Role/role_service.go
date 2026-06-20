package role

import (
	"encoding/json"
	"errors"
	"fmt"
	common "game-server/Common"
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
func (s *Service) CreateRole(req RoleCreateRequest) (*Role, error) {
	// 检查角色名是否已存在
	existingRole, err := DBRoleGetByName(req.Name)
	if err == nil && existingRole != nil && existingRole.ID > 0 {
		return nil, errors.New("角色名已被占用")
	}

	// 检查账号是否已有角色(每个账号最多创建3个角色)
	roles, err := DBRoleList(req.AccountID)
	if err != nil {
		return nil, err
	}
	if len(roles) >= 3 {
		return nil, errors.New("每个账号最多创建3个角色")
	}

	// 通过DBService API创建角色
	roleID, err := common.DBRoleCreate(common.RoleCreateRequest{
		AccountID:  req.AccountID,
		Name:       req.Name,
		Gender:     req.Gender,
		Appearance: req.Appearance,
	})
	if err != nil {
		return nil, err
	}

	// 获取创建的角色信息
	return s.GetRoleByID(roleID)
}

// GetRoleByID 根据ID获取角色
func (s *Service) GetRoleByID(roleID uint64) (*Role, error) {
	roleInfo, err := common.DBRoleGet(roleID)
	if err != nil {
		return nil, err
	}

	return &Role{
		ID:         roleInfo.ID,
		AccountID:  roleInfo.AccountID,
		Name:       roleInfo.Name,
		Level:      roleInfo.Level,
		Exp:        roleInfo.Exp,
		Gold:       roleInfo.Gold,
		BindGold:   roleInfo.BindGold,
		Yuanbao:    int(roleInfo.Yuanbao),
		Gender:     roleInfo.Gender,
		Appearance: 0,
		Hp:         roleInfo.Hp,
		MaxHp:      roleInfo.MaxHp,
		Mp:         roleInfo.Mp,
		MaxMp:      roleInfo.MaxMp,
		Stamina:    100,
		MaxStamina: 100,
		Attack:     roleInfo.Attack,
		Defense:    roleInfo.Defense,
		Speed:      roleInfo.Speed,
		Hit:        roleInfo.Hit,
		Dodge:      roleInfo.Dodge,
		Crit:       roleInfo.Crit,
		CritDamage: roleInfo.CritDamage,
		MapID:      int(roleInfo.MapID),
		MapX:       roleInfo.MapX,
		MapY:       roleInfo.MapY,
		Status:     roleInfo.Status,
	}, nil
}

// GetRoleByName 根据名称获取角色
func (s *Service) GetRoleByName(name string) (*Role, error) {
	roleInfo, err := DBRoleGetByName(name)
	if err != nil {
		return nil, err
	}
	return s.GetRoleByID(roleInfo.ID)
}

// DBRoleGetByName 根据名称获取角色(ID用0表示不存在)
type RoleNameInfo struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

func DBRoleGetByName(name string) (*RoleNameInfo, error) {
	resp, err := common.DBPost("/api/role/get_by_name", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取角色失败: %v", resp["msg"])
	}

	if resp["data"] == nil {
		return &RoleNameInfo{ID: 0}, nil
	}

	data, _ := json.Marshal(resp["data"])
	var role RoleNameInfo
	json.Unmarshal(data, &role)

	return &role, nil
}

// DBRoleList 获取账号下所有角色
func DBRoleList(accountID uint64) ([]map[string]interface{}, error) {
	resp, err := common.DBPost("/api/role/list", map[string]uint64{"account_id": accountID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取角色列表失败: %v", resp["msg"])
	}

	if resp["data"] == nil {
		return []map[string]interface{}{}, nil
	}

	data, _ := json.Marshal(resp["data"])
	var roles []map[string]interface{}
	json.Unmarshal(data, &roles)

	return roles, nil
}

// GetRolesByAccount 获取账号下所有角色
func (s *Service) GetRolesByAccount(accountID uint64) ([]RoleBrief, error) {
	roles, err := DBRoleList(accountID)
	if err != nil {
		return nil, err
	}

	result := make([]RoleBrief, 0, len(roles))
	for _, r := range roles {
		role := RoleBrief{}
		if v, ok := r["id"].(float64); ok {
			role.ID = uint64(v)
		}
		if v, ok := r["name"].(string); ok {
			role.Name = v
		}
		if v, ok := r["level"].(float64); ok {
			role.Level = uint32(v)
		}
		if v, ok := r["gender"].(float64); ok {
			role.Gender = uint8(v)
		}
		if v, ok := r["appearance"].(float64); ok {
			role.Appearance = uint32(v)
		}
		if v, ok := r["map_id"].(float64); ok {
			role.MapID = int(v)
		}
		if v, ok := r["title"].(string); ok {
			role.Title = v
		}
		result = append(result, role)
	}

	return result, nil
}

// UpdateRole 更新角色信息
func (s *Service) UpdateRole(roleID uint64, req RoleUpdateRequest) error {
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

	// 通过DBService API更新
	resp, err := common.DBPost("/api/role/update", map[string]interface{}{
		"id":      roleID,
		"updates": updates,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("更新角色失败: %v", resp["msg"])
	}

	return nil
}

// UpdateRoleAttributes 批量更新角色属性
func (s *Service) UpdateRoleAttributes(roleID uint64, req RoleAttributeRequest) error {
	dbReq := common.RoleAttributeRequest{}
	if req.Hp != nil {
		dbReq.Hp = req.Hp
	}
	if req.MaxHp != nil {
		dbReq.MaxHp = req.MaxHp
	}
	if req.Mp != nil {
		dbReq.Mp = req.Mp
	}
	if req.MaxMp != nil {
		dbReq.MaxMp = req.MaxMp
	}
	if req.Attack != nil {
		dbReq.Attack = req.Attack
	}
	if req.Defense != nil {
		dbReq.Defense = req.Defense
	}
	if req.Speed != nil {
		dbReq.Speed = req.Speed
	}
	if req.Hit != nil {
		dbReq.Hit = req.Hit
	}
	if req.Dodge != nil {
		dbReq.Dodge = req.Dodge
	}
	if req.Crit != nil {
		dbReq.Crit = req.Crit
	}
	if req.Gold != nil {
		gold := int64(*req.Gold)
		dbReq.Gold = &gold
	}
	if req.BindGold != nil {
		bindGold := int64(*req.BindGold)
		dbReq.BindGold = &bindGold
	}
	if req.Yuanbao != nil {
		yuanbao := int64(*req.Yuanbao)
		dbReq.Yuanbao = &yuanbao
	}

	return common.DBRoleUpdateAttributes(roleID, dbReq)
}

// DeleteRole 删除角色
func (s *Service) DeleteRole(roleID uint64, accountID uint64) error {
	return common.DBRoleDelete(roleID, accountID)
}

// AddExp 增加经验值(自动处理升级) - 游戏逻辑在GameService中处理
func (s *Service) AddExp(roleID uint64, exp int64) (bool, uint32, int64, error) {
	leveledUp, level, newExp, err := common.DBRoleAddExp(roleID, exp)
	return leveledUp, uint32(level), newExp, err
}

// AddGold 增加金币
func (s *Service) AddGold(roleID uint64, gold int64) error {
	return common.DBRoleAddGold(roleID, gold)
}

// ConsumeGold 消耗金币
func (s *Service) ConsumeGold(roleID uint64, gold int64) error {
	return common.DBRoleConsumeGold(roleID, gold)
}

// ChangeHP 改变生命值
func (s *Service) ChangeHP(roleID uint64, change int) (int, error) {
	return common.DBRoleChangeHP(roleID, change)
}

// ChangeMP 改变内力值
func (s *Service) ChangeMP(roleID uint64, change int) (int, error) {
	return common.DBRoleChangeMP(roleID, change)
}

// ChangeStamina 改变体力值
func (s *Service) ChangeStamina(roleID uint64, change int) (int, error) {
	return common.DBRoleChangeStamina(roleID, change)
}

// ChangeMap 切换地图
func (s *Service) ChangeMap(roleID uint64, mapID int, x int, y int) error {
	return common.DBRoleChangeMap(roleID, mapID, x, y)
}

// UpdatePosition 更新位置
func (s *Service) UpdatePosition(roleID uint64, x int, y int) error {
	// 先获取当前角色信息
	role, err := s.GetRoleByID(roleID)
	if err != nil {
		return err
	}
	return common.DBRoleChangeMap(roleID, role.MapID, x, y)
}

// SetStatus 设置角色状态
func (s *Service) SetStatus(roleID uint64, status uint8) error {
	return common.DBRoleSetStatus(roleID, status)
}

// SetPKMode 设置PK模式
func (s *Service) SetPKMode(roleID uint64, mode uint8) error {
	if mode > 3 {
		return errors.New("无效的PK模式")
	}
	return common.DBRoleSetPKMode(roleID, mode)
}

// UpdatePkValue 更新善恶值
func (s *Service) UpdatePkValue(roleID uint64, change int) error {
	return common.DBRoleUpdatePKValue(roleID, change)
}

// RecordKill 记录击杀
func (s *Service) RecordKill(roleID uint64) error {
	return common.DBRoleRecordKill(roleID)
}

// RecordDeath 记录死亡
func (s *Service) RecordDeath(roleID uint64) error {
	return common.DBRoleRecordDeath(roleID)
}

// FullRecovery 完全恢复(满血满蓝)
func (s *Service) FullRecovery(roleID uint64) error {
	return common.DBRoleFullRecovery(roleID)
}

// SaveRole 保存角色(手动存档)
func (s *Service) SaveRole(roleID uint64) error {
	// 通过DBService记录保存时间
	resp, err := common.DBPost("/api/role/save", map[string]uint64{"id": roleID})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("保存角色失败: %v", resp["msg"])
	}

	return nil
}

// LoginRecord 记录登录
func (s *Service) LoginRecord(roleID uint64, ip string) error {
	return common.DBRoleLoginRecord(roleID, ip)
}

// LogoutRecord 记录登出
func (s *Service) LogoutRecord(roleID uint64) error {
	return common.DBRoleLogoutRecord(roleID)
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
