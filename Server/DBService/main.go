package main

import (
	"crypto/rand"
	"fmt"
	common "game-server/Common"
	"game-server/DBService/Model"
	"game-server/DBService/Mysql"
	"game-server/DBService/Redis"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 加载全局配置
	err := common.LoadConfig("./Config/DB.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化MySQL
	err = Mysql.Init()
	if err != nil {
		log.Fatal("MySQL初始化失败:", err)
	}
	// 自动建表
	err = Mysql.DB.AutoMigrate(
		&Model.Account{},
		&Model.Role{},
		&Model.RoleSkill{},
		&Model.SkillBase{},
		&Model.ItemBase{},
		&Model.RoleBag{},
		&Model.RoleEquipment{},
		&Model.Friend{},
		&Model.FriendRequest{},
		&Model.Mail{},
		&Model.TaskBase{},
		&Model.RoleTask{},
		&Model.Guild{},
		&Model.GuildMember{},
		&Model.GuildApply{},
		&Model.ChatLog{},
		&Model.RoleSignIn{},
		&Model.SignInReward{},
		&Model.RechargeOrder{},
		&Model.RechargeProduct{},
		&Model.RoleRecharge{},
	)
	if err != nil {
		log.Fatal("建表失败:", err)
	}

	// 初始化Redis
	Redis.Init()

	// 启动HTTP API服务
	go startHTTPServer()

	log.Println("===== 数据微服务启动完成 =====")
	select {} // 常驻
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 连接池状态监控
	r.GET("/api/stats/pool", handlePoolStats)

	// 用户账号相关接口
	accountGroup := r.Group("/api/account")
	{
		accountGroup.POST("/register", handleAccountRegister)
		accountGroup.POST("/login", handleAccountLogin)
		accountGroup.POST("/check", handleAccountCheck)
		accountGroup.POST("/update", handleAccountUpdate)
		accountGroup.POST("/reset_password", handleAccountResetPassword)
	}

	// 角色相关接口
	roleGroup := r.Group("/api/role")
	{
		roleGroup.POST("/create", handleRoleCreate)
		roleGroup.POST("/get", handleRoleGet)
		roleGroup.POST("/update", handleRoleUpdate)
		roleGroup.POST("/delete", handleRoleDelete)
		roleGroup.POST("/list", handleRoleList)
		roleGroup.POST("/update_position", handleRoleUpdatePosition)
		roleGroup.POST("/update_attributes", handleRoleUpdateAttributes)
		roleGroup.POST("/add_exp", handleRoleAddExp)
		roleGroup.POST("/change_hp", handleRoleChangeHP)
		roleGroup.POST("/set_hp", handleRoleSetHP)
		roleGroup.POST("/change_mp", handleRoleChangeMP)
		roleGroup.POST("/change_stamina", handleRoleChangeStamina)
		roleGroup.POST("/set_status", handleRoleSetStatus)
		roleGroup.POST("/add_gold", handleRoleAddGold)
		roleGroup.POST("/consume_gold", handleRoleConsumeGold)
		roleGroup.POST("/record_kill", handleRoleRecordKill)
		roleGroup.POST("/record_death", handleRoleRecordDeath)
		roleGroup.POST("/full_recovery", handleRoleFullRecovery)
		roleGroup.POST("/login_record", handleRoleLoginRecord)
		roleGroup.POST("/logout_record", handleRoleLogoutRecord)
	}

	// 武学相关接口
	skillGroup := r.Group("/api/skill")
	{
		skillGroup.POST("/learn", handleSkillLearn)
		skillGroup.POST("/get_list", handleSkillGetList)
		skillGroup.POST("/get_equipped", handleSkillGetEquipped)
		skillGroup.POST("/equip", handleSkillEquip)
		skillGroup.POST("/unequip", handleSkillUnequip)
		skillGroup.POST("/add_exp", handleSkillAddExp)
		skillGroup.POST("/forget", handleSkillForget)
		skillGroup.POST("/get_base", handleSkillGetBase)
		skillGroup.POST("/get_all_base", handleSkillGetAllBase)
	}

	// 道具相关接口
	itemGroup := r.Group("/api/item")
	{
		itemGroup.POST("/add", handleItemAdd)
		itemGroup.POST("/get_bag", handleItemGetBag)
		itemGroup.POST("/move", handleItemMove)
		itemGroup.POST("/split", handleItemSplit)
		itemGroup.POST("/use", handleItemUse)
		itemGroup.POST("/discard", handleItemDiscard)
		itemGroup.POST("/sell", handleItemSell)
		itemGroup.POST("/equip", handleItemEquip)
		itemGroup.POST("/unequip", handleItemUnequip)
		itemGroup.POST("/get_equipped", handleItemGetEquipped)
		itemGroup.POST("/get_base", handleItemGetBase)
		itemGroup.POST("/get_all_base", handleItemGetAllBase)
		itemGroup.POST("/get_empty_count", handleItemGetEmptyCount)
	}

	// 服务注册中心接口
	registryGroup := r.Group("/api/registry")
	{
		registryGroup.POST("/register", handleRegistryRegister)
		registryGroup.POST("/unregister", handleRegistryUnregister)
		registryGroup.POST("/heartbeat", handleRegistryHeartbeat)
		registryGroup.POST("/list", handleRegistryList)
		registryGroup.POST("/get_by_map", handleRegistryGetByMap)
	}

	// 初始化注册中心
	initRegistry()

	port := common.AppConfig.HTTPPort
	if port == 0 {
		port = 8083
	}

	log.Printf("DBService HTTP API 启动 :%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("HTTP服务启动失败:", err)
	}
}

// generateSalt 生成随机盐值
func generateSalt() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// handleAccountRegister 注册账号
func handleAccountRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var existing Model.Account
	if err := Mysql.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "用户名已存在"})
		return
	}

	salt := generateSalt()
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password+salt), bcrypt.DefaultCost)

	account := Model.Account{
		Username: req.Username,
		Password: string(hash),
		Salt:     salt,
		Status:   1,
	}

	if err := Mysql.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "注册成功", "data": account.ID})
}

// handleAccountLogin 账号登录验证
func handleAccountLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var account Model.Account
	if err := Mysql.DB.Where("username = ?", req.Username).First(&account).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "账号不存在"})
		return
	}

	if account.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "账号已封禁"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password+account.Salt)); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "密码错误"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "验证成功", "data": account.ID})
}

// handleAccountCheck 检查账号是否存在
func handleAccountCheck(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var count int64
	Mysql.DB.Model(&Model.Account{}).Where("username = ?", req.Username).Count(&count)

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": count > 0})
}

// handleAccountUpdate 更新账号信息
func handleAccountUpdate(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	if err := Mysql.DB.Model(&Model.Account{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleAccountResetPassword 重置密码
func handleAccountResetPassword(c *gin.Context) {
	var req struct {
		ID       uint64 `json:"id"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var account Model.Account
	if err := Mysql.DB.Where("id = ?", req.ID).First(&account).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "账号不存在"})
		return
	}

	salt := generateSalt()
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password+salt), bcrypt.DefaultCost)

	if err := Mysql.DB.Model(&account).Updates(map[string]interface{}{
		"password": hash,
		"salt":     salt,
	}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "重置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "重置成功"})
}

// handleRoleCreate 创建角色
func handleRoleCreate(c *gin.Context) {
	var req struct {
		AccountID  uint64 `json:"account_id"`
		Name       string `json:"name"`
		Gender     uint8  `json:"gender"`
		Appearance uint32 `json:"appearance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 输入验证
	if len(req.Name) < 2 || len(req.Name) > 12 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色名长度需2-12位"})
		return
	}
	if req.Gender > 1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "性别参数错误"})
		return
	}
	if req.Appearance > 10 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "外观参数错误"})
		return
	}

	now := time.Now()
	role := Model.Role{
		AccountID:    req.AccountID,
		Name:         req.Name,
		Gender:       req.Gender,
		Appearance:   req.Appearance,
		Level:        1,
		MapID:        1,
		MapX:         400,
		MapY:         300,
		Hp:           100,
		MaxHp:        100,
		Mp:           100,
		MaxMp:        100,
		Stamina:      100,
		MaxStamina:   100,
		Attack:       10,
		Defense:      5,
		Speed:        10,
		Hit:          50,
		Dodge:        10,
		Crit:         5,
		CritDamage:   150,
		CreateTime:   now,
		LastLogin:    now,
		LastSaveTime: now,
		LogoutTime:   now, // 设置为当前时间，避免0000-00-00错误
	}

	if err := Mysql.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "创建成功", "data": role.ID})
}

// handleRoleGet 获取角色信息
func handleRoleGet(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	// 返回小写字段名格式
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": map[string]interface{}{
			"id":          role.ID,
			"account_id":  role.AccountID,
			"name":        role.Name,
			"level":       role.Level,
			"exp":         role.Exp,
			"gold":        role.Gold,
			"gender":      role.Gender,
			"appearance":  role.Appearance,
			"hp":          role.Hp,
			"max_hp":      role.MaxHp,
			"mp":          role.Mp,
			"max_mp":      role.MaxMp,
			"stamina":     role.Stamina,
			"max_stamina": role.MaxStamina,
			"attack":      role.Attack,
			"defense":     role.Defense,
			"speed":       role.Speed,
			"hit":         role.Hit,
			"dodge":       role.Dodge,
			"crit":        role.Crit,
			"crit_damage": role.CritDamage,
			"map_id":      role.MapID,
			"map_x":       role.MapX,
			"map_y":       role.MapY,
		},
	})
}

// handleRoleUpdate 更新角色信息
func handleRoleUpdate(c *gin.Context) {
	var req Model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	if err := Mysql.DB.Save(&req).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleDelete 删除角色
func handleRoleDelete(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	if err := Mysql.DB.Delete(&Model.Role{}, req.ID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// handleRoleList 获取账号角色列表
func handleRoleList(c *gin.Context) {
	var req struct {
		AccountID uint64 `json:"account_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var roles []Model.Role
	if err := Mysql.DB.Where("account_id = ?", req.AccountID).Find(&roles).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}

	// 转换为前端需要的格式（小写字段名）
	roleList := make([]map[string]interface{}, len(roles))
	for i, r := range roles {
		roleList[i] = map[string]interface{}{
			"id":         r.ID,
			"account_id": r.AccountID,
			"name":       r.Name,
			"level":      r.Level,
			"gender":     r.Gender,
			"appearance": r.Appearance,
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": roleList})
}

// handleRoleUpdatePosition 更新角色位置
func handleRoleUpdatePosition(c *gin.Context) {
	var req struct {
		ID    uint64 `json:"id"`
		MapID int    `json:"map_id"`
		MapX  int    `json:"map_x"`
		MapY  int    `json:"map_y"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"map_id": req.MapID,
		"map_x":  req.MapX,
		"map_y":  req.MapY,
	}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleUpdateAttributes 批量更新角色属性
func handleRoleUpdateAttributes(c *gin.Context) {
	var req struct {
		ID       uint64 `json:"id"`
		Hp       *int   `json:"hp"`
		MaxHp    *int   `json:"max_hp"`
		Mp       *int   `json:"mp"`
		MaxMp    *int   `json:"max_mp"`
		Attack   *int   `json:"attack"`
		Defense  *int   `json:"defense"`
		Speed    *int   `json:"speed"`
		Gold     *int64 `json:"gold"`
		BindGold *int64 `json:"bind_gold"`
		Yuanbao  *int64 `json:"yuanbao"`
		Exp      *int64 `json:"exp"`
		Level    *int   `json:"level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	updates := make(map[string]interface{})
	if req.Hp != nil {
		updates["hp"] = *req.Hp
	}
	if req.MaxHp != nil {
		updates["max_hp"] = *req.MaxHp
	}
	if req.Mp != nil {
		updates["mp"] = *req.Mp
	}
	if req.MaxMp != nil {
		updates["max_mp"] = *req.MaxMp
	}
	if req.Attack != nil {
		updates["attack"] = *req.Attack
	}
	if req.Defense != nil {
		updates["defense"] = *req.Defense
	}
	if req.Speed != nil {
		updates["speed"] = *req.Speed
	}
	if req.Gold != nil {
		updates["gold"] = *req.Gold
	}
	if req.BindGold != nil {
		updates["bind_gold"] = *req.BindGold
	}
	if req.Yuanbao != nil {
		updates["yuanbao"] = *req.Yuanbao
	}
	if req.Exp != nil {
		updates["exp"] = *req.Exp
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "没有需要更新的字段"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleAddExp 增加经验值
func handleRoleAddExp(c *gin.Context) {
	var req struct {
		ID  uint64 `json:"id"`
		Exp int64  `json:"exp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	leveledUp := false
	currentLevel := role.Level
	role.Exp += req.Exp

	// 计算升级
	for {
		expNeeded := int64(currentLevel) * 100 * int64(currentLevel)
		if role.Exp >= expNeeded && currentLevel < 200 {
			role.Exp -= expNeeded
			currentLevel++
			leveledUp = true
		} else {
			break
		}
	}

	role.Level = currentLevel
	if err := Mysql.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": gin.H{
		"leveled_up": leveledUp,
		"level":      currentLevel,
		"exp":        role.Exp,
	}})
}

// handleRoleChangeHP 改变生命值
func handleRoleChangeHP(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id"`
		Change int    `json:"change"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	role.Hp += req.Change
	if role.Hp > role.MaxHp {
		role.Hp = role.MaxHp
	}
	if role.Hp < 0 {
		role.Hp = 0
	}

	if err := Mysql.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": role.Hp})
}

// handleRoleSetHP 设置生命值（绝对值）
func handleRoleSetHP(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
		HP int    `json:"hp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	// 确保HP在有效范围内
	if req.HP > role.MaxHp {
		req.HP = role.MaxHp
	}
	if req.HP < 0 {
		req.HP = 0
	}

	if err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Update("hp", req.HP).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": req.HP})
}

// handleRoleChangeMP 改变内力值
func handleRoleChangeMP(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id"`
		Change int    `json:"change"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	role.Mp += req.Change
	if role.Mp > role.MaxMp {
		role.Mp = role.MaxMp
	}
	if role.Mp < 0 {
		role.Mp = 0
	}

	if err := Mysql.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": role.Mp})
}

// handleRoleChangeStamina 改变体力值
func handleRoleChangeStamina(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id"`
		Change int    `json:"change"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}

	role.Stamina += req.Change
	if role.Stamina > role.MaxStamina {
		role.Stamina = role.MaxStamina
	}
	if role.Stamina < 0 {
		role.Stamina = 0
	}

	if err := Mysql.DB.Save(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": role.Stamina})
}

// handleRoleSetStatus 设置角色状态
func handleRoleSetStatus(c *gin.Context) {
	var req struct {
		ID     uint64 `json:"id"`
		Status uint8  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Update("status", req.Status).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleAddGold 增加金币
func handleRoleAddGold(c *gin.Context) {
	var req struct {
		ID   uint64 `json:"id"`
		Gold int64  `json:"gold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).
		Update("gold", Mysql.DB.Raw("gold + ?", req.Gold)).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleConsumeGold 消耗金币
func handleRoleConsumeGold(c *gin.Context) {
	var req struct {
		ID   uint64 `json:"id"`
		Gold int64  `json:"gold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var role Model.Role
	if err := Mysql.DB.Where("id = ?", req.ID).First(&role).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "角色不存在"})
		return
	}
	if role.Gold < req.Gold {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "金币不足"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).
		Update("gold", Mysql.DB.Raw("gold - ?", req.Gold)).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleRecordKill 记录击杀
func handleRoleRecordKill(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"kill_count": Mysql.DB.Raw("kill_count + 1"),
		"pk_value":   Mysql.DB.Raw("pk_value + 50"),
	}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleRecordDeath 记录死亡
func handleRoleRecordDeath(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).
		Update("death_count", Mysql.DB.Raw("death_count + 1")).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleFullRecovery 完全恢复
func handleRoleFullRecovery(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"hp":      Mysql.DB.Raw("max_hp"),
		"mp":      Mysql.DB.Raw("max_mp"),
		"stamina": Mysql.DB.Raw("max_stamina"),
	}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleLoginRecord 记录登录
func handleRoleLoginRecord(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
		IP string `json:"ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":        3,
		"last_login":    time.Now(),
		"last_login_ip": req.IP,
	}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// handleRoleLogoutRecord 记录登出
func handleRoleLogoutRecord(c *gin.Context) {
	var req struct {
		ID uint64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	now := time.Now()
	err := Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":         0,
		"logout_time":    now,
		"last_save_time": now,
	}).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// ========== 武学相关 Handler ==========

// handleSkillLearn 学习武学
func handleSkillLearn(c *gin.Context) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 检查武学是否存在
	var skillBase Model.SkillBase
	if err := Mysql.DB.Where("id = ? AND is_active = 1", req.SkillID).First(&skillBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "武学不存在或已下架"})
		return
	}

	// 检查是否已学习
	var existCount int64
	Mysql.DB.Model(&Model.RoleSkill{}).Where("role_id = ? AND skill_id = ?", req.RoleID, req.SkillID).Count(&existCount)
	if existCount > 0 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "已学习该武学"})
		return
	}

	// 创建武学记录
	roleSkill := Model.RoleSkill{
		RoleID:  req.RoleID,
		SkillID: req.SkillID,
		Level:   1,
		Exp:     0,
		IsEquip: 0,
	}
	if err := Mysql.DB.Create(&roleSkill).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "学习失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "学习成功"})
}

// handleSkillGetList 获取角色武学列表
func handleSkillGetList(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var skills []Model.RoleSkill
	if err := Mysql.DB.Where("role_id = ?", req.RoleID).Find(&skills).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": skills})
}

// handleSkillGetEquipped 获取已装备武学
func handleSkillGetEquipped(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var skills []Model.RoleSkill
	if err := Mysql.DB.Where("role_id = ? AND is_equip = 1", req.RoleID).Find(&skills).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": skills})
}

// handleSkillEquip 装备武学
func handleSkillEquip(c *gin.Context) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 获取武学信息
	var skillBase Model.SkillBase
	if err := Mysql.DB.Where("id = ?", req.SkillID).First(&skillBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "武学不存在"})
		return
	}

	// 外功类武学需要先卸下同类型
	if skillBase.Type >= 5 && skillBase.Type <= 9 {
		Mysql.DB.Model(&Model.RoleSkill{}).
			Joins("LEFT JOIN skill_base ON role_skill.skill_id = skill_base.id").
			Where("role_skill.role_id = ? AND skill_base.type = ? AND role_skill.is_equip = 1", req.RoleID, skillBase.Type).
			Update("is_equip", 0)
	}

	// 装备该武学
	err := Mysql.DB.Model(&Model.RoleSkill{}).
		Where("role_id = ? AND skill_id = ?", req.RoleID, req.SkillID).
		Update("is_equip", 1).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "装备失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "装备成功"})
}

// handleSkillUnequip 卸下武学
func handleSkillUnequip(c *gin.Context) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Model(&Model.RoleSkill{}).
		Where("role_id = ? AND skill_id = ?", req.RoleID, req.SkillID).
		Update("is_equip", 0).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "卸下失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "卸下成功"})
}

// handleSkillAddExp 增加武学熟练度
func handleSkillAddExp(c *gin.Context) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
		Exp     int64  `json:"exp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var roleSkill Model.RoleSkill
	if err := Mysql.DB.Where("role_id = ? AND skill_id = ?", req.RoleID, req.SkillID).First(&roleSkill).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未学习该武学"})
		return
	}

	var skillBase Model.SkillBase
	if err := Mysql.DB.Where("id = ?", req.SkillID).First(&skillBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "武学不存在"})
		return
	}

	leveledUp := false
	roleSkill.Exp += req.Exp

	for roleSkill.Exp >= int64(skillBase.ExpFactor)*int64(roleSkill.Level)*int64(roleSkill.Level) && uint32(roleSkill.Level) < skillBase.MaxLevel {
		roleSkill.Exp -= int64(skillBase.ExpFactor) * int64(roleSkill.Level) * int64(roleSkill.Level)
		roleSkill.Level++
		leveledUp = true
	}

	if err := Mysql.DB.Save(&roleSkill).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": gin.H{"leveled_up": leveledUp, "level": roleSkill.Level, "exp": roleSkill.Exp}})
}

// handleSkillForget 遗忘武学
func handleSkillForget(c *gin.Context) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	result := Mysql.DB.Where("role_id = ? AND skill_id = ?", req.RoleID, req.SkillID).Delete(&Model.RoleSkill{})
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "遗忘失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "未学习该武学"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "遗忘成功"})
}

// handleSkillGetBase 获取武学基础信息
func handleSkillGetBase(c *gin.Context) {
	var req struct {
		SkillID uint32 `json:"skill_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var skill Model.SkillBase
	if err := Mysql.DB.Where("id = ? AND is_active = 1", req.SkillID).First(&skill).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "武学不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": skill})
}

// handleSkillGetAllBase 获取所有武学基础信息
func handleSkillGetAllBase(c *gin.Context) {
	var skills []Model.SkillBase
	if err := Mysql.DB.Where("is_active = 1").Order("type ASC, level ASC").Find(&skills).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": skills})
}

// ========== 道具相关 Handler ==========

// handleItemAdd 添加道具
func handleItemAdd(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
		ItemID uint32 `json:"item_id"`
		Count  uint32 `json:"count"`
		IsBind uint8  `json:"is_bind"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 检查道具是否存在
	var itemBase Model.ItemBase
	if err := Mysql.DB.Where("id = ?", req.ItemID).First(&itemBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "道具不存在"})
		return
	}

	// 查找空格子
	var usedSlots []int
	Mysql.DB.Model(&Model.RoleBag{}).Where("role_id = ?", req.RoleID).Pluck("grid_index", &usedSlots)
	slotMap := make(map[int]bool)
	for _, slot := range usedSlots {
		slotMap[slot] = true
	}
	emptySlot := -1
	for i := 0; i < 80; i++ {
		if !slotMap[i] {
			emptySlot = i
			break
		}
	}
	if emptySlot == -1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "背包已满"})
		return
	}

	bagItem := Model.RoleBag{
		RoleID:    req.RoleID,
		GridIndex: emptySlot,
		ItemID:    req.ItemID,
		Count:     req.Count,
		IsBind:    req.IsBind,
		GetTime:   time.Now(),
	}
	if itemBase.Type == 2 {
		durMax := 100
		bagItem.DurMax = &durMax
		bagItem.DurCurrent = &durMax
	}

	if err := Mysql.DB.Create(&bagItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "添加失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功", "data": emptySlot})
}

// handleItemGetBag 获取背包物品
func handleItemGetBag(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var items []Model.RoleBag
	if err := Mysql.DB.Where("role_id = ?", req.RoleID).Order("grid_index ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// handleItemMove 移动物品
func handleItemMove(c *gin.Context) {
	var req struct {
		RoleID   uint64 `json:"role_id"`
		FromGrid int    `json:"from_grid"`
		ToGrid   int    `json:"to_grid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 检查目标格子是否为空
	var targetItem Model.RoleBag
	err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.ToGrid).First(&targetItem).Error
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "目标格子不为空"})
		return
	}

	// 获取源物品
	var sourceItem Model.RoleBag
	if err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.FromGrid).First(&sourceItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "源格子没有物品"})
		return
	}

	sourceItem.GridIndex = req.ToGrid
	if err := Mysql.DB.Save(&sourceItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "移动失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "移动成功"})
}

// handleItemSplit 拆分物品
func handleItemSplit(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		GridIndex int    `json:"grid_index"`
		Count     uint32 `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var bagItem Model.RoleBag
	if err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.GridIndex).First(&bagItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "物品不存在"})
		return
	}

	if bagItem.Count < req.Count {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "物品数量不足"})
		return
	}

	// 查找空格子
	var usedSlots []int
	Mysql.DB.Model(&Model.RoleBag{}).Where("role_id = ?", req.RoleID).Pluck("grid_index", &usedSlots)
	slotMap := make(map[int]bool)
	for _, slot := range usedSlots {
		slotMap[slot] = true
	}
	emptySlot := -1
	for i := 0; i < 80; i++ {
		if !slotMap[i] {
			emptySlot = i
			break
		}
	}
	if emptySlot == -1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "背包已满"})
		return
	}

	// 减少原物品数量
	bagItem.Count -= req.Count
	Mysql.DB.Save(&bagItem)

	// 创建新物品
	newItem := Model.RoleBag{
		RoleID:    req.RoleID,
		GridIndex: emptySlot,
		ItemID:    bagItem.ItemID,
		Count:     req.Count,
		IsBind:    bagItem.IsBind,
		GetTime:   time.Now(),
	}
	Mysql.DB.Create(&newItem)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "拆分成功"})
}

// handleItemUse 使用道具
func handleItemUse(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		GridIndex int    `json:"grid_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var bagItem Model.RoleBag
	if err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.GridIndex).First(&bagItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "物品不存在"})
		return
	}

	if bagItem.Count <= 1 {
		Mysql.DB.Delete(&bagItem)
	} else {
		bagItem.Count--
		Mysql.DB.Save(&bagItem)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "使用成功"})
}

// handleItemDiscard 丢弃道具
func handleItemDiscard(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		GridIndex int    `json:"grid_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.GridIndex).Delete(&Model.RoleBag{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "丢弃失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "丢弃成功"})
}

// handleItemSell 出售道具
func handleItemSell(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		GridIndex int    `json:"grid_index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var bagItem Model.RoleBag
	if err := Mysql.DB.Where("role_id = ? AND grid_index = ?", req.RoleID, req.GridIndex).First(&bagItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "物品不存在"})
		return
	}

	var itemBase Model.ItemBase
	if err := Mysql.DB.Where("id = ?", bagItem.ItemID).First(&itemBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "道具不存在"})
		return
	}

	sellPrice := itemBase.Price * int(bagItem.Count) / 2

	Mysql.DB.Delete(&bagItem)

	// 增加金币
	Mysql.DB.Model(&Model.Role{}).Where("id = ?", req.RoleID).
		Update("gold", Mysql.DB.Raw("gold + ?", sellPrice))

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "出售成功", "data": sellPrice})
}

// handleItemEquip 穿戴装备
func handleItemEquip(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		BagItemID uint64 `json:"bag_item_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var bagItem Model.RoleBag
	if err := Mysql.DB.Where("id = ? AND role_id = ?", req.BagItemID, req.RoleID).First(&bagItem).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "物品不存在"})
		return
	}

	var itemBase Model.ItemBase
	if err := Mysql.DB.Where("id = ?", bagItem.ItemID).First(&itemBase).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "道具不存在"})
		return
	}

	if itemBase.Type != 2 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该物品不是装备"})
		return
	}

	// 查找该位置是否已有装备
	var existingEquip Model.RoleEquipment
	err := Mysql.DB.Where("role_id = ? AND equip_type = ?", req.RoleID, itemBase.EquipType).First(&existingEquip).Error

	gridIndex := bagItem.GridIndex

	// 卸下原装备
	if err == nil && existingEquip.BagItemID != nil {
		var oldBagItem Model.RoleBag
		Mysql.DB.Where("id = ?", *existingEquip.BagItemID).First(&oldBagItem)
		existingEquip.BagItemID = nil
		Mysql.DB.Save(&existingEquip)
		if oldBagItem.ID > 0 {
			oldBagItem.GridIndex = gridIndex
			Mysql.DB.Save(&oldBagItem)
		}
	}

	// 创建新装备记录
	equip := Model.RoleEquipment{
		RoleID:    req.RoleID,
		EquipType: itemBase.EquipType,
		BagItemID: &req.BagItemID,
		EquipTime: time.Now(),
	}

	Mysql.DB.Delete(&bagItem)

	if err == nil {
		equip.ID = existingEquip.ID
		Mysql.DB.Save(&equip)
	} else {
		Mysql.DB.Create(&equip)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "穿戴成功"})
}

// handleItemUnequip 卸下装备
func handleItemUnequip(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id"`
		EquipType uint8  `json:"equip_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var equip Model.RoleEquipment
	if err := Mysql.DB.Where("role_id = ? AND equip_type = ?", req.RoleID, req.EquipType).First(&equip).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该装备位为空"})
		return
	}

	if equip.BagItemID == nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "该装备位为空"})
		return
	}

	// 查找空格子
	var usedSlots []int
	Mysql.DB.Model(&Model.RoleBag{}).Where("role_id = ?", req.RoleID).Pluck("grid_index", &usedSlots)
	slotMap := make(map[int]bool)
	for _, slot := range usedSlots {
		slotMap[slot] = true
	}
	emptySlot := -1
	for i := 0; i < 80; i++ {
		if !slotMap[i] {
			emptySlot = i
			break
		}
	}
	if emptySlot == -1 {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "背包已满"})
		return
	}

	var bagItem Model.RoleBag
	Mysql.DB.Where("id = ?", *equip.BagItemID).First(&bagItem)
	bagItem.GridIndex = emptySlot
	Mysql.DB.Save(&bagItem)

	equip.BagItemID = nil
	Mysql.DB.Save(&equip)

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "卸下成功"})
}

// handleItemGetEquipped 获取已穿戴装备
func handleItemGetEquipped(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var equips []Model.RoleEquipment
	if err := Mysql.DB.Where("role_id = ? AND bag_item_id IS NOT NULL", req.RoleID).Find(&equips).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": equips})
}

// handleItemGetBase 获取道具基础信息
func handleItemGetBase(c *gin.Context) {
	var req struct {
		ItemID uint32 `json:"item_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var item Model.ItemBase
	if err := Mysql.DB.Where("id = ?", req.ItemID).First(&item).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "道具不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": item})
}

// handleItemGetAllBase 获取所有道具基础信息
func handleItemGetAllBase(c *gin.Context) {
	var items []Model.ItemBase
	if err := Mysql.DB.Order("type ASC, quality ASC, level_req ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": items})
}

// handleItemGetEmptyCount 获取背包空位数
func handleItemGetEmptyCount(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	var count int64
	Mysql.DB.Model(&Model.RoleBag{}).Where("role_id = ?", req.RoleID).Count(&count)
	emptyCount := 80 - int(count)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": emptyCount})
}

// handlePoolStats 连接池状态监控
func handlePoolStats(c *gin.Context) {
	// MySQL连接池状态
	sqlDB, err := Mysql.DB.DB()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "msg": "获取MySQL连接池失败"})
		return
	}

	mysqlStats := gin.H{
		"max_open_connections": sqlDB.Stats().MaxOpenConnections,
		"open_connections":     sqlDB.Stats().OpenConnections,
		"in_use":               sqlDB.Stats().InUse,
		"idle":                 sqlDB.Stats().Idle,
		"wait_count":           sqlDB.Stats().WaitCount,                   // 等待连接的总次数
		"wait_duration_ms":     sqlDB.Stats().WaitDuration.Milliseconds(), // 等待连接的总时间
		"max_idle_closed":      sqlDB.Stats().MaxIdleClosed,               // 因超过最大空闲连接数而关闭的连接数
		"max_lifetime_closed":  sqlDB.Stats().MaxLifetimeClosed,           // 因超过最大生命周期而关闭的连接数
	}

	// Redis连接池状态
	redisStats := gin.H{
		"connected": Redis.RDB != nil,
	}
	if Redis.RDB != nil {
		poolStats := Redis.RDB.PoolStats()
		redisStats["hits"] = poolStats.Hits              // 命中连接池的次数
		redisStats["misses"] = poolStats.Misses          // 未命中连接池的次数
		redisStats["timeouts"] = poolStats.Timeouts      // 获取连接超时的次数
		redisStats["total_conns"] = poolStats.TotalConns // 总连接数
		redisStats["idle_conns"] = poolStats.IdleConns   // 空闲连接数
		redisStats["stale_conns"] = poolStats.StaleConns // 过期连接数
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"mysql": mysqlStats,
			"redis": redisStats,
			"time":  time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}
