package model

import (
	"time"

	"gorm.io/gorm"
)

// Account 账号表
type Account struct {
	gorm.Model
	Username      string     `gorm:"size:32;unique"`
	Password      string     `gorm:"size:64"`
	Status        int        `gorm:"default:0"` // 0正常 1封禁
	LoginIP       string     `gorm:"size:32"`
	LastLoginTime *time.Time `gorm:"size:32"`
	LastLoginIP   string     `gorm:"size:32"`
}

// Role 角色表
type Role struct {
	gorm.Model
	AccountID uint
	Name      string `gorm:"size:32"`
	Level     int
	Hp        int
	Mp        int
	TileX     int
	TileY     int
}
