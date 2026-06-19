package Model

import (
	"time"

	"gorm.io/gorm"
)

// RechargeOrder 充值订单表
type RechargeOrder struct {
	ID         uint64         `gorm:"primaryKey;comment:订单ID"`
	CreatedAt  time.Time      `gorm:"comment:创建时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	OrderID    string         `gorm:"size:64;unique;comment:订单号"`
	RoleID     uint64         `gorm:"index;comment:角色ID"`
	RoleName   string         `gorm:"size:32;comment:角色名称"`
	AccountID  uint64         `gorm:"comment:账号ID"`
	ProductID  uint32         `gorm:"comment:产品ID"`
	Amount     int64          `gorm:"comment:充值金额(分)"`
	Yuanbao    int64          `gorm:"comment:获得元宝"`
	PayTime    time.Time      `gorm:"comment:支付时间"`
	Status     uint8          `gorm:"default:0;comment:状态(0待支付 1已支付 2已发货 3失败)"`
	Channel    string         `gorm:"size:32;comment:支付渠道"`
	NotifyData string         `gorm:"size:1024;comment:回调原始数据"`
}

// RechargeProduct 充值产品表
type RechargeProduct struct {
	ID         uint64         `gorm:"primaryKey;comment:产品ID"`
	CreatedAt  time.Time      `gorm:"comment:创建时间"`
	UpdatedAt  time.Time      `gorm:"comment:更新时间"`
	DeletedAt  gorm.DeletedAt `gorm:"index;comment:删除时间"`
	ProductID  uint32         `gorm:"unique;comment:产品ID(外部)"`
	Name       string         `gorm:"size:32;comment:产品名称"`
	Amount     int64          `gorm:"comment:价格(分)"`
	Yuanbao    int64          `gorm:"comment:获得元宝"`
	FirstBonus int64          `gorm:"default:0;comment:首充赠送"`
	IsFirst    uint8          `gorm:"default:0;comment:是否首充特惠"`
	IsActive   uint8          `gorm:"default:1;comment:是否上架(0下架 1上架)"`
	SortOrder  uint32         `gorm:"default:0;comment:排序"`
}

// RoleRecharge 角色充值记录表
type RoleRecharge struct {
	ID            uint64         `gorm:"primaryKey;comment:记录ID"`
	CreatedAt     time.Time      `gorm:"comment:创建时间"`
	UpdatedAt     time.Time      `gorm:"comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间"`
	RoleID        uint64         `gorm:"index;comment:角色ID"`
	TotalAmount   int64          `gorm:"default:0;comment:累计充值(分)"`
	TotalYuanbao  int64          `gorm:"default:0;comment:累计获得元宝"`
	FirstRecharge uint8          `gorm:"default:0;comment:是否已首充(0否 1是)"`
	LastRecharge  time.Time      `gorm:"comment:最后充值时间"`
}
