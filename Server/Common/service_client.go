package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// GameServiceClient 游戏服务HTTP客户端
type GameServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewGameServiceClient 创建游戏服务客户端(优化连接池)
func NewGameServiceClient(baseURL string) *GameServiceClient {
	timeout := AppConfig.GetHTTPTimeout()

	// 创建优化的Transport，支持连接复用
	transport := &http.Transport{
		// 连接池配置
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 20,               // 每个Host最大空闲连接数
		MaxConnsPerHost:     50,               // 每个Host最大连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时

		// 超时配置
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,  // 连接建立超时
			KeepAlive: 30 * time.Second, // KeepAlive时间
		}).DialContext,

		// TLS配置(如果需要)
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &GameServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// Request 请求结构
type ServiceRequest struct {
	ServiceToken string      `json:"service_token"`
	Method       string      `json:"method"`
	Params       interface{} `json:"params"`
}

// Response 响应结构
type ServiceResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Call 调用游戏服务方法
func (c *GameServiceClient) Call(method string, params interface{}) (*ServiceResponse, error) {
	reqBody := ServiceRequest{
		ServiceToken: ServiceToken,
		Method:       method,
		Params:       params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %v", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/internal", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result ServiceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("响应反序列化失败: %v", err)
	}

	return &result, nil
}

// Post 发送HTTP POST请求
func (c *GameServiceClient) Post(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %v", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return body, nil
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID         uint64 `json:"id"`
	AccountID  uint64 `json:"account_id"`
	Name       string `json:"name"`
	Level      uint32 `json:"level"`
	Exp        int64  `json:"exp"`
	Gold       int64  `json:"gold"`
	BindGold   int64  `json:"bind_gold"`
	Yuanbao    int64  `json:"yuanbao"`
	Gender     uint8  `json:"gender"`
	Appearance uint32 `json:"appearance"`
	Hp         int    `json:"hp"`
	MaxHp      int    `json:"max_hp"`
	Mp         int    `json:"mp"`
	MaxMp      int    `json:"max_mp"`
	Attack     int    `json:"attack"`
	Defense    int    `json:"defense"`
	Speed      int    `json:"speed"`
	Hit        int    `json:"hit"`
	Dodge      int    `json:"dodge"`
	Crit       int    `json:"crit"`
	CritDamage int    `json:"crit_damage"`
	MapID      uint32 `json:"map_id"`
	MapX       int    `json:"map_x"`
	MapY       int    `json:"map_y"`
	Status     uint8  `json:"status"`     // 状态: 0=正常, 1=打坐, 2=死亡, 3=在线
	PkMode     uint8  `json:"pk_mode"`    // PK模式: 0=和平, 1=队伍, 2=帮派, 3=全体
	PkValue    int    `json:"pk_value"`   // PK值（红名程度）
	KillCount  int    `json:"kill_count"` // 击杀数
	WeaponID   uint64 `json:"weapon_id"`  // 装备武器ID(0=徒手)
}

// GetRoleInfo 获取角色信息
func (c *GameServiceClient) GetRoleInfo(roleID uint64) (*RoleInfo, error) {
	resp, err := c.Call("role.get_info", map[string]uint64{"role_id": roleID})
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("获取角色信息失败: %s", resp.Msg)
	}

	var role RoleInfo
	if err := json.Unmarshal(resp.Data, &role); err != nil {
		return nil, fmt.Errorf("角色信息解析失败: %v", err)
	}

	return &role, nil
}

// UpdateRolePosition 更新角色位置
func (c *GameServiceClient) UpdateRolePosition(roleID uint64, mapID uint32, x, y int) error {
	resp, err := c.Call("role.update_position", map[string]interface{}{
		"role_id": roleID,
		"map_id":  mapID,
		"x":       x,
		"y":       y,
	})
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("更新位置失败: %s", resp.Msg)
	}

	return nil
}

// ValidateMove 验证移动
func (c *GameServiceClient) ValidateMove(roleID uint64, mapID uint32, x, y int) (bool, error) {
	resp, err := c.Call("map.validate_move", map[string]interface{}{
		"role_id": roleID,
		"map_id":  mapID,
		"x":       x,
		"y":       y,
	})
	if err != nil {
		return false, err
	}

	if resp.Code != 0 {
		return false, fmt.Errorf("验证移动失败: %s", resp.Msg)
	}

	var result struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return false, fmt.Errorf("解析响应失败: %v", err)
	}

	return result.Valid, nil
}

// 服务地址配置(从配置文件读取)
var (
	DBServiceURL    string
	LoginServiceURL string
)

// GameClient 游戏服务客户端(全局单例) - 已废弃，使用分片的ShardedGameClient
// var GameClient *GameServiceClient

// LoginClient 登录服务客户端
var LoginClient *GameServiceClient

// DBClient 数据库服务客户端
var DBClient *GameServiceClient

// InitServiceClients 初始化服务客户端(需要在LoadConfig之后调用)
func InitServiceClients() {
	DBServiceURL = AppConfig.Services.DBService
	LoginServiceURL = AppConfig.Services.LoginService

	// 如果配置为空，使用默认值
	if DBServiceURL == "" {
		DBServiceURL = "http://localhost:8083"
	}
	if LoginServiceURL == "" {
		LoginServiceURL = "http://localhost:8084"
	}

	DBClient = NewGameServiceClient(DBServiceURL)
	LoginClient = NewGameServiceClient(LoginServiceURL)
}

// ========== DBService API 客户端方法 ==========

// DBSkillRequest 武学请求
type DBSkillRequest struct {
	RoleID  uint64 `json:"role_id"`
	SkillID uint32 `json:"skill_id"`
	Exp     int64  `json:"exp"`
}

// DBSkillResponse 武学响应
type DBSkillResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

// DBItemRequest 道具请求
type DBItemRequest struct {
	RoleID    uint64 `json:"role_id"`
	BagItemID uint64 `json:"bag_item_id"`
	ItemID    uint32 `json:"item_id"`
	GridIndex int    `json:"grid_index"`
	Count     uint32 `json:"count"`
	IsBind    uint8  `json:"is_bind"`
	EquipType uint8  `json:"equip_type"`
}

// DBItemResponse 道具响应
type DBItemResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data,omitempty"`
}

// DBPost 调用 DBService API (通用方法)
func DBPost(path string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("请求序列化失败: %v", err)
	}

	resp, err := DBClient.HTTPClient.Post(DBClient.BaseURL+path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("响应反序列化失败: %v", err)
	}

	return result, nil
}

// ========== 角色相关 API ==========

// DBRoleRequest 角色请求
type DBRoleRequest struct {
	ID     uint64 `json:"id"`
	MapID  int    `json:"map_id"`
	MapX   int    `json:"map_x"`
	MapY   int    `json:"map_y"`
	Hp     *int   `json:"hp,omitempty"`
	MaxHp  *int   `json:"max_hp,omitempty"`
	Mp     *int   `json:"mp,omitempty"`
	MaxMp  *int   `json:"max_mp,omitempty"`
	Attack *int   `json:"attack,omitempty"`
	Exp    *int64 `json:"exp,omitempty"`
	Gold   *int64 `json:"gold,omitempty"`
	Change int    `json:"change"`
	IP     string `json:"ip"`
}

// DBRoleGet 获取角色
func DBRoleGet(roleID uint64) (*RoleInfo, error) {
	resp, err := DBPost("/api/role/get", map[string]uint64{"id": roleID})
	if err != nil {
		return nil, fmt.Errorf("DBService请求失败: %v", err)
	}

	code, ok := resp["code"].(float64)
	if !ok {
		return nil, fmt.Errorf("DBService响应格式错误: %v", resp)
	}
	if code != 0 {
		msg, _ := resp["msg"].(string)
		return nil, fmt.Errorf("获取角色失败: %s", msg)
	}

	data, _ := json.Marshal(resp["data"])
	var role RoleInfo
	if err := json.Unmarshal(data, &role); err != nil {
		return nil, fmt.Errorf("角色信息解析失败: %v", err)
	}

	return &role, nil
}

// DBRoleUpdatePosition 更新角色位置
func DBRoleUpdatePosition(roleID uint64, mapID int, mapX, mapY int) error {
	resp, err := DBPost("/api/role/update_position", map[string]interface{}{
		"id":     roleID,
		"map_id": mapID,
		"map_x":  mapX,
		"map_y":  mapY,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("更新位置失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleChangeHP 改变生命值
func DBRoleChangeHP(roleID uint64, change int) (int, error) {
	resp, err := DBPost("/api/role/change_hp", map[string]interface{}{
		"id":     roleID,
		"change": change,
	})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("改变HP失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// DBRoleSetHP 设置生命值（绝对值）
func DBRoleSetHP(roleID uint64, hp int) error {
	resp, err := DBPost("/api/role/set_hp", map[string]interface{}{
		"id": roleID,
		"hp": hp,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("设置HP失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleChangeMP 改变内力值
func DBRoleChangeMP(roleID uint64, change int) (int, error) {
	resp, err := DBPost("/api/role/change_mp", map[string]interface{}{
		"id":     roleID,
		"change": change,
	})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("改变MP失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// DBRoleAddGold 增加金币
func DBRoleAddGold(roleID uint64, gold int64) error {
	resp, err := DBPost("/api/role/add_gold", map[string]interface{}{
		"id":   roleID,
		"gold": gold,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("增加金币失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleConsumeGold 消耗金币
func DBRoleConsumeGold(roleID uint64, gold int64) error {
	resp, err := DBPost("/api/role/consume_gold", map[string]interface{}{
		"id":   roleID,
		"gold": gold,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("消耗金币失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleAddExp 增加经验值
func DBRoleAddExp(roleID uint64, exp int64) (bool, int, int64, error) {
	resp, err := DBPost("/api/role/add_exp", map[string]interface{}{
		"id":  roleID,
		"exp": exp,
	})
	if err != nil {
		return false, 0, 0, err
	}

	if resp["code"].(float64) != 0 {
		return false, 0, 0, fmt.Errorf("增加经验失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var result struct {
		LeveledUp bool  `json:"leveled_up"`
		Level     int   `json:"level"`
		Exp       int64 `json:"exp"`
	}
	json.Unmarshal(data, &result)

	return result.LeveledUp, result.Level, result.Exp, nil
}

// DBRoleRecordKill 记录击杀
func DBRoleRecordKill(roleID uint64) error {
	resp, err := DBPost("/api/role/record_kill", map[string]uint64{"id": roleID})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("记录击杀失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleRecordDeath 记录死亡
func DBRoleRecordDeath(roleID uint64) error {
	resp, err := DBPost("/api/role/record_death", map[string]uint64{"id": roleID})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("记录死亡失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleSetStatus 设置角色状态
func DBRoleSetStatus(roleID uint64, status uint8) error {
	resp, err := DBPost("/api/role/set_status", map[string]interface{}{
		"id":     roleID,
		"status": status,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("设置状态失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleFullRecovery 完全恢复
func DBRoleFullRecovery(roleID uint64) error {
	resp, err := DBPost("/api/role/full_recovery", map[string]uint64{"id": roleID})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("完全恢复失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleLoginRecord 记录登录
func DBRoleLoginRecord(roleID uint64, ip string) error {
	resp, err := DBPost("/api/role/login_record", map[string]interface{}{
		"id": roleID,
		"ip": ip,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("记录登录失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleLogoutRecord 记录登出
func DBRoleLogoutRecord(roleID uint64) error {
	resp, err := DBPost("/api/role/logout_record", map[string]uint64{"id": roleID})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("记录登出失败: %v", resp["msg"])
	}

	return nil
}

// ========== 武学相关 API ==========

// DBSkillLearn 学习武学
func DBSkillLearn(roleID uint64, skillID uint32) error {
	resp, err := DBPost("/api/skill/learn", map[string]interface{}{
		"role_id":  roleID,
		"skill_id": skillID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("学习武学失败: %v", resp["msg"])
	}

	return nil
}

// DBSkillGetList 获取角色武学列表
func DBSkillGetList(roleID uint64) ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/skill/get_list", map[string]uint64{"role_id": roleID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取武学列表失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var skills []map[string]interface{}
	json.Unmarshal(data, &skills)

	return skills, nil
}

// DBSkillEquip 装备武学
func DBSkillEquip(roleID uint64, skillID uint32) error {
	resp, err := DBPost("/api/skill/equip", map[string]interface{}{
		"role_id":  roleID,
		"skill_id": skillID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("装备武学失败: %v", resp["msg"])
	}

	return nil
}

// DBSkillUnequip 卸下武学
func DBSkillUnequip(roleID uint64, skillID uint32) error {
	resp, err := DBPost("/api/skill/unequip", map[string]interface{}{
		"role_id":  roleID,
		"skill_id": skillID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("卸下武学失败: %v", resp["msg"])
	}

	return nil
}

// DBSkillAddExp 增加武学熟练度（基础版本，仅增加经验）
func DBSkillAddExp(roleID uint64, skillID uint32, exp int64) (bool, int, int64, error) {
	resp, err := DBPost("/api/skill/add_exp", map[string]interface{}{
		"role_id":  roleID,
		"skill_id": skillID,
		"exp":      exp,
	})
	if err != nil {
		return false, 0, 0, err
	}

	if resp["code"].(float64) != 0 {
		return false, 0, 0, fmt.Errorf("增加熟练度失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var result struct {
		LeveledUp bool  `json:"leveled_up"`
		Level     int   `json:"level"`
		Exp       int64 `json:"exp"`
	}
	json.Unmarshal(data, &result)

	return result.LeveledUp, result.Level, result.Exp, nil
}

// DBSkillAddExpWithLevel 增加武学熟练度（带等级更新）
// level 由 GameService 根据 skills.json 计算
func DBSkillAddExpWithLevel(roleID uint64, skillID uint32, exp int64, level int, leveledUp bool) (bool, int, int64, error) {
	resp, err := DBPost("/api/skill/add_exp", map[string]interface{}{
		"role_id":    roleID,
		"skill_id":   skillID,
		"exp":        exp,
		"level":      level,
		"leveled_up": leveledUp,
	})
	if err != nil {
		return false, 0, 0, err
	}

	if resp["code"].(float64) != 0 {
		return false, 0, 0, fmt.Errorf("增加熟练度失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var result struct {
		LeveledUp bool  `json:"leveled_up"`
		Level     int   `json:"level"`
		Exp       int64 `json:"exp"`
	}
	json.Unmarshal(data, &result)

	return result.LeveledUp, result.Level, result.Exp, nil
}

// DBSkillForget 遗忘武学
func DBSkillForget(roleID uint64, skillID uint32) error {
	resp, err := DBPost("/api/skill/forget", map[string]interface{}{
		"role_id":  roleID,
		"skill_id": skillID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("遗忘武学失败: %v", resp["msg"])
	}

	return nil
}

// DBSkillGetBase 获取武学基础信息
func DBSkillGetBase(skillID uint32) (map[string]interface{}, error) {
	resp, err := DBPost("/api/skill/get_base", map[string]uint32{"skill_id": skillID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取武学信息失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var skill map[string]interface{}
	json.Unmarshal(data, &skill)

	return skill, nil
}

// ========== 道具相关 API ==========

// DBItemAdd 添加道具
func DBItemAdd(roleID uint64, itemID uint32, count uint32, isBind uint8) (int, error) {
	resp, err := DBPost("/api/item/add", map[string]interface{}{
		"role_id": roleID,
		"item_id": itemID,
		"count":   count,
		"is_bind": isBind,
	})
	if err != nil {
		return -1, err
	}

	if resp["code"].(float64) != 0 {
		return -1, fmt.Errorf("添加道具失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// DBItemGetBag 获取背包物品
func DBItemGetBag(roleID uint64) ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/item/get_bag", map[string]uint64{"role_id": roleID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取背包失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var items []map[string]interface{}
	json.Unmarshal(data, &items)

	return items, nil
}

// DBItemMove 移动物品
func DBItemMove(roleID uint64, fromGrid, toGrid int) error {
	resp, err := DBPost("/api/item/move", map[string]interface{}{
		"role_id":   roleID,
		"from_grid": fromGrid,
		"to_grid":   toGrid,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("移动物品失败: %v", resp["msg"])
	}

	return nil
}

// DBItemUse 使用道具
func DBItemUse(roleID uint64, gridIndex int) error {
	resp, err := DBPost("/api/item/use", map[string]interface{}{
		"role_id":    roleID,
		"grid_index": gridIndex,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("使用道具失败: %v", resp["msg"])
	}

	return nil
}

// DBItemDiscard 丢弃道具
func DBItemDiscard(roleID uint64, gridIndex int) error {
	resp, err := DBPost("/api/item/discard", map[string]interface{}{
		"role_id":    roleID,
		"grid_index": gridIndex,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("丢弃道具失败: %v", resp["msg"])
	}

	return nil
}

// DBItemSell 出售道具
func DBItemSell(roleID uint64, gridIndex int) (int, error) {
	resp, err := DBPost("/api/item/sell", map[string]interface{}{
		"role_id":    roleID,
		"grid_index": gridIndex,
	})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("出售道具失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// DBItemEquip 穿戴装备
func DBItemEquip(roleID uint64, bagItemID uint64) error {
	resp, err := DBPost("/api/item/equip", map[string]interface{}{
		"role_id":     roleID,
		"bag_item_id": bagItemID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("穿戴装备失败: %v", resp["msg"])
	}

	return nil
}

// DBItemUnequip 卸下装备
func DBItemUnequip(roleID uint64, equipType uint8) error {
	resp, err := DBPost("/api/item/unequip", map[string]interface{}{
		"role_id":    roleID,
		"equip_type": equipType,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("卸下装备失败: %v", resp["msg"])
	}

	return nil
}

// DBItemGetEquipped 获取已穿戴装备
func DBItemGetEquipped(roleID uint64) ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/item/get_equipped", map[string]uint64{"role_id": roleID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取装备失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var equips []map[string]interface{}
	json.Unmarshal(data, &equips)

	return equips, nil
}

// DBItemGetEmptyCount 获取背包空位数
func DBItemGetEmptyCount(roleID uint64) (int, error) {
	resp, err := DBPost("/api/item/get_empty_count", map[string]uint64{"role_id": roleID})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("获取空位数失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// ========== 扩展角色相关 API ==========

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	AccountID  uint64 `json:"account_id"`
	Name       string `json:"name"`
	Gender     uint8  `json:"gender"`
	Appearance uint32 `json:"appearance"`
}

// DBRoleCreate 创建角色
func DBRoleCreate(req RoleCreateRequest) (uint64, error) {
	resp, err := DBPost("/api/role/create", map[string]interface{}{
		"account_id": req.AccountID,
		"name":       req.Name,
		"gender":     req.Gender,
		"appearance": req.Appearance,
	})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("创建角色失败: %v", resp["msg"])
	}

	return uint64(resp["data"].(float64)), nil
}

// DBRoleList 获取账号下所有角色
func DBRoleList(accountID uint64) ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/role/list", map[string]uint64{"account_id": accountID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取角色列表失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var roles []map[string]interface{}
	json.Unmarshal(data, &roles)

	return roles, nil
}

// RoleAttributeRequest 属性修改请求
type RoleAttributeRequest struct {
	Hp       *int   `json:"hp,omitempty"`
	MaxHp    *int   `json:"max_hp,omitempty"`
	Mp       *int   `json:"mp,omitempty"`
	MaxMp    *int   `json:"max_mp,omitempty"`
	Attack   *int   `json:"attack,omitempty"`
	Defense  *int   `json:"defense,omitempty"`
	Speed    *int   `json:"speed,omitempty"`
	Hit      *int   `json:"hit,omitempty"`
	Dodge    *int   `json:"dodge,omitempty"`
	Crit     *int   `json:"crit,omitempty"`
	Gold     *int64 `json:"gold,omitempty"`
	BindGold *int64 `json:"bind_gold,omitempty"`
	Yuanbao  *int64 `json:"yuanbao,omitempty"`
}

// DBRoleUpdateAttributes 批量更新角色属性
func DBRoleUpdateAttributes(roleID uint64, req RoleAttributeRequest) error {
	data := map[string]interface{}{
		"id": roleID,
	}
	if req.Hp != nil {
		data["hp"] = *req.Hp
	}
	if req.MaxHp != nil {
		data["max_hp"] = *req.MaxHp
	}
	if req.Mp != nil {
		data["mp"] = *req.Mp
	}
	if req.MaxMp != nil {
		data["max_mp"] = *req.MaxMp
	}
	if req.Attack != nil {
		data["attack"] = *req.Attack
	}
	if req.Defense != nil {
		data["defense"] = *req.Defense
	}
	if req.Speed != nil {
		data["speed"] = *req.Speed
	}
	if req.Hit != nil {
		data["hit"] = *req.Hit
	}
	if req.Dodge != nil {
		data["dodge"] = *req.Dodge
	}
	if req.Crit != nil {
		data["crit"] = *req.Crit
	}
	if req.Gold != nil {
		data["gold"] = *req.Gold
	}
	if req.BindGold != nil {
		data["bind_gold"] = *req.BindGold
	}
	if req.Yuanbao != nil {
		data["yuanbao"] = *req.Yuanbao
	}

	resp, err := DBPost("/api/role/update_attributes", data)
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("更新属性失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleDelete 删除角色
func DBRoleDelete(roleID uint64, accountID uint64) error {
	resp, err := DBPost("/api/role/delete", map[string]interface{}{
		"id":         roleID,
		"account_id": accountID,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("删除角色失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleChangeStamina 改变体力值
func DBRoleChangeStamina(roleID uint64, change int) (int, error) {
	resp, err := DBPost("/api/role/change_stamina", map[string]interface{}{
		"id":     roleID,
		"change": change,
	})
	if err != nil {
		return 0, err
	}

	if resp["code"].(float64) != 0 {
		return 0, fmt.Errorf("改变体力失败: %v", resp["msg"])
	}

	return int(resp["data"].(float64)), nil
}

// DBRoleSetPKMode 设置PK模式
func DBRoleSetPKMode(roleID uint64, mode uint8) error {
	resp, err := DBPost("/api/role/set_pk_mode", map[string]interface{}{
		"id":   roleID,
		"mode": mode,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("设置PK模式失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleUpdatePKValue 更新善恶值
func DBRoleUpdatePKValue(roleID uint64, change int) error {
	resp, err := DBPost("/api/role/update_pk_value", map[string]interface{}{
		"id":     roleID,
		"change": change,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("更新善恶值失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleChangeMap 切换地图
func DBRoleChangeMap(roleID uint64, mapID int, x, y int) error {
	resp, err := DBPost("/api/role/change_map", map[string]interface{}{
		"id":     roleID,
		"map_id": mapID,
		"x":      x,
		"y":      y,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("切换地图失败: %v", resp["msg"])
	}

	return nil
}

// DBRoleChangePosition 更新玩家位置（不切换地图）
func DBRoleChangePosition(roleID uint64, mapID uint32, x, y int) error {
	resp, err := DBPost("/api/role/update_position", map[string]interface{}{
		"id":     roleID,
		"map_id": mapID,
		"map_x":  x,
		"map_y":  y,
	})
	if err != nil {
		log.Printf("DBRoleChangePosition: 角色 %d 位置更新失败 - err=%v", roleID, err)
		return err
	}

	if resp["code"].(float64) != 0 {
		log.Printf("DBRoleChangePosition: 角色 %d 位置更新失败 - code=%v, msg=%v", roleID, resp["code"], resp["msg"])
		return fmt.Errorf("更新位置失败: %v", resp["msg"])
	}

	log.Printf("DBRoleChangePosition: 角色 %d 位置更新成功 - mapID=%d, x=%d, y=%d", roleID, mapID, x, y)
	return nil
}

// RolePosition 角色位置信息
type RolePosition struct {
	RoleID uint64
	MapID  uint32
	X      int
	Y      int
}

// DBGetRolePosition 获取角色当前位置信息
func DBGetRolePosition(roleID uint64) (*RolePosition, error) {
	resp, err := DBPost("/api/role/get", map[string]interface{}{
		"id": roleID,
	})
	if err != nil {
		log.Printf("DBGetRolePosition: 角色 %d 获取位置失败 - err=%v", roleID, err)
		return nil, err
	}

	// DBService 返回的字段是 data，而不是 role
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		log.Printf("DBGetRolePosition: 角色 %d data字段解析失败 - resp=%v", roleID, resp)
		return nil, fmt.Errorf("角色不存在")
	}

	mapX, _ := data["map_x"].(float64)
	mapY, _ := data["map_y"].(float64)
	mapID, _ := data["map_id"].(float64)

	position := &RolePosition{
		RoleID: roleID,
		MapID:  uint32(mapID),
		X:      int(mapX),
		Y:      int(mapY),
	}

	log.Printf("DBGetRolePosition: 角色 %d 位置 - mapID=%d, x=%d, y=%d", roleID, position.MapID, position.X, position.Y)
	return position, nil
}

// GetGatewayServiceURL 获取Gateway服务地址
func GetGatewayServiceURL() string {
	return AppConfig.Services.GatewayService
}

// DBSkillGetAllBase 获取所有武学基础信息
func DBSkillGetAllBase() ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/skill/get_all_base", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取武学列表失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var skills []map[string]interface{}
	json.Unmarshal(data, &skills)

	return skills, nil
}

// DBSkillGetEquipped 获取角色已装备的武学
func DBSkillGetEquipped(roleID uint64) ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/skill/get_equipped", map[string]uint64{"role_id": roleID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取已装备武学失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var skills []map[string]interface{}
	json.Unmarshal(data, &skills)

	return skills, nil
}

// DBItemGetAllBase 获取所有道具基础信息
func DBItemGetAllBase() ([]map[string]interface{}, error) {
	resp, err := DBPost("/api/item/get_all_base", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取道具列表失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var items []map[string]interface{}
	json.Unmarshal(data, &items)

	return items, nil
}

// DBItemGetBase 获取道具基础信息
func DBItemGetBase(itemID uint32) (map[string]interface{}, error) {
	resp, err := DBPost("/api/item/get_base", map[string]uint32{"item_id": itemID})
	if err != nil {
		return nil, err
	}

	if resp["code"].(float64) != 0 {
		return nil, fmt.Errorf("获取道具信息失败: %v", resp["msg"])
	}

	data, _ := json.Marshal(resp["data"])
	var item map[string]interface{}
	json.Unmarshal(data, &item)

	return item, nil
}

// DBItemSplit 拆分物品
func DBItemSplit(roleID uint64, gridIndex int, count uint32) error {
	resp, err := DBPost("/api/item/split", map[string]interface{}{
		"role_id":    roleID,
		"grid_index": gridIndex,
		"count":      count,
	})
	if err != nil {
		return err
	}

	if resp["code"].(float64) != 0 {
		return fmt.Errorf("拆分物品失败: %v", resp["msg"])
	}

	return nil
}
