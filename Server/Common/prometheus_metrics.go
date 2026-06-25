package common

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics Prometheus监控指标管理器
type PrometheusMetrics struct {
	registry *prometheus.Registry

	// 网关核心指标
	GatewayMoveRequestsTotal      *prometheus.CounterVec
	GatewayValidationDuration     *prometheus.HistogramVec
	GatewayBroadcastMessagesTotal *prometheus.CounterVec
	GatewayConnectionsCurrent     prometheus.Gauge
	GatewayConnectionsTotal       prometheus.Counter

	// 路由器指标
	RouterRequestsTotal         *prometheus.CounterVec
	RouterRequestDuration       *prometheus.HistogramVec
	RouterFallbackRequestsTotal prometheus.Counter
	RouterCircuitBreakerTrips   *prometheus.CounterVec
	RouterInstanceHealth        *prometheus.GaugeVec

	// 广播指标
	BroadcastTotal              *prometheus.CounterVec
	BroadcastLatency            *prometheus.HistogramVec
	BroadcastDuplicatedMessages prometheus.Counter
	BroadcastFailedMessages     prometheus.Counter

	// 怪物同步指标
	MonsterSyncUpdatesTotal *prometheus.CounterVec
	MonsterSyncBatchSize    *prometheus.HistogramVec
	MonsterSyncLatency      *prometheus.HistogramVec
	MonsterActiveCount      *prometheus.GaugeVec

	// 自定义业务指标
	PlayerOnlineCount         prometheus.Gauge
	MapPlayerCount            *prometheus.GaugeVec
	MessageProcessingDuration *prometheus.HistogramVec

	// 代理指标（新增：用于监控网关代理请求）
	ProxyRequestsTotal   *prometheus.CounterVec
	ProxyRequestDuration *prometheus.HistogramVec

	mutex sync.RWMutex
}

var promMetrics *PrometheusMetrics

// InitPrometheusMetrics 初始化Prometheus监控
func InitPrometheusMetrics(metricsPort int) error {
	registry := prometheus.NewRegistry()

	metrics := &PrometheusMetrics{
		registry: registry,
	}

	// 定义网关移动请求计数器
	metrics.GatewayMoveRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_move_requests_total",
			Help: "Total number of move requests received by gateway",
			ConstLabels: map[string]string{
				"service": "gateway",
			},
		},
		[]string{"map_id", "result"}, // result: success, blocked, error
	)

	// 定义验证耗时直方图
	metrics.GatewayValidationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_validation_duration_seconds",
			Help:    "Time spent validating move requests",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"instance_id"},
	)

	// 定义广播消息计数器
	metrics.GatewayBroadcastMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_broadcast_messages_total",
			Help: "Total number of broadcast messages sent",
		},
		[]string{"type", "scope"}, // scope: local, cross_gateway
	)

	// 定义当前连接数
	metrics.GatewayConnectionsCurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gateway_connections_current",
		Help: "Current number of active WebSocket connections",
	})

	// 定义总连接数
	metrics.GatewayConnectionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_connections_total",
		Help: "Total number of connections since startup",
	})

	// 定义路由器请求计数器
	metrics.RouterRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_requests_total",
			Help: "Total number of routing requests",
		},
		[]string{"strategy", "instance_id", "status"},
	)

	// 定义路由请求耗时
	metrics.RouterRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "router_request_duration_seconds",
			Help:    "Time spent processing routing requests",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"strategy"},
	)

	// 定义降级请求数
	metrics.RouterFallbackRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "router_fallback_requests_total",
		Help: "Total number of fallback routing requests",
	})

	// 定义熔断器触发次数
	metrics.RouterCircuitBreakerTrips = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_circuit_breaker_trips_total",
			Help: "Total number of circuit breaker trips",
		},
		[]string{"instance_id", "state"},
	)

	// 定义实例健康状态
	metrics.RouterInstanceHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "router_instance_health",
			Help: "Health status of game service instances (1=healthy, 0=unhealthy)",
		},
		[]string{"instance_id", "url"},
	)

	// 定义广播总数
	metrics.BroadcastTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cross_gateway_broadcast_total",
			Help: "Total number of cross-gateway broadcasts",
		},
		[]string{"channel"}, // channel: redis, rabbitmq
	)

	// 定义广播延迟
	metrics.BroadcastLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cross_gateway_broadcast_latency_seconds",
			Help:    "Latency of cross-gateway broadcasts",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
		[]string{"channel"},
	)

	// 定义重复消息数
	metrics.BroadcastDuplicatedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "broadcast_duplicated_messages_total",
		Help: "Total number of duplicated messages filtered out",
	})

	// 定义失败广播数
	metrics.BroadcastFailedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "broadcast_failed_messages_total",
		Help: "Total number of failed broadcasts",
	})

	// 定义怪物同步更新计数器
	metrics.MonsterSyncUpdatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "monster_sync_updates_total",
			Help: "Total number of monster sync updates",
		},
		[]string{"update_type"}, // position, hp, status, spawn, death
	)

	// 定义批量大小分布
	metrics.MonsterSyncBatchSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "monster_sync_batch_size",
			Help:    "Distribution of monster sync batch sizes",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
		},
		[]string{},
	)

	// 定义怪物同步延迟
	metrics.MonsterSyncLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "monster_sync_latency_seconds",
			Help:    "Time spent syncing monster state",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{},
	)

	// 定义活跃怪物数量
	metrics.MonsterActiveCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "monster_active_count",
			Help: "Number of active (non-dead) monsters per map",
		},
		[]string{"map_id"},
	)

	// 定义在线玩家数量
	metrics.PlayerOnlineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "player_online_count",
		Help: "Current number of online players",
	})

	// 定义地图玩家数量
	metrics.MapPlayerCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "map_player_count",
			Help: "Number of players in each map",
		},
		[]string{"map_id"},
	)

	// 定义消息处理耗时
	metrics.MessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "message_processing_duration_seconds",
			Help:    "Time spent processing different message types",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"message_type"},
	)

	// ★ 新增：定义代理请求指标
	metrics.ProxyRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_proxy_requests_total",
			Help: "Total number of proxy requests forwarded to GameService",
		},
		[]string{"path", "result"}, // result: received, success, error
	)

	metrics.ProxyRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_proxy_request_duration_seconds",
			Help:    "Time spent processing proxy requests to GameService",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"path"},
	)

	// 注册所有指标
	mustRegister := []prometheus.Collector{
		metrics.GatewayMoveRequestsTotal,
		metrics.GatewayValidationDuration,
		metrics.GatewayBroadcastMessagesTotal,
		metrics.GatewayConnectionsCurrent,
		metrics.GatewayConnectionsTotal,
		metrics.RouterRequestsTotal,
		metrics.RouterRequestDuration,
		metrics.RouterFallbackRequestsTotal,
		metrics.RouterCircuitBreakerTrips,
		metrics.RouterInstanceHealth,
		metrics.BroadcastTotal,
		metrics.BroadcastLatency,
		metrics.BroadcastDuplicatedMessages,
		metrics.BroadcastFailedMessages,
		metrics.MonsterSyncUpdatesTotal,
		metrics.MonsterSyncBatchSize,
		metrics.MonsterSyncLatency,
		metrics.MonsterActiveCount,
		metrics.PlayerOnlineCount,
		metrics.MapPlayerCount,
		metrics.MessageProcessingDuration,
		// ★ 新增：注册代理指标
		metrics.ProxyRequestsTotal,
		metrics.ProxyRequestDuration,
	}

	for _, collector := range mustRegister {
		if err := registry.Register(collector); err != nil {
			return fmt.Errorf("注册Prometheus指标失败: %v", err)
		}
	}

	promMetrics = metrics

	// 启动HTTP服务暴露metrics
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		addr := fmt.Sprintf(":%d", metricsPort)
		log.Printf(`📊 Prometheus指标服务启动: http://localhost%s/metrics`, addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf(`❌ Prometheus HTTP服务启动失败: %v`, err)
		}
	}()

	log.Printf(`✅ Prometheus监控初始化完成: port=%d`, metricsPort)
	return nil
}

// RecordMoveRequest 记录移动请求
func RecordMoveRequest(mapID uint32, result string) {
	if promMetrics == nil {
		return
	}
	promMetrics.GatewayMoveRequestsTotal.WithLabelValues(fmt.Sprintf("%d", mapID), result).Inc()
}

// RecordValidationDuration 记录验证耗时
func RecordValidationDuration(instanceID uint32, duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.GatewayValidationDuration.WithLabelValues(fmt.Sprintf("%d", instanceID)).Observe(duration.Seconds())
}

// RecordBroadcastMessage 记录广播消息
func RecordBroadcastMessage(msgType string, scope string) {
	if promMetrics == nil {
		return
	}
	promMetrics.GatewayBroadcastMessagesTotal.WithLabelValues(msgType, scope).Inc()
}

// UpdateConnectionCount 更新连接数
func UpdateConnectionCount(count float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.GatewayConnectionsCurrent.Set(count)
}

// IncrTotalConnections 增加总连接数
func IncrTotalConnections() {
	if promMetrics == nil {
		return
	}
	promMetrics.GatewayConnectionsTotal.Inc()
}

// RecordRoutingRequest 记录路由请求
func RecordRoutingRequest(strategy, instanceID, status string) {
	if promMetrics == nil {
		return
	}
	promMetrics.RouterRequestsTotal.WithLabelValues(strategy, instanceID, status).Inc()
}

// RecordRoutingDuration 记录路由耗时
func RecordRoutingDuration(strategy string, duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.RouterRequestDuration.WithLabelValues(strategy).Observe(duration.Seconds())
}

// IncrFallbackRouting 增加降级路由计数
func IncrFallbackRouting() {
	if promMetrics == nil {
		return
	}
	promMetrics.RouterFallbackRequestsTotal.Inc()
}

// RecordCircuitBreakerTrip 记录熔断器触发
func RecordCircuitBreakerTrip(instanceID uint32, state string) {
	if promMetrics == nil {
		return
	}
	promMetrics.RouterCircuitBreakerTrips.WithLabelValues(fmt.Sprintf("%d", instanceID), state).Inc()
}

// UpdateInstanceHealth 更新实例健康状态
func UpdateInstanceHealth(instanceID uint32, url string, health float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.RouterInstanceHealth.WithLabelValues(fmt.Sprintf("%d", instanceID), url).Set(health)
}

// RecordCrossGatewayBroadcast 记录跨网关广播
func RecordCrossGatewayBroadcast(channel string) {
	if promMetrics == nil {
		return
	}
	promMetrics.BroadcastTotal.WithLabelValues(channel).Inc()
}

// RecordBroadcastLatency 记录广播延迟
func RecordBroadcastLatency(channel string, duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.BroadcastLatency.WithLabelValues(channel).Observe(duration.Seconds())
}

// IncrDuplicatedMessages 增加重复消息计数
func IncrDuplicatedMessages() {
	if promMetrics == nil {
		return
	}
	promMetrics.BroadcastDuplicatedMessages.Inc()
}

// IncrFailedBroadcasts 增加失败广播计数
func IncrFailedBroadcasts() {
	if promMetrics == nil {
		return
	}
	promMetrics.BroadcastFailedMessages.Inc()
}

// RecordMonsterSyncUpdate 记录怪物同步更新
func RecordMonsterSyncUpdate(updateType string) {
	if promMetrics == nil {
		return
	}
	promMetrics.MonsterSyncUpdatesTotal.WithLabelValues(updateType).Inc()
}

// ObserveMonsterBatchSize 观察批量大小
func ObserveMonsterBatchSize(size float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.MonsterSyncBatchSize.WithLabelValues().Observe(size)
}

// ObserveMonsterSyncLatency 观察怪物同步延迟
func ObserveMonsterSyncLatency(duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.MonsterSyncLatency.WithLabelValues().Observe(duration.Seconds())
}

// UpdateMonsterActiveCount 更新活跃怪物数量
func UpdateMonsterActiveCount(mapID uint32, count float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.MonsterActiveCount.WithLabelValues(fmt.Sprintf("%d", mapID)).Set(count)
}

// UpdatePlayerOnlineCount 更新在线玩家数量
func UpdatePlayerOnlineCount(count float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.PlayerOnlineCount.Set(count)
}

// UpdateMapPlayerCount 更新地图玩家数量
func UpdateMapPlayerCount(mapID uint32, count float64) {
	if promMetrics == nil {
		return
	}
	promMetrics.MapPlayerCount.WithLabelValues(fmt.Sprintf("%d", mapID)).Set(count)
}

// RecordMessageProcessing 记录消息处理耗时
func RecordMessageProcessing(msgType string, duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.MessageProcessingDuration.WithLabelValues(msgType).Observe(duration.Seconds())
}

// ★ 新增：RecordProxyRequest 记录代理请求
func RecordProxyRequest(path, result string) {
	if promMetrics == nil {
		return
	}
	promMetrics.ProxyRequestsTotal.WithLabelValues(path, result).Inc()
}

// ★ 新增：RecordProxyDuration 记录代理请求耗时
func RecordProxyDuration(path string, duration time.Duration) {
	if promMetrics == nil {
		return
	}
	promMetrics.ProxyRequestDuration.WithLabelValues(path).Observe(duration.Seconds())
}
