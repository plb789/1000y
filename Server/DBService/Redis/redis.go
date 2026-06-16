package Redis

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v3"
)

var RDB *redis.Client
var ctx = context.Background()

type Config struct {
	Redis struct {
		Addr         string `yaml:"addr"`
		Password     string `yaml:"password"`
		DB           int    `yaml:"db"`
		PoolSize     int    `yaml:"pool_size"`      // 连接池大小
		MinIdleConns int    `yaml:"min_idle_conns"` // 最小空闲连接数
	} `yaml:"redis"`
}

var AppConfig *Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	AppConfig = &Config{}
	return yaml.Unmarshal(data, AppConfig)
}

type RedisConfig struct {
	Addr         string `yaml:"addr"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`      // 连接池大小
	MinIdleConns int    `yaml:"min_idle_conns"` // 最小空闲连接数
}

// InitWithConfig 使用外部配置初始化
func InitWithConfig(cfg *RedisConfig) {
	// 获取连接池配置
	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 50
	}
	minIdle := cfg.MinIdleConns
	if minIdle <= 0 {
		minIdle = 5
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
	})

	// 实际测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		log.Printf("Redis 连接失败: %v (地址: %s)", err, cfg.Addr)
		RDB = nil // 设置为nil表示未连接
	} else {
		log.Printf("Redis 连接成功 (地址: %s, 连接池: PoolSize=%d, MinIdle=%d)", cfg.Addr, poolSize, minIdle)
	}
}

func Init() {
	if AppConfig == nil {
		if err := LoadConfig("./Config/DB.yaml"); err != nil {
			log.Println("Redis配置加载失败，使用默认配置")
			AppConfig = &Config{}
			AppConfig.Redis.Addr = "localhost:6379"
			AppConfig.Redis.PoolSize = 50
			AppConfig.Redis.MinIdleConns = 5
		}
	}

	// 转换为 RedisConfig 并调用 InitWithConfig
	cfg := &RedisConfig{
		Addr:         AppConfig.Redis.Addr,
		Password:     AppConfig.Redis.Password,
		DB:           AppConfig.Redis.DB,
		PoolSize:     AppConfig.Redis.PoolSize,
		MinIdleConns: AppConfig.Redis.MinIdleConns,
	}
	InitWithConfig(cfg)
}

// Set 设置键值对
func Set(key string, value string, expire time.Duration) error {
	return RDB.Set(ctx, key, value, expire).Err()
}

// Get 获取键值
func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// Del 删除键
func Del(key string) error {
	return RDB.Del(ctx, key).Err()
}
