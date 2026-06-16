package Mysql

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

type Config struct {
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

type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxOpenConns    int    `yaml:"max_open_conns"`    // 最大打开连接数
	MaxIdleConns    int    `yaml:"max_idle_conns"`    // 最大空闲连接数
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 连接最大存活时间(秒)
}

// InitWithConfig 使用外部配置初始化
func InitWithConfig(cfg *MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return err
	}

	// 获取连接池配置
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 100
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 20
	}
	maxLifetime := time.Duration(cfg.ConnMaxLifetime) * time.Second
	if maxLifetime <= 0 {
		maxLifetime = 3600 * time.Second
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)

	DB = db
	log.Printf("MySQL 连接成功 (连接池: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v)", maxOpen, maxIdle, maxLifetime)
	return nil
}

func Init() error {
	if AppConfig == nil {
		if err := LoadConfig("./Config/DB.yaml"); err != nil {
			log.Println("MySQL配置加载失败，使用默认配置")
			AppConfig = &Config{}
			AppConfig.MySQL.Host = "localhost"
			AppConfig.MySQL.Port = 3306
			AppConfig.MySQL.User = "root"
			AppConfig.MySQL.Password = "123456"
			AppConfig.MySQL.Database = "millennium"
			AppConfig.MySQL.MaxOpenConns = 100
			AppConfig.MySQL.MaxIdleConns = 20
			AppConfig.MySQL.ConnMaxLifetime = 3600
		}
	}

	// 转换为 MySQLConfig 并调用 InitWithConfig
	cfg := &MySQLConfig{
		Host:            AppConfig.MySQL.Host,
		Port:            AppConfig.MySQL.Port,
		User:            AppConfig.MySQL.User,
		Password:        AppConfig.MySQL.Password,
		Database:        AppConfig.MySQL.Database,
		MaxOpenConns:    AppConfig.MySQL.MaxOpenConns,
		MaxIdleConns:    AppConfig.MySQL.MaxIdleConns,
		ConnMaxLifetime: AppConfig.MySQL.ConnMaxLifetime,
	}
	return InitWithConfig(cfg)
}
