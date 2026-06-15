package redis

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
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
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

func Init() {
	if AppConfig == nil {
		// 尝试加载配置
		if err := LoadConfig("../../Config/DB.yaml"); err != nil {
			log.Println("Redis配置加载失败，使用默认配置")
			AppConfig = &Config{}
			AppConfig.Redis.Addr = "localhost:6379"
		}
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     AppConfig.Redis.Addr,
		Password: AppConfig.Redis.Password,
		DB:       AppConfig.Redis.DB,
	})
	log.Println("Redis 连接成功")
}

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// Del 删除键
func Del(key string) error {
	return RDB.Del(ctx, key).Err()
}
