package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	HTTPPort  int `yaml:"http_port"`
	Heartbeat struct {
		Timeout         int64 `yaml:"timeout"`          // 心跳超时时间(秒)
		CleanupInterval int64 `yaml:"cleanup_interval"` // 清理离线实例间隔(秒)
	} `yaml:"heartbeat"`
	Services struct {
		RegistryService string `yaml:"registry_service"`
	} `yaml:"services"`
}

var registryConfig RegistryConfig

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	InstanceID  uint32   `json:"instance_id"`
	ServiceType string   `json:"service_type"` // game, login, gateway, etc.
	URL         string   `json:"url"`
	HandledMaps []uint32 `json:"handled_maps"`
	StartTime   int64    `json:"start_time"`
	Heartbeat   int64    `json:"heartbeat"`
	Status      string   `json:"status"` // online, offline
}

// RegistryService 服务注册中心
type RegistryService struct {
	instances     map[uint32]*ServiceInstance   // instance_id -> instance
	serviceMap    map[string][]*ServiceInstance // service_type -> instances
	mapToInstance map[uint32]uint32             // map_id -> instance_id
	mu            sync.RWMutex
}

var registry *RegistryService

func NewRegistryService() *RegistryService {
	return &RegistryService{
		instances:     make(map[uint32]*ServiceInstance),
		serviceMap:    make(map[string][]*ServiceInstance),
		mapToInstance: make(map[uint32]uint32),
	}
}

// Register 注册服务实例
func (r *RegistryService) Register(instanceID uint32, serviceType string, url string, handledMaps []uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inst := &ServiceInstance{
		InstanceID:  instanceID,
		ServiceType: serviceType,
		URL:         url,
		HandledMaps: handledMaps,
		StartTime:   time.Now().Unix(),
		Heartbeat:   time.Now().Unix(),
		Status:      "online",
	}

	// 如果已存在，先清理旧数据
	if oldInst, ok := r.instances[instanceID]; ok {
		for _, mapID := range oldInst.HandledMaps {
			delete(r.mapToInstance, mapID)
		}
	}

	r.instances[instanceID] = inst

	// 更新服务类型映射
	r.serviceMap[serviceType] = append(r.serviceMap[serviceType], inst)

	// 更新地图映射（仅游戏服务）
	if serviceType == "game" {
		for _, mapID := range handledMaps {
			r.mapToInstance[mapID] = instanceID
		}
	}

	log.Printf("服务注册: instanceID=%d, type=%s, url=%s, maps=%v", instanceID, serviceType, url, handledMaps)
	return nil
}

// Unregister 注销服务实例
func (r *RegistryService) Unregister(instanceID uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inst, ok := r.instances[instanceID]
	if !ok {
		return fmt.Errorf("实例不存在")
	}

	// 从服务类型映射中移除
	serviceType := inst.ServiceType
	if instances, ok := r.serviceMap[serviceType]; ok {
		for i, in := range instances {
			if in.InstanceID == instanceID {
				r.serviceMap[serviceType] = append(instances[:i], instances[i+1:]...)
				break
			}
		}
	}

	// 从地图映射中移除
	if serviceType == "game" {
		for _, mapID := range inst.HandledMaps {
			if r.mapToInstance[mapID] == instanceID {
				delete(r.mapToInstance, mapID)
			}
		}
	}

	delete(r.instances, instanceID)

	log.Printf("服务注销: instanceID=%d", instanceID)
	return nil
}

// UpdateHeartbeat 更新心跳
func (r *RegistryService) UpdateHeartbeat(instanceID uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inst, ok := r.instances[instanceID]
	if !ok {
		return fmt.Errorf("实例不存在")
	}

	inst.Heartbeat = time.Now().Unix()
	inst.Status = "online"

	return nil
}

// GetInstanceByMapID 根据地图ID获取游戏服务实例
func (r *RegistryService) GetInstanceByMapID(mapID uint32) *ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instanceID, ok := r.mapToInstance[mapID]
	if !ok {
		return nil
	}

	return r.instances[instanceID]
}

// GetInstanceByID 根据实例ID获取实例
func (r *RegistryService) GetInstanceByID(instanceID uint32) *ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.instances[instanceID]
}

// GetInstancesByType 根据服务类型获取所有实例
func (r *RegistryService) GetInstancesByType(serviceType string) []*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.serviceMap[serviceType]
}

// GetAllInstances 获取所有实例
func (r *RegistryService) GetAllInstances() []*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances := make([]*ServiceInstance, 0, len(r.instances))
	for _, inst := range r.instances {
		instances = append(instances, inst)
	}
	return instances
}

// CleanupOfflineInstances 清理离线实例（心跳超时）
func (r *RegistryService) CleanupOfflineInstances(timeoutSeconds int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	toRemove := make([]uint32, 0)

	for id, inst := range r.instances {
		if now-inst.Heartbeat > timeoutSeconds {
			inst.Status = "offline"
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		inst := r.instances[id]

		// 从服务类型映射中移除
		serviceType := inst.ServiceType
		if instances, ok := r.serviceMap[serviceType]; ok {
			for i, in := range instances {
				if in.InstanceID == id {
					r.serviceMap[serviceType] = append(instances[:i], instances[i+1:]...)
					break
				}
			}
		}

		// 从地图映射中移除
		if serviceType == "game" {
			for _, mapID := range inst.HandledMaps {
				if r.mapToInstance[mapID] == id {
					delete(r.mapToInstance, mapID)
				}
			}
		}

		delete(r.instances, id)
		log.Printf("清理离线实例: instanceID=%d", id)
	}
}

func main() {
	// 加载配置文件
	configPath := "Config/Registry.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("加载配置文件失败: %v, 使用默认配置", err)
		// 使用默认配置
		registryConfig.HTTPPort = 8081
		registryConfig.Heartbeat.Timeout = 60
		registryConfig.Heartbeat.CleanupInterval = 30
	} else {
		if err := yaml.Unmarshal(data, &registryConfig); err != nil {
			log.Printf("解析配置文件失败: %v, 使用默认配置", err)
			registryConfig.HTTPPort = 8081
			registryConfig.Heartbeat.Timeout = 60
			registryConfig.Heartbeat.CleanupInterval = 30
		}
	}

	// 设置默认值
	if registryConfig.HTTPPort <= 0 {
		registryConfig.HTTPPort = 8081
	}
	if registryConfig.Heartbeat.Timeout <= 0 {
		registryConfig.Heartbeat.Timeout = 60
	}
	if registryConfig.Heartbeat.CleanupInterval <= 0 {
		registryConfig.Heartbeat.CleanupInterval = 30
	}

	registry = NewRegistryService()

	// 启动定时清理离线实例
	go func() {
		ticker := time.NewTicker(time.Duration(registryConfig.Heartbeat.CleanupInterval) * time.Second)
		for range ticker.C {
			registry.CleanupOfflineInstances(registryConfig.Heartbeat.Timeout)
		}
	}()

	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "registry",
		})
	})

	// 注册服务
	r.POST("/api/registry/register", func(c *gin.Context) {
		var req struct {
			InstanceID  uint32   `json:"instance_id"`
			ServiceType string   `json:"service_type"`
			URL         string   `json:"url"`
			HandledMaps []uint32 `json:"handled_maps"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
			return
		}

		if err := registry.Register(req.InstanceID, req.ServiceType, req.URL, req.HandledMaps); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success"})
	})

	// 注销服务
	r.POST("/api/registry/unregister", func(c *gin.Context) {
		var req struct {
			InstanceID uint32 `json:"instance_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
			return
		}

		if err := registry.Unregister(req.InstanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success"})
	})

	// 更新心跳
	r.POST("/api/registry/heartbeat", func(c *gin.Context) {
		var req struct {
			InstanceID uint32 `json:"instance_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
			return
		}

		if err := registry.UpdateHeartbeat(req.InstanceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success"})
	})

	// 获取所有实例
	r.GET("/api/registry/list", func(c *gin.Context) {
		instances := registry.GetAllInstances()
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": instances})
	})

	// 根据服务类型获取实例
	r.GET("/api/registry/type/:service_type", func(c *gin.Context) {
		serviceType := c.Param("service_type")
		instances := registry.GetInstancesByType(serviceType)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": instances})
	})

	// 根据地图ID获取游戏服务实例
	r.GET("/api/registry/map/:map_id", func(c *gin.Context) {
		var mapID uint32
		fmt.Sscanf(c.Param("map_id"), "%d", &mapID)

		inst := registry.GetInstanceByMapID(mapID)
		if inst == nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "地图未分配"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "data": inst})
	})

	// 根据实例ID获取实例
	r.GET("/api/registry/instance/:instance_id", func(c *gin.Context) {
		var instanceID uint32
		fmt.Sscanf(c.Param("instance_id"), "%d", &instanceID)

		inst := registry.GetInstanceByID(instanceID)
		if inst == nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "实例不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "data": inst})
	})

	// 获取服务统计
	r.GET("/api/registry/stats", func(c *gin.Context) {
		instances := registry.GetAllInstances()

		typeStats := make(map[string]int)
		for _, inst := range instances {
			typeStats[inst.ServiceType]++
		}

		c.JSON(http.StatusOK, gin.H{
			"code":      0,
			"total":     len(instances),
			"by_type":   typeStats,
			"timestamp": time.Now().Unix(),
		})
	})

	port := registryConfig.HTTPPort
	log.Printf("RegistryService 启动，监听端口: %d (配置文件: %s)", port, configPath)
	log.Fatal(r.Run(fmt.Sprintf(":%d", port)))
}

// 辅助函数：JSON序列化
func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
