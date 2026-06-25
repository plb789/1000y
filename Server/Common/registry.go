package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GameServiceInstance 游戏服务实例信息
type GameServiceInstance struct {
	InstanceID  uint32   `json:"instance_id"`
	URL         string   `json:"url"`
	HandledMaps []uint32 `json:"handled_maps"`
	StartTime   int64    `json:"start_time"`
	Heartbeat   int64    `json:"heartbeat"`
}

// Registry 服务注册中心客户端
type Registry struct {
	instances   map[uint32]*GameServiceInstance // 本地缓存
	mapToInst   map[uint32]uint32               // mapID -> instanceID
	registryURL string                          // 注册中心地址
	httpClient  *http.Client                    // HTTP客户端
}

var reg *Registry

// InitRegistry 初始化注册中心
func InitRegistry(registryURL string) {
	if reg == nil {
		reg = &Registry{
			instances:   make(map[uint32]*GameServiceInstance),
			mapToInst:   make(map[uint32]uint32),
			registryURL: registryURL,
			httpClient: &http.Client{
				Timeout: 5 * time.Second,
			},
		}
	}
}

// RegisterGameService 注册游戏服务实例
func RegisterGameService(instanceID uint32, url string, handledMaps []uint32) error {
	if reg == nil {
		return fmt.Errorf("注册中心未初始化")
	}

	inst := &GameServiceInstance{
		InstanceID:  instanceID,
		URL:         url,
		HandledMaps: handledMaps,
		StartTime:   time.Now().Unix(),
		Heartbeat:   time.Now().Unix(),
	}

	// 调用RegistryService API注册
	reqBody, _ := json.Marshal(map[string]interface{}{
		"instance_id":  instanceID,
		"service_type": "game",
		"url":          url,
		"handled_maps": handledMaps,
	})

	resp, err := reg.httpClient.Post(reg.registryURL+"/api/registry/register", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		log.Printf("注册到服务中心失败: %v, 使用本地缓存", err)
		// 降级：使用本地缓存
		reg.instances[instanceID] = inst
		for _, mapID := range handledMaps {
			reg.mapToInst[mapID] = instanceID
		}
		return nil
	}
	defer resp.Body.Close()

	// 更新本地缓存
	reg.instances[instanceID] = inst
	for _, mapID := range handledMaps {
		reg.mapToInst[mapID] = instanceID
	}

	log.Printf("游戏服务实例注册: instanceID=%d, url=%s, maps=%v", instanceID, url, handledMaps)
	return nil
}

// UnregisterGameService 注销游戏服务实例
func UnregisterGameService(instanceID uint32) error {
	if reg == nil {
		return fmt.Errorf("注册中心未初始化")
	}

	// 调用RegistryService API注销
	reqBody, _ := json.Marshal(map[string]interface{}{
		"instance_id": instanceID,
	})

	resp, err := reg.httpClient.Post(reg.registryURL+"/api/registry/unregister", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		log.Printf("从服务中心注销失败: %v", err)
	}

	if resp != nil {
		resp.Body.Close()
	}

	// 从本地缓存移除
	if inst, ok := reg.instances[instanceID]; ok {
		for _, mapID := range inst.HandledMaps {
			delete(reg.mapToInst, mapID)
		}
	}
	delete(reg.instances, instanceID)

	return nil
}

// GetInstanceByMapID 根据地图ID获取实例
func GetInstanceByMapID(mapID uint32) *GameServiceInstance {
	if reg == nil {
		return nil
	}

	instanceID, ok := reg.mapToInst[mapID]
	if !ok {
		// 本地缓存没有，尝试从服务中心刷新
		reg.RefreshFromRegistry()
		instanceID, ok = reg.mapToInst[mapID]
		if !ok {
			return nil
		}
	}

	return reg.instances[instanceID]
}

// GetInstanceByRoleID 根据角色ID获取实例（通过轮询或随机选择可用实例）
func GetInstanceByRoleID(roleID uint64) *GameServiceInstance {
	if reg == nil {
		return nil
	}

	// 优先尝试从所有实例中找一个可用的
	instances := GetAllInstances()
	if len(instances) == 0 {
		return nil
	}

	// 简单策略：根据roleID哈希到固定实例（保证同一角色的请求路由到同一实例）
	index := roleID % uint64(len(instances))
	return instances[index]
}

// GetInstanceByID 根据实例ID获取实例
func GetInstanceByID(instanceID uint32) *GameServiceInstance {
	if reg == nil {
		return nil
	}
	return reg.instances[instanceID]
}

// GetAllInstances 获取所有实例
func GetAllInstances() []*GameServiceInstance {
	if reg == nil {
		return nil
	}

	instances := make([]*GameServiceInstance, 0, len(reg.instances))
	for _, inst := range reg.instances {
		instances = append(instances, inst)
	}
	return instances
}

// RefreshFromRegistry 从注册中心刷新本地缓存
func (r *Registry) RefreshFromRegistry() {
	// 调用RegistryService API获取所有实例
	resp, err := r.httpClient.Get(r.registryURL + "/api/registry/list")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Code int                    `json:"code"`
		Data []*GameServiceInstance `json:"data"`
	}

	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil || result.Code != 0 {
		return
	}

	// 更新本地缓存
	r.instances = make(map[uint32]*GameServiceInstance)
	r.mapToInst = make(map[uint32]uint32)

	for _, inst := range result.Data {
		r.instances[inst.InstanceID] = inst
		for _, mapID := range inst.HandledMaps {
			r.mapToInst[mapID] = inst.InstanceID
		}
	}
}

// UpdateHeartbeat 更新心跳
func UpdateHeartbeat(instanceID uint32) error {
	if reg == nil {
		return fmt.Errorf("注册中心未初始化")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"instance_id": instanceID,
	})

	resp, err := reg.httpClient.Post(reg.registryURL+"/api/registry/heartbeat", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// IsMapHandled 检查地图是否已被处理
func IsMapHandled(mapID uint32) bool {
	if reg == nil {
		return false
	}
	_, ok := reg.mapToInst[mapID]
	return ok
}

// HandleMapInRegistry 检查实例是否处理指定地图
func HandleMapInRegistry(instanceID uint32, mapID uint32) bool {
	if reg == nil {
		return false
	}
	inst, ok := reg.instances[instanceID]
	if !ok {
		return false
	}
	for _, m := range inst.HandledMaps {
		if m == mapID {
			return true
		}
	}
	return false
}
