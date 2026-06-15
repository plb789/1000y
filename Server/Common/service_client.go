package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GameServiceClient 游戏服务HTTP客户端
type GameServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewGameServiceClient 创建游戏服务客户端
func NewGameServiceClient(baseURL string) *GameServiceClient {
	return &GameServiceClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
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
	Appearance string `json:"appearance"`
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

// 服务地址配置
var (
	LoginServiceURL  = "http://localhost:8081"
	GameServiceURL   = "http://localhost:8082"
)

// GameClient 游戏服务客户端(全局单例)
var GameClient = NewGameServiceClient(GameServiceURL)

// LoginClient 登录服务客户端
var LoginClient = NewGameServiceClient(LoginServiceURL)
