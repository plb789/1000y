package mysql

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

type Config struct {
	MySQL struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
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

func Init() error {
	if AppConfig == nil {
		if err := LoadConfig("../../Config/DB.yaml"); err != nil {
			log.Println("MySQL配置加载失败，使用默认配置")
			AppConfig = &Config{}
			AppConfig.MySQL.Host = "localhost"
			AppConfig.MySQL.Port = 3306
			AppConfig.MySQL.User = "root"
			AppConfig.MySQL.Password = "123456"
			AppConfig.MySQL.Database = "millennium"
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		AppConfig.MySQL.User,
		AppConfig.MySQL.Password,
		AppConfig.MySQL.Host,
		AppConfig.MySQL.Port,
		AppConfig.MySQL.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	log.Println("MySQL 连接成功")
	return nil
}
