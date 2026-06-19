package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageBus 消息总线接口
type MessageBus interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler func([]byte)) error
	IsAvailable() bool
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
			available: false,
		}
		if err := bus.Connect(); err != nil {
			log.Printf("RabbitMQ连接失败，降级到HTTP: %v", err)
			return NewHTTPBus(config.HTTPURL)
		}
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

// RabbitMQBus RabbitMQ消息总线
type RabbitMQBus struct {
	url       string
	available bool
	conn      *amqp.Connection
	channel   *amqp.Channel
}

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

	b.available = true
	log.Printf("RabbitMQ连接成功")
	return nil
}

func (b *RabbitMQBus) Publish(topic string, data interface{}) error {
	if !b.available {
		return fmt.Errorf("RabbitMQ不可用")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 发布消息到交换机
	err = b.channel.Publish(
		"game_events", // 交换机
		topic,         // 路由键（topic）
		false,         // 强制
		false,         // 立即
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
	if err != nil {
		log.Printf("RabbitMQ发布消息失败: %v", err)
		return err
	}

	return nil
}

func (b *RabbitMQBus) Subscribe(topic string, handler func([]byte)) error {
	if !b.available {
		return fmt.Errorf("RabbitMQ不可用")
	}

	// 创建队列
	queue, err := b.channel.QueueDeclare(
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
	err = b.channel.QueueBind(
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
	msgs, err := b.channel.Consume(
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
			handler(msg.Body)
		}
	}()

	log.Printf("RabbitMQ订阅成功: topic=%s", topic)
	return nil
}

func (b *RabbitMQBus) IsAvailable() bool {
	return b.available
}

func (b *RabbitMQBus) Close() error {
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

// KafkaBus Kafka消息总线（预留接口）
type KafkaBus struct {
	brokers   []string
	available bool
}

func (b *KafkaBus) Connect() error {
	// Kafka 连接实现（需要 kafka 库）
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

func (b *KafkaBus) Subscribe(topic string, handler func([]byte)) error {
	if !b.available {
		return fmt.Errorf("Kafka不可用")
	}
	log.Printf("Kafka订阅: %s", topic)
	return nil
}

func (b *KafkaBus) IsAvailable() bool {
	return b.available
}

func (b *KafkaBus) Close() error {
	log.Printf("Kafka连接已关闭")
	b.available = false
	return nil
}

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

func (b *HTTPBus) Publish(topic string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 根据topic选择正确的路由
	var url string
	switch topic {
	case "map_move":
		url = b.gatewayURL + "/internal/broadcast_map"
	default:
		url = b.gatewayURL + "/internal/broadcast/" + topic
	}

	resp, err := b.client.Post(url, "application/json", bytes.NewReader(jsonData))
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

func (b *HTTPBus) Subscribe(topic string, handler func([]byte)) error {
	log.Printf("HTTP订阅: %s (HTTP模式不支持订阅，使用轮询或WebSocket)", topic)
	return nil
}

func (b *HTTPBus) IsAvailable() bool {
	return true
}

func (b *HTTPBus) Close() error {
	return nil
}

// GlobalMessageBus 全局消息总线实例
var GlobalMessageBus MessageBus

// InitMessageBus 初始化全局消息总线
func InitMessageBus(config MessageBusConfig) {
	GlobalMessageBus = NewMessageBus(config)
	log.Printf("消息总线初始化完成: %T", GlobalMessageBus)
}
