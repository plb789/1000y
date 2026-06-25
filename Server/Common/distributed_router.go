package common

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// GameServiceConfig GameService实例配置
type GameServiceConfig struct {
	ID                  uint32   `yaml:"id" json:"id"`
	Name                string   `yaml:"name" json:"name"`
	URL                 string   `yaml:"url" json:"url"`
	HandledMaps         []uint32 `yaml:"handled_maps" json:"handled_maps"`
	MaxConnections      int      `yaml:"max_connections" json:"max_connections"`
	HealthCheckInterval int      `yaml:"health_check_interval" json:"health_check_interval"`
	Weight              int      `yaml:"weight" json:"weight"`
}

// RoutingConfig 路由策略配置
type RoutingConfig struct {
	Strategy        string `yaml:"strategy" json:"strategy"`
	FallbackEnabled bool   `yaml:"fallback_enabled" json:"fallback_enabled"`
	CacheTTL        int    `yaml:"cache_ttl" json:"cache_ttl"`
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	Algorithm       string  `yaml:"algorithm" json:"algorithm"`
	HealthThreshold float64 `yaml:"health_threshold" json:"health_threshold"`
}

// DiscoveryConfig 服务发现配置
type DiscoveryConfig struct {
	Type            string `yaml:"type" json:"type"`
	RegistryURL     string `yaml:"registry_url" json:"registry_url"`
	RefreshInterval int    `yaml:"refresh_interval" json:"refresh_interval"`
}

// FailoverConfig 故障转移配置
type FailoverConfig struct {
	Enabled        bool                 `yaml:"enabled" json:"enabled"`
	MaxRetries     int                  `yaml:"max_retries" json:"max_retries"`
	RetryDelayMs   int                  `yaml:"retry_delay_ms" json:"retry_delay_ms"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker" json:"circuit_breaker"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled        bool `yaml:"enabled" json:"enabled"`
	Threshold      int  `yaml:"threshold" json:"threshold"`
	TimeoutSeconds int  `yaml:"timeout_seconds" json:"timeout_seconds"`
}

// InstancesConfig 完整的实例配置
type InstancesConfig struct {
	Instances    []GameServiceConfig `yaml:"instances" json:"instances"`
	Routing      RoutingConfig       `yaml:"routing" json:"routing"`
	LoadBalancer LoadBalancerConfig  `yaml:"load_balancer" json:"load_balancer"`
	Discovery    DiscoveryConfig     `yaml:"discovery" json:"discovery"`
	Failover     FailoverConfig      `yaml:"failover" json:"failover"`
}

// DistributedRouter 分布式路由器
type DistributedRouter struct {
	config          *InstancesConfig
	instances       map[uint32]*GameServiceInstance // instanceID -> instance
	mapToInstances  map[uint32][]uint32             // mapID -> []instanceID (支持多副本)
	roundRobinIndex uint32
	mutex           sync.RWMutex
	httpClient      *http.Client
	circuitBreakers map[uint32]*CircuitBreaker
	healthChecker   *HealthChecker
	metrics         *RouterMetrics
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	state           string // "closed", "open", "half-open"
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	threshold       int
	timeoutSeconds  int
	mutex           sync.Mutex
}

// HealthChecker 健康检查器
type HealthChecker struct {
	router   *DistributedRouter
	stopChan chan struct{}
	interval time.Duration
}

// RouterMetrics 路由器指标
type RouterMetrics struct {
	TotalRequests       int64
	SuccessRequests     int64
	FailedRequests      int64
	FallbackRequests    int64
	CircuitBreakerTrips int64
	AvgResponseTime     time.Duration
	mutex               sync.RWMutex
}

var globalRouter *DistributedRouter

// InitDistributedRouter 初始化分布式路由器
func InitDistributedRouter(configPath string) error {
	config, err := LoadInstancesConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载实例配置失败: %v", err)
	}

	router := &DistributedRouter{
		config:          config,
		instances:       make(map[uint32]*GameServiceInstance),
		mapToInstances:  make(map[uint32][]uint32),
		roundRobinIndex: 0,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		circuitBreakers: make(map[uint32]*CircuitBreaker),
		metrics:         &RouterMetrics{},
	}

	// 初始化所有实例
	for _, instConfig := range config.Instances {
		instance := &GameServiceInstance{
			InstanceID:  instConfig.ID,
			URL:         instConfig.URL,
			HandledMaps: instConfig.HandledMaps,
			StartTime:   time.Now().Unix(),
			Heartbeat:   time.Now().Unix(),
		}

		router.instances[instConfig.ID] = instance

		// 建立mapID到instanceID的映射
		for _, mapID := range instConfig.HandledMaps {
			router.mapToInstances[mapID] = append(router.mapToInstances[mapID], instConfig.ID)
		}

		// 初始化熔断器
		if config.Failover.CircuitBreaker.Enabled {
			router.circuitBreakers[instConfig.ID] = &CircuitBreaker{
				state:          "closed",
				threshold:      config.Failover.CircuitBreaker.Threshold,
				timeoutSeconds: config.Failover.CircuitBreaker.TimeoutSeconds,
			}
		}
	}

	globalRouter = router

	// 启动健康检查
	if config.Discovery.Type == "static" {
		router.startHealthCheck()
	}

	log.Printf("✅ 分布式路由器初始化完成: %d个GameService实例", len(config.Instances))
	for _, inst := range config.Instances {
		log.Printf("   - 实例%d: %s (地图: %v)", inst.ID, inst.URL, inst.HandledMaps)
	}

	return nil
}

// LoadInstancesConfig 加载实例配置
func LoadInstancesConfig(configPath string) (*InstancesConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config InstancesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析YAML配置失败: %v", err)
	}

	return &config, nil
}

// RouteByMapID 根据地图ID路由到对应的GameService实例
func RouteByMapID(mapID uint32) (*GameServiceInstance, error) {
	if globalRouter == nil {
		return nil, fmt.Errorf("路由器未初始化")
	}

	globalRouter.mutex.RLock()
	defer globalRouter.mutex.RUnlock()

	// 根据路由策略选择实例
	switch globalRouter.config.Routing.Strategy {
	case "map_id":
		return globalRouter.routeByMapID(mapID)
	case "round_robin":
		return globalRouter.routeRoundRobin()
	case "random":
		return globalRouter.routeRandom()
	case "hash":
		return globalRouter.routeHash(mapID)
	default:
		return globalRouter.routeByMapID(mapID)
	}
}

// routeByMapID 按地图ID路由（默认策略）
func (r *DistributedRouter) routeByMapID(mapID uint32) (*GameServiceInstance, error) {
	instanceIDs, ok := r.mapToInstances[mapID]
	if !ok || len(instanceIDs) == 0 {
		// 如果没有找到，尝试降级到其他实例
		if r.config.Routing.FallbackEnabled {
			return r.fallbackRoute()
		}
		return nil, fmt.Errorf("没有找到处理地图%d的实例", mapID)
	}

	// 选择第一个可用实例（支持多副本时可以随机选择）
	for _, instanceID := range instanceIDs {
		inst := r.instances[instanceID]
		if inst != nil && r.isInstanceHealthy(instanceID) {
			return inst, nil
		}
	}

	// 所有主实例都不可用，尝试降级
	if r.config.Routing.FallbackEnabled {
		return r.fallbackRoute()
	}

	return nil, fmt.Errorf("处理地图%d的所有实例都不可用", mapID)
}

// routeRoundRobin 轮询路由
func (r *DistributedRouter) routeRoundRobin() (*GameServiceInstance, error) {
	instances := r.getHealthyInstances()
	if len(instances) == 0 {
		if r.config.Routing.FallbackEnabled {
			return r.fallbackRoute()
		}
		return nil, fmt.Errorf("没有可用的健康实例")
	}

	r.roundRobinIndex++
	index := r.roundRobinIndex % uint32(len(instances))
	return instances[index], nil
}

// routeRandom 随机路由
func (r *DistributedRouter) routeRandom() (*GameServiceInstance, error) {
	instances := r.getHealthyInstances()
	if len(instances) == 0 {
		if r.config.Routing.FallbackEnabled {
			return r.fallbackRoute()
		}
		return nil, fmt.Errorf("没有可用的健康实例")
	}

	index := rand.Intn(len(instances))
	return instances[index], nil
}

// routeHash 哈希路由
func (r *DistributedRouter) routeHash(key uint32) (*GameServiceInstance, error) {
	instances := r.getHealthyInstances()
	if len(instances) == 0 {
		if r.config.Routing.FallbackEnabled {
			return r.fallbackRoute()
		}
		return nil, fmt.Errorf("没有可用的健康实例")
	}

	index := key % uint32(len(instances))
	return instances[index], nil
}

// fallbackRoute 降级路由
func (r *DistributedRouter) fallbackRoute() (*GameServiceInstance, error) {
	instances := r.getHealthyInstances()
	if len(instances) > 0 {
		r.metrics.mutex.Lock()
		r.metrics.FallbackRequests++
		r.metrics.mutex.Unlock()

		index := rand.Intn(len(instances))
		return instances[index], nil
	}
	return nil, fmt.Errorf("所有实例都不可用，降级失败")
}

// getHealthyInstances 获取所有健康的实例
func (r *DistributedRouter) getHealthyInstances() []*GameServiceInstance {
	var healthy []*GameServiceInstance
	for _, inst := range r.instances {
		if r.isInstanceHealthy(inst.InstanceID) {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// isInstanceHealthy 检查实例是否健康
func (r *DistributedRouter) isInstanceHealthy(instanceID uint32) bool {
	cb, ok := r.circuitBreakers[instanceID]
	if !ok {
		return true // 没有熔断器则认为健康
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case "closed":
		return true
	case "open":
		// 检查是否超时，可以进入半开状态
		if time.Since(cb.lastFailureTime) > time.Duration(cb.timeoutSeconds)*time.Second {
			cb.state = "half-open"
			cb.successCount = 0
			return true
		}
		return false
	case "half-open":
		return true
	default:
		return true
	}
}

// RecordSuccess 记录成功请求
func (r *DistributedRouter) RecordSuccess(instanceID uint32) {
	r.metrics.mutex.Lock()
	r.metrics.SuccessRequests++
	r.metrics.mutex.Unlock()

	cb, ok := r.circuitBreakers[instanceID]
	if !ok {
		return
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	if cb.state == "half-open" {
		cb.successCount++
		if cb.successCount >= 3 { // 半开状态下连续3次成功则关闭熔断器
			cb.state = "closed"
			log.Printf("🔄 熔断器关闭: instanceID=%d", instanceID)
		}
	}
}

// RecordFailure 记录失败请求
func (r *DistributedRouter) RecordFailure(instanceID uint32) {
	r.metrics.mutex.Lock()
	r.metrics.FailedRequests++
	r.metrics.TotalRequests++
	r.metrics.mutex.Unlock()

	cb, ok := r.circuitBreakers[instanceID]
	if !ok {
		return
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == "half-open" {
		// 半开状态下失败立即打开熔断器
		cb.state = "open"
		r.metrics.mutex.Lock()
		r.metrics.CircuitBreakerTrips++
		r.metrics.mutex.Unlock()
		log.Printf(`⚠️ 熔断器打开(半开状态失败): instanceID=%d`, instanceID)
	} else if cb.state == "closed" && cb.failureCount >= cb.threshold {
		cb.state = "open"
		r.metrics.mutex.Lock()
		r.metrics.CircuitBreakerTrips++
		r.metrics.mutex.Unlock()
		log.Printf(`⚠️ 熔断器打开: instanceID=%d, 连续失败%d次`, instanceID, cb.failureCount)
	}
}

// startHealthCheck 启动健康检查
func (r *DistributedRouter) startHealthCheck() {
	interval := 30 * time.Second
	for _, cfg := range r.config.Instances {
		if cfg.HealthCheckInterval > 0 {
			interval = time.Duration(cfg.HealthCheckInterval) * time.Second
		}
	}

	r.healthChecker = &HealthChecker{
		router:   r,
		stopChan: make(chan struct{}),
		interval: interval,
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				r.performHealthCheck()
			case <-r.healthChecker.stopChan:
				return
			}
		}
	}()

	log.Printf("🏥 健康检查启动: 间隔=%v", interval)
}

// performHealthCheck 执行健康检查
func (r *DistributedRouter) performHealthCheck() {
	for id, inst := range r.instances {
		go func(instanceID uint32, url string) {
			startTime := time.Now()
			resp, err := r.httpClient.Get(url + "/health") // ★ 修复：使用 /health 而非 /api/health
			duration := time.Since(startTime)

			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf(`❌ 健康检查失败: instanceID=%d, url=%s, error=%v, duration=%v`,
					instanceID, url, err, duration)
				r.RecordFailure(instanceID)
			} else {
				resp.Body.Close()
				r.RecordSuccess(instanceID)
			}
		}(id, inst.URL)
	}
}

// GetRouterMetrics 获取路由器指标
func GetRouterMetrics() *RouterMetrics {
	if globalRouter == nil {
		return nil
	}
	return globalRouter.metrics
}

// GetAllInstanceIDs 获取所有实例ID
func GetAllInstanceIDs() []uint32 {
	if globalRouter == nil {
		return nil
	}

	globalRouter.mutex.RLock()
	defer globalRouter.mutex.RUnlock()

	ids := make([]uint32, 0, len(globalRouter.instances))
	for id := range globalRouter.instances {
		ids = append(ids, id)
	}
	return ids
}

// GetInstanceURLByID 根据实例ID获取URL
func GetInstanceURLByID(instanceID uint32) string {
	if globalRouter == nil {
		return ""
	}

	globalRouter.mutex.RLock()
	defer globalRouter.mutex.RUnlock()

	if inst, ok := globalRouter.instances[instanceID]; ok {
		return inst.URL
	}
	return ""
}

// RecordSuccess 记录成功请求（包级函数）
func RecordSuccess(instanceID uint32) {
	if globalRouter != nil {
		globalRouter.RecordSuccess(instanceID)
	}
}

// RecordFailure 记录失败请求（包级函数）
func RecordFailure(instanceID uint32) {
	if globalRouter != nil {
		globalRouter.RecordFailure(instanceID)
	}
}

// GetAnyHealthyInstance 获取任意健康的GameService实例（用于代理请求）
func GetAnyHealthyInstance() (*GameServiceInstance, error) {
	if globalRouter == nil {
		return nil, fmt.Errorf("路由器未初始化")
	}

	globalRouter.mutex.RLock()
	defer globalRouter.mutex.RUnlock()

	// 优先返回健康实例
	for _, inst := range globalRouter.instances {
		cb, ok := globalRouter.circuitBreakers[inst.InstanceID]
		if !ok || cb.state != "open" { // 熔断器未开启
			return inst, nil
		}
	}

	// 如果所有实例都不健康，返回第一个（降级）
	if len(globalRouter.instances) > 0 {
		for _, inst := range globalRouter.instances {
			return inst, nil // 返回第一个实例作为降级方案
		}
	}

	return nil, fmt.Errorf("没有可用的GameService实例")
}
