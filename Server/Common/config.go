package common

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	MySQL struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Database        string `yaml:"database"`
		MaxOpenConns    int    `yaml:"max_open_conns"`    // 最大打开连接数
		MaxIdleConns    int    `yaml:"max_idle_conns"`    // 最大空闲连接数
		ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 连接最大存活时间(秒)
	} `yaml:"mysql"`
	Redis struct {
		Addr         string `yaml:"addr"`
		Password     string `yaml:"password"`
		DB           int    `yaml:"db"`
		PoolSize     int    `yaml:"pool_size"`      // 连接池大小
		MinIdleConns int    `yaml:"min_idle_conns"` // 最小空闲连接数
	} `yaml:"redis"`
	Services struct {
		DBService       string `yaml:"db_service"`
		LoginService    string `yaml:"login_service"`
		GameService     string `yaml:"game_service"`
		GatewayService  string `yaml:"gateway_service"`  // Gateway服务地址
		RegistryService string `yaml:"registry_service"` // 服务注册中心
	} `yaml:"services"`
	HTTPPort int `yaml:"http_port"` // HTTP服务端口
	// GameService实例配置
	InstanceID         uint32   `yaml:"instance_id"`          // 实例ID
	HandledMaps        []uint32 `yaml:"handled_maps"`         // 处理的地图列表
	RegisterToRegistry bool     `yaml:"register_to_registry"` // 是否注册到服务发现
	// HTTP客户端超时配置
	HTTPClient struct {
		Timeout int `yaml:"timeout"` // 超时时间(秒)，默认10秒
	} `yaml:"http_client"`
	// WebSocket配置
	WebSocket struct {
		ReadBufferSize  int `yaml:"read_buffer_size"`  // 读缓冲区大小
		WriteBufferSize int `yaml:"write_buffer_size"` // 写缓冲区大小
		MaxMessageSize  int `yaml:"max_message_size"`  // 最大消息大小
		PingInterval    int `yaml:"ping_interval"`     // 心跳间隔(秒)
		PongTimeout     int `yaml:"pong_timeout"`      // 心跳超时(秒)
	} `yaml:"websocket"`
	// Gateway配置
	Gateway struct {
		SendChannelSize      int `yaml:"send_channel_size"`      // 每玩家发送通道缓冲大小
		BroadcastChannelSize int `yaml:"broadcast_channel_size"` // 广播消息队列大小
		MaxConnectionsPerIP  int `yaml:"max_connections_per_ip"` // 同一IP最大连接数
		ViewRange            int `yaml:"view_range"`             // 视野范围(像素)
		MoveMergeDelay       int `yaml:"move_merge_delay"`       // 移动消息合并延迟(毫秒)
	} `yaml:"gateway"`
	// WebSocket端口
	WSPort int `yaml:"ws_port"`
	// IP白名单
	IPWhite []string `yaml:"ip_white"`
	// 消息总线配置
	MessageBus struct {
		Type         string   `yaml:"type"`          // 消息总线类型: rabbitmq, kafka, http
		RabbitMQURL  string   `yaml:"rabbitmq_url"`  // RabbitMQ连接地址
		KafkaBrokers []string `yaml:"kafka_brokers"` // Kafka broker列表
	} `yaml:"message_bus"`
}

var AppConfig GlobalConfig

// GetHTTPTimeout 获取HTTP超时时间
func (c *GlobalConfig) GetHTTPTimeout() time.Duration {
	if c.HTTPClient.Timeout <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.HTTPClient.Timeout) * time.Second
}

// GetMySQLPoolConfig 获取MySQL连接池配置
func (c *GlobalConfig) GetMySQLPoolConfig() (maxOpen, maxIdle int, maxLifetime time.Duration) {
	maxOpen = c.MySQL.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 100
	}
	maxIdle = c.MySQL.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 20
	}
	maxLifetime = time.Duration(c.MySQL.ConnMaxLifetime) * time.Second
	if maxLifetime <= 0 {
		maxLifetime = 3600 * time.Second
	}
	return
}

// GetRedisPoolConfig 获取Redis连接池配置
func (c *GlobalConfig) GetRedisPoolConfig() (poolSize, minIdle int) {
	poolSize = c.Redis.PoolSize
	if poolSize <= 0 {
		poolSize = 50
	}
	minIdle = c.Redis.MinIdleConns
	if minIdle <= 0 {
		minIdle = 5
	}
	return
}

// GetWebSocketConfig 获取WebSocket配置
func (c *GlobalConfig) GetWebSocketConfig() (readBuf, writeBuf, maxMsg int, pingInterval, pongTimeout time.Duration) {
	readBuf = c.WebSocket.ReadBufferSize
	if readBuf <= 0 {
		readBuf = 4096
	}
	writeBuf = c.WebSocket.WriteBufferSize
	if writeBuf <= 0 {
		writeBuf = 4096
	}
	maxMsg = c.WebSocket.MaxMessageSize
	if maxMsg <= 0 {
		maxMsg = 65536
	}
	pingInterval = time.Duration(c.WebSocket.PingInterval) * time.Second
	if pingInterval <= 0 {
		pingInterval = 30 * time.Second
	}
	pongTimeout = time.Duration(c.WebSocket.PongTimeout) * time.Second
	if pongTimeout <= 0 {
		pongTimeout = 60 * time.Second
	}
	return
}

// GetGatewayConfig 获取Gateway配置
func (c *GlobalConfig) GetGatewayConfig() (sendChan, broadcastChan, maxConnPerIP, viewRange, moveMergeDelay int) {
	sendChan = c.Gateway.SendChannelSize
	if sendChan <= 0 {
		sendChan = 512
	}
	broadcastChan = c.Gateway.BroadcastChannelSize
	if broadcastChan <= 0 {
		broadcastChan = 5000
	}
	maxConnPerIP = c.Gateway.MaxConnectionsPerIP
	if maxConnPerIP <= 0 {
		maxConnPerIP = 10
	}
	viewRange = c.Gateway.ViewRange
	if viewRange <= 0 {
		viewRange = 500
	}
	moveMergeDelay = c.Gateway.MoveMergeDelay
	if moveMergeDelay <= 0 {
		moveMergeDelay = 50
	}
	return
}

// GetSendChannelSize 获取发送通道大小
func (c *GlobalConfig) GetSendChannelSize() int {
	if c.Gateway.SendChannelSize <= 0 {
		return 512
	}
	return c.Gateway.SendChannelSize
}

// GetBroadcastChannelSize 获取广播通道大小
func (c *GlobalConfig) GetBroadcastChannelSize() int {
	if c.Gateway.BroadcastChannelSize <= 0 {
		return 5000
	}
	return c.Gateway.BroadcastChannelSize
}

// GetWSPort 获取WebSocket端口
func (c *GlobalConfig) GetWSPort() int {
	if c.WSPort <= 0 {
		return 8080
	}
	return c.WSPort
}

// LoadConfig 加载全局yaml配置
func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &AppConfig)
}
