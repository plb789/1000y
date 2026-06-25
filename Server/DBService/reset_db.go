//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 删除并重建数据库
	fmt.Println("正在删除数据库...")
	if err := db.Exec("DROP DATABASE IF EXISTS millennium").Error; err != nil {
		log.Fatal("删除数据库失败:", err)
	}
	fmt.Println("数据库已删除")

	fmt.Println("正在创建数据库...")
	if err := db.Exec("CREATE DATABASE millennium CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		log.Fatal("创建数据库失败:", err)
	}
	fmt.Println("数据库已创建: millennium")

	// 连接新数据库
	dsn2 := "root:root@tcp(127.0.0.1:3306)/millennium?charset=utf8mb4&parseTime=True&loc=Local"
	db2, err := gorm.Open(mysql.Open(dsn2), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("连接新数据库失败:", err)
	}

	// 导入Model包并执行AutoMigrate
	fmt.Println("\n正在创建表结构...")

	// 直接导入需要迁移的模型
	type Account struct {
		ID            uint64     `gorm:"primaryKey;comment:账号ID"`
		Username      string     `gorm:"size:32;unique;comment:用户名"`
		Password      string     `gorm:"size:128;comment:密码(bcrypt哈希)"`
		Salt          string     `gorm:"size:32;comment:密码盐值"`
		Status        int        `gorm:"default:0;comment:状态(0正常 1封禁)"`
		LoginIP       string     `gorm:"size:32;comment:登录IP"`
		LastLoginTime *time.Time `gorm:"comment:最后登录时间"`
		LastLoginIP   string     `gorm:"size:32;comment:最后登录IP"`
	}

	type Role struct {
		ID           uint64    `gorm:"primaryKey;comment:角色ID"`
		AccountID    uint64    `gorm:"index;comment:账号ID"`
		Name         string    `gorm:"size:32;unique;comment:角色名"`
		Level        int       `gorm:"default:1;comment:等级"`
		Exp          int64     `gorm:"default:0;comment:经验值"`
		Gold         int64     `gorm:"default:0;comment:金币"`
		BindGold     int64     `gorm:"default:0;comment:绑定金币"`
		Yuanbao      int64     `gorm:"default:0;comment:元宝"`
		Gender       uint8     `gorm:"default:0;comment:性别(0男 1女)"`
		Appearance   uint32    `gorm:"default:0;comment:外观造型"`
		Hp           int       `gorm:"default:100;comment:当前生命"`
		MaxHp        int       `gorm:"default:100;comment:最大生命"`
		Mp           int       `gorm:"default:100;comment:当前内力"`
		MaxMp        int       `gorm:"default:100;comment:最大内力"`
		Stamina      int       `gorm:"default:100;comment:当前体力"`
		MaxStamina   int       `gorm:"default:100;comment:最大体力"`
		Attack       int       `gorm:"default:10;comment:攻击"`
		Defense      int       `gorm:"default:5;comment:防御"`
		Speed        int       `gorm:"default:10;comment:速度"`
		Hit          int       `gorm:"default:50;comment:命中"`
		Dodge        int       `gorm:"default:10;comment:闪避"`
		Crit         int       `gorm:"default:5;comment:暴击率"`
		CritDamage   int       `gorm:"default:150;comment:暴击伤害"`
		MapID        int       `gorm:"default:1;comment:所在地图ID"`
		MapX         int       `gorm:"default:100;comment:地图X坐标"`
		MapY         int       `gorm:"default:100;comment:地图Y坐标"`
		PkMode       uint8     `gorm:"default:0;comment:PK模式"`
		PkValue      int       `gorm:"default:0;comment:惩罚值"`
		Status       uint8     `gorm:"default:0;comment:状态"`
		KillCount    int       `gorm:"default:0;comment:击杀数"`
		DeathCount   int       `gorm:"default:0;comment:死亡数"`
		Title        string    `gorm:"size:64;comment:称号"`
		CreateTime   time.Time `gorm:"comment:创建时间"`
		LastLogin    time.Time `gorm:"comment:最后登录时间"`
		LastLoginIP  string    `gorm:"size:32;comment:最后登录IP"`
		LogoutTime   time.Time `gorm:"comment:最后登出时间"`
		LastSaveTime time.Time `gorm:"comment:最后保存时间"`
	}

	type RoleSkill struct {
		ID      uint64 `gorm:"primaryKey;comment:记录ID"`
		RoleID  uint64 `gorm:"index;comment:角色ID"`
		SkillID uint32 `gorm:"comment:武学ID"`
		Level   int    `gorm:"default:1;comment:武学等级"`
		Exp     int64  `gorm:"default:0;comment:武学经验"`
		IsEquip uint8  `gorm:"default:0;comment:是否装备"`
	}

	type SkillBase struct {
		ID          uint64 `gorm:"primaryKey;comment:武学ID"`
		Name        string `gorm:"size:32;comment:武学名称"`
		Type        uint8  `gorm:"comment:武学类型"`
		Level       uint32 `gorm:"default:1;comment:武学等级"`
		MaxLevel    uint32 `gorm:"default:10;comment:最大等级"`
		ExpFactor   int    `gorm:"comment:升级经验系数"`
		HpBonus     int    `gorm:"comment:生命加成"`
		MpBonus     int    `gorm:"comment:内力加成"`
		AttackBonus int    `gorm:"comment:攻击加成"`
		DefBonus    int    `gorm:"comment:防御加成"`
		SpeedBonus  int    `gorm:"comment:速度加成"`
		HitBonus    int    `gorm:"comment:命中加成"`
		DodgeBonus  int    `gorm:"comment:闪避加成"`
		CritBonus   int    `gorm:"comment:暴击加成"`
		IsActive    uint8  `gorm:"default:1;comment:是否主动"`
	}

	type ItemBase struct {
		ID            uint64 `gorm:"primaryKey;comment:道具ID"`
		Name          string `gorm:"size:32;comment:道具名称"`
		Type          uint8  `gorm:"comment:道具类型"`
		Quality       uint8  `gorm:"comment:品质"`
		LevelReq      uint32 `gorm:"default:1;comment:等级要求"`
		StackMax      uint32 `gorm:"default:1;comment:最大堆叠"`
		Price         int    `gorm:"comment:售价"`
		HpRestore     int    `gorm:"column:hp_restore;comment:HP恢复值"`
		MpRestore     int    `gorm:"column:mp_restore;comment:MP恢复值"`
		IsDropable    uint8  `gorm:"default:1;comment:是否可丢弃"`
		IsSellable    uint8  `gorm:"default:1;comment:是否可出售"`
		IsDestroyable uint8  `gorm:"default:1;comment:是否可销毁"`
		EquipType     uint8  `gorm:"comment:装备类型"`
		Description   string `gorm:"size:256;comment:道具描述"`
	}

	type RoleBag struct {
		ID         uint64    `gorm:"primaryKey;comment:背包记录ID"`
		RoleID     uint64    `gorm:"index;comment:角色ID"`
		GridIndex  int       `gorm:"comment:背包格子索引"`
		ItemID     uint32    `gorm:"comment:道具ID"`
		Count      uint32    `gorm:"default:1;comment:道具数量"`
		IsBind     uint8     `gorm:"default:0;comment:是否绑定"`
		DurCurrent *int      `gorm:"comment:当前耐久度"`
		DurMax     *int      `gorm:"comment:最大耐久度"`
		GetTime    time.Time `gorm:"comment:获得时间"`
	}

	type RoleEquipment struct {
		ID        uint64    `gorm:"primaryKey;comment:装备记录ID"`
		RoleID    uint64    `gorm:"index;comment:角色ID"`
		EquipType uint8     `gorm:"comment:装备位置"`
		BagItemID *uint64   `gorm:"comment:背包物品ID"`
		EquipTime time.Time `gorm:"comment:装备时间"`
	}

	type Friend struct {
		ID         uint64 `gorm:"primaryKey;comment:记录ID"`
		RoleID     uint64 `gorm:"index;comment:角色ID"`
		FriendID   uint64 `gorm:"index;comment:好友ID"`
		FriendName string `gorm:"size:32;comment:好友名称"`
		Status     uint8  `gorm:"default:0;comment:状态(0待确认 1已添加 2黑名单)"`
	}

	type FriendRequest struct {
		ID           uint64     `gorm:"primaryKey;comment:申请ID"`
		FromRoleID   uint64     `gorm:"index;comment:申请人ID"`
		FromRoleName string     `gorm:"size:32;comment:申请人名称"`
		ToRoleID     uint64     `gorm:"index;comment:被申请人ID"`
		Message      string     `gorm:"size:64;comment:申请留言"`
		Status       uint8      `gorm:"default:0;comment:状态"`
		HandledAt    *time.Time `gorm:"comment:处理时间"`
	}

	type Mail struct {
		ID          uint64    `gorm:"primaryKey;comment:邮件ID"`
		Title       string    `gorm:"size:64;comment:邮件标题"`
		Content     string    `gorm:"size:1024;comment:邮件内容"`
		FromRoleID  uint64    `gorm:"comment:发送者ID"`
		FromName    string    `gorm:"size:32;comment:发送者名称"`
		ToRoleID    uint64    `gorm:"index;comment:接收者ID"`
		ItemID      uint32    `gorm:"default:0;comment:附件物品ID"`
		ItemCount   uint32    `gorm:"default:0;comment:附件物品数量"`
		Gold        int64     `gorm:"default:0;comment:附件金币"`
		Yuanbao     int64     `gorm:"default:0;comment:附件元宝"`
		IsRead      uint8     `gorm:"default:0;comment:是否已读"`
		IsGetAttach uint8     `gorm:"default:0;comment:是否领取附件"`
		ExpiredAt   time.Time `gorm:"comment:过期时间"`
	}

	type TaskBase struct {
		ID           uint64 `gorm:"primaryKey;comment:任务ID"`
		Name         string `gorm:"size:32;comment:任务名称"`
		Type         uint8  `gorm:"default:1;comment:任务类型(1主线 2支线 3日常 4周常 5活动)"`
		Desc         string `gorm:"size:256;comment:任务描述"`
		TargetType   uint8  `gorm:"comment:目标类型"`
		TargetID     uint32 `gorm:"comment:目标ID"`
		TargetCount  uint32 `gorm:"default:1;comment:目标数量"`
		ExpReward    int64  `gorm:"default:0;comment:经验奖励"`
		GoldReward   int64  `gorm:"default:0;comment:金币奖励"`
		HonorReward  int64  `gorm:"default:0;comment:声望奖励"`
		ItemReward   uint32 `gorm:"default:0;comment:物品奖励ID"`
		ItemCount    uint32 `gorm:"default:0;comment:物品奖励数量"`
		LevelReq     uint32 `gorm:"default:1;comment:等级要求"`
		PreTaskID    uint32 `gorm:"default:0;comment:前置任务ID"`
		Repeatable   uint8  `gorm:"default:0;comment:是否可重复"`
		TimeLimit    int    `gorm:"default:0;comment:时间限制(秒)"`
		AutoAccept   uint8  `gorm:"default:0;comment:是否自动接取"`
		AutoComplete uint8  `gorm:"default:0;comment:是否自动完成"`
	}

	type RoleTask struct {
		ID           uint64     `gorm:"primaryKey;comment:记录ID"`
		RoleID       uint64     `gorm:"index;comment:角色ID"`
		TaskID       uint32     `gorm:"index;comment:任务ID"`
		Status       uint8      `gorm:"default:0;comment:状态(0未接 1进行中 2已完成 3已领奖)"`
		Progress     uint32     `gorm:"default:0;comment:当前进度"`
		Objectives   string     `gorm:"type:text;comment:多目标进度JSON"`
		AcceptTime   time.Time  `gorm:"comment:接取时间"`
		CompleteTime *time.Time `gorm:"comment:完成时间"`
		FinishTime   *time.Time `gorm:"comment:领奖时间"`
		DailyCount   int        `gorm:"default:0;comment:今日完成次数"`
		TotalCount   int        `gorm:"default:0;comment:总完成次数"`
	}

	type Guild struct {
		ID          uint64 `gorm:"primaryKey;comment:公会ID"`
		Name        string `gorm:"size:32;unique;comment:公会名称"`
		LeaderID    uint64 `gorm:"comment:会长ID"`
		LeaderName  string `gorm:"size:32;comment:会长名称"`
		Level       uint32 `gorm:"default:1;comment:公会等级"`
		Exp         int64  `gorm:"default:0;comment:公会经验"`
		Notice      string `gorm:"size:256;comment:公会公告"`
		Gold        int64  `gorm:"default:0;comment:公会资金"`
		MemberCount uint32 `gorm:"default:0;comment:成员数量"`
		MaxMembers  uint32 `gorm:"default:50;comment:最大成员数"`
	}

	type GuildMember struct {
		ID         uint64    `gorm:"primaryKey;comment:成员ID"`
		GuildID    uint64    `gorm:"index;comment:公会ID"`
		RoleID     uint64    `gorm:"unique;comment:角色ID"`
		RoleName   string    `gorm:"size:32;comment:角色名称"`
		Level      uint32    `gorm:"comment:角色等级"`
		Title      uint8     `gorm:"default:1;comment:职位"`
		Contribute int64     `gorm:"default:0;comment:个人贡献"`
		JoinTime   time.Time `gorm:"comment:加入时间"`
		LastOnline time.Time `gorm:"comment:最后在线时间"`
	}

	type GuildApply struct {
		ID       uint64 `gorm:"primaryKey;comment:申请ID"`
		GuildID  uint64 `gorm:"index;comment:公会ID"`
		RoleID   uint64 `gorm:"index;comment:申请人ID"`
		RoleName string `gorm:"size:32;comment:申请人名称"`
		Level    uint32 `gorm:"comment:申请人等级"`
		Message  string `gorm:"size:64;comment:申请留言"`
		Status   uint8  `gorm:"default:0;comment:状态"`
	}

	type ChatLog struct {
		ID           uint64 `gorm:"primaryKey;comment:记录ID"`
		Channel      uint8  `gorm:"index;comment:聊天频道"`
		SenderID     uint64 `gorm:"index;comment:发送者ID"`
		SenderName   string `gorm:"size:32;comment:发送者名称"`
		ReceiverID   uint64 `gorm:"comment:接收者ID"`
		ReceiverName string `gorm:"size:32;comment:接收者名称"`
		Content      string `gorm:"size:256;comment:聊天内容"`
	}

	type Ranking struct {
		ID        uint64 `gorm:"primaryKey;comment:记录ID"`
		Type      uint8  `gorm:"index;comment:排行类型"`
		RoleID    uint64 `gorm:"index;unique;comment:角色ID"`
		Name      string `gorm:"size:32;comment:角色名称"`
		Value     int64  `gorm:"comment:排行值"`
		GuildID   uint64 `gorm:"comment:公会ID"`
		GuildName string `gorm:"size:32;comment:公会名称"`
	}

	type RoleSignIn struct {
		ID             uint64    `gorm:"primaryKey;comment:记录ID"`
		RoleID         uint64    `gorm:"unique;comment:角色ID"`
		TotalDays      uint32    `gorm:"default:0;comment:累计签到天数"`
		ContinuousDays uint32    `gorm:"default:0;comment:连续签到天数"`
		LastSignIn     time.Time `gorm:"comment:最后签到日期"`
		Month          uint32    `gorm:"comment:当前签到月份"`
	}

	type SignInReward struct {
		ID       uint64 `gorm:"primaryKey;comment:奖励ID"`
		Day      uint32 `gorm:"index;comment:第几天"`
		Type     uint8  `gorm:"comment:奖励类型"`
		ItemID   uint32 `gorm:"default:0;comment:物品ID"`
		Count    uint32 `gorm:"default:0;comment:物品数量"`
		IsDouble uint8  `gorm:"default:0;comment:是否双倍奖励日"`
	}

	type RechargeOrder struct {
		ID         uint64    `gorm:"primaryKey;comment:订单ID"`
		OrderID    string    `gorm:"size:64;unique;comment:订单号"`
		RoleID     uint64    `gorm:"index;comment:角色ID"`
		RoleName   string    `gorm:"size:32;comment:角色名称"`
		AccountID  uint64    `gorm:"comment:账号ID"`
		ProductID  uint32    `gorm:"comment:产品ID"`
		Amount     int64     `gorm:"comment:充值金额(分)"`
		Yuanbao    int64     `gorm:"comment:获得元宝"`
		PayTime    time.Time `gorm:"comment:支付时间"`
		Status     uint8     `gorm:"default:0;comment:状态"`
		Channel    string    `gorm:"size:32;comment:支付渠道"`
		NotifyData string    `gorm:"size:1024;comment:回调原始数据"`
	}

	type RechargeProduct struct {
		ID         uint64 `gorm:"primaryKey;comment:产品ID"`
		ProductID  uint32 `gorm:"unique;comment:产品ID"`
		Name       string `gorm:"size:32;comment:产品名称"`
		Amount     int64  `gorm:"comment:价格(分)"`
		Yuanbao    int64  `gorm:"comment:获得元宝"`
		FirstBonus int64  `gorm:"default:0;comment:首充赠送"`
		IsFirst    uint8  `gorm:"default:0;comment:是否首充特惠"`
		IsActive   uint8  `gorm:"default:1;comment:是否上架"`
		SortOrder  uint32 `gorm:"default:0;comment:排序"`
	}

	type RoleRecharge struct {
		ID            uint64    `gorm:"primaryKey;comment:记录ID"`
		RoleID        uint64    `gorm:"index;comment:角色ID"`
		TotalAmount   int64     `gorm:"default:0;comment:累计充值(分)"`
		TotalYuanbao  int64     `gorm:"default:0;comment:累计获得元宝"`
		FirstRecharge uint8     `gorm:"default:0;comment:是否已首充"`
		LastRecharge  time.Time `gorm:"comment:最后充值时间"`
	}

	// 执行自动迁移
	err = db2.AutoMigrate(
		&Account{},
		&Role{},
		&RoleSkill{},
		&SkillBase{},
		&ItemBase{},
		&RoleBag{},
		&RoleEquipment{},
		&Friend{},
		&FriendRequest{},
		&Mail{},
		&TaskBase{},
		&RoleTask{},
		&Guild{},
		&GuildMember{},
		&GuildApply{},
		&ChatLog{},
		&Ranking{},
		&RoleSignIn{},
		&SignInReward{},
		&RechargeOrder{},
		&RechargeProduct{},
		&RoleRecharge{},
	)
	if err != nil {
		log.Fatal("建表失败:", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("所有表创建完成！共创建 21 张表")
	fmt.Println("========================================")

	// 显示所有表
	var tables []string
	db2.Raw("SHOW TABLES").Scan(&tables)
	fmt.Println("\n已创建的表：")
	for i, t := range tables {
		fmt.Printf("  %d. %s\n", i+1, t)
	}

	// 添加表注释
	fmt.Println("\n正在添加表注释...")
	tableComments := map[string]string{
		"accounts":          "账号表",
		"roles":             "角色表",
		"role_skills":       "角色武学表",
		"skill_bases":       "武学基础表",
		"item_bases":        "道具基础表",
		"role_bags":         "角色背包表",
		"role_equipments":   "角色装备表",
		"friends":           "好友表",
		"friend_requests":   "好友申请表",
		"mails":             "邮件表",
		"task_bases":        "任务基础表",
		"role_tasks":        "角色任务进度表",
		"guilds":            "公会表",
		"guild_members":     "公会成员表",
		"guild_applies":     "公会申请表",
		"chats":             "聊天记录表",
		"rank_rases":        "排行榜表",
		"role_sign_ins":     "角色签到表",
		"sign_in_rewards":   "签到奖励配置表",
		"recharge_orders":   "充值订单表",
		"recharge_products": "充值产品表",
		"role_recharges":    "角色充值记录表",
	}
	for table, comment := range tableComments {
		db2.Exec(fmt.Sprintf("ALTER TABLE %s COMMENT = '%s'", table, comment))
		fmt.Printf("  ✓ %s\n", table)
	}

	// 显示表结构
	fmt.Println("\n========================================")
	fmt.Println("表结构及注释：")
	fmt.Println("========================================")
	for _, table := range tables {
		fmt.Printf("\n【%s】\n", table)
		var cols []struct {
			Field   string
			Type    string
			Comment string
		}
		db2.Raw("SHOW FULL COLUMNS FROM " + table).Scan(&cols)
		for _, col := range cols {
			if col.Comment != "" {
				fmt.Printf("  %s (%s) - %s\n", col.Field, col.Type, col.Comment)
			} else {
				fmt.Printf("  %s (%s)\n", col.Field, col.Type)
			}
		}
	}
}
