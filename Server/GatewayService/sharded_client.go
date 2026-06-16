package main

import (
	"encoding/json"
	"fmt"
	common "game-server/Common"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ShardedGameClient 分片游戏服务客户端
type ShardedGameClient struct {
	defaultURL string
	httpClient *http.Client
}

var shardedGameClient *ShardedGameClient

// NewShardedGameClient 创建分片游戏客户端
func NewShardedGameClient(defaultURL string) *ShardedGameClient {
	return &ShardedGameClient{
		defaultURL: defaultURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetGameServiceURL 根据地图ID获取游戏服务URL
func (c *ShardedGameClient) GetGameServiceURL(mapID uint32) string {
	// 查询注册中心获取对应实例
	inst := common.GetInstanceByMapID(mapID)
	if inst != nil {
		return inst.URL
	}
	
	// 如果注册中心没有，返回默认URL
	return c.defaultURL
}

// GetPlayerFighter 获取玩家战斗信息(路由到正确的实例)
func (c *ShardedGameClient) GetPlayerFighter(roleID uint64, mapID uint32) (*PlayerFighter, error) {
	url := c.GetGameServiceURL(mapID)
	if url == "" {
		return nil, fmt.Errorf("找不到处理地图%d的游戏服务实例", mapID)
	}

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/role/fighter/%d", url, roleID))
	if err != nil {
		// 降级到默认服务
		resp, err = c.httpClient.Get(fmt.Sprintf("%s/api/role/fighter/%d", c.defaultURL, roleID))
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data *PlayerFighter `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf(result.Msg)
	}
	return result.Data, nil
}

// UpdateRolePosition 更新角色位置(路由到正确的实例)
func (c *ShardedGameClient) UpdateRolePosition(roleID uint64, mapID uint32, x, y int) error {
	url := c.GetGameServiceURL(mapID)
	if url == "" {
		return fmt.Errorf("找不到处理地图%d的游戏服务实例", mapID)
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"role_id": roleID,
		"map_id":  mapID,
		"x":       x,
		"y":       y,
	})

	resp, err := c.httpClient.Post(url+"/api/role/position", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		// 降级到默认服务
		resp, err = c.httpClient.Post(c.defaultURL+"/api/role/position", "application/json", strings.NewReader(string(reqBody)))
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	return nil
}

// PlayerFighter 玩家战斗信息
type PlayerFighter struct {
	RoleID   uint64 `json:"role_id"`
	MapID    uint32 `json:"map_id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Name     string `json:"name"`
	Level    uint32 `json:"level"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	MP       int    `json:"mp"`
	MaxMP    int    `json:"max_mp"`
	Attack   int    `json:"attack"`
	Defense  int    `json:"defense"`
	Speed    int    `json:"speed"`
	Hit      int    `json:"hit"`
	Dodge    int    `json:"dodge"`
	Crit     int    `json:"crit"`
	Faction  string `json:"faction"`
}

// InitShardedGameClient 初始化分片游戏客户端
func InitShardedGameClient() {
	registryURL := common.AppConfig.Services.RegistryService
	if registryURL == "" {
		registryURL = common.AppConfig.Services.DBService
	}
	common.InitRegistry(registryURL)

	defaultURL := common.AppConfig.Services.GameService
	if defaultURL == "" {
		defaultURL = "http://localhost:8082"
	}

	shardedGameClient = NewShardedGameClient(defaultURL)
	log.Printf("分片游戏客户端初始化完成, 默认服务: %s", defaultURL)
}

// GetShardedGameClient 获取分片游戏客户端
func GetShardedGameClient() *ShardedGameClient {
	return shardedGameClient
}
