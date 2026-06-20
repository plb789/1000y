package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageBus 消息总线接口
type MessageBus interface {
	// Publish 发布JSON消息
	Publish(topic string, data interface{}) error
	// PublishRaw 发布二进制消息（带元数据，用于二进制协议优化）
	// metadata 例如: {"map_id":"2", "cmd":"3101"}
	PublishRaw(topic string, metadata map[string]string, body []byte) error
	// Subscribe 订阅消息
	Subscribe(topic string, handler func([]byte)) error
	// SubscribeRaw 订阅消息（带元数据回调，用于二进制协议）
	SubscribeRaw(topic string, handler func(metadata map[string]string, body []byte)) error
	// IsAvailable 是否可用
	IsAvailable() bool
	// GetMode 获取当前模式: "rabbitmq", "http"
	GetMode() string
	// Close 关闭
	Close() error
}

// MessageBusConfig 消息总线配置
type MessageBusConfig struct {
	Type         string // rabbitmq, kafka, http
	RabbitMQURL  string
	KafkaBrokers []string
	HTTPURL      string
}

// NewMessageBus 创建消息总线实例
func NewMessageBus(config MessageBusConfig) MessageBus {
	if config.Type == "rabbitmq" {
		bus := &RabbitMQBus{
			url:       config.RabbitMQURL,
			httpURL:   config.HTTPURL,
			available: false,
			stopCh:    make(chan struct{}),
		}
		if err := bus.Connect(); err != nil {
			log.Printf("RabbitMQ连接失败，降级到HTTP: %v", err)
			bus.fallback = NewHTTPBus(config.HTTPURL)
			bus.fallbackActive = true
		}
		// 启动健康检查与自动恢复 goroutine
		go bus.healthCheckLoop()
		return bus
	}
	if config.Type == "kafka" {
		bus := &KafkaBus{
			brokers:   config.KafkaBrokers,
			available: false,
		}
		if err := bus.Connect(); err != nil {
			log.Printf("Kafka连接失败，降级到HTTP: %v", err)
			return NewHTTPBus(config.HTTPURL)
		}
		return bus
	}
	return NewHTTPBus(config.HTTPURL)
}

// ==================== RabbitMQ 实现 ====================

// RabbitMQBus RabbitMQ消息总线（支持运行时降级与自动恢复）
type RabbitMQBus struct {
	url            string
	httpURL        string
	available      bool
	conn           *amqp.Connection
	channel        *amqp.Channel
	mu             sync.RWMutex
	fallback       *HTTPBus // HTTP降级实例
	fallbackActive bool     // 当前是否处于降级状态
	stopCh         chan struct{}
	reconnectCount int64 // 重连次数计数（用于采样日志）
}

// Connect 连接RabbitMQ
func (b *RabbitMQBus) Connect() error {
	var err error

	// 尝试连接 RabbitMQ
	b.conn, err = amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("RabbitMQ连接失败: %v", err)
	}

	// 创建通道
	b.channel, err = b.conn.Channel()
	if err != nil {
		b.conn.Close()
		return fmt.Errorf("RabbitMQ通道创建失败: %v", err)
	}

	// 声明交换机（用于发布消息）
	err = b.channel.ExchangeDeclare(
		"game_events", // 交换机名称
		"topic",       // 类型
		true,          // 持久化
		false,         // 自动删除
		false,         // 内部
		false,         // 等待
		nil,
	)
	if err != nil {
		b.channel.Close()
		b.conn.Close()
		return fmt.Errorf("RabbitMQ交换机声明失败: %v", err)
	}

	b.mu.Lock()
	b.available = true
	b.fallbackActive = false
	b.mu.Unlock()
	log.Printf("RabbitMQ连接成功")
	return nil
}

// Publish 发布JSON消息到RabbitMQ
func (b *RabbitMQBus) Publish(topic string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return b.PublishRaw(topic, nil, jsonData)
}

// PublishRaw 发布二进制消息到RabbitMQ（带元数据）
func (b *RabbitMQBus) PublishRaw(topic string, metadata map[string]string, body []byte) error {
	// 检查是否需要走降级通道
	b.mu.RLock()
	useFallback := b.fallbackActive || !b.available
	b.mu.RUnlock()

	if useFallback {
		if b.fallback != nil {
			return b.fallback.PublishRaw(topic, metadata, body)
		}
		return fmt.Errorf("RabbitMQ不可用且无降级通道")
	}

	b.mu.RLock()
	ch := b.channel
	b.mu.RUnlock()
	if ch == nil {
		// 通道丢失，触发降级
		b.switchToFallback()
		if b.fallback != nil {
			return b.fallback.PublishRaw(topic, metadata, body)
		}
		return fmt.Errorf("RabbitMQ通道不可用")
	}

	// 构造AMQP Headers
	headers := amqp.Table{}
	for k, v := range metadata {
		headers[k] = v
	}

	err := ch.Publish(
		"game_events", // 交换机
		topic,         // 路由键
		false,         // 强制
		false,         // 立即
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        body,
			Headers:     headers,
		},
	)
	if err != nil {
		log.Printf("RabbitMQ发布消息失败，触发降级: %v", err)
		b.switchToFallback()
		// 降级后重试一次
		if b.fallback != nil {
			return b.fallback.PublishRaw(topic, metadata, body)
		}
		return err
	}

	return nil
}

// Subscribe 订阅消息
func (b *RabbitMQBus) Subscribe(topic string, handler func([]byte)) error {
	return b.SubscribeRaw(topic, func(_ map[string]string, body []byte) {
		handler(body)
	})
}

// SubscribeRaw 订阅消息（带元数据回调）
func (b *RabbitMQBus) SubscribeRaw(topic string, handler func(metadata map[string]string, body []byte)) error {
	b.mu.RLock()
	if !b.available || b.channel == nil {
		b.mu.RUnlock()
		return fmt.Errorf("RabbitMQ不可用")
	}
	ch := b.channel
	b.mu.RUnlock()

	// 创建队列
	queue, err := ch.QueueDeclare(
		"",    // 队列名称（随机）
		false, // 持久化
		true,  // 自动删除
		true,  // 独占
		false, // 等待
		nil,
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ队列声明失败: %v", err)
	}

	// 绑定队列到交换机
	err = ch.QueueBind(
		queue.Name,    // 队列名称
		topic,         // 路由键
		"game_events", // 交换机
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ队列绑定失败: %v", err)
	}

	// 开始消费
	msgs, err := ch.Consume(
		queue.Name, // 队列
		"",         // 消费者标签
		true,       // 自动确认
		false,      // 独占
		false,      // 本地
		false,      // 等待
		nil,
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ消费启动失败: %v", err)
	}

	// 异步处理消息
	go func() {
		for msg := range msgs {
			// 提取headers为map[string]string
			metadata := make(map[string]string)
			for k, v := range msg.Headers {
				metadata[k] = fmt.Sprintf("%v", v)
			}
			handler(metadata, msg.Body)
		}
	}()

	log.Printf("RabbitMQ订阅成功: topic=%s", topic)
	return nil
}

// IsAvailable 是否可用
func (b *RabbitMQBus) IsAvailable() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.available && !b.fallbackActive
}

// GetMode 获取当前模式
func (b *RabbitMQBus) GetMode() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.fallbackActive {
		return "http(fallback)"
	}
	if b.available {
		return "rabbitmq"
	}
	return "unavailable"
}

// switchToFallback 切换到HTTP降级模式
func (b *RabbitMQBus) switchToFallback() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.fallbackActive {
		return // 已经在降级模式
	}
	log.Printf("⚠️ RabbitMQ降级到HTTP模式")
	b.fallbackActive = true
	b.available = false
	if b.fallback == nil {
		b.fallback = NewHTTPBus(b.httpURL)
	}
	// 关闭旧连接
	if b.channel != nil {
		b.channel.Close()
		b.channel = nil
	}
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
}

// healthCheckLoop 健康检查与自动恢复循环
// 每10秒检查一次：如果处于降级状态，尝试重连RabbitMQ
func (b *RabbitMQBus) healthCheckLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.mu.RLock()
			needReconnect := b.fallbackActive || !b.available
			b.mu.RUnlock()

			if !needReconnect {
				continue
			}

			// 尝试重连
			atomic.AddInt64(&b.reconnectCount, 1)
			count := atomic.LoadInt64(&b.reconnectCount)
			if err := b.reconnect(); err != nil {
				// 采样日志：每6次打印1次（约1分钟1条）
				if count%6 == 0 {
					log.Printf("RabbitMQ重连失败(第%d次): %v", count, err)
				}
				continue
			}

			// 重连成功，切回RabbitMQ
			b.mu.Lock()
			b.available = true
			b.fallbackActive = false
			b.mu.Unlock()
			log.Printf("✅ RabbitMQ已恢复，从HTTP切回RabbitMQ模式 (经过%d次重连)", count)
			atomic.StoreInt64(&b.reconnectCount, 0)

		case <-b.stopCh:
			return
		}
	}
}

// reconnect 重连RabbitMQ
func (b *RabbitMQBus) reconnect() error {
	// 清理旧连接
	b.mu.Lock()
	if b.channel != nil {
		b.channel.Close()
		b.channel = nil
	}
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
	b.mu.Unlock()

	// 尝试新建连接
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	// 重新声明交换机
	err = ch.ExchangeDeclare(
		"game_events", "topic", true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	b.mu.Lock()
	b.conn = conn
	b.channel = ch
	b.mu.Unlock()
	return nil
}

// Close 关闭
func (b *RabbitMQBus) Close() error {
	close(b.stopCh)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
	log.Printf("RabbitMQ连接已关闭")
	b.available = false
	return nil
}

// ==================== Kafka 实现（预留）====================

// KafkaBus Kafka消息总线（预留接口）
type KafkaBus struct {
	brokers   []string
	available bool
}

func (b *KafkaBus) Connect() error {
	log.Printf("Kafka连接成功 (模拟)")
	b.available = true
	return nil
}

func (b *KafkaBus) Publish(topic string, data interface{}) error {
	if !b.available {
		return fmt.Errorf("Kafka不可用")
	}
	jsonData, _ := json.Marshal(data)
	log.Printf("Kafka发布消息: topic=%s, data=%s", topic, string(jsonData))
	return nil
}

func (b *KafkaBus) PublishRaw(topic string, metadata map[string]string, body []byte) error {
	if !b.available {
		return fmt.Errorf("Kafka不可用")
	}
	log.Printf("Kafka发布二进制: topic=%s, size=%d", topic, len(body))
	return nil
}

func (b *KafkaBus) Subscribe(topic string, handler func([]byte)) error {
	if !b.available {
		return fmt.Errorf("Kafka不可用")
	}
	log.Printf("Kafka订阅: %s", topic)
	return nil
}

func (b *KafkaBus) SubscribeRaw(topic string, handler func(map[string]string, []byte)) error {
	if !b.available {
		return fmt.Errorf("Kafka不可用")
	}
	log.Printf("Kafka订阅(带元数据): %s", topic)
	return nil
}

func (b *KafkaBus) IsAvailable() bool {
	return b.available
}

func (b *KafkaBus) GetMode() string {
	if b.available {
		return "kafka"
	}
	return "unavailable"
}

func (b *KafkaBus) Close() error {
	log.Printf("Kafka连接已关闭")
	b.available = false
	return nil
}

// ==================== HTTP 降级实现 ====================

// HTTPBus HTTP消息总线（降级方案）
type HTTPBus struct {
	gatewayURL string
	client     *http.Client
}

func NewHTTPBus(gatewayURL string) *HTTPBus {
	return &HTTPBus{
		gatewayURL: gatewayURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Publish HTTP发布JSON消息
func (b *HTTPBus) Publish(topic string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return b.PublishRaw(topic, nil, jsonData)
}

// PublishRaw HTTP发布二进制消息
// 根据topic选择对应的Gateway接口，元数据通过HTTP Header传递
func (b *HTTPBus) PublishRaw(topic string, metadata map[string]string, body []byte) error {
	var url string
	switch topic {
	case "map_move":
		url = b.gatewayURL + "/internal/broadcast_map"
	case "monster_position":
		// 怪物位置走二进制广播接口
		url = b.gatewayURL + "/internal/broadcast_binary"
	default:
		url = b.gatewayURL + "/internal/broadcast/" + topic
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	// 根据是否有元数据决定Content-Type
	if metadata != nil {
		// 二进制模式：元数据走Header
		req.Header.Set("Content-Type", "application/octet-stream")
		if v, ok := metadata["map_id"]; ok {
			req.Header.Set("X-Map-Id", v)
		}
		if v, ok := metadata["cmd"]; ok {
			req.Header.Set("X-Cmd", v)
		}
	} else {
		// JSON模式
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := b.client.Do(req)
	if err != nil {
		log.Printf("HTTP发布消息失败: topic=%s, url=%s, err=%v", topic, url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("HTTP发布消息失败: topic=%s, status=%d", topic, resp.StatusCode)
	}

	return nil
}

// Subscribe HTTP不支持订阅
func (b *HTTPBus) Subscribe(topic string, handler func([]byte)) error {
	log.Printf("HTTP订阅: %s (HTTP模式不支持订阅)", topic)
	return nil
}

// SubscribeRaw HTTP不支持订阅
func (b *HTTPBus) SubscribeRaw(topic string, handler func(map[string]string, []byte)) error {
	log.Printf("HTTP订阅(带元数据): %s (HTTP模式不支持订阅)", topic)
	return nil
}

func (b *HTTPBus) IsAvailable() bool {
	return true
}

func (b *HTTPBus) GetMode() string {
	return "http"
}

func (b *HTTPBus) Close() error {
	return nil
}

// ==================== 全局实例 ====================

// GlobalMessageBus 全局消息总线实例
var GlobalMessageBus MessageBus

// InitMessageBus 初始化全局消息总线
func InitMessageBus(config MessageBusConfig) {
	GlobalMessageBus = NewMessageBus(config)
	log.Printf("消息总线初始化完成: mode=%s", GlobalMessageBus.GetMode())
}

// PublishMonsterPositionBinary 便捷函数：通过消息总线发布怪物位置二进制数据
// 优先走RabbitMQ，不可用时自动降级到HTTP
func PublishMonsterPositionBinary(mapID uint32, cmd uint16, body []byte) error {
	if GlobalMessageBus == nil {
		return fmt.Errorf("消息总线未初始化")
	}
	metadata := map[string]string{
		"map_id": strconv.FormatUint(uint64(mapID), 10),
		"cmd":    strconv.FormatUint(uint64(cmd), 10),
	}
	return GlobalMessageBus.PublishRaw("monster_position", metadata, body)
}
